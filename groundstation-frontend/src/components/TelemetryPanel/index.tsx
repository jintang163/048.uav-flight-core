import React from 'react'
import styled from 'styled-components'
import { Card, Descriptions, Tag, Statistic } from 'antd'
import { AimOutlined, WifiOutlined, ThunderboltOutlined, ClockCircleOutlined } from '@ant-design/icons'
import { formatAltitude, formatSpeed, getBatteryColor, getSignalColor, getStatusColor, formatDateTime } from '@/utils'
import type { UAV } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const StyledCard = styled(Card)`
  .ant-card-body {
    padding: 16px;
  }
`

const StatusTag = styled(Tag)<{ $color: string }>`
  background: ${props => props.$color}20;
  color: ${props => props.$color};
  border-color: ${props => props.$color};
  font-weight: 500;
`

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
`

const StatItem = styled.div`
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
`

interface TelemetryPanelProps {
  uav: UAV | null
}

const TelemetryPanel: React.FC<TelemetryPanelProps> = ({ uav }) => {
  if (!uav) {
    return (
      <Container>
        <StyledCard title="遥测数据" size="small">
          <div style={{ textAlign: 'center', color: 'rgba(255,255,255,0.5)', padding: '40px 0' }}>
            暂无无人机数据
          </div>
        </StyledCard>
      </Container>
    )
  }

  const statusColor = getStatusColor(uav.status)
  const batteryColor = getBatteryColor(uav.battery.remaining)
  const signalColor = getSignalColor(uav.signalQuality)

  return (
    <Container>
      <StyledCard 
        title={
          <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
            <span>{uav.name}</span>
            <StatusTag $color={statusColor}>
              {uav.status.toUpperCase()}
            </StatusTag>
          </div>
        } 
        size="small"
      >
        <StatsRow>
          <StatItem>
            <AimOutlined style={{ fontSize: 20, color: '#1890ff', marginBottom: 4 }} />
            <Statistic 
              title={<span style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>GPS卫星</span>}
              value={uav.gpsSatellites}
              valueStyle={{ fontSize: 18, color: uav.gpsFixType >= 3 ? '#52c41a' : '#faad14' }}
            />
          </StatItem>
          <StatItem>
            <WifiOutlined style={{ fontSize: 20, color: signalColor, marginBottom: 4 }} />
            <Statistic 
              title={<span style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>信号质量</span>}
              value={uav.signalQuality}
              suffix="dBm"
              valueStyle={{ fontSize: 18, color: signalColor }}
            />
          </StatItem>
          <StatItem>
            <ThunderboltOutlined style={{ fontSize: 20, color: batteryColor, marginBottom: 4 }} />
            <Statistic 
              title={<span style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>电池</span>}
              value={uav.battery.remaining}
              suffix="%"
              valueStyle={{ fontSize: 18, color: batteryColor }}
            />
          </StatItem>
          <StatItem>
            <ClockCircleOutlined style={{ fontSize: 20, color: '#13c2c2', marginBottom: 4 }} />
            <Statistic 
              title={<span style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>更新时间</span>}
              value={formatDateTime(uav.lastUpdate).split(' ')[1]}
              valueStyle={{ fontSize: 14 }}
            />
          </StatItem>
        </StatsRow>

        <Descriptions size="small" column={2} style={{ marginTop: 16 }}>
          <Descriptions.Item label="飞行模式">{uav.mode.toUpperCase()}</Descriptions.Item>
          <Descriptions.Item label="锁定状态">{uav.armed ? '已解锁' : '已上锁'}</Descriptions.Item>
          <Descriptions.Item label="飞行高度">{formatAltitude(uav.position.alt)}</Descriptions.Item>
          <Descriptions.Item label="相对高度">{formatAltitude(uav.position.relativeAlt)}</Descriptions.Item>
          <Descriptions.Item label="地速">{formatSpeed(uav.velocity.groundSpeed)}</Descriptions.Item>
          <Descriptions.Item label="空速">{formatSpeed(uav.velocity.airSpeed)}</Descriptions.Item>
          <Descriptions.Item label="俯仰角">{uav.attitude.pitch.toFixed(1)}°</Descriptions.Item>
          <Descriptions.Item label="横滚角">{uav.attitude.roll.toFixed(1)}°</Descriptions.Item>
          <Descriptions.Item label="航向角">{uav.heading.toFixed(1)}°</Descriptions.Item>
          <Descriptions.Item label="爬升率">{uav.velocity.climbRate.toFixed(1)} m/s</Descriptions.Item>
          <Descriptions.Item label="油门">{uav.throttle.toFixed(0)}%</Descriptions.Item>
          <Descriptions.Item label="电压">{uav.battery.voltage.toFixed(2)}V</Descriptions.Item>
        </Descriptions>

        <div style={{ marginTop: 12, fontSize: 11, color: 'rgba(255,255,255,0.5)' }}>
          <div>位置: {uav.position.lat.toFixed(6)}, {uav.position.lng.toFixed(6)}</div>
        </div>
      </StyledCard>
    </Container>
  )
}

export default TelemetryPanel
