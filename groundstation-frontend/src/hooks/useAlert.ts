import { useEffect, useCallback } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import { fetchAlerts, fetchAlertStats, acknowledgeAlertById, resolveAlertById } from '@/store/slices/alert'
import { showNotification, playAlertSound, speak } from '@/utils'
import type { Alert, AlertFilter, AlertSeverity } from '@/types'

export const useAlert = (autoFetch = true) => {
  const dispatch = useAppDispatch()
  const { alerts, unreadCount, stats, loading, statsLoading, total, currentPage, pageSize, filter } = useAppSelector(state => state.alert)

  const loadAlerts = useCallback((params?: AlertFilter & { page?: number; pageSize?: number }) => {
    dispatch(fetchAlerts(params || {}))
  }, [dispatch])

  const loadStats = useCallback((params?: { startTime?: number; endTime?: number; uavId?: string }) => {
    dispatch(fetchAlertStats(params || {}))
  }, [dispatch])

  const acknowledge = useCallback((id: string) => {
    dispatch(acknowledgeAlertById(id))
  }, [dispatch])

  const resolve = useCallback((id: string, notes?: string) => {
    dispatch(resolveAlertById({ id, notes }))
  }, [dispatch])

  const handleNewAlert = useCallback((alert: Alert) => {
    if (alert.severity === 'critical' || alert.severity === 'error') {
      playAlertSound()
      speak(alert.title)
    }
    showNotification(alert.title, alert.message)
  }, [])

  useEffect(() => {
    if (autoFetch) {
      loadAlerts()
      loadStats()
    }
  }, [autoFetch, loadAlerts, loadStats])

  const getAlertsBySeverity = (severity: AlertSeverity) => {
    return alerts.filter(alert => alert.severity === severity)
  }

  const getActiveAlerts = () => {
    return alerts.filter(alert => alert.status === 'active')
  }

  return {
    alerts,
    unreadCount,
    stats,
    loading,
    statsLoading,
    total,
    currentPage,
    pageSize,
    filter,
    loadAlerts,
    loadStats,
    acknowledge,
    resolve,
    handleNewAlert,
    getAlertsBySeverity,
    getActiveAlerts
  }
}
