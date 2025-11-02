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
        </el-descriptions-item>
      </el-descriptions>
    </el-card>

    <el-tabs v-model="activeTab" class="detail-tabs">
      <el-tab-pane label="监控指标" name="metrics">
        <ChartCard
          title="CPU 使用率"
          :option="cpuOption"
          height="300px"
        />
        <ChartCard
          title="内存使用率"
          :option="memoryOption"
          height="300px"
        />
      </el-tab-pane>
      
      <el-tab-pane label="采集任务" name="tasks">
        <el-empty description="暂无采集任务" />
      </el-tab-pane>
      
      <el-tab-pane label="告警规则" name="alerts">
        <el-empty description="暂无告警规则" />
      </el-tab-pane>
      
      <el-tab-pane label="历史记录" name="history">
        <el-empty description="暂无历史记录" />
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useDeviceStore } from '@/stores/device'
import StatusBadge from '@/components/common/StatusBadge.vue'
import ChartCard from '@/components/common/ChartCard.vue'
import { Edit, Delete } from '@element-plus/icons-vue'
import { ElMessageBox } from 'element-plus'
import type { EChartsOption } from 'echarts'

const route = useRoute()
const router = useRouter()
const deviceStore = useDeviceStore()

const activeTab = ref('metrics')
const device = computed(() => deviceStore.currentDevice)

const cpuOption = ref<EChartsOption>({
  xAxis: {
    type: 'category',
    data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
    axisLine: { lineStyle: { color: '#666' } }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#666' } }
  },
  series: [{
    data: [45, 52, 48, 65, 58, 55],
    type: 'line',
    smooth: true,
    itemStyle: { color: '#409eff' }
  }]
})

const memoryOption = ref<EChartsOption>({
  xAxis: {
    type: 'category',
    data: ['00:00', '04:00', '08:00', '12:00', '16:00', '20:00'],
    axisLine: { lineStyle: { color: '#666' } }
  },
  yAxis: {
    type: 'value',
    axisLine: { lineStyle: { color: '#666' } }
  },
  series: [{
    data: [62, 68, 65, 72, 70, 68],
    type: 'line',
    smooth: true,
    itemStyle: { color: '#67c23a' }
  }]
})

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

onMounted(() => {
  const id = route.params.id as string
  deviceStore.fetchDevice(id)
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

