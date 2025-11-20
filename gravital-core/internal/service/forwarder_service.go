package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/celestial/gravital-core/internal/forwarder"
	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/pkg/config"
	"github.com/celestial/gravital-core/internal/repository"
	"go.uber.org/zap"
)

// ForwarderService 转发器服务接口
type ForwarderService interface {
	// 配置管理
	CreateForwarder(ctx context.Context, config *model.ForwarderConfig) error
	UpdateForwarder(ctx context.Context, config *model.ForwarderConfig) error
	DeleteForwarder(ctx context.Context, name string) error
	GetForwarder(ctx context.Context, name string) (*model.ForwarderConfig, error)
	ListForwarders(ctx context.Context, enabled *bool) ([]*model.ForwarderConfig, error)

	// 数据转发
	IngestMetrics(ctx context.Context, metrics []*forwarder.Metric) error

	// 统计信息
	GetForwarderStats(ctx context.Context, name string) (map[string]interface{}, error)
	GetAllStats(ctx context.Context) (map[string]interface{}, error)

	// 测试连接
	TestConnection(ctx context.Context, config *model.ForwarderConfig) (*ForwarderTestConnectionResult, error)

	// 生命周期
	Start() error
	Stop() error
	ReloadConfig() error
}

// ForwarderTestConnectionResult 转发器测试连接结果
type ForwarderTestConnectionResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Latency int64  `json:"latency_ms,omitempty"`
}

type forwarderService struct {
	repo    repository.ForwarderRepository
	manager *forwarder.Manager
	config  *config.Config
	logger  *zap.Logger
}

// NewForwarderService 创建转发器服务
func NewForwarderService(
	repo repository.ForwarderRepository,
	cfg *config.Config,
	logger *zap.Logger,
) ForwarderService {
	managerConfig := &forwarder.ManagerConfig{
		BufferSize:    cfg.Forwarder.BufferSize,
		BatchSize:     cfg.Forwarder.BatchSize,
		FlushInterval: cfg.Forwarder.FlushInterval,
		MaxRetries:    cfg.Forwarder.MaxRetries,
		RetryInterval: cfg.Forwarder.RetryInterval,
	}

	return &forwarderService{
		repo:    repo,
		manager: forwarder.NewManager(managerConfig, logger),
		config:  cfg,
		logger:  logger,
	}
}

// Start 启动转发服务
func (s *forwarderService) Start() error {
	// 从配置文件加载转发器
	for _, targetConfig := range s.config.Forwarder.Targets {
		if !targetConfig.Enabled {
			continue
		}

		fwdConfig := &forwarder.ForwarderConfig{
			Name:          targetConfig.Name,
			Type:          forwarder.ForwarderType(targetConfig.Type),
			Enabled:       targetConfig.Enabled,
			Endpoint:      targetConfig.Endpoint,
			DSN:           targetConfig.DSN,
			Table:         targetConfig.Table,
			Username:      targetConfig.Username,
			Password:      targetConfig.Password,
			Timeout:       targetConfig.Timeout,
			BatchSize:     targetConfig.BatchSize,
			FlushInterval: s.config.Forwarder.FlushInterval,
			MaxRetries:    s.config.Forwarder.MaxRetries,
			RetryInterval: s.config.Forwarder.RetryInterval,
		}

		if err := s.createForwarderInstance(fwdConfig); err != nil {
			s.logger.Error("Failed to create forwarder from config",
				zap.String("name", targetConfig.Name),
				zap.Error(err))
			continue
		}
	}

	// 从数据库加载转发器
	ctx := context.Background()
	enabled := true
	dbConfigs, err := s.repo.List(ctx, &enabled)
	if err != nil {
		s.logger.Warn("Failed to load forwarders from database", zap.Error(err))
	} else {
		for _, dbConfig := range dbConfigs {
			fwdConfig := s.modelToForwarderConfig(dbConfig)
			if err := s.createForwarderInstance(fwdConfig); err != nil {
				s.logger.Error("Failed to create forwarder from database",
					zap.String("name", dbConfig.Name),
					zap.Error(err))
			}
		}
	}

	// 启动管理器
	s.manager.Start()
	s.logger.Info("Forwarder service started")

	return nil
}

// Stop 停止转发服务
func (s *forwarderService) Stop() error {
	s.manager.Stop()
	s.logger.Info("Forwarder service stopped")
	return nil
}

// ReloadConfig 重新加载配置
func (s *forwarderService) ReloadConfig() error {
	// 停止现有管理器
	s.manager.Stop()

	// 重新创建管理器
	managerConfig := &forwarder.ManagerConfig{
		BufferSize:    s.config.Forwarder.BufferSize,
		BatchSize:     s.config.Forwarder.BatchSize,
		FlushInterval: s.config.Forwarder.FlushInterval,
		MaxRetries:    s.config.Forwarder.MaxRetries,
		RetryInterval: s.config.Forwarder.RetryInterval,
	}
	s.manager = forwarder.NewManager(managerConfig, s.logger)

	// 重新启动
	return s.Start()
}

// CreateForwarder 创建转发器
func (s *forwarderService) CreateForwarder(ctx context.Context, config *model.ForwarderConfig) error {
	// 保存到数据库
	if err := s.repo.Create(ctx, config); err != nil {
		return err
	}

	// 如果启用，创建实例
	if config.Enabled {
		fwdConfig := s.modelToForwarderConfig(config)
		if err := s.createForwarderInstance(fwdConfig); err != nil {
			return fmt.Errorf("failed to create forwarder instance: %w", err)
		}
	}

	s.logger.Info("Created forwarder", zap.String("name", config.Name))
	return nil
}

// UpdateForwarder 更新转发器
func (s *forwarderService) UpdateForwarder(ctx context.Context, config *model.ForwarderConfig) error {
	// 更新数据库
	if err := s.repo.Update(ctx, config); err != nil {
		return err
	}

	// 移除旧实例
	if err := s.manager.RemoveForwarder(config.Name); err != nil {
		s.logger.Warn("Failed to remove old forwarder instance",
			zap.String("name", config.Name),
			zap.Error(err))
	}

	// 如果启用，创建新实例
	if config.Enabled {
		fwdConfig := s.modelToForwarderConfig(config)
		if err := s.createForwarderInstance(fwdConfig); err != nil {
			return fmt.Errorf("failed to create forwarder instance: %w", err)
		}
	}

	s.logger.Info("Updated forwarder", zap.String("name", config.Name))
	return nil
}

// DeleteForwarder 删除转发器
func (s *forwarderService) DeleteForwarder(ctx context.Context, name string) error {
	// 从数据库删除
	if err := s.repo.Delete(ctx, name); err != nil {
		return err
	}

	// 移除实例
	if err := s.manager.RemoveForwarder(name); err != nil {
		s.logger.Warn("Failed to remove forwarder instance",
			zap.String("name", name),
			zap.Error(err))
	}

	s.logger.Info("Deleted forwarder", zap.String("name", name))
	return nil
}

// GetForwarder 获取转发器
func (s *forwarderService) GetForwarder(ctx context.Context, name string) (*model.ForwarderConfig, error) {
	return s.repo.GetByName(ctx, name)
}

// ListForwarders 列出转发器
func (s *forwarderService) ListForwarders(ctx context.Context, enabled *bool) ([]*model.ForwarderConfig, error) {
	return s.repo.List(ctx, enabled)
}

// IngestMetrics 接收并转发指标数据
func (s *forwarderService) IngestMetrics(ctx context.Context, metrics []*forwarder.Metric) error {
	if len(metrics) == 0 {
		return nil
	}

	// 转发到管理器
	if err := s.manager.ForwardBatch(metrics); err != nil {
		return fmt.Errorf("failed to forward metrics: %w", err)
	}

	s.logger.Info("Queued metrics for forwarding", zap.Int("count", len(metrics)))
	return nil
}

// GetForwarderStats 获取转发器统计
func (s *forwarderService) GetForwarderStats(ctx context.Context, name string) (map[string]interface{}, error) {
	// 获取实时统计
	allStats := s.manager.GetStats()
	stats, exists := allStats[name]
	if !exists {
		return nil, fmt.Errorf("forwarder not found: %s", name)
	}

	// 获取历史统计
	historyStats, err := s.repo.GetStats(ctx, name, 100)
	if err != nil {
		s.logger.Warn("Failed to get history stats", zap.Error(err))
	}

	return map[string]interface{}{
		"name":          name,
		"current":       stats,
		"history":       historyStats,
		"buffer_status": s.getBufferStatus(),
	}, nil
}

// GetAllStats 获取所有转发器统计
func (s *forwarderService) GetAllStats(ctx context.Context) (map[string]interface{}, error) {
	allStats := s.manager.GetStats()
	forwarders := s.manager.ListForwarders()

	result := make(map[string]interface{})
	for _, fwd := range forwarders {
		name := fwd.Name()
		stats := allStats[name]

		result[name] = map[string]interface{}{
			"type":    fwd.Type(),
			"enabled": fwd.IsEnabled(),
			"stats":   stats,
		}
	}

	result["buffer_status"] = s.getBufferStatus()
	result["total_forwarders"] = len(forwarders)

	return result, nil
}

// createForwarderInstance 创建转发器实例
func (s *forwarderService) createForwarderInstance(config *forwarder.ForwarderConfig) error {
	var fwd forwarder.Forwarder
	var err error

	switch config.Type {
	case forwarder.ForwarderTypePrometheus:
		fwd, err = forwarder.NewPrometheusForwarder(config, s.logger)
	case forwarder.ForwarderTypeVictoriaMetrics:
		fwd, err = forwarder.NewVictoriaMetricsForwarder(config, s.logger)
	case forwarder.ForwarderTypeClickHouse:
		fwd, err = forwarder.NewClickHouseForwarder(config, s.logger)
	default:
		return fmt.Errorf("unsupported forwarder type: %s", config.Type)
	}

	if err != nil {
		return err
	}

	return s.manager.AddForwarder(fwd)
}

// modelToForwarderConfig 将模型转换为转发器配置
func (s *forwarderService) modelToForwarderConfig(model *model.ForwarderConfig) *forwarder.ForwarderConfig {
	config := &forwarder.ForwarderConfig{
		Name:          model.Name,
		Type:          forwarder.ForwarderType(model.Type),
		Enabled:       model.Enabled,
		Endpoint:      model.Endpoint,
		BatchSize:     model.BatchSize,
		FlushInterval: time.Duration(model.FlushInterval) * time.Second,
		MaxRetries:    model.RetryTimes,
		Timeout:       time.Duration(model.TimeoutSeconds) * time.Second,
	}

	// 解析认证配置
	if len(model.AuthConfig) > 0 {
		var authConfig map[string]interface{}
		authBytes, err := json.Marshal(model.AuthConfig)
		if err == nil {
			if err := json.Unmarshal(authBytes, &authConfig); err == nil {
				if username, ok := authConfig["username"].(string); ok {
					config.Username = username
				}
				if password, ok := authConfig["password"].(string); ok {
					config.Password = password
				}
				if dsn, ok := authConfig["dsn"].(string); ok {
					config.DSN = dsn
				}
				if table, ok := authConfig["table"].(string); ok {
					config.Table = table
				}
			}
		}
	}

	return config
}

// getBufferStatus 获取缓冲区状态
func (s *forwarderService) getBufferStatus() map[string]interface{} {
	used, capacity := s.manager.GetBufferStatus()
	return map[string]interface{}{
		"used":     used,
		"capacity": capacity,
		"usage":    float64(used) / float64(capacity) * 100,
	}
}

// TestConnection 测试转发器连接
func (s *forwarderService) TestConnection(ctx context.Context, config *model.ForwarderConfig) (*ForwarderTestConnectionResult, error) {
	startTime := time.Now()

	// 转换为转发器配置
	fwdConfig := s.modelToForwarderConfig(config)

	// 创建临时转发器实例进行测试
	var testForwarder forwarder.Forwarder
	var err error

	switch fwdConfig.Type {
	case forwarder.ForwarderTypePrometheus:
		testForwarder, err = forwarder.NewPrometheusForwarder(fwdConfig, s.logger)
	case forwarder.ForwarderTypeVictoriaMetrics:
		testForwarder, err = forwarder.NewVictoriaMetricsForwarder(fwdConfig, s.logger)
	case forwarder.ForwarderTypeClickHouse:
		testForwarder, err = forwarder.NewClickHouseForwarder(fwdConfig, s.logger)
	default:
		return &ForwarderTestConnectionResult{
			Success: false,
			Message: fmt.Sprintf("不支持的转发器类型: %s", config.Type),
		}, nil
	}

	if err != nil {
		return &ForwarderTestConnectionResult{
			Success: false,
			Message: fmt.Sprintf("创建转发器失败: %v", err),
		}, nil
	}
	defer testForwarder.Close()

	// 创建测试指标
	testMetrics := []*forwarder.Metric{
		{
			Name:      "test_connection",
			Value:     1.0,
			Type:      "gauge",
			Labels:    map[string]string{"test": "true"},
			Timestamp: time.Now().Unix(),
		},
	}

	// 尝试写入测试数据
	if err := testForwarder.Write(testMetrics); err != nil {
		latency := time.Since(startTime).Milliseconds()
		return &ForwarderTestConnectionResult{
			Success: false,
			Message: fmt.Sprintf("连接测试失败: %v", err),
			Latency: latency,
		}, nil
	}

	latency := time.Since(startTime).Milliseconds()
	return &ForwarderTestConnectionResult{
		Success: true,
		Message: "连接测试成功",
		Latency: latency,
	}, nil
}
