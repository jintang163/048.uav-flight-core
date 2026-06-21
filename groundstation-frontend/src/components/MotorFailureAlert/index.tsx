import React, { useEffect, useState } from 'react'
import { Modal, Button, Space, Tag, Descriptions, Alert, Divider, Typography } from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  HomeOutlined,
  SafetyCertificateOutlined,
  ExclamationCircleOutlined,
  CloseOutlined,
  ReloadOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { dismissMotorAlert } from '@/store/slices/motor'
import * as motorApi from '@/api/motor'

const { Text } = Typography

interface MotorFailureAlertModalProps {
  visible?: boolean
}

const MotorFailureAlertModal: React.FC<MotorFailureAlertModalProps> = () => {
  const dispatch = useAppDispatch()
  const { activeAlerts } = useAppSelector(state => state.motor)
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  const currentAlert = activeAlerts.find(a => !a.resolved)
  const visible = !!currentAlert

  useEffect(() => {
    if (visible && currentAlert) {
      const audio = new Audio('/alarm.mp3')
      audio.volume = 0.5
      audio.play().catch(() => {})
    }
  }, [visible, currentAlert])

  if (!currentAlert) return null

  const motorLabel = `电机 #${currentAlert.motorIndex + 1}`
  const faultBits = currentAlert.faultFlags.toString(2).padStart(8, '0')

  const handleEmergencyRTH = async () => {
    if (!currentAlert) return
    setActionLoading('rth')
    try {
      await motorApi.emergencyRTH(currentAlert.uavId)
    } catch {}
    setActionLoading(null)
  }

  const handleEmergencyLand = async () => {
    if (!currentAlert) return
    setActionLoading('land')
    try {
      await motorApi.emergencyLand(currentAlert.uavId)
    } catch {}
    setActionLoading(null)
  }

  const handleDismiss = () => {
    if (currentAlert) {
      dispatch(dismissMotorAlert(currentAlert.id))
    }
  }

  return (
    <Modal
      open={visible}
      closable={false}
      footer={null}
      width={520}
      centered
      styles={{
        body: { padding: 0 },
        mask: { backdropFilter: 'blur(4px)' }
      }}
    >
      <div style={{ padding: '24px 24px 16px' }}>
        <Alert
          type="error"
          showIcon
          icon={<ExclamationCircleOutlined style={{ fontSize: 24 }} />}
          message={
            <Space>
              <span style={{ fontSize: 18, fontWeight: 700 }}>{motorLabel} 失效告警</span>
              <Tag color="error" style={{ animation: 'pulse 1.5s infinite' }}>紧急</Tag>
            </Space>
          }
          description={
            <Text style={{ color: 'rgba(0,0,0,0.65)' }}>
              无人机 {currentAlert.uavId} 的 {motorLabel} 检测到故障，飞控已自动调整PID参数并触发紧急返航。
              请立即确认并采取相应措施。
            </Text>
          }
          style={{ marginBottom: 16 }}
        />

        <Descriptions
          bordered
          size="small"
          column={2}
          style={{ marginBottom: 16 }}
        >
          <Descriptions.Item label="失效电机">
            <Tag color="error" icon={<WarningOutlined />}>
              #{currentAlert.motorIndex + 1}
            </Tag>
          </Descriptions.Item>
          <Descriptions.Item label="故障标志">
            <Text code>0x{currentAlert.faultFlags.toString(16).toUpperCase().padStart(4, '0')}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="失效时RPM">
            <Text type="danger">{currentAlert.rpmAtFailure}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="失效时温度">
            <Text type={currentAlert.tempAtFailure > 80 ? 'danger' : 'secondary'}>
              {currentAlert.tempAtFailure}°C
            </Text>
          </Descriptions.Item>
          <Descriptions.Item label="错误码">
            <Text code>{currentAlert.errorCode}</Text>
          </Descriptions.Item>
          <Descriptions.Item label="已执行操作">
            {currentAlert.actionTaken === 'pid_adjusted_rth' ? (
              <Space size={4}>
                <Tag color="blue" icon={<SafetyCertificateOutlined />}>PID已调整</Tag>
                <Tag color="orange" icon={<HomeOutlined />}>返航已触发</Tag>
              </Space>
            ) : (
              <Tag>—</Tag>
            )}
          </Descriptions.Item>
        </Descriptions>

        <Divider style={{ margin: '12px 0' }} />

        <Space direction="vertical" style={{ width: '100%' }} size={12}>
          <Text strong style={{ fontSize: 13 }}>紧急操作：</Text>
          <Space wrap>
            <Button
              type="primary"
              danger
              icon={<HomeOutlined />}
              onClick={handleEmergencyRTH}
              loading={actionLoading === 'rth'}
            >
              紧急返航
            </Button>
            <Button
              danger
              icon={<ThunderboltOutlined />}
              onClick={handleEmergencyLand}
              loading={actionLoading === 'land'}
            >
              紧急降落
            </Button>
            <Button
              icon={<CloseOutlined />}
              onClick={handleDismiss}
            >
              确认并关闭
            </Button>
          </Space>
        </Space>
      </div>
    </Modal>
  )
}

export default MotorFailureAlertModal
