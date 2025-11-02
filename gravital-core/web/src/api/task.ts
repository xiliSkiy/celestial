import request from '@/utils/request'

export interface TaskQuery {
  page?: number
  size?: number
  keyword?: string
  device_id?: string
  sentinel_id?: string
  enabled?: boolean
}

export interface TaskForm {
  name: string
  device_id: string
  sentinel_id?: string
  plugin_type: string
  device_config: Record<string, any>
  interval: string
  timeout: string
  enabled: boolean
  labels?: Record<string, string>
}

export const taskApi = {
  // 获取任务列表
  getTasks: (params: TaskQuery) => {
    return request.get('/v1/tasks', { params })
  },

  // 获取单个任务
  getTask: (id: number) => {
    return request.get(`/v1/tasks/${id}`)
  },

  // 创建任务
  createTask: (data: TaskForm) => {
    return request.post('/v1/tasks', data)
  },

  // 更新任务
  updateTask: (id: number, data: Partial<TaskForm>) => {
    return request.put(`/v1/tasks/${id}`, data)
  },

  // 删除任务
  deleteTask: (id: number) => {
    return request.delete(`/v1/tasks/${id}`)
  },

  // 启用/禁用任务
  toggleTask: (id: number, enabled: boolean) => {
    return request.patch(`/v1/tasks/${id}`, { enabled })
  },

  // 手动触发任务
  triggerTask: (id: number) => {
    return request.post(`/v1/tasks/${id}/trigger`)
  },

  // 获取任务执行历史
  getTaskExecutions: (id: number, params?: any) => {
    return request.get(`/v1/tasks/${id}/executions`, { params })
  }
}

