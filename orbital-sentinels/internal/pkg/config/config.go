package config

import (
	"time"

	"github.com/spf13/viper"
)

// Config 配置结构
type Config struct {
	Sentinel        SentinelConfig  `mapstructure:"sentinel"`
	Core            CoreConfig      `mapstructure:"core"`
	Heartbeat       HeartbeatConfig `mapstructure:"heartbeat"`
	Collector       CollectorConfig `mapstructure:"collector"`
	Buffer          BufferConfig    `mapstructure:"buffer"`
	Sender          SenderConfig    `mapstructure:"sender"`
	Plugins         PluginsConfig   `mapstructure:"plugins"`
	Logging         LoggingConfig   `mapstructure:"logging"`
	Tasks           []TaskConfig    `mapstructure:"tasks"`            // 本地任务配置
	CredentialsPath string          `mapstructure:"credentials_path"` // 凭证文件路径
}

// SentinelConfig Sentinel 配置
type SentinelConfig struct {
	ID     string            `mapstructure:"id"`
	Name   string            `mapstructure:"name"`
	Region string            `mapstructure:"region"`
	Labels map[string]string `mapstructure:"labels"`
}

// CoreConfig 中心端配置
type CoreConfig struct {
	URL                string `mapstructure:"url"`
	APIToken           string `mapstructure:"api_token"`
	RegistrationKey    string `mapstructure:"registration_key"`    // 注册密钥(可选)
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	Interval   time.Duration `mapstructure:"interval"`
	Timeout    time.Duration `mapstructure:"timeout"`
	RetryTimes int           `mapstructure:"retry_times"`
}

// CollectorConfig 采集器配置
type CollectorConfig struct {
	WorkerPoolSize    int           `mapstructure:"worker_pool_size"`
	TaskFetchInterval time.Duration `mapstructure:"task_fetch_interval"`
	MaxExecutionTime  time.Duration `mapstructure:"max_execution_time"`
}

// BufferConfig 缓冲配置
type BufferConfig struct {
	Type          string        `mapstructure:"type"`
	Size          int           `mapstructure:"size"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	DiskPath      string        `mapstructure:"disk_path"`
}

// SenderConfig 发送器配置
type SenderConfig struct {
	Mode          string        `mapstructure:"mode"`
	BatchSize     int           `mapstructure:"batch_size"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	Timeout       time.Duration `mapstructure:"timeout"`
	RetryTimes    int           `mapstructure:"retry_times"`
	RetryInterval time.Duration `mapstructure:"retry_interval"`
	Direct        DirectConfig  `mapstructure:"direct"`
}

// DirectConfig 直连配置
type DirectConfig struct {
	Prometheus      DatabaseConfig `mapstructure:"prometheus"`
	VictoriaMetrics DatabaseConfig `mapstructure:"victoria_metrics"`
	ClickHouse      DatabaseConfig `mapstructure:"clickhouse"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Enabled   bool              `mapstructure:"enabled"`
	URL       string            `mapstructure:"url"`
	DSN       string            `mapstructure:"dsn"`
	Username  string            `mapstructure:"username"`
	Password  string            `mapstructure:"password"`
	TableName string            `mapstructure:"table_name"`
	BatchSize int               `mapstructure:"batch_size"`
	Headers   map[string]string `mapstructure:"headers"`
}

// PluginsConfig 插件配置
type PluginsConfig struct {
	Directory      string        `mapstructure:"directory"`
	AutoReload     bool          `mapstructure:"auto_reload"`
	ReloadInterval time.Duration `mapstructure:"reload_interval"`
}

// LoggingConfig 日志配置
type LoggingConfig struct {
	Level      string `mapstructure:"level"`
	Format     string `mapstructure:"format"`
	Output     string `mapstructure:"output"`
	FilePath   string `mapstructure:"file_path"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

// TaskConfig 任务配置
type TaskConfig struct {
	ID       string                 `mapstructure:"id"`
	DeviceID string                 `mapstructure:"device_id"`
	Plugin   string                 `mapstructure:"plugin"`
	Interval string                 `mapstructure:"interval"` // 如: "60s", "5m", "1h"
	Timeout  string                 `mapstructure:"timeout"`  // 如: "30s", "1m"
	Enabled  bool                   `mapstructure:"enabled"`  // 是否启用
	Config   map[string]interface{} `mapstructure:"config"`   // 插件特定配置
}

// Load 加载配置
func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置环境变量前缀
	viper.SetEnvPrefix("SENTINEL")
	viper.AutomaticEnv()

	// 读取配置文件
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
