package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/celestial/gravital-core/internal/service"
)

// DeviceHandler 设备处理器
type DeviceHandler struct {
	deviceService service.DeviceService
}

// NewDeviceHandler 创建设备处理器
func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// List 获取设备列表
func (h *DeviceHandler) List(c *gin.Context) {
	var req service.ListDeviceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	devices, total, err := h.deviceService.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取设备列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"items":     devices,
		},
	})
}

// Get 获取设备详情
func (h *DeviceHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	device, err := h.deviceService.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    50001,
			"message": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": device,
	})
}

// Create 创建设备
func (h *DeviceHandler) Create(c *gin.Context) {
	var req service.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	device, err := h.deviceService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "创建设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"device_id": device.DeviceID,
		},
	})
}

// Update 更新设备
func (h *DeviceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	var req service.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.deviceService.Update(c.Request.Context(), uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "更新设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// Delete 删除设备
func (h *DeviceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	if err := h.deviceService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "删除设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// TestConnection 测试设备连接
func (h *DeviceHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	result, err := h.deviceService.TestConnection(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "测试连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// BatchImport 批量导入设备
func (h *DeviceHandler) BatchImport(c *gin.Context) {
	// TODO: 实现批量导入逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// GetGroupTree 获取设备分组树
func (h *DeviceHandler) GetGroupTree(c *gin.Context) {
	// TODO: 实现分组树查询
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": []interface{}{},
	})
}

// CreateGroup 创建设备分组
func (h *DeviceHandler) CreateGroup(c *gin.Context) {
	// TODO: 实现创建分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// UpdateGroup 更新设备分组
func (h *DeviceHandler) UpdateGroup(c *gin.Context) {
	// TODO: 实现更新分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// DeleteGroup 删除设备分组
func (h *DeviceHandler) DeleteGroup(c *gin.Context) {
	// TODO: 实现删除分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// GetTags 获取所有设备标签
func (h *DeviceHandler) GetTags(c *gin.Context) {
	tags, err := h.deviceService.GetAllTags(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取标签列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": tags,
	})
}

// GetMetrics 获取设备监控指标
func (h *DeviceHandler) GetMetrics(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	// 获取时间范围参数
	hours := c.DefaultQuery("hours", "24")
	hoursInt, _ := strconv.Atoi(hours)

	metrics, err := h.deviceService.GetMetrics(c.Request.Context(), uint(id), hoursInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取监控指标失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": metrics,
	})
}

// GetTasks 获取设备采集任务
func (h *DeviceHandler) GetTasks(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	tasks, err := h.deviceService.GetTasks(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取采集任务失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": tasks,
	})
}

// GetAlertRules 获取设备告警规则
func (h *DeviceHandler) GetAlertRules(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	rules, err := h.deviceService.GetAlertRules(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取告警规则失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": rules,
	})
}

// GetHistory 获取设备历史记录
func (h *DeviceHandler) GetHistory(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	history, total, err := h.deviceService.GetHistory(c.Request.Context(), uint(id), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取历史记录失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"items": history,
			"total": total,
			"page":  page,
			"page_size": pageSize,
		},
	})
}

