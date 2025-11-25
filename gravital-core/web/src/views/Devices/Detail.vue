<template>
  <div class="device-detail">
    <el-page-header @back="router.back()">
      <template #content>
        <span class="page-title">设备详情</span>
      </template>
    </el-page-header>

    <el-card v-loading="deviceStore.loading" class="device-info-card">
      <template #header>
        <div class="card-header">
          <span>设备信息</span>
          <div>
            <el-button :icon="Edit" @click="handleEdit">编辑</el-button>
            <el-button :icon="Delete" type="danger" @click="handleDelete">删除</el-button>
          </div>
        </div>
      </template>

      <el-descriptions :column="2" border>
        <el-descriptions-item label="设备ID">
          {{ device?.device_id }}
        </el-descriptions-item>
        <el-descriptions-item label="设备名称">
          {{ device?.name }}
        </el-descriptions-item>
        <el-descriptions-item label="设备类型">
          {{ device?.device_type }}
        </el-descriptions-item>
        <el-descriptions-item label="状态">
          <StatusBadge v-if="device" :status="device.status" />
        </el-descriptions-item>
        <el-descriptions-item label="Sentinel">
          {{ device?.sentinel_id || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="最后在线">
          {{ device?.last_seen || '-' }}
        </el-descriptions-item>
        <el-descriptions-item label="创建时间">
          {{ device?.created_at }}
        </el-descriptions-item>
        <el-descriptions-item label="更新时间">
          {{ device?.updated_at }}
        </el-descriptions-item>
        <el-descriptions-item label="标签" :span="2">
          <el-tag
            v-for="(value, key) in device?.labels"
            :key="key"
            size="small"
            style="margin-right: 5px"
          >
            {{ key }}:{{ value }}
          </el-tag>
          <span v-if="!device?.labels || Object.keys(device.labels).length === 0" style="color: #999">
            暂无标签
          </span>
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-tabs v-model="activeTab" class="detail-tabs" @tab-change="handleTabChange">
      <el-tab-pane label="监控指标" name="metrics">
        <div v-loading="metricsLoading">
          <el-empty v-if="!hasMetricsData" description="暂无监控数据" />
          <div v-else>
            <!-- 标准指标 -->
            <ChartCard
              v-if="cpuOption.series && cpuOption.series.length > 0"
              title="CPU 使用率"
              :option="cpuOption"
              height="300px"
            />
            <ChartCard
              v-if="memoryOption.series && memoryOption.series.length > 0"
              title="内存使用率"
              :option="memoryOption"
              height="300px"
            />
            
            <!-- 动态指标 -->
            <ChartCard
              v-for="chart in dynamicCharts"
              :key="chart.title"
              :title="chart.title"
              :option="chart.option"
              height="300px"
            />
          </div>
        </div>
      </el-tab-pane>
      
      <el-tab-pane label="采集任务" name="tasks">
        <div v-loading="tasksLoading">
          <el-empty v-if="tasks.length === 0" description="暂无采集任务" />
          <el-table v-else :data="tasks" style="width: 100%">
            <el-table-column prop="task_id" label="任务ID" width="180" />
            <el-table-column prop="plugin_name" label="插件名称" width="150" />
            <el-table-column prop="interval_seconds" label="采集间隔" width="120">
              <template #default="{ row }">
                {{ row.interval_seconds }}秒
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="last_executed_at" label="最后执行" width="180">
              <template #default="{ row }">
                {{ row.last_executed_at || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" fixed="right" width="150">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="viewTask(row)">
                  详情
                </el-button>
                <el-button text type="primary" size="small" @click="triggerTask(row)">
                  执行
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
      
      <el-tab-pane label="告警规则" name="alerts">
        <div v-loading="alertsLoading">
          <el-empty v-if="alertRules.length === 0" description="暂无告警规则" />
          <el-table v-else :data="alertRules" style="width: 100%">
            <el-table-column prop="rule_name" label="规则名称" width="200" />
            <el-table-column label="监控指标" width="150">
              <template #default="{ row }">
                {{ parseMetricName(row.condition) }}
              </template>
            </el-table-column>
            <el-table-column prop="condition" label="条件" width="200" />
            <el-table-column label="严重程度" width="120">
              <template #default="{ row }">
                <el-tag
                  :type="getSeverityType(row.severity)"
                  size="small"
                >
                  {{ getSeverityText(row.severity) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="状态" width="100">
              <template #default="{ row }">
                <el-tag :type="row.enabled ? 'success' : 'info'" size="small">
                  {{ row.enabled ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="created_at" label="创建时间" width="180" />
            <el-table-column label="操作" fixed="right" width="120">
              <template #default="{ row }">
                <el-button text type="primary" size="small" @click="viewAlertRule(row)">
                  详情
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
      
      <el-tab-pane label="历史记录" name="history">
        <div v-loading="historyLoading">
          <el-empty v-if="history.length === 0" description="暂无历史记录" />
          <div v-else>
            <el-table :data="history" style="width: 100%">
              <el-table-column prop="type" label="类型" width="100">
                <template #default="{ row }">
                  <el-tag size="small">{{ row.type }}</el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="rule_name" label="规则名称" width="200" />
              <el-table-column label="严重程度" width="120">
                <template #default="{ row }">
                  <el-tag
                    :type="getSeverityType(row.severity)"
                    size="small"
                  >
                    {{ getSeverityText(row.severity) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column label="状态" width="120">
                <template #default="{ row }">
                  <el-tag
                    :type="getStatusType(row.status)"
                    size="small"
                  >
                    {{ getStatusText(row.status) }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="message" label="消息" min-width="200" show-overflow-tooltip />
              <el-table-column prop="created_at" label="发生时间" width="180" />
              <el-table-column prop="resolved_at" label="解决时间" width="180">
                <template #default="{ row }">
                  {{ row.resolved_at || '-' }}
                </template>
              </el-table-column>
            </el-table>
            
            <el-pagination
              v-model:current-page="historyPage"
              v-model:page-size="historyPageSize"
              :total="historyTotal"
              :page-sizes="[10, 20, 50]"
              layout="total, sizes, prev, pager, next"
              style="margin-top: 20px"
              @size-change="fetchHistory"
              @current-change="fetchHistory"
            />
          </div>
        </div>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDeviceStore } from '@/stores/device'
import { deviceApi } from '@/api/device'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ChartCard from '@/components/common/ChartCard.vue'
import { Edit, Delete } from '@element-plus/icons-vue'
import { ElMessageBox, ElMessage } from 'element-plus'
import type { EChartsOption } from 'echarts'

const route = useRoute()
const router = useRouter()
const deviceStore = useDeviceStore()

const activeTab = ref('metrics')
const device = computed(() => deviceStore.currentDevice)

// 监控指标
const metricsLoading = ref(false)
const hasMetricsData = ref(false)
const cpuOption = ref<EChartsOption>({})
const memoryOption = ref<EChartsOption>({})
const dynamicCharts = ref<Array<{ title: string; option: EChartsOption }>>([])

// 采集任务
const tasksLoading = ref(false)
const tasks = ref<any[]>([])

// 告警规则
const alertsLoading = ref(false)
const alertRules = ref<any[]>([])

// 历史记录
const historyLoading = ref(false)
const history = ref<any[]>([])
const historyTotal = ref(0)
const historyPage = ref(1)
const historyPageSize = ref(20)

const fetchMetrics = async () => {
  if (!device.value) return
  
  metricsLoading.value = true
  try {
    const res: any = await deviceApi.getDeviceMetrics(device.value.id, 24)
    
    // 检查是否有任何指标数据
    const metrics = res.metrics || {}
    
    // 检查是否有任何指标包含数据
    hasMetricsData.value = Object.values(metrics).some((metric: any) => 
      metric.values && metric.values.length > 0
    )
    
    if (hasMetricsData.value) {
      // 获取标准指标数据
      const cpuData = metrics.cpu || {}
      const memoryData = metrics.memory || {}
      
      // CPU 图表（如果有数据）
      if (cpuData.values && cpuData.values.length > 0) {
        cpuOption.value = {
          tooltip: {
            trigger: 'axis'
          },
          xAxis: {
            type: 'category',
            data: cpuData.timestamps || [],
            axisLine: { lineStyle: { color: '#666' } }
          },
          yAxis: {
            type: 'value',
            name: '使用率 (%)',
            axisLine: { lineStyle: { color: '#666' } }
          },
          series: [{
            data: cpuData.values || [],
            type: 'line',
            smooth: true,
            itemStyle: { color: '#409eff' },
            areaStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(64, 158, 255, 0.3)' },
                  { offset: 1, color: 'rgba(64, 158, 255, 0.1)' }
                ]
              }
            }
          }]
        }
      }
      
      // 内存图表（如果有数据）
      if (memoryData.values && memoryData.values.length > 0) {
        memoryOption.value = {
          tooltip: {
            trigger: 'axis'
          },
          xAxis: {
            type: 'category',
            data: memoryData.timestamps || [],
            axisLine: { lineStyle: { color: '#666' } }
          },
          yAxis: {
            type: 'value',
            name: '使用率 (%)',
            axisLine: { lineStyle: { color: '#666' } }
          },
          series: [{
            data: memoryData.values || [],
            type: 'line',
            smooth: true,
            itemStyle: { color: '#67c23a' },
            areaStyle: {
              color: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: 'rgba(103, 194, 58, 0.3)' },
                  { offset: 1, color: 'rgba(103, 194, 58, 0.1)' }
                ]
              }
            }
          }]
        }
      }
      
      // 为其他可用指标创建图表
      createDynamicCharts(metrics)
    }
  } catch (error) {
    console.error('获取监控指标失败', error)
    hasMetricsData.value = false
  } finally {
    metricsLoading.value = false
  }
}

const fetchTasks = async () => {
  if (!device.value) return
  
  tasksLoading.value = true
  try {
    const res: any = await deviceApi.getDeviceTasks(device.value.id)
    tasks.value = res || []
  } catch (error) {
    console.error('获取采集任务失败', error)
    tasks.value = []
  } finally {
    tasksLoading.value = false
  }
}

const fetchAlertRules = async () => {
  if (!device.value) return
  
  alertsLoading.value = true
  try {
    const res: any = await deviceApi.getDeviceAlertRules(device.value.id)
    alertRules.value = res || []
  } catch (error) {
    console.error('获取告警规则失败', error)
    alertRules.value = []
  } finally {
    alertsLoading.value = false
  }
}

const fetchHistory = async () => {
  if (!device.value) return
  
  historyLoading.value = true
  try {
    const res: any = await deviceApi.getDeviceHistory(device.value.id, {
      page: historyPage.value,
      page_size: historyPageSize.value
    })
    history.value = res.items || []
    historyTotal.value = res.total || 0
  } catch (error) {
    console.error('获取历史记录失败', error)
    history.value = []
    historyTotal.value = 0
  } finally {
    historyLoading.value = false
  }
}

const handleTabChange = (tabName: string) => {
  switch (tabName) {
    case 'metrics':
      fetchMetrics()
      break
    case 'tasks':
      fetchTasks()
      break
    case 'alerts':
      fetchAlertRules()
      break
    case 'history':
      fetchHistory()
      break
  }
}

const handleEdit = () => {
  router.push(`/devices`)
}

const handleDelete = () => {
  ElMessageBox.confirm('确定要删除该设备吗？', '提示', {
    confirmButtonText: '确定',
    cancelButtonText: '取消',
    type: 'warning'
  }).then(async () => {
    if (device.value) {
      await deviceStore.deleteDevice(device.value.id)
      router.back()
    }
  })
}

const viewTask = (task: any) => {
  router.push(`/tasks/${task.id}`)
}

const triggerTask = (task: any) => {
  ElMessage.info('手动触发任务功能开发中...')
}

const viewAlertRule = (rule: any) => {
  router.push(`/alerts?rule_id=${rule.id}`)
}

const getSeverityType = (severity: string) => {
  const map: Record<string, any> = {
    critical: 'danger',
    warning: 'warning',
    info: 'info'
  }
  return map[severity] || 'info'
}

const getSeverityText = (severity: string) => {
  const map: Record<string, string> = {
    critical: '严重',
    warning: '警告',
    info: '信息'
  }
  return map[severity] || severity
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
    firing: '触发中',
    acknowledged: '已确认',
    resolved: '已解决'
  }
  return map[status] || status
}

const parseMetricName = (condition: string) => {
  if (!condition) return '-'
  // 从 condition 中提取指标名称（第一个单词）
  const parts = condition.trim().split(/\s+/)
  return parts[0] || '-'
}

const createDynamicCharts = (metrics: Record<string, any>) => {
  // 清空动态图表
  dynamicCharts.value = []
  
  // 标准指标（已单独处理）
  const standardMetrics = ['cpu', 'memory']
  
  // 指标名称映射
  const metricNameMap: Record<string, string> = {
    'device_status': '设备状态',
    'ping_reachable': 'Ping 可达性',
    'ping_rtt_ms': 'Ping 延迟 (ms)',
    'ping_packet_loss': 'Ping 丢包率 (%)',
    'disk': '磁盘使用率 (%)',
    'network_in': '网络入流量',
    'network_out': '网络出流量'
  }
  
  // 为其他有数据的指标创建图表
  Object.entries(metrics).forEach(([metricName, data]: [string, any]) => {
    // 跳过标准指标和空数据
    if (standardMetrics.includes(metricName) || !data.values || data.values.length === 0) {
      return
    }
    
    const title = metricNameMap[metricName] || metricName
    
    const option: EChartsOption = {
      tooltip: {
        trigger: 'axis'
      },
      xAxis: {
        type: 'category',
        data: data.timestamps || [],
        axisLine: { lineStyle: { color: '#666' } }
      },
      yAxis: {
        type: 'value',
        axisLine: { lineStyle: { color: '#666' } }
      },
      series: [{
        data: data.values || [],
        type: 'line',
        smooth: true,
        itemStyle: { color: '#e6a23c' },
        areaStyle: {
          color: {
            type: 'linear',
            x: 0,
            y: 0,
            x2: 0,
            y2: 1,
            colorStops: [
              { offset: 0, color: 'rgba(230, 162, 60, 0.3)' },
              { offset: 1, color: 'rgba(230, 162, 60, 0.1)' }
            ]
          }
        }
      }]
    }
    
    dynamicCharts.value.push({ title, option })
  })
}

onMounted(async () => {
  const id = route.params.id as string
  await deviceStore.fetchDevice(id)
  
  // 默认加载监控指标
  fetchMetrics()
})
</script>

<style scoped lang="scss">
.device-detail {
  .device-info-card {
    margin: 20px 0;

    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }

  .detail-tabs {
    margin-top: 20px;
  }
}
</style>
