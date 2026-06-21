import React, { useEffect, useMemo, useState } from 'react'
import styled from 'styled-components'
import {
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Space,
  Button,
  Tooltip,
  Divider,
  Slider,
  InputNumber,
  message,
  Badge,
  Typography,
  Spin
} from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  HomeOutlined,
  SafetyCertificateOutlined,
  ReloadOutlined,
  SettingOutlined,
  DashboardOutlined,
  FireOutlined,
  ApiOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { fetchMotorStatuses, fetchMotorFailureState } from '@/store/slices/motor'
import * as motorApi from '@/api/motor'
import type { MotorStatus, MotorStatusType } from '@/types'

const { Text } = Typography

const Container = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-head {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    min-height: 44px;
  }

  .ant-card-body {
    padding: 16px;
  }
`

const HexDiagram = styled.div`
  position: relative;
  width: 280px;
  height: 280px;
  margin: 0 auto 16px;
`

const MotorDot = styled.div<{
  $status: MotorStatusType
  $index: number
  $x: number
  $y: number
}>`
  position: absolute;
  left: ${props => props.$x}px;
  top: ${props => props.$y}px;
  width: 56px;
  height: 56px;
  border-radius: 50%;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  transform: translate(-50%, -50%);
  background: ${props =>
    props.$status === 'fault' ? 'rgba(255, 77, 79, 0.2)' :
    props.$status === 'warning' ? 'rgba(250, 173, 20, 0.2)' :
    props.$status === 'normal' ? 'rgba(82, 196, 26, 0.15)' :
    'rgba(140, 140, 140, 0.15)'};
  border: 2px solid ${props =>
    props.$status === 'fault' ? '#ff4d4f' :
    props.$status === 'warning' ? '#faad14' :
    props.$status === 'normal' ? '#52c41a' :
    '#8c8c8c'};
  ${props => props.$status === 'fault' ? 'box-shadow: 0 0 12px rgba(255, 77, 79, 0.5); animation: pulse 1s infinite;' : ''}
  transition: all 0.3s;
  cursor: pointer;

  &:hover {
    transform: translate(-50%, -50%) scale(1.1);
  }
`

const MotorLabel = styled.div`
  font-size: 10px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  line-height: 1;
`

const MotorRPM = styled.div`
  font-size: 9px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1;
  margin-top: 2px;
  font-family: 'Courier New', monospace;
`

const CenterInfo = styled.div`
  position: absolute;
  left: 50%;
  top: 50%;
  transform: translate(-50%, -50%);
  text-align: center;
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
`

const CenterValue = styled.div`
  font-size: 20px;
  font-weight: 700;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const StatusGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-top: 12px;
`

const StatusCell = styled.div<{ $status: MotorStatusType }>`
  background: ${props =>
    props.$status === 'fault' ? 'rgba(255, 77, 79, 0.08)' :
    props.$status === 'warning' ? 'rgba(250, 173, 20, 0.08)' :
    'rgba(255, 255, 255, 0.03)'};
  border: 1px solid ${props =>
    props.$status === 'fault' ? 'rgba(255, 77, 79, 0.2)' :
    props.$status === 'warning' ? 'rgba(250, 173, 20, 0.2)' :
    'rgba(255, 255, 255, 0.06)'};
  border-radius: 6px;
  padding: 8px;
  text-align: center;
`

const CellLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 4px;
`

const CellValue = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.95);
  font-family: 'Courier New', monospace;
`

const PIDSection = styled.div`
  margin-top: 16px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
`

interface MotorStatusPanelProps {
  uavId?: string
}

const MotorStatusPanel: React.FC<MotorStatusPanelProps> = ({ uavId }) => {
  const dispatch = useAppDispatch()
  const { motorStatuses, failureStates, loading } = useAppSelector(state => state.motor)
  const [pidVisible, setPIDVisible] = useState(false)
  const [pidValues, setPIDValues] = useState({ p: 1.0, i: 1.0, d: 1.0 })
  const [actionLoading, setActionLoading] = useState<string | null>(null)

  useEffect(() => {
    if (uavId) {
      dispatch(fetchMotorStatuses(uavId))
      dispatch(fetchMotorFailureState(uavId))
    }
  }, [dispatch, uavId])

  const motors = useMemo(() => {
    if (!uavId || !motorStatuses[uavId]) return []
    return motorStatuses[uavId].sort((a, b) => a.motor_index - b.motor_index)
  }, [uavId, motorStatuses])

  const failureState = useMemo(() => {
    if (!uavId) return null
    return failureStates[uavId]
  }, [uavId, failureStates])

  const motorCount = motors.length || failureState?.motor_count || 6

  const getMotorPosition = (index: number, count: number): { x: number; y: number } => {
    const cx = 140
    const cy = 140
    const r = 100
    const angle = (index * 2 * Math.PI) / count - Math.PI / 2
    return {
      x: cx + r * Math.cos(angle),
      y: cy + r * Math.sin(angle)
    }
  }

  const failedCount = motors.filter(m => m.status === 'fault').length
  const normalCount = motors.filter(m => m.status === 'normal').length
  const avgTemp = motors.length > 0
    ? Math.round(motors.reduce((sum, m) => sum + m.temperature, 0) / motors.length)
    : 0
  const avgRPM = motors.filter(m => m.status === 'normal').length > 0
    ? Math.round(motors.filter(m => m.status === 'normal').reduce((sum, m) => sum + m.rpm, 0) / motors.filter(m => m.status === 'normal').length)
    : 0

  const handleRefresh = () => {
    if (uavId) {
      dispatch(fetchMotorStatuses(uavId))
      dispatch(fetchMotorFailureState(uavId))
    }
  }

  const handleEmergencyRTH = async () => {
    if (!uavId) return
    setActionLoading('rth')
    try {
      await motorApi.emergencyRTH(uavId)
      message.success('紧急返航指令已发送')
    } catch {
      message.error('指令发送失败')
    }
    setActionLoading(null)
  }

  const handleEmergencyLand = async () => {
    if (!uavId) return
    setActionLoading('land')
    try {
      await motorApi.emergencyLand(uavId)
      message.success('紧急降落指令已发送')
    } catch {
      message.error('指令发送失败')
    }
    setActionLoading(null)
  }

  const handlePIDSubmit = async () => {
    if (!uavId) return
    setActionLoading('pid')
    try {
      await motorApi.manualPIDAdjustment(uavId, {
        p_gain: pidValues.p,
        i_gain: pidValues.i,
        d_gain: pidValues.d
      })
      message.success('PID调整指令已发送')
      setPIDVisible(false)
    } catch {
      message.error('PID调整失败')
    }
    setActionLoading(null)
  }

  const getStatusColor = (status: MotorStatusType) => {
    switch (status) {
      case 'normal': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fault': return '#ff4d4f'
      default: return '#8c8c8c'
    }
  }

  return (
    <Container
      title={
        <Space>
          <DashboardOutlined style={{ color: '#1890ff' }} />
          <span>断桨保护监控</span>
          {failedCount > 0 && (
            <Badge count={failedCount} size="small">
              <Tag color="error" icon={<WarningOutlined />}>失效</Tag>
            </Badge>
          )}
        </Space>
      }
      extra={
        <Space>
          <Tooltip title="刷新">
            <Button
              type="text"
              size="small"
              icon={<ReloadOutlined />}
              onClick={handleRefresh}
              loading={loading}
            />
          </Tooltip>
        </Space>
      }
    >
      {!uavId ? (
        <div style={{ textAlign: 'center', padding: '40px 0', color: 'rgba(255,255,255,0.5)' }}>
          <ApiOutlined style={{ fontSize: 36, marginBottom: 8 }} />
          <div>请先选择无人机</div>
        </div>
      ) : (
        <>
          <HexDiagram>
            {Array.from({ length: motorCount }, (_, i) => {
              const motor = motors.find(m => m.motor_index === i)
              const pos = getMotorPosition(i, motorCount)
              const status: MotorStatusType = motor?.status || 'offline'
              const rpm = motor?.rpm || 0
              const isFailed = failureState?.failed_motors?.includes(i)

              return (
                <Tooltip
                  key={i}
                  title={
                    motor ? (
                      <div>
                        <div>电机 #{i + 1} - {status.toUpperCase()}</div>
                        <div>RPM: {rpm} | 温度: {motor.temperature}°C</div>
                        <div>电压: {motor.voltage?.toFixed(1)}V | 电流: {motor.current?.toFixed(1)}A</div>
                        {motor.fault_flags > 0 && <div>故障标志: 0x{motor.fault_flags.toString(16).toUpperCase()}</div>}
                      </div>
                    ) : (
                      `电机 #${i + 1} - 离线`
                    )
                  }
                >
                  <MotorDot $status={status} $index={i} $x={pos.x} $y={pos.y}>
                    <MotorLabel>M{i + 1}</MotorLabel>
                    <MotorRPM>{rpm > 0 ? rpm : '—'}</MotorRPM>
                  </MotorDot>
                </Tooltip>
              )
            })}
            <CenterInfo>
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>{motorCount}旋翼</div>
              <CenterValue>{normalCount}/{motorCount}</CenterValue>
              <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>正常</div>
            </CenterInfo>
          </HexDiagram>

          {failureState && failureState.failed_motors.length > 0 && (
            <Alert
              type="error"
              showIcon
              icon={<WarningOutlined />}
              message="电机失效"
              description={
                <Space wrap>
                  <span>失效电机: {failureState.failed_motors.map(i => `#${i + 1}`).join(', ')}</span>
                  {failureState.pid_adjusted && <Tag color="blue">PID已调整</Tag>}
                  {failureState.rth_triggered && <Tag color="orange">返航已触发</Tag>}
                </Space>
              }
              style={{ marginBottom: 12, background: 'rgba(255,77,79,0.05)', border: '1px solid rgba(255,77,79,0.2)' }}
            />
          )}

          <Row gutter={[16, 12]} style={{ marginBottom: 12 }}>
            <Col span={8}>
              <Statistic
                title={<span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)' }}><FireOutlined /> 平均温度</span>}
                value={avgTemp}
                suffix="°C"
                valueStyle={{ color: avgTemp > 80 ? '#ff4d4f' : '#fff', fontSize: 18 }}
              />
            </Col>
            <Col span={8}>
              <Statistic
                title={<span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)' }}><DashboardOutlined /> 平均RPM</span>}
                value={avgRPM}
                valueStyle={{ fontSize: 18 }}
              />
            </Col>
            <Col span={8}>
              <Statistic
                title={<span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)' }}><SafetyCertificateOutlined /> 冗余状态</span>}
                valueRender={() => (
                  <Tag color={failedCount === 0 ? 'success' : failedCount <= 1 ? 'warning' : 'error'}>
                    {failedCount === 0 ? '正常' : failedCount <= 1 ? '降级运行' : '危险'}
                  </Tag>
                )}
              />
            </Col>
          </Row>

          {motors.length > 0 && (
            <StatusGrid>
              {motors.map(motor => (
                <StatusCell key={motor.motor_index} $status={motor.status}>
                  <CellLabel>M{motor.motor_index + 1}</CellLabel>
                  <CellValue style={{ color: getStatusColor(motor.status) }}>
                    {motor.rpm}
                  </CellValue>
                  <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)' }}>
                    {motor.temperature}°C
                  </div>
                </StatusCell>
              ))}
            </StatusGrid>
          )}

          <Divider style={{ margin: '16px 0 12px' }} />

          <Space wrap>
            <Button
              type="primary"
              danger={failedCount > 0}
              icon={<HomeOutlined />}
              onClick={handleEmergencyRTH}
              loading={actionLoading === 'rth'}
              disabled={!uavId}
            >
              紧急返航
            </Button>
            <Button
              danger
              icon={<ThunderboltOutlined />}
              onClick={handleEmergencyLand}
              loading={actionLoading === 'land'}
              disabled={!uavId}
            >
              紧急降落
            </Button>
            <Button
              icon={<SettingOutlined />}
              onClick={() => setPIDVisible(!pidVisible)}
            >
              PID调整
            </Button>
          </Space>

          {pidVisible && (
            <PIDSection>
              <div style={{ marginBottom: 12 }}>
                <div style={{ marginBottom: 8, fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>
                  P增益 (比例)
                </div>
                <Slider
                  min={0.5} max={2.0} step={0.01}
                  value={pidValues.p}
                  onChange={v => setPIDValues({ ...pidValues, p: v })}
                  marks={{ 0.5: '0.5', 1: '1.0', 2: '2.0' }}
                />
              </div>
              <div style={{ marginBottom: 12 }}>
                <div style={{ marginBottom: 8, fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>
                  I增益 (积分)
                </div>
                <Slider
                  min={0.5} max={2.0} step={0.01}
                  value={pidValues.i}
                  onChange={v => setPIDValues({ ...pidValues, i: v })}
                  marks={{ 0.5: '0.5', 1: '1.0', 2: '2.0' }}
                />
              </div>
              <div style={{ marginBottom: 12 }}>
                <div style={{ marginBottom: 8, fontSize: 13, color: 'rgba(255,255,255,0.7)' }}>
                  D增益 (微分)
                </div>
                <Slider
                  min={0.5} max={2.0} step={0.01}
                  value={pidValues.d}
                  onChange={v => setPIDValues({ ...pidValues, d: v })}
                  marks={{ 0.5: '0.5', 1: '1.0', 2: '2.0' }}
                />
              </div>
              <Button
                type="primary"
                block
                onClick={handlePIDSubmit}
                loading={actionLoading === 'pid'}
              >
                发送PID参数
              </Button>
            </PIDSection>
          )}
        </>
      )}
    </Container>
  )
}

export default MotorStatusPanel
