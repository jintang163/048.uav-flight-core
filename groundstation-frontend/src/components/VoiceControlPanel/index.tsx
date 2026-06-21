import React, { useEffect, useState } from 'react'
import styled, { keyframes } from 'styled-components'
import { Button, Badge, Tag, Space, Modal, Tooltip } from 'antd'
import {
  AudioOutlined,
  AudioMutedOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ExclamationCircleOutlined,
  SoundOutlined,
  DeleteOutlined,
  RocketOutlined,
  HomeOutlined,
  DownOutlined,
  CameraOutlined,
  VideoCameraOutlined,
  PauseOutlined,
  UnlockOutlined,
  LockOutlined
} from '@ant-design/icons'
import { useVoiceControl } from '@/hooks/useVoiceControl'
import { useUAV } from '@/hooks/useUAV'

const pulseGlow = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(24, 144, 255, 0.7); }
  70% { box-shadow: 0 0 0 12px rgba(24, 144, 255, 0); }
  100% { box-shadow: 0 0 0 0 rgba(24, 144, 255, 0); }
`

const warningPulse = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(250, 173, 20, 0.7); }
  70% { box-shadow: 0 0 0 12px rgba(250, 173, 20, 0); }
  100% { box-shadow: 0 0 0 0 rgba(250, 173, 20, 0); }
`

const Container = styled.div`
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const MicButton = styled(Button)<{ $listening?: boolean; $pending?: boolean }>`
  width: 64px;
  height: 64px;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 24px;
  margin: 0 auto;
  position: relative;
  transition: all 0.3s;
  border: 2px solid ${props => {
    if (props.$pending) return '#faad14'
    if (props.$listening) return '#1890ff'
    return 'rgba(255, 255, 255, 0.2)'
  }};

  background: ${props => {
    if (props.$pending) return 'rgba(250, 173, 20, 0.2)'
    if (props.$listening) return 'rgba(24, 144, 255, 0.2)'
    return 'rgba(255, 255, 255, 0.05)'
  }};

  color: ${props => {
    if (props.$pending) return '#faad14'
    if (props.$listening) return '#1890ff'
    return 'rgba(255, 255, 255, 0.6)'
  }};

  ${props => props.$listening && !props.$pending && `
    animation: ${pulseGlow} 2s infinite;
  `}

  ${props => props.$pending && `
    animation: ${warningPulse} 1s infinite;
  `}

  &:hover {
    transform: scale(1.05);
  }
`

const StatusText = styled.div<{ $listening?: boolean; $pending?: boolean }>`
  text-align: center;
  font-size: 12px;
  color: ${props => {
    if (props.$pending) return '#faad14'
    if (props.$listening) return '#1890ff'
    return 'rgba(255, 255, 255, 0.5)'
  }};
  margin-top: 4px;
`

const LastTextDisplay = styled.div`
  text-align: center;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  min-height: 28px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-style: italic;
`

const CommandList = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 6px;
  margin-top: 4px;
`

const CommandTag = styled.div<{ $type?: string }>`
  font-size: 11px;
  padding: 4px 8px;
  border-radius: 4px;
  display: flex;
  align-items: center;
  gap: 4px;
  background: ${props => {
    switch (props.$type) {
      case 'danger': return 'rgba(255, 77, 79, 0.1)'
      case 'warning': return 'rgba(250, 173, 20, 0.1)'
      case 'success': return 'rgba(82, 196, 26, 0.1)'
      default: return 'rgba(24, 144, 255, 0.1)'
    }
  }};
  color: ${props => {
    switch (props.$type) {
      case 'danger': return '#ff4d4f'
      case 'warning': return '#faad14'
      case 'success': return '#52c41a'
      default: return '#1890ff'
    }
  }};
  border: 1px solid ${props => {
    switch (props.$type) {
      case 'danger': return 'rgba(255, 77, 79, 0.3)'
      case 'warning': return 'rgba(250, 173, 20, 0.3)'
      case 'success': return 'rgba(82, 196, 26, 0.3)'
      default: return 'rgba(24, 144, 255, 0.3)'
    }
  }};
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    filter: brightness(1.2);
  }
`

const ConfirmationPanel = styled.div`
  background: rgba(250, 173, 20, 0.1);
  border: 1px solid rgba(250, 173, 20, 0.3);
  border-radius: 8px;
  padding: 12px;
  text-align: center;
`

const ConfirmationTitle = styled.div`
  font-size: 13px;
  font-weight: 600;
  color: #faad14;
  margin-bottom: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
`

const ConfirmationText = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 10px;
`

const ConfirmationButtons = styled.div`
  display: flex;
  gap: 8px;
  justify-content: center;
`

const LogSection = styled.div`
  max-height: 200px;
  overflow-y: auto;
  margin-top: 8px;

  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-track {
    background: transparent;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.1);
    border-radius: 2px;
  }
`

const LogItem = styled.div<{ $status?: string }>`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 8px;
  font-size: 12px;
  border-radius: 4px;
  margin-bottom: 4px;
  background: ${props => {
    switch (props.$status) {
      case 'executed': return 'rgba(82, 196, 26, 0.05)'
      case 'cancelled': return 'rgba(255, 255, 255, 0.02)'
      case 'executing': return 'rgba(24, 144, 255, 0.05)'
      case 'rejected': return 'rgba(255, 77, 79, 0.05)'
      default: return 'rgba(255, 255, 255, 0.02)'
    }
  }};
  border-left: 2px solid ${props => {
    switch (props.$status) {
      case 'executed': return '#52c41a'
      case 'cancelled': return '#8c8c8c'
      case 'executing': return '#1890ff'
      case 'rejected': return '#ff4d4f'
      default: return 'rgba(255, 255, 255, 0.2)'
    }
  }};
`

const LogTime = styled.span`
  color: rgba(255, 255, 255, 0.3);
  font-family: 'Courier New', monospace;
  font-size: 10px;
  min-width: 50px;
`

const LogText = styled.span`
  color: rgba(255, 255, 255, 0.7);
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

const LogStatus = styled.span`
  font-size: 10px;
  min-width: 40px;
  text-align: right;
`

const NotSupportedWarning = styled.div`
  text-align: center;
  padding: 16px;
  color: rgba(255, 255, 255, 0.5);
  font-size: 12px;
`

const voiceCommandList = [
  { label: '起飞', keywords: '起飞/升空', icon: <RocketOutlined />, type: 'success', confirm: false },
  { label: '降落', keywords: '降落/着陆', icon: <DownOutlined />, type: 'danger', confirm: true },
  { label: '返航', keywords: '返航/回家', icon: <HomeOutlined />, type: 'warning', confirm: true },
  { label: '悬停', keywords: '悬停/停下', icon: <PauseOutlined />, type: 'primary', confirm: false },
  { label: '拍照', keywords: '拍照/照相', icon: <CameraOutlined />, type: 'primary', confirm: false },
  { label: '录像', keywords: '录像/录制', icon: <VideoCameraOutlined />, type: 'primary', confirm: false },
  { label: '解锁', keywords: '解锁', icon: <UnlockOutlined />, type: 'success', confirm: false },
  { label: '上锁', keywords: '上锁/锁定', icon: <LockOutlined />, type: 'danger', confirm: true },
]

interface VoiceControlPanelProps {
  compact?: boolean
}

const VoiceControlPanel: React.FC<VoiceControlPanelProps> = ({ compact = false }) => {
  const { selectedUAVId } = useUAV()
  const {
    isListening,
    isSupported,
    lastText,
    logs,
    pendingConfirmation,
    toggleListening,
    confirmPending,
    cancelPending,
    clearLogs
  } = useVoiceControl(selectedUAVId || undefined)

  const [confirmModalVisible, setConfirmModalVisible] = useState(false)

  useEffect(() => {
    if (pendingConfirmation && !compact) {
      setConfirmModalVisible(true)
    }
  }, [pendingConfirmation, compact])

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'executed': return <CheckCircleOutlined style={{ color: '#52c41a' }} />
      case 'cancelled': return <CloseCircleOutlined style={{ color: '#8c8c8c' }} />
      case 'executing': return <SoundOutlined style={{ color: '#1890ff' }} />
      case 'rejected': return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />
      default: return null
    }
  }

  const getStatusText = (status: string) => {
    switch (status) {
      case 'executed': return '已执行'
      case 'cancelled': return '已取消'
      case 'executing': return '执行中'
      case 'rejected': return '已拒绝'
      default: return '已识别'
    }
  }

  const formatLogTime = (timestamp: number) => {
    const d = new Date(timestamp)
    return `${d.getHours().toString().padStart(2, '0')}:${d.getMinutes().toString().padStart(2, '0')}:${d.getSeconds().toString().padStart(2, '0')}`
  }

  if (!isSupported) {
    return (
      <Container>
        <NotSupportedWarning>
          <AudioMutedOutlined style={{ fontSize: 24, marginBottom: 8, display: 'block' }} />
          当前浏览器不支持语音识别<br />
          请使用 Chrome 或 Edge 浏览器
        </NotSupportedWarning>
      </Container>
    )
  }

  if (compact) {
    return (
      <Tooltip title={isListening ? '语音控制已开启' : '语音控制已关闭'}>
        <Button
          type={isListening ? 'primary' : 'default'}
          shape="circle"
          icon={isListening ? <AudioOutlined /> : <AudioMutedOutlined />}
          onClick={toggleListening}
          style={{
            background: isListening ? 'rgba(24, 144, 255, 0.2)' : 'rgba(255,255,255,0.05)',
            borderColor: isListening ? '#1890ff' : 'rgba(255,255,255,0.2)',
            color: isListening ? '#1890ff' : 'rgba(255,255,255,0.5)'
          }}
        />
      </Tooltip>
    )
  }

  return (
    <Container>
      <MicButton
        $listening={isListening}
        $pending={!!pendingConfirmation}
        icon={isListening ? <AudioOutlined /> : <AudioMutedOutlined />}
        onClick={toggleListening}
      />
      <StatusText $listening={isListening} $pending={!!pendingConfirmation}>
        {pendingConfirmation
          ? '⏳ 等待确认...'
          : isListening
            ? '🎤 语音识别中...'
            : '点击开启语音控制'
        }
      </StatusText>

      {lastText && (
        <LastTextDisplay>
          "{lastText}"
        </LastTextDisplay>
      )}

      {pendingConfirmation && (
        <ConfirmationPanel>
          <ConfirmationTitle>
            <ExclamationCircleOutlined />
            二次确认
          </ConfirmationTitle>
          <ConfirmationText>
            语音指令「<b style={{ color: '#faad14' }}>{pendingConfirmation.command.label}</b>」
            需要确认才能执行
          </ConfirmationText>
          <ConfirmationButtons>
            <Button
              size="small"
              danger
              icon={<CheckCircleOutlined />}
              onClick={confirmPending}
            >
              确认执行
            </Button>
            <Button
              size="small"
              icon={<CloseCircleOutlined />}
              onClick={cancelPending}
            >
              取消
            </Button>
          </ConfirmationButtons>
          <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)', marginTop: 6 }}>
            也可以语音说"确认"或"取消" · 15秒后自动取消
          </div>
        </ConfirmationPanel>
      )}

      <CommandList>
        {voiceCommandList.map(cmd => (
          <Tooltip key={cmd.label} title={`语音指令: ${cmd.keywords}${cmd.confirm ? ' (需确认)' : ''}`}>
            <CommandTag $type={cmd.type}>
              {cmd.icon}
              <span>{cmd.label}</span>
              {cmd.confirm && <span style={{ fontSize: 9, opacity: 0.6 }}>⚠</span>}
            </CommandTag>
          </Tooltip>
        ))}
      </CommandList>

      {logs.length > 0 && (
        <>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginTop: 8 }}>
            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>语音日志</span>
            <Button
              type="text"
              size="small"
              icon={<DeleteOutlined />}
              onClick={clearLogs}
              style={{ color: 'rgba(255,255,255,0.3)', fontSize: 10 }}
            />
          </div>
          <LogSection>
            {logs.map(log => (
              <LogItem key={log.id} $status={log.status}>
                <LogTime>{formatLogTime(log.timestamp)}</LogTime>
                <LogText>"{log.text}"</LogText>
                <LogStatus>
                  {log.action && (
                    <Tag color={
                      log.status === 'executed' ? 'success' :
                      log.status === 'rejected' ? 'error' :
                      log.status === 'cancelled' ? 'default' : 'processing'
                    } style={{ fontSize: 10, margin: 0, padding: '0 4px', lineHeight: '18px' }}>
                      {log.action}
                    </Tag>
                  )}
                  {!log.action && getStatusIcon(log.status)}
                </LogStatus>
              </LogItem>
            ))}
          </LogSection>
        </>
      )}

      <Modal
        title={
          <Space>
            <ExclamationCircleOutlined style={{ color: '#faad14', fontSize: 20 }} />
            语音指令二次确认
          </Space>
        }
        open={confirmModalVisible && !!pendingConfirmation}
        onCancel={() => {
          cancelPending()
          setConfirmModalVisible(false)
        }}
        footer={[
          <Button key="cancel" onClick={() => {
            cancelPending()
            setConfirmModalVisible(false)
          }}>
            取消
          </Button>,
          <Button key="confirm" type="primary" danger onClick={() => {
            confirmPending()
            setConfirmModalVisible(false)
          }}>
            确认执行
          </Button>
        ]}
        width={400}
      >
        {pendingConfirmation && (
          <div style={{ textAlign: 'center', padding: '16px 0' }}>
            <div style={{
              width: 64, height: 64, borderRadius: '50%',
              background: 'rgba(250, 173, 20, 0.1)',
              border: '2px solid rgba(250, 173, 20, 0.3)',
              display: 'flex', alignItems: 'center', justifyContent: 'center',
              margin: '0 auto 16px', fontSize: 28, color: '#faad14'
            }}>
              <SoundOutlined />
            </div>
            <div style={{ fontSize: 16, fontWeight: 600, color: 'rgba(255,255,255,0.9)', marginBottom: 8 }}>
              确认执行「{pendingConfirmation.command.label}」？
            </div>
            <div style={{ fontSize: 13, color: 'rgba(255,255,255,0.6)', marginBottom: 4 }}>
              识别到语音指令: "{pendingConfirmation.text}"
            </div>
            <div style={{
              padding: '8px 12px',
              background: 'rgba(255, 255, 255, 0.05)',
              borderRadius: 6,
              fontSize: 12,
              color: 'rgba(255, 255, 255, 0.5)',
              marginTop: 12
            }}>
              💡 也可以通过语音说"确认"或"取消"来操作，15秒后自动取消
            </div>
          </div>
        )}
      </Modal>
    </Container>
  )
}

export default VoiceControlPanel
