import React, { useState } from 'react'
import styled from 'styled-components'
import { Button, Space, Modal, message, InputNumber, Divider } from 'antd'
import {
  RocketOutlined,
  DownOutlined,
  HomeOutlined,
  LockOutlined,
  UnlockOutlined,
  PauseOutlined,
  PlayCircleOutlined,
  StopOutlined
} from '@ant-design/icons'
import { useUAV } from '@/hooks/useUAV'
import { sendCommand } from '@/websocket/command'
import { speakAlert } from '@/utils'
import type { UAVMode } from '@/types'
import PreflightCheckPanel from '@/components/PreflightCheckPanel'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
`

const Title = styled.div`
  font-size: 15px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
`

const ButtonGroup = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 12px;
  margin-bottom: 16px;
`

const ControlButton = styled(Button)<{ $danger?: boolean; $success?: boolean }>`
  height: 56px;
  font-size: 14px;
  font-weight: 500;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;

  .anticon {
    font-size: 20px;
  }

  ${props => props.$danger && `
    &.ant-btn-primary {
      background: #ff4d4f;
      border-color: #ff4d4f;
    }
    &.ant-btn-primary:hover {
      background: #ff7875 !important;
      border-color: #ff7875 !important;
    }
  `}

  ${props => props.$success && `
    &.ant-btn-primary {
      background: #52c41a;
      border-color: #52c41a;
    }
    &.ant-btn-primary:hover {
      background: #73d13d !important;
      border-color: #73d13d !important;
    }
  `}
`

const ButtonLabel = styled.span`
  font-size: 12px;
`

const StatusSection = styled.div`
  margin-top: auto;
  padding: 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
`

const StatusRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }
`

const StatusLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
`

const StatusValue = styled.span<{ $color?: string }>`
  font-weight: 500;
  color: ${props => props.$color || '#fff'};
`

const AltitudeInput = styled.div`
  margin-top: 16px;

  .ant-input-number {
    width: 100%;
  }
`

interface ControlPanelProps {
  showTitle?: boolean
}

const getModeText = (mode: UAVMode): string => {
  const modeMap: Record<UAVMode, string> = {
    manual: '手动模式',
    stabilize: '自稳模式',
    acro: '特技模式',
    guided: '引导模式',
    auto: '自动模式',
    rtl: '返航模式',
    land: '降落模式',
    circle: '环绕模式',
    loiter: '悬停模式',
    follow: '跟随模式',
    unknown: '未知模式'
  }
  return modeMap[mode] || mode
}

const getModeColor = (mode: UAVMode): string => {
  const colorMap: Record<UAVMode, string> = {
    manual: '#ff4d4f',
    stabilize: '#faad14',
    acro: '#eb2f96',
    guided: '#1890ff',
    auto: '#52c41a',
    rtl: '#13c2c2',
    land: '#fa8c16',
    circle: '#722ed1',
    loiter: '#2f54eb',
    follow: '#a0d911',
    unknown: '#8c8c8c'
  }
  return colorMap[mode] || '#8c8c8c'
}

const ControlPanel: React.FC<ControlPanelProps> = ({ showTitle = true }) => {
  const { currentUAV, loading } = useUAV()
  const [takeoffAltitude, setTakeoffAltitude] = useState<number>(5)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [preflightPassed, setPreflightPassed] = useState(false)

  const confirmAction = (title: string, content: string, onConfirm: () => void) => {
    Modal.confirm({
      title,
      content,
      okText: '确认执行',
      cancelText: '取消',
      okButtonProps: { danger: title.includes('紧急') || title.includes('降落') },
      onOk: () => {
        onConfirm()
      }
    })
  }

  const handleArm = async () => {
    if (!currentUAV) return
    setActionLoading('arm')
    try {
      await sendCommand(currentUAV.id, 'arm')
      message.success('无人机已解锁')
      speakAlert('无人机已解锁')
    } catch (error) {
      message.error('解锁失败')
    } finally {
      setActionLoading(null)
    }
  }

  const handleDisarm = async () => {
    if (!currentUAV) return
    confirmAction(
      '确认上锁',
      '请确保无人机已安全降落，上锁后电机将停止转动。',
      async () => {
        setActionLoading('disarm')
        try {
          await sendCommand(currentUAV.id, 'disarm')
          message.success('无人机已上锁')
          speakAlert('无人机已上锁')
        } catch (error) {
          message.error('上锁失败')
        } finally {
          setActionLoading(null)
        }
      }
    )
  }

  const handleTakeoff = async () => {
    if (!currentUAV) return
    if (!preflightPassed) {
      message.error('飞行前自检未通过，禁止起飞！请先完成自检并修复所有不通过项')
      speakAlert('飞行前自检未通过，禁止起飞')
      return
    }
    confirmAction(
      '确认起飞',
      `无人机将起飞至 ${takeoffAltitude} 米高度，请确保周围环境安全。`,
      async () => {
        setActionLoading('takeoff')
        try {
          await sendCommand(currentUAV.id, 'takeoff', { altitude: takeoffAltitude })
          message.success(`起飞指令已发送，目标高度 ${takeoffAltitude} 米`)
          speakAlert(`起飞指令已发送，目标高度 ${takeoffAltitude} 米`)
        } catch (error) {
          message.error('起飞失败')
        } finally {
          setActionLoading(null)
        }
      }
    )
  }

  const handleLand = async () => {
    if (!currentUAV) return
    confirmAction(
      '确认降落',
      '无人机将执行降落操作，请确保下方无障碍物。',
      async () => {
        setActionLoading('land')
        try {
          await sendCommand(currentUAV.id, 'land')
          message.success('降落指令已发送')
          speakAlert('降落指令已发送')
        } catch (error) {
          message.error('降落失败')
        } finally {
          setActionLoading(null)
        }
      }
    )
  }

  const handleRTL = async () => {
    if (!currentUAV) return
    confirmAction(
      '确认返航',
      '无人机将自动返航至起飞点，是否继续？',
      async () => {
        setActionLoading('rtl')
        try {
          await sendCommand(currentUAV.id, 'rtl')
          message.success('返航指令已发送')
          speakAlert('返航指令已发送')
        } catch (error) {
          message.error('返航失败')
        } finally {
          setActionLoading(null)
        }
      }
    )
  }

  const handlePause = async () => {
    if (!currentUAV) return
    setActionLoading('pause')
    try {
      await sendCommand(currentUAV.id, 'pause')
      message.success('已悬停')
      speakAlert('已悬停')
    } catch (error) {
      message.error('悬停失败')
    } finally {
      setActionLoading(null)
    }
  }

  const handleResume = async () => {
    if (!currentUAV) return
    setActionLoading('resume')
    try {
      await sendCommand(currentUAV.id, 'resume')
      message.success('已继续任务')
      speakAlert('已继续任务')
    } catch (error) {
      message.error('继续失败')
    } finally {
      setActionLoading(null)
    }
  }

  const handleEmergencyStop = async () => {
    if (!currentUAV) return
    confirmAction(
      '紧急停桨',
      '警告：此操作将立即停止所有电机，无人机将会坠落！仅在紧急情况下使用！',
      async () => {
        setActionLoading('emergency')
        try {
          await sendCommand(currentUAV.id, 'emergency_stop')
          message.warning('紧急停桨指令已发送')
          speakAlert('紧急停桨')
        } catch (error) {
          message.error('紧急停桨失败')
        } finally {
          setActionLoading(null)
        }
      }
    )
  }

  const isArmed = currentUAV?.armed
  const isFlying = currentUAV?.status === 'flying'

  return (
    <Container>
      {showTitle && (
        <Title>
          <RocketOutlined style={{ color: '#1890ff' }} />
          飞行控制
        </Title>
      )}

      {!currentUAV ? (
        <div style={{ textAlign: 'center', color: 'rgba(255,255,255,0.5)', padding: 40 }}>
          请先选择无人机
        </div>
      ) : (
        <>
          <ButtonGroup>
            <ControlButton
              type="primary"
              icon={<UnlockOutlined />}
              onClick={handleArm}
              loading={actionLoading === 'arm'}
              disabled={isArmed || isFlying || loading}
              $success
            >
              <ButtonLabel>解锁</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<LockOutlined />}
              onClick={handleDisarm}
              loading={actionLoading === 'disarm'}
              disabled={!isArmed || isFlying || loading}
              $danger
            >
              <ButtonLabel>上锁</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<RocketOutlined />}
              onClick={handleTakeoff}
              loading={actionLoading === 'takeoff'}
              disabled={!isArmed || isFlying || loading || !preflightPassed}
              $success
            >
              <ButtonLabel>{!preflightPassed ? '需自检' : '起飞'}</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<DownOutlined />}
              onClick={handleLand}
              loading={actionLoading === 'land'}
              disabled={!isFlying || loading}
            >
              <ButtonLabel>降落</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<HomeOutlined />}
              onClick={handleRTL}
              loading={actionLoading === 'rtl'}
              disabled={!isFlying || loading}
            >
              <ButtonLabel>返航</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<PauseOutlined />}
              onClick={handlePause}
              loading={actionLoading === 'pause'}
              disabled={!isFlying || loading}
            >
              <ButtonLabel>悬停</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={handleResume}
              loading={actionLoading === 'resume'}
              disabled={!isFlying || loading}
            >
              <ButtonLabel>继续</ButtonLabel>
            </ControlButton>

            <ControlButton
              type="primary"
              icon={<StopOutlined />}
              onClick={handleEmergencyStop}
              loading={actionLoading === 'emergency'}
              disabled={loading}
              $danger
            >
              <ButtonLabel>紧急停桨</ButtonLabel>
            </ControlButton>
          </ButtonGroup>

          {!isFlying && (
            <AltitudeInput>
              <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.7)', marginBottom: 8 }}>
                起飞高度 (米)
              </div>
              <Space.Compact style={{ width: '100%' }}>
                <InputNumber
                  type="number"
                  value={takeoffAltitude}
                  onChange={(v) => setTakeoffAltitude(Number(v) || 5)}
                  min={1}
                  max={120}
                  style={{ width: '100%' }}
                />
              </Space.Compact>
            </AltitudeInput>
          )}

          {!isFlying && currentUAV && (
            <div style={{ marginTop: 16 }}>
              <PreflightCheckPanel
                uavId={Number(currentUAV.id)}
                uavName={currentUAV.name}
                showTitle={false}
                compact
                onPassChange={setPreflightPassed}
                onTakeoff={handleTakeoff}
              />
            </div>
          )}

          <StatusSection>
            <StatusRow>
              <StatusLabel>当前模式</StatusLabel>
              <StatusValue $color={getModeColor(currentUAV.mode)}>
                {getModeText(currentUAV.mode)}
              </StatusValue>
            </StatusRow>
            <StatusRow>
              <StatusLabel>飞行状态</StatusLabel>
              <StatusValue $color={isFlying ? '#52c41a' : '#faad14'}>
                {isFlying ? '飞行中' : '地面'}
              </StatusValue>
            </StatusRow>
            <StatusRow>
              <StatusLabel>锁定状态</StatusLabel>
              <StatusValue $color={isArmed ? '#52c41a' : '#ff4d4f'}>
                {isArmed ? '已解锁' : '已上锁'}
              </StatusValue>
            </StatusRow>
          </StatusSection>
        </>
      )}
    </Container>
  )
}

export default ControlPanel
