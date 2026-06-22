export type WeatherCondition = 'clear' | 'cloudy' | 'rain' | 'snow' | 'thunderstorm' | 'fog' | 'hail'

export type WeatherAlertType = 'high_wind' | 'gust' | 'thunderstorm' | 'low_temperature' | 'heavy_rain'

export type WeatherAlertLevel = 'info' | 'warning' | 'critical'

export interface WeatherData {
  id: number
  uav_id: number
  source: string
  wind_speed: number
  wind_direction: number
  wind_gust_speed: number
  temperature: number
  humidity: number
  pressure: number
  visibility: number
  condition: WeatherCondition
  is_thunderstorm: boolean
  precipitation: number
  latitude: number
  longitude: number
  altitude: number
  created_at: string
}

export interface WeatherAlertEvent {
  id: number
  uav_id: number
  alert_type: WeatherAlertType
  alert_level: WeatherAlertLevel
  wind_speed: number
  gust_speed: number
  temperature: number
  message: string
  action_taken: string
  is_resolved: boolean
  resolved_at: string | null
  created_at: string
}

export interface WeatherThresholds {
  wind_speed_return_ms: number
  gust_protect_ms: number
  wind_adapt_ms: number
  low_temp_c: number
  thunderstorm_reject: boolean
}

export interface WeatherCheckResult {
  can_takeoff: boolean
  warnings: string[]
  block_reasons: string[]
  weather_data: WeatherData | null
}

export interface FlightWeatherLog {
  id: number
  flight_id: number
  uav_id: number
  avg_wind_speed: number
  max_wind_speed: number
  max_gust_speed: number
  avg_temperature: number
  min_temperature: number
  condition: WeatherCondition
  had_thunderstorm: boolean
  sample_count: number
  created_at: string
}
