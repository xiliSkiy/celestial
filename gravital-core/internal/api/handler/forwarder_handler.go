package handler

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/celestial/gravital-core/internal/forwarder"
	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// ForwarderHandler 转发器处理器
type ForwarderHandler struct {
	service service.ForwarderService
	db      *gorm.DB
	logger  *zap.Logger
}

// NewForwarderHandler 创建转发器处理器
func NewForwarderHandler(service service.ForwarderService, db *gorm.DB, logger *zap.Logger) *ForwarderHandler {
	return &ForwarderHandler{
		service: service,
		db:      db,
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

	// 获取所有转发器的统计信息
	allStats, err := h.service.GetAllStats(c.Request.Context())
	if err != nil {
		h.logger.Warn("Failed to get forwarder stats", zap.Error(err))
	}

	// 将统计信息合并到配置中
	items := make([]map[string]interface{}, 0, len(configs))
	for _, config := range configs {
		item := map[string]interface{}{
			"id":              config.ID,
			"name":            config.Name,
			"type":            config.Type,
			"enabled":         config.Enabled,
			"endpoint":        config.Endpoint,
			"auth_config":     config.AuthConfig,
			"batch_size":      config.BatchSize,
			"flush_interval":  config.FlushInterval,
			"retry_times":     config.RetryTimes,
			"timeout_seconds": config.TimeoutSeconds,
			"created_at":      config.CreatedAt,
			"updated_at":      config.UpdatedAt,
			"success_count":   int64(0),
			"failure_count":   int64(0),
			"avg_latency":     int64(0),
		}

		// 如果有统计信息，添加到结果中
		if allStats != nil {
			if fwdStats, ok := allStats[config.Name].(map[string]interface{}); ok {
				if stats, ok := fwdStats["stats"].(forwarder.Stats); ok {
					item["success_count"] = stats.SuccessCount
					item["failure_count"] = stats.FailedCount
					item["avg_latency"] = stats.AvgLatencyMs
				}
			}
		}

		items = append(items, item)
	}

	SuccessResponse(c, gin.H{
		"total": len(items),
		"items": items,
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

	// 检查是否是 gzip 压缩数据
	var reader io.Reader = c.Request.Body
	if c.GetHeader("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(c.Request.Body)
		if err != nil {
			h.logger.Error("Failed to create gzip reader", zap.Error(err))
			ErrorResponse(c, http.StatusBadRequest, 40001, "invalid gzip data: "+err.Error())
			return
		}
		defer gzipReader.Close()
		reader = gzipReader
	}

	// 读取并解析 JSON
	body, err := io.ReadAll(reader)
	if err != nil {
		h.logger.Error("Failed to read request body", zap.Error(err))
		ErrorResponse(c, http.StatusBadRequest, 40001, "failed to read body: "+err.Error())
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		h.logger.Error("Failed to unmarshal JSON",
			zap.Error(err),
			zap.String("body_preview", string(body[:min(len(body), 200)])))
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

	// 提取设备状态信息并更新 PostgreSQL
	deviceStatusMap := h.extractDeviceStatus(req.Metrics)
	if len(deviceStatusMap) > 0 {
		if err := h.updateDeviceStatusInDB(c.Request.Context(), deviceStatusMap); err != nil {
			h.logger.Error("Failed to update device status",
				zap.Int("device_count", len(deviceStatusMap)),
				zap.Error(err))
			// 不中断流程，继续转发指标
		}
	}

	// 转发指标到时序库（包含设备状态指标）
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
		zap.Int("count", len(req.Metrics)),
		zap.Int("devices_updated", len(deviceStatusMap)))

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

// TestConnection 测试转发器连接
// @Summary 测试转发器连接
// @Tags forwarder
// @Accept json
// @Produce json
// @Param config body model.ForwarderConfig true "转发器配置"
// @Success 200 {object} Response{data=service.ForwarderTestConnectionResult}
// @Router /api/v1/forwarders/test [post]
func (h *ForwarderHandler) TestConnection(c *gin.Context) {
	var req model.ForwarderConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		ErrorResponse(c, http.StatusBadRequest, 40001, err.Error())
		return
	}

	result, err := h.service.TestConnection(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("Failed to test connection", zap.Error(err))
		ErrorResponse(c, http.StatusInternalServerError, 10001, err.Error())
		return
	}

	SuccessResponse(c, result)
}

// DeviceStatusInfo 设备状态信息
type DeviceStatusInfo struct {
	Status   string
	LastSeen time.Time
}

// extractDeviceStatus 从指标中提取设备状态信息
func (h *ForwarderHandler) extractDeviceStatus(metrics []*forwarder.Metric) map[string]DeviceStatusInfo {
	statusMap := make(map[string]DeviceStatusInfo)

	for _, m := range metrics {
		// 只处理 device_status 指标
		if m.Name != "device_status" {
			continue
		}

		// 获取 device_id
		deviceID, ok := m.Labels["device_id"]
		if !ok || deviceID == "" {
			continue
		}

		// 解析状态值（1=online, 0=offline）
		status := "offline"
		if m.Value == 1.0 {
			status = "online"
		}

		// 使用指标的时间戳
		lastSeen := time.UnixMilli(m.Timestamp)

		statusMap[deviceID] = DeviceStatusInfo{
			Status:   status,
			LastSeen: lastSeen,
		}
	}

	return statusMap
}

// updateDeviceStatusInDB 更新设备状态到 PostgreSQL
func (h *ForwarderHandler) updateDeviceStatusInDB(ctx context.Context, statusMap map[string]DeviceStatusInfo) error {
	if len(statusMap) == 0 {
		return nil
	}

	// 批量更新设备状态
	for deviceID, info := range statusMap {
		updates := map[string]interface{}{
			"status":     info.Status,
			"last_seen":  info.LastSeen,
			"updated_at": time.Now(),
		}

		result := h.db.WithContext(ctx).
			Model(&model.Device{}).
			Where("device_id = ?", deviceID).
			Updates(updates)

		if result.Error != nil {
			h.logger.Error("Failed to update device status",
				zap.String("device_id", deviceID),
				zap.String("status", info.Status),
				zap.Error(result.Error))
			// 继续处理其他设备，不中断
			continue
		}

		if result.RowsAffected > 0 {
			h.logger.Debug("Updated device status",
				zap.String("device_id", deviceID),
				zap.String("status", info.Status))
		}
	}

	return nil
}
