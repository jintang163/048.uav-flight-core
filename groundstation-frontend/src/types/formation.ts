export enum FormationType {
  LINE = 'line',
  TRIANGLE = 'triangle',
  CIRCLE = 'circle'
}

export enum FormationStatus {
  IDLE = 'idle',
  READY = 'ready',
  EXECUTING = 'executing',
  PAUSED = 'paused',
  COMPLETED = 'completed'
}

export enum LightEffect {
  STATIC = 'static',
  BLINK = 'blink',
  RAINBOW = 'rainbow',
  BREATHING = 'breathing'
}

export interface FormationMember {
  id: string
  formationId: string
  uavId: string
  positionIndex: number
  offsetX: number
  offsetY: number
  offsetZ: number
  isLeader: boolean
  status: string
  uav?: {
    id: string
    name: string
    status: string
  }
}

export interface Formation {
  id: string
  uuid: string
  name: string
  type: FormationType
  status: FormationStatus
  leaderId: string
  spacing: number
  description: string
  ownerId: string
  createdAt: number
  updatedAt: number
  members?: FormationMember[]
}

export interface FormationCollisionWarning {
  id: string
  formationId: string
  uavId1: string
  uavId2: string
  distance: number
  warningLevel: 'warning' | 'critical'
  timestamp: number
  resolved: boolean
  resolvedAt?: number
}

export interface FormationLightConfig {
  red: number
  green: number
  blue: number
  effect: LightEffect
}

export interface CreateFormationRequest {
  name: string
  type: FormationType
  spacing?: number
  description?: string
  uavIds?: string[]
  leaderId?: string
}

export interface UpdateFormationRequest {
  name?: string
  type?: FormationType
  spacing?: number
  description?: string
}
