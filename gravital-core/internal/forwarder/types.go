package forwarder

import "time"

// Metric 指标数据
type Metric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Type      string            `json:"type"`
	Labels    map[string]string `json:"labels"`
	Timestamp int64             `json:"timestamp"`
}

// ForwarderType 转发器类型
type ForwarderType string

const (
	ForwarderTypePrometheus      ForwarderType = "prometheus"
	ForwarderTypeVictoriaMetrics ForwarderType = "victoria-metrics"
	ForwarderTypeClickHouse      ForwarderType = "clickhouse"
)

// Forwarder 转发器接口
type Forwarder interface {
	// Write 写入指标数据
	Write(metrics []*Metric) error

	// Close 关闭转发器
	Close() error

	// Name 获取转发器名称
	Name() string

	// Type 获取转发器类型
	Type() ForwarderType

	// IsEnabled 是否启用
	IsEnabled() bool
}

// ForwarderConfig 转发器配置
type ForwarderConfig struct {
	Name          string
	Type          ForwarderType
	Enabled       bool
	Endpoint      string
	DSN           string
	Table         string
	Username      string
	Password      string
	Timeout       time.Duration
	BatchSize     int
	FlushInterval time.Duration
	MaxRetries    int
	RetryInterval time.Duration
}

// Stats 统计信息
type Stats struct {
	SuccessCount int64
	FailedCount  int64
	TotalBytes   int64
	AvgLatencyMs int64
	LastError    string
	LastSuccess  time.Time
}
