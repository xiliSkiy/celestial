package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// WebhookSender Webhook 发送器
type WebhookSender struct {
	config *WebhookConfig
	client *http.Client
	logger *zap.Logger
}

// NewWebhookSender 创建 Webhook 发送器
func NewWebhookSender(config *WebhookConfig, logger *zap.Logger) *WebhookSender {
	timeout := 30 * time.Second
	if config.Timeout > 0 {
		timeout = time.Duration(config.Timeout) * time.Second
	}
	
	return &WebhookSender{
		config: config,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

// Name 获取发送器名称
func (s *WebhookSender) Name() string {
	return "Webhook"
}

// Validate 验证配置
func (s *WebhookSender) Validate() error {
	if s.config.URL == "" {
		return fmt.Errorf("webhook URL is required")
	}
	
	method := s.config.Method
	if method == "" {
		method = "POST"
	}
	if method != "GET" && method != "POST" && method != "PUT" {
		return fmt.Errorf("invalid HTTP method: %s", method)
	}
	
	return nil
}

// Send 发送 Webhook
func (s *WebhookSender) Send(ctx context.Context, notification *Notification) error {
	// 构建请求体
	payload := map[string]interface{}{
		"id":         notification.ID,
		"channel":    notification.Channel,
		"recipient":  notification.Recipient,
		"subject":    notification.Subject,
		"content":    notification.Content,
		"priority":   notification.Priority,
		"alert_id":   notification.AlertID,
		"metadata":   notification.Metadata,
		"created_at": notification.CreatedAt.Format(time.RFC3339),
	}
	
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	
	// 创建 HTTP 请求
	method := s.config.Method
	if method == "" {
		method = "POST"
	}
	
	req, err := http.NewRequestWithContext(ctx, method, s.config.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Celestial-Alert-System/1.0")
	
	// 添加自定义请求头
	for key, value := range s.config.Headers {
		req.Header.Set(key, value)
	}
	
	// 发送请求
	s.logger.Debug("Sending webhook",
		zap.String("url", s.config.URL),
		zap.String("method", method),
		zap.String("alert_id", notification.AlertID))
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, _ := io.ReadAll(resp.Body)
	
	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d: %s", resp.StatusCode, string(body))
	}
	
	s.logger.Debug("Webhook sent successfully",
		zap.String("url", s.config.URL),
		zap.Int("status_code", resp.StatusCode))
	
	return nil
}

