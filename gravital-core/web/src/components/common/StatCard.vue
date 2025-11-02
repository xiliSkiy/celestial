<template>
  <div class="stat-card" @click="handleClick">
    <div class="stat-icon" :style="{ background: iconBg }">
      <el-icon :size="32" :color="iconColor">
        <component :is="icon" />
      </el-icon>
    </div>
    <div class="stat-content">
      <div class="stat-label">{{ label }}</div>
      <div class="stat-value">{{ formatValue(value) }}</div>
      <div v-if="trend" class="stat-trend" :class="trendClass">
        <el-icon>
          <CaretTop v-if="trendType === 'up'" />
          <CaretBottom v-else />
        </el-icon>
        <span>{{ trend }}</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  label: string
  value: string | number
  icon?: any
  iconColor?: string
  iconBg?: string
  trend?: string
  trendType?: 'up' | 'down'
  clickable?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  iconColor: '#409eff',
  iconBg: 'rgba(64, 158, 255, 0.1)',
  clickable: false
})

const emit = defineEmits(['click'])

const trendClass = computed(() => {
  return props.trendType === 'up' ? 'trend-up' : 'trend-down'
})

const formatValue = (val: string | number) => {
  if (typeof val === 'number') {
    return val.toLocaleString()
  }
  return val
}

const handleClick = () => {
  if (props.clickable) {
    emit('click')
  }
}
</script>

<style scoped lang="scss">
.stat-card {
  display: flex;
  align-items: center;
  gap: 20px;
  padding: 24px;
  background: var(--bg-secondary);
  border-radius: 8px;
  border: 1px solid var(--border-color);
  transition: all 0.3s;
  cursor: pointer;

  &:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  .stat-icon {
    width: 64px;
    height: 64px;
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 8px;
  }

  .stat-content {
    flex: 1;

    .stat-label {
      font-size: 14px;
      color: var(--text-secondary);
      margin-bottom: 8px;
    }

    .stat-value {
      font-size: 28px;
      font-weight: bold;
      color: var(--text-primary);
      margin-bottom: 4px;
    }

    .stat-trend {
      display: flex;
      align-items: center;
      gap: 4px;
      font-size: 12px;

      &.trend-up {
        color: var(--color-success);
      }

      &.trend-down {
        color: var(--color-danger);
      }
    }
  }
}
</style>

