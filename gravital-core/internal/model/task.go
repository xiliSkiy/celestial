package model

import "time"

// CollectionTask 采集任务
type CollectionTask struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TaskID          string     `gorm:"uniqueIndex;size:64;not null" json:"task_id"`
	DeviceID        string     `gorm:"size:64;not null;index" json:"device_id"`
	Device          *Device    `gorm:"foreignKey:DeviceID;references:DeviceID" json:"device,omitempty"`
	SentinelID      string     `gorm:"size:64;not null;index" json:"sentinel_id"`
	PluginName      string     `gorm:"size:64;not null" json:"plugin_name"`
	Config          JSONB      `gorm:"type:jsonb" json:"config"`
	IntervalSeconds int        `json:"interval_seconds"`
	Enabled         bool       `gorm:"default:true" json:"enabled"`
	Priority        int        `gorm:"default:5" json:"priority"`
	RetryCount      int        `gorm:"default:3" json:"retry_count"`
	TimeoutSeconds  int        `gorm:"default:30" json:"timeout_seconds"`
	LastExecutedAt  *time.Time `json:"last_executed_at"`
	NextExecutionAt *time.Time `gorm:"index" json:"next_execution_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// TaskExecution 任务执行记录
type TaskExecution struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	TaskID           string    `gorm:"size:64;not null;index" json:"task_id"`
	SentinelID       string    `gorm:"size:64" json:"sentinel_id"`
	Status           string    `gorm:"size:32" json:"status"`
	MetricsCollected int       `json:"metrics_collected"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message"`
	ExecutionTimeMs  int       `json:"execution_time_ms"`
	ExecutedAt       time.Time `gorm:"index" json:"executed_at"`
}

// TableName 指定表名
func (CollectionTask) TableName() string {
	return "collection_tasks"
}

func (TaskExecution) TableName() string {
	return "task_executions"
}

