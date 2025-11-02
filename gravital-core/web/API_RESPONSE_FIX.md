# API 响应处理修复说明

## 问题描述

前端登录时虽然后端返回成功（`code: 0`），但前端无法正确处理响应数据。

## 问题原因

### 1. 响应数据结构不匹配

**后端返回格式**:
```json
{
  "code": 0,
  "data": {
    "token": "...",
    "refresh_token": "...",
    "expires_in": 86400,
    "user": {
      "id": 1,
      "username": "admin",
      "role": {
        "permissions": ["*"]
      }
    }
  }
}
```

**前端期望格式**:
```javascript
// 直接访问 res.token，但实际应该是 res.data.token
token.value = res.token  // ❌ 错误
```

### 2. 权限获取路径错误

权限数据在 `user.role.permissions` 中，而不是 `user.permissions`。

## 解决方案

### 1. 修改响应拦截器 (`src/utils/request.ts`)

添加统一的响应处理逻辑：

```typescript
// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    
    // 如果返回的是 Blob 类型（文件下载），直接返回
    if (response.config.responseType === 'blob') {
      return response
    }
    
    // 后端统一返回格式: { code: 0, data: {...} }
    // 如果 code 为 0，返回 data 字段
    if (data.code === 0) {
      return data.data
    }
    
    // 如果 code 不为 0，抛出错误
    const error: any = new Error(data.message || '请求失败')
    error.code = data.code
    error.data = data
    return Promise.reject(error)
  },
  // ... 错误处理
)
```

**关键改动**:
- 检查 `data.code === 0` 判断请求是否成功
- 成功时返回 `data.data`，这样前端可以直接使用 `res.token`
- 失败时抛出包含错误信息的异常

### 2. 修改权限获取逻辑 (`src/stores/user.ts`)

```typescript
// 登录
const login = async (username: string, password: string) => {
  try {
    const res: any = await authApi.login({ username, password })
    token.value = res.token
    refreshToken.value = res.refresh_token
    userInfo.value = res.user
    // 权限在 user.role.permissions 中
    permissions.value = res.user?.role?.permissions || []
    
    // 保存到本地存储
    localStorage.setItem('token', res.token)
    localStorage.setItem('refreshToken', res.refresh_token)
    
    ElMessage.success('登录成功')
    router.push('/')
  } catch (error) {
    ElMessage.error('登录失败')
    throw error
  }
}
```

**关键改动**:
- `res.user.permissions` → `res.user?.role?.permissions`
- 使用可选链操作符 `?.` 避免空指针错误

## 数据流程

### 修复前
```
后端响应: { code: 0, data: { token: "..." } }
    ↓
Axios 拦截器: 直接返回整个响应
    ↓
前端代码: res.token (undefined) ❌
```

### 修复后
```
后端响应: { code: 0, data: { token: "..." } }
    ↓
Axios 拦截器: 检查 code === 0，返回 data 字段
    ↓
前端代码: res.token ("...") ✅
```

## 测试验证

### 1. 登录测试
```bash
# 访问前端
http://localhost:5173

# 使用以下凭据登录
用户名: admin
密码: admin123
```

### 2. 预期结果
- ✅ 显示 "登录成功" 提示
- ✅ 自动跳转到仪表盘页面
- ✅ Token 保存到 localStorage
- ✅ 用户信息和权限正确加载

### 3. 浏览器控制台检查
```javascript
// 检查 Token
localStorage.getItem('token')

// 检查用户 Store
// 在 Vue DevTools 中查看 user store
// permissions 应该为 ["*"]
```

## 后端 API 响应规范

所有 API 响应都应遵循以下格式：

### 成功响应
```json
{
  "code": 0,
  "data": {
    // 实际数据
  }
}
```

### 错误响应
```json
{
  "code": 20001,  // 错误码
  "message": "错误信息"
}
```

## 相关文件

- `web/src/utils/request.ts` - Axios 响应拦截器
- `web/src/stores/user.ts` - 用户状态管理
- `web/src/api/auth.ts` - 认证 API
- `web/src/types/api.ts` - API 类型定义

## 注意事项

1. **统一响应格式**: 确保所有后端 API 都返回 `{ code, data/message }` 格式
2. **错误处理**: 前端需要同时处理业务错误（code !== 0）和 HTTP 错误（4xx, 5xx）
3. **类型安全**: 建议为所有 API 响应定义 TypeScript 类型

---

**修复日期**: 2025-11-02  
**状态**: ✅ 已修复

