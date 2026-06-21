import { get, post, put, del } from './http'
import type { BlackboxLog, ParsedLogData, AnalysisReport, LogAnalysisReport } from '@/types/blackbox'

export const getBlackboxList = (params?: {
  page?: number
  page_size?: number
  uav_id?: string
  mission_id?: string
  status?: string
  crash_detected?: boolean
}): Promise<{ list: BlackboxLog[]; total: number }> => {
  return get<{ list: BlackboxLog[]; total: number }>('/blackbox', params)
}

export const getBlackboxDetail = (id: string): Promise<BlackboxLog> => {
  return get<BlackboxLog>(`/blackbox/${id}`)
}

export const uploadBlackbox = (data: FormData): Promise<BlackboxLog> => {
  return post<BlackboxLog>('/blackbox', data)
}

export const updateBlackbox = (id: string, data: Partial<BlackboxLog>): Promise<BlackboxLog> => {
  return put<BlackboxLog>(`/blackbox/${id}`, data)
}

export const deleteBlackbox = (id: string): Promise<void> => {
  return del<void>(`/blackbox/${id}`)
}

export const downloadBlackbox = (id: string): Promise<{ download_url: string; file_name: string }> => {
  return get<{ download_url: string; file_name: string }>(`/blackbox/${id}/download`)
}

export const parseBlackboxLog = (id: string): Promise<ParsedLogData> => {
  return get<ParsedLogData>(`/blackbox/${id}/parse`)
}

export const getAnalysisReport = (id: string): Promise<AnalysisReport> => {
  return get<AnalysisReport>(`/blackbox/${id}/analysis`)
}

export const analyzeBlackbox = (id: string): Promise<void> => {
  return post<void>(`/blackbox/${id}/analyze`)
}

export const exportBlackboxCSV = (id: string): string => {
  const token = localStorage.getItem('accessToken')
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
  return `${baseURL}/v1/blackbox/${id}/export/csv?access_token=${token}`
}

export const exportBlackboxReport = (id: string): string => {
  const token = localStorage.getItem('accessToken')
  const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'
  return `${baseURL}/v1/blackbox/${id}/export/report?access_token=${token}`
}

export const getBlackboxReports = (id: string): Promise<LogAnalysisReport[]> => {
  return get<LogAnalysisReport[]>(`/blackbox/${id}/reports`)
}

export const getBlackboxStatistics = (params?: {
  uav_id?: string
  start_time?: string
  end_time?: string
}): Promise<Record<string, unknown>> => {
  return get<Record<string, unknown>>('/blackbox/statistics', params)
}

export const autoUploadBlackbox = (data: {
  uav_id: string | number
  mission_id?: string | number
  file?: File
}): Promise<BlackboxLog> => {
  const formData = new FormData()
  formData.append('uav_id', String(data.uav_id))
  if (data.mission_id !== undefined && data.mission_id !== null) {
    formData.append('mission_id', String(data.mission_id))
  }
  if (data.file) {
    formData.append('file', data.file)
  }
  return post<BlackboxLog>('/blackbox/auto-upload', formData)
}
