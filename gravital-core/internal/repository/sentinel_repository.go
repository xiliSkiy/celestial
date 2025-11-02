package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// SentinelRepository Sentinel 仓库接口
type SentinelRepository interface {
	Create(ctx context.Context, sentinel *model.Sentinel) error
	GetByID(ctx context.Context, id uint) (*model.Sentinel, error)
	GetBySentinelID(ctx context.Context, sentinelID string) (*model.Sentinel, error)
	Update(ctx context.Context, sentinel *model.Sentinel) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter *SentinelFilter) ([]*model.Sentinel, int64, error)
	UpdateHeartbeat(ctx context.Context, sentinelID string, heartbeat *model.SentinelHeartbeat) error
	UpdateStatus(ctx context.Context, sentinelID string, status string) error
}

// SentinelFilter Sentinel 过滤条件
type SentinelFilter struct {
	Page     int
	PageSize int
	Status   string
	Region   string
}

type sentinelRepository struct {
	db *gorm.DB
}

// NewSentinelRepository 创建 Sentinel 仓库
func NewSentinelRepository(db *gorm.DB) SentinelRepository {
	return &sentinelRepository{db: db}
}

func (r *sentinelRepository) Create(ctx context.Context, sentinel *model.Sentinel) error {
	return r.db.WithContext(ctx).Create(sentinel).Error
}

func (r *sentinelRepository) GetByID(ctx context.Context, id uint) (*model.Sentinel, error) {
	var sentinel model.Sentinel
	err := r.db.WithContext(ctx).First(&sentinel, id).Error
	if err != nil {
		return nil, err
	}
	return &sentinel, nil
}

func (r *sentinelRepository) GetBySentinelID(ctx context.Context, sentinelID string) (*model.Sentinel, error) {
	var sentinel model.Sentinel
	err := r.db.WithContext(ctx).Where("sentinel_id = ?", sentinelID).First(&sentinel).Error
	if err != nil {
		return nil, err
	}
	return &sentinel, nil
}

func (r *sentinelRepository) Update(ctx context.Context, sentinel *model.Sentinel) error {
	return r.db.WithContext(ctx).Save(sentinel).Error
}

func (r *sentinelRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Sentinel{}, id).Error
}

func (r *sentinelRepository) List(ctx context.Context, filter *SentinelFilter) ([]*model.Sentinel, int64, error) {
	var sentinels []*model.Sentinel
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Sentinel{})

	// 应用过滤条件
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Region != "" {
		query = query.Where("region = ?", filter.Region)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).Order("last_heartbeat DESC").Find(&sentinels).Error

	return sentinels, total, err
}

func (r *sentinelRepository) UpdateHeartbeat(ctx context.Context, sentinelID string, heartbeat *model.SentinelHeartbeat) error {
	// 保存心跳记录
	if err := r.db.WithContext(ctx).Create(heartbeat).Error; err != nil {
		return err
	}

	// 更新 Sentinel 的最后心跳时间
	now := time.Now()
	return r.db.WithContext(ctx).Model(&model.Sentinel{}).
		Where("sentinel_id = ?", sentinelID).
		Updates(map[string]interface{}{
			"last_heartbeat": now,
			"status":         "online",
			"updated_at":     now,
		}).Error
}

func (r *sentinelRepository) UpdateStatus(ctx context.Context, sentinelID string, status string) error {
	return r.db.WithContext(ctx).Model(&model.Sentinel{}).
		Where("sentinel_id = ?", sentinelID).
		Update("status", status).Error
}

