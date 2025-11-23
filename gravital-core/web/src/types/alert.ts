// 告警类型定义
export interface AlertRule {
  id: string
  rule_name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  condition: string
  duration: number
  filters: Record<string, any>
  notification_config: Record<string, any>
  enabled: boolean
  created_at: string
  updated_at: string
}

// 告警事件
export interface AlertEvent {
  id: string
  rule_id: string
  rule_name: string
  severity: 'critical' | 'warning' | 'info'
  device_id: string
  device_name: string
  metric_name: string
  current_value: number
  threshold: number
  message: string
  status: 'firing' | 'resolved' | 'acknowledged'
  triggered_at: string
  resolved_at?: string
  acknowledged_at?: string
  acknowledged_by?: string
}

// 告警规则查询参数
export interface AlertRuleQuery {
  page?: number
  size?: number
  keyword?: string
  severity?: string
  enabled?: boolean
}

// 告警事件查询参数
export interface AlertEventQuery {
  page?: number
  size?: number
  severity?: string
  status?: string
  start_time?: string
  end_time?: string
}

// 通知渠道配置
export interface NotificationChannelConfig {
  channel: 'email' | 'webhook' | 'dingtalk' | 'wechat' | 'sms'
  enabled: boolean
  recipients: string[]
  template?: string
  config?: Record<string, any>
}

// 通知配置
export interface NotificationConfig {
  enabled: boolean
  channels: NotificationChannelConfig[]
  dedupe_interval?: number      // 去重间隔（秒）
  escalation_enabled?: boolean  // 是否启用升级
  escalation_after?: number     // 升级时间（秒）
  escalation_channels?: string[] // 升级通知渠道
}

// 创建告警规则请求
export interface AlertRuleCreate {
  rule_name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  condition: string
  duration: number
  filters?: Record<string, any>
  notification_config?: NotificationConfig
  enabled?: boolean
}

// 告警聚合信息
export interface AlertAggregation {
  rule_id: number
  rule_name: string
  severity: 'critical' | 'warning' | 'info'
  description: string
  total_count: number
  firing_count: number
  acked_count: number
  first_fired: string
  last_fired: string
  devices: Array<{
    device_id: string
    device_name: string
    status: string
    triggered_at: string
  }>
}

