import React, { useState, useEffect, useMemo } from 'react'
import styled from 'styled-components'
import {
  Tabs,
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Space,
  Button,
  Select,
  message,
  Badge,
  Divider
} from 'antd'
import {
  CameraOutlined,
  EnvironmentOutlined,
  SoundOutlined,
  RotateLeftOutlined,
  AreaChartOutlined,
  ApiOutlined,
  PlayCircleOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { useUAV } from '@/hooks/useUAV'
import { fetchPayloads, fetchOrbitMissions, fetchOrthoMissions } from '@/store/slices/payload'
import OrbitPlanner from '@/components/OrbitPlanner'
import OrthoPlanner from '@/components/OrthoPlanner'
import PayloadControlPanel from '@/components/PayloadControlPanel'
import FlightMap from '@/components/FlightMap'
import type { PayloadType } from '@/types'
import type { PayloadDevice } from '@/types'

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
  gap: 16px;
`

const Title = styled.div`
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 480px;
  gap: 16px;
  overflow: hidden;
  min-height: 0;
`

const MapContainer = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
  display: flex;
  flex-direction: column;
`

const MapStatsBar = styled.div`
  position: absolute;
  bottom: 16px;
  left: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.2);
`

const StatItem = styled.div`
  flex: 1;
  text-align: center;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 4px;
`

const StatValue = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const SidePanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const PayloadSummary = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-head {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    min-height: 44px;
  }

  .ant-card-body {
    padding: 12px 16px;
  }
`

const PayloadBadge = styled.div<{ type: PayloadType; online: boolean }>`
  display: inline-flex;
  align-items: center;
  gap: 6px;
  padding: 4px 10px;
  border-radius: 12px;
  font-size: 12px;
  background: ${props => props.online ? 'rgba(82, 196, 26, 0.1)' : 'rgba(140, 140, 140, 0.1)'};
  color: ${props => props.online ? '#52c41a' : '#8c8c8c'};
  border: 1px solid ${props => props.online ? 'rgba(82, 196, 26, 0.3)' : 'rgba(140, 140, 140, 0.3)'};
`

const PayloadManagement: React.FC = () => {
  const dispatch = useAppDispatch()
  const { selectedUAVId, currentUAV } = useUAV()
  const { payloads, orbitMissions, orthoMissions, loading, selectedArea, orbitCenter } = useAppSelector(state => state.payload)
  const [activeTab, setActiveTab] = useState('orbit')
  const [cameraPayloadId, setCameraPayloadId] = useState<string | undefined>()

  useEffect(() => {
    if (selectedUAVId) {
      dispatch(fetchPayloads({ uavId: selectedUAVId }))
      dispatch(fetchOrbitMissions({ uavId: selectedUAVId }))
      dispatch(fetchOrthoMissions({ uavId: selectedUAVId }))
    }
  }, [dispatch, selectedUAVId])

  useEffect(() => {
    if (payloads.length > 0 && !cameraPayloadId) {
      const camera = payloads.find(p => p.type === 'camera' || p.type === 'thermal_camera')
      if (camera) {
        setCameraPayloadId(camera.id)
      }
    }
  }, [payloads, cameraPayloadId])

  const uavPosition = useMemo(() => {
    if (currentUAV?.latitude && currentUAV?.longitude) {
      return { lat: currentUAV.latitude, lng: currentUAV.longitude }
    }
    return undefined
  }, [currentUAV])

  const payloadSummary = useMemo(() => {
    const groups: Record<string, { total: number; online: number }> = {
      camera: { total: 0, online: 0 },
      thermal_camera: { total: 0, online: 0 },
      sprayer: { total: 0, online: 0 },
      speaker: { total: 0, online: 0 },
      gripper: { total: 0, online: 0 },
      other: { total: 0, online: 0 }
    }
    payloads.forEach((p: PayloadDevice) => {
      const key = groups[p.type] ? p.type : 'other'
      groups[key].total++
      if (p.status === 'online' || p.status === 'active') {
        groups[key].online++
      }
    })
    return groups
  }, [payloads])

  const getPayloadTypeLabel = (type: PayloadType): string => {
    const map: Record<PayloadType, string> = {
      camera: '可见光相机',
      thermal_camera: '热成像',
      sprayer: '喷药器',
      speaker: '喊话器',
      gripper: '机械爪',
      sensor: '传感器',
      lidar: '激光雷达',
      parachute: '降落伞',
      other: '通用载荷'
    }
    return map[type] || type
  }

  const getPayloadTypeIcon = (type: PayloadType) => {
    switch (type) {
      case 'camera':
      case 'thermal_camera':
        return <CameraOutlined />
      case 'sprayer':
        return <RotateLeftOutlined />
      case 'speaker':
        return <SoundOutlined />
      default:
        return <ApiOutlined />
    }
  }

  const tabItems = [
    {
      key: 'orbit',
      label: (
        <Space>
          <EnvironmentOutlined />
          兴趣点环绕
        </Space>
      ),
      children: selectedUAVId ? (
        <OrbitPlanner
          uavId={selectedUAVId}
          cameraPayloadId={cameraPayloadId}
          uavPosition={uavPosition}
        />
      ) : (
        <div style={{ padding: '40px 20px', textAlign: 'center', color: 'rgba(255,255,255,0.5)' }}>
          <ApiOutlined style={{ fontSize: 48, marginBottom: 12 }} />
          <div>请先在顶部选择无人机</div>
        </div>
      )
    },
    {
      key: 'ortho',
      label: (
        <Space>
          <AreaChartOutlined />
          正射影像采集
        </Space>
      ),
      children: selectedUAVId ? (
        <OrthoPlanner
          uavId={selectedUAVId}
          cameraPayloadId={cameraPayloadId}
          uavPosition={uavPosition}
        />
      ) : (
        <div style={{ padding: '40px 20px', textAlign: 'center', color: 'rgba(255,255,255,0.5)' }}>
          <ApiOutlined style={{ fontSize: 48, marginBottom: 12 }} />
          <div>请先在顶部选择无人机</div>
        </div>
      )
    },
    {
      key: 'control',
      label: (
        <Space>
          <PlayCircleOutlined />
          设备控制
        </Space>
      ),
      children: <PayloadControlPanel uavId={selectedUAVId} />
    }
  ]

  const runningOrbits = orbitMissions.filter(o => o.status === 'running').length
  const runningOrthos = orthoMissions.filter(o => o.status === 'running').length
  const totalRunningMissions = runningOrbits + runningOrthos

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <CameraOutlined style={{ color: '#1890ff' }} />
            载荷任务管理
          </Title>
          <Space size="large">
            <Space>
              <Tag color={payloads.length > 0 ? '#1890ff' : 'default'}>
                载荷设备: {payloads.filter(p => p.status === 'online' || p.status === 'active').length}/{payloads.length} 在线
              </Tag>
            </Space>
            <Space>
              <Badge count={totalRunningMissions} size="small" offset={[-4, 4]}>
                <Tag color={totalRunningMissions > 0 ? 'processing' : 'default'}>
                  执行中任务
                </Tag>
              </Badge>
            </Space>
          </Space>
        </HeaderLeft>

        <Space>
          <Select
            placeholder="选择相机载荷"
            value={cameraPayloadId}
            onChange={setCameraPayloadId}
            style={{ width: 200 }}
            allowClear
            options={payloads
              .filter(p => p.type === 'camera' || p.type === 'thermal_camera')
              .map(p => ({
                value: p.id,
                label: (
                  <Space>
                    {getPayloadTypeIcon(p.type)}
                    <span>{p.name || getPayloadTypeLabel(p.type)}</span>
                    <Badge
                      status={(p.status === 'online' || p.status === 'active') ? 'success' : 'default'}
                      text={p.status === 'online' || p.status === 'active' ? '在线' : '离线'}
                    />
                  </Space>
                )
              }))}
          />
        </Space>
      </Header>

      <Content>
        <MapContainer>
          <FlightMap
            showMission={false}
            showTrajectory
            editable={false}
          />

          <MapStatsBar>
            <StatItem>
              <StatLabel>在线载荷</StatLabel>
              <StatValue>
                {payloads.filter(p => p.status === 'online' || p.status === 'active').length}/{payloads.length}
              </StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>环绕任务</StatLabel>
              <StatValue>
                {runningOrbits} / {orbitMissions.length}
              </StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>正射任务</StatLabel>
              <StatValue>
                {runningOrthos} / {orthoMissions.length}
              </StatValue>
            </StatItem>
            {selectedArea && selectedArea.length >= 3 && (
              <StatItem>
                <StatLabel>测区顶点</StatLabel>
                <StatValue>{selectedArea.length}</StatValue>
              </StatItem>
            )}
            {orbitCenter && (
              <StatItem>
                <StatLabel>环绕中心</StatLabel>
                <StatValue style={{ fontSize: 12 }}>
                  {orbitCenter.lat.toFixed(4)}, {orbitCenter.lng.toFixed(4)}
                </StatValue>
              </StatItem>
            )}
          </MapStatsBar>
        </MapContainer>

        <SidePanel>
          <PayloadSummary size="small" title="载荷概览">
            <Row gutter={[8, 8]}>
              {Object.entries(payloadSummary)
                .filter(([, v]) => v.total > 0)
                .map(([type, info]) => (
                  <Col span={12} key={type}>
                    <PayloadBadge type={type as PayloadType} online={info.online > 0}>
                      {getPayloadTypeIcon(type as PayloadType)}
                      <span>{getPayloadTypeLabel(type as PayloadType)}: {info.online}/{info.total}</span>
                    </PayloadBadge>
                  </Col>
                ))}
              {payloads.length === 0 && (
                <Col span={24}>
                  <div style={{ textAlign: 'center', color: 'rgba(255,255,255,0.5)', padding: '8px 0' }}>
                    暂无载荷设备
                  </div>
                </Col>
              )}
            </Row>
          </PayloadSummary>

          <div style={{ flex: 1, overflow: 'hidden', display: 'flex', flexDirection: 'column' }}>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={tabItems}
              size="small"
              style={{ flex: 1, display: 'flex', flexDirection: 'column' }}
              tabBarStyle={{ marginBottom: 8 }}
            />
          </div>
        </SidePanel>
      </Content>
    </Container>
  )
}

export default PayloadManagement
