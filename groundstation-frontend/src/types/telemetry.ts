import type { UAVAttitude, UAVPosition, UAVVelocity, UAVBattery } from './uav'

export interface TelemetryData {
  timestamp: number
  uavId: string
  position: UAVPosition
  attitude: UAVAttitude
  velocity: UAVVelocity
  battery: UAVBattery
  gps: GPSData
  system: SystemStatus
  rc: RCChannels
  motors: MotorData[]
  vibration: VibrationData
  wind: WindData
}

export interface GPSData {
  fixType: number
  satellitesVisible: number
  satellitesUsed: number
  hdop: number
  vdop: number
  pdop: number
  eph: number
  epv: number
}

export interface RCChannels {
  channel1: number
  channel2: number
  channel3: number
  channel4: number
  channel5: number
  channel6: number
  channel7: number
  channel8: number
  channel9?: number
  channel10?: number
  channel11?: number
  channel12?: number
  channel13?: number
  channel14?: number
  channel15?: number
  channel16?: number
  rssi: number
}

export interface MotorData {
  id: number
  rpm: number
  temperature: number
  voltage: number
  current: number
}

export interface VibrationData {
  x: number
  y: number
  z: number
  clipping0: number
  clipping1: number
  clipping2: number
}

export interface WindData {
  direction: number
  speed: number
  speedZ: number
  temperature: number
}

export interface SystemStatus {
  load: number
  voltageBattery: number
  currentBattery: number
  batteryRemaining: number
  dropRateComm: number
  errorsCount1: number
  errorsCount2: number
  errorsCount3: number
  errorsCount4: number
}

export interface TelemetryHistoryPoint {
  timestamp: number
  altitude: number
  speed: number
  throttle: number
  battery: number
}

export interface TelemetryHistory {
  uavId: string
  startTime: number
  endTime: number
  points: TelemetryHistoryPoint[]
}
