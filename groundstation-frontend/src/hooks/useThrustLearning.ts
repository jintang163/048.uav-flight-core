import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  setLearningStatus,
  setThrustCurve,
  setPIDGains,
  setSamples,
  setLoading,
  setError
} from '@/store/slices/thrust-learning'
import type { PIDGainProfile, ThrustLearningStatus, ThrustCurvePoint } from '@/types/thrust-learning'
import {
  getThrustLearningStatus,
  triggerLearning,
  optimizeModel,
  getThrustCurve,
  updateThrustCurve,
  getPIDGains,
  updatePIDGains,
  applyAutoTunedPID,
  getLearningSamples
} from '@/api/thrust-learning'

export const useThrustLearning = (uavId?: number) => {
  const dispatch = useAppDispatch()
  const state = useAppSelector(s => s.thrustLearning)

  const fetchStatus = useCallback(async (id: number) => {
    try {
      dispatch(setLoading(true))
      const status = await getThrustLearningStatus(id)
      dispatch(setLearningStatus(status))
    } catch (e: any) {
      dispatch(setError(e.message))
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch])

  const fetchCurve = useCallback(async (id: number) => {
    try {
      const curve = await getThrustCurve(id)
      dispatch(setThrustCurve(curve))
    } catch (e) {
      console.error('Failed to fetch thrust curve:', e)
    }
  }, [dispatch])

  const fetchPIDGains = useCallback(async (id: number) => {
    try {
      const gains = await getPIDGains(id)
      dispatch(setPIDGains(gains))
    } catch (e) {
      console.error('Failed to fetch PID gains:', e)
    }
  }, [dispatch])

  const fetchSamples = useCallback(async (id: number, limit?: number) => {
    try {
      const samples = await getLearningSamples(id, limit)
      dispatch(setSamples(samples))
    } catch (e) {
      console.error('Failed to fetch samples:', e)
    }
  }, [dispatch])

  const fetchAll = useCallback(async (id: number) => {
    await Promise.all([
      fetchStatus(id),
      fetchCurve(id),
      fetchPIDGains(id),
      fetchSamples(id, 200)
    ])
  }, [fetchStatus, fetchCurve, fetchPIDGains, fetchSamples])

  const startLearning = useCallback(async (id: number) => {
    try {
      dispatch(setLoading(true))
      const result = await triggerLearning(id)
      if (result.success) {
        await fetchStatus(id)
      }
      return result
    } catch (e: any) {
      dispatch(setError(e.message))
      return { success: false, message: e.message }
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch, fetchStatus])

  const startOptimize = useCallback(async (id: number) => {
    try {
      dispatch(setLoading(true))
      const result = await optimizeModel(id)
      if (result.success) {
        await fetchStatus(id)
        await fetchCurve(id)
        await fetchPIDGains(id)
      }
      return result
    } catch (e: any) {
      dispatch(setError(e.message))
      return { success: false, message: e.message }
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch, fetchStatus, fetchCurve, fetchPIDGains])

  const saveThrustCurve = useCallback(async (id: number, points: ThrustCurvePoint[]) => {
    try {
      await updateThrustCurve(id, points)
      dispatch(setThrustCurve(points))
    } catch (e) {
      console.error('Failed to save thrust curve:', e)
    }
  }, [dispatch])

  const savePIDGains = useCallback(async (id: number, gains: Partial<PIDGainProfile>) => {
    try {
      dispatch(setLoading(true))
      const updated = await updatePIDGains(id, gains)
      dispatch(setPIDGains(updated))
      return { success: true, message: 'PID参数已保存' }
    } catch (e: any) {
      dispatch(setError(e.message))
      return { success: false, message: e.message }
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch])

  const handlePIDGainsChange = useCallback((gains: Partial<PIDGainProfile>) => {
    if (state.pidGains) {
      dispatch(setPIDGains({ ...state.pidGains, ...gains }))
    }
  }, [dispatch, state.pidGains])

  const applyPID = useCallback(async (id: number) => {
    try {
      dispatch(setLoading(true))
      const result = await applyAutoTunedPID(id)
      return result
    } catch (e: any) {
      dispatch(setError(e.message))
      return { success: false, message: e.message }
    } finally {
      dispatch(setLoading(false))
    }
  }, [dispatch])

  useEffect(() => {
    if (uavId !== undefined) {
      fetchAll(uavId)
    }
  }, [uavId, fetchAll])

  return {
    ...state,
    fetchStatus,
    fetchCurve,
    fetchPIDGains,
    fetchSamples,
    fetchAll,
    startLearning,
    startOptimize,
    saveThrustCurve,
    savePIDGains,
    applyPID,
    handlePIDGainsChange
  }
}
