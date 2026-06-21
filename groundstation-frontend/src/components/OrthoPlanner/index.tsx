import React, { useState, useMemo, useCallback } from 'react'
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
  Divider,
  Table,
  message,
  Alert,
  Empty
} from 'antd'
import {
  AimOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  ExclamationCircleOutlined,
  CameraOutlined,
  DeleteOutlined,
  AreaChartOutlined,
  ScheduleOutlined,
  RouteOutlined
} from '@ant-design/icons'
import { useDispatch, useSelector } from 'react-redux'
import type { RootState } from '@/store'
import {
  setSelectedArea,
  updateOrthoMission,
  setCurrentOrtho,
  fetchOrthoMissions
} from '@/store/slices/payload'
import {
  createOrthoMission,
  planOrthoMission,
  startOrthoMission,
  pauseOrthoMission,
  resumeOrthoMission,
  abortOrthoMission
} from '@/api/payload'
import type { OrthoMission, OrthoStatus, OrthoWaypoint } from '@/types'

const { Option } = Select

interface OrthoPlannerProps {
  uavId: string
  cameraPayloadId?: string
  uavPosition?: { lat: number; lng: number }
}

const OrthoPlanner: React.FC<OrthoPlannerProps> = ({
  uavId,
  cameraPayloadId,
  uavPosition
}) => {
  const dispatch = useDispatch()
  const [form] = Form.useForm()

  const { currentOrtho, orthoMissions, selectedArea } = useSelector(
    (state: RootState) => state.payload
  )
  const orthoMissionsOfUAV = useMemo(
    () => orthoMissions.filter((o) => o.uavId === uavId),
    [orthoMissions, uavId]
  )

  const [isDrawingArea, setIsDrawingArea] = useState(false)
  const [loadingAction, setLoadingAction] = useState<string | null>(null)

  const statusColors: Record<OrthoStatus, string> = {
    pending: 'default',
    planning: 'processing',
    planned: 'blue',
    running: 'processing',
    paused: 'warning',
    completed: 'success',
    aborted: 'default',
    failed: 'error'
  }

  const areaInfo = useMemo(() => {
    if (!selectedArea || selectedArea.length < 3) return null
    return calculatePolygonArea(selectedArea)
  }, [selectedArea])

  const calculatedInfo = useMemo(() => {
    if (!currentOrtho) return null
    return {
      area: currentOrtho.totalAreaKm2?.toFixed(3) || '—',
      duration: currentOrtho.estimatedDurationSec
        ? `${Math.ceil(currentOrtho.estimatedDurationSec / 60)} 分钟`
        : '—',
      photos: currentOrtho.estimatedPhotos || 0,
      distance: currentOrtho.totalDistanceKm?.toFixed(2) || '—',
      waypoints: currentOrtho.waypointsCount || currentOrtho.waypoints?.length || 0,
      gsd: currentOrtho.gsD?.toFixed(1) || '—'
    }
  }, [currentOrtho])

  function calculatePolygonArea(coords: { lat: number; lng: number }[]) {
    if (coords.length < 3) return { area: 0, perimeter: 0 }
    let area = 0
    let perimeter = 0
    const R = 6378137
    const rad = (d: number) => (d * Math.PI) / 180

    for (let i = 0; i < coords.length; i++) {
      const p1 = coords[i]
      const p2 = coords[(i + 1) % coords.length]
      area +=
        (rad(p2.lng) - rad(p1.lng)) *
        (2 + Math.sin(rad(p1.lat)) + Math.sin(rad(p2.lat)))
      const dLat = rad(p2.lat - p1.lat)
      const dLng = rad(p2.lng - p1.lng)
      const a =
        Math.sin(dLat / 2) ** 2 +
        Math.cos(rad(p1.lat)) * Math.cos(rad(p2.lat)) * Math.sin(dLng / 2) ** 2
      perimeter += 2 * R * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a))
    }
    area = Math.abs((area * R * R) / 2)
    return {
      area: (area / 1_000_000).toFixed(3),
      perimeter: (perimeter / 1000).toFixed(2)
    }
  }

  const handleAreaPointClick = useCallback(
    (lat: number, lng: number) => {
      if (!isDrawingArea) return
      dispatch(
        setSelectedArea([
          ...(selectedArea || []),
          { lat, lng }
        ])
      )
    },
    [isDrawingArea, selectedArea, dispatch]
  )

  const startDrawArea = () => {
    setIsDrawingArea(true)
    dispatch(setSelectedArea([]))
    message.info('请在地图上依次点击选择测区多边形顶点，完成后点击"完成选点"')
  }

  const finishDrawArea = () => {
    if (!selectedArea || selectedArea.length < 3) {
      message.warning('至少需要3个点才能构成测区')
      return
    }
    setIsDrawingArea(false)
    message.success(`测区已选择，共 ${selectedArea.length} 个顶点`)
  }

  const clearArea = () => {
    dispatch(setSelectedArea(null))
    setIsDrawingArea(false)
  }

  const handleCreateMission = async (values: any) => {
    if (!selectedArea || selectedArea.length < 3) {
      message.error('请先在地图上框选测区范围')
      return
    }
    try {
      setLoadingAction('create')
      const mission = await createOrthoMission({
        uavId,
        name: values.name,
        altitude: values.altitude,
        speed: values.speed,
        overlapForward: values.overlapForward,
        overlapSide: values.overlapSide,
        areaCoordinates: selectedArea,
        cameraFocalLength: values.focalLength,
        sensorWidth: values.sensorWidth,
        sensorHeight: values.sensorHeight,
        imageWidth: values.imageWidth,
        imageHeight: values.imageHeight
      })

      setLoadingAction('plan')
      const planned = await planOrthoMission(mission.id, {
        uavId,
        name: values.name,
        areaCoordinates: selectedArea,
        altitude: values.altitude,
        overlapForward: values.overlapForward,
        overlapSide: values.overlapSide,
        speed: values.speed,
        cameraFocalLength: values.focalLength,
        sensorWidth: values.sensorWidth,
        sensorHeight: values.sensorHeight,
        imageWidth: values.imageWidth,
        imageHeight: values.imageHeight
      })

      dispatch(updateOrthoMission(planned))
      dispatch(setCurrentOrtho(planned))
      dispatch(fetchOrthoMissions() as any)
      message.success('正射任务创建并规划成功')
    } catch (err: any) {
      message.error(err.message || '创建任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleStartMission = async () => {
    if (!currentOrtho) return
    try {
      setLoadingAction('start')
      const mission = await startOrthoMission(currentOrtho.id)
      dispatch(updateOrthoMission(mission))
      message.success('正射采集任务已启动')
    } catch (err: any) {
      message.error(err.message || '启动任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handlePauseMission = async () => {
    if (!currentOrtho) return
    try {
      setLoadingAction('pause')
      const mission = await pauseOrthoMission(currentOrtho.id)
      dispatch(updateOrthoMission(mission))
      message.info('正射任务已暂停')
    } catch (err: any) {
      message.error(err.message || '暂停任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleResumeMission = async () => {
    if (!currentOrtho) return
    try {
      setLoadingAction('resume')
      const mission = await resumeOrthoMission(currentOrtho.id)
      dispatch(updateOrthoMission(mission))
      message.success('正射任务已恢复')
    } catch (err: any) {
      message.error(err.message || '恢复任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleAbortMission = async () => {
    if (!currentOrtho) return
    try {
      setLoadingAction('abort')
      const mission = await abortOrthoMission(currentOrtho.id)
      dispatch(updateOrthoMission(mission))
      message.warning('正射任务已中止')
    } catch (err: any) {
      message.error(err.message || '中止任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const selectMission = (mission: OrthoMission) => {
    dispatch(setCurrentOrtho(mission))
    dispatch(setSelectedArea(mission.areaCoordinates))
    form.setFieldsValue({
      name: mission.name,
      altitude: mission.altitude,
      speed: mission.speed,
      overlapForward: mission.overlapForward,
      overlapSide: mission.overlapSide,
      focalLength: mission.cameraFocalLength,
      sensorWidth: mission.sensorWidth,
      sensorHeight: mission.sensorHeight,
      imageWidth: mission.imageWidth,
      imageHeight: mission.imageHeight
    })
  }

  const waypointColumns = useMemo(
    () => [
      { title: '序号', dataIndex: 'sequence', width: 60 },
      {
        title: '纬度',
        dataIndex: 'lat',
        render: (v: number) => v.toFixed(7)
      },
      {
        title: '经度',
        dataIndex: 'lng',
        render: (v: number) => v.toFixed(7)
      },
      { title: '高度(m)', dataIndex: 'alt' },
      {
        title: '拍照点',
        dataIndex: 'capturePoint',
        render: (v: boolean) => (v ? <Tag color="green">是</Tag> : <Tag>否</Tag>)
      }
    ],
    []
  )

  return (
    <div className="ortho-planner h-full flex flex-col gap-4">
      <Card
        size="small"
        title={
          <Space>
            <RouteOutlined />
            <span>正射影像采集规划</span>
            {isDrawingArea && (
              <Tag color="blue" icon={<AimOutlined />}>正在框选测区 ({selectedArea?.length || 0} 个点)</Tag>
            )}
          </Space>
        }
        extra={
          <Select
            size="small"
            placeholder="选择已有任务"
            style={{ width: 180 }}
            allowClear
            value={currentOrtho?.id}
            onChange={(_, option: any) => {
              if (option?.mission) selectMission(option.mission)
            }}
          >
            {orthoMissionsOfUAV.map((mission) => (
              <Option key={mission.id} value={mission.id} mission={mission}>
                <Space size="small">
                  <Tag color={statusColors[mission.status as OrthoStatus]}>{mission.status}</Tag>
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
            altitude: 100,
            speed: 8,
            overlapForward: 80,
            overlapSide: 70,
            focalLength: 8.4,
            sensorWidth: 13.2,
            sensorHeight: 8.8,
            imageWidth: 4000,
            imageHeight: 3000
          }}
        >
          <Form.Item
            label="任务名称"
            name="name"
            rules={[{ required: true, message: '请输入任务名称' }]}
          >
            <Input placeholder="如：XX地块正射采集" />
          </Form.Item>

          <Space style={{ marginBottom: 12 }}>
            {!isDrawingArea ? (
              <Button
                size="small"
                type="primary"
                icon={<AreaChartOutlined />}
                onClick={startDrawArea}
              >
                开始框选测区
              </Button>
            ) : (
              <>
                <Button
                  size="small"
                  type="primary"
                  icon={<AreaChartOutlined />}
                  onClick={finishDrawArea}
                  disabled={!selectedArea || selectedArea.length < 3}
                >
                  完成选点
                </Button>
                <Button
                  size="small"
                  danger
                  icon={<DeleteOutlined />}
                  onClick={clearArea}
                >
                  取消
                </Button>
              </>
            )}
            {selectedArea && selectedArea.length > 0 && (
              <Tag color="green">
                测区: {selectedArea.length} 点
                {areaInfo && ` | ${areaInfo.area} km² | ${areaInfo.perimeter} km`}
              </Tag>
            )}
          </Space>

          {(!selectedArea || selectedArea.length < 3) && (
            <Alert
              style={{ marginBottom: 12 }}
              type="info"
              showIcon
              message="请先框选测区"
              description="点击"开始框选测区"按钮，然后在地图上依次点击多边形顶点，至少3个点构成闭合区域"
            />
          )}

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item label="飞行高度 (m)" name="altitude" rules={[{ required: true }]}>
                <Slider min={30} max={500} marks={{ 50: '50', 100: '100', 200: '200' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="飞行速度 (m/s)" name="speed" rules={[{ required: true }]}>
                <Slider min={1} max={20} step={0.5} marks={{ 5: '5', 10: '10', 15: '15' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item label="航向重叠度 (%)" name="overlapForward" rules={[{ required: true }]}>
                <Slider min={50} max={90} marks={{ 60: '60', 70: '70', 80: '80' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="旁向重叠度 (%)" name="overlapSide" rules={[{ required: true }]}>
                <Slider min={40} max={85} marks={{ 50: '50', 60: '60', 70: '70' }} />
              </Form.Item>
            </Col>
          </Row>

          <Divider orientation="left" plain style={{ margin: '8px 0' }}>
            <CameraOutlined /> 相机参数
          </Divider>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item
                label={<Tooltip title="镜头焦距 (mm)">焦距 (mm)</Tooltip>}
                name="focalLength"
              >
                <InputNumber step={0.1} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item label="传感器宽" name="sensorWidth">
                <InputNumber step={0.1} addonAfter="mm" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={6}>
              <Form.Item label="传感器高" name="sensorHeight">
                <InputNumber step={0.1} addonAfter="mm" style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Row gutter={8}>
            <Col span={12}>
              <Form.Item label="图像宽 (px)" name="imageWidth">
                <InputNumber step={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item label="图像高 (px)" name="imageHeight">
                <InputNumber step={100} style={{ width: '100%' }} />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item>
            <Space>
              <Button
                type="primary"
                htmlType="submit"
                icon={<ScheduleOutlined />}
                loading={loadingAction === 'create' || loadingAction === 'plan'}
                disabled={!selectedArea || selectedArea.length < 3}
              >
                {currentOrtho ? '重新规划' : '创建并规划航线'}
              </Button>
              <Button
                icon={<DeleteOutlined />}
                danger
                onClick={() => {
                  dispatch(setCurrentOrtho(null))
                  dispatch(setSelectedArea(null))
                  form.resetFields()
                }}
              >
                清除
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Card>

      {calculatedInfo && (
        <Card size="small" title={<Space><AreaChartOutlined />规划结果</Space>}>
          <Row gutter={16}>
            <Col span={8}><Statistic title="测区面积" value={calculatedInfo.area} suffix="km²" /></Col>
            <Col span={8}><Statistic title="预计耗时" value={calculatedInfo.duration} /></Col>
            <Col span={8}><Statistic title="预计照片" value={calculatedInfo.photos} suffix="张" /></Col>
            <Col span={8}><Statistic title="飞行距离" value={calculatedInfo.distance} suffix="km" /></Col>
            <Col span={8}><Statistic title="航点数" value={calculatedInfo.waypoints} /></Col>
            <Col span={8}><Statistic title="地面分辨率" value={calculatedInfo.gsd} suffix="cm/px" /></Col>
          </Row>
        </Card>
      )}

      {currentOrtho && (
        <>
          <Card size="small" title="任务进度">
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="任务状态"
                  valueRender={() => (
                    <Tag color={statusColors[currentOrtho.status as OrthoStatus]}>
                      {currentOrtho.status}
                    </Tag>
                  )}
                />
              </Col>
              <Col span={8}>
                <Statistic
                  title="当前航点"
                  value={`${currentOrtho.currentWaypointIndex || 0}/${currentOrtho.waypointsCount || currentOrtho.waypoints?.length || 0}`}
                />
              </Col>
              <Col span={8}>
                <Statistic title="已拍照片" value={currentOrtho.photosCaptured || 0} suffix="张" />
              </Col>
            </Row>
            <div style={{ marginTop: 12 }}>
              <Progress percent={Math.floor((currentOrtho.progress || 0) * 100)} size="small" />
            </div>
          </Card>

          {currentOrtho.waypoints && currentOrtho.waypoints.length > 0 && (
            <Card
              size="small"
              title="航线航点列表"
              bodyStyle={{ padding: 0, maxHeight: 260, overflowY: 'auto' }}
            >
              <Table<OrthoWaypoint>
                size="small"
                dataSource={currentOrtho.waypoints}
                columns={waypointColumns}
                rowKey="id"
                pagination={false}
                locale={{ emptyText: <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无航点数据" /> }}
              />
            </Card>
          )}

          <Card
            size="small"
            title={<Space><PlayCircleOutlined />任务控制</Space>}
          >
            <Space wrap>
              {(currentOrtho.status === 'pending' || currentOrtho.status === 'planned' ||
                currentOrtho.status === 'completed') && (
                <Button
                  type="primary"
                  icon={<PlayCircleOutlined />}
                  onClick={handleStartMission}
                  loading={loadingAction === 'start'}
                >
                  开始采集
                </Button>
              )}
              {currentOrtho.status === 'running' && (
                <Button
                  icon={<PauseCircleOutlined />}
                  onClick={handlePauseMission}
                  loading={loadingAction === 'pause'}
                >
                  暂停
                </Button>
              )}
              {currentOrtho.status === 'paused' && (
                <Button
                  type="primary"
                  icon={<ReloadOutlined />}
                  onClick={handleResumeMission}
                  loading={loadingAction === 'resume'}
                >
                  恢复
                </Button>
              )}
              {(currentOrtho.status === 'running' || currentOrtho.status === 'paused') && (
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

export default OrthoPlanner
