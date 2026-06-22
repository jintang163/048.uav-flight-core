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
import { VideoCodecText, getResolutionDimensions } from '@/types/remote-cockpit'
import { getVideoStreamUrl } from '@/api/remote-cockpit'

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
  const [streamUrl, setStreamUrl] = useState<string>('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const loadAttemptsRef = useRef<number>(0)

  const loadStreamUrl = useCallback(async () => {
    if (!uavId) return
    setLoading(true)
    setError(null)
    try {
      const result = await getVideoStreamUrl(uavId, 'webrtc')
      if (result?.url) {
        setStreamUrl(result.url)
      } else {
        setError('无法获取视频流地址')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载视频流失败')
    } finally {
      setLoading(false)
    }
  }, [uavId])

  useEffect(() => {
    if (autoPlay && videoStatus.active) {
      loadStreamUrl()
    }
    return () => {
      if (videoRef.current?.srcObject) {
        videoRef.current.srcObject = null
      }
    }
  }, [autoPlay, videoStatus.active, loadStreamUrl, uavId])

  useEffect(() => {
    if (videoStatus.active && streamUrl && videoRef.current) {
      if (videoRef.current.src !== streamUrl) {
        videoRef.current.src = streamUrl
      }
      videoRef.current.play().catch(() => {
        setIsPlaying(false)
      })
    } else if (!videoStatus.active && videoRef.current) {
      videoRef.current.pause()
      setIsPlaying(false)
    }
  }, [videoStatus.active, streamUrl])

  const handleVideoPlay = () => setIsPlaying(true)
  const handleVideoPause = () => setIsPlaying(false)
  const handleVideoError = () => {
    setError('视频播放错误，正在重试...')
    loadAttemptsRef.current += 1
    if (loadAttemptsRef.current < 3) {
      setTimeout(loadStreamUrl, 2000)
    }
  }

  const handleStart = async () => {
    onStart?.()
    await loadStreamUrl()
  }

  const handleStop = () => {
    if (videoRef.current) {
      videoRef.current.pause()
      videoRef.current.removeAttribute('src')
      videoRef.current.load()
    }
    setIsPlaying(false)
    onStop?.()
  }

  const handleReload = async () => {
    handleStop()
    setTimeout(handleStart, 500)
  }

  const dimensions = getResolutionDimensions(videoConfig.resolution)
  const bitrateProgress = videoStatus.target_bitrate_kbps > 0
    ? Math.min(100, (videoStatus.current_bitrate_kbps / videoStatus.target_bitrate_kbps) * 100)
    : 0

  const frameDropRate = videoStatus.frames_decoded > 0
    ? (videoStatus.frames_dropped / videoStatus.frames_decoded) * 100
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
        />

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
                <ClockCircleOutlined style={{ color: videoStatus.latency_ms < 100 ? '#52c41a' : videoStatus.latency_ms < 200 ? '#faad14' : '#ff4d4f' }} />
                <span>延迟 {videoStatus.latency_ms}ms</span>
              </MetricItem>
            </MetricGroup>
            <MetricGroup>
              <MetricItem>
                <ThunderboltOutlined style={{ color: frameDropRate > 5 ? '#ff4d4f' : '#52c41a' }} />
                <span>丢帧 {frameDropRate.toFixed(1)}%</span>
              </MetricItem>
              <MetricItem>
                <Badge
                  status={videoStatus.packet_loss < 1 ? 'success' : videoStatus.packet_loss < 5 ? 'warning' : 'error'}
                  text={`丢包 ${videoStatus.packet_loss.toFixed(2)}%`}
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
