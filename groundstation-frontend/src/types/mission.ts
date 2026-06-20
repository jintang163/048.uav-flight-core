export enum WaypointAction {
  WAYPOINT = 'waypoint',
  LOITER_UNLIMITED = 'loiter_unlimited',
  LOITER_TIME = 'loiter_time',
  LOITER_TURNS = 'loiter_turns',
  RETURN_TO_LAUNCH = 'rtl',
  LAND = 'land',
  TAKEOFF = 'takeoff',
  DELAY = 'delay',
  CONDITION_DELAY = 'condition_delay',
  CONDITION_DISTANCE = 'condition_distance',
  CONDITION_YAW = 'condition_yaw',
  DO_SET_MODE = 'do_set_mode',
  DO_JUMP = 'do_jump',
  DO_CHANGE_SPEED = 'do_change_speed',
  DO_SET_HOME = 'do_set_home',
  CAMERA_TRIGGER = 'camera_trigger',
  CAMERA_CONTROL = 'camera_control',
  MOUNT_CONTROL = 'mount_control'
}

export enum MissionStatus {
  DRAFT = 'draft',
  UPLOADED = 'uploaded',
  EXECUTING = 'executing',
  PAUSED = 'paused',
  COMPLETED = 'completed',
  ABORTED = 'aborted'
}

export interface Waypoint {
  id: string
  sequence: number
  action: WaypointAction
  lat: number
  lng: number
  altitude: number
  parameters: WaypointParameters
  isCurrent?: boolean
  isReached?: boolean
}

export interface WaypointParameters {
  holdTime?: number
  acceptanceRadius?: number
  passRadius?: number
  yaw?: number
  speed?: number
  turns?: number
  delay?: number
  distance?: number
  mode?: string
  jumpSequence?: number
  jumpRepeat?: number
}

export interface Mission {
  id: string
  name: string
  description?: string
  waypoints: Waypoint[]
  status: MissionStatus
  uavId?: string
  createdAt: number
  updatedAt: number
  plannedStartTime?: number
  actualStartTime?: number
  actualEndTime?: number
  distance?: number
  duration?: number
}

export interface MissionExecutionState {
  currentWaypointIndex: number
  distanceToNextWaypoint: number
  estimatedTimeToNextWaypoint: number
  totalDistance: number
  completedDistance: number
  remainingDistance: number
  estimatedTotalTime: number
  elapsedTime: number
  remainingTime: number
}

export interface MissionPlan {
  missionId: string
  waypoints: MissionWaypoint[]
  homePosition: Position
  takeoffAltitude: number
  landAltitude: number
  cruiseSpeed: number
  returnAltitude: number
}

export interface MissionWaypoint {
  seq: number
  frame: number
  command: number
  current: number
  autocontinue: number
  param1: number
  param2: number
  param3: number
  param4: number
  x: number
  y: number
  z: number
  missionType: number
}

export interface Position {
  lat: number
  lng: number
  alt: number
}

export interface TrajectoryPoint {
  lat: number
  lng: number
  alt: number
  timestamp: number
}
