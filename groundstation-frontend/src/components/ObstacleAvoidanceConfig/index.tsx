import React, { useState } from 'react'
import styled from 'styled-components'
import { Switch, Select, Slider, Button, message, Space, Card, Tooltip } from 'antd'
import {
  SafetyCertificateOutlined,
  WarningOutlined,
  ArrowUpOutlined,
  UndoOutlined,
  PauseOutlined
} from '@ant-design/icons'
import type { AvoidanceSensitivity, AvoidanceStrategy, ObstacleSensorType } from '@/types/obstacle-avoidance'
import { SENSITIVITY_CONFIG, STRATEGY_LABELS, SENSOR_TYPE_LABELS } from '@/types/obstacle-avoidance'

const Container = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
`

const ConfigSection = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  padding: 14px;
`

const SectionTitle = styled.div`
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
`

const ConfigRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 12px;

  &:last-child {
    margin-bottom: 0;
  }
`

const ConfigLabel = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  display: flex;
  align-items: center;
  gap: 6px;
`

const ConfigValue = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.9);
`

const SliderContainer = styled.div`
  width: 100%;
  margin-top: 8px;
  padding: 0 4px;
`

const StrategyCards = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
`

const StrategyCard = styled.div<{ $active: boolean; $color: string }>`
  background: ${props => props.$active ? `rgba(${props.$color}, 0.15)` : 'rgba(255, 255, 255, 0.03)'};
  border: 1px solid ${props => props.$active ? `rgba(${props.$color}, 0.5)` : 'rgba(255, 255, 255, 0.1)'};
  border-radius: 8px;
  padding: 12px 8px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    border-color: rgba(${props => props.$color}, 0.6);
    background: rgba(${props => props.$color}, 0.1);
  }
`

const StrategyIcon = styled.div<{ $color: string }>`
  font-size: 20px;
  color: rgba(${props => props.$color}, 1);
  margin-bottom: 6px;
`

const StrategyName = styled.div<{ $active: boolean }>`
  font-size: 11px;
  font-weight: ${props => props.$active ? 600 : 400};
  color: ${props => props.$active ? '#fff' : 'rgba(255, 255, 255, 0.6)'};
`

const ActionButtons = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 4px;
`

interface ObstacleAvoidanceConfigProps {
  enabled: boolean
  sensitivity: AvoidanceSensitivity
  strategy: AvoidanceStrategy
  sensorType?: ObstacleSensorType
  detectionRange?: number
  ascendHeight?: number
  retreatDistance?: number
  bypassAngle?: number
  onEnabledChange?: (enabled: boolean) => void
  onSensitivityChange?: (sensitivity: AvoidanceSensitivity) => void
  onStrategyChange?: (strategy: AvoidanceStrategy) => void
  onSensorTypeChange?: (sensorType: ObstacleSensorType) => void
  onDetectionRangeChange?: (range: number) => void
  onAscendHeightChange?: (height: number) => void
  onRetreatDistanceChange?: (distance: number) => void
  onBypassAngleChange?: (angle: number) => void
  onApply?: () => void
  onReset?: () => void
}

const STRATEGY_CONFIG: { key: AvoidanceStrategy; label: string; icon: React.ReactNode; color: string }[] = [
  { key: 'hover', label: '悬停', icon: <PauseOutlined />, color: '250, 173, 20' },
  { key: 'ascend_bypass', label: '上升绕行', icon: <ArrowUpOutlined />, color: '114, 46, 209' },
  { key: 'retreat_bypass', label: '后退绕行', icon: <UndoOutlined />, color: '24, 144, 255' }
]

const ObstacleAvoidanceConfigComponent: React.FC<ObstacleAvoidanceConfigProps> = ({
  enabled,
  sensitivity,
  strategy,
  sensorType = 'millimeter_wave_radar',
  detectionRange = 15,
  ascendHeight = 5,
  retreatDistance = 10,
  bypassAngle = 45,
  onEnabledChange,
  onSensitivityChange,
  onStrategyChange,
  onSensorTypeChange,
  onDetectionRangeChange,
  onAscendHeightChange,
  onRetreatDistanceChange,
  onBypassAngleChange,
  onApply,
  onReset
}) => {
  return (
    <Container>
      <ConfigSection>
        <SectionTitle>
          <SafetyCertificateOutlined style={{ color: '#1890ff' }} />
          避障开关
        </SectionTitle>
        <ConfigRow>
          <ConfigLabel>
            避障功能
          </ConfigLabel>
          <Switch
            checked={enabled}
            onChange={onEnabledChange}
            checkedChildren="开启"
            unCheckedChildren="关闭"
          />
        </ConfigRow>
      </ConfigSection>

      <ConfigSection>
        <SectionTitle>
          <WarningOutlined style={{ color: '#faad14' }} />
          灵敏度设置
        </SectionTitle>
        <ConfigRow>
          <ConfigLabel>检测灵敏度</ConfigLabel>
          <ConfigValue>
            <Select
              value={sensitivity}
              onChange={onSensitivityChange}
              size="small"
              style={{ width: 140 }}
              options={Object.entries(SENSITIVITY_CONFIG).map(([key, cfg]) => ({
                value: key,
                label: cfg.label
              }))}
            />
          </ConfigValue>
        </ConfigRow>
        <SliderContainer>
          <Slider
            min={5}
            max={15}
            step={1}
            value={detectionRange}
            onChange={onDetectionRangeChange}
            marks={{
              5: '5m',
              10: '10m',
              15: '15m'
            }}
            tooltip={{ formatter: v => `${v}m` }}
          />
        </SliderContainer>
        <ConfigRow>
          <ConfigLabel>传感器类型</ConfigLabel>
          <ConfigValue>
            <Select
              value={sensorType}
              onChange={onSensorTypeChange}
              size="small"
              style={{ width: 140 }}
              options={Object.entries(SENSOR_TYPE_LABELS).map(([key, label]) => ({
                value: key,
                label
              }))}
            />
          </ConfigValue>
        </ConfigRow>
      </ConfigSection>

      <ConfigSection>
        <SectionTitle>
          避障策略
        </SectionTitle>
        <StrategyCards>
          {STRATEGY_CONFIG.map(s => (
            <StrategyCard
              key={s.key}
              $active={strategy === s.key}
              $color={s.color}
              onClick={() => onStrategyChange?.(s.key)}
            >
              <StrategyIcon $color={s.color}>{s.icon}</StrategyIcon>
              <StrategyName $active={strategy === s.key}>{s.label}</StrategyName>
            </StrategyCard>
          ))}
        </StrategyCards>
      </ConfigSection>

      {(strategy === 'ascend_bypass' || strategy === 'retreat_bypass') && (
        <ConfigSection>
          <SectionTitle>绕行参数</SectionTitle>
          {strategy === 'ascend_bypass' && (
            <>
              <ConfigRow>
                <ConfigLabel>上升高度</ConfigLabel>
                <ConfigValue>{ascendHeight}m</ConfigValue>
              </ConfigRow>
              <SliderContainer>
                <Slider
                  min={2}
                  max={20}
                  step={1}
                  value={ascendHeight}
                  onChange={onAscendHeightChange}
                  marks={{ 2: '2m', 10: '10m', 20: '20m' }}
                  tooltip={{ formatter: v => `${v}m` }}
                />
              </SliderContainer>
            </>
          )}
          {strategy === 'retreat_bypass' && (
            <>
              <ConfigRow>
                <ConfigLabel>后退距离</ConfigLabel>
                <ConfigValue>{retreatDistance}m</ConfigValue>
              </ConfigRow>
              <SliderContainer>
                <Slider
                  min={3}
                  max={30}
                  step={1}
                  value={retreatDistance}
                  onChange={onRetreatDistanceChange}
                  marks={{ 3: '3m', 15: '15m', 30: '30m' }}
                  tooltip={{ formatter: v => `${v}m` }}
                />
              </SliderContainer>
            </>
          )}
          <ConfigRow>
            <ConfigLabel>绕行角度</ConfigLabel>
            <ConfigValue>{bypassAngle}°</ConfigValue>
          </ConfigRow>
          <SliderContainer>
            <Slider
              min={15}
              max={90}
              step={5}
              value={bypassAngle}
              onChange={onBypassAngleChange}
              marks={{ 15: '15°', 45: '45°', 90: '90°' }}
              tooltip={{ formatter: v => `${v}°` }}
            />
          </SliderContainer>
        </ConfigSection>
      )}

      <ActionButtons>
        <Button
          type="primary"
          onClick={onApply}
          style={{ flex: 1 }}
        >
          应用配置
        </Button>
        <Button onClick={onReset}>
          重置
        </Button>
      </ActionButtons>
    </Container>
  )
}

export default ObstacleAvoidanceConfigComponent
