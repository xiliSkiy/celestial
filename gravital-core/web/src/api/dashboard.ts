import request from '@/utils/request'

export interface DashboardStats {
  total_devices: number
  online_devices: number
  offline_devices: number
  error_devices: number
  active_alerts: number
  total_tasks: number
  active_sentinels: number
  total_sentinels: number
}

export interface DeviceStatusData {
  status: string
  count: number
}

export interface AlertTrendData {
  time: string
  critical: number
  warning: number
  info: number
}

export interface SentinelStatusData {
  region: string
  online: number
  offline: number
}

export interface ForwarderStatsData {
  name: string
  success_count: number
  failure_count: number
}

export interface Activity {
  id: number
  type: 'info' | 'warning' | 'danger' | 'success'
  content: string
  time: string
  created_at: string
}

export const dashboardApi = {
  // 获取统计数据
  getStats: () => {
    return request.get<DashboardStats>('/v1/dashboard/stats')
  },

  // 获取设备状态分布
  getDeviceStatus: () => {
    return request.get<DeviceStatusData[]>('/v1/dashboard/device-status')
  },

  // 获取告警趋势
  getAlertTrend: (hours: number = 24) => {
    return request.get<AlertTrendData[]>('/v1/dashboard/alert-trend', {
      params: { hours }
    })
  },

  // 获取 Sentinel 状态
  getSentinelStatus: () => {
    return request.get<SentinelStatusData[]>('/v1/dashboard/sentinel-status')
  },

  // 获取转发器统计
  getForwarderStats: () => {
    return request.get<ForwarderStatsData[]>('/v1/dashboard/forwarder-stats')
  },

  // 获取最近活动
  getActivities: (limit: number = 10) => {
    return request.get<Activity[]>('/v1/dashboard/activities', {
      params: { limit }
    })
  }
}

