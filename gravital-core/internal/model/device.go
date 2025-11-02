package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// Device 设备模型
type Device struct {
	ID               uint           `gorm:"primaryKey" json:"id"`
	DeviceID         string         `gorm:"uniqueIndex;size:64;not null" json:"device_id"`
	Name             string         `gorm:"size:255;not null" json:"name"`
	DeviceType       string         `gorm:"size:64;not null;index" json:"device_type"`
	GroupID          *uint          `gorm:"index" json:"group_id"`
	Group            *DeviceGroup   `gorm:"foreignKey:GroupID" json:"group,omitempty"`
	SentinelID       string         `gorm:"size:64;index" json:"sentinel_id"`
	ConnectionConfig JSONB          `gorm:"type:jsonb" json:"connection_config"`
	Labels           JSONB          `gorm:"type:jsonb" json:"labels"`
	Status           string         `gorm:"size:32;default:'unknown';index" json:"status"`
	LastSeen         *time.Time     `json:"last_seen"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

// DeviceGroup 设备分组
type DeviceGroup struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	Name        string          `gorm:"size:255;not null" json:"name"`
	ParentID    *uint           `gorm:"index" json:"parent_id"`
	Parent      *DeviceGroup    `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children    []DeviceGroup   `gorm:"foreignKey:ParentID" json:"children,omitempty"`
	Description string          `gorm:"type:text" json:"description"`
	CreatedAt   time.Time       `json:"created_at"`
}

// DeviceTemplate 设备模板
type DeviceTemplate struct {
	ID               uint      `gorm:"primaryKey" json:"id"`
	Name             string    `gorm:"size:255;not null" json:"name"`
	DeviceType       string    `gorm:"size:64;not null" json:"device_type"`
	ConnectionSchema JSONB     `gorm:"type:jsonb" json:"connection_schema"`
	DefaultConfig    JSONB     `gorm:"type:jsonb" json:"default_config"`
	CreatedAt        time.Time `json:"created_at"`
}

// JSONB 自定义类型，用于处理 PostgreSQL JSONB
type JSONB map[string]interface{}

// Value 实现 driver.Valuer 接口
func (j JSONB) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// Scan 实现 sql.Scanner 接口
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		*j = nil
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	
	result := make(map[string]interface{})
	if err := json.Unmarshal(bytes, &result); err != nil {
		return err
	}
	
	*j = result
	return nil
}

// TableName 指定表名
func (Device) TableName() string {
	return "devices"
}

func (DeviceGroup) TableName() string {
	return "device_groups"
}

func (DeviceTemplate) TableName() string {
	return "device_templates"
}

