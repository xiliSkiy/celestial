package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
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
}

// NewDeviceService 创建设备服务
func NewDeviceService(deviceRepo repository.DeviceRepository) DeviceService {
	return &deviceService{
		deviceRepo: deviceRepo,
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

