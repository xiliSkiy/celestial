<template>
  <div class="dashboard">
    <!-- 统计卡片区 -->
    <div class="stat-cards">
      <StatCard
        label="设备总数"
        :value="stats.totalDevices"
        :icon="Monitor"
        icon-color="#409eff"
        icon-bg="rgba(64, 158, 255, 0.1)"
        trend="+5.2%"
        trend-type="up"
        clickable
        @click="router.push('/devices')"
      />
      <StatCard
        label="在线设备"
        :value="stats.onlineDevices"
        :icon="CircleCheck"
        icon-color="#67c23a"
        icon-bg="rgba(103, 194, 58, 0.1)"
        trend="+2.1%"
        trend-type="up"
      />
      <StatCard
        label="活跃告警"
        :value="stats.activeAlerts"
        :icon="Bell"
        icon-color="#f56c6c"
        icon-bg="rgba(245, 108, 108, 0.1)"
        trend="-12%"
        trend-type="down"
        clickable
        @click="router.push('/alerts')"
      />
      <StatCard
        label="任务数"
        :value="stats.totalTasks"
        :icon="List"
        icon-color="#e6a23c"
        icon-bg="rgba(230, 162, 60, 0.1)"
        trend="+8.3%"
        trend-type="up"
      />
    </div>

    <!-- 图表区 -->
    <div class="chart-grid">
      <ChartCard
        title="设备状态分布"
        :option="deviceStatusOption"
        height="350px"
        @refresh="fetchDeviceStatus"
      />
      <ChartCard
        title="告警趋势"
        :option="alertTrendOption"
        height="350px"
        @refresh="fetchAlertTrend"
      />
    </div>

    <div class="chart-grid">
      <ChartCard
        title="Sentinel 状态"
        :option="sentinelStatusOption"
        height="350px"
        @refresh="fetchSentinelStatus"
      />
      <ChartCard
        title="数据转发统计"
        :option="forwarderStatsOption"
        height="350px"
        @refresh="fetchForwarderStats"
      />
    </div>

    <!-- 最近活动 -->
    <el-card class="activity-card">
      <template #header>
        <div class="card-header">
          <span>最近活动</span>
          <el-button text @click="fetchActivities">刷新</el-button>
        </div>
      </template>
      <el-timeline>
        <el-timeline-item
          v-for="activity in activities"
          :key="activity.id"
          :timestamp="activity.time"
          :type="activity.type"
        >
          {{ activity.content }}
        </el-timeline-item>
      </el-timeline>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRouter } from 'vue-router'
import StatCard from '@/components/common/StatCard.vue'
import ChartCard from '@/components/common/ChartCard.vue'
import { Monitor, CircleCheck, Bell, List } from '@element-plus/icons-vue'
import type { EChartsOption } from 'echarts'
import { dashboardApi } from '@/api/dashboard'
import { ElMessage } from 'element-plus'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'
import 'dayjs/locale/zh-cn'

dayjs.extend(relativeTime)
dayjs.locale('zh-cn')

const router = useRouter()

const stats = ref({
  totalDevices: 0,
  onlineDevices: 0,
  activeAlerts: 0,
  totalTasks: 0
})

const activities = ref<any[]>([])
const loading = ref(false)

// 计算趋势
const deviceTrend = computed(() => {
  if (stats.value.totalDevices === 0) return '+0%'
  const percentage = ((stats.value.onlineDevices / stats.value.totalDevices) * 100).toFixed(1)
  return `${percentage}%`
})

// 设备状态分布图表
const deviceStatusOption = ref<EChartsOption>({
  tooltip: {
    trigger: 'item',
    formatter: '{b}: {c} ({d}%)'
  },
  legend: {
    orient: 'vertical',
    left: 'left',
    textStyle: { color: '#fff' }
  },
  series: [
    {
      name: '设备状态',
      type: 'pie',
      radius: '50%',
      data: [],
      emphasis: {
        itemStyle: {
          shadowBlur: 10,
          shadowOffsetX: 0,
          shadowColor: 'rgba(0, 0, 0, 0.5)'
        }
      }
    }
  ]
})

// 告警趋势图表
const alertTrendOption = ref<EChartsOption>({
  tooltip: {
    trigger: 'axis'
  },
  legend: {
    data: ['Critical', 'Warning', 'Info'],
    textStyle: { color: '#fff' }
  },
  xAxis: {
    type: 'category',
    data: [],
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' },
    splitLine: { lineStyle: { color: '#333' } }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  series: []
})

// Sentinel 状态图表
const sentinelStatusOption = ref<EChartsOption>({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' }
  },
  legend: {
    data: ['在线', '离线'],
    textStyle: { color: '#fff' }
  },
  xAxis: {
    type: 'category',
    data: [],
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' },
    splitLine: { lineStyle: { color: '#333' } }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  series: []
})

// 数据转发统计图表
const forwarderStatsOption = ref<EChartsOption>({
  tooltip: {
    trigger: 'axis',
    axisPointer: { type: 'shadow' }
  },
  legend: {
    data: ['成功', '失败'],
    textStyle: { color: '#fff' }
  },
  xAxis: {
    type: 'category',
    data: [],
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#666' } },
    axisLabel: { color: '#999' },
    splitLine: { lineStyle: { color: '#333' } }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  series: []
})

// 获取统计数据
const fetchStats = async () => {
  try {
    const res: any = await dashboardApi.getStats()
    stats.value = {
      totalDevices: res.total_devices || 0,
      onlineDevices: res.online_devices || 0,
      activeAlerts: res.active_alerts || 0,
      totalTasks: res.total_tasks || 0
    }
  } catch (error) {
    console.error('获取统计数据失败:', error)
  }
}

// 获取设备状态
const fetchDeviceStatus = async () => {
  try {
    const res: any = await dashboardApi.getDeviceStatus()
    const data = res || []
    
    const statusMap: Record<string, { name: string; color: string }> = {
      online: { name: '在线', color: '#67c23a' },
      offline: { name: '离线', color: '#909399' },
      error: { name: '错误', color: '#f56c6c' }
    }
    
    const pieData = data.map((item: any) => ({
      value: item.count,
      name: statusMap[item.status]?.name || item.status,
      itemStyle: { color: statusMap[item.status]?.color }
    }))
    
    if (deviceStatusOption.value.series && deviceStatusOption.value.series[0]) {
      (deviceStatusOption.value.series[0] as any).data = pieData
    }
  } catch (error) {
    console.error('获取设备状态失败:', error)
  }
}

// 获取告警趋势
const fetchAlertTrend = async () => {
  try {
    const res: any = await dashboardApi.getAlertTrend(24)
    const data = res || []
    
    const times = data.map((item: any) => dayjs(item.time).format('HH:mm'))
    const critical = data.map((item: any) => item.critical || 0)
    const warning = data.map((item: any) => item.warning || 0)
    const info = data.map((item: any) => item.info || 0)
    
    alertTrendOption.value.xAxis = {
      type: 'category',
      data: times,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#999' }
    }
    
    alertTrendOption.value.series = [
      {
        name: 'Critical',
        type: 'line',
        data: critical,
        smooth: true,
        itemStyle: { color: '#f56c6c' }
      },
      {
        name: 'Warning',
        type: 'line',
        data: warning,
        smooth: true,
        itemStyle: { color: '#e6a23c' }
      },
      {
        name: 'Info',
        type: 'line',
        data: info,
        smooth: true,
        itemStyle: { color: '#909399' }
      }
    ]
  } catch (error) {
    console.error('获取告警趋势失败:', error)
  }
}

// 获取 Sentinel 状态
const fetchSentinelStatus = async () => {
  try {
    const res: any = await dashboardApi.getSentinelStatus()
    const data = res || []
    
    const regions = data.map((item: any) => item.region)
    const online = data.map((item: any) => item.online || 0)
    const offline = data.map((item: any) => item.offline || 0)
    
    sentinelStatusOption.value.xAxis = {
      type: 'category',
      data: regions,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#999' }
    }
    
    sentinelStatusOption.value.series = [
      {
        name: '在线',
        type: 'bar',
        data: online,
        itemStyle: { color: '#67c23a' }
      },
      {
        name: '离线',
        type: 'bar',
        data: offline,
        itemStyle: { color: '#f56c6c' }
      }
    ]
  } catch (error) {
    console.error('获取 Sentinel 状态失败:', error)
  }
}

// 获取转发器统计
const fetchForwarderStats = async () => {
  try {
    const res: any = await dashboardApi.getForwarderStats()
    const data = res || []
    
    const names = data.map((item: any) => item.name)
    const success = data.map((item: any) => item.success_count || 0)
    const failure = data.map((item: any) => item.failure_count || 0)
    
    forwarderStatsOption.value.xAxis = {
      type: 'category',
      data: names,
      axisLine: { lineStyle: { color: '#666' } },
      axisLabel: { color: '#999' }
    }
    
    forwarderStatsOption.value.series = [
      {
        name: '成功',
        type: 'bar',
        data: success,
        itemStyle: { color: '#67c23a' }
      },
      {
        name: '失败',
        type: 'bar',
        data: failure,
        itemStyle: { color: '#f56c6c' }
      }
    ]
  } catch (error) {
    console.error('获取转发器统计失败:', error)
  }
}

// 获取最近活动
const fetchActivities = async () => {
  try {
    const res: any = await dashboardApi.getActivities(10)
    activities.value = (res || []).map((item: any) => ({
      ...item,
      time: dayjs(item.created_at).fromNow()
    }))
  } catch (error) {
    console.error('获取活动失败:', error)
  }
}

// 刷新所有数据
const refreshAll = async () => {
  loading.value = true
  try {
    await Promise.all([
      fetchStats(),
      fetchDeviceStatus(),
      fetchAlertTrend(),
      fetchSentinelStatus(),
      fetchForwarderStats(),
      fetchActivities()
    ])
    ElMessage.success('数据已刷新')
  } catch (error) {
    ElMessage.error('刷新数据失败')
  } finally {
    loading.value = false
  }
}

onMounted(() => {
  refreshAll()
  
  // 每 30 秒自动刷新
  const timer = setInterval(refreshAll, 30000)
  
  // 组件卸载时清除定时器
  return () => clearInterval(timer)
})
</script>

<style scoped lang="scss">
.dashboard {
  .stat-cards {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 20px;
    margin-bottom: 20px;
  }

  .chart-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(500px, 1fr));
    gap: 20px;
    margin-bottom: 20px;
  }

  .activity-card {
    .card-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
    }
  }
}
</style>

