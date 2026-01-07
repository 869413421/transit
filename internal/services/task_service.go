package services

import (
	"context"
	"time"

	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// TaskService 任务服务接口
type TaskService interface {
	CreateTask(ctx context.Context, userID, channelID, taskType, modelName, upstreamTaskID string, cost float64) (*models.Task, error)
	GetTask(ctx context.Context, taskID string) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, taskID, status, resultURL string) error
	GetPendingTasks(ctx context.Context, limit int) ([]*models.Task, error)
}

type taskService struct {
	taskRepo repository.TaskRepository
}

// NewTaskService 创建任务服务
func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, userID, channelID, taskType, modelName, upstreamTaskID string, cost float64) (*models.Task, error) {
	now := time.Now()
	task := &models.Task{
		ID:             uuid.New().String(),
		UserID:         userID,
		ChannelID:      channelID,
		Type:           taskType,
		ModelName:      modelName,
		UpstreamTaskID: upstreamTaskID,
		Status:         "running",
		Cost:           cost,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		logger.Error("Failed to create task", zap.Error(err))
		return nil, err
	}

	logger.Info("Task created",
		zap.String("task_id", task.ID),
		zap.String("type", taskType),
		zap.String("model", modelName),
	)

	return task, nil
}

func (s *taskService) GetTask(ctx context.Context, taskID string) (*models.Task, error) {
	return s.taskRepo.FindByID(ctx, taskID)
}

func (s *taskService) UpdateTaskStatus(ctx context.Context, taskID, status, resultURL string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}

	task.Status = status
	task.ResultURL = resultURL
	task.UpdatedAt = time.Now()

	if err := s.taskRepo.Update(ctx, task); err != nil {
		logger.Error("Failed to update task",
			zap.String("task_id", taskID),
			zap.Error(err),
		)
		return err
	}

	logger.Info("Task updated",
		zap.String("task_id", taskID),
		zap.String("status", status),
	)

	return nil
}

func (s *taskService) GetPendingTasks(ctx context.Context, limit int) ([]*models.Task, error) {
	return s.taskRepo.FindPendingTasks(ctx, limit)
}
