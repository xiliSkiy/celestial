package model

import "time"

// Sentinel 采集端模型
type Sentinel struct {
	ID            uint       `gorm:"primaryKey" json:"id"`
	SentinelID    string     `gorm:"uniqueIndex;size:64;not null" json:"sentinel_id"`
	Name          string     `gorm:"size:255" json:"name"`
	Hostname      string     `gorm:"size:255" json:"hostname"`
	IPAddress     string     `gorm:"size:64" json:"ip_address"`
	Version       string     `gorm:"size:32" json:"version"`
	OS            string     `gorm:"size:64" json:"os"`
	Arch          string     `gorm:"size:32" json:"arch"`
	Region        string     `gorm:"size:64;index" json:"region"`
	Labels        JSONB      `gorm:"type:jsonb" json:"labels"`
	APIToken      string     `gorm:"size:255" json:"-"` // 不返回给前端
	Status        string     `gorm:"size:32;index" json:"status"`
	LastHeartbeat *time.Time `json:"last_heartbeat"`
	RegisteredAt  time.Time  `json:"registered_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SentinelHeartbeat 心跳记录
type SentinelHeartbeat struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	SentinelID   string    `gorm:"size:64;not null;index:idx_sentinel_time" json:"sentinel_id"`
	CPUUsage     float64   `json:"cpu_usage"`
	MemoryUsage  float64   `json:"memory_usage"`
	DiskUsage    float64   `json:"disk_usage"`
	TaskCount    int       `json:"task_count"`
	PluginCount  int       `json:"plugin_count"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	ReceivedAt   time.Time `gorm:"index:idx_sentinel_time" json:"received_at"`
}

// TableName 指定表名
func (Sentinel) TableName() string {
	return "sentinels"
}

func (SentinelHeartbeat) TableName() string {
	return "sentinel_heartbeats"
}

