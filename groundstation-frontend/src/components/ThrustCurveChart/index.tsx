import React, { useRef, useEffect, useState, useCallback } from 'react'
import styled from 'styled-components'
import type { ThrustCurvePoint, ThrustLearningSample } from '@/types/thrust-learning'

const Container = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 6px;
  overflow: hidden;
`

const CanvasWrapper = styled.div`
  width: 100%;
  height: 100%;
`

const Tooltip = styled.div<{ visible: boolean; x: number; y: number }>`
  position: absolute;
  display: ${props => props.visible ? 'block' : 'none'};
  background: rgba(15, 23, 42, 0.95);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 6px;
  padding: 8px 12px;
  color: rgba(255, 255, 255, 0.9);
  font-size: 12px;
  pointer-events: none;
  transform: translate(-50%, -100%);
  margin-top: -8px;
  z-index: 10;
  white-space: nowrap;
`

interface ThrustCurveChartProps {
  curvePoints: ThrustCurvePoint[]
  samples: ThrustLearningSample[]
  hoverThrottle?: number
  estimatedWeight?: number
}

interface HoverInfo {
  visible: boolean
  x: number
  y: number
  throttle?: number
  thrust?: number
  sampleCount?: number
}

const ThrustCurveChart: React.FC<ThrustCurveChartProps> = ({
  curvePoints,
  samples,
  hoverThrottle,
  estimatedWeight
}) => {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const containerRef = useRef<HTMLDivElement>(null)
  const [hoverInfo, setHoverInfo] = useState<HoverInfo>({ visible: false, x: 0, y: 0 })
  const [dims, setDims] = useState({ width: 600, height: 400 })

  const padding = { top: 30, right: 30, bottom: 50, left: 60 }

  useEffect(() => {
    const updateSize = () => {
      if (containerRef.current) {
        const rect = containerRef.current.getBoundingClientRect()
        setDims({ width: rect.width, height: rect.height })
      }
    }
    updateSize()
    window.addEventListener('resize', updateSize)
    return () => window.removeEventListener('resize', updateSize)
  }, [])

  const getThrustMax = useCallback(() => {
    let max = 10
    if (estimatedWeight) {
      max = Math.max(max, estimatedWeight * 9.81 * 2.5)
    }
    curvePoints.forEach(p => {
      max = Math.max(max, p.thrust_n * 1.2)
    })
    return max
  }, [curvePoints, estimatedWeight])

  const draw = useCallback(() => {
    const canvas = canvasRef.current
    if (!canvas) return

    const ctx = canvas.getContext('2d')
    if (!ctx) return

    const dpr = window.devicePixelRatio || 1
    canvas.width = dims.width * dpr
    canvas.height = dims.height * dpr
    ctx.scale(dpr, dpr)

    const width = dims.width
    const height = dims.height
    const plotWidth = width - padding.left - padding.right
    const plotHeight = height - padding.top - padding.bottom
    const thrustMax = getThrustMax()

    ctx.clearRect(0, 0, width, height)

    ctx.strokeStyle = 'rgba(255, 255, 255, 0.06)'
    ctx.lineWidth = 1
    for (let i = 0; i <= 10; i++) {
      const y = padding.top + (plotHeight / 10) * i
      ctx.beginPath()
      ctx.moveTo(padding.left, y)
      ctx.lineTo(padding.left + plotWidth, y)
      ctx.stroke()
    }
    for (let i = 0; i <= 10; i++) {
      const x = padding.left + (plotWidth / 10) * i
      ctx.beginPath()
      ctx.moveTo(x, padding.top)
      ctx.lineTo(x, padding.top + plotHeight)
      ctx.stroke()
    }

    ctx.strokeStyle = 'rgba(255, 255, 255, 0.3)'
    ctx.lineWidth = 1.5
    ctx.beginPath()
    ctx.moveTo(padding.left, padding.top)
    ctx.lineTo(padding.left, padding.top + plotHeight)
    ctx.lineTo(padding.left + plotWidth, padding.top + plotHeight)
    ctx.stroke()

    ctx.fillStyle = 'rgba(255, 255, 255, 0.5)'
    ctx.font = '11px -apple-system, BlinkMacSystemFont, sans-serif'
    ctx.textAlign = 'center'
    for (let i = 0; i <= 10; i++) {
      const x = padding.left + (plotWidth / 10) * i
      const val = (i / 10).toFixed(1)
      ctx.fillText(val, x, padding.top + plotHeight + 20)
    }
    ctx.textAlign = 'right'
    for (let i = 0; i <= 10; i++) {
      const y = padding.top + (plotHeight / 10) * (10 - i)
      const val = ((thrustMax / 10) * i).toFixed(0)
      ctx.fillText(val, padding.left - 8, y + 4)
    }

    ctx.fillStyle = 'rgba(255, 255, 255, 0.7)'
    ctx.font = '12px -apple-system, BlinkMacSystemFont, sans-serif'
    ctx.textAlign = 'center'
    ctx.fillText('油门 (0~1.0)', padding.left + plotWidth / 2, height - 8)
    ctx.save()
    ctx.translate(14, padding.top + plotHeight / 2)
    ctx.rotate(-Math.PI / 2)
    ctx.fillText('推力 (N)', 0, 0)
    ctx.restore()

    if (samples.length > 0) {
      samples.forEach(sample => {
        const x = padding.left + Math.max(0, Math.min(1, sample.throttle)) * plotWidth
        const thrustFromAccel = (sample.accel_z + 9.81) * (estimatedWeight || 2)
        const y = padding.top + plotHeight - Math.max(0, Math.min(thrustMax, thrustFromAccel)) / thrustMax * plotHeight
        ctx.fillStyle = 'rgba(24, 144, 255, 0.4)'
        ctx.beginPath()
        ctx.arc(x, y, 2.5, 0, Math.PI * 2)
        ctx.fill()
      })
    }

    if (curvePoints.length > 1) {
      ctx.strokeStyle = '#52c41a'
      ctx.lineWidth = 2.5
      ctx.beginPath()
      curvePoints.forEach((p, idx) => {
        const x = padding.left + Math.max(0, Math.min(1, p.throttle)) * plotWidth
        const y = padding.top + plotHeight - Math.max(0, Math.min(thrustMax, p.thrust_n)) / thrustMax * plotHeight
        if (idx === 0) {
          ctx.moveTo(x, y)
        } else {
          ctx.lineTo(x, y)
        }
      })
      ctx.stroke()

      curvePoints.forEach(p => {
        const x = padding.left + Math.max(0, Math.min(1, p.throttle)) * plotWidth
        const y = padding.top + plotHeight - Math.max(0, Math.min(thrustMax, p.thrust_n)) / thrustMax * plotHeight
        ctx.fillStyle = '#52c41a'
        ctx.beginPath()
        ctx.arc(x, y, 4, 0, Math.PI * 2)
        ctx.fill()
        ctx.strokeStyle = '#0f172a'
        ctx.lineWidth = 1.5
        ctx.stroke()
      })
    }

    if (hoverThrottle !== undefined && hoverThrottle > 0) {
      const x = padding.left + Math.max(0, Math.min(1, hoverThrottle)) * plotWidth
      const hoverThrust = (estimatedWeight || 2) * 9.81
      const y = padding.top + plotHeight - Math.max(0, Math.min(thrustMax, hoverThrust)) / thrustMax * plotHeight

      ctx.strokeStyle = '#faad14'
      ctx.lineWidth = 1.5
      ctx.setLineDash([5, 5])
      ctx.beginPath()
      ctx.moveTo(x, padding.top)
      ctx.lineTo(x, padding.top + plotHeight)
      ctx.stroke()
      ctx.beginPath()
      ctx.moveTo(padding.left, y)
      ctx.lineTo(padding.left + plotWidth, y)
      ctx.stroke()
      ctx.setLineDash([])

      ctx.fillStyle = '#faad14'
      ctx.beginPath()
      ctx.arc(x, y, 7, 0, Math.PI * 2)
      ctx.fill()
      ctx.strokeStyle = '#fff'
      ctx.lineWidth = 2
      ctx.stroke()

      ctx.fillStyle = '#faad14'
      ctx.font = 'bold 11px -apple-system, BlinkMacSystemFont, sans-serif'
      ctx.textAlign = 'left'
      ctx.fillText(`悬停点: ${(hoverThrottle * 100).toFixed(1)}%`, x + 10, y - 8)
    }
  }, [dims, curvePoints, samples, hoverThrottle, estimatedWeight, getThrustMax])

  useEffect(() => {
    draw()
  }, [draw])

  const handleMouseMove = useCallback((e: React.MouseEvent<HTMLDivElement>) => {
    if (!containerRef.current || curvePoints.length === 0) return
    const rect = containerRef.current.getBoundingClientRect()
    const mx = e.clientX - rect.left
    const my = e.clientY - rect.top
    const plotWidth = dims.width - padding.left - padding.right
    const plotHeight = dims.height - padding.top - padding.bottom
    const thrustMax = getThrustMax()

    if (mx >= padding.left && mx <= padding.left + plotWidth &&
        my >= padding.top && my <= padding.top + plotHeight) {
      const throttle = (mx - padding.left) / plotWidth
      let nearestPoint: ThrustCurvePoint | null = null
      let minDist = Infinity
      curvePoints.forEach(p => {
        const px = padding.left + p.throttle * plotWidth
        const dist = Math.abs(px - mx)
        if (dist < minDist) {
          minDist = dist
          nearestPoint = p
        }
      })

      if (nearestPoint && minDist < 20) {
        const px = padding.left + nearestPoint.throttle * plotWidth
        const py = padding.top + plotHeight - Math.min(thrustMax, nearestPoint.thrust_n) / thrustMax * plotHeight
        setHoverInfo({
          visible: true,
          x: px,
          y: py,
          throttle: nearestPoint.throttle,
          thrust: nearestPoint.thrust_n,
          sampleCount: nearestPoint.sample_count
        })
      } else {
        setHoverInfo({ visible: false, x: mx, y: my })
      }
    } else {
      setHoverInfo(prev => ({ ...prev, visible: false }))
    }
  }, [dims, curvePoints, getThrustMax])

  const handleMouseLeave = useCallback(() => {
    setHoverInfo(prev => ({ ...prev, visible: false }))
  }, [])

  return (
    <Container ref={containerRef} onMouseMove={handleMouseMove} onMouseLeave={handleMouseLeave}>
      <CanvasWrapper>
        <canvas ref={canvasRef} style={{ width: dims.width, height: dims.height }} />
      </CanvasWrapper>
      <Tooltip visible={hoverInfo.visible} x={hoverInfo.x} y={hoverInfo.y}>
        <div>油门: {((hoverInfo.throttle || 0) * 100).toFixed(1)}%</div>
        <div>推力: {(hoverInfo.thrust || 0).toFixed(2)} N</div>
        {hoverInfo.sampleCount !== undefined && (
          <div>样本数: {hoverInfo.sampleCount}</div>
        )}
      </Tooltip>
    </Container>
  )
}

export default ThrustCurveChart
