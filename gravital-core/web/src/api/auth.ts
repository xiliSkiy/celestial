import request from '@/utils/request'
import type { LoginRequest, LoginResponse, UserInfo } from '@/types/api'

export const authApi = {
  // 登录
  login: (data: LoginRequest) => 
    request.post<LoginResponse>('/v1/auth/login', data),
  
  // 刷新 token
  refreshToken: (refreshToken: string) => 
    request.post('/v1/auth/refresh', { refresh_token: refreshToken }),
  
  // 登出
  logout: () => 
    request.post('/v1/auth/logout'),
  
  // 获取当前用户信息
  getUserInfo: () => 
    request.get<UserInfo>('/v1/auth/me')
}

