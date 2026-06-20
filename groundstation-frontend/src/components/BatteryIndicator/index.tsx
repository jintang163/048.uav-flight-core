import React from 'react'
import styled from 'styled-components'
import { Card, Empty, Tooltip } from 'antd'
import { BatteryOutlined, ThunderboltOutlined, WarningOutlined } from '@ant-design/icons'
import { useTelemetry } from '@/hooks/useTelemetry'
import { useUAV } from '@/hooks/useUAV'
import { formatTime } from '@/utils'

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

const MainBattery = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  margin-bottom: 16px;

  .ant-card-body {
    padding: 20px;
  }
`

const BatteryHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`

const BatteryIconWrapper = styled.div<{ $level: number; $charging: boolean }>`
  font-size: 48px;
  color: ${props => {
    if (props.$level <= 15) return '#ff4d4f'
    if (props.$level <= 30) return '#faad14'
    if (props.$charging) return '#52c41a'
    return '#1890ff'
  }};
  display: flex;
  align-items: center;
  position: relative;

  .charging-icon {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    font-size: 20px;
    color: #fff;
  }
`

const BatteryLevel = styled.div`
  text-align: right;
`

const LevelValue = styled.div<{ $level: number }>`
  font-size: 36px;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  color: ${props => {
    if (props.$level <= 15) return '#ff4d4f'
    if (props.$level <= 30) return '#faad14'
    return '#52c41a'
  }};
  line-height: 1;
`

const LevelLabel = styled.div`
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 4px;
`

const BatteryBarContainer = styled.div`
  width: 100%;
  height: 24px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  overflow: hidden;
  position: relative;
  border: 2px solid rgba(255, 255, 255, 0.2);
`

const BatteryBarFill = styled.div<{ $percent: number; $color: string }>`
  height: 100%;
  width: ${props => props.$percent}%;
  background: linear-gradient(90deg, ${props => props.$color} 0%, ${props => props.$color}dd 100%);
  border-radius: 10px;
  transition: width 0.3s ease;
`

const BatteryBarText = styled.div`
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  font-size: 12px;
  font-weight: 600;
  color: #fff;
  text-shadow: 0 1px 2px rgba(0, 0, 0, 0.5);
`

const BatteryInfo = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-top: 16px;
`

const InfoItem = styled.div`
  background: rgba(255, 255, 255, 0.05);
  padding: 12px;
  border-radius: 8px;
`

const InfoLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 4px;
`

const InfoValue = styled.div<{ $color?: string }>`
  font-size: 16px;
  font-weight: 600;
  color: ${props => props.$color || '#fff'};
  font-family: 'Courier New', monospace;
`

const InfoUnit = styled.span`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  margin-left: 4px;
`

const WarningItem = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 77, 79, 0.1);
  border: 1px solid rgba(255, 77, 79, 0.3);
  border-radius: 6px;
  margin-top: 12px;

  .anticon {
    color: #ff4d4f;
  }
`

const WarningText = styled.span`
  font-size: 12px;
  color: #ff4d4f;
`

const CellsContainer = styled.div`
  margin-top: 16px;
`

const CellsTitle = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 8px;
`

const CellsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(60px, 1fr));
  gap: 8px;
`

const CellCard = styled.div<{ $voltage: number; $minVoltage: number; $maxVoltage: number }>`
  background: rgba(255, 255, 255, 0.05);
  padding: 8px;
  border-radius: 6px;
  text-align: center;
  border: 1px solid ${props => {
    const diff = props.$maxVoltage - props.$minVoltage
    if (props.$voltage < 3.5) return 'rgba(255, 77, 79, 0.5)'
    if (diff > 0.1) return 'rgba(250, 173, 20, 0.5)'
    return 'rgba(255, 255, 255, 0.1)'
  }};
`

const CellNumber = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 2px;
`

const CellVoltage = styled.div<{ $voltage: number }>`
  font-size: 13px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: ${props => {
    if (props.$voltage < 3.5) return '#ff4d4f'
    if (props.$voltage < 3.7) return '#faad14'
    return '#52c41a'
  }};
`

const TimeRemaining = styled.div`
  margin-top: 16px;
  padding: 12px;
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.1) 0%, rgba(82, 196, 26, 0.1) 100%);
  border-radius: 8px;
  text-align: center;
`

const TimeLabel = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 4px;
`

const TimeValue = styled.div`
  font-size: 24px;
  font-weight: 700;
  color: #1890ff;
  font-family: 'Courier New', monospace;
`

interface BatteryIndicatorProps {
  showTitle?: boolean
  showCells?: boolean
}

const getBatteryColor = (level: number): string => {
  if (level <= 15) return '#ff4d4f'
  if (level <= 30) return '#faad14'
  return '#52c41a'
}

const getCellVoltage = (voltage: number, totalCells: number): number => {
  return totalCells > 0 ? voltage / totalCells : 0
}

const BatteryIndicator: React.FC<BatteryIndicatorProps> = ({ showTitle = true, showCells = true }) => {
  const { selectedUAVId } = useUAV()
  const { battery } = useTelemetry(selectedUAVId || undefined)

  if (!selectedUAVId) {
    return (
      <Container>
        {showTitle && (
          <Title>
            <BatteryOutlined style={{ color: '#52c41a' }} />
            电池状态
          </Title>
        )}
        <Empty
          description="请先选择无人机"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ margin: 'auto' }}
        />
      </Container>
    )
  }

  if (!battery) {
    return (
      <Container>
        {showTitle && (
          <Title>
            <BatteryOutlined style={{ color: '#52c41a' }} />
            电池状态
          </Title>
        )}
        <Empty
          description="暂无电池数据"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ margin: 'auto' }}
        />
      </Container>
    )
  }

  const { remaining = 0, voltage = 0, current = 0, cells = [], temperature = 0 } = battery
  const cellCount = cells.length || 3
  const charging = false
  const timeRemaining = 0
  const color = getBatteryColor(remaining)
  const avgCellVoltage = getCellVoltage(voltage, cellCount)
  const minCellVoltage = cells.length > 0 ? Math.min(...cells) : avgCellVoltage
  const maxCellVoltage = cells.length > 0 ? Math.max(...cells) : avgCellVoltage
  const cellDiff = maxCellVoltage - minCellVoltage

  const warnings: string[] = []
  if (remaining <= 15) warnings.push('电池电量低')
  if (remaining <= 5) warnings.push('电池电量严重不足，请立即降落！')
  if (avgCellVoltage < 3.5) warnings.push('电芯电压过低')
  if (cellDiff > 0.1) warnings.push(`电芯压差过大 (${cellDiff.toFixed(3)}V)`)
  if (temperature > 60) warnings.push('电池温度过高')

  return (
    <Container>
      {showTitle && (
        <Title>
          <BatteryOutlined style={{ color }} />
          电池状态
          {charging && (
            <Tooltip title="充电中">
              <ThunderboltOutlined style={{ color: '#52c41a', marginLeft: 4 }} />
            </Tooltip>
          )}
        </Title>
      )}

      <MainBattery size="small">
        <BatteryHeader>
          <BatteryIconWrapper $level={remaining} $charging={charging}>
            <BatteryOutlined />
            {charging && <ThunderboltOutlined className="charging-icon" />}
          </BatteryIconWrapper>
          <BatteryLevel>
            <LevelValue $level={remaining}>{remaining.toFixed(0)}%</LevelValue>
            <LevelLabel>剩余电量</LevelLabel>
          </BatteryLevel>
        </BatteryHeader>

        <BatteryBarContainer>
          <BatteryBarFill $percent={remaining} $color={color} />
          <BatteryBarText>{remaining.toFixed(0)}%</BatteryBarText>
        </BatteryBarContainer>

        <BatteryInfo>
          <InfoItem>
            <InfoLabel>总电压</InfoLabel>
            <InfoValue $color="#1890ff">
              {voltage.toFixed(2)}<InfoUnit>V</InfoUnit>
            </InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>平均电压</InfoLabel>
            <InfoValue $color={avgCellVoltage < 3.7 ? '#faad14' : '#52c41a'}>
              {avgCellVoltage.toFixed(3)}<InfoUnit>V/Cell</InfoUnit>
            </InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>放电电流</InfoLabel>
            <InfoValue $color={current > 20 ? '#faad14' : '#1890ff'}>
              {current.toFixed(1)}<InfoUnit>A</InfoUnit>
            </InfoValue>
          </InfoItem>
          <InfoItem>
            <InfoLabel>温度</InfoLabel>
            <InfoValue $color={temperature > 50 ? '#ff4d4f' : '#52c41a'}>
              {temperature.toFixed(0)}<InfoUnit>°C</InfoUnit>
            </InfoValue>
          </InfoItem>
        </BatteryInfo>

        {timeRemaining > 0 && (
          <TimeRemaining>
            <TimeLabel>预计剩余飞行时间</TimeLabel>
            <TimeValue>{formatTime(timeRemaining)}</TimeValue>
          </TimeRemaining>
        )}

        {warnings.length > 0 && warnings.map((warning, index) => (
          <WarningItem key={index}>
            <WarningOutlined />
            <WarningText>{warning}</WarningText>
          </WarningItem>
        ))}

        {showCells && cells.length > 0 && (
          <CellsContainer>
            <CellsTitle>
              电芯电压
              {cellDiff > 0.1 && (
                <span style={{ color: '#faad14', marginLeft: 8 }}>
                  压差: {cellDiff.toFixed(3)}V
                </span>
              )}
            </CellsTitle>
            <CellsGrid>
              {cells.map((cellVoltage, index) => (
                <CellCard
                  key={index}
                  $voltage={cellVoltage}
                  $minVoltage={minCellVoltage}
                  $maxVoltage={maxCellVoltage}
                >
                  <CellNumber>Cell {index + 1}</CellNumber>
                  <CellVoltage $voltage={cellVoltage}>
                    {cellVoltage.toFixed(3)}
                  </CellVoltage>
                </CellCard>
              ))}
            </CellsGrid>
          </CellsContainer>
        )}
      </MainBattery>
    </Container>
  )
}

export default BatteryIndicator
