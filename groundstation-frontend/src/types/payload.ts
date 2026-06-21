export type PayloadType =
  | 'camera'
  | 'thermal_camera'
  | 'sprayer'
  | 'speaker'
  | 'gripper'
  | 'sensor'
  | 'lidar'
  | 'parachute'
  | 'other'

export type PayloadStatus = 'offline' | 'online' | 'active' | 'error' | string

export type CameraMode = 'idle' | 'photo' | 'video' | 'playback'

export interface CameraStatus {
  id?: string
  payloadId?: string
  mode?: CameraMode
  recording?: boolean
  isRecording?: boolean
  recordingTimeSec?: number
  recordTimeSec?: number
  photoCount?: number
  storageFreeMB?: number
  storageTotalMB?: number
  storageUsedPercent?: number
  storagePercent?: number
  lensTemperatureC?: number
  lensTemperature?: number
  lensTempC?: number
  sensorTemperatureC?: number
  zoomLevel?: number
  focusLevel?: number
  iso?: number
  shutterSpeedMs?: number
  shutterSpeed?: string
  resolution?: string
  frameRate?: number
  batteryPercent?: number
  lastCaptureAt?: number
}

export interface SprayerStatus {
  id?: string
  payloadId?: string
  flowRate?: number
  flowRateLpm?: number
  targetFlowRate?: number
  remainingVolume?: number
  remainingPercent?: number
  tankLevelPercent?: number
  totalCapacity?: number
  tankCapacityL?: number
  tankRemainingL?: number
  pressure?: number
  pressureBar?: number
  spraying?: boolean
  isSpraying?: boolean
  nozzleCount?: number
  nozzleActive?: number
  sprayWidthM?: number
  totalSprayedL?: number
}

export interface SpeakerAudio {
  id: string
  uuid?: string
  payloadId?: string
  name?: string
  type?: 'pre_recorded' | 'tts'
  format?: string
  content?: string
  filePath?: string
  fileURL?: string
  fileUrl?: string
  fileSizeKB?: number
  fileSizeBytes?: number
  durationSec?: number
  audioDurationSec?: number
  isTextToSpeech?: boolean
  voice?: string
  speed?: number
  pitch?: number
  volume?: number
  createdBy?: string
  createdAt?: number | string
}

export interface PayloadDevice {
  id: string
  uuid?: string
  uavId: string
  uavID?: string
  name: string
  type: PayloadType
  model?: string
  description?: string
  status: PayloadStatus
  port?: number
  config?: string
  temperature?: number
  battery?: number
  health?: number
  lastActiveAt?: number
  createdAt?: number
  cameraStatus?: CameraStatus
  sprayerStatus?: SprayerStatus
}

export type OrbitStatus =
  | 'pending'
  | 'running'
  | 'paused'
  | 'completed'
  | 'aborted'
  | 'failed'
  | string

export interface OrbitMission {
  id: string
  uuid?: string
  uavId: string
  uavID?: string
  name: string
  description?: string
  centerLat: number
  centerLng: number
  centerAlt: number
  altitude: number
  radius: number
  loops: number
  currentLoop?: number
  direction: 1 | -1
  velocity: number
  gimbalPitch?: number
  autoCapture: boolean
  captureInterval: number
  status: OrbitStatus
  progress?: number
  estimatedDurationSec?: number
  elapsedTimeSec?: number
  photosCaptured?: number
  creatorId?: string
  createdAt?: number
  startedAt?: number
  completedAt?: number
}

export type OrthoStatus =
  | 'pending'
  | 'planning'
  | 'planned'
  | 'running'
  | 'paused'
  | 'completed'
  | 'aborted'
  | 'failed'
  | string

export interface OrthoWaypoint {
  id: string
  missionId: string
  sequence: number
  lat: number
  lng: number
  alt: number
  speed?: number
  capturePoint?: boolean
  heading?: number
}

export interface OrthoMission {
  id: string
  uuid?: string
  uavId: string
  uavID?: string
  name: string
  description?: string
  areaCoordinates: { lat: number; lng: number }[]
  altitude: number
  speed: number
  overlapForward: number
  overlapSide: number
  cameraFocalLength?: number
  sensorWidth?: number
  sensorHeight?: number
  imageWidth?: number
  imageHeight?: number
  gsD?: number
  totalAreaKm2?: number
  estimatedDurationSec?: number
  estimatedPhotos?: number
  totalDistanceKm?: number
  waypointsCount?: number
  currentWaypointIndex?: number
  photosCaptured?: number
  status: OrthoStatus
  progress?: number
  waypoints?: OrthoWaypoint[]
  creatorId?: string
  createdAt?: number
  startedAt?: number
  completedAt?: number
}

export type TTSTaskStatus = 'pending' | 'processing' | 'completed' | 'failed' | 'playing' | string

export interface TextToSpeechTask {
  id: string
  uuid?: string
  uavId?: string
  uavID?: string
  speakerPayloadId?: string
  payloadId?: string
  text: string
  language?: string
  voice?: string
  speed?: number
  pitch?: number
  volume?: number
  status: TTSTaskStatus
  autoPlay?: boolean
  audio?: SpeakerAudio
  audioURL?: string
  audioUrl?: string
  durationSec?: number
  audioDurationSec?: number
  errorMessage?: string
  createdBy?: string
  createdAt?: number | string
  completedAt?: number | string
}

export interface PayloadTelemetry {
  uavId: string
  payloadId: string
  payloadType: PayloadType
  timestamp: number
  data: Record<string, any>
}

export interface OrbitCreateRequest {
  uavId: string
  name?: string
  centerLat: number
  centerLng: number
  altitude: number
  radius: number
  loops: number
  direction: 1 | -1
  velocity: number
  autoCapture?: boolean
  captureInterval?: number
  gimbalPitch?: number
  payloadId?: string
}

export interface OrthoPlanRequest {
  uavId: string
  name?: string
  areaCoordinates: { lat: number; lng: number }[]
  altitude: number
  overlapForward: number
  overlapSide: number
  speed?: number
  cameraFocalLength?: number
  sensorWidth?: number
  sensorHeight?: number
  imageWidth?: number
  imageHeight?: number
}

export interface TTSCreateRequest {
  uavId: string
  speakerPayloadId?: string
  payloadId?: string
  text: string
  language?: string
  voice?: string
  speed?: number
  pitch?: number
  volume?: number
  autoPlay?: boolean
}
