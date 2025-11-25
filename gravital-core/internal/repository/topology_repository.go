package repository

import (
	"context"

	"github.com/celestial/gravital-core/internal/model"
	"gorm.io/gorm"
)

// TopologyRepository 拓扑仓储接口
type TopologyRepository interface {
	// 拓扑管理
	Create(ctx context.Context, topology *model.Topology) error
	GetByID(ctx context.Context, id uint) (*model.Topology, error)
	Update(ctx context.Context, topology *model.Topology) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, filter TopologyFilter) ([]model.Topology, int64, error)

	// 节点管理
	CreateNode(ctx context.Context, node *model.TopologyNode) error
	GetNodeByID(ctx context.Context, id uint) (*model.TopologyNode, error)
	UpdateNode(ctx context.Context, node *model.TopologyNode) error
	DeleteNode(ctx context.Context, id uint) error
	BatchUpdateNodes(ctx context.Context, nodes []model.TopologyNode) error
	GetNodesByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyNode, error)

	// 链路管理
	CreateLink(ctx context.Context, link *model.TopologyLink) error
	GetLinkByID(ctx context.Context, id uint) (*model.TopologyLink, error)
	UpdateLink(ctx context.Context, link *model.TopologyLink) error
	DeleteLink(ctx context.Context, id uint) error
	GetLinksByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyLink, error)

	// 分组管理
	CreateGroup(ctx context.Context, group *model.TopologyGroup) error
	GetGroupByID(ctx context.Context, id uint) (*model.TopologyGroup, error)
	UpdateGroup(ctx context.Context, group *model.TopologyGroup) error
	DeleteGroup(ctx context.Context, id uint) error
	GetGroupsByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyGroup, error)

	// 版本管理
	CreateVersion(ctx context.Context, version *model.TopologyVersion) error
	GetVersionsByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyVersion, error)
	GetVersionByID(ctx context.Context, topologyID uint, version int) (*model.TopologyVersion, error)

	// LLDP 邻居管理
	UpsertLLDPNeighbor(ctx context.Context, neighbor *model.LLDPNeighbor) error
	GetLLDPNeighborsByDeviceID(ctx context.Context, deviceID string) ([]model.LLDPNeighbor, error)
	GetAllLLDPNeighbors(ctx context.Context) ([]model.LLDPNeighbor, error)
	DeleteStaleLLDPNeighbors(ctx context.Context, ttl int) error
}

// TopologyFilter 拓扑过滤器
type TopologyFilter struct {
	Page     int
	PageSize int
	Type     string
	Scope    string
	Keyword  string
}

type topologyRepository struct {
	db *gorm.DB
}

// NewTopologyRepository 创建拓扑仓储实例
func NewTopologyRepository(db *gorm.DB) TopologyRepository {
	return &topologyRepository{db: db}
}

// Create 创建拓扑
func (r *topologyRepository) Create(ctx context.Context, topology *model.Topology) error {
	return r.db.WithContext(ctx).Create(topology).Error
}

// GetByID 根据 ID 获取拓扑
func (r *topologyRepository) GetByID(ctx context.Context, id uint) (*model.Topology, error) {
	var topology model.Topology
	err := r.db.WithContext(ctx).
		Preload("Nodes").
		Preload("Links").
		Preload("Links.SourceNode").
		Preload("Links.TargetNode").
		Preload("Groups").
		First(&topology, id).Error
	if err != nil {
		return nil, err
	}
	return &topology, nil
}

// Update 更新拓扑
func (r *topologyRepository) Update(ctx context.Context, topology *model.Topology) error {
	return r.db.WithContext(ctx).Save(topology).Error
}

// Delete 删除拓扑
func (r *topologyRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Topology{}, id).Error
}

// List 获取拓扑列表
func (r *topologyRepository) List(ctx context.Context, filter TopologyFilter) ([]model.Topology, int64, error) {
	var topologies []model.Topology
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Topology{})

	// 过滤条件
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
	}
	if filter.Scope != "" {
		query = query.Where("scope = ?", filter.Scope)
	}
	if filter.Keyword != "" {
		query = query.Where("name LIKE ? OR description LIKE ?",
			"%"+filter.Keyword+"%", "%"+filter.Keyword+"%")
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 分页查询
	offset := (filter.Page - 1) * filter.PageSize
	err := query.Offset(offset).Limit(filter.PageSize).
		Order("created_at DESC").
		Find(&topologies).Error

	return topologies, total, err
}

// CreateNode 创建节点
func (r *topologyRepository) CreateNode(ctx context.Context, node *model.TopologyNode) error {
	return r.db.WithContext(ctx).Create(node).Error
}

// GetNodeByID 根据 ID 获取节点
func (r *topologyRepository) GetNodeByID(ctx context.Context, id uint) (*model.TopologyNode, error) {
	var node model.TopologyNode
	err := r.db.WithContext(ctx).First(&node, id).Error
	if err != nil {
		return nil, err
	}
	return &node, nil
}

// UpdateNode 更新节点
func (r *topologyRepository) UpdateNode(ctx context.Context, node *model.TopologyNode) error {
	return r.db.WithContext(ctx).Save(node).Error
}

// DeleteNode 删除节点
func (r *topologyRepository) DeleteNode(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.TopologyNode{}, id).Error
}

// BatchUpdateNodes 批量更新节点
func (r *topologyRepository) BatchUpdateNodes(ctx context.Context, nodes []model.TopologyNode) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, node := range nodes {
			if err := tx.Save(&node).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// GetNodesByTopologyID 获取拓扑的所有节点
func (r *topologyRepository) GetNodesByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyNode, error) {
	var nodes []model.TopologyNode
	err := r.db.WithContext(ctx).
		Where("topology_id = ?", topologyID).
		Find(&nodes).Error
	return nodes, err
}

// CreateLink 创建链路
func (r *topologyRepository) CreateLink(ctx context.Context, link *model.TopologyLink) error {
	return r.db.WithContext(ctx).Create(link).Error
}

// GetLinkByID 根据 ID 获取链路
func (r *topologyRepository) GetLinkByID(ctx context.Context, id uint) (*model.TopologyLink, error) {
	var link model.TopologyLink
	err := r.db.WithContext(ctx).
		Preload("SourceNode").
		Preload("TargetNode").
		First(&link, id).Error
	if err != nil {
		return nil, err
	}
	return &link, nil
}

// UpdateLink 更新链路
func (r *topologyRepository) UpdateLink(ctx context.Context, link *model.TopologyLink) error {
	return r.db.WithContext(ctx).Save(link).Error
}

// DeleteLink 删除链路
func (r *topologyRepository) DeleteLink(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.TopologyLink{}, id).Error
}

// GetLinksByTopologyID 获取拓扑的所有链路
func (r *topologyRepository) GetLinksByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyLink, error) {
	var links []model.TopologyLink
	err := r.db.WithContext(ctx).
		Where("topology_id = ?", topologyID).
		Preload("SourceNode").
		Preload("TargetNode").
		Find(&links).Error
	return links, err
}

// CreateGroup 创建分组
func (r *topologyRepository) CreateGroup(ctx context.Context, group *model.TopologyGroup) error {
	return r.db.WithContext(ctx).Create(group).Error
}

// GetGroupByID 根据 ID 获取分组
func (r *topologyRepository) GetGroupByID(ctx context.Context, id uint) (*model.TopologyGroup, error) {
	var group model.TopologyGroup
	err := r.db.WithContext(ctx).First(&group, id).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup 更新分组
func (r *topologyRepository) UpdateGroup(ctx context.Context, group *model.TopologyGroup) error {
	return r.db.WithContext(ctx).Save(group).Error
}

// DeleteGroup 删除分组
func (r *topologyRepository) DeleteGroup(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.TopologyGroup{}, id).Error
}

// GetGroupsByTopologyID 获取拓扑的所有分组
func (r *topologyRepository) GetGroupsByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyGroup, error) {
	var groups []model.TopologyGroup
	err := r.db.WithContext(ctx).
		Where("topology_id = ?", topologyID).
		Find(&groups).Error
	return groups, err
}

// CreateVersion 创建版本
func (r *topologyRepository) CreateVersion(ctx context.Context, version *model.TopologyVersion) error {
	return r.db.WithContext(ctx).Create(version).Error
}

// GetVersionsByTopologyID 获取拓扑的所有版本
func (r *topologyRepository) GetVersionsByTopologyID(ctx context.Context, topologyID uint) ([]model.TopologyVersion, error) {
	var versions []model.TopologyVersion
	err := r.db.WithContext(ctx).
		Where("topology_id = ?", topologyID).
		Order("version DESC").
		Find(&versions).Error
	return versions, err
}

// GetVersionByID 获取指定版本
func (r *topologyRepository) GetVersionByID(ctx context.Context, topologyID uint, version int) (*model.TopologyVersion, error) {
	var ver model.TopologyVersion
	err := r.db.WithContext(ctx).
		Where("topology_id = ? AND version = ?", topologyID, version).
		First(&ver).Error
	if err != nil {
		return nil, err
	}
	return &ver, nil
}

// UpsertLLDPNeighbor 创建或更新 LLDP 邻居
func (r *topologyRepository) UpsertLLDPNeighbor(ctx context.Context, neighbor *model.LLDPNeighbor) error {
	return r.db.WithContext(ctx).
		Where("device_id = ? AND local_interface = ? AND neighbor_chassis_id = ? AND neighbor_port_id = ?",
			neighbor.DeviceID, neighbor.LocalInterface, neighbor.NeighborChassisID, neighbor.NeighborPortID).
		Assign(neighbor).
		FirstOrCreate(neighbor).Error
}

// GetLLDPNeighborsByDeviceID 获取设备的 LLDP 邻居
func (r *topologyRepository) GetLLDPNeighborsByDeviceID(ctx context.Context, deviceID string) ([]model.LLDPNeighbor, error) {
	var neighbors []model.LLDPNeighbor
	err := r.db.WithContext(ctx).
		Where("device_id = ?", deviceID).
		Find(&neighbors).Error
	return neighbors, err
}

// GetAllLLDPNeighbors 获取所有 LLDP 邻居
func (r *topologyRepository) GetAllLLDPNeighbors(ctx context.Context) ([]model.LLDPNeighbor, error) {
	var neighbors []model.LLDPNeighbor
	err := r.db.WithContext(ctx).Find(&neighbors).Error
	return neighbors, err
}

// DeleteStaleLLDPNeighbors 删除过期的 LLDP 邻居
func (r *topologyRepository) DeleteStaleLLDPNeighbors(ctx context.Context, ttl int) error {
	return r.db.WithContext(ctx).
		Where("last_seen < NOW() - INTERVAL '? seconds'", ttl).
		Delete(&model.LLDPNeighbor{}).Error
}

