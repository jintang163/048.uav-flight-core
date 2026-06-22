import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  setEnabled,
  setSensitivity,
  setStrategy,
  updateConfig,
  setLoading,
  setError,
  clearDetections
} from '@/store/slices/obstacle-avoidance'
import type { AvoidanceSensitivity, AvoidanceStrategy, ObstacleSensorType } from '@/types/obstacle-avoidance'
import {
  getObstacleAvoidanceConfig,
  updateObstacleAvoidanceConfig,
  getObstacleHeatmap,
  getObstacleAvoidanceLogs,
  getObstacleAvoidanceStatistics,
  clearHeatmapData
} from '@/api/obstacle-avoidance'

export const useObstacleAvoidance = (uavId?: string) => {
  const dispatch = useAppDispatch()
  const state = useAppSelector(s => s.obstacleAvoidance)

  const fetchConfig = useCallback(async (id: string) => {
    try {
      dispatch(setLoading(true))
      const config = await getObstacleAvoidanceConfig(id)
      dispatch(updateConfig(config))
    } catch (e: any) {
      dispatch(setError(e.message))
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch])

  const saveConfig = useCallback(async (id: string) => {
    try {
      dispatch(setLoading(true))
      await updateObstacleAvoidanceConfig(id, {
        enabled: state.enabled,
        sensitivity: state.sensitivity,
        strategy: state.strategy,
        detectionRange: state.config?.detectionRange,
        ascendHeight: state.config?.ascendHeight,
        retreatDistance: state.config?.retreatDistance,
        bypassAngle: state.config?.bypassAngle
      })
    } catch (e: any) {
      dispatch(setError(e.message))
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch, state.enabled, state.sensitivity, state.strategy, state.config])

  const fetchHeatmap = useCallback(async (params?: { uavId?: string; startTime?: number; endTime?: number }) => {
    try {
      const points = await getObstacleHeatmap(params)
      dispatch({ type: 'obstacleAvoidance/updateHeatmapPoints', payload: points })
    } catch (e) {
      console.error('Failed to fetch heatmap:', e)
    }
  }, [dispatch])

  const fetchLogs = useCallback(async (params?: {
    page?: number
    pageSize?: number
    uavId?: string
    strategy?: AvoidanceStrategy
    status?: string
    startTime?: number
    endTime?: number
  }) => {
    try {
      const result = await getObstacleAvoidanceLogs(params)
      dispatch({ type: 'obstacleAvoidance/setLogs', payload: result.list })
    } catch (e) {
      console.error('Failed to fetch logs:', e)
    }
  }, [dispatch])

  const fetchStatistics = useCallback(async (params?: { uavId?: string; startTime?: number; endTime?: number }) => {
    try {
      const stats = await getObstacleAvoidanceStatistics(params)
      dispatch({ type: 'obstacleAvoidance/setStatistics', payload: stats })
    } catch (e) {
      console.error('Failed to fetch statistics:', e)
    }
  }, [dispatch])

  const handleEnabledChange = useCallback((enabled: boolean) => {
    dispatch(setEnabled(enabled))
  }, [dispatch])

  const handleSensitivityChange = useCallback((sensitivity: AvoidanceSensitivity) => {
    dispatch(setSensitivity(sensitivity))
  }, [dispatch])

  const handleStrategyChange = useCallback((strategy: AvoidanceStrategy) => {
    dispatch(setStrategy(strategy))
  }, [dispatch])

  const handleSensorTypeChange = useCallback((sensorType: ObstacleSensorType) => {
    dispatch(updateConfig({ sensorType }))
  }, [dispatch])

  const handleDetectionRangeChange = useCallback((range: number) => {
    dispatch(updateConfig({ detectionRange: range }))
  }, [dispatch])

  const handleAscendHeightChange = useCallback((height: number) => {
    dispatch(updateConfig({ ascendHeight: height }))
  }, [dispatch])

  const handleRetreatDistanceChange = useCallback((distance: number) => {
    dispatch(updateConfig({ retreatDistance: distance }))
  }, [dispatch])

  const handleBypassAngleChange = useCallback((angle: number) => {
    dispatch(updateConfig({ bypassAngle: angle }))
  }, [dispatch])

  const handleClearHeatmap = useCallback(async (id?: string) => {
    try {
      await clearHeatmapData(id)
      dispatch({ type: 'obstacleAvoidance/updateHeatmapPoints', payload: [] })
    } catch (e) {
      console.error('Failed to clear heatmap:', e)
    }
  }, [dispatch])

  useEffect(() => {
    if (uavId) {
      fetchConfig(uavId)
    }
  }, [uavId, fetchConfig])

  return {
    ...state,
    fetchConfig,
    saveConfig,
    fetchHeatmap,
    fetchLogs,
    fetchStatistics,
    handleEnabledChange,
    handleSensitivityChange,
    handleStrategyChange,
    handleSensorTypeChange,
    handleDetectionRangeChange,
    handleAscendHeightChange,
    handleRetreatDistanceChange,
    handleBypassAngleChange,
    handleClearHeatmap,
    clearDetections: () => dispatch(clearDetections())
  }
}
