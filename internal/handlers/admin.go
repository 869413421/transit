package handlers

import (
	"net/http"

	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/pool"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AdminHandler 管理后台处理器
type AdminHandler struct {
	db      *gorm.DB
	pool    *pool.RedisPool
	billing *billing.Service
}

// NewAdminHandler 创建管理处理器
func NewAdminHandler(db *gorm.DB, pool *pool.RedisPool, billing *billing.Service) *AdminHandler {
	return &AdminHandler{
		db:      db,
		pool:    pool,
		billing: billing,
	}
}

// AdminAuth 管理员鉴权中间件
func (h *AdminHandler) AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := c.GetHeader("X-Admin-Token")
		// TODO: 从配置或数据库读取管理员 Token
		if token != "transit-admin-secret-2026" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// AddChannel 添加渠道
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

	channel := &models.Channel{
		ID:             uuid.New().String(),
		Name:           req.Name,
		SecretKey:      req.SecretKey,
		BaseURL:        req.BaseURL,
		MaxConcurrency: req.MaxConcurrency,
		Weight:         req.Weight,
		IsActive:       true,
	}

	if channel.MaxConcurrency == 0 {
		channel.MaxConcurrency = 200
	}
	if channel.Weight == 0 {
		channel.Weight = 10
	}

	if err := h.db.Create(channel).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel added successfully", "channel": channel})
}

// ListChannels 列出所有渠道
func (h *AdminHandler) ListChannels(c *gin.Context) {
	var channels []models.Channel
	if err := h.db.Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}

	// 获取实时并发数
	for i := range channels {
		concurrency, _ := h.pool.GetConcurrency(c.Request.Context(), channels[i].ID)
		channels[i].CurrentConcurrency = concurrency
	}

	c.JSON(http.StatusOK, gin.H{"channels": channels})
}

// DeleteChannel 删除渠道
func (h *AdminHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	if err := h.db.Delete(&models.Channel{}, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// Recharge 用户充值
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to recharge"})
		return
	}

	// 记录流水
	log := &models.BillingLog{
		ID:      uuid.New().String(),
		UserID:  req.UserID,
		Amount:  req.Amount,
		LogType: "recharge",
		Remark:  req.Remark,
	}
	h.db.Create(log)

	c.JSON(http.StatusOK, gin.H{"message": "Recharge successful", "new_balance": req.Amount})
}

// Monitor 系统监控
func (h *AdminHandler) Monitor(c *gin.Context) {
	var channels []models.Channel
	h.db.Where("is_active = ?", true).Find(&channels)

	var totalConcurrency int
	channelStats := make([]map[string]interface{}, 0)

	for _, ch := range channels {
		concurrency, _ := h.pool.GetConcurrency(c.Request.Context(), ch.ID)
		totalConcurrency += concurrency

		channelStats = append(channelStats, map[string]interface{}{
			"id":          ch.ID,
			"name":        ch.Name,
			"concurrency": concurrency,
			"max":         ch.MaxConcurrency,
			"usage":       float64(concurrency) / float64(ch.MaxConcurrency) * 100,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"total_concurrency": totalConcurrency,
		"channels":          channelStats,
		"status":            "healthy",
	})
}
