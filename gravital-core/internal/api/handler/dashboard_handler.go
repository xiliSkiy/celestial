package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// DashboardHandler Dashboard 处理器
type DashboardHandler struct {
	db *gorm.DB
}

// NewDashboardHandler 创建 Dashboard 处理器
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{
		db: db,
	}
}

// GetStats 获取统计数据
func (h *DashboardHandler) GetStats(c *gin.Context) {
	var stats struct {
		TotalDevices    int64 `json:"total_devices"`
		OnlineDevices   int64 `json:"online_devices"`
		OfflineDevices  int64 `json:"offline_devices"`
		ErrorDevices    int64 `json:"error_devices"`
		ActiveAlerts    int64 `json:"active_alerts"`
		TotalTasks      int64 `json:"total_tasks"`
		ActiveSentinels int64 `json:"active_sentinels"`
		TotalSentinels  int64 `json:"total_sentinels"`
	}

	// 设备统计
	h.db.Model(&model.Device{}).Count(&stats.TotalDevices)
	h.db.Model(&model.Device{}).Where("status = ?", "online").Count(&stats.OnlineDevices)
	h.db.Model(&model.Device{}).Where("status = ?", "offline").Count(&stats.OfflineDevices)
	h.db.Model(&model.Device{}).Where("status = ?", "error").Count(&stats.ErrorDevices)

	// 告警统计
	h.db.Model(&model.AlertEvent{}).Where("status = ?", "firing").Count(&stats.ActiveAlerts)

	// 任务统计
	h.db.Model(&model.CollectionTask{}).Count(&stats.TotalTasks)

	// Sentinel 统计
	h.db.Model(&model.Sentinel{}).Count(&stats.TotalSentinels)
	h.db.Model(&model.Sentinel{}).Where("status = ?", "online").Count(&stats.ActiveSentinels)

	SuccessResponse(c, stats)
}

// GetDeviceStatus 获取设备状态分布
func (h *DashboardHandler) GetDeviceStatus(c *gin.Context) {
	var results []struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}

	h.db.Model(&model.Device{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&results)

	SuccessResponse(c, results)
}

// GetAlertTrend 获取告警趋势
func (h *DashboardHandler) GetAlertTrend(c *gin.Context) {
	hoursStr := c.DefaultQuery("hours", "24")
	hours, _ := strconv.Atoi(hoursStr)

	startTime := time.Now().Add(-time.Duration(hours) * time.Hour)

	var results []struct {
		Time     time.Time `json:"time"`
		Critical int64     `json:"critical"`
		Warning  int64     `json:"warning"`
		Info     int64     `json:"info"`
	}

	// 按小时分组统计
	h.db.Model(&model.AlertEvent{}).
		Select(`
			DATE_TRUNC('hour', triggered_at) as time,
			COUNT(CASE WHEN severity = 'critical' THEN 1 END) as critical,
			COUNT(CASE WHEN severity = 'warning' THEN 1 END) as warning,
			COUNT(CASE WHEN severity = 'info' THEN 1 END) as info
		`).
		Where("triggered_at >= ?", startTime).
		Group("DATE_TRUNC('hour', triggered_at)").
		Order("time").
		Scan(&results)

	SuccessResponse(c, results)
}

// GetSentinelStatus 获取 Sentinel 状态
func (h *DashboardHandler) GetSentinelStatus(c *gin.Context) {
	var results []struct {
		Region  string `json:"region"`
		Online  int64  `json:"online"`
		Offline int64  `json:"offline"`
	}

	h.db.Model(&model.Sentinel{}).
		Select(`
			region,
			COUNT(CASE WHEN status = 'online' THEN 1 END) as online,
			COUNT(CASE WHEN status = 'offline' THEN 1 END) as offline
		`).
		Group("region").
		Scan(&results)

	SuccessResponse(c, results)
}

// GetForwarderStats 获取转发器统计
func (h *DashboardHandler) GetForwarderStats(c *gin.Context) {
	var results []struct {
		Name         string `json:"name"`
		SuccessCount int64  `json:"success_count"`
		FailureCount int64  `json:"failure_count"`
	}

	h.db.Model(&model.ForwarderStats{}).
		Select("name, success_count, failure_count").
		Scan(&results)

	SuccessResponse(c, results)
}

// GetActivities 获取最近活动
func (h *DashboardHandler) GetActivities(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "10")
	limit, _ := strconv.Atoi(limitStr)

	type Activity struct {
		ID        uint      `json:"id"`
		Type      string    `json:"type"`
		Content   string    `json:"content"`
		CreatedAt time.Time `json:"created_at"`
	}

	var activities []Activity

	// 从告警事件获取活动
	var alertEvents []model.AlertEvent
	h.db.Order("triggered_at DESC").Limit(limit).Find(&alertEvents)

	for _, event := range alertEvents {
		actType := "info"
		if event.Severity == "critical" {
			actType = "danger"
		} else if event.Severity == "warning" {
			actType = "warning"
		}

		activities = append(activities, Activity{
			ID:        event.ID,
			Type:      actType,
			Content:   event.Message,
			CreatedAt: event.TriggeredAt,
		})
	}

	SuccessResponse(c, activities)
}

