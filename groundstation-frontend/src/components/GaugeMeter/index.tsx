import React from 'react'
import ReactECharts from 'echarts-for-react'
import styled from 'styled-components'
import { getBatteryColor } from '@/utils'

const Container = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
`

const GaugeTitle = styled.div`
  position: absolute;
  bottom: 10px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.8);
  text-align: center;
`

interface GaugeMeterProps {
  title: string
  value: number
  min?: number
  max?: number
  unit?: string
  warningThreshold?: number
  dangerThreshold?: number
  color?: string
  precision?: number
}

const GaugeMeter: React.FC<GaugeMeterProps> = ({
  title,
  value,
  min = 0,
  max = 100,
  unit = '',
  warningThreshold,
  dangerThreshold,
  color,
  precision = 0
}) => {
  const getGaugeColor = () => {
    if (color) return color
    if (dangerThreshold !== undefined && value >= dangerThreshold) return '#ff4d4f'
    if (warningThreshold !== undefined && value >= warningThreshold) return '#faad14'
    return '#52c41a'
  }

  const gaugeColor = getGaugeColor()

  const option = {
    series: [
      {
        type: 'gauge',
        startAngle: 225,
        endAngle: -45,
        min: min,
        max: max,
        progress: {
          show: true,
          width: 16,
          itemStyle: {
            color: gaugeColor
          }
        },
        axisLine: {
          lineStyle: {
            width: 16,
            color: [[1, 'rgba(255, 255, 255, 0.1)']]
          }
        },
        axisTick: {
          show: false
        },
        splitLine: {
          length: 8,
          lineStyle: {
            width: 2,
            color: 'rgba(255, 255, 255, 0.3)'
          }
        },
        axisLabel: {
          distance: 22,
          color: 'rgba(255, 255, 255, 0.6)',
          fontSize: 10
        },
        pointer: {
          show: false
        },
        anchor: {
          show: false
        },
        title: {
          show: false
        },
        detail: {
          valueAnimation: true,
          width: '60%',
          lineHeight: 40,
          borderRadius: 8,
          offsetCenter: [0, '0%'],
          fontSize: 24,
          fontWeight: 'bold',
          formatter: `{value}${unit}`,
          color: gaugeColor
        },
        data: [
          {
            value: Number(value.toFixed(precision))
          }
        ]
      }
    ]
  }

  return (
    <Container>
      <ReactECharts
        option={option}
        style={{ width: '100%', height: '100%' }}
        notMerge={true}
      />
      <GaugeTitle>{title}</GaugeTitle>
    </Container>
  )
}

export const AltitudeGauge: React.FC<{ value: number }> = ({ value }) => (
  <GaugeMeter
    title="高度"
    value={value}
    min={0}
    max={500}
    unit="m"
    warningThreshold={400}
    dangerThreshold={450}
    precision={0}
  />
)

export const AirspeedGauge: React.FC<{ value: number }> = ({ value }) => (
  <GaugeMeter
    title="空速"
    value={value}
    min={0}
    max={30}
    unit="m/s"
    warningThreshold={20}
    dangerThreshold={25}
    precision={1}
  />
)

export const ThrottleGauge: React.FC<{ value: number }> = ({ value }) => (
  <GaugeMeter
    title="油门"
    value={value}
    min={0}
    max={100}
    unit="%"
    warningThreshold={80}
    dangerThreshold={95}
    precision={0}
  />
)

export const VoltageGauge: React.FC<{ value: number }> = ({ value }) => {
  const batteryColor = getBatteryColor((value / 12.6) * 100)
  return (
    <GaugeMeter
      title="电压"
      value={value}
      min={0}
      max={14.8}
      unit="V"
      warningThreshold={11.1}
      dangerThreshold={10.5}
      color={batteryColor}
      precision={2}
    />
  )
}

export default GaugeMeter
