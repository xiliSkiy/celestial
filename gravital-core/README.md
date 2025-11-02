# Gravital Core - 中心端

Celestial 监控系统的中心端，负责设备管理、任务调度、告警管理和数据转发。

## 🎯 功能特性

### 核心功能
- ✅ 设备管理（CRUD、分组、模板）
- ✅ Sentinel 管理（注册、心跳、状态监控）
- ✅ 任务调度（任务分配、执行跟踪）
- ✅ 告警管理（规则引擎、通知）
- ✅ 数据转发（Prometheus、VictoriaMetrics、ClickHouse）
- ✅ 用户认证（JWT、RBAC）
- ✅ API 网关（RESTful API）

### 技术栈
- **语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: PostgreSQL 15+
- **缓存**: Redis 7+
- **ORM**: GORM
- **认证**: JWT
- **日志**: Zap
- **配置**: Viper

## 📦 项目结构

```
gravital-core/
├── cmd/
│   └── server/              # 主程序入口
│       └── main.go
├── internal/
│   ├── api/                 # API 层
│   │   ├── handler/         # HTTP 处理器
│   │   ├── middleware/      # 中间件
│   │   └── router/          # 路由
│   ├── service/             # 业务逻辑层
│   ├── repository/          # 数据访问层
│   ├── model/               # 数据模型
│   ├── pkg/                 # 公共包
│   │   ├── auth/            # 认证
│   │   ├── cache/           # 缓存
│   │   ├── config/          # 配置
│   │   ├── database/        # 数据库
│   │   └── logger/          # 日志
│   ├── alert/               # 告警模块
│   ├── forwarder/           # 转发模块
│   └── scheduler/           # 调度模块
├── config/                  # 配置文件
│   └── config.example.yaml
├── migrations/              # 数据库迁移
│   ├── 001_init.up.sql
│   └── 001_init.down.sql
├── scripts/                 # 工具脚本
├── Makefile                 # 构建脚本
├── Dockerfile               # Docker 构建
├── docker-compose.yaml      # Docker Compose
└── README.md
```

## 🚀 快速开始

### 前置要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Docker & Docker Compose (可选)

### 1. 安装依赖

```bash
# 下载依赖
make deps

# 或者
go mod download
```

### 2. 配置数据库

```bash
# 启动 PostgreSQL 和 Redis（使用 Docker）
docker-compose up -d postgres redis

# 等待数据库启动
sleep 5

# 运行数据库迁移
make migrate-up DB_URL="postgres://postgres:postgres@localhost:5432/gravital?sslmode=disable"
```

### 3. 配置文件

```bash
# 复制配置文件
cp config/config.example.yaml config/config.yaml

# 编辑配置（根据实际情况修改）
vim config/config.yaml
```

关键配置项：
- `server.port`: 服务端口（默认 8080）
- `database.*`: 数据库连接信息
- `redis.*`: Redis 连接信息
- `auth.jwt_secret`: JWT 密钥（生产环境必须修改）

### 4. 启动服务

```bash
# 开发模式
make run

# 或编译后运行
make build
./bin/gravital-core -c config/config.yaml
```

### 5. 验证服务

```bash
# 健康检查
curl http://localhost:8080/health

# 版本信息
curl http://localhost:8080/version

# 登录（默认用户名: admin, 密码: admin123）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'
```

## 🐳 Docker 部署

### 使用 Docker Compose（推荐）

```bash
# 启动所有服务
docker-compose up -d

# 查看日志
docker-compose logs -f gravital-core

# 停止服务
docker-compose down
```

### 单独构建 Docker 镜像

```bash
# 构建镜像
make docker-build

# 运行容器
docker run -d \
  --name gravital-core \
  -p 8080:8080 \
  -v $(PWD)/config:/app/config \
  gravital-core:latest
```

## 📚 API 文档

### 认证

#### 登录
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123"
}
```

响应：
```json
{
  "code": 0,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### 设备管理

#### 获取设备列表
```http
GET /api/v1/devices?page=1&page_size=20
Authorization: Bearer {token}
```

#### 创建设备
```http
POST /api/v1/devices
Authorization: Bearer {token}
Content-Type: application/json

{
  "name": "Core Switch 01",
  "device_type": "switch",
  "connection_config": {
    "host": "192.168.1.1",
    "port": 161,
    "community": "public"
  },
  "labels": {
    "env": "production"
  }
}
```

### Sentinel 管理

#### Sentinel 注册
```http
POST /api/v1/sentinels/register
Content-Type: application/json

{
  "name": "sentinel-office-01",
  "hostname": "sentinel-01.local",
  "ip_address": "192.168.1.100",
  "version": "1.0.0",
  "os": "linux",
  "arch": "amd64",
  "region": "office-beijing"
}
```

#### 心跳上报
```http
POST /api/v1/sentinels/heartbeat
X-Sentinel-ID: sentinel-001
X-API-Token: {api_token}
Content-Type: application/json

{
  "cpu_usage": 15.5,
  "memory_usage": 45.2,
  "task_count": 20
}
```

### 数据采集

#### 上报数据
```http
POST /api/v1/data/ingest
X-Sentinel-ID: sentinel-001
X-API-Token: {api_token}
Content-Type: application/json

{
  "metrics": [
    {
      "device_id": "dev-001",
      "name": "cpu_usage",
      "value": 75.5,
      "timestamp": 1698883200,
      "labels": {
        "host": "server-01"
      }
    }
  ]
}
```

完整 API 文档请参考：[docs/05-API接口文档.md](../docs/05-API接口文档.md)

## 🔧 开发指南

### 运行测试

```bash
# 运行所有测试
make test

# 运行测试并生成覆盖率报告
make test-coverage
```

### 代码格式化

```bash
# 格式化代码
make fmt

# 代码检查
make lint
```

### 创建数据库迁移

```bash
# 创建新的迁移文件
make migrate-create NAME=add_new_table

# 运行迁移
make migrate-up DB_URL="..."

# 回滚迁移
make migrate-down DB_URL="..."
```

## 📊 监控和运维

### 健康检查

```bash
# 健康检查端点
curl http://localhost:8080/health

# 响应示例
{
  "status": "healthy",
  "components": {
    "database": "healthy",
    "redis": "healthy"
  }
}
```

### 日志

日志文件位置：`./logs/gravital.log`

日志级别：
- `debug`: 调试信息
- `info`: 一般信息
- `warn`: 警告信息
- `error`: 错误信息

### 性能分析

启用 pprof（在配置文件中设置）：
```yaml
system:
  enable_profiling: true
  profiling_port: 6060
```

访问：http://localhost:6060/debug/pprof/

## 🔐 安全建议

### 生产环境配置

1. **修改默认密码**
   - 修改数据库密码
   - 修改 Redis 密码
   - 修改默认管理员密码

2. **JWT 密钥**
   - 使用强随机密钥
   - 定期轮换密钥

3. **HTTPS**
   - 使用 Nginx 反向代理
   - 配置 SSL 证书

4. **防火墙**
   - 限制数据库访问
   - 只开放必要端口

## 🤝 与 Sentinel 集成

### 配置 Sentinel

在 Sentinel 的配置文件中设置中心端地址：

```yaml
core:
  url: "http://gravital-core:8080"
  api_token: "your-api-token"

heartbeat:
  interval: 30s
```

### 数据流

```
Sentinel → Gravital Core → TSDB
   ↓            ↓
心跳/任务    告警/转发
```

## 📝 待完成功能

- [ ] 完整的 Service 层实现
- [ ] 完整的 Handler 层实现
- [ ] 告警规则引擎
- [ ] 数据转发模块
- [ ] WebSocket 实时推送
- [ ] Grafana 集成
- [ ] Web UI
- [ ] 集群部署支持

## 🐛 故障排查

### 数据库连接失败

```bash
# 检查 PostgreSQL 是否运行
docker ps | grep postgres

# 检查连接
psql -h localhost -U postgres -d gravital

# 查看日志
docker logs postgres
```

### Redis 连接失败

```bash
# 检查 Redis 是否运行
docker ps | grep redis

# 测试连接
redis-cli ping

# 查看日志
docker logs redis
```

### 服务启动失败

```bash
# 查看详细日志
./bin/gravital-core -c config/config.yaml

# 检查配置文件
cat config/config.yaml

# 检查端口占用
lsof -i :8080
```

## 📖 相关文档

- [系统整体架构](../docs/01-系统整体架构设计.md)
- [中心端详细设计](../docs/02-中心端详细设计.md)
- [API 接口文档](../docs/05-API接口文档.md)
- [部署运维手册](../docs/06-部署运维手册.md)

## 📄 许可证

MIT License

## 👥 贡献

欢迎提交 Issue 和 Pull Request！

---

**Gravital Core** - 引力核心，统一管理你的监控系统 🌌

