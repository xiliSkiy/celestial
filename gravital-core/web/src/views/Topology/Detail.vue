<template>
  <div class="topology-detail">
    <el-page-header @back="goBack" :title="topology?.name || '拓扑详情'" />
    
    <div v-loading="loading" class="topology-container">
      <TopologyCanvas
        v-if="topology"
        :nodes="topology.nodes || []"
        :links="topology.links || []"
        :editable="true"
        :show-minimap="true"
        @node-click="handleNodeClick"
        @link-click="handleLinkClick"
        @node-position-change="handleNodePositionChange"
      />
      
      <el-empty v-else description="暂无拓扑数据" />
    </div>
    
    <!-- 操作按钮 -->
    <div class="action-buttons">
      <el-button-group>
        <el-button type="primary" @click="handleAddNode">添加节点</el-button>
        <el-button type="primary" @click="handleAddLink">添加链路</el-button>
        <el-button @click="handleSnapshot">创建快照</el-button>
        <el-button @click="handleVersions">版本历史</el-button>
        <el-button @click="handlePathAnalysis">路径分析</el-button>
        <el-button @click="handleImpactAnalysis">影响分析</el-button>
      </el-button-group>
    </div>
    
    <!-- 添加节点对话框 -->
    <el-dialog v-model="addNodeDialogVisible" title="添加节点" width="500px">
      <el-form :model="nodeForm" label-width="100px">
        <el-form-item label="设备ID">
          <el-input v-model="nodeForm.device_id" placeholder="请输入设备ID" />
        </el-form-item>
        <el-form-item label="标签">
          <el-input v-model="nodeForm.label" placeholder="请输入节点标签" />
        </el-form-item>
        <el-form-item label="节点类型">
          <el-select v-model="nodeForm.node_type">
            <el-option label="设备" value="device" />
            <el-option label="分组" value="group" />
            <el-option label="云服务" value="cloud" />
            <el-option label="互联网" value="internet" />
          </el-select>
        </el-form-item>
        <el-form-item label="层级">
          <el-input-number v-model="nodeForm.layer" :min="0" :max="10" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addNodeDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddNodeSubmit">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 添加链路对话框 -->
    <el-dialog v-model="addLinkDialogVisible" title="添加链路" width="500px">
      <el-form :model="linkForm" label-width="100px">
        <el-form-item label="源节点">
          <el-select v-model="linkForm.source_node_id" placeholder="请选择源节点">
            <el-option
              v-for="node in topology?.nodes"
              :key="node.id"
              :label="node.label"
              :value="node.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="目标节点">
          <el-select v-model="linkForm.target_node_id" placeholder="请选择目标节点">
            <el-option
              v-for="node in topology?.nodes"
              :key="node.id"
              :label="node.label"
              :value="node.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="链路类型">
          <el-select v-model="linkForm.link_type">
            <el-option label="物理连接" value="physical" />
            <el-option label="逻辑连接" value="logical" />
            <el-option label="虚拟连接" value="virtual" />
          </el-select>
        </el-form-item>
        <el-form-item label="源接口">
          <el-input v-model="linkForm.source_interface" placeholder="例如: GigabitEthernet1/0/1" />
        </el-form-item>
        <el-form-item label="目标接口">
          <el-input v-model="linkForm.target_interface" placeholder="例如: GigabitEthernet1/0/1" />
        </el-form-item>
        <el-form-item label="带宽(bps)">
          <el-input-number v-model="linkForm.bandwidth" :min="0" :step="1000000" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addLinkDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleAddLinkSubmit">确定</el-button>
      </template>
    </el-dialog>
    
    <!-- 版本历史对话框 -->
    <el-dialog v-model="versionsDialogVisible" title="版本历史" width="800px">
      <el-table :data="versions" stripe>
        <el-table-column prop="version" label="版本号" width="100" />
        <el-table-column prop="change_description" label="变更说明" min-width="200" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="120">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleRestore(row.version)">
              恢复
            </el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, reactive } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { useTopologyStore } from '@/stores/topology'
import TopologyCanvas from '@/components/topology/TopologyCanvas.vue'
import type { TopologyDetailResponse, TopologyNode, TopologyLink, AddNodeRequest, AddLinkRequest, TopologyVersion } from '@/types/topology'
import { formatDateTime } from '@/utils/format'

const route = useRoute()
const router = useRouter()
const topologyStore = useTopologyStore()

const loading = ref(false)
const topology = ref<TopologyDetailResponse | null>(null)
const topologyId = ref(Number(route.params.id))

const addNodeDialogVisible = ref(false)
const addLinkDialogVisible = ref(false)
const versionsDialogVisible = ref(false)
const versions = ref<TopologyVersion[]>([])

const nodeForm = reactive<AddNodeRequest>({
  device_id: '',
  label: '',
  node_type: 'device',
  layer: 0
})

const linkForm = reactive<AddLinkRequest>({
  source_node_id: 0,
  target_node_id: 0,
  link_type: 'physical',
  source_interface: '',
  target_interface: '',
  bandwidth: 1000000000
})

onMounted(() => {
  fetchTopology()
})

const fetchTopology = async () => {
  loading.value = true
  try {
    topology.value = await topologyStore.fetchTopology(topologyId.value)
  } catch (error) {
    ElMessage.error('获取拓扑详情失败')
    console.error(error)
  } finally {
    loading.value = false
  }
}

const goBack = () => {
  router.push('/topology')
}

const handleNodeClick = (node: TopologyNode) => {
  console.log('Node clicked:', node)
}

const handleLinkClick = (link: TopologyLink) => {
  console.log('Link clicked:', link)
}

const handleNodePositionChange = async (nodeId: number, x: number, y: number) => {
  try {
    await topologyStore.updateNodePosition(topologyId.value, nodeId, x, y)
  } catch (error) {
    console.error('更新节点位置失败:', error)
  }
}

const handleAddNode = () => {
  Object.assign(nodeForm, {
    device_id: '',
    label: '',
    node_type: 'device',
    layer: 0
  })
  addNodeDialogVisible.value = true
}

const handleAddNodeSubmit = async () => {
  try {
    await topologyStore.addNode(topologyId.value, nodeForm)
    addNodeDialogVisible.value = false
    await fetchTopology()
  } catch (error) {
    console.error('添加节点失败:', error)
  }
}

const handleAddLink = () => {
  Object.assign(linkForm, {
    source_node_id: 0,
    target_node_id: 0,
    link_type: 'physical',
    source_interface: '',
    target_interface: '',
    bandwidth: 1000000000
  })
  addLinkDialogVisible.value = true
}

const handleAddLinkSubmit = async () => {
  if (!linkForm.source_node_id || !linkForm.target_node_id) {
    ElMessage.warning('请选择源节点和目标节点')
    return
  }
  
  if (linkForm.source_node_id === linkForm.target_node_id) {
    ElMessage.warning('源节点和目标节点不能相同')
    return
  }
  
  try {
    await topologyStore.addLink(topologyId.value, linkForm)
    addLinkDialogVisible.value = false
    await fetchTopology()
  } catch (error) {
    console.error('添加链路失败:', error)
  }
}

const handleSnapshot = async () => {
  try {
    const { value: description } = await ElMessageBox.prompt('请输入快照描述', '创建快照', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      inputPattern: /.+/,
      inputErrorMessage: '快照描述不能为空'
    })
    
    await topologyStore.createSnapshot(topologyId.value, description)
  } catch (error) {
    if (error !== 'cancel') {
      console.error('创建快照失败:', error)
    }
  }
}

const handleVersions = async () => {
  try {
    versions.value = await topologyStore.fetchVersions(topologyId.value)
    versionsDialogVisible.value = true
  } catch (error) {
    console.error('获取版本列表失败:', error)
  }
}

const handleRestore = async (version: number) => {
  try {
    await ElMessageBox.confirm(
      `确定要恢复到版本 ${version} 吗？当前拓扑将被覆盖。`,
      '恢复版本',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await topologyStore.restoreVersion(topologyId.value, version)
    versionsDialogVisible.value = false
    await fetchTopology()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('恢复版本失败:', error)
    }
  }
}

const handlePathAnalysis = async () => {
  ElMessage.info('路径分析功能开发中')
}

const handleImpactAnalysis = async () => {
  ElMessage.info('影响分析功能开发中')
}
</script>

<style scoped lang="scss">
.topology-detail {
  height: 100vh;
  display: flex;
  flex-direction: column;
  padding: 20px;

  .el-page-header {
    margin-bottom: 20px;
  }

  .topology-container {
    flex: 1;
    min-height: 0;
    background: white;
    border-radius: 4px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }

  .action-buttons {
    margin-top: 20px;
    text-align: center;
  }
}
</style>

