package main

import (
	"log"

	"github.com/869413421/transit/internal/app"
	"github.com/869413421/transit/internal/config"
	"github.com/869413421/transit/pkg/logger"
)

func main() {
	// 1. 加载应用程序配置
	// 尝试从本地 config.yaml 或环境变量加载配置信息
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("无法加载配置信息: %v", err)
	}

	// 2. 初始化全局日志服务
	// 根据运行环境（开发/生产）设置不同的日志输出级别和格式
	if err := logger.Init(cfg.Server.Environment); err != nil {
		log.Fatalf("无法初始化日志服务: %v", err)
	}
	defer logger.Sync() // 确保在程序退出前所有日志都已刷入磁盘

	// 3. 初始化并启动应用程序容器
	// App 封装了所有的依赖项（DB, Redis）以及 HTTP 服务的启动逻辑
	application := app.NewApp(cfg)
	defer application.Stop() // 程序退出时释放持有的资源（如数据库连接池）

	// 启动应用程序并监听配置的端口
	if err := application.Start(); err != nil {
		log.Fatalf("应用程序运行失败: %v", err)
	}
}
