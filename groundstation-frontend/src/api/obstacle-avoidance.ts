import { get, post, put, del } from './http'
import type {
  ObstacleAvoidanceConfig,
  ObstacleAvoidanceLog,
  ObstacleAvoidanceStatistics,
  ObstacleHeatmapPoint,
  AvoidanceSensitivity,
  AvoidanceStrategy,
  ObstacleAvoidanceEvent,
  PageResult
} from '@/types'

export const getObstacleAvoidanceConfig = (uavId: string): Promise<ObstacleAvoidanceConfig> => {
  return get<ObstacleAvoidanceConfig>(`/obstacle-avoidance/config/${uavId}`)
}

export const updateObstacleAvoidanceConfig = (uavId: string, data: {
  enabled?: boolean
  sensitivity?: AvoidanceSensitivity
  strategy?: AvoidanceStrategy
  detectionRange?: number
  ascendHeight?: number
  retreatDistance?: number
  bypassAngle?: number
}): Promise<ObstacleAvoidanceConfig> => {
  return put<ObstacleAvoidanceConfig>(`/obstacle-avoidance/config/${uavId}`, data)
}

export const getObstacleHeatmap = (params?: {
  uavId?: string
  startTime?: number
  endTime?: number
  bounds?: { minLat: number; maxLat: number; minLng: number; maxLng: number }
}): Promise<ObstacleHeatmapPoint[]> => {
  return get<ObstacleHeatmapPoint[]>('/obstacle-avoidance/heatmap', params)
}

export const getObstacleAvoidanceLogs = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  strategy?: AvoidanceStrategy
  status?: string
  startTime?: number
  endTime?: number
}): Promise<PageResult<ObstacleAvoidanceLog>> => {
  return get<PageResult<ObstacleAvoidanceLog>>('/obstacle-avoidance/logs', params)
}

export const getObstacleAvoidanceStatistics = (params?: {
  uavId?: string
  startTime?: number
  endTime?: number
}): Promise<ObstacleAvoidanceStatistics> => {
  return get<ObstacleAvoidanceStatistics>('/obstacle-avoidance/statistics', params)
}

export const getAvoidanceEvents = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  status?: string
}): Promise<PageResult<ObstacleAvoidanceEvent>> => {
  return get<PageResult<ObstacleAvoidanceEvent>>('/obstacle-avoidance/events', params)
}

export const getAvoidanceEventDetail = (id: string): Promise<ObstacleAvoidanceEvent> => {
  return get<ObstacleAvoidanceEvent>(`/obstacle-avoidance/events/${id}`)
}

export const triggerAvoidanceTest = (uavId: string, strategy: AvoidanceStrategy): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>(`/obstacle-avoidance/test/${uavId}`, { strategy })
}

export const clearHeatmapData = (uavId?: string): Promise<void> => {
  return del<void>('/obstacle-avoidance/heatmap', { uavId })
}
