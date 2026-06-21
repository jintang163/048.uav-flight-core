import request from '@/utils/request'
import type { MotorStatus, MotorFailureState, ManualPIDRequest } from '@/types'

const BASE = '/motor-protection'

export const getMotorStatuses = (uavId: string | number): Promise<MotorStatus[]> => {
  return request.get(`${BASE}/uav/${uavId}/status`)
}

export const getMotorFailureState = (uavId: string | number): Promise<MotorFailureState | null> => {
  return request.get(`${BASE}/uav/${uavId}/failure-state`)
}

export const manualPIDAdjustment = (uavId: string | number, data: ManualPIDRequest): Promise<void> => {
  return request.post(`${BASE}/uav/${uavId}/pid-adjust`, data)
}

export const emergencyRTH = (uavId: string | number): Promise<void> => {
  return request.post(`${BASE}/uav/${uavId}/emergency-rth`)
}

export const emergencyLand = (uavId: string | number): Promise<void> => {
  return request.post(`${BASE}/uav/${uavId}/emergency-land`)
}

export const resolveMotorFailure = (uavId: string | number, motorIndex: number): Promise<void> => {
  return request.post(`${BASE}/uav/${uavId}/resolve/${motorIndex}`)
}
