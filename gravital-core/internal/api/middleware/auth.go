package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/celestial/gravital-core/internal/pkg/auth"
)

// Auth JWT 认证中间件
func Auth(jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    20001,
				"message": "未认证",
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    20002,
				"message": "Token 格式错误",
				"error":   "InvalidToken",
			})
			c.Abort()
			return
		}

		// 验证 token
		claims, err := jwtManager.VerifyToken(parts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    20003,
					"message": "Token 已过期",
					"error":   "TokenExpired",
				})
			} else {
				c.JSON(http.StatusUnauthorized, gin.H{
					"code":    20002,
					"message": "Token 无效",
					"error":   "InvalidToken",
				})
			}
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("permissions", claims.Permissions)

		c.Next()
	}
}

// RequirePermission 权限检查中间件
func RequirePermission(required string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取用户权限
		permsInterface, exists := c.Get("permissions")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    20004,
				"message": "权限不足",
				"error":   "Forbidden",
			})
			c.Abort()
			return
		}

		perms, ok := permsInterface.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    20004,
				"message": "权限不足",
				"error":   "Forbidden",
			})
			c.Abort()
			return
		}

		// 检查权限
		permission := auth.NewPermission(perms)
		if !permission.Has(required) {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    20004,
				"message": "权限不足: " + required,
				"error":   "Forbidden",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// SentinelAuth Sentinel 认证中间件
func SentinelAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取 Sentinel ID
		sentinelID := c.GetHeader("X-Sentinel-ID")
		if sentinelID == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    20001,
				"message": "未提供 Sentinel ID",
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}

		// 获取 API Token
		apiToken := c.GetHeader("X-API-Token")
		if apiToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    20001,
				"message": "未提供 API Token",
				"error":   "Unauthorized",
			})
			c.Abort()
			return
		}

		// TODO: 验证 API Token
		// 这里应该从数据库查询验证 token 是否有效
		// 为了简化，暂时只检查是否存在

		// 将 Sentinel ID 存入上下文
		c.Set("sentinel_id", sentinelID)

		c.Next()
	}
}

