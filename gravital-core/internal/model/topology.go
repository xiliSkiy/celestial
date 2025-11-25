package model

import (
	"time"
)

// Topology 拓扑图
type Topology struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	Name              string    `gorm:"type:varchar(255);not null" json:"name"`
	Description       string    `gorm:"type:text" json:"description"`
	Type              string    `gorm:"type:varchar(32);not null;index" json:"type"` // physical, logical, custom
	Scope             string    `gorm:"type:varchar(32);index" json:"scope"`          // global, datacenter, region
	LayoutType        string    `gorm:"type:varchar(32);default:'force'" json:"layout_type"`
	LayoutConfig      JSONB     `gorm:"type:jsonb" json:"layout_config"`
	ViewConfig        JSONB     `gorm:"type:jsonb" json:"view_config"`
	IsAutoDiscovery   bool      `gorm:"default:false" json:"is_auto_discovery"`
	DiscoveryInterval int       `gorm:"type:int" json:"discovery_interval"` // 秒
	LastDiscoveryAt   *time.Time `json:"last_discovery_at"`
	Version           int       `gorm:"default:1" json:"version"`
	CreatedBy         uint      `json:"created_by"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`

	// 关联
	Nodes  []TopologyNode  `gorm:"foreignKey:TopologyID;constraint:OnDelete:CASCADE" json:"nodes,omitempty"`
	Links  []TopologyLink  `gorm:"foreignKey:TopologyID;constraint:OnDelete:CASCADE" json:"links,omitempty"`
	Groups []TopologyGroup `gorm:"foreignKey:TopologyID;constraint:OnDelete:CASCADE" json:"groups,omitempty"`
}

// TopologyNode 拓扑节点
type TopologyNode struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	TopologyID uint      `gorm:"not null;index:idx_topology_device" json:"topology_id"`
	DeviceID   string    `gorm:"type:varchar(64);not null;index:idx_topology_device" json:"device_id"`
	NodeType   string    `gorm:"type:varchar(32);not null" json:"node_type"` // device, group, cloud, internet
	Label      string    `gorm:"type:varchar(255)" json:"label"`
	Icon       string    `gorm:"type:varchar(64)" json:"icon"`
	PositionX  float64   `json:"position_x"`
	PositionY  float64   `json:"position_y"`
	Layer      int       `gorm:"default:0;index:idx_layer" json:"layer"` // 层级
	Size       int       `gorm:"default:40" json:"size"`
	Color      string    `gorm:"type:varchar(32)" json:"color"`
	Shape      string    `gorm:"type:varchar(32);default:'circle'" json:"shape"`
	Properties JSONB     `gorm:"type:jsonb" json:"properties"`
	IsLocked   bool      `gorm:"default:false" json:"is_locked"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Topology *Topology `gorm:"foreignKey:TopologyID" json:"-"`
}

// TopologyLink 拓扑链路
type TopologyLink struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	TopologyID      uint       `gorm:"not null;index" json:"topology_id"`
	SourceNodeID    uint       `gorm:"not null;index" json:"source_node_id"`
	TargetNodeID    uint       `gorm:"not null;index" json:"target_node_id"`
	LinkType        string     `gorm:"type:varchar(32);not null" json:"link_type"` // physical, logical, virtual
	SourceInterface string     `gorm:"type:varchar(128)" json:"source_interface"`
	TargetInterface string     `gorm:"type:varchar(128)" json:"target_interface"`
	Bandwidth       int64      `json:"bandwidth"` // bps
	Protocol        string     `gorm:"type:varchar(32)" json:"protocol"`
	Status          string     `gorm:"type:varchar(32);default:'unknown';index" json:"status"` // up, down, degraded, unknown
	Utilization     float64    `json:"utilization"`                                            // 0-100
	Latency         float64    `json:"latency"`                                                // ms
	PacketLoss      float64    `json:"packet_loss"`                                            // %
	LineStyle       string     `gorm:"type:varchar(32);default:'solid'" json:"line_style"`
	LineWidth       int        `gorm:"default:2" json:"line_width"`
	Color           string     `gorm:"type:varchar(32)" json:"color"`
	Label           string     `gorm:"type:varchar(255)" json:"label"`
	Properties      JSONB      `gorm:"type:jsonb" json:"properties"`
	DiscoveredBy    string     `gorm:"type:varchar(32)" json:"discovered_by"` // lldp, cdp, manual
	DiscoveredAt    *time.Time `json:"discovered_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`

	// 关联
	Topology   *Topology     `gorm:"foreignKey:TopologyID" json:"-"`
	SourceNode *TopologyNode `gorm:"foreignKey:SourceNodeID" json:"source_node,omitempty"`
	TargetNode *TopologyNode `gorm:"foreignKey:TargetNodeID" json:"target_node,omitempty"`
}

// TopologyGroup 拓扑分组
type TopologyGroup struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TopologyID  uint      `gorm:"not null;index" json:"topology_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description string    `gorm:"type:text" json:"description"`
	ParentID    *uint     `gorm:"index" json:"parent_id"`
	PositionX   float64   `json:"position_x"`
	PositionY   float64   `json:"position_y"`
	Width       float64   `json:"width"`
	Height      float64   `json:"height"`
	Color       string    `gorm:"type:varchar(32)" json:"color"`
	BorderColor string    `gorm:"type:varchar(32)" json:"border_color"`
	IsCollapsed bool      `gorm:"default:false" json:"is_collapsed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	// 关联
	Topology *Topology       `gorm:"foreignKey:TopologyID" json:"-"`
	Parent   *TopologyGroup  `gorm:"foreignKey:ParentID" json:"parent,omitempty"`
	Children []TopologyGroup `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

// TopologyVersion 拓扑版本历史
type TopologyVersion struct {
	ID                uint      `gorm:"primaryKey" json:"id"`
	TopologyID        uint      `gorm:"not null;uniqueIndex:idx_topology_version" json:"topology_id"`
	Version           int       `gorm:"not null;uniqueIndex:idx_topology_version" json:"version"`
	Snapshot          JSONB     `gorm:"type:jsonb;not null" json:"snapshot"`
	ChangeDescription string    `gorm:"type:text" json:"change_description"`
	ChangedBy         uint      `json:"changed_by"`
	CreatedAt         time.Time `json:"created_at"`

	// 关联
	Topology *Topology `gorm:"foreignKey:TopologyID" json:"-"`
}

// LLDPNeighbor LLDP 邻居信息
type LLDPNeighbor struct {
	ID                 uint       `gorm:"primaryKey" json:"id"`
	DeviceID           string     `gorm:"type:varchar(64);not null;uniqueIndex:idx_lldp_unique;index" json:"device_id"`
	LocalInterface     string     `gorm:"type:varchar(128);not null;uniqueIndex:idx_lldp_unique" json:"local_interface"`
	NeighborChassisID  string     `gorm:"type:varchar(128);uniqueIndex:idx_lldp_unique;index" json:"neighbor_chassis_id"`
	NeighborPortID     string     `gorm:"type:varchar(128);uniqueIndex:idx_lldp_unique" json:"neighbor_port_id"`
	NeighborSystemName string     `gorm:"type:varchar(255)" json:"neighbor_system_name"`
	NeighborSystemDesc string     `gorm:"type:text" json:"neighbor_system_desc"`
	NeighborPortDesc   string     `gorm:"type:varchar(255)" json:"neighbor_port_desc"`
	NeighborMgmtAddr   string     `gorm:"type:varchar(64)" json:"neighbor_mgmt_addr"`
	TTL                int        `json:"ttl"` // 秒
	DiscoveredAt       *time.Time `json:"discovered_at"`
	LastSeen           *time.Time `json:"last_seen"`
	CreatedAt          time.Time  `json:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at"`
}

// TableName 指定表名
func (Topology) TableName() string {
	return "topologies"
}

func (TopologyNode) TableName() string {
	return "topology_nodes"
}

func (TopologyLink) TableName() string {
	return "topology_links"
}

func (TopologyGroup) TableName() string {
	return "topology_groups"
}

func (TopologyVersion) TableName() string {
	return "topology_versions"
}

func (LLDPNeighbor) TableName() string {
	return "lldp_neighbors"
}

