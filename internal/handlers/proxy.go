package handlers

import (
	"net/http"
	"time"

	"github.com/869413421/transit/internal/config"
	"github.com/869413421/transit/internal/services"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/loadbalancer"
	"github.com/869413421/transit/pkg/logger"
	"github.com/869413421/transit/pkg/upstream"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ProxyHandler 代理转发处理器
type ProxyHandler struct {
	cfg         *config.Config
	selector    *loadbalancer.Selector
	taskService services.TaskService
	billing     *billing.Service
}

// NewProxyHandler 创建代理转发处理器
func NewProxyHandler(
	cfg *config.Config,
	selector *loadbalancer.Selector,
	taskService services.TaskService,
	billing *billing.Service,
) *ProxyHandler {
	return &ProxyHandler{
		cfg:         cfg,
		selector:    selector,
		taskService: taskService,
		billing:     billing,
	}
}

// ChatCompletions 文本对话(同步)
// @Summary 文本对话
// @Description 同步文本对话API,兼容OpenAI格式
// @Tags Proxy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body upstream.ChatCompletionRequest true "对话请求"
// @Success 200 {object} upstream.ChatCompletionResponse
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/chat/completions [post]
func (h *ProxyHandler) ChatCompletions(c *gin.Context) {
	// 获取用户ID
	userID, _ := c.Get("user_id")

	// 解析请求
	var req upstream.ChatCompletionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取模型配置
	modelCfg := h.cfg.Models.GetModelByName(req.Model)
	if modelCfg == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported model: " + req.Model})
		return
	}

	// 检查余额(预估费用,假设1000 tokens)
	estimatedCost := 1000 * (modelCfg.PricePer1KInputTokens + modelCfg.PricePer1KOutputTokens) / 1000
	balance, err := h.billing.GetBalance(c.Request.Context(), userID.(string))
	if err != nil || balance < estimatedCost {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		return
	}

	// 选择渠道
	channel, err := h.selector.SelectChannel(c.Request.Context())
	if err != nil {
		logger.Error("Failed to select channel", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available channels"})
		return
	}
	defer h.selector.ReleaseChannel(c.Request.Context(), channel.ID)

	// 创建APIMart适配器
	adapter := upstream.NewAPIMartAdapter(channel.BaseURL, channel.SecretKey)

	// 转发请求
	startTime := time.Now()
	resp, err := adapter.ChatCompletion(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Upstream request failed",
			zap.String("channel_id", channel.ID),
			zap.Error(err),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upstream request failed"})
		return
	}

	// 计算实际费用
	inputCost := float64(resp.Usage.PromptTokens) * modelCfg.PricePer1KInputTokens / 1000
	outputCost := float64(resp.Usage.CompletionTokens) * modelCfg.PricePer1KOutputTokens / 1000
	actualCost := inputCost + outputCost

	// 扣费
	if err := h.billing.PostDeduct(c.Request.Context(), userID.(string), resp.Usage.TotalTokens, actualCost/float64(resp.Usage.TotalTokens)*1000); err != nil {
		logger.Error("Failed to deduct balance", zap.Error(err))
	}

	logger.Info("Chat completion success",
		zap.String("user_id", userID.(string)),
		zap.String("model", req.Model),
		zap.Int("total_tokens", resp.Usage.TotalTokens),
		zap.Float64("cost", actualCost),
		zap.Duration("latency", time.Since(startTime)),
	)

	c.JSON(http.StatusOK, resp)
}

// ImageGeneration 图片生成(异步)
// @Summary 图片生成
// @Description 异步图片生成API,返回任务ID
// @Tags Proxy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body upstream.ImageGenerationRequest true "图片生成请求"
// @Success 200 {object} object{task_id=string,status=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/images/generations [post]
func (h *ProxyHandler) ImageGeneration(c *gin.Context) {
	// 获取用户ID
	userID, _ := c.Get("user_id")

	// 解析请求
	var req upstream.ImageGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取模型配置
	modelCfg := h.cfg.Models.GetModelByName(req.Model)
	if modelCfg == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported model: " + req.Model})
		return
	}

	// 预扣费
	cost := modelCfg.PricePerGeneration
	if err := h.billing.PreDeduct(c.Request.Context(), userID.(string), cost); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		return
	}

	// 选择渠道
	channel, err := h.selector.SelectChannel(c.Request.Context())
	if err != nil {
		// 退费
		h.billing.Refund(c.Request.Context(), userID.(string), cost)
		logger.Error("Failed to select channel", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available channels"})
		return
	}

	// 创建APIMart适配器
	adapter := upstream.NewAPIMartAdapter(channel.BaseURL, channel.SecretKey)

	// 转发请求
	resp, err := adapter.ImageGeneration(c.Request.Context(), &req)
	if err != nil {
		// 退费并释放并发位
		h.billing.Refund(c.Request.Context(), userID.(string), cost)
		h.selector.ReleaseChannel(c.Request.Context(), channel.ID)
		logger.Error("Upstream request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upstream request failed"})
		return
	}

	// 创建任务记录
	task, err := h.taskService.CreateTask(
		c.Request.Context(),
		userID.(string),
		channel.ID,
		"async",
		req.Model,
		resp.TaskID,
		cost,
	)
	if err != nil {
		logger.Error("Failed to create task", zap.Error(err))
	}

	logger.Info("Image generation submitted",
		zap.String("user_id", userID.(string)),
		zap.String("task_id", task.ID),
		zap.String("upstream_task_id", resp.TaskID),
	)

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  resp.Status,
	})
}

// VideoGeneration 视频生成(异步)
// @Summary 视频生成
// @Description 异步视频生成API,返回任务ID
// @Tags Proxy
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body upstream.VideoGenerationRequest true "视频生成请求"
// @Success 200 {object} object{task_id=string,status=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/videos/generations [post]
func (h *ProxyHandler) VideoGeneration(c *gin.Context) {
	// 获取用户ID
	userID, _ := c.Get("user_id")

	// 解析请求
	var req upstream.VideoGenerationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: " + err.Error()})
		return
	}

	// 获取模型配置
	modelCfg := h.cfg.Models.GetModelByName(req.Model)
	if modelCfg == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported model: " + req.Model})
		return
	}

	// 预扣费
	cost := modelCfg.PricePerGeneration
	if err := h.billing.PreDeduct(c.Request.Context(), userID.(string), cost); err != nil {
		c.JSON(http.StatusPaymentRequired, gin.H{"error": "Insufficient balance"})
		return
	}

	// 选择渠道
	channel, err := h.selector.SelectChannel(c.Request.Context())
	if err != nil {
		// 退费
		h.billing.Refund(c.Request.Context(), userID.(string), cost)
		logger.Error("Failed to select channel", zap.Error(err))
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "No available channels"})
		return
	}

	// 创建APIMart适配器
	adapter := upstream.NewAPIMartAdapter(channel.BaseURL, channel.SecretKey)

	// 转发请求
	resp, err := adapter.VideoGeneration(c.Request.Context(), &req)
	if err != nil {
		// 退费并释放并发位
		h.billing.Refund(c.Request.Context(), userID.(string), cost)
		h.selector.ReleaseChannel(c.Request.Context(), channel.ID)
		logger.Error("Upstream request failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Upstream request failed"})
		return
	}

	// 创建任务记录
	task, err := h.taskService.CreateTask(
		c.Request.Context(),
		userID.(string),
		channel.ID,
		"async",
		req.Model,
		resp.TaskID,
		cost,
	)
	if err != nil {
		logger.Error("Failed to create task", zap.Error(err))
	}

	logger.Info("Video generation submitted",
		zap.String("user_id", userID.(string)),
		zap.String("task_id", task.ID),
		zap.String("upstream_task_id", resp.TaskID),
	)

	c.JSON(http.StatusOK, gin.H{
		"task_id": task.ID,
		"status":  resp.Status,
	})
}

// GetTask 查询任务状态
// @Summary 查询任务
// @Description 查询异步任务的状态和结果
// @Tags Proxy
// @Produce json
// @Security BearerAuth
// @Param task_id path string true "任务ID"
// @Success 200 {object} models.Task
// @Failure 404 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/tasks/{task_id} [get]
func (h *ProxyHandler) GetTask(c *gin.Context) {
	taskID := c.Param("task_id")

	task, err := h.taskService.GetTask(c.Request.Context(), taskID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetBalance 查询余额
// @Summary 查询余额
// @Description 查询用户账户余额
// @Tags Proxy
// @Produce json
// @Security BearerAuth
// @Success 200 {object} object{balance=number}
// @Failure 500 {object} object{error=string}
// @Router /api/v1/balance [get]
func (h *ProxyHandler) GetBalance(c *gin.Context) {
	userID, _ := c.Get("user_id")

	balance, err := h.billing.GetBalance(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get balance"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})
}
