import React from 'react'
import styled from 'styled-components'
import { InputNumber, Button, Space, Tag, Divider } from 'antd'
import {
  ThunderboltOutlined,
  RocketOutlined,
  RiseOutlined,
  CheckCircleOutlined,
  PlayCircleOutlined
} from '@ant-design/icons'
import type { PIDGainProfile } from '@/types/thrust-learning'

const Container = styled.div`
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const GroupHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
`

const GroupTitle = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
`

const GroupBody = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  padding: 0 4px;
`

const GainItem = styled.div`
  display: flex;
  flex-direction: column;
  gap: 4px;
`

const GainLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  text-align: center;
`

const StyledInput = styled(InputNumber)`
  width: 100%;

  .ant-input-number-input {
    background: rgba(0, 0, 0, 0.3);
    color: rgba(255, 255, 255, 0.9);
    text-align: center;
    font-family: 'Courier New', monospace;
    font-size: 13px;
  }

  .ant-input-number-handler-wrap {
    background: rgba(0, 0, 0, 0.3);
  }

  .ant-input-number-handler {
    color: rgba(255, 255, 255, 0.5);
    border-color: rgba(255, 255, 255, 0.06);

    &:hover {
      color: #1890ff;
    }
  }
`

const ActionsRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding-top: 8px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
`

interface PIDGainsPanelProps {
  gains: PIDGainProfile | null
  onChange: (gains: Partial<PIDGainProfile>) => void
  onApply: () => void
  onAutoTune: () => void
  applying?: boolean
  autoTuning?: boolean
}

interface AngleGains {
  roll_kp: number; roll_ki: number; roll_kd: number;
  pitch_kp: number; pitch_ki: number; pitch_kd: number;
  yaw_kp: number; yaw_ki: number; yaw_kd: number;
}

interface RateGains {
  rate_roll_kp: number; rate_roll_ki: number; rate_roll_kd: number;
  rate_pitch_kp: number; rate_pitch_ki: number; rate_pitch_kd: number;
  rate_yaw_kp: number; rate_yaw_ki: number; rate_yaw_kd: number;
}

interface AltGains {
  alt_kp: number; alt_ki: number; alt_kd: number;
}

const PIDGainsPanel: React.FC<PIDGainsPanelProps> = ({
  gains,
  onChange,
  onApply,
  onAutoTune,
  applying,
  autoTuning
}) => {
  const handleAngleChange = (axis: 'roll' | 'pitch' | 'yaw', param: 'kp' | 'ki' | 'kd', value: number | null) => {
    if (value === null) return
    const key = `${axis}_${param}` as keyof AngleGains
    onChange({ [key]: value } as Partial<PIDGainProfile>)
  }

  const handleRateChange = (axis: 'roll' | 'pitch' | 'yaw', param: 'kp' | 'ki' | 'kd', value: number | null) => {
    if (value === null) return
    const key = `rate_${axis}_${param}` as keyof RateGains
    onChange({ [key]: value } as Partial<PIDGainProfile>)
  }

  const handleAltChange = (param: 'kp' | 'ki' | 'kd', value: number | null) => {
    if (value === null) return
    const key = `alt_${param}` as keyof AltGains
    onChange({ [key]: value } as Partial<PIDGainProfile>)
  }

  const g = gains || {
    roll_kp: 0, roll_ki: 0, roll_kd: 0,
    pitch_kp: 0, pitch_ki: 0, pitch_kd: 0,
    yaw_kp: 0, yaw_ki: 0, yaw_kd: 0,
    rate_roll_kp: 0, rate_roll_ki: 0, rate_roll_kd: 0,
    rate_pitch_kp: 0, rate_pitch_ki: 0, rate_pitch_kd: 0,
    rate_yaw_kp: 0, rate_yaw_ki: 0, rate_yaw_kd: 0,
    alt_kp: 0, alt_ki: 0, alt_kd: 0,
    is_auto_tuned: false,
    profile_name: '',
    uav_id: 0
  }

  return (
    <Container>
      <GroupHeader>
        <GroupTitle>
          <ThunderboltOutlined style={{ color: '#1890ff' }} />
          角度环 (Angle)
        </GroupTitle>
        {gains?.is_auto_tuned && (
          <Tag color="#52c41a" icon={<CheckCircleOutlined />}>自动调参</Tag>
        )}
      </GroupHeader>
      <GroupBody>
        {(['roll', 'pitch', 'yaw'] as const).map(axis => (
          <React.Fragment key={axis}>
            {(['kp', 'ki', 'kd'] as const).map(param => (
              <GainItem key={`${axis}-${param}`}>
                <GainLabel>{axis.toUpperCase()} {param.toUpperCase()}</GainLabel>
                <StyledInput
                  size="small"
                  step={0.01}
                  precision={3}
                  value={g[`${axis}_${param}` as keyof AngleGains]}
                  onChange={(v) => handleAngleChange(axis, param, v)}
                />
              </GainItem>
            ))}
          </React.Fragment>
        ))}
      </GroupBody>

      <Divider style={{ margin: '4px 0', borderColor: 'rgba(255,255,255,0.06)' }} />

      <GroupHeader>
        <GroupTitle>
          <RocketOutlined style={{ color: '#52c41a' }} />
          角速度环 (Rate)
        </GroupTitle>
      </GroupHeader>
      <GroupBody>
        {(['roll', 'pitch', 'yaw'] as const).map(axis => (
          <React.Fragment key={axis}>
            {(['kp', 'ki', 'kd'] as const).map(param => (
              <GainItem key={`rate-${axis}-${param}`}>
                <GainLabel>R{axis.charAt(0).toUpperCase()} {param.toUpperCase()}</GainLabel>
                <StyledInput
                  size="small"
                  step={0.001}
                  precision={4}
                  value={g[`rate_${axis}_${param}` as keyof RateGains]}
                  onChange={(v) => handleRateChange(axis, param, v)}
                />
              </GainItem>
            ))}
          </React.Fragment>
        ))}
      </GroupBody>

      <Divider style={{ margin: '4px 0', borderColor: 'rgba(255,255,255,0.06)' }} />

      <GroupHeader>
        <GroupTitle>
          <RiseOutlined style={{ color: '#faad14' }} />
          高度环 (Altitude)
        </GroupTitle>
      </GroupHeader>
      <GroupBody>
        {(['kp', 'ki', 'kd'] as const).map(param => (
          <GainItem key={`alt-${param}`}>
            <GainLabel>ALT {param.toUpperCase()}</GainLabel>
            <StyledInput
              size="small"
              step={0.01}
              precision={3}
              value={g[`alt_${param}` as keyof AltGains]}
              onChange={(v) => handleAltChange(param, v)}
            />
          </GainItem>
        ))}
        <GainItem />
        <GainItem />
      </GroupBody>

      <ActionsRow>
        <Button
          size="small"
          icon={<PlayCircleOutlined />}
          onClick={onAutoTune}
          loading={autoTuning}
        >
          自动调参
        </Button>
        <Space>
          <Button
            size="small"
            type="primary"
            icon={<CheckCircleOutlined />}
            onClick={onApply}
            loading={applying}
          >
            应用到飞控
          </Button>
        </Space>
      </ActionsRow>
    </Container>
  )
}

export default PIDGainsPanel
