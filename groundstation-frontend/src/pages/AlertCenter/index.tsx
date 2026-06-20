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
  Form,
  message,
  Tag,
  Tooltip,
  Tabs,
  Empty,
  Checkbox,
  Badge,
  Descriptions
} from 'antd'
import {
  SearchOutlined,
  ReloadOutlined,
  CheckCircleOutlined,
  CheckOutlined,
  BellOutlined,
  WarningOutlined,
  ExclamationCircleOutlined,
  InfoCircleOutlined,
  SettingOutlined,
  FilterOutlined,
  DownloadOutlined,
  EyeOutlined
} from '@ant-design/icons'
import { useAlert } from '@/hooks/useAlert'
import { formatDateTime, getSeverityColor } from '@/utils'
import type { Alert, AlertSeverity, AlertCategory, AlertStatus } from '@/types'
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

const StatCard = styled(Card)<{ $color?: string }>`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-left: 3px solid ${props => props.$color || '#1890ff'};

  .ant-statistic-title {
    color: rgba(255, 255, 255, 0.6);
  }

  .ant-statistic-content {
    color: ${props => props.$color || '#fff'};
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
    max-height: 60vh;
    overflow-y: auto;
  }
`

const SettingsModal = styled(Modal)`
  .ant-modal-body {
    max-height: 70vh;
    overflow-y: auto;
  }
`

const AlertCenter: React.FC = () => {
  const {
    alerts,
    unreadCount,
    stats,
    loading,
    total,
    currentPage,
    pageSize,
    filter,
    loadAlerts,
    loadStats,
    acknowledge,
    resolve
  } = useAlert()

  const [keyword, setKeyword] = useState('')
  const [severityFilter, setSeverityFilter] = useState<AlertSeverity | ''>('')
  const [statusFilter, setStatusFilter] = useState<AlertStatus | ''>('')
  const [categoryFilter, setCategoryFilter] = useState<AlertCategory | ''>('')
  const [dateRange, setDateRange] = useState<[dayjs.Dayjs, dayjs.Dayjs] | null>(null)
  const [detailVisible, setDetailVisible] = useState(false)
  const [selectedAlert, setSelectedAlert] = useState<Alert | null>(null)
  const [settingsVisible, setSettingsVisible] = useState(false)
  const [settingsForm] = Form.useForm()
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])

  useEffect(() => {
    loadAlerts({
      page: currentPage,
      pageSize,
      keyword: keyword || undefined,
      severity: severityFilter || undefined,
      status: statusFilter || undefined,
      category: categoryFilter || undefined,
      startTime: dateRange?.[0]?.valueOf(),
      endTime: dateRange?.[1]?.valueOf()
    })
    loadStats()
  }, [currentPage, pageSize, keyword, severityFilter, statusFilter, categoryFilter, dateRange])

  const handleViewDetail = async (alert: Alert) => {
    setSelectedAlert(alert)
    setDetailVisible(true)
    if (alert.status === 'active') {
      await acknowledge(alert.id)
    }
  }

  const handleAcknowledge = async (alert: Alert) => {
    try {
      await acknowledge(alert.id)
      message.success('已确认')
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleResolve = async (alert: Alert) => {
    try {
      await resolve(alert.id, '已解决')
      message.success('已解决')
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleBatchAcknowledge = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择告警')
      return
    }
    try {
      for (const id of selectedRowKeys) {
        await acknowledge(id as string)
      }
      message.success(`已确认 ${selectedRowKeys.length} 条告警`)
      setSelectedRowKeys([])
    } catch (error) {
      message.error('操作失败')
    }
  }

  const handleBatchResolve = async () => {
    if (selectedRowKeys.length === 0) {
      message.warning('请先选择告警')
      return
    }
    Modal.confirm({
      title: '确认解决',
      content: `确定要将选中的 ${selectedRowKeys.length} 条告警标记为已解决吗？`,
      onOk: async () => {
        try {
          for (const id of selectedRowKeys) {
            await resolve(id as string, '批量解决')
          }
          message.success(`已解决 ${selectedRowKeys.length} 条告警`)
          setSelectedRowKeys([])
        } catch (error) {
          message.error('操作失败')
        }
      }
    })
  }

  const handleSaveSettings = async (values: any) => {
    try {
      message.success('设置已保存')
      setSettingsVisible(false)
    } catch (error) {
      message.error('保存失败')
    }
  }

  const getSeverityIcon = (severity: AlertSeverity) => {
    switch (severity) {
      case 'critical':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f', fontSize: 18 }} />
      case 'error':
        return <WarningOutlined style={{ color: '#ff7875', fontSize: 18 }} />
      case 'warning':
        return <WarningOutlined style={{ color: '#faad14', fontSize: 18 }} />
      case 'info':
        return <InfoCircleOutlined style={{ color: '#1890ff', fontSize: 18 }} />
      default:
        return <BellOutlined style={{ fontSize: 18 }} />
    }
  }

  const getSeverityText = (severity: AlertSeverity) => {
    const map: Record<AlertSeverity, string> = {
      critical: '严重',
      error: '错误',
      warning: '警告',
      info: '信息',
      debug: '调试'
    }
    return map[severity] || severity
  }

  const getStatusText = (status: AlertStatus) => {
    const map: Record<AlertStatus, string> = {
      active: '待处理',
      acknowledged: '已确认',
      resolved: '已解决'
    }
    return map[status] || status
  }

  const getCategoryText = (category: AlertCategory) => {
    const map: Record<AlertCategory, string> = {
      system: '系统',
      battery: '电池',
      gps: 'GPS',
      connection: '连接',
      geofence: '围栏',
      mission: '任务',
      safety: '安全',
      other: '其他'
    }
    return map[category] || category
  }

  const columns = [
    {
      title: '级别',
      dataIndex: 'severity',
      key: 'severity',
      width: 80,
      render: (severity: AlertSeverity) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
          {getSeverityIcon(severity)}
          <span style={{ color: getSeverityColor(severity), fontSize: 12, fontWeight: 600 }}>
            {getSeverityText(severity)}
          </span>
        </div>
      )
    },
    {
      title: '标题',
      dataIndex: 'title',
      key: 'title',
      render: (text: string, record: Alert) => (
        <div>
          <div style={{ fontWeight: record.status === 'active' ? 600 : 400 }}>
            {text}
            {record.status === 'active' && (
              <Badge status="processing" color={getSeverityColor(record.severity)} style={{ marginLeft: 8 }} />
            )}
          </div>
          <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.5)', marginTop: 2 }}>
            {record.message}
          </div>
        </div>
      )
    },
    {
      title: '类别',
      dataIndex: 'category',
      key: 'category',
      width: 100,
      render: (category: AlertCategory) => (
        <Tag>{getCategoryText(category)}</Tag>
      )
    },
    {
      title: '无人机',
      dataIndex: 'uavName',
      key: 'uavName',
      width: 120,
      render: (name?: string) => name || '-'
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: AlertStatus) => {
        const colorMap: Record<AlertStatus, string> = {
          active: '#ff4d4f',
          acknowledged: '#faad14',
          resolved: '#52c41a'
        }
        return (
          <Tag color={colorMap[status]}>
            {getStatusText(status)}
          </Tag>
        )
      }
    },
    {
      title: '时间',
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
      render: (_: unknown, record: Alert) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          {record.status === 'active' && (
            <Tooltip title="确认">
              <Button
                type="link"
                size="small"
                icon={<CheckOutlined />}
                onClick={() => handleAcknowledge(record)}
              />
            </Tooltip>
          )}
          {record.status !== 'resolved' && (
            <Tooltip title="解决">
              <Button
                type="link"
                size="small"
                icon={<CheckCircleOutlined />}
                onClick={() => handleResolve(record)}
              />
            </Tooltip>
          )}
        </Space>
      )
    }
  ]

  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys)
    }
  }

  return (
    <Container>
      <Header>
        <Title>
          <BellOutlined style={{ color: '#1890ff' }} />
          告警中心
          <Badge count={unreadCount} size="small" />
        </Title>

        <Space>
          <Button
            icon={<DownloadOutlined />}
          >
            导出
          </Button>
          <Button
            icon={<SettingOutlined />}
            onClick={() => setSettingsVisible(true)}
          >
            通知设置
          </Button>
          <Button
            icon={<ReloadOutlined />}
            onClick={() => {
              loadAlerts()
              loadStats()
            }}
          >
            刷新
          </Button>
        </Space>
      </Header>

      <SearchBar>
        <Input
          placeholder="搜索告警标题/内容"
          prefix={<SearchOutlined />}
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
          style={{ width: 240 }}
          allowClear
        />
        <Select
          placeholder="级别筛选"
          value={severityFilter}
          onChange={setSeverityFilter}
          style={{ width: 120 }}
          allowClear
        >
          <Select.Option value="critical">严重</Select.Option>
          <Select.Option value="error">错误</Select.Option>
          <Select.Option value="warning">警告</Select.Option>
          <Select.Option value="info">信息</Select.Option>
        </Select>
        <Select
          placeholder="状态筛选"
          value={statusFilter}
          onChange={setStatusFilter}
          style={{ width: 120 }}
          allowClear
        >
          <Select.Option value="active">待处理</Select.Option>
          <Select.Option value="acknowledged">已确认</Select.Option>
          <Select.Option value="resolved">已解决</Select.Option>
        </Select>
        <Select
          placeholder="类别筛选"
          value={categoryFilter}
          onChange={setCategoryFilter}
          style={{ width: 120 }}
          allowClear
        >
          <Select.Option value="system">系统</Select.Option>
          <Select.Option value="battery">电池</Select.Option>
          <Select.Option value="gps">GPS</Select.Option>
          <Select.Option value="connection">连接</Select.Option>
          <Select.Option value="geofence">围栏</Select.Option>
          <Select.Option value="mission">任务</Select.Option>
          <Select.Option value="safety">安全</Select.Option>
        </Select>
        <DatePicker.RangePicker
          value={dateRange}
          onChange={(dates) => setDateRange(dates as [dayjs.Dayjs, dayjs.Dayjs])}
          style={{ width: 280 }}
        />
      </SearchBar>

      <StatsRow gutter={16}>
        <Col span={5}>
          <StatCard $color="#ff4d4f">
            <Statistic
              title="严重"
              value={stats?.bySeverity?.critical || 0}
              prefix={<ExclamationCircleOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={5}>
          <StatCard $color="#ff7875">
            <Statistic
              title="错误"
              value={stats?.bySeverity?.error || 0}
              prefix={<WarningOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={5}>
          <StatCard $color="#faad14">
            <Statistic
              title="警告"
              value={stats?.bySeverity?.warning || 0}
              prefix={<WarningOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={5}>
          <StatCard $color="#1890ff">
            <Statistic
              title="信息"
              value={stats?.bySeverity?.info || 0}
              prefix={<InfoCircleOutlined />}
            />
          </StatCard>
        </Col>
        <Col span={4}>
          <StatCard $color="#52c41a">
            <Statistic
              title="待处理"
              value={unreadCount}
              prefix={<BellOutlined />}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer>
        <div style={{ padding: 12, borderBottom: '1px solid rgba(255,255,255,0.1)' }}>
          <Space>
            <Button
              size="small"
              icon={<CheckOutlined />}
              onClick={handleBatchAcknowledge}
              disabled={selectedRowKeys.length === 0}
            >
              批量确认
            </Button>
            <Button
              size="small"
              icon={<CheckCircleOutlined />}
              onClick={handleBatchResolve}
              disabled={selectedRowKeys.length === 0}
            >
              批量解决
            </Button>
            <span style={{ color: 'rgba(255,255,255,0.5)', fontSize: 12 }}>
              已选择 {selectedRowKeys.length} 条
            </span>
          </Space>
        </div>
        <Table
          columns={columns}
          dataSource={alerts}
          rowKey="id"
          loading={loading}
          rowSelection={rowSelection}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 条告警`,
            onChange: (page, size) => {}
          }}
          scroll={{ y: 'calc(100vh - 520px)' }}
          locale={{
            emptyText: (
              <Empty
                description="暂无告警"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )
          }}
        />
      </TableContainer>

      <DetailModal
        title="告警详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          selectedAlert && selectedAlert.status === 'active' && (
            <Button key="ack" icon={<CheckOutlined />} onClick={() => {
              handleAcknowledge(selectedAlert)
              setDetailVisible(false)
            }}>
              确认
            </Button>
          ),
          selectedAlert && selectedAlert.status !== 'resolved' && (
            <Button key="resolve" type="primary" icon={<CheckCircleOutlined />} onClick={() => {
              handleResolve(selectedAlert)
              setDetailVisible(false)
            }}>
              解决
            </Button>
          )
        ]}
        width={700}
      >
        {selectedAlert && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="级别" span={2}>
              <Space>
                {getSeverityIcon(selectedAlert.severity)}
                <Tag color={getSeverityColor(selectedAlert.severity)}>
                  {getSeverityText(selectedAlert.severity)}
                </Tag>
              </Space>
            </Descriptions.Item>
            <Descriptions.Item label="标题" span={2}>
              {selectedAlert.title}
            </Descriptions.Item>
            <Descriptions.Item label="内容" span={2}>
              {selectedAlert.message}
            </Descriptions.Item>
            <Descriptions.Item label="类别">
              {getCategoryText(selectedAlert.category)}
            </Descriptions.Item>
            <Descriptions.Item label="状态">
              <Tag color={selectedAlert.status === 'active' ? '#ff4d4f' : selectedAlert.status === 'acknowledged' ? '#faad14' : '#52c41a'}>
                {getStatusText(selectedAlert.status)}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="无人机">
              {selectedAlert.uavName || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="创建时间">
              {formatDateTime(selectedAlert.createdAt)}
            </Descriptions.Item>
            {selectedAlert.acknowledgedAt && (
              <Descriptions.Item label="确认时间" span={2}>
                {formatDateTime(selectedAlert.acknowledgedAt)}
              </Descriptions.Item>
            )}
            {selectedAlert.resolvedAt && (
              <Descriptions.Item label="解决时间" span={2}>
                {formatDateTime(selectedAlert.resolvedAt)}
              </Descriptions.Item>
            )}
            {selectedAlert.resolutionNotes && (
              <Descriptions.Item label="解决备注" span={2}>
                {selectedAlert.resolutionNotes}
              </Descriptions.Item>
            )}
          </Descriptions>
        )}
      </DetailModal>

      <SettingsModal
        title="通知设置"
        open={settingsVisible}
        onCancel={() => setSettingsVisible(false)}
        footer={null}
        width={600}
      >
        <Form
          form={settingsForm}
          layout="vertical"
          onFinish={handleSaveSettings}
          initialValues={{
            soundEnabled: true,
            voiceEnabled: true,
            notificationEnabled: true,
            criticalSound: true,
            criticalVoice: true,
            criticalNotification: true,
            errorSound: true,
            errorVoice: false,
            errorNotification: true,
            warningSound: false,
            warningVoice: false,
            warningNotification: true,
            infoSound: false,
            infoVoice: false,
            infoNotification: false
          }}
        >
          <Card size="small" title="全局设置" style={{ marginBottom: 16 }}>
            <Form.Item name="soundEnabled" valuePropName="checked">
              <Checkbox>启用告警音效</Checkbox>
            </Form.Item>
            <Form.Item name="voiceEnabled" valuePropName="checked">
              <Checkbox>启用语音播报</Checkbox>
            </Form.Item>
            <Form.Item name="notificationEnabled" valuePropName="checked">
              <Checkbox>启用桌面通知</Checkbox>
            </Form.Item>
          </Card>

          <Card size="small" title="分级通知设置" style={{ marginBottom: 16 }}>
            <Row gutter={16}>
              <Col span={8} style={{ textAlign: 'center', padding: '8px 0', fontWeight: 600 }}>
                级别
              </Col>
              <Col span={5} style={{ textAlign: 'center', padding: '8px 0', fontWeight: 600 }}>
                音效
              </Col>
              <Col span={5} style={{ textAlign: 'center', padding: '8px 0', fontWeight: 600 }}>
                语音
              </Col>
              <Col span={6} style={{ textAlign: 'center', padding: '8px 0', fontWeight: 600 }}>
                桌面通知
              </Col>
            </Row>
            <Row gutter={16} style={{ alignItems: 'center' }}>
              <Col span={8} style={{ padding: '8px 0' }}>
                <Tag color="#ff4d4f">严重</Tag>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="criticalSound" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="criticalVoice" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={6} style={{ textAlign: 'center' }}>
                <Form.Item name="criticalNotification" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16} style={{ alignItems: 'center' }}>
              <Col span={8} style={{ padding: '8px 0' }}>
                <Tag color="#ff7875">错误</Tag>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="errorSound" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="errorVoice" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={6} style={{ textAlign: 'center' }}>
                <Form.Item name="errorNotification" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16} style={{ alignItems: 'center' }}>
              <Col span={8} style={{ padding: '8px 0' }}>
                <Tag color="#faad14">警告</Tag>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="warningSound" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="warningVoice" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={6} style={{ textAlign: 'center' }}>
                <Form.Item name="warningNotification" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
            </Row>
            <Row gutter={16} style={{ alignItems: 'center' }}>
              <Col span={8} style={{ padding: '8px 0' }}>
                <Tag color="#1890ff">信息</Tag>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="infoSound" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={5} style={{ textAlign: 'center' }}>
                <Form.Item name="infoVoice" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
              <Col span={6} style={{ textAlign: 'center' }}>
                <Form.Item name="infoNotification" valuePropName="checked" style={{ margin: 0 }}>
                  <Checkbox />
                </Form.Item>
              </Col>
            </Row>
          </Card>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                保存设置
              </Button>
              <Button onClick={() => setSettingsVisible(false)}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </SettingsModal>
    </Container>
  )
}

export default AlertCenter
