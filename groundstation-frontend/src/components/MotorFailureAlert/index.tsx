import React, { useEffect, useMemo, useState } from 'react'
import { Modal, Button, Space, Tag, Descriptions, Alert, Divider, Typography, List, Avatar, Badge } from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  HomeOutlined,
  SafetyCertificateOutlined,
  ExclamationCircleOutlined,
  CloseOutlined,
  ThunderboltFilled
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { dismissMotorAlert, clearMotorAlerts } from '@/store/slices/motor'
import * as motorApi from '@/api/motor'
import type { MotorFailureAlert as MotorFailureAlertType } from '@/types'

const { Text, Paragraph } = Typography

const MotorFailureAlertModal: React.FC = () => {
  const dispatch = useAppDispatch()
  const { activeAlerts } = useAppSelector(state => state.motor)
  const [actionLoading, setActionLoading] = useState<string | null>(null)
  const [visible, setVisible] = useState(false)
  const [audioPlayed, setAudioPlayed] = useState(false)

  const unresolvedAlerts = useMemo(
    () => activeAlerts.filter(a => !a.resolved).sort((a, b) => a.motorIndex - b.motorIndex),
    [activeAlerts]
  )

  const alertsByUAV = useMemo(() => {
    const groups = new Map<number, MotorFailureAlertType[]>()
    unresolvedAlerts.forEach(a => {
      if (!groups.has(a.uavId)) groups.set(a.uavId, [])
      groups.get(a.uavId)!.push(a)
    })
    return groups
  }, [unresolvedAlerts])

  useEffect(() => {
    const shouldShow = unresolvedAlerts.length > 0
    if (shouldShow !== visible) {
      setVisible(shouldShow)
    }
    if (shouldShow && !audioPlayed) {
      try {
        const AudioCtx = (window as any).AudioContext || (window as any).webkitAudioContext
        if (AudioCtx) {
          const ctx = new AudioCtx()
          const osc = ctx.createOscillator()
          const gain = ctx.createGain()
          osc.type = 'square'
          osc.frequency.setValueAtTime(880, ctx.currentTime)
          osc.frequency.setValueAtTime(440, ctx.currentTime + 0.25)
          osc.frequency.setValueAtTime(880, ctx.currentTime + 0.5)
          gain.gain.setValueAtTime(0.3, ctx.currentTime)
          gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.75)
          osc.connect(gain)
          gain.connect(ctx.destination)
          osc.start()
          osc.stop(ctx.currentTime + 0.75)
          setTimeout(() => ctx.close(), 1000)
        }
        const audio = new Audio('/alarm.mp3')
        audio.volume = 0.5
        audio.play().catch(() => {})
      } catch {}
      setAudioPlayed(true)
    }
    if (!shouldShow) {
      setAudioPlayed(false)
    }
  }, [unresolvedAlerts.length, visible, audioPlayed])

  if (unresolvedAlerts.length === 0) return null

  const uniqueUAVIDs = Array.from(alertsByUAV.keys())
  const currentUAVId = uniqueUAVIDs[0]
  const currentAlerts = alertsByUAV.get(currentUAVId) || []
  const totalFailedMotors = currentAlerts.length

  const handleEmergencyRTH = async () => {
    setActionLoading('rth')
    try {
      for (const uid of uniqueUAVIDs) {
        await motorApi.emergencyRTH(uid)
      }
    } catch {}
    setActionLoading(null)
  }

  const handleEmergencyLand = async () => {
    setActionLoading('land')
    try {
      for (const uid of uniqueUAVIDs) {
        await motorApi.emergencyLand(uid)
      }
    } catch {}
    setActionLoading(null)
  }

  const handleDismissAll = () => {
    dispatch(clearMotorAlerts())
  }

  const handleDismiss = (alertId: string) => {
    dispatch(dismissMotorAlert(alertId))
  }

  return (
    <Modal
      open={visible}
      closable={false}
      footer={null}
      width={560}
      centered
      maskClosable={false}
      keyboard={false}
      destroyOnClose
      styles={{
        mask: {
          backdropFilter: 'blur(4px)',
          background: 'rgba(0, 0, 0, 0.65)'
        }
      }}
    >
      <div style={{
        border: '2px solid #ff4d4f',
        borderRadius: 8,
        overflow: 'hidden',
        background: '#fff'
      }}>
        <div style={{
          background: 'linear-gradient(90deg, #ff4d4f 0%, #ff7875 100%)',
          padding: '16px 20px',
          color: '#fff',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between'
        }}>
          <Space size={12}>
            <Badge count={totalFailedMotors} size="large" offset={[4, -2]}>
              <ThunderboltFilled style={{ fontSize: 28 }} />
            </Badge>
            <div>
              <div style={{ fontSize: 18, fontWeight: 700 }}>电机失效紧急告警</div>
              <div style={{ fontSize: 12, opacity: 0.9 }}>
                无人机 {currentUAVId} · 检测到 {totalFailedMotors} 个电机故障 · 飞控已自动调整控制参数
              </div>
            </div>
          </Space>
          <Tag color="white" style={{
            background: 'rgba(255,255,255,0.2)',
            border: '1px solid rgba(255,255,255,0.3)',
            color: '#fff',
            animation: 'pulse 1s infinite'
          }}>
            紧急
          </Tag>
        </div>

        <div style={{ padding: '16px 20px' }}>
          <Alert
            type="error"
            showIcon
            icon={<ExclamationCircleOutlined style={{ fontSize: 20 }} />}
            message={<strong>断桨保护已触发</strong>}
            description={
              <Paragraph style={{ margin: 0 }}>
                六/八旋翼冗余系统已介入。失效电机的动力已重新分配到剩余电机，PID参数已自适应调整，并正在执行紧急返航。
                请保持冷静，监控飞行状态，必要时采取人工干预措施。
              </Paragraph>
            }
            style={{ marginBottom: 16 }}
          />

          <Space direction="vertical" size={12} style={{ width: '100%' }}>
            <Text strong style={{ fontSize: 14 }}>失效电机详情：</Text>
            <List
              size="small"
              dataSource={currentAlerts}
              locale={{ emptyText: '暂无失效电机' }}
              renderItem={(alert) => {
                const motorLabel = `#${alert.motorIndex + 1}`
                const faultHex = '0x' + alert.faultFlags.toString(16).toUpperCase().padStart(4, '0')

                return (
                  <List.Item
                    key={alert.id}
                    style={{
                      background: '#fff2f0',
                      border: '1px solid #ffccc7',
                      borderRadius: 6,
                      padding: '10px 12px',
                      marginBottom: 8
                    }}
                  >
                    <List.Item.Meta
                      avatar={
                        <Avatar
                          size={40}
                          icon={<ThunderboltOutlined />}
                          style={{ background: '#ff4d4f', fontWeight: 700 }}
                        />
                      }
                      title={
                        <Space size={8}>
                          <span style={{ fontWeight: 700, fontSize: 15 }}>
                            电机 {motorLabel}
                          </span>
                          <Tag color="error" icon={<WarningOutlined />}>
                            失效
                          </Tag>
                          {alert.timestamp && (
                            <Text type="secondary" style={{ fontSize: 11 }}>
                              {new Date(alert.timestamp).toLocaleTimeString()}
                            </Text>
                          )}
                        </Space>
                      }
                      description={
                        <Space size={[16, 6]} wrap>
                          <Space size={4}>
                            <Text type="secondary" style={{ fontSize: 12 }}>故障标志:</Text>
                            <Text code style={{ fontSize: 12 }}>{faultHex}</Text>
                          </Space>
                          <Space size={4}>
                            <Text type="secondary" style={{ fontSize: 12 }}>失效时RPM:</Text>
                            <Text type="danger" strong style={{ fontSize: 12 }}>{alert.rpmAtFailure}</Text>
                          </Space>
                          <Space size={4}>
                            <Text type="secondary" style={{ fontSize: 12 }}>温度:</Text>
                            <Text
                              strong
                              type={alert.tempAtFailure > 80 ? 'danger' : 'secondary'}
                              style={{ fontSize: 12 }}
                            >
                              {alert.tempAtFailure}°C
                            </Text>
                          </Space>
                          <Space size={4}>
                            <Text type="secondary" style={{ fontSize: 12 }}>错误码:</Text>
                            <Text code style={{ fontSize: 12 }}>{alert.errorCode}</Text>
                          </Space>
                        </Space>
                      }
                    />
                    <Button
                      type="text"
                      size="small"
                      icon={<CloseOutlined />}
                      onClick={() => handleDismiss(alert.id)}
                    />
                  </List.Item>
                )
              }}
            />
          </Space>

          <Descriptions
            bordered
            size="small"
            column={2}
            style={{ marginTop: 16 }}
          >
            <Descriptions.Item label="失效电机">
              <Tag color="error">
                {currentAlerts.map(a => `#${a.motorIndex + 1}`).join(', ')}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="已执行操作">
              <Space size={4} wrap>
                <Tag color="blue" icon={<SafetyCertificateOutlined />}>
                  混控重分配
                </Tag>
                <Tag color="cyan" icon={<SafetyCertificateOutlined />}>
                  PID自适应
                </Tag>
                <Tag color="orange" icon={<HomeOutlined />}>
                  返航已触发
                </Tag>
              </Space>
            </Descriptions.Item>
          </Descriptions>

          <Divider style={{ margin: '16px 0' }} />

          <Space direction="vertical" style={{ width: '100%' }} size={12}>
            <Text strong style={{ fontSize: 13 }}>紧急操作：</Text>
            <Space wrap size={8}>
              <Button
                type="primary"
                danger
                size="large"
                icon={<HomeOutlined />}
                onClick={handleEmergencyRTH}
                loading={actionLoading === 'rth'}
              >
                立即紧急返航
              </Button>
              <Button
                danger
                size="large"
                icon={<ThunderboltOutlined />}
                onClick={handleEmergencyLand}
                loading={actionLoading === 'land'}
              >
                就地紧急降落
              </Button>
              <Button
                size="large"
                icon={<CloseOutlined />}
                onClick={handleDismissAll}
              >
                全部确认关闭
              </Button>
            </Space>
          </Space>
        </div>
      </div>
    </Modal>
  )
}

export default MotorFailureAlertModal
