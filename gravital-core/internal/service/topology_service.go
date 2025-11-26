package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/celestial/gravital-core/internal/model"
	"github.com/celestial/gravital-core/internal/repository"
	"go.uber.org/zap"
)

// TopologyService 拓扑服务接口
type TopologyService interface {
	// 拓扑管理
	CreateTopology(ctx context.Context, req *CreateTopologyRequest) (*model.Topology, error)
	GetTopology(ctx context.Context, id uint) (*TopologyDetailResponse, error)
	UpdateTopology(ctx context.Context, id uint, req *UpdateTopologyRequest) error
	DeleteTopology(ctx context.Context, id uint) error
	ListTopologies(ctx context.Context, req *ListTopologyRequest) (*ListTopologyResponse, error)

	// 节点管理
	AddNode(ctx context.Context, topologyID uint, req *AddNodeRequest) (*model.TopologyNode, error)
	UpdateNodePosition(ctx context.Context, topologyID, nodeID uint, x, y float64) error
	BatchUpdateNodes(ctx context.Context, topologyID uint, req *BatchUpdateNodesRequest) error
	DeleteNode(ctx context.Context, topologyID, nodeID uint) error

	// 链路管理
	AddLink(ctx context.Context, topologyID uint, req *AddLinkRequest) (*model.TopologyLink, error)
	UpdateLinkStatus(ctx context.Context, topologyID, linkID uint, req *UpdateLinkStatusRequest) error
	DeleteLink(ctx context.Context, topologyID, linkID uint) error

	// 布局
	ApplyLayout(ctx context.Context, topologyID uint, req *ApplyLayoutRequest) (*ApplyLayoutResponse, error)

	// 分析
	AnalyzePath(ctx context.Context, topologyID uint, req *PathAnalysisRequest) (*PathAnalysisResponse, error)
	AnalyzeImpact(ctx context.Context, topologyID uint, req *ImpactAnalysisRequest) (*ImpactAnalysisResponse, error)

	// 版本管理
	CreateSnapshot(ctx context.Context, topologyID uint, description string, userID uint) error
	GetVersions(ctx context.Context, topologyID uint) ([]model.TopologyVersion, error)
	RestoreVersion(ctx context.Context, topologyID uint, version int) error

	// LLDP 邻居
	UpsertLLDPNeighbor(ctx context.Context, neighbor *model.LLDPNeighbor) error
	GetLLDPNeighbors(ctx context.Context, deviceID string) ([]model.LLDPNeighbor, error)
	
	// 自动发现
	DiscoverTopology(ctx context.Context, topologyID uint) (*DiscoverTopologyResponse, error)
}

type topologyService struct {
	topologyRepo        repository.TopologyRepository
	deviceRepo          repository.DeviceRepository
	discoveryService    TopologyDiscoveryService
	logger              *zap.Logger
}

// NewTopologyService 创建拓扑服务实例
func NewTopologyService(
	topologyRepo repository.TopologyRepository,
	deviceRepo repository.DeviceRepository,
	discoveryService TopologyDiscoveryService,
	logger *zap.Logger,
) TopologyService {
	return &topologyService{
		topologyRepo:     topologyRepo,
		deviceRepo:       deviceRepo,
		discoveryService: discoveryService,
		logger:           logger,
	}
}

// CreateTopologyRequest 创建拓扑请求
type CreateTopologyRequest struct {
	Name              string                 `json:"name" binding:"required"`
	Description       string                 `json:"description"`
	Type              string                 `json:"type" binding:"required"` // physical, logical, custom
	Scope             string                 `json:"scope"`
	LayoutType        string                 `json:"layout_type"`
	IsAutoDiscovery   bool                   `json:"is_auto_discovery"`
	DiscoveryInterval int                    `json:"discovery_interval"`
	LayoutConfig      map[string]interface{} `json:"layout_config"`
	ViewConfig        map[string]interface{} `json:"view_config"`
}

// UpdateTopologyRequest 更新拓扑请求
type UpdateTopologyRequest struct {
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	LayoutType        string                 `json:"layout_type"`
	IsAutoDiscovery   bool                   `json:"is_auto_discovery"`
	DiscoveryInterval int                    `json:"discovery_interval"`
	LayoutConfig      map[string]interface{} `json:"layout_config"`
	ViewConfig        map[string]interface{} `json:"view_config"`
}

// ListTopologyRequest 列表请求
type ListTopologyRequest struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	Type     string `json:"type"`
	Scope    string `json:"scope"`
	Keyword  string `json:"keyword"`
}

// ListTopologyResponse 列表响应
type ListTopologyResponse struct {
	Total int64             `json:"total"`
	Items []TopologyListItem `json:"items"`
}

// TopologyListItem 拓扑列表项
type TopologyListItem struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Scope     string    `json:"scope"`
	NodeCount int       `json:"node_count"`
	LinkCount int       `json:"link_count"`
	CreatedAt time.Time `json:"created_at"`
}

// TopologyDetailResponse 拓扑详情响应
type TopologyDetailResponse struct {
	model.Topology
	NodeCount int `json:"node_count"`
	LinkCount int `json:"link_count"`
}

// AddNodeRequest 添加节点请求
type AddNodeRequest struct {
	DeviceID  string                 `json:"device_id" binding:"required"`
	NodeType  string                 `json:"node_type" binding:"required"`
	Label     string                 `json:"label"`
	PositionX float64                `json:"position_x"`
	PositionY float64                `json:"position_y"`
	Layer     int                    `json:"layer"`
	Icon      string                 `json:"icon"`
	Size      int                    `json:"size"`
	Color     string                 `json:"color"`
	Shape     string                 `json:"shape"`
	Properties map[string]interface{} `json:"properties"`
}

// BatchUpdateNodesRequest 批量更新节点请求
type BatchUpdateNodesRequest struct {
	Nodes []NodeUpdate `json:"nodes" binding:"required"`
}

// NodeUpdate 节点更新
type NodeUpdate struct {
	ID        uint    `json:"id" binding:"required"`
	PositionX float64 `json:"position_x"`
	PositionY float64 `json:"position_y"`
}

// AddLinkRequest 添加链路请求
type AddLinkRequest struct {
	SourceNodeID    uint                   `json:"source_node_id" binding:"required"`
	TargetNodeID    uint                   `json:"target_node_id" binding:"required"`
	LinkType        string                 `json:"link_type" binding:"required"`
	SourceInterface string                 `json:"source_interface"`
	TargetInterface string                 `json:"target_interface"`
	Bandwidth       int64                  `json:"bandwidth"`
	Protocol        string                 `json:"protocol"`
	Label           string                 `json:"label"`
	Properties      map[string]interface{} `json:"properties"`
}

// UpdateLinkStatusRequest 更新链路状态请求
type UpdateLinkStatusRequest struct {
	Status      string  `json:"status"`
	Utilization float64 `json:"utilization"`
	Latency     float64 `json:"latency"`
	PacketLoss  float64 `json:"packet_loss"`
}

// ApplyLayoutRequest 应用布局请求
type ApplyLayoutRequest struct {
	LayoutType string                 `json:"layout_type" binding:"required"`
	Options    map[string]interface{} `json:"options"`
}

// ApplyLayoutResponse 应用布局响应
type ApplyLayoutResponse struct {
	Nodes []NodePosition `json:"nodes"`
}

// NodePosition 节点位置
type NodePosition struct {
	ID       uint    `json:"id"`
	Position Position `json:"position"`
}

// Position 位置
type Position struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

// PathAnalysisRequest 路径分析请求
type PathAnalysisRequest struct {
	SourceNodeID uint   `json:"source_node_id" binding:"required"`
	TargetNodeID uint   `json:"target_node_id" binding:"required"`
	Algorithm    string `json:"algorithm"` // shortest, all
}

// PathAnalysisResponse 路径分析响应
type PathAnalysisResponse struct {
	Paths []Path `json:"paths"`
}

// Path 路径
type Path struct {
	Nodes        []uint  `json:"nodes"`
	Links        []uint  `json:"links"`
	HopCount     int     `json:"hop_count"`
	TotalLatency float64 `json:"total_latency"`
}

// ImpactAnalysisRequest 影响分析请求
type ImpactAnalysisRequest struct {
	NodeID   uint   `json:"node_id" binding:"required"`
	Scenario string `json:"scenario"` // failure, maintenance
}

// ImpactAnalysisResponse 影响分析响应
type ImpactAnalysisResponse struct {
	AffectedNodes  []uint `json:"affected_nodes"`
	AffectedLinks  []uint `json:"affected_links"`
	IsolatedNodes  []uint `json:"isolated_nodes"`
	ImpactLevel    string `json:"impact_level"` // low, medium, high
}

// CreateTopology 创建拓扑
func (s *topologyService) CreateTopology(ctx context.Context, req *CreateTopologyRequest) (*model.Topology, error) {
	topology := &model.Topology{
		Name:              req.Name,
		Description:       req.Description,
		Type:              req.Type,
		Scope:             req.Scope,
		LayoutType:        req.LayoutType,
		IsAutoDiscovery:   req.IsAutoDiscovery,
		DiscoveryInterval: req.DiscoveryInterval,
		LayoutConfig:      req.LayoutConfig,
		ViewConfig:        req.ViewConfig,
		Version:           1,
	}

	if err := s.topologyRepo.Create(ctx, topology); err != nil {
		s.logger.Error("Failed to create topology", zap.Error(err))
		return nil, err
	}

	return topology, nil
}

// GetTopology 获取拓扑详情
func (s *topologyService) GetTopology(ctx context.Context, id uint) (*TopologyDetailResponse, error) {
	topology, err := s.topologyRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return &TopologyDetailResponse{
		Topology:  *topology,
		NodeCount: len(topology.Nodes),
		LinkCount: len(topology.Links),
	}, nil
}

// UpdateTopology 更新拓扑
func (s *topologyService) UpdateTopology(ctx context.Context, id uint, req *UpdateTopologyRequest) error {
	topology, err := s.topologyRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if req.Name != "" {
		topology.Name = req.Name
	}
	if req.Description != "" {
		topology.Description = req.Description
	}
	if req.LayoutType != "" {
		topology.LayoutType = req.LayoutType
	}
	topology.IsAutoDiscovery = req.IsAutoDiscovery
	topology.DiscoveryInterval = req.DiscoveryInterval
	if req.LayoutConfig != nil {
		topology.LayoutConfig = req.LayoutConfig
	}
	if req.ViewConfig != nil {
		topology.ViewConfig = req.ViewConfig
	}

	return s.topologyRepo.Update(ctx, topology)
}

// DeleteTopology 删除拓扑
func (s *topologyService) DeleteTopology(ctx context.Context, id uint) error {
	return s.topologyRepo.Delete(ctx, id)
}

// ListTopologies 获取拓扑列表
func (s *topologyService) ListTopologies(ctx context.Context, req *ListTopologyRequest) (*ListTopologyResponse, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	filter := repository.TopologyFilter{
		Page:     req.Page,
		PageSize: req.PageSize,
		Type:     req.Type,
		Scope:    req.Scope,
		Keyword:  req.Keyword,
	}

	topologies, total, err := s.topologyRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]TopologyListItem, 0, len(topologies))
	for _, t := range topologies {
		// 获取节点和链路数量
		nodes, _ := s.topologyRepo.GetNodesByTopologyID(ctx, t.ID)
		links, _ := s.topologyRepo.GetLinksByTopologyID(ctx, t.ID)

		items = append(items, TopologyListItem{
			ID:        t.ID,
			Name:      t.Name,
			Type:      t.Type,
			Scope:     t.Scope,
			NodeCount: len(nodes),
			LinkCount: len(links),
			CreatedAt: t.CreatedAt,
		})
	}

	return &ListTopologyResponse{
		Total: total,
		Items: items,
	}, nil
}

// AddNode 添加节点
func (s *topologyService) AddNode(ctx context.Context, topologyID uint, req *AddNodeRequest) (*model.TopologyNode, error) {
	node := &model.TopologyNode{
		TopologyID: topologyID,
		DeviceID:   req.DeviceID,
		NodeType:   req.NodeType,
		Label:      req.Label,
		PositionX:  req.PositionX,
		PositionY:  req.PositionY,
		Layer:      req.Layer,
		Icon:       req.Icon,
		Size:       req.Size,
		Color:      req.Color,
		Shape:      req.Shape,
		Properties: req.Properties,
	}

	if err := s.topologyRepo.CreateNode(ctx, node); err != nil {
		s.logger.Error("Failed to create node", zap.Error(err))
		return nil, err
	}

	return node, nil
}

// UpdateNodePosition 更新节点位置
func (s *topologyService) UpdateNodePosition(ctx context.Context, topologyID, nodeID uint, x, y float64) error {
	node, err := s.topologyRepo.GetNodeByID(ctx, nodeID)
	if err != nil {
		return err
	}

	if node.TopologyID != topologyID {
		return fmt.Errorf("node does not belong to topology")
	}

	node.PositionX = x
	node.PositionY = y

	return s.topologyRepo.UpdateNode(ctx, node)
}

// BatchUpdateNodes 批量更新节点
func (s *topologyService) BatchUpdateNodes(ctx context.Context, topologyID uint, req *BatchUpdateNodesRequest) error {
	nodes := make([]model.TopologyNode, 0, len(req.Nodes))

	for _, update := range req.Nodes {
		node, err := s.topologyRepo.GetNodeByID(ctx, update.ID)
		if err != nil {
			return err
		}

		if node.TopologyID != topologyID {
			return fmt.Errorf("node %d does not belong to topology", update.ID)
		}

		node.PositionX = update.PositionX
		node.PositionY = update.PositionY
		nodes = append(nodes, *node)
	}

	return s.topologyRepo.BatchUpdateNodes(ctx, nodes)
}

// DeleteNode 删除节点
func (s *topologyService) DeleteNode(ctx context.Context, topologyID, nodeID uint) error {
	node, err := s.topologyRepo.GetNodeByID(ctx, nodeID)
	if err != nil {
		return err
	}

	if node.TopologyID != topologyID {
		return fmt.Errorf("node does not belong to topology")
	}

	return s.topologyRepo.DeleteNode(ctx, nodeID)
}

// AddLink 添加链路
func (s *topologyService) AddLink(ctx context.Context, topologyID uint, req *AddLinkRequest) (*model.TopologyLink, error) {
	link := &model.TopologyLink{
		TopologyID:      topologyID,
		SourceNodeID:    req.SourceNodeID,
		TargetNodeID:    req.TargetNodeID,
		LinkType:        req.LinkType,
		SourceInterface: req.SourceInterface,
		TargetInterface: req.TargetInterface,
		Bandwidth:       req.Bandwidth,
		Protocol:        req.Protocol,
		Label:           req.Label,
		Properties:      req.Properties,
		Status:          "unknown",
		DiscoveredBy:    "manual",
	}

	now := time.Now()
	link.DiscoveredAt = &now

	if err := s.topologyRepo.CreateLink(ctx, link); err != nil {
		s.logger.Error("Failed to create link", zap.Error(err))
		return nil, err
	}

	return link, nil
}

// UpdateLinkStatus 更新链路状态
func (s *topologyService) UpdateLinkStatus(ctx context.Context, topologyID, linkID uint, req *UpdateLinkStatusRequest) error {
	link, err := s.topologyRepo.GetLinkByID(ctx, linkID)
	if err != nil {
		return err
	}

	if link.TopologyID != topologyID {
		return fmt.Errorf("link does not belong to topology")
	}

	link.Status = req.Status
	link.Utilization = req.Utilization
	link.Latency = req.Latency
	link.PacketLoss = req.PacketLoss

	return s.topologyRepo.UpdateLink(ctx, link)
}

// DeleteLink 删除链路
func (s *topologyService) DeleteLink(ctx context.Context, topologyID, linkID uint) error {
	link, err := s.topologyRepo.GetLinkByID(ctx, linkID)
	if err != nil {
		return err
	}

	if link.TopologyID != topologyID {
		return fmt.Errorf("link does not belong to topology")
	}

	return s.topologyRepo.DeleteLink(ctx, linkID)
}

// ApplyLayout 应用布局（简化版，实际布局计算在前端完成）
func (s *topologyService) ApplyLayout(ctx context.Context, topologyID uint, req *ApplyLayoutRequest) (*ApplyLayoutResponse, error) {
	// 这里只是更新拓扑的布局类型和配置
	topology, err := s.topologyRepo.GetByID(ctx, topologyID)
	if err != nil {
		return nil, err
	}

	topology.LayoutType = req.LayoutType
	topology.LayoutConfig = req.Options

	if err := s.topologyRepo.Update(ctx, topology); err != nil {
		return nil, err
	}

	// 返回当前节点位置（前端会根据布局类型重新计算）
	nodes, err := s.topologyRepo.GetNodesByTopologyID(ctx, topologyID)
	if err != nil {
		return nil, err
	}

	positions := make([]NodePosition, 0, len(nodes))
	for _, node := range nodes {
		positions = append(positions, NodePosition{
			ID: node.ID,
			Position: Position{
				X: node.PositionX,
				Y: node.PositionY,
			},
		})
	}

	return &ApplyLayoutResponse{
		Nodes: positions,
	}, nil
}

// AnalyzePath 路径分析（简化版）
func (s *topologyService) AnalyzePath(ctx context.Context, topologyID uint, req *PathAnalysisRequest) (*PathAnalysisResponse, error) {
	// 获取所有链路
	links, err := s.topologyRepo.GetLinksByTopologyID(ctx, topologyID)
	if err != nil {
		return nil, err
	}

	// 构建邻接表
	graph := make(map[uint][]uint)
	linkMap := make(map[string]uint)

	for _, link := range links {
		graph[link.SourceNodeID] = append(graph[link.SourceNodeID], link.TargetNodeID)
		graph[link.TargetNodeID] = append(graph[link.TargetNodeID], link.SourceNodeID)
		linkMap[fmt.Sprintf("%d-%d", link.SourceNodeID, link.TargetNodeID)] = link.ID
		linkMap[fmt.Sprintf("%d-%d", link.TargetNodeID, link.SourceNodeID)] = link.ID
	}

	// 使用 BFS 查找最短路径
	path := s.findShortestPath(graph, req.SourceNodeID, req.TargetNodeID)
	if len(path) == 0 {
		return &PathAnalysisResponse{Paths: []Path{}}, nil
	}

	// 构建链路列表
	linkIDs := make([]uint, 0)
	for i := 0; i < len(path)-1; i++ {
		key := fmt.Sprintf("%d-%d", path[i], path[i+1])
		if linkID, ok := linkMap[key]; ok {
			linkIDs = append(linkIDs, linkID)
		}
	}

	return &PathAnalysisResponse{
		Paths: []Path{
			{
				Nodes:        path,
				Links:        linkIDs,
				HopCount:     len(path) - 1,
				TotalLatency: 0, // TODO: 计算总延迟
			},
		},
	}, nil
}

// findShortestPath 使用 BFS 查找最短路径
func (s *topologyService) findShortestPath(graph map[uint][]uint, start, end uint) []uint {
	if start == end {
		return []uint{start}
	}

	visited := make(map[uint]bool)
	parent := make(map[uint]uint)
	queue := []uint{start}
	visited[start] = true

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		if current == end {
			// 回溯构建路径
			path := []uint{}
			for node := end; node != start; node = parent[node] {
				path = append([]uint{node}, path...)
			}
			path = append([]uint{start}, path...)
			return path
		}

		for _, neighbor := range graph[current] {
			if !visited[neighbor] {
				visited[neighbor] = true
				parent[neighbor] = current
				queue = append(queue, neighbor)
			}
		}
	}

	return []uint{} // 未找到路径
}

// AnalyzeImpact 影响分析（简化版）
func (s *topologyService) AnalyzeImpact(ctx context.Context, topologyID uint, req *ImpactAnalysisRequest) (*ImpactAnalysisResponse, error) {
	// 获取所有节点和链路
	nodes, err := s.topologyRepo.GetNodesByTopologyID(ctx, topologyID)
	if err != nil {
		return nil, err
	}

	links, err := s.topologyRepo.GetLinksByTopologyID(ctx, topologyID)
	if err != nil {
		return nil, err
	}

	// 构建邻接表（排除故障节点）
	graph := make(map[uint][]uint)
	affectedLinks := make([]uint, 0)

	for _, link := range links {
		if link.SourceNodeID == req.NodeID || link.TargetNodeID == req.NodeID {
			affectedLinks = append(affectedLinks, link.ID)
		} else {
			graph[link.SourceNodeID] = append(graph[link.SourceNodeID], link.TargetNodeID)
			graph[link.TargetNodeID] = append(graph[link.TargetNodeID], link.SourceNodeID)
		}
	}

	// 查找孤立节点（使用 DFS）
	visited := make(map[uint]bool)
	isolatedNodes := make([]uint, 0)

	for _, node := range nodes {
		if node.ID == req.NodeID {
			continue
		}
		if !visited[node.ID] {
			// 从该节点开始 DFS
			component := s.dfs(graph, node.ID, visited)
			// 如果连通分量只有一个节点，说明它是孤立的
			if len(component) == 1 {
				isolatedNodes = append(isolatedNodes, component[0])
			}
		}
	}

	// 计算影响级别
	impactLevel := "low"
	affectedRatio := float64(len(affectedLinks)) / float64(len(links))
	if affectedRatio > 0.5 || len(isolatedNodes) > 0 {
		impactLevel = "high"
	} else if affectedRatio > 0.2 {
		impactLevel = "medium"
	}

	return &ImpactAnalysisResponse{
		AffectedNodes:  []uint{req.NodeID},
		AffectedLinks:  affectedLinks,
		IsolatedNodes:  isolatedNodes,
		ImpactLevel:    impactLevel,
	}, nil
}

// dfs 深度优先搜索
func (s *topologyService) dfs(graph map[uint][]uint, node uint, visited map[uint]bool) []uint {
	visited[node] = true
	component := []uint{node}

	for _, neighbor := range graph[node] {
		if !visited[neighbor] {
			component = append(component, s.dfs(graph, neighbor, visited)...)
		}
	}

	return component
}

// CreateSnapshot 创建快照
func (s *topologyService) CreateSnapshot(ctx context.Context, topologyID uint, description string, userID uint) error {
	topology, err := s.topologyRepo.GetByID(ctx, topologyID)
	if err != nil {
		return err
	}

	// 序列化拓扑数据
	snapshot, err := json.Marshal(topology)
	if err != nil {
		return err
	}

	var snapshotMap map[string]interface{}
	if err := json.Unmarshal(snapshot, &snapshotMap); err != nil {
		return err
	}

	version := &model.TopologyVersion{
		TopologyID:        topologyID,
		Version:           topology.Version + 1,
		Snapshot:          snapshotMap,
		ChangeDescription: description,
		ChangedBy:         userID,
	}

	if err := s.topologyRepo.CreateVersion(ctx, version); err != nil {
		return err
	}

	// 更新拓扑版本号
	topology.Version++
	return s.topologyRepo.Update(ctx, topology)
}

// GetVersions 获取版本列表
func (s *topologyService) GetVersions(ctx context.Context, topologyID uint) ([]model.TopologyVersion, error) {
	return s.topologyRepo.GetVersionsByTopologyID(ctx, topologyID)
}

// RestoreVersion 恢复版本
func (s *topologyService) RestoreVersion(ctx context.Context, topologyID uint, version int) error {
	ver, err := s.topologyRepo.GetVersionByID(ctx, topologyID, version)
	if err != nil {
		return err
	}

	// 反序列化快照
	snapshotBytes, err := json.Marshal(ver.Snapshot)
	if err != nil {
		return err
	}

	var topology model.Topology
	if err := json.Unmarshal(snapshotBytes, &topology); err != nil {
		return err
	}

	// 保持 ID 不变
	topology.ID = topologyID

	return s.topologyRepo.Update(ctx, &topology)
}

// UpsertLLDPNeighbor 创建或更新 LLDP 邻居
func (s *topologyService) UpsertLLDPNeighbor(ctx context.Context, neighbor *model.LLDPNeighbor) error {
	now := time.Now()
	neighbor.LastSeen = &now
	if neighbor.DiscoveredAt == nil {
		neighbor.DiscoveredAt = &now
	}

	return s.topologyRepo.UpsertLLDPNeighbor(ctx, neighbor)
}

// GetLLDPNeighbors 获取 LLDP 邻居
func (s *topologyService) GetLLDPNeighbors(ctx context.Context, deviceID string) ([]model.LLDPNeighbor, error) {
	return s.topologyRepo.GetLLDPNeighborsByDeviceID(ctx, deviceID)
}

// DiscoverTopologyResponse 自动发现响应
type DiscoverTopologyResponse struct {
	TopologyID      uint  `json:"topology_id"`
	DiscoveredNodes int   `json:"discovered_nodes"`
	DiscoveredLinks int   `json:"discovered_links"`
	DurationMs      int64 `json:"duration_ms"`
}

// DiscoverTopology 自动发现拓扑
func (s *topologyService) DiscoverTopology(ctx context.Context, topologyID uint) (*DiscoverTopologyResponse, error) {
	startTime := time.Now()
	
	result, err := s.discoveryService.DiscoverTopology(ctx, topologyID)
	if err != nil {
		return nil, fmt.Errorf("failed to discover topology: %w", err)
	}
	
	duration := time.Since(startTime)
	
	return &DiscoverTopologyResponse{
		TopologyID:      topologyID,
		DiscoveredNodes: result.DiscoveredNodes,
		DiscoveredLinks: result.DiscoveredLinks,
		DurationMs:      duration.Milliseconds(),
	}, nil
}

