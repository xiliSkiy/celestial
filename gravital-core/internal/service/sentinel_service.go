package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
)

// SentinelService Sentinel 服务接口
type SentinelService interface {
	Register(ctx context.Context, req *RegisterSentinelRequest) (*RegisterSentinelResponse, error)
	Heartbeat(ctx context.Context, sentinelID string, req *HeartbeatRequest) error
	Get(ctx context.Context, id uint) (*model.Sentinel, error)
	List(ctx context.Context, req *ListSentinelRequest) ([]*model.Sentinel, int64, error)
	Delete(ctx context.Context, id uint) error
	Control(ctx context.Context, sentinelID string, action string) error
}

// RegisterSentinelRequest Sentinel 注册请求
type RegisterSentinelRequest struct {
	Name      string                 `json:"name" binding:"required"`
	Hostname  string                 `json:"hostname" binding:"required"`
	IPAddress string                 `json:"ip_address"`
	Version   string                 `json:"version" binding:"required"`
	OS        string                 `json:"os" binding:"required"`
	Arch      string                 `json:"arch" binding:"required"`
	Region    string                 `json:"region"`
	Labels    map[string]interface{} `json:"labels"`
}

// RegisterSentinelResponse Sentinel 注册响应
type RegisterSentinelResponse struct {
	SentinelID string                 `json:"sentinel_id"`
	APIToken   string                 `json:"api_token"`
	Config     map[string]interface{} `json:"config"`
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	CPUUsage      float64 `json:"cpu_usage"`
	MemoryUsage   float64 `json:"memory_usage"`
	DiskUsage     float64 `json:"disk_usage"`
	TaskCount     int     `json:"task_count"`
	PluginCount   int     `json:"plugin_count"`
	UptimeSeconds int64   `json:"uptime_seconds"`
	Version       string  `json:"version"`
}

// ListSentinelRequest Sentinel 列表请求
type ListSentinelRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Status   string `form:"status"`
	Region   string `form:"region"`
}

type sentinelService struct {
	sentinelRepo repository.SentinelRepository
}

// NewSentinelService 创建 Sentinel 服务
func NewSentinelService(sentinelRepo repository.SentinelRepository) SentinelService {
	return &sentinelService{
		sentinelRepo: sentinelRepo,
	}
}

func (s *sentinelService) Register(ctx context.Context, req *RegisterSentinelRequest) (*RegisterSentinelResponse, error) {
	// 生成 Sentinel ID
	sentinelID := fmt.Sprintf("sentinel-%s-%d", req.Hostname, time.Now().Unix())

	// 生成 API Token
	apiToken, err := generateAPIToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate api token: %w", err)
	}

	// 检查是否已存在
	existing, err := s.sentinelRepo.GetBySentinelID(ctx, sentinelID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing sentinel: %w", err)
	}

	now := time.Now()

	if existing != nil {
		// 更新已存在的 Sentinel
		existing.Name = req.Name
		existing.IPAddress = req.IPAddress
		existing.Version = req.Version
		existing.OS = req.OS
		existing.Arch = req.Arch
		existing.Region = req.Region
		existing.Labels = req.Labels
		existing.Status = "online"
		existing.LastHeartbeat = &now
		existing.UpdatedAt = now

		if err := s.sentinelRepo.Update(ctx, existing); err != nil {
			return nil, fmt.Errorf("failed to update sentinel: %w", err)
		}

		return &RegisterSentinelResponse{
			SentinelID: existing.SentinelID,
			APIToken:   existing.APIToken,
			Config: map[string]interface{}{
				"heartbeat_interval": 30,
				"task_fetch_interval": 60,
			},
		}, nil
	}

	// 创建新的 Sentinel
	sentinel := &model.Sentinel{
		SentinelID:    sentinelID,
		Name:          req.Name,
		Hostname:      req.Hostname,
		IPAddress:     req.IPAddress,
		Version:       req.Version,
		OS:            req.OS,
		Arch:          req.Arch,
		Region:        req.Region,
		Labels:        req.Labels,
		APIToken:      apiToken,
		Status:        "online",
		LastHeartbeat: &now,
		RegisteredAt:  now,
		UpdatedAt:     now,
	}

	if err := s.sentinelRepo.Create(ctx, sentinel); err != nil {
		return nil, fmt.Errorf("failed to create sentinel: %w", err)
	}

	return &RegisterSentinelResponse{
		SentinelID: sentinelID,
		APIToken:   apiToken,
		Config: map[string]interface{}{
			"heartbeat_interval": 30,
			"task_fetch_interval": 60,
		},
	}, nil
}

func (s *sentinelService) Heartbeat(ctx context.Context, sentinelID string, req *HeartbeatRequest) error {
	// 检查 Sentinel 是否存在
	sentinel, err := s.sentinelRepo.GetBySentinelID(ctx, sentinelID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("sentinel not found")
		}
		return fmt.Errorf("failed to get sentinel: %w", err)
	}

	// 创建心跳记录
	heartbeat := &model.SentinelHeartbeat{
		SentinelID:    sentinelID,
		CPUUsage:      req.CPUUsage,
		MemoryUsage:   req.MemoryUsage,
		DiskUsage:     req.DiskUsage,
		TaskCount:     req.TaskCount,
		PluginCount:   req.PluginCount,
		UptimeSeconds: req.UptimeSeconds,
		ReceivedAt:    time.Now(),
	}

	// 更新心跳
	if err := s.sentinelRepo.UpdateHeartbeat(ctx, sentinelID, heartbeat); err != nil {
		return fmt.Errorf("failed to update heartbeat: %w", err)
	}

	// 如果版本不同，更新版本
	if req.Version != "" && req.Version != sentinel.Version {
		sentinel.Version = req.Version
		if err := s.sentinelRepo.Update(ctx, sentinel); err != nil {
			return fmt.Errorf("failed to update version: %w", err)
		}
	}

	return nil
}

func (s *sentinelService) Get(ctx context.Context, id uint) (*model.Sentinel, error) {
	sentinel, err := s.sentinelRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("sentinel not found")
		}
		return nil, fmt.Errorf("failed to get sentinel: %w", err)
	}
	return sentinel, nil
}

func (s *sentinelService) List(ctx context.Context, req *ListSentinelRequest) ([]*model.Sentinel, int64, error) {
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

	filter := &repository.SentinelFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Status:   req.Status,
		Region:   req.Region,
	}

	return s.sentinelRepo.List(ctx, filter)
}

func (s *sentinelService) Delete(ctx context.Context, id uint) error {
	return s.sentinelRepo.Delete(ctx, id)
}

func (s *sentinelService) Control(ctx context.Context, sentinelID string, action string) error {
	// TODO: 实现远程控制逻辑
	// 这里需要通过某种机制（如 WebSocket、消息队列）向 Sentinel 发送控制命令
	return fmt.Errorf("control action not implemented: %s", action)
}

// generateAPIToken 生成 API Token
func generateAPIToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "sentinel_" + hex.EncodeToString(bytes), nil
}

