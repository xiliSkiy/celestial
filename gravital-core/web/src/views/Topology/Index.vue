<template>
  <div class="topology-page">
    <el-card class="search-card">
      <el-form :inline="true" :model="query">
        <el-form-item label="类型">
          <el-select v-model="query.type" placeholder="全部" clearable style="width: 150px">
            <el-option label="物理拓扑" value="physical" />
            <el-option label="逻辑拓扑" value="logical" />
            <el-option label="自定义拓扑" value="custom" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="范围">
          <el-select v-model="query.scope" placeholder="全部" clearable style="width: 150px">
            <el-option label="全局" value="global" />
            <el-option label="数据中心" value="datacenter" />
            <el-option label="区域" value="region" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="关键词">
          <el-input
            v-model="query.keyword"
            placeholder="搜索拓扑名称或描述"
            clearable
            style="width: 250px"
            @keyup.enter="handleSearch"
          />
        </el-form-item>
        
        <el-form-item>
          <el-button type="primary" @click="handleSearch">搜索</el-button>
          <el-button @click="handleReset">重置</el-button>
          <el-button type="success" @click="handleCreate">创建拓扑</el-button>
        </el-form-item>
      </el-form>
    </el-card>

    <el-card class="list-card">
      <el-table
        v-loading="loading"
        :data="topologies"
        stripe
        style="width: 100%"
      >
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="名称" min-width="200">
          <template #default="{ row }">
            <el-link type="primary" @click="handleView(row)">
              {{ row.name }}
            </el-link>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="120">
          <template #default="{ row }">
            <el-tag :type="getTypeTag(row.type)">
              {{ getTypeLabel(row.type) }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="scope" label="范围" width="120" />
        <el-table-column prop="node_count" label="节点数" width="100" />
        <el-table-column prop="link_count" label="链路数" width="100" />
        <el-table-column prop="created_at" label="创建时间" width="180">
          <template #default="{ row }">
            {{ formatDateTime(row.created_at) }}
          </template>
        </el-table-column>
        <el-table-column label="操作" width="200" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" @click="handleView(row)">查看</el-button>
            <el-button link type="primary" @click="handleEdit(row)">编辑</el-button>
            <el-button link type="danger" @click="handleDelete(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.page_size"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchData"
        @current-change="fetchData"
      />
    </el-card>

    <!-- 创建/编辑对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="600px"
      @close="handleDialogClose"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="formRules"
        label-width="120px"
      >
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入拓扑名称" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input
            v-model="form.description"
            type="textarea"
            :rows="3"
            placeholder="请输入拓扑描述"
          />
        </el-form-item>
        
        <el-form-item label="类型" prop="type">
          <el-select v-model="form.type" placeholder="请选择类型">
            <el-option label="物理拓扑" value="physical" />
            <el-option label="逻辑拓扑" value="logical" />
            <el-option label="自定义拓扑" value="custom" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="范围" prop="scope">
          <el-select v-model="form.scope" placeholder="请选择范围">
            <el-option label="全局" value="global" />
            <el-option label="数据中心" value="datacenter" />
            <el-option label="区域" value="region" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="布局类型" prop="layout_type">
          <el-select v-model="form.layout_type" placeholder="请选择布局类型">
            <el-option label="力导向" value="force" />
            <el-option label="层次布局" value="hierarchical" />
            <el-option label="环形布局" value="circular" />
            <el-option label="树形布局" value="tree" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="自动发现">
          <el-switch v-model="form.is_auto_discovery" />
        </el-form-item>
        
        <el-form-item v-if="form.is_auto_discovery" label="发现间隔(秒)">
          <el-input-number
            v-model="form.discovery_interval"
            :min="60"
            :max="86400"
            :step="60"
          />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import { useTopologyStore } from '@/stores/topology'
import type { TopologyQuery, CreateTopologyRequest, UpdateTopologyRequest, TopologyListItem } from '@/types/topology'
import { formatDateTime } from '@/utils/format'

const router = useRouter()
const topologyStore = useTopologyStore()

const loading = ref(false)
const topologies = ref<TopologyListItem[]>([])
const total = ref(0)

const query = reactive<TopologyQuery>({
  page: 1,
  page_size: 20,
  type: '',
  scope: '',
  keyword: ''
})

const dialogVisible = ref(false)
const dialogTitle = ref('创建拓扑')
const editingId = ref<number | null>(null)
const formRef = ref<FormInstance>()

const form = reactive<CreateTopologyRequest>({
  name: '',
  description: '',
  type: 'physical',
  scope: 'global',
  layout_type: 'force',
  is_auto_discovery: false,
  discovery_interval: 300
})

const formRules: FormRules = {
  name: [
    { required: true, message: '请输入拓扑名称', trigger: 'blur' }
  ],
  type: [
    { required: true, message: '请选择拓扑类型', trigger: 'change' }
  ]
}

onMounted(() => {
  fetchData()
})

const fetchData = async () => {
  loading.value = true
  try {
    await topologyStore.fetchTopologies(query)
    topologies.value = topologyStore.topologies
    total.value = topologyStore.total
  } finally {
    loading.value = false
  }
}

const handleSearch = () => {
  query.page = 1
  fetchData()
}

const handleReset = () => {
  query.type = ''
  query.scope = ''
  query.keyword = ''
  query.page = 1
  fetchData()
}

const handleCreate = () => {
  dialogTitle.value = '创建拓扑'
  editingId.value = null
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row: TopologyListItem) => {
  dialogTitle.value = '编辑拓扑'
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    type: row.type,
    scope: row.scope
  })
  dialogVisible.value = true
}

const handleView = (row: TopologyListItem) => {
  router.push(`/topology/${row.id}`)
}

const handleDelete = async (row: TopologyListItem) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除拓扑"${row.name}"吗？此操作不可恢复。`,
      '删除确认',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await topologyStore.deleteTopology(row.id)
    ElMessage.success('删除成功')
    fetchData()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除失败:', error)
    }
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    try {
      if (editingId.value) {
        await topologyStore.updateTopology(editingId.value, form as UpdateTopologyRequest)
      } else {
        await topologyStore.createTopology(form)
      }
      
      dialogVisible.value = false
      fetchData()
    } catch (error) {
      console.error('提交失败:', error)
    }
  })
}

const handleDialogClose = () => {
  formRef.value?.resetFields()
  resetForm()
}

const resetForm = () => {
  Object.assign(form, {
    name: '',
    description: '',
    type: 'physical',
    scope: 'global',
    layout_type: 'force',
    is_auto_discovery: false,
    discovery_interval: 300
  })
}

const getTypeTag = (type: string) => {
  const tags: Record<string, any> = {
    physical: 'primary',
    logical: 'success',
    custom: 'warning'
  }
  return tags[type] || 'info'
}

const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    physical: '物理拓扑',
    logical: '逻辑拓扑',
    custom: '自定义拓扑'
  }
  return labels[type] || type
}
</script>

<style scoped lang="scss">
.topology-page {
  padding: 20px;

  .search-card {
    margin-bottom: 20px;
  }

  .list-card {
    .el-pagination {
      margin-top: 20px;
      justify-content: flex-end;
    }
  }
}
</style>

