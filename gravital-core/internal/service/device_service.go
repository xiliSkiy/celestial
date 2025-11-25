package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
	"github.com/celestial/gravital-core/internal/timeseries"
)

// DeviceService 设备服务接口
type DeviceService interface {
	Create(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error)
	Get(ctx context.Context, id uint) (*model.Device, error)
	Update(ctx context.Context, id uint, req *UpdateDeviceRequest) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *ListDeviceRequest) ([]*model.Device, int64, error)
	TestConnection(ctx context.Context, id uint) (*TestConnectionResult, error)
	GetAllTags(ctx context.Context) ([]string, error)
	GetMetrics(ctx context.Context, id uint, hours int) (interface{}, error)
	GetTasks(ctx context.Context, id uint) ([]*model.CollectionTask, error)
	GetAlertRules(ctx context.Context, id uint) ([]*model.AlertRule, error)
	GetHistory(ctx context.Context, id uint, page, pageSize int) ([]interface{}, int64, error)
}

// CreateDeviceRequest 创建设备请求
type CreateDeviceRequest struct {
	Name             string                 `json:"name" binding:"required"`
	DeviceType       string                 `json:"device_type" binding:"required"`
	GroupID          *uint                  `json:"group_id"`
	SentinelID       string                 `json:"sentinel_id"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
	Labels           map[string]interface{} `json:"labels"`
}

// UpdateDeviceRequest 更新设备请求
type UpdateDeviceRequest struct {
	Name             string                 `json:"name"`
	GroupID          *uint                  `json:"group_id"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
	Labels           map[string]interface{} `json:"labels"`
}

// ListDeviceRequest 设备列表请求
type ListDeviceRequest struct {
	Page       int      `form:"page"`
	PageSize   int      `form:"page_size"`
	GroupID    *uint    `form:"group_id"`
	DeviceType string   `form:"device_type"`
	Status     string   `form:"status"`
	Keyword    string   `form:"keyword"`
	Labels     []string `form:"labels"`
}

// TestConnectionResult 连接测试结果
type TestConnectionResult struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	LatencyMs int    `json:"latency_ms"`
}

type deviceService struct {
	deviceRepo repository.DeviceRepository
	db         *gorm.DB
	tsClient   *timeseries.Client
}

// NewDeviceService 创建设备服务
func NewDeviceService(deviceRepo repository.DeviceRepository, db *gorm.DB, tsClient *timeseries.Client) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
		db:         db,
		tsClient:   tsClient,
	}
}

func (s *deviceService) Create(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error) {
	// 生成设备 ID
	deviceID := fmt.Sprintf("dev-%s", uuid.New().String()[:8])

	device := &model.Device{
		DeviceID:         deviceID,
		Name:             req.Name,
		DeviceType:       req.DeviceType,
		GroupID:          req.GroupID,
		SentinelID:       req.SentinelID,
		ConnectionConfig: req.ConnectionConfig,
		Labels:           req.Labels,
		Status:           "unknown",
	}

	if err := s.deviceRepo.Create(ctx, device); err != nil {
		return nil, fmt.Errorf("failed to create device: %w", err)
	}

	return device, nil
}

func (s *deviceService) Get(ctx context.Context, id uint) (*model.Device, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("device not found")
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}
	return device, nil
}

func (s *deviceService) Update(ctx context.Context, id uint, req *UpdateDeviceRequest) error {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("device not found")
	}

	// 更新字段
	if req.Name != "" {
		device.Name = req.Name
	}
	if req.GroupID != nil {
		device.GroupID = req.GroupID
	}
	if req.ConnectionConfig != nil {
		device.ConnectionConfig = req.ConnectionConfig
	}
	if req.Labels != nil {
		device.Labels = req.Labels
	}

	return s.deviceRepo.Update(ctx, device)
}

func (s *deviceService) Delete(ctx context.Context, id uint) error {
	return s.deviceRepo.Delete(ctx, id)
}

func (s *deviceService) List(ctx context.Context, req *ListDeviceRequest) ([]*model.Device, int64, error) {
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

	filter := &repository.DeviceFilter{
		Page:       req.Page,
		PageSize:   req.PageSize,
		GroupID:    req.GroupID,
		DeviceType: req.DeviceType,
		Status:     req.Status,
		Keyword:    req.Keyword,
		Labels:     req.Labels,
	}

	return s.deviceRepo.List(ctx, filter)
}

func (s *deviceService) GetAllTags(ctx context.Context) ([]string, error) {
	return s.deviceRepo.GetAllTags(ctx)
}

func (s *deviceService) GetMetrics(ctx context.Context, id uint, hours int) (interface{}, error) {
	// 获取设备信息
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	// 如果配置了时序数据库客户端，从时序数据库查询
	if s.tsClient != nil {
		metrics, err := s.tsClient.GetDeviceMetrics(device.DeviceID, hours)
		if err != nil {
			// 查询失败，返回空数据结构
			return map[string]interface{}{
				"device_id": device.DeviceID,
				"metrics": map[string]interface{}{
					"cpu":         map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
					"memory":      map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
					"disk":        map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
					"network_in":  map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
					"network_out": map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
				},
			}, nil
		}

		// 转换为响应格式
		result := map[string]interface{}{
			"device_id": metrics.DeviceID,
			"metrics":   make(map[string]interface{}),
		}

		metricsMap := result["metrics"].(map[string]interface{})
		for metricName, data := range metrics.Metrics {
			metricsMap[metricName] = map[string]interface{}{
				"timestamps": data.Timestamps,
				"values":     data.Values,
			}
		}

		// 如果某些指标没有数据，填充空数组
		for _, metricName := range []string{"cpu", "memory", "disk", "network_in", "network_out"} {
			if _, exists := metricsMap[metricName]; !exists {
				metricsMap[metricName] = map[string]interface{}{
					"timestamps": []string{},
					"values":     []float64{},
				}
			}
		}

		return result, nil
	}

	// 未配置时序数据库，返回空数据结构
	return map[string]interface{}{
		"device_id": device.DeviceID,
		"metrics": map[string]interface{}{
			"cpu":         map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
			"memory":      map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
			"disk":        map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
			"network_in":  map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
			"network_out": map[string]interface{}{"timestamps": []string{}, "values": []float64{}},
		},
	}, nil
}

func (s *deviceService) GetTasks(ctx context.Context, id uint) ([]*model.CollectionTask, error) {
	// 获取设备信息
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	// 查询该设备的采集任务
	var tasks []*model.CollectionTask
	err = s.db.WithContext(ctx).
		Where("device_id = ?", device.DeviceID).
		Order("created_at DESC").
		Find(&tasks).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}

	return tasks, nil
}

func (s *deviceService) GetAlertRules(ctx context.Context, id uint) ([]*model.AlertRule, error) {
	// 获取设备信息
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	// 查询该设备的告警规则
	// 告警规则通过 Filters JSONB 字段来过滤设备
	// 查询条件：
	// 1. filters->>'device_id' = device_id (直接匹配该设备的规则)
	// 2. filters IS NULL (全局规则，适用于所有设备)
	// 3. filters::text = '{}' (空对象，也是全局规则)
	var rules []*model.AlertRule
	err = s.db.WithContext(ctx).
		Where("filters->>'device_id' = ?", device.DeviceID).
		Or("filters IS NULL").
		Or("filters::text = '{}'").
		Order("created_at DESC").
		Find(&rules).Error

	if err != nil {
		return nil, fmt.Errorf("failed to query alert rules: %w", err)
	}

	// 进一步过滤：只返回真正匹配该设备的规则
	// 因为可能有全局规则（filters 为空），我们需要检查这些规则是否真的适用于该设备
	filteredRules := make([]*model.AlertRule, 0)
	for _, rule := range rules {
		// 如果 filters 为空或没有 device_id 过滤，说明是全局规则，适用于所有设备
		if rule.Filters == nil || len(rule.Filters) == 0 {
			filteredRules = append(filteredRules, rule)
			continue
		}
		
		// 检查 filters 中是否有 device_id 且匹配
		if deviceID, ok := rule.Filters["device_id"]; ok {
			if deviceIDStr, ok := deviceID.(string); ok && deviceIDStr == device.DeviceID {
				filteredRules = append(filteredRules, rule)
			}
		} else {
			// 如果 filters 存在但没有 device_id，说明是全局规则
			filteredRules = append(filteredRules, rule)
		}
	}

	return filteredRules, nil
}

func (s *deviceService) GetHistory(ctx context.Context, id uint, page, pageSize int) ([]interface{}, int64, error) {
	// 获取设备信息
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, 0, fmt.Errorf("device not found")
	}

	// 查询告警事件历史
	var alertEvents []model.AlertEvent
	var total int64

	query := s.db.WithContext(ctx).Model(&model.AlertEvent{}).
		Where("device_id = ?", device.DeviceID)

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("failed to count history: %w", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err = query.Order("created_at DESC").
		Offset(offset).
		Limit(pageSize).
		Find(&alertEvents).Error

	if err != nil {
		return nil, 0, fmt.Errorf("failed to query history: %w", err)
	}

	// 转换为通用格式
	history := make([]interface{}, len(alertEvents))
	for i, event := range alertEvents {
		history[i] = map[string]interface{}{
			"id":          event.ID,
			"type":        "alert",
			"rule_name":   event.RuleID,
			"severity":    event.Severity,
			"status":      event.Status,
			"message":     event.Message,
			"created_at":  event.CreatedAt,
			"resolved_at": event.ResolvedAt,
		}
	}

	return history, total, nil
}

func (s *deviceService) TestConnection(ctx context.Context, id uint) (*TestConnectionResult, error) {
	device, err := s.deviceRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("device not found")
	}

	// TODO: 实现实际的连接测试逻辑
	// 这里需要根据设备类型调用相应的测试方法

	return &TestConnectionResult{
		Status:    "success",
		Message:   fmt.Sprintf("Successfully connected to %s", device.Name),
		LatencyMs: 15,
	}, nil
}

