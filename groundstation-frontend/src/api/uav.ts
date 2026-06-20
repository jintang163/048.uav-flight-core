import { get, post, put, del } from './http'
import type { UAV, UAVListItem, UAVMode } from '@/types'

export const getUAVList = (params?: { page?: number; pageSize?: number; keyword?: string; status?: string }): Promise<{ list: UAVListItem[]; total: number }> => {
  return get<{ list: UAVListItem[]; total: number }>('/uav/list', params)
}

export const getUAVDetail = (id: string): Promise<UAV> => {
  return get<UAV>(`/uav/${id}`)
}

export const createUAV = (data: Partial<UAV>): Promise<UAV> => {
  return post<UAV>('/uav', data)
}

export const updateUAV = (id: string, data: Partial<UAV>): Promise<UAV> => {
  return put<UAV>(`/uav/${id}`, data)
}

export const deleteUAV = (id: string): Promise<void> => {
  return del<void>(`/uav/${id}`)
}

export const setUAVMode = (id: string, mode: UAVMode): Promise<void> => {
  return post<void>(`/uav/${id}/mode`, { mode })
}

export const armUAV = (id: string): Promise<void> => {
  return post<void>(`/uav/${id}/arm`)
}

export const disarmUAV = (id: string): Promise<void> => {
  return post<void>(`/uav/${id}/disarm`)
}

export const takeoff = (id: string, altitude: number): Promise<void> => {
  return post<void>(`/uav/${id}/takeoff`, { altitude })
}

export const land = (id: string): Promise<void> => {
  return post<void>(`/uav/${id}/land`)
}

export const returnToHome = (id: string): Promise<void> => {
  return post<void>(`/uav/${id}/rtl`)
}

export const goTo = (id: string, lat: number, lng: number, alt: number): Promise<void> => {
  return post<void>(`/uav/${id}/goto`, { lat, lng, alt })
}

export const setHome = (id: string, lat: number, lng: number, alt: number): Promise<void> => {
  return post<void>(`/uav/${id}/home`, { lat, lng, alt })
}

export const resetUAV = (id: string): Promise<void> => {
  return post<void>(`/uav/${id}/reset`)
}

export const calibrateSensors = (id: string, type: 'gyro' | 'compass' | 'accelerometer' | 'level'): Promise<void> => {
  return post<void>(`/uav/${id}/calibrate`, { type })
}
