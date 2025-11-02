// Sentinel 类型定义
export interface Sentinel {
  id: string
  sentinel_id: string
  name: string
  region: string
  ip_address: string
  hostname: string
  os_type: string
  os_version: string
  arch: string
  version: string
  status: 'online' | 'offline'
  last_heartbeat: string
  cpu_usage?: number
  memory_usage?: number
  disk_usage?: number
  device_count?: number
  task_count?: number
  created_at: string
  updated_at: string
}

// Sentinel 查询参数
export interface SentinelQuery {
  page?: number
  size?: number
  keyword?: string
  status?: string
  region?: string
}

// Sentinel 统计信息
export interface SentinelStats {
  sentinel_id: string
  cpu_usage: number
  memory_usage: number
  disk_usage: number
  device_count: number
  task_count: number
  uptime: number
}

