package plugin

import (
	"context"
	"time"
)

// PluginMeta 插件元信息
type PluginMeta struct {
	Name        string   `yaml:"name"`
	Version     string   `yaml:"version"`
	Description string   `yaml:"description"`
	Author      string   `yaml:"author"`
	DeviceTypes []string `yaml:"device_types"` // 支持的设备类型
}

// DeviceField 设备字段定义
type DeviceField struct {
	Name        string      `yaml:"name"`
	Type        string      `yaml:"type"` // string, int, bool, password
	Required    bool        `yaml:"required"`
	Default     interface{} `yaml:"default"`
	Description string      `yaml:"description"`
	Validation  string      `yaml:"validation"` // 正则表达式
	Min         int         `yaml:"min"`
	Max         int         `yaml:"max"`
}

// PluginSchema 插件配置 Schema
type PluginSchema struct {
	Meta         PluginMeta    `yaml:"meta"`
	DeviceFields []DeviceField `yaml:"device_fields"`
	ConfigFields []DeviceField `yaml:"config_fields"` // 插件级配置
}

// Plugin 插件接口
type Plugin interface {
	// Meta 获取插件元信息
	Meta() PluginMeta

	// Schema 获取配置 Schema
	Schema() PluginSchema

	// Init 初始化插件
	Init(config map[string]interface{}) error

	// ValidateConfig 验证设备连接配置
	ValidateConfig(deviceConfig map[string]interface{}) error

	// TestConnection 测试连接
	TestConnection(deviceConfig map[string]interface{}) error

	// Collect 采集数据
	Collect(ctx context.Context, task *CollectionTask) ([]*Metric, error)

	// Close 关闭插件
	Close() error
}

// CollectionTask 采集任务
type CollectionTask struct {
	TaskID       string
	DeviceID     string
	PluginName   string
	DeviceConfig map[string]interface{}
	PluginConfig map[string]interface{}
	Timeout      time.Duration
}

// Metric 指标数据
type Metric struct {
	Name      string
	Value     float64
	Timestamp int64
	Labels    map[string]string
	Type      MetricType // gauge, counter, histogram
}

// MetricType 指标类型
type MetricType string

const (
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeCounter   MetricType = "counter"
	MetricTypeHistogram MetricType = "histogram"
)
