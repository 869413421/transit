.PHONY: help build run dev test clean docker-build docker-run

help: ## 显示帮助信息
	@echo "可用命令:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译应用程序
	@echo "编译 Transit..."
	@go build -o bin/transit cmd/api/main.go

run: ## 运行应用程序
	@echo "启动 Transit..."
	@go run cmd/api/main.go

dev: ## 使用 Air 热重载运行
	@echo "使用 Air 启动开发模式..."
	@air

test: ## 运行测试
	@echo "运行测试..."
	@go test -v ./...

clean: ## 清理编译产物
	@echo "清理编译产物..."
	@rm -rf bin/
	@rm -rf tmp/

docker-build: ## 构建 Docker 镜像
	@echo "构建 Docker 镜像..."
	@docker build -t transit:latest .

docker-run: ## 运行 Docker 容器
	@echo "运行 Docker 容器..."
	@docker run -p 8080:8080 --env-file configs/.env transit:latest

tidy: ## 整理依赖
	@echo "整理 Go 依赖..."
	@go mod tidy

migrate: ## 运行数据库迁移
	@echo "运行数据库迁移..."
	@mysql -u root -p transit < migrations/001_init_schema.sql
