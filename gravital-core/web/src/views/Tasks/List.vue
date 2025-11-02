<template>
  <div class="tasks-list">
    <!-- 操作栏 -->
    <el-card class="toolbar-card">
      <div class="toolbar">
        <div class="toolbar-left">
          <el-button type="primary" :icon="Plus" @click="handleCreate">
            创建任务
          </el-button>
        </div>
        <div class="toolbar-right">
          <el-input
            v-model="query.keyword"
            placeholder="搜索任务"
            :prefix-icon="Search"
            style="width: 200px"
            clearable
            @change="fetchTasks"
          />
          <el-select
            v-model="query.enabled"
            placeholder="状态"
            style="width: 120px"
            clearable
            @change="fetchTasks"
          >
            <el-option label="全部" :value="undefined" />
            <el-option label="启用" :value="true" />
            <el-option label="禁用" :value="false" />
          </el-select>
        </div>
      </div>
    </el-card>

    <!-- 任务列表 -->
    <el-card class="table-card">
      <el-table :data="tasks" v-loading="loading" style="width: 100%">
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="任务名称" width="200" />
        <el-table-column prop="device_id" label="目标设备" width="180" />
        <el-table-column prop="sentinel_id" label="Sentinel" width="180" />
        <el-table-column prop="plugin_type" label="插件类型" width="120" />
        <el-table-column prop="interval" label="采集间隔" width="120" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <StatusBadge :status="row.enabled ? 'online' : 'offline'" />
          </template>
        </el-table-column>
        <el-table-column prop="last_execution" label="最后执行" width="180" />
        <el-table-column label="操作" width="250" fixed="right">
          <template #default="{ row }">
            <el-button 
              text 
              type="primary" 
              @click="handleTrigger(row)"
              :loading="triggeringId === row.id"
            >
              执行
            </el-button>
            <el-button text type="primary" @click="handleEdit(row)">
              编辑
            </el-button>
            <el-button 
              text 
              :type="row.enabled ? 'warning' : 'success'" 
              @click="handleToggle(row)"
            >
              {{ row.enabled ? '禁用' : '启用' }}
            </el-button>
            <el-button text type="danger" @click="handleDelete(row)">
              删除
            </el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 分页 -->
      <el-pagination
        v-model:current-page="query.page"
        v-model:page-size="query.size"
        :total="total"
        :page-sizes="[10, 20, 50, 100]"
        layout="total, sizes, prev, pager, next, jumper"
        @size-change="fetchTasks"
        @current-change="fetchTasks"
      />
    </el-card>

    <!-- 创建/编辑任务对话框 -->
    <el-dialog
      v-model="dialogVisible"
      :title="dialogTitle"
      width="700px"
      @close="resetForm"
    >
      <el-form
        ref="formRef"
        :model="form"
        :rules="rules"
        label-width="120px"
      >
        <el-form-item label="任务名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入任务名称" />
        </el-form-item>
        
        <el-form-item label="目标设备" prop="device_id">
          <el-select 
            v-model="form.device_id" 
            placeholder="请选择设备"
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="device in devices"
              :key="device.id"
              :label="`${device.name} (${device.device_id})`"
              :value="device.device_id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="Sentinel" prop="sentinel_id">
          <el-select 
            v-model="form.sentinel_id" 
            placeholder="请选择 Sentinel（可选）"
            clearable
            filterable
            style="width: 100%"
          >
            <el-option
              v-for="sentinel in sentinels"
              :key="sentinel.id"
              :label="sentinel.name"
              :value="sentinel.sentinel_id"
            />
          </el-select>
        </el-form-item>
        
        <el-form-item label="插件类型" prop="plugin_type">
          <el-select v-model="form.plugin_type" placeholder="请选择插件类型">
            <el-option label="Ping" value="ping" />
            <el-option label="SNMP" value="snmp" />
            <el-option label="SSH" value="ssh" />
            <el-option label="HTTP" value="http" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="采集间隔" prop="interval">
          <el-input v-model="form.interval" placeholder="例如: 30s, 1m, 5m">
            <template #append>秒/分钟</template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="超时时间" prop="timeout">
          <el-input v-model="form.timeout" placeholder="例如: 10s, 30s">
            <template #append>秒</template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="设备配置" prop="device_config">
          <el-input
            v-model="deviceConfigStr"
            type="textarea"
            :rows="4"
            placeholder='请输入 JSON 格式配置，例如: {"host": "192.168.1.1", "port": 22}'
          />
        </el-form-item>
        
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, computed } from 'vue'
import { taskApi, type TaskQuery, type TaskForm } from '@/api/task'
import { deviceApi } from '@/api/device'
import { sentinelApi } from '@/api/sentinel'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

const query = reactive<TaskQuery>({
  page: 1,
  size: 20,
  keyword: '',
  enabled: undefined
})

const loading = ref(false)
const tasks = ref<any[]>([])
const total = ref(0)
const devices = ref<any[]>([])
const sentinels = ref<any[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref('创建任务')
const formRef = ref<FormInstance>()
const currentTask = ref<any>(null)
const submitting = ref(false)
const triggeringId = ref<number | null>(null)

const form = reactive<TaskForm>({
  name: '',
  device_id: '',
  sentinel_id: '',
  plugin_type: '',
  device_config: {},
  interval: '60s',
  timeout: '30s',
  enabled: true,
  labels: {}
})

const deviceConfigStr = ref('{}')

const rules: FormRules = {
  name: [{ required: true, message: '请输入任务名称', trigger: 'blur' }],
  device_id: [{ required: true, message: '请选择目标设备', trigger: 'change' }],
  plugin_type: [{ required: true, message: '请选择插件类型', trigger: 'change' }],
  interval: [{ required: true, message: '请输入采集间隔', trigger: 'blur' }],
  timeout: [{ required: true, message: '请输入超时时间', trigger: 'blur' }]
}

// 获取任务列表
const fetchTasks = async () => {
  loading.value = true
  try {
    const res: any = await taskApi.getTasks(query)
    tasks.value = res.items || []
    total.value = res.total || 0
  } catch (error) {
    ElMessage.error('获取任务列表失败')
  } finally {
    loading.value = false
  }
}

// 获取设备列表
const fetchDevices = async () => {
  try {
    const res: any = await deviceApi.getDevices({ page: 1, size: 1000 })
    devices.value = res.items || []
  } catch (error) {
    console.error('获取设备列表失败', error)
  }
}

// 获取 Sentinel 列表
const fetchSentinels = async () => {
  try {
    const res: any = await sentinelApi.getSentinels({ page: 1, size: 1000 })
    sentinels.value = res.items || []
  } catch (error) {
    console.error('获取 Sentinel 列表失败', error)
  }
}

// 创建任务
const handleCreate = () => {
  dialogTitle.value = '创建任务'
  currentTask.value = null
  dialogVisible.value = true
}

// 编辑任务
const handleEdit = (row: any) => {
  dialogTitle.value = '编辑任务'
  currentTask.value = row
  Object.assign(form, {
    name: row.name,
    device_id: row.device_id,
    sentinel_id: row.sentinel_id || '',
    plugin_type: row.plugin_type,
    device_config: row.device_config || {},
    interval: row.interval,
    timeout: row.timeout,
    enabled: row.enabled,
    labels: row.labels || {}
  })
  deviceConfigStr.value = JSON.stringify(row.device_config || {}, null, 2)
  dialogVisible.value = true
}

// 删除任务
const handleDelete = (row: any) => {
  ElMessageBox.confirm('确定要删除该任务吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await taskApi.deleteTask(row.id)
      ElMessage.success('删除成功')
      fetchTasks()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 启用/禁用任务
const handleToggle = async (row: any) => {
  try {
    await taskApi.toggleTask(row.id, !row.enabled)
    ElMessage.success(row.enabled ? '已禁用' : '已启用')
    fetchTasks()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 手动触发任务
const handleTrigger = async (row: any) => {
  triggeringId.value = row.id
  try {
    await taskApi.triggerTask(row.id)
    ElMessage.success('任务已触发')
  } catch (error) {
    ElMessage.error('触发失败')
  } finally {
    triggeringId.value = null
  }
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 解析 device_config JSON
      try {
        form.device_config = JSON.parse(deviceConfigStr.value)
      } catch (error) {
        ElMessage.error('设备配置 JSON 格式错误')
        return
      }
      
      submitting.value = true
      try {
        if (currentTask.value) {
          await taskApi.updateTask(currentTask.value.id, form)
          ElMessage.success('更新成功')
        } else {
          await taskApi.createTask(form)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchTasks()
      } catch (error: any) {
        ElMessage.error(error.message || '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

// 重置表单
const resetForm = () => {
  formRef.value?.resetFields()
  Object.assign(form, {
    name: '',
    device_id: '',
    sentinel_id: '',
    plugin_type: '',
    device_config: {},
    interval: '60s',
    timeout: '30s',
    enabled: true,
    labels: {}
  })
  deviceConfigStr.value = '{}'
}

onMounted(() => {
  fetchTasks()
  fetchDevices()
  fetchSentinels()
})
</script>

<style scoped lang="scss">
.tasks-list {
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      align-items: center;
      justify-content: space-between;

      .toolbar-left,
      .toolbar-right {
        display: flex;
        gap: 10px;
      }
    }
  }

  .table-card {
    :deep(.el-pagination) {
      margin-top: 20px;
      justify-content: flex-end;
    }
  }
}
</style>

