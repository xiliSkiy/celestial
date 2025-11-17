<template>
  <div class="forwarders-list">
    <!-- 操作栏 -->
    <el-card class="toolbar-card">
      <div class="toolbar">
        <el-button type="primary" :icon="Plus" @click="handleCreate">
          添加转发器
        </el-button>
        <el-button 
          :icon="Refresh" 
          @click="handleReload"
          :loading="reloading"
        >
          重新加载配置
        </el-button>
      </div>
    </el-card>

    <!-- 转发器卡片网格 -->
    <div v-loading="loading" class="forwarder-grid">
      <el-card
        v-for="forwarder in forwarders"
        :key="forwarder.id"
        class="forwarder-card"
        shadow="hover"
      >
        <div class="forwarder-header">
          <h3>{{ forwarder.name }}</h3>
          <StatusBadge :status="forwarder.enabled ? 'online' : 'offline'" />
        </div>
        
        <div class="forwarder-body">
          <div class="forwarder-info">
            <div class="info-item">
              <span class="label">类型:</span>
              <span class="value">{{ getTypeLabel(forwarder.type) }}</span>
            </div>
            <div class="info-item">
              <span class="label">端点:</span>
              <span class="value">{{ forwarder.endpoint }}</span>
            </div>
            <div class="info-item">
              <span class="label">批次大小:</span>
              <span class="value">{{ forwarder.batch_size || 1000 }}</span>
            </div>
            <div class="info-item">
              <span class="label">刷新间隔:</span>
              <span class="value">{{ forwarder.flush_interval || '10s' }}</span>
            </div>
          </div>
          
          <el-divider />
          
          <div class="forwarder-stats">
            <div class="stat-item">
              <div class="stat-label">成功</div>
              <div class="stat-value success">
                {{ formatNumber(forwarder.success_count || 0) }}
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-label">失败</div>
              <div class="stat-value danger">
                {{ formatNumber(forwarder.failure_count || 0) }}
              </div>
            </div>
            <div class="stat-item">
              <div class="stat-label">延迟</div>
              <div class="stat-value">
                {{ forwarder.avg_latency || 0 }}ms
              </div>
            </div>
          </div>
        </div>
        
        <div class="forwarder-footer">
          <el-button 
            text 
            type="primary" 
            @click="handleToggle(forwarder)"
          >
            {{ forwarder.enabled ? '禁用' : '启用' }}
          </el-button>
          <el-button text type="primary" @click="handleEdit(forwarder)">
            编辑
          </el-button>
          <el-button text type="danger" @click="handleDelete(forwarder)">
            删除
          </el-button>
        </div>
      </el-card>
    </div>

    <!-- 创建/编辑转发器对话框 -->
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
        <el-form-item label="名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入转发器名称" />
        </el-form-item>
        
        <el-form-item label="类型" prop="type">
          <el-select 
            v-model="form.type" 
            placeholder="请选择类型"
            @change="handleTypeChange"
          >
            <el-option label="Prometheus" value="prometheus" />
            <el-option label="VictoriaMetrics" value="victoria-metrics" />
            <el-option label="ClickHouse" value="clickhouse" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="端点" prop="endpoint">
          <el-input 
            v-model="form.endpoint" 
            placeholder="例如: http://localhost:8428/api/v1/write"
          />
        </el-form-item>
        
        <el-form-item label="批次大小" prop="batch_size">
          <el-input-number 
            v-model="form.batch_size" 
            :min="100"
            :max="10000"
            :step="100"
            style="width: 100%"
          />
        </el-form-item>
        
        <el-form-item label="刷新间隔" prop="flush_interval">
          <el-input v-model="form.flush_interval" placeholder="例如: 10s, 1m">
            <template #append>秒</template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="认证配置">
          <el-input
            v-model="authConfigStr"
            type="textarea"
            :rows="3"
            placeholder='可选，JSON 格式，例如: {"username": "admin", "password": "secret"}'
          />
        </el-form-item>
        
        <el-form-item label="TLS 配置">
          <el-input
            v-model="tlsConfigStr"
            type="textarea"
            :rows="3"
            placeholder='可选，JSON 格式，例如: {"insecure_skip_verify": true}'
          />
        </el-form-item>
        
        <el-form-item label="启用">
          <el-switch v-model="form.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button 
          type="primary" 
          @click="handleTest"
          :loading="testing"
        >
          测试连接
        </el-button>
        <el-button 
          type="primary" 
          @click="handleSubmit"
          :loading="submitting"
        >
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { forwarderApi, type ForwarderForm } from '@/api/forwarder'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Plus, Refresh } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

const loading = ref(false)
const reloading = ref(false)
const submitting = ref(false)
const testing = ref(false)
const forwarders = ref<any[]>([])

const dialogVisible = ref(false)
const dialogTitle = ref('添加转发器')
const formRef = ref<FormInstance>()
const currentForwarder = ref<any>(null)

const form = reactive<ForwarderForm>({
  name: '',
  type: 'victoria-metrics',
  endpoint: '',
  enabled: true,
  batch_size: 1000,
  flush_interval: '10s',
  auth_config: {},
  tls_config: {}
})

const authConfigStr = ref('{}')
const tlsConfigStr = ref('{}')

const rules: FormRules = {
  name: [{ required: true, message: '请输入转发器名称', trigger: 'blur' }],
  type: [{ required: true, message: '请选择类型', trigger: 'change' }],
  endpoint: [{ required: true, message: '请输入端点地址', trigger: 'blur' }],
  batch_size: [{ required: true, message: '请输入批次大小', trigger: 'blur' }],
  flush_interval: [{ required: true, message: '请输入刷新间隔', trigger: 'blur' }]
}

// 获取转发器列表
const fetchForwarders = async () => {
  loading.value = true
  try {
    const res: any = await forwarderApi.getForwarders()
    forwarders.value = res.items || res || []
  } catch (error) {
    ElMessage.error('获取转发器列表失败')
  } finally {
    loading.value = false
  }
}

// 重新加载配置
const handleReload = async () => {
  reloading.value = true
  try {
    await forwarderApi.reloadConfig()
    ElMessage.success('配置已重新加载')
    fetchForwarders()
  } catch (error) {
    ElMessage.error('重新加载失败')
  } finally {
    reloading.value = false
  }
}

// 创建转发器
const handleCreate = () => {
  dialogTitle.value = '添加转发器'
  currentForwarder.value = null
  dialogVisible.value = true
}

// 编辑转发器
const handleEdit = (forwarder: any) => {
  dialogTitle.value = '编辑转发器'
  currentForwarder.value = forwarder
  Object.assign(form, {
    name: forwarder.name,
    type: forwarder.type,
    endpoint: forwarder.endpoint,
    enabled: forwarder.enabled,
    batch_size: forwarder.batch_size || 1000,
    flush_interval: formatSecondsToDuration(forwarder.flush_interval || 10),
    auth_config: forwarder.auth_config || {},
    tls_config: forwarder.tls_config || {}
  })
  authConfigStr.value = JSON.stringify(forwarder.auth_config || {}, null, 2)
  tlsConfigStr.value = JSON.stringify(forwarder.tls_config || {}, null, 2)
  dialogVisible.value = true
}

// 删除转发器
const handleDelete = (forwarder: any) => {
  ElMessageBox.confirm('确定要删除该转发器吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await forwarderApi.deleteForwarder(forwarder.name)
      ElMessage.success('删除成功')
      fetchForwarders()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 启用/禁用转发器
const handleToggle = async (forwarder: any) => {
  try {
    // 获取完整配置，然后更新 enabled 字段
    const res: any = await forwarderApi.getForwarder(forwarder.name)
    const fullConfig = res.data || res
    const updateData = {
      ...fullConfig,
      enabled: !forwarder.enabled,
      // 确保 flush_interval 是数字
      flush_interval: typeof fullConfig.flush_interval === 'string' 
        ? parseDurationToSeconds(fullConfig.flush_interval) 
        : fullConfig.flush_interval
    }
    await forwarderApi.updateForwarder(forwarder.name, updateData)
    ElMessage.success(forwarder.enabled ? '已禁用' : '已启用')
    fetchForwarders()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 类型变更
const handleTypeChange = (type: string) => {
  // 根据类型设置默认端点
  const defaultEndpoints: Record<string, string> = {
    'prometheus': 'http://localhost:9090/api/v1/write',
    'victoria-metrics': 'http://localhost:8428/api/v1/write',
    'clickhouse': 'tcp://localhost:9000/default'
  }
  if (!form.endpoint || form.endpoint === '') {
    form.endpoint = defaultEndpoints[type] || ''
  }
}

// 解析时间字符串为秒数（如 "10s" -> 10, "1m" -> 60）
const parseDurationToSeconds = (duration: string): number => {
  if (!duration) return 10 // 默认值
  const match = duration.match(/^(\d+)([smh])?$/)
  if (!match) return 10
  const value = parseInt(match[1])
  const unit = match[2] || 's'
  switch (unit) {
    case 's': return value
    case 'm': return value * 60
    case 'h': return value * 3600
    default: return value
  }
}

// 将秒数转换为时间字符串（如 10 -> "10s", 60 -> "1m"）
const formatSecondsToDuration = (seconds: number | string): string => {
  if (typeof seconds === 'string') {
    // 如果已经是字符串，直接返回
    return seconds
  }
  if (!seconds || seconds <= 0) return '10s'
  
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

// 测试连接
const handleTest = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 解析配置
      try {
        form.auth_config = JSON.parse(authConfigStr.value)
        form.tls_config = JSON.parse(tlsConfigStr.value)
      } catch (error) {
        ElMessage.error('配置 JSON 格式错误')
        return
      }
      
      testing.value = true
      try {
        // 准备测试数据（转换格式以匹配后端）
        const testData = {
          name: form.name || 'test',
          type: form.type,
          endpoint: form.endpoint,
          enabled: form.enabled,
          batch_size: form.batch_size || 1000,
          flush_interval: parseDurationToSeconds(form.flush_interval || '10s'),
          retry_times: 3,
          timeout_seconds: 30,
          auth_config: form.auth_config || {}
        }
        
        const result = await forwarderApi.testConnection(testData as any)
        if (result.success) {
          ElMessage.success(result.message || '连接测试成功')
        } else {
          ElMessage.error(result.message || '连接测试失败')
        }
      } catch (error: any) {
        ElMessage.error(error.response?.data?.message || error.message || '连接测试失败')
      } finally {
        testing.value = false
      }
    }
  })
}

// 提交表单
const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (valid) {
      // 解析配置
      try {
        form.auth_config = JSON.parse(authConfigStr.value)
        form.tls_config = JSON.parse(tlsConfigStr.value)
      } catch (error) {
        ElMessage.error('配置 JSON 格式错误')
        return
      }
      
      submitting.value = true
      try {
        // 准备提交数据（转换格式以匹配后端）
        const submitData = {
          name: form.name,
          type: form.type,
          endpoint: form.endpoint,
          enabled: form.enabled,
          batch_size: form.batch_size || 1000,
          flush_interval: parseDurationToSeconds(form.flush_interval || '10s'),
          retry_times: 3,
          timeout_seconds: 30,
          auth_config: form.auth_config || {}
        }
        
        if (currentForwarder.value) {
          await forwarderApi.updateForwarder(currentForwarder.value.name, submitData as any)
          ElMessage.success('更新成功')
        } else {
          await forwarderApi.createForwarder(submitData as any)
          ElMessage.success('创建成功')
        }
        dialogVisible.value = false
        fetchForwarders()
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
    type: 'victoria-metrics',
    endpoint: '',
    enabled: true,
    batch_size: 1000,
    flush_interval: '10s',
    auth_config: {},
    tls_config: {}
  })
  authConfigStr.value = '{}'
  tlsConfigStr.value = '{}'
}

// 辅助函数
const getTypeLabel = (type: string) => {
  const labels: Record<string, string> = {
    'prometheus': 'Prometheus',
    'victoria-metrics': 'VictoriaMetrics',
    'clickhouse': 'ClickHouse'
  }
  return labels[type] || type
}

const formatNumber = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + 'M'
  } else if (num >= 1000) {
    return (num / 1000).toFixed(1) + 'K'
  }
  return num.toString()
}

onMounted(() => {
  fetchForwarders()
})
</script>

<style scoped lang="scss">
.forwarders-list {
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      gap: 10px;
    }
  }

  .forwarder-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(400px, 1fr));
    gap: 20px;
    margin-bottom: 20px;

    .forwarder-card {
      .forwarder-header {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 16px;

        h3 {
          margin: 0;
          font-size: 18px;
          font-weight: 600;
        }
      }

      .forwarder-body {
        .forwarder-info {
          .info-item {
            display: flex;
            margin-bottom: 8px;

            .label {
              width: 100px;
              color: var(--text-secondary);
            }

            .value {
              flex: 1;
              color: var(--text-primary);
              word-break: break-all;
            }
          }
        }

        .forwarder-stats {
          display: grid;
          grid-template-columns: repeat(3, 1fr);
          gap: 16px;

          .stat-item {
            text-align: center;

            .stat-label {
              font-size: 12px;
              color: var(--text-secondary);
              margin-bottom: 4px;
            }

            .stat-value {
              font-size: 18px;
              font-weight: 600;
              color: var(--text-primary);

              &.success {
                color: #67c23a;
              }

              &.danger {
                color: #f56c6c;
              }
            }
          }
        }
      }

      .forwarder-footer {
        display: flex;
        justify-content: flex-end;
        gap: 8px;
        margin-top: 16px;
        padding-top: 16px;
        border-top: 1px solid var(--border-color);
      }
    }
  }

  .buffer-card {
    .buffer-info {
      display: flex;
      justify-content: space-between;
      margin-top: 10px;
      font-size: 14px;
      color: var(--text-secondary);
    }
  }
}
</style>
