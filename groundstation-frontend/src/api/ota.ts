import { get, post, put, del } from './http'
import type { FirmwareInfo } from '@/types'

export const getFirmwareList = (params?: { page?: number; pageSize?: number; hardware?: string }): Promise<{ list: FirmwareInfo[]; total: number }> => {
  return get<{ list: FirmwareInfo[]; total: number }>('/ota/firmware/list', params)
}

export const getFirmwareDetail = (id: string): Promise<FirmwareInfo> => {
  return get<FirmwareInfo>(`/ota/firmware/${id}`)
}

export const uploadFirmware = (file: File, data: { name: string; version: string; hardware: string; description?: string; changelog?: string }): Promise<FirmwareInfo> => {
  const formData = new FormData()
  formData.append('file', file)
  Object.entries(data).forEach(([key, value]) => {
    formData.append(key, value)
  })
  return post<FirmwareInfo>('/ota/firmware/upload', formData)
}

export const updateFirmware = (id: string, data: Partial<FirmwareInfo>): Promise<FirmwareInfo> => {
  return put<FirmwareInfo>(`/ota/firmware/${id}`, data)
}

export const deleteFirmware = (id: string): Promise<void> => {
  return del<void>(`/ota/firmware/${id}`)
}

export const startFirmwareUpdate = (uavId: string, firmwareId: string): Promise<{ taskId: string }> => {
  return post<{ taskId: string }>('/ota/update/start', { uavId, firmwareId })
}

export const getUpdateProgress = (taskId: string): Promise<{
  taskId: string
  uavId: string
  status: 'pending' | 'uploading' | 'verifying' | 'flashing' | 'rebooting' | 'success' | 'failed'
  progress: number
  message?: string
  startTime: number
  endTime?: number
}> => {
  return get<{
    taskId: string
    uavId: string
    status: 'pending' | 'uploading' | 'verifying' | 'flashing' | 'rebooting' | 'success' | 'failed'
    progress: number
    message?: string
    startTime: number
    endTime?: number
  }>(`/ota/update/${taskId}/progress`)
}

export const cancelFirmwareUpdate = (taskId: string): Promise<void> => {
  return post<void>(`/ota/update/${taskId}/cancel`)
}

export const getLatestFirmware = (hardware: string): Promise<FirmwareInfo | null> => {
  return get<FirmwareInfo | null>('/ota/firmware/latest', { hardware })
}

export const checkFirmwareVersion = (uavId: string): Promise<{
  currentVersion: string
  latestVersion: string
  hasUpdate: boolean
  firmware?: FirmwareInfo
}> => {
  return get<{
    currentVersion: string
    latestVersion: string
    hasUpdate: boolean
    firmware?: FirmwareInfo
  }>(`/ota/check/${uavId}`)
}
