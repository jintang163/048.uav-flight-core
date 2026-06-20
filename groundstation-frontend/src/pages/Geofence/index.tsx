import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Table,
  Button,
  Space,
  Input,
  Select,
  Modal,
  Form,
  Card,
  Row,
  Col,
  Statistic,
  message,
  Tag,
  Tooltip,
  Switch,
  Popconfirm,
  Tabs,
  Empty
} from 'antd'
import {
  PlusOutlined,
  SearchOutlined,
  ReloadOutlined,
  EditOutlined,
  DeleteOutlined,
  EyeOutlined,
  SafetyOutlined,
  WarningOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PolygonOutlined,
  StopOutlined
} from '@ant-design/icons'
import FlightMap from '@/components/FlightMap'
import { useGeofence } from '@/hooks/useGeofence'
import { formatDateTime } from '@/utils'
import type { Geofence, GeofenceType, GeofenceAction } from '@/types'

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

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 420px 1fr;
  gap: 16px;
  overflow: hidden;
  min-height: 0;
`

const SidePanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
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
`

const MapContainer = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  overflow: hidden;
  display: flex;
  flex-direction: column;

  .ant-card-body {
    flex: 1;
    overflow: hidden;
    padding: 0;
  }
`

const Toolbar = styled.div`
  position: absolute;
  top: 16px;
  left: 16px;
  z-index: 10;
  display: flex;
  flex-direction: column;
  gap: 8px;
`

const ToolButton = styled(Button)`
  width: 40px;
  height: 40px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  color: #fff;

  &:hover {
    background: rgba(24, 144, 255, 0.8) !important;
    border-color: #1890ff !important;
    color: #fff !important;
  }

  &.active {
    background: #1890ff;
    border-color: #1890ff;
  }
`

const Geofence: React.FC = () => {
  const {
    geofences,
    loading,
    total,
    stats,
    drawMode,
    selectedGeofenceId,
    fetchGeofences,
    createGeofence,
    updateGeofence,
    deleteGeofence,
    toggleGeofence,
    setDrawMode,
    selectGeofence
  } = useGeofence()

  const [keyword, setKeyword] = useState('')
  const [typeFilter, setTypeFilter] = useState<GeofenceType | ''>('')
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [formVisible, setFormVisible] = useState(false)
  const [form] = Form.useForm()
  const [editingGeofence, setEditingGeofence] = useState<Geofence | null>(null)
  const [activeTab, setActiveTab] = useState('fences')

  useEffect(() => {
    fetchGeofences({
      page: currentPage,
      pageSize,
      keyword: keyword || undefined,
      type: typeFilter || undefined
    })
  }, [currentPage, pageSize, keyword, typeFilter])

  const handleCreate = () => {
    setEditingGeofence(null)
    form.resetFields()
    setFormVisible(true)
  }

  const handleEdit = (geofence: Geofence) => {
    setEditingGeofence(geofence)
    form.setFieldsValue(geofence)
    setFormVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteGeofence(id)
      message.success('删除成功')
    } catch (error) {
      message.error('删除失败')
    }
  }

  const handleToggle = async (geofence: Geofence) => {
    try {
      await toggleGeofence(geofence.id, !geofence.enabled)
      message.success(geofence.enabled ? '已禁用' : '已启用')
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleFormSubmit = async (values: any) => {
    try {
      if (editingGeofence) {
        await updateGeofence(editingGeofence.id, values)
        message.success('更新成功')
      } else {
        await createGeofence(values)
        message.success('创建成功')
      }
      setFormVisible(false)
      form.resetFields()
      setEditingGeofence(null)
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleDrawModeChange = (mode: 'polygon' | 'circle' | 'none') => {
    setDrawMode(mode)
  }

  const getTypeIcon = (type: GeofenceType) => {
    switch (type) {
      case 'polygon':
        return <PolygonOutlined />
      case 'circle':
        return <StopOutlined />
      case 'rectangle':
        return <SafetyOutlined />
      default:
        return <SafetyOutlined />
    }
  }

  const getTypeText = (type: GeofenceType) => {
    const map: Record<GeofenceType, string> = {
      polygon: '多边形',
      circle: '圆形',
      rectangle: '矩形'
    }
    return map[type] || type
  }

  const getActionText = (action: GeofenceAction) => {
    const map: Record<GeofenceAction, string> = {
      warning: '警告',
      rtl: '返航',
      land: '降落',
      hold: '悬停'
    }
    return map[action] || action
  }

  const getActionColor = (action: GeofenceAction) => {
    const map: Record<GeofenceAction, string> = {
      warning: '#faad14',
      rtl: '#1890ff',
      land: '#52c41a',
      hold: '#722ed1'
    }
    return map[action] || '#8c8c8c'
  }

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Geofence) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
          {getTypeIcon(record.type)}
          <span>{text}</span>
        </div>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (type: GeofenceType) => (
        <Tag>{getTypeText(type)}</Tag>
      )
    },
    {
      title: '触发动作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: GeofenceAction) => (
        <Tag color={getActionColor(action)}>
          {getActionText(action)}
        </Tag>
      )
    },
    {
      title: '状态',
      dataIndex: 'enabled',
      key: 'enabled',
      width: 80,
      render: (enabled: boolean) => (
        <Tag color={enabled ? '#52c41a' : '#8c8c8c'} icon={enabled ? <CheckCircleOutlined /> : <CloseCircleOutlined />}>
          {enabled ? '启用' : '禁用'}
        </Tag>
      )
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
      render: (time: number) => formatDateTime(time)
    },
    {
      title: '操作',
      key: 'actions',
      width: 160,
      fixed: 'right',
      render: (_: unknown, record: Geofence) => (
        <Space size="small">
          <Tooltip title={record.enabled ? '禁用' : '启用'}>
            <Switch
              size="small"
              checked={record.enabled}
              onChange={() => handleToggle(record)}
            />
          </Tooltip>
          <Tooltip title="查看">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => selectGeofence(record.id)}
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
            title="确认删除此围栏？"
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

  const violationColumns = [
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      width: 160,
      render: (time: number) => formatDateTime(time)
    },
    {
      title: '无人机',
      dataIndex: 'uavName',
      key: 'uavName',
      width: 120
    },
    {
      title: '围栏名称',
      data: 'geofenceName',
      key: 'geofenceName'
    },
    {
      title: '位置',
      dataIndex: 'position',
      key: 'position',
      render: (pos: { lat: number; lng: number }) => (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
          {pos.lat.toFixed(4)}, {pos.lng.toFixed(4)}
        </span>
      )
    },
    {
      title: '高度',
      dataIndex: 'altitude',
      key: 'altitude',
      width: 80,
      render: (alt: number) => `${alt.toFixed(1)} m`
    },
    {
      title: '触发动作',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: GeofenceAction) => (
        <Tag color={getActionColor(action)}>
          {getActionText(action)}
        </Tag>
      )
    }
  ]

  return (
    <Container>
      <Header>
        <Title>
          <SafetyOutlined style={{ color: '#1890ff' }} />
          电子围栏管理
        </Title>

        <SearchBar>
          <Input
            placeholder="搜索围栏名称"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            style={{ width: 200 }}
            allowClear
          />
          <Select
            placeholder="类型筛选"
            value={typeFilter}
            onChange={setTypeFilter}
            style={{ width: 120 }}
            allowClear
          >
            <Select.Option value="polygon">多边形</Select.Option>
            <Select.Option value="circle">圆形</Select.Option>
            <Select.Option value="rectangle">矩形</Select.Option>
          </Select>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => fetchGeofences()}
          >
            刷新
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={handleCreate}
          >
            新建围栏
          </Button>
        </SearchBar>
      </Header>

      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="围栏总数"
              value={stats?.total || 0}
              prefix={<SafetyOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="已启用"
              value={stats?.enabled || 0}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="已禁用"
              value={stats?.disabled || 0}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#8c8c8c' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="违规次数"
              value={stats?.violations || 0}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <Content>
        <SidePanel>
          <TableContainer>
            <Tabs
              activeKey={activeTab}
              onChange={setActiveTab}
              items={[
                {
                  key: 'fences',
                  label: '围栏列表',
                  children: (
                    <Table
                      columns={columns}
                      dataSource={geofences}
                      rowKey="id"
                      loading={loading}
                      pagination={{
                        current: currentPage,
                        pageSize,
                        total,
                        showSizeChanger: true,
                        showQuickJumper: true,
                        showTotal: (t) => `共 ${t} 个围栏`,
                        onChange: (page, size) => {
                          setCurrentPage(page)
                          setPageSize(size)
                        }
                      }}
                      scroll={{ y: 'calc(100vh - 520px)' }}
                      locale={{
                        emptyText: (
                          <Empty
                            description="暂无围栏"
                            image={Empty.PRESENTED_IMAGE_SIMPLE}
                          />
                        )
                      }}
                    />
                  )
                },
                {
                  key: 'violations',
                  label: '违规记录',
                  children: (
                    <Table
                      columns={violationColumns}
                      dataSource={[]}
                      rowKey="id"
                      scroll={{ y: 'calc(100vh - 520px)' }}
                      locale={{
                        emptyText: (
                          <Empty
                            description="暂无违规记录"
                            image={Empty.PRESENTED_IMAGE_SIMPLE}
                          />
                        )
                      }}
                    />
                  )
                }
              ]}
            />
          </TableContainer>
        </SidePanel>

        <MapContainer title="地图视图" extra={
          <Space>
            <Tag color={drawMode === 'polygon' ? 'blue' : 'default'}>多边形</Tag>
            <Tag color={drawMode === 'circle' ? 'blue' : 'default'}>圆形</Tag>
          </Space>
        }>
          <Toolbar>
            <ToolButton
              className={drawMode === 'polygon' ? 'active' : ''}
              icon={<PolygonOutlined />}
              onClick={() => handleDrawModeChange(drawMode === 'polygon' ? 'none' : 'polygon')}
              title="绘制多边形围栏"
            />
            <ToolButton
              className={drawMode === 'circle' ? 'active' : ''}
              icon={<StopOutlined />}
              onClick={() => handleDrawModeChange(drawMode === 'circle' ? 'none' : 'circle')}
              title="绘制圆形围栏"
            />
          </Toolbar>
          <FlightMap
            showGeofence
            showMission={false}
            showTrajectory={false}
            editable={drawMode !== 'none'}
          />
        </MapContainer>
      </Content>

      <Modal
        title={editingGeofence ? '编辑围栏' : '新建围栏'}
        open={formVisible}
        onCancel={() => {
          setFormVisible(false)
          form.resetFields()
          setEditingGeofence(null)
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
            label="围栏名称"
            rules={[{ required: true, message: '请输入围栏名称' }]}
          >
            <Input placeholder="请输入围栏名称" />
          </Form.Item>
          <Form.Item
            name="type"
            label="围栏类型"
            rules={[{ required: true, message: '请选择围栏类型' }]}
          >
            <Select placeholder="请选择围栏类型">
              <Select.Option value="polygon">多边形</Select.Option>
              <Select.Option value="circle">圆形</Select.Option>
              <Select.Option value="rectangle">矩形</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="action"
            label="触发动作"
            rules={[{ required: true, message: '请选择触发动作' }]}
          >
            <Select placeholder="请选择触发动作">
              <Select.Option value="warning">仅警告</Select.Option>
              <Select.Option value="hold">悬停</Select.Option>
              <Select.Option value="rtl">返航</Select.Option>
              <Select.Option value="land">降落</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item
            name="minAltitude"
            label="最低高度 (米)"
          >
            <Input.Number style={{ width: '100%' }} placeholder="可不填" min={0} />
          </Form.Item>
          <Form.Item
            name="maxAltitude"
            label="最高高度 (米)"
          >
            <Input.Number style={{ width: '100%' }} placeholder="可不填" min={0} />
          </Form.Item>
          <Form.Item
            name="description"
            label="描述"
          >
            <Input.TextArea rows={3} placeholder="请输入描述（可选）" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                {editingGeofence ? '更新' : '创建'}
              </Button>
              <Button onClick={() => {
                setFormVisible(false)
                form.resetFields()
                setEditingGeofence(null)
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

export default Geofence
