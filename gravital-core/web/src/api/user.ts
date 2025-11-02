import request from '@/utils/request'

export interface UserQuery {
  page?: number
  size?: number
  keyword?: string
  role_id?: number
  enabled?: boolean
}

export interface UserForm {
  username: string
  email: string
  password?: string
  role_id: number
  enabled: boolean
}

export interface RoleForm {
  name: string
  permissions: string[]
  description?: string
}

export interface SystemConfig {
  site_name?: string
  site_url?: string
  alert_email?: string
  alert_webhook?: string
  retention_days?: number
  max_devices?: number
  max_sentinels?: number
}

export const userApi = {
  // 用户管理
  getUsers: (params: UserQuery) => {
    return request.get('/v1/users', { params })
  },

  getUser: (id: number) => {
    return request.get(`/v1/users/${id}`)
  },

  createUser: (data: UserForm) => {
    return request.post('/v1/users', data)
  },

  updateUser: (id: number, data: Partial<UserForm>) => {
    return request.put(`/v1/users/${id}`, data)
  },

  deleteUser: (id: number) => {
    return request.delete(`/v1/users/${id}`)
  },

  toggleUser: (id: number, enabled: boolean) => {
    return request.patch(`/v1/users/${id}`, { enabled })
  },

  resetPassword: (id: number, password: string) => {
    return request.post(`/v1/users/${id}/reset-password`, { password })
  },

  // 角色管理
  getRoles: () => {
    return request.get('/v1/roles')
  },

  getRole: (id: number) => {
    return request.get(`/v1/roles/${id}`)
  },

  createRole: (data: RoleForm) => {
    return request.post('/v1/roles', data)
  },

  updateRole: (id: number, data: Partial<RoleForm>) => {
    return request.put(`/v1/roles/${id}`, data)
  },

  deleteRole: (id: number) => {
    return request.delete(`/v1/roles/${id}`)
  },

  // 系统配置
  getSystemConfig: () => {
    return request.get('/v1/system/config')
  },

  updateSystemConfig: (data: SystemConfig) => {
    return request.put('/v1/system/config', data)
  }
}

