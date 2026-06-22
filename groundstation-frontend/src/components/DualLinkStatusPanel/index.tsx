import React from 'react'
import styled, { keyframes, css } from 'styled-components'
import { Card, Tag, Switch, Progress, Space, Tooltip, Button, Divider, Badge } from 'antd'
import {
  WifiOutlined,
  MobileOutlined,
  SignalOutlined,
  SwapOutlined,
  SafetyCertificateOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  ArrowLeftOutlined,
  ArrowRightOutlined,
  RocketOutlined
} from '@ant-design/icons'
import type { CockpitLinkStatus } from '@/types'
import { LinkType, LinkState, LinkTypeText, LinkStateText, getRSSIColor, getRSSILevel } from '@/types/link'

const pulse = keyframes`
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
`

const switchFlash = keyframes`
  0% { box-shadow: 0 0 0 0 rgba(24, 144, 255, 0.6); }
  70% { box-shadow: 0 0 0 12px rgba(24, 144, 255, 0); }
  100% { box-shadow: 0 0 0 0 rgba(24, 144, 255, 0); }
`

const signalJump = keyframes`
  0%, 100% { transform: scaleY(1); }
  50% { transform: scaleY(1.2); }
`

const Container = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

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

const LinkCard = styled.div<{ $active: boolean; $connected: boolean }>`
  border-radius: 10px;
  padding: 14px;
  margin-bottom: 10px;
  background: rgba(255, 255, 255, 0.02);
  border: 1px solid rgba(255, 255, 255, 0.08);
  transition: all 0.3s ease;
  position: relative;

  &:last-child {
    margin-bottom: 0;
  }

  ${props => props.$active && css`
    background: rgba(24, 144, 255, 0.08);
    border-color: rgba(24, 144, 255, 0.3);
    animation: ${switchFlash} 2s ease-out;
  `}

  ${props => !props.$connected && css`
    opacity: 0.5;
  `}
`

const LinkHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
`

const LinkName = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #fff;
`

const LinkIcon = styled.div<{ $color: string; $connected: boolean }>`
  width: 32px;
  height: 32px;
  border-radius: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 16px;
  background: ${props => `${props.$color}20`};
  color: ${props => props.$color};
  transition: all 0.3s;

  ${props => props.$connected && css`
    animation: ${pulse} 2s ease-in-out infinite;
  `}
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
  background: ${props => props.$active ? props.$color : 'rgba(255,255,255,0.15)'};
  transition: all 0.3s;
  transform-origin: bottom;

  ${props => props.$active && css`
    animation: ${signalJump} 1.5s ease-in-out infinite;
    animation-delay: ${props.$index * 0.1}s;
  `}
`

const MetricsRow = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
`

const Metric = styled.div`
  text-align: center;
  padding: 4px;
`

const MetricLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
  margin-bottom: 2px;
`

const MetricValue = styled.div<{ $color?: string }>`
  font-size: 12px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: ${props => props.$color || '#fff'};
`

const SwitchButton = styled(Button)<{ $canSwitch: boolean }>`
  ${props => !props.$canSwitch && css`
    opacity: 0.5;
    cursor: not-allowed;
  `}
`

const ActiveBadge = styled(Tag)<{ $active: boolean }>`
  ${props => props.$active && css`
    animation: ${pulse} 1.5s ease-in-out infinite;
  `}
`

interface DualLinkStatusPanelProps {
  linkStatus: CockpitLinkStatus | null
  onFailoverToggle: (enabled: boolean) => void
  onSwitchPrimary: (linkType: LinkType) => void
  onRefresh: () => void
  radioRSSI?: number
  lteRSSI?: number
}

const DualLinkStatusPanel: React.FC<DualLinkStatusPanelProps> = ({
  linkStatus,
  onFailoverToggle,
  onSwitchPrimary,
  onRefresh,
  radioRSSI = 0,
  lteRSSI = 0
}) => {
  if (!linkStatus) {
    return (
      <Container>
        <Header>
          <Title>
            <WifiOutlined style={{ color: '#1890ff' }} />
            双链路备份
          </Title>
        </Header>
        <div style={{ textAlign: 'center', padding: '24px 0', color: 'rgba(255,255,255,0.4)' }}>
          暂无链路数据
        </div>
      </Container>
    )
  }

  const renderSignalBars = (rssi: number, type: LinkType, connected: boolean) => {
    const level = getRSSILevel(rssi, type)
    const color = getRSSIColor(rssi, type)
    const heights = [4, 8, 12, 16]

    return (
      <SignalBars>
        {heights.map((h, i) => (
          <SignalBar
            key={i}
            $active={connected && i < level}
            $color={color}
            $index={i}
            style={{ height: h }}
          />
        ))}
      </SignalBars>
    )
  }

  const getLatencyColor = (latency: number): string => {
    if (latency < 100) return '#52c41a'
    if (latency < 200) return '#faad14'
    return '#ff4d4f'
  }

  const getPacketLossColor = (loss: number): string => {
    if (loss < 1) return '#52c41a'
    if (loss < 5) return '#faad14'
    return '#ff4d4f'
  }

  const getLinkColor = (type: LinkType): string => {
    return type === LinkType.LTE ? '#13c2c2' : '#1890ff'
  }

  const getLinkIcon = (type: LinkType) => {
    return type === LinkType.LTE ? <MobileOutlined /> : <SignalOutlined />
  }

  const renderLinkCard = (type: LinkType) => {
    const isPrimary = linkStatus.primary_link === type
    const isSecondary = linkStatus.secondary_link === type
    const state = isPrimary ? linkStatus.primary_state : linkStatus.secondary_state
    const connected = state === LinkState.CONNECTED || state === LinkState.DEGRADED
    const latency = isPrimary ? linkStatus.primary_latency_ms : linkStatus.secondary_latency_ms
    const packetLoss = isPrimary ? linkStatus.primary_packet_loss : linkStatus.secondary_packet_loss
    const rssi = type === LinkType.LTE ? lteRSSI : radioRSSI
    const color = getLinkColor(type)
    const isActive = isPrimary

    return (
      <LinkCard key={type} $active={isActive} $connected={connected}>
        <LinkHeader>
          <LinkName>
            <LinkIcon $color={color} $connected={connected}>
              {getLinkIcon(type)}
            </LinkIcon>
            {LinkTypeText[type]}
            {connected ? (
              <CheckCircleOutlined style={{ color: '#52c41a', fontSize: 12 }} />
            ) : (
              <CloseCircleOutlined style={{ color: '#ff4d4f', fontSize: 12 }} />
            )}
            {isPrimary && (
              <ActiveBadge color="blue" $active={isActive}>
                主链路
              </ActiveBadge>
            )}
            {isSecondary && (
              <Tag color="default" style={{ fontSize: 10 }}>
                备用
              </Tag>
            )}
          </LinkName>
          <Space size={4}>
            {renderSignalBars(rssi, type, connected)}
            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)', fontFamily: 'monospace' }}>
              {rssi} dBm
            </span>
          </Space>
        </LinkHeader>

        <MetricsRow>
          <Metric>
            <MetricLabel>
              <ClockCircleOutlined /> 延迟
            </MetricLabel>
            <MetricValue $color={getLatencyColor(latency)}>{latency} ms</MetricValue>
          </Metric>
          <Metric>
            <MetricLabel>
              <WarningOutlined /> 丢包
            </MetricLabel>
            <MetricValue $color={getPacketLossColor(packetLoss)}>{packetLoss.toFixed(2)}%</MetricValue>
          </Metric>
          <Metric>
            <MetricLabel>状态</MetricLabel>
            <MetricValue $color={connected ? '#52c41a' : '#ff4d4f'}>
              {LinkStateText[state]}
            </MetricValue>
          </Metric>
        </MetricsRow>

        {!isPrimary && connected && (
          <div style={{ marginTop: 10, textAlign: 'right' }}>
            <SwitchButton
              size="small"
              type="primary"
              ghost
              icon={<SwapOutlined />}
              onClick={() => onSwitchPrimary(type)}
              $canSwitch={connected}
            >
              切换为主链路
            </SwitchButton>
          </div>
        )}
      </LinkCard>
    )
  }

  return (
    <Container>
      <Header>
        <Title>
          <WifiOutlined style={{ color: '#1890ff' }} />
          双链路备份
          {linkStatus.failover_count > 0 && (
            <Tag color="purple" style={{ fontSize: 10 }}>
              已切换 {linkStatus.failover_count} 次
            </Tag>
          )}
        </Title>
        <Tooltip title="刷新链路状态">
          <Button
            size="small"
            type="text"
            icon={<ReloadOutlined />}
            onClick={onRefresh}
            style={{ color: 'rgba(255,255,255,0.6)' }}
          />
        </Tooltip>
      </Header>

      {renderLinkCard(linkStatus.primary_link)}
      {renderLinkCard(linkStatus.secondary_link)}

      <Divider style={{ margin: '12px 0', borderColor: 'rgba(255,255,255,0.05)' }} />

      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <Space>
          <SafetyCertificateOutlined style={{ color: linkStatus.failover_enabled ? '#52c41a' : 'rgba(255,255,255,0.4)' }} />
          <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.7)' }}>
            自动故障切换
          </span>
          <Switch
            checked={linkStatus.failover_enabled}
            onChange={onFailoverToggle}
            size="small"
          />
        </Space>
        <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>
          阈值: {linkStatus.failover_threshold_ms}ms
        </span>
      </div>

      {linkStatus.last_failover_time && (
        <div style={{ marginTop: 8, fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>
          上次切换: {new Date(linkStatus.last_failover_time).toLocaleString()}
        </div>
      )}
    </Container>
  )
}

export default DualLinkStatusPanel
