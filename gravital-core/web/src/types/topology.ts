// 拓扑类型定义

export interface Topology {
  id: number
  name: string
  description: string
  type: 'physical' | 'logical' | 'custom'
  scope?: string
  layout_type: 'force' | 'hierarchical' | 'circular' | 'tree'
  layout_config?: Record<string, any>
  view_config?: Record<string, any>
  is_auto_discovery: boolean
  discovery_interval?: number
  last_discovery_at?: string
  version: number
  created_by?: number
  created_at: string
  updated_at: string
  nodes?: TopologyNode[]
  links?: TopologyLink[]
  groups?: TopologyGroup[]
}

export interface TopologyNode {
  id: number
  topology_id: number
  device_id: string
  node_type: 'device' | 'group' | 'cloud' | 'internet'
  label: string
  icon?: string
  position_x: number
  position_y: number
  layer: number
  size: number
  color?: string
  shape: string
  properties?: Record<string, any>
  is_locked: boolean
  created_at: string
  updated_at: string
}

export interface TopologyLink {
  id: number
  topology_id: number
  source_node_id: number
  target_node_id: number
  link_type: 'physical' | 'logical' | 'virtual'
  source_interface?: string
  target_interface?: string
  bandwidth?: number
  protocol?: string
  status: 'up' | 'down' | 'degraded' | 'unknown'
  utilization?: number
  latency?: number
  packet_loss?: number
  line_style: 'solid' | 'dashed' | 'dotted'
  line_width: number
  color?: string
  label?: string
  properties?: Record<string, any>
  discovered_by?: 'lldp' | 'cdp' | 'manual'
  discovered_at?: string
  created_at: string
  updated_at: string
  source_node?: TopologyNode
  target_node?: TopologyNode
}

export interface TopologyGroup {
  id: number
  topology_id: number
  name: string
  description?: string
  parent_id?: number
  position_x: number
  position_y: number
  width: number
  height: number
  color?: string
  border_color?: string
  is_collapsed: boolean
  created_at: string
  updated_at: string
}

export interface TopologyVersion {
  id: number
  topology_id: number
  version: number
  snapshot: Record<string, any>
  change_description?: string
  changed_by?: number
  created_at: string
}

export interface LLDPNeighbor {
  id: number
  device_id: string
  local_interface: string
  neighbor_chassis_id: string
  neighbor_port_id: string
  neighbor_system_name?: string
  neighbor_system_desc?: string
  neighbor_port_desc?: string
  neighbor_mgmt_addr?: string
  ttl: number
  discovered_at?: string
  last_seen?: string
  created_at: string
  updated_at: string
}

// API 请求和响应类型

export interface TopologyQuery {
  page?: number
  page_size?: number
  type?: string
  scope?: string
  keyword?: string
}

export interface TopologyListItem {
  id: number
  name: string
  type: string
  scope?: string
  node_count: number
  link_count: number
  created_at: string
}

export interface TopologyListResponse {
  total: number
  items: TopologyListItem[]
}

export interface TopologyDetailResponse extends Topology {
  node_count: number
  link_count: number
}

export interface CreateTopologyRequest {
  name: string
  description?: string
  type: 'physical' | 'logical' | 'custom'
  scope?: string
  layout_type?: string
  is_auto_discovery?: boolean
  discovery_interval?: number
  layout_config?: Record<string, any>
  view_config?: Record<string, any>
}

export interface UpdateTopologyRequest {
  name?: string
  description?: string
  layout_type?: string
  is_auto_discovery?: boolean
  discovery_interval?: number
  layout_config?: Record<string, any>
  view_config?: Record<string, any>
}

export interface AddNodeRequest {
  device_id: string
  node_type: 'device' | 'group' | 'cloud' | 'internet'
  label: string
  position_x?: number
  position_y?: number
  layer?: number
  icon?: string
  size?: number
  color?: string
  shape?: string
  properties?: Record<string, any>
}

export interface NodeUpdate {
  id: number
  position_x: number
  position_y: number
}

export interface BatchUpdateNodesRequest {
  nodes: NodeUpdate[]
}

export interface AddLinkRequest {
  source_node_id: number
  target_node_id: number
  link_type: 'physical' | 'logical' | 'virtual'
  source_interface?: string
  target_interface?: string
  bandwidth?: number
  protocol?: string
  label?: string
  properties?: Record<string, any>
}

export interface UpdateLinkStatusRequest {
  status: 'up' | 'down' | 'degraded' | 'unknown'
  utilization?: number
  latency?: number
  packet_loss?: number
}

export interface ApplyLayoutRequest {
  layout_type: 'force' | 'hierarchical' | 'circular' | 'tree'
  options?: Record<string, any>
}

export interface Position {
  x: number
  y: number
}

export interface NodePosition {
  id: number
  position: Position
}

export interface ApplyLayoutResponse {
  nodes: NodePosition[]
}

export interface PathAnalysisRequest {
  source_node_id: number
  target_node_id: number
  algorithm?: 'shortest' | 'all'
}

export interface Path {
  nodes: number[]
  links: number[]
  hop_count: number
  total_latency: number
}

export interface PathAnalysisResponse {
  paths: Path[]
}

export interface ImpactAnalysisRequest {
  node_id: number
  scenario: 'failure' | 'maintenance'
}

export interface ImpactAnalysisResponse {
  affected_nodes: number[]
  affected_links: number[]
  isolated_nodes: number[]
  impact_level: 'low' | 'medium' | 'high'
}

export interface CreateSnapshotRequest {
  description: string
}

// G6 图数据格式
export interface G6Node {
  id: string
  label: string
  type?: string
  x?: number
  y?: number
  size?: number
  color?: string
  style?: Record<string, any>
  stateStyles?: Record<string, any>
  data?: TopologyNode
}

export interface G6Edge {
  id: string
  source: string
  target: string
  label?: string
  type?: string
  style?: Record<string, any>
  stateStyles?: Record<string, any>
  data?: TopologyLink
}

export interface G6GraphData {
  nodes: G6Node[]
  edges: G6Edge[]
}

