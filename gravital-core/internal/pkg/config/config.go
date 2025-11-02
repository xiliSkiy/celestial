package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 全局配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Auth      AuthConfig      `mapstructure:"auth"`
	Alert     AlertConfig     `mapstructure:"alert"`
	Forwarder ForwarderConfig `mapstructure:"forwarder"`
	Sentinel  SentinelConfig  `mapstructure:"sentinel"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Logging   LoggingConfig   `mapstructure:"logging"`
	Grafana   GrafanaConfig   `mapstructure:"grafana"`
	System    SystemConfig    `mapstructure:"system"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host           string        `mapstructure:"host"`
	Port           int           `mapstructure:"port"`
	Mode           string        `mapstructure:"mode"`
	ReadTimeout    time.Duration `mapstructure:"read_timeout"`
	WriteTimeout   time.Duration `mapstructure:"write_timeout"`
	MaxHeaderBytes int           `mapstructure:"max_header_bytes"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Database        string        `mapstructure:"database"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	LogLevel        string        `mapstructure:"log_level"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret     string        `mapstructure:"jwt_secret"`
	JWTExpire     time.Duration `mapstructure:"jwt_expire"`
	RefreshExpire time.Duration `mapstructure:"refresh_expire"`
	BcryptCost    int           `mapstructure:"bcrypt_cost"`
}

// AlertConfig 告警配置
type AlertConfig struct {
	EvaluationInterval      time.Duration `mapstructure:"evaluation_interval"`
	NotificationTimeout     time.Duration `mapstructure:"notification_timeout"`
	MaxConcurrentEvaluations int           `mapstructure:"max_concurrent_evaluations"`
	RetentionDays           int           `mapstructure:"retention_days"`
}

// ForwarderConfig 转发器配置
type ForwarderConfig struct {
	BufferSize    int                 `mapstructure:"buffer_size"`
	BatchSize     int                 `mapstructure:"batch_size"`
	FlushInterval time.Duration       `mapstructure:"flush_interval"`
	MaxRetries    int                 `mapstructure:"max_retries"`
	RetryInterval time.Duration       `mapstructure:"retry_interval"`
	Targets       []ForwarderTarget   `mapstructure:"targets"`
}

// ForwarderTarget 转发目标配置
type ForwarderTarget struct {
	Name      string        `mapstructure:"name"`
	Type      string        `mapstructure:"type"`
	Enabled   bool          `mapstructure:"enabled"`
	Endpoint  string        `mapstructure:"endpoint"`
	DSN       string        `mapstructure:"dsn"`
	Table     string        `mapstructure:"table"`
	Timeout   time.Duration `mapstructure:"timeout"`
	BatchSize int           `mapstructure:"batch_size"`
	Username  string        `mapstructure:"username"`
	Password  string        `mapstructure:"password"`
}

// SentinelConfig Sentinel 配置
type SentinelConfig struct {
	HeartbeatTimeout  time.Duration `mapstructure:"heartbeat_timeout"`
	OfflineThreshold  time.Duration `mapstructure:"offline_threshold"`
	TaskFetchInterval time.Duration `mapstructure:"task_fetch_interval"`
	AutoAssign        bool          `mapstructure:"auto_assign"`
}

// SchedulerConfig 调度器配置
type SchedulerConfig struct {
	WorkerPoolSize  int           `mapstructure:"worker_pool_size"`
	TaskQueueSize   int           `mapstructure:"task_queue_size"`
	DefaultInterval time.Duration `mapstructure:"default_interval"`
	DefaultTimeout  time.Duration `mapstructure:"default_timeout"`
	DefaultRetry    int           `mapstructure:"default_retry"`
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
	Compress   bool   `mapstructure:"compress"`
}

// GrafanaConfig Grafana 配置
type GrafanaConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	URL           string `mapstructure:"url"`
	AdminUser     string `mapstructure:"admin_user"`
	AdminPassword string `mapstructure:"admin_password"`
	OrgID         int    `mapstructure:"org_id"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	MaxDevices          int  `mapstructure:"max_devices"`
	MaxSentinels        int  `mapstructure:"max_sentinels"`
	MaxTasksPerSentinel int  `mapstructure:"max_tasks_per_sentinel"`
	DataRetentionDays   int  `mapstructure:"data_retention_days"`
	EnableMetrics       bool `mapstructure:"enable_metrics"`
	EnableProfiling     bool `mapstructure:"enable_profiling"`
	ProfilingPort       int  `mapstructure:"profiling_port"`
}

var globalConfig *Config

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	v := viper.New()
	
	// 设置配置文件
	if configPath != "" {
		v.SetConfigFile(configPath)
	} else {
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath("./config")
		v.AddConfigPath(".")
	}
	
	// 读取环境变量
	v.AutomaticEnv()
	
	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}
	
	// 解析配置
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}
	
	// 验证配置
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}
	
	globalConfig = &cfg
	return &cfg, nil
}

// Get 获取全局配置
func Get() *Config {
	return globalConfig
}

// Validate 验证配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	
	if c.Auth.JWTSecret == "" {
		return fmt.Errorf("jwt secret is required")
	}
	
	return nil
}

// GetDSN 获取数据库 DSN
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.Username, c.Password, c.Database, c.SSLMode)
}

// GetAddr 获取 Redis 地址
func (c *RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr 获取服务器地址
func (c *ServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

