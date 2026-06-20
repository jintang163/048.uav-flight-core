export enum GeofenceType {
  POLYGON = 'polygon',
  CIRCLE = 'circle',
  RECTANGLE = 'rectangle'
}

export enum GeofenceAction {
  WARN = 'warn',
  RTL = 'rtl',
  LAND = 'land',
  HOLD = 'hold'
}

export interface GeofencePolygon {
  type: GeofenceType.POLYGON
  points: { lat: number; lng: number }[]
}

export interface GeofenceCircle {
  type: GeofenceType.CIRCLE
  center: { lat: number; lng: number }
  radius: number
}

export interface GeofenceRectangle {
  type: GeofenceType.RECTANGLE
  northeast: { lat: number; lng: number }
  southwest: { lat: number; lng: number }
}

export type GeofenceShape = GeofencePolygon | GeofenceCircle | GeofenceRectangle

export interface Geofence {
  id: string
  name: string
  description?: string
  shape: GeofenceShape
  action: GeofenceAction
  altitudeMin?: number
  altitudeMax?: number
  isInclusion: boolean
  isEnabled: boolean
  createdAt: number
  updatedAt: number
  color?: string
}

export interface GeofenceViolation {
  id: string
  geofenceId: string
  geofenceName: string
  uavId: string
  uavName: string
  timestamp: number
  position: {
    lat: number
    lng: number
    alt: number
  }
  resolved: boolean
  resolvedAt?: number
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
