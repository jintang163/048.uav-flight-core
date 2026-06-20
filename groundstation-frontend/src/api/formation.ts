import { get, post, put, del } from './http'
import type {
  Formation,
  FormationMember,
  FormationCollisionWarning,
  FormationLightConfig,
  CreateFormationRequest,
  UpdateFormationRequest
} from '@/types'

export const getFormationList = (params?: {
  page?: number
  pageSize?: number
}): Promise<{ list: Formation[]; total: number }> => {
  return get<{ list: Formation[]; total: number }>('/formations', params)
}

export const getFormationDetail = (id: string): Promise<Formation> => {
  return get<Formation>(`/formations/${id}`)
}

export const createFormation = (data: CreateFormationRequest): Promise<Formation> => {
  return post<Formation>('/formations', data)
}

export const updateFormation = (
  id: string,
  data: UpdateFormationRequest
): Promise<Formation> => {
  return put<Formation>(`/formations/${id}`, data)
}

export const deleteFormation = (id: string): Promise<void> => {
  return del<void>(`/formations/${id}`)
}

export const getFormationMembers = (id: string): Promise<FormationMember[]> => {
  return get<FormationMember[]>(`/formations/${id}/members`)
}

export const addFormationMember = (id: string, uavId: string): Promise<void> => {
  return post<void>(`/formations/${id}/members`, { uav_id: uavId })
}

export const removeFormationMember = (id: string, uavId: string): Promise<void> => {
  return del<void>(`/formations/${id}/members/${uavId}`)
}

export const setFormationLeader = (id: string, uavId: string): Promise<void> => {
  return post<void>(`/formations/${id}/leader/${uavId}`)
}

export const startFormation = (id: string): Promise<void> => {
  return post<void>(`/formations/${id}/start`)
}

export const pauseFormation = (id: string): Promise<void> => {
  return post<void>(`/formations/${id}/pause`)
}

export const resumeFormation = (id: string): Promise<void> => {
  return post<void>(`/formations/${id}/resume`)
}

export const stopFormation = (id: string): Promise<void> => {
  return post<void>(`/formations/${id}/stop`)
}

export const getActiveFormations = (): Promise<Formation[]> => {
  return get<Formation[]>('/formations/active')
}

export const getCollisionWarnings = (
  id: string,
  params?: { page?: number; pageSize?: number }
): Promise<{ list: FormationCollisionWarning[]; total: number }> => {
  return get<{ list: FormationCollisionWarning[]; total: number }>(
    `/formations/${id}/collisions`,
    params
  )
}

export const setFormationLight = (
  id: string,
  config: FormationLightConfig
): Promise<void> => {
  return post<void>(`/formations/${id}/light`, config)
}

export const syncFormationWaypoints = (
  id: string,
  missionId: string
): Promise<void> => {
  return post<void>(`/formations/${id}/sync-waypoints`, { mission_id: missionId })
}

export const multiTakeoff = (
  id: string,
  altitude?: number
): Promise<void> => {
  return post<void>(`/formations/${id}/takeoff`, { altitude: altitude || 5.0 })
}
