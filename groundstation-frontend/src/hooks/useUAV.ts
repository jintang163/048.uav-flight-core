import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import { fetchUAVList, fetchUAVDetail, selectUAV, updateUAVRealtime } from '@/store/slices/uav'
import type { UAV } from '@/types'

export const useUAV = (uavId?: string) => {
  const dispatch = useAppDispatch()
  const { currentUAV, uavList, selectedUAVId, loading, listLoading, error, total } = useAppSelector(state => state.uav)

  const loadUAVList = useCallback((params?: { page?: number; pageSize?: number; keyword?: string; status?: string }) => {
    dispatch(fetchUAVList(params))
  }, [dispatch])

  const loadUAVDetail = useCallback((id: string) => {
    dispatch(fetchUAVDetail(id))
  }, [dispatch])

  const selectCurrentUAV = useCallback((id: string | null) => {
    dispatch(selectUAV(id))
  }, [dispatch])

  const updateCurrentUAV = useCallback((data: Partial<UAV> & { id: string }) => {
    dispatch(updateUAVRealtime(data))
  }, [dispatch])

  useEffect(() => {
    if (uavId) {
      loadUAVDetail(uavId)
    }
  }, [uavId, loadUAVDetail])

  return {
    currentUAV,
    uavList,
    selectedUAVId,
    loading,
    listLoading,
    error,
    total,
    loadUAVList,
    loadUAVDetail,
    selectCurrentUAV,
    updateCurrentUAV
  }
}
