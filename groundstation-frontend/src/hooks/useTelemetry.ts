import { useEffect, useRef, useMemo } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import { updateTelemetry, setTelemetryConnected } from '@/store/slices/telemetry'
import { getWebSocketClient } from '@/websocket/client'
import type { TelemetryData, GPSData, UAVBattery, UAVAttitude, RCChannels as RCChannelsType } from '@/types'

export const useTelemetry = (uavId?: string) => {
  const dispatch = useAppDispatch()
  const { currentData, history, connected } = useAppSelector(state => state.telemetry)
  const wsClient = useRef(getWebSocketClient())

  useEffect(() => {
    const client = wsClient.current
    if (!client) return

    const handleTelemetry = (data: unknown) => {
      const telemetry = data as TelemetryData
      if (!uavId || telemetry.uavId === uavId) {
        dispatch(updateTelemetry(telemetry))
      }
    }

    const handleConnect = () => {
      dispatch(setTelemetryConnected(true))
    }

    const handleDisconnect = () => {
      dispatch(setTelemetryConnected(false))
    }

    client.on('telemetry', handleTelemetry)
    client.on('connect', handleConnect)
    client.on('disconnect', handleDisconnect)

    if (uavId && client.connected) {
      client.send('subscribe_uav', { uavId })
    }

    return () => {
      if (!client) return
      client.off('telemetry', handleTelemetry)
      client.off('connect', handleConnect)
      client.off('disconnect', handleDisconnect)

      if (uavId && client.connected) {
        client.send('unsubscribe_uav', { uavId })
      }
    }
  }, [dispatch, uavId])

  const attitude: UAVAttitude | null = useMemo(() => {
    if (!currentData) return null
    return currentData.attitude
  }, [currentData])

  const altitude: number | null = useMemo(() => {
    if (!currentData) return null
    return currentData.position?.alt || 0
  }, [currentData])

  const airspeed: number = useMemo(() => {
    if (!currentData) return 0
    return currentData.velocity?.airSpeed || 0
  }, [currentData])

  const throttle: number = useMemo(() => {
    if (!currentData) return 0
    return currentData.system?.load || 0
  }, [currentData])

  const battery: UAVBattery | null = useMemo(() => {
    if (!currentData) return null
    if (currentData.battery) return currentData.battery
    if (currentData.system) {
      return {
        voltage: currentData.system.voltageBattery / 1000,
        current: currentData.system.currentBattery / 100,
        remaining: currentData.system.batteryRemaining,
        temperature: 0,
        cells: []
      }
    }
    return null
  }, [currentData])

  const gps: GPSData | null = useMemo(() => {
    if (!currentData) return null
    return currentData.gps || null
  }, [currentData])

  const position: { lat: number; lng: number; alt: number; heading: number } | null = useMemo(() => {
    if (!currentData || !currentData.position) return null
    return {
      lat: currentData.position.lat,
      lng: currentData.position.lng,
      alt: currentData.position.alt,
      heading: currentData.attitude?.yaw || 0
    }
  }, [currentData])

  const trajectory: { lat: number; lng: number }[] = useMemo(() => {
    return history.map(p => ({
      lat: p.altitude,
      lng: 0
    }))
  }, [history])

  const flightTime: number | null = useMemo(() => {
    if (!currentData) return null
    return 0
  }, [currentData])

  const rcChannels: RCChannelsType | null = useMemo(() => {
    if (!currentData) return null
    return currentData.rc || null
  }, [currentData])

  const rssi: number = useMemo(() => {
    if (!currentData || !currentData.rc) return 0
    return currentData.rc.rssi || 0
  }, [currentData])

  return {
    currentData,
    history,
    connected,
    attitude,
    altitude,
    airspeed,
    throttle,
    battery,
    gps,
    position,
    trajectory,
    flightTime,
    rcChannels,
    rssi
  }
}
