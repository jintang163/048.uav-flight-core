import type { Dispatch } from '@reduxjs/toolkit'
import {
  addDetection,
  updateDetections,
  addAvoidanceEvent,
  updateAvoidanceEvent,
  completeAvoidanceEvent,
  addHeatmapPoint,
  updateHeatmapPoints,
  addLog,
  setConfig
} from '@/store/slices/obstacle-avoidance'
import type {
  ObstacleDetection,
  ObstacleAvoidanceEvent,
  ObstacleHeatmapPoint,
  ObstacleAvoidanceLog,
  ObstacleAvoidanceConfig
} from '@/types/obstacle-avoidance'
import type WebSocketClient from './client'

export const setupObstacleAvoidanceHandlers = (wsClient: WebSocketClient, dispatch: Dispatch): void => {
  wsClient.on('obstacle_detection', (data: unknown) => {
    const detections = data as { uavId: string; detections: ObstacleDetection[] }
    if (detections.detections && Array.isArray(detections.detections)) {
      dispatch(updateDetections(detections.detections))
      detections.detections.forEach(detection => {
        dispatch(addHeatmapPoint({
          lat: detection.position.lat,
          lng: detection.position.lng,
          alt: detection.position.alt,
          triggerCount: 1,
          lastTriggerTime: detection.timestamp,
          intensity: 1 / Math.max(detection.distance, 1),
          avgDistance: detection.distance,
          minDistance: detection.distance
        }))
      })
    } else {
      const detection = data as ObstacleDetection
      if (detection.id) {
        dispatch(addDetection(detection))
      }
    }
  })

  wsClient.on('obstacle_avoidance_start', (data: unknown) => {
    const event = data as ObstacleAvoidanceEvent
    if (event.id) {
      dispatch(addAvoidanceEvent(event))
      dispatch(addLog({
        id: event.id,
        uavId: event.uavId,
        timestamp: event.timestamp,
        sensorType: event.detection.sensorType,
        direction: event.detection.direction,
        distance: event.detection.distance,
        strategy: event.strategy,
        status: 'triggered',
        position: event.startPosition,
        description: `检测到${event.detection.direction}方向障碍物，距离${event.detection.distance.toFixed(1)}m，执行${event.strategy}策略`
      }))
    }
  })

  wsClient.on('obstacle_avoidance_update', (data: unknown) => {
    const update = data as { id: string; status: string; bypassPath?: unknown[] }
    if (update.id) {
      dispatch(updateAvoidanceEvent({
        id: update.id,
        status: update.status as ObstacleAvoidanceEvent['status'],
        bypassPath: update.bypassPath as ObstacleAvoidanceEvent['bypassPath']
      }))
    }
  })

  wsClient.on('obstacle_avoidance_complete', (data: unknown) => {
    const completion = data as { id: string; completedAt: number; status: string }
    if (completion.id) {
      dispatch(completeAvoidanceEvent({
        id: completion.id,
        completedAt: completion.completedAt
      }))
    }
  })

  wsClient.on('obstacle_avoidance_failed', (data: unknown) => {
    const failure = data as { id: string; reason: string }
    if (failure.id) {
      dispatch(updateAvoidanceEvent({
        id: failure.id,
        status: 'failed',
        failReason: failure.reason
      }))
    }
  })

  wsClient.on('obstacle_heatmap_update', (data: unknown) => {
    const heatmapData = data as { points: ObstacleHeatmapPoint[] }
    if (heatmapData.points && Array.isArray(heatmapData.points)) {
      dispatch(updateHeatmapPoints(heatmapData.points))
    }
  })

  wsClient.on('obstacle_avoidance_config', (data: unknown) => {
    const config = data as ObstacleAvoidanceConfig
    if (config.uavId) {
      dispatch(setConfig(config))
    }
  })

  wsClient.on('obstacle_avoidance_log', (data: unknown) => {
    const log = data as ObstacleAvoidanceLog
    if (log.id) {
      dispatch(addLog(log))
    }
  })
}

export const subscribeToObstacleAvoidance = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('subscribe_obstacle_avoidance', { uavId })
}

export const unsubscribeFromObstacleAvoidance = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('unsubscribe_obstacle_avoidance', { uavId })
}

export const setAvoidanceConfig = (wsClient: WebSocketClient, uavId: string, config: {
  enabled?: boolean
  sensitivity?: string
  strategy?: string
}): void => {
  wsClient.send('set_obstacle_avoidance_config', { uavId, ...config })
}
