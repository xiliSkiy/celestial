package heartbeat

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"runtime"
	"time"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"go.uber.org/zap"
)

// Manager 心跳管理器
type Manager struct {
	client         *http.Client
	coreURL        string
	apiToken       string
	sentinelID     string
	interval       time.Duration
	timeout        time.Duration
	retryTimes     int
	metrics        *SystemMetrics
	ctx            context.Context
	cancel         context.CancelFunc
	onConfigUpdate func(version int)
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	SentinelID    string  `json:"sentinel_id"`
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	TaskCount     int     `json:"task_count"`
	PluginCount   int     `json:"plugin_count"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Version       string  `json:"version"`
}

// HeartbeatResponse 心跳响应
type HeartbeatResponse struct {
	Status        string   `json:"status"`
	ConfigVersion int      `json:"config_version"`
	Commands      []string `json:"commands"`
}

// NewManager 创建心跳管理器
func NewManager(
	coreURL, apiToken, sentinelID string,
	interval, timeout time.Duration,
	retryTimes int,
) *Manager {
	return &Manager{
		client: &http.Client{
			Timeout: timeout,
		},
		coreURL:    coreURL,
		apiToken:   apiToken,
		sentinelID: sentinelID,
		interval:   interval,
		timeout:    timeout,
		retryTimes: retryTimes,
		metrics:    NewSystemMetrics(),
	}
}

// SetConfigUpdateHandler 设置配置更新处理器
func (m *Manager) SetConfigUpdateHandler(handler func(version int)) {
	m.onConfigUpdate = handler
}

// Start 启动心跳
func (m *Manager) Start(ctx context.Context) {
	m.ctx, m.cancel = context.WithCancel(ctx)

	go m.heartbeatLoop()

	logger.Info("Heartbeat started", zap.Duration("interval", m.interval))
}

// Stop 停止心跳
func (m *Manager) Stop() {
	if m.cancel != nil {
		m.cancel()
	}

	logger.Info("Heartbeat stopped")
}

// heartbeatLoop 心跳循环
func (m *Manager) heartbeatLoop() {
	ticker := time.NewTicker(m.interval)
	defer ticker.Stop()

	// 立即发送一次心跳
	m.sendHeartbeat()

	for {
		select {
		case <-ticker.C:
			m.sendHeartbeat()
		case <-m.ctx.Done():
			return
		}
	}
}

// sendHeartbeat 发送心跳
func (m *Manager) sendHeartbeat() {
	ctx, cancel := context.WithTimeout(m.ctx, m.timeout)
	defer cancel()

	// 收集系统指标
	metrics := m.metrics.Collect()

	// 构建请求
	req := &HeartbeatRequest{
		SentinelID:    m.sentinelID,
		CPUUsage:      metrics.CPUUsage,
		MemoryUsage:   metrics.MemoryUsage,
		DiskUsage:     metrics.DiskUsage,
		TaskCount:     metrics.TaskCount,
		PluginCount:   metrics.PluginCount,
		UptimeSeconds: metrics.UptimeSeconds,
		Version:       "1.0.0", // TODO: 从配置或编译时注入
	}

	// 发送请求
	resp, err := m.sendRequest(ctx, req)
	if err != nil {
		// 中心端不可用时只记录警告，不影响采集端运行
		logger.Warn("Failed to send heartbeat (core may be unavailable)", zap.Error(err))
		return
	}

	// 处理响应
	if m.onConfigUpdate != nil && resp.ConfigVersion > 0 {
		m.onConfigUpdate(resp.ConfigVersion)
	}

	logger.Debug("Heartbeat sent successfully")
}

// sendRequest 发送请求
func (m *Manager) sendRequest(ctx context.Context, req *HeartbeatRequest) (*HeartbeatResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		"POST",
		m.coreURL+"/api/v1/sentinels/heartbeat",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Sentinel-ID", m.sentinelID)
	httpReq.Header.Set("X-API-Token", m.apiToken)

	httpResp, err := m.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer httpResp.Body.Close()

	body, _ := io.ReadAll(httpResp.Body)

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d, body: %s", httpResp.StatusCode, string(body))
	}

	var resp struct {
		Code int               `json:"code"`
		Data HeartbeatResponse `json:"data"`
	}

	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &resp.Data, nil
}

// SystemMetrics 系统指标
type SystemMetrics struct {
	startTime time.Time
}

// Metrics 指标数据
type Metrics struct {
	CPUUsage      float64
	MemoryUsage   float64
	DiskUsage     float64
	TaskCount     int
	PluginCount   int
	UptimeSeconds int64
}

// NewSystemMetrics 创建系统指标收集器
func NewSystemMetrics() *SystemMetrics {
	return &SystemMetrics{
		startTime: time.Now(),
	}
}

// Collect 收集指标
func (sm *SystemMetrics) Collect() *Metrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &Metrics{
		CPUUsage:      0, // TODO: 实现 CPU 使用率采集
		MemoryUsage:   float64(m.Alloc) / float64(m.Sys) * 100,
		DiskUsage:     0, // TODO: 实现磁盘使用率采集
		TaskCount:     0, // 由外部设置
		PluginCount:   0, // 由外部设置
		UptimeSeconds: int64(time.Since(sm.startTime).Seconds()),
	}
}
