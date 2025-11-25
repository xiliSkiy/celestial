<template>
  <div class="topology-canvas">
    <div ref="containerRef" class="canvas-container"></div>
    
    <!-- 工具栏 -->
    <div class="canvas-toolbar">
      <el-button-group>
        <el-tooltip content="适应画布">
          <el-button :icon="FullScreen" @click="fitView" />
        </el-tooltip>
        <el-tooltip content="放大">
          <el-button :icon="ZoomIn" @click="zoomIn" />
        </el-tooltip>
        <el-tooltip content="缩小">
          <el-button :icon="ZoomOut" @click="zoomOut" />
        </el-tooltip>
        <el-tooltip content="重置缩放">
          <el-button :icon="Refresh" @click="resetZoom" />
        </el-tooltip>
      </el-button-group>
      
      <el-divider direction="vertical" />
      
      <el-select
        v-model="currentLayout"
        placeholder="选择布局"
        style="width: 120px"
        @change="handleLayoutChange"
      >
        <el-option label="力导向" value="force" />
        <el-option label="层次布局" value="hierarchical" />
        <el-option label="环形布局" value="circular" />
        <el-option label="网格布局" value="grid" />
      </el-select>
    </div>
    
    <!-- 迷你地图 -->
    <div v-if="showMinimap" ref="minimapRef" class="canvas-minimap"></div>
    
    <!-- 节点详情面板 -->
    <el-drawer
      v-model="drawerVisible"
      :title="selectedNode ? '节点详情' : '链路详情'"
      direction="rtl"
      size="400px"
    >
      <div v-if="selectedNode" class="node-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="标签">{{ selectedNode.label }}</el-descriptions-item>
          <el-descriptions-item label="设备ID">{{ selectedNode.device_id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ selectedNode.node_type }}</el-descriptions-item>
          <el-descriptions-item label="层级">{{ selectedNode.layer }}</el-descriptions-item>
          <el-descriptions-item label="位置">
            ({{ selectedNode.position_x.toFixed(2) }}, {{ selectedNode.position_y.toFixed(2) }})
          </el-descriptions-item>
        </el-descriptions>
      </div>
      
      <div v-if="selectedLink" class="link-detail">
        <el-descriptions :column="1" border>
          <el-descriptions-item label="类型">{{ selectedLink.link_type }}</el-descriptions-item>
          <el-descriptions-item label="状态">
            <el-tag :type="getLinkStatusType(selectedLink.status)">
              {{ selectedLink.status }}
            </el-tag>
          </el-descriptions-item>
          <el-descriptions-item label="源接口">{{ selectedLink.source_interface }}</el-descriptions-item>
          <el-descriptions-item label="目标接口">{{ selectedLink.target_interface }}</el-descriptions-item>
          <el-descriptions-item label="带宽">
            {{ formatBandwidth(selectedLink.bandwidth) }}
          </el-descriptions-item>
          <el-descriptions-item label="利用率">
            {{ selectedLink.utilization?.toFixed(2) }}%
          </el-descriptions-item>
          <el-descriptions-item label="延迟">
            {{ selectedLink.latency?.toFixed(2) }} ms
          </el-descriptions-item>
          <el-descriptions-item label="丢包率">
            {{ selectedLink.packet_loss?.toFixed(2) }}%
          </el-descriptions-item>
        </el-descriptions>
      </div>
    </el-drawer>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onBeforeUnmount, watch, nextTick } from 'vue'
import { FullScreen, ZoomIn, ZoomOut, Refresh } from '@element-plus/icons-vue'
import type { TopologyNode, TopologyLink, G6GraphData } from '@/types/topology'

// 注意：G6 需要单独安装：npm install @antv/g6
// 这里使用动态导入以避免构建错误
let G6: any = null

interface Props {
  nodes: TopologyNode[]
  links: TopologyLink[]
  editable?: boolean
  showMinimap?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editable: false,
  showMinimap: true
})

const emit = defineEmits<{
  nodeClick: [node: TopologyNode]
  linkClick: [link: TopologyLink]
  nodePositionChange: [nodeId: number, x: number, y: number]
  addNode: [x: number, y: number]
  addLink: [sourceId: number, targetId: number]
}>()

const containerRef = ref<HTMLDivElement>()
const minimapRef = ref<HTMLDivElement>()
let graph: any = null

const currentLayout = ref('force')
const drawerVisible = ref(false)
const selectedNode = ref<TopologyNode | null>(null)
const selectedLink = ref<TopologyLink | null>(null)

// 初始化 G6
onMounted(async () => {
  try {
    // 动态导入 G6
    const G6Module = await import('@antv/g6')
    G6 = G6Module.default || G6Module
    
    await nextTick()
    initGraph()
    updateGraph()
  } catch (error) {
    console.error('Failed to load G6:', error)
    console.warn('请先安装 @antv/g6: npm install @antv/g6')
  }
})

onBeforeUnmount(() => {
  if (graph) {
    graph.destroy()
  }
})

// 监听数据变化
watch(() => [props.nodes, props.links], () => {
  updateGraph()
}, { deep: true })

// 初始化图
const initGraph = () => {
  if (!containerRef.value || !G6) return

  const width = containerRef.value.offsetWidth
  const height = containerRef.value.offsetHeight

  graph = new G6.Graph({
    container: containerRef.value,
    width,
    height,
    modes: {
      default: props.editable
        ? ['drag-canvas', 'zoom-canvas', 'drag-node', 'click-select']
        : ['drag-canvas', 'zoom-canvas', 'click-select']
    },
    layout: {
      type: 'force',
      preventOverlap: true,
      nodeSpacing: 100,
      linkDistance: 150
    },
    defaultNode: {
      size: 40,
      style: {
        fill: '#409EFF',
        stroke: '#fff',
        lineWidth: 2
      },
      labelCfg: {
        position: 'bottom',
        offset: 10,
        style: {
          fontSize: 12,
          fill: '#333'
        }
      }
    },
    defaultEdge: {
      type: 'line',
      style: {
        stroke: '#e2e2e2',
        lineWidth: 2,
        endArrow: {
          path: G6.Arrow.triangle(8, 10, 0),
          fill: '#e2e2e2'
        }
      },
      labelCfg: {
        autoRotate: true,
        style: {
          fontSize: 10,
          fill: '#666'
        }
      }
    },
    nodeStateStyles: {
      selected: {
        stroke: '#409EFF',
        lineWidth: 3
      },
      hover: {
        stroke: '#67C23A',
        lineWidth: 3
      }
    },
    edgeStateStyles: {
      selected: {
        stroke: '#409EFF',
        lineWidth: 3
      },
      hover: {
        stroke: '#67C23A',
        lineWidth: 3
      }
    }
  })

  // 绑定事件
  graph.on('node:click', (evt: any) => {
    const node = evt.item.getModel()
    selectedNode.value = node.data
    selectedLink.value = null
    drawerVisible.value = true
    emit('nodeClick', node.data)
  })

  graph.on('edge:click', (evt: any) => {
    const edge = evt.item.getModel()
    selectedLink.value = edge.data
    selectedNode.value = null
    drawerVisible.value = true
    emit('linkClick', edge.data)
  })

  graph.on('node:dragend', (evt: any) => {
    if (props.editable) {
      const node = evt.item.getModel()
      emit('nodePositionChange', node.data.id, node.x, node.y)
    }
  })

  // 添加迷你地图
  if (props.showMinimap && minimapRef.value) {
    const minimap = new G6.Minimap({
      container: minimapRef.value,
      size: [200, 150]
    })
    graph.addPlugin(minimap)
  }
}

// 更新图数据
const updateGraph = () => {
  if (!graph) return

  const graphData: G6GraphData = {
    nodes: props.nodes.map(node => ({
      id: String(node.id),
      label: node.label,
      x: node.position_x,
      y: node.position_y,
      size: node.size,
      style: {
        fill: node.color || getNodeColor(node.node_type),
        stroke: '#fff'
      },
      data: node
    })),
    edges: props.links.map(link => ({
      id: String(link.id),
      source: String(link.source_node_id),
      target: String(link.target_node_id),
      label: link.label,
      style: {
        stroke: link.color || getLinkColor(link.status),
        lineWidth: link.line_width,
        lineDash: link.line_style === 'dashed' ? [5, 5] : undefined
      },
      data: link
    }))
  }

  graph.data(graphData)
  graph.render()
}

// 获取节点颜色
const getNodeColor = (nodeType: string) => {
  const colors: Record<string, string> = {
    device: '#409EFF',
    group: '#67C23A',
    cloud: '#E6A23C',
    internet: '#F56C6C'
  }
  return colors[nodeType] || '#909399'
}

// 获取链路颜色
const getLinkColor = (status: string) => {
  const colors: Record<string, string> = {
    up: '#67C23A',
    down: '#F56C6C',
    degraded: '#E6A23C',
    unknown: '#909399'
  }
  return colors[status] || '#e2e2e2'
}

// 获取链路状态类型
const getLinkStatusType = (status: string) => {
  const types: Record<string, any> = {
    up: 'success',
    down: 'danger',
    degraded: 'warning',
    unknown: 'info'
  }
  return types[status] || 'info'
}

// 格式化带宽
const formatBandwidth = (bandwidth?: number) => {
  if (!bandwidth) return '-'
  if (bandwidth >= 1000000000) {
    return `${(bandwidth / 1000000000).toFixed(2)} Gbps`
  }
  if (bandwidth >= 1000000) {
    return `${(bandwidth / 1000000).toFixed(2)} Mbps`
  }
  return `${(bandwidth / 1000).toFixed(2)} Kbps`
}

// 工具栏方法
const fitView = () => {
  graph?.fitView()
}

const zoomIn = () => {
  const zoom = graph.getZoom()
  graph.zoomTo(zoom + 0.1)
}

const zoomOut = () => {
  const zoom = graph.getZoom()
  graph.zoomTo(zoom - 0.1)
}

const resetZoom = () => {
  graph?.zoomTo(1)
  graph?.fitCenter()
}

const handleLayoutChange = (layout: string) => {
  if (!graph) return

  const layouts: Record<string, any> = {
    force: {
      type: 'force',
      preventOverlap: true,
      nodeSpacing: 100,
      linkDistance: 150
    },
    hierarchical: {
      type: 'dagre',
      rankdir: 'TB',
      nodesep: 50,
      ranksep: 100
    },
    circular: {
      type: 'circular',
      radius: 300
    },
    grid: {
      type: 'grid',
      rows: 5,
      cols: 5
    }
  }

  graph.updateLayout(layouts[layout] || layouts.force)
}

// 暴露方法给父组件
defineExpose({
  fitView,
  zoomIn,
  zoomOut,
  resetZoom,
  getGraph: () => graph
})
</script>

<style scoped lang="scss">
.topology-canvas {
  position: relative;
  width: 100%;
  height: 100%;
  background: #f5f5f5;

  .canvas-container {
    width: 100%;
    height: 100%;
  }

  .canvas-toolbar {
    position: absolute;
    top: 20px;
    left: 20px;
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px;
    background: white;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    z-index: 10;
  }

  .canvas-minimap {
    position: absolute;
    bottom: 20px;
    right: 20px;
    border: 1px solid #ddd;
    border-radius: 4px;
    background: white;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    z-index: 10;
  }

  .node-detail,
  .link-detail {
    padding: 20px;
  }
}
</style>

