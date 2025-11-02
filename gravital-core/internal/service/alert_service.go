package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
)

// AlertService 告警服务接口
type AlertService interface {
	// 告警规则
	CreateRule(ctx context.Context, req *CreateAlertRuleRequest) (*model.AlertRule, error)
	GetRule(ctx context.Context, id uint) (*model.AlertRule, error)
	UpdateRule(ctx context.Context, id uint, req *UpdateAlertRuleRequest) error
	DeleteRule(ctx context.Context, id uint) error
	ListRules(ctx context.Context, req *ListAlertRuleRequest) ([]*model.AlertRule, int64, error)
	ToggleRule(ctx context.Context, id uint, enabled bool) error

	// 告警事件
	GetEvent(ctx context.Context, id uint) (*model.AlertEvent, error)
	ListEvents(ctx context.Context, req *ListAlertEventRequest) ([]*model.AlertEvent, int64, error)
	AcknowledgeEvent(ctx context.Context, id uint, userID uint, comment string) error
	ResolveEvent(ctx context.Context, id uint, comment string) error
	SilenceEvent(ctx context.Context, id uint, duration time.Duration, comment string) error
	GetStats(ctx context.Context) (map[string]interface{}, error)
}

// CreateAlertRuleRequest 创建告警规则请求
type CreateAlertRuleRequest struct {
	RuleName           string                 `json:"rule_name" binding:"required"`
	Enabled            bool                   `json:"enabled"`
	Severity           string                 `json:"severity" binding:"required"`
	Condition          string                 `json:"condition" binding:"required"`
	Filters            map[string]interface{} `json:"filters"`
	Duration           int                    `json:"duration"`
	NotificationConfig map[string]interface{} `json:"notification_config"`
	Description        string                 `json:"description"`
}

// UpdateAlertRuleRequest 更新告警规则请求
type UpdateAlertRuleRequest struct {
	RuleName           string                 `json:"rule_name"`
	Enabled            *bool                  `json:"enabled"`
	Severity           string                 `json:"severity"`
	Condition          string                 `json:"condition"`
	Filters            map[string]interface{} `json:"filters"`
	Duration           int                    `json:"duration"`
	NotificationConfig map[string]interface{} `json:"notification_config"`
	Description        string                 `json:"description"`
}

// ListAlertRuleRequest 告警规则列表请求
type ListAlertRuleRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Enabled  *bool  `form:"enabled"`
	Severity string `form:"severity"`
	Keyword  string `form:"keyword"`
}

// ListAlertEventRequest 告警事件列表请求
type ListAlertEventRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Severity string `form:"severity"`
	DeviceID string `form:"device_id"`
	RuleID   *uint  `form:"rule_id"`
}

type alertService struct {
	alertRepo repository.AlertRepository
}

// NewAlertService 创建告警服务
func NewAlertService(alertRepo repository.AlertRepository) AlertService {
	return &alertService{
		alertRepo: alertRepo,
	}
}

func (s *alertService) CreateRule(ctx context.Context, req *CreateAlertRuleRequest) (*model.AlertRule, error) {
	rule := &model.AlertRule{
		RuleName:           req.RuleName,
		Enabled:            req.Enabled,
		Severity:           req.Severity,
		Condition:          req.Condition,
		Filters:            req.Filters,
		Duration:           req.Duration,
		NotificationConfig: req.NotificationConfig,
		Description:        req.Description,
	}

	if err := s.alertRepo.CreateRule(ctx, rule); err != nil {
		return nil, fmt.Errorf("failed to create rule: %w", err)
	}

	return rule, nil
}

func (s *alertService) GetRule(ctx context.Context, id uint) (*model.AlertRule, error) {
	rule, err := s.alertRepo.GetRuleByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("rule not found")
		}
		return nil, fmt.Errorf("failed to get rule: %w", err)
	}
	return rule, nil
}

func (s *alertService) UpdateRule(ctx context.Context, id uint, req *UpdateAlertRuleRequest) error {
	rule, err := s.alertRepo.GetRuleByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rule not found")
	}

	// 更新字段
	if req.RuleName != "" {
		rule.RuleName = req.RuleName
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Severity != "" {
		rule.Severity = req.Severity
	}
	if req.Condition != "" {
		rule.Condition = req.Condition
	}
	if req.Filters != nil {
		rule.Filters = req.Filters
	}
	if req.Duration > 0 {
		rule.Duration = req.Duration
	}
	if req.NotificationConfig != nil {
		rule.NotificationConfig = req.NotificationConfig
	}
	if req.Description != "" {
		rule.Description = req.Description
	}

	return s.alertRepo.UpdateRule(ctx, rule)
}

func (s *alertService) DeleteRule(ctx context.Context, id uint) error {
	return s.alertRepo.DeleteRule(ctx, id)
}

func (s *alertService) ListRules(ctx context.Context, req *ListAlertRuleRequest) ([]*model.AlertRule, int64, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filter := &repository.AlertRuleFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Enabled:  req.Enabled,
		Severity: req.Severity,
		Keyword:  req.Keyword,
	}

	return s.alertRepo.ListRules(ctx, filter)
}

func (s *alertService) ToggleRule(ctx context.Context, id uint, enabled bool) error {
	rule, err := s.alertRepo.GetRuleByID(ctx, id)
	if err != nil {
		return fmt.Errorf("rule not found")
	}

	rule.Enabled = enabled
	return s.alertRepo.UpdateRule(ctx, rule)
}

func (s *alertService) GetEvent(ctx context.Context, id uint) (*model.AlertEvent, error) {
	event, err := s.alertRepo.GetEventByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to get event: %w", err)
	}
	return event, nil
}

func (s *alertService) ListEvents(ctx context.Context, req *ListAlertEventRequest) ([]*model.AlertEvent, int64, error) {
	// 设置默认值
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	filter := &repository.AlertEventFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Severity: req.Severity,
		DeviceID: req.DeviceID,
		RuleID:   req.RuleID,
	}

	return s.alertRepo.ListEvents(ctx, filter)
}

func (s *alertService) AcknowledgeEvent(ctx context.Context, id uint, userID uint, comment string) error {
	event, err := s.alertRepo.GetEventByID(ctx, id)
	if err != nil {
		return fmt.Errorf("event not found")
	}

	now := time.Now()
	event.Acknowledged = true
	event.AcknowledgedBy = &userID
	event.AcknowledgedAt = &now
	event.Comment = comment

	return s.alertRepo.UpdateEvent(ctx, event)
}

func (s *alertService) ResolveEvent(ctx context.Context, id uint, comment string) error {
	event, err := s.alertRepo.GetEventByID(ctx, id)
	if err != nil {
		return fmt.Errorf("event not found")
	}

	now := time.Now()
	event.Status = "resolved"
	event.ResolvedAt = &now
	event.Comment = comment

	return s.alertRepo.UpdateEvent(ctx, event)
}

func (s *alertService) SilenceEvent(ctx context.Context, id uint, duration time.Duration, comment string) error {
	event, err := s.alertRepo.GetEventByID(ctx, id)
	if err != nil {
		return fmt.Errorf("event not found")
	}

	event.Status = "silenced"
	event.Comment = comment

	return s.alertRepo.UpdateEvent(ctx, event)
}

func (s *alertService) GetStats(ctx context.Context) (map[string]interface{}, error) {
	// TODO: 实现统计逻辑
	return map[string]interface{}{
		"total":       0,
		"by_severity": map[string]int{},
		"by_status":   map[string]int{},
	}, nil
}

