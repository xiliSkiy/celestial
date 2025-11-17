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
        <el-table-column prop="name" label="任务名称" width="200" show-overflow-tooltip />
        <el-table-column label="目标设备" width="180">
          <template #default="{ row }">
            {{ row.device?.name || row.device_id || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="sentinel_id" label="Sentinel" width="180" show-overflow-tooltip />
        <el-table-column prop="plugin_type" label="插件类型" width="120">
          <template #default="{ row }">
            {{ row.plugin_type || '-' }}
          </template>
        </el-table-column>
        <el-table-column prop="interval" label="采集间隔" width="120">
          <template #default="{ row }">
            {{ row.interval || '-' }}
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <StatusBadge :status="row.enabled ? 'online' : 'offline'" />
          </template>
        </el-table-column>
        <el-table-column prop="last_execution" label="最后执行" width="180">
          <template #default="{ row }">
            {{ row.last_execution || '-' }}
          </template>
        </el-table-column>
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
  sentinel_id: [{ required: true, message: '请选择 Sentinel', trigger: 'change' }],
  plugin_type: [{ required: true, message: '请选择插件类型', trigger: 'change' }],
  interval: [{ required: true, message: '请输入采集间隔', trigger: 'blur' }],
  timeout: [{ required: true, message: '请输入超时时间', trigger: 'blur' }]
}

// 将秒数转换为时间字符串（如 60 -> "60s", 120 -> "2m"）
const formatSecondsToDuration = (seconds: number | string): string => {
  if (typeof seconds === 'string') {
    // 如果已经是字符串，直接返回
    return seconds
  }
  if (!seconds || seconds <= 0) return '60s'
  
  if (seconds < 60) {
    return `${seconds}s`
  } else if (seconds < 3600) {
    const minutes = Math.floor(seconds / 60)
    return `${minutes}m`
  } else {
    const hours = Math.floor(seconds / 3600)
    return `${hours}h`
  }
}

// 格式化日期时间
const formatDateTime = (date: string | null | undefined): string => {
  if (!date) return '-'
  try {
    const d = new Date(date)
    return d.toLocaleString('zh-CN', {
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      second: '2-digit'
    })
  } catch (error) {
    return '-'
  }
}

// 获取任务列表
const fetchTasks = async () => {
  loading.value = true
  try {
    const res: any = await taskApi.getTasks(query)
    // 转换数据格式以匹配前端表格列
    tasks.value = (res.items || []).map((task: any) => ({
      ...task,
      // 任务名称：使用设备名称 + 插件类型，或使用 task_id
      name: task.device?.name 
        ? `${task.device.name} - ${task.plugin_name || '未知'}`
        : `任务 ${task.task_id}`,
      // 插件类型：后端返回 plugin_name，前端期望 plugin_type
      plugin_type: task.plugin_name || task.plugin_type || '-',
      // 采集间隔：后端返回 interval_seconds (数字)，前端期望 interval (字符串)
      interval: formatSecondsToDuration(task.interval_seconds || task.interval || 60),
      // 最后执行：后端返回 last_executed_at，前端期望 last_execution (格式化)
      last_execution: formatDateTime(task.last_executed_at),
      // 目标设备：显示设备名称
      device_name: task.device?.name || task.device_id || '-'
    }))
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
    name: row.name || '',
    device_id: row.device_id,
    sentinel_id: row.sentinel_id || '',
    plugin_type: row.plugin_name || row.plugin_type,  // 后端返回 plugin_name
    device_config: row.config || row.device_config || {},  // 后端返回 config
    interval: formatSecondsToDuration(row.interval_seconds || row.interval || 60),  // 后端返回 interval_seconds (数字)
    timeout: formatSecondsToDuration(row.timeout_seconds || row.timeout || 30),  // 后端返回 timeout_seconds (数字)
    enabled: row.enabled,
    labels: row.labels || {}
  })
  deviceConfigStr.value = JSON.stringify(row.config || row.device_config || {}, null, 2)
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

// 解析时间字符串为秒数（如 "60s" -> 60, "1m" -> 60）
const parseDurationToSeconds = (duration: string): number => {
  if (!duration) return 60 // 默认值
  const match = duration.match(/^(\d+)([smh])?$/)
  if (!match) return 60
  const value = parseInt(match[1])
  const unit = match[2] || 's'
  switch (unit) {
    case 's': return value
    case 'm': return value * 60
    case 'h': return value * 3600
    default: return value
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
          // 更新任务：只发送 UpdateTaskRequest 需要的字段
          const updateData = {
            config: form.device_config,  // device_config -> config
            interval_seconds: parseDurationToSeconds(form.interval || '60s'),  // interval -> interval_seconds (数字)
            enabled: form.enabled
          }
          await taskApi.updateTask(currentTask.value.id, updateData as any)
          ElMessage.success('更新成功')
        } else {
          // 创建任务：发送 CreateTaskRequest 需要的字段
          if (!form.sentinel_id) {
            ElMessage.error('请选择 Sentinel')
            submitting.value = false
            return
          }
          const createData = {
            device_id: form.device_id,
            sentinel_id: form.sentinel_id,
            plugin_name: form.plugin_type,  // plugin_type -> plugin_name
            config: form.device_config,  // device_config -> config
            interval_seconds: parseDurationToSeconds(form.interval || '60s'),  // interval -> interval_seconds (数字)
            timeout_seconds: parseDurationToSeconds(form.timeout || '30s'),  // timeout -> timeout_seconds (数字)
            enabled: form.enabled
          }
          await taskApi.createTask(createData as any)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchTasks()
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || error.message || '操作失败')
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

