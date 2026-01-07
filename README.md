# Transit - 高性能 API 中转站

Transit 是一个专为高并发场景设计的 API 中转站，支持文本、图片、视频等多模型 API 的统一转发与计费管理。

## 核心特性

- ✅ **高并发支撑**：基于 Redis + Lua 脚本的原子并发控制，单 Key 支持 200+ 并发
- ✅ **多模型适配**：统一支持同步文本对话和异步图片/视频生成
- ✅ **灵活计费**：预扣费 + 失败退费机制，精准按 Token 或次数计费
- ✅ **热更新管理**：Admin API 支持动态添加/下线 Key，无需重启服务
- ✅ **工程化架构**：分层设计、依赖注入、结构化日志、优雅关机

## 技术栈

- **语言**: Go 1.21+
- **框架**: Gin
- **数据库**: MySQL 8.0+
- **缓存**: Redis 6.0+
- **ORM**: GORM
- **配置**: Viper
- **日志**: Zap

## 项目结构

```
transit/
├── cmd/
│   └── api/              # 主程序入口
├── internal/
│   ├── api/              # 路由配置
│   ├── app/              # 应用生命周期管理
│   ├── config/           # 配置管理
│   ├── database/         # 数据库连接
│   ├── handlers/         # HTTP 处理器
│   ├── models/           # 数据模型
│   ├── repository/       # 数据访问层
│   └── services/         # 业务逻辑层
├── pkg/
│   ├── billing/          # 计费服务
│   ├── logger/           # 日志服务
│   └── pool/             # Redis 并发池
├── configs/              # 配置文件
├── migrations/           # 数据库迁移
└── Makefile             # 开发工具
```

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

### 3. 配置应用

```bash
# 复制配置文件
cp configs/config.yaml configs/config.local.yaml

# 编辑 config.local.yaml 填入你的配置
# 或使用环境变量
export DATABASE_HOST=127.0.0.1
export DATABASE_PASSWORD=your_password
export REDIS_ADDR=127.0.0.1:6379
```

### 4. 初始化数据库

```bash
# 创建数据库
mysql -u root -p -e "CREATE DATABASE transit CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 运行迁移
make migrate
```

### 5. 启动服务

```bash
# 开发模式
make run

# 或编译后运行
make build
./bin/transit
```

服务将在 `http://localhost:8080` 启动。

## 开发命令

```bash
make help          # 查看所有可用命令
make build         # 编译应用程序
make run           # 运行应用程序
make test          # 运行测试
make clean         # 清理编译产物
make tidy          # 整理依赖
```

## Admin API 使用

所有管理接口需要在请求头中携带 `X-Admin-Token`（在 config.yaml 中配置）。

### 添加上游 Key

```bash
curl -X POST http://localhost:8080/admin/channels \
  -H "X-Admin-Token: your-admin-token" \
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
  -H "X-Admin-Token: your-admin-token"
```

### 用户充值

```bash
curl -X POST http://localhost:8080/admin/recharge \
  -H "X-Admin-Token: your-admin-token" \
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
  -H "X-Admin-Token: your-admin-token"
```

## 配置说明

配置文件位于 `configs/config.yaml`，支持以下配置项：

```yaml
server:
  port: "8080"
  environment: "development"  # development, production

database:
  host: "127.0.0.1"
  port: "3306"
  user: "root"
  password: "password"
  dbname: "transit"

redis:
  addr: "127.0.0.1:6379"
  password: ""
  db: 0

admin:
  token: "your-admin-token"  # 请修改为强密码
```

也可以通过环境变量覆盖配置，例如：
- `DATABASE_HOST`
- `DATABASE_PASSWORD`
- `REDIS_ADDR`
- `ADMIN_TOKEN`

## 下一步开发

- [ ] 实现用户 API 转发逻辑
- [ ] 接入 APIMart 适配器
- [ ] 添加支付回调接口
- [ ] 开发可视化管理后台

## License

MIT

## 作者

[@869413421](https://github.com/869413421)
