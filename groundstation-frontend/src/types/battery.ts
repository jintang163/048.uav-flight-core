export enum BatteryStatus {
  IDLE = 'idle',
  CHARGING = 'charging',
  IN_USE = 'in_use',
  DISCHARGING = 'discharging',
  STORAGE = 'storage',
  FAULT = 'fault'
}

export enum BatteryHealthStatus {
  EXCELLENT = 'excellent',
  GOOD = 'good',
  FAIR = 'fair',
  POOR = 'poor',
  CRITICAL = 'critical'
}

export enum AlertLevel {
  INFO = 'info',
  WARNING = 'warning',
  CRITICAL = 'critical',
  FATAL = 'fatal'
}

export enum AlertStatus {
  NEW = 'new',
  ACKNOWLEDGED = 'acknowledged',
  RESOLVED = 'resolved'
}

export interface Battery {
  id: string
  uuid: string
  battery_id: string
  model: string
  manufacturer: string
  capacity: number
  capacity_unit: string
  voltage: number
  cell_count: number
  current_voltage: number
  current_level: number
  current_temperature: number
  current_current: number
  soh: number
  health_status: BatteryHealthStatus
  status: BatteryStatus
  cycle_count: number
  total_flight_time: number
  total_charge_count: number
  manufacture_date?: string
  first_use_date?: string
  last_used_at?: string
  last_charged_at?: string
  storage_days: number
  needs_maintenance: boolean
  maintenance_message?: string
  uav_id?: string
  location: string
  notes: string
  created_at: string
  updated_at: string
}

export interface BatteryUsageRecord {
  id: string
  uuid: string
  battery_id: string
  uav_id?: string
  flight_mission_id?: string
  start_level: number
  end_level: number
  start_voltage: number
  end_voltage: number
  max_temperature: number
  avg_current: number
  max_current: number
  duration: number
  distance: number
  start_time?: string
  end_time?: string
  cell_voltages?: number[]
  created_at: string
  updated_at: string
}

export interface BatteryCellData {
  id: string
  battery_id: string
  cell_index: number
  voltage: number
  resistance: number
  status: string
  recorded_at: string
  created_at: string
  updated_at: string
}

export enum ChargingStationStatus {
  ONLINE = 'online',
  OFFLINE = 'offline',
  FAULT = 'fault',
  MAINTENANCE = 'maintenance'
}

export interface ChargingStation {
  id: string
  uuid: string
  station_id: string
  name: string
  model: string
  manufacturer: string
  status: ChargingStationStatus
  slot_count: number
  occupied_slots: number
  charging_slots: number
  total_charged: number
  location: string
  ip_address: string
  port: number
  protocol: string
  last_online_at?: string
  firmware_version?: string
  max_voltage: number
  max_current: number
  description?: string
  created_at: string
  updated_at: string
}

export interface ChargingSlot {
  id: string
  station_id: string
  slot_index: number
  slot_name: string
  status: string
  battery_id?: string
  charging_mode?: string
  target_voltage: number
  target_current: number
  current_voltage: number
  current_current: number
  current_level: number
  temperature: number
  charged_capacity: number
  charging_time: number
  remaining_time: number
  start_time?: string
  end_time?: string
  fault_code?: number
  fault_message?: string
  created_at: string
  updated_at: string
  station?: ChargingStation
  battery?: Battery
}

export interface ChargingRecord {
  id: string
  uuid: string
  battery_id: string
  station_id: string
  slot_id: string
  start_level: number
  end_level: number
  start_voltage: number
  end_voltage: number
  charging_mode: string
  charging_current: number
  max_temperature: number
  avg_temperature: number
  charged_capacity: number
  charging_time: number
  energy_consumed: number
  status: string
  start_time?: string
  end_time?: string
  cell_voltages_start?: number[]
  cell_voltages_end?: number[]
  fault_code?: number
  fault_message?: string
  created_at: string
  updated_at: string
  battery?: Battery
  station?: ChargingStation
}

export interface BatteryMaintenanceAlert {
  id: string
  uuid: string
  battery_id: string
  alert_type: string
  level: AlertLevel
  title: string
  message: string
  status: AlertStatus
  storage_days: number
  soh: number
  acknowledged_by?: string
  acknowledged_at?: string
  resolved_by?: string
  resolved_at?: string
  resolved_note?: string
  created_at: string
  updated_at: string
  battery?: Battery
}

export interface BatteryStatistics {
  total: number
  charging: number
  in_use: number
  idle: number
  fault: number
  excellent: number
  good: number
  fair: number
  poor: number
  critical: number
  needs_maintenance: number
  total_cycle_count: number
}

export interface ChargingStatistics {
  total_stations: number
  online_stations: number
  offline_stations: number
  fault_stations: number
  total_slots: number
  occupied_slots: number
  charging_slots: number
  available_slots: number
  today_charging_count: number
  total_charged_capacity: number
}

export interface CreateBatteryRequest {
  battery_id: string
  model?: string
  manufacturer?: string
  capacity?: number
  capacity_unit?: string
  voltage?: number
  cell_count?: number
  location?: string
  notes?: string
}

export interface UpdateBatteryRequest {
  model?: string
  manufacturer?: string
  capacity?: number
  capacity_unit?: string
  voltage?: number
  cell_count?: number
  status?: string
  location?: string
  notes?: string
}

export interface BatteryTelemetryRequest {
  voltage: number
  level: number
  temperature?: number
  current?: number
  cell_voltages?: number[]
}

export interface CreateChargingStationRequest {
  station_id: string
  name: string
  model?: string
  manufacturer?: string
  slot_count?: number
  location?: string
  ip_address?: string
  port?: number
  protocol?: string
  max_voltage?: number
  max_current?: number
  description?: string
}

export interface UpdateChargingStationRequest {
  name?: string
  model?: string
  manufacturer?: string
  slot_count?: number
  location?: string
  ip_address?: string
  port?: number
  protocol?: string
  max_voltage?: number
  max_current?: number
  description?: string
  status?: string
}

export interface StartChargingRequest {
  battery_id: string
  charging_mode?: string
  target_voltage?: number
  target_current?: number
}

export interface SlotTelemetryRequest {
  voltage?: number
  current?: number
  level?: number
  temperature?: number
  charged_capacity?: number
  charging_time?: number
  remaining_time?: number
}
