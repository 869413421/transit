# Transit - 高性能 API 中转站

Transit 是一个专为高并发场景设计的 API 中转站，支持文本、图片、视频等多模型 API 的统一转发与计费管理。

## 核心特性

- ✅ **高并发支撑**：基于 Redis + Lua 脚本的原子并发控制，单 Key 支持 200+ 并发
- ✅ **多模型适配**：统一支持同步文本对话和异步图片/视频生成
- ✅ **灵活计费**：预扣费 + 失败退费机制，精准按 Token 或次数计费
- ✅ **热更新管理**：Admin API 支持动态添加/下线 Key，无需重启服务
- ✅ **实时监控**：查看全站并发水位和 Key 负载状况

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+
- **ORM**: GORM

## 快速开始

### 1. 环境准备

确保已安装：
- Go 1.21+
- MySQL 8.0+
- Redis 6.0+

### 2. 克隆项目

```bash
git clone https://github.com/869413421/transit.git
cd transit
```

### 3. 配置环境变量

```bash
cp configs/.env.example configs/.env
# 编辑 configs/.env 填入你的数据库和 Redis 配置
```

### 4. 安装依赖

```bash
go mod tidy
```

### 5. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE transit CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 导入 Schema
mysql -u root -p transit < migrations/001_init_schema.sql
```

### 6. 启动服务

```bash
go run cmd/api/main.go
```

服务将在 `http://localhost:8080` 启动。

## Admin API 使用

所有管理接口需要在请求头中携带 `X-Admin-Token`。

### 添加上游 Key

```bash
curl -X POST http://localhost:8080/admin/channels \
  -H "X-Admin-Token: transit-admin-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "APIMart-Key-1",
    "secret_key": "your-apimart-key",
    "base_url": "https://api.apimart.ai",
    "max_concurrency": 200,
    "weight": 10
  }'
```

### 查看所有渠道

```bash
curl http://localhost:8080/admin/channels \
  -H "X-Admin-Token: transit-admin-secret-2026"
```

### 用户充值

```bash
curl -X POST http://localhost:8080/admin/recharge \
  -H "X-Admin-Token: transit-admin-secret-2026" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user-uuid",
    "amount": 100.00,
    "remark": "手动充值"
  }'
```

### 系统监控

```bash
curl http://localhost:8080/admin/monitor \
  -H "X-Admin-Token: transit-admin-secret-2026"
```

## 项目结构

```
transit/
├── cmd/
│   └── api/              # 主程序入口
├── internal/
│   ├── handlers/         # HTTP 处理器
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── services/         # 业务逻辑
├── pkg/
│   ├── pool/             # Redis 并发池
│   └── billing/          # 计费服务
├── configs/              # 配置文件
├── migrations/           # 数据库迁移
└── README.md
```

## 架构设计

详见项目文档中的 `implementation_plan.md`，核心包括：

1. **并发控制**：Redis Lua 脚本保证原子性
2. **负载均衡**：最小连接数策略选择 Key
3. **计费系统**：异步预扣 + 同步后扣双轨制
4. **任务追踪**：本地 Task ID 与上游 Task ID 映射

## 下一步开发

- [ ] 实现用户 API 转发逻辑
- [ ] 接入 APIMart 适配器
- [ ] 添加支付回调接口
- [ ] 开发可视化管理后台

## License

MIT

## 作者

[@869413421](https://github.com/869413421)
