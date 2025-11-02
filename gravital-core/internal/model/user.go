package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// User 用户模型
type User struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	Username     string     `gorm:"uniqueIndex;size:64;not null" json:"username"`
	Email        string     `gorm:"uniqueIndex;size:255" json:"email"`
	PasswordHash string     `gorm:"size:255" json:"-"` // 不返回给前端
	RoleID       uint       `json:"role_id"`
	Role         *Role      `gorm:"foreignKey:RoleID" json:"role,omitempty"`
	Enabled      bool       `gorm:"default:true" json:"enabled"`
	LastLogin    *time.Time `json:"last_login"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// Role 角色模型
type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"uniqueIndex;size:64;not null" json:"name"`
	Permissions StringArray `gorm:"type:jsonb" json:"permissions"`
	Description string    `gorm:"type:text" json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// APIToken API Token 模型
type APIToken struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	Token       string     `gorm:"uniqueIndex;size:255;not null" json:"token"`
	Name        string     `gorm:"size:255" json:"name"`
	UserID      *uint      `json:"user_id"`
	User        *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	SentinelID  string     `gorm:"size:64" json:"sentinel_id"`
	Permissions StringArray `gorm:"type:jsonb" json:"permissions"`
	ExpiresAt   *time.Time `json:"expires_at"`
	LastUsed    *time.Time `json:"last_used"`
	CreatedAt   time.Time  `json:"created_at"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

func (Role) TableName() string {
	return "roles"
}

func (APIToken) TableName() string {
	return "api_tokens"
}

// StringArray 字符串数组类型，用于存储 JSON 数组
type StringArray []string

// Scan 实现 sql.Scanner 接口
func (s *StringArray) Scan(value interface{}) error {
	if value == nil {
		*s = []string{}
		return nil
	}
	
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to unmarshal JSONB value")
	}
	
	return json.Unmarshal(bytes, s)
}

// Value 实现 driver.Valuer 接口
func (s StringArray) Value() (driver.Value, error) {
	if len(s) == 0 {
		return "[]", nil
	}
	return json.Marshal(s)
}

