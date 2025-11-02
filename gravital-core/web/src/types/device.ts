// 设备类型定义
export interface Device {
  id: string
  device_id: string
  name: string
  device_type: string
  group_id?: string
  sentinel_id?: string
  connection_config: Record<string, any>
  labels: Record<string, string>
  status: 'online' | 'offline' | 'error' | 'unknown'
  last_seen?: string
  created_at: string
  updated_at: string
}

// 设备分组
export interface DeviceGroup {
  id: string
  name: string
  parent_id?: string
  description?: string
  created_at: string
  children?: DeviceGroup[]
}

// 设备查询参数
export interface DeviceQuery {
  page?: number
  size?: number
  keyword?: string
  group_id?: string
  device_type?: string
  status?: string
}

// 创建设备请求
export interface DeviceCreate {
  name: string
  device_type: string
  group_id?: string
  sentinel_id?: string
  connection_config: Record<string, any>
  labels?: Record<string, string>
}

// 更新设备请求
export interface DeviceUpdate {
  name?: string
  device_type?: string
  group_id?: string
  sentinel_id?: string
  connection_config?: Record<string, any>
  labels?: Record<string, string>
}

