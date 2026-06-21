import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import { Card, Tooltip, Tag, Progress, Badge, Space } from 'antd'
import {
  WifiOutlined,
  SignalOutlined,
  MobileOutlined,
  SwapOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined
} from '@ant-design/icons'
import type { LinkStatus, LinkType } from '@/types/link'
import { LinkTypeText, LinkStateText, getRSSIColor, getRSSILevel } from '@/types/link'
import { getLinkStatus } from '@/api/link'

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

const ActiveLinkBadge = styled(Tag)`
  font-size: 12px;
  font-weight: 500;
`

const LinkRow = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px;
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  margin-bottom: 8px;

  &:last-child {
    margin-bottom: 0;
  }

  &.active {
    background: rgba(24, 144, 255, 0.1);
    border: 1px solid rgba(24, 144, 255, 0.3);
  }
`

const LinkIcon = styled.div<{ color: string }>`
  width: 36px;
  height: 36px;
  border-radius: 8px;
  background: ${props => props.color}20;
  display: flex;
  align-items: center;
  justify-content: center;
  color: ${props => props.color};
  font-size: 18px;
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
`

const SignalBars = styled.div`
  display: flex;
  align-items: flex-end;
  gap: 2px;
  height: 16px;
`

const SignalBar = styled.div<{ active: boolean; color: string }>`
  width: 3px;
  border-radius: 1px;
  background: ${props => props.active ? props.color : 'rgba(255,255,255,0.2)'};
  transition: height 0.3s ease;
`

const LatencyBadge = styled.div<{ value: number }>`
  padding: 2px 8px;
  border-radius: 4px;
  background: ${props => props.value < 100 ? 'rgba(82, 196, 26, 0.2)' : props.value < 200 ? 'rgba(250, 173, 20, 0.2)' : 'rgba(255, 77, 79, 0.2)'};
  color: ${props => props.value < 100 ? '#52c41a' : props.value < 200 ? '#faad14' : '#ff4d4f'};
  font-size: 11px;
  font-weight: 500;
`

const PacketLossBadge = styled.div<{ value: number }>`
  padding: 2px 8px;
  border-radius: 4px;
  background: ${props => props.value < 1 ? 'rgba(82, 196, 26, 0.2)' : props.value < 5 ? 'rgba(250, 173, 20, 0.2)' : 'rgba(255, 77, 79, 0.2)'};
  color: ${props => props.value < 1 ? '#52c41a' : props.value < 5 ? '#faad14' : '#ff4d4f'};
  font-size: 11px;
  font-weight: 500;
`

const RSSIProgress = styled(Progress)`
  width: 100px;

  .ant-progress-text {
    color: rgba(255,255,255,0.6) !important;
    font-size: 11px;
  }
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

const LinkStatusIndicator: React.FC<{
  uavId: string
  refreshInterval?: number
  showDetails?: boolean
}> = ({ uavId, refreshInterval = 3000, showDetails = true }) => {
  const [linkStatus, setLinkStatus] = useState<LinkStatus | null>(null)
  const [loading, setLoading] = useState(false)

  const loadStatus = async () => {
    if (!uavId) return
    try {
      setLoading(true)
      const status = await getLinkStatus(uavId)
      setLinkStatus(status)
    } catch (error) {
      // 静默失败，保持上一次的状态
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    loadStatus()
    const interval = setInterval(loadStatus, refreshInterval)
    return () => clearInterval(interval)
  }, [uavId, refreshInterval])

  const renderSignalBars = (rssi: number, type: LinkType) => {
    const level = getRSSILevel(rssi, type)
    const color = getRSSIColor(rssi, type)
    const bars = [4, 8, 12, 16]

    return (
      <SignalBars>
        {bars.map((height, index) => (
          <SignalBar
            key={index}
            active={index < level}
            color={color}
            style={{ height }}
          />
        ))}
      </SignalBars>
    )
  }

  const renderLinkRow = (type: LinkType, status: LinkStatus) => {
    const isRadio = type === LinkType.RADIO
    const connected = isRadio ? status.radio_connected : status.lte_connected
    const rssi = isRadio ? status.radio_rssi : status.lte_rssi
    const state = isRadio ? status.radio_state : status.lte_state
    const isActive = status.active_link === type
    const color = getRSSIColor(rssi, type)
    const rssiPercent = Math.max(0, Math.min(100, ((rssi + 120) / 60) * 100))

    return (
      <LinkRow key={type} className={isActive ? 'active' : ''}>
        <LinkIcon color={color}>
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
              {renderSignalBars(rssi, type)}
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
            {!isRadio && status.lte_network_type && (
              <Tag color="purple" style={{ fontSize: 10, padding: '0 6px' }}>
                {status.lte_network_type}
              </Tag>
            )}
          </LinkDetails>
          {showDetails && (
            <div style={{ display: 'flex', gap: 8, marginTop: 8 }}>
              <LatencyBadge value={status.latency_ms}>
                <ClockCircleOutlined style={{ marginRight: 2 }} />
                {status.latency_ms} ms
              </LatencyBadge>
              <PacketLossBadge value={status.packet_loss}>
                丢包 {status.packet_loss.toFixed(1)}%
              </PacketLossBadge>
            </div>
          )}
        </LinkInfo>
      </LinkRow>
    )
  }

  if (!linkStatus) {
    return (
      <Container loading={loading}>
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
    <Container loading={loading}>
      <Header>
        <Title>
          <WifiOutlined style={{ color: '#1890ff' }} />
          通信链路
        </Title>
        <Space>
          <ActiveLinkBadge color={linkStatus.active_link === 1 ? 'blue' : 'cyan'}>
            <SwapOutlined style={{ marginRight: 4 }} />
            {LinkTypeText[linkStatus.active_link]}
          </ActiveLinkBadge>
          <Badge
            status={linkStatus.radio_connected || linkStatus.lte_connected ? 'processing' : 'error'}
            text={linkStatus.radio_connected || linkStatus.lte_connected ? '在线' : '离线'}
          />
        </Space>
      </Header>

      {renderLinkRow(1 as LinkType, linkStatus)}
      {renderLinkRow(2 as LinkType, linkStatus)}

      <AutoSwitchIndicator>
        {linkStatus.auto_switch_enabled ? (
          <>
            <CheckCircleOutlined style={{ color: '#52c41a' }} />
            <span>自动切换已启用</span>
          </>
        ) : (
          <>
            <CloseCircleOutlined style={{ color: 'rgba(255,255,255,0.4)' }} />
            <span style={{ color: 'rgba(255,255,255,0.4)' }}>自动切换已禁用</span>
          </>
        )}
        <span style={{ marginLeft: 'auto', color: 'rgba(255,255,255,0.4)' }}>
          上行 {linkStatus.bytes_sent || '0'} | 下行 {linkStatus.bytes_received || '0'}
        </span>
      </AutoSwitchIndicator>
    </Container>
  )
}

export default LinkStatusIndicator
