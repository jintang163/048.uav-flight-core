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
  Avatar,
  Tooltip,
  Switch,
  Descriptions
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  RocketOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  CloseCircleOutlined,
  ReloadOutlined,
  ExportOutlined,
  SettingOutlined
} from '@ant-design/icons'
import { useUAV } from '@/hooks/useUAV'
import { getStatusColor, formatDateTime, getModeText } from '@/utils'
import type { UAVListItem, UAVStatus, UAVMode } from '@/types'

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
  }

  .ant-table-tbody > tr:hover > td {
    background: rgba(255, 255, 255, 0.05) !important;
  }
`

const UAVAvatar = styled(Avatar)`
  background: linear-gradient(135deg, #1890ff 0%, #52c41a 100%);
`

const BatteryIndicator = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const BatteryBar = styled.div`
  width: 60px;
  height: 8px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 4px;
  overflow: hidden;
`

const BatteryFill = styled.div<{ $level: number }>`
  height: 100%;
  width: ${props => props.$level}%;
  background: ${props => {
    if (props.$level <= 15) return '#ff4d4f'
    if (props.$level <= 30) return '#faad14'
    return '#52c41a'
  }};
  border-radius: 4px;
`

const DetailModal = styled(Modal)`
  .ant-modal-body {
    max-height: 60vh;
    overflow-y: auto;
  }
`

const UAVList: React.FC = () => {
  const {
    uavList,
    total,
    listLoading,
    loadUAVList: fetchUAVList,
    selectCurrentUAV: selectUAV,
    loadUAVDetail: fetchUAVDetail,
    currentUAV,
    loading: detailLoading
  } = useUAV()

  const [keyword, setKeyword] = useState('')
  const [statusFilter, setStatusFilter] = useState<UAVStatus | ''>('')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [detailVisible, setDetailVisible] = useState(false)
  const [formVisible, setFormVisible] = useState(false)
  const [form] = Form.useForm()
  const [editingUAV, setEditingUAV] = useState<UAVListItem | null>(null)

  useEffect(() => {
    loadUAVList()
  }, [currentPage, pageSize, keyword, statusFilter])

  const loadUAVList = () => {
    fetchUAVList({
      page: currentPage,
      pageSize,
      keyword: keyword || undefined,
      status: statusFilter || undefined
    })
  }

  const handleViewDetail = async (uav: UAVListItem) => {
    await fetchUAVDetail(uav.id)
    setDetailVisible(true)
  }

  const handleEdit = (uav: UAVListItem) => {
    setEditingUAV(uav)
    form.setFieldsValue(uav)
    setFormVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      message.success('删除成功')
      loadUAVList()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleSelectUAV = (uav: UAVListItem) => {
    selectUAV(uav.id)
    message.success(`已选择无人机: ${uav.name}`)
  }

  const handleFormSubmit = async (values: any) => {
    try {
      if (editingUAV) {
        message.success('更新成功')
      } else {
        message.success('创建成功')
      }
      setFormVisible(false)
      form.resetFields()
      setEditingUAV(null)
      loadUAVList()
    } catch (error) {
      message.error('操作失败')
    }
  }

  const stats = {
    total: uavList.length,
    online: uavList.filter((u: UAVListItem) => u.status !== 'disconnected' && u.status !== 'error').length,
    flying: uavList.filter((u: UAVListItem) => u.status === 'flying' || u.status === 'takeoff' || u.status === 'hovering').length,
    offline: uavList.filter((u: UAVListItem) => u.status === 'disconnected').length
  }

  const columns = [
    {
      title: '无人机',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: UAVListItem) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <UAVAvatar size={40} icon={<RocketOutlined />} />
          <div>
            <div style={{ fontWeight: 500 }}>{text}</div>
            <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.5)' }}>
              {record.model} | {record.id.slice(0, 8)}
            </div>
          </div>
        </div>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: UAVStatus) => {
        const statusMap: Record<string, { text: string; color: string; icon: React.ReactNode }> = {
          disconnected: { text: '离线', color: '#8c8c8c', icon: <CloseCircleOutlined /> },
          connected: { text: '已连接', color: '#52c41a', icon: <CheckCircleOutlined /> },
          armed: { text: '已解锁', color: '#1890ff', icon: <CheckCircleOutlined /> },
          disarmed: { text: '已上锁', color: '#faad14', icon: <CloseCircleOutlined /> },
          takeoff: { text: '起飞中', color: '#722ed1', icon: <RocketOutlined /> },
          landing: { text: '降落中', color: '#eb2f96', icon: <WarningOutlined /> },
          hovering: { text: '悬停中', color: '#13c2c2', icon: <CheckCircleOutlined /> },
          flying: { text: '飞行中', color: '#1890ff', icon: <RocketOutlined /> },
          return_to_home: { text: '返航中', color: '#fa8c16', icon: <RocketOutlined /> },
          error: { text: '故障', color: '#ff4d4f', icon: <WarningOutlined /> }
        }
        const info = statusMap[status] || statusMap.disconnected
        return (
          <Tag color={info.color} icon={info.icon}>
            {info.text}
          </Tag>
        )
      }
    },
    {
      title: '模式',
      dataIndex: 'mode',
      key: 'mode',
      width: 100,
      render: (mode: UAVMode) => (
        <span style={{ fontFamily: 'monospace' }}>{getModeText(mode)}</span>
      )
    },
    {
      title: '电池',
      dataIndex: 'battery',
      key: 'battery',
      width: 150,
      render: (battery: number) => (
        <BatteryIndicator>
          <BatteryBar>
            <BatteryFill $level={battery} />
          </BatteryBar>
          <span style={{
            color: battery <= 15 ? '#ff4d4f' : battery <= 30 ? '#faad14' : '#52c41a',
            fontFamily: 'monospace',
            fontWeight: 500
          }}>
            {battery}%
          </span>
        </BatteryIndicator>
      )
    },
    {
      title: 'GPS',
      dataIndex: 'gpsSatellites',
      key: 'gpsSatellites',
      width: 100,
      render: (satellites: number) => (
        <span style={{
          color: satellites >= 8 ? '#52c41a' : satellites >= 5 ? '#faad14' : '#ff4d4f',
          fontFamily: 'monospace'
        }}>
          {satellites || 0} 颗
        </span>
      )
    },
    {
      title: '最后在线',
      dataIndex: 'lastSeen',
      key: 'lastSeen',
      width: 180,
      render: (lastSeen: number) => formatDateTime(lastSeen)
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: UAVListItem) => (
        <Space size="small">
          <Tooltip title="选择此无人机">
            <Button
              type="link"
              size="small"
              icon={<RocketOutlined />}
              onClick={() => handleSelectUAV(record)}
            />
          </Tooltip>
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="编辑">
            <Button
              type="link"
              size="small"
              icon={<EditOutlined />}
              onClick={() => handleEdit(record)}
            />
          </Tooltip>
          <Popconfirm
            title="确认删除此无人机？"
            onConfirm={() => handleDelete(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <Container>
      <Header>
        <Title>
          <RocketOutlined style={{ color: '#1890ff' }} />
          无人机管理
        </Title>

        <SearchBar>
          <Input
            placeholder="搜索无人机名称/ID"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            style={{ width: 240 }}
            allowClear
          />
          <Select
            placeholder="状态筛选"
            value={statusFilter}
            onChange={setStatusFilter}
            style={{ width: 120 }}
            allowClear
          >
            <Select.Option value="connected">已连接</Select.Option>
            <Select.Option value="flying">飞行中</Select.Option>
            <Select.Option value="disconnected">离线</Select.Option>
            <Select.Option value="error">故障</Select.Option>
          </Select>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadUAVList}
          >
            刷新
          </Button>
          <Button
            icon={<ExportOutlined />}
          >
            导出
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              setEditingUAV(null)
              form.resetFields()
              setFormVisible(true)
            }}
          >
            添加无人机
          </Button>
        </SearchBar>
      </Header>

      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总数"
              value={stats.total}
              prefix={<RocketOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="在线"
              value={stats.online}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="飞行中"
              value={stats.flying}
              prefix={<RocketOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="离线"
              value={stats.offline}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={uavList}
          rowKey="id"
          loading={listLoading}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 架无人机`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            }
          }}
          scroll={{ y: 'calc(100vh - 420px)' }}
        />
      </TableContainer>

      <DetailModal
        title="无人机详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          <Button key="select" type="primary" onClick={() => {
            if (currentUAV) {
              selectUAV(currentUAV.id)
              setDetailVisible(false)
            }
          }}>
            选择此无人机
          </Button>
        ]}
        width={800}
      >
        {currentUAV && !detailLoading && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="名称" span={2}>{currentUAV.name}</Descriptions.Item>
            <Descriptions.Item label="ID">{currentUAV.id}</Descriptions.Item>
            <Descriptions.Item label="型号">{currentUAV.info?.model}</Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={getStatusColor(currentUAV.status)}>
                {currentUAV.status}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="模式">{getModeText(currentUAV.mode)}</Descriptions.Item>
            <Descriptions.Item label="固件版本">{currentUAV.info?.firmwareVersion}</Descriptions.Item>
            <Descriptions.Item label="硬件版本">{currentUAV.info?.hardwareVersion}</Descriptions.Item>
            <Descriptions.Item label="解锁状态">{currentUAV.armed ? '已解锁' : '已上锁'}</Descriptions.Item>
            <Descriptions.Item label="电池电压">{currentUAV.battery?.voltage?.toFixed(2)} V</Descriptions.Item>
            <Descriptions.Item label="电池电量">{currentUAV.battery?.remaining?.toFixed(0)}%</Descriptions.Item>
            <Descriptions.Item label="GPS卫星">{currentUAV.gpsSatellites || 0} 颗</Descriptions.Item>
            <Descriptions.Item label="GPS定位类型">{currentUAV.gpsFixType || '无'}</Descriptions.Item>
            <Descriptions.Item label="当前位置" span={2}>
              {currentUAV.position?.lat?.toFixed(6)}, {currentUAV.position?.lng?.toFixed(6)}
            </Descriptions.Item>
            <Descriptions.Item label="当前高度">{currentUAV.position?.alt?.toFixed(1)} m</Descriptions.Item>
            <Descriptions.Item label="航向">{currentUAV.heading?.toFixed(0)}°</Descriptions.Item>
            <Descriptions.Item label="空速">{currentUAV.velocity?.airSpeed?.toFixed(1)} m/s</Descriptions.Item>
            <Descriptions.Item label="地速">{currentUAV.velocity?.groundSpeed?.toFixed(1)} m/s</Descriptions.Item>
            <Descriptions.Item label="俯仰角">{currentUAV.attitude?.pitch?.toFixed(1)}°</Descriptions.Item>
            <Descriptions.Item label="横滚角">{currentUAV.attitude?.roll?.toFixed(1)}°</Descriptions.Item>
            <Descriptions.Item label="偏航角">{currentUAV.attitude?.yaw?.toFixed(1)}°</Descriptions.Item>
            <Descriptions.Item label="最后更新" span={2}>
              {formatDateTime(currentUAV.lastUpdate)}
            </Descriptions.Item>
          </Descriptions>
        )}
      </DetailModal>

      <Modal
        title={editingUAV ? '编辑无人机' : '添加无人机'}
        open={formVisible}
        onCancel={() => {
          setFormVisible(false)
          form.resetFields()
          setEditingUAV(null)
        }}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleFormSubmit}
        >
          <Form.Item
            name="name"
            label="无人机名称"
            rules={[{ required: true, message: '请输入无人机名称' }]}
          >
            <Input placeholder="请输入无人机名称" />
          </Form.Item>
          <Form.Item
            name="model"
            label="型号"
            rules={[{ required: true, message: '请输入型号' }]}
          >
            <Input placeholder="例如: MAVIC 3, Phantom 4" />
          </Form.Item>
          <Form.Item
            name="id"
            label="设备ID"
            rules={[{ required: true, message: '请输入设备ID' }]}
          >
            <Input placeholder="请输入设备唯一标识" disabled={!!editingUAV} />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="请输入描述信息（可选）" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingUAV ? '更新' : '添加'}
              </Button>
              <Button onClick={() => {
                setFormVisible(false)
                form.resetFields()
                setEditingUAV(null)
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Container>
  )
}

export default UAVList
