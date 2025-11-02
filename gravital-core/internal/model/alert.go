package model

import "time"

// AlertRule 告警规则
type AlertRule struct {
	ID                 uint      `gorm:"primaryKey" json:"id"`
	RuleName           string    `gorm:"size:255;not null" json:"rule_name"`
	Enabled            bool      `gorm:"default:true" json:"enabled"`
	Severity           string    `gorm:"size:32" json:"severity"`
	Condition          string    `gorm:"type:text;not null" json:"condition"`
	Filters            JSONB     `gorm:"type:jsonb" json:"filters"`
	Duration           int       `json:"duration"`
	NotificationConfig JSONB     `gorm:"type:jsonb" json:"notification_config"`
	InhibitRules       JSONB     `gorm:"type:jsonb" json:"inhibit_rules"`
	MutePeriods        JSONB     `gorm:"type:jsonb" json:"mute_periods"`
	Description        string    `gorm:"type:text" json:"description"`
	CreatedBy          *uint     `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AlertEvent 告警事件
type AlertEvent struct {
	ID               uint       `gorm:"primaryKey" json:"id"`
	AlertID          string     `gorm:"uniqueIndex;size:64;not null" json:"alert_id"`
	RuleID           uint       `gorm:"index" json:"rule_id"`
	Rule             *AlertRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
	DeviceID         string     `gorm:"size:64;index" json:"device_id"`
	MetricName       string     `gorm:"size:255" json:"metric_name"`
	Severity         string     `gorm:"size:32" json:"severity"`
	Message          string     `gorm:"type:text" json:"message"`
	Labels           JSONB      `gorm:"type:jsonb" json:"labels"`
	TriggeredAt      time.Time  `gorm:"index" json:"triggered_at"`
	ResolvedAt       *time.Time `json:"resolved_at"`
	Status           string     `gorm:"size:32;index" json:"status"`
	NotificationSent bool       `gorm:"default:false" json:"notification_sent"`
	Acknowledged     bool       `gorm:"default:false" json:"acknowledged"`
	AcknowledgedBy   *uint      `json:"acknowledged_by"`
	AcknowledgedAt   *time.Time `json:"acknowledged_at"`
	Comment          string     `gorm:"type:text" json:"comment"`
	CreatedAt        time.Time  `json:"created_at"`
}

// AlertNotification 告警通知记录
type AlertNotification struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	AlertEventID uint       `gorm:"index" json:"alert_event_id"`
	Channel      string     `gorm:"size:64" json:"channel"`
	Recipient    string     `gorm:"size:255" json:"recipient"`
	Status       string     `gorm:"size:32" json:"status"`
	SentAt       *time.Time `json:"sent_at"`
	ErrorMessage string     `gorm:"type:text" json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

// TableName 指定表名
func (AlertRule) TableName() string {
	return "alert_rules"
}

func (AlertEvent) TableName() string {
	return "alert_events"
}

func (AlertNotification) TableName() string {
	return "alert_notifications"
}

