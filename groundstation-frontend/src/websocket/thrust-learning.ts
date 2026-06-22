import type { Dispatch } from '@reduxjs/toolkit'
import {
  setLearningStatus,
  setThrustCurve,
  setPIDGains,
  appendSample
} from '@/store/slices/thrust-learning'
import type {
  ThrustLearningStatus,
  ThrustCurvePoint,
  PIDGainProfile,
  ThrustLearningSample
} from '@/types/thrust-learning'
import type WebSocketClient from './client'

export const setupThrustLearningHandlers = (wsClient: WebSocketClient, dispatch: Dispatch): void => {
  wsClient.on('thrust_learning_status', (data: unknown) => {
    const status = data as ThrustLearningStatus
    if (status.uav_id !== undefined) {
      dispatch(setLearningStatus(status))
    }
  })

  wsClient.on('thrust_curve_update', (data: unknown) => {
    const payload = data as { uav_id: number; points: ThrustCurvePoint[] }
    if (payload.points && Array.isArray(payload.points)) {
      dispatch(setThrustCurve(payload.points))
    }
  })

  wsClient.on('pid_gains_update', (data: unknown) => {
    const gains = data as PIDGainProfile
    if (gains.uav_id !== undefined) {
      dispatch(setPIDGains(gains))
    }
  })

  wsClient.on('thrust_learning_sample', (data: unknown) => {
    const sample = data as ThrustLearningSample
    if (sample.throttle !== undefined && sample.accel_z !== undefined) {
      dispatch(appendSample(sample))
    }
  })
}

export const subscribeToThrustLearning = (wsClient: WebSocketClient, uavId: number): void => {
  wsClient.send('subscribe_thrust_learning', { uav_id: uavId, uavId })
}

export const unsubscribeFromThrustLearning = (wsClient: WebSocketClient, uavId: number): void => {
  wsClient.send('unsubscribe_thrust_learning', { uav_id: uavId, uavId })
}

export const triggerLearningWS = (wsClient: WebSocketClient, uavId: number): void => {
  wsClient.send('trigger_thrust_learning', { uav_id: uavId, uavId })
}

export const setPIDGainsWS = (wsClient: WebSocketClient, uavId: number, gains: Partial<PIDGainProfile>): void => {
  wsClient.send('set_pid_gains', { uav_id: uavId, uavId, gains })
}
