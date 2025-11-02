import { defineStore } from 'pinia'
import { ref } from 'vue'
import { authApi } from '@/api/auth'
import type { UserInfo } from '@/types/api'
import { ElMessage } from 'element-plus'
import router from '@/router'

export const useUserStore = defineStore('user', () => {
  const token = ref('')
  const refreshToken = ref('')
  const userInfo = ref<UserInfo | null>(null)
  const permissions = ref<string[]>([])

  // 登录
  const login = async (username: string, password: string) => {
    try {
      const res: any = await authApi.login({ username, password })
      token.value = res.token
      refreshToken.value = res.refresh_token
      userInfo.value = res.user
      // 权限在 user.role.permissions 中
      permissions.value = res.user?.role?.permissions || []
      
      // 保存到本地存储
      localStorage.setItem('token', res.token)
      localStorage.setItem('refreshToken', res.refresh_token)
      
      ElMessage.success('登录成功')
      router.push('/')
    } catch (error) {
      ElMessage.error('登录失败')
      throw error
    }
  }

  // 登出
  const logout = async () => {
    try {
      await authApi.logout()
    } catch (error) {
      console.error('Logout error:', error)
    } finally {
      token.value = ''
      refreshToken.value = ''
      userInfo.value = null
      permissions.value = []
      
      localStorage.removeItem('token')
      localStorage.removeItem('refreshToken')
      
      router.push('/login')
    }
  }

  // 获取用户信息
  const getUserInfo = async () => {
    try {
      const res: any = await authApi.getUserInfo()
      userInfo.value = res
      // 权限在 user.role.permissions 中
      permissions.value = res?.role?.permissions || []
    } catch (error) {
      console.error('Get user info error:', error)
      logout()
    }
  }

  // 检查权限
  const hasPermission = (permission: string) => {
    return permissions.value.includes(permission) || permissions.value.includes('*')
  }

  return {
    token,
    refreshToken,
    userInfo,
    permissions,
    login,
    logout,
    getUserInfo,
    hasPermission
  }
})

