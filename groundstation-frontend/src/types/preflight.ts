export type PreflightCheckType =
  | 'gps'
  | 'battery'
  | 'imu'
  | 'storage'
  | 'link'
  | 'compass'
  | 'barometer'
  | 'arm'

export type PreflightCheckStatus = 'pass' | 'warning' | 'fail' | 'pending'

export interface PreflightCheckItem {
  check_type: PreflightCheckType
  name: string
  description: string
  status: PreflightCheckStatus
  threshold: string
  actual_value: string
  message: string
  detail: Record<string, any>
  checked_at: string
}

export interface PreflightCheckResult {
  uav_id: number
  uav_name: string
  overall_status: PreflightCheckStatus
  can_takeoff: boolean
  passed_count: number
  warning_count: number
  failed_count: number
  total_count: number
  checks: PreflightCheckItem[]
  started_at: string
  finished_at: string
  summary: string
}

export interface PreflightCheckThresholds {
  min_satellites: number
  min_gps_fix_type: number
  max_hdop: number
  min_voltage: number
  min_voltage_per_cell: number
  cell_count: number
  max_accel_offset: number
  max_gyro_offset: number
  min_storage_space_mb: number
  min_signal_strength: number
  min_link_quality: number
}

export interface RunPreflightRequest {
  uav_id: number
  min_satellites?: number
  min_gps_fix_type?: number
  max_hdop?: number
  min_voltage?: number
  min_voltage_per_cell?: number
  cell_count?: number
  min_storage_space_mb?: number
  min_signal_strength?: number
  min_link_quality?: number
}

export interface BatchRunPreflightRequest {
  uav_ids: number[]
}

export interface BatchPreflightResponse {
  results: PreflightCheckResult[]
  errors: Array<{ uav_id: number; error: string }>
  total: number
  success: number
  failed: number
}
