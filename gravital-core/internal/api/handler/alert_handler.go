package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/service"
)

// AlertHandler 告警处理器
type AlertHandler struct {
	alertService service.AlertService
	db           *gorm.DB
}

// NewAlertHandler 创建告警处理器
func NewAlertHandler(alertService service.AlertService, db *gorm.DB) *AlertHandler {
	return &AlertHandler{
		alertService: alertService,
		db:           db,
	}
}

// ListRules 获取告警规则列表
func (h *AlertHandler) ListRules(c *gin.Context) {
	var req service.ListAlertRuleRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	rules, total, err := h.alertService.ListRules(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取规则列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"items":     rules,
		},
	})
}

// GetRule 获取告警规则详情
func (h *AlertHandler) GetRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的规则 ID",
		})
		return
	}

	rule, err := h.alertService.GetRule(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    50001,
			"message": "规则不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rule,
	})
}

// CreateRule 创建告警规则
func (h *AlertHandler) CreateRule(c *gin.Context) {
	var req service.CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	rule, err := h.alertService.CreateRule(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "创建规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"id": rule.ID,
		},
	})
}

// UpdateRule 更新告警规则
func (h *AlertHandler) UpdateRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的规则 ID",
		})
		return
	}

	var req service.UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.alertService.UpdateRule(c.Request.Context(), uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "更新规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// DeleteRule 删除告警规则
func (h *AlertHandler) DeleteRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的规则 ID",
		})
		return
	}

	if err := h.alertService.DeleteRule(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "删除规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ToggleRule 启用/禁用告警规则
func (h *AlertHandler) ToggleRule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的规则 ID",
		})
		return
	}

	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.alertService.ToggleRule(c.Request.Context(), uint(id), req.Enabled); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "操作失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ListEvents 获取告警事件列表
func (h *AlertHandler) ListEvents(c *gin.Context) {
	var req service.ListAlertEventRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	events, total, err := h.alertService.ListEvents(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取事件列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"items":     events,
		},
	})
}

// GetEvent 获取告警事件详情
func (h *AlertHandler) GetEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的事件 ID",
		})
		return
	}

	event, err := h.alertService.GetEvent(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    50001,
			"message": "事件不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": event,
	})
}

// AcknowledgeEvent 确认告警事件
func (h *AlertHandler) AcknowledgeEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的事件 ID",
		})
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	// 使用 ShouldBindJSON，但忽略 EOF 错误（空请求体）
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 获取当前用户 ID
	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(0) // 如果没有用户信息，使用 0
	}

	if err := h.alertService.AcknowledgeEvent(c.Request.Context(), uint(id), userID.(uint), req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "确认失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ResolveEvent 解决告警事件
func (h *AlertHandler) ResolveEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的事件 ID",
		})
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	// 使用 ShouldBindJSON，但忽略 EOF 错误（空请求体）
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.alertService.ResolveEvent(c.Request.Context(), uint(id), req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "解决失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// SilenceEvent 静默告警事件
func (h *AlertHandler) SilenceEvent(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的事件 ID",
		})
		return
	}

	var req struct {
		Duration string `json:"duration" binding:"required"`
		Comment  string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的时间格式",
		})
		return
	}

	if err := h.alertService.SilenceEvent(c.Request.Context(), uint(id), duration, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "静默失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// GetStats 获取告警统计
func (h *AlertHandler) GetStats(c *gin.Context) {
	stats, err := h.alertService.GetStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取统计失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": stats,
	})
}

// GetAggregations 获取告警聚合信息
func (h *AlertHandler) GetAggregations(c *gin.Context) {
	aggregations, err := service.GetAlertAggregations(c.Request.Context(), h.db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取聚合信息失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": aggregations,
	})
}

// BatchAcknowledge 批量确认告警
func (h *AlertHandler) BatchAcknowledge(c *gin.Context) {
	var req struct {
		IDs     []uint `json:"ids" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		userID = uint(0)
	}

	if err := service.BatchAcknowledgeEvents(c.Request.Context(), h.db, req.IDs, userID.(uint), req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "批量确认失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// BatchResolve 批量解决告警
func (h *AlertHandler) BatchResolve(c *gin.Context) {
	var req struct {
		IDs     []uint `json:"ids" binding:"required"`
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.BatchResolveEvents(c.Request.Context(), h.db, req.IDs, req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "批量解决失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// ResolveByRule 解决某个规则的所有告警
func (h *AlertHandler) ResolveByRule(c *gin.Context) {
	ruleID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的规则 ID",
		})
		return
	}

	var req struct {
		Comment string `json:"comment"`
	}
	if err := c.ShouldBindJSON(&req); err != nil && err.Error() != "EOF" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := service.ResolveEventsByRule(c.Request.Context(), h.db, uint(ruleID), req.Comment); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "解决失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

