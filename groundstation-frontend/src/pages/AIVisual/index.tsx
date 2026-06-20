import React, { useEffect, useRef, useState, useCallback } from 'react'
import styled from 'styled-components'
import {
  Row,
  Col,
  Card,
  Button,
  Space,
  Tag,
  Select,
  Switch,
  Badge,
  Tooltip,
  Slider,
  Divider,
  message,
  Spin
} from 'antd'
import {
  ScanOutlined,
  AimOutlined,
  StopOutlined,
  VideoCameraOutlined,
  EyeOutlined,
  AlertOutlined,
  SearchOutlined,
  CompassOutlined,
  ThunderboltOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import {
  lockTarget,
  stopTracking as stopTrackingAction,
  fetchActiveTracking,
  setSelectedBbox,
  setIsDrawing,
  fetchDetections
} from '@/store/slices/tracking'
import { getUAVList } from '@/api/uav'
import { useWebSocket } from '@/hooks/useWebSocket'
import { subscribeToTracking, unsubscribeFromTracking } from '@/websocket/telemetry'
import {
  DetectionClass,
  DetectionClassLabels,
  DetectionClassColors,
  TrackingStatus,
  TrackingStatusLabels,
  TrackingStatusColors
} from '@/types/tracking'
import type { UAVListItem, DetectionTarget, TrackingTask, BoundingBox } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 16px;
  overflow: hidden;
`

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 24px;
`

const Title = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
`

const UAVSelector = styled(Select)`
  width: 250px;
  .ant-select-selector {
    background: rgba(255, 255, 255, 0.1) !important;
    border: 1px solid rgba(255, 255, 255, 0.2) !important;
  }
  .ant-select-selection-item {
    color: #fff !important;
  }
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 340px;
  gap: 16px;
  overflow: hidden;
`

const VideoPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const InfoPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
`

const VideoCard = styled(Card)`
  flex: 1;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;

  .ant-card-head {
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    min-height: 48px;
  }

  .ant-card-body {
    height: calc(100% - 57px);
    padding: 0;
    position: relative;
  }
`

const VideoContainer = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  background: #000;
  display: flex;
  align-items: center;
  justify-content: center;
  overflow: hidden;
  cursor: crosshair;
`

const VideoCanvas = styled.canvas`
  max-width: 100%;
  max-height: 100%;
  object-fit: contain;
`

const OverlayCanvas = styled.canvas`
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
  cursor: crosshair;
`

const VideoPlaceholder = styled.div`
  color: rgba(255, 255, 255, 0.4);
  font-size: 14px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 12px;
`

const DetectionBox = styled.div<{
  $x: number
  $y: number
  $width: number
  $height: number
  $color: string
  $selected: boolean
}>`
  position: absolute;
  border: 2px solid ${props => (props.$selected ? '#fff' : props.$color)};
  background: ${props => props.$color}15;
  left: ${props => props.$x}%;
  top: ${props => props.$y}%;
  width: ${props => props.$width}%;
  height: ${props => props.$height}%;
  pointer-events: auto;
  cursor: pointer;
  border-radius: 2px;
  transition: all 0.15s ease;

  &:hover {
    background: ${props => props.$color}30;
    border-width: 3px;
  }
`

const DetectionLabel = styled.div<{ $color: string }>`
  position: absolute;
  top: -22px;
  left: -2px;
  background: ${props => props.$color};
  color: #fff;
  padding: 2px 6px;
  font-size: 11px;
  font-weight: 600;
  border-radius: 2px;
  white-space: nowrap;
`

const ConfidenceBar = styled.div`
  position: absolute;
  bottom: -6px;
  left: 0;
  width: 100%;
  height: 4px;
  background: rgba(255, 255, 255, 0.2);
  border-radius: 2px;
  overflow: hidden;
`

const ConfidenceFill = styled.div<{ $width: number; $color: string }>`
  height: 100%;
  width: ${props => props.$width}%;
  background: ${props => props.$color};
  transition: width 0.3s;
`

const Crosshair = styled.div<{ $x: number; $y: number; $visible: boolean }>`
  position: absolute;
  left: ${props => props.$x}%;
  top: ${props => props.$y}%;
  width: 40px;
  height: 40px;
  transform: translate(-50%, -50%);
  pointer-events: none;
  opacity: ${props => (props.$visible ? 1 : 0)};
  transition: opacity 0.2s;

  &::before,
  &::after {
    content: '';
    position: absolute;
    background: rgba(255, 77, 79, 0.8);
  }

  &::before {
    left: 50%;
    top: 0;
    width: 2px;
    height: 100%;
    transform: translateX(-50%);
  }

  &::after {
    top: 50%;
    left: 0;
    height: 2px;
    width: 100%;
    transform: translateY(-50%);
  }
`

const InfoCard = styled(Card)`
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;

  .ant-card-head {
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    min-height: 48px;
  }

  .ant-card-body {
    padding: 16px;
  }
`

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  &:last-child {
    border-bottom: none;
  }
`

const InfoLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
`

const InfoValue = styled.span`
  color: #fff;
  font-size: 13px;
  font-weight: 500;
`

const DetectionItem = styled.div<{ $selected: boolean }>`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 12px;
  margin-bottom: 6px;
  background: ${props => (props.$selected ? 'rgba(24, 144, 255, 0.15)' : 'rgba(255, 255, 255, 0.03)')};
  border: 1px solid ${props => (props.$selected ? 'rgba(24, 144, 255, 0.4)' : 'rgba(255, 255, 255, 0.08)')};
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.15s ease;

  &:hover {
    background: rgba(255, 255, 255, 0.08);
  }
`

const DetectionClassBadge = styled.span<{ $color: string }>`
  display: inline-block;
  padding: 2px 8px;
  background: ${props => props.$color}30;
  color: ${props => props.$color};
  border-radius: 4px;
  font-size: 12px;
  font-weight: 600;
`

const TrackingStatusCard = styled(InfoCard)<{ $status: TrackingStatus }>`
  border-color: ${props =>
    props.$status === TrackingStatus.TRACKING
      ? 'rgba(82, 196, 26, 0.3)'
      : props.$status === TrackingStatus.SEARCHING
      ? 'rgba(250, 173, 20, 0.3)'
      : props.$status === TrackingStatus.LOST
      ? 'rgba(255, 77, 79, 0.3)'
      : 'rgba(255, 255, 255, 0.1)'};

  .ant-card-head {
    border-color: ${props =>
      props.$status === TrackingStatus.TRACKING
        ? 'rgba(82, 196, 26, 0.2)'
        : props.$status === TrackingStatus.SEARCHING
        ? 'rgba(250, 173, 20, 0.2)'
        : props.$status === TrackingStatus.LOST
        ? 'rgba(255, 77, 79, 0.2)'
        : 'rgba(255, 255, 255, 0.1)'};
  }
`

const AIVisual: React.FC = () => {
  const dispatch = useAppDispatch()
  const { detections, activeTask, selectedBbox, isDrawing, loading, detectionLoading } = useAppSelector(
    state => state.tracking
  )
  const { uavList, selectedUAVId, currentUAV } = useAppSelector(state => state.uav)
  const { client: wsClient, isConnected } = useWebSocket()

  const videoCanvasRef = useRef<HTMLCanvasElement>(null)
  const overlayCanvasRef = useRef<HTMLCanvasElement>(null)
  const [drawingStart, setDrawingStart] = useState<{ x: number; y: number } | null>(null)
  const [drawingEnd, setDrawingEnd] = useState<{ x: number; y: number } | null>(null)
  const [mousePos, setMousePos] = useState<{ x: number; y: number } | null>(null)
  const [autoDetect, setAutoDetect] = useState(true)
  const [confidenceThreshold, setConfidenceThreshold] = useState(0.5)
  const [streaming, setStreaming] = useState(false)
  const [drawingBbox, setDrawingBbox] = useState<BoundingBox | null>(null)

  const videoContainerRef = useRef<HTMLDivElement>(null)

  useEffect(() => {
    if (uavList.length === 0) {
      getUAVList({ page: 1, pageSize: 50 }).then(result => {
        if (!selectedUAVId && result.list.length > 0) {
          // dispatch will be handled via store
        }
      })
    }
  }, [uavList.length, selectedUAVId])

  useEffect(() => {
    if (selectedUAVId) {
      dispatch(fetchActiveTracking(selectedUAVId))
      dispatch(fetchDetections({ uavId: selectedUAVId, page: 1, pageSize: 20 }))
    }
  }, [selectedUAVId, dispatch])

  useEffect(() => {
    if (wsClient && isConnected && selectedUAVId) {
      subscribeToTracking(wsClient, selectedUAVId)
      return () => {
        if (wsClient) {
          unsubscribeFromTracking(wsClient, selectedUAVId)
        }
      }
    }
  }, [wsClient, isConnected, selectedUAVId])

  useEffect(() => {
    drawOverlay()
  }, [detections, selectedBbox, drawingBbox, activeTask, mousePos, confidenceThreshold])

  const drawOverlay = useCallback(() => {
    const canvas = overlayCanvasRef.current
    const container = videoContainerRef.current
    if (!canvas || !container) return

    const rect = container.getBoundingClientRect()
    canvas.width = rect.width
    canvas.height = rect.height
    const ctx = canvas.getContext('2d')
    if (!ctx) return

    ctx.clearRect(0, 0, canvas.width, canvas.height)

    const frameWidth = 1280
    const frameHeight = 720
    const scaleX = canvas.width / frameWidth
    const scaleY = canvas.height / frameHeight
    const scale = Math.min(scaleX, scaleY)
    const offsetX = (canvas.width - frameWidth * scale) / 2
    const offsetY = (canvas.height - frameHeight * scale) / 2

    const filteredDetections = detections.filter(d => d.confidence >= confidenceThreshold)

    for (const d of filteredDetections) {
      const x = offsetX + d.bbox_x * scale
      const y = offsetY + d.bbox_y * scale
      const w = d.bbox_width * scale
      const h = d.bbox_height * scale
      const color = DetectionClassColors[d.class] || '#8c8c8c'

      ctx.strokeStyle = color
      ctx.lineWidth = 2
      ctx.strokeRect(x, y, w, h)

      const label = `${DetectionClassLabels[d.class] || d.class} ${Math.round(d.confidence * 100)}%`
      ctx.font = 'bold 12px sans-serif'
      const metrics = ctx.measureText(label)
      ctx.fillStyle = color
      ctx.fillRect(x, y - 20, metrics.width + 12, 20)
      ctx.fillStyle = '#fff'
      ctx.fillText(label, x + 6, y - 6)

      ctx.fillStyle = 'rgba(255,255,255,0.2)'
      ctx.fillRect(x, y + h + 2, w, 4)
      ctx.fillStyle = color
      ctx.fillRect(x, y + h + 2, w * d.confidence, 4)
    }

    if (activeTask && activeTask.current_bbox_x !== undefined) {
      const x = offsetX + (activeTask.current_bbox_x || 0) * scale
      const y = offsetY + (activeTask.current_bbox_y || 0) * scale
      const w = (activeTask.current_bbox_width || 0) * scale
      const h = (activeTask.current_bbox_height || 0) * scale

      ctx.strokeStyle = '#1890ff'
      ctx.lineWidth = 3
      ctx.setLineDash([8, 4])
      ctx.strokeRect(x, y, w, h)
      ctx.setLineDash([])

      const centerX = x + w / 2
      const centerY = y + h / 2
      ctx.strokeStyle = 'rgba(24, 144, 255, 0.6)'
      ctx.lineWidth = 1

      ctx.beginPath()
      ctx.moveTo(centerX - 20, centerY)
      ctx.lineTo(centerX + 20, centerY)
      ctx.moveTo(centerX, centerY - 20)
      ctx.lineTo(centerX, centerY + 20)
      ctx.stroke()
    }

    if (drawingBbox) {
      const x = offsetX + drawingBbox.x * scale
      const y = offsetY + drawingBbox.y * scale
      const w = drawingBbox.width * scale
      const h = drawingBbox.height * scale

      ctx.strokeStyle = '#ff4d4f'
      ctx.lineWidth = 2
      ctx.setLineDash([6, 3])
      ctx.strokeRect(x, y, w, h)
      ctx.setLineDash([])
    }
  }, [detections, selectedBbox, drawingBbox, activeTask, confidenceThreshold])

  const getRelativeCoords = useCallback((clientX: number, clientY: number) => {
    const container = videoContainerRef.current
    if (!container) return null
    const rect = container.getBoundingClientRect()
    const frameWidth = 1280
    const frameHeight = 720
    const scaleX = rect.width / frameWidth
    const scaleY = rect.height / frameHeight
    const scale = Math.min(scaleX, scaleY)
    const offsetX = (rect.width - frameWidth * scale) / 2
    const offsetY = (rect.height - frameHeight * scale) / 2

    const x = (clientX - rect.left - offsetX) / scale
    const y = (clientY - rect.top - offsetY) / scale
    return { x, y, width: frameWidth, height: frameHeight }
  }, [])

  const handleMouseDown = useCallback(
    (e: React.MouseEvent) => {
      if (activeTask?.status === TrackingStatus.TRACKING) return
      const coords = getRelativeCoords(e.clientX, e.clientY)
      if (!coords) return
      setDrawingStart({ x: coords.x, y: coords.y })
      setDrawingEnd({ x: coords.x, y: coords.y })
      dispatch(setIsDrawing(true))
    },
    [getRelativeCoords, activeTask?.status, dispatch]
  )

  const handleMouseMove = useCallback(
    (e: React.MouseEvent) => {
      const coords = getRelativeCoords(e.clientX, e.clientY)
      if (!coords) return
      setMousePos({ x: ((e.clientX - e.currentTarget.getBoundingClientRect().left) / e.currentTarget.getBoundingClientRect().width) * 100, y: ((e.clientY - e.currentTarget.getBoundingClientRect().top) / e.currentTarget.getBoundingClientRect().height) * 100 })

      if (drawingStart) {
        setDrawingEnd({ x: coords.x, y: coords.y })
        const x = Math.min(drawingStart.x, coords.x)
        const y = Math.min(drawingStart.y, coords.y)
        const width = Math.abs(coords.x - drawingStart.x)
        const height = Math.abs(coords.y - drawingStart.y)
        setDrawingBbox({ x, y, width, height })
      }
    },
    [drawingStart, getRelativeCoords]
  )

  const handleMouseUp = useCallback(
    (e: React.MouseEvent) => {
      if (!drawingStart) return
      const coords = getRelativeCoords(e.clientX, e.clientY)
      if (!coords) return

      const x = Math.min(drawingStart.x, coords.x)
      const y = Math.min(drawingStart.y, coords.y)
      const width = Math.abs(coords.x - drawingStart.x)
      const height = Math.abs(coords.y - drawingStart.y)

      dispatch(setIsDrawing(false))
      setDrawingStart(null)
      setDrawingEnd(null)

      if (width < 10 || height < 10) {
        setDrawingBbox(null)
        return
      }

      const bbox: BoundingBox = { x, y, width, height }
      setDrawingBbox(bbox)
      dispatch(setSelectedBbox(bbox))
    },
    [drawingStart, getRelativeCoords, dispatch]
  )

  const handleLockTarget = useCallback(async () => {
    if (!selectedUAVId) {
      message.warning('请先选择无人机')
      return
    }
    const bbox = drawingBbox || selectedBbox
    if (!bbox) {
      message.warning('请先框选目标')
      return
    }
    if (bbox.width < 10 || bbox.height < 10) {
      message.warning('框选区域过小')
      return
    }

    try {
      await dispatch(
        lockTarget({
          uav_id: selectedUAVId,
          bbox_x: bbox.x,
          bbox_y: bbox.y,
          bbox_width: bbox.width,
          bbox_height: bbox.height,
          frame_width: 1280,
          frame_height: 720
        })
      ).unwrap()
      message.success('目标锁定成功，无人机开始追踪')
      setDrawingBbox(null)
    } catch (err) {
      message.error(err instanceof Error ? err.message : '锁定目标失败')
    }
  }, [selectedUAVId, drawingBbox, selectedBbox, dispatch])

  const handleStopTracking = useCallback(async () => {
    if (!activeTask) return
    try {
      await dispatch(stopTrackingAction(activeTask.id)).unwrap()
      message.success('已停止追踪')
    } catch (err) {
      message.error(err instanceof Error ? err.message : '停止追踪失败')
    }
  }, [activeTask, dispatch])

  const handleDetectFromList = useCallback(
    (detection: DetectionTarget) => {
      const bbox: BoundingBox = {
        x: detection.bbox_x,
        y: detection.bbox_y,
        width: detection.bbox_width,
        height: detection.bbox_height
      }
      setDrawingBbox(bbox)
      dispatch(setSelectedBbox(bbox))
    },
    [dispatch]
  )

  const trackingStatus = activeTask?.status || TrackingStatus.IDLE

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <ScanOutlined style={{ color: '#722ed1' }} />
            AI 视觉识别追踪
          </Title>
          <UAVSelector
            placeholder="选择无人机"
            value={selectedUAVId}
            onChange={val => {
              // UAV selection handled via redux in real app
            }}
            options={uavList.map(uav => ({
              label: uav.name,
              value: uav.id
            }))}
          />
          <Badge status={isConnected ? 'success' : 'error'} text={<span style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12 }}>{isConnected ? '实时连接' : '未连接'}</span>} />
        </HeaderLeft>
        <Space>
          <Tooltip title={autoDetect ? '关闭自动检测' : '开启自动检测'}>
            <Button
              type={autoDetect ? 'primary' : 'default'}
              icon={<EyeOutlined />}
              onClick={() => setAutoDetect(!autoDetect)}
            >
              自动检测: {autoDetect ? '开' : '关'}
            </Button>
          </Tooltip>
          {activeTask?.status === TrackingStatus.TRACKING ||
          activeTask?.status === TrackingStatus.LOCKING ||
          activeTask?.status === TrackingStatus.SEARCHING ? (
            <Button danger icon={<StopOutlined />} onClick={handleStopTracking} loading={loading}>
              停止追踪
            </Button>
          ) : (
            <Button
              type="primary"
              icon={<AimOutlined />}
              onClick={handleLockTarget}
              disabled={!drawingBbox && !selectedBbox}
              loading={loading}
              style={{ background: '#722ed1', borderColor: '#722ed1' }}
            >
              锁定目标
            </Button>
          )}
        </Space>
      </Header>

      <Content>
        <VideoPanel>
          <VideoCard
            title={
              <Space>
                <VideoCameraOutlined />
                实时画面
                {streaming && <Badge status="processing" text={<span style={{ color: '#52c41a', fontSize: 12 }}>直播中</span>} />}
              </Space>
            }
            extra={
              <Space size="small">
                <Tooltip title="检测置信度阈值">
                  <span style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12 }}>置信度: {Math.round(confidenceThreshold * 100)}%</span>
                </Tooltip>
                <Slider
                  min={0.1}
                  max={0.95}
                  step={0.05}
                  value={confidenceThreshold}
                  onChange={setConfidenceThreshold}
                  style={{ width: 100 }}
                />
              </Space>
            }
          >
            <VideoContainer
              ref={videoContainerRef}
              onMouseDown={handleMouseDown}
              onMouseMove={handleMouseMove}
              onMouseUp={handleMouseUp}
              onMouseLeave={() => {
                setDrawingStart(null)
                setDrawingEnd(null)
                dispatch(setIsDrawing(false))
                setMousePos(null)
              }}
            >
              {!streaming ? (
                <VideoPlaceholder>
                  <VideoCameraOutlined style={{ fontSize: 48, opacity: 0.4 }} />
                  <div>视频流未连接</div>
                  <div style={{ fontSize: 12 }}>请确保无人机视频传输正常</div>
                </VideoPlaceholder>
              ) : (
                <VideoCanvas ref={videoCanvasRef} />
              )}
              <OverlayCanvas ref={overlayCanvasRef} />

              {activeTask?.status === TrackingStatus.TRACKING && activeTask?.center_offset_x !== undefined && (
                <Crosshair
                  $x={50 + (activeTask.center_offset_x || 0) * -50}
                  $y={50 + (activeTask.center_offset_y || 0) * -50}
                  $visible={true}
                />
              )}

              {detections
                .filter(d => d.confidence >= confidenceThreshold)
                .map((d, i) => {
                  const isSelected =
                    selectedBbox &&
                    Math.abs(selectedBbox.x - d.bbox_x) < 5 &&
                    Math.abs(selectedBbox.y - d.bbox_y) < 5
                  return (
                    <DetectionBox
                      key={d.id || i}
                      $x={(d.bbox_x / 1280) * 100}
                      $y={(d.bbox_y / 720) * 100}
                      $width={(d.bbox_width / 1280) * 100}
                      $height={(d.bbox_height / 720) * 100}
                      $color={DetectionClassColors[d.class] || '#8c8c8c'}
                      $selected={!!isSelected}
                      onClick={e => {
                        e.stopPropagation()
                        handleDetectFromList(d)
                      }}
                    >
                      <DetectionLabel $color={DetectionClassColors[d.class] || '#8c8c8c'}>
                        {DetectionClassLabels[d.class] || d.class} {Math.round(d.confidence * 100)}%
                      </DetectionLabel>
                      <ConfidenceBar>
                        <ConfidenceFill
                          $width={d.confidence * 100}
                          $color={DetectionClassColors[d.class] || '#8c8c8c'}
                        />
                      </ConfidenceBar>
                    </DetectionBox>
                  )
                })}
            </VideoContainer>
          </VideoCard>
        </VideoPanel>

        <InfoPanel>
          <TrackingStatusCard title={
            <Space>
              {trackingStatus === TrackingStatus.TRACKING ? (
                <ThunderboltOutlined style={{ color: '#52c41a' }} />
              ) : trackingStatus === TrackingStatus.SEARCHING ? (
                <SearchOutlined style={{ color: '#faad14' }} />
              ) : trackingStatus === TrackingStatus.LOST ? (
                <AlertOutlined style={{ color: '#ff4d4f' }} />
              ) : (
                <CompassOutlined />
              )}
              追踪状态
            </Space>
          } $status={trackingStatus}>
            <InfoRow>
              <InfoLabel>当前状态</InfoLabel>
              <InfoValue>
                <Tag color={TrackingStatusColors[trackingStatus] as any}>
                  {TrackingStatusLabels[trackingStatus]}
                </Tag>
              </InfoValue>
            </InfoRow>
            {activeTask && (
              <>
                <InfoRow>
                  <InfoLabel>目标类型</InfoLabel>
                  <InfoValue>
                    <DetectionClassBadge $color={DetectionClassColors[activeTask.target_class] || '#8c8c8c'}>
                      {DetectionClassLabels[activeTask.target_class] || activeTask.target_class}
                    </DetectionClassBadge>
                  </InfoValue>
                </InfoRow>
                <InfoRow>
                  <InfoLabel>可见帧数</InfoLabel>
                  <InfoValue>{activeTask.frames_visible}</InfoValue>
                </InfoRow>
                <InfoRow>
                  <InfoLabel>丢失帧数</InfoLabel>
                  <InfoValue>{activeTask.frames_lost}</InfoValue>
                </InfoRow>
                <InfoRow>
                  <InfoLabel>当前置信度</InfoLabel>
                  <InfoValue>
                    {activeTask.confidence ? `${Math.round(activeTask.confidence * 100)}%` : '-'}
                  </InfoValue>
                </InfoRow>
                {activeTask.status === TrackingStatus.SEARCHING && (
                  <>
                    <InfoRow>
                      <InfoLabel>搜索半径</InfoLabel>
                      <InfoValue>
                        <span style={{ color: '#faad14' }}>{activeTask.search_radius.toFixed(1)} m</span>
                      </InfoValue>
                    </InfoRow>
                    <InfoRow>
                      <InfoLabel>最大搜索半径</InfoLabel>
                      <InfoValue>{activeTask.max_search_radius.toFixed(1)} m</InfoValue>
                    </InfoRow>
                  </>
                )}
                {activeTask.center_offset_x !== undefined && (
                  <>
                    <Divider style={{ margin: '8px 0', borderColor: 'rgba(255,255,255,0.1)' }} />
                    <InfoRow>
                      <InfoLabel>画面偏移 X</InfoLabel>
                      <InfoValue>
                        {activeTask.center_offset_x !== undefined
                          ? `${(activeTask.center_offset_x * 100).toFixed(1)}%`
                          : '-'}
                      </InfoValue>
                    </InfoRow>
                    <InfoRow>
                      <InfoLabel>画面偏移 Y</InfoLabel>
                      <InfoValue>
                        {activeTask.center_offset_y !== undefined
                          ? `${(activeTask.center_offset_y * 100).toFixed(1)}%`
                          : '-'}
                      </InfoValue>
                    </InfoRow>
                  </>
                )}
                {activeTask.target_latitude !== undefined && (
                  <>
                    <Divider style={{ margin: '8px 0', borderColor: 'rgba(255,255,255,0.1)' }} />
                    <InfoRow>
                      <InfoLabel>目标纬度</InfoLabel>
                      <InfoValue>{activeTask.target_latitude?.toFixed(6)}</InfoValue>
                    </InfoRow>
                    <InfoRow>
                      <InfoLabel>目标经度</InfoLabel>
                      <InfoValue>{activeTask.target_longitude?.toFixed(6)}</InfoValue>
                    </InfoRow>
                  </>
                )}
              </>
            )}
          </TrackingStatusCard>

          <InfoCard
            title={
              <Space>
                <EyeOutlined />
                检测结果
                <Badge count={detections.filter(d => d.confidence >= confidenceThreshold).length} size="small" />
              </Space>
            }
          >
            {detectionLoading ? (
              <div style={{ textAlign: 'center', padding: 20 }}>
                <Spin size="small" />
              </div>
            ) : detections.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '20px 0', color: 'rgba(255,255,255,0.4)', fontSize: 12 }}>
                暂无检测结果
              </div>
            ) : (
              detections
                .filter(d => d.confidence >= confidenceThreshold)
                .slice(0, 10)
                .map((d, i) => {
                  const isSelected =
                    selectedBbox &&
                    Math.abs(selectedBbox.x - d.bbox_x) < 5 &&
                    Math.abs(selectedBbox.y - d.bbox_y) < 5
                  return (
                    <DetectionItem key={d.id || i} $selected={!!isSelected} onClick={() => handleDetectFromList(d)}>
                      <Space>
                        <DetectionClassBadge $color={DetectionClassColors[d.class] || '#8c8c8c'}>
                          {DetectionClassLabels[d.class] || d.class}
                        </DetectionClassBadge>
                        <span style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12 }}>
                          {Math.round(d.bbox_width)}×{Math.round(d.bbox_height)}
                        </span>
                      </Space>
                      <Tag color={d.confidence >= 0.8 ? 'success' : d.confidence >= 0.6 ? 'processing' : 'warning'}>
                        {Math.round(d.confidence * 100)}%
                      </Tag>
                    </DetectionItem>
                  )
                })
            )}
          </InfoCard>

          <InfoCard
            title={
              <Space>
                <AimOutlined />
                操作说明
              </Space>
            }
          >
            <div style={{ color: 'rgba(255,255,255,0.6)', fontSize: 12, lineHeight: 1.8 }}>
              <div>1. 在视频画面上 <b style={{ color: '#ff4d4f' }}>按住鼠标拖拽</b> 框选目标</div>
              <div>2. 或点击下方「检测结果」快速选择</div>
              <div>3. 点击 <b style={{ color: '#722ed1' }}>「锁定目标」</b> 开始追踪</div>
              <div>4. 追踪中无人机会自动保持目标画面居中</div>
              <div>5. 目标丢失时自动扩大搜索半径（最大 50m）</div>
            </div>
          </InfoCard>
        </InfoPanel>
      </Content>
    </Container>
  )
}

export default AIVisual
