import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  fetchFormationList,
  fetchFormationDetail,
  fetchActiveFormations,
  selectFormation,
  updateFormationRealtime,
  startFormationById,
  pauseFormationById,
  resumeFormationById,
  stopFormationById
} from '@/store/slices/formation'
import type { Formation, CreateFormationRequest, UpdateFormationRequest } from '@/types'

export const useFormation = (formationId?: string) => {
  const dispatch = useAppDispatch()
  const {
    formations,
    currentFormation,
    selectedFormationId,
    activeFormations,
    collisionWarnings,
    lightConfig,
    loading,
    listLoading,
    error,
    total
  } = useAppSelector(state => state.formation)

  const loadFormationList = useCallback(
    (params?: { page?: number; pageSize?: number }) => {
      dispatch(fetchFormationList(params))
    },
    [dispatch]
  )

  const loadFormationDetail = useCallback(
    (id: string) => {
      dispatch(fetchFormationDetail(id))
    },
    [dispatch]
  )

  const loadActiveFormations = useCallback(() => {
    dispatch(fetchActiveFormations())
  }, [dispatch])

  const selectCurrentFormation = useCallback(
    (id: string | null) => {
      dispatch(selectFormation(id))
    },
    [dispatch]
  )

  const updateCurrentFormation = useCallback(
    (data: Partial<Formation> & { id: string }) => {
      dispatch(updateFormationRealtime(data))
    },
    [dispatch]
  )

  const startFormation = useCallback(
    (id: string) => {
      dispatch(startFormationById(id))
    },
    [dispatch]
  )

  const pauseFormation = useCallback(
    (id: string) => {
      dispatch(pauseFormationById(id))
    },
    [dispatch]
  )

  const resumeFormation = useCallback(
    (id: string) => {
      dispatch(resumeFormationById(id))
    },
    [dispatch]
  )

  const stopFormation = useCallback(
    (id: string) => {
      dispatch(stopFormationById(id))
    },
    [dispatch]
  )

  useEffect(() => {
    if (formationId) {
      loadFormationDetail(formationId)
    }
  }, [formationId, loadFormationDetail])

  return {
    formations,
    currentFormation,
    selectedFormationId,
    activeFormations,
    collisionWarnings,
    lightConfig,
    loading,
    listLoading,
    error,
    total,
    loadFormationList,
    loadFormationDetail,
    loadActiveFormations,
    selectCurrentFormation,
    updateCurrentFormation,
    startFormation,
    pauseFormation,
    resumeFormation,
    stopFormation
  }
}
