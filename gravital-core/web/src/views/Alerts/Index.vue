<template>
  <div class="alerts">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 告警规则 -->
      <el-tab-pane label="告警规则" name="rules">
        <el-card class="toolbar-card">
          <div class="toolbar">
            <el-button type="primary" :icon="Plus" @click="handleCreateRule">
              创建规则
            </el-button>
            <el-input
              v-model="ruleQuery.keyword"
              placeholder="搜索规则"
              :prefix-icon="Search"
              style="width: 200px"
              clearable
              @change="fetchRules"
            />
            <el-select
              v-model="ruleQuery.severity"
              placeholder="级别"
              style="width: 120px"
              clearable
              @change="fetchRules"
            >
              <el-option label="全部" value="" />
              <el-option label="Critical" value="critical" />
              <el-option label="Warning" value="warning" />
              <el-option label="Info" value="info" />
            </el-select>
          </div>
        </el-card>

        <el-card class="table-card">
          <el-table :data="rules" v-loading="loading">
            <el-table-column prop="name" label="规则名称" width="200" />
            <el-table-column label="级别" width="120">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)">
                  {{ row.severity }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="metric_name" label="指标" width="150" />
            <el-table-column label="条件" width="150">
              <template #default="{ row }">
                {{ row.operator }} {{ row.threshold }}
              </template>
            </el-table-column>
            <el-table-column prop="duration" label="持续时间" width="120" />
            <el-table-column prop="description" label="描述" />
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <StatusBadge :status="row.enabled ? 'online' : 'offline'" />
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200" fixed="right">
              <template #default="{ row }">
                <el-button text type="primary" @click="handleEditRule(row)">
                  编辑
                </el-button>
                <el-button 
                  text 
                  :type="row.enabled ? 'warning' : 'success'" 
                  @click="handleToggleRule(row)"
                >
                  {{ row.enabled ? '禁用' : '启用' }}
                </el-button>
                <el-button text type="danger" @click="handleDeleteRule(row)">
                  删除
                </el-button>
              </template>
            </el-table-column>
          </el-table>

          <el-pagination
            v-model:current-page="ruleQuery.page"
            v-model:page-size="ruleQuery.size"
            :total="ruleTotal"
            :page-sizes="[10, 20, 50, 100]"
            layout="total, sizes, prev, pager, next, jumper"
            @size-change="fetchRules"
            @current-change="fetchRules"
          />
        </el-card>
      </el-tab-pane>

      <!-- 告警事件 -->
      <el-tab-pane label="告警事件" name="events">
        <el-card class="toolbar-card">
          <div class="toolbar">
            <el-select
              v-model="eventQuery.severity"
              placeholder="级别"
              style="width: 120px"
              clearable
              @change="fetchEvents"
            >
              <el-option label="全部" value="" />
              <el-option label="Critical" value="critical" />
              <el-option label="Warning" value="warning" />
              <el-option label="Info" value="info" />
            </el-select>
            <el-select
              v-model="eventQuery.status"
              placeholder="状态"
              style="width: 120px"
              clearable
              @change="fetchEvents"
            >
              <el-option label="全部" value="" />
              <el-option label="告警中" value="firing" />
              <el-option label="已解决" value="resolved" />
              <el-option label="已确认" value="acknowledged" />
            </el-select>
          </div>
        </el-card>

        <el-card class="timeline-card">
          <el-timeline v-loading="loading">
            <el-timeline-item
              v-for="event in events"
              :key="event.id"
              :timestamp="event.triggered_at"
              :type="getEventType(event.severity)"
            >
              <div class="event-item">
                <div class="event-header">
                  <el-tag :type="getSeverityType(event.severity)">
                    {{ event.severity }}
                  </el-tag>
                  <span class="event-title">{{ event.message }}</span>
                  <el-tag v-if="event.status" :type="getStatusType(event.status)" size="small">
                    {{ getStatusText(event.status) }}
                  </el-tag>
                </div>
                <div class="event-content">
                  <p><strong>设备:</strong> {{ event.device_name || event.device_id }}</p>
                  <p><strong>指标:</strong> {{ event.metric_name }}</p>
                  <p><strong>当前值:</strong> {{ event.current_value }}</p>
                  <p v-if="event.resolved_at"><strong>解决时间:</strong> {{ event.resolved_at }}</p>
                </div>
                <div class="event-actions">
                  <el-button 
                    v-if="event.status === 'firing'"
                    size="small" 
                    @click="handleAcknowledge(event)"
                    :loading="acknowledgingId === event.id"
                  >
                    确认
                  </el-button>
                  <el-button 
                    v-if="event.status !== 'resolved'"
                    size="small" 
                    type="success"
                    @click="handleResolve(event)"
                    :loading="resolvingId === event.id"
                  >
                    解决
                  </el-button>
                </div>
              </div>
            </el-timeline-item>
          </el-timeline>

          <el-pagination
            v-model:current-page="eventQuery.page"
            v-model:page-size="eventQuery.size"
            :total="eventTotal"
            :page-sizes="[10, 20, 50]"
            layout="total, sizes, prev, pager, next"
            @size-change="fetchEvents"
            @current-change="fetchEvents"
          />
        </el-card>
      </el-tab-pane>
    </el-tabs>

    <!-- 创建/编辑规则对话框 -->
    <el-dialog
      v-model="ruleDialogVisible"
      :title="ruleDialogTitle"
      width="600px"
      @close="resetRuleForm"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="name">
          <el-input v-model="ruleForm.name" placeholder="请输入规则名称" />
        </el-form-item>
        
        <el-form-item label="描述" prop="description">
          <el-input 
            v-model="ruleForm.description" 
            type="textarea"
            :rows="2"
            placeholder="请输入描述"
          />
        </el-form-item>
        
        <el-form-item label="级别" prop="severity">
          <el-select v-model="ruleForm.severity" placeholder="请选择级别">
            <el-option label="Critical" value="critical" />
            <el-option label="Warning" value="warning" />
            <el-option label="Info" value="info" />
          </el-select>
        </el-form-item>
        
        <el-form-item label="指标名称" prop="metric_name">
          <el-input v-model="ruleForm.metric_name" placeholder="例如: cpu_usage" />
        </el-form-item>
        
        <el-form-item label="条件" prop="operator">
          <el-row :gutter="10">
            <el-col :span="8">
              <el-select v-model="ruleForm.operator" placeholder="运算符">
                <el-option label=">" value=">" />
                <el-option label="<" value="<" />
                <el-option label=">=" value=">=" />
                <el-option label="<=" value="<=" />
                <el-option label="==" value="==" />
                <el-option label="!=" value="!=" />
              </el-select>
            </el-col>
            <el-col :span="16">
              <el-input-number 
                v-model="ruleForm.threshold" 
                placeholder="阈值"
                style="width: 100%"
              />
            </el-col>
          </el-row>
        </el-form-item>
        
        <el-form-item label="持续时间" prop="duration">
          <el-input v-model="ruleForm.duration" placeholder="例如: 5m, 10m">
            <template #append>分钟</template>
          </el-input>
        </el-form-item>
        
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
      </el-form>
      
      <template #footer>
        <el-button @click="ruleDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="handleSubmitRule" :loading="submitting">
          确定
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { alertApi } from '@/api/alert'
import StatusBadge from '@/components/common/StatusBadge.vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'

const activeTab = ref('rules')
const loading = ref(false)
const submitting = ref(false)
const acknowledgingId = ref<number | null>(null)
const resolvingId = ref<number | null>(null)

// 规则相关
const ruleQuery = reactive({
  page: 1,
  size: 20,
  keyword: '',
  severity: '',
  enabled: undefined
})

const rules = ref<any[]>([])
const ruleTotal = ref(0)
const ruleDialogVisible = ref(false)
const ruleDialogTitle = ref('创建规则')
const ruleFormRef = ref<FormInstance>()
const currentRule = ref<any>(null)

const ruleForm = reactive({
  name: '',
  description: '',
  severity: 'warning' as 'critical' | 'warning' | 'info',
  metric_name: '',
  operator: '>' as '>' | '<' | '>=' | '<=' | '==' | '!=',
  threshold: 0,
  duration: '5m',
  enabled: true,
  labels: {}
})

const ruleRules: FormRules = {
  name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
  severity: [{ required: true, message: '请选择级别', trigger: 'change' }],
  metric_name: [{ required: true, message: '请输入指标名称', trigger: 'blur' }],
  operator: [{ required: true, message: '请选择运算符', trigger: 'change' }],
  threshold: [{ required: true, message: '请输入阈值', trigger: 'blur' }],
  duration: [{ required: true, message: '请输入持续时间', trigger: 'blur' }]
}

// 事件相关
const eventQuery = reactive({
  page: 1,
  size: 20,
  severity: '',
  status: ''
})

const events = ref<any[]>([])
const eventTotal = ref(0)

// 获取规则列表
const fetchRules = async () => {
  loading.value = true
  try {
    const res: any = await alertApi.getRules(ruleQuery)
    rules.value = res.items || []
    ruleTotal.value = res.total || 0
  } catch (error) {
    ElMessage.error('获取规则列表失败')
  } finally {
    loading.value = false
  }
}

// 获取事件列表
const fetchEvents = async () => {
  loading.value = true
  try {
    const res: any = await alertApi.getEvents(eventQuery)
    events.value = res.items || []
    eventTotal.value = res.total || 0
  } catch (error) {
    ElMessage.error('获取事件列表失败')
  } finally {
    loading.value = false
  }
}

// Tab 切换
const handleTabChange = (tab: string) => {
  if (tab === 'rules') {
    fetchRules()
  } else if (tab === 'events') {
    fetchEvents()
  }
}

// 创建规则
const handleCreateRule = () => {
  ruleDialogTitle.value = '创建规则'
  currentRule.value = null
  ruleDialogVisible.value = true
}

// 编辑规则
const handleEditRule = (row: any) => {
  ruleDialogTitle.value = '编辑规则'
  currentRule.value = row
  Object.assign(ruleForm, {
    name: row.name,
    description: row.description || '',
    severity: row.severity,
    metric_name: row.metric_name,
    operator: row.operator,
    threshold: row.threshold,
    duration: row.duration,
    enabled: row.enabled,
    labels: row.labels || {}
  })
  ruleDialogVisible.value = true
}

// 删除规则
const handleDeleteRule = (row: any) => {
  ElMessageBox.confirm('确定要删除该规则吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    try {
      await alertApi.deleteRule(row.id)
      ElMessage.success('删除成功')
      fetchRules()
    } catch (error) {
      ElMessage.error('删除失败')
    }
  })
}

// 启用/禁用规则
const handleToggleRule = async (row: any) => {
  try {
    await alertApi.toggleRule(row.id, !row.enabled)
    ElMessage.success(row.enabled ? '已禁用' : '已启用')
    fetchRules()
  } catch (error) {
    ElMessage.error('操作失败')
  }
}

// 提交规则
const handleSubmitRule = async () => {
  if (!ruleFormRef.value) return
  
  await ruleFormRef.value.validate(async (valid) => {
    if (valid) {
      submitting.value = true
      try {
        if (currentRule.value) {
          await alertApi.updateRule(currentRule.value.id, ruleForm)
          ElMessage.success('更新成功')
        } else {
          await alertApi.createRule(ruleForm)
          ElMessage.success('创建成功')
        }
        ruleDialogVisible.value = false
        fetchRules()
      } catch (error: any) {
        ElMessage.error(error.message || '操作失败')
      } finally {
        submitting.value = false
      }
    }
  })
}

// 重置规则表单
const resetRuleForm = () => {
  ruleFormRef.value?.resetFields()
  Object.assign(ruleForm, {
    name: '',
    description: '',
    severity: 'warning',
    metric_name: '',
    operator: '>',
    threshold: 0,
    duration: '5m',
    enabled: true,
    labels: {}
  })
}

// 确认告警
const handleAcknowledge = async (event: any) => {
  acknowledgingId.value = event.id
  try {
    await alertApi.acknowledgeEvent(event.id)
    ElMessage.success('已确认')
    fetchEvents()
  } catch (error) {
    ElMessage.error('确认失败')
  } finally {
    acknowledgingId.value = null
  }
}

// 解决告警
const handleResolve = async (event: any) => {
  resolvingId.value = event.id
  try {
    await alertApi.resolveEvent(event.id)
    ElMessage.success('已解决')
    fetchEvents()
  } catch (error) {
    ElMessage.error('解决失败')
  } finally {
    resolvingId.value = null
  }
}

// 辅助函数
const getSeverityType = (severity: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info'
  }
  return map[severity] || 'info'
}

const getEventType = (severity: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    warning: 'warning',
    info: 'primary'
  }
  return map[severity] || 'primary'
}

const getStatusType = (status: string) => {
  const map: Record<string, any> = {
    firing: 'danger',
    acknowledged: 'warning',
    resolved: 'success'
  }
  return map[status] || 'info'
}

const getStatusText = (status: string) => {
  const map: Record<string, string> = {
    firing: '告警中',
    acknowledged: '已确认',
    resolved: '已解决'
  }
  return map[status] || status
}

onMounted(() => {
  fetchRules()
})
</script>

<style scoped lang="scss">
.alerts {
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      gap: 10px;
    }
  }

  .table-card {
    :deep(.el-pagination) {
      margin-top: 20px;
      justify-content: flex-end;
    }
  }

  .timeline-card {
    :deep(.el-pagination) {
      margin-top: 20px;
      justify-content: center;
    }

    .event-item {
      .event-header {
        display: flex;
        align-items: center;
        gap: 10px;
        margin-bottom: 10px;

        .event-title {
          font-weight: 600;
          flex: 1;
        }
      }

      .event-content {
        margin: 10px 0;
        color: var(--text-secondary);

        p {
          margin: 5px 0;
        }
      }

      .event-actions {
        display: flex;
        gap: 8px;
      }
    }
  }
}
</style>
