package poller

import (
	"context"
	"time"

	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/internal/services"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/loadbalancer"
	"github.com/869413421/transit/pkg/logger"
	"github.com/869413421/transit/pkg/upstream"
	"go.uber.org/zap"
)

// Poller 异步任务轮询器
type Poller struct {
	taskService  services.TaskService
	channelRepo  repository.ChannelRepository
	selector     *loadbalancer.Selector
	billing      *billing.Service
	pollInterval time.Duration
	batchSize    int
	stopChan     chan struct{}
}

// NewPoller 创建轮询器
func NewPoller(
	taskService services.TaskService,
	channelRepo repository.ChannelRepository,
	selector *loadbalancer.Selector,
	billing *billing.Service,
) *Poller {
	return &Poller{
		taskService:  taskService,
		channelRepo:  channelRepo,
		selector:     selector,
		billing:      billing,
		pollInterval: 10 * time.Second, // 每10秒轮询一次
		batchSize:    100,              // 每次处理100个任务
		stopChan:     make(chan struct{}),
	}
}

// Start 启动轮询器
func (p *Poller) Start(ctx context.Context) {
	logger.Info("Task poller started",
		zap.Duration("poll_interval", p.pollInterval),
		zap.Int("batch_size", p.batchSize),
	)

	ticker := time.NewTicker(p.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Task poller stopped")
			return
		case <-p.stopChan:
			logger.Info("Task poller stopped")
			return
		case <-ticker.C:
			p.pollTasks(ctx)
		}
	}
}

// Stop 停止轮询器
func (p *Poller) Stop() {
	close(p.stopChan)
}

// pollTasks 轮询任务
func (p *Poller) pollTasks(ctx context.Context) {
	// 获取待处理任务
	tasks, err := p.taskService.GetPendingTasks(ctx, p.batchSize)
	if err != nil {
		logger.Error("Failed to get pending tasks", zap.Error(err))
		return
	}

	if len(tasks) == 0 {
		return
	}

	logger.Debug("Polling tasks", zap.Int("count", len(tasks)))

	// 处理每个任务
	for _, task := range tasks {
		if err := p.processTask(ctx, task); err != nil {
			logger.Error("Failed to process task",
				zap.String("task_id", task.ID),
				zap.Error(err),
			)
		}
	}
}

// processTask 处理单个任务
func (p *Poller) processTask(ctx context.Context, task *models.Task) error {
	// 获取渠道信息
	channel, err := p.channelRepo.FindByID(ctx, task.ChannelID)
	if err != nil {
		return err
	}

	// 创建APIMart适配器
	adapter := upstream.NewAPIMartAdapter(channel.BaseURL, channel.SecretKey)

	// 查询上游任务状态
	status, err := adapter.GetTaskStatus(ctx, task.UpstreamTaskID)
	if err != nil {
		logger.Warn("Failed to get upstream task status",
			zap.String("task_id", task.ID),
			zap.String("upstream_task_id", task.UpstreamTaskID),
			zap.Error(err),
		)
		return err
	}

	// 根据状态更新任务
	switch status.Status {
	case "completed":
		// 任务完成
		resultURL := ""
		if len(status.Result.Images) > 0 {
			resultURL = status.Result.Images[0]
		} else if len(status.Result.Videos) > 0 {
			resultURL = status.Result.Videos[0]
		}

		if err := p.taskService.UpdateTaskStatus(ctx, task.ID, "completed", resultURL); err != nil {
			return err
		}

		// 释放并发位
		p.selector.ReleaseChannel(ctx, task.ChannelID)

		logger.Info("Task completed",
			zap.String("task_id", task.ID),
			zap.String("result_url", resultURL),
		)

	case "failed", "cancelled":
		// 任务失败,退费
		if err := p.taskService.UpdateTaskStatus(ctx, task.ID, status.Status, ""); err != nil {
			return err
		}

		// 退费
		if err := p.billing.Refund(ctx, task.UserID, task.Cost); err != nil {
			logger.Error("Failed to refund",
				zap.String("task_id", task.ID),
				zap.Error(err),
			)
		}

		// 释放并发位
		p.selector.ReleaseChannel(ctx, task.ChannelID)

		logger.Warn("Task failed",
			zap.String("task_id", task.ID),
			zap.String("status", status.Status),
			zap.String("error", status.Error.Message),
		)

	case "pending", "processing":
		// 任务进行中,继续等待
		logger.Debug("Task in progress",
			zap.String("task_id", task.ID),
			zap.String("status", status.Status),
			zap.Int("progress", status.Progress),
		)
	}

	return nil
}
