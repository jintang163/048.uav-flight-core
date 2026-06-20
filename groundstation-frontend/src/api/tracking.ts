import { get, post } from './http'
import type {
  LockTargetRequest,
  TrackingTask,
  DetectionTarget,
  PageResult
} from '@/types'

export const lockTarget = (data: LockTargetRequest): Promise<TrackingTask> => {
  return post<TrackingTask>('/tracking/lock', data)
}

export const stopTracking = (id: string): Promise<void> => {
  return post<void>(`/tracking/${id}/stop`, {})
}

export const getTrackingTask = (id: string): Promise<TrackingTask> => {
  return get<TrackingTask>(`/tracking/${id}`)
}

export const listTrackingTasks = (params: {
  page?: number
  pageSize?: number
  uav_id?: string
  status?: string
}): Promise<PageResult<TrackingTask>> => {
  return get<PageResult<TrackingTask>>('/tracking', params)
}

export const getActiveTracking = (uavId: string): Promise<TrackingTask | null> => {
  return get<TrackingTask | null>(`/tracking/uav/${uavId}/active`)
}

export const listDetections = (
  uavId: string,
  params?: { page?: number; pageSize?: number }
): Promise<PageResult<DetectionTarget>> => {
  return get<PageResult<DetectionTarget>>(`/tracking/uav/${uavId}/detections`, params)
}

export const detectImage = (
  uavId: string,
  imageFile: File,
  frameWidth?: number,
  frameHeight?: number
): Promise<DetectionTarget[]> => {
  const formData = new FormData()
  formData.append('image', imageFile)
  const params = new URLSearchParams()
  if (frameWidth) params.append('width', String(frameWidth))
  if (frameHeight) params.append('height', String(frameHeight))
  return post<DetectionTarget[]>(
    `/tracking/uav/${uavId}/detect${params.toString() ? '?' + params.toString() : ''}`,
    formData
  )
}
