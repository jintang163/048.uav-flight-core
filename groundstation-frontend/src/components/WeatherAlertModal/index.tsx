import React, { useEffect, useMemo, useState, useCallback } from 'react'
import { Modal, Button, Space, Tag, Descriptions, Alert, Typography, List } from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  HomeOutlined,
  CloudOutlined,
  SafetyCertificateOutlined,
  CloseOutlined,
  CloudFilled,
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { removeWeatherAlert } from '@/store/slices/weather'
import { resolveWeatherAlert } from '@/api/weather'
import type { WeatherAlertEvent, WeatherAlertLevel, WeatherAlertType } from '@/types'

const { Text, Paragraph } = Typography

const alertLevelConfig: Record<WeatherAlertLevel, { color: string; icon: React.ReactNode; title: string }> = {
  info: { color: 'blue', icon: <CloudOutlined />, title: '提示' },
  warning: { color: 'orange', icon: <WarningOutlined />, title: '警告' },
  critical: { color: 'red', icon: <ThunderboltOutlined />, title: '严重' },
}

const alertTypeLabel: Record<WeatherAlertType, string> = {
  high_wind: '大风预警',
  gust: '阵风保护',
  thunderstorm: '雷暴警告',
  low_temperature: '低温预警',
  heavy_rain: '暴雨预警',
}

const actionTakenLabel: Record<string, string> = {
  suggest_rth: '建议返航',
  attitude_protection: '姿态保护已触发',
  reject_takeoff: '起飞已拒绝',
  warn_low_temp: '低温提醒',
  wind_adapt: '风速自适应降速',
}

const WeatherAlertModal: React.FC = () => {
  const dispatch = useAppDispatch()
  const { activeAlerts } = useAppSelector(state => state.weather)
  const [visible, setVisible] = useState(false)
  const [resolving, setResolving] = useState<number | null>(null)

  const allAlerts = useMemo(() => {
    return Object.values(activeAlerts).flat().filter(a => !a.is_resolved)
  }, [activeAlerts])

  useEffect(() => {
    const shouldShow = allAlerts.length > 0
    if (shouldShow !== visible) {
      setVisible(shouldShow)
    }
  }, [allAlerts.length, visible])

  useEffect(() => {
    if (allAlerts.length > 0) {
      const hasCritical = allAlerts.some(a => a.alert_level === 'critical')
      if (hasCritical) {
        try {
          const AudioCtx = (window as any).AudioContext || (window as any).webkitAudioContext
          if (AudioCtx) {
            const ctx = new AudioCtx()
            const osc = ctx.createOscillator()
            const gain = ctx.createGain()
            osc.type = 'square'
            osc.frequency.setValueAtTime(660, ctx.currentTime)
            osc.frequency.setValueAtTime(440, ctx.currentTime + 0.15)
            osc.frequency.setValueAtTime(660, ctx.currentTime + 0.3)
            gain.gain.setValueAtTime(0.2, ctx.currentTime)
            gain.gain.exponentialRampToValueAtTime(0.01, ctx.currentTime + 0.45)
            osc.connect(gain)
            gain.connect(ctx.destination)
            osc.start()
            osc.stop(ctx.currentTime + 0.45)
          }
        } catch {}
      }
    }
  }, [allAlerts.length])

  const handleResolve = useCallback(async (alert: WeatherAlertEvent) => {
    setResolving(alert.id)
    try {
      await resolveWeatherAlert(alert.id)
      dispatch(removeWeatherAlert({ uavId: alert.uav_id, alertId: alert.id }))
    } catch {}
    setResolving(null)
  }, [dispatch])

  const handleClose = useCallback(() => {
    setVisible(false)
  }, [])

  const criticalAlerts = allAlerts.filter(a => a.alert_level === 'critical')
  const warningAlerts = allAlerts.filter(a => a.alert_level === 'warning')
  const infoAlerts = allAlerts.filter(a => a.alert_level === 'info')

  const renderAlertItem = (alert: WeatherAlertEvent) => {
    const config = alertLevelConfig[alert.alert_level]
    return (
      <List.Item
        key={alert.id}
        actions={[
          <Button
            size="small"
            type="link"
            loading={resolving === alert.id}
            onClick={() => handleResolve(alert)}
          >
            解除
          </Button>,
        ]}
      >
        <List.Item.Meta
          avatar={
            <span style={{ fontSize: 20, color: config.color === 'red' ? '#ff4d4f' : config.color === 'orange' ? '#fa8c16' : '#1890ff' }}>
              {config.icon}
            </span>
          }
          title={
            <Space>
              <Tag color={config.color}>{config.title}</Tag>
              <Text strong>{alertTypeLabel[alert.alert_type] || alert.alert_type}</Text>
              {alert.action_taken && (
                <Tag color="purple">{actionTakenLabel[alert.action_taken] || alert.action_taken}</Tag>
              )}
            </Space>
          }
          description={
            <div>
              <Paragraph style={{ marginBottom: 4 }}>{alert.message}</Paragraph>
              <Space size="large">
                {alert.wind_speed > 0 && <Text type="secondary">风速: {alert.wind_speed.toFixed(1)}m/s</Text>}
                {alert.gust_speed > 0 && <Text type="secondary">阵风: {alert.gust_speed.toFixed(1)}m/s</Text>}
                {alert.temperature !== 0 && <Text type="secondary">温度: {alert.temperature.toFixed(1)}°C</Text>}
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
          <CloudFilled style={{ color: '#fa8c16' }} />
          <span>气象预警</span>
          {allAlerts.length > 0 && <Tag color="red">{allAlerts.length}</Tag>}
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
          message="严重气象警告"
          description="存在影响飞行安全的严重气象条件，请立即采取行动"
          style={{ marginBottom: 16 }}
        />
      )}

      {criticalAlerts.length > 0 && (
        <>
          <Text strong style={{ color: '#ff4d4f' }}>严重预警</Text>
          <List
            size="small"
            dataSource={criticalAlerts}
            renderItem={renderAlertItem}
            style={{ marginBottom: 16 }}
          />
        </>
      )}

      {warningAlerts.length > 0 && (
        <>
          <Text strong style={{ color: '#fa8c16' }}>警告</Text>
          <List
            size="small"
            dataSource={warningAlerts}
            renderItem={renderAlertItem}
            style={{ marginBottom: 16 }}
          />
        </>
      )}

      {infoAlerts.length > 0 && (
        <>
          <Text strong style={{ color: '#1890ff' }}>提示</Text>
          <List
            size="small"
            dataSource={infoAlerts}
            renderItem={renderAlertItem}
          />
        </>
      )}
    </Modal>
  )
}

export default WeatherAlertModal
