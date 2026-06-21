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
  Tooltip,
  Tabs,
  Empty,
  Progress,
  Descriptions,
  List,
  Badge
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  ThunderboltOutlined,
  ReloadOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  PlayCircleOutlined,
  StopOutlined,
  BatteryOutlined,
  ExportOutlined,
  SettingOutlined,
  ThunderboltFilled,
  MinusOutlined
} from '@ant-design/icons'
import {
  getChargingStationList,
  getChargingStationDetail,
  createChargingStation,
  updateChargingStation,
  deleteChargingStation,
  getChargingStationSlots,
  getChargingStatistics,
  startCharging,
  stopCharging,
  getStationChargingRecords,
  assignBatteryToSlot,
  removeBatteryFromSlot,
  getBatteryList
} from '@/api/charging'
import { formatDateTime, formatDuration } from '@/utils'
import type {
  ChargingStation,
  ChargingSlot,
  ChargingRecord,
  ChargingStationStatus,
  ChargingStatistics as ChargingStatsType,
  CreateChargingStationRequest,
  UpdateChargingStationRequest,
  StartChargingRequest,
  Battery
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

const SlotGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: 12px;
  margin-top: 16px;
`

const SlotCard = styled(Card)<{ $status: string; $level: number }>`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid ${props => {
    switch (props.$status) {
      case 'charging': return 'rgba(82, 196, 26, 0.5)'
      case 'occupied': return 'rgba(24, 144, 255, 0.5)'
      case 'fault': return 'rgba(255, 77, 79, 0.5)'
      default: return 'rgba(255, 255, 255, 0.1)'
    }
  }};
  cursor: pointer;
  transition: all 0.3s;

  &:hover {
    border-color: rgba(24, 144, 255, 0.8);
    transform: translateY(-2px);
  }

  .ant-card-body {
    padding: 16px;
  }
`

const SlotHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
`

const SlotName = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: #fff;
`

const SlotStatusBadge = styled(Tag)`
  margin: 0;
`

const SlotBatteryInfo = styled.div`
  margin-bottom: 12px;
`

const BatteryID = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.7);
  font-family: 'Courier New', monospace;
  margin-bottom: 4px;
`

const SlotProgress = styled(Progress)`
  .ant-progress-text {
    color: rgba(255, 255, 255, 0.7);
    font-size: 12px;
  }
`

const SlotStats = styled.div`
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 8px;
  margin-top: 12px;
`

const SlotStat = styled.div`
  text-align: center;
`

const SlotStatLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 2px;
`

const SlotStatValue = styled.div<{ $color?: string }>`
  font-size: 13px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: ${props => props.$color || '#fff'};
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

const RecordItem = styled.div`
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  margin-bottom: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
`

const RecordHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
`

const RecordStats = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 8px;
`

const ChargingManagement: React.FC = () => {
  const [stations, setStations] = useState<ChargingStation[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [statistics, setStatistics] = useState<ChargingStatsType | null>(null)
  const [searchKeyword, setSearchKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<string | undefined>()

  const [modalVisible, setModalVisible] = useState(false)
  const [modalType, setModalType] = useState<'create' | 'edit' | 'detail'>('create')
  const [selectedStation, setSelectedStation] = useState<ChargingStation | null>(null)
  const [form] = Form.useForm()

  const [slots, setSlots] = useState<ChargingSlot[]>([])
  const [records, setRecords] = useState<ChargingRecord[]>([])
  const [detailTab, setDetailTab] = useState('slots')
  const [recordsTotal, setRecordsTotal] = useState(0)
  const [recordsPage, setRecordsPage] = useState(1)

  const [batteries, setBatteries] = useState<Battery[]>([])
  const [slotModalVisible, setSlotModalVisible] = useState(false)
  const [selectedSlot, setSelectedSlot] = useState<ChargingSlot | null>(null)
  const [slotForm] = Form.useForm()
  const [slotAction, setSlotAction] = useState<'start' | 'assign'>('start')

  const fetchStatistics = async () => {
    try {
      const data = await getChargingStatistics()
      setStatistics(data)
    } catch (error) {
      console.error('Failed to fetch statistics:', error)
    }
  }

  const fetchStations = async () => {
    setLoading(true)
    try {
      const data = await getChargingStationList({
        page,
        pageSize,
        keyword: searchKeyword || undefined,
        status: statusFilter
      })
      setStations(data.list)
      setTotal(data.total)
    } catch (error) {
      message.error('获取充电柜列表失败')
    } finally {
      setLoading(false)
    }
  }

  const fetchBatteries = async () => {
    try {
      const data = await getBatteryList({ page: 1, pageSize: 100, status: 'idle' })
      setBatteries(data.list)
    } catch (error) {
      console.error('Failed to fetch batteries:', error)
    }
  }

  useEffect(() => {
    fetchStatistics()
    fetchStations()
  }, [page, pageSize, statusFilter])

  const handleSearch = () => {
    setPage(1)
    fetchStations()
  }

  const fetchStationDetail = async (id: string) => {
    try {
      const [station, slotData, recordData] = await Promise.all([
        getChargingStationDetail(id),
        getChargingStationSlots(id),
        getStationChargingRecords(id, { page: recordsPage, pageSize: 10 })
      ])
      setSelectedStation(station)
      setSlots(slotData)
      setRecords(recordData.list)
      setRecordsTotal(recordData.total)
    } catch (error) {
      console.error('Failed to fetch station detail:', error)
    }
  }

  const handleCreate = () => {
    setModalType('create')
    setSelectedStation(null)
    form.resetFields()
    setModalVisible(true)
  }

  const handleEdit = (record: ChargingStation) => {
    setModalType('edit')
    setSelectedStation(record)
    form.setFieldsValue({
      name: record.name,
      model: record.model,
      manufacturer: record.manufacturer,
      slot_count: record.slot_count,
      location: record.location,
      ip_address: record.ip_address,
      port: record.port,
      protocol: record.protocol,
      max_voltage: record.max_voltage,
      max_current: record.max_current,
      description: record.description,
      status: record.status
    })
    setModalVisible(true)
  }

  const handleView = (record: ChargingStation) => {
    setModalType('detail')
    setSelectedStation(record)
    fetchStationDetail(record.id)
    setModalVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteChargingStation(id)
      message.success('删除成功')
      fetchStations()
      fetchStatistics()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleModalOk = async () => {
    try {
      const values = await form.validateFields()
      
      if (modalType === 'create') {
        await createChargingStation(values as CreateChargingStationRequest)
        message.success('创建成功')
      } else if (modalType === 'edit' && selectedStation) {
        await updateChargingStation(selectedStation.id, values as UpdateChargingStationRequest)
        message.success('更新成功')
      }
      
      setModalVisible(false)
      fetchStations()
      fetchStatistics()
    } catch (error) {
      console.error('Form error:', error)
    }
  }

  const getStatusColor = (status: ChargingStationStatus): string => {
    switch (status) {
      case 'online': return '#52c41a'
      case 'offline': return '#8c8c8c'
      case 'fault': return '#ff4d4f'
      case 'maintenance': return '#faad14'
      default: return '#8c8c8c'
    }
  }

  const getStatusText = (status: ChargingStationStatus): string => {
    switch (status) {
      case 'online': return '在线'
      case 'offline': return '离线'
      case 'fault': return '故障'
      case 'maintenance': return '维护中'
      default: return '未知'
    }
  }

  const getSlotStatusColor = (status: string): string => {
    switch (status) {
      case 'charging': return '#52c41a'
      case 'occupied': return '#1890ff'
      case 'empty': return '#8c8c8c'
      case 'fault': return '#ff4d4f'
      default: return '#8c8c8c'
    }
  }

  const getSlotStatusText = (status: string): string => {
    switch (status) {
      case 'charging': return '充电中'
      case 'occupied': return '已占用'
      case 'empty': return '空闲'
      case 'fault': return '故障'
      default: return '未知'
    }
  }

  const handleSlotClick = (slot: ChargingSlot) => {
    setSelectedSlot(slot)
    if (slot.status === 'empty') {
      setSlotAction('assign')
      fetchBatteries()
    } else {
      setSlotAction('start')
    }
    slotForm.resetFields()
    setSlotModalVisible(true)
  }

  const handleSlotAction = async () => {
    if (!selectedSlot) return

    try {
      if (slotAction === 'assign') {
        const values = await slotForm.validateFields()
        await assignBatteryToSlot(selectedSlot.id, values.battery_id)
        message.success('分配成功')
      } else if (slotAction === 'start') {
        const values = await slotForm.validateFields()
        await startCharging(selectedSlot.id, values as StartChargingRequest)
        message.success('开始充电')
      }
      
      setSlotModalVisible(false)
      if (selectedStation) {
        fetchStationDetail(selectedStation.id)
      }
      fetchStatistics()
    } catch (error) {
      console.error('Slot action error:', error)
      message.error('操作失败')
    }
  }

  const handleStopCharging = async (slot: ChargingSlot) => {
    try {
      await stopCharging(slot.id, slot.current_level)
      message.success('已停止充电')
      if (selectedStation) {
        fetchStationDetail(selectedStation.id)
      }
      fetchStatistics()
    } catch (error) {
      message.error('停止充电失败')
    }
  }

  const handleRemoveBattery = async (slot: ChargingSlot) => {
    try {
      await removeBatteryFromSlot(slot.id)
      message.success('已移除电池')
      if (selectedStation) {
        fetchStationDetail(selectedStation.id)
      }
      fetchStatistics()
    } catch (error) {
      message.error('移除失败')
    }
  }

  const columns = [
    {
      title: '充电柜ID',
      dataIndex: 'station_id',
      key: 'station_id',
      width: 140,
      render: (text: string) => (
        <span style={{ fontFamily: 'Courier New, monospace' }}>{text}</span>
      )
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      width: 160
    },
    {
      title: '型号',
      dataIndex: 'model',
      key: 'model',
      width: 120
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: ChargingStationStatus) => (
        <Tag color={getStatusColor(status)}>
          <Badge status={status === 'online' ? 'success' : status === 'fault' ? 'error' : 'default'} />
          {getStatusText(status)}
        </Tag>
      )
    },
    {
      title: '槽位数',
      dataIndex: 'slot_count',
      key: 'slot_count',
      width: 100,
      render: (count: number, record: ChargingStation) => (
        <div style={{ fontFamily: 'Courier New, monospace' }}>
          {record.occupied_slots} / {count}
        </div>
      )
    },
    {
      title: '充电中',
      dataIndex: 'charging_slots',
      key: 'charging_slots',
      width: 100,
      render: (count: number) => (
        <span style={{ color: '#52c41a', fontWeight: 600, fontFamily: 'Courier New, monospace' }}>
          {count} 个
        </span>
      )
    },
    {
      title: '位置',
      dataIndex: 'location',
      key: 'location',
      width: 140
    },
    {
      title: 'IP地址',
      dataIndex: 'ip_address',
      key: 'ip_address',
      width: 140,
      render: (ip: string, record: ChargingStation) => (
        <span style={{ fontFamily: 'Courier New, monospace' }}>
          {ip}:{record.port}
        </span>
      )
    },
    {
      title: '最后在线',
      dataIndex: 'last_online_at',
      key: 'last_online_at',
      width: 160,
      render: (time: string) => time ? formatDateTime(time) : '-'
    },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right' as const,
      render: (_: unknown, record: ChargingStation) => (
        <Space>
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => handleView(record)}>
            详情
          </Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(record)}>
            编辑
          </Button>
          <Popconfirm
            title="确定删除该充电柜？"
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
    if (!selectedStation) return null

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
                  <Descriptions.Item label="充电柜ID">{selectedStation.station_id}</Descriptions.Item>
                  <Descriptions.Item label="名称">{selectedStation.name}</Descriptions.Item>
                  <Descriptions.Item label="型号">{selectedStation.model || '-'}</Descriptions.Item>
                  <Descriptions.Item label="制造商">{selectedStation.manufacturer || '-'}</Descriptions.Item>
                  <Descriptions.Item label="状态">
                    <Tag color={getStatusColor(selectedStation.status)}>
                      {getStatusText(selectedStation.status)}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="槽位数量">{selectedStation.slot_count} 个</Descriptions.Item>
                  <Descriptions.Item label="已占用">{selectedStation.occupied_slots} 个</Descriptions.Item>
                  <Descriptions.Item label="充电中">{selectedStation.charging_slots} 个</Descriptions.Item>
                  <Descriptions.Item label="累计充电">{selectedStation.total_charged} 次</Descriptions.Item>
                  <Descriptions.Item label="位置">{selectedStation.location || '-'}</Descriptions.Item>
                  <Descriptions.Item label="IP地址">{selectedStation.ip_address}</Descriptions.Item>
                  <Descriptions.Item label="端口">{selectedStation.port}</Descriptions.Item>
                  <Descriptions.Item label="协议">{selectedStation.protocol}</Descriptions.Item>
                  <Descriptions.Item label="固件版本">{selectedStation.firmware_version || '-'}</Descriptions.Item>
                  <Descriptions.Item label="最大电压">{selectedStation.max_voltage}V</Descriptions.Item>
                  <Descriptions.Item label="最大电流">{selectedStation.max_current}A</Descriptions.Item>
                  <Descriptions.Item label="最后在线">
                    {selectedStation.last_online_at ? formatDateTime(selectedStation.last_online_at) : '-'}
                  </Descriptions.Item>
                  <Descriptions.Item label="描述" span={2}>
                    {selectedStation.description || '-'}
                  </Descriptions.Item>
                </Descriptions>
              )
            },
            {
              key: 'slots',
              label: '充电槽位',
              children: slots.length > 0 ? (
                <SlotGrid>
                  {slots.map(slot => (
                    <SlotCard
                      key={slot.id}
                      $status={slot.status}
                      $level={slot.current_level}
                      onClick={() => handleSlotClick(slot)}
                    >
                      <SlotHeader>
                        <SlotName>{slot.slot_name || `槽位 ${slot.slot_index + 1}`}</SlotName>
                        <SlotStatusBadge color={getSlotStatusColor(slot.status)}>
                          {getSlotStatusText(slot.status)}
                        </SlotStatusBadge>
                      </SlotHeader>
                      
                      {slot.battery_id ? (
                        <>
                          <SlotBatteryInfo>
                            <BatteryID>
                              <BatteryOutlined style={{ marginRight: 4 }} />
                              {slot.battery?.battery_id || '未知电池'}
                            </BatteryID>
                          </SlotBatteryInfo>
                          <SlotProgress
                            percent={slot.current_level}
                            size="small"
                            strokeColor={slot.current_level <= 15 ? '#ff4d4f' : slot.current_level <= 30 ? '#faad14' : '#52c41a'}
                          />
                          <SlotStats>
                            <SlotStat>
                              <SlotStatLabel>电压</SlotStatLabel>
                              <SlotStatValue $color="#1890ff">{slot.current_voltage.toFixed(2)}V</SlotStatValue>
                            </SlotStat>
                            <SlotStat>
                              <SlotStatLabel>电流</SlotStatLabel>
                              <SlotStatValue $color="#52c41a">{slot.current_current.toFixed(2)}A</SlotStatValue>
                            </SlotStat>
                            <SlotStat>
                              <SlotStatLabel>温度</SlotStatLabel>
                              <SlotStatValue $color={slot.temperature > 50 ? '#ff4d4f' : '#faad14'}>
                                {slot.temperature.toFixed(1)}°C
                              </SlotStatValue>
                            </SlotStat>
                            <SlotStat>
                              <SlotStatLabel>剩余</SlotStatLabel>
                              <SlotStatValue>{slot.remaining_time > 0 ? `${slot.remaining_time}min` : '-'}</SlotStatValue>
                            </SlotStat>
                          </SlotStats>
                          {slot.status === 'charging' && (
                            <Button
                              type="primary"
                              danger
                              size="small"
                              icon={<StopOutlined />}
                              style={{ width: '100%', marginTop: 12 }}
                              onClick={e => { e.stopPropagation(); handleStopCharging(slot); }}
                            >
                              停止充电
                            </Button>
                          )}
                          {slot.status !== 'charging' && slot.status !== 'empty' && (
                            <Button
                              type="primary"
                              size="small"
                              icon={<MinusOutlined />}
                              style={{ width: '100%', marginTop: 12 }}
                              onClick={e => { e.stopPropagation(); handleRemoveBattery(slot); }}
                            >
                              移除电池
                            </Button>
                          )}
                        </>
                      ) : (
                        <div style={{ textAlign: 'center', padding: '20px 0', color: 'rgba(255,255,255,0.4)' }}>
                          <BatteryOutlined style={{ fontSize: 32, marginBottom: 8 }} />
                          <div>空闲槽位</div>
                        </div>
                      )}
                    </SlotCard>
                  ))}
                </SlotGrid>
              ) : (
                <Empty description="暂无槽位数据" image={Empty.PRESENTED_IMAGE_SIMPLE} />
              )
            },
            {
              key: 'records',
              label: '充电记录',
              children: records.length > 0 ? (
                <div>
                  {records.map(record => (
                    <RecordItem key={record.id}>
                      <RecordHeader>
                        <Space>
                          <ThunderboltOutlined style={{ color: '#52c41a' }} />
                          <span style={{ color: 'rgba(255,255,255,0.9)', fontWeight: 500 }}>
                            {record.start_time ? formatDateTime(record.start_time) : '未知时间'}
                          </span>
                        </Space>
                        <Tag color={record.status === 'completed' ? 'green' : record.status === 'charging' ? 'blue' : 'orange'}>
                          {record.status === 'completed' ? '已完成' : record.status === 'charging' ? '充电中' : record.status}
                        </Tag>
                      </RecordHeader>
                      <RecordStats>
                        <div>
                          <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.5)' }}>起始电量</div>
                          <div style={{ fontSize: 14, fontWeight: 600, color: '#faad14', fontFamily: 'Courier New, monospace' }}>
                            {record.start_level.toFixed(1)}%
                          </div>
                        </div>
                        <div>
                          <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.5)' }}>结束电量</div>
                          <div style={{ fontSize: 14, fontWeight: 600, color: '#52c41a', fontFamily: 'Courier New, monospace' }}>
                            {record.end_level.toFixed(1)}%
                          </div>
                        </div>
                        <div>
                          <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.5)' }}>充电时长</div>
                          <div style={{ fontSize: 14, fontWeight: 600, color: '#1890ff', fontFamily: 'Courier New, monospace' }}>
                            {formatDuration(record.charging_time * 60)}
                          </div>
                        </div>
                        <div>
                          <div style={{ fontSize: 10, color: 'rgba(255,255,255,0.5)' }}>充电量</div>
                          <div style={{ fontSize: 14, fontWeight: 600, color: '#722ed1', fontFamily: 'Courier New, monospace' }}>
                            {record.charged_capacity.toFixed(1)}%
                          </div>
                        </div>
                      </RecordStats>
                    </RecordItem>
                  ))}
                </div>
              ) : (
                <Empty description="暂无充电记录" image={Empty.PRESENTED_IMAGE_SIMPLE} />
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
          <ThunderboltOutlined style={{ color: '#52c41a' }} />
          智能充电柜管理
        </Title>
        <Space>
          <Button icon={<ReloadOutlined />} onClick={() => { fetchStations(); fetchStatistics(); }}>
            刷新
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={handleCreate}>
            新增充电柜
          </Button>
        </Space>
      </Header>

      <StatsRow gutter={16}>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="充电柜总数"
              value={statistics?.total_stations || 0}
              prefix={<ThunderboltOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="在线"
              value={statistics?.online_stations || 0}
              valueStyle={{ color: '#52c41a' }}
              prefix={<CheckCircleOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="离线"
              value={statistics?.offline_stations || 0}
              valueStyle={{ color: '#8c8c8c' }}
              prefix={<ClockCircleOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="故障"
              value={statistics?.fault_stations || 0}
              valueStyle={{ color: '#ff4d4f' }}
              prefix={<WarningOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="充电中"
              value={statistics?.charging_slots || 0}
              valueStyle={{ color: '#52c41a' }}
              prefix={<ThunderboltFilled />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard>
            <Statistic
              title="可用槽位"
              value={statistics?.available_slots || 0}
              valueStyle={{ color: '#1890ff' }}
              prefix={<BatteryOutlined />}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <SearchBar>
        <Input
          placeholder="搜索充电柜ID、名称、位置..."
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
          <Select.Option value="online">在线</Select.Option>
          <Select.Option value="offline">离线</Select.Option>
          <Select.Option value="fault">故障</Select.Option>
          <Select.Option value="maintenance">维护中</Select.Option>
        </Select>
        <Button type="primary" onClick={handleSearch}>
          搜索
        </Button>
      </SearchBar>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={stations}
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
        title={modalType === 'create' ? '新增充电柜' : modalType === 'edit' ? '编辑充电柜' : '充电柜详情'}
        open={modalVisible}
        onOk={modalType !== 'detail' ? handleModalOk : undefined}
        onCancel={() => setModalVisible(false)}
        width={modalType === 'detail' ? 900 : 600}
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
                name="station_id"
                label="充电柜ID"
                rules={[{ required: true, message: '请输入充电柜ID' }]}
              >
                <Input placeholder="请输入充电柜唯一标识" />
              </Form.Item>
            )}
            <Form.Item
              name="name"
              label="名称"
              rules={modalType === 'create' ? [{ required: true, message: '请输入名称' }] : []}
            >
              <Input placeholder="请输入充电柜名称" />
            </Form.Item>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="model" label="型号">
                  <Input placeholder="请输入型号" />
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
                <Form.Item name="slot_count" label="槽位数量">
                  <Input type="number" placeholder="请输入槽位数量" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="location" label="位置">
                  <Input placeholder="请输入安装位置" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={16}>
                <Form.Item name="ip_address" label="IP地址">
                  <Input placeholder="请输入IP地址" />
                </Form.Item>
              </Col>
              <Col span={8}>
                <Form.Item name="port" label="端口">
                  <Input type="number" placeholder="端口" />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="protocol" label="通信协议">
                  <Select defaultValue="tcp">
                    <Select.Option value="tcp">TCP</Select.Option>
                    <Select.Option value="udp">UDP</Select.Option>
                    <Select.Option value="can">CAN</Select.Option>
                  </Select>
                </Form.Item>
              </Col>
              <Col span={12}>
                {modalType === 'edit' && (
                  <Form.Item name="status" label="状态">
                    <Select>
                      <Select.Option value="online">在线</Select.Option>
                      <Select.Option value="offline">离线</Select.Option>
                      <Select.Option value="fault">故障</Select.Option>
                      <Select.Option value="maintenance">维护中</Select.Option>
                    </Select>
                  </Form.Item>
                )}
              </Col>
            </Row>
            <Row gutter={16}>
              <Col span={12}>
                <Form.Item name="max_voltage" label="最大电压(V)">
                  <Input type="number" step="0.1" placeholder="最大输出电压" />
                </Form.Item>
              </Col>
              <Col span={12}>
                <Form.Item name="max_current" label="最大电流(A)">
                  <Input type="number" step="0.1" placeholder="最大输出电流" />
                </Form.Item>
              </Col>
            </Row>
            <Form.Item name="description" label="描述">
              <Input.TextArea rows={3} placeholder="请输入描述信息" />
            </Form.Item>
          </Form>
        )}
      </Modal>

      <Modal
        title={slotAction === 'assign' ? '分配电池' : '开始充电'}
        open={slotModalVisible}
        onOk={handleSlotAction}
        onCancel={() => setSlotModalVisible(false)}
        okText="确定"
        cancelText="取消"
        width={400}
      >
        <Form form={slotForm} layout="vertical">
          {slotAction === 'assign' ? (
            <Form.Item
              name="battery_id"
              label="选择电池"
              rules={[{ required: true, message: '请选择电池' }]}
            >
              <Select placeholder="请选择要分配的电池">
                {batteries.map(battery => (
                  <Select.Option key={battery.id} value={battery.id}>
                    {battery.battery_id} - {battery.current_level.toFixed(1)}%
                  </Select.Option>
                ))}
              </Select>
            </Form.Item>
          ) : (
            <>
              <Form.Item name="charging_mode" label="充电模式">
                <Select defaultValue="standard">
                  <Select.Option value="standard">标准充电</Select.Option>
                  <Select.Option value="fast">快速充电</Select.Option>
                  <Select.Option value="storage">存储充电</Select.Option>
                  <Select.Option value="balance">均衡充电</Select.Option>
                </Select>
              </Form.Item>
              <Row gutter={16}>
                <Col span={12}>
                  <Form.Item name="target_voltage" label="目标电压(V)">
                    <Input type="number" step="0.1" placeholder="目标电压" />
                  </Form.Item>
                </Col>
                <Col span={12}>
                  <Form.Item name="target_current" label="目标电流(A)">
                    <Input type="number" step="0.1" placeholder="目标电流" />
                  </Form.Item>
                </Col>
              </Row>
            </>
          )}
        </Form>
      </Modal>
    </Container>
  )
}

export default ChargingManagement
