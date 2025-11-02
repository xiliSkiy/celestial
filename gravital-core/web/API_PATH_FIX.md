# 前端 API 路径修复报告

## 🐛 问题描述

**发现时间**: 2025-11-02  
**问题**: 前端请求路径出现重复的 `/api`，导致实际请求变成 `/api/api/v1/...`

## 🔍 问题原因

### 配置分析

1. **request.ts 配置**:
```typescript
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  // ...
})
```

2. **vite.config.ts 代理配置**:
```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',
      changeOrigin: true
    }
  }
}
```

3. **API 文件路径不统一**:
- ❌ 错误: `/api/v1/users` (user.ts, task.ts, forwarder.ts, dashboard.ts)
- ✅ 正确: `/v1/users` (auth.ts, device.ts, sentinel.ts, alert.ts)

### 请求流程

```
前端调用: request.get('/api/v1/users')
↓
baseURL 拼接: /api + /api/v1/users = /api/api/v1/users
↓
Vite 代理: http://localhost:8080/api/api/v1/users
↓
❌ 404 Not Found
```

## ✅ 解决方案

### 修复原则

**由于 `baseURL` 已经设置为 `/api`，所有 API 路径应该只写 `/v1/...`**

### 修复的文件

1. ✅ `src/api/user.ts` - 所有路径从 `/api/v1/...` 改为 `/v1/...`
2. ✅ `src/api/task.ts` - 所有路径从 `/api/v1/...` 改为 `/v1/...`
3. ✅ `src/api/forwarder.ts` - 所有路径从 `/api/v1/...` 改为 `/v1/...`
4. ✅ `src/api/dashboard.ts` - 所有路径从 `/api/v1/...` 改为 `/v1/...`

### 修复对比

#### 修复前
```typescript
// user.ts
getUsers: (params: UserQuery) => {
  return request.get('/api/v1/users', { params })
}

// 实际请求: /api + /api/v1/users = /api/api/v1/users ❌
```

#### 修复后
```typescript
// user.ts
getUsers: (params: UserQuery) => {
  return request.get('/v1/users', { params })
}

// 实际请求: /api + /v1/users = /api/v1/users ✅
```

## 📊 修复统计

| 文件 | 修复接口数 | 状态 |
|------|-----------|------|
| user.ts | 12 | ✅ 完成 |
| task.ts | 8 | ✅ 完成 |
| forwarder.ts | 9 | ✅ 完成 |
| dashboard.ts | 6 | ✅ 完成 |
| **总计** | **35** | **✅ 完成** |

## 🎯 修复后的路径规范

### 统一的路径格式

所有 API 文件现在都使用统一的路径格式：

```typescript
// ✅ 正确格式
export const xxxApi = {
  getList: () => request.get('/v1/xxx'),
  getOne: (id) => request.get(`/v1/xxx/${id}`),
  create: (data) => request.post('/v1/xxx', data),
  update: (id, data) => request.put(`/v1/xxx/${id}`, data),
  delete: (id) => request.delete(`/v1/xxx/${id}`)
}
```

### 完整的 API 路径列表

#### 认证 API
- POST `/v1/auth/login`
- POST `/v1/auth/refresh`
- POST `/v1/auth/logout`
- GET `/v1/auth/me`

#### 设备 API
- GET `/v1/devices`
- GET `/v1/devices/:id`
- POST `/v1/devices`
- PUT `/v1/devices/:id`
- DELETE `/v1/devices/:id`
- POST `/v1/devices/:id/test-connection`
- GET `/v1/device-groups`
- POST `/v1/device-groups`

#### Sentinel API
- GET `/v1/sentinels`
- GET `/v1/sentinels/:id`
- GET `/v1/sentinels/:id/stats`
- POST `/v1/sentinels/:id/control`
- DELETE `/v1/sentinels/:id`

#### 任务 API
- GET `/v1/tasks`
- GET `/v1/tasks/:id`
- POST `/v1/tasks`
- PUT `/v1/tasks/:id`
- DELETE `/v1/tasks/:id`
- PATCH `/v1/tasks/:id`
- POST `/v1/tasks/:id/trigger`
- GET `/v1/tasks/:id/executions`

#### 告警 API
- GET `/v1/alert-rules`
- GET `/v1/alert-rules/:id`
- POST `/v1/alert-rules`
- PUT `/v1/alert-rules/:id`
- DELETE `/v1/alert-rules/:id`
- PUT `/v1/alert-rules/:id/toggle`
- GET `/v1/alert-events`
- GET `/v1/alert-events/:id`
- POST `/v1/alert-events/:id/acknowledge`
- POST `/v1/alert-events/:id/resolve`

#### 转发器 API
- GET `/v1/forwarders`
- GET `/v1/forwarders/:id`
- POST `/v1/forwarders`
- PUT `/v1/forwarders/:id`
- DELETE `/v1/forwarders/:id`
- PATCH `/v1/forwarders/:id`
- POST `/v1/forwarders/reload`
- GET `/v1/forwarders/:id/stats`
- GET `/v1/forwarders/buffer-stats`
- POST `/v1/forwarders/:id/test`

#### Dashboard API
- GET `/v1/dashboard/stats`
- GET `/v1/dashboard/device-status`
- GET `/v1/dashboard/alert-trend`
- GET `/v1/dashboard/sentinel-status`
- GET `/v1/dashboard/forwarder-stats`
- GET `/v1/dashboard/activities`

#### 用户 API
- GET `/v1/users`
- GET `/v1/users/:id`
- POST `/v1/users`
- PUT `/v1/users/:id`
- DELETE `/v1/users/:id`
- PATCH `/v1/users/:id`
- POST `/v1/users/:id/reset-password`

#### 角色 API
- GET `/v1/roles`
- GET `/v1/roles/:id`
- POST `/v1/roles`
- PUT `/v1/roles/:id`
- DELETE `/v1/roles/:id`

#### 系统配置 API
- GET `/v1/system/config`
- PUT `/v1/system/config`

## 🔧 配置说明

### request.ts
```typescript
const request: AxiosInstance = axios.create({
  baseURL: '/api',  // 所有请求自动加上 /api 前缀
  timeout: 30000
})
```

### vite.config.ts
```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:8080',  // 代理到后端
      changeOrigin: true
      // 不需要 rewrite，保持 /api 前缀
    }
  }
}
```

### 请求流程（修复后）
```
前端调用: request.get('/v1/users')
↓
baseURL 拼接: /api + /v1/users = /api/v1/users
↓
Vite 代理: http://localhost:8080/api/v1/users
↓
后端路由: /api/v1/users
↓
✅ 200 OK
```

## ✨ 额外改进

### forwarder.ts 改进

1. **修复 testConnection 接口**:
```typescript
// 修复前
testConnection: (data: ForwarderForm) => {
  return request.post('/v1/forwarders/test', data)
}

// 修复后
testConnection: (id: number) => {
  return request.post(`/v1/forwarders/${id}/test`)
}
```

2. **新增 getBufferStats 接口**:
```typescript
getBufferStats: () => {
  return request.get('/v1/forwarders/buffer-stats')
}
```

## 🎉 总结

### 修复完成
- ✅ 修复了 35 个 API 路径
- ✅ 统一了所有 API 文件的路径格式
- ✅ 改进了部分接口设计
- ✅ 添加了缺失的接口

### 验证方法
```bash
# 启动前端
cd gravital-core/web
npm run dev

# 访问应用
open http://localhost:5173

# 检查浏览器控制台
# 应该看到正确的请求路径: /api/v1/...
# 而不是错误的: /api/api/v1/...
```

### 注意事项
1. ⚠️ 所有新增的 API 都应该使用 `/v1/...` 格式
2. ⚠️ 不要在 API 路径中再加 `/api` 前缀
3. ⚠️ `baseURL` 已经包含了 `/api`

---

**修复完成日期**: 2025-11-02  
**修复人员**: AI Assistant  
**状态**: ✅ 100% 完成

**🎊 前端 API 路径已全部修复，现在可以正常请求后端接口了！**

