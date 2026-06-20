import React, { useState } from 'react'
import styled from 'styled-components'
import { Card, Button, Input, Select, Table, Modal, Form, InputNumber, message, Popconfirm } from 'antd'
import { PlusOutlined, DeleteOutlined, EditOutlined, SaveOutlined, PlayCircleOutlined, PauseCircleOutlined, StopOutlined, UploadOutlined } from '@ant-design/icons'
import { useMission } from '@/hooks/useMission'
import type { Waypoint, WaypointAction } from '@/types'

const Container = styled.div`
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  gap: 12px;
`

const Toolbar = styled.div`
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
`

const WaypointActionTag = styled.span<{ $color: string }>`
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 12px;
  background: ${props => props.$color}20;
  color: ${props => props.$color};
`

const actionOptions = [
  { value: 'waypoint', label: '航点', color: '#1890ff' },
  { value: 'loiter_unlimited', label: '无限悬停', color: '#52c41a' },
  { value: 'loiter_time', label: '定时悬停', color: '#52c41a' },
  { value: 'loiter_turns', label: '定圈悬停', color: '#52c41a' },
  { value: 'rtl', label: '返航', color: '#faad14' },
  { value: 'land', label: '降落', color: '#ff4d4f' },
  { value: 'takeoff', label: '起飞', color: '#722ed1' },
  { value: 'delay', label: '延迟', color: '#13c2c2' },
  { value: 'camera_trigger', label: '拍照', color: '#eb2f96' }
]

interface MissionEditorProps {
  missionId?: string
  onStart?: () => void
  onPause?: () => void
  onStop?: () => void
  onUpload?: () => void
}

const MissionEditor: React.FC<MissionEditorProps> = ({
  missionId,
  onStart,
  onPause,
  onStop,
  onUpload
}) => {
  const {
    currentMission,
    waypoints,
    loading,
    createMission,
    updateMission,
    updateWaypoint,
    deleteWaypoint,
    reorderWaypoints
  } = useMission(missionId)

  const [editModalVisible, setEditModalVisible] = useState(false)
  const [editingWaypoint, setEditingWaypoint] = useState<Waypoint | null>(null)
  const [form] = Form.useForm()
  const [missionName, setMissionName] = useState(currentMission?.name || '新航线')

  const handleAddWaypoint = () => {
    message.info('请在地图上点击添加航点')
  }

  const handleEditWaypoint = (waypoint: Waypoint) => {
    setEditingWaypoint(waypoint)
    form.setFieldsValue({
      action: waypoint.action,
      altitude: waypoint.altitude,
      holdTime: waypoint.parameters.holdTime,
      speed: waypoint.parameters.speed,
      yaw: waypoint.parameters.yaw
    })
    setEditModalVisible(true)
  }

  const handleSaveWaypoint = async () => {
    try {
      const values = await form.validateFields()
      if (editingWaypoint) {
        const updated: Waypoint = {
          ...editingWaypoint,
          action: values.action,
          altitude: values.altitude,
          parameters: {
            ...editingWaypoint.parameters,
            holdTime: values.holdTime,
            speed: values.speed,
            yaw: values.yaw
          }
        }
        updateWaypoint(updated)
        message.success('航点更新成功')
        setEditModalVisible(false)
      }
    } catch (error) {
      console.error('Form validation failed:', error)
    }
  }

  const handleDeleteWaypoint = (waypointId: string) => {
    deleteWaypoint(waypointId)
    message.success('航点删除成功')
  }

  const handleSaveMission = () => {
    if (currentMission) {
      updateMission(currentMission.id, {
        name: missionName,
        waypoints
      })
    } else {
      createMission({
        name: missionName,
        waypoints
      })
    }
    message.success('航线保存成功')
  }

  const getActionColor = (action: WaypointAction): string => {
    const opt = actionOptions.find(o => o.value === action)
    return opt?.color || '#1890ff'
  }

  const columns = [
    {
      title: '序号',
      dataIndex: 'sequence',
      key: 'sequence',
      width: 60,
      render: (seq: number, _: unknown, index: number) => (
        <span style={{ fontWeight: 'bold', color: currentMission?.waypoints[index]?.isCurrent ? '#1890ff' : undefined }}>
          {seq}
        </span>
      )
    },
    {
      title: '类型',
      dataIndex: 'action',
      key: 'action',
      width: 100,
      render: (action: WaypointAction) => {
        const opt = actionOptions.find(o => o.value === action)
        return <WaypointActionTag $color={opt?.color || '#1890ff'}>{opt?.label || action}</WaypointActionTag>
      }
    },
    {
      title: '纬度',
      dataIndex: 'lat',
      key: 'lat',
      render: (val: number) => val.toFixed(6)
    },
    {
      title: '经度',
      dataIndex: 'lng',
      key: 'lng',
      render: (val: number) => val.toFixed(6)
    },
    {
      title: '高度(m)',
      dataIndex: 'altitude',
      key: 'altitude',
      width: 80
    },
    {
      title: '状态',
      key: 'status',
      width: 80,
      render: (_: unknown, record: Waypoint) => (
        <>
          {record.isCurrent && <span style={{ color: '#1890ff' }}>当前</span>}
          {record.isReached && <span style={{ color: '#52c41a' }}>已到达</span>}
        </>
      )
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: Waypoint) => (
        <div style={{ display: 'flex', gap: 8 }}>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => handleEditWaypoint(record)}
          />
          <Popconfirm
            title="确定删除此航点？"
            onConfirm={() => handleDeleteWaypoint(record.id)}
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
            />
          </Popconfirm>
        </div>
      )
    }
  ]

  const totalDistance = waypoints.reduce((acc, curr, index) => {
    if (index === 0) return 0
    const prev = waypoints[index - 1]
    return acc + Math.sqrt(
      Math.pow(curr.lat - prev.lat, 2) +
      Math.pow(curr.lng - prev.lng, 2)
    ) * 111000
  }, 0)

  return (
    <Container>
      <Card
        title={
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <Input
              value={missionName}
              onChange={e => setMissionName(e.target.value)}
              style={{ width: 200 }}
              placeholder="航线名称"
            />
            <span style={{ fontSize: 12, color: 'rgba(255,255,255,0.6)' }}>
              {waypoints.length} 个航点 | {totalDistance.toFixed(0)}m
            </span>
          </div>
        }
        size="small"
        extra={
          <Toolbar>
            <Button
              type="primary"
              icon={<PlusOutlined />}
              onClick={handleAddWaypoint}
              size="small"
            >
              添加航点
            </Button>
            <Button
              icon={<SaveOutlined />}
              onClick={handleSaveMission}
              size="small"
              loading={loading}
            >
              保存
            </Button>
            <Button
              icon={<UploadOutlined />}
              onClick={onUpload}
              size="small"
            >
              上传
            </Button>
            <Button
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={onStart}
              size="small"
              disabled={!currentMission || waypoints.length === 0}
            >
              开始
            </Button>
            <Button
              icon={<PauseCircleOutlined />}
              onClick={onPause}
              size="small"
            >
              暂停
            </Button>
            <Button
              danger
              icon={<StopOutlined />}
              onClick={onStop}
              size="small"
            >
              停止
            </Button>
          </Toolbar>
        }
      >
        <Table
          size="small"
          dataSource={waypoints}
          columns={columns}
          rowKey="id"
          pagination={false}
          scroll={{ y: 300 }}
          rowClassName={(record) => record.isCurrent ? 'current-waypoint' : ''}
        />
      </Card>

      <Modal
        title="编辑航点"
        open={editModalVisible}
        onOk={handleSaveWaypoint}
        onCancel={() => setEditModalVisible(false)}
      >
        <Form form={form} layout="vertical">
          <Form.Item
            name="action"
            label="动作类型"
            rules={[{ required: true, message: '请选择动作类型' }]}
          >
            <Select>
              {actionOptions.map(opt => (
                <Select.Option key={opt.value} value={opt.value}>
                  <span style={{ color: opt.color }}>{opt.label}</span>
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item
            name="altitude"
            label="高度 (m)"
            rules={[{ required: true, message: '请输入高度' }]}
          >
            <InputNumber min={0} max={500} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="holdTime"
            label="停留时间 (s)"
          >
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="speed"
            label="速度 (m/s)"
          >
            <InputNumber min={0} max={30} step={0.5} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item
            name="yaw"
            label="航向角 (°)"
          >
            <InputNumber min={0} max={360} style={{ width: '100%' }} />
          </Form.Item>
        </Form>
      </Modal>
    </Container>
  )
}

export default MissionEditor
