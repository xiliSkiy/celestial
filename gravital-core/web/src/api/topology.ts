import request from '@/utils/request'
import type {
  Topology,
  TopologyQuery,
  TopologyListResponse,
  TopologyDetailResponse,
  CreateTopologyRequest,
  UpdateTopologyRequest,
  TopologyNode,
  AddNodeRequest,
  BatchUpdateNodesRequest,
  TopologyLink,
  AddLinkRequest,
  UpdateLinkStatusRequest,
  ApplyLayoutRequest,
  ApplyLayoutResponse,
  PathAnalysisRequest,
  PathAnalysisResponse,
  ImpactAnalysisRequest,
  ImpactAnalysisResponse,
  TopologyVersion,
  CreateSnapshotRequest
} from '@/types/topology'

export const topologyApi = {
  // 拓扑管理
  listTopologies: (params: TopologyQuery) =>
    request.get<TopologyListResponse>('/v1/topologies', { params }),

  getTopology: (id: number) =>
    request.get<TopologyDetailResponse>(`/v1/topologies/${id}`),

  createTopology: (data: CreateTopologyRequest) =>
    request.post<Topology>('/v1/topologies', data),

  updateTopology: (id: number, data: UpdateTopologyRequest) =>
    request.put(`/v1/topologies/${id}`, data),

  deleteTopology: (id: number) =>
    request.delete(`/v1/topologies/${id}`),

  // 节点管理
  addNode: (topologyId: number, data: AddNodeRequest) =>
    request.post<TopologyNode>(`/v1/topologies/${topologyId}/nodes`, data),

  updateNodePosition: (topologyId: number, nodeId: number, position: { x: number; y: number }) =>
    request.patch(`/v1/topologies/${topologyId}/nodes/${nodeId}/position`, position),

  batchUpdateNodes: (topologyId: number, data: BatchUpdateNodesRequest) =>
    request.patch(`/v1/topologies/${topologyId}/nodes/batch`, data),

  deleteNode: (topologyId: number, nodeId: number) =>
    request.delete(`/v1/topologies/${topologyId}/nodes/${nodeId}`),

  // 链路管理
  addLink: (topologyId: number, data: AddLinkRequest) =>
    request.post<TopologyLink>(`/v1/topologies/${topologyId}/links`, data),

  updateLinkStatus: (topologyId: number, linkId: number, data: UpdateLinkStatusRequest) =>
    request.patch(`/v1/topologies/${topologyId}/links/${linkId}/status`, data),

  deleteLink: (topologyId: number, linkId: number) =>
    request.delete(`/v1/topologies/${topologyId}/links/${linkId}`),

  // 布局
  applyLayout: (topologyId: number, data: ApplyLayoutRequest) =>
    request.post<ApplyLayoutResponse>(`/v1/topologies/${topologyId}/layout`, data),

  // 分析
  analyzePath: (topologyId: number, data: PathAnalysisRequest) =>
    request.post<PathAnalysisResponse>(`/v1/topologies/${topologyId}/analyze/path`, data),

  analyzeImpact: (topologyId: number, data: ImpactAnalysisRequest) =>
    request.post<ImpactAnalysisResponse>(`/v1/topologies/${topologyId}/analyze/impact`, data),

  // 版本管理
  getVersions: (topologyId: number) =>
    request.get<TopologyVersion[]>(`/v1/topologies/${topologyId}/versions`),

  createSnapshot: (topologyId: number, data: CreateSnapshotRequest) =>
    request.post(`/v1/topologies/${topologyId}/versions`, data),

  restoreVersion: (topologyId: number, version: number) =>
    request.post(`/v1/topologies/${topologyId}/versions/${version}/restore`)
}

