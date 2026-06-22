export type ObstacleSensorType = 'millimeter_wave_radar' | 'stereo_vision' | 'lidar' | 'ultrasonic'

export type AvoidanceSensitivity = 'far' | 'medium' | 'near'

export type AvoidanceStrategy = 'hover' | 'ascend_bypass' | 'retreat_bypass'

export type ObstacleDirection = 'front' | 'left' | 'right' | 'top' | 'bottom' | 'rear'

export type AvoidanceActionStatus = 'detecting' | 'triggered' | 'avoiding' | 'bypassing' | 'completed' | 'failed'

export interface ObstacleDetection {
  id: string
  uavId: string
  timestamp: number
  sensorType: ObstacleSensorType
  direction: ObstacleDirection
  distance: number
  relativeAngle: number
  obstacleSize: number
  confidence: number
  position: {
    lat: number
    lng: number
    alt: number
  }
}

export interface ObstacleAvoidanceEvent {
  id: string
  uavId: string
  timestamp: number
  detection: ObstacleDetection
  strategy: AvoidanceStrategy
  status: AvoidanceActionStatus
  startPosition: {
    lat: number
    lng: number
    alt: number
  }
  bypassPath: BypassWaypoint[]
  completedAt?: number
  failReason?: string
}

export interface BypassWaypoint {
  lat: number
  lng: number
  alt: number
  timestamp: number
  type: 'detection_point' | 'bypass_start' | 'bypass_waypoint' | 'bypass_end' | 'resume_point'
}

export interface ObstacleAvoidanceConfig {
  uavId: string
  enabled: boolean
  sensitivity: AvoidanceSensitivity
  strategy: AvoidanceStrategy
  sensorType: ObstacleSensorType
  detectionRange: number
  minObstacleSize: number
  ascendHeight: number
  retreatDistance: number
  bypassAngle: number
}

export interface ObstacleHeatmapPoint {
  lat: number
  lng: number
  alt: number
  triggerCount: number
  lastTriggerTime: number
  intensity: number
  avgDistance: number
  minDistance: number
}

export interface ObstacleAvoidanceLog {
  id: string
  uavId: string
  uavName?: string
  timestamp: number
  sensorType: ObstacleSensorType
  direction: ObstacleDirection
  distance: number
  strategy: AvoidanceStrategy
  status: AvoidanceActionStatus
  duration?: number
  position: {
    lat: number
    lng: number
    alt: number
  }
  bypassPathLength?: number
  description: string
}

export interface ObstacleAvoidanceStatistics {
  totalDetections: number
  totalAvoidanceEvents: number
  successfulAvoidances: number
  failedAvoidances: number
  avgReactionTime: number
  avgAvoidanceDuration: number
  nearestObstacleDistance: number
  strategyDistribution: Record<AvoidanceStrategy, number>
  directionDistribution: Record<ObstacleDirection, number>
  sensitivitySettings: AvoidanceSensitivity
}

export const SENSITIVITY_CONFIG: Record<AvoidanceSensitivity, { detectionRange: number; label: string; reactionDistance: number }> = {
  far: { detectionRange: 15, label: '远距 (15m)', reactionDistance: 12 },
  medium: { detectionRange: 10, label: '中距 (10m)', reactionDistance: 8 },
  near: { detectionRange: 5, label: '近距 (5m)', reactionDistance: 4 }
}

export const STRATEGY_LABELS: Record<AvoidanceStrategy, string> = {
  hover: '悬停',
  ascend_bypass: '上升绕行',
  retreat_bypass: '后退绕行'
}

export const SENSOR_TYPE_LABELS: Record<ObstacleSensorType, string> = {
  millimeter_wave_radar: '毫米波雷达',
  stereo_vision: '双目视觉',
  lidar: '激光雷达',
  ultrasonic: '超声波'
}

export const DIRECTION_LABELS: Record<ObstacleDirection, string> = {
  front: '前方',
  left: '左侧',
  right: '右侧',
  top: '上方',
  bottom: '下方',
  rear: '后方'
}

export const ACTION_STATUS_LABELS: Record<AvoidanceActionStatus, string> = {
  detecting: '检测中',
  triggered: '已触发',
  avoiding: '避障中',
  bypassing: '绕行中',
  completed: '已完成',
  failed: '失败'
}

export const ACTION_STATUS_COLORS: Record<AvoidanceActionStatus, string> = {
  detecting: '#1890ff',
  triggered: '#faad14',
  avoiding: '#fa8c16',
  bypassing: '#722ed1',
  completed: '#52c41a',
  failed: '#ff4d4f'
}
