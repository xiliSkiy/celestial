package forwarder

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

// PrometheusForwarder Prometheus 转发器
type PrometheusForwarder struct {
	config *ForwarderConfig
	client *http.Client
	logger *zap.Logger
	stats  Stats
	mu     sync.RWMutex
}

// NewPrometheusForwarder 创建 Prometheus 转发器
func NewPrometheusForwarder(config *ForwarderConfig, logger *zap.Logger) (*PrometheusForwarder, error) {
	if config.Endpoint == "" {
		return nil, fmt.Errorf("prometheus endpoint is required")
	}

	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	return &PrometheusForwarder{
		config: config,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger,
	}, nil
}

// Write 写入指标数据
func (f *PrometheusForwarder) Write(metrics []*Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	startTime := time.Now()

	// 转换为 Prometheus Remote Write 格式
	writeRequest := f.convertToPrometheusFormat(metrics)

	// 序列化
	data, err := writeRequest.Marshal()
	if err != nil {
		f.recordError()
		return fmt.Errorf("failed to marshal write request: %w", err)
	}

	// Snappy 压缩
	compressed := snappy.Encode(nil, data)

	// 创建 HTTP 请求
	ctx, cancel := context.WithTimeout(context.Background(), f.config.Timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", f.config.Endpoint, bytes.NewReader(compressed))
	if err != nil {
		f.recordError()
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	// 设置认证
	if f.config.Username != "" && f.config.Password != "" {
		req.SetBasicAuth(f.config.Username, f.config.Password)
	}

	// 发送请求
	resp, err := f.client.Do(req)
	if err != nil {
		f.recordError()
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		f.recordError()
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	// 记录成功
	latency := time.Since(startTime).Milliseconds()
	f.recordSuccess(int64(len(compressed)), latency)

	f.logger.Debug("Wrote to Prometheus",
		zap.String("forwarder", f.config.Name),
		zap.Int("metrics", len(metrics)),
		zap.Int("bytes", len(compressed)),
		zap.Int64("latency_ms", latency))

	return nil
}

// convertToPrometheusFormat 转换为 Prometheus 格式
func (f *PrometheusForwarder) convertToPrometheusFormat(metrics []*Metric) *prompb.WriteRequest {
	timeseries := make([]prompb.TimeSeries, 0, len(metrics))

	for _, metric := range metrics {
		// 构建标签
		labels := make([]prompb.Label, 0, len(metric.Labels)+1)
		labels = append(labels, prompb.Label{
			Name:  "__name__",
			Value: metric.Name,
		})

		for key, value := range metric.Labels {
			labels = append(labels, prompb.Label{
				Name:  key,
				Value: value,
			})
		}

		// 构建样本
		timestamp := metric.Timestamp
		if timestamp == 0 {
			timestamp = time.Now().Unix()
		}

		sample := prompb.Sample{
			Value:     metric.Value,
			Timestamp: timestamp * 1000, // 转换为毫秒
		}

		timeseries = append(timeseries, prompb.TimeSeries{
			Labels:  labels,
			Samples: []prompb.Sample{sample},
		})
	}

	return &prompb.WriteRequest{
		Timeseries: timeseries,
	}
}

// Close 关闭转发器
func (f *PrometheusForwarder) Close() error {
	f.client.CloseIdleConnections()
	return nil
}

// Name 获取转发器名称
func (f *PrometheusForwarder) Name() string {
	return f.config.Name
}

// Type 获取转发器类型
func (f *PrometheusForwarder) Type() ForwarderType {
	return ForwarderTypePrometheus
}

// IsEnabled 是否启用
func (f *PrometheusForwarder) IsEnabled() bool {
	return f.config.Enabled
}

// GetStats 获取统计信息
func (f *PrometheusForwarder) GetStats() Stats {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.stats
}

// recordSuccess 记录成功
func (f *PrometheusForwarder) recordSuccess(bytes int64, latencyMs int64) {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.stats.SuccessCount++
	f.stats.TotalBytes += bytes
	f.stats.LastSuccess = time.Now()

	// 计算平均延迟
	if f.stats.AvgLatencyMs == 0 {
		f.stats.AvgLatencyMs = latencyMs
	} else {
		f.stats.AvgLatencyMs = (f.stats.AvgLatencyMs + latencyMs) / 2
	}
}

// recordError 记录错误
func (f *PrometheusForwarder) recordError() {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.stats.FailedCount++
}

