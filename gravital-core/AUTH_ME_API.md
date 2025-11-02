# /api/v1/auth/me 接口实现说明

## 问题描述

前端请求 `http://localhost:5173/api/v1/auth/me` 时返回 404 错误，因为该接口尚未实现。

## 问题原因

前端在用户登录后需要调用 `/api/v1/auth/me` 接口来获取当前用户的详细信息，但后端缺少这个接口的实现。

## 解决方案

### 1. 添加 Service 接口方法

在 `internal/service/auth_service.go` 中添加 `GetUserInfo` 方法：

```go
// AuthService 认证服务接口
type AuthService interface {
    Login(ctx context.Context, username, password string) (*LoginResponse, error)
    RefreshToken(ctx context.Context, refreshToken string) (*LoginResponse, error)
    Logout(ctx context.Context, userID uint) error
    GetUserInfo(ctx context.Context, userID uint) (*model.User, error) // 新增
}
```

### 2. 实现 Service 方法

```go
func (s *authService) GetUserInfo(ctx context.Context, userID uint) (*model.User, error) {
    user, err := s.userRepo.GetByID(ctx, userID)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            return nil, fmt.Errorf("用户不存在")
        }
        return nil, fmt.Errorf("failed to get user: %w", err)
    }

    // 检查用户是否启用
    if !user.Enabled {
        return nil, fmt.Errorf("用户已被禁用")
    }

    return user, nil
}
```

### 3. 添加 Handler 方法

在 `internal/api/handler/auth_handler.go` 中添加：

```go
// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
    userID, exists := c.Get("user_id")
    if !exists {
        c.JSON(http.StatusUnauthorized, gin.H{
            "code":    20001,
            "message": "未认证",
        })
        return
    }

    user, err := h.authService.GetUserInfo(c.Request.Context(), userID.(uint))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "code":    10001,
            "message": "获取用户信息失败: " + err.Error(),
        })
        return
    }

    c.JSON(http.StatusOK, gin.H{
        "code": 0,
        "data": user,
    })
}
```

### 4. 注册路由

在 `internal/api/router/router.go` 中添加路由：

```go
auth := v1.Group("/auth")
{
    auth.POST("/login", authHandler.Login)
    auth.POST("/refresh", authHandler.RefreshToken)
    auth.POST("/logout", middleware.Auth(jwtManager), authHandler.Logout)
    auth.GET("/me", middleware.Auth(jwtManager), authHandler.GetCurrentUser) // 新增
}
```

## API 说明

### 请求

```http
GET /api/v1/auth/me
Authorization: Bearer <token>
```

### 成功响应

```json
{
  "code": 0,
  "data": {
    "id": 1,
    "username": "admin",
    "email": "admin@gravital-core.local",
    "role_id": 1,
    "role": {
      "id": 1,
      "name": "admin",
      "permissions": ["*"],
      "description": "管理员，拥有所有权限",
      "created_at": "2025-11-02T04:44:02.802343Z"
    },
    "enabled": true,
    "last_login": "2025-11-02T13:47:21.726248+08:00",
    "created_at": "2025-11-02T04:44:02.802343Z",
    "updated_at": "2025-11-02T05:47:21.742468Z"
  }
}
```

### 错误响应

#### 未认证
```json
{
  "code": 20001,
  "message": "未认证"
}
```

#### 用户不存在
```json
{
  "code": 10001,
  "message": "获取用户信息失败: 用户不存在"
}
```

#### 用户已禁用
```json
{
  "code": 10001,
  "message": "获取用户信息失败: 用户已被禁用"
}
```

## 测试

### 使用测试脚本

```bash
cd gravital-core
./scripts/test-me-api.sh
```

### 手动测试

```bash
# 1. 登录获取 token
TOKEN=$(curl -s -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 2. 获取当前用户信息
curl -X GET "http://localhost:8080/api/v1/auth/me" \
  -H "Authorization: Bearer $TOKEN" \
  | jq '.'
```

## 相关文件

- `internal/service/auth_service.go` - 认证服务
- `internal/api/handler/auth_handler.go` - 认证处理器
- `internal/api/router/router.go` - 路由配置
- `scripts/test-me-api.sh` - 测试脚本

## 前端使用

前端在 `src/stores/user.ts` 中调用此接口：

```typescript
// 获取用户信息
const getUserInfo = async () => {
  try {
    const res: any = await authApi.getUserInfo()
    userInfo.value = res
    permissions.value = res?.role?.permissions || []
  } catch (error) {
    console.error('Get user info error:', error)
    logout()
  }
}
```

## 注意事项

1. **认证要求**: 此接口需要 JWT token 认证
2. **权限检查**: 自动检查用户是否被禁用
3. **关联查询**: 自动加载用户的角色和权限信息
4. **错误处理**: 统一的错误响应格式

## 状态码说明

| Code | 说明 |
|------|------|
| 0 | 成功 |
| 10001 | 服务器内部错误 |
| 20001 | 未认证或认证失败 |

---

**实现日期**: 2025-11-02  
**状态**: ✅ 已实现并测试

