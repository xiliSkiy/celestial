package notification

import "time"

// Channel 通知渠道类型
type Channel string

const (
	ChannelEmail      Channel = "email"       // 邮件
	ChannelWebhook    Channel = "webhook"     // Webhook
	ChannelDingTalk   Channel = "dingtalk"    // 钉钉
	ChannelWeChat     Channel = "wechat"      // 企业微信
	ChannelSMS        Channel = "sms"         // 短信
	ChannelSlack      Channel = "slack"       // Slack
	ChannelTelegram   Channel = "telegram"    // Telegram
)

// Status 通知状态
type Status string

const (
	StatusPending Status = "pending" // 待发送
	StatusSending Status = "sending" // 发送中
	StatusSent    Status = "sent"    // 已发送
	StatusFailed  Status = "failed"  // 发送失败
)

// Priority 通知优先级
type Priority string

const (
	PriorityLow      Priority = "low"      // 低优先级
	PriorityNormal   Priority = "normal"   // 普通优先级
	PriorityHigh     Priority = "high"     // 高优先级
	PriorityCritical Priority = "critical" // 紧急优先级
)

// Notification 通知消息
type Notification struct {
	ID          string                 `json:"id"`
	Channel     Channel                `json:"channel"`
	Recipient   string                 `json:"recipient"`
	Subject     string                 `json:"subject"`
	Content     string                 `json:"content"`
	Priority    Priority               `json:"priority"`
	Metadata    map[string]interface{} `json:"metadata"`
	AlertID     string                 `json:"alert_id,omitempty"`
	AlertRuleID uint                   `json:"alert_rule_id,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled           bool                   `json:"enabled"`
	Channels          []ChannelConfig        `json:"channels"`
	DedupeInterval    int                    `json:"dedupe_interval"`     // 去重间隔（秒）
	EscalationEnabled bool                   `json:"escalation_enabled"`  // 是否启用升级
	EscalationAfter   int                    `json:"escalation_after"`    // 升级时间（秒）
	EscalationChannels []Channel             `json:"escalation_channels"` // 升级通知渠道
	Metadata          map[string]interface{} `json:"metadata,omitempty"`
}

// ChannelConfig 通知渠道配置
type ChannelConfig struct {
	Channel    Channel  `json:"channel"`
	Enabled    bool     `json:"enabled"`
	Recipients []string `json:"recipients"`
	Template   string   `json:"template,omitempty"`
	Config     map[string]interface{} `json:"config,omitempty"`
}

// NotificationResult 通知发送结果
type NotificationResult struct {
	NotificationID string    `json:"notification_id"`
	Channel        Channel   `json:"channel"`
	Recipient      string    `json:"recipient"`
	Status         Status    `json:"status"`
	Error          string    `json:"error,omitempty"`
	SentAt         time.Time `json:"sent_at,omitempty"`
}

// EmailConfig 邮件配置
type EmailConfig struct {
	SMTPHost     string `json:"smtp_host"`
	SMTPPort     int    `json:"smtp_port"`
	SMTPUser     string `json:"smtp_user"`
	SMTPPassword string `json:"smtp_password"`
	From         string `json:"from"`
	UseTLS       bool   `json:"use_tls"`
}

// WebhookConfig Webhook 配置
type WebhookConfig struct {
	URL     string            `json:"url"`
	Method  string            `json:"method"`  // GET, POST, PUT
	Headers map[string]string `json:"headers"`
	Timeout int               `json:"timeout"` // 超时时间（秒）
}

// DingTalkConfig 钉钉配置
type DingTalkConfig struct {
	WebhookURL string `json:"webhook_url"`
	Secret     string `json:"secret,omitempty"` // 签名密钥
	AtMobiles  []string `json:"at_mobiles,omitempty"`
	AtAll      bool   `json:"at_all,omitempty"`
}

// WeChatConfig 企业微信配置
type WeChatConfig struct {
	WebhookURL string   `json:"webhook_url"`
	MentionedList []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}

// SMSConfig 短信配置
type SMSConfig struct {
	Provider  string `json:"provider"` // aliyun, tencent, twilio
	AccessKey string `json:"access_key"`
	SecretKey string `json:"secret_key"`
	SignName  string `json:"sign_name"`
	Template  string `json:"template"`
}

