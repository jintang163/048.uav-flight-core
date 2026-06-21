import { get, post } from './http'
import type { LinkStatus, LinkStatusReport, LinkStatistics } from '@/types/link'

export const getLinkStatus = (uavId: string): Promise<LinkStatus> => {
  return get<LinkStatus>(`/link/${uavId}/latest`)
}

export const getLinkHistory = (uavId: string, params?: {
  page?: number
  page_size?: number
  start_time?: string
  end_time?: string
}): Promise<{ list: LinkStatus[]; total: number }> => {
  return get<{ list: LinkStatus[]; total: number }>(`/link/${uavId}/history`, params)
}

export const getLinkStatistics = (params?: {
  uav_id?: string
  start_time?: string
  end_time?: string
}): Promise<LinkStatistics> => {
  return get<LinkStatistics>('/link/statistics', params)
}

export const reportLinkStatus = (data: LinkStatusReport): Promise<LinkStatus> => {
  return post<LinkStatus>('/link/status', data)
}
