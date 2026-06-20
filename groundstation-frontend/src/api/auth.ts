import { get, post } from './http'
import type { LoginRequest, LoginResponse, UserInfo } from '@/types'

export const login = (data: LoginRequest): Promise<LoginResponse> => {
  return post<LoginResponse>('/auth/login', data)
}

export const logout = (): Promise<void> => {
  return post<void>('/auth/logout')
}

export const refreshToken = (refreshToken: string): Promise<{ accessToken: string; refreshToken: string }> => {
  return post<{ accessToken: string; refreshToken: string }>('/auth/refresh', { refreshToken })
}

export const getCurrentUser = (): Promise<UserInfo> => {
  return get<UserInfo>('/auth/me')
}

export const changePassword = (data: { oldPassword: string; newPassword: string }): Promise<void> => {
  return post<void>('/auth/password', data)
}
