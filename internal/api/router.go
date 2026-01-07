package api

import (
	"github.com/869413421/transit/internal/handlers"
	"github.com/gin-gonic/gin"
)

// Router 路由器
type Router struct {
	engine       *gin.Engine
	adminHandler *handlers.AdminHandler
}

// NewRouter 创建路由器
func NewRouter(engine *gin.Engine, adminHandler *handlers.AdminHandler) *Router {
	return &Router{
		engine:       engine,
		adminHandler: adminHandler,
	}
}

// Setup 配置路由
func (r *Router) Setup() {
	// 健康检查
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 管理后台路由
	admin := r.engine.Group("/admin", r.adminHandler.AdminAuth())
	{
		admin.POST("/channels", r.adminHandler.AddChannel)
		admin.GET("/channels", r.adminHandler.ListChannels)
		admin.DELETE("/channels/:id", r.adminHandler.DeleteChannel)
		admin.POST("/recharge", r.adminHandler.Recharge)
		admin.GET("/monitor", r.adminHandler.Monitor)
	}
}
