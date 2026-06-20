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
  Progress,
  Upload,
  Descriptions,
  Empty
} from 'antd'
import {
  UploadOutlined,
  ReloadOutlined,
  CloudUploadOutlined,
  DeleteOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  WarningOutlined,
  ClockCircleOutlined,
  RocketOutlined,
  SearchOutlined
} from '@ant-design/icons'
import { getFirmwareList, getFirmwareDetail, uploadFirmware, updateFirmware, getUpdateProgress, checkNewVersion } from '@/api/ota'
import { useUAV } from '@/hooks/useUAV'
import { formatDateTime, formatFileSize } from '@/utils'
import type { FirmwareInfo } from '@/types'
import type { UploadProps } from 'antd'

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
`

const UpdateModal = styled(Modal)`
  .ant-modal-body {
    max-height: 60vh;
    overflow-y: auto;
  }
`

const ProgressContainer = styled.div`
  padding: 20px 0;
`

const Firmware: React.FC = () => {
  const { uavList, selectedUAVId, currentUAV } = useUAV()
  const [firmwares, setFirmwares] = useState<FirmwareInfo[]>([])
  const [loading, setLoading] = useState(false)
  const [total, setTotal] = useState(0)
  const [currentPage, setCurrentPage] = useState(1)
  const [pageSize, setPageSize] = useState(10)
  const [keyword, setKeyword] = useState('')
  const [hardwareFilter, setHardwareFilter] = useState<string>('')
  const [detailVisible, setDetailVisible] = useState(false)
  const [updateVisible, setUpdateVisible] = useState(false)
  const [uploadVisible, setUploadVisible] = useState(false)
  const [selectedFirmware, setSelectedFirmware] = useState<FirmwareInfo | null>(null)
  const [updateProgress, setUpdateProgress] = useState(0)
  const [updating, setUpdating] = useState(false)
  const [updateForm] = Form.useForm()
  const [uploadForm] = Form.useForm()

  useEffect(() => {
    loadFirmwares()
  }, [currentPage, pageSize, keyword, hardwareFilter])

  const loadFirmwares = async () => {
    setLoading(true)
    try {
      const params: any = {
        page: currentPage,
        pageSize
      }
      if (keyword) params.keyword = keyword
      if (hardwareFilter) params.hardware = hardwareFilter
      const result = await getFirmwareList(params)
      setFirmwares(result.list)
      setTotal(result.total)
    } catch (error) {
      message.error('加载固件列表失败')
    } finally {
      setLoading(false)
    }
  }

  const handleViewDetail = async (firmware: FirmwareInfo) => {
    try {
      const detail = await getFirmwareDetail(firmware.id)
      setSelectedFirmware(detail)
      setDetailVisible(true)
    } catch (error) {
      message.error('加载详情失败')
    }
  }

  const handleUpdate = (firmware: FirmwareInfo) => {
    if (!selectedUAVId) {
      message.warning('请先选择无人机')
      return
    }
    setSelectedFirmware(firmware)
    updateForm.setFieldsValue({
      firmwareId: firmware.id,
      uavId: selectedUAVId
    })
    setUpdateVisible(true)
  }

  const handleUpdateSubmit = async (values: any) => {
    Modal.confirm({
      title: '确认更新固件',
      content: `确定要将无人机固件更新到 ${selectedFirmware?.version} 版本吗？更新过程中请勿断电或断开连接。',
      okText: '确认更新',
      cancelText: '取消',
      onOk: async () => {
        setUpdating(true)
        setUpdateProgress(0)
        try {
          await updateFirmware(values.uavId, values.firmwareId)
          
          const progressInterval = setInterval(async () => {
            try {
              const progress = await getUpdateProgress(values.uavId)
              setUpdateProgress(progress.progress)
              if (progress.status === 'completed') {
                clearInterval(progressInterval)
                setUpdating(false)
                message.success('固件更新成功')
                setUpdateVisible(false)
              } else if (progress.status === 'failed') {
                clearInterval(progressInterval)
                setUpdating(false)
                message.error('固件更新失败')
              }
            } catch (error) {
              clearInterval(progressInterval)
              setUpdating(false)
            }
          }, 1000)
        } catch (error) {
          setUpdating(false)
          message.error('更新失败')
        }
      }
    })
  }

  const handleCheckNewVersion = async () => {
    try {
      const result = await checkNewVersion()
      if (result.hasNewVersion) {
        message.info(`发现新版本: ${result.latestVersion}`)
      } else {
        message.success('当前已是最新版本')
      }
    } catch (error) {
      message.error('检查更新失败')
    }
  }

  const handleDelete = async (id: string) => {
    try {
      message.success('删除成功')
      loadFirmwares()
    } catch (error) {
      message.error('删除失败')
    }
  }

  const uploadProps: UploadProps = {
    name: 'file',
    action: '/api/ota/upload',
    headers: {
      authorization: `Bearer ${localStorage.getItem('accessToken')}`
    },
    onChange(info) {
      if (info.file.status === 'done') {
        message.success(`${info.file.name} 上传成功`)
        loadFirmwares()
      } else if (info.file.status === 'error') {
        message.error(`${info.file.name} 上传失败`)
      }
    }
  }

  const stats = {
    total: firmwares.length,
    stable: firmwares.filter(f => f.version.includes('stable')).length,
    beta: firmwares.filter(f => f.version.includes('beta')).length,
    totalSize: firmwares.reduce((sum, f) => sum + f.fileSize, 0)
  }

  const columns = [
    {
      title: '固件名称',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: FirmwareInfo) => (
        <div>
          <div style={{ fontWeight: 500 }}>{text}</div>
          <div style={{ fontSize: 12, color: 'rgba(255,255,255,0.5)', marginTop: 2 }}>
            {record.hardware}
          </div>
        </div>
      )
    },
    {
      title: '版本号',
      dataIndex: 'version',
      key: 'version',
      width: 140,
      render: (version: string) => {
        const isStable = version.includes('stable')
        return (
          <Tag color={isStable ? '#52c41a' : '#faad14'}>
            {version}
          </Tag>
        )
      }
    },
    {
      title: '硬件平台',
      dataIndex: 'hardware',
      key: 'hardware',
      width: 120
    },
    {
      title: '文件大小',
      dataIndex: 'fileSize',
      key: 'fileSize',
      width: 120,
      render: (size: number) => formatFileSize(size)
    },
    {
      title: '发布日期',
      dataIndex: 'releaseDate',
      key: 'releaseDate',
      width: 160,
      render: (date: number) => formatDateTime(date)
    },
    {
      title: 'MD5校验',
      dataIndex: 'checksum',
      key: 'checksum',
      width: 140,
      render: (checksum?: string) => checksum ? (
        <span style={{ fontFamily: 'monospace', fontSize: 12 }}>
          {checksum.slice(0, 8)}...
        </span>
      ) : '-'
    },
    {
      title: '操作',
      key: 'actions',
      width: 200,
      fixed: 'right',
      render: (_: unknown, record: FirmwareInfo) => (
        <Space size="small">
          <Tooltip title="查看详情">
            <Button
              type="link"
              size="small"
              icon={<EyeOutlined />}
              onClick={() => handleViewDetail(record)}
            />
          </Tooltip>
          <Tooltip title="更新固件">
            <Button
              type="link"
              size="small"
              icon={<CloudUploadOutlined />}
              onClick={() => handleUpdate(record)}
            />
          </Tooltip>
          <Tooltip title="删除">
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDelete(record.id)}
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
          <CloudUploadOutlined style={{ color: '#1890ff' }} />
          固件管理
        </Title>

        <SearchBar>
          <Input
            placeholder="搜索固件名称"
            prefix={<SearchOutlined />}
            value={keyword}
            onChange={(e) => setKeyword(e.target.value)}
            style={{ width: 200 }}
            allowClear
          />
          <Select
            placeholder="硬件平台"
            value={hardwareFilter}
            onChange={setHardwareFilter}
            style={{ width: 140 }}
            allowClear
          >
            <Select.Option value="px4">PX4</Select.Option>
            <Select.Option value="apm">APM</Select.Option>
            <Select.Option value="ardupilot">ArduPilot</Select.Option>
            <Select.Option value="betaflight">Betaflight</Select.Option>
          </Select>
          <Button
            icon={<ReloadOutlined />}
            onClick={loadFirmwares}
          >
            刷新
          </Button>
          <Button
            icon={<CheckCircleOutlined />}
            onClick={handleCheckNewVersion}
          >
            检查更新
          </Button>
          <Upload {...uploadProps} showUploadList={false}>
            <Button
              icon={<UploadOutlined />}
            >
              上传固件
            </Button>
          </Upload>
        </SearchBar>
      </Header>

      <StatsRow gutter={16}>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="固件总数"
              value={stats.total}
              prefix={<CloudUploadOutlined />}
              valueStyle={{ color: '#1890ff' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="稳定版"
              value={stats.stable}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#52c41a' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="测试版"
              value={stats.beta}
              prefix={<WarningOutlined />}
              valueStyle={{ color: '#faad14' }}
            />
          </StatCard>
        </Col>
        <Col span={6}>
          <StatCard>
            <Statistic
              title="总大小"
              value={formatFileSize(stats.totalSize)}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#722ed1' }}
            />
          </StatCard>
        </Col>
      </StatsRow>

      <TableContainer>
        <Table
          columns={columns}
          dataSource={firmwares}
          rowKey="id"
          loading={loading}
          pagination={{
            current: currentPage,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (t) => `共 ${t} 个固件`,
            onChange: (page, size) => {
              setCurrentPage(page)
              setPageSize(size)
            }
          }}
          scroll={{ y: 'calc(100vh - 420px)' }}
          locale={{
            emptyText: (
              <Empty
                description="暂无固件"
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )
          }}
        />
      </TableContainer>

      <Modal
        title="固件详情"
        open={detailVisible}
        onCancel={() => setDetailVisible(false)}
        footer={[
          <Button key="close" onClick={() => setDetailVisible(false)}>
            关闭
          </Button>,
          <Button key="update" type="primary" icon={<CloudUploadOutlined />} onClick={() => {
            if (selectedFirmware) {
              handleUpdate(selectedFirmware)
              setDetailVisible(false)
            }
          }}>
            更新此固件
          </Button>
        ]}
        width={700}
      >
        {selectedFirmware && (
          <Descriptions bordered column={2} size="small">
            <Descriptions.Item label="固件名称" span={2}>
              {selectedFirmware.name}
            </Descriptions.Item>
            <Descriptions.Item label="版本号">
              <Tag color={selectedFirmware.version.includes('stable') ? '#52c41a' : '#faad14'}>
                {selectedFirmware.version}
              </Tag>
            </Descriptions.Item>
            <Descriptions.Item label="硬件平台">
              {selectedFirmware.hardware}
            </Descriptions.Item>
            <Descriptions.Item label="文件大小">
              {formatFileSize(selectedFirmware.fileSize)}
            </Descriptions.Item>
            <Descriptions.Item label="发布日期">
              {formatDateTime(selectedFirmware.releaseDate)}
            </Descriptions.Item>
            <Descriptions.Item label="MD5校验">
              {selectedFirmware.checksum || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="描述" span={2}>
              {selectedFirmware.description || '-'}
            </Descriptions.Item>
            <Descriptions.Item label="更新日志" span={2}>
              {selectedFirmware.changelog || '-'}
            </Descriptions.Item>
          </Descriptions>
        )}
      </Modal>

      <UpdateModal
        title="更新固件"
        open={updateVisible}
        onCancel={() => {
          setUpdateVisible(false)
          setUpdateProgress(0)
          setUpdating(false)
        }}
        footer={null}
        width={500}
      >
        {updating ? (
          <ProgressContainer>
            <Progress
              percent={updateProgress}
              status={updateProgress === 100 ? 'success' : 'active'}
              strokeColor={{
                '0%': '#1890ff',
                '100%': '#52c41a'
              }}
            />
            <div style={{ textAlign: 'center', marginTop: 16, color: 'rgba(255,255,255,0.7)' }}>
              正在更新固件，请勿关闭页面或断开连接...
            </div>
          </ProgressContainer>
        ) : (
          <Form
            form={updateForm}
            layout="vertical"
            onFinish={handleUpdateSubmit}
          >
            <Form.Item
              name="uavId"
              label="选择无人机"
              rules={[{ required: true, message: '请选择无人机' }]}
            >
              <Select placeholder="请选择要更新的无人机">
                {uavList.filter((u: any) => u.status === 'online').map((uav: any) => (
                <Select.Option key={uav.id} value={uav.id}>
                  {uav.name}
                </Select.Option>
              ))}
              </Select>
            </Form.Item>
            <Form.Item
              name="firmwareId"
              label="固件版本"
              rules={[{ required: true, message: '请选择固件' }]}
            >
              <Select disabled>
                <Select.Option value={selectedFirmware?.id}>
                  {selectedFirmware?.name} - {selectedFirmware?.version}
                </Select.Option>
              </Select>
            </Form.Item>
            <Form.Item>
              <Space>
                <Button type="primary" htmlType="submit" loading={updating}>
                  开始更新
                </Button>
                <Button onClick={() => setUpdateVisible(false)}>
                  取消
                </Button>
              </Space>
            </Form.Item>
          </Form>
        )}
      </UpdateModal>
    </Container>
  )
}

export default Firmware