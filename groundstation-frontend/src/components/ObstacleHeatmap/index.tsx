import React, { useEffect, useRef, useCallback } from 'react'
import styled from 'styled-components'
import { Tooltip } from 'antd'
import type { ObstacleHeatmapPoint } from '@/types/obstacle-avoidance'

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

const Legend = styled.div`
  position: absolute;
  bottom: 12px;
  right: 12px;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 11px;
  color: rgba(255, 255, 255, 0.8);
`

const LegendTitle = styled.div`
  font-weight: 600;
  margin-bottom: 6px;
  font-size: 12px;
`

const LegendGradient = styled.div`
  width: 120px;
  height: 12px;
  border-radius: 3px;
  margin: 4px 0;
  background: linear-gradient(to right, #52c41a, #faad14, #ff4d4f);
`

const LegendLabels = styled.div`
  display: flex;
  justify-content: space-between;
  font-size: 10px;
  color: rgba(255, 255, 255, 0.6);
`

const StatsOverlay = styled.div`
  position: absolute;
  top: 12px;
  left: 12px;
  background: rgba(0, 0, 0, 0.75);
  backdrop-filter: blur(8px);
  padding: 10px 14px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-size: 11px;
  color: rgba(255, 255, 255, 0.8);
`

const StatsRow = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 4px;

  &:last-child {
    margin-bottom: 0;
  }
`

const StatsDot = styled.div<{ $color: string }>`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: ${props => props.$color};
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

interface ObstacleHeatmapProps {
  points: ObstacleHeatmapPoint[]
  center?: { lat: number; lng: number }
  zoom?: number
  width?: number
  height?: number
}

const getHeatColor = (intensity: number): string => {
  if (intensity < 0.33) {
    const t = intensity / 0.33
    const r = Math.round(82 + (250 - 82) * t)
    const g = Math.round(196 + (173 - 196) * t)
    const b = Math.round(26 + (20 - 26) * t)
    return `rgb(${r}, ${g}, ${b})`
  } else if (intensity < 0.66) {
    const t = (intensity - 0.33) / 0.33
    const r = Math.round(250 + (255 - 250) * t)
    const g = Math.round(173 + (77 - 173) * t)
    const b = Math.round(20 + (79 - 20) * t)
    return `rgb(${r}, ${g}, ${b})`
  } else {
    const t = (intensity - 0.66) / 0.34
    const r = 255
    const g = Math.round(77 - 77 * t)
    const b = Math.round(79 - 79 * t)
    return `rgb(${r}, ${g}, ${b})`
  }
}

const ObstacleHeatmap: React.FC<ObstacleHeatmapProps> = ({
  points,
  center,
  zoom = 15,
  width,
  height
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const animFrameRef = useRef<number>(0)

  const drawHeatmap = useCallback(() => {
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

    const gridSpacing = 30
    ctx.strokeStyle = 'rgba(255, 255, 255, 0.03)'
    ctx.lineWidth = 1
    for (let x = 0; x < w; x += gridSpacing) {
      ctx.beginPath()
      ctx.moveTo(x, 0)
      ctx.lineTo(x, h)
      ctx.stroke()
    }
    for (let y = 0; y < h; y += gridSpacing) {
      ctx.beginPath()
      ctx.moveTo(0, y)
      ctx.lineTo(w, y)
      ctx.stroke()
    }

    if (points.length === 0) return

    const centerLat = center?.lat ?? points.reduce((s, p) => s + p.lat, 0) / points.length
    const centerLng = center?.lng ?? points.reduce((s, p) => s + p.lng, 0) / points.length

    const scale = Math.pow(2, zoom) * 0.5
    const projectX = (lng: number) => w / 2 + (lng - centerLng) * scale * 1000
    const projectY = (lat: number) => h / 2 - (lat - centerLat) * scale * 1000

    const maxIntensity = Math.max(...points.map(p => p.intensity), 0.01)

    for (const point of points) {
      const x = projectX(point.lng)
      const y = projectY(point.lat)
      const normalizedIntensity = point.intensity / maxIntensity
      const baseRadius = 15 + normalizedIntensity * 35
      const gradient = ctx.createRadialGradient(x, y, 0, x, y, baseRadius)

      const color = getHeatColor(normalizedIntensity)
      gradient.addColorStop(0, color.replace('rgb', 'rgba').replace(')', `, ${0.6 + normalizedIntensity * 0.4})`))
      gradient.addColorStop(0.4, color.replace('rgb', 'rgba').replace(')', `, ${0.3 + normalizedIntensity * 0.3})`))
      gradient.addColorStop(1, color.replace('rgb', 'rgba').replace(')', ', 0)'))

      ctx.beginPath()
      ctx.arc(x, y, baseRadius, 0, Math.PI * 2)
      ctx.fillStyle = gradient
      ctx.fill()
    }

    for (const point of points) {
      const x = projectX(point.lng)
      const y = projectY(point.lat)
      const normalizedIntensity = point.intensity / maxIntensity

      if (normalizedIntensity > 0.3) {
        ctx.beginPath()
        ctx.arc(x, y, 3, 0, Math.PI * 2)
        ctx.fillStyle = '#fff'
        ctx.fill()

        ctx.font = '10px monospace'
        ctx.fillStyle = 'rgba(255, 255, 255, 0.8)'
        ctx.fillText(`${point.minDistance.toFixed(1)}m`, x + 6, y - 6)
        ctx.fillText(`×${point.triggerCount}`, x + 6, y + 6)
      }
    }
  }, [points, center, zoom])

  useEffect(() => {
    const animate = () => {
      drawHeatmap()
      animFrameRef.current = requestAnimationFrame(animate)
    }
    animate()
    return () => cancelAnimationFrame(animFrameRef.current)
  }, [drawHeatmap])

  useEffect(() => {
    const handleResize = () => drawHeatmap()
    window.addEventListener('resize', handleResize)
    return () => window.removeEventListener('resize', handleResize)
  }, [drawHeatmap])

  const totalTriggers = points.reduce((s, p) => s + p.triggerCount, 0)
  const nearestDistance = points.length > 0 ? Math.min(...points.map(p => p.minDistance)) : Infinity

  return (
    <Container ref={containerRef}>
      <Canvas ref={canvasRef} style={{ width: width || '100%', height: height || '100%' }} />

      {points.length > 0 && (
        <>
          <StatsOverlay>
            <StatsRow>
              <StatsDot $color="#1890ff" />
              触发点: {points.length}
            </StatsRow>
            <StatsRow>
              <StatsDot $color="#faad14" />
              总触发次数: {totalTriggers}
            </StatsRow>
            <StatsRow>
              <StatsDot $color="#ff4d4f" />
              最近距离: {nearestDistance === Infinity ? '-' : `${nearestDistance.toFixed(1)}m`}
            </StatsRow>
          </StatsOverlay>

          <Legend>
            <LegendTitle>避障热力图</LegendTitle>
            <LegendGradient />
            <LegendLabels>
              <span>低频</span>
              <span>中频</span>
              <span>高频</span>
            </LegendLabels>
          </Legend>
        </>
      )}

      {points.length === 0 && (
        <EmptyHint>暂无避障触发数据</EmptyHint>
      )}
    </Container>
  )
}

export default ObstacleHeatmap
