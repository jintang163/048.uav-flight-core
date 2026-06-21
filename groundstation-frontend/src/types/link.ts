export enum LinkType {
  RADIO = 1,
  LTE = 2,
  DUAL = 3
}

export enum LinkState {
  DISCONNECTED = 0,
  CONNECTING = 1,
  CONNECTED = 2,
  DEGRADED = 3
}

export interface LinkQuality {
  rssi: number
  snr: number
  packet_loss: number
  latency_ms: number
}

export interface LinkStatus {
  id: string
  uav_id: string
  active_link: LinkType
  radio_rssi: number
  radio_state: LinkState
  radio_connected: boolean
  lte_rssi: number
  lte_state: LinkState
  lte_connected: boolean
  lte_network_type: string
  packet_loss: number
  latency_ms: number
  bytes_sent: string
  bytes_received: string
  auto_switch_enabled: boolean
  timestamp: number
}

export interface LinkStatusReport {
  uav_id: string | number
  active_link: number
  radio_rssi: number
  radio_connected: boolean
  lte_rssi: number
  lte_connected: boolean
  lte_network_type: string
  packet_loss: number
  latency_ms: number
}

export interface LinkStatistics {
  total_uavs: number
  radio_connected: number
  lte_connected: number
  dual_connected: number
  disconnected: number
  avg_latency_ms: number
  avg_packet_loss: number
}

export const LinkTypeText: Record<LinkType, string> = {
  [LinkType.RADIO]: '数传电台',
  [LinkType.LTE]: '4G网络',
  [LinkType.DUAL]: '双链路'
}

export const LinkStateText: Record<LinkState, string> = {
  [LinkState.DISCONNECTED]: '断开',
  [LinkState.CONNECTING]: '连接中',
  [LinkState.CONNECTED]: '已连接',
  [LinkState.DEGRADED]: '降级'
}

export const getRSSIColor = (rssi: number, type: LinkType): string => {
  if (type === LinkType.RADIO) {
    if (rssi >= -60) return '#52c41a'
    if (rssi >= -80) return '#faad14'
    if (rssi >= -90) return '#fa8c16'
    return '#ff4d4f'
  } else {
    if (rssi >= -70) return '#52c41a'
    if (rssi >= -85) return '#faad14'
    if (rssi >= -100) return '#fa8c16'
    return '#ff4d4f'
  }
}

export const getRSSILevel = (rssi: number, type: LinkType): number => {
  if (type === LinkType.RADIO) {
    if (rssi >= -60) return 4
    if (rssi >= -80) return 3
    if (rssi >= -90) return 2
    if (rssi >= -100) return 1
    return 0
  } else {
    if (rssi >= -70) return 4
    if (rssi >= -85) return 3
    if (rssi >= -100) return 2
    if (rssi >= -110) return 1
    return 0
  }
}
