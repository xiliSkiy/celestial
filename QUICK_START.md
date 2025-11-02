# Gravital Core 快速启动指南

## 🚀 快速开始

### 前置要求

- Docker & Docker Compose
- Go 1.21+
- Node.js 18+
- npm 9+

### 1️⃣ 启动后端服务

```bash
# 进入后端目录
cd gravital-core

# 启动数据库和 Redis
docker-compose up -d postgres redis

# 等待服务启动
sleep 5

# 运行数据库迁移
make migrate-up

# 创建管理员用户
docker-compose exec -T postgres psql -U postgres -d gravital <<'SQL'
INSERT INTO roles (name, permissions, description) VALUES
    ('admin', '["*"]', '管理员，拥有所有权限')
ON CONFLICT (name) DO UPDATE SET 
    permissions = EXCLUDED.permissions;

INSERT INTO users (username, email, password_hash, role_id, enabled) 
SELECT 
    'admin',
    'admin@gravital-core.local',
    '$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG',
    (SELECT id FROM roles WHERE name = 'admin'),
    true
ON CONFLICT (username) DO UPDATE SET
    password_hash = EXCLUDED.password_hash;
SQL

# 编译并启动后端
make build
./bin/gravital-core -c config/config.yaml
```

### 2️⃣ 启动前端服务

```bash
# 新开一个终端窗口
cd gravital-core/web

# 安装依赖（首次运行）
npm install

# 创建环境变量文件
cp .env.example .env.local

# 启动开发服务器
npm run dev
```

### 3️⃣ 访问应用

- 🌐 **前端**: http://localhost:5173
- 🔧 **后端 API**: http://localhost:8080
- 📊 **Grafana**: http://localhost:3000 (如果启用)

### 4️⃣ 登录

- 👤 **用户名**: `admin`
- 🔑 **密码**: `admin123`

## 📋 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| 前端 (Vue) | 5173 | Web UI |
| 后端 (Go) | 8080 | RESTful API |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Grafana | 3000 | 可视化 |
| VictoriaMetrics | 8428 | 时序数据库 |

## 🔧 常用命令

### 后端

```bash
# 编译
make build

# 运行
make run

# 测试
make test

# 数据库迁移
make migrate-up    # 升级
make migrate-down  # 回滚

# 清理
make clean
```

### 前端

```bash
# 开发
npm run dev

# 构建
npm run build

# 预览
npm run preview

# 代码检查
npm run lint
```

## 🧪 测试

### 测试登录接口

```bash
cd gravital-core
./scripts/test-login.sh
```

### 测试前端登录

1. 访问 http://localhost:5173
2. 输入用户名: `admin`
3. 输入密码: `admin123`
4. 点击登录

## 🐛 常见问题

### 1. 端口冲突

**问题**: 端口已被占用

**解决**:
```bash
# 查看端口占用
lsof -i :5173  # 前端
lsof -i :8080  # 后端

# 杀死进程
kill -9 <PID>
```

### 2. 数据库连接失败

**问题**: `failed to connect database`

**解决**:
```bash
# 检查 PostgreSQL 是否运行
docker-compose ps postgres

# 重启 PostgreSQL
docker-compose restart postgres

# 查看日志
docker-compose logs postgres
```

### 3. Redis 认证失败

**问题**: `NOAUTH Authentication required`

**解决**:
```bash
# 检查 config/config.yaml 中的 Redis 密码
# 应该是: redis123

# 或者查看 docker-compose.yaml 中的密码
grep "requirepass" docker-compose.yaml
```

### 4. 前端登录失败

**问题**: 登录后没有跳转

**解决**:
```bash
# 检查后端是否运行
curl http://localhost:8080/api/v1/health

# 检查浏览器控制台错误
# F12 -> Console

# 清除浏览器缓存和 localStorage
# F12 -> Application -> Local Storage -> Clear
```

### 5. 密码错误

**问题**: `用户名或密码错误`

**解决**:
```bash
# 重新设置管理员密码
cd gravital-core
docker-compose exec -T postgres psql -U postgres -d gravital <<'SQL'
UPDATE users 
SET password_hash = '$2a$10$C/d6qPp3yGedbXT9kxnTieZYlwboRXIy.FcFjrie/yghKedWwR8yG'
WHERE username = 'admin';
SQL
```

## 📚 相关文档

### 设计文档
- [系统整体架构设计](docs/01-系统整体架构设计.md)
- [中心端详细设计](docs/02-中心端详细设计.md)
- [采集端详细设计](docs/03-采集端详细设计.md)
- [前端 UI 设计方案](docs/07-前端UI设计方案.md)

### 实现文档
- [后端实现指南](gravital-core/IMPLEMENTATION_GUIDE.md)
- [前端实现总结](gravital-core/web/FRONTEND_IMPLEMENTATION_SUMMARY.md)
- [API 接口文档](docs/05-API接口文档.md)

### 问题修复
- [登录问题修复](gravital-core/LOGIN_FIX.md)
- [密码更新说明](gravital-core/PASSWORD_UPDATE.md)
- [API 响应处理修复](gravital-core/web/API_RESPONSE_FIX.md)

## 🎯 下一步

1. **配置数据源**
   - 添加设备
   - 配置 Sentinel
   - 创建采集任务

2. **配置告警**
   - 创建告警规则
   - 配置通知渠道

3. **配置数据转发**
   - 添加 VictoriaMetrics
   - 添加 ClickHouse
   - 配置 Prometheus

4. **配置 Grafana**
   - 添加数据源
   - 导入仪表盘
   - 创建可视化

## 💡 提示

- 开发环境下，前端会自动代理 API 请求到后端
- 生产环境需要配置 Nginx 反向代理
- 建议使用 Docker Compose 部署完整环境
- 定期备份数据库

## 🆘 获取帮助

- 查看日志: `docker-compose logs -f`
- 查看后端日志: `tail -f /tmp/gravital-core.log`
- 查看前端控制台: 浏览器 F12 -> Console
- 运行测试脚本: `./scripts/test-login.sh`

---

**更新日期**: 2025-11-02  
**版本**: v1.0.0

**🎉 现在可以开始使用 Gravital Core 了！**

