package sender

import (
	"context"
	"testing"
	"time"

	"github.com/celestial/orbital-sentinels/internal/plugin"
)

func TestPrometheusWriter_ConvertFormat(t *testing.T) {
	writer := NewPrometheusWriter("http://localhost:9090/api/v1/write", 30*time.Second)

	metrics := []*plugin.Metric{
		{
			Name:      "cpu_usage",
			Value:     75.5,
			Type:      plugin.MetricTypeGauge,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"host":   "server1",
				"cpu":    "cpu0",
				"region": "us-west",
			},
		},
		{
			Name:      "memory_usage",
			Value:     8589934592, // 8GB in bytes
			Type:      plugin.MetricTypeGauge,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"host":   "server1",
				"region": "us-west",
			},
		},
	}

	// 测试转换
	writeRequest := writer.convertToPrometheusFormat(metrics)

	if writeRequest == nil {
		t.Fatal("Write request is nil")
	}

	if len(writeRequest.Timeseries) != 2 {
		t.Errorf("Expected 2 timeseries, got %d", len(writeRequest.Timeseries))
	}

	// 检查第一个时间序列
	ts := writeRequest.Timeseries[0]
	if len(ts.Labels) != 4 { // __name__ + 3 labels
		t.Errorf("Expected 4 labels, got %d", len(ts.Labels))
	}

	// 检查 __name__ 标签
	found := false
	for _, label := range ts.Labels {
		if label.Name == "__name__" && label.Value == "cpu_usage" {
			found = true
			break
		}
	}
	if !found {
		t.Error("__name__ label not found or incorrect")
	}

	// 检查样本
	if len(ts.Samples) != 1 {
		t.Errorf("Expected 1 sample, got %d", len(ts.Samples))
	}

	if ts.Samples[0].Value != 75.5 {
		t.Errorf("Expected value 75.5, got %f", ts.Samples[0].Value)
	}
}

func TestVictoriaMetricsWriter_ConvertFormat(t *testing.T) {
	writer := NewVictoriaMetricsWriter("http://localhost:8428/api/v1/write", 30*time.Second)

	metrics := []*plugin.Metric{
		{
			Name:      "disk_usage",
			Value:     85.2,
			Type:      plugin.MetricTypeGauge,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"host":       "server1",
				"mount":      "/data",
				"filesystem": "ext4",
			},
		},
	}

	// 测试转换
	writeRequest := writer.convertToVictoriaMetricsFormat(metrics)

	if writeRequest == nil {
		t.Fatal("Write request is nil")
	}

	if len(writeRequest.Timeseries) != 1 {
		t.Errorf("Expected 1 timeseries, got %d", len(writeRequest.Timeseries))
	}

	ts := writeRequest.Timeseries[0]
	if len(ts.Labels) != 4 { // __name__ + 3 labels
		t.Errorf("Expected 4 labels, got %d", len(ts.Labels))
	}
}

func TestDirectSender_Send(t *testing.T) {
	// 创建直连发送器
	directSender := NewDirectSender()

	// 创建测试指标
	metrics := []*plugin.Metric{
		{
			Name:      "test_metric",
			Value:     100.0,
			Type:      plugin.MetricTypeCounter,
			Timestamp: time.Now().Unix(),
			Labels: map[string]string{
				"test": "true",
			},
		},
	}

	// 测试没有配置任何写入器的情况
	ctx := context.Background()
	err := directSender.Send(ctx, metrics)
	if err != nil {
		t.Errorf("Expected no error when no writers configured, got %v", err)
	}

	// 测试空指标
	err = directSender.Send(ctx, []*plugin.Metric{})
	if err != nil {
		t.Errorf("Expected no error for empty metrics, got %v", err)
	}
}

func TestPrometheusWriter_SetBasicAuth(t *testing.T) {
	writer := NewPrometheusWriter("http://localhost:9090/api/v1/write", 30*time.Second)

	writer.SetBasicAuth("user", "pass")

	if writer.username != "user" {
		t.Errorf("Expected username 'user', got '%s'", writer.username)
	}

	if writer.password != "pass" {
		t.Errorf("Expected password 'pass', got '%s'", writer.password)
	}
}

func TestPrometheusWriter_SetHeader(t *testing.T) {
	writer := NewPrometheusWriter("http://localhost:9090/api/v1/write", 30*time.Second)

	writer.SetHeader("X-Custom-Header", "custom-value")

	if writer.headers["X-Custom-Header"] != "custom-value" {
		t.Errorf("Expected header 'custom-value', got '%s'", writer.headers["X-Custom-Header"])
	}
}

func TestVictoriaMetricsWriter_SetBasicAuth(t *testing.T) {
	writer := NewVictoriaMetricsWriter("http://localhost:8428/api/v1/write", 30*time.Second)

	writer.SetBasicAuth("admin", "secret")

	if writer.username != "admin" {
		t.Errorf("Expected username 'admin', got '%s'", writer.username)
	}

	if writer.password != "secret" {
		t.Errorf("Expected password 'secret', got '%s'", writer.password)
	}
}

func TestVictoriaMetricsWriter_SetHeader(t *testing.T) {
	writer := NewVictoriaMetricsWriter("http://localhost:8428/api/v1/write", 30*time.Second)

	writer.SetHeader("X-Scope-OrgID", "tenant1")

	if writer.headers["X-Scope-OrgID"] != "tenant1" {
		t.Errorf("Expected header 'tenant1', got '%s'", writer.headers["X-Scope-OrgID"])
	}
}

func TestDirectSender_Close(t *testing.T) {
	directSender := NewDirectSender()

	// 测试没有 ClickHouse 写入器的情况
	err := directSender.Close()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}
