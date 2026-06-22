import { useEffect, useCallback, useRef } from 'react'
import { useAppDispatch, useAppSelector } from '@/store'
import {
  setCockpitActive,
  setCockpitMode,
  setCurrentUAV,
  setAvailableUAVs,
  setVideoConfig,
  updateVideoConfig,
  setVideoStatus,
  updateVideoStatus,
  setNetworkMetrics,
  setLinkStatus,
  updateLinkStatus,
  setHIDEnabled,
  setActiveHIDDevice,
  setHIDAxes,
  setHIDButtons,
  startCockpitSession as startSessionAction,
  endCockpitSession as endSessionAction,
  setAutoMissionFallback,
  setVideoDisconnectTime,
  incrementQualityAdjustmentCount,
  applyVideoQualityPreset,
  resetRemoteCockpit,
  selectRemoteCockpitState,
  selectIsCockpitActive,
  selectCockpitMode,
  selectCurrentUAVId,
  selectVideoConfig,
  selectVideoStatus,
  selectNetworkMetrics,
  selectLinkStatus,
  selectHIDState,
  selectCockpitSession
} from '@/store/slices/remote-cockpit'
import {
  startCockpitSessionWS,
  endCockpitSessionWS,
  startVideoStreamWS,
  stopVideoStreamWS,
  requestVideoQualityChangeWS,
  setAdaptiveQualityWS,
  sendFlightControlCommandWS,
  requestCockpitLinkStatusWS,
  requestAvailableCockpitUAVsWS,
  switchCockpitUAVWS,
  setLinkFailoverEnabledWS,
  setAutoMissionFallbackWS,
  setPrimaryLinkWS
} from '@/websocket/remote-cockpit'
import {
  startCockpitSession,
  endCockpitSession,
  startVideoStream,
  stopVideoStream,
  getVideoStreamStatus,
  adjustVideoQuality,
  setAdaptiveQuality,
  getNetworkMetrics,
  getCockpitLinkStatus,
  setLinkFailover,
  setPrimaryLink,
  setAutoMissionFallback as setAutoMissionFallbackApi,
  sendFlightControlCommand,
  getAvailableCockpitUAVs,
  switchCockpitUAV,
  getVideoStreamUrl
} from '@/api/remote-cockpit'
import { getWebSocketClient } from '@/websocket/client'
import type {
  CockpitMode,
  VideoStreamConfig,
  VideoQualityPreset,
  VideoResolution,
  HIDAxisState,
  HIDButtonState,
  ResolutionOrder
} from '@/types'
import { LinkType } from '@/types/link'
import type { UAVMode } from '@/types/uav'

const VIDEO_DISCONNECT_THRESHOLD_MS = 2000
const COMMAND_SEND_INTERVAL_MS = 50
const QUALITY_ADJUSTMENT_COOLDOWN_MS = 3000
const NETWORK_METRICS_POLL_INTERVAL_MS = 1000

export const useRemoteCockpit = (uavId?: string) => {
  const dispatch = useAppDispatch()
  const state = useAppSelector(selectRemoteCockpitState)
  const isActive = useAppSelector(selectIsCockpitActive)
  const mode = useAppSelector(selectCockpitMode)
  const currentUAVId = useAppSelector(selectCurrentUAVId)
  const videoConfig = useAppSelector(selectVideoConfig)
  const videoStatus = useAppSelector(selectVideoStatus)
  const networkMetrics = useAppSelector(selectNetworkMetrics)
  const linkStatus = useAppSelector(selectLinkStatus)
  const hidState = useAppSelector(selectHIDState)
  const session = useAppSelector(selectCockpitSession)

  const wsClientRef = useRef(getWebSocketClient())
  const commandSendTimerRef = useRef<number | null>(null)
  const networkMetricsTimerRef = useRef<number | null>(null)
  const videoDisconnectCheckTimerRef = useRef<number | null>(null)
  const lastQualityAdjustmentRef = useRef<number>(0)
  const pendingCommandRef = useRef<{ pitch: number; roll: number; yaw: number; throttle: number } | null>(null)
  const commandsSentRef = useRef<number>(0)
  const sessionStartTimeRef = useRef<number>(0)
  const flightTimeTimerRef = useRef<number | null>(null)

  const effectiveUAVId = uavId || currentUAVId

  const startSession = useCallback(async (targetUavId: string, pilotId?: string) => {
    try {
      dispatch(setCockpitMode(CockpitMode.CONNECTING))
      dispatch(setCurrentUAV(targetUavId))

      const client = wsClientRef.current
      if (client?.connected) {
        startCockpitSessionWS(client, targetUavId, pilotId)
      }

      const result = await startCockpitSession(targetUavId, pilotId)
      dispatch(startSessionAction(result))
      sessionStartTimeRef.current = Date.now()

      if (flightTimeTimerRef.current) {
        clearInterval(flightTimeTimerRef.current)
      }
      flightTimeTimerRef.current = window.setInterval(() => {
        const elapsed = Date.now() - sessionStartTimeRef.current
        dispatch({
          type: 'remoteCockpit/updateSessionStats',
          payload: { total_flight_time_ms: elapsed }
        })
      }, 1000)

      return result
    } catch (error) {
      dispatch(setCockpitMode(CockpitMode.DISCONNECTED))
      throw error
    }
  }, [dispatch])

  const endSession = useCallback(async () => {
    const targetId = effectiveUAVId
    if (!targetId) return

    try {
      await stopVideoStreaming()

      const client = wsClientRef.current
      if (client?.connected) {
        endCockpitSessionWS(client, targetId)
      }

      await endCockpitSession(targetId)
      dispatch(endSessionAction())

      if (flightTimeTimerRef.current) {
        clearInterval(flightTimeTimerRef.current)
        flightTimeTimerRef.current = null
      }
    } catch (error) {
      console.error('Failed to end cockpit session:', error)
    }
  }, [dispatch, effectiveUAVId])

  const startVideoStreaming = useCallback(async (config?: Partial<VideoStreamConfig>) => {
    const targetId = effectiveUAVId
    if (!targetId) return

    try {
      const client = wsClientRef.current
      if (client?.connected) {
        startVideoStreamWS(client, targetId, config)
      }

      const status = await startVideoStream(targetId, config)
      dispatch(setVideoStatus(status))
      dispatch(setVideoDisconnectTime(null))

      startVideoDisconnectMonitor()
      return status
    } catch (error) {
      console.error('Failed to start video stream:', error)
      throw error
    }
  }, [dispatch, effectiveUAVId])

  const stopVideoStreaming = useCallback(async () => {
    const targetId = effectiveUAVId
    if (!targetId) return

    try {
      if (videoDisconnectCheckTimerRef.current) {
        clearInterval(videoDisconnectCheckTimerRef.current)
        videoDisconnectCheckTimerRef.current = null
      }

      const client = wsClientRef.current
      if (client?.connected) {
        stopVideoStreamWS(client, targetId)
      }

      await stopVideoStream(targetId)
      dispatch(updateVideoStatus({ active: false, fps: 0, current_bitrate_kbps: 0 }))
    } catch (error) {
      console.error('Failed to stop video stream:', error)
    }
  }, [dispatch, effectiveUAVId])

  const startVideoDisconnectMonitor = useCallback(() => {
    if (videoDisconnectCheckTimerRef.current) {
      clearInterval(videoDisconnectCheckTimerRef.current)
    }

    videoDisconnectCheckTimerRef.current = window.setInterval(() => {
      if (!videoStatus.active) return

      const now = Date.now()
      const lastFrameTime = videoStatus.last_frame_time
      if (lastFrameTime && now - lastFrameTime > VIDEO_DISCONNECT_THRESHOLD_MS) {
        handleVideoDisconnect()
      }
    }, 500)
  }, [videoStatus.active, videoStatus.last_frame_time])

  const handleVideoDisconnect = useCallback(() => {
    dispatch(setVideoDisconnectTime(Date.now()))
    dispatch(updateVideoStatus({ active: false }))

    if (state.auto_mission_fallback && mode === CockpitMode.FLYING) {
      triggerAutoMissionFallback('视频流连接断开')
    }
  }, [dispatch, state.auto_mission_fallback, mode])

  const triggerAutoMissionFallback = useCallback((reason: string) => {
    dispatch(setCockpitMode(CockpitMode.MISSION))
    const targetId = effectiveUAVId
    if (targetId) {
      const client = wsClientRef.current
      if (client?.connected) {
        client.send('trigger_mission_fallback', { uav_id: targetId, reason })
      }
    }
  }, [dispatch, effectiveUAVId])

  const adjustVideoQualityAutomatically = useCallback(() => {
    if (!videoConfig.adaptive_enabled) return
    if (!networkMetrics) return

    const now = Date.now()
    if (now - lastQualityAdjustmentRef.current < QUALITY_ADJUSTMENT_COOLDOWN_MS) {
      return
    }

    const bandwidth = networkMetrics.bandwidth_estimate_kbps
    const packetLoss = networkMetrics.packet_loss
    const latency = networkMetrics.rtt_ms

    let shouldAdjust = false
    let newBitrate = videoConfig.bitrate_kbps
    let newResolution = videoConfig.resolution
    let reason = ''

    const currentResIndex = ResolutionOrder.indexOf(videoConfig.resolution)
    const minResIndex = ResolutionOrder.indexOf(videoConfig.min_resolution)
    const maxResIndex = ResolutionOrder.indexOf(videoConfig.max_resolution)

    if (packetLoss > 5 || latency > 200) {
      shouldAdjust = true
      reason = `网络质量下降 (丢包:${packetLoss.toFixed(1)}%, 延迟:${latency}ms)`
      newBitrate = Math.max(videoConfig.min_bitrate_kbps, videoConfig.bitrate_kbps * 0.7)
      if (currentResIndex > minResIndex) {
        newResolution = ResolutionOrder[currentResIndex - 1]
      }
    } else if (bandwidth > videoConfig.bitrate_kbps * 1.5 && packetLoss < 1 && latency < 100) {
      shouldAdjust = true
      reason = `网络质量良好 (带宽:${bandwidth}kbps, 丢包:${packetLoss.toFixed(1)}%)`
      newBitrate = Math.min(videoConfig.max_bitrate_kbps, videoConfig.bitrate_kbps * 1.2)
      if (currentResIndex < maxResIndex) {
        newResolution = ResolutionOrder[currentResIndex + 1]
      }
    }

    if (shouldAdjust && (newBitrate !== videoConfig.bitrate_kbps || newResolution !== videoConfig.resolution)) {
      lastQualityAdjustmentRef.current = now
      dispatch(incrementQualityAdjustmentCount())
      dispatch(updateVideoConfig({ bitrate_kbps: Math.round(newBitrate), resolution: newResolution }))

      const targetId = effectiveUAVId
      if (targetId) {
        const client = wsClientRef.current
        if (client?.connected) {
          requestVideoQualityChangeWS(client, targetId, Math.round(newBitrate), newResolution)
        }
        adjustVideoQuality(targetId, Math.round(newBitrate), newResolution).catch(console.error)
      }
    }
  }, [videoConfig, networkMetrics, dispatch, effectiveUAVId])

  const setVideoQualityPreset = useCallback((preset: VideoQualityPreset) => {
    dispatch(applyVideoQualityPreset(preset))
    const targetId = effectiveUAVId
    if (targetId) {
      const client = wsClientRef.current
      if (client?.connected) {
        const newConfig = { ...videoConfig }
        requestVideoQualityChangeWS(client, targetId, newConfig.bitrate_kbps, newConfig.resolution)
      }
    }
  }, [dispatch, effectiveUAVId, videoConfig])

  const enableAdaptiveQuality = useCallback((enabled: boolean) => {
    dispatch(updateVideoConfig({ adaptive_enabled: enabled }))
    const targetId = effectiveUAVId
    if (targetId) {
      const client = wsClientRef.current
      if (client?.connected) {
        setAdaptiveQualityWS(client, targetId, enabled)
      }
      setAdaptiveQuality(targetId, enabled).catch(console.error)
    }
  }, [dispatch, effectiveUAVId])

  const sendControlCommand = useCallback((
    pitch: number,
    roll: number,
    yaw: number,
    throttle: number,
    source: 'keyboard' | 'gamepad' | 'autopilot' = 'gamepad'
  ) => {
    pendingCommandRef.current = { pitch, roll, yaw, throttle }

    const targetId = effectiveUAVId
    if (!targetId) return

    const client = wsClientRef.current
    if (client?.connected) {
      sendFlightControlCommandWS(client, targetId, pitch, roll, yaw, throttle, source)
    }
  }, [effectiveUAVId])

  const startCommandSender = useCallback(() => {
    if (commandSendTimerRef.current) return

    commandSendTimerRef.current = window.setInterval(() => {
      const cmd = pendingCommandRef.current
      if (!cmd || !effectiveUAVId) return

      sendFlightControlCommand(
        effectiveUAVId,
        cmd.pitch,
        cmd.roll,
        cmd.yaw,
        cmd.throttle
      ).catch(() => {})

      commandsSentRef.current += 1
      if (commandsSentRef.current % 20 === 0) {
        dispatch({
          type: 'remoteCockpit/updateSessionStats',
          payload: { commands_sent: commandsSentRef.current }
        })
      }
    }, COMMAND_SEND_INTERVAL_MS)
  }, [dispatch, effectiveUAVId])

  const stopCommandSender = useCallback(() => {
    if (commandSendTimerRef.current) {
      clearInterval(commandSendTimerRef.current)
      commandSendTimerRef.current = null
    }
  }, [])

  const fetchLinkStatus = useCallback(async () => {
    const targetId = effectiveUAVId
    if (!targetId) return

    try {
      const client = wsClientRef.current
      if (client?.connected) {
        requestCockpitLinkStatusWS(client, targetId)
      }

      const status = await getCockpitLinkStatus(targetId)
      dispatch(setLinkStatus(status))
    } catch (error) {
      console.error('Failed to fetch link status:', error)
    }
  }, [dispatch, effectiveUAVId])

  const enableLinkFailover = useCallback((enabled: boolean) => {
    dispatch(updateLinkStatus({ failover_enabled: enabled }))
    const targetId = effectiveUAVId
    if (targetId) {
      const client = wsClientRef.current
      if (client?.connected) {
        setLinkFailoverEnabledWS(client, targetId, enabled)
      }
      setLinkFailover(targetId, enabled).catch(console.error)
    }
  }, [dispatch, effectiveUAVId])

  const switchPrimaryLink = useCallback((linkType: LinkType) => {
    const targetId = effectiveUAVId
    if (!targetId) return

    const client = wsClientRef.current
    if (client?.connected) {
      setPrimaryLinkWS(client, targetId, linkType)
    }
    setPrimaryLink(targetId, linkType).catch(console.error)
  }, [effectiveUAVId])

  const enableAutoMissionFallback = useCallback((enabled: boolean) => {
    dispatch(setAutoMissionFallback(enabled))
    const targetId = effectiveUAVId
    if (targetId) {
      const client = wsClientRef.current
      if (client?.connected) {
        setAutoMissionFallbackWS(client, targetId, enabled)
      }
      setAutoMissionFallbackApi(targetId, enabled).catch(console.error)
    }
  }, [dispatch, effectiveUAVId])

  const fetchAvailableUAVs = useCallback(async () => {
    try {
      const client = wsClientRef.current
      if (client?.connected) {
        requestAvailableCockpitUAVsWS(client)
      }

      const result = await getAvailableCockpitUAVs()
      dispatch(setAvailableUAVs(result.uav_ids))
      return result.uav_ids
    } catch (error) {
      console.error('Failed to fetch available UAVs:', error)
      return []
    }
  }, [dispatch])

  const switchUAV = useCallback(async (toUavId: string) => {
    const fromUavId = effectiveUAVId
    if (!fromUavId || fromUavId === toUavId) return

    try {
      await stopVideoStreaming()
      stopCommandSender()

      const client = wsClientRef.current
      if (client?.connected) {
        switchCockpitUAVWS(client, fromUavId, toUavId)
      }

      const result = await switchCockpitUAV(fromUavId, toUavId)
      dispatch(setCurrentUAV(toUavId))

      await startVideoStreaming()
      startCommandSender()

      return result
    } catch (error) {
      console.error('Failed to switch UAV:', error)
      throw error
    }
  }, [dispatch, effectiveUAVId, startVideoStreaming, stopVideoStreaming, startCommandSender, stopCommandSender])

  const fetchStreamUrl = useCallback(async (protocol: 'webrtc' | 'hls' | 'ws' = 'webrtc') => {
    const targetId = effectiveUAVId
    if (!targetId) return null
    try {
      return await getVideoStreamUrl(targetId, protocol)
    } catch (error) {
      console.error('Failed to get stream URL:', error)
      return null
    }
  }, [effectiveUAVId])

  const setHIDAxesState = useCallback((axes: HIDAxisState) => {
    dispatch(setHIDAxes(axes))
    if (hidState.enabled && isActive) {
      sendControlCommand(axes.pitch, axes.roll, axes.yaw, axes.throttle, 'gamepad')
    }
  }, [dispatch, hidState.enabled, isActive, sendControlCommand])

  const setHIDButtonsState = useCallback((buttons: HIDButtonState) => {
    dispatch(setHIDButtons(buttons))
  }, [dispatch])

  const toggleHIDEnabled = useCallback((enabled: boolean) => {
    dispatch(setHIDEnabled(enabled))
  }, [dispatch])

  const selectHIDDevice = useCallback((deviceId: string | null) => {
    dispatch(setActiveHIDDevice(deviceId))
  }, [dispatch])

  useEffect(() => {
    if (!isActive || !effectiveUAVId) return

    if (networkMetricsTimerRef.current) {
      clearInterval(networkMetricsTimerRef.current)
    }

    networkMetricsTimerRef.current = window.setInterval(async () => {
      try {
        const metrics = await getNetworkMetrics(effectiveUAVId)
        dispatch(setNetworkMetrics({ ...metrics, timestamp: Date.now() }))
        adjustVideoQualityAutomatically()
      } catch (error) {
      }
    }, NETWORK_METRICS_POLL_INTERVAL_MS)

    fetchLinkStatus()
    startCommandSender()

    return () => {
      if (networkMetricsTimerRef.current) {
        clearInterval(networkMetricsTimerRef.current)
        networkMetricsTimerRef.current = null
      }
      stopCommandSender()
    }
  }, [isActive, effectiveUAVId, dispatch, fetchLinkStatus, adjustVideoQualityAutomatically, startCommandSender, stopCommandSender])

  useEffect(() => {
    return () => {
      if (flightTimeTimerRef.current) {
        clearInterval(flightTimeTimerRef.current)
      }
      if (commandSendTimerRef.current) {
        clearInterval(commandSendTimerRef.current)
      }
      if (networkMetricsTimerRef.current) {
        clearInterval(networkMetricsTimerRef.current)
      }
      if (videoDisconnectCheckTimerRef.current) {
        clearInterval(videoDisconnectCheckTimerRef.current)
      }
    }
  }, [])

  return {
    state,
    isActive,
    mode,
    currentUAVId: effectiveUAVId,
    videoConfig,
    videoStatus,
    networkMetrics,
    linkStatus,
    hidState,
    session,
    startSession,
    endSession,
    startVideoStreaming,
    stopVideoStreaming,
    setVideoQualityPreset,
    enableAdaptiveQuality,
    sendControlCommand,
    fetchLinkStatus,
    enableLinkFailover,
    switchPrimaryLink,
    enableAutoMissionFallback,
    fetchAvailableUAVs,
    switchUAV,
    fetchStreamUrl,
    setHIDAxesState,
    setHIDButtonsState,
    toggleHIDEnabled,
    selectHIDDevice,
    setCockpitMode: (mode: CockpitMode) => dispatch(setCockpitMode(mode))
  }
}

export default useRemoteCockpit
