import { get, post, put } from './http'
import type {
  CollisionAlert,
  RouteIntersection,
  UAVLivePosition,
  CollisionStatus,
  CollisionStats,
  PageResult,
} from '@/types'

export const getCollisionStatus = (): Promise<CollisionStatus> => {
  return get('/collision/status')
}

export const toggleCollisionAvoidance = (enabled: boolean): Promise<{ enabled: boolean }> => {
  return put('/collision/enabled', { enabled })
}

export const getActiveCollisionAlerts = (): Promise<CollisionAlert[]> => {
  return get('/collision/alerts/active')
}

export const listCollisionAlerts = (params: {
  risk_level?: string
  uav_id?: number
  resolved?: boolean
  page?: number
  page_size?: number
}): Promise<PageResult<CollisionAlert>> => {
  return get('/collision/alerts', params as Record<string, unknown>)
}

export const resolveCollisionAlert = (id: number): Promise<void> => {
  return post(`/collision/alerts/${id}/resolve`)
}

export const getAllUAVPositions = (): Promise<Record<string, UAVLivePosition>> => {
  return get('/collision/positions')
}

export const getRouteIntersections = (uavId?: number): Promise<RouteIntersection[]> => {
  const params: Record<string, unknown> = {}
  if (uavId) params.uav_id = uavId
  return get('/collision/intersections', params)
}

export const detectRouteIntersections = (): Promise<{
  intersections: RouteIntersection[]
  count: number
}> => {
  return post('/collision/intersections/detect')
}

export const manualCollisionAvoidance = (params: {
  uav_id: number
  action: string
  param?: number
}): Promise<void> => {
  return post('/collision/manual', params)
}

export const getUAVSpeedFactor = (uavId: number): Promise<{
  uav_id: number
  speed_factor: number
}> => {
  return get(`/collision/uav/${uavId}/speed-factor`)
}

export const getCollisionStats = (params?: {
  start_time?: number
  end_time?: number
}): Promise<CollisionStats> => {
  return get('/collision/stats', params as Record<string, unknown>)
}
