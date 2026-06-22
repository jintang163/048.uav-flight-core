import type { Dispatch } from '@reduxjs/toolkit'
import {
  setVideoStatus,
  updateVideoStatus,
  setNetworkMetrics,
  setLinkStatus,
  updateLinkStatus,
  setCockpitMode,
  setCurrentUAV,
  setAvailableUAVs,
  setVideoDisconnectTime,
  incrementQualityAdjustmentCount,
  updateSessionStats,
  startCockpitSession,
  endCockpitSession
} from '@/store/slices/remote-cockpit'
import { addAlert } from '@/store/slices/alert'
import type {
  VideoStreamStatus,
  NetworkMetrics,
  CockpitLinkStatus,
  CockpitMode,
  CockpitSession,
  VideoStreamConfig
} from '@/types'
import { LinkType, LinkState } from '@/types/link'
import type WebSocketClient from './client'

export const setupRemoteCockpitHandlers = (wsClient: WebSocketClient, dispatch: Dispatch): void => {
  wsClient.on('video_stream_status', (data: unknown) => {
    const status = data as VideoStreamStatus
    dispatch(setVideoStatus(status))
  })

  wsClient.on('video_stream_update', (data: unknown) => {
    const update = data as Partial<VideoStreamStatus>
    dispatch(updateVideoStatus(update))
  })

  wsClient.on('video_stream_disconnected', () => {
    dispatch(setVideoDisconnectTime(Date.now()))
    dispatch(updateVideoStatus({ active: false, fps: 0, current_bitrate_kbps: 0 }))
  })

  wsClient.on('video_stream_reconnected', () => {
    dispatch(setVideoDisconnectTime(null))
    dispatch(updateVideoStatus({ active: true }))
  })

  wsClient.on('network_metrics', (data: unknown) => {
    const metrics = data as NetworkMetrics
    dispatch(setNetworkMetrics({ ...metrics, timestamp: Date.now() }))
  })

  wsClient.on('cockpit_link_status', (data: unknown) => {
    const status = data as CockpitLinkStatus
    dispatch(setLinkStatus(status))
  })

  wsClient.on('cockpit_link_update', (data: unknown) => {
    const update = data as Partial<CockpitLinkStatus>
    dispatch(updateLinkStatus(update))
  })

  wsClient.on('cockpit_link_failover', (data: unknown) => {
    const payload = data as { from_link: LinkType; to_link: LinkType; reason: string }
    dispatch(updateSessionStats({ failover_events: Date.now() }))
    dispatch(addAlert({
      id: `failover_${Date.now()}`,
      title: '链路自动切换',
      message: `从 ${payload.from_link === LinkType.LTE ? '4G网络' : '数传电台'} 切换至 ${payload.to_link === LinkType.LTE ? '4G网络' : '数传电台'}，原因：${payload.reason}`,
      severity: 'warning',
      status: 'active',
      createdAt: Date.now()
    } as any))
  })

  wsClient.on('cockpit_mode_change', (data: unknown) => {
    const payload = data as { mode: CockpitMode; uav_id?: string }
    dispatch(setCockpitMode(payload.mode))
    if (payload.uav_id) {
      dispatch(setCurrentUAV(payload.uav_id))
    }
  })

  wsClient.on('cockpit_available_uavs', (data: unknown) => {
    const payload = data as { uav_ids: string[] }
    dispatch(setAvailableUAVs(payload.uav_ids))
  })

  wsClient.on('video_quality_adjusted', (data: unknown) => {
    const payload = data as {
      old_bitrate: number
      new_bitrate: number
      old_resolution: string
      new_resolution: string
      reason: string
    }
    dispatch(incrementQualityAdjustmentCount())
    dispatch(addAlert({
      id: `quality_${Date.now()}`,
      title: '画质自适应调整',
      message: `码率: ${payload.old_bitrate}kbps → ${payload.new_bitrate}kbps, 分辨率: ${payload.old_resolution} → ${payload.new_resolution}, 原因: ${payload.reason}`,
      severity: 'info',
      status: 'active',
      createdAt: Date.now()
    } as any))
  })

  wsClient.on('cockpit_session_started', (data: unknown) => {
    const session = data as CockpitSession
    dispatch(startCockpitSession(session))
  })

  wsClient.on('cockpit_session_ended', () => {
    dispatch(endCockpitSession())
  })

  wsClient.on('auto_mission_fallback_triggered', (data: unknown) => {
    const payload = data as { uav_id: string; reason: string }
    dispatch(setCockpitMode(CockpitMode.MISSION))
    dispatch(addAlert({
      id: `fallback_${Date.now()}`,
      title: '自动切换至航线飞行',
      message: `无人机 ${payload.uav_id} 已自动切换至航线飞行模式，原因：${payload.reason}`,
      severity: 'warning',
      status: 'active',
      createdAt: Date.now()
    } as any))
  })
}

export const startCockpitSessionWS = (wsClient: WebSocketClient, uavId: string, pilotId?: string): void => {
  wsClient.send('start_cockpit_session', { uav_id: uavId, pilot_id: pilotId })
}

export const endCockpitSessionWS = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('end_cockpit_session', { uav_id: uavId })
}

export const startVideoStreamWS = (wsClient: WebSocketClient, uavId: string, config?: Partial<VideoStreamConfig>): void => {
  wsClient.send('start_video_stream', { uav_id: uavId, config })
}

export const stopVideoStreamWS = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('stop_video_stream', { uav_id: uavId })
}

export const requestVideoQualityChangeWS = (
  wsClient: WebSocketClient,
  uavId: string,
  bitrateKbps?: number,
  resolution?: string
): void => {
  wsClient.send('adjust_video_quality', {
    uav_id: uavId,
    bitrate_kbps: bitrateKbps,
    resolution
  })
}

export const setAdaptiveQualityWS = (wsClient: WebSocketClient, uavId: string, enabled: boolean): void => {
  wsClient.send('set_adaptive_quality', { uav_id: uavId, enabled })
}

export const sendFlightControlCommandWS = (
  wsClient: WebSocketClient,
  uavId: string,
  pitch: number,
  roll: number,
  yaw: number,
  throttle: number,
  source: 'keyboard' | 'gamepad' | 'autopilot' = 'gamepad'
): void => {
  wsClient.send('flight_control', {
    uav_id: uavId,
    pitch,
    roll,
    yaw,
    throttle,
    source,
    timestamp: Date.now()
  })
}

export const requestCockpitLinkStatusWS = (wsClient: WebSocketClient, uavId: string): void => {
  wsClient.send('get_cockpit_link_status', { uav_id: uavId })
}

export const requestAvailableCockpitUAVsWS = (wsClient: WebSocketClient): void => {
  wsClient.send('get_available_cockpit_uavs', {})
}

export const switchCockpitUAVWS = (wsClient: WebSocketClient, fromUavId: string, toUavId: string): void => {
  wsClient.send('switch_cockpit_uav', { from_uav_id: fromUavId, to_uav_id: toUavId })
}

export const setLinkFailoverEnabledWS = (wsClient: WebSocketClient, uavId: string, enabled: boolean): void => {
  wsClient.send('set_link_failover', { uav_id: uavId, enabled })
}

export const setAutoMissionFallbackWS = (wsClient: WebSocketClient, uavId: string, enabled: boolean): void => {
  wsClient.send('set_auto_mission_fallback', { uav_id: uavId, enabled })
}

export const setPrimaryLinkWS = (wsClient: WebSocketClient, uavId: string, linkType: LinkType): void => {
  wsClient.send('set_primary_link', { uav_id: uavId, link_type: linkType })
}
