import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import { useNavigate } from 'react-router-dom'
import { Row, Col, Select, Badge, Tag, Button, Space, Card } from 'antd'
import {
  RocketOutlined,
  DashboardOutlined,
  BulbOutlined,
  ThunderboltOutlined,
  BatteryFullOutlined,
  PoweroffOutlined,
  AlertOutlined
} from '@ant-design/icons'
import ArtificialHorizon from '@/components/ArtificialHorizon'
import { AltitudeGauge, AirspeedGauge, ThrottleGauge, VoltageGauge } from '@/components/GaugeMeter'
import FlightMap from '@/components/FlightMap'
import TelemetryPanel from '@/components/TelemetryPanel'
import AlertPanel from '@/components/AlertPanel'
import ControlPanel from '@/components/ControlPanel'
import BatteryIndicator from '@/components/BatteryIndicator'
import RCChannels from '@/components/RCChannels'
import MotorStatusPanel from '@/components/MotorStatusPanel'
import LinkStatusIndicator from '@/components/LinkStatusIndicator'
import { useUAV } from '@/hooks/useUAV'
import { useTelemetry } from '@/hooks/useTelemetry'
import { useAlert } from '@/hooks/useAlert'
import { formatDateTime, getStatusColor } from '@/utils'
import type { UAVListItem } from '@/types'
import { getBatteryList, getBatteryStatistics, getUnacknowledgedMaintenanceCount } from '@/api/battery'
import { getChargingStatistics } from '@/api/charging'
import type { BatteryStatistics, ChargingStatistics } from '@/types'

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
`

const UAVSelector = styled(Select)`
  width: 200px;

  .ant-select-selector {
    background: rgba(255, 255, 255, 0.1) !important;
    border: 1px solid rgba(255, 255, 255, 0.2) !important;
  }

  .ant-select-selection-item {
    color: #fff !important;
  }
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 20px;
`

const StatusItem = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
`

const StatusLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
`

const StatusValue = styled.span<{ $color?: string }>`
  font-weight: 600;
  color: ${props => props.$color || '#fff'};
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 320px 1fr 360px;
  grid-template-rows: 1fr;
  gap: 16px;
  overflow: hidden;
  min-height: 0;
`

const LeftPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const CenterPanel = styled.div`
  display: grid;
  grid-template-rows: 1fr auto;
  gap: 16px;
  overflow: hidden;
`

const RightPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const PanelCard = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
`

const TopRow = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  grid-template-rows: 280px;
  gap: 16px;
`

const GaugesRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 16px;
  height: 200px;
`

const BottomRow = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  height: 280px;
`

const BatteryOverviewCard = styled(PanelCard)`
  padding: 16px;
`

const BatteryOverviewTitle = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
`

const BatteryStatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 10px;
  margin-bottom: 12px;
`

const BatteryStatItem = styled.div`
  background: rgba(255, 255, 255, 0.05);
  padding: 10px;
  border-radius: 6px;
  text-align: center;
`

const BatteryStatValue = styled.div<{ $color?: string }>`
  font-size: 20px;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  color: ${props => props.$color || '#fff'};
`

const BatteryStatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
`

const BatteryQuickActions = styled.div`
  display: flex;
  gap: 8px;
`

const QuickActionBtn = styled(Button)`
  flex: 1;
  font-size: 12px;
  background: rgba(24, 144, 255, 0.1);
  border: 1px solid rgba(24, 144, 255, 0.3);
  color: #1890ff;

  &:hover {
    background: rgba(24, 144, 255, 0.2) !important;
    border-color: #1890ff !important;
    color: #1890ff !important;
  }
`

const MapContainer = styled(PanelCard)`
  flex: 1;
  min-height: 0;
  position: relative;
`

const FlightTimeBadge = styled.div`
  position: absolute;
  top: 16px;
  right: 16px;
  z-index: 10;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  padding: 8px 16px;
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.2);
`

const FlightTimeLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 2px;
`

const FlightTimeValue = styled.div`
  font-size: 18px;
  font-weight: 700;
  color: #52c41a;
  font-family: 'Courier New', monospace;
`

const Dashboard: React.FC = () => {
  const navigate = useNavigate()
  const { uavList, selectedUAVId, selectCurrentUAV, currentUAV, listLoading } = useUAV()
  const { attitude, altitude, airspeed, throttle, battery, gps, position, trajectory, flightTime } = useTelemetry(selectedUAVId || undefined)
  const { unreadCount } = useAlert(false)
  const [batteryStats, setBatteryStats] = useState<BatteryStatistics | null>(null)
  const [chargingStats, setChargingStats] = useState<ChargingStatistics | null>(null)
  const [maintenanceCount, setMaintenanceCount] = useState(0)

  useEffect(() => {
    if (uavList.length > 0 && !selectedUAVId) {
      selectCurrentUAV(uavList[0].id)
    }
  }, [uavList, selectedUAVId, selectCurrentUAV])

  useEffect(() => {
    const fetchStats = async () => {
      try {
        const [batteryData, chargingData, maintenanceData] = await Promise.all([
          getBatteryStatistics(),
          getChargingStatistics(),
          getUnacknowledgedMaintenanceCount()
        ])
        if (batteryData) setBatteryStats(batteryData)
        if (chargingData) setChargingStats(chargingData)
        if (maintenanceData && maintenanceData.count !== undefined) {
          setMaintenanceCount(maintenanceData.count)
        }
      } catch (e) {
        console.error('Failed to fetch battery stats:', e)
      }
    }
    fetchStats()
    const interval = setInterval(fetchStats, 30000)
    return () => clearInterval(interval)
  }, [])

  const uavOptions = uavList.map((uav: UAVListItem) => ({
    value: uav.id,
    label: (
      <Space>
        <Badge status={uav.status === 'connected' || uav.status === 'armed' || uav.status === 'flying' ? 'processing' : 'default'} color={getStatusColor(uav.status)} />
        <span>{uav.name}</span>
        <Tag color={uav.status === 'flying' ? '#52c41a' : '#8c8c8c'} style={{ marginLeft: 'auto' }}>
          {uav.status === 'flying' ? '飞行中' : uav.status === 'disconnected' ? '离线' : '在线'}
        </Tag>
      </Space>
    )
  }))

  const renderUAVOption = (uav: UAVListItem) => (
    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
        <Badge status={uav.status === 'connected' || uav.status === 'armed' || uav.status === 'flying' ? 'processing' : 'default'} color={getStatusColor(uav.status)} />
        <span>{uav.name}</span>
      </div>
      <Tag color={uav.status === 'flying' ? '#52c41a' : '#8c8c8c'}>
        {uav.status === 'flying' ? '飞行中' : uav.status === 'disconnected' ? '离线' : '在线'}
      </Tag>
    </div>
  )

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <DashboardOutlined style={{ color: '#1890ff' }} />
            飞行控制台
          </Title>
          <UAVSelector
            placeholder="选择无人机"
            value={selectedUAVId}
            onChange={selectCurrentUAV}
            loading={listLoading}
            optionLabelProp="label"
            options={uavList.map((uav: UAVListItem) => ({
              value: uav.id,
              label: renderUAVOption(uav)
            }))}
          />
        </HeaderLeft>

        <HeaderRight>
          {currentUAV && (
            <>
              <StatusItem>
                <StatusLabel>飞行模式:</StatusLabel>
                <StatusValue $color="#1890ff">{currentUAV.mode.toUpperCase()}</StatusValue>
              </StatusItem>
              <StatusItem>
                <StatusLabel>GPS:</StatusLabel>
                <StatusValue $color={gps?.fixType && gps.fixType >= 3 ? '#52c41a' : '#faad14'}>
                  {gps?.satellitesVisible || 0} 颗卫星
                </StatusValue>
              </StatusItem>
              <StatusItem>
                <StatusLabel>更新时间:</StatusLabel>
                <StatusValue>{formatDateTime(Date.now())}</StatusValue>
              </StatusItem>
            </>
          )}
        </HeaderRight>
      </Header>

      <Content>
        <LeftPanel>
          <PanelCard style={{ flex: 1, minHeight: 0, overflow: 'auto' }}>
            <TelemetryPanel uav={currentUAV} />
          </PanelCard>
          <PanelCard style={{ height: 320 }}>
            <ControlPanel showTitle={false} />
          </PanelCard>
        </LeftPanel>

        <CenterPanel>
          <TopRow>
            <PanelCard>
              <ArtificialHorizon
                pitch={attitude?.pitch || 0}
                roll={attitude?.roll || 0}
                heading={position?.heading || 0}
              />
            </PanelCard>
            <PanelCard>
              <BatteryIndicator showTitle={false} showCells={false} />
            </PanelCard>
          </TopRow>

          <GaugesRow>
            <PanelCard>
              <AltitudeGauge value={altitude || 0} />
            </PanelCard>
            <PanelCard>
              <AirspeedGauge value={airspeed || 0} />
            </PanelCard>
            <PanelCard>
              <ThrottleGauge value={throttle || 0} />
            </PanelCard>
            <PanelCard>
              <VoltageGauge value={battery?.voltage || 0} />
            </PanelCard>
          </GaugesRow>

          <MapContainer>
            <FlightTimeBadge>
              <FlightTimeLabel>飞行时间</FlightTimeLabel>
              <FlightTimeValue>
                {flightTime ? `${Math.floor(flightTime / 60)}:${(flightTime % 60).toString().padStart(2, '0')}` : '00:00'}
              </FlightTimeValue>
            </FlightTimeBadge>
            <FlightMap
              uavPosition={position ? {
                lat: position.lat,
                lng: position.lng,
                alt: altitude || 0,
                heading: position.heading || 0
              } : undefined}
              trajectory={trajectory}
              showTrajectory
              showGeofence
              showMission
            />
          </MapContainer>
        </CenterPanel>

        <RightPanel>
          {selectedUAVId && (
            <LinkStatusIndicator uavId={selectedUAVId} showDetails={true} />
          )}
          
          <BatteryOverviewCard>
            <BatteryOverviewTitle>
              <BatteryFullOutlined style={{ color: '#52c41a' }} />
              电池概览
              {maintenanceCount > 0 && (
                <Badge count={maintenanceCount} size="small" style={{ marginLeft: 'auto' }}>
                  <AlertOutlined style={{ color: '#faad14', fontSize: '14px' }} />
                </Badge>
              )}
            </BatteryOverviewTitle>
            <BatteryStatsGrid>
              <BatteryStatItem>
                <BatteryStatValue $color="#1890ff">{batteryStats?.total || 0}</BatteryStatValue>
                <BatteryStatLabel>电池总数</BatteryStatLabel>
              </BatteryStatItem>
              <BatteryStatItem>
                <BatteryStatValue $color="#52c41a">{batteryStats?.charging || 0}</BatteryStatValue>
                <BatteryStatLabel>充电中</BatteryStatLabel>
              </BatteryStatItem>
              <BatteryStatItem>
                <BatteryStatValue $color="#faad14">{batteryStats?.in_use || 0}</BatteryStatValue>
                <BatteryStatLabel>使用中</BatteryStatLabel>
              </BatteryStatItem>
              <BatteryStatItem>
                <BatteryStatValue $color={maintenanceCount > 0 ? '#ff4d4f' : '#8c8c8c'}>
                  {maintenanceCount}
                </BatteryStatValue>
                <BatteryStatLabel>需保养</BatteryStatLabel>
              </BatteryStatItem>
            </BatteryStatsGrid>
            <BatteryQuickActions>
              <QuickActionBtn size="small" onClick={() => navigate('/battery-management')}>
                <BatteryFullOutlined /> 电池管理
              </QuickActionBtn>
              <QuickActionBtn size="small" onClick={() => navigate('/charging-management')}>
                <PoweroffOutlined /> 充电管理
              </QuickActionBtn>
            </BatteryQuickActions>
          </BatteryOverviewCard>

          <PanelCard style={{ flex: 1, minHeight: 0, overflow: 'hidden' }}>
            <AlertPanel showTitle maxItems={8} />
          </PanelCard>
          <PanelCard style={{ height: 260 }}>
            <RCChannels showTitle={false} maxChannels={8} />
          </PanelCard>
          <MotorStatusPanel uavId={selectedUAVId || undefined} />
        </RightPanel>
      </Content>
    </Container>
  )
}

export default Dashboard
