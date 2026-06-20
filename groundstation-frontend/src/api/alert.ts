import { get, post, put } from './http'
import type { Alert, AlertFilter, AlertStats, PageResult } from '@/types'

export const getAlertList = (params?: AlertFilter & { page?: number; pageSize?: number }): Promise<PageResult<Alert>> => {
  return get<PageResult<Alert>>('/alert/list', params)
}

export const getAlertDetail = (id: string): Promise<Alert> => {
  return get<Alert>(`/alert/${id}`)
}

export const acknowledgeAlert = (id: string): Promise<Alert> => {
  return post<Alert>(`/alert/${id}/acknowledge`)
}

export const resolveAlert = (id: string, notes?: string): Promise<Alert> => {
  return post<Alert>(`/alert/${id}/resolve`, { notes })
}

export const batchAcknowledge = (ids: string[]): Promise<void> => {
  return post<void>('/alert/batch-acknowledge', { ids })
}

export const batchResolve = (ids: string[], notes?: string): Promise<void> => {
  return post<void>('/alert/batch-resolve', { ids, notes })
}

export const getAlertStats = (params?: { startTime?: number; endTime?: number; uavId?: string }): Promise<AlertStats> => {
  return get<AlertStats>('/alert/stats', params)
}

export const markAllAsRead = (): Promise<void> => {
  return post<void>('/alert/mark-all-read')
}

export const getNotificationSettings = (): Promise<import('@/types').NotificationSettings> => {
  return get<import('@/types').NotificationSettings>('/alert/notification-settings')
}

export const updateNotificationSettings = (settings: Partial<import('@/types').NotificationSettings>): Promise<import('@/types').NotificationSettings> => {
  return put<import('@/types').NotificationSettings>('/alert/notification-settings', settings)
}
