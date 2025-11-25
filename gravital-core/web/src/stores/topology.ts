import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { topologyApi } from '@/api/topology'
import type {
  Topology,
  TopologyQuery,
  TopologyListItem,
  TopologyDetailResponse,
  CreateTopologyRequest,
  UpdateTopologyRequest,
  TopologyNode,
  TopologyLink,
  AddNodeRequest,
  AddLinkRequest
} from '@/types/topology'
import { ElMessage } from 'element-plus'

export const useTopologyStore = defineStore('topology', () => {
  // 状态
  const topologies = ref<TopologyListItem[]>([])
  const currentTopology = ref<TopologyDetailResponse | null>(null)
  const total = ref(0)
  const loading = ref(false)

  // 计算属性
  const hasTopology = computed(() => currentTopology.value !== null)
  const nodes = computed(() => currentTopology.value?.nodes || [])
  const links = computed(() => currentTopology.value?.links || [])
  const groups = computed(() => currentTopology.value?.groups || [])

  // 获取拓扑列表
  const fetchTopologies = async (query: TopologyQuery = {}) => {
    loading.value = true
    try {
      const res: any = await topologyApi.listTopologies(query)
      topologies.value = res.items || []
      total.value = res.total || 0
    } catch (error) {
      ElMessage.error('获取拓扑列表失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 获取拓扑详情
  const fetchTopology = async (id: number) => {
    loading.value = true
    try {
      const res: TopologyDetailResponse = await topologyApi.getTopology(id)
      currentTopology.value = res
      return res
    } catch (error) {
      ElMessage.error('获取拓扑详情失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 创建拓扑
  const createTopology = async (data: CreateTopologyRequest) => {
    try {
      const res: Topology = await topologyApi.createTopology(data)
      ElMessage.success('创建拓扑成功')
      return res
    } catch (error) {
      ElMessage.error('创建拓扑失败')
      throw error
    }
  }

  // 更新拓扑
  const updateTopology = async (id: number, data: UpdateTopologyRequest) => {
    try {
      await topologyApi.updateTopology(id, data)
      ElMessage.success('更新拓扑成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === id) {
        await fetchTopology(id)
      }
    } catch (error) {
      ElMessage.error('更新拓扑失败')
      throw error
    }
  }

  // 删除拓扑
  const deleteTopology = async (id: number) => {
    try {
      await topologyApi.deleteTopology(id)
      ElMessage.success('删除拓扑成功')
      // 如果删除的是当前拓扑，清空
      if (currentTopology.value?.id === id) {
        currentTopology.value = null
      }
    } catch (error) {
      ElMessage.error('删除拓扑失败')
      throw error
    }
  }

  // 添加节点
  const addNode = async (topologyId: number, data: AddNodeRequest) => {
    try {
      const res: TopologyNode = await topologyApi.addNode(topologyId, data)
      ElMessage.success('添加节点成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === topologyId) {
        await fetchTopology(topologyId)
      }
      return res
    } catch (error) {
      ElMessage.error('添加节点失败')
      throw error
    }
  }

  // 更新节点位置
  const updateNodePosition = async (topologyId: number, nodeId: number, x: number, y: number) => {
    try {
      await topologyApi.updateNodePosition(topologyId, nodeId, { x, y })
      // 更新本地状态
      if (currentTopology.value?.nodes) {
        const node = currentTopology.value.nodes.find(n => n.id === nodeId)
        if (node) {
          node.position_x = x
          node.position_y = y
        }
      }
    } catch (error) {
      console.error('更新节点位置失败:', error)
      throw error
    }
  }

  // 批量更新节点
  const batchUpdateNodes = async (topologyId: number, nodes: Array<{ id: number; position_x: number; position_y: number }>) => {
    try {
      await topologyApi.batchUpdateNodes(topologyId, { nodes })
      // 更新本地状态
      if (currentTopology.value?.nodes) {
        nodes.forEach(update => {
          const node = currentTopology.value!.nodes!.find(n => n.id === update.id)
          if (node) {
            node.position_x = update.position_x
            node.position_y = update.position_y
          }
        })
      }
    } catch (error) {
      ElMessage.error('批量更新节点失败')
      throw error
    }
  }

  // 删除节点
  const deleteNode = async (topologyId: number, nodeId: number) => {
    try {
      await topologyApi.deleteNode(topologyId, nodeId)
      ElMessage.success('删除节点成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === topologyId) {
        await fetchTopology(topologyId)
      }
    } catch (error) {
      ElMessage.error('删除节点失败')
      throw error
    }
  }

  // 添加链路
  const addLink = async (topologyId: number, data: AddLinkRequest) => {
    try {
      const res: TopologyLink = await topologyApi.addLink(topologyId, data)
      ElMessage.success('添加链路成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === topologyId) {
        await fetchTopology(topologyId)
      }
      return res
    } catch (error) {
      ElMessage.error('添加链路失败')
      throw error
    }
  }

  // 删除链路
  const deleteLink = async (topologyId: number, linkId: number) => {
    try {
      await topologyApi.deleteLink(topologyId, linkId)
      ElMessage.success('删除链路成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === topologyId) {
        await fetchTopology(topologyId)
      }
    } catch (error) {
      ElMessage.error('删除链路失败')
      throw error
    }
  }

  // 应用布局
  const applyLayout = async (topologyId: number, layoutType: string, options?: Record<string, any>) => {
    try {
      const res: any = await topologyApi.applyLayout(topologyId, { layout_type: layoutType, options })
      ElMessage.success('应用布局成功')
      return res
    } catch (error) {
      ElMessage.error('应用布局失败')
      throw error
    }
  }

  // 路径分析
  const analyzePath = async (topologyId: number, sourceNodeId: number, targetNodeId: number) => {
    try {
      const res: any = await topologyApi.analyzePath(topologyId, {
        source_node_id: sourceNodeId,
        target_node_id: targetNodeId,
        algorithm: 'shortest'
      })
      return res
    } catch (error) {
      ElMessage.error('路径分析失败')
      throw error
    }
  }

  // 影响分析
  const analyzeImpact = async (topologyId: number, nodeId: number, scenario: 'failure' | 'maintenance' = 'failure') => {
    try {
      const res: any = await topologyApi.analyzeImpact(topologyId, {
        node_id: nodeId,
        scenario
      })
      return res
    } catch (error) {
      ElMessage.error('影响分析失败')
      throw error
    }
  }

  // 创建快照
  const createSnapshot = async (topologyId: number, description: string) => {
    try {
      await topologyApi.createSnapshot(topologyId, { description })
      ElMessage.success('创建快照成功')
    } catch (error) {
      ElMessage.error('创建快照失败')
      throw error
    }
  }

  // 获取版本列表
  const fetchVersions = async (topologyId: number) => {
    try {
      const res: any = await topologyApi.getVersions(topologyId)
      return res
    } catch (error) {
      ElMessage.error('获取版本列表失败')
      throw error
    }
  }

  // 恢复版本
  const restoreVersion = async (topologyId: number, version: number) => {
    try {
      await topologyApi.restoreVersion(topologyId, version)
      ElMessage.success('恢复版本成功')
      // 刷新当前拓扑
      if (currentTopology.value?.id === topologyId) {
        await fetchTopology(topologyId)
      }
    } catch (error) {
      ElMessage.error('恢复版本失败')
      throw error
    }
  }

  // 清空当前拓扑
  const clearCurrentTopology = () => {
    currentTopology.value = null
  }

  return {
    // 状态
    topologies,
    currentTopology,
    total,
    loading,

    // 计算属性
    hasTopology,
    nodes,
    links,
    groups,

    // 方法
    fetchTopologies,
    fetchTopology,
    createTopology,
    updateTopology,
    deleteTopology,
    addNode,
    updateNodePosition,
    batchUpdateNodes,
    deleteNode,
    addLink,
    deleteLink,
    applyLayout,
    analyzePath,
    analyzeImpact,
    createSnapshot,
    fetchVersions,
    restoreVersion,
    clearCurrentTopology
  }
})

