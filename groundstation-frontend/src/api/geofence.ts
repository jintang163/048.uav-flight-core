import { get, post, put, del } from './http'
import type {
  Geofence,
  GeofenceViolation,
  TemporaryUnlocking,
  ViolationStatistics,
  PageResult
} from '@/types'

export const getGeofenceList = (params?: {
  page?: number
  pageSize?: number
  type?: string
  isActive?: boolean
  category?: string
  source?: string
}): Promise<PageResult<Geofence>> => {
  return get<PageResult<Geofence>>('/geofences', params)
}

export const getGeofenceDetail = (id: string): Promise<Geofence> => {
  return get<Geofence>(`/geofences/${id}`)
}

export const createGeofence = (data: Partial<Geofence> & {
  coordinates?: number[][]
  uavIds?: number[]
}): Promise<Geofence> => {
  return post<Geofence>('/geofences', data)
}

export const updateGeofence = (id: string, data: Partial<Geofence> & {
  coordinates?: number[][]
  uavIds?: number[]
}): Promise<Geofence> => {
  return put<Geofence>(`/geofences/${id}`, data)
}

export const deleteGeofence = (id: string): Promise<void> => {
  return del<void>(`/geofences/${id}`)
}

export const getUAVGeofences = (uavId: string): Promise<Geofence[]> => {
  return get<Geofence[]>(`/geofences/uav/${uavId}`)
}

export const checkGeofenceViolation = (uavId: string, lat: number, lng: number, altitude: number): Promise<{
  hasViolation: boolean
  violations: Array<{
    geofenceId: string
    geofenceName: string
    geofenceCategory: string
    violationType: string
    severity: string
    distance: number
    action: string
  }>
}> => {
  return get(`/geofences/uav/${uavId}/check`, { lat, lng, altitude })
}

export const getViolationList = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  geofenceId?: string
  severity?: string
  violationType?: string
  isResolved?: boolean
  startTime?: string
  endTime?: string
}): Promise<PageResult<GeofenceViolation>> => {
  return get<PageResult<GeofenceViolation>>('/geofence-violations', params)
}

export const getViolationDetail = (id: string): Promise<GeofenceViolation> => {
  return get<GeofenceViolation>(`/geofence-violations/${id}`)
}

export const resolveViolation = (id: string, notes?: string): Promise<void> => {
  return post<void>(`/geofence-violations/${id}/resolve`, { notes })
}

export const batchResolveViolations = (ids: string[], notes?: string): Promise<{
  successCount: number
  total: number
}> => {
  return post('/geofence-violations/batch/resolve', { ids, notes })
}

export const getViolationStatistics = (params?: {
  uavId?: string
  geofenceId?: string
  startTime?: string
  endTime?: string
}): Promise<ViolationStatistics> => {
  return get<ViolationStatistics>('/geofence-violations/statistics', params)
}

export const getUnlockingList = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  applicantId?: string
  status?: string
  category?: string
  startTime?: string
  endTime?: string
}): Promise<PageResult<TemporaryUnlocking>> => {
  return get<PageResult<TemporaryUnlocking>>('/temporary-unlockings', params)
}

export const getUnlockingDetail = (id: string): Promise<TemporaryUnlocking> => {
  return get<TemporaryUnlocking>(`/temporary-unlockings/${id}`)
}

export const applyUnlocking = (data: Partial<TemporaryUnlocking>): Promise<TemporaryUnlocking> => {
  return post<TemporaryUnlocking>('/temporary-unlockings', data)
}

export const approveUnlocking = (id: string, remark?: string): Promise<void> => {
  return post<void>(`/temporary-unlockings/${id}/approve`, { remark })
}

export const rejectUnlocking = (id: string, remark?: string): Promise<void> => {
  return post<void>(`/temporary-unlockings/${id}/reject`, { remark })
}

export const cancelUnlocking = (id: string): Promise<void> => {
  return post<void>(`/temporary-unlockings/${id}/cancel`)
}

export const getActiveUnlockings = (uavId: string, category?: string): Promise<TemporaryUnlocking[]> => {
  return get<TemporaryUnlocking[]>(`/temporary-unlockings/uav/${uavId}/active`, { category })
}
