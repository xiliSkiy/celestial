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

