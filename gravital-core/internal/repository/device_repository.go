package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// DeviceRepository 设备仓库接口
type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id uint) (*model.Device, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter *DeviceFilter) ([]*model.Device, int64, error)
	GetAllTags(ctx context.Context) ([]string, error)
}

// DeviceFilter 设备过滤条件
type DeviceFilter struct {
	Page       int
	PageSize   int
	GroupID    *uint
	DeviceType string
	Status     string
	Keyword    string
	Labels     []string
}

type deviceRepository struct {
	db *gorm.DB
}

// NewDeviceRepository 创建设备仓库
func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *deviceRepository) GetByID(ctx context.Context, id uint) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).Preload("Group").First(&device, id).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).Preload("Group").Where("device_id = ?", deviceID).First(&device).Error
	if err != nil {
		return nil, err
	}
	return &device, nil
}

func (r *deviceRepository) Update(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Save(device).Error
}

func (r *deviceRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Device{}, id).Error
}

func (r *deviceRepository) List(ctx context.Context, filter *DeviceFilter) ([]*model.Device, int64, error) {
	var devices []*model.Device
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Device{})

	// 应用过滤条件
	if filter.GroupID != nil {
		query = query.Where("group_id = ?", *filter.GroupID)
	}
	if filter.DeviceType != "" {
		query = query.Where("device_type = ?", filter.DeviceType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Keyword != "" {
		query = query.Where("name LIKE ? OR device_id LIKE ?", "%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}
	
	// 标签过滤
	if len(filter.Labels) > 0 {
		for _, label := range filter.Labels {
			// 标签格式: key:value
			query = query.Where("labels::text LIKE ?", "%"+label+"%")
		}
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Preload("Group").Offset(offset).Limit(filter.PageSize).Find(&devices).Error

	return devices, total, err
}

func (r *deviceRepository) GetAllTags(ctx context.Context) ([]string, error) {
	var devices []*model.Device
	if err := r.db.WithContext(ctx).Select("labels").Find(&devices).Error; err != nil {
		return nil, err
	}

	// 收集所有唯一的标签
	tagSet := make(map[string]bool)
	for _, device := range devices {
		if device.Labels != nil {
			for key, value := range device.Labels {
				if strValue, ok := value.(string); ok {
					tag := key + ":" + strValue
					tagSet[tag] = true
				}
			}
		}
	}

	// 转换为切片
	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

