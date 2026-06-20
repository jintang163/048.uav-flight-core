import React, { useState, useEffect, useCallback } from 'react'
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
  Switch,
  Popconfirm,
  Tabs,
  DatePicker,
  InputNumber,
  Empty,
  Tooltip
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
  StopOutlined,
  ExclamationCircleOutlined,
  UnlockOutlined,
  ClockCircleOutlined
} from '@ant-design/icons'
import FlightMap from '@/components/FlightMap'
import {
  getGeofenceList,
  createGeofence,
  updateGeofence,
  deleteGeofence,
  getViolationList,
  resolveViolation,
  batchResolveViolations,
  getViolationStatistics,
  getUnlockingList,
  applyUnlocking,
  approveUnlocking,
  rejectUnlocking,
  cancelUnlocking
} from '@/api/geofence'
import { formatDateTime } from '@/utils'
import type {
  Geofence,
  GeofenceType,
  GeofenceShape,
  GeofenceCategory,
  FailAction,
  GeofenceViolation,
  ViolationSeverity,
  TemporaryUnlocking,
  UnlockStatus
} from '@/types'
import type { ColumnsType } from 'antd/es/table'

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

  .ant-table-wrapper {
    flex: 1;
    overflow: auto;
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

const FilterBar = styled.div`
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
  flex-wrap: wrap;
`

const GeofencePage: React.FC = () => {
  const [activeTab, setActiveTab] = useState('fences')
  const [selectedGeofenceId, setSelectedGeofenceId] = useState<string | null>(null)

  return (
    <Container>
      <Header>
        <Title>
          <SafetyOutlined style={{ color: '#1890ff' }} />
          电子围栏与禁飞区管理
        </Title>
      </Header>

      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          { key: 'fences', label: '围栏管理' },
          { key: 'violations', label: '越界日志' },
          { key: 'unlockings', label: '临时解禁' },
          { key: 'restricted', label: '国家禁飞区' }
        ]}
      />

      {activeTab === 'fences' && <FencesTab onSelect={setSelectedGeofenceId} selectedId={selectedGeofenceId} />}
      {activeTab === 'violations' && <ViolationsTab />}
      {activeTab === 'unlockings' && <UnlockingsTab />}
      {activeTab === 'restricted' && <RestrictedTab onSelect={setSelectedGeofenceId} />}
    </Container>
  )
}

const FencesTab: React.FC<{ onSelect: (id: string | null) => void; selectedId: string | null }> = ({ onSelect, selectedId }) => {
  const [fences, setFences] = useState<Geofence[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [category, setCategory] = useState<string>('')
  const [type, setType] = useState<string>('')
  const [keyword, setKeyword] = useState('')
  const [formVisible, setFormVisible] = useState(false)
  const [editingFence, setEditingFence] = useState<Geofence | null>(null)
  const [form] = Form.useForm()

  const loadFences = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getGeofenceList({ page, pageSize, category: category || undefined, type: type || undefined })
      setFences(result.list || [])
      setTotal(result.total || 0)
    } catch {
      message.error('加载围栏列表失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, category, type])

  useEffect(() => { loadFences() }, [loadFences])

  const handleCreate = () => {
    setEditingFence(null)
    form.resetFields()
    form.setFieldsValue({ type: 'exclusion', shape: 'circle', category: 'custom', failAction: 'warn', maxAltitude: 120, maxDistance: 500 })
    setFormVisible(true)
  }

  const handleEdit = (fence: Geofence) => {
    setEditingFence(fence)
    form.setFieldsValue({
      ...fence,
      coordinates: fence.coordinates?.map(c => [c.lat, c.lng])
    })
    setFormVisible(true)
  }

  const handleDelete = async (id: string) => {
    try {
      await deleteGeofence(id)
      message.success('删除成功')
      loadFences()
    } catch {
      message.error('删除失败')
    }
  }

  const handleSubmit = async (values: any) => {
    try {
      if (editingFence) {
        await updateGeofence(editingFence.id, values)
        message.success('更新成功')
      } else {
        await createGeofence(values)
        message.success('创建成功')
      }
      setFormVisible(false)
      form.resetFields()
      setEditingFence(null)
      loadFences()
    } catch {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<Geofence> = [
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
      render: (text, rec) => (
        <a onClick={() => onSelect(rec.id)} style={{ color: selectedId === rec.id ? '#1890ff' : '#fff' }}>{text}</a>
      )
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 80,
      render: (val: GeofenceType) => (
        <Tag color={val === 'exclusion' ? 'red' : 'green'}>
          {val === 'exclusion' ? '禁飞' : '允许'}
        </Tag>
      )
    },
    {
      title: '形状',
      dataIndex: 'shape',
      key: 'shape',
      width: 80,
      render: (val: GeofenceShape) => shapeText(val)
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      width: 80,
      render: (val: GeofenceCategory) => categoryTag(val)
    },
    {
      title: '状态',
      dataIndex: 'isActive',
      key: 'isActive',
      width: 60,
      render: (val: boolean) => (
        <Tag color={val ? 'green' : 'default'}>{val ? '启用' : '禁用'}</Tag>
      )
    },
    {
      title: '操作',
      key: 'actions',
      width: 140,
      render: (_, rec) => (
        <Space size="small">
          <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onSelect(rec.id)}>查看</Button>
          <Button type="link" size="small" icon={<EditOutlined />} onClick={() => handleEdit(rec)}>编辑</Button>
          <Popconfirm title="确认删除?" onConfirm={() => handleDelete(rec.id)}>
            <Button type="link" size="small" danger icon={<DeleteOutlined />}>删除</Button>
          </Popconfirm>
        </Space>
      )
    }
  ]

  return (
    <Content>
      <SidePanel>
        <StatsRow gutter={16}>
          <Col span={12}>
            <StatCard>
              <Statistic title="围栏总数" value={total} />
            </StatCard>
          </Col>
          <Col span={12}>
            <StatCard>
              <Statistic title="启用中" value={fences.filter(f => f.isActive).length} />
            </StatCard>
          </Col>
        </StatsRow>

        <TableContainer title="围栏列表" extra={
          <Button type="primary" size="small" icon={<PlusOutlined />} onClick={handleCreate}>
            新建
          </Button>
        }>
          <FilterBar>
            <Select
              style={{ width: 100 }}
              placeholder="类别"
              allowClear
              value={category || undefined}
              onChange={v => { setCategory(v || ''); setPage(1) }}
              options={[
                { value: 'custom', label: '自定义' },
                { value: 'airport', label: '机场' },
                { value: 'military', label: '军事' }
              ]}
            />
            <Select
              style={{ width: 100 }}
              placeholder="类型"
              allowClear
              value={type || undefined}
              onChange={v => { setType(v || ''); setPage(1) }}
              options={[
                { value: 'exclusion', label: '禁飞区' },
                { value: 'inclusion', label: '允许区' }
              ]}
            />
            <Input
              placeholder="搜索"
              prefix={<SearchOutlined />}
              value={keyword}
              onChange={e => { setKeyword(e.target.value); setPage(1) }}
              style={{ width: 160 }}
            />
          </FilterBar>

          <Table
            columns={columns}
            dataSource={fences}
            rowKey="id"
            size="small"
            loading={loading}
            pagination={{
              current: page,
              pageSize,
              total,
              showSizeChanger: true,
              showQuickJumper: true,
              size: 'small',
              onChange: (p, ps) => { setPage(p); setPageSize(ps) }
            }}
            scroll={{ y: 280 }}
          />
        </TableContainer>
      </SidePanel>

      <MapContainer title="地图展示">
        <FlightMap geofences={fences} selectedGeofenceId={selectedId} />
      </MapContainer>

      <Modal
        title={editingFence ? '编辑围栏' : '新建围栏'}
        open={formVisible}
        onCancel={() => setFormVisible(false)}
        footer={null}
        width={520}
      >
        <Form form={form} layout="vertical" onFinish={handleSubmit}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请输入名称' }]}>
            <Input placeholder="请输入围栏名称" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="type" label="类型" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'exclusion', label: '禁飞区' },
                  { value: 'inclusion', label: '允许区' }
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="shape" label="形状" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'circle', label: '圆形' },
                  { value: 'polygon', label: '多边形' }
                ]} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="category" label="类别">
                <Select options={[
                  { value: 'custom', label: '自定义' },
                  { value: 'airport', label: '机场' },
                  { value: 'military', label: '军事' },
                  { value: 'national', label: '国家级' }
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="failAction" label="越界处置">
                <Select options={[
                  { value: 'warn', label: '仅警告' },
                  { value: 'hover', label: '悬停' },
                  { value: 'rtl', label: '返航' },
                  { value: 'land', label: '降落' }
                ]} />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="description" label="描述">
            <Input.TextArea rows={2} placeholder="请输入描述" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="maxAltitude" label="最大高度(m)">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="maxDistance" label="最远距离(m)">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item noStyle shouldUpdate={(prev, curr) => prev.shape !== curr.shape}>
            {({ getFieldValue }) => {
              const shape = getFieldValue('shape')
              if (shape === 'circle') {
                return (
                  <Row gutter={16}>
                    <Col span={12}>
                      <Form.Item name="centerLat" label="圆心纬度">
                        <InputNumber style={{ width: '100%' }} step={0.0000001} precision={7} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item name="centerLng" label="圆心经度">
                        <InputNumber style={{ width: '100%' }} step={0.0000001} precision={7} />
                      </Form.Item>
                    </Col>
                    <Col span={12}>
                      <Form.Item name="radius" label="半径(m)">
                        <InputNumber style={{ width: '100%' }} min={0} />
                      </Form.Item>
                    </Col>
                  </Row>
                )
              }
              return null
            }}
          </Form.Item>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">保存</Button>
              <Button onClick={() => setFormVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Content>
  )
}

const ViolationsTab: React.FC = () => {
  const [violations, setViolations] = useState<GeofenceViolation[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [severity, setSeverity] = useState<string>('')
  const [isResolved, setIsResolved] = useState<string>('')
  const [stats, setStats] = useState<{ total: number; unresolved: number; critical: number }>({ total: 0, unresolved: 0, critical: 0 })

  const loadViolations = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getViolationList({
        page, pageSize,
        severity: severity || undefined,
        isResolved: isResolved ? isResolved === 'true' : undefined
      })
      setViolations(result.list || [])
      setTotal(result.total || 0)
    } catch {
      message.error('加载越界日志失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, severity, isResolved])

  const loadStats = useCallback(async () => {
    try {
      const data = await getViolationStatistics()
      setStats(data as any)
    } catch { /* ignore */ }
  }, [])

  useEffect(() => { loadViolations() }, [loadViolations])
  useEffect(() => { loadStats() }, [loadStats])

  const handleResolve = async (id: string) => {
    try {
      await resolveViolation(id)
      message.success('已标记为已处理')
      loadViolations()
      loadStats()
    } catch {
      message.error('操作失败')
    }
  }

  const handleBatchResolve = async () => {
    if (violations.filter(v => !v.isResolved).length === 0) {
      message.info('没有待处理的越界记录')
      return
    }
    Modal.confirm({
      title: '批量处理',
      content: '确定将当前所有未处理的越界记录标记为已处理？',
      onOk: async () => {
        try {
          await batchResolveViolations(violations.filter(v => !v.isResolved).map(v => v.id))
          message.success('批量处理成功')
          loadViolations()
          loadStats()
        } catch {
          message.error('操作失败')
        }
      }
    })
  }

  const columns: ColumnsType<GeofenceViolation> = [
    {
      title: '时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 170,
      render: (val: string) => formatDateTime(val)
    },
    {
      title: '围栏',
      dataIndex: 'geofenceName',
      key: 'geofenceName',
      width: 120,
      ellipsis: true
    },
    {
      title: '类别',
      dataIndex: 'geofenceCategory',
      key: 'geofenceCategory',
      width: 80,
      render: (val: GeofenceCategory) => categoryTag(val)
    },
    {
      title: '违规类型',
      dataIndex: 'violationType',
      key: 'violationType',
      width: 130,
      render: (val: string) => violationTypeText(val)
    },
    {
      title: '严重程度',
      dataIndex: 'severity',
      key: 'severity',
      width: 80,
      render: (val: ViolationSeverity) => severityTag(val)
    },
    {
      title: '处置动作',
      dataIndex: 'actionTaken',
      key: 'actionTaken',
      width: 80,
      render: (val: FailAction) => failActionText(val)
    },
    {
      title: '状态',
      dataIndex: 'isResolved',
      key: 'isResolved',
      width: 80,
      render: (val: boolean) => (
        <Tag color={val ? 'green' : 'red'}>{val ? '已处理' : '未处理'}</Tag>
      )
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_, rec) => (
        !rec.isResolved ? (
          <Button type="link" size="small" onClick={() => handleResolve(rec.id)}>
            标记处理
          </Button>
        ) : null
      )
    }
  ]

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 16, overflow: 'hidden' }}>
      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic title="总越界次数" value={stats.total} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="未处理" value={stats.unresolved} valueStyle={{ color: '#ff4d4f' }} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="严重级别" value={stats.critical} valueStyle={{ color: '#faad14' }} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="今日越界" value={0} />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer
        title="越界日志"
        extra={
          <Space>
            <Button size="small" icon={<ReloadOutlined />} onClick={loadViolations}>刷新</Button>
            <Button size="small" type="primary" icon={<CheckCircleOutlined />} onClick={handleBatchResolve}>批量处理</Button>
          </Space>
        }
      >
        <FilterBar>
          <Select
            style={{ width: 120 }}
            placeholder="严重程度"
            allowClear
            value={severity || undefined}
            onChange={v => { setSeverity(v || ''); setPage(1) }}
            options={[
              { value: 'warning', label: '警告' },
              { value: 'critical', label: '严重' },
              { value: 'fatal', label: '致命' }
            ]}
          />
          <Select
            style={{ width: 120 }}
            placeholder="处理状态"
            allowClear
            value={isResolved || undefined}
            onChange={v => { setIsResolved(v || ''); setPage(1) }}
            options={[
              { value: 'true', label: '已处理' },
              { value: 'false', label: '未处理' }
            ]}
          />
        </FilterBar>

        <Table
          columns={columns}
          dataSource={violations}
          rowKey="id"
          size="small"
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            size: 'small',
            onChange: (p, ps) => { setPage(p); setPageSize(ps) }
          }}
          scroll={{ y: 320 }}
        />
      </TableContainer>
    </div>
  )
}

const UnlockingsTab: React.FC = () => {
  const [items, setItems] = useState<TemporaryUnlocking[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [page, setPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [status, setStatus] = useState<string>('')
  const [formVisible, setFormVisible] = useState(false)
  const [form] = Form.useForm()
  const [isAdmin] = useState(true)

  const loadItems = useCallback(async () => {
    setLoading(false)
    try {
      const result = await getUnlockingList({
        page, pageSize,
        status: status || undefined
      })
      setItems(result.list || [])
      setTotal(result.total || 0)
    } catch {
      message.error('加载临时解禁列表失败')
    } finally {
      setLoading(false)
    }
  }, [page, pageSize, status])

  useEffect(() => { loadItems() }, [loadItems])

  const handleApply = async () => {
    try {
      const values = await form.validateFields()
      await applyUnlocking(values)
      message.success('申请提交成功')
      setFormVisible(false)
      form.resetFields()
      loadItems()
    } catch {
      message.error('提交失败')
    }
  }

  const handleApprove = async (id: string) => {
    Modal.confirm({
      title: '批准解禁申请',
      content: '确定批准该临时解禁申请？',
      onOk: async () => {
        try {
          await approveUnlocking(id)
          message.success('已批准')
          loadItems()
        } catch {
          message.error('操作失败')
        }
      }
    })
  }

  const handleReject = async (id: string) => {
    Modal.confirm({
      title: '驳回解禁申请',
      content: '确定驳回该临时解禁申请？',
      okButtonProps: { danger: true },
      onOk: async () => {
        try {
          await rejectUnlocking(id)
          message.success('已驳回')
          loadItems()
        } catch {
          message.error('操作失败')
        }
      }
    })
  }

  const handleCancel = async (id: string) => {
    try {
      await cancelUnlocking(id)
      message.success('已取消')
      loadItems()
    } catch {
      message.error('操作失败')
    }
  }

  const columns: ColumnsType<TemporaryUnlocking> = [
    {
      title: '申请标题',
      dataIndex: 'title',
      key: 'title',
      ellipsis: true
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      width: 80,
      render: (val: GeofenceCategory) => categoryTag(val)
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 90,
      render: (val: UnlockStatus) => unlockStatusTag(val)
    },
    {
      title: '有效期',
      key: 'period',
      width: 200,
      render: (_, rec) => (
        <span style={{ fontSize: 12 }}>
          <ClockCircleOutlined style={{ marginRight: 4 }} />
          {rec.startTime ? formatDateTime(rec.startTime) : '-'} ~ {rec.endTime ? formatDateTime(rec.endTime) : '-'}
        </span>
      )
    },
    {
      title: '申请时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 160,
      render: (val: string) => formatDateTime(val)
    },
    {
      title: '操作',
      key: 'actions',
      width: 180,
      render: (_, rec) => (
        <Space size="small">
          {isAdmin && rec.status === 'pending' && (
            <>
              <Button type="link" size="small" onClick={() => handleApprove(rec.id)}>批准</Button>
              <Button type="link" size="small" danger onClick={() => handleReject(rec.id)}>驳回</Button>
            </>
          )}
          {rec.status === 'pending' && (
            <Button type="link" size="small" onClick={() => handleCancel(rec.id)}>取消</Button>
          )}
        </Space>
      )
    }
  ]

  return (
    <div style={{ flex: 1, display: 'flex', flexDirection: 'column', gap: 16, overflow: 'hidden' }}>
      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic title="总申请数" value={total} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="待审批" value={items.filter(i => i.status === 'pending').length} valueStyle={{ color: '#faad14' }} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="已通过" value={items.filter(i => i.status === 'approved').length} valueStyle={{ color: '#52c41a' }} />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic title="已驳回" value={items.filter(i => i.status === 'rejected').length} valueStyle={{ color: '#ff4d4f' }} />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer
        title="临时解禁申请"
        extra={
          <Button type="primary" size="small" icon={<UnlockOutlined />} onClick={() => setFormVisible(true)}>
            申请解禁
          </Button>
        }
      >
        <FilterBar>
          <Select
            style={{ width: 120 }}
            placeholder="状态"
            allowClear
            value={status || undefined}
            onChange={v => { setStatus(v || ''); setPage(1) }}
            options={[
              { value: 'pending', label: '待审批' },
              { value: 'approved', label: '已通过' },
              { value: 'rejected', label: '已驳回' },
              { value: 'expired', label: '已过期' }
            ]}
          />
        </FilterBar>

        <Table
          columns={columns}
          dataSource={items}
          rowKey="id"
          size="small"
          loading={loading}
          pagination={{
            current: page,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            size: 'small',
            onChange: (p, ps) => { setPage(p); setPageSize(ps) }
          }}
          scroll={{ y: 320 }}
        />
      </TableContainer>

      <Modal title="申请临时解禁" open={formVisible} onCancel={() => setFormVisible(false)} footer={null} width={520}>
        <Form form={form} layout="vertical">
          <Form.Item name="title" label="申请标题" rules={[{ required: true }]}>
            <Input placeholder="请输入申请标题" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="category" label="禁飞区类别" rules={[{ required: true }]}>
                <Select options={[
                  { value: 'custom', label: '自定义' },
                  { value: 'airport', label: '机场' },
                  { value: 'military', label: '军事' }
                ]} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="uavId" label="目标无人机">
                <Input placeholder="请输入无人机ID" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item name="reason" label="申请理由" rules={[{ required: true }]}>
            <Input.TextArea rows={3} placeholder="请详细说明解禁事由" />
          </Form.Item>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="startTime" label="开始时间" rules={[{ required: true }]}>
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="endTime" label="结束时间" rules={[{ required: true }]}>
                <DatePicker showTime style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="maxAltitude" label="最大高度(m)">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="radius" label="活动半径(m)">
                <InputNumber style={{ width: '100%' }} min={0} />
              </Form.Item>
            </Col>
          </Row>
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item name="contactName" label="联系人">
                <Input placeholder="联系人姓名" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="contactPhone" label="联系电话">
                <Input placeholder="联系电话" />
              </Form.Item>
            </Col>
          </Row>
          <Form.Item>
            <Space>
              <Button type="primary" onClick={handleApply}>提交申请</Button>
              <Button onClick={() => setFormVisible(false)}>取消</Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

const RestrictedTab: React.FC<{ onSelect: (id: string | null) => void }> = ({ onSelect }) => {
  const [fences, setFences] = useState<Geofence[]>([])
  const [loading, setLoading] = useState(false)
  const [category, setCategory] = useState<string>('')

  const loadFences = useCallback(async () => {
    setLoading(true)
    try {
      const result = await getGeofenceList({
        page: 1, pageSize: 200,
        source: 'national',
        category: category || undefined
      })
      setFences(result.list || [])
    } catch {
      message.error('加载禁飞区失败')
    } finally {
      setLoading(false)
    }
  }, [category])

  useEffect(() => { loadFences() }, [loadFences])

  const columns: ColumnsType<Geofence> = [
    { title: '名称', dataIndex: 'name', key: 'name' },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (val: GeofenceCategory) => categoryTag(val)
    },
    {
      title: '形状',
      dataIndex: 'shape',
      key: 'shape',
      width: 80,
      render: (val: GeofenceShape) => shapeText(val)
    },
    {
      title: '高度限制',
      dataIndex: 'maxAltitude',
      key: 'maxAltitude',
      width: 100,
      render: (val: number) => val ? `${val}m` : '-'
    },
    {
      title: '所在城市',
      dataIndex: 'cityName',
      key: 'cityName',
      width: 120
    },
    {
      title: '操作',
      key: 'action',
      width: 80,
      render: (_, rec) => (
        <Button type="link" size="small" icon={<EyeOutlined />} onClick={() => onSelect(rec.id)}>
          定位
        </Button>
      )
    }
  ]

  return (
    <Content>
      <SidePanel>
        <TableContainer title="国家禁飞区" extra={
          <Tooltip title="数据来源：民用航空局公布的禁飞区域">
            <ExclamationCircleOutlined style={{ color: '#faad14' }} />
          </Tooltip>
        }>
          <FilterBar>
            <Select
              style={{ width: 140 }}
              placeholder="禁飞区类别"
              allowClear
              value={category || undefined}
              onChange={v => setCategory(v || '')}
              options={[
                { value: 'airport', label: '机场净空区' },
                { value: 'military', label: '军事禁区' },
                { value: 'nuclear', label: '核设施' },
                { value: 'government', label: '政府机关' },
                { value: 'national', label: '国家级' }
              ]}
            />
          </FilterBar>
          <Table
            columns={columns}
            dataSource={fences}
            rowKey="id"
            size="small"
            loading={loading}
            scroll={{ y: 380 }}
            pagination={{ pageSize: 50, size: 'small' }}
          />
          {fences.length === 0 && !loading && (
            <Empty description="暂无禁飞区数据" style={{ padding: 40 }} />
          )}
        </TableContainer>
      </SidePanel>
      <MapContainer title="禁飞区分布">
        <FlightMap geofences={fences} showNationalOnly />
      </MapContainer>
    </Content>
  )
}

const shapeText = (shape: GeofenceShape) => {
  const map: Record<GeofenceShape, string> = {
    polygon: '多边形',
    circle: '圆形',
    rectangle: '矩形'
  }
  return map[shape] || shape
}

const categoryTag = (category: GeofenceCategory) => {
  const colorMap: Record<GeofenceCategory, string> = {
    custom: 'blue',
    airport: 'red',
    military: 'orange',
    nuclear: 'purple',
    government: 'cyan',
    national: 'red'
  }
  const textMap: Record<GeofenceCategory, string> = {
    custom: '自定义',
    airport: '机场',
    military: '军事',
    nuclear: '核设施',
    government: '政府',
    national: '国家级'
  }
  return <Tag color={colorMap[category] || 'default'}>{textMap[category] || category}</Tag>
}

const severityTag = (severity: ViolationSeverity) => {
  const colorMap: Record<ViolationSeverity, string> = {
    warning: 'gold',
    critical: 'red',
    fatal: 'magenta'
  }
  const textMap: Record<ViolationSeverity, string> = {
    warning: '警告',
    critical: '严重',
    fatal: '致命'
  }
  return <Tag color={colorMap[severity] || 'default'}>{textMap[severity] || severity}</Tag>
}

const violationTypeText = (type: string) => {
  const map: Record<string, string> = {
    altitude_exceeded: '超高度',
    altitude_too_low: '高度过低',
    inside_exclusion_zone: '进入禁飞区',
    outside_inclusion_zone: '超出允许区',
    distance_exceeded: '超距离'
  }
  return map[type] || type
}

const failActionText = (action: FailAction) => {
  const map: Record<FailAction, string> = {
    warn: '警告',
    hover: '悬停',
    rtl: '返航',
    land: '降落'
  }
  return map[action] || action
}

const unlockStatusTag = (status: UnlockStatus) => {
  const colorMap: Record<UnlockStatus, string> = {
    pending: 'gold',
    approved: 'green',
    rejected: 'red',
    expired: 'default',
    cancelled: 'default'
  }
  const textMap: Record<UnlockStatus, string> = {
    pending: '待审批',
    approved: '已通过',
    rejected: '已驳回',
    expired: '已过期',
    cancelled: '已取消'
  }
  return <Tag color={colorMap[status] || 'default'}>{textMap[status] || status}</Tag>
}

export default GeofencePage
