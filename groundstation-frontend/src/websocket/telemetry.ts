import type { Dispatch } from '@reduxjs/toolkit'
import { updateTelemetry, setTelemetryConnected } from '@/store/slices/telemetry'
import { updateUAVRealtime, updateUAVStatus, updateUAVBattery, updateUAVMode } from '@/store/slices/uav'
import { updateExecutionState, setCurrentWaypoint, setWaypointReached, setMissionStatus } from '@/store/slices/mission'
import { addAlert } from '@/store/slices/alert'
import { addViolation } from '@/store/slices/geofence'
import {
  updateFormationStatus,
  updateFormationRealtime,
  addCollisionWarning,
  setLightConfig as setFormationLightConfig
} from '@/store/slices/formation'
import {
  updateDetections,
  updateActiveTask
} from '@/store/slices/tracking'
import type { TelemetryData, Alert, GeofenceViolation, UAVStatus, UAVMode, DetectionTarget, TrackingTask } from '@/types'
import type { FormationCollisionWarning } from '@/types/formation'
import type WebSocketClient from './client'

export const setupTelemetryHandlers = (wsClient: WebSocketClient, dispatch: Dispatch): void => {
  wsClient.on('telemetry', (data: unknown) => {
    const telemetryData = data as TelemetryData
    dispatch(updateTelemetry(telemetryData))
    
    dispatch(updateUAVRealtime({
      id: telemetryData.uavId,
      position: telemetryData.position,
      attitude: telemetryData.attitude,
      velocity: telemetryData.velocity,
      battery: telemetryData.battery,
      heading: telemetryData.position.alt,
      gpsFixType: telemetryData.gps.fixType,
      gpsSatellites: telemetryData.gps.satellitesVisible,
      signalQuality: telemetryData.rc.rssi,
      lastUpdate: telemetryData.timestamp
    }))
  })

  wsClient.on('uav_status', (data: unknown) => {
    const { uavId, status } = data as { uavId: string; status: UAVStatus }
    dispatch(updateUAVStatus({ id: uavId, status }))
  })

  wsClient.on('uav_mode', (data: unknown) => {
    const { uavId, mode } = data as { uavId: string; mode: UAVMode }
    dispatch(updateUAVMode({ id: uavId, mode }))
  })

  wsClient.on('battery', (data: unknown) => {
    const { uavId, remaining, voltage } = data as { uavId: string; remaining: number; voltage: number }
    dispatch(updateUAVBattery({ id: uavId, remaining, voltage }))
  })

  wsClient.on('mission_progress', (data: unknown) => {
    const { missionId, currentWaypoint, distanceToNext, estimatedTime, totalDistance, completedDistance, elapsedTime, remainingTime } = data as {
      missionId: string
      currentWaypoint: number
      distanceToNext: number
      estimatedTime: number
      totalDistance: number
      completedDistance: number
      elapsedTime: number
      remainingTime: number
    }
    
    dispatch(setCurrentWaypoint(currentWaypoint))
    dispatch(updateExecutionState({
      currentWaypointIndex: currentWaypoint,
      distanceToNextWaypoint: distanceToNext,
      estimatedTimeToNextWaypoint: estimatedTime,
      totalDistance,
      completedDistance,
      remainingDistance: totalDistance - completedDistance,
      estimatedTotalTime: elapsedTime + remainingTime,
      elapsedTime,
      remainingTime
    }))
  })

  wsClient.on('waypoint_reached', (data: unknown) => {
    const { waypointIndex } = data as { waypointIndex: number }
    dispatch(setWaypointReached(waypointIndex))
  })

  wsClient.on('mission_status', (data: unknown) => {
    const { missionId, status } = data as { missionId: string; status: string }
    dispatch(setMissionStatus({ missionId, status: status as never }))
  })

  wsClient.on('alert', (data: unknown) => {
    const alert = data as Alert
    dispatch(addAlert(alert))
    
    if (Notification.permission === 'granted') {
      new Notification(alert.title, {
        body: alert.message,
        icon: '/icon.png'
      })
    }
  })

  wsClient.on('geofence_violation', (data: unknown) => {
    const violation = data as GeofenceViolation
    dispatch(addViolation(violation))
  })

  wsClient.on('formation_update', (data: unknown) => {
    const payload = data as { formation_id: string; [key: string]: unknown }
    if (payload.formation_id) {
      dispatch(updateFormationRealtime({
        id: payload.formation_id,
        ...payload
      }))
    }
  })

  wsClient.on('formation_status', (data: unknown) => {
    const { formation_id, status } = data as { formation_id: string; status: string }
    if (formation_id && status) {
      dispatch(updateFormationStatus({ id: formation_id, status: status as any }))
    }
  })

  wsClient.on('formation_collision_warning', (data: unknown) => {
    const warning = data as FormationCollisionWarning
    dispatch(addCollisionWarning(warning))
  })

  wsClient.on('formation_light', (data: unknown) => {
    const { light } = data as { light: { red: number; green: number; blue: number; effect: string } }
    if (light) {
      dispatch(setFormationLightConfig({
        red: light.red,
        green: light.green,
        blue: light.blue,
        effect: light.effect as any
      }))
    }
  })

  wsClient.on('detections_update', (data: unknown) => {
    const { detections } = data as { detections: DetectionTarget[] }
    if (detections && Array.isArray(detections)) {
      dispatch(updateDetections(detections))
    }
  })

  wsClient.on('tracking_update', (data: unknown) => {
    const { tracking_task } = data as { tracking_task: TrackingTask }
    if (tracking_task) {
      dispatch(updateActiveTask(tracking_task))
    }
  })

  wsClient.on('connect', () => {
    dispatch(setTelemetryConnected(true))
  })

  wsClient.on('disconnect', () => {
    dispatch(setTelemetryConnected(false))
  })
}

export const subscribeToUAV = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('subscribe_uav', { uavId })
}

export const unsubscribeFromUAV = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('unsubscribe_uav', { uavId })
}

export const subscribeToAlerts = (wsClient: WebSocketClient): void => {
  wsClient.send('subscribe_alerts', {})
}

export const unsubscribeFromAlerts = (wsClient: WebSocketClient): void => {
  wsClient.send('unsubscribe_alerts', {})
}

export const requestTelemetry = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('request_telemetry', { uavId })
}

export const subscribeToFormation = (wsClient: WebSocketClient, formationId: string): void => {
  wsClient.send('subscribe_formation', { formation_id: formationId })
}

export const unsubscribeFromFormation = (wsClient: WebSocketClient, formationId: string): void => {
  wsClient.send('unsubscribe_formation', { formation_id: formationId })
}

export const subscribeToTracking = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('subscribe_uav', { uavId: uavId })
}

export const unsubscribeFromTracking = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('unsubscribe_uav', { uavId: uavId })
}
