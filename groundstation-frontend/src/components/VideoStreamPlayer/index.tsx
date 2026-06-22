import React, { useState, useEffect, useRef, useCallback } from 'react'
import styled, { keyframes, css } from 'styled-components'
import { Card, Tag, Spin, Alert, Progress, Space, Button, Tooltip, Badge } from 'antd'
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  WarningOutlined,
  EyeOutlined,
  DashboardOutlined,
  ThunderboltOutlined,
  ClockCircleOutlined,
  DisconnectOutlined
} from '@ant-design/icons'
import type { VideoStreamStatus, VideoStreamConfig } from '@/types'
import { VideoCodec, VideoCodecText, getResolutionDimensions } from '@/types/remote-cockpit'
import { getVideoStreamUrl } from '@/api/remote-cockpit'
import { useAppDispatch } from '@/store'
import { updateVideoStatus } from '@/store/slices/remote-cockpit'

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
`

const livePulse = keyframes`
  0%, 100% { transform: scale(1); }
  50% { transform: scale(1.15); }
`

const Container = styled(Card)<{ $disconnected: boolean }>`
  background: rgba(0, 0, 0, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  position: relative;

  .ant-card-body {
    padding: 0;
    position: relative;
  }

  ${props => props.$disconnected && css`
    border-color: rgba(255, 77, 79, 0.3);
  `}
`

const VideoWrapper = styled.div`
  position: relative;
  width: 100%;
  aspect-ratio: 16 / 9;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
`

const VideoElement = styled.video`
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
`

const CanvasElement = styled.canvas<{ $visible: boolean }>`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  object-fit: contain;
  background: #000;
  display: ${props => props.$visible ? 'block' : 'none'};
`

const Overlay = styled.div<{ $visible: boolean }>`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  transition: opacity 0.3s;
  opacity: ${props => props.$visible ? 1 : 0};
  pointer-events: ${props => props.$visible ? 'auto' : 'none'};
  z-index: 10;
`

const TopBar = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(180deg, rgba(0,0,0,0.7) 0%, transparent 100%);
  z-index: 5;
  pointer-events: none;
`

const BottomBar = styled.div`
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  padding: 8px 12px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  background: linear-gradient(0deg, rgba(0,0,0,0.7) 0%, transparent 100%);
  z-index: 5;
  pointer-events: none;
`

const LiveBadge = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 10px;
  background: rgba(255, 77, 79, 0.9);
  border-radius: 12px;
  color: #fff;
  font-size: 11px;
  font-weight: 600;
`

const LiveDot = styled.div`
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #fff;
  animation: ${livePulse} 1.2s ease-in-out infinite;
`

const StatusTag = styled(Tag)<{ $active: boolean }>`
  font-size: 11px;
  font-weight: 500;

  ${props => !props.$active && css`
    animation: ${pulse} 1.5s ease-in-out infinite;
  `}
`

const MetricGroup = styled.div`
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
`

const MetricItem = styled.div`
  display: flex;
  align-items: center;
  gap: 4px;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.8);
  font-family: 'Courier New', monospace;
`

const CenterMessage = styled.div`
  text-align: center;
  color: rgba(255, 255, 255, 0.6);
`

const MessageIcon = styled.div`
  font-size: 48px;
  margin-bottom: 12px;
  color: rgba(255, 255, 255, 0.2);
`

const MessageText = styled.div`
  font-size: 14px;
  margin-bottom: 16px;
`

interface RealtimeStats {
  latency_ms: number
  jitter_ms: number
  packet_loss: number
  frames_decoded: number
  frames_dropped: number
  packetsReceived: number
  packetsLost: number
}

class H265Decoder {
  private canvas: HTMLCanvasElement | null = null
  private ws: WebSocket | null = null
  private decoderWorker: Worker | null = null
  private rendering = false

  async init(canvas: HTMLCanvasElement): Promise<void> {
    this.canvas = canvas
    try {
      this.decoderWorker = new Worker(
        URL.createObjectURL(
          new Blob(
            [
              'self.onmessage=function(e){if(e.data.type==="decode"){console.debug("[H265Decoder Worker] Would decode NAL unit, size:",e.data.nal.byteLength);self.postMessage({type:"frame",width:0,height:0})}else if(e.data.type==="init"){console.debug("[H265Decoder Worker] WASM decoder initialized (stub)")}}'
            ],
            { type: 'application/javascript' }
          )
        )
      )
      this.decoderWorker.onmessage = (e: MessageEvent) => {
        if (e.data.type === 'frame' && this.canvas && e.data.bitmap) {
          const ctx = this.canvas.getContext('2d')
          if (ctx) {
            ctx.drawImage(e.data.bitmap, 0, 0, this.canvas.width, this.canvas.height)
            e.data.bitmap.close()
          }
        }
      }
      this.decoderWorker.postMessage({ type: 'init' })
    } catch {
      console.error('[H265Decoder] Failed to initialize decoder worker')
    }
  }

  connect(url: string): void {
    if (this.ws) {
      this.ws.close()
    }
    this.rendering = true
    this.ws = new WebSocket(url)
    this.ws.binaryType = 'arraybuffer'
    this.ws.onopen = () => {
      console.debug('[H265Decoder] WebSocket connected to', url)
    }
    this.ws.onmessage = (event: MessageEvent) => {
      if (event.data instanceof ArrayBuffer) {
        this.onMessage(event.data)
      }
    }
    this.ws.onclose = () => {
      console.debug('[H265Decoder] WebSocket closed')
    }
    this.ws.onerror = () => {
      console.error('[H265Decoder] WebSocket error')
    }
  }

  disconnect(): void {
    this.rendering = false
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  private onMessage(data: ArrayBuffer): void {
    if (!this.decoderWorker) return
    const nal = new Uint8Array(data)
    this.decoderWorker.postMessage({ type: 'decode', nal }, [nal.buffer])
  }

  destroy(): void {
    this.disconnect()
    if (this.decoderWorker) {
      this.decoderWorker.terminate()
      this.decoderWorker = null
    }
    this.canvas = null
  }
}

function checkH265Support(): boolean {
  try {
    const capabilities = RTCRtpReceiver.getCapabilities('video')
    if (!capabilities) return false
    return capabilities.codecs.some(
      c =>
        c.mimeType.toLowerCase().includes('h265') ||
        c.mimeType.toLowerCase().includes('hevc')
    )
  } catch {
    return false
  }
}

interface VideoStreamPlayerProps {
  uavId: string
  videoStatus: VideoStreamStatus
  videoConfig: VideoStreamConfig
  onStart?: () => void
  onStop?: () => void
  autoPlay?: boolean
  showControls?: boolean
  showMetrics?: boolean
  height?: number | string
}

const VideoStreamPlayer: React.FC<VideoStreamPlayerProps> = ({
  uavId,
  videoStatus,
  videoConfig,
  onStart,
  onStop,
  autoPlay = true,
  showControls = true,
  showMetrics = true,
  height
}) => {
  const videoRef = useRef<HTMLVideoElement>(null)
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const pcRef = useRef<RTCPeerConnection | null>(null)
  const decoderRef = useRef<H265Decoder | null>(null)
  const statsTimerRef = useRef<ReturnType<typeof setInterval> | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [isWasmMode, setIsWasmMode] = useState(false)
  const [realtimeStats, setRealtimeStats] = useState<RealtimeStats>({
    latency_ms: 0,
    jitter_ms: 0,
    packet_loss: 0,
    frames_decoded: 0,
    frames_dropped: 0,
    packetsReceived: 0,
    packetsLost: 0
  })
  const loadAttemptsRef = useRef<number>(0)
  const dispatch = useAppDispatch()

  const getToken = useCallback((): string | null => {
    try {
      return localStorage.getItem('accessToken')
    } catch {
      return null
    }
  }, [])

  const collectStats = useCallback(async () => {
    const pc = pcRef.current
    if (!pc) return
    try {
      const stats = await pc.getStats()
      let latencyMs = 0
      let jitterMs = 0
      let packetLoss = 0
      let framesDecoded = 0
      let framesDropped = 0
      let packetsReceived = 0
      let packetsLost = 0

      stats.forEach(report => {
        if (report.type === 'inbound-rtp' && report.kind === 'video') {
          packetsReceived = report.packetsReceived ?? 0
          packetsLost = report.packetsLost ?? 0
          framesDecoded = report.framesDecoded ?? 0
          framesDropped = report.framesDropped ?? 0
          const jitterDelay = report.jitterBufferDelay ?? 0
          const jitterEmitted = report.jitterBufferEmittedCount ?? 1
          jitterMs = (jitterDelay / jitterEmitted) * 1000
          packetLoss =
            packetsReceived + packetsLost > 0
              ? (packetsLost / (packetsReceived + packetsLost)) * 100
              : 0
        }

        if (report.type === 'candidate-pair' && report.state === 'succeeded') {
          if (report.currentRoundTripTime !== undefined) {
            latencyMs = report.currentRoundTripTime * 1000
          }
        }
      })

      const newStats: RealtimeStats = {
        latency_ms: Math.round(latencyMs),
        jitter_ms: Math.round(jitterMs * 100) / 100,
        packet_loss: Math.round(packetLoss * 100) / 100,
        frames_decoded: framesDecoded,
        frames_dropped: framesDropped,
        packetsReceived,
        packetsLost
      }

      setRealtimeStats(newStats)
      dispatch(
        updateVideoStatus({
          latency_ms: newStats.latency_ms,
          jitter_ms: newStats.jitter_ms,
          packet_loss: newStats.packet_loss,
          frames_decoded: newStats.frames_decoded,
          frames_dropped: newStats.frames_dropped
        })
      )
    } catch {
      // stats collection failed silently
    }
  }, [dispatch])

  const startStatsCollection = useCallback(() => {
    if (statsTimerRef.current) clearInterval(statsTimerRef.current)
    statsTimerRef.current = setInterval(collectStats, 1000)
  }, [collectStats])

  const stopStatsCollection = useCallback(() => {
    if (statsTimerRef.current) {
      clearInterval(statsTimerRef.current)
      statsTimerRef.current = null
    }
  }, [])

  const closePeerConnection = useCallback(() => {
    if (pcRef.current) {
      pcRef.current.close()
      pcRef.current = null
    }
  }, [])

  const setupWebRTC = useCallback(async () => {
    if (!uavId) return
    setLoading(true)
    setError(null)

    try {
      const isH265 = videoConfig.codec === VideoCodec.H265
      const browserSupportsH265 = checkH265Support()

      if (isH265 && !browserSupportsH265) {
        await setupWasmFallback()
        return
      }

      const pc = new RTCPeerConnection({
        iceServers: [],
        bundlePolicy: 'max-bundle'
      })
      pcRef.current = pc

      pc.addTransceiver('video', { direction: 'recvonly' })

      pc.ontrack = (event: RTCTrackEvent) => {
        if (videoRef.current && event.streams[0]) {
          videoRef.current.srcObject = event.streams[0]
          videoRef.current.play().catch(() => {
            setIsPlaying(false)
          })
        }
      }

      pc.onicecandidate = (event: RTCPeerConnectionIceEvent) => {
        if (event.candidate === null) {
          sendSDPOffer(pc)
        }
      }

      pc.onconnectionstatechange = () => {
        if (pc.connectionState === 'failed' || pc.connectionState === 'disconnected') {
          setError('WebRTC 连接断开')
          setIsPlaying(false)
          stopStatsCollection()
        } else if (pc.connectionState === 'connected') {
          setIsPlaying(true)
          startStatsCollection()
        }
      }

      const offer = await pc.createOffer()
      await pc.setLocalDescription(offer)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'WebRTC 连接失败')
      setLoading(false)
    }
  }, [uavId, videoConfig.codec, startStatsCollection, stopStatsCollection])

  const sendSDPOffer = useCallback(async (pc: RTCPeerConnection) => {
    if (!pc.localDescription) return
    const token = getToken()
    try {
      const response = await fetch(`/api/v1/remote-cockpit/video/${uavId}/sdp`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/sdp',
          ...(token ? { Authorization: `Bearer ${token}` } : {})
        },
        body: pc.localDescription.sdp
      })

      if (!response.ok) {
        throw new Error(`SDP exchange failed: ${response.status}`)
      }

      const answer = await response.text()
      await pc.setRemoteDescription(new RTCSessionDescription({ type: 'answer', sdp: answer }))
      setLoading(false)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'SDP 交换失败')
      setLoading(false)
      closePeerConnection()
    }
  }, [uavId, getToken, closePeerConnection])

  const setupWasmFallback = useCallback(async () => {
    setIsWasmMode(true)

    if (canvasRef.current && !decoderRef.current) {
      const decoder = new H265Decoder()
      await decoder.init(canvasRef.current)
      decoderRef.current = decoder
    }

    try {
      const result = await getVideoStreamUrl(uavId, 'ws')
      if (result?.url) {
        decoderRef.current?.connect(result.url)
        setIsPlaying(true)
        setLoading(false)
      } else {
        const wsUrl = `ws://${window.location.hostname}:8889/ws/uav_${uavId}`
        decoderRef.current?.connect(wsUrl)
        setIsPlaying(true)
        setLoading(false)
      }
    } catch (err) {
      const wsUrl = `ws://${window.location.hostname}:8889/ws/uav_${uavId}`
      decoderRef.current?.connect(wsUrl)
      setIsPlaying(true)
      setLoading(false)
    }
  }, [uavId])

  const handleStart = useCallback(async () => {
    onStart?.()
    loadAttemptsRef.current = 0
    await setupWebRTC()
  }, [onStart, setupWebRTC])

  const handleStop = useCallback(() => {
    closePeerConnection()
    stopStatsCollection()

    if (videoRef.current) {
      videoRef.current.srcObject = null
    }

    if (decoderRef.current) {
      decoderRef.current.disconnect()
    }

    setIsWasmMode(false)
    setIsPlaying(false)
    onStop?.()
  }, [closePeerConnection, stopStatsCollection, onStop])

  const handleReload = useCallback(async () => {
    handleStop()
    setTimeout(handleStart, 500)
  }, [handleStop, handleStart])

  useEffect(() => {
    if (autoPlay && videoStatus.active) {
      setupWebRTC()
    }
    return () => {
      closePeerConnection()
      stopStatsCollection()
      if (decoderRef.current) {
        decoderRef.current.destroy()
        decoderRef.current = null
      }
      if (videoRef.current) {
        videoRef.current.srcObject = null
      }
    }
  }, [autoPlay, videoStatus.active, uavId])

  useEffect(() => {
    return () => {
      if (decoderRef.current) {
        decoderRef.current.destroy()
        decoderRef.current = null
      }
    }
  }, [])

  const handleVideoPlay = () => setIsPlaying(true)
  const handleVideoPause = () => setIsPlaying(false)
  const handleVideoError = () => {
    setError('视频播放错误，正在重试...')
    loadAttemptsRef.current += 1
    if (loadAttemptsRef.current < 3) {
      setTimeout(handleReload, 2000)
    }
  }

  const dimensions = getResolutionDimensions(videoConfig.resolution)
  const bitrateProgress = videoStatus.target_bitrate_kbps > 0
    ? Math.min(100, (videoStatus.current_bitrate_kbps / videoStatus.target_bitrate_kbps) * 100)
    : 0

  const effectiveLatency = isWasmMode ? videoStatus.latency_ms : realtimeStats.latency_ms
  const effectiveJitter = isWasmMode ? videoStatus.jitter_ms : realtimeStats.jitter_ms
  const effectivePacketLoss = isWasmMode ? videoStatus.packet_loss : realtimeStats.packet_loss
  const effectiveFramesDecoded = isWasmMode ? videoStatus.frames_decoded : realtimeStats.frames_decoded
  const effectiveFramesDropped = isWasmMode ? videoStatus.frames_dropped : realtimeStats.frames_dropped

  const frameDropRate = effectiveFramesDecoded > 0
    ? (effectiveFramesDropped / effectiveFramesDecoded) * 100
    : 0

  return (
    <Container
      $disconnected={!videoStatus.active || !!error}
      styles={{ body: { padding: 0 } }}
    >
      <VideoWrapper style={height ? { height, aspectRatio: 'unset' } : undefined}>
        <VideoElement
          ref={videoRef}
          autoPlay
          playsInline
          muted
          onPlay={handleVideoPlay}
          onPause={handleVideoPause}
          onError={handleVideoError}
          style={isWasmMode ? { display: 'none' } : undefined}
        />
        <CanvasElement ref={canvasRef} $visible={isWasmMode} />

        <TopBar>
          <Space size={8}>
            {videoStatus.active && (
              <LiveBadge>
                <LiveDot />
                LIVE
              </LiveBadge>
            )}
            <StatusTag color="cyan" $active={videoStatus.active}>
              {VideoCodecText[videoConfig.codec]} {dimensions.width}x{dimensions.height}
            </StatusTag>
            <StatusTag color={videoConfig.adaptive_enabled ? 'green' : 'default'} $active>
              {videoConfig.adaptive_enabled ? '画质自适应' : '固定画质'}
            </StatusTag>
            {isWasmMode && (
              <StatusTag color="orange" $active>
                WASM 解码
              </StatusTag>
            )}
          </Space>
          <Space size={8}>
            {showControls && (
              <>
                {isPlaying ? (
                  <Tooltip title="停止">
                    <Button
                      type="text"
                      size="small"
                      icon={<PauseCircleOutlined />}
                      onClick={handleStop}
                      style={{ color: '#fff' }}
                    />
                  </Tooltip>
                ) : (
                  <Tooltip title="播放">
                    <Button
                      type="text"
                      size="small"
                      icon={<PlayCircleOutlined />}
                      onClick={handleStart}
                      style={{ color: '#fff' }}
                    />
                  </Tooltip>
                )}
                <Tooltip title="重新连接">
                  <Button
                    type="text"
                    size="small"
                    icon={<ReloadOutlined />}
                    onClick={handleReload}
                    style={{ color: '#fff' }}
                  />
                </Tooltip>
              </>
            )}
          </Space>
        </TopBar>

        {showMetrics && (
          <BottomBar>
            <MetricGroup>
              <MetricItem>
                <EyeOutlined style={{ color: '#52c41a' }} />
                <span>{videoStatus.fps} FPS</span>
              </MetricItem>
              <MetricItem>
                <DashboardOutlined style={{ color: '#1890ff' }} />
                <span>{videoStatus.current_bitrate_kbps.toFixed(0)} / {videoStatus.target_bitrate_kbps} kbps</span>
              </MetricItem>
              <MetricItem>
                <ClockCircleOutlined style={{ color: effectiveLatency < 100 ? '#52c41a' : effectiveLatency < 200 ? '#faad14' : '#ff4d4f' }} />
                <span>延迟 {effectiveLatency}ms</span>
              </MetricItem>
            </MetricGroup>
            <MetricGroup>
              <MetricItem>
                <ThunderboltOutlined style={{ color: frameDropRate > 5 ? '#ff4d4f' : '#52c41a' }} />
                <span>丢帧 {frameDropRate.toFixed(1)}%</span>
              </MetricItem>
              <MetricItem>
                <Badge
                  status={effectivePacketLoss < 1 ? 'success' : effectivePacketLoss < 5 ? 'warning' : 'error'}
                  text={`丢包 ${effectivePacketLoss.toFixed(2)}%`}
                />
              </MetricItem>
            </MetricGroup>
          </BottomBar>
        )}

        <Overlay $visible={loading || !!error || !videoStatus.active}>
          {loading ? (
            <CenterMessage>
              <Spin size="large" style={{ color: '#1890ff' }} />
              <MessageText style={{ marginTop: 16 }}>正在连接视频流...</MessageText>
            </CenterMessage>
          ) : error ? (
            <CenterMessage>
              <MessageIcon>
                <WarningOutlined />
              </MessageIcon>
              <MessageText>{error}</MessageText>
              <Button type="primary" icon={<ReloadOutlined />} onClick={handleReload}>
                重试连接
              </Button>
            </CenterMessage>
          ) : !videoStatus.active ? (
            <CenterMessage>
              <MessageIcon>
                <DisconnectOutlined />
              </MessageIcon>
              <MessageText>视频流已断开</MessageText>
              {showControls && (
                <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart}>
                  开始图传
                </Button>
              )}
            </CenterMessage>
          ) : null}
        </Overlay>
      </VideoWrapper>

      {showMetrics && bitrateProgress > 0 && (
        <div style={{ padding: '8px 12px', background: 'rgba(255,255,255,0.02)' }}>
          <Progress
            percent={Math.round(bitrateProgress)}
            showInfo
            format={() => `码率利用率 ${bitrateProgress.toFixed(0)}%`}
            strokeColor={{
              '0%': '#52c41a',
              '100%': bitrateProgress > 90 ? '#ff4d4f' : '#1890ff'
            }}
            size="small"
          />
        </div>
      )}
    </Container>
  )
}

export default VideoStreamPlayer
