export type GeofenceType = 'inclusion' | 'exclusion'

export type GeofenceShape = 'polygon' | 'circle' | 'rectangle'

export type GeofenceCategory = 'custom' | 'airport' | 'military' | 'nuclear' | 'government' | 'national'

export type GeofenceSource = 'user' | 'national' | 'system'

export type FailAction = 'warn' | 'hover' | 'rtl' | 'land'

export interface Coordinate {
  lat: number
  lng: number
}

export interface Geofence {
  id: string
  uuid: string
  name: string
  description: string
  type: GeofenceType
  shape: GeofenceShape
  category: GeofenceCategory
  source: GeofenceSource
  creatorId: string
  isActive: boolean
  maxAltitude: number
  minAltitude: number
  maxDistance: number
  centerLat: number
  centerLng: number
  radius: number
  coordinates: Coordinate[]
  failAction: FailAction
  countryCode: string
  cityName: string
  adminLevel: number
  createdAt: string
  updatedAt: string
  uavs?: Array<{ id: string; name: string }>
}

export type ViolationType = 'altitude_exceeded' | 'altitude_too_low' | 'inside_exclusion_zone' | 'outside_inclusion_zone' | 'distance_exceeded'

export type ViolationSeverity = 'warning' | 'critical' | 'fatal'

export interface GeofenceViolation {
  id: string
  uavId: string
  uavName?: string
  geofenceId: string
  geofenceName: string
  geofenceCategory: GeofenceCategory
  violationType: ViolationType
  severity: ViolationSeverity
  latitude: number
  longitude: number
  altitude: number
  distance: number
  duration: number
  actionTaken: FailAction
  actionResult: string
  isResolved: boolean
  resolvedAt?: string
  notes: string
  createdAt: string
  updatedAt: string
}

export type UnlockStatus = 'pending' | 'approved' | 'rejected' | 'expired' | 'cancelled'

export interface TemporaryUnlocking {
  id: string
  uuid: string
  uavId: string
  uavName?: string
  geofenceId: string
  geofenceName?: string
  applicantId: string
  applicantName?: string
  approverId?: string
  approverName?: string
  title: string
  reason: string
  status: UnlockStatus
  category: GeofenceCategory
  unlockType: string
  startTime?: string
  endTime?: string
  maxAltitude: number
  maxDistance: number
  centerLat: number
  centerLng: number
  radius: number
  approvalRemark: string
  approvedAt?: string
  cancelledAt?: string
  missionId?: string
  contactName: string
  contactPhone: string
  createdAt: string
  updatedAt: string
}

export interface ViolationStatistics {
  total: number
  unresolved: number
  critical: number
  byType: Record<string, number>
}

export interface FlightRestrictionZone {
  id: string
  name: string
  type: string
  category: string
  shape: GeofenceShape
  altitudeMin: number
  altitudeMax: number
  reason: string
  effectiveFrom?: number
  effectiveTo?: number
  isActive: boolean
}
