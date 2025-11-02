// API 响应基础类型
export interface ApiResponse<T = any> {
  code: number
  message: string
  data: T
}

// 分页查询参数
export interface PageQuery {
  page?: number
  size?: number
  keyword?: string
}

// 分页响应
export interface PageResponse<T> {
  total: number
  items: T[]
  page: number
  size: number
}

// 用户信息
export interface UserInfo {
  id: string
  username: string
  email: string
  role: string
  permissions: string[]
  created_at: string
}

// 登录请求
export interface LoginRequest {
  username: string
  password: string
}

// 登录响应
export interface LoginResponse {
  token: string
  refresh_token: string
  expires_in: number
  user: UserInfo
}

