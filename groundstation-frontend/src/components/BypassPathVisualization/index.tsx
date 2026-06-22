import React, { useEffect, useRef, useCallback } from 'react'
import styled from 'styled-components'
import { Tag } from 'antd'
import type { ObstacleAvoidanceEvent, BypassWaypoint } from '@/types/obstacle-avoidance'
import { STRATEGY_LABELS, ACTION_STATUS_LABELS, ACTION_STATUS_COLORS } from '@/types/obstacle-avoidance'

const Container = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  overflow: hidden;
`

const Canvas = styled.canvas`
  width: 100%;
  height: 100%;
  display: block;
`

const PathInfo = styled.div`
  position: absolute;
  top: 12px;
  right: 12px;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 11px;
  color: rgba(255, 255, 255, 0.8);
  max-width: 200px;
`

const PathInfoTitle = styled.div`
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
`

const PathInfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  margin-bottom: 3px;
  gap: 12px;
`

const EmptyHint = styled.div`
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  color: rgba(255, 255, 255, 0.3);
  font-size: 14px;
  text-align: center;
`

interface BypassPathVisualizationProps {
  event: ObstacleAvoidanceEvent | null
  currentUAVPosition?: { lat: number; lng: number; alt: number }
  animated?: boolean
}

const BypassPathVisualization: React.FC<BypassPathVisualizationProps> = ({
  event,
  currentUAVPosition,
  animated = true
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const animFrameRef = useRef<number>(0)
  const animProgressRef = useRef<number>(0)
  const trailRef = useRef<{ x: number; y: number; alpha: number }[]>([])

  const draw = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = window.devicePixelRatio || 1
    const rect = canvas.getBoundingClientRect()
    canvas.width = rect.width * dpr
    canvas.height = rect.height * dpr
    ctx.scale(dpr, dpr)

    const w = rect.width
    const h = rect.height

    ctx.clearRect(0, 0, w, h)
    ctx.fillStyle = '#0a1628'
    ctx.fillRect(0, 0, w, h)

    if (!event || !event.bypassPath || event.bypassPath.length === 0) {
      ctx.fillStyle = 'rgba(255, 255, 255, 0.15)'
      ctx.font = '14px sans-serif'
      ctx.textAlign = 'center'
      ctx.fillText('暂无绕行路径数据', w / 2, h / 2)
      return
    }

    const allPoints = [
      { lat: event.startPosition.lat, lng: event.startPosition.lng, alt: event.startPosition.alt },
      ...event.bypassPath
    ]

    const centerLat = allPoints.reduce((s, p) => s + p.lat, 0) / allPoints.length
    const centerLng = allPoints.reduce((s, p) => s + p.lng, 0) / allPoints.length

    let maxDist = 0
    for (const p of allPoints) {
      const dx = (p.lng - centerLng) * 111000
      const dy = (p.lat - centerLat) * 111000
      maxDist = Math.max(maxDist, Math.sqrt(dx * dx + dy * dy))
    }
    maxDist = Math.max(maxDist, 10)

    const scale = Math.min(w, h) * 0.35 / maxDist
    const projectX = (lng: number) => w / 2 + (lng - centerLng) * 111000 * scale
    const projectY = (lat: number) => h / 2 - (lat - centerLat) * 111000 * scale

    ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)'
    ctx.lineWidth = 1
    for (let r = 20; r < Math.max(w, h); r += 40) {
      ctx.beginPath()
      ctx.arc(w / 2, h / 2, r, 0, Math.PI * 2)
      ctx.stroke()
    }

    if (animated) {
      animProgressRef.current += 0.008
      if (animProgressRef.current > 1) animProgressRef.current = 0
    }

    const totalPoints = allPoints.length
    const animatedCount = animated
      ? Math.floor(animProgressRef.current * totalPoints) + 1
      : totalPoints
    const visiblePoints = allPoints.slice(0, Math.min(animatedCount, totalPoints))

    for (let i = 0; i < visiblePoints.length - 1; i++) {
      const p1 = visiblePoints[i]
      const p2 = visiblePoints[i + 1]
      const isBypass = i > 0 && event.bypassPath[i - 1]?.type !== 'resume_point'

      ctx.beginPath()
      ctx.moveTo(projectX(p1.lng), projectY(p1.lat))
      ctx.lineTo(projectX(p2.lng), projectY(p2.lat))

      if (isBypass) {
        ctx.strokeStyle = '#722ed1'
        ctx.lineWidth = 3
        ctx.setLineDash([6, 4])
      } else {
        ctx.strokeStyle = '#1890ff'
        ctx.lineWidth = 2
        ctx.setLineDash([])
      }
      ctx.stroke()
      ctx.setLineDash([])
    }

    if (animated && visiblePoints.length > 0) {
      const lastPoint = visiblePoints[visiblePoints.length - 1]
      const lx = projectX(lastPoint.lng)
      const ly = projectY(lastPoint.lat)

      trailRef.current.push({ x: lx, y: ly, alpha: 1 })
      trailRef.current = trailRef.current
        .map(t => ({ ...t, alpha: t.alpha - 0.03 }))
        .filter(t => t.alpha > 0)

      const pulse = Math.sin(Date.now() / 200) * 0.3 + 0.7
      const gradient = ctx.createRadialGradient(lx, ly, 0, lx, ly, 12)
      gradient.addColorStop(0, `rgba(82, 196, 26, ${pulse})`)
      gradient.addColorStop(1, 'rgba(82, 196, 26, 0)')
      ctx.beginPath()
      ctx.arc(lx, ly, 12, 0, Math.PI * 2)
      ctx.fillStyle = gradient
      ctx.fill()

      ctx.beginPath()
      ctx.arc(lx, ly, 4, 0, Math.PI * 2)
      ctx.fillStyle = '#52c41a'
      ctx.fill()

      for (const t of trailRef.current) {
        ctx.beginPath()
        ctx.arc(t.x, t.y, 2, 0, Math.PI * 2)
        ctx.fillStyle = `rgba(82, 196, 26, ${t.alpha * 0.5})`
        ctx.fill()
      }
    }

    for (let i = 0; i < allPoints.length; i++) {
      const p = allPoints[i]
      const x = projectX(p.lng)
      const y = projectY(p.lat)
      const bypassPoint = event.bypassPath[i - 1]

      if (i === 0) {
        ctx.beginPath()
        ctx.arc(x, y, 6, 0, Math.PI * 2)
        ctx.fillStyle = '#ff4d4f'
        ctx.fill()
        ctx.strokeStyle = '#fff'
        ctx.lineWidth = 2
        ctx.stroke()

        ctx.font = 'bold 10px monospace'
        ctx.fillStyle = '#fff'
        ctx.textAlign = 'center'
        ctx.fillText('起', x, y - 10)
      } else if (bypassPoint?.type === 'obstacle_point') {
        ctx.beginPath()
        ctx.arc(x, y, 8, 0, Math.PI * 2)
        ctx.fillStyle = 'rgba(255, 77, 79, 0.3)'
        ctx.fill()
        ctx.beginPath()
        ctx.arc(x, y, 5, 0, Math.PI * 2)
        ctx.fillStyle = '#ff4d4f'
        ctx.fill()

        ctx.font = '10px monospace'
        ctx.fillStyle = '#ff4d4f'
        ctx.textAlign = 'center'
        ctx.fillText('⚠', x, y - 10)
      } else if (bypassPoint?.type === 'bypass_start' || bypassPoint?.type === 'bypass_end') {
        ctx.beginPath()
        ctx.arc(x, y, 5, 0, Math.PI * 2)
        ctx.fillStyle = '#722ed1'
        ctx.fill()
        ctx.strokeStyle = '#fff'
        ctx.lineWidth = 1.5
        ctx.stroke()
      } else if (bypassPoint?.type === 'resume_point') {
        ctx.beginPath()
        ctx.arc(x, y, 6, 0, Math.PI * 2)
        ctx.fillStyle = '#52c41a'
        ctx.fill()
        ctx.strokeStyle = '#fff'
        ctx.lineWidth = 2
        ctx.stroke()

        ctx.font = 'bold 10px monospace'
        ctx.fillStyle = '#fff'
        ctx.textAlign = 'center'
        ctx.fillText('终', x, y - 10)
      } else {
        ctx.beginPath()
        ctx.arc(x, y, 3, 0, Math.PI * 2)
        ctx.fillStyle = '#1890ff'
        ctx.fill()
      }
    }

    const detectionPoint = projectX(event.detection.position.lng)
    const detectionY = projectY(event.detection.position.lat)
    ctx.beginPath()
    ctx.arc(detectionPoint, detectionY, 12, 0, Math.PI * 2)
    ctx.strokeStyle = '#ff4d4f'
    ctx.lineWidth = 2
    ctx.setLineDash([4, 3])
    ctx.stroke()
    ctx.setLineDash([])

    ctx.font = '10px monospace'
    ctx.fillStyle = '#ff4d4f'
    ctx.textAlign = 'center'
    ctx.fillText(`${event.detection.distance.toFixed(1)}m`, detectionPoint, detectionY + 20)
  }, [event, animated])

  useEffect(() => {
    const animate = () => {
      draw()
      animFrameRef.current = requestAnimationFrame(animate)
    }
    animate()
    return () => cancelAnimationFrame(animFrameRef.current)
  }, [draw])

  const bypassPathLength = event?.bypassPath
    ? event.bypassPath.reduce((acc, p, i) => {
        if (i === 0) return 0
        const prev = event.bypassPath[i - 1]
        const dx = (p.lng - prev.lng) * 111000
        const dy = (p.lat - prev.lat) * 111000
        const dz = (p.alt - prev.alt)
        return acc + Math.sqrt(dx * dx + dy * dy + dz * dz)
      }, 0)
    : 0

  const duration = event?.completedAt && event?.timestamp
    ? ((event.completedAt - event.timestamp) / 1000).toFixed(1)
    : '-'

  return (
    <Container>
      <Canvas ref={canvasRef} />

      {event && (
        <PathInfo>
          <PathInfoTitle>
            绕行路径
            <Tag color={ACTION_STATUS_COLORS[event.status]} style={{ fontSize: '10px', lineHeight: '18px', padding: '0 4px' }}>
              {ACTION_STATUS_LABELS[event.status]}
            </Tag>
          </PathInfoTitle>
          <PathInfoRow>
            <span>策略:</span>
            <span>{STRATEGY_LABELS[event.strategy]}</span>
          </PathInfoRow>
          <PathInfoRow>
            <span>路径长度:</span>
            <span>{bypassPathLength.toFixed(1)}m</span>
          </PathInfoRow>
          <PathInfoRow>
            <span>耗时:</span>
            <span>{duration}s</span>
          </PathInfoRow>
          <PathInfoRow>
            <span>障碍距离:</span>
            <span>{event.detection.distance.toFixed(1)}m</span>
          </PathInfoRow>
        </PathInfo>
      )}

      {!event && (
        <EmptyHint>暂无绕行路径数据</EmptyHint>
      )}
    </Container>
  )
}

export default BypassPathVisualization
