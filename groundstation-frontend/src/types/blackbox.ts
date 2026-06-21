export enum BlackboxLogStatus {
  UPLOADING = 'uploading',
  UPLOADED = 'uploaded',
  ANALYZED = 'analyzed',
  ERROR = 'error'
}

export interface BlackboxLog {
  id: string
  uuid: string
  uav_id: string
  mission_id?: string
  flight_name: string
  start_time?: string
  end_time?: string
  duration: number
  file_size: number
  file_name: string
  file_url: string
  log_type: string
  status: BlackboxLogStatus
  file_hash: string
  uploader_id: string
  max_altitude: number
  max_speed: number
  distance: number
  battery_used: number
  crash_detected: boolean
  tags?: string
  notes?: string
  created_at: string
  updated_at: string
  uav?: {
    id: string
    name: string
  }
  mission?: {
    id: string
    name: string
  }
}

export interface LogDataPoint {
  timestamp: number
  latitude: number
  longitude: number
  altitude: number
  roll: number
  pitch: number
  yaw: number
  vx: number
  vy: number
  vz: number
  voltage: number
  current: number
  throttle: number
  flight_mode: number
  satellites: number
  gps_fix_type: number
  error_flags: number
  motor_pwm: number[]
  rc_channels: number[]
}

export interface LogEvent {
  timestamp: number
  event_type: string
  event_type_id: number
  severity: number
  description: string
  param1: number
  param2: number
  param3: number
  param4: number
}

export interface LogStatistics {
  total_duration: number
  max_altitude: number
  max_speed: number
  total_distance: number
  avg_voltage: number
  min_voltage: number
  battery_used: number
  anomaly_count: number
  max_roll: number
  max_pitch: number
  avg_satellites: number
  crash_detected: boolean
}

export interface ParsedLogData {
  header: Record<string, unknown>
  data_points: LogDataPoint[]
  events: LogEvent[]
  statistics: LogStatistics
}

export interface FlightPhase {
  phase_name: string
  start_time: number
  end_time: number
  duration: number
}

export interface AnalysisReport {
  log_id: string
  flight_summary: string
  flight_score: number
  anomalies: LogEvent[]
  statistics: LogStatistics
  recommendations: string[]
  flight_phases: FlightPhase[]
}

export interface LogAnalysisReport {
  id: string
  log_id: string
  analyzer_id?: string
  report_type: string
  summary: string
  flight_score: number
  anomalies: string
  recommendations: string
  report_data: string
  report_url?: string
  created_at: string
  updated_at: string
}

export interface UploadLogRequest {
  uav_id: string
  mission_id?: string
  log_type?: string
  start_time?: string
  end_time?: string
  flight_name?: string
  notes?: string
  file: File
}
