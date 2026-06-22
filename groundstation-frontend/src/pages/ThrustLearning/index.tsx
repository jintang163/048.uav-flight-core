import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import { Select, Button, Space, Badge, Table, message, Tag } from 'antd'
import {
  RadarChartOutlined,
  ReloadOutlined,
  ApiOutlined,
  ThunderboltOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  ExperimentOutlined
} from '@ant-design/icons'
import ThrustCurveChart from '@/components/ThrustCurveChart'
import PIDGainsPanel from '@/components/PIDGainsPanel'
import LearningStatusPanel from '@/components/LearningStatusPanel'
import { useThrustLearning } from '@/hooks/useThrustLearning'
import { useUAV } from '@/hooks/useUAV'
import { LEARNING_STATE_LABELS, LEARNING_STATE_COLORS } from '@/types/thrust-learning'
import type { ThrustLearningSample } from '@/types/thrust-learning'

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

const StatusBadge = styled.div<{ $color: string }>`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: ${props => `${props.$color}15`};
  border-radius: 16px;
  border: 1px solid ${props => `${props.$color}40`};
  font-size: 12px;
  color: ${props => props.$color};
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 10px;
`

const StatCard = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
`

const StatIcon = styled.div<{ $color: string }>`
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin: 0 auto 6px;
  background: ${props => `${props.$color}15`};
  border-radius: 8px;
  color: ${props => props.$color};
  font-size: 16px;
`

const StatValue = styled.div`
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 360px 340px;
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
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 8px;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
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
  padding: 12px;
`

const SamplesPanelBody = styled(PanelBody)`
  padding: 0;
  overflow: auto;
`

const StyledTable = styled(Table)`
  .ant-table {
    background: transparent;
  }

  .ant-table-thead > tr > th {
    background: rgba(255, 255, 255, 0.04);
    color: rgba(255, 255, 255, 0.7);
    border-bottom: 1px solid rgba(255, 255, 255, 0.08);
    padding: 8px 12px;
    font-size: 11px;
  }

  .ant-table-tbody > tr > td {
    border-bottom: 1px solid rgba(255, 255, 255, 0.04);
    color: rgba(255, 255, 255, 0.85);
    padding: 6px 12px;
    font-size: 12px;
    font-family: 'Courier New', monospace;
  }

  .ant-table-tbody > tr:hover > td {
    background: rgba(255, 255, 255, 0.03);
  }

  .ant-table-placeholder {
    color: rgba(255, 255, 255, 0.4);
  }
`

const ThrustLearning: React.FC = () => {
  const { uavList, selectedUAVId, selectCurrentUAV, currentUAV } = useUAV()
  const selectedUAVNum = selectedUAVId ? parseInt(selectedUAVId, 10) : undefined

  const {
    status,
    thrustCurve,
    pidGains,
    samples,
    loading,
    fetchAll,
    startLearning,
    startOptimize,
    savePIDGains,
    applyPID,
    handlePIDGainsChange
  } = useThrustLearning(selectedUAVNum)

  const [learningLoading, setLearningLoading] = useState(false)
  const [optimizeLoading, setOptimizeLoading] = useState(false)
  const [applyLoading, setApplyLoading] = useState(false)
  const [autoTuneLoading, setAutoTuneLoading] = useState(false)

  useEffect(() => {
    const interval = setInterval(() => {
      if (selectedUAVNum !== undefined) {
        fetchAll(selectedUAVNum)
      }
    }, 3000)
    return () => clearInterval(interval)
  }, [selectedUAVNum, fetchAll])

  const handleStartLearning = async () => {
    if (selectedUAVNum === undefined) return
    setLearningLoading(true)
    const result = await startLearning(selectedUAVNum)
    if (result.success) {
      message.success('推力自学习已启动')
    } else {
      message.error(result.message || '启动失败')
    }
    setLearningLoading(false)
  }

  const handleOptimize = async () => {
    if (selectedUAVNum === undefined) return
    setOptimizeLoading(true)
    const result = await startOptimize(selectedUAVNum)
    if (result.success) {
      message.success('模型优化已启动')
    } else {
      message.error(result.message || '优化失败')
    }
    setOptimizeLoading(false)
  }

  const handleApplyPID = async () => {
    if (selectedUAVNum === undefined) return
    setApplyLoading(true)
    const result = await savePIDGains(selectedUAVNum, pidGains || {})
    if (result.success) {
      message.success('PID参数已应用到飞控')
    } else {
      message.error(result.message || '应用失败')
    }
    setApplyLoading(false)
  }

  const handleAutoTune = async () => {
    if (selectedUAVNum === undefined) return
    setAutoTuneLoading(true)
    const result = await applyPID(selectedUAVNum)
    if (result.success) {
      message.success('自动调参已完成')
    } else {
      message.error(result.message || '自动调参失败')
    }
    setAutoTuneLoading(false)
  }

  const state = status?.state || 'idle'
  const stateColor = LEARNING_STATE_COLORS[state]

  const sampleColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 50
    },
    {
      title: '油门',
      dataIndex: 'throttle',
      key: 'throttle',
      width: 70,
      render: (v: number) => `${(v * 100).toFixed(0)}%`
    },
    {
      title: 'AccZ',
      dataIndex: 'accel_z',
      key: 'accel_z',
      width: 70,
      render: (v: number) => v.toFixed(2)
    },
    {
      title: '高度(m)',
      dataIndex: 'altitude',
      key: 'altitude',
      width: 70,
      render: (v: number) => v.toFixed(1)
    },
    {
      title: '电压(V)',
      dataIndex: 'voltage',
      key: 'voltage',
      width: 70,
      render: (v: number) => v.toFixed(1)
    }
  ]

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <RadarChartOutlined style={{ color: '#1890ff' }} />
            推力曲线自学习
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
          <StatusBadge $color={stateColor}>
            <ApiOutlined />
            {LEARNING_STATE_LABELS[state]}
          </StatusBadge>
        </HeaderLeft>
        <HeaderRight>
          <Button
            size="small"
            icon={<ReloadOutlined />}
            onClick={() => selectedUAVNum !== undefined && fetchAll(selectedUAVNum)}
            loading={loading}
          >
            刷新
          </Button>
        </HeaderRight>
      </Header>

      <StatsRow>
        <StatCard>
          <StatIcon $color="#1890ff">
            <ThunderboltOutlined />
          </StatIcon>
          <StatValue>{(status?.estimated_weight_kg || 0).toFixed(2)}</StatValue>
          <StatLabel>估计重量 (kg)</StatLabel>
        </StatCard>
        <StatCard>
          <StatIcon $color="#52c41a">
            <DashboardOutlined />
          </StatIcon>
          <StatValue>{((status?.hover_throttle || 0) * 100).toFixed(1)}%</StatValue>
          <StatLabel>悬停油门</StatLabel>
        </StatCard>
        <StatCard>
          <StatIcon $color="#faad14">
            <DatabaseOutlined />
          </StatIcon>
          <StatValue>{status?.sample_count || 0}</StatValue>
          <StatLabel>已采集样本</StatLabel>
        </StatCard>
        <StatCard>
          <StatIcon $color="#722ed1">
            <ExperimentOutlined />
          </StatIcon>
          <StatValue>
            {pidGains?.is_auto_tuned ? (
              <Tag color="#52c41a" style={{ margin: 0, fontSize: '12px' }}>已调参</Tag>
            ) : (
              <Tag color="#999" style={{ margin: 0, fontSize: '12px' }}>未调参</Tag>
            )}
          </StatValue>
          <StatLabel>PID调参状态</StatLabel>
        </StatCard>
      </StatsRow>

      <Content>
        <Panel>
          <PanelCard style={{ flex: 1, minHeight: 0 }}>
            <PanelHeader>
              <PanelTitle>
                <RadarChartOutlined style={{ color: '#1890ff' }} />
                推力曲线
              </PanelTitle>
              <Tag color="#1890ff">{thrustCurve.length} 个曲线点</Tag>
            </PanelHeader>
            <PanelBody style={{ padding: 8 }}>
              <ThrustCurveChart
                curvePoints={thrustCurve}
                samples={samples}
                hoverThrottle={status?.hover_throttle}
                estimatedWeight={status?.estimated_weight_kg}
              />
            </PanelBody>
          </PanelCard>
        </Panel>

        <Panel>
          <PanelCard style={{ flex: 1, minHeight: 0, overflow: 'auto' }}>
            <PanelHeader>
              <PanelTitle>
                <ThunderboltOutlined style={{ color: '#52c41a' }} />
                PID 增益配置
              </PanelTitle>
              {pidGains?.is_auto_tuned && (
                <Tag color="#52c41a">自动调参</Tag>
              )}
            </PanelHeader>
            <PanelBody>
              <PIDGainsPanel
                gains={pidGains}
                onChange={handlePIDGainsChange}
                onApply={handleApplyPID}
                onAutoTune={handleAutoTune}
                applying={applyLoading}
                autoTuning={autoTuneLoading}
              />
            </PanelBody>
          </PanelCard>
        </Panel>

        <Panel>
          <PanelCard style={{ flex: 0, flexShrink: 0 }}>
            <PanelHeader>
              <PanelTitle>
                <DashboardOutlined style={{ color: '#faad14' }} />
                学习状态
              </PanelTitle>
            </PanelHeader>
            <PanelBody>
              <LearningStatusPanel
                status={status}
                onStartLearning={handleStartLearning}
                onOptimize={handleOptimize}
                learningLoading={learningLoading}
                optimizeLoading={optimizeLoading}
              />
            </PanelBody>
          </PanelCard>

          <PanelCard style={{ flex: 1, minHeight: 0 }}>
            <PanelHeader>
              <PanelTitle>
                <DatabaseOutlined style={{ color: '#722ed1' }} />
                采集样本
              </PanelTitle>
              <Tag color="#722ed1">{samples.length}</Tag>
            </PanelHeader>
            <SamplesPanelBody>
              <StyledTable<ThrustLearningSample>
                dataSource={samples.slice(-50)}
                columns={sampleColumns}
                rowKey="id"
                size="small"
                pagination={false}
                scroll={{ y: '100%' }}
                locale={{ emptyText: '暂无样本数据' }}
              />
            </SamplesPanelBody>
          </PanelCard>
        </Panel>
      </Content>
    </Container>
  )
}

export default ThrustLearning
