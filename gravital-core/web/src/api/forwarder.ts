import request from '@/utils/request'

export interface ForwarderForm {
  name: string
  type: 'prometheus' | 'victoria-metrics' | 'clickhouse'
  endpoint: string
  enabled: boolean
  batch_size?: number
  flush_interval?: string
  auth_config?: Record<string, any>
  tls_config?: Record<string, any>
}

export const forwarderApi = {
  // 获取转发器列表
  getForwarders: () => {
    return request.get('/v1/forwarders')
  },

  // 获取单个转发器
  getForwarder: (id: number) => {
    return request.get(`/v1/forwarders/${id}`)
  },

  // 创建转发器
  createForwarder: (data: ForwarderForm) => {
    return request.post('/v1/forwarders', data)
  },

  // 更新转发器
  updateForwarder: (id: number, data: Partial<ForwarderForm>) => {
    return request.put(`/v1/forwarders/${id}`, data)
  },

  // 删除转发器
  deleteForwarder: (id: number) => {
    return request.delete(`/v1/forwarders/${id}`)
  },

  // 启用/禁用转发器
  toggleForwarder: (id: number, enabled: boolean) => {
    return request.patch(`/v1/forwarders/${id}`, { enabled })
  },

  // 重新加载配置
  reloadConfig: () => {
    return request.post('/v1/forwarders/reload')
  },

  // 获取转发器统计
  getStats: (id: number) => {
    return request.get(`/v1/forwarders/${id}/stats`)
  },

  // 获取缓冲区统计
  getBufferStats: () => {
    return request.get('/v1/forwarders/buffer-stats')
  },

  // 测试连接
  testConnection: (id: number) => {
    return request.post(`/v1/forwarders/${id}/test`)
  }
}

