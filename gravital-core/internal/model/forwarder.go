package model

import "time"

// ForwarderConfig 转发器配置
type ForwarderConfig struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Name          string    `gorm:"uniqueIndex;size:255;not null" json:"name"`
	Type          string    `gorm:"size:64;not null" json:"type"`
	Enabled       bool      `gorm:"default:true" json:"enabled"`
	Endpoint      string    `gorm:"type:text" json:"endpoint"`
	AuthConfig    JSONB     `gorm:"type:jsonb" json:"auth_config"`
	BatchSize     int       `json:"batch_size"`
	FlushInterval int       `json:"flush_interval"`
	RetryTimes    int       `json:"retry_times"`
	TimeoutSeconds int      `json:"timeout_seconds"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// ForwarderStats 转发器统计
type ForwarderStats struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	ForwarderName  string    `gorm:"size:255;index" json:"forwarder_name"`
	SuccessCount   int64     `json:"success_count"`
	FailedCount    int64     `json:"failed_count"`
	TotalBytes     int64     `json:"total_bytes"`
	AvgLatencyMs   int       `json:"avg_latency_ms"`
	RecordedAt     time.Time `gorm:"index" json:"recorded_at"`
}

// TableName 指定表名
func (ForwarderConfig) TableName() string {
	return "forwarder_configs"
}

func (ForwarderStats) TableName() string {
	return "forwarder_stats"
}

