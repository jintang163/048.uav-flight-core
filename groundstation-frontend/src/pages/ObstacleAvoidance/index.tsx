import React, { useEffect, useState } from 'react'
import styled from 'styled-components'
import { Row, Col, Select, Tabs, Card, Statistic, Tag, Button, Space, Badge, message, Popconfirm } from 'antd'
import {
  SafetyCertificateOutlined,
  WarningOutlined,
  ThunderboltOutlined,
  EnvironmentOutlined,
  ReloadOutlined,
  DeleteOutlined,
  ApiOutlined
} from '@ant-design/icons'
import ObstacleHeatmap from '@/components/ObstacleHeatmap'
import BypassPathVisualization from '@/components/BypassPathVisualization'
import ObstacleAvoidanceConfigComponent from '@/components/ObstacleAvoidanceConfig'
import ObstacleAvoidanceLogPanel from '@/components/ObstacleAvoidanceLog'
import { useObstacleAvoidance } from '@/hooks/useObstacleAvoidance'
import { useUAV } from '@/hooks/useUAV'
import { STRATEGY_LABELS, ACTION_STATUS_LABELS, ACTION_STATUS_COLORS } from '@/types/obstacle-avoidance'

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
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
`

const StatusBadge = styled.div<{ $enabled: boolean }>`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: ${props => props.$enabled ? 'rgba(82, 196, 26, 0.1)' : 'rgba(255, 77, 79, 0.1)'};
  border-radius: 16px;
  border: 1px solid ${props => props.$enabled ? 'rgba(82, 196, 26, 0.3)' : 'rgba(255, 77, 79, 0.3)'};
  font-size: 12px;
  color: ${props => props.$enabled ? '#52c41a' : '#ff4d4f'};
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 1fr 320px;
  grid-template-rows: 1fr;
  gap: 16px;
  overflow: hidden;
  min-height: 0;
`

const Panel = styled.div`
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
  display: flex;
  flex-direction: column;
`

const PanelHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 10px 14px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
  background: rgba(255, 255, 255, 0.02);
`

const PanelTitle = styled.div`
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  gap: 8px;
`

const PanelBody = styled.div`
  flex: 1;
  overflow: hidden;
  min-height: 0;
`

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
`

const StatCard = styled(PanelCard)`
  padding: 12px;
  text-align: center;
`

const StatValue = styled.div<{ $color?: string }>`
  font-size: 22px;
  font-weight: 700;
  color: ${props => props.$color || '#fff'};
  font-family: 'Courier New', monospace;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
`

const ActiveEventBanner = styled.div<{ $status: string }>`
  background: ${props => {
    if (props.$status === 'avoiding' || props.$status === 'bypassing') return 'rgba(250, 140, 22, 0.1)'
    if (props.$status === 'triggered') return 'rgba(255, 77, 79, 0.1)'
    return 'rgba(24, 144, 255, 0.1)'
  }};
  border: 1px solid ${props => {
    if (props.$status === 'avoiding' || props.$status === 'bypassing') return 'rgba(250, 140, 22, 0.3)'
    if (props.$status === 'triggered') return 'rgba(255, 77, 79, 0.3)'
    return 'rgba(24, 144, 255, 0.3)'
  }};
  border-radius: 8px;
  padding: 10px 16px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
`

const EventInfo = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
`

const ObstacleAvoidance: React.FC = () => {
  const { uavList, selectedUAVId, selectCurrentUAV, currentUAV } = useUAV()
  const {
    currentDetections,
    activeAvoidanceEvent,
    recentAvoidanceEvents,
    heatmapPoints,
    enabled,
    sensitivity,
    strategy,
    config,
    logs,
    statistics,
    loading,
    saveConfig,
    fetchHeatmap,
    fetchLogs,
    fetchStatistics,
    handleEnabledChange,
    handleSensitivityChange,
    handleStrategyChange,
    handleSensorTypeChange,
    handleDetectionRangeChange,
    handleAscendHeightChange,
    handleRetreatDistanceChange,
    handleBypassAngleChange,
    handleClearHeatmap
  } = useObstacleAvoidance(selectedUAVId || undefined)

  const [activeTab, setActiveTab] = useState('heatmap')

  useEffect(() => {
    fetchHeatmap({ uavId: selectedUAVId })
    fetchLogs({ uavId: selectedUAVId, pageSize: 50 })
    fetchStatistics({ uavId: selectedUAVId })
  }, [selectedUAVId, fetchHeatmap, fetchLogs, fetchStatistics])

  useEffect(() => {
    const interval = setInterval(() => {
      fetchHeatmap({ uavId: selectedUAVId })
      fetchStatistics({ uavId: selectedUAVId })
    }, 5000)
    return () => clearInterval(interval)
  }, [selectedUAVId, fetchHeatmap, fetchStatistics])

  const handleApply = async () => {
    if (!selectedUAVId) return
    try {
      await saveConfig(selectedUAVId)
      message.success('避障配置已应用')
    } catch {
      message.error('配置应用失败')
    }
  }

  const handleReset = () => {
    if (!selectedUAVId) return
    handleEnabledChange(true)
    handleSensitivityChange('medium')
    handleStrategyChange('ascend_bypass')
    message.info('配置已重置')
  }

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <SafetyCertificateOutlined style={{ color: '#1890ff' }} />
            空中避障与感知
          </Title>
          <Select
            placeholder="选择无人机"
            value={selectedUAVId}
            onChange={selectCurrentUAV}
            style={{ width: 200 }}
            options={uavList.map(uav => ({
              value: uav.id,
              label: (
                <Space>
                  <Badge status={uav.status === 'flying' ? 'processing' : 'default'} />
                  <span>{uav.name}</span>
                </Space>
              )
            }))}
          />
          <StatusBadge $enabled={enabled}>
            <ApiOutlined />
            {enabled ? '避障已启用' : '避障已关闭'}
          </StatusBadge>
        </HeaderLeft>
        <HeaderRight>
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => {
              fetchHeatmap({ uavId: selectedUAVId })
              fetchLogs({ uavId: selectedUAVId, pageSize: 50 })
              fetchStatistics({ uavId: selectedUAVId })
            }}
          >
            刷新
          </Button>
          <Popconfirm
            title="确定清除热力图数据？"
            onConfirm={() => handleClearHeatmap(selectedUAVId)}
          >
            <Button size="small" icon={<DeleteOutlined />} danger>
              清除热力图
            </Button>
          </Popconfirm>
        </HeaderRight>
      </Header>

      {activeAvoidanceEvent && (
        <ActiveEventBanner $status={activeAvoidanceEvent.status}>
          <EventInfo>
            <WarningOutlined style={{ color: ACTION_STATUS_COLORS[activeAvoidanceEvent.status], fontSize: '18px' }} />
            <span>避障事件进行中</span>
            <Tag color={ACTION_STATUS_COLORS[activeAvoidanceEvent.status]}>
              {ACTION_STATUS_LABELS[activeAvoidanceEvent.status]}
            </Tag>
            <Tag>{STRATEGY_LABELS[activeAvoidanceEvent.strategy]}</Tag>
            <span>障碍距离: {activeAvoidanceEvent.detection.distance.toFixed(1)}m</span>
          </EventInfo>
        </ActiveEventBanner>
      )}

      <StatsRow>
        <StatCard>
          <StatValue $color="#1890ff">{statistics?.totalDetections ?? currentDetections.length}</StatValue>
          <StatLabel>检测次数</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue $color="#52c41a">{statistics?.successfulAvoidances ?? recentAvoidanceEvents.filter(e => e.status === 'completed').length}</StatValue>
          <StatLabel>成功避障</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue $color="#ff4d4f">{statistics?.failedAvoidances ?? recentAvoidanceEvents.filter(e => e.status === 'failed').length}</StatValue>
          <StatLabel>失败次数</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue $color="#faad14">{statistics?.nearestObstacleDistance?.toFixed(1) ?? '-'}</StatValue>
          <StatLabel>最近距离(m)</StatLabel>
        </StatCard>
      </StatsRow>

      <Content>
        <Panel>
          <PanelCard style={{ flex: 1, minHeight: 0 }}>
            <PanelHeader>
              <PanelTitle>
                <EnvironmentOutlined style={{ color: '#1890ff' }} />
                {activeTab === 'heatmap' ? '避障触发点热力图' : '绕行路径可视化'}
              </PanelTitle>
              <Space size="small">
                <Button
                  size="small"
                  type={activeTab === 'heatmap' ? 'primary' : 'text'}
                  onClick={() => setActiveTab('heatmap')}
                >
                  热力图
                </Button>
                <Button
                  size="small"
                  type={activeTab === 'path' ? 'primary' : 'text'}
                  onClick={() => setActiveTab('path')}
                >
                  绕行路径
                </Button>
              </Space>
            </PanelHeader>
            <PanelBody>
              {activeTab === 'heatmap' ? (
                <ObstacleHeatmap
                  points={heatmapPoints}
                  center={currentUAV ? { lat: currentUAV.position?.lat ?? 0, lng: currentUAV.position?.lng ?? 0 } : undefined}
                />
              ) : (
                <BypassPathVisualization
                  event={activeAvoidanceEvent || recentAvoidanceEvents[0] || null}
                  currentUAVPosition={currentUAV?.position ? { lat: currentUAV.position.lat, lng: currentUAV.position.lng, alt: currentUAV.position.alt } : undefined}
                />
              )}
            </PanelBody>
          </PanelCard>
        </Panel>

        <Panel>
          <PanelCard style={{ flex: 1, minHeight: 0 }}>
            <PanelHeader>
              <PanelTitle>
                <ThunderboltOutlined style={{ color: '#faad14' }} />
                避障日志
              </PanelTitle>
              <Badge count={logs.length} overflowCount={99} style={{ background: '#1890ff' }} />
            </PanelHeader>
            <PanelBody style={{ padding: '0 4px' }}>
              <ObstacleAvoidanceLogPanel logs={logs} maxItems={50} />
            </PanelBody>
          </PanelCard>
        </Panel>

        <Panel>
          <PanelCard style={{ flex: 1, minHeight: 0, overflow: 'auto' }}>
            <PanelBody style={{ padding: '14px' }}>
              <ObstacleAvoidanceConfigComponent
                enabled={enabled}
                sensitivity={sensitivity}
                strategy={strategy}
                sensorType={config?.sensorType}
                detectionRange={config?.detectionRange}
                ascendHeight={config?.ascendHeight}
                retreatDistance={config?.retreatDistance}
                bypassAngle={config?.bypassAngle}
                onEnabledChange={handleEnabledChange}
                onSensitivityChange={handleSensitivityChange}
                onStrategyChange={handleStrategyChange}
                onSensorTypeChange={handleSensorTypeChange}
                onDetectionRangeChange={handleDetectionRangeChange}
                onAscendHeightChange={handleAscendHeightChange}
                onRetreatDistanceChange={handleRetreatDistanceChange}
                onBypassAngleChange={handleBypassAngleChange}
                onApply={handleApply}
                onReset={handleReset}
              />
            </PanelBody>
          </PanelCard>
        </Panel>
      </Content>
    </Container>
  )
}

export default ObstacleAvoidance
