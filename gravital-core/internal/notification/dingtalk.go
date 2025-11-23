package notification

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"
)

// DingTalkSender 钉钉发送器
type DingTalkSender struct {
	config *DingTalkConfig
	client *http.Client
	logger *zap.Logger
}

// NewDingTalkSender 创建钉钉发送器
func NewDingTalkSender(config *DingTalkConfig, logger *zap.Logger) *DingTalkSender {
	return &DingTalkSender{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Name 获取发送器名称
func (s *DingTalkSender) Name() string {
	return "DingTalk"
}

// Validate 验证配置
func (s *DingTalkSender) Validate() error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("DingTalk webhook URL is required")
	}
	return nil
}

// Send 发送钉钉通知
func (s *DingTalkSender) Send(ctx context.Context, notification *Notification) error {
	// 构建消息内容
	message := s.buildMessage(notification)
	
	// 序列化为 JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// 构建 URL（带签名）
	webhookURL := s.buildURL()
	
	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	s.logger.Debug("Sending DingTalk notification",
		zap.String("alert_id", notification.AlertID))
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send DingTalk notification: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	
	// 解析响应
	var result struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}
	
	if result.ErrCode != 0 {
		return fmt.Errorf("DingTalk API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}
	
	s.logger.Debug("DingTalk notification sent successfully",
		zap.String("alert_id", notification.AlertID))
	
	return nil
}

// buildMessage 构建钉钉消息
func (s *DingTalkSender) buildMessage(notification *Notification) map[string]interface{} {
	// 构建 Markdown 内容
	content := fmt.Sprintf("### %s\n\n", notification.Subject)
	content += fmt.Sprintf("**优先级**: %s\n\n", s.getPriorityText(notification.Priority))
	content += fmt.Sprintf("```\n%s\n```\n\n", notification.Content)
	
	// 添加元数据
	if len(notification.Metadata) > 0 {
		content += "**详细信息**:\n\n"
		for key, value := range notification.Metadata {
			content += fmt.Sprintf("- %s: %v\n", key, value)
		}
	}
	
	content += fmt.Sprintf("\n> 发送时间: %s", notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"title": notification.Subject,
			"text":  content,
		},
	}
	
	// 添加 @ 功能
	at := map[string]interface{}{
		"isAtAll": s.config.AtAll,
	}
	
	if len(s.config.AtMobiles) > 0 {
		at["atMobiles"] = s.config.AtMobiles
	}
	
	message["at"] = at
	
	return message
}

// buildURL 构建带签名的 URL
func (s *DingTalkSender) buildURL() string {
	if s.config.Secret == "" {
		return s.config.WebhookURL
	}
	
	// 生成签名
	timestamp := time.Now().UnixMilli()
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, s.config.Secret)
	
	h := hmac.New(sha256.New, []byte(s.config.Secret))
	h.Write([]byte(stringToSign))
	signature := base64.StdEncoding.EncodeToString(h.Sum(nil))
	
	// 构建 URL
	u, _ := url.Parse(s.config.WebhookURL)
	q := u.Query()
	q.Set("timestamp", fmt.Sprintf("%d", timestamp))
	q.Set("sign", signature)
	u.RawQuery = q.Encode()
	
	return u.String()
}

// getPriorityText 获取优先级文本
func (s *DingTalkSender) getPriorityText(priority Priority) string {
	switch priority {
	case PriorityCritical:
		return "🔴 紧急"
	case PriorityHigh:
		return "🟠 高"
	case PriorityNormal:
		return "🟡 普通"
	case PriorityLow:
		return "🟢 低"
	default:
		return "⚪ 未知"
	}
}

