import React, { useState, useEffect, useRef, useCallback } from 'react'
import styled from 'styled-components'
import {
  Card,
  Button,
  Space,
  Slider,
  Tag,
  Tooltip,
  List,
  Progress,
  Statistic,
  Row,
  Col
} from 'antd'
import {
  PlayCircleOutlined,
  PauseCircleOutlined,
  FastForwardOutlined,
  RewindOutlined,
  ClockCircleOutlined,
  AlertOutlined,
  ArrowLeftOutlined
} from '@ant-design/icons'
import type { LogDataPoint, LogEvent, LogStatistics } from '@/types/blackbox'
import { formatDuration } from '@/utils'
import ArtificialHorizon from '@/components/ArtificialHorizon'
import GaugeMeter from '@/components/GaugeMeter'
import BatteryIndicator from '@/components/BatteryIndicator'

const Container = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  height: 100%;
  display: flex;
  flex-direction: column;

  .ant-card-body {
    flex: 1;
    display: flex;
    flex-direction: column;
    padding: 16px;
    gap: 16px;
    overflow: hidden;
  }
`

const ReplayHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`

const Title = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
`

const MapContainer = styled.div`
  flex: 1;
  position: relative;
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.05) 0%, rgba(82, 196, 26, 0.05) 100%);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  min-height: 300px;
`

const MapCanvas = styled.canvas`
  width: 100%;
  height: 100%;
`

const MapOverlay = styled.div`
  position: absolute;
  top: 12px;
  left: 12px;
  right: 12px;
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
`

const InfoBadge = styled.div`
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(4px);
  padding: 6px 12px;
  border-radius: 6px;
  color: #fff;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
`

const ControlBar = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 12px;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
`

const PlayButton = styled(Button)`
  width: 48px;
  height: 48px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
`

const TimeDisplay = styled.div`
  min-width: 120px;
  text-align: center;
  font-family: monospace;
  font-size: 14px;
  color: #fff;
`

const SpeedSelector = styled.div`
  display: flex;
  gap: 4px;
`

const SpeedBtn = styled(Button)<{ active: boolean }>`
  ${props => props.active ? `
    background: #1890ff;
    color: #fff;
    border-color: #1890ff;
  ` : ''}
`

const GaugesRow = styled(Row)`
  gap: 12px;
`

const GaugeCard = styled(Card)`
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.08);
  text-align: center;

  .ant-card-body {
    padding: 12px;
  }
`

const GaugeLabel = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 8px;
`

const EventsPanel = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  max-height: 300px;
  overflow-y: auto;

  .ant-card-body {
    padding: 12px;
  }

  .ant-list-item {
    padding: 8px 0;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  }
`

interface FlightReplayProps {
  dataPoints: LogDataPoint[]
  events: LogEvent[]
  statistics: LogStatistics
  title?: string
}

const FlightReplay: React.FC<FlightReplayProps> = ({
  dataPoints,
  events,
  statistics,
  title = '飞行回放'
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentIndex, setCurrentIndex] = useState(0)
  const speedRef = useRef(1)
  const [speedValue, setSpeedValue] = useState(1)
  const animationRef = useRef<number>()
  const lastTimeRef = useRef<number>(0)

  const currentPoint = dataPoints[currentIndex] || null

  const totalDuration = dataPoints.length > 0
    ? (dataPoints[dataPoints.length - 1]?.timestamp || 0) / 1000
    : 0

  const drawMap = useCallback(() => {
    const canvas = canvasRef.current
    const container = containerRef.current
    if (!canvas || !container || dataPoints.length === 0) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const rect = container.getBoundingClientRect()
    canvas.width = rect.width * window.devicePixelRatio
    canvas.height = rect.height * window.devicePixelRatio
    ctx.scale(window.devicePixelRatio, window.devicePixelRatio)

    const width = rect.width
    const height = rect.height
    const padding = 40

    let minLat = Infinity
    let maxLat = -Infinity
    let minLon = Infinity
    let maxLon = -Infinity

    dataPoints.forEach(p => {
      if (p.latitude < minLat) minLat = p.latitude
      if (p.latitude > maxLat) maxLat = p.latitude
      if (p.longitude < minLon) minLon = p.longitude
      if (p.longitude > maxLon) maxLon = p.longitude
    })

    if (minLat === maxLat) {
      minLat -= 0.001
      maxLat += 0.001
    }
    if (minLon === maxLon) {
      minLon -= 0.001
      maxLon += 0.001
    }

    const latRange = maxLat - minLat
    const lonRange = maxLon - minLon
    const scale = Math.min((width - padding * 2) / lonRange, (height - padding * 2) / latRange)

    const latToY = (lat: number) => height - padding - (lat - minLat) * scale
    const lonToX = (lon: number) => padding + (lon - minLon) * scale

    ctx.clearRect(0, 0, width, height)

    ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)'
    ctx.lineWidth = 1
    const gridCount = 5
    for (let i = 0; i <= gridCount; i++) {
      const x = padding + (width - padding * 2) * (i / gridCount)
      const y = padding + (height - padding * 2) * (i / gridCount)
      ctx.beginPath()
      ctx.moveTo(x, padding)
      ctx.lineTo(x, height - padding)
      ctx.stroke()
      ctx.beginPath()
      ctx.moveTo(padding, y)
      ctx.lineTo(width - padding, y)
      ctx.stroke()
    }

    if (dataPoints.length > 1) {
      const gradient = ctx.createLinearGradient(
        lonToX(minLon), 0,
        lonToX(maxLon), 0
      )
      gradient.addColorStop(0, '#1890ff')
      gradient.addColorStop(0.5, '#52c41a')
      gradient.addColorStop(1, '#fa8c16')

      ctx.strokeStyle = gradient
      ctx.lineWidth = 2
      ctx.beginPath()
      dataPoints.forEach((p, i) => {
        const x = lonToX(p.longitude)
        const y = latToY(p.latitude)
        if (i === 0) {
          ctx.moveTo(x, y)
        } else {
          ctx.lineTo(x, y)
        }
      })
      ctx.stroke()
    }

    if (currentPoint) {
      const x = lonToX(currentPoint.longitude)
      const y = latToY(currentPoint.latitude)

      ctx.beginPath()
      ctx.arc(x, y, 20, 0, Math.PI * 2)
      ctx.fillStyle = 'rgba(24, 144, 255, 0.2)'
      ctx.fill()

      ctx.beginPath()
      ctx.arc(x, y, 8, 0, Math.PI * 2)
      ctx.fillStyle = '#1890ff'
      ctx.fill()
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 2
      ctx.stroke()

      ctx.save()
      ctx.translate(x, y)
      ctx.rotate(currentPoint.yaw * Math.PI / 180)
      ctx.beginPath()
      ctx.moveTo(0, -12)
      ctx.lineTo(8, 8)
      ctx.lineTo(0, 4)
      ctx.lineTo(-8, 8)
      ctx.closePath()
      ctx.fillStyle = '#fff'
      ctx.fill()
      ctx.restore()
    }

    if (dataPoints.length > 0) {
      const start = dataPoints[0]
      const end = dataPoints[dataPoints.length - 1]

      ctx.fillStyle = '#52c41a'
      ctx.beginPath()
      ctx.arc(lonToX(start.longitude), latToY(start.latitude), 6, 0, Math.PI * 2)
      ctx.fill()
      ctx.fillStyle = '#fff'
      ctx.font = '10px sans-serif'
      ctx.textAlign = 'center'
      ctx.fillText('起', lonToX(start.longitude), latToY(start.latitude) + 3)

      ctx.fillStyle = '#ff4d4f'
      ctx.beginPath()
      ctx.arc(lonToX(end.longitude), latToY(end.latitude), 6, 0, Math.PI * 2)
      ctx.fill()
      ctx.fillStyle = '#fff'
      ctx.font = '10px sans-serif'
      ctx.fillText('终', lonToX(end.longitude), latToY(end.latitude) + 3)
    }
  }, [dataPoints, currentPoint])

  useEffect(() => {
    drawMap()
  }, [drawMap])

  useEffect(() => {
    const handleResize = () => drawMap()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [drawMap])

  useEffect(() => {
    if (!isPlaying || dataPoints.length === 0) {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current)
      }
      return
    }

    const animate = (timestamp: number) => {
      if (lastTimeRef.current === 0) {
        lastTimeRef.current = timestamp
      }

      const delta = timestamp - lastTimeRef.current
      lastTimeRef.current = timestamp

      const sampleInterval = 100
      const increment = Math.floor((delta * speedRef.current * 10) / sampleInterval)

      setCurrentIndex(prev => {
        const next = prev + increment
        if (next >= dataPoints.length - 1) {
          setIsPlaying(false)
          return dataPoints.length - 1
        }
        return Math.min(next, dataPoints.length - 1)
      })

      animationRef.current = requestAnimationFrame(animate)
    }

    animationRef.current = requestAnimationFrame(animate)

    return () => {
      if (animationRef.current) {
        cancelAnimationFrame(animationRef.current)
      }
    }
  }, [isPlaying, dataPoints.length])

  const togglePlay = () => {
    if (currentIndex >= dataPoints.length - 1) {
      setCurrentIndex(0)
    }
    lastTimeRef.current = 0
    setIsPlaying(!isPlaying)
  }

  const handleSliderChange = (value: number) => {
    setCurrentIndex(Math.floor(value / 100 * (dataPoints.length - 1)))
  }

  const handleSpeedChange = (speed: number) => {
    speedRef.current = speed
    setSpeedValue(speed)
  }

  const currentTime = currentPoint ? currentPoint.timestamp / 1000 : 0

  const nearbyEvents = events.filter(e => {
    const diff = Math.abs(e.timestamp - (currentPoint?.timestamp || 0))
    return diff < 5000
  })

  const getSeverityColor = (severity: number) => {
    switch (severity) {
      case 3: return 'red'
      case 2: return 'orange'
      default: return 'blue'
    }
  }

  return (
    <Container>
      <ReplayHeader>
        <Title>
          <PlayCircleOutlined style={{ color: '#1890ff' }} />
          {title}
        </Title>
        <Space>
          <Tag color="blue">{dataPoints.length} 个数据点</Tag>
          <Tag color="orange">{events.length} 个事件</Tag>
        </Space>
      </ReplayHeader>

      <MapContainer ref={containerRef}>
        <MapCanvas ref={canvasRef} />
        <MapOverlay>
          <InfoBadge>
            <span style={{ color: '#52c41a' }}>●</span>
            起点
          </InfoBadge>
          <InfoBadge>
            <span style={{ color: '#ff4d4f' }}>●</span>
            终点
          </InfoBadge>
          {currentPoint && (
            <>
              <InfoBadge>
                高度: {currentPoint.altitude.toFixed(1)} m
              </InfoBadge>
              <InfoBadge>
                速度: {Math.sqrt(currentPoint.vx ** 2 + currentPoint.vy ** 2 + currentPoint.vz ** 2).toFixed(1)} m/s
              </InfoBadge>
            </>
          )}
        </MapOverlay>
      </MapContainer>

      <ControlBar>
        <Button
          icon={<RewindOutlined />}
          onClick={() => setCurrentIndex(0)}
          disabled={dataPoints.length === 0}
        />
        <PlayButton
          type="primary"
          shape="circle"
          icon={isPlaying ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
          onClick={togglePlay}
          disabled={dataPoints.length === 0}
        />
        <Button
          icon={<FastForwardOutlined />}
          onClick={() => setCurrentIndex(dataPoints.length - 1)}
          disabled={dataPoints.length === 0}
        />

        <div style={{ flex: 1, margin: '0 16px' }}>
          <Slider
            min={0}
            max={100}
            value={dataPoints.length > 1 ? (currentIndex / (dataPoints.length - 1)) * 100 : 0}
            onChange={handleSliderChange}
            tooltip={{
              formatter: () => formatDuration(currentTime)
            }}
          />
        </div>

        <TimeDisplay>
          {formatDuration(currentTime)} / {formatDuration(totalDuration)}
        </TimeDisplay>

        <SpeedSelector>
          {[0.5, 1, 2, 4].map(s => (
            <SpeedBtn
              key={s}
              size="small"
              active={speedValue === s}
              onClick={() => handleSpeedChange(s)}
            >
              {s}x
            </SpeedBtn>
          ))}
        </SpeedSelector>
      </ControlBar>

      <Row gutter={12}>
        <Col span={6}>
          <GaugeCard>
            <ArtificialHorizon
              pitch={currentPoint?.pitch || 0}
              roll={currentPoint?.roll || 0}
              size={120}
            />
            <GaugeLabel>姿态</GaugeLabel>
          </GaugeCard>
        </Col>
        <Col span={6}>
          <GaugeCard>
            <GaugeMeter
              value={currentPoint?.altitude || 0}
              max={Math.max(statistics.max_altitude || 50)}
              label="高度"
              unit="m"
              size={120}
            />
            <GaugeLabel>高度</GaugeLabel>
          </GaugeCard>
        </Col>
        <Col span={6}>
          <GaugeCard>
            <GaugeMeter
              value={currentPoint ? Math.sqrt(currentPoint.vx ** 2 + currentPoint.vy ** 2) : 0}
              max={Math.max(statistics.max_speed || 20)}
              label="速度"
              unit="m/s"
              size={120}
              color="#52c41a"
            />
            <GaugeLabel>速度</GaugeLabel>
          </GaugeCard>
        </Col>
        <Col span={6}>
          <GaugeCard>
            <BatteryIndicator
              voltage={currentPoint?.voltage || 0}
              percentage={currentPoint?.throttle || 0}
              size={120}
            />
            <GaugeLabel>电池</GaugeLabel>
          </GaugeCard>
        </Col>
      </Row>

      <EventsPanel title={
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          <AlertOutlined style={{ color: '#fa8c16' }} />
          <span style={{ color: '#fff', fontWeight: 500, fontSize: 14 }}>事件时间线</span>
          <Tag color="orange" style={{ marginLeft: 'auto' }}>{events.length}</Tag>
        </div>
      }>
        <List
          size="small"
          dataSource={events.slice(0, 10)}
          locale={{ emptyText: '暂无事件' }}
          renderItem={(event) => (
            <List.Item key={event.timestamp}>
              <List.Item.Meta
                avatar={
                  <Tag color={getSeverityColor(event.severity)}>
                    {event.severity === 3 ? '严重' : event.severity === 2 ? '警告' : '提示'}
                  </Tag>
                }
                title={
                  <Space size={8}>
                    <span style={{ color: '#fff', fontSize: 13 }}>{event.description}</span>
                    <Tag color="blue" style={{ fontSize: 11 }}>
                      {formatDuration(event.timestamp / 1000)}
                    </Tag>
                  </Space>
                }
                description={
                  <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>
                    {event.event_type}
                  </span>
                }
              />
            </List.Item>
          )}
        />
      </EventsPanel>
    </Container>
  )
}

export default FlightReplay
