import React from 'react'
import styled from 'styled-components'
import { Progress, Button, Tag, Steps } from 'antd'
import {
  PlayCircleOutlined,
  ExperimentOutlined,
  DashboardOutlined,
  ThunderboltOutlined,
  CheckCircleOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
import type { ThrustLearningStatus, LearningState } from '@/types/thrust-learning'
import { LEARNING_STATE_LABELS, LEARNING_STATE_COLORS } from '@/types/thrust-learning'

const Container = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
  width: 100%;
`

const StatusRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
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
  font-weight: 600;
`

const StatsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
`

const StatCard = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.06);
  border-radius: 8px;
  padding: 10px 12px;
`

const StatValue = styled.div`
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-top: 2px;
`

const ActionsRow = styled.div`
  display: flex;
  gap: 8px;
`

const TimelineWrapper = styled.div`
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  padding: 12px;
  border: 1px solid rgba(255, 255, 255, 0.06);
`

const TimelineTitle = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 6px;
`

interface LearningStatusPanelProps {
  status: ThrustLearningStatus | null
  onStartLearning: () => void
  onOptimize: () => void
  learningLoading?: boolean
  optimizeLoading?: boolean
}

const stateOrder: LearningState[] = ['idle', 'weight_estimation', 'data_collecting', 'model_optimizing', 'applied']

const getStepIndex = (state: LearningState): number => {
  const idx = stateOrder.indexOf(state)
  return idx === -1 ? 0 : idx
}

const LearningStatusPanel: React.FC<LearningStatusPanelProps> = ({
  status,
  onStartLearning,
  onOptimize,
  learningLoading,
  optimizeLoading
}) => {
  const state: LearningState = status?.state || 'idle'
  const color = LEARNING_STATE_COLORS[state]
  const currentStep = getStepIndex(state)

  const canStart = state === 'idle' || state === 'applied'
  const canOptimize = state === 'data_collecting' && (status?.sample_count || 0) >= 10

  return (
    <Container>
      <StatusRow>
        <StatusBadge $color={color}>
          <DashboardOutlined />
          {LEARNING_STATE_LABELS[state]}
        </StatusBadge>
        {status?.progress_pct !== undefined && (
          <Tag color={color} style={{ margin: 0 }}>
            {status.progress_pct}%
          </Tag>
        )}
      </StatusRow>

      {status?.progress_pct !== undefined && (
        <Progress
          percent={status.progress_pct}
          showInfo={false}
          strokeColor={color}
          trailColor="rgba(255,255,255,0.06)"
          size="small"
        />
      )}

      <StatsGrid>
        <StatCard>
          <StatValue>{(status?.estimated_weight_kg || 0).toFixed(2)}</StatValue>
          <StatLabel>估计重量 (kg)</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue>{((status?.hover_throttle || 0) * 100).toFixed(1)}%</StatValue>
          <StatLabel>悬停油门</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue>{status?.sample_count || 0}</StatValue>
          <StatLabel>已采集样本</StatLabel>
        </StatCard>
        <StatCard>
          <StatValue>
            {status?.started_at
              ? new Date(status.started_at).toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' })
              : '--'}
          </StatValue>
          <StatLabel>开始时间</StatLabel>
        </StatCard>
      </StatsGrid>

      <ActionsRow>
        <Button
          type="primary"
          icon={<PlayCircleOutlined />}
          onClick={onStartLearning}
          loading={learningLoading}
          disabled={!canStart}
          size="small"
          block
        >
          {state === 'applied' ? '重新学习' : '开始学习'}
        </Button>
        <Button
          icon={<ExperimentOutlined />}
          onClick={onOptimize}
          loading={optimizeLoading}
          disabled={!canOptimize}
          size="small"
          block
        >
          优化模型
        </Button>
      </ActionsRow>

      <TimelineWrapper>
        <TimelineTitle>
          <ThunderboltOutlined style={{ color: '#1890ff' }} />
          学习流程时间线
        </TimelineTitle>
        <Steps
          direction="vertical"
          size="small"
          current={currentStep}
          status={state === 'applied' ? 'finish' : state === 'idle' ? 'wait' : 'process'}
          items={[
            {
              title: '空闲等待',
              description: '等待开始自学习',
              icon: <DashboardOutlined />
            },
            {
              title: '重量估算',
              description: '通过悬停油门估算无人机重量',
              icon: <DatabaseOutlined />
            },
            {
              title: '数据采集',
              description: '采集不同油门下的推力数据',
              icon: <DatabaseOutlined />
            },
            {
              title: '模型优化',
              description: '拟合推力曲线并优化PID参数',
              icon: <ExperimentOutlined />
            },
            {
              title: '已应用',
              description: '优化参数已应用到飞控',
              icon: <CheckCircleOutlined />
            }
          ]}
          styles={{
            item: { paddingBottom: 8 },
            itemTitle: { color: 'rgba(255,255,255,0.9)', fontSize: '12px', fontWeight: 600 },
            itemDescription: { color: 'rgba(255,255,255,0.5)', fontSize: '11px' }
          }}
        />
      </TimelineWrapper>
    </Container>
  )
}

export default LearningStatusPanel
