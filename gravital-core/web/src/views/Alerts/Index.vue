<template>
  <div class="alerts">
    <el-tabs v-model="activeTab" @tab-change="handleTabChange">
      <!-- 告警概览 -->
      <el-tab-pane label="告警概览" name="overview">
        <el-card v-loading="loading">
          <div v-if="aggregations.length === 0" class="empty-state">
            <el-empty description="暂无活跃告警" />
          </div>
          
          <div v-else class="aggregation-list">
            <el-card 
              v-for="agg in aggregations" 
              :key="agg.rule_id"
              class="aggregation-card"
              shadow="hover"
            >
              <div class="agg-header">
                <el-tag :type="getSeverityType(agg.severity)" size="large">
                  {{ agg.severity }}
                </el-tag>
                <h3>{{ agg.rule_name }}</h3>
                <el-badge :value="agg.firing_count" :type="agg.firing_count > 0 ? 'danger' : 'info'" />
              </div>
              
              <div class="agg-description">
                {{ agg.description || '无描述' }}
              </div>
              
              <div class="agg-stats">
                <div class="stat-item">
                  <span class="stat-label">总计</span>
                  <span class="stat-value">{{ agg.total_count }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">告警中</span>
                  <span class="stat-value danger">{{ agg.firing_count }}</span>
                </div>
                <div class="stat-item">
                  <span class="stat-label">已确认</span>
                  <span class="stat-value warning">{{ agg.acked_count }}</span>
                </div>
              </div>
              
              <div class="agg-time">
                <span>首次: {{ agg.first_fired }}</span>
                <span>最近: {{ agg.last_fired }}</span>
              </div>
              
              <div class="agg-actions">
                <el-button size="small" @click="viewRuleEvents(agg.rule_id)">
                  查看详情
                </el-button>
                <el-button 
                  size="small" 
                  type="success" 
                  @click="handleResolveByRule(agg.rule_id)"
                  :loading="resolvingRuleId === agg.rule_id"
                >
                  全部解决
                </el-button>
              </div>
              
              <!-- 展开显示受影响的设备 -->
              <el-collapse v-model="expandedRules" class="device-collapse">
                <el-collapse-item :name="agg.rule_id">
                  <template #title>
                    <span class="collapse-title">
                      受影响设备 ({{ agg.devices.length }})
                    </span>
                  </template>
                  <el-table :data="agg.devices" size="small">
                    <el-table-column prop="device_id" label="设备ID" width="200" />
                    <el-table-column label="状态" width="100">
                      <template #default="{ row }">
                        <el-tag :type="getStatusType(row.status)" size="small">
                          {{ getStatusText(row.status) }}
                        </el-tag>
                      </template>
                    </el-table-column>
                    <el-table-column prop="triggered_at" label="触发时间" />
                  </el-table>
                </el-collapse-item>
              </el-collapse>
            </el-card>
          </div>
        </el-card>
      </el-tab-pane>
      
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
            <el-table-column prop="rule_name" label="规则名称" width="200" />
            <el-table-column label="级别" width="120">
              <template #default="{ row }">
                <el-tag :type="getSeverityType(row.severity)">
                  {{ row.severity }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="condition" label="条件" width="200" />
            <el-table-column label="持续时间" width="120">
              <template #default="{ row }">
                {{ formatDuration(row.duration) }}
              </template>
            </el-table-column>
            <el-table-column prop="description" label="描述" />
            <el-table-column label="通知" width="80">
              <template #default="{ row }">
                <el-tag 
                  v-if="row.notification_config?.enabled" 
                  type="success" 
                  size="small"
                >
                  已启用
                </el-tag>
                <el-tag 
                  v-else 
                  type="info" 
                  size="small"
                >
                  未启用
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <StatusBadge :status="row.enabled ? 'enabled' : 'disabled'" />
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
            <div class="toolbar-left">
              <el-button 
                v-if="selectedEvents.length > 0"
                size="small"
                @click="handleBatchAcknowledge"
                :loading="batchAcknowledging"
              >
                批量确认 ({{ selectedEvents.length }})
              </el-button>
              <el-button 
                v-if="selectedEvents.length > 0"
                size="small"
                type="success"
                @click="handleBatchResolve"
                :loading="batchResolving"
              >
                批量解决 ({{ selectedEvents.length }})
              </el-button>
            </div>
            <div class="toolbar-right">
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
                  <el-checkbox 
                    v-model="selectedEventIds" 
                    :label="event.id"
                    @change="handleEventSelection"
                  />
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
      width="800px"
      @close="resetRuleForm"
    >
      <el-form
        ref="ruleFormRef"
        :model="ruleForm"
        :rules="ruleRules"
        label-width="100px"
      >
        <el-form-item label="规则名称" prop="rule_name">
          <el-input v-model="ruleForm.rule_name" placeholder="请输入规则名称" />
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
          <el-input v-model="ruleForm.metric_name" placeholder="例如: device_status, cpu_usage" />
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
          <el-input-number 
            v-model="ruleForm.duration" 
            :min="1"
            placeholder="持续时间（分钟）"
            style="width: 100%"
          >
            <template #append>分钟</template>
          </el-input-number>
        </el-form-item>
        
        <el-form-item label="启用">
          <el-switch v-model="ruleForm.enabled" />
        </el-form-item>
        
        <!-- 通知配置 -->
        <el-divider content-position="left">通知配置</el-divider>
        
        <NotificationConfig v-model="ruleForm.notification_config" />
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
import NotificationConfig from '@/components/alert/NotificationConfig.vue'
import { Plus, Search } from '@element-plus/icons-vue'
import { ElMessage, ElMessageBox, type FormInstance, type FormRules } from 'element-plus'
import type { NotificationConfig as NotificationConfigType } from '@/types/alert'

const activeTab = ref('overview')
const loading = ref(false)
const submitting = ref(false)
const acknowledgingId = ref<number | null>(null)
const resolvingId = ref<number | null>(null)
const resolvingRuleId = ref<number | null>(null)
const batchAcknowledging = ref(false)
const batchResolving = ref(false)

// 聚合视图
const aggregations = ref<any[]>([])
const expandedRules = ref<number[]>([])

// 批量操作
const selectedEventIds = ref<number[]>([])
const selectedEvents = ref<any[]>([])

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
  rule_name: '',
  description: '',
  severity: 'warning' as 'critical' | 'warning' | 'info',
  metric_name: '',
  operator: '>' as '>' | '<' | '>=' | '<=' | '==' | '!=',
  threshold: 0,
  duration: 5,
  enabled: true,
  filters: {},
  notification_config: {
    enabled: false,
    channels: [],
    dedupe_interval: 300,
    escalation_enabled: false,
    escalation_after: 1800,
    escalation_channels: []
  } as NotificationConfigType
})

const ruleRules: FormRules = {
  rule_name: [{ required: true, message: '请输入规则名称', trigger: 'blur' }],
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

// 获取聚合信息
const fetchAggregations = async () => {
  loading.value = true
  try {
    const res: any = await alertApi.getAggregations()
    aggregations.value = res.data || []
  } catch (error) {
    ElMessage.error('获取聚合信息失败')
  } finally {
    loading.value = false
  }
}

// Tab 切换
const handleTabChange = (tab: string) => {
  if (tab === 'overview') {
    fetchAggregations()
  } else if (tab === 'rules') {
    fetchRules()
  } else if (tab === 'events') {
    fetchEvents()
  }
}

// 查看规则的所有事件
const viewRuleEvents = (ruleId: number) => {
  activeTab.value = 'events'
  eventQuery.rule_id = ruleId
  fetchEvents()
}

// 按规则解决所有告警
const handleResolveByRule = async (ruleId: number) => {
  try {
    await ElMessageBox.confirm('确定要解决该规则的所有告警吗？', '提示', {
      confirmButtonText: '确定',
      cancelButtonText: '取消',
      type: 'warning'
    })
    
    resolvingRuleId.value = ruleId
    await alertApi.resolveByRule(ruleId)
    ElMessage.success('已解决该规则的所有告警')
    fetchAggregations()
  } catch (error: any) {
    if (error !== 'cancel') {
      ElMessage.error(error.message || '操作失败')
    }
  } finally {
    resolvingRuleId.value = null
  }
}

// 事件选择
const handleEventSelection = () => {
  selectedEvents.value = events.value.filter(e => selectedEventIds.value.includes(e.id))
}

// 批量确认
const handleBatchAcknowledge = async () => {
  if (selectedEventIds.value.length === 0) return
  
  batchAcknowledging.value = true
  try {
    await alertApi.batchAcknowledge(selectedEventIds.value)
    ElMessage.success(`已确认 ${selectedEventIds.value.length} 条告警`)
    selectedEventIds.value = []
    selectedEvents.value = []
    fetchEvents()
    if (activeTab.value === 'overview') {
      fetchAggregations()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '批量确认失败')
  } finally {
    batchAcknowledging.value = false
  }
}

// 批量解决
const handleBatchResolve = async () => {
  if (selectedEventIds.value.length === 0) return
  
  batchResolving.value = true
  try {
    await alertApi.batchResolve(selectedEventIds.value)
    ElMessage.success(`已解决 ${selectedEventIds.value.length} 条告警`)
    selectedEventIds.value = []
    selectedEvents.value = []
    fetchEvents()
    if (activeTab.value === 'overview') {
      fetchAggregations()
    }
  } catch (error: any) {
    ElMessage.error(error.message || '批量解决失败')
  } finally {
    batchResolving.value = false
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
  
  // 解析 condition 字符串，提取 metric_name, operator, threshold
  const conditionMatch = row.condition?.match(/^(\w+)\s*([><=!]+)\s*(.+)$/)
  
  Object.assign(ruleForm, {
    rule_name: row.rule_name,
    description: row.description || '',
    severity: row.severity,
    metric_name: conditionMatch ? conditionMatch[1] : '',
    operator: conditionMatch ? conditionMatch[2] : '>',
    threshold: conditionMatch ? parseFloat(conditionMatch[3]) : 0,
    duration: Math.floor(row.duration / 60), // 转换秒为分钟
    enabled: row.enabled,
    filters: row.filters || {},
    notification_config: row.notification_config || {
      enabled: false,
      channels: [],
      dedupe_interval: 300,
      escalation_enabled: false,
      escalation_after: 1800,
      escalation_channels: []
    }
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
        // 构建符合后端格式的请求数据
        const requestData = {
          rule_name: ruleForm.rule_name,
          description: ruleForm.description,
          severity: ruleForm.severity,
          condition: `${ruleForm.metric_name} ${ruleForm.operator} ${ruleForm.threshold}`,
          duration: ruleForm.duration * 60, // 转换分钟为秒
          enabled: ruleForm.enabled,
          filters: ruleForm.filters || {},
          notification_config: ruleForm.notification_config
        }
        
        if (currentRule.value) {
          await alertApi.updateRule(currentRule.value.id, requestData)
          ElMessage.success('更新成功')
        } else {
          await alertApi.createRule(requestData)
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
    rule_name: '',
    description: '',
    severity: 'warning',
    metric_name: '',
    operator: '>',
    threshold: 0,
    duration: 5,
    enabled: true,
    notification_config: {
      enabled: false,
      channels: [],
      dedupe_interval: 300,
      escalation_enabled: false,
      escalation_after: 1800,
      escalation_channels: []
    },
    filters: {}
  })
}

// 格式化持续时间（秒转为可读格式）
const formatDuration = (seconds: number) => {
  if (seconds < 60) {
    return `${seconds}秒`
  } else if (seconds < 3600) {
    return `${Math.floor(seconds / 60)}分钟`
  } else {
    return `${Math.floor(seconds / 3600)}小时`
  }
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
  fetchAggregations()
  
  // 每 30 秒自动刷新聚合信息
  setInterval(() => {
    if (activeTab.value === 'overview') {
      fetchAggregations()
    }
  }, 30000)
})
</script>

<style scoped lang="scss">
.alerts {
  .empty-state {
    padding: 40px 0;
    text-align: center;
  }
  
  .aggregation-list {
    display: grid;
    gap: 16px;
  }
  
  .aggregation-card {
    .agg-header {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-bottom: 12px;
      
      h3 {
        flex: 1;
        margin: 0;
        font-size: 16px;
        font-weight: 600;
      }
    }
    
    .agg-description {
      color: var(--el-text-color-secondary);
      font-size: 14px;
      margin-bottom: 16px;
    }
    
    .agg-stats {
      display: flex;
      gap: 24px;
      margin-bottom: 12px;
      padding: 12px;
      background: var(--el-fill-color-lighter);
      border-radius: 4px;
      
      .stat-item {
        display: flex;
        flex-direction: column;
        gap: 4px;
        
        .stat-label {
          font-size: 12px;
          color: var(--el-text-color-secondary);
        }
        
        .stat-value {
          font-size: 20px;
          font-weight: 600;
          
          &.danger {
            color: var(--el-color-danger);
          }
          
          &.warning {
            color: var(--el-color-warning);
          }
        }
      }
    }
    
    .agg-time {
      display: flex;
      justify-content: space-between;
      font-size: 12px;
      color: var(--el-text-color-secondary);
      margin-bottom: 12px;
    }
    
    .agg-actions {
      display: flex;
      gap: 8px;
      margin-bottom: 12px;
    }
    
    .device-collapse {
      margin-top: 12px;
      border-top: 1px solid var(--el-border-color);
      padding-top: 12px;
      
      .collapse-title {
        font-size: 14px;
        font-weight: 500;
      }
    }
  }
  
  .toolbar-card {
    margin-bottom: 20px;

    .toolbar {
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 10px;
      
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
        
        :deep(.el-checkbox) {
          margin-right: 8px;
        }

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
