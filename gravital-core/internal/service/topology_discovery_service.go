package service

import (
	"context"
	"fmt"
	"time"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
	"go.uber.org/zap"
)

// TopologyDiscoveryService 拓扑自动发现服务
type TopologyDiscoveryService interface {
	// 自动发现拓扑
	DiscoverTopology(ctx context.Context, topologyID uint) error
	// 清理过期的 LLDP 邻居
	CleanupStaleNeighbors(ctx context.Context) error
}

type topologyDiscoveryService struct {
	topologyRepo repository.TopologyRepository
	deviceRepo   repository.DeviceRepository
	logger       *zap.Logger
}

// NewTopologyDiscoveryService 创建拓扑自动发现服务实例
func NewTopologyDiscoveryService(
	topologyRepo repository.TopologyRepository,
	deviceRepo repository.DeviceRepository,
	logger *zap.Logger,
) TopologyDiscoveryService {
	return &topologyDiscoveryService{
		topologyRepo: topologyRepo,
		deviceRepo:   deviceRepo,
		logger:       logger,
	}
}

// DiscoverTopology 自动发现拓扑
func (s *topologyDiscoveryService) DiscoverTopology(ctx context.Context, topologyID uint) error {
	s.logger.Info("Starting topology discovery", zap.Uint("topology_id", topologyID))

	// 获取拓扑信息
	topology, err := s.topologyRepo.GetByID(ctx, topologyID)
	if err != nil {
		return fmt.Errorf("failed to get topology: %w", err)
	}

	if !topology.IsAutoDiscovery {
		return fmt.Errorf("topology auto discovery is disabled")
	}

	// 获取所有 LLDP 邻居信息
	neighbors, err := s.topologyRepo.GetAllLLDPNeighbors(ctx)
	if err != nil {
		return fmt.Errorf("failed to get LLDP neighbors: %w", err)
	}

	s.logger.Info("Found LLDP neighbors", zap.Int("count", len(neighbors)))

	// 构建设备映射（用于快速查找）
	deviceMap := make(map[string]*model.Device)
	devices, _, err := s.deviceRepo.List(ctx, &repository.DeviceFilter{})
	if err != nil {
		return fmt.Errorf("failed to list devices: %w", err)
	}

	for i := range devices {
		deviceMap[devices[i].DeviceID] = devices[i]
	}

	// 构建节点映射（已存在的节点）
	existingNodes := make(map[string]*model.TopologyNode)
	nodes, err := s.topologyRepo.GetNodesByTopologyID(ctx, topologyID)
	if err != nil {
		return fmt.Errorf("failed to get existing nodes: %w", err)
	}

	for i := range nodes {
		existingNodes[nodes[i].DeviceID] = &nodes[i]
	}

	// 构建链路映射（已存在的链路）
	existingLinks := make(map[string]*model.TopologyLink)
	links, err := s.topologyRepo.GetLinksByTopologyID(ctx, topologyID)
	if err != nil {
		return fmt.Errorf("failed to get existing links: %w", err)
	}

	for i := range links {
		key := fmt.Sprintf("%d-%d", links[i].SourceNodeID, links[i].TargetNodeID)
		existingLinks[key] = &links[i]
	}

	// 处理 LLDP 邻居，创建节点和链路
	discoveredNodes := 0
	discoveredLinks := 0

	for _, neighbor := range neighbors {
		// 查找本地设备
		localDevice, ok := deviceMap[neighbor.DeviceID]
		if !ok {
			s.logger.Warn("Local device not found", zap.String("device_id", neighbor.DeviceID))
			continue
		}

		// 查找或创建本地节点
		localNode, err := s.findOrCreateNode(ctx, topology, localDevice, existingNodes)
		if err != nil {
			s.logger.Error("Failed to find or create local node",
				zap.String("device_id", neighbor.DeviceID),
				zap.Error(err))
			continue
		}
		if localNode.ID == 0 {
			discoveredNodes++
		}

		// 匹配邻居设备
		neighborDevice := s.matchNeighborDevice(neighbor, deviceMap)
		if neighborDevice == nil {
			s.logger.Warn("Neighbor device not found",
				zap.String("neighbor_chassis_id", neighbor.NeighborChassisID),
				zap.String("neighbor_system_name", neighbor.NeighborSystemName))
			continue
		}

		// 查找或创建邻居节点
		neighborNode, err := s.findOrCreateNode(ctx, topology, neighborDevice, existingNodes)
		if err != nil {
			s.logger.Error("Failed to find or create neighbor node",
				zap.String("device_id", neighborDevice.DeviceID),
				zap.Error(err))
			continue
		}
		if neighborNode.ID == 0 {
			discoveredNodes++
		}

		// 创建链路
		linkKey := fmt.Sprintf("%d-%d", localNode.ID, neighborNode.ID)
		reverseLinkKey := fmt.Sprintf("%d-%d", neighborNode.ID, localNode.ID)

		if _, exists := existingLinks[linkKey]; !exists {
			if _, reverseExists := existingLinks[reverseLinkKey]; !reverseExists {
				// 创建新链路
				link := &model.TopologyLink{
					TopologyID:      topologyID,
					SourceNodeID:    localNode.ID,
					TargetNodeID:    neighborNode.ID,
					LinkType:        "physical",
					SourceInterface: neighbor.LocalInterface,
					TargetInterface: neighbor.NeighborPortID,
					Status:          "unknown",
					DiscoveredBy:    "lldp",
				}

				now := time.Now()
				link.DiscoveredAt = &now

				if err := s.topologyRepo.CreateLink(ctx, link); err != nil {
					s.logger.Error("Failed to create link",
						zap.Uint("source_node_id", localNode.ID),
						zap.Uint("target_node_id", neighborNode.ID),
						zap.Error(err))
					continue
				}

				discoveredLinks++
			}
		}
	}

	// 更新拓扑的最后发现时间
	now := time.Now()
	topology.LastDiscoveryAt = &now
	if err := s.topologyRepo.Update(ctx, topology); err != nil {
		s.logger.Error("Failed to update topology last discovery time", zap.Error(err))
	}

	s.logger.Info("Topology discovery completed",
		zap.Uint("topology_id", topologyID),
		zap.Int("discovered_nodes", discoveredNodes),
		zap.Int("discovered_links", discoveredLinks))

	return nil
}

// findOrCreateNode 查找或创建节点
func (s *topologyDiscoveryService) findOrCreateNode(
	ctx context.Context,
	topology *model.Topology,
	device *model.Device,
	existingNodes map[string]*model.TopologyNode,
) (*model.TopologyNode, error) {
	// 检查节点是否已存在
	if node, exists := existingNodes[device.DeviceID]; exists {
		return node, nil
	}

	// 创建新节点
	// 从 connection_config 中提取 IP 和其他信息
	var ip string
	if device.ConnectionConfig != nil {
		if ipVal, ok := device.ConnectionConfig["host"].(string); ok {
			ip = ipVal
		}
	}
	
	node := &model.TopologyNode{
		TopologyID: topology.ID,
		DeviceID:   device.DeviceID,
		NodeType:   "device",
		Label:      device.Name,
		Layer:      0, // 默认层级，可以后续调整
		Size:       40,
		Shape:      "circle",
		Properties: map[string]interface{}{
			"ip":          ip,
			"device_type": device.DeviceType,
			"status":      device.Status,
		},
	}

	if err := s.topologyRepo.CreateNode(ctx, node); err != nil {
		return nil, err
	}

	// 添加到已存在节点映射
	existingNodes[device.DeviceID] = node

	return node, nil
}

// matchNeighborDevice 匹配邻居设备
func (s *topologyDiscoveryService) matchNeighborDevice(
	neighbor model.LLDPNeighbor,
	deviceMap map[string]*model.Device,
) *model.Device {
	// 优先级 1: 通过 Chassis ID 匹配
	for _, device := range deviceMap {
		// 从 Labels 中查找 chassis_id
		if device.Labels != nil {
			if chassisID, ok := device.Labels["chassis_id"].(string); ok {
				if chassisID == neighbor.NeighborChassisID {
					return device
				}
			}
		}
	}

	// 优先级 2: 通过管理 IP 匹配
	if neighbor.NeighborMgmtAddr != "" {
		for _, device := range deviceMap {
			// 从 connection_config 中获取 host
			if device.ConnectionConfig != nil {
				if host, ok := device.ConnectionConfig["host"].(string); ok {
					if host == neighbor.NeighborMgmtAddr {
						return device
					}
				}
			}
		}
	}

	// 优先级 3: 通过系统名称模糊匹配
	if neighbor.NeighborSystemName != "" {
		for _, device := range deviceMap {
			if device.Name == neighbor.NeighborSystemName {
				return device
			}
		}
	}

	return nil
}

// CleanupStaleNeighbors 清理过期的 LLDP 邻居
func (s *topologyDiscoveryService) CleanupStaleNeighbors(ctx context.Context) error {
	// 清理 24 小时未更新的邻居信息
	ttl := 24 * 60 * 60 // 24 小时
	return s.topologyRepo.DeleteStaleLLDPNeighbors(ctx, ttl)
}

