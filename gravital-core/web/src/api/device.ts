import request from '@/utils/request'
import type { Device, DeviceGroup, DeviceQuery, DeviceCreate, DeviceUpdate } from '@/types/device'
import type { PageResponse } from '@/types/api'

export const deviceApi = {
  // 获取设备列表
  getDevices: (params: DeviceQuery) => 
    request.get<PageResponse<Device>>('/v1/devices', { params }),
  
  // 获取设备详情
  getDevice: (id: string) => 
    request.get<Device>(`/v1/devices/${id}`),
  
  // 创建设备
  createDevice: (data: DeviceCreate) => 
    request.post<{ device_id: string }>('/v1/devices', data),
  
  // 更新设备
  updateDevice: (id: string, data: DeviceUpdate) => 
    request.put(`/v1/devices/${id}`, data),
  
  // 删除设备
  deleteDevice: (id: string) => 
    request.delete(`/v1/devices/${id}`),
  
  // 测试连接
  testConnection: (id: string) => 
    request.post(`/v1/devices/${id}/test-connection`),
  
  // 获取设备分组
  getDeviceGroups: () => 
    request.get<DeviceGroup[]>('/v1/device-groups'),
  
  // 创建设备分组
  createDeviceGroup: (data: { name: string; parent_id?: string; description?: string }) => 
    request.post('/v1/device-groups', data),
  
  // 批量导入设备
  importDevices: (file: File) => {
    const formData = new FormData()
    formData.append('file', file)
    return request.post('/v1/devices/import', formData, {
      headers: { 'Content-Type': 'multipart/form-data' }
    })
  },
  
  // 导出设备
  exportDevices: (params: DeviceQuery) => 
    request.get('/v1/devices/export', { 
      params, 
      responseType: 'blob' 
    })
}

