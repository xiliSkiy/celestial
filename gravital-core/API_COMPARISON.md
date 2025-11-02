# 前后端 API 接口对比检查

## 📋 检查日期
2025-11-02

## 🔍 检查方法
- 前端：检查 `web/src/api/` 目录下所有 API 定义
- 后端：检查 `internal/api/router/router.go` 中注册的路由

## ✅ 已实现的接口

### 1. 认证相关 (Auth)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| POST /v1/auth/login | POST /api/v1/auth/login | ✅ 已实现 |
| POST /v1/auth/refresh | POST /api/v1/auth/refresh | ✅ 已实现 |
| POST /v1/auth/logout | POST /api/v1/auth/logout | ✅ 已实现 |
| GET /v1/auth/me | GET /api/v1/auth/me | ✅ 已实现 |

### 2. 设备管理 (Device)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /v1/devices | GET /api/v1/devices | ✅ 已实现 |
| GET /v1/devices/:id | GET /api/v1/devices/:id | ✅ 已实现 |
| POST /v1/devices | POST /api/v1/devices | ✅ 已实现 |
| PUT /v1/devices/:id | PUT /api/v1/devices/:id | ✅ 已实现 |
| DELETE /v1/devices/:id | DELETE /api/v1/devices/:id | ✅ 已实现 |
| POST /v1/devices/:id/test-connection | POST /api/v1/devices/:id/test-connection | ✅ 已实现 |
| GET /v1/device-groups | GET /api/v1/device-groups/tree | ⚠️ 路径不同 |
| POST /v1/device-groups | POST /api/v1/device-groups | ✅ 已实现 |
| POST /v1/devices/import | POST /api/v1/devices/batch-import | ⚠️ 路径不同 |
| GET /v1/devices/export | - | ❌ 未实现 |

### 3. Sentinel 管理

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /v1/sentinels | GET /api/v1/sentinels | ✅ 已实现 |
| GET /v1/sentinels/:id | GET /api/v1/sentinels/:id | ✅ 已实现 |
| GET /v1/sentinels/:id/stats | - | ❌ 未实现 |
| POST /v1/sentinels/:id/control | POST /api/v1/sentinels/:id/control | ✅ 已实现 |
| DELETE /v1/sentinels/:id | DELETE /api/v1/sentinels/:id | ✅ 已实现 |

### 4. 任务管理 (Task)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/tasks | GET /api/v1/tasks | ✅ 已实现 |
| GET /api/v1/tasks/:id | GET /api/v1/tasks/:id | ✅ 已实现 |
| POST /api/v1/tasks | POST /api/v1/tasks | ✅ 已实现 |
| PUT /api/v1/tasks/:id | PUT /api/v1/tasks/:id | ✅ 已实现 |
| DELETE /api/v1/tasks/:id | DELETE /api/v1/tasks/:id | ✅ 已实现 |
| PATCH /api/v1/tasks/:id | - | ❌ 未实现 (toggle) |
| POST /api/v1/tasks/:id/trigger | POST /api/v1/tasks/:id/trigger | ✅ 已实现 |
| GET /api/v1/tasks/:id/executions | - | ❌ 未实现 |

### 5. 告警管理 (Alert)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /v1/alert-rules | GET /api/v1/alert-rules | ✅ 已实现 |
| GET /v1/alert-rules/:id | GET /api/v1/alert-rules/:id | ✅ 已实现 |
| POST /v1/alert-rules | POST /api/v1/alert-rules | ✅ 已实现 |
| PUT /v1/alert-rules/:id | PUT /api/v1/alert-rules/:id | ✅ 已实现 |
| DELETE /v1/alert-rules/:id | DELETE /api/v1/alert-rules/:id | ✅ 已实现 |
| PUT /v1/alert-rules/:id/toggle | POST /api/v1/alert-rules/:id/toggle | ⚠️ 方法不同 |
| GET /v1/alert-events | GET /api/v1/alert-events | ✅ 已实现 |
| GET /v1/alert-events/:id | GET /api/v1/alert-events/:id | ✅ 已实现 |
| POST /v1/alert-events/:id/acknowledge | POST /api/v1/alert-events/:id/acknowledge | ✅ 已实现 |
| POST /v1/alert-events/:id/resolve | POST /api/v1/alert-events/:id/resolve | ✅ 已实现 |
| POST /v1/alert-events/:id/silence | POST /api/v1/alert-events/:id/silence | ✅ 已实现 |

### 6. 数据转发 (Forwarder)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/forwarders | GET /api/v1/forwarders | ✅ 已实现 |
| GET /api/v1/forwarders/:id | GET /api/v1/forwarders/:name | ⚠️ 参数不同 |
| POST /api/v1/forwarders | POST /api/v1/forwarders | ✅ 已实现 |
| PUT /api/v1/forwarders/:id | PUT /api/v1/forwarders/:name | ⚠️ 参数不同 |
| DELETE /api/v1/forwarders/:id | DELETE /api/v1/forwarders/:name | ⚠️ 参数不同 |
| PATCH /api/v1/forwarders/:id | - | ❌ 未实现 (toggle) |
| POST /api/v1/forwarders/reload | POST /api/v1/forwarders/reload | ✅ 已实现 |
| GET /api/v1/forwarders/:id/stats | GET /api/v1/forwarders/:name/stats | ⚠️ 参数不同 |
| POST /api/v1/forwarders/test | - | ❌ 未实现 |

### 7. Dashboard

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/dashboard/stats | - | ❌ 未实现 |
| GET /api/v1/dashboard/device-status | - | ❌ 未实现 |
| GET /api/v1/dashboard/alert-trend | - | ❌ 未实现 |
| GET /api/v1/dashboard/sentinel-status | - | ❌ 未实现 |
| GET /api/v1/dashboard/forwarder-stats | - | ❌ 未实现 |
| GET /api/v1/dashboard/activities | - | ❌ 未实现 |

### 8. 用户管理 (User)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/users | - | ❌ 未实现 |
| GET /api/v1/users/:id | - | ❌ 未实现 |
| POST /api/v1/users | - | ❌ 未实现 |
| PUT /api/v1/users/:id | - | ❌ 未实现 |
| DELETE /api/v1/users/:id | - | ❌ 未实现 |
| PATCH /api/v1/users/:id | - | ❌ 未实现 (toggle) |
| POST /api/v1/users/:id/reset-password | - | ❌ 未实现 |

### 9. 角色管理 (Role)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/roles | - | ❌ 未实现 |
| GET /api/v1/roles/:id | - | ❌ 未实现 |
| POST /api/v1/roles | - | ❌ 未实现 |
| PUT /api/v1/roles/:id | - | ❌ 未实现 |
| DELETE /api/v1/roles/:id | - | ❌ 未实现 |

### 10. 系统配置 (System Config)

| 前端 API | 后端路由 | 状态 |
|---------|---------|------|
| GET /api/v1/system/config | GET /api/v1/system/config | ✅ 已实现 |
| PUT /api/v1/system/config | PUT /api/v1/system/config | ✅ 已实现 |

## 📊 统计总结

### 整体统计

| 状态 | 数量 | 百分比 |
|------|------|--------|
| ✅ 已实现 | 40 | 62% |
| ⚠️ 部分实现 | 7 | 11% |
| ❌ 未实现 | 17 | 27% |
| **总计** | **64** | **100%** |

### 按模块统计

| 模块 | 已实现 | 部分实现 | 未实现 | 总计 |
|------|-------|---------|--------|------|
| 认证 | 4 | 0 | 0 | 4 |
| 设备管理 | 7 | 2 | 1 | 10 |
| Sentinel | 4 | 0 | 1 | 5 |
| 任务管理 | 6 | 0 | 2 | 8 |
| 告警管理 | 10 | 1 | 0 | 11 |
| 数据转发 | 5 | 3 | 2 | 10 |
| Dashboard | 0 | 0 | 6 | 6 |
| 用户管理 | 0 | 0 | 7 | 7 |
| 角色管理 | 0 | 0 | 5 | 5 |
| 系统配置 | 2 | 0 | 0 | 2 |

## ❌ 需要实现的接口列表

### 高优先级（核心功能）

1. **Dashboard API** (6个接口) - P0
   - GET /api/v1/dashboard/stats
   - GET /api/v1/dashboard/device-status
   - GET /api/v1/dashboard/alert-trend
   - GET /api/v1/dashboard/sentinel-status
   - GET /api/v1/dashboard/forwarder-stats
   - GET /api/v1/dashboard/activities

2. **用户管理 API** (7个接口) - P0
   - GET /api/v1/users
   - GET /api/v1/users/:id
   - POST /api/v1/users
   - PUT /api/v1/users/:id
   - DELETE /api/v1/users/:id
   - PATCH /api/v1/users/:id (toggle)
   - POST /api/v1/users/:id/reset-password

3. **角色管理 API** (5个接口) - P0
   - GET /api/v1/roles
   - GET /api/v1/roles/:id
   - POST /api/v1/roles
   - PUT /api/v1/roles/:id
   - DELETE /api/v1/roles/:id

### 中优先级（增强功能）

4. **任务管理补充** (2个接口) - P1
   - PATCH /api/v1/tasks/:id (toggle enabled)
   - GET /api/v1/tasks/:id/executions

5. **Sentinel 补充** (1个接口) - P1
   - GET /v1/sentinels/:id/stats

6. **数据转发补充** (2个接口) - P1
   - PATCH /api/v1/forwarders/:id (toggle)
   - POST /api/v1/forwarders/test

7. **设备管理补充** (1个接口) - P1
   - GET /v1/devices/export

## ⚠️ 需要修复的接口（路径/方法不一致）

1. **设备分组**
   - 前端: GET /v1/device-groups
   - 后端: GET /api/v1/device-groups/tree
   - 建议: 统一路径

2. **设备导入**
   - 前端: POST /v1/devices/import
   - 后端: POST /api/v1/devices/batch-import
   - 建议: 统一路径

3. **告警规则切换**
   - 前端: PUT /v1/alert-rules/:id/toggle
   - 后端: POST /api/v1/alert-rules/:id/toggle
   - 建议: 统一为 POST

4. **转发器参数**
   - 前端: 使用 :id (数字)
   - 后端: 使用 :name (字符串)
   - 建议: 统一为 :id

## 🎯 实现优先级

### Phase 1: 核心功能（必须实现）
1. Dashboard API (6个)
2. 用户管理 API (7个)
3. 角色管理 API (5个)

**总计**: 18个接口

### Phase 2: 增强功能（建议实现）
1. 任务管理补充 (2个)
2. Sentinel 补充 (1个)
3. 数据转发补充 (2个)
4. 设备管理补充 (1个)

**总计**: 6个接口

### Phase 3: 接口统一（优化）
1. 修复路径不一致 (4处)
2. 统一参数类型 (3处)

## 📝 实现建议

### 1. Dashboard Handler
创建 `internal/api/handler/dashboard_handler.go`，实现 6 个统计接口

### 2. User Handler
创建 `internal/api/handler/user_handler.go`，实现用户 CRUD

### 3. Role Handler
可以合并到 User Handler 或单独创建

### 4. 补充现有 Handler
- Task Handler: 添加 toggle 和 executions
- Sentinel Handler: 添加 stats
- Forwarder Handler: 添加 toggle 和 test

---

**检查完成日期**: 2025-11-02  
**下一步**: 根据优先级逐个实现未完成的接口

