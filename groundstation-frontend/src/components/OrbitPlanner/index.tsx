import React, { useState, useCallback, useMemo, useRef } from 'react'
import {
  Card,
  Form,
  Input,
  InputNumber,
  Button,
  Select,
  Space,
  Slider,
  Row,
  Col,
  Statistic,
  Tag,
  Progress,
  Tooltip,
  Switch,
  Divider,
  message
} from 'antd'
import {
  EnvironmentOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  ExclamationCircleOutlined,
  CameraOutlined,
  DeleteOutlined
} from '@ant-design/icons'
import { useDispatch, useSelector } from 'react-redux'
import type { RootState } from '@/store'
import {
  setOrbitCenter,
  updateOrbitMission,
  setCurrentOrbit,
  fetchOrbitMissions
} from '@/store/slices/payload'
import {
  createOrbitMission,
  startOrbitMission,
  pauseOrbitMission,
  resumeOrbitMission,
  abortOrbitMission
} from '@/api/payload'
import type { OrbitMission, OrbitStatus } from '@/types'

const { Option } = Select

interface OrbitPlannerProps {
  uavId: string
  cameraPayloadId?: string
  onMapClick?: (callback: (lat: number, lng: number) => void) => void
  mapInstance?: any
  uavPosition?: { lat: number; lng: number }
}

const OrbitPlanner: React.FC<OrbitPlannerProps> = ({
  uavId,
  cameraPayloadId,
  uavPosition
}) => {
  const dispatch = useDispatch()
  const [form] = Form.useForm()
  const mapClickHandlerRef = useRef<((lat: number, lng: number) => void) | null>(null)

  const { currentOrbit, orbitMissions, orbitCenter } = useSelector(
    (state: RootState) => state.payload
  )
  const orbitMissionsOfUAV = useMemo(
    () => orbitMissions.filter((o) => o.uavId === uavId),
    [orbitMissions, uavId]
  )

  const [isPickingCenter, setIsPickingCenter] = useState(false)
  const [loadingAction, setLoadingAction] = useState<string | null>(null)
  const [isAutoCapture, setIsAutoCapture] = useState(true)

  const statusColors: Record<OrbitStatus, string> = {
    pending: 'default',
    running: 'processing',
    paused: 'warning',
    completed: 'success',
    aborted: 'default',
    failed: 'error'
  }

  const handleMapClickForCenter = useCallback(
    (lat: number, lng: number) => {
      if (isPickingCenter) {
        dispatch(setOrbitCenter({ lat, lng }))
        form.setFieldsValue({
          centerLat: Number(lat.toFixed(7)),
          centerLng: Number(lng.toFixed(7))
        })
        setIsPickingCenter(false)
      }
    },
    [isPickingCenter, dispatch, form]
  )

  mapClickHandlerRef.current = handleMapClickForCenter

  const startPickCenter = () => {
    setIsPickingCenter(true)
    message.info('请在地图上点击选择环绕中心点')
  }

  const useCurrentUAVPosition = () => {
    if (uavPosition) {
      dispatch(setOrbitCenter(uavPosition))
      form.setFieldsValue({
        centerLat: Number(uavPosition.lat.toFixed(7)),
        centerLng: Number(uavPosition.lng.toFixed(7))
      })
      message.success('已使用无人机当前位置')
    } else {
      message.warning('无人机位置暂不可用')
    }
  }

  const handleCreateMission = async (values: any) => {
    try {
      setLoadingAction('create')
      const mission = await createOrbitMission({
        uavId,
        name: values.name,
        centerLat: values.centerLat,
        centerLng: values.centerLng,
        altitude: values.altitude,
        radius: values.radius,
        loops: values.loops,
        direction: values.direction,
        velocity: values.velocity,
        autoCapture: isAutoCapture,
        captureInterval: values.captureInterval,
        gimbalPitch: values.gimbalPitch,
        payloadId: cameraPayloadId
      })
      dispatch(updateOrbitMission(mission))
      dispatch(setCurrentOrbit(mission))
      dispatch(fetchOrbitMissions() as any)
      message.success('环绕任务创建成功')
    } catch (err: any) {
      message.error(err.message || '创建任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleStartMission = async () => {
    if (!currentOrbit) return
    try {
      setLoadingAction('start')
      const mission = await startOrbitMission(currentOrbit.id)
      dispatch(updateOrbitMission(mission))
      message.success('环绕任务已启动')
    } catch (err: any) {
      message.error(err.message || '启动任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handlePauseMission = async () => {
    if (!currentOrbit) return
    try {
      setLoadingAction('pause')
      const mission = await pauseOrbitMission(currentOrbit.id)
      dispatch(updateOrbitMission(mission))
      message.info('环绕任务已暂停')
    } catch (err: any) {
      message.error(err.message || '暂停任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleResumeMission = async () => {
    if (!currentOrbit) return
    try {
      setLoadingAction('resume')
      const mission = await resumeOrbitMission(currentOrbit.id)
      dispatch(updateOrbitMission(mission))
      message.success('环绕任务已恢复')
    } catch (err: any) {
      message.error(err.message || '恢复任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleAbortMission = async () => {
    if (!currentOrbit) return
    try {
      setLoadingAction('abort')
      const mission = await abortOrbitMission(currentOrbit.id)
      dispatch(updateOrbitMission(mission))
      message.warning('环绕任务已中止')
    } catch (err: any) {
      message.error(err.message || '中止任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const selectMission = (mission: OrbitMission) => {
    dispatch(setCurrentOrbit(mission))
    dispatch(setOrbitCenter({ lat: mission.centerLat, lng: mission.centerLng }))
    form.setFieldsValue({
      name: mission.name,
      centerLat: mission.centerLat,
      centerLng: mission.centerLng,
      altitude: mission.altitude,
      radius: mission.radius,
      loops: mission.loops,
      direction: mission.direction,
      velocity: mission.velocity,
      captureInterval: mission.captureInterval,
      gimbalPitch: mission.gimbalPitch || -90
    })
    setIsAutoCapture(mission.autoCapture)
  }

  const estimatedTime = useMemo(() => {
    if (!currentOrbit) return '—'
    const circumference = 2 * Math.PI * currentOrbit.radius
    const totalDistance = circumference * currentOrbit.loops
    const timeMin = Math.ceil(totalDistance / (currentOrbit.velocity * 60))
    return `${timeMin} 分钟`
  }, [currentOrbit])

  return (
    <div className="orbit-planner h-full flex flex-col gap-4">
      <Card
        size="small"
        title={
          <Space>
            <EnvironmentOutlined />
            <span>兴趣点环绕规划</span>
            {isPickingCenter && <Tag color="blue" icon={<EnvironmentOutlined />}>正在选点...</Tag>}
          </Space>
        }
        extra={
          <Select
            size="small"
            placeholder="选择已有任务"
            style={{ width: 180 }}
            allowClear
            value={currentOrbit?.id}
            onChange={(_, option: any) => {
              if (option?.mission) selectMission(option.mission)
            }}
          >
            {orbitMissionsOfUAV.map((mission) => (
              <Option key={mission.id} value={mission.id} mission={mission}>
                <Space size="small">
                  <Tag color={statusColors[mission.status as OrbitStatus]}>{mission.status}</Tag>
                  <span>{mission.name}</span>
                </Space>
              </Option>
            ))}
          </Select>
        }
      >
        <Form
          form={form}
          layout="vertical"
          size="small"
          onFinish={handleCreateMission}
          initialValues={{
            altitude: 50,
            radius: 30,
            loops: 1,
            direction: 1,
            velocity: 5,
            captureInterval: 5,
            gimbalPitch: -90
          }}
        >
          <Form.Item
            label="任务名称"
            name="name"
            rules={[{ required: true, message: '请输入任务名称' }]}
          >
            <Input placeholder="如：建筑环绕拍摄" />
          </Form.Item>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item
                label="中心纬度"
                name="centerLat"
                rules={[{ required: true, message: '必填' }]}
              >
                <InputNumber
                  step={0.0000001}
                  precision={7}
                  style={{ width: '100%' }}
                  placeholder="纬度"
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label="中心经度"
                name="centerLng"
                rules={[{ required: true, message: '必填' }]}
              >
                <InputNumber
                  step={0.0000001}
                  precision={7}
                  style={{ width: '100%' }}
                  placeholder="经度"
                />
              </Form.Item>
            </Col>
          </Row>

          <Space style={{ marginBottom: 16, marginTop: -8 }}>
            <Button
              size="small"
              type={isPickingCenter ? 'primary' : 'default'}
              icon={<EnvironmentOutlined />}
              onClick={startPickCenter}
              danger={isPickingCenter}
            >
              {isPickingCenter ? '取消选点' : '地图选点'}
            </Button>
            <Button
              size="small"
              icon={<EnvironmentOutlined />}
              onClick={useCurrentUAVPosition}
              disabled={!uavPosition}
            >
              使用无人机位置
            </Button>
            {orbitCenter && (
              <Tag color="green">
                已选点: {orbitCenter.lat.toFixed(6)}, {orbitCenter.lng.toFixed(6)}
              </Tag>
            )}
          </Space>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item label="飞行高度 (m)" name="altitude" rules={[{ required: true }]}>
                <Slider min={10} max={500} marks={{ 50: '50', 100: '100', 200: '200' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="环绕半径 (m)" name="radius" rules={[{ required: true }]}>
                <Slider min={5} max={200} marks={{ 20: '20', 50: '50', 100: '100' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={8}>
            <Col span={8}>
              <Form.Item label="环绕圈数" name="loops" rules={[{ required: true }]}>
                <InputNumber min={1} max={10} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="飞行速度 (m/s)" name="velocity" rules={[{ required: true }]}>
                <InputNumber min={1} max={20} step={0.5} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item label="环绕方向" name="direction" rules={[{ required: true }]}>
                <Select>
                  <Option value={1}>顺时针</Option>
                  <Option value={-1}>逆时针</Option>
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left" plain style={{ margin: '8px 0' }}>
            <CameraOutlined /> 拍摄设置
          </Divider>

          <Space style={{ marginBottom: 12 }} align="center">
            <span>自动拍照:</span>
            <Switch checked={isAutoCapture} onChange={setIsAutoCapture} />
            <Tag color={isAutoCapture ? 'green' : 'default'}>
              {isAutoCapture ? '已开启' : '已关闭'}
            </Tag>
          </Space>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item
                label={
                  <Tooltip title="每隔多少秒拍摄一次">拍照间隔 (秒)</Tooltip>
                }
                name="captureInterval"
              >
                <InputNumber min={1} max={60} disabled={!isAutoCapture} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                label={<Tooltip title="云台俯仰角，-90为正向下">云台俯仰 (°)</Tooltip>}
                name="gimbalPitch"
              >
                <InputNumber
                  min={-90}
                  max={30}
                  step={5}
                  disabled={!isAutoCapture}
                  style={{ width: '100%' }}
                />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit" loading={loadingAction === 'create'}>
                {currentOrbit ? '更新任务' : '创建任务'}
              </Button>
              <Button
                icon={<DeleteOutlined />}
                danger
                onClick={() => {
                  dispatch(setCurrentOrbit(null))
                  dispatch(setOrbitCenter(null))
                  form.resetFields()
                }}
              >
                清除
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {currentOrbit && (
        <>
          <Card size="small" title="任务信息" className="flex-shrink-0">
            <Row gutter={16}>
              <Col span={12}>
                <Statistic title="任务状态" valueRender={() => (
                  <Tag color={statusColors[currentOrbit.status as OrbitStatus]}>
                    {currentOrbit.status}
                  </Tag>
                )} />
              </Col>
              <Col span={12}>
                <Statistic title="预计耗时" value={estimatedTime} />
              </Col>
              <Col span={12}>
                <Statistic title="当前圈数" value={`${currentOrbit.currentLoop || 0}/${currentOrbit.loops}`} />
              </Col>
              <Col span={12}>
                <Statistic title="已拍照片" value={currentOrbit.photosCaptured || 0} suffix="张" />
              </Col>
            </Row>
            <div style={{ marginTop: 12 }}>
              <Progress
                percent={Math.floor((currentOrbit?.progress || 0) * 100)}
                size="small"
              />
            </div>
          </Card>

          <Card
            size="small"
            title={
              <Space>
                <PlayCircleOutlined />
                <span>任务控制</span>
              </Space>
            }
          >
            <Space wrap>
              {currentOrbit.status === 'pending' || currentOrbit.status === 'completed' ? (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={handleStartMission}
                  loading={loadingAction === 'start'}
                >
                  开始执行
                </Button>
              ) : null}
              {currentOrbit.status === 'running' && (
                <Button
                  icon={<PauseCircleOutlined />}
                  onClick={handlePauseMission}
                  loading={loadingAction === 'pause'}
                >
                  暂停
                </Button>
              )}
              {currentOrbit.status === 'paused' && (
                <Button
                  type="primary"
                  icon={<ReloadOutlined />}
                  onClick={handleResumeMission}
                  loading={loadingAction === 'resume'}
                >
                  恢复
                </Button>
              )}
              {(currentOrbit.status === 'running' || currentOrbit.status === 'paused') && (
                <Tooltip title="中止任务并返回原点">
                  <Button
                    danger
                    icon={<ExclamationCircleOutlined />}
                    onClick={handleAbortMission}
                    loading={loadingAction === 'abort'}
                  >
                    中止
                  </Button>
                </Tooltip>
              )}
            </Space>
          </Card>
        </>
      )}
    </div>
  )
}

export default OrbitPlanner
