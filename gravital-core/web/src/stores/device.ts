import { defineStore } from 'pinia'
import { ref } from 'vue'
import { deviceApi } from '@/api/device'
import type { Device, DeviceGroup, DeviceQuery } from '@/types/device'
import { ElMessage } from 'element-plus'

export const useDeviceStore = defineStore('device', () => {
  const devices = ref<Device[]>([])
  const currentDevice = ref<Device | null>(null)
  const deviceGroups = ref<DeviceGroup[]>([])
  const total = ref(0)
  const loading = ref(false)

  // 获取设备列表
  const fetchDevices = async (params: DeviceQuery) => {
    loading.value = true
    try {
      const res: any = await deviceApi.getDevices(params)
      devices.value = res.items
      total.value = res.total
    } catch (error) {
      ElMessage.error('获取设备列表失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 获取设备详情
  const fetchDevice = async (id: string) => {
    loading.value = true
    try {
      const res: any = await deviceApi.getDevice(id)
      currentDevice.value = res
    } catch (error) {
      ElMessage.error('获取设备详情失败')
      throw error
    } finally {
      loading.value = false
    }
  }

  // 创建设备
  const createDevice = async (data: any) => {
    try {
      await deviceApi.createDevice(data)
      ElMessage.success('创建设备成功')
    } catch (error) {
      ElMessage.error('创建设备失败')
      throw error
    }
  }

  // 更新设备
  const updateDevice = async (id: string, data: any) => {
    try {
      await deviceApi.updateDevice(id, data)
      ElMessage.success('更新设备成功')
    } catch (error) {
      ElMessage.error('更新设备失败')
      throw error
    }
  }

  // 删除设备
  const deleteDevice = async (id: string) => {
    try {
      await deviceApi.deleteDevice(id)
      ElMessage.success('删除设备成功')
    } catch (error) {
      ElMessage.error('删除设备失败')
      throw error
    }
  }

  // 获取设备分组
  const fetchDeviceGroups = async () => {
    try {
      const res: any = await deviceApi.getDeviceGroups()
      deviceGroups.value = res
    } catch (error) {
      ElMessage.error('获取设备分组失败')
      throw error
    }
  }

  return {
    devices,
    currentDevice,
    deviceGroups,
    total,
    loading,
    fetchDevices,
    fetchDevice,
    createDevice,
    updateDevice,
    deleteDevice,
    fetchDeviceGroups
  }
})

