export * from './uav'
export * from './telemetry'
export * from './mission'
export * from './geofence'
export * from './alert'
export * from './formation'
export * from './tracking'
export * from './payload'
export * from './motor'
export * from './blackbox'

export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
  timestamp: number
}

export interface PageResult<T> {
  list: T[]
  total: number
  page: number
  pageSize: number
}

export interface LoginRequest {
  username: string
  password: string
}

export interface LoginResponse {
  accessToken: string
  refreshToken: string
  user: UserInfo
}

export interface UserInfo {
  id: string
  username: string
  nickname: string
  avatar?: string
  roles: string[]
  permissions: string[]
}

export interface AuthState {
  isAuthenticated: boolean
  user: UserInfo | null
  accessToken: string | null
  refreshToken: string | null
  loading: boolean
  error: string | null
}

export interface FlightRecord {
  id: string
  uavId: string
  uavName: string
  startTime: number
  endTime: number
  duration: number
  maxAltitude: number
  maxSpeed: number
  distance: number
  startPosition: { lat: number; lng: number }
  endPosition: { lat: number; lng: number }
  purpose?: string
  pilot?: string
  notes?: string
}

export interface FirmwareInfo {
  id: string
  name: string
  version: string
  hardware: string
  fileSize: number
  releaseDate: number
  description?: string
  changelog?: string
  checksum?: string
}

export interface SystemSettings {
  theme: 'light' | 'dark'
  language: 'zh-CN' | 'en-US'
  mapType: 'standard' | 'satellite' | 'hybrid'
  unitSystem: 'metric' | 'imperial'
  autoRefresh: boolean
  refreshInterval: number
  notificationSettings: NotificationSettings
}
