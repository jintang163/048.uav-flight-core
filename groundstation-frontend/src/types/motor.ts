export type MotorStatusType = 'normal' | 'warning' | 'fault' | 'offline'

export interface MotorStatus {
  id?: number
  uav_id: number
  motor_index: number
  status: MotorStatusType
  rpm: number
  voltage: number
  current: number
  temperature: number
  throttle: number
  fault_flags: number
  error_count: number
  vendor?: string
  model?: string
  error_code?: number
  created_at?: string
  updated_at?: string
}

export interface MotorFailureAlert {
  id: string
  uavId: number
  motorIndex: number
  faultFlags: number
  errorCode: number
  rpmAtFailure: number
  tempAtFailure: number
  actionTaken: string
  timestamp: number
  resolved: boolean
}

export interface MotorFailureState {
  uav_id: number
  failed_motors: number[]
  pid_adjusted: boolean
  rth_triggered: boolean
  start_time: string
  last_update: string
  motor_count: number
}

export interface ManualPIDRequest {
  p_gain: number
  i_gain: number
  d_gain: number
}
