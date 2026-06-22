import React, { useMemo } from 'react'
import styled from 'styled-components'
import { Tag, Empty } from 'antd'
import type { ObstacleAvoidanceLog } from '@/types/obstacle-avoidance'
import {
  STRATEGY_LABELS,
  SENSOR_TYPE_LABELS,
  DIRECTION_LABELS,
  ACTION_STATUS_LABELS,
  ACTION_STATUS_COLORS
} from '@/types/obstacle-avoidance'
import { formatDateTime } from '@/utils'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
`

const LogList = styled.div`
  flex: 1;
  overflow-y: auto;
  padding-right: 4px;

  &::-webkit-scrollbar {
    width: 4px;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 2px;
  }
`

const LogItem = styled.div<{ $status: string }>`
  padding: 10px 12px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  background: ${props => {
    if (props.$status === 'failed') return 'rgba(255, 77, 79, 0.05)'
    if (props.$status === 'avoiding' || props.$status === 'bypassing') return 'rgba(24, 144, 255, 0.05)'
    return 'transparent'
  }};
  transition: background 0.2s;

  &:hover {
    background: rgba(255, 255, 255, 0.03);
  }
`

const LogHeader = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 4px;
`

const LogTime = styled.span`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  font-family: 'Courier New', monospace;
`

const LogTags = styled.div`
  display: flex;
  gap: 4px;
`

const LogDescription = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  line-height: 1.5;
`

const LogDetails = styled.div`
  display: flex;
  gap: 16px;
  margin-top: 6px;
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
`

const LogDetail = styled.span`
  display: flex;
  align-items: center;
  gap: 4px;
`

const EmptyContainer = styled.div`
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
`

interface ObstacleAvoidanceLogProps {
  logs: ObstacleAvoidanceLog[]
  maxItems?: number
}

const ObstacleAvoidanceLogPanel: React.FC<ObstacleAvoidanceLogProps> = ({
  logs,
  maxItems
}) => {
  const displayLogs = useMemo(() => {
    const items = maxItems ? logs.slice(0, maxItems) : logs
    return items
  }, [logs, maxItems])

  return (
    <Container>
      {displayLogs.length > 0 ? (
        <LogList>
          {displayLogs.map(log => (
            <LogItem key={log.id} $status={log.status}>
              <LogHeader>
                <LogTime>{formatDateTime(log.timestamp)}</LogTime>
                <LogTags>
                  <Tag color={ACTION_STATUS_COLORS[log.status]} style={{ fontSize: '10px', lineHeight: '16px', padding: '0 4px', margin: 0 }}>
                    {ACTION_STATUS_LABELS[log.status]}
                  </Tag>
                  <Tag style={{ fontSize: '10px', lineHeight: '16px', padding: '0 4px', margin: 0, background: 'rgba(255,255,255,0.05)', color: 'rgba(255,255,255,0.7)', border: '1px solid rgba(255,255,255,0.1)' }}>
                    {STRATEGY_LABELS[log.strategy]}
                  </Tag>
                </LogTags>
              </LogHeader>
              <LogDescription>{log.description}</LogDescription>
              <LogDetails>
                <LogDetail>
                  传感器: {SENSOR_TYPE_LABELS[log.sensorType]}
                </LogDetail>
                <LogDetail>
                  方向: {DIRECTION_LABELS[log.direction]}
                </LogDetail>
                <LogDetail>
                  距离: {log.distance.toFixed(1)}m
                </LogDetail>
                {log.duration !== undefined && (
                  <LogDetail>
                    耗时: {log.duration.toFixed(1)}s
                  </LogDetail>
                )}
              </LogDetails>
            </LogItem>
          ))}
        </LogList>
      ) : (
        <EmptyContainer>
          <Empty description="暂无避障日志" image={Empty.PRESENTED_IMAGE_SIMPLE} />
        </EmptyContainer>
      )}
    </Container>
  )
}

export default ObstacleAvoidanceLogPanel
