import request from '@/utils/request'
import type { AlertRule, AlertEvent, AlertRuleQuery, AlertEventQuery, AlertRuleCreate } from '@/types/alert'
import type { PageResponse } from '@/types/api'

export const alertApi = {
  // 获取告警规则列表
  getRules: (params: AlertRuleQuery) => 
    request.get<PageResponse<AlertRule>>('/v1/alert-rules', { params }),
  
  // 获取告警规则详情
  getRule: (id: string) => 
    request.get<AlertRule>(`/v1/alert-rules/${id}`),
  
  // 创建告警规则
  createRule: (data: AlertRuleCreate) => 
    request.post<{ id: string }>('/v1/alert-rules', data),
  
  // 更新告警规则
  updateRule: (id: string, data: Partial<AlertRuleCreate>) => 
    request.put(`/v1/alert-rules/${id}`, data),
  
  // 删除告警规则
  deleteRule: (id: string) => 
    request.delete(`/v1/alert-rules/${id}`),
  
  // 启用/禁用告警规则
  toggleRule: (id: string, enabled: boolean) => 
    request.post(`/v1/alert-rules/${id}/toggle`, { enabled }),
  
  // 获取告警事件列表
  getEvents: (params: AlertEventQuery) => 
    request.get<PageResponse<AlertEvent>>('/v1/alert-events', { params }),
  
  // 获取告警事件详情
  getEvent: (id: string) => 
    request.get<AlertEvent>(`/v1/alert-events/${id}`),
  
  // 确认告警
  acknowledgeEvent: (id: string, comment?: string) => 
    request.post(`/v1/alert-events/${id}/acknowledge`, { comment: comment || '' }),
  
  // 解决告警
  resolveEvent: (id: string, comment?: string) => 
    request.post(`/v1/alert-events/${id}/resolve`, { comment: comment || '' }),
  
  // 静默告警
  silenceEvent: (id: string, duration: number) => 
    request.post(`/v1/alert-events/${id}/silence`, { duration }),
  
  // 获取告警聚合信息
  getAggregations: () => 
    request.get('/v1/alert-aggregations'),
  
  // 批量确认告警
  batchAcknowledge: (ids: number[], comment?: string) => 
    request.post('/v1/alert-events/batch-acknowledge', { ids, comment: comment || '' }),
  
  // 批量解决告警
  batchResolve: (ids: number[], comment?: string) => 
    request.post('/v1/alert-events/batch-resolve', { ids, comment: comment || '' }),
  
  // 按规则解决所有告警
  resolveByRule: (ruleId: number, comment?: string) => 
    request.post(`/v1/alert-rules/${ruleId}/resolve-all`, { comment: comment || '' }),
  
  // 获取通知历史
  getNotificationHistory: (eventId: number) =>
    request.get(`/v1/alert-events/${eventId}/notifications`),
  
  // 测试通知
  testNotification: (data: {
    channel: string
    recipient: string
    subject: string
    content: string
  }) =>
    request.post('/v1/notifications/test', data)
}

