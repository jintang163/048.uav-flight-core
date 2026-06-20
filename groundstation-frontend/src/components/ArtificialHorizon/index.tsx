import React, { useEffect, useRef, useMemo } from 'react'
import ReactECharts from 'echarts-for-react'
import * as echarts from 'echarts'
import styled from 'styled-components'
import { radToDeg } from '@/utils'

const Container = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  background: #000;
  border-radius: 8px;
  overflow: hidden;
`

const DataOverlay = styled.div`
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  pointer-events: none;
  z-index: 10;
  color: white;
  text-align: center;
  font-family: monospace;
`

const ValueDisplay = styled.div`
  font-size: 24px;
  font-weight: bold;
  text-shadow: 0 0 10px rgba(0, 0, 0, 0.8);
  margin: 8px 0;
`

const Label = styled.div`
  font-size: 12px;
  opacity: 0.8;
`

interface ArtificialHorizonProps {
  pitch: number
  roll: number
  heading?: number
  altitude?: number
  airspeed?: number
  climbRate?: number
  throttle?: number
}

const ArtificialHorizon: React.FC<ArtificialHorizonProps> = ({
  pitch,
  roll,
  heading = 0,
  altitude = 0,
  airspeed = 0,
  climbRate = 0,
  throttle = 0
}) => {
  const chartRef = useRef<ReactECharts>(null)
  const pitchDeg = radToDeg(pitch)
  const rollDeg = radToDeg(roll)
  const headingDeg = radToDeg(heading)

  const option = useMemo(() => {
    const pitchOffset = pitchDeg * 2
    const rotation = rollDeg

    const graphicElements: echarts.GraphicComponentOption[] = [
      {
        type: 'group',
        left: 'center',
        top: 'center',
        rotation: (rotation * Math.PI) / 180,
        children: [
          {
            type: 'rect',
            left: -400,
            top: -400 + pitchOffset,
            shape: {
              width: 800,
              height: 800
            },
            style: {
              fill: {
                type: 'linear',
                x: 0,
                y: 0,
                x2: 0,
                y2: 1,
                colorStops: [
                  { offset: 0, color: '#1e3a8a' },
                  { offset: 0.5, color: '#3b82f6' },
                  { offset: 0.5, color: '#8b4513' },
                  { offset: 1, color: '#654321' }
                ]
              }
            }
          },
          {
            type: 'line',
            left: -400,
            top: -300 + pitchOffset,
            shape: {
              x1: 0,
              y1: 0,
              x2: 800,
              y2: 0
            },
            style: {
              stroke: 'rgba(255, 255, 255, 0.6)',
              lineWidth: 2
            }
          },
          ...[-40, -30, -20, -10, 10, 20, 30, 40].map(deg => ({
            type: 'group',
            left: 'center',
            top: 'center',
            children: [
              {
                type: 'line',
                left: -60,
                top: -deg * 2 + pitchOffset - 1,
                shape: {
                  x1: 0,
                  y1: 0,
                  x2: deg % 20 === 0 ? 120 : 80,
                  y2: 0
                },
                style: {
                  stroke: 'rgba(255, 255, 255, 0.8)',
                  lineWidth: deg % 20 === 0 ? 2 : 1
                }
              },
              deg % 20 === 0
                ? {
                    type: 'text',
                    left: 70,
                    top: -deg * 2 + pitchOffset - 8,
                    style: {
                      text: `${deg > 0 ? '+' : ''}${deg}`,
                      fill: 'white',
                      fontSize: 12,
                      fontWeight: 'bold'
                    }
                  }
                : null
            ].filter(Boolean)
          }))
        ]
      },
      {
        type: 'line',
        left: 'center',
        top: 'center',
        shape: {
          x1: -150,
          y1: 0,
          x2: -40,
          y2: 0
        },
        style: {
          stroke: '#ff6b6b',
          lineWidth: 3
        }
      },
      {
        type: 'line',
        left: 'center',
        top: 'center',
        shape: {
          x1: 40,
          y1: 0,
          x2: 150,
          y2: 0
        },
        style: {
          stroke: '#ff6b6b',
          lineWidth: 3
        }
      },
      {
        type: 'line',
        left: 'center',
        top: 'center',
        shape: {
          x1: -15,
          y1: -50,
          x2: -15,
          y2: 50
        },
        style: {
          stroke: '#ff6b6b',
          lineWidth: 2
        }
      },
      {
        type: 'circle',
        left: 'center',
        top: 'center',
        shape: {
          cx: 0,
          cy: 0,
          r: 8
        },
        style: {
          stroke: '#ff6b6b',
          lineWidth: 2,
          fill: 'none'
        }
      }
    ]

    return {
      backgroundColor: 'transparent',
      animation: false,
      graphic: graphicElements,
      series: [],
      grid: {
        show: false
      },
      xAxis: {
        show: false,
        min: -200,
        max: 200
      },
      yAxis: {
        show: false,
        min: -200,
        max: 200
      }
    }
  }, [pitchOffset, rotation])

  useEffect(() => {
    const chart = chartRef.current?.getEchartsInstance()
    if (chart) {
      chart.setOption(option)
    }
  }, [option])

  return (
    <Container>
      <ReactECharts
        ref={chartRef}
        option={option}
        style={{ width: '100%', height: '100%' }}
        notMerge={true}
      />
      <DataOverlay>
        <div style={{ display: 'flex', justifyContent: 'space-between', width: 400, position: 'absolute', left: -200, top: -150 }}>
          <div>
            <ValueDisplay>{airspeed.toFixed(1)}</ValueDisplay>
            <Label>空速 m/s</Label>
          </div>
          <div>
            <ValueDisplay>{altitude.toFixed(0)}</ValueDisplay>
            <Label>高度 m</Label>
          </div>
        </div>
        <div style={{ position: 'absolute', width: 300, left: -150, top: 80 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between' }}>
            <div>
              <ValueDisplay style={{ fontSize: 16 }}>{climbRate.toFixed(1)}</ValueDisplay>
              <Label>爬升 m/s</Label>
            </div>
            <div>
              <ValueDisplay style={{ fontSize: 16 }}>{headingDeg.toFixed(0)}°</ValueDisplay>
              <Label>航向</Label>
            </div>
            <div>
              <ValueDisplay style={{ fontSize: 16 }}>{throttle.toFixed(0)}%</ValueDisplay>
              <Label>油门</Label>
            </div>
          </div>
        </div>
      </DataOverlay>
    </Container>
  )
}

export default ArtificialHorizon
