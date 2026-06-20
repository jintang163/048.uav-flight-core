export enum UAVStatus {
  DISCONNECTED = 'disconnected',
  CONNECTED = 'connected',
  ARMED = 'armed',
  DISARMED = 'disarmed',
  TAKEOFF = 'takeoff',
  LANDING = 'landing',
  HOVERING = 'hovering',
  FLYING = 'flying',
  RETURN_TO_HOME = 'return_to_home',
  ERROR = 'error'
}

export enum UAVMode {
  MANUAL = 'manual',
  STABILIZE = 'stabilize',
  ALT_HOLD = 'alt_hold',
  LOITER = 'loiter',
  AUTO = 'auto',
  GUIDED = 'guided',
  RTL = 'rtl',
  LAND = 'land',
  CIRCLE = 'circle'
}

export interface UAVPosition {
  lat: number
  lng: number
  alt: number
  relativeAlt: number
}

export interface UAVAttitude {
  pitch: number
  roll: number
  yaw: number
  pitchSpeed: number
  rollSpeed: number
  yawSpeed: number
}

export interface UAVVelocity {
  x: number
  y: number
  z: number
  groundSpeed: number
  airSpeed: number
  climbRate: number
}

export interface UAVBattery {
  voltage: number
  current: number
  remaining: number
  temperature: number
  cells: number[]
}

export interface UAVInfo {
  id: string
  name: string
  model: string
  serialNumber: string
  firmwareVersion: string
  hardwareVersion: string
}

export interface UAV {
  id: string
  name: string
  status: UAVStatus
  mode: UAVMode
  position: UAVPosition
  attitude: UAVAttitude
  velocity: UAVVelocity
  battery: UAVBattery
  info: UAVInfo
  signalQuality: number
  gpsFixType: number
  gpsSatellites: number
  heading: number
  throttle: number
  altitude: number
  armed: boolean
  connected: boolean
  lastUpdate: number
}

export interface UAVListItem {
  id: string
  name: string
  status: UAVStatus
  battery: number
  signal: number
  lastSeen: number
  model: string
}
