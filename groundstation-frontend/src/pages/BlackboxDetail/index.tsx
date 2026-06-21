import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Card,
  Button,
  Space,
  Tabs,
  Descriptions,
  Tag,
  Row,
  Col,
  Statistic,
  Progress,
  List,
  message,
  Spin,
  Empty,
  Tooltip
} from 'antd'
import {
  ArrowLeftOutlined,
  DownloadOutlined,
  BarChartOutlined,
  FileTextOutlined,
  AlertOutlined,
  ClockCircleOutlined,
  RiseOutlined,
  DashboardOutlined,
  EnvironmentOutlined,
  ThunderboltOutlined,
  SafetyOutlined
} from '@ant-design/icons'
import { useParams, useNavigate } from 'react-router-dom'
import type { BlackboxLog, ParsedLogData, AnalysisReport } from '@/types/blackbox'
import {
  getBlackboxDetail,
  parseBlackboxLog,
  getAnalysisReport,
  exportBlackboxCSV,
  exportBlackboxReport
} from '@/api/blackbox'
import { formatDuration, formatDistance, formatDateTime, formatFileSize } from '@/utils'
import FlightReplay from '@/components/FlightReplay'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  padding: 16px;
  gap: 16px;
  overflow: hidden;
`

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
`

const Title = styled.div`
  font-size: 18px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 12px;
`

const BackButton = styled(Button)`
  display: flex;
  align-items: center;
  gap: 8px;
`

const Content = styled.div`
  flex: 1;
  display: flex;
  gap: 16px;
  overflow: hidden;
`

const LeftPanel = styled.div`
  flex: 2;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const RightPanel = styled.div`
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
`

const StatsCard = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-body {
    padding: 16px;
  }
`

const StatItem = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);

  &:last-child {
    border-bottom: none;
  }
`

const StatIcon = styled.div<{ color: string }>`
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

const StatInfo = styled.div`
  flex: 1;
`

const StatLabel = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
`

const StatValue = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #fff;
`

const ScoreCard = styled(Card)`
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.1) 0%, rgba(82, 196, 26, 0.1) 100%);
  border: 1px solid rgba(24, 144, 255, 0.2);
  text-align: center;

  .ant-card-body {
    padding: 24px;
  }
`

const ScoreValue = styled.div`
  font-size: 48px;
  font-weight: 700;
  color: #fff;
  line-height: 1;
  margin-bottom: 8px;
`

const ScoreLabel = styled.div`
  font-size: 14px;
  color: rgba(255, 255, 255, 0.7);
  margin-bottom: 16px;
`

const SectionTitle = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 12px;
  display: flex;
  align-items: center;
  gap: 8px;
`

const RecommendationItem = styled.div`
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  margin-bottom: 8px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: flex-start;
  gap: 8px;

  &:last-child {
    margin-bottom: 0;
  }
`

const FlightPhaseItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  margin-bottom: 6px;

  &:last-child {
    margin-bottom: 0;
  }
`

const PhaseName = styled.div`
  font-size: 13px;
  color: #fff;
`

const PhaseDuration = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
`

const BlackboxDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>()
  const navigate = useNavigate()
  const [log, setLog] = useState<BlackboxLog | null>(null)
  const [parsedData, setParsedData] = useState<ParsedLogData | null>(null)
  const [analysisReport, setAnalysisReport] = useState<AnalysisReport | null>(null)
  const [loading, setLoading] = useState(false)
  const [parsing, setParsing] = useState(false)

  useEffect(() => {
    if (id) {
      loadData()
    }
  }, [id])

  const loadData = async () => {
    if (!id) return
    setLoading(true)
    try {
      const [logData, analysis] = await Promise.all([
        getBlackboxDetail(id),
        getAnalysisReport(id).catch(() => null)
      ])
      setLog(logData)
      setAnalysisReport(analysis)
    } catch (error) {
      message.error('加载日志详情失败')
    } finally {
      setLoading(false)
    }
  }

  const handleParseLog = async () => {
    if (!id) return
    setParsing(true)
    try {
      const data = await parseBlackboxLog(id)
      setParsedData(data)
    } catch (error) {
      message.error('解析日志失败')
    } finally {
      setParsing(false)
    }
  }

  const handleExportCSV = () => {
    if (!id) return
    const url = exportBlackboxCSV(id)
    window.open(url, '_blank')
  }

  const handleExportReport = () => {
    if (!id) return
    const url = exportBlackboxReport(id)
    window.open(url, '_blank')
  }

  const getSeverityColor = (severity: number) => {
    switch (severity) {
      case 3: return 'red'
      case 2: return 'orange'
      default: return 'blue'
    }
  }

  const getScoreColor = (score: number) => {
    if (score >= 80) return '#52c41a'
    if (score >= 60) return '#fa8c16'
    return '#ff4d4f'
  }

  const stats = [
    { label: '飞行时长', value: log?.duration ? formatDuration(log.duration) : '-', icon: <ClockCircleOutlined />, color: '#1890ff' },
    { label: '飞行距离', value: log?.distance ? formatDistance(log.distance) : '-', icon: <EnvironmentOutlined />, color: '#722ed1' },
    { label: '最大高度', value: log?.max_altitude ? `${log.max_altitude.toFixed(1)} m` : '-', icon: <RiseOutlined />, color: '#fa8c16' },
    { label: '最大速度', value: log?.max_speed ? `${log.max_speed.toFixed(1)} m/s` : '-', icon: <DashboardOutlined />, color: '#13c2c2' },
    { label: '电池使用', value: log?.battery_used ? `${log.battery_used.toFixed(1)}%` : '-', icon: <ThunderboltOutlined />, color: '#faad14' },
    { label: '文件大小', value: log?.file_size ? formatFileSize(log.file_size) : '-', icon: <FileTextOutlined />, color: '#eb2f96' },
  ]

  if (loading) {
    return (
      <Container>
        <div style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
          <Spin size="large" tip="加载中..." />
        </div>
      </Container>
    )
  }

  return (
    <Container>
      <Header>
        <Space>
          <BackButton icon={<ArrowLeftOutlined />} onClick={() => navigate('/blackbox')}>
            返回列表
          </BackButton>
          <Title>
            <ThunderboltOutlined style={{ color: '#1890ff' }} />
            {log?.flight_name || '日志详情'}
          </Title>
        </Space>
        <Space>
          <Tooltip title="导出CSV">
            <Button icon={<DownloadOutlined />} onClick={handleExportCSV}>
              导出CSV
            </Button>
          </Tooltip>
          <Tooltip title="导出报告">
            <Button icon={<FileTextOutlined />} onClick={handleExportReport}>
              导出报告
            </Button>
          </Tooltip>
        </Space>
      </Header>

      <Content>
        <LeftPanel>
          {parsedData ? (
            <FlightReplay
              dataPoints={parsedData.data_points}
              events={parsedData.events}
              statistics={parsedData.statistics}
              title="飞行回放"
            />
          ) : (
            <StatsCard style={{ flex: 1, display: 'flex', alignItems: 'center', justifyContent: 'center' }}>
              <Empty
                description="点击下方按钮解析日志以查看回放"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
              <Button
                type="primary"
                icon={<BarChartOutlined />}
                loading={parsing}
                onClick={handleParseLog}
                style={{ marginTop: 16 }}
              >
                解析日志
              </Button>
            </StatsCard>
          )}
        </LeftPanel>

        <RightPanel>
          {log && (
            <StatsCard size="small">
              <SectionTitle>
                <FileTextOutlined style={{ color: '#1890ff' }} />
                基本信息
              </SectionTitle>
              <StatItem>
                <StatIcon color="#1890ff">
                  <FileTextOutlined />
                </StatIcon>
                <StatInfo>
                  <StatLabel>飞行名称</StatLabel>
                  <StatValue>{log.flight_name}</StatValue>
                </StatInfo>
              </StatItem>
              <StatItem>
                <StatIcon color="#52c41a">
                  <SafetyOutlined />
                </StatIcon>
                <StatInfo>
                  <StatLabel>状态</StatLabel>
                  <StatValue>
                    <Tag color={log.status === 'analyzed' ? 'purple' : log.status === 'uploaded' ? 'green' : 'blue'}>
                      {log.status === 'uploading' ? '上传中' : log.status === 'uploaded' ? '已上传' : log.status === 'analyzed' ? '已分析' : '错误'}
                    </Tag>
                  </StatValue>
                </StatInfo>
              </StatItem>
              <StatItem>
                <StatIcon color="#fa8c16">
                  <ClockCircleOutlined />
                </StatIcon>
                <StatInfo>
                  <StatLabel>开始时间</StatLabel>
                  <StatValue style={{ fontSize: 13 }}>
                    {log.start_time ? formatDateTime(new Date(log.start_time).getTime()) : '-'}
                  </StatValue>
                </StatInfo>
              </StatItem>
              {log.crash_detected && (
                <StatItem>
                  <StatIcon color="#ff4d4f">
                    <AlertOutlined />
                  </StatIcon>
                  <StatInfo>
                    <StatLabel>坠毁检测</StatLabel>
                    <StatValue style={{ color: '#ff4d4f' }}>检测到坠毁</StatValue>
                  </StatInfo>
                </StatItem>
              )}
            </StatsCard>
          )}

          {analysisReport && (
            <ScoreCard>
              <ScoreValue style={{ color: getScoreColor(analysisReport.flight_score) }}>
                {analysisReport.flight_score}
              </ScoreValue>
              <ScoreLabel>飞行评分</ScoreLabel>
              <Progress
                percent={analysisReport.flight_score}
                strokeColor={getScoreColor(analysisReport.flight_score)}
                showInfo={false}
              />
              <div style={{ marginTop: 12, fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>
                {analysisReport.flight_summary}
              </div>
            </ScoreCard>
          )}

          <StatsCard size="small">
            <SectionTitle>
              <BarChartOutlined style={{ color: '#52c41a' }} />
              飞行统计
            </SectionTitle>
            {stats.map((stat, index) => (
              <StatItem key={index}>
                <StatIcon color={stat.color}>{stat.icon}</StatIcon>
                <StatInfo>
                  <StatLabel>{stat.label}</StatLabel>
                  <StatValue>{stat.value}</StatValue>
                </StatInfo>
              </StatItem>
            ))}
          </StatsCard>

          {analysisReport?.flight_phases && analysisReport.flight_phases.length > 0 && (
            <StatsCard size="small">
              <SectionTitle>
                <ClockCircleOutlined style={{ color: '#722ed1' }} />
                飞行阶段
              </SectionTitle>
              {analysisReport.flight_phases.map((phase, index) => (
                <FlightPhaseItem key={index}>
                  <PhaseName>{phase.phase_name}</PhaseName>
                  <PhaseDuration>{formatDuration(phase.duration)}</PhaseDuration>
                </FlightPhaseItem>
              ))}
            </StatsCard>
          )}

          {analysisReport?.recommendations && analysisReport.recommendations.length > 0 && (
            <StatsCard size="small">
              <SectionTitle>
                <AlertOutlined style={{ color: '#fa8c16' }} />
                改进建议
              </SectionTitle>
              {analysisReport.recommendations.map((rec, index) => (
                <RecommendationItem key={index}>
                  <span style={{ color: '#fa8c16' }}>•</span>
                  <span>{rec}</span>
                </RecommendationItem>
              ))}
            </StatsCard>
          )}

          {analysisReport?.anomalies && analysisReport.anomalies.length > 0 && (
            <StatsCard size="small">
              <SectionTitle>
                <AlertOutlined style={{ color: '#ff4d4f' }} />
                异常事件 ({analysisReport.anomalies.length})
              </SectionTitle>
              <List
                size="small"
                dataSource={analysisReport.anomalies.slice(0, 5)}
                renderItem={(event) => (
                  <List.Item key={event.timestamp}>
                    <List.Item.Meta
                      avatar={
                        <Tag color={getSeverityColor(event.severity)} style={{ marginRight: 8 }}>
                          {event.severity === 3 ? '严重' : event.severity === 2 ? '警告' : '提示'}
                        </Tag>
                      }
                      title={<span style={{ color: '#fff', fontSize: 12 }}>{event.description}</span>}
                      description={
                        <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 11 }}>
                          {formatDuration(event.timestamp / 1000)}
                        </span>
                      }
                    />
                  </List.Item>
                )}
              />
            </StatsCard>
          )}
        </RightPanel>
      </Content>
    </Container>
  )
}

export default BlackboxDetail
