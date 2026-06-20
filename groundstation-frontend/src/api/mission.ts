import { get, post, put, del } from './http'
import type { Mission, Waypoint, PageResult } from '@/types'

export const getMissionList = (params?: { page?: number; pageSize?: number; keyword?: string; status?: string }): Promise<PageResult<Mission>> => {
  return get<PageResult<Mission>>('/mission/list', params)
}

export const getMissionDetail = (id: string): Promise<Mission> => {
  return get<Mission>(`/mission/${id}`)
}

export const createMission = (data: Partial<Mission>): Promise<Mission> => {
  return post<Mission>('/mission', data)
}

export const updateMission = (id: string, data: Partial<Mission>): Promise<Mission> => {
  return put<Mission>(`/mission/${id}`, data)
}

export const deleteMission = (id: string): Promise<void> => {
  return del<void>(`/mission/${id}`)
}

export const duplicateMission = (id: string): Promise<Mission> => {
  return post<Mission>(`/mission/${id}/duplicate`)
}

export const uploadMission = (uavId: string, missionId: string): Promise<void> => {
  return post<void>(`/mission/${missionId}/upload`, { uavId })
}

export const downloadMission = (uavId: string): Promise<Waypoint[]> => {
  return get<Waypoint[]>(`/mission/download`, { uavId })
}

export const startMission = (uavId: string, missionId: string): Promise<void> => {
  return post<void>(`/mission/${missionId}/start`, { uavId })
}

export const pauseMission = (uavId: string): Promise<void> => {
  return post<void>(`/mission/${uavId}/pause`)
}

export const resumeMission = (uavId: string): Promise<void> => {
  return post<void>(`/mission/${uavId}/resume`)
}

export const stopMission = (uavId: string): Promise<void> => {
  return post<void>(`/mission/${uavId}/stop`)
}

export const setCurrentWaypoint = (uavId: string, waypointIndex: number): Promise<void> => {
  return post<void>(`/mission/${uavId}/waypoint/current`, { index: waypointIndex })
}

export const addWaypoint = (missionId: string, waypoint: Partial<Waypoint>): Promise<Waypoint> => {
  return post<Waypoint>(`/mission/${missionId}/waypoint`, waypoint)
}

export const updateWaypoint = (missionId: string, waypointId: string, data: Partial<Waypoint>): Promise<Waypoint> => {
  return put<Waypoint>(`/mission/${missionId}/waypoint/${waypointId}`, data)
}

export const deleteWaypoint = (missionId: string, waypointId: string): Promise<void> => {
  return del<void>(`/mission/${missionId}/waypoint/${waypointId}`)
}

export const reorderWaypoints = (missionId: string, waypointIds: string[]): Promise<void> => {
  return put<void>(`/mission/${missionId}/waypoints/reorder`, { waypointIds })
}
