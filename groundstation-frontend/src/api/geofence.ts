import { get, post, put, del } from './http'
import type { Geofence, GeofenceViolation, FlightRestrictionZone, PageResult } from '@/types'

export const getGeofenceList = (params?: { page?: number; pageSize?: number; keyword?: string; isEnabled?: boolean }): Promise<PageResult<Geofence>> => {
  return get<PageResult<Geofence>>('/geofence/list', params)
}

export const getGeofenceDetail = (id: string): Promise<Geofence> => {
  return get<Geofence>(`/geofence/${id}`)
}

export const createGeofence = (data: Partial<Geofence>): Promise<Geofence> => {
  return post<Geofence>('/geofence', data)
}

export const updateGeofence = (id: string, data: Partial<Geofence>): Promise<Geofence> => {
  return put<Geofence>(`/geofence/${id}`, data)
}

export const deleteGeofence = (id: string): Promise<void> => {
  return del<void>(`/geofence/${id}`)
}

export const toggleGeofence = (id: string, enabled: boolean): Promise<void> => {
  return put<void>(`/geofence/${id}/toggle`, { enabled })
}

export const getViolationList = (params?: { page?: number; pageSize?: number; geofenceId?: string; uavId?: string; resolved?: boolean }): Promise<PageResult<GeofenceViolation>> => {
  return get<PageResult<GeofenceViolation>>('/geofence/violations', params)
}

export const resolveViolation = (id: string): Promise<void> => {
  return post<void>(`/geofence/violation/${id}/resolve`)
}

export const getRestrictionZones = (params?: { lat: number; lng: number; radius: number }): Promise<FlightRestrictionZone[]> => {
  return get<FlightRestrictionZone[]>('/geofence/restrictions', params)
}

export const checkPosition = (lat: number, lng: number, alt: number): Promise<{
  insideFence: Geofence[]
  outsideFence: Geofence[]
  restrictions: FlightRestrictionZone[]
}> => {
  return get<{
    insideFence: Geofence[]
    outsideFence: Geofence[]
    restrictions: FlightRestrictionZone[]
  }>('/geofence/check', { lat, lng, alt })
}
