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
  getForwarder: (name: string) => {
    return request.get(`/v1/forwarders/${name}`)
  },

  // 创建转发器
  createForwarder: (data: ForwarderForm) => {
    return request.post('/v1/forwarders', data)
  },

  // 更新转发器
  updateForwarder: (name: string, data: Partial<ForwarderForm>) => {
    return request.put(`/v1/forwarders/${name}`, data)
  },

  // 删除转发器
  deleteForwarder: (name: string) => {
    return request.delete(`/v1/forwarders/${name}`)
  },

  // 启用/禁用转发器
  toggleForwarder: (name: string, enabled: boolean) => {
    // 使用 PUT 方法更新，只更新 enabled 字段
    return request.put(`/v1/forwarders/${name}`, { enabled })
  },

  // 重新加载配置
  reloadConfig: () => {
    return request.post('/v1/forwarders/reload')
  },

  // 获取转发器统计
  getStats: (name: string) => {
    return request.get(`/v1/forwarders/${name}/stats`)
  },

  // 获取缓冲区统计
  getBufferStats: () => {
    return request.get('/v1/forwarders/buffer-stats')
  },

  // 测试连接（接受配置对象，不需要已存在的转发器）
  testConnection: (data: ForwarderForm) => {
    return request.post('/v1/forwarders/test', data)
  }
}

