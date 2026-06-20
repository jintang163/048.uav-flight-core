import { get, post } from './http'
import type { FlightRecord, TelemetryHistory, PageResult } from '@/types'

export const getFlightHistory = (params?: { page?: number; pageSize?: number; uavId?: string; startTime?: number; endTime?: number }): Promise<PageResult<FlightRecord>> => {
  return get<PageResult<FlightRecord>>('/flight/history', params)
}

export const getFlightDetail = (id: string): Promise<FlightRecord> => {
  return get<FlightRecord>(`/flight/${id}`)
}

export const getFlightTelemetry = (flightId: string): Promise<TelemetryHistory> => {
  return get<TelemetryHistory>(`/flight/${flightId}/telemetry`)
}

export const getFlightTrajectory = (flightId: string): Promise<{ lat: number; lng: number; alt: number; timestamp: number }[]> => {
  return get<{ lat: number; lng: number; alt: number; timestamp: number }[]>(`/flight/${flightId}/trajectory`)
}

export const exportFlightLog = (flightId: string, format: 'csv' | 'json' | 'kml'): Promise<Blob> => {
  return get<Blob>(`/flight/${flightId}/export`, { format })
}

export const getFlightStats = (params?: { uavId?: string; startTime?: number; endTime?: number }): Promise<{
  totalFlights: number
  totalFlightTime: number
  totalDistance: number
  maxAltitude: number
  maxSpeed: number
  averageFlightTime: number
}> => {
  return get<{
    totalFlights: number
    totalFlightTime: number
    totalDistance: number
    maxAltitude: number
    maxSpeed: number
    averageFlightTime: number
  }>('/flight/stats', params)
}

export const createFlightRecord = (data: Partial<FlightRecord>): Promise<FlightRecord> => {
  return post<FlightRecord>('/flight', data)
}

export const updateFlightRecord = (id: string, data: Partial<FlightRecord>): Promise<FlightRecord> => {
  return post<FlightRecord>(`/flight/${id}`, data)
}
