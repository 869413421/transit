package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/869413421/transit/internal/handlers"
	"github.com/869413421/transit/internal/models"
	"github.com/869413421/transit/pkg/billing"
	"github.com/869413421/transit/pkg/pool"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func main() {
	// 加载配置
	dbDSN := getEnv("DB_DSN", "root:password@tcp(127.0.0.1:3306)/transit?charset=utf8mb4&parseTime=True&loc=Local")
	redisAddr := getEnv("REDIS_ADDR", "127.0.0.1:6379")
	port := getEnv("PORT", "8080")

	// 连接数据库
	db, err := gorm.Open(mysql.Open(dbDSN), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// 自动迁移
	if err := db.AutoMigrate(
		&models.User{},
		&models.UserAPIKey{},
		&models.Channel{},
		&models.Task{},
		&models.BillingLog{},
	); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// 连接 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	// 初始化服务
	redisPool := pool.NewRedisPool(rdb)
	billingService := billing.NewService(rdb)

	// 初始化处理器
	adminHandler := handlers.NewAdminHandler(db, redisPool, billingService)

	// 设置路由
	r := gin.Default()

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 管理后台路由
	admin := r.Group("/admin", adminHandler.AdminAuth())
	{
		admin.POST("/channels", adminHandler.AddChannel)
		admin.GET("/channels", adminHandler.ListChannels)
		admin.DELETE("/channels/:id", adminHandler.DeleteChannel)
		admin.POST("/recharge", adminHandler.Recharge)
		admin.GET("/monitor", adminHandler.Monitor)
	}

	// 启动服务
	addr := fmt.Sprintf(":%s", port)
	log.Printf("Transit API Relay Station starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
