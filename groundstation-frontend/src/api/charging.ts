import { get, post, put, del } from './http'
import type {
  ChargingStation,
  ChargingSlot,
  ChargingRecord,
  ChargingStatistics,
  CreateChargingStationRequest,
  UpdateChargingStationRequest,
  StartChargingRequest,
  SlotTelemetryRequest
} from '@/types'

export const getChargingStationList = (params?: {
  page?: number
  pageSize?: number
  status?: string
  keyword?: string
}): Promise<{ list: ChargingStation[]; total: number }> => {
  return get<{ list: ChargingStation[]; total: number }>('/charging/stations', params)
}

export const getChargingStationDetail = (id: string): Promise<ChargingStation> => {
  return get<ChargingStation>(`/charging/stations/${id}`)
}

export const createChargingStation = (data: CreateChargingStationRequest): Promise<ChargingStation> => {
  return post<ChargingStation>('/charging/stations', data)
}

export const updateChargingStation = (id: string, data: UpdateChargingStationRequest): Promise<ChargingStation> => {
  return put<ChargingStation>(`/charging/stations/${id}`, data)
}

export const deleteChargingStation = (id: string): Promise<void> => {
  return del<void>(`/charging/stations/${id}`)
}

export const getChargingStationSlots = (id: string): Promise<ChargingSlot[]> => {
  return get<ChargingSlot[]>(`/charging/stations/${id}/slots`)
}

export const getChargingStatistics = (): Promise<ChargingStatistics> => {
  return get<ChargingStatistics>('/charging/stations/statistics')
}

export const stationHeartbeat = (id: string, data: {
  status?: string
  firmware_version?: string
  occupied_slots?: number
  charging_slots?: number
}): Promise<void> => {
  return post<void>(`/charging/stations/${id}/heartbeat`, data)
}

export const getChargingSlotDetail = (slotId: string): Promise<ChargingSlot> => {
  return get<ChargingSlot>(`/charging/slots/${slotId}`)
}

export const startCharging = (slotId: string, data: StartChargingRequest): Promise<ChargingRecord> => {
  return post<ChargingRecord>(`/charging/slots/${slotId}/start`, data)
}

export const stopCharging = (slotId: string, endLevel?: number): Promise<ChargingRecord> => {
  return post<ChargingRecord>(`/charging/slots/${slotId}/stop`, { end_level: endLevel })
}

export const updateSlotTelemetry = (slotId: string, data: SlotTelemetryRequest): Promise<void> => {
  return post<void>(`/charging/slots/${slotId}/telemetry`, data)
}

export const assignBatteryToSlot = (slotId: string, batteryId: string): Promise<void> => {
  return post<void>(`/charging/slots/${slotId}/assign`, { battery_id: batteryId })
}

export const removeBatteryFromSlot = (slotId: string): Promise<void> => {
  return post<void>(`/charging/slots/${slotId}/remove`, {})
}

export const setSlotFault = (slotId: string, faultCode: number, faultMessage: string): Promise<void> => {
  return post<void>(`/charging/slots/${slotId}/fault`, { fault_code: faultCode, fault_message: faultMessage })
}

export const getChargingRecords = (params?: {
  page?: number
  pageSize?: number
  battery_id?: string
  station_id?: string
  status?: string
}): Promise<{ list: ChargingRecord[]; total: number }> => {
  return get<{ list: ChargingRecord[]; total: number }>('/charging/records', params)
}

export const getChargingRecordDetail = (id: string): Promise<ChargingRecord> => {
  return get<ChargingRecord>(`/charging/records/${id}`)
}

export const getBatteryChargingRecords = (batteryId: string, params?: {
  page?: number
  pageSize?: number
}): Promise<{ list: ChargingRecord[]; total: number }> => {
  return get<{ list: ChargingRecord[]; total: number }>(`/charging/records/battery/${batteryId}`, params)
}

export const getStationChargingRecords = (stationId: string, params?: {
  page?: number
  pageSize?: number
}): Promise<{ list: ChargingRecord[]; total: number }> => {
  return get<{ list: ChargingRecord[]; total: number }>(`/charging/stations/${stationId}/records`, params)
}
