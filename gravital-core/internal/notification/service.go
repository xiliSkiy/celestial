package notification

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// Service 通知服务接口
type Service interface {
	// Send 发送通知
	Send(ctx context.Context, notification *Notification) (*NotificationResult, error)
	
	// SendBatch 批量发送通知
	SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error)
	
	// SendAlert 发送告警通知
	SendAlert(ctx context.Context, event *model.AlertEvent, config *NotificationConfig) error
	
	// RegisterChannel 注册通知渠道
	RegisterChannel(channel Channel, sender Sender) error
	
	// GetChannel 获取通知渠道
	GetChannel(channel Channel) (Sender, error)
	
	// ShouldNotify 判断是否应该发送通知（去重检查）
	ShouldNotify(ctx context.Context, alertID string, ruleID uint) (bool, error)
	
	// RecordNotification 记录通知
	RecordNotification(ctx context.Context, notification *model.AlertNotification) error
	
	// GetNotificationHistory 获取通知历史
	GetNotificationHistory(ctx context.Context, alertEventID uint) ([]*model.AlertNotification, error)
}

// Sender 通知发送器接口
type Sender interface {
	// Send 发送通知
	Send(ctx context.Context, notification *Notification) error
	
	// Name 获取发送器名称
	Name() string
	
	// Validate 验证配置
	Validate() error
}

// service 通知服务实现
type service struct {
	db              *gorm.DB
	logger          *zap.Logger
	channels        map[Channel]Sender
	channelsMu      sync.RWMutex
	dedupeCache     map[string]time.Time // alertID -> lastNotifyTime
	dedupeCacheMu   sync.RWMutex
	escalationCache map[string]time.Time // alertID -> firstNotifyTime
	escalationMu    sync.RWMutex
}

// NewService 创建通知服务
func NewService(db *gorm.DB, logger *zap.Logger) Service {
	s := &service{
		db:              db,
		logger:          logger,
		channels:        make(map[Channel]Sender),
		dedupeCache:     make(map[string]time.Time),
		escalationCache: make(map[string]time.Time),
	}
	
	// 启动清理协程
	go s.cleanupCache()
	
	return s
}

// RegisterChannel 注册通知渠道
func (s *service) RegisterChannel(channel Channel, sender Sender) error {
	s.channelsMu.Lock()
	defer s.channelsMu.Unlock()
	
	if err := sender.Validate(); err != nil {
		return fmt.Errorf("invalid sender configuration: %w", err)
	}
	
	s.channels[channel] = sender
	s.logger.Info("Notification channel registered",
		zap.String("channel", string(channel)),
		zap.String("sender", sender.Name()))
	
	return nil
}

// GetChannel 获取通知渠道
func (s *service) GetChannel(channel Channel) (Sender, error) {
	s.channelsMu.RLock()
	defer s.channelsMu.RUnlock()
	
	sender, ok := s.channels[channel]
	if !ok {
		return nil, fmt.Errorf("channel %s not registered", channel)
	}
	
	return sender, nil
}

// Send 发送通知
func (s *service) Send(ctx context.Context, notification *Notification) (*NotificationResult, error) {
	result := &NotificationResult{
		NotificationID: notification.ID,
		Channel:        notification.Channel,
		Recipient:      notification.Recipient,
		Status:         StatusPending,
	}
	
	// 获取发送器
	sender, err := s.GetChannel(notification.Channel)
	if err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
		s.logger.Error("Failed to get notification channel",
			zap.String("channel", string(notification.Channel)),
			zap.Error(err))
		return result, err
	}
	
	// 发送通知
	result.Status = StatusSending
	if err := sender.Send(ctx, notification); err != nil {
		result.Status = StatusFailed
		result.Error = err.Error()
		s.logger.Error("Failed to send notification",
			zap.String("channel", string(notification.Channel)),
			zap.String("recipient", notification.Recipient),
			zap.Error(err))
		return result, err
	}
	
	result.Status = StatusSent
	result.SentAt = time.Now()
	
	s.logger.Info("Notification sent successfully",
		zap.String("channel", string(notification.Channel)),
		zap.String("recipient", notification.Recipient),
		zap.String("alert_id", notification.AlertID))
	
	return result, nil
}

// SendBatch 批量发送通知
func (s *service) SendBatch(ctx context.Context, notifications []*Notification) ([]*NotificationResult, error) {
	results := make([]*NotificationResult, len(notifications))
	var wg sync.WaitGroup
	
	for i, notification := range notifications {
		wg.Add(1)
		go func(idx int, notif *Notification) {
			defer wg.Done()
			result, _ := s.Send(ctx, notif)
			results[idx] = result
		}(i, notification)
	}
	
	wg.Wait()
	return results, nil
}

// SendAlert 发送告警通知
func (s *service) SendAlert(ctx context.Context, event *model.AlertEvent, config *NotificationConfig) error {
	if config == nil || !config.Enabled {
		s.logger.Debug("Notification disabled for alert", zap.String("alert_id", event.AlertID))
		return nil
	}
	
	// 去重检查
	shouldNotify, err := s.ShouldNotify(ctx, event.AlertID, event.RuleID)
	if err != nil {
		return fmt.Errorf("failed to check dedupe: %w", err)
	}
	if !shouldNotify {
		s.logger.Debug("Alert notification skipped due to deduplication",
			zap.String("alert_id", event.AlertID))
		return nil
	}
	
	// 检查是否需要升级通知
	shouldEscalate := s.shouldEscalate(event.AlertID, config)
	
	// 准备通知内容
	subject := fmt.Sprintf("[%s] %s", event.Severity, event.Message)
	content := s.formatAlertContent(event)
	
	// 确定使用的通知渠道
	channels := config.Channels
	if shouldEscalate && config.EscalationEnabled {
		// 使用升级渠道
		s.logger.Info("Alert escalation triggered",
			zap.String("alert_id", event.AlertID),
			zap.Uint("rule_id", event.RuleID))
		
		// 添加升级渠道
		for _, escalationChannel := range config.EscalationChannels {
			found := false
			for _, ch := range channels {
				if ch.Channel == escalationChannel {
					found = true
					break
				}
			}
			if !found {
				channels = append(channels, ChannelConfig{
					Channel: escalationChannel,
					Enabled: true,
				})
			}
		}
	}
	
	// 发送通知到所有启用的渠道
	var notifications []*Notification
	for _, channelConfig := range channels {
		if !channelConfig.Enabled {
			continue
		}
		
		for _, recipient := range channelConfig.Recipients {
			notification := &Notification{
				ID:          fmt.Sprintf("notif-%s-%s-%d", event.AlertID, channelConfig.Channel, time.Now().Unix()),
				Channel:     channelConfig.Channel,
				Recipient:   recipient,
				Subject:     subject,
				Content:     content,
				Priority:    s.severityToPriority(event.Severity),
				AlertID:     event.AlertID,
				AlertRuleID: event.RuleID,
				CreatedAt:   time.Now(),
				Metadata: map[string]interface{}{
					"event_id":   event.ID,
					"device_id":  event.DeviceID,
					"metric_name": event.MetricName,
					"escalated":  shouldEscalate,
				},
			}
			notifications = append(notifications, notification)
		}
	}
	
	// 批量发送
	results, _ := s.SendBatch(ctx, notifications)
	
	// 记录通知结果
	for _, result := range results {
		notificationRecord := &model.AlertNotification{
			AlertEventID: event.ID,
			Channel:      string(result.Channel),
			Recipient:    result.Recipient,
			Status:       string(result.Status),
			SentAt:       &result.SentAt,
			ErrorMessage: result.Error,
		}
		
		if err := s.RecordNotification(ctx, notificationRecord); err != nil {
			s.logger.Error("Failed to record notification",
				zap.String("alert_id", event.AlertID),
				zap.Error(err))
		}
	}
	
	// 更新去重缓存
	s.updateDedupeCache(event.AlertID)
	
	// 更新升级缓存
	if shouldEscalate {
		s.updateEscalationCache(event.AlertID, true)
	} else {
		s.updateEscalationCache(event.AlertID, false)
	}
	
	return nil
}

// ShouldNotify 判断是否应该发送通知（去重检查）
func (s *service) ShouldNotify(ctx context.Context, alertID string, ruleID uint) (bool, error) {
	s.dedupeCacheMu.RLock()
	lastNotifyTime, exists := s.dedupeCache[alertID]
	s.dedupeCacheMu.RUnlock()
	
	if !exists {
		return true, nil
	}
	
	// 从数据库获取规则配置
	var rule model.AlertRule
	if err := s.db.WithContext(ctx).First(&rule, ruleID).Error; err != nil {
		return true, nil // 如果查询失败，默认发送
	}
	
	// 解析通知配置
	dedupeInterval := 300 // 默认 5 分钟
	if rule.NotificationConfig != nil {
		if interval, ok := rule.NotificationConfig["dedupe_interval"].(float64); ok {
			dedupeInterval = int(interval)
		}
	}
	
	// 检查是否超过去重间隔
	elapsed := time.Since(lastNotifyTime).Seconds()
	return elapsed >= float64(dedupeInterval), nil
}

// shouldEscalate 判断是否应该升级通知
func (s *service) shouldEscalate(alertID string, config *NotificationConfig) bool {
	if !config.EscalationEnabled {
		return false
	}
	
	s.escalationMu.RLock()
	firstNotifyTime, exists := s.escalationCache[alertID]
	s.escalationMu.RUnlock()
	
	if !exists {
		return false
	}
	
	elapsed := time.Since(firstNotifyTime).Seconds()
	return elapsed >= float64(config.EscalationAfter)
}

// updateDedupeCache 更新去重缓存
func (s *service) updateDedupeCache(alertID string) {
	s.dedupeCacheMu.Lock()
	defer s.dedupeCacheMu.Unlock()
	s.dedupeCache[alertID] = time.Now()
}

// updateEscalationCache 更新升级缓存
func (s *service) updateEscalationCache(alertID string, reset bool) {
	s.escalationMu.Lock()
	defer s.escalationMu.Unlock()
	
	if reset {
		delete(s.escalationCache, alertID)
	} else {
		if _, exists := s.escalationCache[alertID]; !exists {
			s.escalationCache[alertID] = time.Now()
		}
	}
}

// RecordNotification 记录通知
func (s *service) RecordNotification(ctx context.Context, notification *model.AlertNotification) error {
	return s.db.WithContext(ctx).Create(notification).Error
}

// GetNotificationHistory 获取通知历史
func (s *service) GetNotificationHistory(ctx context.Context, alertEventID uint) ([]*model.AlertNotification, error) {
	var notifications []*model.AlertNotification
	err := s.db.WithContext(ctx).
		Where("alert_event_id = ?", alertEventID).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

// formatAlertContent 格式化告警内容
func (s *service) formatAlertContent(event *model.AlertEvent) string {
	return fmt.Sprintf(`告警详情：
- 告警ID: %s
- 设备ID: %s
- 指标名称: %s
- 严重级别: %s
- 告警消息: %s
- 触发时间: %s
- 当前状态: %s`,
		event.AlertID,
		event.DeviceID,
		event.MetricName,
		event.Severity,
		event.Message,
		event.TriggeredAt.Format("2006-01-02 15:04:05"),
		event.Status)
}

// severityToPriority 将严重级别转换为优先级
func (s *service) severityToPriority(severity string) Priority {
	switch severity {
	case "critical":
		return PriorityCritical
	case "warning":
		return PriorityHigh
	case "info":
		return PriorityNormal
	default:
		return PriorityNormal
	}
}

// cleanupCache 定期清理缓存
func (s *service) cleanupCache() {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	
	for range ticker.C {
		now := time.Now()
		
		// 清理去重缓存（保留 24 小时内的）
		s.dedupeCacheMu.Lock()
		for alertID, lastTime := range s.dedupeCache {
			if now.Sub(lastTime) > 24*time.Hour {
				delete(s.dedupeCache, alertID)
			}
		}
		s.dedupeCacheMu.Unlock()
		
		// 清理升级缓存（保留 24 小时内的）
		s.escalationMu.Lock()
		for alertID, firstTime := range s.escalationCache {
			if now.Sub(firstTime) > 24*time.Hour {
				delete(s.escalationCache, alertID)
			}
		}
		s.escalationMu.Unlock()
		
		s.logger.Debug("Notification cache cleaned up")
	}
}

