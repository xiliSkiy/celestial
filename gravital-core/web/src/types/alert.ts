// 告警类型定义
export interface AlertRule {
  id: string
  name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  metric_name: string
  operator: '>' | '<' | '>=' | '<=' | '==' | '!='
  threshold: number
  duration: number
  filters: Record<string, any>
  notification_channels: string[]
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

// 创建告警规则请求
export interface AlertRuleCreate {
  name: string
  description?: string
  severity: 'critical' | 'warning' | 'info'
  metric_name: string
  operator: '>' | '<' | '>=' | '<=' | '==' | '!='
  threshold: number
  duration: number
  filters?: Record<string, any>
  notification_channels: string[]
}

