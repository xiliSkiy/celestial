package engine

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/notification"
	"github.com/celestial/gravital-core/internal/repository"
)

// AlertEngine 告警引擎
type AlertEngine struct {
	db               *gorm.DB
	logger           *zap.Logger
	alertRepo        repository.AlertRepository
	vmClient         *VMClient
	notificationSvc  notification.Service
	checkInterval    time.Duration
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	activeAlerts     map[uint]map[string]*ActiveAlert // rule_id -> device_id -> alert
	activeAlertsMu   sync.RWMutex
}

// ActiveAlert 活跃的告警
type ActiveAlert struct {
	RuleID       uint
	DeviceID     string
	EventID      uint
	FirstFiredAt time.Time
	LastFiredAt  time.Time
}

// Config 引擎配置
type Config struct {
	VMURL            string
	CheckInterval    time.Duration
	NotificationSvc  notification.Service
}

// NewAlertEngine 创建告警引擎
func NewAlertEngine(db *gorm.DB, logger *zap.Logger, cfg *Config) *AlertEngine {
	ctx, cancel := context.WithCancel(context.Background())
	
	// 创建 VictoriaMetrics 客户端
	vmClient := NewVMClient(cfg.VMURL, logger)
	
	// 检查 VictoriaMetrics 健康状态
	if cfg.VMURL != "" {
		if err := vmClient.Health(); err != nil {
			logger.Warn("VictoriaMetrics health check failed, will retry during queries",
				zap.String("url", cfg.VMURL),
				zap.Error(err))
		} else {
			logger.Info("VictoriaMetrics connection established",
				zap.String("url", cfg.VMURL))
		}
	} else {
		logger.Warn("VictoriaMetrics URL not configured, alert engine will use fallback mode")
	}
	
	return &AlertEngine{
		db:              db,
		logger:          logger,
		alertRepo:       repository.NewAlertRepository(db),
		vmClient:        vmClient,
		notificationSvc: cfg.NotificationSvc,
		checkInterval:   cfg.CheckInterval,
		ctx:             ctx,
		cancel:          cancel,
		activeAlerts:    make(map[uint]map[string]*ActiveAlert),
	}
}

// Start 启动告警引擎
func (e *AlertEngine) Start() {
	e.logger.Info("Starting alert engine",
		zap.Duration("check_interval", e.checkInterval))

	e.wg.Add(1)
	go e.evaluationLoop()
}

// Stop 停止告警引擎
func (e *AlertEngine) Stop() {
	e.logger.Info("Stopping alert engine...")
	e.cancel()
	e.wg.Wait()
	e.logger.Info("Alert engine stopped")
}

// evaluationLoop 评估循环
func (e *AlertEngine) evaluationLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.checkInterval)
	defer ticker.Stop()

	// 立即执行一次
	e.evaluateAllRules()

	for {
		select {
		case <-ticker.C:
			e.evaluateAllRules()
		case <-e.ctx.Done():
			return
		}
	}
}

// evaluateAllRules 评估所有规则
func (e *AlertEngine) evaluateAllRules() {
	// 获取所有启用的规则
	rules, _, err := e.alertRepo.ListRules(e.ctx, &repository.AlertRuleFilter{
		Enabled:  boolPtr(true),
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		e.logger.Error("Failed to list alert rules", zap.Error(err))
		return
	}

	if len(rules) == 0 {
		return
	}

	e.logger.Debug("Evaluating alert rules", zap.Int("count", len(rules)))

	// 并发评估所有规则
	var wg sync.WaitGroup
	for _, rule := range rules {
		wg.Add(1)
		go func(r *model.AlertRule) {
			defer wg.Done()
			e.evaluateRule(r)
		}(rule)
	}
	wg.Wait()
}

// evaluateRule 评估单个规则
func (e *AlertEngine) evaluateRule(rule *model.AlertRule) {
	// 解析条件：metric_name operator threshold
	// 例如：device_status != 0
	parts := strings.Fields(rule.Condition)
	if len(parts) != 3 {
		e.logger.Error("Invalid condition format",
			zap.String("rule", rule.RuleName),
			zap.String("condition", rule.Condition))
		return
	}

	metricName := parts[0]
	operator := parts[1]
	thresholdStr := parts[2]

	threshold, err := strconv.ParseFloat(thresholdStr, 64)
	if err != nil {
		e.logger.Error("Invalid threshold value",
			zap.String("rule", rule.RuleName),
			zap.String("threshold", thresholdStr),
			zap.Error(err))
		return
	}

	// 查询指标数据
	// 构造 PromQL 查询
	query := metricName
	if rule.Filters != nil && len(rule.Filters) > 0 {
		// 添加过滤条件
		var filters []string
		for k, v := range rule.Filters {
			if strVal, ok := v.(string); ok {
				filters = append(filters, fmt.Sprintf(`%s="%s"`, k, strVal))
			}
		}
		if len(filters) > 0 {
			query = fmt.Sprintf("%s{%s}", metricName, strings.Join(filters, ","))
		}
	}

	// 查询当前值
	results, err := e.queryMetric(query)
	if err != nil {
		e.logger.Error("Failed to query metric",
			zap.String("rule", rule.RuleName),
			zap.String("query", query),
			zap.Error(err))
		return
	}

	// 评估每个时间序列
	for _, result := range results {
		deviceID := result.Labels["device_id"]
		if deviceID == "" {
			continue
		}

		// 检查是否满足告警条件
		if e.checkCondition(result.Value, operator, threshold) {
			// 满足条件，触发告警
			e.triggerAlert(rule, deviceID, metricName, result.Value, threshold, operator)
		} else {
			// 不满足条件，解决告警
			e.resolveAlert(rule, deviceID)
		}
	}
}

// queryMetric 查询指标数据
func (e *AlertEngine) queryMetric(query string) ([]MetricResult, error) {
	// 优先使用 VictoriaMetrics 查询
	if e.vmClient != nil && e.vmClient.baseURL != "" {
		results, err := e.vmClient.Query(query)
		if err != nil {
			e.logger.Warn("Failed to query VictoriaMetrics, falling back to database",
				zap.String("query", query),
				zap.Error(err))
			// 如果 VM 查询失败，回退到数据库查询
			return e.queryMetricFromDB(query)
		}
		
		// 如果 VM 返回空结果，也尝试从数据库查询（兼容性）
		if len(results) == 0 {
			e.logger.Debug("VictoriaMetrics returned no results, trying database fallback",
				zap.String("query", query))
			return e.queryMetricFromDB(query)
		}
		
		return results, nil
	}
	
	// 如果没有配置 VictoriaMetrics，使用数据库查询
	e.logger.Debug("VictoriaMetrics not configured, using database fallback",
		zap.String("query", query))
	return e.queryMetricFromDB(query)
}

// queryMetricFromDB 从数据库查询指标（回退方案）
func (e *AlertEngine) queryMetricFromDB(query string) ([]MetricResult, error) {
	// 解析查询
	metricName := query
	filters := make(map[string]string)
	
	if strings.Contains(query, "{") {
		parts := strings.SplitN(query, "{", 2)
		metricName = parts[0]
		filterStr := strings.TrimSuffix(parts[1], "}")
		
		// 解析过滤条件
		for _, filter := range strings.Split(filterStr, ",") {
			kv := strings.SplitN(filter, "=", 2)
			if len(kv) == 2 {
				key := strings.TrimSpace(kv[0])
				value := strings.Trim(strings.TrimSpace(kv[1]), `"`)
				filters[key] = value
			}
		}
	}

	// 对于 device_status 指标，从数据库查询
	if metricName == "device_status" {
		var devices []model.Device
		dbQuery := e.db.Model(&model.Device{})
		
		// 应用过滤条件
		if deviceID, ok := filters["device_id"]; ok {
			dbQuery = dbQuery.Where("device_id = ?", deviceID)
		}
		if deviceType, ok := filters["device_type"]; ok {
			dbQuery = dbQuery.Where("device_type = ?", deviceType)
		}
		
		if err := dbQuery.Find(&devices).Error; err != nil {
			return nil, err
		}

		results := make([]MetricResult, 0, len(devices))
		for _, device := range devices {
			value := 0.0
			if device.Status == "online" {
				value = 1.0
			}
			
			results = append(results, MetricResult{
				Labels: map[string]string{
					"device_id":   device.DeviceID,
					"device_type": device.DeviceType,
				},
				Value: value,
			})
		}
		
		e.logger.Debug("Queried metric from database",
			zap.String("metric", metricName),
			zap.Int("result_count", len(results)))
		
		return results, nil
	}

	// 其他指标不支持数据库查询
	return nil, fmt.Errorf("unsupported metric for database fallback: %s", metricName)
}

// checkCondition 检查条件
func (e *AlertEngine) checkCondition(value float64, operator string, threshold float64) bool {
	switch operator {
	case ">":
		return value > threshold
	case ">=":
		return value >= threshold
	case "<":
		return value < threshold
	case "<=":
		return value <= threshold
	case "==":
		return value == threshold
	case "!=":
		return value != threshold
	default:
		return false
	}
}

// triggerAlert 触发告警
func (e *AlertEngine) triggerAlert(rule *model.AlertRule, deviceID, metricName string, currentValue, threshold float64, operator string) {
	e.activeAlertsMu.Lock()
	defer e.activeAlertsMu.Unlock()

	// 检查是否已经有活跃的告警
	if deviceAlerts, ok := e.activeAlerts[rule.ID]; ok {
		if alert, exists := deviceAlerts[deviceID]; exists {
			// 已经有活跃告警，更新最后触发时间
			alert.LastFiredAt = time.Now()
			return
		}
	}

	// 创建新的告警事件
	alertID := fmt.Sprintf("alert-%s-%s-%d", rule.RuleName, deviceID, time.Now().Unix())
	message := fmt.Sprintf("%s: 当前值 %.2f %s 阈值 %.2f", rule.RuleName, currentValue, operator, threshold)

	event := &model.AlertEvent{
		AlertID:     alertID,
		RuleID:      rule.ID,
		DeviceID:    deviceID,
		MetricName:  metricName,
		Severity:    rule.Severity,
		Message:     message,
		Labels:      make(map[string]interface{}),
		TriggeredAt: time.Now(),
		Status:      "firing",
	}

	if err := e.alertRepo.CreateEvent(e.ctx, event); err != nil {
		e.logger.Error("Failed to create alert event",
			zap.String("rule", rule.RuleName),
			zap.String("device_id", deviceID),
			zap.Error(err))
		return
	}

	// 记录活跃告警
	if e.activeAlerts[rule.ID] == nil {
		e.activeAlerts[rule.ID] = make(map[string]*ActiveAlert)
	}
	e.activeAlerts[rule.ID][deviceID] = &ActiveAlert{
		RuleID:       rule.ID,
		DeviceID:     deviceID,
		EventID:      event.ID,
		FirstFiredAt: time.Now(),
		LastFiredAt:  time.Now(),
	}

	e.logger.Info("Alert triggered",
		zap.String("rule", rule.RuleName),
		zap.String("device_id", deviceID),
		zap.Float64("value", currentValue),
		zap.Float64("threshold", threshold))
	
	// 发送通知
	if e.notificationSvc != nil && rule.NotificationConfig != nil {
		go func() {
			config := e.parseNotificationConfig(rule.NotificationConfig)
			if err := e.notificationSvc.SendAlert(context.Background(), event, config); err != nil {
				e.logger.Error("Failed to send alert notification",
					zap.String("alert_id", event.AlertID),
					zap.Error(err))
			}
		}()
	}
}

// parseNotificationConfig 解析通知配置
func (e *AlertEngine) parseNotificationConfig(config map[string]interface{}) *notification.NotificationConfig {
	notifConfig := &notification.NotificationConfig{
		Enabled:           true,
		DedupeInterval:    300,  // 默认 5 分钟
		EscalationEnabled: false,
		EscalationAfter:   1800, // 默认 30 分钟
	}
	
	if enabled, ok := config["enabled"].(bool); ok {
		notifConfig.Enabled = enabled
	}
	
	if interval, ok := config["dedupe_interval"].(float64); ok {
		notifConfig.DedupeInterval = int(interval)
	}
	
	if escalationEnabled, ok := config["escalation_enabled"].(bool); ok {
		notifConfig.EscalationEnabled = escalationEnabled
	}
	
	if escalationAfter, ok := config["escalation_after"].(float64); ok {
		notifConfig.EscalationAfter = int(escalationAfter)
	}
	
	// 解析通知渠道
	if channels, ok := config["channels"].([]interface{}); ok {
		for _, ch := range channels {
			if channelMap, ok := ch.(map[string]interface{}); ok {
				channelConfig := notification.ChannelConfig{}
				
				if channel, ok := channelMap["channel"].(string); ok {
					channelConfig.Channel = notification.Channel(channel)
				}
				
				if enabled, ok := channelMap["enabled"].(bool); ok {
					channelConfig.Enabled = enabled
				}
				
				if recipients, ok := channelMap["recipients"].([]interface{}); ok {
					for _, recipient := range recipients {
						if recipientStr, ok := recipient.(string); ok {
							channelConfig.Recipients = append(channelConfig.Recipients, recipientStr)
						}
					}
				}
				
				notifConfig.Channels = append(notifConfig.Channels, channelConfig)
			}
		}
	}
	
	// 解析升级渠道
	if escalationChannels, ok := config["escalation_channels"].([]interface{}); ok {
		for _, ch := range escalationChannels {
			if channelStr, ok := ch.(string); ok {
				notifConfig.EscalationChannels = append(notifConfig.EscalationChannels, notification.Channel(channelStr))
			}
		}
	}
	
	return notifConfig
}

// resolveAlert 解决告警
func (e *AlertEngine) resolveAlert(rule *model.AlertRule, deviceID string) {
	e.activeAlertsMu.Lock()
	defer e.activeAlertsMu.Unlock()

	// 检查是否有活跃的告警
	deviceAlerts, ok := e.activeAlerts[rule.ID]
	if !ok {
		return
	}

	alert, exists := deviceAlerts[deviceID]
	if !exists {
		return
	}

	// 更新告警事件状态为已解决
	now := time.Now()
	if err := e.db.Model(&model.AlertEvent{}).
		Where("id = ?", alert.EventID).
		Updates(map[string]interface{}{
			"status":      "resolved",
			"resolved_at": now,
		}).Error; err != nil {
		e.logger.Error("Failed to resolve alert event",
			zap.String("rule", rule.RuleName),
			zap.String("device_id", deviceID),
			zap.Error(err))
		return
	}

	// 从活跃告警中移除
	delete(deviceAlerts, deviceID)
	if len(deviceAlerts) == 0 {
		delete(e.activeAlerts, rule.ID)
	}

	e.logger.Info("Alert resolved",
		zap.String("rule", rule.RuleName),
		zap.String("device_id", deviceID))
}

func boolPtr(b bool) *bool {
	return &b
}

