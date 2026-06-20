import React, { useEffect } from 'react'
import styled from 'styled-components'
import { List, Badge, Tag, Button, Empty } from 'antd'
import { BellOutlined, CheckCircleOutlined, ExclamationCircleOutlined, WarningOutlined, InfoCircleOutlined } from '@ant-design/icons'
import { useAlert } from '@/hooks/useAlert'
import { getSeverityColor, formatDateTime } from '@/utils'
import type { Alert, AlertSeverity } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
`

const Header = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.1);
`

const Title = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  font-weight: 600;
  font-size: 15px;
`

const AlertList = styled(List)`
  flex: 1;
  overflow-y: auto;
  
  .ant-list-item {
    padding: 12px 16px;
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    cursor: pointer;
    transition: background 0.2s;
    
    &:hover {
      background: rgba(255, 255, 255, 0.05);
    }
  }
`

const AlertItem = styled.div`
  width: 100%;
  display: flex;
  gap: 12px;
`

const AlertIcon = styled.div<{ $severity: AlertSeverity }>`
  font-size: 20px;
  color: ${props => getSeverityColor(props.$severity)};
  flex-shrink: 0;
`

const AlertContent = styled.div`
  flex: 1;
  min-width: 0;
`

const AlertTitle = styled.div`
  font-weight: 500;
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

const AlertMessage = styled.div`
  font-size: 13px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 4px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
`

const AlertMeta = styled.div`
  display: flex;
  gap: 8px;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
`

interface AlertPanelProps {
  maxItems?: number
  showTitle?: boolean
}

const getAlertIcon = (severity: AlertSeverity) => {
  switch (severity) {
    case 'critical':
      return <ExclamationCircleOutlined />
    case 'error':
      return <WarningOutlined />
    case 'warning':
      return <WarningOutlined />
    case 'info':
      return <InfoCircleOutlined />
    default:
      return <BellOutlined />
  }
}

const AlertPanel: React.FC<AlertPanelProps> = ({
  maxItems = 10,
  showTitle = true
}) => {
  const { alerts, unreadCount, loading, acknowledge, loadAlerts } = useAlert()

  useEffect(() => {
    loadAlerts({ pageSize: maxItems })
  }, [loadAlerts, maxItems])

  const handleAlertClick = (alert: Alert) => {
    if (alert.status === 'active') {
      acknowledge(alert.id)
    }
  }

  const displayAlerts = alerts.slice(0, maxItems)

  return (
    <Container>
      {showTitle && (
        <Header>
          <Title>
            <BellOutlined style={{ color: '#faad14' }} />
            告警通知
            <Badge count={unreadCount} size="small" style={{ marginLeft: 8 }} />
          </Title>
          <Button type="link" size="small">
            查看全部
          </Button>
        </Header>
      )}

      {displayAlerts.length === 0 ? (
        <Empty
          description="暂无告警"
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          style={{ margin: 'auto' }}
        />
      ) : (
        <AlertList
          dataSource={displayAlerts}
          loading={loading}
          renderItem={(alert) => (
            <List.Item key={alert.id} onClick={() => handleAlertClick(alert)}>
              <AlertItem>
                <AlertIcon $severity={alert.severity}>
                  {getAlertIcon(alert.severity)}
                </AlertIcon>
                <AlertContent>
                  <AlertTitle>
                    {alert.title}
                    {alert.status === 'active' && (
                      <Badge status="processing" color={getSeverityColor(alert.severity)} style={{ marginLeft: 8 }} />
                    )}
                  </AlertTitle>
                  <AlertMessage>{alert.message}</AlertMessage>
                  <AlertMeta>
                    <Tag color={getSeverityColor(alert.severity)} style={{ margin: 0, padding: '0 4px' }}>
                      {alert.severity.toUpperCase()}
                    </Tag>
                    <span>{alert.category}</span>
                    <span>{formatDateTime(alert.createdAt)}</span>
                    {alert.uavName && <span>无人机: {alert.uavName}</span>}
                  </AlertMeta>
                </AlertContent>
                {alert.status === 'active' && (
                  <Button
                    type="text"
                    size="small"
                    icon={<CheckCircleOutlined />}
                    onClick={(e) => {
                      e.stopPropagation()
                      acknowledge(alert.id)
                    }}
                  />
                )}
              </AlertItem>
            </List.Item>
          )}
        />
      )}
    </Container>
  )
}

export default AlertPanel
