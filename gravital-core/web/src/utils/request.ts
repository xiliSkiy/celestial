import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios'
import { ElMessage } from 'element-plus'
import router from '@/router'

// 创建 axios 实例
const request: AxiosInstance = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '/api',
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

// 请求拦截器
request.interceptors.request.use(
  (config: any) => {
    // 添加 token
    const token = localStorage.getItem('token')
    if (token) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    console.error('Request error:', error)
    return Promise.reject(error)
  }
)

// 响应拦截器
request.interceptors.response.use(
  (response: AxiosResponse) => {
    const { data } = response
    
    // 如果返回的是 Blob 类型（文件下载），直接返回
    if (response.config.responseType === 'blob') {
      return response
    }
    
    // 后端统一返回格式: { code: 0, data: {...} }
    // 如果 code 为 0，返回 data 字段
    if (data.code === 0) {
      return data.data
    }
    
    // 如果 code 不为 0，抛出错误
    const error: any = new Error(data.message || '请求失败')
    error.code = data.code
    error.data = data
    return Promise.reject(error)
  },
  (error) => {
    // 如果是业务错误（code 不为 0）
    if (error.code && error.data) {
      ElMessage.error(error.message || '请求失败')
      return Promise.reject(error)
    }
    
    // 如果是 HTTP 错误
    if (error.response) {
      const { status, data } = error.response
      
      switch (status) {
        case 401:
          ElMessage.error('未授权，请重新登录')
          localStorage.removeItem('token')
          router.push('/login')
          break
        case 403:
          ElMessage.error('没有权限访问')
          break
        case 404:
          ElMessage.error('请求的资源不存在')
          break
        case 500:
          ElMessage.error(data?.message || '服务器错误')
          break
        default:
          ElMessage.error(data?.message || '请求失败')
      }
    } else if (error.request) {
      ElMessage.error('网络错误，请检查网络连接')
    } else {
      ElMessage.error(error.message || '请求配置错误')
    }
    
    return Promise.reject(error)
  }
)

export default request

