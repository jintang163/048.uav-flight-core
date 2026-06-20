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
  Empty
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
  EnvironmentOutlined
} from '@ant-design/icons'
import type { FlightRecord } from '@/types'
import { formatDateTime, formatDuration, formatDistance } from '@/utils'
import { getFlightHistory as getFlightList, getFlightDetail, exportFlightLog as exportFlightTrajectory } from '@/api/flight'
import { useUAV } from '@/hooks/useUAV'
import dayjs from 'dayjs'

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
`

const SearchBar = styled.div`
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
`

const StatsRow = styled(Row)`
  margin-bottom: 16px;
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

const DetailModal = styled(Modal)`
  .ant-modal-body {
    max-height: 70vh;
    overflow-y: auto;
  }
`

const MapPreview = styled.div`
  width: 100%;
  height: 300px;
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.1) 0%, rgba(82, 196, 26, 0.1) 100%);
  border-radius: 8px;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed rgba(255, 255, 255, 0.2);
`

const FlightHistory: React.FC = () => {
  const { uavList } = useUAV()
  const [flights, setFlights] = useState<FlightRecord[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [keyword, setKeyword] = useState('')
  const [uavFilter, setUavFilter] = useState<string>('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedFlight, setSelectedFlight] = useState<FlightRecord | null>(null)

  useEffect(() => {
    loadFlights()
  }, [currentPage, pageSize, keyword, uavFilter, dateRange])

  const loadFlights = async () => {
    setLoading(true)
    try {
      const params: any = {
        page: currentPage,
        pageSize
      }
      if (keyword) params.keyword = keyword
      if (uavFilter) params.uavId = uavFilter
      if (dateRange) {
        params.startTime = dateRange[0].valueOf()
        params.endTime = dateRange[1].valueOf()
      }
      const result = await getFlightList(params)
      setFlights(result.list)
      setTotal(result.total)
    } catch (error) {
      message.error('加载飞行记录失败')
    } finally {
      setLoading(false)
    }
  }

  const handleViewDetail = async (flight: FlightRecord) => {
    try {
      const detail = await getFlightDetail(flight.id)
      setSelectedFlight(detail)
      setDetailVisible(true)
    } catch (error) {
      message.error('加载详情失败')
    }
  }

  const handleExport = async (flight: FlightRecord) => {
    try {
      await exportFlightTrajectory(flight.id)
      message.success('轨迹导出成功')
    } catch (error) {
      message.error('导出失败')
    }
  }

  const handleExportAll = () => {
    message.info('正在导出所有飞行记录...')
  }

  const stats = {
    totalFlights: total,
    totalDuration: flights.reduce((sum, f) => sum + f.duration, 0),
    totalDistance: flights.reduce((sum, f) => sum + f.distance, 0),
    maxAltitude: Math.max(...flights.map(f => f.maxAltitude), 0)
  }

  const columns = [
    {
      title: '飞行编号',
      dataIndex: 'id',
      key: 'id',
      width: 120,
      render: (id: string) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
          {id.slice(0, 8)}
        </span>
      )
    },
    {
      title: '无人机',
      dataIndex: 'uavName',
      key: 'uavName',
      width: 120
    },
    {
      title: '起飞时间',
      dataIndex: 'startTime',
      key: 'startTime',
      width: 160,
      render: (time: number) => formatDateTime(time)
    },
    {
      title: '降落时间',
      dataIndex: 'endTime',
      key: 'endTime',
      width: 160,
      render: (time: number) => formatDateTime(time)
    },
    {
      title: '飞行时长',
      dataIndex: 'duration',
      key: 'duration',
      width: 100,
      render: (duration: number) => (
        <Tag color="blue" icon={<ClockCircleOutlined />}>
          {formatDuration(duration)}
        </Tag>
      )
    },
    {
      title: '飞行距离',
      dataIndex: 'distance',
      key: 'distance',
      width: 100,
      render: (distance: number) => formatDistance(distance)
    },
    {
      title: '最大高度',
      dataIndex: 'maxAltitude',
      key: 'maxAltitude',
      width: 100,
      render: (alt: number) => `${alt.toFixed(1)} m`
    },
    {
      title: '最大速度',
      dataIndex: 'maxSpeed',
      key: 'maxSpeed',
      width: 100,
      render: (speed: number) => `${speed.toFixed(1)} m/s`
    },
    {
      title: '飞手',
      dataIndex: 'pilot',
      key: 'pilot',
      width: 100,
      render: (pilot?: string) => pilot || '-'
    },
    {
      title: '用途',
      dataIndex: 'purpose',
      key: 'purpose',
      width: 100,
      render: (purpose?: string) => purpose || '-'
    },
    {
      title: '操作',
      key: 'actions',
      width: 140,
      fixed: 'right',
      render: (_: unknown, record: FlightRecord) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="回放">
            <Button
              type="link"
              size="small"
              icon={<PlayCircleOutlined />}
            />
          </Tooltip>
          <Tooltip title="导出轨迹">
            <Button
              type="link"
              size="small"
              icon={<DownloadOutlined />}
              onClick={() => handleExport(record)}
            />
          </Tooltip>
        </Space>
      )
    }
  ]

  return (
    <Container>
      <Header>
        <Title>
          <ClockCircleOutlined style={{ color: '#1890ff' }} />
          历史飞行记录
        </Title>

        <Space>
          <Button
            icon={<ExportOutlined />}
            onClick={handleExportAll}
          >
            导出全部
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadFlights}
          >
            刷新
          </Button>
        </Space>
      </Header>

      <SearchBar>
        <Input
          placeholder="搜索飞行编号/飞手"
          prefix={<SearchOutlined />}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 200 }}
          allowClear
        />
        <Select
          placeholder="选择无人机"
          value={uavFilter}
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
              value={stats.totalFlights}
              prefix={<DashboardOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总飞行时长"
              value={formatDuration(stats.totalDuration)}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总飞行距离"
              value={formatDistance(stats.totalDistance)}
              prefix={<EnvironmentOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="最高飞行高度"
              value={stats.maxAltitude.toFixed(1)}
              suffix="m"
              prefix={<RiseOutlined />}
              valueStyle={{ color: '#fa8c16' }}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={flights}
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
          scroll={{ y: 'calc(100vh - 480px)' }}
          locale={{
            emptyText: (
              <Empty
                description="暂无飞行记录"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )
          }}
        />
      </TableContainer>

      <DetailModal
        title="飞行详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          <Button key="replay" icon={<PlayCircleOutlined />}>
            飞行回放
          </Button>,
          <Button key="export" type="primary" icon={<DownloadOutlined />} onClick={() => selectedFlight && handleExport(selectedFlight)}>
            导出轨迹
          </Button>
        ]}
        width={900}
      >
        {selectedFlight && (
          <>
            <MapPreview>
              <div style={{ textAlign: 'center', color: 'rgba(255,255,255,0.5)' }}>
                <EnvironmentOutlined style={{ fontSize: 48, marginBottom: 8 }} />
                <div>地图轨迹预览</div>
              </div>
            </MapPreview>

            <Descriptions bordered column={2} size="small">
              <Descriptions.Item label="飞行编号" span={2}>
                {selectedFlight.id}
              </Descriptions.Item>
              <Descriptions.Item label="无人机">{selectedFlight.uavName}</Descriptions.Item>
              <Descriptions.Item label="飞手">{selectedFlight.pilot || '-'}</Descriptions.Item>
              <Descriptions.Item label="用途">{selectedFlight.purpose || '-'}</Descriptions.Item>
              <Descriptions.Item label="备注">{selectedFlight.notes || '-'}</Descriptions.Item>
              <Descriptions.Item label="起飞时间" span={2}>
                {formatDateTime(selectedFlight.startTime)}
              </Descriptions.Item>
              <Descriptions.Item label="降落时间" span={2}>
                {formatDateTime(selectedFlight.endTime)}
              </Descriptions.Item>
              <Descriptions.Item label="飞行时长">
                <Tag color="blue">{formatDuration(selectedFlight.duration)}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="飞行距离">
                {formatDistance(selectedFlight.distance)}
              </Descriptions.Item>
              <Descriptions.Item label="最大高度">{selectedFlight.maxAltitude.toFixed(1)} m</Descriptions.Item>
              <Descriptions.Item label="最大速度">{selectedFlight.maxSpeed.toFixed(1)} m/s</Descriptions.Item>
              <Descriptions.Item label="起飞位置" span={2}>
                {selectedFlight.startPosition.lat.toFixed(6)}, {selectedFlight.startPosition.lng.toFixed(6)}
              </Descriptions.Item>
              <Descriptions.Item label="降落位置" span={2}>
                {selectedFlight.endPosition.lat.toFixed(6)}, {selectedFlight.endPosition.lng.toFixed(6)}
              </Descriptions.Item>
            </Descriptions>
          </>
        )}
      </DetailModal>
    </Container>
  )
}

export default FlightHistory
