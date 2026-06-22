import { get, post, del } from './http'
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
  return post<CockpitSession>('/remote-cockpit/sessions', { uav_id: uavId, pilot_id: pilotId })
}

export const endCockpitSession = (uavId: string): Promise<CockpitSession> => {
  return del<CockpitSession>(`/remote-cockpit/sessions/${uavId}`)
}

export const getCockpitSession = (uavId: string): Promise<CockpitSession> => {
  return get<CockpitSession>(`/remote-cockpit/sessions/${uavId}`)
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
  return post<VideoStreamStatus>(`/remote-cockpit/video/${uavId}/start`, config)
}

export const stopVideoStream = (uavId: string): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>(`/remote-cockpit/video/${uavId}/stop`)
}

export const getVideoStreamStatus = (uavId: string): Promise<VideoStreamStatus> => {
  return get<VideoStreamStatus>(`/remote-cockpit/video/${uavId}`)
}

export const adjustVideoQuality = (
  uavId: string,
  bitrateKbps?: number,
  resolution?: string
): Promise<VideoStreamConfig> => {
  return post<VideoStreamConfig>(`/remote-cockpit/video/${uavId}/quality`, {
    bitrate_kbps: bitrateKbps,
    resolution
  })
}

export const setAdaptiveQuality = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>('/remote-cockpit/video/adaptive', { uav_id: uavId, enabled })
}

export const getNetworkMetrics = (uavId: string): Promise<NetworkMetrics> => {
  return get<NetworkMetrics>(`/remote-cockpit/metrics/${uavId}/network`)
}

export const getNetworkMetricsHistory = (
  uavId: string,
  params?: { startTime?: number; endTime?: number; interval?: number }
): Promise<NetworkMetrics[]> => {
  return get<NetworkMetrics[]>(`/remote-cockpit/network/history/${uavId}`, params)
}

export const getCockpitLinkStatus = (uavId: string): Promise<CockpitLinkStatus> => {
  return get<CockpitLinkStatus>(`/remote-cockpit/link/${uavId}`)
}

export const setLinkFailover = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>(`/remote-cockpit/link/${uavId}/failover`, { enabled })
}

export const setPrimaryLink = (uavId: string, linkType: LinkType): Promise<CockpitLinkStatus> => {
  return post<CockpitLinkStatus>(`/remote-cockpit/link/${uavId}/primary`, { link_type: linkType })
}

export const setAutoMissionFallback = (uavId: string, enabled: boolean): Promise<{ success: boolean }> => {
  return post<{ success: boolean }>(`/remote-cockpit/link/${uavId}/fallback`, { enabled })
}

export const sendFlightControlCommand = (
  uavId: string,
  pitch: number,
  roll: number,
  yaw: number,
  throttle: number,
  source: 'keyboard' | 'gamepad' | 'autopilot' = 'gamepad'
): Promise<FlightControlCommand> => {
  return post<FlightControlCommand>(`/remote-cockpit/control/${uavId}`, {
    pitch,
    roll,
    yaw,
    throttle,
    source,
    timestamp: Date.now()
  })
}

export const getAvailableCockpitUAVs = (): Promise<{ uav_ids: string[] }> => {
  return get<{ uav_ids: string[] }>('/remote-cockpit/uavs')
}

export const switchCockpitUAV = (fromUavId: string, toUavId: string): Promise<{ success: boolean; uav_id: string }> => {
  return post<{ success: boolean; uav_id: string }>('/remote-cockpit/switch', {
    from_uav_id: fromUavId,
    to_uav_id: toUavId
  })
}

export const getVideoStreamUrl = (uavId: string, protocol: 'webrtc' | 'hls' | 'ws' = 'webrtc'): Promise<{ url: string; protocol: string }> => {
  return get<{ url: string; protocol: string }>(`/remote-cockpit/video/${uavId}/url`, { protocol })
}

export const triggerMissionFallback = (uavId: string, reason: string): Promise<{ message: string; reason: string }> => {
  return post<{ message: string; reason: string }>(`/remote-cockpit/link/${uavId}/fallback/trigger`, { reason })
}
