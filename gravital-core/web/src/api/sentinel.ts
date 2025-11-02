import request from '@/utils/request'
import type { Sentinel, SentinelQuery, SentinelStats } from '@/types/sentinel'
import type { PageResponse } from '@/types/api'

export const sentinelApi = {
  // 获取 Sentinel 列表
  getSentinels: (params: SentinelQuery) => 
    request.get<PageResponse<Sentinel>>('/v1/sentinels', { params }),
  
  // 获取 Sentinel 详情
  getSentinel: (id: string) => 
    request.get<Sentinel>(`/v1/sentinels/${id}`),
  
  // 获取 Sentinel 统计信息
  getSentinelStats: (id: string) => 
    request.get<SentinelStats>(`/v1/sentinels/${id}/stats`),
  
  // 控制 Sentinel（重启、停止等）
  controlSentinel: (id: string, action: 'restart' | 'stop' | 'update') => 
    request.post(`/v1/sentinels/${id}/control`, { action }),
  
  // 删除 Sentinel
  deleteSentinel: (id: string) => 
    request.delete(`/v1/sentinels/${id}`)
}

