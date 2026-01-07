package app

import (
	"context"
	"fmt"

	"github.com/869413421/transit/internal/api"
	"github.com/869413421/transit/internal/config"
	"github.com/869413421/transit/internal/database"
	"github.com/869413421/transit/internal/handlers"
	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/internal/repository"
	"github.com/869413421/transit/internal/services"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/logger"
	"github.com/869413421/transit/pkg/pool"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// App 是应用程序的核心容器，负责管理依赖注入和生命周期
type App struct {
	cfg   *config.Config
	db    *gorm.DB
	redis *redis.Client
}

// NewApp 创建一个新的 App 实例
func NewApp(cfg *config.Config) *App {
	return &App{
		cfg: cfg,
	}
}

// Start 初始化并启动应用程序
func (a *App) Start() error {
	// 1. 连接数据库
	db, err := database.NewMySQLDB(&a.cfg.Database)
	if err != nil {
		return fmt.Errorf("连接数据库失败: %w", err)
	}
	a.db = db
	logger.Info("数据库连接成功")

	// 2. 自动迁移数据库
	if err := a.db.AutoMigrate(
		&models.User{},
		&models.UserAPIKey{},
		&models.Channel{},
		&models.Task{},
		&models.BillingLog{},
	); err != nil {
		return fmt.Errorf("数据库迁移失败: %w", err)
	}
	logger.Info("数据库迁移完成")

	// 3. 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     a.cfg.Redis.Addr,
		Password: a.cfg.Redis.Password,
		DB:       a.cfg.Redis.DB,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("连接 Redis 失败: %w", err)
	}
	a.redis = rdb
	logger.Info("Redis 连接成功")

	// 4. 初始化持久层 (Repositories)
	channelRepo := repository.NewChannelRepository(a.db)
	userRepo := repository.NewUserRepository(a.db)

	// 5. 初始化基础设施层
	redisPool := pool.NewRedisPool(a.redis)
	billingService := billing.NewService(a.redis)

	// 6. 初始化业务逻辑层 (Services)
	channelService := services.NewChannelService(channelRepo, redisPool)

	// 7. 初始化接口层 (Handlers)
	adminHandler := handlers.NewAdminHandler(
		a.cfg,
		channelService,
		userRepo,
		billingService,
		redisPool,
	)

	// 8. 配置路由
	if a.cfg.Server.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	router := api.NewRouter(engine, adminHandler)
	router.Setup()

	// 9. 启动 HTTP 服务
	addr := ":" + a.cfg.Server.Port
	logger.Info("服务器正在启动", zap.String("address", addr), zap.String("environment", a.cfg.Server.Environment))
	return engine.Run(addr)
}

// Stop 执行优雅关机流程
func (a *App) Stop() {
	if a.db != nil {
		sqlDB, _ := a.db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
		logger.Info("数据库连接已关闭")
	}
	if a.redis != nil {
		a.redis.Close()
		logger.Info("Redis 连接已关闭")
	}
}
