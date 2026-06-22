import { get, post } from './http'
import type {
  CockpitSession,
  VideoStreamConfig,
  VideoStreamStatus,
  NetworkMetrics,
  CockpitLinkStatus,
  FlightControlCommand
} from '@/types'
import type { LinkType } from '@/types/link'

export interface CockpitSessionListResponse {
  list: CockpitSession[]
  total: number
  page: number
  pageSize: number
}

export const startCockpitSession = (uavId: string, pilotId?: string): Promise<CockpitSession> => {
  return post<CockpitSession>('/remote-cockpit/session/start', { uav_id: uavId, pilot_id: pilotId })
}

export const endCockpitSession = (uavId: string): Promise<CockpitSession> => {
  return post<CockpitSession>('/remote-cockpit/session/end', { uav_id: uavId })
}

export const getCockpitSession = (sessionId: string): Promise<CockpitSession> => {
  return get<CockpitSession>(`/remote-cockpit/session/${sessionId}`)
}

export const getCockpitSessionList = (params?: {
  page?: number
  pageSize?: number
  uavId?: string
  pilotId?: string
  startTime?: number
  endTime?: number
}): Promise<CockpitSessionListResponse> => {
  return get<CockpitSessionListResponse>('/remote-cockpit/session/list', params)
}

export const startVideoStream = (uavId: string, config?: Partial<VideoStreamConfig>): Promise<VideoStreamStatus> => {
  return post<VideoStreamStatus>('/remote-cockpit/video/start', { uav_id: uavId, config })
}

export const stopVideoStream = (uavId: string): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>('/remote-cockpit/video/stop', { uav_id: uavId })
}

export const getVideoStreamStatus = (uavId: string): Promise<VideoStreamStatus> => {
  return get<VideoStreamStatus>(`/remote-cockpit/video/status/${uavId}`)
}

export const adjustVideoQuality = (
  uavId: string,
  bitrateKbps?: number,
  resolution?: string
): Promise<VideoStreamConfig> => {
  return post<VideoStreamConfig>('/remote-cockpit/video/adjust-quality', {
    uav_id: uavId,
    bitrate_kbps: bitrateKbps,
    resolution
  })
}

export const setAdaptiveQuality = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>('/remote-cockpit/video/adaptive', { uav_id: uavId, enabled })
}

export const getNetworkMetrics = (uavId: string): Promise<NetworkMetrics> => {
  return get<NetworkMetrics>(`/remote-cockpit/network/metrics/${uavId}`)
}

export const getNetworkMetricsHistory = (
  uavId: string,
  params?: { startTime?: number; endTime?: number; interval?: number }
): Promise<NetworkMetrics[]> => {
  return get<NetworkMetrics[]>(`/remote-cockpit/network/history/${uavId}`, params)
}

export const getCockpitLinkStatus = (uavId: string): Promise<CockpitLinkStatus> => {
  return get<CockpitLinkStatus>(`/remote-cockpit/link/status/${uavId}`)
}

export const setLinkFailover = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>('/remote-cockpit/link/failover', { uav_id: uavId, enabled })
}

export const setPrimaryLink = (uavId: string, linkType: LinkType): Promise<CockpitLinkStatus> => {
  return post<CockpitLinkStatus>('/remote-cockpit/link/primary', { uav_id: uavId, link_type: linkType })
}

export const setAutoMissionFallback = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>('/remote-cockpit/fallback/mission', { uav_id: uavId, enabled })
}

export const sendFlightControlCommand = (
  uavId: string,
  pitch: number,
  roll: number,
  yaw: number,
  throttle: number,
  source: 'keyboard' | 'gamepad' | 'autopilot' = 'gamepad'
): Promise<FlightControlCommand> => {
  return post<FlightControlCommand>('/remote-cockpit/control/command', {
    uav_id: uavId,
    pitch,
    roll,
    yaw,
    throttle,
    source,
    timestamp: Date.now()
  })
}

export const getAvailableCockpitUAVs = (): Promise<{ uav_ids: string[] }> => {
  return get<{ uav_ids: string[] }>('/remote-cockpit/available-uavs')
}

export const switchCockpitUAV = (fromUavId: string, toUavId: string): Promise<{ success: boolean; uav_id: string }> => {
  return post<{ success: boolean; uav_id: string }>('/remote-cockpit/switch-uav', {
    from_uav_id: fromUavId,
    to_uav_id: toUavId
  })
}

export const getVideoStreamUrl = (uavId: string, protocol: 'webrtc' | 'hls' | 'ws' = 'webrtc'): Promise<{ url: string; protocol: string }> => {
  return get<{ url: string; protocol: string }>(`/remote-cockpit/video/stream-url/${uavId}`, { protocol })
}
