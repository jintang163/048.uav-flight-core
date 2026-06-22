import React, { useEffect, useState, useMemo, useCallback } from 'react'
import { Modal, List, Tag, Space, Button, Typography, Alert } from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  PauseCircleOutlined,
  SafetyOutlined,
  CheckCircleOutlined,
  CloseOutlined,
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { fetchActiveAlerts, dismissAlert } from '@/store/slices/collision'
import { initCollisionWebSocket } from '@/websocket/collision'
import type { CollisionAlert, CollisionRiskLevel } from '@/types'

const { Text, Paragraph } = Typography

const riskConfig: Record<CollisionRiskLevel, { color: string; icon: React.ReactNode; label: string }> = {
  safe: { color: 'green', icon: <SafetyOutlined />, label: '安全' },
  warning: { color: 'orange', icon: <WarningOutlined />, label: '警告' },
  critical: { color: 'red', icon: <ThunderboltOutlined />, label: '严重' },
  avoiding: { color: 'purple', icon: <PauseCircleOutlined />, label: '避让中' },
  resolved: { color: 'blue', icon: <CheckCircleOutlined />, label: '已解除' },
}

const actionLabel: Record<string, string> = {
  speed_reduce: '减速避让',
  speed_adjust: '速度调整',
  hold_position: '悬停等待',
  waypoint_hold: '航点等待',
  altitude_change: '高度调整',
  resume: '恢复正常',
}

const CollisionAlertModal: React.FC = () => {
  const dispatch = useAppDispatch()
  const { alerts } = useAppSelector(state => state.collision)
  const [visible, setVisible] = useState(false)
  const initialized = React.useRef(false)

  useEffect(() => {
    if (!initialized.current) {
      initCollisionWebSocket()
      dispatch(fetchActiveAlerts())
      initialized.current = true
    }
  }, [dispatch])

  const activeAlerts = useMemo(
    () => alerts.filter(a => !a.is_resolved),
    [alerts]
  )

  useEffect(() => {
    if (activeAlerts.length > 0 && !visible) {
      setVisible(true)
    }
  }, [activeAlerts.length, visible])

  useEffect(() => {
    if (activeAlerts.some(a => a.risk_level === 'critical' || a.risk_level === 'avoiding')) {
      try {
        const AudioCtx = (window as any).AudioContext || (window as any).webkitAudioContext
        if (AudioCtx) {
          const ctx = new AudioCtx()
          const osc = ctx.createOscillator()
          const gain = ctx.createGain()
          osc.type = 'sawtooth'
          osc.frequency.setValueAtTime(880, ctx.currentTime)
          osc.frequency.setValueAtTime(440, ctx.currentTime + 0.12)
          osc.frequency.setValueAtTime(880, ctx.currentTime + 0.24)
          gain.gain.setValueAtTime(0.15, ctx.currentTime)
          gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.36)
          osc.connect(gain)
          gain.connect(ctx.destination)
          osc.start()
          osc.stop(ctx.currentTime + 0.36)
        }
      } catch {}
    }
  }, [activeAlerts.length])

  const handleClose = useCallback(() => {
    setVisible(false)
  }, [])

  const handleDismiss = useCallback((id: number) => {
    dispatch(dismissAlert(id))
  }, [dispatch])

  const criticalAlerts = activeAlerts.filter(a => a.risk_level === 'critical' || a.risk_level === 'avoiding')
  const warningAlerts = activeAlerts.filter(a => a.risk_level === 'warning')

  const renderAlertItem = (alert: CollisionAlert) => {
    const cfg = riskConfig[alert.risk_level] || riskConfig.warning
    return (
      <List.Item
        key={alert.id}
        actions={[
          <Button
            type="link"
            size="small"
            danger={alert.risk_level === 'critical'}
            onClick={() => handleDismiss(alert.id)}
          >
            解除
          </Button>,
        ]}
      >
        <List.Item.Meta
          avatar={
            <span style={{ fontSize: 20 }}>
              {cfg.icon}
            </span>
          }
          title={
            <Space>
              <Tag color={cfg.color}>{cfg.label}</Tag>
              <Text strong>碰撞预警</Text>
              {alert.action_taken && (
                <Tag color="purple">{actionLabel[alert.action_taken] || alert.action_taken}</Tag>
              )}
            </Space>
          }
          description={
            <div>
              <Paragraph style={{ marginBottom: 4 }}>
                无人机 #{alert.uav_id_1} 与 #{alert.uav_id_2}
              </Paragraph>
              <Space size="large">
                <Text type="secondary">
                  当前距离: <Text strong>{alert.current_distance.toFixed(1)}m</Text>
                </Text>
                {alert.time_to_collision > 0 && (
                  <Text type="secondary">
                    TTC: <Text strong>{alert.time_to_collision.toFixed(1)}s</Text>
                  </Text>
                )}
                {alert.action_detail && (
                  <Text type="secondary">{alert.action_detail}</Text>
                )}
              </Space>
            </div>
          }
        />
      </List.Item>
    )
  }

  return (
    <Modal
      title={
        <Space>
          <ThunderboltOutlined style={{ color: '#ff4d4f' }} />
          <span>碰撞预警</span>
          {activeAlerts.length > 0 && <Tag color="red">{activeAlerts.length}</Tag>}
        </Space>
      }
      open={visible}
      onCancel={handleClose}
      width={640}
      footer={[
        <Button key="close" onClick={handleClose}>
          关闭
        </Button>,
      ]}
    >
      {criticalAlerts.length > 0 && (
        <Alert
          type="error"
          showIcon
          icon={<ThunderboltOutlined />}
          message="严重碰撞风险"
          description="系统已自动触发避让机制，请注意监控"
          style={{ marginBottom: 16 }}
        />
      )}

      {criticalAlerts.length > 0 && (
        <>
          <Text strong style={{ color: '#ff4d4f' }}>严重预警</Text>
          <List size="small" dataSource={criticalAlerts} renderItem={renderAlertItem} style={{ marginBottom: 16 }} />
        </>
      )}

      {warningAlerts.length > 0 && (
        <>
          <Text strong style={{ color: '#fa8c16' }}>警告</Text>
          <List size="small" dataSource={warningAlerts} renderItem={renderAlertItem} />
        </>
      )}

      {activeAlerts.length === 0 && (
        <div style={{ textAlign: 'center', padding: 32 }}>
          <SafetyOutlined style={{ fontSize: 48, color: '#52c41a' }} />
          <Paragraph style={{ marginTop: 16 }}>所有无人机均处于安全距离</Paragraph>
        </div>
      )}
    </Modal>
  )
}

export default CollisionAlertModal
