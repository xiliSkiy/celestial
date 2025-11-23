package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
)

// TaskService 任务服务接口
type TaskService interface {
	Create(ctx context.Context, req *CreateTaskRequest) (*model.CollectionTask, error)
	Get(ctx context.Context, id uint) (*model.CollectionTask, error)
	Update(ctx context.Context, id uint, req *UpdateTaskRequest) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *ListTaskRequest) ([]*model.CollectionTask, int64, error)
	GetSentinelTasks(ctx context.Context, sentinelID string) ([]*model.CollectionTask, error)
	ReportExecution(ctx context.Context, taskID string, req *ReportExecutionRequest) error
	Trigger(ctx context.Context, id uint) error
	Toggle(ctx context.Context, id uint, enabled bool) error
	GetExecutions(ctx context.Context, id uint, page, pageSize int) ([]*model.TaskExecution, int64, error)
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	DeviceID        string                 `json:"device_id" binding:"required"`
	SentinelID      string                 `json:"sentinel_id" binding:"required"`
	PluginName      string                 `json:"plugin_name" binding:"required"`
	Config          map[string]interface{} `json:"config"`
	IntervalSeconds int                    `json:"interval_seconds" binding:"required"`
	Priority        int                    `json:"priority"`
	RetryCount      int                    `json:"retry_count"`
	TimeoutSeconds  int                    `json:"timeout_seconds"`
	Enabled         bool                   `json:"enabled"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Config          map[string]interface{} `json:"config"`
	IntervalSeconds int                    `json:"interval_seconds"`
	Enabled         *bool                  `json:"enabled"`
}

// ListTaskRequest 任务列表请求
type ListTaskRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	DeviceID   string `form:"device_id"`
	SentinelID string `form:"sentinel_id"`
	PluginName string `form:"plugin_name"`
	Enabled    *bool  `form:"enabled"`
}

// ReportExecutionRequest 执行结果上报请求
type ReportExecutionRequest struct {
	Status           string `json:"status" binding:"required"`
	MetricsCollected int    `json:"metrics_collected"`
	ErrorMessage     string `json:"error_message"`
	ExecutionTimeMs  int    `json:"execution_time_ms"`
	ExecutedAt       string `json:"executed_at"`
}

type taskService struct {
	taskRepo     repository.TaskRepository
	deviceRepo   repository.DeviceRepository
	sentinelRepo repository.SentinelRepository
}

// NewTaskService 创建任务服务
func NewTaskService(taskRepo repository.TaskRepository, deviceRepo repository.DeviceRepository, sentinelRepo repository.SentinelRepository) TaskService {
	return &taskService{
		taskRepo:     taskRepo,
		deviceRepo:   deviceRepo,
		sentinelRepo: sentinelRepo,
	}
}

func (s *taskService) Create(ctx context.Context, req *CreateTaskRequest) (*model.CollectionTask, error) {
	// 验证设备存在
	_, err := s.deviceRepo.GetByDeviceID(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device not found: %s", req.DeviceID)
	}

	// 验证 Sentinel 存在
	_, err = s.sentinelRepo.GetBySentinelID(ctx, req.SentinelID)
	if err != nil {
		return nil, fmt.Errorf("sentinel not found: %s", req.SentinelID)
	}

	// 生成任务 ID
	taskID := fmt.Sprintf("task-%s", uuid.New().String()[:8])

	// 设置默认值
	if req.Priority == 0 {
		req.Priority = 5
	}
	if req.RetryCount == 0 {
		req.RetryCount = 3
	}
	if req.TimeoutSeconds == 0 {
		req.TimeoutSeconds = 30
	}

	// 计算下次执行时间
	nextExecution := time.Now().Add(time.Duration(req.IntervalSeconds) * time.Second)

	task := &model.CollectionTask{
		TaskID:          taskID,
		DeviceID:        req.DeviceID,
		SentinelID:      req.SentinelID,
		PluginName:      req.PluginName,
		Config:          req.Config,
		IntervalSeconds: req.IntervalSeconds,
		Enabled:         req.Enabled,
		Priority:        req.Priority,
		RetryCount:      req.RetryCount,
		TimeoutSeconds:  req.TimeoutSeconds,
		NextExecutionAt: &nextExecution,
	}

	if err := s.taskRepo.Create(ctx, task); err != nil {
		return nil, fmt.Errorf("failed to create task: %w", err)
	}

	return task, nil
}

func (s *taskService) Get(ctx context.Context, id uint) (*model.CollectionTask, error) {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("task not found")
		}
		return nil, fmt.Errorf("failed to get task: %w", err)
	}
	return task, nil
}

func (s *taskService) Update(ctx context.Context, id uint, req *UpdateTaskRequest) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	// 更新字段
	if req.Config != nil {
		task.Config = req.Config
	}
	if req.IntervalSeconds > 0 {
		task.IntervalSeconds = req.IntervalSeconds
	}
	if req.Enabled != nil {
		task.Enabled = *req.Enabled
	}

	return s.taskRepo.Update(ctx, task)
}

func (s *taskService) Delete(ctx context.Context, id uint) error {
	return s.taskRepo.Delete(ctx, id)
}

func (s *taskService) List(ctx context.Context, req *ListTaskRequest) ([]*model.CollectionTask, int64, error) {
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

	filter := &repository.TaskFilter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		DeviceID:   req.DeviceID,
		SentinelID: req.SentinelID,
		PluginName: req.PluginName,
		Enabled:    req.Enabled,
	}

	return s.taskRepo.List(ctx, filter)
}

func (s *taskService) GetSentinelTasks(ctx context.Context, sentinelID string) ([]*model.CollectionTask, error) {
	tasks, err := s.taskRepo.GetBySentinelID(ctx, sentinelID)
	if err != nil {
		return nil, err
	}

	// 为每个任务合并设备的连接配置和设备类型
	for _, task := range tasks {
		// 如果任务的 Config 为空，初始化为空 map
		if task.Config == nil {
			task.Config = make(map[string]interface{})
		}

		// 添加设备类型到配置中（用于状态指标的标签）
		if task.Device != nil && task.Device.DeviceType != "" {
			task.Config["device_type"] = task.Device.DeviceType
		}

		// 合并设备的连接配置
		if task.Device != nil && task.Device.ConnectionConfig != nil && len(task.Device.ConnectionConfig) > 0 {
			// 将设备的连接配置合并到任务配置中
			// 只有当任务配置中不存在该字段时，才从设备配置中复制
			// JSONB 类型本身就是 map[string]interface{}
			for key, value := range task.Device.ConnectionConfig {
				if _, exists := task.Config[key]; !exists {
					task.Config[key] = value
				}
			}
		}
	}

	return tasks, nil
}

func (s *taskService) ReportExecution(ctx context.Context, taskID string, req *ReportExecutionRequest) error {
	// 获取任务
	task, err := s.taskRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	// 解析执行时间
	var executedAt time.Time
	if req.ExecutedAt != "" {
		executedAt, err = time.Parse(time.RFC3339, req.ExecutedAt)
		if err != nil {
			executedAt = time.Now()
		}
	} else {
		executedAt = time.Now()
	}

	// 记录执行结果
	execution := &model.TaskExecution{
		TaskID:           taskID,
		SentinelID:       task.SentinelID,
		Status:           req.Status,
		MetricsCollected: req.MetricsCollected,
		ErrorMessage:     req.ErrorMessage,
		ExecutionTimeMs:  req.ExecutionTimeMs,
		ExecutedAt:       executedAt,
	}

	if err := s.taskRepo.RecordExecution(ctx, execution); err != nil {
		return fmt.Errorf("failed to record execution: %w", err)
	}

	// 更新任务的执行时间
	nextExecution := executedAt.Add(time.Duration(task.IntervalSeconds) * time.Second)
	if err := s.taskRepo.UpdateExecutionTime(ctx, taskID, executedAt, nextExecution); err != nil {
		return fmt.Errorf("failed to update execution time: %w", err)
	}

	return nil
}

func (s *taskService) Trigger(ctx context.Context, id uint) error {
	// 获取任务
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("task not found")
		}
		return fmt.Errorf("failed to get task: %w", err)
	}

	// 检查任务是否启用
	if !task.Enabled {
		return fmt.Errorf("task is disabled")
	}

	// 将下次执行时间设置为当前时间，这样 Sentinel 下次拉取任务时就会立即执行
	now := time.Now()
	task.NextExecutionAt = &now

	// 更新任务
	if err := s.taskRepo.Update(ctx, task); err != nil {
		return fmt.Errorf("failed to update task: %w", err)
	}

	return nil
}

func (s *taskService) Toggle(ctx context.Context, id uint, enabled bool) error {
	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("task not found")
	}

	task.Enabled = enabled
	return s.taskRepo.Update(ctx, task)
}

func (s *taskService) GetExecutions(ctx context.Context, id uint, page, pageSize int) ([]*model.TaskExecution, int64, error) {
	// 设置默认值
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, fmt.Errorf("task not found")
	}

	return s.taskRepo.GetExecutions(ctx, task.TaskID, page, pageSize)
}
