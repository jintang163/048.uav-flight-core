import React, { useState, useCallback, useEffect } from 'react'
import styled, { keyframes } from 'styled-components'
import { Button, Progress, Tag, Space, Modal, Tooltip, Result, Divider } from 'antd'
import {
  SafetyCertificateOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  RocketOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  WifiOutlined,
  CompassOutlined,
  BarChartOutlined,
  UnlockOutlined,
  LockOutlined,
  ExpandOutlined,
  ClockCircleOutlined,
  InfoCircleOutlined
} from '@ant-design/icons'
import { runPreflightCheck, getPreflightThresholds } from '@/api/preflight'
import { speakAlert } from '@/utils'
import type {
  PreflightCheckResult,
  PreflightCheckItem,
  PreflightCheckStatus,
  PreflightCheckType,
  PreflightCheckThresholds
} from '@/types'

const scanLine = keyframes`
  0% { transform: translateY(-100%); }
  100% { transform: translateY(2000%); }
`

const Container = styled.div`
  width: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const Header = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
`

const Title = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
  gap: 8px;
`

const OverallStatusCard = styled.div<{ $status: PreflightCheckStatus }>`
  padding: 16px;
  border-radius: 8px;
  text-align: center;
  background: ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.08)'
      case 'warning': return 'rgba(250, 173, 20, 0.08)'
      case 'fail': return 'rgba(255, 77, 79, 0.08)'
      default: return 'rgba(255, 255, 255, 0.03)'
    }
  }};
  border: 1px solid ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.3)'
      case 'warning': return 'rgba(250, 173, 20, 0.3)'
      case 'fail': return 'rgba(255, 77, 79, 0.3)'
      default: return 'rgba(255, 255, 255, 0.1)'
    }
  }};
  position: relative;
  overflow: hidden;
`

const ScanningOverlay = styled.div`
  position: absolute;
  inset: 0;
  background: linear-gradient(180deg, transparent, rgba(24, 144, 255, 0.15), transparent);
  animation: ${scanLine} 1.5s linear infinite;
  pointer-events: none;
`

const OverallIcon = styled.div<{ $status: PreflightCheckStatus }>`
  font-size: 36px;
  margin-bottom: 8px;
  color: ${props => {
    switch (props.$status) {
      case 'pass': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fail': return '#ff4d4f'
      default: return '#8c8c8c'
    }
  }};
`

const OverallTitle = styled.div<{ $status: PreflightCheckStatus }>`
  font-size: 18px;
  font-weight: 700;
  margin-bottom: 4px;
  color: ${props => {
    switch (props.$status) {
      case 'pass': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fail': return '#ff4d4f'
      default: return 'rgba(255, 255, 255, 0.9)'
    }
  }};
`

const OverallSummary = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 10px;
`

const StatsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-top: 8px;
`

const StatItem = styled.div`
  text-align: center;
  padding: 4px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 4px;
`

const StatValue = styled.div<{ $color?: string }>`
  font-size: 20px;
  font-weight: 700;
  font-family: 'Courier New', monospace;
  color: ${props => props.$color || 'rgba(255, 255, 255, 0.9)'};
`

const StatLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
  margin-top: 2px;
`

const CheckList = styled.div`
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 320px;
  overflow-y: auto;

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

const CheckItem = styled.div<{ $status: PreflightCheckStatus; $expanded: boolean }>`
  padding: 8px 10px;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s;
  background: ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.05)'
      case 'warning': return 'rgba(250, 173, 20, 0.05)'
      case 'fail': return 'rgba(255, 77, 79, 0.05)'
      default: return 'rgba(255, 255, 255, 0.02)'
    }
  }};
  border-left: 3px solid ${props => {
    switch (props.$status) {
      case 'pass': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fail': return '#ff4d4f'
      default: return 'rgba(255, 255, 255, 0.2)'
    }
  }};
  opacity: ${props => props.$status === 'pass' ? 0.9 : 1};
  border-bottom: 1px solid rgba(255, 255, 255, 0.04);

  &:hover {
    background: rgba(255, 255, 255, 0.06);
  }
`

const CheckItemHeader = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
`

const CheckIcon = styled.div<{ $status: PreflightCheckStatus }>`
  width: 20px;
  height: 20px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${props => {
    switch (props.$status) {
      case 'pass': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fail': return '#ff4d4f'
      default: return '#8c8c8c'
    }
  }};
  flex-shrink: 0;
`

const CheckCategoryIcon = styled.div<{ $color?: string }>`
  width: 18px;
  height: 18px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${props => props.$color || 'rgba(255, 255, 255, 0.5)'};
  font-size: 13px;
  flex-shrink: 0;
`

const CheckName = styled.div`
  font-size: 12px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.85);
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
`

const CheckStatusBadge = styled.div<{ $status: PreflightCheckStatus }>`
  font-size: 10px;
  padding: 1px 6px;
  border-radius: 10px;
  font-weight: 600;
  background: ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.15)'
      case 'warning': return 'rgba(250, 173, 20, 0.15)'
      case 'fail': return 'rgba(255, 77, 79, 0.15)'
      default: return 'rgba(255, 255, 255, 0.1)'
    }
  }};
  color: ${props => {
    switch (props.$status) {
      case 'pass': return '#52c41a'
      case 'warning': return '#faad14'
      case 'fail': return '#ff4d4f'
      default: return 'rgba(255, 255, 255, 0.5)'
    }
  }};
`

const CheckItemDetail = styled.div`
  margin-top: 8px;
  padding: 8px 8px 8px 38px;
  background: rgba(0, 0, 0, 0.2);
  border-radius: 4px;
  font-size: 11px;
`

const DetailRow = styled.div`
  display: flex;
  justify-content: space-between;
  padding: 2px 0;
  color: rgba(255, 255, 255, 0.6);
`

const DetailLabel = styled.span`
  color: rgba(255, 255, 255, 0.4);
`

const DetailValue = styled.span<{ $type?: 'good' | 'bad' | 'warn' }>`
  font-family: 'Courier New', monospace;
  color: ${props => {
    switch (props.$type) {
      case 'good': return '#52c41a'
      case 'bad': return '#ff4d4f'
      case 'warn': return '#faad14'
      default: return 'rgba(255, 255, 255, 0.85)'
    }
  }};
`

const MessageBar = styled.div<{ $status: PreflightCheckStatus }>`
  margin-top: 4px;
  margin-left: 38px;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 10px;
  background: ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.08)'
      case 'warning': return 'rgba(250, 173, 20, 0.08)'
      case 'fail': return 'rgba(255, 77, 79, 0.08)'
      default: return 'rgba(255, 255, 255, 0.03)'
    }
  }};
  color: ${props => {
    switch (props.$status) {
      case 'pass': return 'rgba(82, 196, 26, 0.9)'
      case 'warning': return 'rgba(250, 173, 20, 0.95)'
      case 'fail': return 'rgba(255, 77, 79, 0.95)'
      default: return 'rgba(255, 255, 255, 0.6)'
    }
  }};
`

const BottomActions = styled.div`
  display: flex;
  gap: 8px;
  padding-top: 4px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
`

const getCheckIcon = (type: PreflightCheckType): { icon: React.ReactNode; color: string } => {
  switch (type) {
    case 'gps':
      return { icon: <EnvironmentOutlined />, color: '#52c41a' }
    case 'battery':
      return { icon: <ThunderboltOutlined />, color: '#faad14' }
    case 'imu':
      return { icon: <DashboardOutlined />, color: '#1890ff' }
    case 'storage':
      return { icon: <DatabaseOutlined />, color: '#eb2f96' }
    case 'link':
      return { icon: <WifiOutlined />, color: '#13c2c2' }
    case 'compass':
      return { icon: <CompassOutlined />, color: '#722ed1' }
    case 'barometer':
      return { icon: <BarChartOutlined />, color: '#fa8c16' }
    case 'arm':
      return { icon: <UnlockOutlined />, color: '#1890ff' }
    default:
      return { icon: <InfoCircleOutlined />, color: '#8c8c8c' }
  }
}

const getStatusIcon = (status: PreflightCheckStatus) => {
  switch (status) {
    case 'pass': return <CheckCircleOutlined />
    case 'warning': return <WarningOutlined />
    case 'fail': return <CloseCircleOutlined />
    case 'pending': return <ClockCircleOutlined />
  }
}

const getStatusLabel = (status: PreflightCheckStatus) => {
  switch (status) {
    case 'pass': return '通过'
    case 'warning': return '警告'
    case 'fail': return '不通过'
    case 'pending': return '待检'
  }
}

interface PreflightCheckPanelProps {
  uavId?: number
  uavName?: string
  showTitle?: boolean
  compact?: boolean
  onPassChange?: (canTakeoff: boolean) => void
  onTakeoff?: () => void
}

const PreflightCheckPanel: React.FC<PreflightCheckPanelProps> = ({
  uavId,
  uavName,
  showTitle = true,
  compact = false,
  onPassChange,
  onTakeoff
}) => {
  const [result, setResult] = useState<PreflightCheckResult | null>(null)
  const [isScanning, setIsScanning] = useState(false)
  const [expandedItem, setExpandedItem] = useState<PreflightCheckType | null>(null)
  const [thresholds, setThresholds] = useState<PreflightCheckThresholds | null>(null)
  const [showDetailModal, setShowDetailModal] = useState(false)

  useEffect(() => {
    const fetchThresholds = async () => {
      try {
        const data = await getPreflightThresholds()
        setThresholds(data)
      } catch (e) {
        console.warn('Failed to get preflight thresholds')
      }
    }
    fetchThresholds()
  }, [])

  useEffect(() => {
    if (onPassChange) {
      onPassChange(result?.can_takeoff ?? false)
    }
  }, [result, onPassChange])

  const runCheck = useCallback(async () => {
    if (!uavId) {
      speakAlert('请先选择无人机')
      return
    }
    setIsScanning(true)
    setExpandedItem(null)
    try {
      await new Promise(r => setTimeout(r, 800))
      const data = await runPreflightCheck({ uav_id: uavId })
      setResult(data)

      if (data.overall_status === 'pass') {
        speakAlert('自检通过，具备起飞条件')
      } else if (data.overall_status === 'warning') {
        speakAlert(`自检通过，但有${data.warning_count}项警告`)
      } else {
        speakAlert(`自检未通过，${data.failed_count}项不通过，禁止起飞`)
      }
    } catch (error: any) {
      console.error('Preflight check failed:', error)
    } finally {
      setIsScanning(false)
    }
  }, [uavId])

  useEffect(() => {
    if (uavId && !result) {
      runCheck()
    }
  }, [uavId])

  const failedItems = result?.checks.filter(c => c.status === 'fail') || []
  const warningItems = result?.checks.filter(c => c.status === 'warning') || []

  const getOverallIcon = () => {
    if (isScanning) return <SafetyCertificateOutlined style={{ animation: 'pulse 1s infinite' }} />
    if (!result) return <SafetyCertificateOutlined />
    switch (result.overall_status) {
      case 'pass': return <CheckCircleOutlined />
      case 'warning': return <WarningOutlined />
      case 'fail': return <CloseCircleOutlined />
      default: return <SafetyCertificateOutlined />
    }
  }

  const getOverallTitle = () => {
    if (isScanning) return '正在进行飞行前检查...'
    if (!result) return '尚未进行自检'
    switch (result.overall_status) {
      case 'pass': return '自检全部通过'
      case 'warning': return '自检通过（有警告项）'
      case 'fail': return '自检未通过，禁止起飞'
      default: return '尚未进行自检'
    }
  }

  const getOverallStatus = (): PreflightCheckStatus => {
    if (!result || isScanning) return 'pending'
    return result.overall_status
  }

  if (compact) {
    return (
      <Container>
        {showTitle && (
          <Header>
            <Title>
              <SafetyCertificateOutlined />
              飞行前检查
            </Title>
            <Tooltip title="重新检查">
              <Button
                type="text"
                size="small"
                icon={<ReloadOutlined spin={isScanning} />}
                onClick={runCheck}
                disabled={isScanning || !uavId}
                style={{ color: 'rgba(255,255,255,0.5)' }}
              />
            </Tooltip>
          </Header>
        )}

        {!result && !isScanning ? (
          <OverallStatusCard $status="pending">
            <OverallIcon $status="pending">
              <SafetyCertificateOutlined />
            </OverallIcon>
            <OverallTitle $status="pending" style={{ fontSize: 14 }}>
              点击开始自检
            </OverallTitle>
            <OverallSummary>
              {uavName ? `UAV: ${uavName}` : '请先选择无人机'}
            </OverallSummary>
            <Button
              type="primary"
              icon={<ReloadOutlined />}
              loading={isScanning}
              onClick={runCheck}
              disabled={!uavId}
              block
              size="small"
            >
              开始飞行前检查
            </Button>
          </OverallStatusCard>
        ) : (
          <Space direction="vertical" style={{ width: '100%' }} size={8}>
            <Tag
              color={getOverallStatus() === 'pass' ? 'green' : getOverallStatus() === 'warning' ? 'orange' : 'red'}
              style={{ textAlign: 'center', padding: '4px 12px', fontSize: 12, margin: 0 }}
              onClick={() => setShowDetailModal(true)}
            >
              {getOverallIcon()} {getOverallTitle()}
            </Tag>

            {result && (
              <div style={{ display: 'flex', gap: 4 }}>
                <Progress
                  percent={Math.round((result.passed_count / result.total_count) * 100)}
                  size="small"
                  status={getOverallStatus() === 'fail' ? 'exception' : undefined}
                  strokeColor={
                    getOverallStatus() === 'pass' ? '#52c41a' :
                    getOverallStatus() === 'warning' ? '#faad14' : '#ff4d4f'
                  }
                  showInfo={false}
                  style={{ flex: 1 }}
                />
                <span style={{
                  fontSize: 10,
                  fontFamily: 'monospace',
                  color: 'rgba(255,255,255,0.5)',
                  alignSelf: 'center'
                }}>
                  {result.passed_count}/{result.total_count}
                </span>
              </div>
            )}

            <Space style={{ width: '100%', justifyContent: 'space-between' }} size={4}>
              <Button
                size="small"
                type="text"
                icon={<ExpandOutlined />}
                onClick={() => setShowDetailModal(true)}
                style={{ color: 'rgba(255,255,255,0.5)', fontSize: 10, padding: '0 4px' }}
              >
                详情
              </Button>
              <Button
                size="small"
                icon={<ReloadOutlined />}
                onClick={runCheck}
                loading={isScanning}
                disabled={!uavId}
                style={{ fontSize: 10, padding: '0 8px' }}
              >
                重检
              </Button>
              {onTakeoff && result?.can_takeoff && (
                <Button
                  size="small"
                  type="primary"
                  danger={!result?.can_takeoff}
                  icon={<RocketOutlined />}
                  onClick={onTakeoff}
                  disabled={!result?.can_takeoff}
                  style={{ fontSize: 10, padding: '0 8px' }}
                >
                  起飞
                </Button>
              )}
            </Space>
          </Space>
        )}

        <DetailModal
          visible={showDetailModal}
          onClose={() => setShowDetailModal(false)}
          result={result}
          isScanning={isScanning}
          onRecheck={runCheck}
          onTakeoff={onTakeoff}
          uavName={uavName}
          thresholds={thresholds}
          expandedItem={expandedItem}
          setExpandedItem={setExpandedItem}
          failedItems={failedItems}
          warningItems={warningItems}
        />
      </Container>
    )
  }

  return (
    <Container>
      {showTitle && (
        <Header>
          <Title>
            <SafetyCertificateOutlined />
            飞行前检查
            {uavName && (
              <Tag color="geekblue" style={{ fontSize: 10, padding: '0 6px', marginLeft: 4 }}>
                {uavName}
              </Tag>
            )}
          </Title>
          <Space>
            <Tooltip title="查看阈值">
              <Button
                type="text"
                size="small"
                icon={<InfoCircleOutlined />}
                disabled={!thresholds}
                onClick={() => console.log('Thresholds:', thresholds)}
                style={{ color: 'rgba(255,255,255,0.5)' }}
              />
            </Tooltip>
            <Button
              size="small"
              icon={<ReloadOutlined />}
              loading={isScanning}
              onClick={runCheck}
              disabled={isScanning || !uavId}
              type="primary"
              ghost
            >
              {result ? '重新检查' : '开始自检'}
            </Button>
          </Space>
        </Header>
      )}

      <OverallStatusCard $status={getOverallStatus()}>
        {isScanning && <ScanningOverlay />}
        <OverallIcon $status={getOverallStatus()}>
          {getOverallIcon()}
        </OverallIcon>
        <OverallTitle $status={getOverallStatus()}>
          {getOverallTitle()}
        </OverallTitle>
        {result && (
          <>
            <OverallSummary>
              {result.summary}
            </OverallSummary>
            <StatsRow>
              <StatItem>
                <StatValue $color="#52c41a">{result.passed_count}</StatValue>
                <StatLabel>通过</StatLabel>
              </StatItem>
              <StatItem>
                <StatValue $color="#faad14">{result.warning_count}</StatValue>
                <StatLabel>警告</StatLabel>
              </StatItem>
              <StatItem>
                <StatValue $color="#ff4d4f">{result.failed_count}</StatValue>
                <StatLabel>不通过</StatLabel>
              </StatItem>
            </StatsRow>
          </>
        )}
        {!result && !isScanning && (
          <OverallSummary>
            {uavId ? '点击按钮开始检查起飞条件' : '请先选择无人机'}
          </OverallSummary>
        )}
      </OverallStatusCard>

      {result && (
        <>
          <CheckList>
            {result.checks.map(check => (
              <CheckItemComponent
                key={check.check_type}
                check={check}
                expanded={expandedItem === check.check_type}
                onToggle={() => setExpandedItem(
                  expandedItem === check.check_type ? null : check.check_type
                )}
              />
            ))}
          </CheckList>

          {(failedItems.length > 0 || warningItems.length > 0) && (
            <div style={{
              padding: '8px 10px',
              background: failedItems.length > 0 ? 'rgba(255,77,79,0.06)' : 'rgba(250,173,20,0.06)',
              border: `1px solid ${failedItems.length > 0 ? 'rgba(255,77,79,0.2)' : 'rgba(250,173,20,0.2)'}`,
              borderRadius: 6,
              fontSize: 11
            }}>
              <div style={{
                fontWeight: 600,
                color: failedItems.length > 0 ? '#ff4d4f' : '#faad14',
                marginBottom: 4
              }}>
                {failedItems.length > 0 ? '⚠️ 需要修复的问题' : '⚠️ 注意事项'}
              </div>
              {failedItems.length > 0 && (
                <div style={{ color: '#ff4d4f', marginBottom: 2 }}>
                  <b>不通过项：</b>{failedItems.map(c => c.name).join('、')}
                </div>
              )}
              {warningItems.length > 0 && (
                <div style={{ color: '#faad14' }}>
                  <b>警告项：</b>{warningItems.map(c => c.name).join('、')}
                </div>
              )}
            </div>
          )}
        </>
      )}

      <BottomActions>
        <Button
          icon={<ReloadOutlined />}
          loading={isScanning}
          onClick={runCheck}
          disabled={isScanning || !uavId}
          style={{ flex: 1 }}
          size="small"
        >
          重新检查
        </Button>
        {onTakeoff && (
          <Button
            type="primary"
            danger={!result?.can_takeoff}
            icon={<RocketOutlined />}
            onClick={onTakeoff}
            disabled={!result?.can_takeoff || isScanning}
            style={{ flex: 1 }}
            size="small"
          >
            {result?.can_takeoff ? '允许起飞' : result ? '禁止起飞' : '等待检查'}
          </Button>
        )}
      </BottomActions>
    </Container>
  )
}

interface CheckItemComponentProps {
  check: PreflightCheckItem
  expanded: boolean
  onToggle: () => void
}

const CheckItemComponent: React.FC<CheckItemComponentProps> = ({ check, expanded, onToggle }) => {
  const iconInfo = getCheckIcon(check.check_type)

  const formatDetail = (detail: Record<string, any>): Array<{ label: string; value: string; type?: 'good' | 'bad' | 'warn' }> => {
    const rows: Array<{ label: string; value: string; type?: 'good' | 'bad' | 'warn' }> = []

    const labelMap: Record<string, string> = {
      satellites: '卫星数量',
      fix_type: '定位类型',
      hdop: 'HDOP精度',
      voltage: '总电压',
      current: '电流',
      level_percent: '剩余电量',
      cell_count: '电芯数',
      per_cell_volt: '单节电压',
      soh_percent: '电池健康度',
      cycle_count: '循环次数',
      gyro_roll: '横滚角速度',
      gyro_pitch: '俯仰角速度',
      gyro_yaw: '偏航角速度',
      gyro_total: '陀螺合计',
      accel_roll: '横滚加速度',
      accel_pitch: '俯仰加速度',
      accel_z: '垂直加速度',
      is_stable: '机体稳定',
      is_leveled: '机体水平',
      roll_deg: '横滚角度',
      pitch_deg: '俯仰角度',
      free_mb: '剩余空间',
      total_mb: '总空间',
      used_percent: '已用百分比',
      min_required_mb: '最小要求',
      rssi_dbm: '信号强度',
      link_quality: '链路质量',
      heading: '航向角度',
      variance: '磁偏方差',
      interference: '磁场干扰',
      heading_valid: '航向有效',
      altitude_msl: '海拔高度',
      altitude_rel: '相对高度',
      vertical_speed: '垂直速度',
      altitude_valid: '高度有效',
      climb_rate_valid: '升降速度',
      is_armed: '解锁状态',
      uav_status: '无人机状态',
      no_error: '有无错误',
      is_ground: '在地面'
    }

    Object.entries(detail).forEach(([key, val]) => {
      let value = ''
      let type: 'good' | 'bad' | 'warn' | undefined

      if (typeof val === 'boolean') {
        value = val ? '是' : '否'
        type = val ? 'good' : 'bad'
      } else if (typeof val === 'number') {
        if (key.includes('percent') || key.includes('hdop')) {
          value = val.toFixed(2)
        } else if (key === 'rssi_dbm') {
          value = `${val} dBm`
          type = val > -80 ? 'good' : 'warn'
        } else if (key === 'link_quality' || key === 'level_percent' || key === 'soh_percent') {
          value = `${val.toFixed(0)}%`
          type = val > 70 ? 'good' : val > 40 ? 'warn' : 'bad'
        } else if (key.includes('space') || key === 'free_mb' || key === 'total_mb') {
          value = `${(val / 1024).toFixed(1)} GB`
        } else if (key === 'heading') {
          value = `${val.toFixed(1)}°`
        } else if (key === 'satellites') {
          value = `${val} 颗`
          type = val >= 10 ? 'good' : val >= 6 ? 'warn' : 'bad'
        } else {
          value = val.toFixed(3)
        }
      } else {
        value = String(val)
      }

      rows.push({
        label: labelMap[key] || key,
        value,
        type
      })
    })

    return rows
  }

  return (
    <>
      <CheckItem $status={check.status} $expanded={expanded} onClick={onToggle}>
        <CheckItemHeader>
          <CheckIcon $status={check.status}>
            {getStatusIcon(check.status)}
          </CheckIcon>
          <CheckCategoryIcon $color={iconInfo.color}>
            {iconInfo.icon}
          </CheckCategoryIcon>
          <CheckName title={check.name}>
            {check.name}
          </CheckName>
          <CheckStatusBadge $status={check.status}>
            {getStatusLabel(check.status)}
          </CheckStatusBadge>
        </CheckItemHeader>

        {(check.status === 'fail' || check.status === 'warning') && (
          <MessageBar $status={check.status}>
            {check.message}
          </MessageBar>
        )}

        {expanded && (
          <CheckItemDetail>
            <DetailRow>
              <DetailLabel>要求</DetailLabel>
              <DetailValue>{check.threshold}</DetailValue>
            </DetailRow>
            <DetailRow>
              <DetailLabel>实际值</DetailLabel>
              <DetailValue $type={check.status === 'pass' ? 'good' : check.status === 'fail' ? 'bad' : 'warn'}>
                {check.actual_value}
              </DetailValue>
            </DetailRow>
            <Divider style={{ margin: '6px 0', borderColor: 'rgba(255,255,255,0.05)' }} />
            {formatDetail(check.detail).map((d, i) => (
              <DetailRow key={i}>
                <DetailLabel>{d.label}</DetailLabel>
                <DetailValue $type={d.type}>{d.value}</DetailValue>
              </DetailRow>
            ))}
          </CheckItemDetail>
        )}
      </CheckItem>
    </>
  )
}

interface DetailModalProps {
  visible: boolean
  onClose: () => void
  result: PreflightCheckResult | null
  isScanning: boolean
  onRecheck: () => void
  onTakeoff?: () => void
  uavName?: string
  thresholds: PreflightCheckThresholds | null
  expandedItem: PreflightCheckType | null
  setExpandedItem: (t: PreflightCheckType | null) => void
  failedItems: PreflightCheckItem[]
  warningItems: PreflightCheckItem[]
}

const DetailModal: React.FC<DetailModalProps> = ({
  visible,
  onClose,
  result,
  isScanning,
  onRecheck,
  onTakeoff,
  uavName,
  thresholds,
  expandedItem,
  setExpandedItem,
  failedItems,
  warningItems
}) => {
  if (!result) return null

  const status = result.overall_status

  return (
    <Modal
      title={
        <Space>
          <SafetyCertificateOutlined style={{
            color: status === 'pass' ? '#52c41a' : status === 'warning' ? '#faad14' : '#ff4d4f',
            fontSize: 18
          }} />
          飞行前检查报告
          {uavName && (
            <Tag color="geekblue" style={{ marginLeft: 8 }}>{uavName}</Tag>
          )}
        </Space>
      }
      open={visible}
      onCancel={onClose}
      width={640}
      footer={[
        <Button key="close" onClick={onClose}>关闭</Button>,
        <Button key="recheck" icon={<ReloadOutlined />} onClick={onRecheck} loading={isScanning}>
          重新检查
        </Button>,
        onTakeoff && (
          <Button
            key="takeoff"
            type="primary"
            danger={!result.can_takeoff}
            icon={<RocketOutlined />}
            onClick={() => {
              onTakeoff()
              onClose()
            }}
            disabled={!result.can_takeoff || isScanning}
          >
            {result.can_takeoff ? '确认起飞' : '禁止起飞'}
          </Button>
        )
      ].filter(Boolean)}
    >
      <Result
        status={
          status === 'pass' ? 'success' :
          status === 'warning' ? 'warning' : 'error'
        }
        title={
          status === 'pass' ? '自检通过' :
          status === 'warning' ? '自检通过（有警告）' : '自检未通过'
        }
        subTitle={result.summary}
        style={{ padding: '8px 0 16px' }}
        extra={
          <Space>
            <Progress
              type="dashboard"
              percent={Math.round((result.passed_count / result.total_count) * 100)}
              width={100}
              strokeColor={
                status === 'pass' ? '#52c41a' :
                status === 'warning' ? '#faad14' : '#ff4d4f'
              }
              format={(p) => <span style={{ fontSize: 18, fontWeight: 700 }}>{p}%</span>}
            />
            <div style={{ fontSize: 12 }}>
              <div><Tag color="success">通过 {result.passed_count}</Tag></div>
              <div><Tag color="warning">警告 {result.warning_count}</Tag></div>
              <div><Tag color="error">不通过 {result.failed_count}</Tag></div>
            </div>
          </Space>
        }
      />

      {(failedItems.length > 0 || warningItems.length > 0) && (
        <div style={{
          padding: '12px 16px',
          background: failedItems.length > 0 ? 'rgba(255,77,79,0.05)' : 'rgba(250,173,20,0.05)',
          border: `1px solid ${failedItems.length > 0 ? 'rgba(255,77,79,0.2)' : 'rgba(250,173,20,0.2)'}`,
          borderRadius: 8,
          marginBottom: 16
        }}>
          <div style={{
            fontWeight: 600,
            marginBottom: 8,
            color: failedItems.length > 0 ? '#ff4d4f' : '#faad14',
            fontSize: 13
          }}>
            <WarningOutlined /> 检查要点：
          </div>
          {failedItems.map(item => (
            <div key={item.check_type} style={{
              fontSize: 12,
              color: '#ff4d4f',
              padding: '4px 0',
              paddingLeft: 16
            }}>
              <LockOutlined /> <b>[{item.name}]</b> {item.message}
            </div>
          ))}
          {warningItems.map(item => (
            <div key={item.check_type} style={{
              fontSize: 12,
              color: '#faad14',
              padding: '4px 0',
              paddingLeft: 16
            }}>
              <InfoCircleOutlined /> <b>[{item.name}]</b> {item.message}
            </div>
          ))}
        </div>
      )}

      <div style={{
        fontSize: 12,
        color: 'rgba(255,255,255,0.5)',
        marginBottom: 8,
        display: 'flex',
        justifyContent: 'space-between'
      }}>
        <span>检查项详情（点击展开）</span>
        <span>
          <ClockCircleOutlined /> 耗时：{
            ((new Date(result.finished_at).getTime() - new Date(result.started_at).getTime()) / 1000).toFixed(2)
          }s
        </span>
      </div>

      <div style={{
        display: 'flex',
        flexDirection: 'column',
        gap: 6,
        maxHeight: 360,
        overflowY: 'auto',
        padding: '0 4px 0 0'
      }}>
        {result.checks.map(check => (
          <CheckItemComponent
            key={check.check_type}
            check={check}
            expanded={expandedItem === check.check_type}
            onToggle={() => setExpandedItem(
              expandedItem === check.check_type ? null : check.check_type
            )}
          />
        ))}
      </div>

      {thresholds && (
        <div style={{ marginTop: 16, paddingTop: 12, borderTop: '1px solid rgba(255,255,255,0.06)' }}>
          <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)', marginBottom: 6 }}>
            检查阈值：
          </div>
          <div style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(2, 1fr)',
            gap: '4px 16px',
            fontSize: 10,
            color: 'rgba(255,255,255,0.5)',
            fontFamily: 'monospace'
          }}>
            <span>📡 GPS ≥ {thresholds.min_satellites}颗</span>
            <span>🔋 电压 ≥ {thresholds.min_voltage}V</span>
            <span>📍 HDOP ≤ {thresholds.max_hdop}</span>
            <span>💾 存储 ≥ {(thresholds.min_storage_space_mb / 1024).toFixed(1)}GB</span>
            <span>📶 信号 ≥ {thresholds.min_signal_strength}dBm</span>
            <span>🔗 LQ ≥ {thresholds.min_link_quality}%</span>
          </div>
        </div>
      )}
    </Modal>
  )
}

export default PreflightCheckPanel
