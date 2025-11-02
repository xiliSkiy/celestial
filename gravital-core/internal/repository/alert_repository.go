package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// AlertRepository 告警仓库接口
type AlertRepository interface {
	// 告警规则
	CreateRule(ctx context.Context, rule *model.AlertRule) error
	GetRuleByID(ctx context.Context, id uint) (*model.AlertRule, error)
	UpdateRule(ctx context.Context, rule *model.AlertRule) error
	DeleteRule(ctx context.Context, id uint) error
	ListRules(ctx context.Context, filter *AlertRuleFilter) ([]*model.AlertRule, int64, error)

	// 告警事件
	CreateEvent(ctx context.Context, event *model.AlertEvent) error
	GetEventByID(ctx context.Context, id uint) (*model.AlertEvent, error)
	UpdateEvent(ctx context.Context, event *model.AlertEvent) error
	ListEvents(ctx context.Context, filter *AlertEventFilter) ([]*model.AlertEvent, int64, error)
}

// AlertRuleFilter 告警规则过滤条件
type AlertRuleFilter struct {
	Page     int
	PageSize int
	Enabled  *bool
	Severity string
	Keyword  string
}

// AlertEventFilter 告警事件过滤条件
type AlertEventFilter struct {
	Page     int
	PageSize int
	Status   string
	Severity string
	DeviceID string
	RuleID   *uint
}

type alertRepository struct {
	db *gorm.DB
}

// NewAlertRepository 创建告警仓库
func NewAlertRepository(db *gorm.DB) AlertRepository {
	return &alertRepository{db: db}
}

func (r *alertRepository) CreateRule(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Create(rule).Error
}

func (r *alertRepository) GetRuleByID(ctx context.Context, id uint) (*model.AlertRule, error) {
	var rule model.AlertRule
	err := r.db.WithContext(ctx).First(&rule, id).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

func (r *alertRepository) UpdateRule(ctx context.Context, rule *model.AlertRule) error {
	return r.db.WithContext(ctx).Save(rule).Error
}

func (r *alertRepository) DeleteRule(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.AlertRule{}, id).Error
}

func (r *alertRepository) ListRules(ctx context.Context, filter *AlertRuleFilter) ([]*model.AlertRule, int64, error) {
	var rules []*model.AlertRule
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertRule{})

	// 应用过滤条件
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.Keyword != "" {
		query = query.Where("rule_name LIKE ?", "%"+filter.Keyword+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Find(&rules).Error

	return rules, total, err
}

func (r *alertRepository) CreateEvent(ctx context.Context, event *model.AlertEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *alertRepository) GetEventByID(ctx context.Context, id uint) (*model.AlertEvent, error) {
	var event model.AlertEvent
	err := r.db.WithContext(ctx).Preload("Rule").First(&event, id).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *alertRepository) UpdateEvent(ctx context.Context, event *model.AlertEvent) error {
	return r.db.WithContext(ctx).Save(event).Error
}

func (r *alertRepository) ListEvents(ctx context.Context, filter *AlertEventFilter) ([]*model.AlertEvent, int64, error) {
	var events []*model.AlertEvent
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AlertEvent{})

	// 应用过滤条件
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Severity != "" {
		query = query.Where("severity = ?", filter.Severity)
	}
	if filter.DeviceID != "" {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.RuleID != nil {
		query = query.Where("rule_id = ?", *filter.RuleID)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Preload("Rule").Offset(offset).Limit(filter.PageSize).Order("triggered_at DESC").Find(&events).Error

	return events, total, err
}

