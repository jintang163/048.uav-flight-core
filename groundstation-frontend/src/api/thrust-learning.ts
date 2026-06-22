import { get, post, put } from './http'
import type {
  ThrustLearningStatus,
  ThrustCurvePoint,
  PIDGainProfile,
  ThrustLearningSample
} from '@/types'

export const getThrustLearningStatus = (uavId: number): Promise<ThrustLearningStatus> => {
  return get<ThrustLearningStatus>(`/thrust-learning/status/${uavId}`)
}

export const triggerLearning = (uavId: number): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>(`/thrust-learning/trigger/${uavId}`)
}

export const optimizeModel = (uavId: number): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>(`/thrust-learning/optimize/${uavId}`)
}

export const getThrustCurve = (uavId: number): Promise<ThrustCurvePoint[]> => {
  return get<ThrustCurvePoint[]>(`/thrust-learning/curve/${uavId}`)
}

export const updateThrustCurve = (uavId: number, points: ThrustCurvePoint[]): Promise<ThrustCurvePoint[]> => {
  return put<ThrustCurvePoint[]>(`/thrust-learning/curve/${uavId}`, { points })
}

export const getPIDGains = (uavId: number): Promise<PIDGainProfile> => {
  return get<PIDGainProfile>(`/thrust-learning/pid/${uavId}`)
}

export const updatePIDGains = (uavId: number, gains: Partial<PIDGainProfile>): Promise<PIDGainProfile> => {
  return put<PIDGainProfile>(`/thrust-learning/pid/${uavId}`, gains)
}

export const applyAutoTunedPID = (uavId: number): Promise<{ success: boolean; message: string }> => {
  return post<{ success: boolean; message: string }>(`/thrust-learning/pid/apply/${uavId}`)
}

export const getLearningSamples = (uavId: number, limit?: number): Promise<ThrustLearningSample[]> => {
  return get<ThrustLearningSample[]>(`/thrust-learning/samples/${uavId}`, { limit })
}
