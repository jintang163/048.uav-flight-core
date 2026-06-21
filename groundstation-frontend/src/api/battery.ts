import { get, post, put, del } from './http'
import type {
  Battery,
  BatteryUsageRecord,
  BatteryCellData,
  BatteryMaintenanceAlert,
  BatteryStatistics,
  CreateBatteryRequest,
  UpdateBatteryRequest,
  BatteryTelemetryRequest
} from '@/types'

export const getBatteryList = (params?: {
  page?: number
  pageSize?: number
  status?: string
  health_status?: string
  uav_id?: string
  keyword?: string
}): Promise<{ list: Battery[]; total: number }> => {
  return get<{ list: Battery[]; total: number }>('/batteries', params)
}

export const getBatteryDetail = (id: string): Promise<Battery> => {
  return get<Battery>(`/batteries/${id}`)
}

export const createBattery = (data: CreateBatteryRequest): Promise<Battery> => {
  return post<Battery>('/batteries', data)
}

export const updateBattery = (id: string, data: UpdateBatteryRequest): Promise<Battery> => {
  return put<Battery>(`/batteries/${id}`, data)
}

export const deleteBattery = (id: string): Promise<void> => {
  return del<void>(`/batteries/${id}`)
}

export const getBatteryStatistics = (): Promise<BatteryStatistics> => {
  return get<BatteryStatistics>('/batteries/statistics')
}

export const getBatteryUsageRecords = (id: string, params?: {
  page?: number
  pageSize?: number
}): Promise<{ list: BatteryUsageRecord[]; total: number }> => {
  return get<{ list: BatteryUsageRecord[]; total: number }>(`/batteries/${id}/usage-records`, params)
}

export const getBatteryCellData = (id: string): Promise<BatteryCellData[]> => {
  return get<BatteryCellData[]>(`/batteries/${id}/cell-data`)
}

export const updateBatteryTelemetry = (id: string, data: BatteryTelemetryRequest): Promise<void> => {
  return post<void>(`/batteries/${id}/telemetry`, data)
}

export const identifyBattery = (batteryId: string): Promise<Battery> => {
  return get<Battery>('/batteries/identify', { battery_id: batteryId })
}

export const getMaintenanceAlerts = (params?: {
  page?: number
  pageSize?: number
  battery_id?: string
  status?: string
  alert_type?: string
  level?: string
}): Promise<{ list: BatteryMaintenanceAlert[]; total: number }> => {
  return get<{ list: BatteryMaintenanceAlert[]; total: number }>('/batteries/maintenance/alerts', params)
}

export const getUnacknowledgedMaintenanceCount = (): Promise<{ count: number }> => {
  return get<{ count: number }>('/batteries/maintenance/alerts/unacknowledged-count')
}

export const acknowledgeMaintenanceAlert = (id: string): Promise<void> => {
  return post<void>(`/batteries/maintenance/alerts/${id}/acknowledge`, {})
}

export const resolveMaintenanceAlert = (id: string, note: string): Promise<void> => {
  return post<void>(`/batteries/maintenance/alerts/${id}/resolve`, { note })
}

export const checkMaintenanceReminders = (maxDays?: number): Promise<{ alerts: BatteryMaintenanceAlert[]; count: number }> => {
  return post<{ alerts: BatteryMaintenanceAlert[]; count: number }>('/batteries/maintenance/check', { max_days: maxDays })
}

export const registerBatteryUse = (id: string, uavId: string): Promise<void> => {
  return post<void>(`/batteries/${id}/register-use`, {}, { params: { uav_id: uavId } })
}

export const updateBatterySOH = (id: string, soh: number): Promise<void> => {
  return put<void>(`/batteries/${id}/soh`, { soh })
}
