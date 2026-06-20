import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  fetchMissions,
  fetchMissionDetail,
  createNewMission,
  updateExistingMission,
  deleteMissionById,
  selectMission,
  setCurrentMission,
  addWaypointToMission,
  updateWaypointInMission,
  removeWaypointFromMission,
  reorderWaypointsInMission
} from '@/store/slices/mission'
import type { Mission, Waypoint, WaypointAction } from '@/types'
import { generateId } from '@/utils'
import {
  startMission as apiStartMission,
  pauseMission as apiPauseMission,
  resumeMission as apiResumeMission,
  resumeMissionFromBreakpoint as apiResumeMissionFromBreakpoint,
  stopMission as apiStopMission,
  setCurrentWaypoint as apiSetCurrentWaypoint,
  uploadMission as apiUploadMission,
  downloadMission as apiDownloadMission,
  exportMission as apiExportMission,
  importMission as apiImportMission,
  addWaypoint as apiAddWaypoint,
  updateWaypoint as apiUpdateWaypoint,
  deleteWaypoint as apiDeleteWaypoint,
  reorderWaypoints as apiReorderWaypoints
} from '@/api/mission'

export const useMission = (missionId?: string) => {
  const dispatch = useAppDispatch()
  const { missions, currentMission, waypoints, executionState, loading, listLoading, error, total } = useAppSelector(state => state.mission)

  const loadMissions = useCallback((params?: { page?: number; pageSize?: number; keyword?: string; status?: string }) => {
    dispatch(fetchMissions(params || {}))
  }, [dispatch])

  const loadMissionDetail = useCallback((id: string) => {
    dispatch(fetchMissionDetail(id))
  }, [dispatch])

  const createMission = useCallback((data: Partial<Mission>) => {
    return dispatch(createNewMission({
      name: data.name || '新航线',
      description: data.description,
      waypoints: data.waypoints || [],
      status: 'draft',
      createdAt: Date.now(),
      updatedAt: Date.now(),
      ...data
    }))
  }, [dispatch])

  const updateMission = useCallback((id: string, data: Partial<Mission>) => {
    return dispatch(updateExistingMission({ id, data }))
  }, [dispatch])

  const deleteMission = useCallback((id: string) => {
    return dispatch(deleteMissionById(id))
  }, [dispatch])

  const selectCurrentMission = useCallback((id: string | null) => {
    dispatch(selectMission(id))
  }, [dispatch])

  const setMission = useCallback((mission: Mission | null) => {
    dispatch(setCurrentMission(mission))
  }, [dispatch])

  const addWaypoint = useCallback((waypoint: Omit<Waypoint, 'id' | 'sequence'>) => {
    const newWaypoint: Waypoint = {
      ...waypoint,
      id: generateId(),
      sequence: waypoints.length + 1
    }
    dispatch(addWaypointToMission(newWaypoint))
  }, [dispatch, waypoints.length])

  const updateWaypoint = useCallback((waypoint: Waypoint) => {
    dispatch(updateWaypointInMission(waypoint))
  }, [dispatch])

  const deleteWaypoint = useCallback((waypointId: string) => {
    dispatch(removeWaypointFromMission(waypointId))
  }, [dispatch])

  const reorderWaypoints = useCallback((waypointIds: string[]) => {
    dispatch(reorderWaypointsInMission(waypointIds))
  }, [dispatch])

  const setWaypointAction = useCallback((waypointId: string, actionType: WaypointAction) => {
    const waypoint = waypoints.find(w => w.id === waypointId)
    if (waypoint) {
      updateWaypoint({ ...waypoint, action: actionType })
    }
  }, [waypoints, updateWaypoint])

  const duplicateMission = useCallback((id: string) => {
    const mission = missions.find(m => m.id === id)
    if (mission) {
      return createMission({
        ...mission,
        name: `${mission.name} (副本)`,
        status: 'draft'
      })
    }
  }, [missions, createMission])

  const startMission = useCallback((uavId: string, missionId: string) => {
    return apiStartMission(uavId, missionId)
  }, [])

  const pauseMission = useCallback((missionId: string) => {
    return apiPauseMission(missionId)
  }, [])

  const resumeMission = useCallback((missionId: string) => {
    return apiResumeMission(missionId)
  }, [])

  const resumeMissionFromBreakpoint = useCallback((missionId: string) => {
    return apiResumeMissionFromBreakpoint(missionId)
  }, [])

  const stopMission = useCallback((missionId: string) => {
    return apiStopMission(missionId)
  }, [])

  const setCurrentWaypoint = useCallback((missionId: string, index: number) => {
    return apiSetCurrentWaypoint(missionId, index)
  }, [])

  const uploadMission = useCallback((uavId: string, missionId: string) => {
    return apiUploadMission(uavId, missionId)
  }, [])

  const downloadMission = useCallback((uavId: string) => {
    return apiDownloadMission(uavId)
  }, [])

  const exportMission = useCallback((missionId: string) => {
    return apiExportMission(missionId)
  }, [])

  const importMission = useCallback((file: File) => {
    return apiImportMission(file)
  }, [])

  useEffect(() => {
    if (missionId) {
      loadMissionDetail(missionId)
    }
  }, [missionId, loadMissionDetail])

  return {
    missions,
    currentMission,
    waypoints,
    executionState,
    loading,
    listLoading,
    error,
    total,
    loadMissions,
    loadMissionDetail,
    createMission,
    updateMission,
    deleteMission,
    selectCurrentMission,
    selectMission: selectCurrentMission,
    setMission,
    addWaypoint,
    updateWaypoint,
    deleteWaypoint,
    reorderWaypoints,
    setWaypointAction,
    duplicateMission,
    startMission,
    pauseMission,
    resumeMission,
    resumeMissionFromBreakpoint,
    stopMission,
    setCurrentWaypoint,
    uploadMission,
    downloadMission,
    exportMission,
    importMission
  }
}
