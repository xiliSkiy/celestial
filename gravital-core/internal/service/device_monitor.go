package service

import (
	"context"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// DeviceMonitor 设备监控服务（定时检查设备状态）
type DeviceMonitor struct {
	db              *gorm.DB
	logger          *zap.Logger
	checkInterval   time.Duration // 检查间隔
	offlineTimeout  time.Duration // 离线超时时间
	ticker          *time.Ticker
	done            chan struct{}
	ctx             context.Context
	cancel          context.CancelFunc
}

// DeviceMonitorConfig 设备监控配置
type DeviceMonitorConfig struct {
	CheckInterval  time.Duration // 检查间隔，默认 1 分钟
	OfflineTimeout time.Duration // 离线超时，默认 5 分钟
}

// NewDeviceMonitor 创建设备监控服务
func NewDeviceMonitor(db *gorm.DB, logger *zap.Logger, config *DeviceMonitorConfig) *DeviceMonitor {
	if config == nil {
		config = &DeviceMonitorConfig{
			CheckInterval:  1 * time.Minute,
			OfflineTimeout: 5 * time.Minute,
		}
	}

	// 设置默认值
	if config.CheckInterval == 0 {
		config.CheckInterval = 1 * time.Minute
	}
	if config.OfflineTimeout == 0 {
		config.OfflineTimeout = 5 * time.Minute
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &DeviceMonitor{
		db:             db,
		logger:         logger,
		checkInterval:  config.CheckInterval,
		offlineTimeout: config.OfflineTimeout,
		done:           make(chan struct{}),
		ctx:            ctx,
		cancel:         cancel,
	}
}

// Start 启动监控
func (m *DeviceMonitor) Start() {
	m.ticker = time.NewTicker(m.checkInterval)

	go func() {
		m.logger.Info("Device monitor started",
			zap.Duration("check_interval", m.checkInterval),
			zap.Duration("offline_timeout", m.offlineTimeout))

		// 立即执行一次检查
		m.checkDeviceStatus()

		for {
			select {
			case <-m.ticker.C:
				m.checkDeviceStatus()
			case <-m.done:
				m.logger.Info("Device monitor stopped")
				return
			}
		}
	}()
}

// Stop 停止监控
func (m *DeviceMonitor) Stop() {
	if m.ticker != nil {
		m.ticker.Stop()
	}
	m.cancel()
	close(m.done)
}

// checkDeviceStatus 检查设备状态
func (m *DeviceMonitor) checkDeviceStatus() {
	// 计算超时时间点
	timeoutThreshold := time.Now().Add(-m.offlineTimeout)

	// 查找长时间未上报的在线设备
	var offlineCount int64
	result := m.db.WithContext(m.ctx).
		Model(&model.Device{}).
		Where("status = ?", "online").
		Where("last_seen < ? OR last_seen IS NULL", timeoutThreshold).
		Updates(map[string]interface{}{
			"status":     "offline",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		m.logger.Error("Failed to update offline devices", zap.Error(result.Error))
		return
	}

	offlineCount = result.RowsAffected

	if offlineCount > 0 {
		m.logger.Info("Marked devices as offline",
			zap.Int64("count", offlineCount),
			zap.Time("timeout_threshold", timeoutThreshold))
	}

	// 统计当前设备状态
	m.logDeviceStats()
}

// logDeviceStats 记录设备状态统计
func (m *DeviceMonitor) logDeviceStats() {
	var stats struct {
		Total   int64
		Online  int64
		Offline int64
		Unknown int64
	}

	m.db.WithContext(m.ctx).Model(&model.Device{}).Count(&stats.Total)
	m.db.WithContext(m.ctx).Model(&model.Device{}).Where("status = ?", "online").Count(&stats.Online)
	m.db.WithContext(m.ctx).Model(&model.Device{}).Where("status = ?", "offline").Count(&stats.Offline)
	m.db.WithContext(m.ctx).Model(&model.Device{}).Where("status = ?", "unknown").Count(&stats.Unknown)

	m.logger.Debug("Device status statistics",
		zap.Int64("total", stats.Total),
		zap.Int64("online", stats.Online),
		zap.Int64("offline", stats.Offline),
		zap.Int64("unknown", stats.Unknown))
}

