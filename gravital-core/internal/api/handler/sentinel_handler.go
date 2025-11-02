package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/celestial/gravital-core/internal/service"
)

// SentinelHandler Sentinel 处理器
type SentinelHandler struct {
	sentinelService service.SentinelService
}

// NewSentinelHandler 创建 Sentinel 处理器
func NewSentinelHandler(sentinelService service.SentinelService) *SentinelHandler {
	return &SentinelHandler{
		sentinelService: sentinelService,
	}
}

// Register Sentinel 注册
func (h *SentinelHandler) Register(c *gin.Context) {
	var req service.RegisterSentinelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	resp, err := h.sentinelService.Register(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "注册失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": resp,
	})
}

// Heartbeat 心跳上报
func (h *SentinelHandler) Heartbeat(c *gin.Context) {
	sentinelID := c.GetHeader("X-Sentinel-ID")
	if sentinelID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "缺少 Sentinel ID",
		})
		return
	}

	var req service.HeartbeatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.sentinelService.Heartbeat(c.Request.Context(), sentinelID, &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "心跳处理失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"status":         "ok",
			"config_version": 1,
		},
	})
}

// List 获取 Sentinel 列表
func (h *SentinelHandler) List(c *gin.Context) {
	var req service.ListSentinelRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	sentinels, total, err := h.sentinelService.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"items":     sentinels,
		},
	})
}

// Get 获取 Sentinel 详情
func (h *SentinelHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的 Sentinel ID",
		})
		return
	}

	sentinel, err := h.sentinelService.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    50001,
			"message": "Sentinel 不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": sentinel,
	})
}

// Delete 删除 Sentinel
func (h *SentinelHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的 Sentinel ID",
		})
		return
	}

	if err := h.sentinelService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// Control 远程控制 Sentinel
func (h *SentinelHandler) Control(c *gin.Context) {
	sentinelID := c.Param("id")

	var req struct {
		Action string                 `json:"action" binding:"required"`
		Params map[string]interface{} `json:"params"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.sentinelService.Control(c.Request.Context(), sentinelID, req.Action); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "控制失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

