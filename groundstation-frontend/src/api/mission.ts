import { get, post, put, del } from './http'
import type { Mission, Waypoint, PageResult } from '@/types'

export const getMissionList = (params?: { page?: number; pageSize?: number; keyword?: string; status?: string; uav_id?: string }): Promise<PageResult<Mission>> => {
  return get<PageResult<Mission>>('/missions', params)
}

export const getMissionDetail = (id: string): Promise<Mission> => {
  return get<Mission>(`/missions/${id}`)
}

export const createMission = (data: Partial<Mission>): Promise<Mission> => {
  return post<Mission>('/missions', data)
}

export const updateMission = (id: string, data: Partial<Mission>): Promise<Mission> => {
  return put<Mission>(`/missions/${id}`, data)
}

export const deleteMission = (id: string): Promise<void> => {
  return del<void>(`/missions/${id}`)
}

export const duplicateMission = (id: string): Promise<Mission> => {
  return post<Mission>(`/missions/${id}/duplicate`)
}

export const uploadMission = (uavId: string, missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/upload`, { uavId })
}

export const downloadMission = (uavId: string): Promise<Waypoint[]> => {
  return get<Waypoint[]>('/missions/download', { uavId })
}

export const startMission = (_uavId: string, missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/start`)
}

export const pauseMission = (missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/pause`)
}

export const resumeMission = (missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/resume`)
}

export const resumeMissionFromBreakpoint = (missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/resume-breakpoint`)
}

export const stopMission = (missionId: string): Promise<void> => {
  return post<void>(`/missions/${missionId}/stop`)
}

export const setCurrentWaypoint = (missionId: string, waypointIndex: number): Promise<void> => {
  return post<void>(`/missions/${missionId}/waypoint/current`, { index: waypointIndex })
}

export const addWaypoint = (missionId: string, waypoint: Partial<Waypoint>): Promise<Waypoint> => {
  return post<Waypoint>(`/missions/${missionId}/waypoint`, waypoint)
}

export const updateWaypoint = (missionId: string, waypointId: string, data: Partial<Waypoint>): Promise<Waypoint> => {
  return put<Waypoint>(`/missions/${missionId}/waypoint/${waypointId}`, data)
}

export const deleteWaypoint = (missionId: string, waypointId: string): Promise<void> => {
  return del<void>(`/missions/${missionId}/waypoint/${waypointId}`)
}

export const reorderWaypoints = (missionId: string, waypointIds: string[]): Promise<void> => {
  return put<void>(`/missions/${missionId}/waypoints/reorder`, { waypointIds })
}

export const exportMission = (missionId: string): Promise<void> => {
  return get<void>(`/missions/${missionId}/export`)
}

export const importMission = (file: File): Promise<Mission> => {
  const formData = new FormData()
  formData.append('file', file)
  return post<Mission>('/missions/import', formData)
}

export const getTemplateList = (params?: { page?: number; pageSize?: number; category?: string; keyword?: string }): Promise<PageResult<Mission>> => {
  return get<PageResult<Mission>>('/missions/templates', params)
}

export const getTemplateDetail = (id: string): Promise<Mission> => {
  return get<Mission>(`/missions/templates/${id}`)
}

export const createTemplate = (data: Partial<Mission>): Promise<Mission> => {
  return post<Mission>('/missions/templates', data)
}

export const updateTemplate = (id: string, data: Partial<Mission>): Promise<Mission> => {
  return put<Mission>(`/missions/templates/${id}`, data)
}

export const deleteTemplate = (id: string): Promise<void> => {
  return del<void>(`/missions/templates/${id}`)
}
