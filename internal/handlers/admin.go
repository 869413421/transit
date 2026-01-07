package handlers

import (
	"net/http"
	"time"

	"github.com/869413421/transit/internal/config"
	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/internal/services"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/logger"
	"github.com/869413421/transit/pkg/pool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	cfg            *config.Config
	channelService services.ChannelService
	userRepo       repository.UserRepository
	billing        *billing.Service
	pool           *pool.RedisPool
}

// NewAdminHandler 创建管理处理器
func NewAdminHandler(
	cfg *config.Config,
	channelService services.ChannelService,
	userRepo repository.UserRepository,
	billing *billing.Service,
	pool *pool.RedisPool,
) *AdminHandler {
	return &AdminHandler{
		cfg:            cfg,
		channelService: channelService,
		userRepo:       userRepo,
		billing:        billing,
		pool:           pool,
	}
}

// AdminAuth 管理员鉴权中间件
func (h *AdminHandler) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Token")
		if token != h.cfg.Admin.Token {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AddChannel 添加渠道
// @Summary 添加上游渠道
// @Description 添加一个新的上游 API Key 到渠道池
// @Tags Admin
// @Accept json
// @Produce json
// @Security AdminToken
// @Param channel body object{name=string,secret_key=string,base_url=string,max_concurrency=int,weight=int} true "渠道信息"
// @Success 200 {object} object{message=string,channel=models.Channel}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/channels [post]
func (h *AdminHandler) AddChannel(c *gin.Context) {
	var req struct {
		Name           string `json:"name" binding:"required"`
		SecretKey      string `json:"secret_key" binding:"required"`
		BaseURL        string `json:"base_url"`
		MaxConcurrency int    `json:"max_concurrency"`
		Weight         int    `json:"weight"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	now := time.Now()
	channel := &models.Channel{
		ID:             uuid.New().String(),
		Name:           req.Name,
		SecretKey:      req.SecretKey,
		BaseURL:        req.BaseURL,
		MaxConcurrency: req.MaxConcurrency,
		Weight:         req.Weight,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if channel.MaxConcurrency == 0 {
		channel.MaxConcurrency = 200
	}
	if channel.Weight == 0 {
		channel.Weight = 10
	}

	if err := h.channelService.Create(c.Request.Context(), channel); err != nil {
		logger.Error("Failed to create channel", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	logger.Info("Channel created", zap.String("id", channel.ID), zap.String("name", channel.Name))
	c.JSON(http.StatusOK, gin.H{"message": "Channel added successfully", "channel": channel})
}

// ListChannels 列出所有渠道
// @Summary 查看所有渠道
// @Description 获取所有上游渠道列表及实时并发数
// @Tags Admin
// @Produce json
// @Security AdminToken
// @Success 200 {object} object{channels=[]models.Channel}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/channels [get]
func (h *AdminHandler) ListChannels(c *gin.Context) {
	channels, err := h.channelService.GetAllWithConcurrency(c.Request.Context())
	if err != nil {
		logger.Error("Failed to fetch channels", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// DeleteChannel 删除渠道
// @Summary 删除渠道
// @Description 从渠道池中删除指定的上游 Key
// @Tags Admin
// @Produce json
// @Security AdminToken
// @Param id path string true "渠道 ID"
// @Success 200 {object} object{message=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/channels/{id} [delete]
func (h *AdminHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if err := h.channelService.Delete(c.Request.Context(), id); err != nil {
		logger.Error("Failed to delete channel", zap.String("id", id), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	logger.Info("Channel deleted", zap.String("id", id))
	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// Recharge 用户充值
// @Summary 用户充值
// @Description 为指定用户账户充值
// @Tags Admin
// @Accept json
// @Produce json
// @Security AdminToken
// @Param recharge body object{user_id=string,amount=number,remark=string} true "充值信息"
// @Success 200 {object} object{message=string}
// @Failure 400 {object} object{error=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/recharge [post]
func (h *AdminHandler) Recharge(c *gin.Context) {
	var req struct {
		UserID string  `json:"user_id" binding:"required"`
		Amount float64 `json:"amount" binding:"required,gt=0"`
		Remark string  `json:"remark"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Redis 充值
	if err := h.billing.Recharge(c.Request.Context(), req.UserID, req.Amount); err != nil {
		logger.Error("Failed to recharge", zap.String("user_id", req.UserID), zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recharge"})
		return
	}

	logger.Info("User recharged", zap.String("user_id", req.UserID), zap.Float64("amount", req.Amount))
	c.JSON(http.StatusOK, gin.H{"message": "Recharge successful"})
}

// Monitor 系统监控
// @Summary 系统监控
// @Description 查看全站实时并发水位和渠道负载状况
// @Tags Admin
// @Produce json
// @Security AdminToken
// @Success 200 {object} object{total_concurrency=int,channels=[]object,status=string}
// @Failure 401 {object} object{error=string}
// @Failure 500 {object} object{error=string}
// @Router /admin/monitor [get]
func (h *AdminHandler) Monitor(c *gin.Context) {
	channels, err := h.channelService.GetAllWithConcurrency(c.Request.Context())
	if err != nil {
		logger.Error("Failed to fetch channels for monitoring", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monitoring data"})
		return
	}

	var totalConcurrency int
	channelStats := make([]map[string]interface{}, 0)

	for _, ch := range channels {
		if !ch.IsActive {
			continue
		}
		totalConcurrency += ch.CurrentConcurrency

		usage := float64(0)
		if ch.MaxConcurrency > 0 {
			usage = float64(ch.CurrentConcurrency) / float64(ch.MaxConcurrency) * 100
		}

		channelStats = append(channelStats, map[string]interface{}{
			"id":          ch.ID,
			"name":        ch.Name,
			"concurrency": ch.CurrentConcurrency,
			"max":         ch.MaxConcurrency,
			"usage":       usage,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_concurrency": totalConcurrency,
		"channels":          channelStats,
		"status":            "healthy",
	})
}
