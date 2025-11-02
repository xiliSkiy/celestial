package repository

import (
	"context"
	"fmt"

	"github.com/celestial/gravital-core/internal/model"
	"gorm.io/gorm"
)

// ForwarderRepository 转发器仓库接口
type ForwarderRepository interface {
	Create(ctx context.Context, config *model.ForwarderConfig) error
	Update(ctx context.Context, config *model.ForwarderConfig) error
	Delete(ctx context.Context, name string) error
	GetByName(ctx context.Context, name string) (*model.ForwarderConfig, error)
	List(ctx context.Context, enabled *bool) ([]*model.ForwarderConfig, error)
	RecordStats(ctx context.Context, stats *model.ForwarderStats) error
	GetStats(ctx context.Context, forwarderName string, limit int) ([]*model.ForwarderStats, error)
}

type forwarderRepository struct {
	db *gorm.DB
}

// NewForwarderRepository 创建转发器仓库
func NewForwarderRepository(db *gorm.DB) ForwarderRepository {
	return &forwarderRepository{db: db}
}

// Create 创建转发器配置
func (r *forwarderRepository) Create(ctx context.Context, config *model.ForwarderConfig) error {
	if err := r.db.WithContext(ctx).Create(config).Error; err != nil {
		return fmt.Errorf("failed to create forwarder config: %w", err)
	}
	return nil
}

// Update 更新转发器配置
func (r *forwarderRepository) Update(ctx context.Context, config *model.ForwarderConfig) error {
	if err := r.db.WithContext(ctx).Save(config).Error; err != nil {
		return fmt.Errorf("failed to update forwarder config: %w", err)
	}
	return nil
}

// Delete 删除转发器配置
func (r *forwarderRepository) Delete(ctx context.Context, name string) error {
	if err := r.db.WithContext(ctx).Where("name = ?", name).Delete(&model.ForwarderConfig{}).Error; err != nil {
		return fmt.Errorf("failed to delete forwarder config: %w", err)
	}
	return nil
}

// GetByName 根据名称获取转发器配置
func (r *forwarderRepository) GetByName(ctx context.Context, name string) (*model.ForwarderConfig, error) {
	var config model.ForwarderConfig
	if err := r.db.WithContext(ctx).Where("name = ?", name).First(&config).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("forwarder config not found: %s", name)
		}
		return nil, fmt.Errorf("failed to get forwarder config: %w", err)
	}
	return &config, nil
}

// List 列出转发器配置
func (r *forwarderRepository) List(ctx context.Context, enabled *bool) ([]*model.ForwarderConfig, error) {
	var configs []*model.ForwarderConfig
	query := r.db.WithContext(ctx)

	if enabled != nil {
		query = query.Where("enabled = ?", *enabled)
	}

	if err := query.Order("created_at DESC").Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to list forwarder configs: %w", err)
	}

	return configs, nil
}

// RecordStats 记录转发器统计
func (r *forwarderRepository) RecordStats(ctx context.Context, stats *model.ForwarderStats) error {
	if err := r.db.WithContext(ctx).Create(stats).Error; err != nil {
		return fmt.Errorf("failed to record forwarder stats: %w", err)
	}
	return nil
}

// GetStats 获取转发器统计
func (r *forwarderRepository) GetStats(ctx context.Context, forwarderName string, limit int) ([]*model.ForwarderStats, error) {
	var stats []*model.ForwarderStats
	query := r.db.WithContext(ctx).Where("forwarder_name = ?", forwarderName)

	if limit > 0 {
		query = query.Limit(limit)
	}

	if err := query.Order("recorded_at DESC").Find(&stats).Error; err != nil {
		return nil, fmt.Errorf("failed to get forwarder stats: %w", err)
	}

	return stats, nil
}

