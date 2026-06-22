import { createSlice, type PayloadAction } from '@reduxjs/toolkit'
import type {
  RemoteCockpitState,
  CockpitMode,
  VideoStreamConfig,
  VideoStreamStatus,
  NetworkMetrics,
  CockpitLinkStatus,
  HIDState,
  HIDDeviceInfo,
  HIDAxisState,
  HIDButtonState,
  CockpitSession,
  VideoQualityPreset,
  VideoQualityPresetConfig
} from '@/types'

const initialHIDState: HIDState = {
  devices: [],
  active_device_id: null,
  axes: {
    pitch: 0,
    roll: 0,
    yaw: 0,
    throttle: 0
  },
  buttons: {
    arm: false,
    disarm: false,
    takeoff: false,
    land: false,
    rtl: false,
    pause: false,
    mode_switch: false,
    emergency_stop: false
  },
  enabled: true,
  calibration_needed: false
}

const initialVideoConfig: VideoStreamConfig = VideoQualityPresetConfig[VideoQualityPreset.MEDIUM]

const initialVideoStatus: VideoStreamStatus = {
  active: false,
  codec: initialVideoConfig.codec,
  resolution: initialVideoConfig.resolution,
  current_bitrate_kbps: 0,
  target_bitrate_kbps: initialVideoConfig.bitrate_kbps,
  fps: 0,
  latency_ms: 0,
  jitter_ms: 0,
  packet_loss: 0,
  frames_decoded: 0,
  frames_dropped: 0,
  last_frame_time: 0
}

const initialState: RemoteCockpitState = {
  active: false,
  mode: CockpitMode.IDLE,
  current_uav_id: null,
  available_uav_ids: [],
  video_config: initialVideoConfig,
  video_status: initialVideoStatus,
  network_metrics: null,
  network_metrics_history: [],
  link_status: null,
  hid: initialHIDState,
  current_session: null,
  auto_mission_fallback: true,
  last_video_disconnect_time: null,
  quality_adjustment_count: 0,
  last_quality_adjustment_time: null
}

const MAX_NETWORK_METRICS_HISTORY = 60

const remoteCockpitSlice = createSlice({
  name: 'remoteCockpit',
  initialState,
  reducers: {
    setCockpitActive: (state, action: PayloadAction<boolean>) => {
      state.active = action.payload
      if (!action.payload) {
        state.mode = CockpitMode.IDLE
      }
    },
    setCockpitMode: (state, action: PayloadAction<CockpitMode>) => {
      state.mode = action.payload
    },
    setCurrentUAV: (state, action: PayloadAction<string | null>) => {
      state.current_uav_id = action.payload
    },
    setAvailableUAVs: (state, action: PayloadAction<string[]>) => {
      state.available_uav_ids = action.payload
    },
    setVideoConfig: (state, action: PayloadAction<VideoStreamConfig>) => {
      state.video_config = action.payload
    },
    updateVideoConfig: (state, action: PayloadAction<Partial<VideoStreamConfig>>) => {
      state.video_config = { ...state.video_config, ...action.payload }
    },
    setVideoStatus: (state, action: PayloadAction<VideoStreamStatus>) => {
      state.video_status = action.payload
    },
    updateVideoStatus: (state, action: PayloadAction<Partial<VideoStreamStatus>>) => {
      state.video_status = { ...state.video_status, ...action.payload }
    },
    setNetworkMetrics: (state, action: PayloadAction<NetworkMetrics>) => {
      state.network_metrics = action.payload
      state.network_metrics_history.push(action.payload)
      if (state.network_metrics_history.length > MAX_NETWORK_METRICS_HISTORY) {
        state.network_metrics_history = state.network_metrics_history.slice(-MAX_NETWORK_METRICS_HISTORY)
      }
    },
    clearNetworkMetrics: (state) => {
      state.network_metrics = null
      state.network_metrics_history = []
    },
    setLinkStatus: (state, action: PayloadAction<CockpitLinkStatus>) => {
      state.link_status = action.payload
    },
    updateLinkStatus: (state, action: PayloadAction<Partial<CockpitLinkStatus>>) => {
      if (state.link_status) {
        state.link_status = { ...state.link_status, ...action.payload }
      }
    },
    setHIDEnabled: (state, action: PayloadAction<boolean>) => {
      state.hid.enabled = action.payload
    },
    setHIDDevices: (state, action: PayloadAction<HIDDeviceInfo[]>) => {
      state.hid.devices = action.payload
    },
    addHIDDevice: (state, action: PayloadAction<HIDDeviceInfo>) => {
      const existing = state.hid.devices.find(d => d.id === action.payload.id)
      if (!existing) {
        state.hid.devices.push(action.payload)
      } else {
        const idx = state.hid.devices.findIndex(d => d.id === action.payload.id)
        state.hid.devices[idx] = action.payload
      }
    },
    removeHIDDevice: (state, action: PayloadAction<string>) => {
      state.hid.devices = state.hid.devices.filter(d => d.id !== action.payload)
      if (state.hid.active_device_id === action.payload) {
        state.hid.active_device_id = null
      }
    },
    setActiveHIDDevice: (state, action: PayloadAction<string | null>) => {
      state.hid.active_device_id = action.payload
    },
    setHIDAxes: (state, action: PayloadAction<HIDAxisState>) => {
      state.hid.axes = action.payload
    },
    setHIDButtons: (state, action: PayloadAction<HIDButtonState>) => {
      state.hid.buttons = action.payload
    },
    setHIDCalibrationNeeded: (state, action: PayloadAction<boolean>) => {
      state.hid.calibration_needed = action.payload
    },
    startCockpitSession: (state, action: PayloadAction<CockpitSession>) => {
      state.current_session = action.payload
      state.active = true
    },
    endCockpitSession: (state) => {
      if (state.current_session) {
        state.current_session.end_time = Date.now()
      }
      state.active = false
      state.mode = CockpitMode.IDLE
    },
    updateSessionStats: (state, action: PayloadAction<{ commands_sent?: number; failover_events?: number; total_flight_time_ms?: number }>) => {
      if (state.current_session) {
        if (action.payload.commands_sent !== undefined) {
          state.current_session.commands_sent = action.payload.commands_sent
        }
        if (action.payload.failover_events !== undefined) {
          state.current_session.failover_events = action.payload.failover_events
        }
        if (action.payload.total_flight_time_ms !== undefined) {
          state.current_session.total_flight_time_ms = action.payload.total_flight_time_ms
        }
      }
    },
    setAutoMissionFallback: (state, action: PayloadAction<boolean>) => {
      state.auto_mission_fallback = action.payload
    },
    setVideoDisconnectTime: (state, action: PayloadAction<number | null>) => {
      state.last_video_disconnect_time = action.payload
    },
    incrementQualityAdjustmentCount: (state) => {
      state.quality_adjustment_count += 1
      state.last_quality_adjustment_time = Date.now()
    },
    applyVideoQualityPreset: (state, action: PayloadAction<VideoQualityPreset>) => {
      state.video_config = VideoQualityPresetConfig[action.payload]
    },
    resetRemoteCockpit: () => initialState
  }
})

export const {
  setCockpitActive,
  setCockpitMode,
  setCurrentUAV,
  setAvailableUAVs,
  setVideoConfig,
  updateVideoConfig,
  setVideoStatus,
  updateVideoStatus,
  setNetworkMetrics,
  clearNetworkMetrics,
  setLinkStatus,
  updateLinkStatus,
  setHIDEnabled,
  setHIDDevices,
  addHIDDevice,
  removeHIDDevice,
  setActiveHIDDevice,
  setHIDAxes,
  setHIDButtons,
  setHIDCalibrationNeeded,
  startCockpitSession,
  endCockpitSession,
  updateSessionStats,
  setAutoMissionFallback,
  setVideoDisconnectTime,
  incrementQualityAdjustmentCount,
  applyVideoQualityPreset,
  resetRemoteCockpit
} = remoteCockpitSlice.actions

export const selectRemoteCockpitState = (state: { remoteCockpit: RemoteCockpitState }): RemoteCockpitState =>
  state.remoteCockpit

export const selectIsCockpitActive = (state: { remoteCockpit: RemoteCockpitState }): boolean =>
  state.remoteCockpit.active

export const selectCockpitMode = (state: { remoteCockpit: RemoteCockpitState }): CockpitMode =>
  state.remoteCockpit.mode

export const selectCurrentUAVId = (state: { remoteCockpit: RemoteCockpitState }): string | null =>
  state.remoteCockpit.current_uav_id

export const selectVideoConfig = (state: { remoteCockpit: RemoteCockpitState }): VideoStreamConfig =>
  state.remoteCockpit.video_config

export const selectVideoStatus = (state: { remoteCockpit: RemoteCockpitState }): VideoStreamStatus =>
  state.remoteCockpit.video_status

export const selectNetworkMetrics = (state: { remoteCockpit: RemoteCockpitState }): NetworkMetrics | null =>
  state.remoteCockpit.network_metrics

export const selectLinkStatus = (state: { remoteCockpit: RemoteCockpitState }): CockpitLinkStatus | null =>
  state.remoteCockpit.link_status

export const selectHIDState = (state: { remoteCockpit: RemoteCockpitState }): HIDState =>
  state.remoteCockpit.hid

export const selectCockpitSession = (state: { remoteCockpit: RemoteCockpitState }): CockpitSession | null =>
  state.remoteCockpit.current_session

export default remoteCockpitSlice.reducer
