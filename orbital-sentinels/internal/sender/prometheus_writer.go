package sender

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/celestial/orbital-sentinels/internal/pkg/logger"
	"github.com/celestial/orbital-sentinels/internal/plugin"
	"github.com/golang/snappy"
	"github.com/prometheus/prometheus/prompb"
	"go.uber.org/zap"
)

// PrometheusWriter Prometheus Remote Write 实现
type PrometheusWriter struct {
	client   *http.Client
	url      string
	username string
	password string
	headers  map[string]string
}

// NewPrometheusWriter 创建 Prometheus 写入器
func NewPrometheusWriter(url string, timeout time.Duration) *PrometheusWriter {
	return &PrometheusWriter{
		client: &http.Client{
			Timeout: timeout,
		},
		url:     url,
		headers: make(map[string]string),
	}
}

// SetBasicAuth 设置基本认证
func (pw *PrometheusWriter) SetBasicAuth(username, password string) {
	pw.username = username
	pw.password = password
}

// SetHeader 设置自定义请求头
func (pw *PrometheusWriter) SetHeader(key, value string) {
	pw.headers[key] = value
}

// Write 写入数据
func (pw *PrometheusWriter) Write(ctx context.Context, metrics []*plugin.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// 转换为 Prometheus Remote Write 格式
	writeRequest := pw.convertToPrometheusFormat(metrics)

	// 序列化
	data, err := writeRequest.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal write request: %w", err)
	}

	// Snappy 压缩
	compressed := snappy.Encode(nil, data)

	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", pw.url, bytes.NewReader(compressed))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	// 设置自定义请求头
	for key, value := range pw.headers {
		req.Header.Set(key, value)
	}

	// 设置认证
	if pw.username != "" && pw.password != "" {
		req.SetBasicAuth(pw.username, pw.password)
	}

	// 发送请求
	resp, err := pw.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := io.ReadAll(resp.Body)

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	logger.Debug("Wrote to Prometheus",
		zap.Int("metrics", len(metrics)),
		zap.Int("bytes", len(compressed)))

	return nil
}

// convertToPrometheusFormat 转换为 Prometheus 格式
func (pw *PrometheusWriter) convertToPrometheusFormat(metrics []*plugin.Metric) *prompb.WriteRequest {
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
		sample := prompb.Sample{
			Value:     metric.Value,
			Timestamp: metric.Timestamp * 1000, // 转换为毫秒
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
