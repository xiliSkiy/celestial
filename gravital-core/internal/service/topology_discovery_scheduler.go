package service

import (
	"context"
	"time"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// TopologyDiscoveryScheduler 拓扑自动发现调度器
type TopologyDiscoveryScheduler struct {
	topologyRepo     repository.TopologyRepository
	discoveryService TopologyDiscoveryService
	db               *gorm.DB
	logger           *zap.Logger
	checkInterval    time.Duration // 检查间隔
	cleanupInterval  time.Duration // 清理间隔
	ticker           *time.Ticker
	cleanupTicker    *time.Ticker
	done             chan struct{}
	ctx              context.Context
	cancel           context.CancelFunc
}

// TopologyDiscoverySchedulerConfig 调度器配置
type TopologyDiscoverySchedulerConfig struct {
	CheckInterval   time.Duration // 检查间隔，默认 5 分钟
	CleanupInterval time.Duration // 清理间隔，默认 1 小时
}

// NewTopologyDiscoveryScheduler 创建拓扑自动发现调度器
func NewTopologyDiscoveryScheduler(
	topologyRepo repository.TopologyRepository,
	discoveryService TopologyDiscoveryService,
	db *gorm.DB,
	logger *zap.Logger,
	config *TopologyDiscoverySchedulerConfig,
) *TopologyDiscoveryScheduler {
	if config == nil {
		config = &TopologyDiscoverySchedulerConfig{
			CheckInterval:   5 * time.Minute,
			CleanupInterval: 1 * time.Hour,
		}
	}

	// 设置默认值
	if config.CheckInterval == 0 {
		config.CheckInterval = 5 * time.Minute
	}
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 1 * time.Hour
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &TopologyDiscoveryScheduler{
		topologyRepo:     topologyRepo,
		discoveryService: discoveryService,
		db:               db,
		logger:           logger,
		checkInterval:    config.CheckInterval,
		cleanupInterval:  config.CleanupInterval,
		done:             make(chan struct{}),
		ctx:              ctx,
		cancel:           cancel,
	}
}

// Start 启动调度器
func (s *TopologyDiscoveryScheduler) Start() {
	s.ticker = time.NewTicker(s.checkInterval)
	s.cleanupTicker = time.NewTicker(s.cleanupInterval)

	s.logger.Info("Topology discovery scheduler started",
		zap.Duration("check_interval", s.checkInterval),
		zap.Duration("cleanup_interval", s.cleanupInterval))

	// 启动自动发现检查循环
	go func() {
		// 立即执行一次检查
		s.checkAndDiscoverTopologies()

		for {
			select {
			case <-s.ticker.C:
				s.checkAndDiscoverTopologies()
			case <-s.done:
				s.logger.Info("Topology discovery scheduler stopped")
				return
			}
		}
	}()

	// 启动清理循环
	go func() {
		// 立即执行一次清理
		s.cleanupStaleNeighbors()

		for {
			select {
			case <-s.cleanupTicker.C:
				s.cleanupStaleNeighbors()
			case <-s.done:
				return
			}
		}
	}()
}

// Stop 停止调度器
func (s *TopologyDiscoveryScheduler) Stop() {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.cleanupTicker != nil {
		s.cleanupTicker.Stop()
	}
	s.cancel()
	close(s.done)
	s.logger.Info("Topology discovery scheduler stopped")
}

// checkAndDiscoverTopologies 检查并触发自动发现
func (s *TopologyDiscoveryScheduler) checkAndDiscoverTopologies() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	// 获取所有启用自动发现的拓扑
	topologies, err := s.getAutoDiscoveryTopologies(ctx)
	if err != nil {
		s.logger.Error("Failed to get auto-discovery topologies", zap.Error(err))
		return
	}

	if len(topologies) == 0 {
		s.logger.Debug("No topologies with auto-discovery enabled")
		return
	}

	s.logger.Debug("Checking topologies for auto-discovery",
		zap.Int("count", len(topologies)))

	now := time.Now()
	discoveredCount := 0

	for _, topology := range topologies {
		// 检查是否需要触发发现
		if s.shouldDiscover(topology, now) {
			s.logger.Info("Triggering auto-discovery for topology",
				zap.Uint("topology_id", topology.ID),
				zap.String("name", topology.Name))

			// 异步执行发现，避免阻塞
			go func(topoID uint) {
				discoverCtx, discoverCancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer discoverCancel()

				result, err := s.discoveryService.DiscoverTopology(discoverCtx, topoID)
				if err != nil {
					s.logger.Error("Auto-discovery failed",
						zap.Uint("topology_id", topoID),
						zap.Error(err))
					return
				}

				s.logger.Info("Auto-discovery completed",
					zap.Uint("topology_id", topoID),
					zap.Int("discovered_nodes", result.DiscoveredNodes),
					zap.Int("discovered_links", result.DiscoveredLinks))
			}(topology.ID)

			discoveredCount++
		}
	}

	if discoveredCount > 0 {
		s.logger.Info("Triggered auto-discovery",
			zap.Int("count", discoveredCount))
	}
}

// getAutoDiscoveryTopologies 获取启用自动发现的拓扑
func (s *TopologyDiscoveryScheduler) getAutoDiscoveryTopologies(ctx context.Context) ([]model.Topology, error) {
	var topologies []model.Topology
	err := s.db.WithContext(ctx).
		Where("is_auto_discovery = ?", true).
		Find(&topologies).Error
	return topologies, err
}

// shouldDiscover 判断是否需要触发发现
func (s *TopologyDiscoveryScheduler) shouldDiscover(topology model.Topology, now time.Time) bool {
	// 如果没有设置发现间隔，默认不触发
	if topology.DiscoveryInterval <= 0 {
		return false
	}

	// 如果从未发现过，立即触发
	if topology.LastDiscoveryAt == nil {
		return true
	}

	// 计算下次发现时间
	nextDiscoveryTime := topology.LastDiscoveryAt.Add(time.Duration(topology.DiscoveryInterval) * time.Second)

	// 如果已经超过发现间隔，触发发现
	return now.After(nextDiscoveryTime)
}

// cleanupStaleNeighbors 清理过期的 LLDP 邻居
func (s *TopologyDiscoveryScheduler) cleanupStaleNeighbors() {
	ctx, cancel := context.WithTimeout(s.ctx, 30*time.Second)
	defer cancel()

	s.logger.Debug("Starting cleanup of stale LLDP neighbors")

	if err := s.discoveryService.CleanupStaleNeighbors(ctx); err != nil {
		s.logger.Error("Failed to cleanup stale neighbors", zap.Error(err))
		return
	}

	s.logger.Debug("Completed cleanup of stale LLDP neighbors")
}
