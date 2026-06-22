import { getWebSocketClient } from '../client'
import { updateWeatherData, addWeatherAlert } from '@/store/slices/weather'
import type { WeatherData, WeatherAlertEvent } from '@/types'
import store from '@/store'

export const initWeatherWebSocket = (): void => {
  const wsClient = getWebSocketClient()
  if (!wsClient) return

  wsClient.on('weather_data', (data: unknown) => {
    const payload = data as WeatherData & { uav_id: number }
    if (payload.uav_id) {
      store.dispatch(updateWeatherData(payload as WeatherData))
    }
  })

  wsClient.on('weather_alert', (data: unknown) => {
    const payload = data as {
      alert_id: number
      uav_id: number
      alert_type: string
      alert_level: string
      wind_speed: number
      gust_speed: number
      temperature: number
      message: string
      action_taken: string
      timestamp: number
    }
    if (payload.uav_id) {
      const alert: WeatherAlertEvent = {
        id: payload.alert_id,
        uav_id: payload.uav_id,
        alert_type: payload.alert_type as WeatherAlertEvent['alert_type'],
        alert_level: payload.alert_level as WeatherAlertEvent['alert_level'],
        wind_speed: payload.wind_speed,
        gust_speed: payload.gust_speed,
        temperature: payload.temperature,
        message: payload.message,
        action_taken: payload.action_taken,
        is_resolved: false,
        resolved_at: null,
        created_at: new Date(payload.timestamp).toISOString(),
      }
      store.dispatch(addWeatherAlert({ uavId: payload.uav_id, alert }))
    }
  })
}
