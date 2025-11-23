<template>
  <span class="status-badge" :class="`status-${status}`">
    <i class="status-dot"></i>
    <span class="status-text">{{ text || statusText }}</span>
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  status: 'online' | 'offline' | 'error' | 'unknown' | 'firing' | 'resolved' | 'acknowledged' | 'enabled' | 'disabled'
  text?: string
}

const props = defineProps<Props>()

const statusText = computed(() => {
  const statusMap: Record<string, string> = {
    online: '在线',
    offline: '离线',
    error: '错误',
    unknown: '未知',
    firing: '告警中',
    resolved: '已解决',
    acknowledged: '已确认',
    enabled: '已启用',
    disabled: '已禁用'
  }
  return statusMap[props.status] || props.status
})
</script>

<style scoped lang="scss">
.status-badge {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 500;

  .status-dot {
    width: 6px;
    height: 6px;
    border-radius: 50%;
  }

  &.status-online {
    color: var(--color-success);
    background: rgba(103, 194, 58, 0.1);

    .status-dot {
      background: var(--color-success);
    }
  }

  &.status-offline {
    color: var(--color-danger);
    background: rgba(245, 108, 108, 0.1);

    .status-dot {
      background: var(--color-danger);
    }
  }

  &.status-error {
    color: var(--color-warning);
    background: rgba(230, 162, 60, 0.1);

    .status-dot {
      background: var(--color-warning);
    }
  }

  &.status-unknown {
    color: var(--color-info);
    background: rgba(144, 147, 153, 0.1);

    .status-dot {
      background: var(--color-info);
    }
  }

  &.status-firing {
    color: var(--color-danger);
    background: rgba(245, 108, 108, 0.1);

    .status-dot {
      background: var(--color-danger);
      animation: pulse 2s infinite;
    }
  }

  &.status-resolved {
    color: var(--color-success);
    background: rgba(103, 194, 58, 0.1);

    .status-dot {
      background: var(--color-success);
    }
  }

  &.status-acknowledged {
    color: var(--color-warning);
    background: rgba(230, 162, 60, 0.1);

    .status-dot {
      background: var(--color-warning);
    }
  }

  &.status-enabled {
    color: var(--color-success);
    background: rgba(103, 194, 58, 0.1);

    .status-dot {
      background: var(--color-success);
    }
  }

  &.status-disabled {
    color: var(--color-info);
    background: rgba(144, 147, 153, 0.1);

    .status-dot {
      background: var(--color-info);
    }
  }
}

@keyframes pulse {
  0%, 100% {
    opacity: 1;
  }
  50% {
    opacity: 0.5;
  }
}
</style>

