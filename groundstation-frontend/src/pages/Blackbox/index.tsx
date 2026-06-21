import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Table,
  Button,
  Space,
  Input,
  Select,
  DatePicker,
  Card,
  Row,
  Col,
  Statistic,
  Modal,
  message,
  Tag,
  Tooltip,
  Descriptions,
  Empty,
  Upload,
  Progress,
  Badge
} from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  ExportOutlined,
  EyeOutlined,
  DownloadOutlined,
  PlayCircleOutlined,
  ClockCircleOutlined,
  RiseOutlined,
  DashboardOutlined,
  EnvironmentOutlined,
  UploadOutlined,
  FileTextOutlined,
  AlertOutlined,
  ThunderboltOutlined,
  BarChartOutlined,
  DeleteOutlined
} from '@ant-design/icons'
import type { UploadFile, UploadProps } from 'antd/es/upload/interface'
import type { BlackboxLog, BlackboxLogStatus } from '@/types/blackbox'
import {
  getBlackboxList,
  getBlackboxDetail,
  uploadBlackbox,
  deleteBlackbox,
  exportBlackboxCSV,
  exportBlackboxReport,
  analyzeBlackbox,
  getAnalysisReport,
  getBlackboxStatistics
} from '@/api/blackbox'
import { useUAV } from '@/hooks/useUAV'
import { formatDateTime, formatDuration, formatDistance, formatFileSize } from '@/utils'
import dayjs from 'dayjs'
import { useNavigate } from 'react-router-dom'

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
  display: flex;
  align-items: center;
  gap: 10px;
  color: #fff;
`

const SearchBar = styled.div`
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
`

const StatsRow = styled(Row)`
  margin-bottom: 8px;
`

const StatCard = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-statistic-title {
    color: rgba(255, 255, 255, 0.6);
  }

  .ant-statistic-content {
    color: #fff;
  }
`

const TableContainer = styled(Card)`
  flex: 1;
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .ant-card-body {
    flex: 1;
    overflow: hidden;
    display: flex;
    flex-direction: column;
    padding: 0;
  }
`

const StatusTag = styled(Tag)<{ status: BlackboxLogStatus }>`
  ${props => {
    switch (props.status) {
      case 'uploading':
        return 'background: rgba(24, 144, 255, 0.1); color: #1890ff; border-color: rgba(24, 144, 255, 0.3);'
      case 'uploaded':
        return 'background: rgba(82, 196, 26, 0.1); color: #52c41a; border-color: rgba(82, 196, 26, 0.3);'
      case 'analyzed':
        return 'background: rgba(114, 46, 209, 0.1); color: #722ed1; border-color: rgba(114, 46, 209, 0.3);'
      case 'error':
        return 'background: rgba(255, 77, 79, 0.1); color: #ff4d4f; border-color: rgba(255, 77, 79, 0.3);'
      default:
        return ''
    }
  }}
`

const CrashBadge = styled(Badge)`
  .ant-badge-status-dot {
    width: 8px;
    height: 8px;
  }
`

const BlackboxPage: React.FC = () => {
  const navigate = useNavigate()
  const { uavList } = useUAV()
  const [logs, setLogs] = useState<BlackboxLog[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [keyword, setKeyword] = useState('')
  const [uavFilter, setUavFilter] = useState<string>('')
  const [statusFilter, setStatusFilter] = useState<string>('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedLog, setSelectedLog] = useState<BlackboxLog | null>(null)
  const [uploadVisible, setUploadVisible] = useState(false)
  const [uploadUavId, setUploadUavId] = useState<string>('')
  const [fileList, setFileList] = useState<UploadFile[]>([])
  const [stats, setStats] = useState<Record<string, unknown>>({})

  useEffect(() => {
    loadLogs()
    loadStatistics()
  }, [currentPage, pageSize, uavFilter, statusFilter, dateRange])

  const loadLogs = async () => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = {
        page: currentPage,
        page_size: pageSize
      }
      if (uavFilter) params.uav_id = uavFilter
      if (statusFilter) params.status = statusFilter
      if (dateRange) {
        params.start_time = dateRange[0].toISOString()
        params.end_time = dateRange[1].toISOString()
      }
      const result = await getBlackboxList(params)
      setLogs(result.list)
      setTotal(result.total)
    } catch (error) {
      message.error('加载日志列表失败')
    } finally {
      setLoading(false)
    }
  }

  const loadStatistics = async () => {
    try {
      const params: Record<string, unknown> = {}
      if (uavFilter) params.uav_id = uavFilter
      if (dateRange) {
        params.start_time = dateRange[0].toISOString()
        params.end_time = dateRange[1].toISOString()
      }
      const result = await getBlackboxStatistics(params)
      setStats(result)
    } catch {
      // 忽略统计加载错误
    }
  }

  const handleViewDetail = async (log: BlackboxLog) => {
    try {
      const detail = await getBlackboxDetail(log.id)
      setSelectedLog(detail)
      setDetailVisible(true)
    } catch (error) {
      message.error('加载日志详情失败')
    }
  }

  const handlePlayback = (log: BlackboxLog) => {
    navigate(`/blackbox/${log.id}`)
  }

  const handleDelete = async (log: BlackboxLog) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除日志 "${log.flight_name}" 吗？此操作不可恢复。`,
      okText: '删除',
      cancelText: '取消',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await deleteBlackbox(log.id)
          message.success('删除成功')
          loadLogs()
        } catch (error) {
          message.error('删除失败')
        }
      }
    })
  }

  const handleExportCSV = (log: BlackboxLog) => {
    const url = exportBlackboxCSV(log.id)
    window.open(url, '_blank')
  }

  const handleExportReport = async (log: BlackboxLog) => {
    try {
      const url = exportBlackboxReport(log.id)
      window.open(url, '_blank')
    } catch (error) {
      message.error('导出报告失败')
    }
  }

  const handleAnalyze = async (log: BlackboxLog) => {
    try {
      await analyzeBlackbox(log.id)
      message.success('分析任务已启动')
      setTimeout(() => loadLogs(), 2000)
    } catch (error) {
      message.error('启动分析失败')
    }
  }

  const handleUpload = async () => {
    if (!uploadUavId) {
      message.error('请选择无人机')
      return
    }
    if (fileList.length === 0) {
      message.error('请选择文件')
      return
    }

    const formData = new FormData()
    formData.append('uav_id', uploadUavId)
    formData.append('file', fileList[0] as unknown as File)

    try {
      await uploadBlackbox(formData)
      message.success('上传成功')
      setUploadVisible(false)
      setFileList([])
      setUploadUavId('')
      loadLogs()
    } catch (error) {
      message.error('上传失败')
    }
  }

  const uploadProps: UploadProps = {
    fileList,
    beforeUpload: (file) => {
      setFileList([file])
      return false
    },
    onRemove: () => {
      setFileList([])
    },
    maxCount: 1,
    accept: '.bin,.ulg,.log,.csv'
  }

  const getStatusText = (status: BlackboxLogStatus) => {
    const map: Record<BlackboxLogStatus, string> = {
      uploading: '上传中',
      uploaded: '已上传',
      analyzed: '已分析',
      error: '错误'
    }
    return map[status] || status
  }

  const columns = [
    {
      title: '飞行名称',
      dataIndex: 'flight_name',
      key: 'flight_name',
      width: 200,
      render: (name: string, record: BlackboxLog) => (
        <Space size={8}>
          <FileTextOutlined style={{ color: '#1890ff' }} />
          <span style={{ color: '#fff' }}>{name}</span>
          {record.crash_detected && (
            <CrashBadge status="error" text={<Tag color="red" icon={<AlertOutlined />}>坠毁</Tag>} />
          )}
        </Space>
      )
    },
    {
      title: '无人机',
      dataIndex: ['uav', 'name'],
      key: 'uav_name',
      width: 120,
      render: (name: string) => name || '-'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: BlackboxLogStatus) => (
        <StatusTag status={status}>{getStatusText(status)}</StatusTag>
      )
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      width: 160,
      render: (time: string) => time ? formatDateTime(new Date(time).getTime()) : '-'
    },
    {
      title: '飞行时长',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (duration: number) => duration > 0 ? (
        <Tag color="blue" icon={<ClockCircleOutlined />}>
          {formatDuration(duration)}
        </Tag>
      ) : '-'
    },
    {
      title: '飞行距离',
      dataIndex: 'distance',
      key: 'distance',
      width: 100,
      render: (distance: number) => distance > 0 ? formatDistance(distance) : '-'
    },
    {
      title: '最大高度',
      dataIndex: 'max_altitude',
      key: 'max_altitude',
      width: 100,
      render: (alt: number) => alt > 0 ? `${alt.toFixed(1)} m` : '-'
    },
    {
      title: '文件大小',
      dataIndex: 'file_size',
      key: 'file_size',
      width: 100,
      render: (size: number) => formatFileSize(size)
    },
    {
      title: '上传时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 160,
      render: (time: string) => formatDateTime(new Date(time).getTime())
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: BlackboxLog) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="飞行回放">
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
              onClick={() => handlePlayback(record)}
            />
          </Tooltip>
          <Tooltip title="分析报告">
            <Button
              type="link"
              size="small"
              icon={<BarChartOutlined />}
              onClick={() => handleExportReport(record)}
            />
          </Tooltip>
          <Tooltip title="导出CSV">
            <Button
              type="link"
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => handleExportCSV(record)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record)}
            />
          </Tooltip>
        </Space>
      )
    }
  ]

  const totalFlights = (stats.total_flights as number) || 0
  const totalDuration = (stats.total_duration as number) || 0
  const totalDistance = (stats.total_distance as number) || 0
  const crashCount = (stats.crash_count as number) || 0

  return (
    <Container>
      <Header>
        <Title>
          <ThunderboltOutlined style={{ color: '#1890ff' }} />
          飞行日志黑匣子
        </Title>

        <Space>
          <Button
            type="primary"
            icon={<UploadOutlined />}
            onClick={() => setUploadVisible(true)}
          >
            上传日志
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadLogs}
          >
            刷新
          </Button>
        </Space>
      </Header>

      <SearchBar>
        <Select
          placeholder="选择无人机"
          value={uavFilter || undefined}
          onChange={setUavFilter}
          style={{ width: 160 }}
          allowClear
        >
          {uavList.map((uav: any) => (
            <Select.Option key={uav.id} value={uav.id}>
              {uav.name}
            </Select.Option>
          ))}
        </Select>
        <Select
          placeholder="状态筛选"
          value={statusFilter || undefined}
          onChange={setStatusFilter}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="uploading">上传中</Select.Option>
          <Select.Option value="uploaded">已上传</Select.Option>
          <Select.Option value="analyzed">已分析</Select.Option>
          <Select.Option value="error">错误</Select.Option>
        </Select>
        <DatePicker.RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs])}
          style={{ width: 280 }}
        />
      </SearchBar>

      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总飞行架次"
              value={totalFlights}
              prefix={<DashboardOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总飞行时长"
              value={formatDuration(totalDuration)}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总飞行距离"
              value={formatDistance(totalDistance)}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="异常/坠毁次数"
              value={crashCount}
              prefix={<AlertOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={logs}
          rowKey="id"
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条记录`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            }
          }}
          scroll={{ y: 'calc(100vh - 500px)' }}
          locale={{
            emptyText: (
              <Empty
                description="暂无日志记录"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )
          }}
        />
      </TableContainer>

      <Modal
        title="日志详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          <Button key="analyze" icon={<BarChartOutlined />} onClick={() => selectedLog && handleAnalyze(selectedLog)}>
            重新分析
          </Button>,
          <Button key="replay" icon={<PlayCircleOutlined />} type="primary" onClick={() => selectedLog && handlePlayback(selectedLog)}>
            飞行回放
          </Button>
        ]}
        width={800}
      >
        {selectedLog && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="飞行名称" span={2}>
              {selectedLog.flight_name}
            </Descriptions.Item>
            <Descriptions.Item label="无人机">
              {selectedLog.uav?.name || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <StatusTag status={selectedLog.status}>
                {getStatusText(selectedLog.status)}
              </StatusTag>
            </Descriptions.Item>
            <Descriptions.Item label="开始时间">
              {selectedLog.start_time ? formatDateTime(new Date(selectedLog.start_time).getTime()) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="结束时间">
              {selectedLog.end_time ? formatDateTime(new Date(selectedLog.end_time).getTime()) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="飞行时长">
              {selectedLog.duration > 0 ? formatDuration(selectedLog.duration) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="飞行距离">
              {selectedLog.distance > 0 ? formatDistance(selectedLog.distance) : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最大高度">
              {selectedLog.max_altitude > 0 ? `${selectedLog.max_altitude.toFixed(1)} m` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="最大速度">
              {selectedLog.max_speed > 0 ? `${selectedLog.max_speed.toFixed(1)} m/s` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="电池使用">
              {selectedLog.battery_used > 0 ? `${selectedLog.battery_used.toFixed(1)}%` : '-'}
            </Descriptions.Item>
            <Descriptions.Item label="文件大小">
              {formatFileSize(selectedLog.file_size)}
            </Descriptions.Item>
            <Descriptions.Item label="文件名">
              {selectedLog.file_name}
            </Descriptions.Item>
            <Descriptions.Item label="坠毁检测">
              {selectedLog.crash_detected ? (
                <Tag color="red" icon={<AlertOutlined />}>是</Tag>
              ) : (
                <Tag color="green">否</Tag>
              )}
            </Descriptions.Item>
            <Descriptions.Item label="备注" span={2}>
              {selectedLog.notes || '-'}
            </Descriptions.Item>
            {selectedLog.tags && (
              <Descriptions.Item label="标签" span={2}>
                {selectedLog.tags}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </Modal>

      <Modal
        title="上传飞行日志"
        open={uploadVisible}
        onCancel={() => {
          setUploadVisible(false)
          setFileList([])
          setUploadUavId('')
        }}
        onOk={handleUpload}
        okText="上传"
        cancelText="取消"
        width={500}
      >
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <div>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>选择无人机 *</div>
            <Select
              placeholder="请选择无人机"
              value={uploadUavId || undefined}
              onChange={setUploadUavId}
              style={{ width: '100%' }}
            >
              {uavList.map((uav: any) => (
                <Select.Option key={uav.id} value={uav.id}>
                  {uav.name}
                </Select.Option>
              ))}
            </Select>
          </div>
          <div>
            <div style={{ marginBottom: 8, fontWeight: 500 }}>日志文件 *</div>
            <Upload {...uploadProps}>
              <Button icon={<UploadOutlined />}>选择文件</Button>
            </Upload>
            <div style={{ marginTop: 8, fontSize: 12, color: 'rgba(0,0,0,0.45)' }}>
              支持 .bin, .ulg, .log, .csv 格式
            </div>
          </div>
        </Space>
      </Modal>
    </Container>
  )
}

export default BlackboxPage
