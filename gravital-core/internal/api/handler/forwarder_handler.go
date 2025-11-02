package handler

import (
	"net/http"

	"github.com/celestial/gravital-core/internal/forwarder"
	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ForwarderHandler 转发器处理器
type ForwarderHandler struct {
	service service.ForwarderService
	logger  *zap.Logger
}

// NewForwarderHandler 创建转发器处理器
func NewForwarderHandler(service service.ForwarderService, logger *zap.Logger) *ForwarderHandler {
	return &ForwarderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateForwarder 创建转发器
// @Summary 创建转发器
// @Tags forwarder
// @Accept json
// @Produce json
// @Param forwarder body model.ForwarderConfig true "转发器配置"
// @Success 200 {object} Response
// @Router /api/v1/forwarders [post]
func (h *ForwarderHandler) CreateForwarder(c *gin.Context) {
	var req model.ForwarderConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	if err := h.service.CreateForwarder(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to create forwarder", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"name": req.Name,
	})
}

// UpdateForwarder 更新转发器
// @Summary 更新转发器
// @Tags forwarder
// @Accept json
// @Produce json
// @Param name path string true "转发器名称"
// @Param forwarder body model.ForwarderConfig true "转发器配置"
// @Success 200 {object} Response
// @Router /api/v1/forwarders/{name} [put]
func (h *ForwarderHandler) UpdateForwarder(c *gin.Context) {
	name := c.Param("name")

	var req model.ForwarderConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	// 确保名称一致
	req.Name = name

	if err := h.service.UpdateForwarder(c.Request.Context(), &req); err != nil {
		h.logger.Error("Failed to update forwarder", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "success"})
}

// DeleteForwarder 删除转发器
// @Summary 删除转发器
// @Tags forwarder
// @Produce json
// @Param name path string true "转发器名称"
// @Success 200 {object} Response
// @Router /api/v1/forwarders/{name} [delete]
func (h *ForwarderHandler) DeleteForwarder(c *gin.Context) {
	name := c.Param("name")

	if err := h.service.DeleteForwarder(c.Request.Context(), name); err != nil {
		h.logger.Error("Failed to delete forwarder", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, gin.H{"message": "success"})
}

// GetForwarder 获取转发器
// @Summary 获取转发器
// @Tags forwarder
// @Produce json
// @Param name path string true "转发器名称"
// @Success 200 {object} Response{data=model.ForwarderConfig}
// @Router /api/v1/forwarders/{name} [get]
func (h *ForwarderHandler) GetForwarder(c *gin.Context) {
	name := c.Param("name")

	config, err := h.service.GetForwarder(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get forwarder", zap.Error(err))
		ErrorResponse(c, http.StatusNotFound, 40004, err.Error())
		return
	}

	SuccessResponse(c, config)
}

// ListForwarders 列出转发器
// @Summary 列出转发器
// @Tags forwarder
// @Produce json
// @Param enabled query bool false "是否启用"
// @Success 200 {object} Response{data=[]model.ForwarderConfig}
// @Router /api/v1/forwarders [get]
func (h *ForwarderHandler) ListForwarders(c *gin.Context) {
	var enabled *bool
	if enabledStr := c.Query("enabled"); enabledStr != "" {
		val := enabledStr == "true"
		enabled = &val
	}

	configs, err := h.service.ListForwarders(c.Request.Context(), enabled)
	if err != nil {
		h.logger.Error("Failed to list forwarders", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"total": len(configs),
		"items": configs,
	})
}

// IngestMetrics 接收指标数据
// @Summary 接收指标数据
// @Tags forwarder
// @Accept json
// @Produce json
// @Param metrics body []forwarder.Metric true "指标数据"
// @Success 200 {object} Response
// @Router /api/v1/data/ingest [post]
func (h *ForwarderHandler) IngestMetrics(c *gin.Context) {
	var req struct {
		Metrics []*forwarder.Metric `json:"metrics"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	if len(req.Metrics) == 0 {
		ErrorResponse(c, http.StatusBadRequest, 40001, "no metrics provided")
		return
	}

	// 从请求头获取 Sentinel ID
	sentinelID := c.GetHeader("X-Sentinel-ID")
	if sentinelID != "" {
		// 为所有指标添加 sentinel_id 标签
		for _, metric := range req.Metrics {
			if metric.Labels == nil {
				metric.Labels = make(map[string]string)
			}
			metric.Labels["sentinel_id"] = sentinelID
		}
	}

	if err := h.service.IngestMetrics(c.Request.Context(), req.Metrics); err != nil {
		h.logger.Error("Failed to ingest metrics",
			zap.String("sentinel_id", sentinelID),
			zap.Int("count", len(req.Metrics)),
			zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	h.logger.Info("Ingested metrics",
		zap.String("sentinel_id", sentinelID),
		zap.Int("count", len(req.Metrics)))

	SuccessResponse(c, gin.H{
		"received": len(req.Metrics),
	})
}

// GetForwarderStats 获取转发器统计
// @Summary 获取转发器统计
// @Tags forwarder
// @Produce json
// @Param name path string true "转发器名称"
// @Success 200 {object} Response
// @Router /api/v1/forwarders/{name}/stats [get]
func (h *ForwarderHandler) GetForwarderStats(c *gin.Context) {
	name := c.Param("name")

	stats, err := h.service.GetForwarderStats(c.Request.Context(), name)
	if err != nil {
		h.logger.Error("Failed to get forwarder stats", zap.Error(err))
		ErrorResponse(c, http.StatusNotFound, 40004, err.Error())
		return
	}

	SuccessResponse(c, stats)
}

// GetAllStats 获取所有转发器统计
// @Summary 获取所有转发器统计
// @Tags forwarder
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/forwarders/stats [get]
func (h *ForwarderHandler) GetAllStats(c *gin.Context) {
	stats, err := h.service.GetAllStats(c.Request.Context())
	if err != nil {
		h.logger.Error("Failed to get all stats", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, stats)
}

// ReloadConfig 重新加载配置
// @Summary 重新加载配置
// @Tags forwarder
// @Produce json
// @Success 200 {object} Response
// @Router /api/v1/forwarders/reload [post]
func (h *ForwarderHandler) ReloadConfig(c *gin.Context) {
	if err := h.service.ReloadConfig(); err != nil {
		h.logger.Error("Failed to reload config", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, gin.H{
		"message": "config reloaded successfully",
	})
}

