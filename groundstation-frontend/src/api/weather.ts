import { get, post, put } from './http'
import type { WeatherData, WeatherAlertEvent, WeatherThresholds, WeatherCheckResult, FlightWeatherLog, PageResult } from '@/types'

export const getLatestWeather = (uavId: number): Promise<WeatherData> => {
  return get(`/weather/uav/${uavId}/latest`)
}

export const getWeatherHistory = (uavId: number, startTime?: number, endTime?: number): Promise<WeatherData[]> => {
  const params: Record<string, unknown> = {}
  if (startTime) params.start_time = startTime
  if (endTime) params.end_time = endTime
  return get(`/weather/uav/${uavId}/history`, params)
}

export const getActiveWeatherAlerts = (uavId: number): Promise<WeatherAlertEvent[]> => {
  return get(`/weather/uav/${uavId}/alerts/active`)
}

export const checkTakeoffWeather = (uavId: number): Promise<WeatherCheckResult> => {
  return get(`/weather/uav/${uavId}/takeoff-check`)
}

export const reportWeatherSensorData = (data: {
  uav_id: number
  wind_speed: number
  wind_direction: number
  wind_gust_speed: number
  temperature: number
  humidity: number
  pressure: number
  condition: string
  latitude: number
  longitude: number
  altitude: number
}): Promise<void> => {
  return post('/weather/sensor', data)
}

export const fetchWeatherFromAPI = (lat: number, lon: number): Promise<WeatherData> => {
  return get('/weather/fetch', { lat, lon })
}

export const getWeatherAlerts = (params: {
  uav_id?: number
  alert_type?: string
  alert_level?: string
  page?: number
  page_size?: number
}): Promise<PageResult<WeatherAlertEvent>> => {
  return get('/weather/alerts', params as Record<string, unknown>)
}

export const resolveWeatherAlert = (id: number): Promise<void> => {
  return post(`/weather/alerts/${id}/resolve`)
}

export const getWeatherThresholds = (): Promise<WeatherThresholds> => {
  return get('/weather/thresholds')
}

export const updateWeatherThresholds = (thresholds: WeatherThresholds): Promise<WeatherThresholds> => {
  return put('/weather/thresholds', thresholds)
}

export const getFlightWeatherLog = (flightId: number): Promise<FlightWeatherLog> => {
  return get(`/weather/flight/${flightId}`)
}
