# Gravital Core 实现指南

本文档提供完整的实现示例，帮助快速完成剩余的代码。

## 📁 实现顺序

```
Repository → Service → Handler
```

## 1. Repository 层实现示例

### DeviceRepository

```go
// internal/repository/device_repository.go
package repository

import (
	"context"
	"gorm.io/gorm"
	"github.com/celestial/gravital-core/internal/model"
)

type DeviceRepository interface {
	Create(ctx context.Context, device *model.Device) error
	GetByID(ctx context.Context, id uint) (*model.Device, error)
	GetByDeviceID(ctx context.Context, deviceID string) (*model.Device, error)
	Update(ctx context.Context, device *model.Device) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter *DeviceFilter) ([]*model.Device, int64, error)
}

type DeviceFilter struct {
	Page       int
	PageSize   int
	GroupID    *uint
	DeviceType string
	Status     string
	Keyword    string
}

type deviceRepository struct {
	db *gorm.DB
}

func NewDeviceRepository(db *gorm.DB) DeviceRepository {
	return &deviceRepository{db: db}
}

func (r *deviceRepository) Create(ctx context.Context, device *model.Device) error {
	return r.db.WithContext(ctx).Create(device).Error
}

func (r *deviceRepository) GetByID(ctx context.Context, id uint) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).Preload("Group").First(&device, id).Error
	return &device, err
}

func (r *deviceRepository) GetByDeviceID(ctx context.Context, deviceID string) (*model.Device, error) {
	var device model.Device
	err := r.db.WithContext(ctx).Preload("Group").Where("device_id = ?", deviceID).First(&device).Error
	return &device, err
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

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Preload("Group").Offset(offset).Limit(filter.PageSize).Find(&devices).Error

	return devices, total, err
}
```

## 2. Service 层实现示例

### DeviceService

```go
// internal/service/device_service.go
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

type DeviceService interface {
	Create(ctx context.Context, req *CreateDeviceRequest) (*model.Device, error)
	Get(ctx context.Context, id uint) (*model.Device, error)
	Update(ctx context.Context, id uint, req *UpdateDeviceRequest) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, req *ListDeviceRequest) ([]*model.Device, int64, error)
	TestConnection(ctx context.Context, id uint) (*TestConnectionResult, error)
}

type CreateDeviceRequest struct {
	Name             string                 `json:"name" binding:"required"`
	DeviceType       string                 `json:"device_type" binding:"required"`
	GroupID          *uint                  `json:"group_id"`
	SentinelID       string                 `json:"sentinel_id"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
	Labels           map[string]interface{} `json:"labels"`
}

type UpdateDeviceRequest struct {
	Name             string                 `json:"name"`
	GroupID          *uint                  `json:"group_id"`
	ConnectionConfig map[string]interface{} `json:"connection_config"`
	Labels           map[string]interface{} `json:"labels"`
}

type ListDeviceRequest struct {
	Page       int    `form:"page"`
	PageSize   int    `form:"page_size"`
	GroupID    *uint  `form:"group_id"`
	DeviceType string `form:"device_type"`
	Status     string `form:"status"`
	Keyword    string `form:"keyword"`
}

type TestConnectionResult struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	LatencyMs int    `json:"latency_ms"`
}

type deviceService struct {
	deviceRepo repository.DeviceRepository
}

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
	}

	return s.deviceRepo.List(ctx, filter)
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
```

## 3. Handler 层实现示例

### DeviceHandler

```go
// internal/api/handler/device_handler.go
package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/celestial/gravital-core/internal/service"
)

type DeviceHandler struct {
	deviceService service.DeviceService
}

func NewDeviceHandler(deviceService service.DeviceService) *DeviceHandler {
	return &DeviceHandler{
		deviceService: deviceService,
	}
}

// List 获取设备列表
func (h *DeviceHandler) List(c *gin.Context) {
	var req service.ListDeviceRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	devices, total, err := h.deviceService.List(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "获取设备列表失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"total":     total,
			"page":      req.Page,
			"page_size": req.PageSize,
			"items":     devices,
		},
	})
}

// Get 获取设备详情
func (h *DeviceHandler) Get(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	device, err := h.deviceService.Get(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    50001,
			"message": "设备不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": device,
	})
}

// Create 创建设备
func (h *DeviceHandler) Create(c *gin.Context) {
	var req service.CreateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	device, err := h.deviceService.Create(c.Request.Context(), &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "创建设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"device_id": device.DeviceID,
		},
	})
}

// Update 更新设备
func (h *DeviceHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	var req service.UpdateDeviceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	if err := h.deviceService.Update(c.Request.Context(), uint(id), &req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "更新设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// Delete 删除设备
func (h *DeviceHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	if err := h.deviceService.Delete(c.Request.Context(), uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "删除设备失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// TestConnection 测试设备连接
func (h *DeviceHandler) TestConnection(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    40001,
			"message": "无效的设备 ID",
		})
		return
	}

	result, err := h.deviceService.TestConnection(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    10001,
			"message": "测试连接失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": result,
	})
}

// BatchImport 批量导入设备
func (h *DeviceHandler) BatchImport(c *gin.Context) {
	// TODO: 实现批量导入逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// GetGroupTree 获取设备分组树
func (h *DeviceHandler) GetGroupTree(c *gin.Context) {
	// TODO: 实现分组树查询
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": []interface{}{},
	})
}

// CreateGroup 创建设备分组
func (h *DeviceHandler) CreateGroup(c *gin.Context) {
	// TODO: 实现创建分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// UpdateGroup 更新设备分组
func (h *DeviceHandler) UpdateGroup(c *gin.Context) {
	// TODO: 实现更新分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}

// DeleteGroup 删除设备分组
func (h *DeviceHandler) DeleteGroup(c *gin.Context) {
	// TODO: 实现删除分组
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "功能开发中",
	})
}
```

## 4. 其他 Handler 占位实现

```go
// internal/api/handler/common.go
package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "healthy",
		"components": gin.H{
			"database": "healthy",
			"redis":    "healthy",
		},
	})
}

// Version 版本信息
func Version(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version":    "1.0.0",
		"build_time": "2025-11-02",
	})
}

// SystemInfo 系统信息
func SystemInfo(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"version": "1.0.0",
			"uptime":  "1h30m",
		},
	})
}

// GetConfig 获取配置
func GetConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{},
	})
}

// UpdateConfig 更新配置
func UpdateConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
	})
}

// IngestData 数据采集
func IngestData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"data": gin.H{
			"received": 0,
		},
	})
}
```

## 5. 按照相同模式实现其他模块

### SentinelRepository/Service/Handler
### TaskRepository/Service/Handler
### AlertRepository/Service/Handler

每个模块都遵循相同的三层架构模式。

## 6. 编译和运行

```bash
# 下载依赖
cd /Users/liangxin/Downloads/code/celestial/gravital-core
go mod tidy

# 编译
make build

# 启动数据库
docker-compose up -d postgres redis

# 运行迁移
make migrate-up DB_URL="postgres://postgres:postgres@localhost:5432/gravital?sslmode=disable"

# 运行
./bin/gravital-core -c config/config.yaml
```

## 7. 测试 API

```bash
# 登录
TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.data.token')

# 创建设备
curl -X POST http://localhost:8080/api/v1/devices \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Switch",
    "device_type": "switch",
    "connection_config": {"host": "192.168.1.1"}
  }'

# 获取设备列表
curl -X GET "http://localhost:8080/api/v1/devices?page=1&page_size=10" \
  -H "Authorization: Bearer $TOKEN"
```

## 📝 注意事项

1. **错误处理**: 所有层都要有完善的错误处理
2. **日志记录**: 关键操作要记录日志
3. **参数验证**: Handler 层要验证所有输入参数
4. **事务处理**: Service 层处理需要事务的操作
5. **上下文传递**: 使用 context 传递请求上下文
6. **权限检查**: 敏感操作要检查权限

## 🚀 快速完成开发

按照以上示例，依次实现：
1. ✅ DeviceRepository/Service/Handler（示例已提供）
2. ⏳ SentinelRepository/Service/Handler
3. ⏳ TaskRepository/Service/Handler
4. ⏳ AlertRepository/Service/Handler
5. ⏳ AuthService（完善）

完成这些后，系统就可以与 Sentinel 进行基本的交互了！

