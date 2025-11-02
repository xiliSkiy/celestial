package repository

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
)

// TaskRepository 任务仓库接口
type TaskRepository interface {
	Create(ctx context.Context, task *model.CollectionTask) error
	GetByID(ctx context.Context, id uint) (*model.CollectionTask, error)
	GetByTaskID(ctx context.Context, taskID string) (*model.CollectionTask, error)
	Update(ctx context.Context, task *model.CollectionTask) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter *TaskFilter) ([]*model.CollectionTask, int64, error)
	GetBySentinelID(ctx context.Context, sentinelID string) ([]*model.CollectionTask, error)
	RecordExecution(ctx context.Context, execution *model.TaskExecution) error
	UpdateExecutionTime(ctx context.Context, taskID string, lastExecuted, nextExecution time.Time) error
	GetExecutions(ctx context.Context, taskID string, page, pageSize int) ([]*model.TaskExecution, int64, error)
}

// TaskFilter 任务过滤条件
type TaskFilter struct {
	Page       int
	PageSize   int
	DeviceID   string
	SentinelID string
	PluginName string
	Enabled    *bool
}

type taskRepository struct {
	db *gorm.DB
}

// NewTaskRepository 创建任务仓库
func NewTaskRepository(db *gorm.DB) TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) Create(ctx context.Context, task *model.CollectionTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetByID(ctx context.Context, id uint) (*model.CollectionTask, error) {
	var task model.CollectionTask
	err := r.db.WithContext(ctx).Preload("Device").First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) GetByTaskID(ctx context.Context, taskID string) (*model.CollectionTask, error) {
	var task model.CollectionTask
	err := r.db.WithContext(ctx).Preload("Device").Where("task_id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) Update(ctx context.Context, task *model.CollectionTask) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *taskRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.CollectionTask{}, id).Error
}

func (r *taskRepository) List(ctx context.Context, filter *TaskFilter) ([]*model.CollectionTask, int64, error) {
	var tasks []*model.CollectionTask
	var total int64

	query := r.db.WithContext(ctx).Model(&model.CollectionTask{})

	// 应用过滤条件
	if filter.DeviceID != "" {
		query = query.Where("device_id = ?", filter.DeviceID)
	}
	if filter.SentinelID != "" {
		query = query.Where("sentinel_id = ?", filter.SentinelID)
	}
	if filter.PluginName != "" {
		query = query.Where("plugin_name = ?", filter.PluginName)
	}
	if filter.Enabled != nil {
		query = query.Where("enabled = ?", *filter.Enabled)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Preload("Device").Offset(offset).Limit(filter.PageSize).Find(&tasks).Error

	return tasks, total, err
}

func (r *taskRepository) GetBySentinelID(ctx context.Context, sentinelID string) ([]*model.CollectionTask, error) {
	var tasks []*model.CollectionTask
	err := r.db.WithContext(ctx).
		Preload("Device").
		Where("sentinel_id = ? AND enabled = ?", sentinelID, true).
		Find(&tasks).Error
	return tasks, err
}

func (r *taskRepository) RecordExecution(ctx context.Context, execution *model.TaskExecution) error {
	return r.db.WithContext(ctx).Create(execution).Error
}

func (r *taskRepository) UpdateExecutionTime(ctx context.Context, taskID string, lastExecuted, nextExecution time.Time) error {
	return r.db.WithContext(ctx).Model(&model.CollectionTask{}).
		Where("task_id = ?", taskID).
		Updates(map[string]interface{}{
			"last_executed_at":  lastExecuted,
			"next_execution_at": nextExecution,
		}).Error
}

func (r *taskRepository) GetExecutions(ctx context.Context, taskID string, page, pageSize int) ([]*model.TaskExecution, int64, error) {
	var executions []*model.TaskExecution
	var total int64

	query := r.db.WithContext(ctx).Model(&model.TaskExecution{}).Where("task_id = ?", taskID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.Order("executed_at DESC").Offset(offset).Limit(pageSize).Find(&executions).Error

	return executions, total, err
}

