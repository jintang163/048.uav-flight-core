export enum AlertSeverity {
  INFO = 'info',
  WARNING = 'warning',
  ERROR = 'error',
  CRITICAL = 'critical'
}

export enum AlertCategory {
  SYSTEM = 'system',
  BATTERY = 'battery',
  GPS = 'gps',
  CONNECTION = 'connection',
  FLIGHT = 'flight',
  GEOFENCE = 'geofence',
  FAILSAFE = 'failsafe',
  FIRMWARE = 'firmware',
  SENSOR = 'sensor'
}

export enum AlertStatus {
  ACTIVE = 'active',
  ACKNOWLEDGED = 'acknowledged',
  RESOLVED = 'resolved'
}

export interface Alert {
  id: string
  severity: AlertSeverity
  category: AlertCategory
  title: string
  message: string
  uavId?: string
  uavName?: string
  status: AlertStatus
  createdAt: number
  acknowledgedAt?: number
  resolvedAt?: number
  acknowledgedBy?: string
  resolvedBy?: string
  source?: string
  code?: string
  metadata?: Record<string, unknown>
}

export interface AlertFilter {
  severity?: AlertSeverity[]
  category?: AlertCategory[]
  status?: AlertStatus[]
  uavId?: string
  startTime?: number
  endTime?: number
}

export interface AlertStats {
  total: number
  active: number
  acknowledged: number
  resolved: number
  critical: number
  error: number
  warning: number
  info: number
}

export interface NotificationSettings {
  enabled: boolean
  soundEnabled: boolean
  desktopEnabled: boolean
  voiceEnabled: boolean
  minSeverity: AlertSeverity
  categories: AlertCategory[]
}
