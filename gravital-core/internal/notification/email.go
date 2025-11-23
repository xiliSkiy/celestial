package notification

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"go.uber.org/zap"
)

// EmailSender 邮件发送器
type EmailSender struct {
	config *EmailConfig
	logger *zap.Logger
}

// NewEmailSender 创建邮件发送器
func NewEmailSender(config *EmailConfig, logger *zap.Logger) *EmailSender {
	return &EmailSender{
		config: config,
		logger: logger,
	}
}

// Name 获取发送器名称
func (s *EmailSender) Name() string {
	return "Email"
}

// Validate 验证配置
func (s *EmailSender) Validate() error {
	if s.config.SMTPHost == "" {
		return fmt.Errorf("SMTP host is required")
	}
	if s.config.SMTPPort == 0 {
		return fmt.Errorf("SMTP port is required")
	}
	if s.config.From == "" {
		return fmt.Errorf("from address is required")
	}
	return nil
}

// Send 发送邮件
func (s *EmailSender) Send(ctx context.Context, notification *Notification) error {
	// 构建邮件内容
	message := s.buildMessage(notification)
	
	// SMTP 认证
	auth := smtp.PlainAuth("", s.config.SMTPUser, s.config.SMTPPassword, s.config.SMTPHost)
	
	// 发送邮件
	addr := fmt.Sprintf("%s:%d", s.config.SMTPHost, s.config.SMTPPort)
	
	if s.config.UseTLS {
		// 使用 TLS
		return s.sendWithTLS(addr, auth, notification.Recipient, message)
	}
	
	// 不使用 TLS
	return smtp.SendMail(addr, auth, s.config.From, []string{notification.Recipient}, []byte(message))
}

// sendWithTLS 使用 TLS 发送邮件
func (s *EmailSender) sendWithTLS(addr string, auth smtp.Auth, to string, message string) error {
	// 创建 TLS 配置
	tlsConfig := &tls.Config{
		ServerName: s.config.SMTPHost,
	}
	
	// 连接到 SMTP 服务器
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()
	
	// 创建 SMTP 客户端
	client, err := smtp.NewClient(conn, s.config.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()
	
	// 认证
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("SMTP authentication failed: %w", err)
	}
	
	// 设置发件人
	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}
	
	// 设置收件人
	if err := client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}
	
	// 发送邮件内容
	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}
	
	if _, err := w.Write([]byte(message)); err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}
	
	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	
	return nil
}

// buildMessage 构建邮件消息
func (s *EmailSender) buildMessage(notification *Notification) string {
	var builder strings.Builder
	
	// 邮件头
	builder.WriteString(fmt.Sprintf("From: %s\r\n", s.config.From))
	builder.WriteString(fmt.Sprintf("To: %s\r\n", notification.Recipient))
	builder.WriteString(fmt.Sprintf("Subject: %s\r\n", notification.Subject))
	builder.WriteString("MIME-Version: 1.0\r\n")
	builder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	builder.WriteString("\r\n")
	
	// 邮件正文（HTML 格式）
	builder.WriteString("<html><body>")
	builder.WriteString("<div style='font-family: Arial, sans-serif;'>")
	
	// 优先级标签
	priorityColor := s.getPriorityColor(notification.Priority)
	builder.WriteString(fmt.Sprintf("<div style='background-color: %s; color: white; padding: 10px; border-radius: 5px; margin-bottom: 20px;'>", priorityColor))
	builder.WriteString(fmt.Sprintf("<h2 style='margin: 0;'>%s</h2>", notification.Subject))
	builder.WriteString("</div>")
	
	// 内容
	builder.WriteString("<div style='padding: 20px; background-color: #f5f5f5; border-radius: 5px;'>")
	builder.WriteString("<pre style='white-space: pre-wrap; font-family: monospace;'>")
	builder.WriteString(notification.Content)
	builder.WriteString("</pre>")
	builder.WriteString("</div>")
	
	// 元数据
	if len(notification.Metadata) > 0 {
		builder.WriteString("<div style='margin-top: 20px; padding: 10px; background-color: #e8e8e8; border-radius: 5px;'>")
		builder.WriteString("<h4>详细信息：</h4>")
		builder.WriteString("<ul>")
		for key, value := range notification.Metadata {
			builder.WriteString(fmt.Sprintf("<li><strong>%s:</strong> %v</li>", key, value))
		}
		builder.WriteString("</ul>")
		builder.WriteString("</div>")
	}
	
	// 页脚
	builder.WriteString("<div style='margin-top: 30px; padding-top: 20px; border-top: 1px solid #ccc; color: #888; font-size: 12px;'>")
	builder.WriteString("<p>此邮件由 Celestial 监控系统自动发送，请勿回复。</p>")
	builder.WriteString(fmt.Sprintf("<p>发送时间: %s</p>", notification.CreatedAt.Format("2006-01-02 15:04:05")))
	builder.WriteString("</div>")
	
	builder.WriteString("</div>")
	builder.WriteString("</body></html>")
	
	return builder.String()
}

// getPriorityColor 获取优先级颜色
func (s *EmailSender) getPriorityColor(priority Priority) string {
	switch priority {
	case PriorityCritical:
		return "#d32f2f" // 红色
	case PriorityHigh:
		return "#f57c00" // 橙色
	case PriorityNormal:
		return "#1976d2" // 蓝色
	case PriorityLow:
		return "#388e3c" // 绿色
	default:
		return "#757575" // 灰色
	}
}

