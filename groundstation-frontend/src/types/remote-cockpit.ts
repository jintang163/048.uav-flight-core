import type { LinkType, LinkState } from './link'
import type { UAVMode } from './uav'

export enum VideoCodec {
  H264 = 'h264',
  H265 = 'h265'
}

export enum VideoResolution {
  RES_1920x1080 = '1920x1080',
  RES_1280x720 = '1280x720',
  RES_960x540 = '960x540',
  RES_640x480 = '640x480',
  RES_480x360 = '480x360'
}

export enum VideoQualityPreset {
  ULTRA_LOW = 'ultra_low',
  LOW = 'low',
  MEDIUM = 'medium',
  HIGH = 'high',
  ULTRA_HIGH = 'ultra_high'
}

export enum CockpitMode {
  IDLE = 'idle',
  CONNECTING = 'connecting',
  FLYING = 'flying',
  MISSION = 'mission',
  EMERGENCY = 'emergency',
  DISCONNECTED = 'disconnected'
}

export enum HIDDeviceType {
  GAMEPAD = 'gamepad',
  JOYSTICK = 'joystick',
  RC_TRANSMITTER = 'rc_transmitter',
  UNKNOWN = 'unknown'
}

export interface VideoStreamConfig {
  codec: VideoCodec
  resolution: VideoResolution
  bitrate_kbps: number
  fps: number
  keyframe_interval: number
  adaptive_enabled: boolean
  min_bitrate_kbps: number
  max_bitrate_kbps: number
  min_resolution: VideoResolution
  max_resolution: VideoResolution
}

export interface VideoStreamStatus {
  active: boolean
  codec: VideoCodec
  resolution: VideoResolution
  current_bitrate_kbps: number
  target_bitrate_kbps: number
  fps: number
  latency_ms: number
  jitter_ms: number
  packet_loss: number
  frames_decoded: number
  frames_dropped: number
  last_frame_time: number
}

export interface NetworkMetrics {
  bandwidth_estimate_kbps: number
  rtt_ms: number
  packet_loss: number
  jitter_ms: number
  throughput_kbps: number
  timestamp: number
}

export interface HIDAxisState {
  pitch: number
  roll: number
  yaw: number
  throttle: number
}

export interface HIDButtonState {
  arm: boolean
  disarm: boolean
  takeoff: boolean
  land: boolean
  rtl: boolean
  pause: boolean
  mode_switch: boolean
  emergency_stop: boolean
}

export interface HIDDeviceInfo {
  id: string
  name: string
  type: HIDDeviceType
  connected: boolean
  vendor_id: number
  product_id: number
  mapping_profile: string
  last_active: number
}

export interface HIDState {
  devices: HIDDeviceInfo[]
  active_device_id: string | null
  axes: HIDAxisState
  buttons: HIDButtonState
  enabled: boolean
  calibration_needed: boolean
}

export interface CockpitLinkStatus {
  primary_link: LinkType
  secondary_link: LinkType
  primary_state: LinkState
  secondary_state: LinkState
  primary_latency_ms: number
  secondary_latency_ms: number
  primary_packet_loss: number
  secondary_packet_loss: number
  failover_enabled: boolean
  failover_threshold_ms: number
  failover_count: number
  last_failover_time: number | null
}

export interface FlightControlCommand {
  uav_id: string
  pitch: number
  roll: number
  yaw: number
  throttle: number
  mode?: UAVMode
  timestamp: number
  source: 'keyboard' | 'gamepad' | 'autopilot'
}

export interface CockpitSession {
  session_id: string
  uav_id: string
  pilot_id: string
  start_time: number
  end_time: number | null
  mode: CockpitMode
  total_flight_time_ms: number
  commands_sent: number
  failover_events: number
}

export interface RemoteCockpitState {
  active: boolean
  mode: CockpitMode
  current_uav_id: string | null
  available_uav_ids: string[]
  video_config: VideoStreamConfig
  video_status: VideoStreamStatus
  network_metrics: NetworkMetrics | null
  network_metrics_history: NetworkMetrics[]
  link_status: CockpitLinkStatus | null
  hid: HIDState
  current_session: CockpitSession | null
  auto_mission_fallback: boolean
  last_video_disconnect_time: number | null
  quality_adjustment_count: number
  last_quality_adjustment_time: number | null
}

export const VideoQualityPresetConfig: Record<VideoQualityPreset, VideoStreamConfig> = {
  [VideoQualityPreset.ULTRA_LOW]: {
    codec: VideoCodec.H265,
    resolution: VideoResolution.RES_480x360,
    bitrate_kbps: 500,
    fps: 15,
    keyframe_interval: 30,
    adaptive_enabled: true,
    min_bitrate_kbps: 200,
    max_bitrate_kbps: 1000,
    min_resolution: VideoResolution.RES_480x360,
    max_resolution: VideoResolution.RES_640x480
  },
  [VideoQualityPreset.LOW]: {
    codec: VideoCodec.H265,
    resolution: VideoResolution.RES_640x480,
    bitrate_kbps: 1000,
    fps: 20,
    keyframe_interval: 40,
    adaptive_enabled: true,
    min_bitrate_kbps: 500,
    max_bitrate_kbps: 2000,
    min_resolution: VideoResolution.RES_480x360,
    max_resolution: VideoResolution.RES_960x540
  },
  [VideoQualityPreset.MEDIUM]: {
    codec: VideoCodec.H265,
    resolution: VideoResolution.RES_960x540,
    bitrate_kbps: 2500,
    fps: 25,
    keyframe_interval: 50,
    adaptive_enabled: true,
    min_bitrate_kbps: 1000,
    max_bitrate_kbps: 4000,
    min_resolution: VideoResolution.RES_640x480,
    max_resolution: VideoResolution.RES_1280x720
  },
  [VideoQualityPreset.HIGH]: {
    codec: VideoCodec.H265,
    resolution: VideoResolution.RES_1280x720,
    bitrate_kbps: 5000,
    fps: 30,
    keyframe_interval: 60,
    adaptive_enabled: true,
    min_bitrate_kbps: 2000,
    max_bitrate_kbps: 8000,
    min_resolution: VideoResolution.RES_960x540,
    max_resolution: VideoResolution.RES_1920x1080
  },
  [VideoQualityPreset.ULTRA_HIGH]: {
    codec: VideoCodec.H265,
    resolution: VideoResolution.RES_1920x1080,
    bitrate_kbps: 10000,
    fps: 60,
    keyframe_interval: 60,
    adaptive_enabled: true,
    min_bitrate_kbps: 5000,
    max_bitrate_kbps: 15000,
    min_resolution: VideoResolution.RES_1280x720,
    max_resolution: VideoResolution.RES_1920x1080
  }
}

export const ResolutionOrder: VideoResolution[] = [
  VideoResolution.RES_480x360,
  VideoResolution.RES_640x480,
  VideoResolution.RES_960x540,
  VideoResolution.RES_1280x720,
  VideoResolution.RES_1920x1080
]

export const getResolutionDimensions = (resolution: VideoResolution): { width: number; height: number } => {
  const [w, h] = resolution.split('x').map(Number)
  return { width: w, height: h }
}

export const CockpitModeText: Record<CockpitMode, string> = {
  [CockpitMode.IDLE]: '待机',
  [CockpitMode.CONNECTING]: '连接中',
  [CockpitMode.FLYING]: '飞行中',
  [CockpitMode.MISSION]: '航线飞行',
  [CockpitMode.EMERGENCY]: '紧急状态',
  [CockpitMode.DISCONNECTED]: '已断开'
}

export const VideoCodecText: Record<VideoCodec, string> = {
  [VideoCodec.H264]: 'H.264',
  [VideoCodec.H265]: 'H.265'
}

export const VideoQualityPresetText: Record<VideoQualityPreset, string> = {
  [VideoQualityPreset.ULTRA_LOW]: '超低画质',
  [VideoQualityPreset.LOW]: '低画质',
  [VideoQualityPreset.MEDIUM]: '中等画质',
  [VideoQualityPreset.HIGH]: '高画质',
  [VideoQualityPreset.ULTRA_HIGH]: '超高画质'
}

export const HIDDeviceTypeText: Record<HIDDeviceType, string> = {
  [HIDDeviceType.GAMEPAD]: '游戏手柄',
  [HIDDeviceType.JOYSTICK]: '飞行摇杆',
  [HIDDeviceType.RC_TRANSMITTER]: '遥控器',
  [HIDDeviceType.UNKNOWN]: '未知设备'
}
