import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  fetchGeofences,
  fetchGeofenceDetail,
  createNewGeofence,
  updateExistingGeofence,
  deleteGeofenceById,
  toggleGeofenceById,
  fetchViolations,
  fetchRestrictionZones,
  selectGeofence,
  setDrawMode,
  setDrawType,
  setEditingGeofence
} from '@/store/slices/geofence'
import type { Geofence, GeofenceType, GeofenceShape } from '@/types'

export const useGeofence = () => {
  const dispatch = useAppDispatch()
  const {
    geofences,
    selectedGeofence,
    violations,
    restrictionZones,
    loading,
    listLoading,
    violationLoading,
    error,
    total,
    violationTotal,
    drawMode,
    drawType,
    editingGeofence
  } = useAppSelector(state => state.geofence)

  const loadGeofences = useCallback((params?: { page?: number; pageSize?: number; keyword?: string; isEnabled?: boolean }) => {
    dispatch(fetchGeofences(params || {}))
  }, [dispatch])

  const loadGeofenceDetail = useCallback((id: string) => {
    dispatch(fetchGeofenceDetail(id))
  }, [dispatch])

  const createGeofence = useCallback((data: Partial<Geofence>) => {
    return dispatch(createNewGeofence({
      name: data.name || '新围栏',
      isEnabled: true,
      isInclusion: false,
      action: 'warn',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      ...data
    }))
  }, [dispatch])

  const updateGeofence = useCallback((id: string, data: Partial<Geofence>) => {
    return dispatch(updateExistingGeofence({ id, data }))
  }, [dispatch])

  const deleteGeofence = useCallback((id: string) => {
    return dispatch(deleteGeofenceById(id))
  }, [dispatch])

  const toggleGeofence = useCallback((id: string, enabled: boolean) => {
    return dispatch(toggleGeofenceById({ id, enabled }))
  }, [dispatch])

  const loadViolations = useCallback((params?: { page?: number; pageSize?: number; geofenceId?: string; uavId?: string; resolved?: boolean }) => {
    dispatch(fetchViolations(params || {}))
  }, [dispatch])

  const loadRestrictionZones = useCallback((params: { lat: number; lng: number; radius: number }) => {
    dispatch(fetchRestrictionZones(params))
  }, [dispatch])

  const selectFence = useCallback((geofence: Geofence | null) => {
    dispatch(selectGeofence(geofence))
  }, [dispatch])

  const startDraw = useCallback((type: GeofenceType) => {
    dispatch(setDrawType(type))
    dispatch(setDrawMode(true))
  }, [dispatch])

  const stopDraw = useCallback(() => {
    dispatch(setDrawMode(false))
    dispatch(setDrawType(null))
  }, [dispatch])

  const startEdit = useCallback((geofence: Geofence) => {
    dispatch(setEditingGeofence(geofence))
  }, [dispatch])

  const stopEdit = useCallback(() => {
    dispatch(setEditingGeofence(null))
  }, [dispatch])

  const updateShape = useCallback((id: string, shape: GeofenceShape) => {
    updateGeofence(id, { shape, updatedAt: Date.now() })
  }, [updateGeofence])

  useEffect(() => {
    loadGeofences()
  }, [loadGeofences])

  return {
    geofences,
    selectedGeofence,
    violations,
    restrictionZones,
    loading,
    listLoading,
    violationLoading,
    error,
    total,
    violationTotal,
    drawMode,
    drawType,
    editingGeofence,
    loadGeofences,
    loadGeofenceDetail,
    createGeofence,
    updateGeofence,
    deleteGeofence,
    toggleGeofence,
    loadViolations,
    loadRestrictionZones,
    selectFence,
    startDraw,
    stopDraw,
    startEdit,
    stopEdit,
    updateShape
  }
}
