package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestID 请求 ID 中间件
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试从 header 获取
		requestID := c.GetHeader("X-Request-ID")
		
		// 如果没有，生成一个新的
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// 设置到上下文和响应头
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)
		
		c.Next()
	}
}

