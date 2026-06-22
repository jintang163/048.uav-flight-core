import { useEffect, useRef } from 'react'
import { useAppSelector, useAppDispatch } from '@/store'
import { addWeatherAlert, updateWeatherData } from '@/store/slices/weather'
import { initWeatherWebSocket } from '@/websocket/weather'
import type { WeatherAlertEvent, WeatherData } from '@/types'

export const useWeatherAlert = (): void => {
  const dispatch = useAppDispatch()
  const { activeAlerts } = useAppSelector(state => state.weather)
  const initialized = useRef(false)

  useEffect(() => {
    if (!initialized.current) {
      initWeatherWebSocket()
      initialized.current = true
    }
  }, [])
}

export const useWeatherData = (uavId: number | null) => {
  const weather = useAppSelector(state =>
    uavId ? state.weather.currentWeather[uavId] : null
  )
  return weather
}

export const useWeatherAlertsForUAV = (uavId: number | null) => {
  const alerts = useAppSelector(state =>
    uavId ? (state.weather.activeAlerts[uavId] || []) : []
  )
  return alerts.filter(a => !a.is_resolved)
}
