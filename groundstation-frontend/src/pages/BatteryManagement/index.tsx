import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Table,
  Button,
  Space,
  Input,
  Select,
  Tag,
  Card,
  Row,
  Col,
  Statistic,
  Modal,
  Form,
  message,
  Popconfirm,
  Badge,
  Tooltip,
  Descriptions,
  Progress,
  Tabs,
  List,
  Empty
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  BatteryOutlined,
  ThunderboltOutlined,
  ReloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  HistoryOutlined,
  ThunderboltFilled
} from '@ant-design/icons'
import {
  getBatteryList,
  getBatteryDetail,
  createBattery,
  updateBattery,
  deleteBattery,
  getBatteryStatistics,
  getBatteryUsageRecords,
  getBatteryCellData,
  identifyBattery
} from '@/api/battery'
import { formatDateTime, formatDuration } from '@/utils'
import type {
  Battery,
  BatteryStatus,
  BatteryHealthStatus,
  BatteryStatistics as BatteryStatsType,
  BatteryUsageRecord,
  BatteryCellData,
  CreateBatteryRequest,
  UpdateBatteryRequest
} from '@/types'

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
`

const StatsRow = styled(Row)`
  margin-bottom: 0;
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

  .ant-table {
    background: transparent;
  }

  .ant-table-thead > tr > th {
    background: rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.8);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .ant-table-tbody > tr > td {
    border-bottom: 1px solid rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.85);
  }

  .ant-table-tbody > tr:hover > td {
    background: rgba(255, 255, 255, 0.05) !important;
  }

  .ant-pagination {
    padding: 16px;
  }
`

const BatteryIconWrapper = styled.div<{ $level: number; $charging?: boolean }>`
  font-size: 24px;
  color: ${props => {
    if (props.$level <= 15) return '#ff4d4f'
    if (props.$level <= 30) return '#faad14'
    if (props.$charging) return '#52c41a'
    return '#1890ff'
  }};
  display: inline-flex;
  align-items: center;
  position: relative;
`

const SOHBar = styled(Progress)`
  .ant-progress-text {
    color: rgba(255, 255, 255, 0.7);
  }
`

const DetailContainer = styled.div`
  .ant-descriptions-title {
    color: #fff;
    font-size: 16px;
    font-weight: 600;
    margin-bottom: 12px;
  }

  .ant-descriptions-item-label {
    color: rgba(255, 255, 255, 0.5);
  }

  .ant-descriptions-item-content {
    color: rgba(255, 255, 255, 0.9);
  }
`

const CellGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(80px, 1fr));
  gap: 8px;
  margin-top: 12px;
`

const CellCard = styled.div<{ $voltage: number; $min: number; $max: number }>`
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid ${props => {
    const diff = props.$max - props.$min
    if (props.$voltage < 3.5) return 'rgba(255, 77, 79, 0.5)'
    if (diff > 0.1) return 'rgba(250, 173, 20, 0.5)'
    return 'rgba(255, 255, 255, 0.1)'
  }};
  padding: 8px;
  border-radius: 6px;
  text-align: center;
`

const CellIndex = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 4px;
`

const CellVoltage = styled.div<{ $voltage: number }>`
  font-size: 14px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: ${props => {
    if (props.$voltage < 3.5) return '#ff4d4f'
    if (props.$voltage < 3.7) return '#faad14'
    return '#52c41a'
  }};
`

const UsageRecordItem = styled.div`
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  margin-bottom: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
`

const UsageRecordHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
`

const UsageRecordStats = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
`

const StatItem = styled.div`
  text-align: center;
`

const StatLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 2px;
`

const StatValue = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: #1890ff;
  font-family: 'Courier New', monospace;
`

const getStatusColor = (status: BatteryStatus): string => {
  switch (status) {
    case 'charging': return '#52c41a'
    case 'in_use': return '#1890ff'
    case 'idle': return '#faad14'
    case 'fault': return '#ff4d4f'
    case 'discharging': return '#eb2f96'
    case 'storage': return '#722ed1'
    default: return '#8c8c8c'
  }
}

const getStatusText = (status: BatteryStatus): string => {
  switch (status) {
    case 'charging': return '充电中'
    case 'in_use': return '使用中'
    case 'idle': return '空闲'
    case 'fault': return '故障'
    case 'discharging': return '放电中'
    case 'storage': return '存储'
    default: return '未知'
  }
}

const getHealthColor = (status: BatteryHealthStatus): string => {
  switch (status) {
    case 'excellent': return '#52c41a'
    case 'good': return '#1890ff'
    case 'fair': return '#faad14'
    case 'poor': return '#ff7a45'
    case 'critical': return '#ff4d4f'
    default: return '#8c8c8c'
  }
}

const getHealthText = (status: BatteryHealthStatus): string => {
  switch (status) {
    case 'excellent': return '优秀'
    case 'good': return '良好'
    case 'fair': return '一般'
    case 'poor': return '较差'
    case 'critical': return '危险'
    default: return '未知'
  }
}

const BatteryManagement: React.FC = () => {
  const [batteries, setBatteries] = useState<Battery[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statistics, setStatistics] = useState<BatteryStatsType | null>(null)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()
  const [healthFilter, setHealthFilter] = useState<string | undefined>()
  const [identifyID, setIdentifyID] = useState('')

  const [modalVisible, setModalVisible] = useState(false)
  const [modalType, setModalType] = useState<'create' | 'edit' | 'detail'>('create')
  const [selectedBattery, setSelectedBattery] = useState<Battery | null>(null)
  const [form] = Form.useForm()

  const [usageRecords, setUsageRecords] = useState<BatteryUsageRecord[]>([])
  const [cellData, setCellData] = useState<BatteryCellData[]>([])
  const [detailTab, setDetailTab] = useState('info')
  const [recordsTotal, setRecordsTotal] = useState(0)
  const [recordsPage, setRecordsPage] = useState(1)

  const fetchStatistics = async () => {
    try {
      const data = await getBatteryStatistics()
      setStatistics(data)
    } catch (error) {
      console.error('Failed to fetch statistics:', error)
    }
  }

  const fetchBatteries = async () => {
    setLoading(true)
    try {
      const data = await getBatteryList({
        page,
        pageSize,
        keyword: searchKeyword || undefined,
        status: statusFilter,
        health_status: healthFilter
      })
      setBatteries(data.list)
      setTotal(data.total)
    } catch (error) {
      message.error('获取电池列表失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    fetchStatistics()
    fetchBatteries()
  }, [page, pageSize, statusFilter, healthFilter])

  const handleSearch = () => {
    setPage(1)
    fetchBatteries()
  }

  const handleIdentify = async () => {
    if (!identifyID) {
      message.warning('请输入电池ID')
      return
    }
    try {
      const battery = await identifyBattery(identifyID)
      setSelectedBattery(battery)
      setModalType('detail')
      setModalVisible(true)
      fetchBatteryDetail(battery.id)
    } catch (error) {
      message.error('未找到该电池')
    }
  }

  const fetchBatteryDetail = async (id: string) => {
    try {
      const [battery, records, cells] = await Promise.all([
        getBatteryDetail(id),
        getBatteryUsageRecords(id, { page: recordsPage, pageSize: 10 }),
        getBatteryCellData(id)
      ])
      setSelectedBattery(battery)
      setUsageRecords(records.list)
      setRecordsTotal(records.total)
      setCellData(cells)
    } catch (error) {
      console.error('Failed to fetch battery detail:', error)
    }
  }

  const handleCreate = () => {
    setModalType('create')
    setSelectedBattery(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: Battery) => {
    setModalType('edit')
    setSelectedBattery(record)
    form.setFieldsValue({
      model: record.model,
      manufacturer: record.manufacturer,
      capacity: record.capacity,
      capacity_unit: record.capacity_unit,
      voltage: record.voltage,
      cell_count: record.cell_count,
      status: record.status,
      location: record.location,
      notes: record.notes
    })
    setModalVisible(true)
  }

  const handleView = (record: Battery) => {
    setModalType('detail')
    setSelectedBattery(record)
    fetchBatteryDetail(record.id)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteBattery(id)
      message.success('删除成功')
      fetchBatteries()
      fetchStatistics()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      
      if (modalType === 'create') {
        await createBattery(values as CreateBatteryRequest)
        message.success('创建成功')
      } else if (modalType === 'edit' && selectedBattery) {
        await updateBattery(selectedBattery.id, values as UpdateBatteryRequest)
        message.success('更新成功')
      }
      
      setModalVisible(false)
      fetchBatteries()
      fetchStatistics()
    } catch (error) {
      console.error('Form error:', error)
    }
  }

  const columns = [
    {
      title: '电池ID',
      dataIndex: 'battery_id',
      key: 'battery_id',
      width: 140,
      render: (text: string, record: Battery) => (
        <Space>
          <BatteryIconWrapper $level={record.current_level} $charging={record.status === 'charging'}>
            {record.status === 'charging' ? <ThunderboltFilled /> : <BatteryOutlined />}
          </BatteryIconWrapper>
          <span style={{ fontFamily: 'Courier New, monospace' }}>{text}</span>
        </Space>
      )
    },
    {
      title: '型号',
      dataIndex: 'model',
      key: 'model',
      width: 120
    },
    {
      title: '电量',
      dataIndex: 'current_level',
      key: 'current_level',
      width: 140,
      render: (level: number, record: Battery) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
          <Progress
            percent={level}
            size="small"
            strokeColor={level <= 15 ? '#ff4d4f' : level <= 30 ? '#faad14' : '#52c41a'}
            style={{ width: 80 }}
          />
          <span style={{ 
            color: level <= 15 ? '#ff4d4f' : level <= 30 ? '#faad14' : '#52c41a',
            fontWeight: 600,
            fontFamily: 'Courier New, monospace'
          }}>
            {level.toFixed(1)}%
          </span>
        </div>
      )
    },
    {
      title: '电压',
      dataIndex: 'current_voltage',
      key: 'current_voltage',
      width: 100,
      render: (v: number) => <span style={{ fontFamily: 'Courier New, monospace' }}>{v.toFixed(2)}V</span>
    },
    {
      title: '温度',
      dataIndex: 'current_temperature',
      key: 'current_temperature',
      width: 100,
      render: (t: number) => (
        <span style={{ 
          color: t > 50 ? '#ff4d4f' : t > 40 ? '#faad14' : '#52c41a',
          fontFamily: 'Courier New, monospace'
        }}>
          {t.toFixed(1)}°C
        </span>
      )
    },
    {
      title: '健康度',
      dataIndex: 'soh',
      key: 'soh',
      width: 140,
      render: (soh: number, record: Battery) => (
        <Space>
          <Progress
            type="circle"
            size={40}
            percent={soh}
            strokeColor={getHealthColor(record.health_status)}
          />
          <div>
            <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>
              {getHealthText(record.health_status)}
            </div>
            <div style={{ fontSize: 14, fontWeight: 600, fontFamily: 'Courier New, monospace' }}>
              {soh.toFixed(1)}%
            </div>
          </div>
        </Space>
      )
    },
    {
      title: '循环次数',
      dataIndex: 'cycle_count',
      key: 'cycle_count',
      width: 100,
      render: (c: number) => <span style={{ fontFamily: 'Courier New, monospace' }}>{c} 次</span>
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: BatteryStatus) => (
        <Tag color={getStatusColor(status)}>{getStatusText(status)}</Tag>
      )
    },
    {
      title: '存放天数',
      dataIndex: 'storage_days',
      key: 'storage_days',
      width: 100,
      render: (days: number, record: Battery) => (
        <Space>
          <span style={{ fontFamily: 'Courier New, monospace' }}>{days} 天</span>
          {record.needs_maintenance && (
            <Tooltip title={record.maintenance_message || '需要保养'}>
              <WarningOutlined style={{ color: '#faad14' }} />
            </Tooltip>
          )}
        </Space>
      )
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      width: 120
    },
    {
      title: '操作',
      key: 'action',
      width: 160,
      fixed: 'right' as const,
      render: (_: unknown, record: Battery) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleView(record)}>
            详情
          </Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该电池？"
            onConfirm={() => handleDelete(record.id)}
            okText="确定"
            cancelText="取消"
          >
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>
              删除
            </Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  const renderDetailContent = () => {
    if (!selectedBattery) return null

    const minVoltage = cellData.length > 0 ? Math.min(...cellData.map(c => c.voltage)) : 0
    const maxVoltage = cellData.length > 0 ? Math.max(...cellData.map(c => c.voltage)) : 0

    return (
      <DetailContainer>
        <Tabs
          activeKey={detailTab}
          onChange={setDetailTab}
          items={[
            {
              key: 'info',
              label: '基本信息',
              children: (
                <Descriptions column={2} bordered size="small">
                  <Descriptions.Item label="电池ID">{selectedBattery.battery_id}</Descriptions.Item>
                  <Descriptions.Item label="型号">{selectedBattery.model || '-'}</Descriptions.Item>
                  <Descriptions.Item label="制造商">{selectedBattery.manufacturer || '-'}</Descriptions.Item>
                  <Descriptions.Item label="容量">
                    {selectedBattery.capacity} {selectedBattery.capacity_unit}
                  </Descriptions.Item>
                  <Descriptions.Item label="标称电压">{selectedBattery.voltage}V</Descriptions.Item>
                  <Descriptions.Item label="电芯数量">{selectedBattery.cell_count} 串</Descriptions.Item>
                  <Descriptions.Item label="当前电压">{selectedBattery.current_voltage.toFixed(2)}V</Descriptions.Item>
                  <Descriptions.Item label="当前电量">{selectedBattery.current_level.toFixed(1)}%</Descriptions.Item>
                  <Descriptions.Item label="温度">{selectedBattery.current_temperature.toFixed(1)}°C</Descriptions.Item>
                  <Descriptions.Item label="电流">{selectedBattery.current_current.toFixed(2)}A</Descriptions.Item>
                  <Descriptions.Item label="健康度">
                    <Tag color={getHealthColor(selectedBattery.health_status)}>
                      {getHealthText(selectedBattery.health_status)} ({selectedBattery.soh.toFixed(1)}%)
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={getStatusColor(selectedBattery.status)}>
                      {getStatusText(selectedBattery.status)}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="循环次数">{selectedBattery.cycle_count} 次</Descriptions.Item>
                  <Descriptions.Item label="总飞行时间">{formatDuration(selectedBattery.total_flight_time)}</Descriptions.Item>
                  <Descriptions.Item label="总充电次数">{selectedBattery.total_charge_count} 次</Descriptions.Item>
                  <Descriptions.Item label="存放天数">{selectedBattery.storage_days} 天</Descriptions.Item>
                  <Descriptions.Item label="位置">{selectedBattery.location || '-'}</Descriptions.Item>
                  <Descriptions.Item label="最后使用">
                    {selectedBattery.last_used_at ? formatDateTime(selectedBattery.last_used_at) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="最后充电">
                    {selectedBattery.last_charged_at ? formatDateTime(selectedBattery.last_charged_at) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="首次使用">
                    {selectedBattery.first_use_date ? formatDateTime(selectedBattery.first_use_date) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="生产日期">
                    {selectedBattery.manufacture_date ? formatDateTime(selectedBattery.manufacture_date) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="备注" span={2}>
                    {selectedBattery.notes || '-'}
                  </Descriptions.Item>
                </Descriptions>
              )
            },
            {
              key: 'cells',
              label: '电芯数据',
              children: cellData.length > 0 ? (
                <div>
                  <div style={{ marginBottom: 12, color: 'rgba(255,255,255,0.7)' }}>
                    最高电压: <span style={{ color: '#faad14', fontWeight: 600 }}>{maxVoltage.toFixed(3)}V</span>
                    {' / '}
                    最低电压: <span style={{ color: '#1890ff', fontWeight: 600 }}>{minVoltage.toFixed(3)}V</span>
                    {' / '}
                    压差: <span style={{ 
                      color: (maxVoltage - minVoltage) > 0.1 ? '#ff4d4f' : '#52c41a',
                      fontWeight: 600 
                    }}>
                      {(maxVoltage - minVoltage).toFixed(3)}V
                    </span>
                  </div>
                  <CellGrid>
                    {cellData.map(cell => (
                      <CellCard key={cell.id} $voltage={cell.voltage} $min={minVoltage} $max={maxVoltage}>
                        <CellIndex>Cell {cell.cell_index + 1}</CellIndex>
                        <CellVoltage $voltage={cell.voltage}>{cell.voltage.toFixed(3)}</CellVoltage>
                        <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.4)', marginTop: 2 }}>
                          {cell.resistance.toFixed(2)}mΩ
                        </div>
                      </CellCard>
                    ))}
                  </CellGrid>
                </div>
              ) : (
                <Empty description="暂无电芯数据" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )
            },
            {
              key: 'usage',
              label: '使用记录',
              children: usageRecords.length > 0 ? (
                <div>
                  {usageRecords.map(record => (
                    <UsageRecordItem key={record.id}>
                      <UsageRecordHeader>
                        <Space>
                          <ClockCircleOutlined style={{ color: '#1890ff' }} />
                          <span style={{ color: 'rgba(255,255,255,0.9)', fontWeight: 500 }}>
                            {record.start_time ? formatDateTime(record.start_time) : '未知时间'}
                          </span>
                        </Space>
                        <Tag color="blue">飞行 {formatDuration(record.duration)}</Tag>
                      </UsageRecordHeader>
                      <UsageRecordStats>
                        <StatItem>
                          <StatLabel>起始电量</StatLabel>
                          <StatValue>{record.start_level.toFixed(1)}%</StatValue>
                        </StatItem>
                        <StatItem>
                          <StatLabel>结束电量</StatLabel>
                          <StatValue>{record.end_level.toFixed(1)}%</StatValue>
                        </StatItem>
                        <StatItem>
                          <StatLabel>最高温度</StatLabel>
                          <StatValue>{record.max_temperature.toFixed(1)}°C</StatValue>
                        </StatItem>
                        <StatItem>
                          <StatLabel>飞行距离</StatLabel>
                          <StatValue>{record.distance.toFixed(1)}m</StatValue>
                        </StatItem>
                      </UsageRecordStats>
                    </UsageRecordItem>
                  ))}
                </div>
              ) : (
                <Empty description="暂无使用记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )
            }
          ]}
        />
      </DetailContainer>
    )
  }

  return (
    <Container>
      <Header>
        <Title>
          <BatteryOutlined style={{ color: '#1890ff' }} />
          智能电池管理
        </Title>
        <Space>
          <Input
            placeholder="电池ID识别"
            value={identifyID}
            onChange={e => setIdentifyID(e.target.value)}
            onPressEnter={handleIdentify}
            style={{ width: 180 }}
            prefix={<SearchOutlined />}
          />
          <Button type="primary" icon={<BatteryOutlined />} onClick={handleIdentify}>
            识别电池
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => { fetchBatteries(); fetchStatistics(); }}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新增电池
          </Button>
        </Space>
      </Header>

      <StatsRow gutter={16}>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="电池总数"
              value={statistics?.total || 0}
              prefix={<BatteryOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="充电中"
              value={statistics?.charging || 0}
              valueStyle={{ color: '#52c41a' }}
              prefix={<ThunderboltOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="使用中"
              value={statistics?.in_use || 0}
              valueStyle={{ color: '#1890ff' }}
              prefix={<BatteryOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="空闲"
              value={statistics?.idle || 0}
              valueStyle={{ color: '#faad14' }}
              prefix={<ClockCircleOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="故障"
              value={statistics?.fault || 0}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="需保养"
              value={statistics?.needs_maintenance || 0}
              valueStyle={{ color: '#faad14' }}
              prefix={<CheckCircleOutlined />}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <SearchBar>
        <Input
          placeholder="搜索电池ID、型号、位置..."
          value={searchKeyword}
          onChange={e => setSearchKeyword(e.target.value)}
          onPressEnter={handleSearch}
          style={{ width: 240 }}
          prefix={<SearchOutlined />}
          allowClear
        />
        <Select
          placeholder="状态筛选"
          value={statusFilter}
          onChange={value => { setStatusFilter(value); setPage(1) }}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="idle">空闲</Select.Option>
          <Select.Option value="charging">充电中</Select.Option>
          <Select.Option value="in_use">使用中</Select.Option>
          <Select.Option value="discharging">放电中</Select.Option>
          <Select.Option value="storage">存储</Select.Option>
          <Select.Option value="fault">故障</Select.Option>
        </Select>
        <Select
          placeholder="健康状态"
          value={healthFilter}
          onChange={value => { setHealthFilter(value); setPage(1) }}
          style={{ width: 140 }}
          allowClear
        >
          <Select.Option value="excellent">优秀</Select.Option>
          <Select.Option value="good">良好</Select.Option>
          <Select.Option value="fair">一般</Select.Option>
          <Select.Option value="poor">较差</Select.Option>
          <Select.Option value="critical">危险</Select.Option>
        </Select>
        <Button type="primary" onClick={handleSearch}>
          搜索
        </Button>
      </SearchBar>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={batteries}
          rowKey="id"
          loading={loading}
          scroll={{ x: 1400, y: 'calc(100vh - 380px)' }}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条`,
            onChange: (p, ps) => { setPage(p); setPageSize(ps) }
          }}
        />
      </TableContainer>

      <Modal
        title={modalType === 'create' ? '新增电池' : modalType === 'edit' ? '编辑电池' : '电池详情'}
        open={modalVisible}
        onOk={modalType !== 'detail' ? handleModalOk : undefined}
        onCancel={() => setModalVisible(false)}
        width={modalType === 'detail' ? 800 : 600}
        okText="确定"
        cancelText="取消"
        footer={modalType === 'detail' ? [
          <Button key="close" onClick={() => setModalVisible(false)}>关闭</Button>
        ] : undefined}
      >
        {modalType === 'detail' ? (
          renderDetailContent()
        ) : (
          <Form form={form} layout="vertical">
            {modalType === 'create' && (
              <Form.Item
                name="battery_id"
                label="电池ID"
                rules={[{ required: true, message: '请输入电池ID' }]}
              >
                <Input placeholder="请输入电池唯一标识" />
              </Form.Item>
            )}
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="model" label="型号">
                  <Input placeholder="请输入电池型号" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="manufacturer" label="制造商">
                  <Input placeholder="请输入制造商" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="capacity" label="容量">
                  <Input type="number" placeholder="请输入容量" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="capacity_unit" label="容量单位">
                  <Select defaultValue="mAh">
                    <Select.Option value="mAh">mAh</Select.Option>
                    <Select.Option value="Ah">Ah</Select.Option>
                    <Select.Option value="Wh">Wh</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="voltage" label="标称电压(V)">
                  <Input type="number" step="0.1" placeholder="请输入标称电压" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="cell_count" label="电芯数量">
                  <Input type="number" placeholder="请输入电芯串数" />
                </Form.Item>
              </Col>
            </Row>
            {modalType === 'edit' && (
              <Form.Item name="status" label="状态">
                <Select>
                  <Select.Option value="idle">空闲</Select.Option>
                  <Select.Option value="charging">充电中</Select.Option>
                  <Select.Option value="in_use">使用中</Select.Option>
                  <Select.Option value="storage">存储</Select.Option>
                  <Select.Option value="fault">故障</Select.Option>
                </Select>
              </Form.Item>
            )}
            <Form.Item name="location" label="存放位置">
              <Input placeholder="请输入存放位置" />
            </Form.Item>
            <Form.Item name="notes" label="备注">
              <Input.TextArea rows={3} placeholder="请输入备注信息" />
            </Form.Item>
          </Form>
        )}
      </Modal>
    </Container>
  )
}

export default BatteryManagement
