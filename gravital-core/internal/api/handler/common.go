package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// SuccessResponse 成功响应
func SuccessResponse(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": data,
	})
}

// ErrorResponse 错误响应
func ErrorResponse(c *gin.Context, httpStatus, code int, message string) {
	c.JSON(httpStatus, gin.H{
		"code":    code,
		"message": message,
	})
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"components": gin.H{
			"database": "healthy",
			"redis":    "healthy",
		},
	})
}

// Version 版本信息
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    "1.0.0",
		"build_time": "2025-11-02",
	})
}

// SystemInfo 系统信息
func SystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"version": "1.0.0",
			"uptime":  "1h30m",
		},
	})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
	})
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// IngestData 数据采集
func IngestData(c *gin.Context) {
	// TODO: 实现数据采集逻辑
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"received": 0,
			"accepted": 0,
			"rejected": 0,
		},
	})
}

