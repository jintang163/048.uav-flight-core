import React, { useState, useEffect, useRef, useMemo } from 'react'
import styled, { keyframes, css } from 'styled-components'
import { Card, Tooltip, Tag, Progress, Badge, Space } from 'antd'
import {
  WifiOutlined,
  SignalOutlined,
  MobileOutlined,
  SwapOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  ReloadOutlined,
  GlobalOutlined,
  SafetyCertificateOutlined
} from '@ant-design/icons'
import type { LinkStatus, LinkType } from '@/types/link'
import { LinkTypeText, LinkStateText, getRSSIColor, getRSSILevel } from '@/types/link'
import { getLinkStatus } from '@/api/link'
import { useAppSelector, useAppDispatch } from '@/store'
import {
  selectCurrentLink,
  selectIsLinkSwitching,
  selectLinkSwitchCount,
  selectLastSwitchTime,
  setSwitchingComplete
} from '@/store/slices/link'
import { formatFileSize } from '@/utils'

const pulse = keyframes`
  0%, 100% {
    opacity: 1;
    transform: scale(1);
  }
  50% {
    opacity: 0.6;
    transform: scale(1.05);
  }
`

const signalJump = keyframes`
  0%, 100% {
    transform: scaleY(1);
  }
  50% {
    transform: scaleY(1.2);
  }
`

const switchFlash = keyframes`
  0% {
    box-shadow: 0 0 0 0 rgba(24, 144, 255, 0.7);
  }
  70% {
    box-shadow: 0 0 0 15px rgba(24, 144, 255, 0);
  }
  100% {
    box-shadow: 0 0 0 0 rgba(24, 144, 255, 0);
  }
`

const numberRoll = keyframes`
  0% {
    transform: translateY(-100%);
    opacity: 0;
  }
  100% {
    transform: translateY(0);
    opacity: 1;
  }
`

const slideIn = keyframes`
  from {
    opacity: 0;
    transform: translateX(-10px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
`

const NetworkIconGlow = keyframes`
  0%, 100% {
    filter: drop-shadow(0 0 2px currentColor);
  }
  50% {
    filter: drop-shadow(0 0 8px currentColor);
  }
`

const Container = styled(Card)<{ $switching: boolean }>`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  transition: all 0.3s ease;

  ${props => props.$switching && css`
    animation: ${switchFlash} 1.5s ease-out;
  `}

  .ant-card-body {
    padding: 16px;
  }
`

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`

const Title = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
`

const UpdateTime = styled.span`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.4);
  font-weight: normal;
  margin-left: 8px;
`

const ActiveLinkBadge = styled(Tag)<{ $switching: boolean }>`
  font-size: 12px;
  font-weight: 500;
  transition: all 0.3s ease;

  ${props => props.$switching && css`
    animation: ${pulse} 0.5s ease-in-out infinite;
  `}
`

const LinkRow = styled.div<{ $active: boolean; $switching: boolean }>`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  margin-bottom: 8px;
  border: 1px solid transparent;
  transition: all 0.3s ease;

  &:last-child {
    margin-bottom: 0;
  }

  ${props => props.$active && css`
    background: rgba(24, 144, 255, 0.1);
    border-color: rgba(24, 144, 255, 0.3);
  `}

  ${props => props.$active && props.$switching && css`
    animation: ${switchFlash} 1.5s ease-out;
  `}
`

const LinkIcon = styled.div<{ $color: string; $connected: boolean }>`
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: ${props => props.$color}20;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${props => props.$color};
  font-size: 18px;
  transition: all 0.3s ease;
  position: relative;

  ${props => !props.$connected && css`
    opacity: 0.4;
  `}

  ${props => props.$connected && css`
    animation: ${pulse} 2s ease-in-out infinite;
  `}
`

const LinkInfo = styled.div`
  flex: 1;
  min-width: 0;
`

const LinkName = styled.div`
  font-size: 13px;
  font-weight: 500;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
`

const LinkDetails = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
  margin-top: 4px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  flex-wrap: wrap;
`

const SignalBars = styled.div`
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 16px;
`

const SignalBar = styled.div<{ $active: boolean; $color: string; $index: number }>`
  width: 3px;
  border-radius: 1px;
  background: ${props => props.$active ? props.$color : 'rgba(255,255,255,0.2)'};
  transition: height 0.3s ease, background 0.3s ease;
  transform-origin: bottom;

  ${props => props.$active && css`
    animation: ${signalJump} 1.5s ease-in-out infinite;
    animation-delay: ${props.$index * 0.1}s;
  `}
`

const LatencyBadge = styled.div<{ $value: number }>`
  padding: 2px 8px;
  border-radius: 4px;
  background: ${props => props.$value < 100 ? 'rgba(82, 196, 26, 0.2)' : props.$value < 200 ? 'rgba(250, 173, 20, 0.2)' : 'rgba(255, 77, 79, 0.2)'};
  color: ${props => props.$value < 100 ? '#52c41a' : props.$value < 200 ? '#faad14' : '#ff4d4f'};
  font-size: 11px;
  font-weight: 500;
  font-family: 'Courier New', monospace;
  transition: all 0.3s ease;
`

const PacketLossBadge = styled.div<{ $value: number }>`
  padding: 2px 8px;
  border-radius: 4px;
  background: ${props => props.$value < 1 ? 'rgba(82, 196, 26, 0.2)' : props.$value < 5 ? 'rgba(250, 173, 20, 0.2)' : 'rgba(255, 77, 79, 0.2)'};
  color: ${props => props.$value < 1 ? '#52c41a' : props.$value < 5 ? '#faad14' : '#ff4d4f'};
  font-size: 11px;
  font-weight: 500;
  font-family: 'Courier New', monospace;
  transition: all 0.3s ease;
`

const RSSIProgress = styled(Progress)`
  width: 100px;

  .ant-progress-text {
    color: rgba(255,255,255,0.6) !important;
    font-size: 11px;
    font-family: 'Courier New', monospace;
  }
`

const NetworkTypeTag = styled(Tag)<{ $type: string }>`
  font-size: 10px;
  padding: 0 6px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 3px;

  .anticon {
    animation: ${NetworkIconGlow} 2s ease-in-out infinite;
  }

  ${props => {
    const type = props.$type.toUpperCase()
    if (type.includes('5G')) return css`
      background: rgba(114, 46, 209, 0.2) !important;
      color: #722ed1 !important;
      border-color: rgba(114, 46, 209, 0.3) !important;
    `
    if (type.includes('LTE') || type.includes('4G')) return css`
      background: rgba(19, 194, 194, 0.2) !important;
      color: #13c2c2 !important;
      border-color: rgba(19, 194, 194, 0.3) !important;
    `
    return css`
      background: rgba(250, 173, 20, 0.2) !important;
      color: #faad14 !important;
      border-color: rgba(250, 173, 20, 0.3) !important;
    `
  }}
`

const AutoSwitchIndicator = styled.div`
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
`

const QuickInfoSection = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.05);
`

const InfoCard = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  padding: 8px 10px;
  text-align: center;
  animation: ${slideIn} 0.3s ease-out;
`

const InfoLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
  margin-bottom: 4px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
`

const InfoValue = styled.div`
  font-size: 13px;
  font-weight: 600;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const SwitchingOverlay = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.6);
  backdrop-filter: blur(2px);
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  z-index: 10;
  animation: ${slideIn} 0.3s ease-out;
`

const SwitchingText = styled.div`
  color: #1890ff;
  font-size: 13px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 8px;

  .anticon {
    animation: ${numberRoll} 1s linear infinite;
  }
`

const NumberRollWrapper = styled.span`
  display: inline-block;
  animation: ${numberRoll} 0.3s ease-out;
`

const getRelativeTime = (timestamp: number): string => {
  const now = Date.now()
  const diff = Math.floor((now - timestamp) / 1000)

  if (diff < 1) return '刚刚'
  if (diff < 60) return `${diff}秒前`
  if (diff < 3600) return `${Math.floor(diff / 60)}分钟前`
  if (diff < 86400) return `${Math.floor(diff / 3600)}小时前`
  return `${Math.floor(diff / 86400)}天前`
}

const parseBytes = (value: string | number | undefined): number => {
  if (!value) return 0
  if (typeof value === 'number') return value
  const parsed = parseInt(value, 10)
  return isNaN(parsed) ? 0 : parsed
}

const LinkStatusIndicator: React.FC<{
  uavId: string
  refreshInterval?: number
  showDetails?: boolean
}> = ({ uavId, refreshInterval = 3000, showDetails = true }) => {
  const dispatch = useAppDispatch()
  const reduxLinkStatus = useAppSelector(selectCurrentLink)
  const isSwitching = useAppSelector(selectIsLinkSwitching)
  const switchCount = useAppSelector(selectLinkSwitchCount)
  const lastSwitchTime = useAppSelector(selectLastSwitchTime)
  const telemetryConnected = useAppSelector(state => state.telemetry.connected)

  const [httpLinkStatus, setHttpLinkStatus] = useState<LinkStatus | null>(null)
  const [loading, setLoading] = useState(false)
  const [lastUpdateTime, setLastUpdateTime] = useState<number>(Date.now())
  const [displayLatency, setDisplayLatency] = useState(0)
  const [displayPacketLoss, setDisplayPacketLoss] = useState(0)
  const switchTimerRef = useRef<number | null>(null)
  const prevLatencyRef = useRef(0)
  const prevPacketLossRef = useRef(0)

  const useRealtimeData = telemetryConnected && reduxLinkStatus !== null
  const linkStatus = useRealtimeData ? reduxLinkStatus : httpLinkStatus

  useEffect(() => {
    if (!linkStatus) return

    if (linkStatus.latency_ms !== prevLatencyRef.current) {
      prevLatencyRef.current = linkStatus.latency_ms
      setDisplayLatency(linkStatus.latency_ms)
    }
    if (linkStatus.packet_loss !== prevPacketLossRef.current) {
      prevPacketLossRef.current = linkStatus.packet_loss
      setDisplayPacketLoss(linkStatus.packet_loss)
    }
  }, [linkStatus?.latency_ms, linkStatus?.packet_loss])

  const loadStatus = async () => {
    if (!uavId || useRealtimeData) return
    try {
      setLoading(true)
      const status = await getLinkStatus(uavId)
      setHttpLinkStatus(status)
      setLastUpdateTime(Date.now())
    } catch (error) {
      // 静默失败，保持上一次的状态
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    if (useRealtimeData && reduxLinkStatus) {
      setLastUpdateTime(reduxLinkStatus.timestamp || Date.now())
    }
  }, [reduxLinkStatus, useRealtimeData])

  useEffect(() => {
    loadStatus()
    if (!useRealtimeData) {
      const interval = setInterval(loadStatus, refreshInterval)
      return () => clearInterval(interval)
    }
  }, [uavId, refreshInterval, useRealtimeData])

  useEffect(() => {
    if (isSwitching) {
      if (switchTimerRef.current) {
        clearTimeout(switchTimerRef.current)
      }
      switchTimerRef.current = window.setTimeout(() => {
        dispatch(setSwitchingComplete())
      }, 1500)
    }
    return () => {
      if (switchTimerRef.current) {
        clearTimeout(switchTimerRef.current)
      }
    }
  }, [isSwitching, dispatch])

  const renderSignalBars = (rssi: number, type: LinkType, connected: boolean) => {
    const level = getRSSILevel(rssi, type)
    const color = getRSSIColor(rssi, type)
    const bars = [4, 8, 12, 16]

    return (
      <SignalBars>
        {bars.map((height, index) => (
          <SignalBar
            key={index}
            $active={connected && index < level}
            $color={color}
            $index={index}
            style={{ height }}
          />
        ))}
      </SignalBars>
    )
  }

  const getNetworkIcon = (networkType: string) => {
    const type = networkType.toUpperCase()
    if (type.includes('5G')) return <GlobalOutlined />
    if (type.includes('LTE') || type.includes('4G')) return <MobileOutlined />
    return <SignalOutlined />
  }

  const renderLinkRow = (type: LinkType, status: LinkStatus) => {
    const isRadio = type === LinkType.RADIO
    const connected = isRadio ? status.radio_connected : status.lte_connected
    const rssi = isRadio ? status.radio_rssi : status.lte_rssi
    const state = isRadio ? status.radio_state : status.lte_state
    const isActive = status.active_link === type
    const color = getRSSIColor(rssi, type)
    const rssiPercent = Math.max(0, Math.min(100, ((rssi + 120) / 60) * 100))
    const networkType = !isRadio ? status.lte_network_type : ''

    return (
      <LinkRow key={type} $active={isActive} $switching={isSwitching && isActive}>
        <LinkIcon $color={color} $connected={connected}>
          {isRadio ? <SignalOutlined /> : <MobileOutlined />}
        </LinkIcon>
        <LinkInfo>
          <LinkName>
            {LinkTypeText[type]}
            {connected ? (
              <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 12 }} />
            ) : (
              <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 12 }} />
            )}
            {isActive && (
              <Tag color="blue" style={{ fontSize: 10, padding: '0 6px', marginLeft: 4 }}>
                当前
              </Tag>
            )}
          </LinkName>
          <LinkDetails>
            <Tooltip title={`信号强度: ${rssi} dBm`}>
              {renderSignalBars(rssi, type, connected)}
            </Tooltip>
            <RSSIProgress
              percent={rssiPercent}
              showInfo
              format={() => `${rssi} dBm`}
              strokeColor={color}
              size="small"
            />
            <span style={{ color: 'rgba(255,255,255,0.5)' }}>
              {LinkStateText[state]}
            </span>
            {!isRadio && networkType && (
              <NetworkTypeTag $type={networkType}>
                {getNetworkIcon(networkType)}
                {networkType}
              </NetworkTypeTag>
            )}
          </LinkDetails>
          {showDetails && (
            <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
              <LatencyBadge $value={displayLatency}>
                <ClockCircleOutlined style={{ marginRight: 2 }} />
                <NumberRollWrapper key={displayLatency}>
                  {displayLatency}
                </NumberRollWrapper>
                {' '}ms
              </LatencyBadge>
              <PacketLossBadge $value={displayPacketLoss}>
                丢包{' '}
                <NumberRollWrapper key={displayPacketLoss}>
                  {displayPacketLoss.toFixed(1)}
                </NumberRollWrapper>
                %
              </PacketLossBadge>
            </div>
          )}
        </LinkInfo>
      </LinkRow>
    )
  }

  const totalBytesSent = useMemo(() => {
    return parseBytes(linkStatus?.bytes_sent)
  }, [linkStatus?.bytes_sent])

  const totalBytesReceived = useMemo(() => {
    return parseBytes(linkStatus?.bytes_received)
  }, [linkStatus?.bytes_received])

  const updateTimeDisplay = useMemo(() => {
    return getRelativeTime(lastUpdateTime)
  }, [lastUpdateTime])

  if (!linkStatus) {
    return (
      <Container loading={loading} $switching={false}>
        <Header>
          <Title>
            <WifiOutlined style={{ color: '#1890ff' }} />
            通信链路
          </Title>
          <Tag color="default">未连接</Tag>
        </Header>
        <div style={{ textAlign: 'center', padding: '24px 0', color: 'rgba(255,255,255,0.4)' }}>
          暂无链路数据
        </div>
      </Container>
    )
  }

  return (
    <Container loading={loading && !useRealtimeData} $switching={isSwitching}>
      <Header>
        <Title>
          <WifiOutlined style={{ color: '#1890ff' }} />
          通信链路
          <UpdateTime>
            {useRealtimeData ? '实时' : ''}
            {updateTimeDisplay}
          </UpdateTime>
        </Title>
        <Space>
          <ActiveLinkBadge
            color={linkStatus.active_link === 1 ? 'blue' : 'cyan'}
            $switching={isSwitching}
          >
            <SwapOutlined style={{ marginRight: 4 }} />
            {isSwitching ? '切换中...' : LinkTypeText[linkStatus.active_link]}
          </ActiveLinkBadge>
          <Badge
            status={linkStatus.radio_connected || linkStatus.lte_connected ? 'processing' : 'error'}
            text={linkStatus.radio_connected || linkStatus.lte_connected ? '在线' : '离线'}
          />
        </Space>
      </Header>

      <div style={{ position: 'relative' }}>
        {isSwitching && (
          <SwitchingOverlay>
            <SwitchingText>
              <ReloadOutlined spin />
              链路切换中...
            </SwitchingText>
          </SwitchingOverlay>
        )}

        {renderLinkRow(1 as LinkType, linkStatus)}
        {renderLinkRow(2 as LinkType, linkStatus)}
      </div>

      <QuickInfoSection>
        <InfoCard>
          <InfoLabel>
            <ArrowUpOutlined style={{ color: '#52c41a' }} />
            上行流量
          </InfoLabel>
          <InfoValue style={{ color: '#52c41a' }}>
            {formatFileSize(totalBytesSent)}
          </InfoValue>
        </InfoCard>
        <InfoCard>
          <InfoLabel>
            <ArrowDownOutlined style={{ color: '#1890ff' }} />
            下行流量
          </InfoLabel>
          <InfoValue style={{ color: '#1890ff' }}>
            {formatFileSize(totalBytesReceived)}
          </InfoValue>
        </InfoCard>
        <InfoCard>
          <InfoLabel>
            <SwapOutlined style={{ color: '#722ed1' }} />
            切换次数
          </InfoLabel>
          <InfoValue style={{ color: '#722ed1' }}>
            {switchCount} 次
          </InfoValue>
        </InfoCard>
      </QuickInfoSection>

      <AutoSwitchIndicator>
        {linkStatus.auto_switch_enabled ? (
          <>
            <SafetyCertificateOutlined style={{ color: '#52c41a' }} />
            <span>自动切换已启用</span>
          </>
        ) : (
          <>
            <CloseCircleOutlined style={{ color: 'rgba(255,255,255,0.4)' }} />
            <span style={{ color: 'rgba(255,255,255,0.4)' }}>自动切换已禁用</span>
          </>
        )}
      </AutoSwitchIndicator>
    </Container>
  )
}

export default LinkStatusIndicator
