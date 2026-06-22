export type CollisionRiskLevel = 'safe' | 'warning' | 'critical' | 'avoiding' | 'resolved'

export type AvoidanceActionType =
  | 'speed_reduce'
  | 'speed_adjust'
  | 'hold_position'
  | 'waypoint_hold'
  | 'altitude_change'
  | 'resume'

export interface CollisionAlert {
  id: number
  alert_id: string
  uav_id_1: number
  uav_id_2: number
  risk_level: CollisionRiskLevel
  min_distance: number
  current_distance: number
  time_to_collision: number
  alert_type: string
  action_taken: AvoidanceActionType
  action_detail: string
  is_resolved: boolean
  resolved_at: string | null
  created_at: string
  updated_at: string
}

export interface RouteIntersection {
  id: number
  uav_id_1: number
  uav_id_2: number
  mission_id_1: number
  mission_id_2: number
  waypoint_seq_1: number
  waypoint_seq_2: number
  latitude: number
  longitude: number
  altitude: number
  distance_m: number
  eta_1: string
  eta_2: string
  time_diff_sec: number
  risk_level: CollisionRiskLevel
  is_active: boolean
  created_at: string
}

export interface UAVLivePosition {
  uav_id: number
  latitude: number
  longitude: number
  altitude: number
  ground_speed: number
  heading: number
  velocity_x: number
  velocity_y: number
  velocity_z: number
  mode: string
  timestamp: string
}

export interface AvoidanceDecision {
  pair_key: string
  uav_id_1: number
  uav_id_2: number
  distance: number
  risk_level: CollisionRiskLevel
  primary_action: AvoidanceActionType
  secondary_action?: AvoidanceActionType
  speed_factor_1: number
  speed_factor_2: number
  hold_duration: number
  reason: string
}

export interface CollisionStatus {
  enabled: boolean
  active_uavs: number
  active_alerts: number
  intersections: number
  safe_distance_m: number
  warning_distance_m: number
}

export interface CollisionStats {
  total: number
  critical: number
  warning: number
  avoided: number
  resolved: number
}
