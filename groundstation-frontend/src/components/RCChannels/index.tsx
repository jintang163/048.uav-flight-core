import React from 'react'
import styled from 'styled-components'
import { Card, Empty } from 'antd'
import { ControlOutlined } from '@ant-design/icons'
import { useTelemetry } from '@/hooks/useTelemetry'
import { useUAV } from '@/hooks/useUAV'
import type { RCChannels as RCChannelsType } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
`

const Title = styled.div`
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
`

const ChannelsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 12px;
  flex: 1;
`

const ChannelCard = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;

  .ant-card-body {
    padding: 12px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
  }
`

const ChannelName = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  text-transform: uppercase;
`

const ChannelValue = styled.div`
  font-size: 18px;
  font-weight: 600;
  color: #1890ff;
  font-family: 'Courier New', monospace;
`

const BarContainer = styled.div`
  width: 100%;
  height: 6px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 3px;
  overflow: hidden;
`

const BarFill = styled.div<{ $percent: number; $color?: string }>`
  height: 100%;
  width: ${props => props.$percent}%;
  background: ${props => props.$color || '#1890ff'};
  border-radius: 3px;
  transition: width 0.1s ease;
`

const ChannelPercent = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
`

const SignalStrength = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
  margin-top: 16px;
`

const SignalBars = styled.div`
  display: flex;
  align-items: flex-end;
  gap: 3px;
  height: 24px;
`

const SignalBar = styled.div<{ $active: boolean; $height: number }>`
  width: 4px;
  height: ${props => props.$height}%;
  background: ${props => props.$active ? '#52c41a' : 'rgba(255,255,255,0.2)'};
  border-radius: 2px;
  transition: background 0.3s;
`

const SignalInfo = styled.div`
  flex: 1;
`

const SignalLabel = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 2px;
`

const SignalValue = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #52c41a;
`

interface RCChannelsProps {
  showTitle?: boolean
  maxChannels?: number
}

const channelNames: Record<number, string> = {
  1: 'ROLL',
  2: 'PITCH',
  3: 'THROTTLE',
  4: 'YAW',
  5: 'AUX1',
  6: 'AUX2',
  7: 'AUX3',
  8: 'AUX4',
  9: 'AUX5',
  10: 'AUX6',
  11: 'AUX7',
  12: 'AUX8',
  13: 'AUX9',
  14: 'AUX10',
  15: 'AUX11',
  16: 'AUX12'
}

const getChannelColor = (channel: number): string => {
  const colors: Record<number, string> = {
    1: '#1890ff',
    2: '#52c41a',
    3: '#faad14',
    4: '#eb2f96',
    5: '#722ed1',
    6: '#13c2c2',
    7: '#fa8c16',
    8: '#2f54eb'
  }
  return colors[channel] || '#8c8c8c'
}

const calculatePercent = (value: number, min: number = 1000, max: number = 2000): number => {
  return Math.max(0, Math.min(100, ((value - min) / (max - min)) * 100))
}

const RCChannels: React.FC<RCChannelsProps> = ({ showTitle = true, maxChannels = 16 }) => {
  const { selectedUAVId } = useUAV()
  const { rcChannels, rssi } = useTelemetry(selectedUAVId || undefined)

  const renderChannel = (index: number, value: number) => {
    const percent = calculatePercent(value)
    const color = getChannelColor(index)

    return (
      <ChannelCard key={index} size="small">
        <ChannelName>{channelNames[index] || `CH${index}`}</ChannelName>
        <ChannelValue style={{ color }}>{value}</ChannelValue>
        <BarContainer>
          <BarFill $percent={percent} $color={color} />
        </BarContainer>
        <ChannelPercent>{percent.toFixed(0)}%</ChannelPercent>
      </ChannelCard>
    )
  }

  const renderSignalBars = () => {
    const bars = 5
    const signalPercent = rssi ? Math.min(100, (rssi / 100) * 100) : 0
    const activeBars = Math.ceil((signalPercent / 100) * bars)

    return (
      <SignalBars>
        {Array.from({ length: bars }, (_, i) => {
          const height = ((i + 1) / bars) * 100
          return <SignalBar key={i} $active={i < activeBars} $height={height} />
        })}
      </SignalBars>
    )
  }

  const channelsArray = rcChannels ? Object.entries(rcChannels)
    .filter(([key]) => key.startsWith('ch'))
    .map(([key, value]) => {
      const index = parseInt(key.replace('ch', ''))
      return { index, value: value as number }
    })
    .filter(c => c.index <= maxChannels)
    .sort((a, b) => a.index - b.index) : []

  return (
    <Container>
      {showTitle && (
        <Title>
          <ControlOutlined style={{ color: '#722ed1' }} />
          遥控器通道
        </Title>
      )}

      {!selectedUAVId ? (
        <Empty
          description="请先选择无人机"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ margin: 'auto' }}
        />
      ) : channelsArray.length === 0 ? (
        <Empty
          description="暂无通道数据"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ margin: 'auto' }}
        />
      ) : (
        <>
          <ChannelsGrid>
            {channelsArray.map(({ index, value }) => renderChannel(index, value))}
          </ChannelsGrid>

          <SignalStrength>
            {renderSignalBars()}
            <SignalInfo>
              <SignalLabel>信号强度 (RSSI)</SignalLabel>
              <SignalValue>{rssi || 0}%</SignalValue>
            </SignalInfo>
          </SignalStrength>
        </>
      )}
    </Container>
  )
}

export default RCChannels
