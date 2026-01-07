package api

import (
	"github.com/869413421/transit/internal/handlers"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Router 路由器
type Router struct {
	engine       *gin.Engine
	adminHandler *handlers.AdminHandler
	proxyHandler *handlers.ProxyHandler
}

// NewRouter 创建路由器
func NewRouter(engine *gin.Engine, adminHandler *handlers.AdminHandler, proxyHandler *handlers.ProxyHandler) *Router {
	return &Router{
		engine:       engine,
		adminHandler: adminHandler,
		proxyHandler: proxyHandler,
	}
}

// Setup 配置路由
func (r *Router) Setup() {
	// Swagger 文档
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

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

	// 用户API路由(需要API Key认证)
	// 注意: UserAuth中间件需要在app初始化时传入
	api := r.engine.Group("/api/v1")
	{
		// 文本对话(同步)
		api.POST("/chat/completions", r.proxyHandler.ChatCompletions)

		// 图片生成(异步)
		api.POST("/images/generations", r.proxyHandler.ImageGeneration)

		// 视频生成(异步)
		api.POST("/videos/generations", r.proxyHandler.VideoGeneration)

		// 任务查询
		api.GET("/tasks/:task_id", r.proxyHandler.GetTask)

		// 余额查询
		api.GET("/balance", r.proxyHandler.GetBalance)
	}
}
