import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Button,
  Space,
  Select,
  Modal,
  Form,
  Input,
  InputNumber,
  message,
  Tabs,
  Card,
  Row,
  Col,
  Statistic,
  Tag,
  Divider
} from 'antd'
import {
  PlusOutlined,
  SaveOutlined,
  UploadOutlined,
  PlayCircleOutlined,
  PauseOutlined,
  StopOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  CopyOutlined,
  ImportOutlined,
  DownloadOutlined
} from '@ant-design/icons'
import FlightMap from '@/components/FlightMap'
import MissionEditor from '@/components/MissionEditor'
import { useMission } from '@/hooks/useMission'
import { useUAV } from '@/hooks/useUAV'
import { formatDistance, formatDuration } from '@/utils'
import type { Waypoint, Mission as MissionType, WaypointAction } from '@/types'

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
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 8px;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
`

const Title = styled.div`
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
`

const MissionSelector = styled(Select)`
  width: 240px;
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 420px;
  gap: 16px;
  overflow: hidden;
  min-height: 0;
`

const MapContainer = styled.div`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;
  position: relative;
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

const StatsBar = styled.div`
  position: absolute;
  bottom: 16px;
  left: 16px;
  right: 16px;
  z-index: 10;
  display: flex;
  gap: 16px;
  padding: 12px 16px;
  background: rgba(0, 0, 0, 0.7);
  backdrop-filter: blur(8px);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.2);
`

const StatItem = styled.div`
  flex: 1;
  text-align: center;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 4px;
`

const StatValue = styled.div`
  font-size: 16px;
  font-weight: 600;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const EditorContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const MissionInfo = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-head {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .ant-card-body {
    padding: 16px;
  }
`

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;

  &:last-child {
    margin-bottom: 0;
  }
`

const InfoLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
`

const InfoValue = styled.span`
  font-weight: 500;
  color: #fff;
`

interface MissionProps {
  missionId?: string
}

const Mission: React.FC<MissionProps> = ({ missionId }) => {
  const { selectedUAVId } = useUAV()
  const {
    missions,
    currentMission,
    waypoints,
    loading,
    executionState,
    createMission,
    updateMission,
    deleteMission,
    selectMission,
    addWaypoint,
    updateWaypoint,
    deleteWaypoint,
    uploadMission,
    startMission,
    pauseMission,
    resumeMission,
    resumeMissionFromBreakpoint,
    stopMission,
    exportMission,
    importMission,
    duplicateMission
  } = useMission()

  const [editMode, setEditMode] = useState<'add' | 'edit' | 'none'>('none')
  const [selectedWaypoint, setSelectedWaypoint] = useState<Waypoint | null>(null)
  const [missionFormVisible, setMissionFormVisible] = useState(false)
  const [form] = Form.useForm()

  useEffect(() => {
    if (missionId) {
      selectMission(missionId)
    }
  }, [missionId, selectMission])

  const handleMissionSelect = (id: string) => {
    selectMission(id)
  }

  const handleWaypointAdd = (waypoint: Waypoint) => {
    if (currentMission) {
      addWaypoint(currentMission.id, waypoint)
    }
  }

  const handleWaypointUpdate = (waypoint: Waypoint) => {
    if (currentMission) {
      updateWaypoint(currentMission.id, waypoint)
    }
  }

  const handleWaypointDelete = (waypointId: string) => {
    if (currentMission) {
      deleteWaypoint(currentMission.id, waypointId)
    }
  }

  const handleCreateMission = async () => {
    try {
      const values = await form.validateFields()
      await createMission(values)
      setMissionFormVisible(false)
      form.resetFields()
      message.success('航线创建成功')
    } catch (error) {
      message.error('创建失败')
    }
  }

  const handleSaveMission = async () => {
    if (currentMission) {
      try {
        await updateMission(currentMission.id, { waypoints })
        message.success('航线保存成功')
      } catch (error) {
        message.error('保存失败')
      }
    }
  }

  const handleUploadMission = async () => {
    if (currentMission && selectedUAVId) {
      try {
        await uploadMission(selectedUAVId, currentMission.id)
        message.success('航线上传成功')
      } catch (error) {
        message.error('上传失败')
      }
    } else {
      message.warning('请先选择无人机')
    }
  }

  const handleStartMission = async () => {
    if (currentMission && selectedUAVId) {
      Modal.confirm({
        title: '确认开始航线任务',
        content: '无人机将按照预设航线执行任务，请确保周围环境安全。',
        okText: '确认开始',
        cancelText: '取消',
        onOk: async () => {
          try {
            await startMission(selectedUAVId, currentMission.id)
            message.success('航线任务已开始')
          } catch (error) {
            message.error('开始失败')
          }
        }
      })
    } else {
      message.warning('请先选择无人机')
    }
  }

  const handlePauseMission = async () => {
    if (currentMission) {
      try {
        await pauseMission(currentMission.id)
        message.success('任务已暂停')
      } catch (error) {
        message.error('暂停失败')
      }
    }
  }

  const handleResumeMission = async () => {
    if (currentMission) {
      try {
        await resumeMission(currentMission.id)
        message.success('任务已继续')
      } catch (error) {
        message.error('继续失败')
      }
    }
  }

  const handleResumeFromBreakpoint = async () => {
    if (currentMission) {
      Modal.confirm({
        title: '确认从断点续飞',
        content: '无人机将从上次中断的航点继续执行任务。',
        okText: '确认续飞',
        cancelText: '取消',
        onOk: async () => {
          try {
            await resumeMissionFromBreakpoint(currentMission.id)
            message.success('已从断点续飞')
          } catch (error) {
            message.error('断点续飞失败')
          }
        }
      })
    }
  }

  const handleStopMission = async () => {
    if (currentMission) {
      Modal.confirm({
        title: '确认停止任务',
        content: '停止后无人机将悬停在当前位置，请选择后续操作（返航/降落）。',
        okText: '确认停止',
        cancelText: '取消',
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            await stopMission(currentMission.id)
            message.success('任务已停止')
          } catch (error) {
            message.error('停止失败')
          }
        }
      })
    }
  }

  const handleExportMission = () => {
    if (currentMission) {
      exportMission(currentMission.id)
      message.success('航线已导出')
    }
  }

  const handleDuplicateMission = async () => {
    if (currentMission) {
      try {
        await duplicateMission(currentMission.id)
        message.success('航线已复制')
      } catch (error) {
        message.error('复制失败')
      }
    }
  }

  const handleDeleteMission = async () => {
    if (currentMission) {
      Modal.confirm({
        title: '确认删除航线',
        content: '删除后将无法恢复，是否继续？',
        okText: '确认删除',
        cancelText: '取消',
        okButtonProps: { danger: true },
        onOk: async () => {
          try {
            await deleteMission(currentMission.id)
            message.success('航线已删除')
          } catch (error) {
            message.error('删除失败')
          }
        }
      })
    }
  }

  const calculateStats = () => {
    if (waypoints.length < 2) {
      return { distance: 0, duration: 0, maxAltitude: 0, minAltitude: 0 }
    }

    let totalDistance = 0
    let maxAlt = -Infinity
    let minAlt = Infinity

    for (let i = 1; i < waypoints.length; i++) {
      const prev = waypoints[i - 1]
      const curr = waypoints[i]
      const dist = Math.sqrt(
        Math.pow(curr.lat - prev.lat, 2) +
        Math.pow(curr.lng - prev.lng, 2) +
        Math.pow((curr.altitude - prev.altitude) / 1000, 2)
      ) * 111000
      totalDistance += dist
      maxAlt = Math.max(maxAlt, curr.altitude, prev.altitude)
      minAlt = Math.min(minAlt, curr.altitude, prev.altitude)
    }

    const avgSpeed = 10
    const duration = totalDistance / avgSpeed

    return {
      distance: totalDistance,
      duration,
      maxAltitude: maxAlt === -Infinity ? 0 : maxAlt,
      minAltitude: minAlt === Infinity ? 0 : minAlt
    }
  }

  const stats = calculateStats()
  const isExecuting = executionState?.status === 'running' || executionState?.status === 'paused'

  const missionOptions = missions.map((m: MissionType) => ({
    value: m.id,
    label: (
      <Space>
        <span>{m.name}</span>
        <Tag color={m.status === 'active' ? '#52c41a' : '#8c8c8c'} style={{ marginLeft: 'auto' }}>
          {m.waypoints?.length || 0} 个航点
        </Tag>
      </Space>
    )
  }))

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <EditOutlined style={{ color: '#1890ff' }} />
            航线规划
          </Title>
          <MissionSelector
            placeholder="选择航线"
            value={currentMission?.id}
            onChange={handleMissionSelect}
            allowClear
            options={missionOptions}
          />
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => setMissionFormVisible(true)}
          >
            新建航线
          </Button>
        </HeaderLeft>

        <Space>
          {currentMission && (
            <>
              <Button
                icon={<SaveOutlined />}
                onClick={handleSaveMission}
                disabled={loading || isExecuting}
              >
                保存
              </Button>
              <Button
                icon={<UploadOutlined />}
                onClick={handleUploadMission}
                disabled={loading || !selectedUAVId || isExecuting}
              >
                上传
              </Button>
              <Button
                icon={<CopyOutlined />}
                onClick={handleDuplicateMission}
                disabled={loading}
              >
                复制
              </Button>
              <Button
                icon={<DownloadOutlined />}
                onClick={handleExportMission}
                disabled={loading}
              >
                导出
              </Button>
              <Divider type="vertical" style={{ height: 30 }} />
              {!isExecuting ? (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={handleStartMission}
                  disabled={loading || !selectedUAVId || waypoints.length === 0}
                  $success
                >
                  开始任务
                </Button>
              ) : (
                <>
                  {executionState?.status === 'paused' ? (
                    <>
                      <Button
                        type="primary"
                        icon={<PlayCircleOutlined />}
                        onClick={handleResumeMission}
                        disabled={loading}
                      >
                        继续
                      </Button>
                      <Button
                        icon={<ImportOutlined />}
                        onClick={handleResumeFromBreakpoint}
                        disabled={loading}
                      >
                        断点续飞
                      </Button>
                    </>
                  ) : (
                    <Button
                      icon={<PauseOutlined />}
                      onClick={handlePauseMission}
                      disabled={loading}
                    >
                      暂停
                    </Button>
                  )}
                  <Button
                    danger
                    icon={<StopOutlined />}
                    onClick={handleStopMission}
                    disabled={loading}
                  >
                    停止
                  </Button>
                </>
              )}
              <Divider type="vertical" style={{ height: 30 }} />
              <Button
                danger
                icon={<DeleteOutlined />}
                onClick={handleDeleteMission}
                disabled={loading || isExecuting}
              >
                删除
              </Button>
            </>
          )}
        </Space>
      </Header>

      <Content>
        <MapContainer>
          <Toolbar>
            <ToolButton
              className={editMode === 'add' ? 'active' : ''}
              icon={<PlusOutlined />}
              onClick={() => setEditMode(editMode === 'add' ? 'none' : 'add')}
              title="添加航点"
            />
            <ToolButton
              className={editMode === 'edit' ? 'active' : ''}
              icon={<EditOutlined />}
              onClick={() => setEditMode(editMode === 'edit' ? 'none' : 'edit')}
              title="编辑航点"
            />
            <ToolButton
              icon={<EyeOutlined />}
              title="查看航线"
            />
            <ToolButton
              icon={<ImportOutlined />}
              title="导入航点"
            />
          </Toolbar>

          <FlightMap
            editable={editMode !== 'none'}
            showMission
            showTrajectory={false}
            onWaypointAdd={handleWaypointAdd}
            onWaypointUpdate={handleWaypointUpdate}
            onWaypointDelete={handleWaypointDelete}
          />

          <StatsBar>
            <StatItem>
              <StatLabel>航点数量</StatLabel>
              <StatValue>{waypoints.length}</StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>总航程</StatLabel>
              <StatValue>{formatDistance(stats.distance)}</StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>预计时间</StatLabel>
              <StatValue>{formatDuration(stats.duration)}</StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>最高高度</StatLabel>
              <StatValue>{stats.maxAltitude.toFixed(1)} m</StatValue>
            </StatItem>
            <StatItem>
              <StatLabel>最低高度</StatLabel>
              <StatValue>{stats.minAltitude.toFixed(1)} m</StatValue>
            </StatItem>
          </StatsBar>
        </MapContainer>

        <EditorContainer>
          {currentMission && (
            <MissionInfo size="small" title="航线信息">
              <InfoRow>
                <InfoLabel>航线名称</InfoLabel>
                <InfoValue>{currentMission.name}</InfoValue>
              </InfoRow>
              <InfoRow>
                <InfoLabel>任务类型</InfoLabel>
                <InfoValue>{currentMission.type || '普通航线'}</InfoValue>
              </InfoRow>
              <InfoRow>
                <InfoLabel>创建时间</InfoLabel>
                <InfoValue>{new Date(currentMission.createdAt).toLocaleString()}</InfoValue>
              </InfoRow>
              {currentMission.description && (
                <InfoRow>
                  <InfoLabel>描述</InfoLabel>
                  <InfoValue>{currentMission.description}</InfoValue>
                </InfoRow>
              )}
            </MissionInfo>
          )}

          <div style={{ flex: 1, overflow: 'hidden' }}>
            <MissionEditor
              onWaypointAdd={handleWaypointAdd}
              onWaypointUpdate={handleWaypointUpdate}
              onWaypointDelete={handleWaypointDelete}
              editable={editMode !== 'none'}
            />
          </div>
        </EditorContainer>
      </Content>

      <Modal
        title="新建航线"
        open={missionFormVisible}
        onCancel={() => {
          setMissionFormVisible(false)
          form.resetFields()
        }}
        footer={null}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={handleCreateMission}
        >
          <Form.Item
            name="name"
            label="航线名称"
            rules={[{ required: true, message: '请输入航线名称' }]}
          >
            <Input placeholder="请输入航线名称" />
          </Form.Item>
          <Form.Item
            name="description"
            label="航线描述"
          >
            <Input.TextArea rows={3} placeholder="请输入航线描述（可选）" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loading}>
                创建
              </Button>
              <Button onClick={() => {
                setMissionFormVisible(false)
                form.resetFields()
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

export default Mission
