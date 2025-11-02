<template>
  <div class="chart-card">
    <div class="chart-header">
      <h3>{{ title }}</h3>
      <div class="chart-actions">
        <el-button :icon="Refresh" circle @click="handleRefresh" />
        <el-button :icon="Download" circle @click="handleDownload" />
      </div>
    </div>
    <div class="chart-body">
      <div ref="chartRef" class="chart-container"></div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { Refresh, Download } from '@element-plus/icons-vue'

interface Props {
  title: string
  option: echarts.EChartsOption
  height?: string
}

const props = withDefaults(defineProps<Props>(), {
  height: '400px'
})

const emit = defineEmits(['refresh', 'download'])

const chartRef = ref<HTMLElement>()
let chart: echarts.ECharts | null = null

onMounted(() => {
  if (chartRef.value) {
    chart = echarts.init(chartRef.value, 'dark')
    chart.setOption(props.option)
    
    // 监听窗口大小变化
    window.addEventListener('resize', handleResize)
  }
})

onUnmounted(() => {
  if (chart) {
    chart.dispose()
  }
  window.removeEventListener('resize', handleResize)
})

watch(() => props.option, (newOption) => {
  if (chart) {
    chart.setOption(newOption, true)
  }
}, { deep: true })

const handleResize = () => {
  if (chart) {
    chart.resize()
  }
}

const handleRefresh = () => {
  emit('refresh')
}

const handleDownload = () => {
  if (chart) {
    const url = chart.getDataURL({
      type: 'png',
      backgroundColor: '#fff'
    })
    const link = document.createElement('a')
    link.href = url
    link.download = `${props.title}.png`
    link.click()
  }
  emit('download')
}
</script>

<style scoped lang="scss">
.chart-card {
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  padding: 20px;

  .chart-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 20px;

    h3 {
      font-size: 16px;
      font-weight: 600;
      color: var(--text-primary);
      margin: 0;
    }

    .chart-actions {
      display: flex;
      gap: 8px;
    }
  }

  .chart-body {
    .chart-container {
      width: 100%;
      height: v-bind(height);
    }
  }
}
</style>

