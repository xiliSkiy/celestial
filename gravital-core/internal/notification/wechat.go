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

// WeChatSender 企业微信发送器
type WeChatSender struct {
	config *WeChatConfig
	client *http.Client
	logger *zap.Logger
}

// NewWeChatSender 创建企业微信发送器
func NewWeChatSender(config *WeChatConfig, logger *zap.Logger) *WeChatSender {
	return &WeChatSender{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// Name 获取发送器名称
func (s *WeChatSender) Name() string {
	return "WeChat Work"
}

// Validate 验证配置
func (s *WeChatSender) Validate() error {
	if s.config.WebhookURL == "" {
		return fmt.Errorf("WeChat webhook URL is required")
	}
	return nil
}

// Send 发送企业微信通知
func (s *WeChatSender) Send(ctx context.Context, notification *Notification) error {
	// 构建消息内容
	message := s.buildMessage(notification)
	
	// 序列化为 JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}
	
	// 创建 HTTP 请求
	req, err := http.NewRequestWithContext(ctx, "POST", s.config.WebhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	
	// 发送请求
	s.logger.Debug("Sending WeChat notification",
		zap.String("alert_id", notification.AlertID))
	
	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send WeChat notification: %w", err)
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
		return fmt.Errorf("WeChat API error: %s (code: %d)", result.ErrMsg, result.ErrCode)
	}
	
	s.logger.Debug("WeChat notification sent successfully",
		zap.String("alert_id", notification.AlertID))
	
	return nil
}

// buildMessage 构建企业微信消息
func (s *WeChatSender) buildMessage(notification *Notification) map[string]interface{} {
	// 构建 Markdown 内容
	content := fmt.Sprintf("### %s\n", notification.Subject)
	content += fmt.Sprintf("> **优先级**: <font color=\"%s\">%s</font>\n", 
		s.getPriorityColor(notification.Priority),
		s.getPriorityText(notification.Priority))
	content += "\n---\n\n"
	content += fmt.Sprintf("```\n%s\n```\n", notification.Content)
	
	// 添加元数据
	if len(notification.Metadata) > 0 {
		content += "\n**详细信息**:\n"
		for key, value := range notification.Metadata {
			content += fmt.Sprintf("> %s: %v\n", key, value)
		}
	}
	
	content += fmt.Sprintf("\n<font color=\"comment\">发送时间: %s</font>", 
		notification.CreatedAt.Format("2006-01-02 15:04:05"))
	
	message := map[string]interface{}{
		"msgtype": "markdown",
		"markdown": map[string]interface{}{
			"content": content,
		},
	}
	
	// 添加 @ 功能
	if len(s.config.MentionedList) > 0 || len(s.config.MentionedMobileList) > 0 {
		mentioned := make(map[string]interface{})
		
		if len(s.config.MentionedList) > 0 {
			mentioned["mentioned_list"] = s.config.MentionedList
		}
		
		if len(s.config.MentionedMobileList) > 0 {
			mentioned["mentioned_mobile_list"] = s.config.MentionedMobileList
		}
		
		message["markdown"].(map[string]interface{})["mentioned_list"] = s.config.MentionedList
		message["markdown"].(map[string]interface{})["mentioned_mobile_list"] = s.config.MentionedMobileList
	}
	
	return message
}

// getPriorityText 获取优先级文本
func (s *WeChatSender) getPriorityText(priority Priority) string {
	switch priority {
	case PriorityCritical:
		return "紧急"
	case PriorityHigh:
		return "高"
	case PriorityNormal:
		return "普通"
	case PriorityLow:
		return "低"
	default:
		return "未知"
	}
}

// getPriorityColor 获取优先级颜色
func (s *WeChatSender) getPriorityColor(priority Priority) string {
	switch priority {
	case PriorityCritical:
		return "warning" // 红色
	case PriorityHigh:
		return "warning" // 橙色（企业微信只支持 warning 和 comment）
	case PriorityNormal:
		return "info"    // 蓝色
	case PriorityLow:
		return "comment" // 灰色
	default:
		return "comment"
	}
}

