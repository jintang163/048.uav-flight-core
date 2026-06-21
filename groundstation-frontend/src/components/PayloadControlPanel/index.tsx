import React, { useState, useEffect, useMemo, useRef } from 'react'
import styled from 'styled-components'
import {
  Card,
  Tabs,
  Form,
  Input,
  Button,
  Select,
  Switch,
  Slider,
  Space,
  Row,
  Col,
  Statistic,
  Tag,
  Progress,
  List,
  Avatar,
  Tooltip,
  Divider,
  InputNumber,
  Modal,
  message,
  Badge,
  Alert
} from 'antd'
import {
  CameraOutlined,
  VideoCameraOutlined,
  PlayCircleOutlined,
  StopOutlined,
  RotateLeftOutlined,
  SoundOutlined,
  FileTextOutlined,
  SendOutlined,
  DeleteOutlined,
  DownloadOutlined,
  ReloadOutlined,
  BulbOutlined,
  CloudUploadOutlined,
  SettingOutlined,
  EnvironmentOutlined,
  ExpandOutlined,
  ZoomInOutlined,
  ZoomOutOutlined,
  ThermometerOutlined,
  DatabaseOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import { addTTSTask, updateTTSTask } from '@/store/slices/payload'
import * as payloadApi from '@/api/payload'
import type { CameraMode, PayloadDevice, SprayerStatus, CameraStatus, TextToSpeechTask, SpeakerAudio } from '@/types'
import type { PayloadType } from '@/types'

const Container = styled.div`
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 4px;
  overflow-y: auto;
  height: 100%;
  padding-right: 8px;

  &::-webkit-scrollbar {
    width: 6px;
  }
  &::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.03);
    border-radius: 3px;
  }
  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.15);
    border-radius: 3px;
    &:hover {
      background: rgba(255, 255, 255, 0.25);
    }
  }
`

const PanelCard = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-head {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    min-height: 40px;
  }

  .ant-card-body {
    padding: 12px;
  }

  .ant-statistic-content {
    color: rgba(255, 255, 255, 0.95);
  }

  .ant-statistic-title {
    color: rgba(255, 255, 255, 0.6);
    margin-bottom: 2px;
  }
`

const DeviceSelector = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  background: rgba(255, 255, 255, 0.04);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  margin-bottom: 12px;
`

const StatusRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);

  &:last-child {
    border-bottom: none;
  }
`

const StatusLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
`

const StatusValue = styled.span`
  color: rgba(255, 255, 255, 0.95);
  font-weight: 500;
  font-family: 'Courier New', monospace;
  font-size: 13px;
`

const ControlRow = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  gap: 12px;
`

const ControlLabel = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.85);
  font-size: 13px;
`

const AudioItem = styled(List.Item)`
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  padding: 8px 12px !important;
  margin-bottom: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
`

interface PayloadControlPanelProps {
  uavId?: string
}

const PayloadControlPanel: React.FC<PayloadControlPanelProps> = ({ uavId }) => {
  const dispatch = useAppDispatch()
  const { payloads, cameraStatusMap, sprayerStatusMap, ttsTasks, speakerAudios } = useAppSelector(state => state.payload)
  const [cameraForm] = Form.useForm()
  const [sprayerForm] = Form.useForm()
  const [ttsForm] = Form.useForm()
  const audioRef = useRef<HTMLAudioElement>(null)
  const [playingAudio, setPlayingAudio] = useState<string | null>(null)
  const [loadingAction, setLoadingAction] = useState<string | null>(null)

  const cameras = useMemo(
    () => payloads.filter(p => p.type === 'camera' || p.type === 'thermal_camera'),
    [payloads]
  )
  const sprayers = useMemo(
    () => payloads.filter(p => p.type === 'sprayer'),
    [payloads]
  )
  const speakers = useMemo(
    () => payloads.filter(p => p.type === 'speaker'),
    [payloads]
  )

  const [selectedCameraId, setSelectedCameraId] = useState<string | undefined>()
  const [selectedSprayerId, setSelectedSprayerId] = useState<string | undefined>()
  const [selectedSpeakerId, setSelectedSpeakerId] = useState<string | undefined>()

  useEffect(() => {
    if (cameras.length > 0 && !selectedCameraId) setSelectedCameraId(cameras[0].id)
  }, [cameras, selectedCameraId])

  useEffect(() => {
    if (sprayers.length > 0 && !selectedSprayerId) setSelectedSprayerId(sprayers[0].id)
  }, [sprayers, selectedSprayerId])

  useEffect(() => {
    if (speakers.length > 0 && !selectedSpeakerId) setSelectedSpeakerId(speakers[0].id)
  }, [speakers, selectedSpeakerId])

  const currentCamera = cameras.find(c => c.id === selectedCameraId)
  const currentCameraStatus: CameraStatus | undefined = selectedCameraId ? cameraStatusMap[selectedCameraId] : undefined
  const currentSprayer = sprayers.find(s => s.id === selectedSprayerId)
  const currentSprayerStatus: SprayerStatus | undefined = selectedSprayerId ? sprayerStatusMap[selectedSprayerId] : undefined
  const currentSpeaker = speakers.find(s => s.id === selectedSpeakerId)

  const handleTakePhoto = async () => {
    if (!selectedCameraId || !uavId) {
      message.warning('请先选择无人机和相机')
      return
    }
    try {
      setLoadingAction('photo')
      await payloadApi.cameraTakePhoto(selectedCameraId, { count: 1 })
      message.success('拍照指令已发送')
    } catch (e) {
      message.error('拍照失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleToggleRecording = async () => {
    if (!selectedCameraId || !uavId) {
      message.warning('请先选择无人机和相机')
      return
    }
    try {
      setLoadingAction('record')
      const isRecording = currentCameraStatus?.isRecording
      if (isRecording) {
        await payloadApi.cameraStopRecording(selectedCameraId)
        message.success('停止录制')
      } else {
        await payloadApi.cameraStartRecording(selectedCameraId)
        message.success('开始录制')
      }
    } catch (e) {
      message.error('操作失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleSetMode = async (mode: CameraMode) => {
    if (!selectedCameraId) return
    try {
      await payloadApi.setCameraMode(selectedCameraId, mode)
      message.success(`切换到${mode === 'photo' ? '拍照' : mode === 'video' ? '录像' : '回放'}模式`)
    } catch (e) {
      message.error('模式切换失败')
    }
  }

  const handleZoom = async (direction: 'in' | 'out') => {
    if (!selectedCameraId) return
    try {
      await payloadApi.setCameraZoom(selectedCameraId, {
        zoomType: direction,
        zoomSpeed: 50
      })
    } catch (e) {
      message.error('变焦失败')
    }
  }

  const handleToggleSprayer = async (on: boolean) => {
    if (!selectedSprayerId || !uavId) {
      message.warning('请先选择无人机和喷药器')
      return
    }
    try {
      setLoadingAction('sprayer')
      if (on) {
        await payloadApi.sprayerStart(selectedSprayerId)
        message.success('开始喷药')
      } else {
        await payloadApi.sprayerStop(selectedSprayerId)
        message.success('停止喷药')
      }
    } catch (e) {
      message.error('操作失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handleSetFlowRate = async (value: number) => {
    if (!selectedSprayerId) return
    try {
      await payloadApi.setSprayerFlowRate(selectedSprayerId, { flowRate: value })
    } catch (e) {
      message.error('设置流量失败')
    }
  }

  const handleCreateTTS = async () => {
    if (!selectedSpeakerId || !uavId) {
      message.warning('请先选择无人机和喊话器')
      return
    }
    try {
      const values = await ttsForm.validateFields()
      setLoadingAction('tts')
      const res = await payloadApi.createTextToSpeechTask({
        uavId,
        speakerPayloadId: selectedSpeakerId,
        text: values.text,
        voice: values.voice || 'xiaoxiao',
        volume: values.volume || 80,
        speed: values.speed || 1.0,
        pitch: values.pitch || 1.0,
        autoPlay: values.autoPlay !== false
      })
      dispatch(addTTSTask(res))
      message.success('TTS任务已创建，正在合成...')
      ttsForm.resetFields()
    } catch (e: any) {
      message.error(e?.message || '创建TTS任务失败')
    } finally {
      setLoadingAction(null)
    }
  }

  const handlePlayAudio = (audio: SpeakerAudio) => {
    if (!audio.fileURL) {
      message.warning('音频文件未就绪')
      return
    }
    if (audioRef.current) {
      audioRef.current.src = audio.fileURL
      audioRef.current.play().then(() => {
        setPlayingAudio(audio.id)
      }).catch(() => {
        message.error('音频播放失败')
      })
    }
  }

  const handleBroadcastAudio = async (audioId: string) => {
    if (!selectedSpeakerId) return
    try {
      await payloadApi.playSpeakerAudio(selectedSpeakerId, { audioId, loop: false })
      message.success('已发送广播指令')
    } catch (e) {
      message.error('广播失败')
    }
  }

  const formatDuration = (sec?: number): string => {
    if (!sec) return '—'
    const m = Math.floor(sec / 60)
    const s = Math.floor(sec % 60)
    return `${m}:${s.toString().padStart(2, '0')}`
  }

  const formatFileSize = (bytes?: number): string => {
    if (!bytes) return '—'
    if (bytes < 1024) return `${bytes} B`
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
    return `${(bytes / 1024 / 1024).toFixed(2)} MB`
  }

  const getPayloadTypeLabel = (type: PayloadType): string => {
    const map: Record<PayloadType, string> = {
      camera: '可见光相机',
      thermal_camera: '热成像相机',
      sprayer: '喷药器',
      speaker: '喊话器',
      gripper: '机械爪',
      sensor: '传感器',
      lidar: '激光雷达',
      parachute: '降落伞',
      other: '通用载荷'
    }
    return map[type] || type
  }

  const deviceSelectStyle = { flex: 1 } as React.CSSProperties

  const cameraTab = (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <DeviceSelector>
        <CameraOutlined style={{ color: '#1890ff' }} />
        <Select
          style={deviceSelectStyle}
          size="small"
          placeholder="选择相机"
          value={selectedCameraId}
          onChange={setSelectedCameraId}
          options={cameras.map(c => ({
            value: c.id,
            label: (
              <Space>
                <Badge status={(c.status === 'online' || c.status === 'active') ? 'success' : 'default'} />
                <span>{c.name || getPayloadTypeLabel(c.type)}</span>
              </Space>
            )
          }))}
        />
        {currentCamera && (
          <Tag color={(currentCamera.status === 'online' || currentCamera.status === 'active') ? 'success' : 'default'}>
            {currentCamera.status === 'online' || currentCamera.status === 'active' ? '在线' : '离线'}
          </Tag>
        )}
      </DeviceSelector>

      {currentCameraStatus && (
        <PanelCard size="small" title="相机状态">
          <Row gutter={[12, 12]}>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <DatabaseOutlined />
                    存储卡剩余
                  </span>
                }
                value={currentCameraStatus.storagePercent ?? 0}
                suffix="%"
                valueStyle={{ color: (currentCameraStatus.storagePercent ?? 0) > 15 ? '#52c41a' : '#ff4d4f' }}
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <ThermometerOutlined />
                    镜头温度
                  </span>
                }
                value={currentCameraStatus.lensTempC ?? 0}
                suffix="°C"
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <CameraOutlined />
                    已拍照片
                  </span>
                }
                value={currentCameraStatus.photoCount ?? 0}
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <VideoCameraOutlined />
                    录制状态
                  </span>
                }
                valueRender={() => currentCameraStatus.isRecording ? (
                  <Tag color="error" icon={<span style={{ width: 6, height: 6, background: '#ff4d4f', borderRadius: '50%', display: 'inline-block', animation: 'pulse 1s infinite' }} />}>录制中</Tag>
                ) : (
                  <Tag color="default">空闲</Tag>
                )}
              />
            </Col>
          </Row>

          {currentCameraStatus.recordTimeSec !== undefined && currentCameraStatus.recordTimeSec > 0 && (
            <>
              <Divider style={{ margin: '12px 0' }} />
              <StatusRow>
                <StatusLabel>当前录制时长</StatusLabel>
                <StatusValue>{formatDuration(currentCameraStatus.recordTimeSec)}</StatusValue>
              </StatusRow>
            </>
          )}
          {currentCameraStatus.zoomLevel !== undefined && (
            <StatusRow>
              <StatusLabel>光学变焦</StatusLabel>
              <StatusValue>{currentCameraStatus.zoomLevel.toFixed(1)}x</StatusValue>
            </StatusRow>
          )}
        </PanelCard>
      )}

      <PanelCard size="small" title="拍摄控制">
        <Row gutter={[8, 12]} style={{ marginBottom: 12 }}>
          <Col span={24}>
            <ControlRow>
              <ControlLabel>
                <SettingOutlined />
                拍摄模式
              </ControlLabel>
              <Select
                size="small"
                style={{ width: 120 }}
                value={currentCameraStatus?.mode || 'photo'}
                onChange={handleSetMode}
                options={[
                  { value: 'photo', label: '拍照模式' },
                  { value: 'video', label: '录像模式' },
                  { value: 'playback', label: '回放模式' }
                ]}
              />
            </ControlRow>
          </Col>
        </Row>

        <Space style={{ width: '100%' }} direction="vertical" size={8}>
          <Row gutter={8}>
            <Col span={12}>
              <Button
                block
                type="primary"
                icon={<CameraOutlined />}
                onClick={handleTakePhoto}
                loading={loadingAction === 'photo'}
                disabled={!currentCamera || currentCamera.status !== 'online' && currentCamera.status !== 'active'}
              >
                拍照
              </Button>
            </Col>
            <Col span={12}>
              <Button
                block
                danger={currentCameraStatus?.isRecording}
                icon={currentCameraStatus?.isRecording ? <StopOutlined /> : <VideoCameraOutlined />}
                onClick={handleToggleRecording}
                loading={loadingAction === 'record'}
                disabled={!currentCamera || currentCamera.status !== 'online' && currentCamera.status !== 'active'}
                type={currentCameraStatus?.isRecording ? undefined : 'primary'}
              >
                {currentCameraStatus?.isRecording ? '停止录像' : '开始录像'}
              </Button>
            </Col>
          </Row>

          <Row gutter={8}>
            <Col span={12}>
              <Button
                block
                icon={<ZoomOutOutlined />}
                onClick={() => handleZoom('out')}
                disabled={!currentCamera || currentCamera.status !== 'online' && currentCamera.status !== 'active'}
              >
                缩小
              </Button>
            </Col>
            <Col span={12}>
              <Button
                block
                icon={<ZoomInOutlined />}
                onClick={() => handleZoom('in')}
                disabled={!currentCamera || currentCamera.status !== 'online' && currentCamera.status !== 'active'}
              >
                放大
              </Button>
            </Col>
          </Row>
        </Space>
      </PanelCard>
    </Space>
  )

  const sprayerTab = (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <DeviceSelector>
        <RotateLeftOutlined style={{ color: '#52c41a' }} />
        <Select
          style={deviceSelectStyle}
          size="small"
          placeholder="选择喷药器"
          value={selectedSprayerId}
          onChange={setSelectedSprayerId}
          options={sprayers.map(s => ({
            value: s.id,
            label: (
              <Space>
                <Badge status={(s.status === 'online' || s.status === 'active') ? 'success' : 'default'} />
                <span>{s.name || getPayloadTypeLabel(s.type)}</span>
              </Space>
            )
          }))}
        />
        {currentSprayer && (
          <Tag color={(currentSprayer.status === 'online' || currentSprayer.status === 'active') ? 'success' : 'default'}>
            {currentSprayer.status === 'online' || currentSprayer.status === 'active' ? '在线' : '离线'}
          </Tag>
        )}
      </DeviceSelector>

      {currentSprayerStatus && (
        <PanelCard size="small" title="喷药状态">
          <Row gutter={[12, 12]}>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <DatabaseOutlined />
                    药液剩余
                  </span>
                }
                value={currentSprayerStatus.tankLevelPercent ?? 0}
                suffix="%"
                valueStyle={{ color: (currentSprayerStatus.tankLevelPercent ?? 0) > 10 ? '#52c41a' : '#ff4d4f' }}
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <BulbOutlined />
                    泵压
                  </span>
                }
                value={currentSprayerStatus.pressureBar ?? 0}
                suffix="bar"
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <ReloadOutlined />
                    流量
                  </span>
                }
                value={currentSprayerStatus.flowRateLpm ?? 0}
                suffix="L/min"
              />
            </Col>
            <Col span={12}>
              <Statistic
                title={
                  <span style={{ display: 'flex', alignItems: 'center', gap: 4 }}>
                    <CloudUploadOutlined />
                    喷幅
                  </span>
                }
                value={currentSprayerStatus.sprayWidthM ?? 0}
                suffix="m"
              />
            </Col>
          </Row>

          <Divider style={{ margin: '12px 0' }} />

          <StatusRow>
            <StatusLabel>喷药状态</StatusLabel>
            <StatusValue>
              {currentSprayerStatus.isSpraying ? (
                <Tag color="processing" icon={<span style={{ width: 6, height: 6, background: '#1890ff', borderRadius: '50%', display: 'inline-block', animation: 'pulse 1s infinite' }} />}>喷药中</Tag>
              ) : (
                <Tag color="default">已停止</Tag>
              )}
            </StatusValue>
          </StatusRow>
          {currentSprayerStatus.totalSprayedL !== undefined && (
            <StatusRow>
              <StatusLabel>累计喷洒</StatusLabel>
              <StatusValue>{currentSprayerStatus.totalSprayedL.toFixed(1)} L</StatusValue>
            </StatusRow>
          )}
        </PanelCard>
      )}

      <PanelCard size="small" title="喷药控制">
        <Form form={sprayerForm} layout="vertical" size="small">
          <Form.Item label="流量控制 (L/min)" style={{ marginBottom: 16 }}>
            <Slider
              min={0}
              max={10}
              step={0.5}
              value={currentSprayerStatus?.flowRateLpm ?? 2}
              onChange={handleSetFlowRate}
              marks={{ 0: '0', 2: '2', 5: '5', 10: '10' }}
              tooltip={{ formatter: (v: number | null) => `${v} L/min` }}
              disabled={!currentSprayer || currentSprayer.status !== 'online' && currentSprayer.status !== 'active'}
            />
          </Form.Item>
        </Form>

        <Row gutter={8}>
          <Col span={12}>
            <Button
              block
              type="primary"
              icon={<PlayCircleOutlined />}
              onClick={() => handleToggleSprayer(true)}
              loading={loadingAction === 'sprayer' && !currentSprayerStatus?.isSpraying}
              disabled={!currentSprayer || currentSprayer.status !== 'online' && currentSprayer.status !== 'active' || currentSprayerStatus?.isSpraying}
            >
              开始喷药
            </Button>
          </Col>
          <Col span={12}>
            <Button
              block
              danger
              icon={<StopOutlined />}
              onClick={() => handleToggleSprayer(false)}
              loading={loadingAction === 'sprayer' && currentSprayerStatus?.isSpraying}
              disabled={!currentSprayer || currentSprayer.status !== 'online' && currentSprayer.status !== 'active' || !currentSprayerStatus?.isSpraying}
            >
              停止喷药
            </Button>
          </Col>
        </Row>
      </PanelCard>
    </Space>
  )

  const speakerTab = (
    <Space direction="vertical" size="small" style={{ width: '100%' }}>
      <DeviceSelector>
        <SoundOutlined style={{ color: '#fa8c16' }} />
        <Select
          style={deviceSelectStyle}
          size="small"
          placeholder="选择喊话器"
          value={selectedSpeakerId}
          onChange={setSelectedSpeakerId}
          options={speakers.map(s => ({
            value: s.id,
            label: (
              <Space>
                <Badge status={(s.status === 'online' || s.status === 'active') ? 'success' : 'default'} />
                <span>{s.name || getPayloadTypeLabel(s.type)}</span>
              </Space>
            )
          }))}
        />
        {currentSpeaker && (
          <Tag color={(currentSpeaker.status === 'online' || currentSpeaker.status === 'active') ? 'success' : 'default'}>
            {currentSpeaker.status === 'online' || currentSpeaker.status === 'active' ? '在线' : '离线'}
          </Tag>
        )}
      </DeviceSelector>

      <PanelCard size="small" title={<span><FileTextOutlined style={{ marginRight: 6 }} />文字转语音 (TTS)</span>}>
        <Form form={ttsForm} layout="vertical" size="small">
          <Form.Item
            name="text"
            label="输入文字"
            rules={[{ required: true, message: '请输入要转换的文字' }, { max: 500, message: '最多500字' }]}
            style={{ marginBottom: 12 }}
          >
            <Input.TextArea
              rows={3}
              placeholder="请输入要转换为语音的文字内容..."
              showCount
              maxLength={500}
              style={{ resize: 'none' }}
            />
          </Form.Item>

          <Row gutter={[8, 8]}>
            <Col span={12}>
              <Form.Item name="voice" label="音色" initialValue="xiaoxiao" style={{ marginBottom: 8 }}>
                <Select
                  size="small"
                  options={[
                    { value: 'xiaoxiao', label: '晓晓 (女声)' },
                    { value: 'yunxi', label: '云希 (男声)' },
                    { value: 'yunjian', label: '云健 (男声)' },
                    { value: 'xiaoyi', label: '晓伊 (女声)' },
                    { value: 'aria', label: 'Aria (英文女)' },
                    { value: 'guy', label: 'Guy (英文男)' }
                  ]}
                />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="volume" label="音量" initialValue={80} style={{ marginBottom: 8 }}>
                <Slider min={0} max={100} size="small" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="speed" label="语速" initialValue={1.0} style={{ marginBottom: 8 }}>
                <Slider min={0.5} max={2.0} step={0.1} marks={{ 0.5: '0.5x', 1: '1x', 2: '2x' }} size="small" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item name="pitch" label="音调" initialValue={1.0} style={{ marginBottom: 8 }}>
                <Slider min={0.5} max={2.0} step={0.1} marks={{ 0.5: '0.5x', 1: '1x', 2: '2x' }} size="small" />
              </Form.Item>
            </Col>
          </Row>

          <Form.Item name="autoPlay" label="合成后自动广播" initialValue={true} valuePropName="checked" style={{ marginBottom: 12 }}>
            <Switch size="small" />
          </Form.Item>

          <Button
            block
            type="primary"
            icon={<SendOutlined />}
            onClick={handleCreateTTS}
            loading={loadingAction === 'tts'}
            disabled={!currentSpeaker || currentSpeaker.status !== 'online' && currentSpeaker.status !== 'active'}
          >
            生成语音并广播
          </Button>
        </Form>
      </PanelCard>

      <PanelCard size="small" title={<span><SoundOutlined style={{ marginRight: 6 }} />任务与音频</span>}>
        <Tabs
          size="small"
          items={[
            {
              key: 'tasks',
              label: `TTS任务 (${ttsTasks.length})`,
              children: ttsTasks.length === 0 ? (
                <div style={{ textAlign: 'center', padding: '20px 0', color: 'rgba(255,255,255,0.5)' }}>
                  暂无TTS任务
                </div>
              ) : (
                <List
                  size="small"
                  dataSource={ttsTasks.slice(0, 10)}
                  locale={{ emptyText: '暂无TTS任务' }}
                  renderItem={(task: TextToSpeechTask) => (
                    <AudioItem>
                      <List.Item.Meta
                        avatar={
                          <Avatar
                            size="small"
                            icon={task.status === 'completed' ? <SoundOutlined /> : task.status === 'processing' ? <ReloadOutlined spin /> : task.status === 'failed' ? <StopOutlined /> : <FileTextOutlined />}
                            style={{
                              background: task.status === 'completed' ? '#52c41a' : task.status === 'processing' ? '#1890ff' : task.status === 'failed' ? '#ff4d4f' : '#8c8c8c'
                            }}
                          />
                        }
                        title={
                          <Space size="small">
                            <span style={{ color: 'rgba(255,255,255,0.9)', fontSize: 13 }}>
                              {task.text.length > 20 ? task.text.substring(0, 20) + '...' : task.text}
                            </span>
                            <Tag color={
                              task.status === 'completed' ? 'success' :
                              task.status === 'processing' ? 'processing' :
                              task.status === 'failed' ? 'error' :
                              task.status === 'playing' ? 'blue' : 'default'
                            }>
                              {task.status === 'pending' ? '等待中' :
                               task.status === 'processing' ? '合成中' :
                               task.status === 'completed' ? '已完成' :
                               task.status === 'playing' ? '播放中' :
                               task.status === 'failed' ? '失败' : task.status}
                            </Tag>
                          </Space>
                        }
                        description={
                          <Space size="large">
                            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)' }}>
                              时长: {formatDuration(task.audioDurationSec)}
                            </span>
                            <span style={{ fontSize: 11, color: 'rgba(255,255,255,0.5)' }}>
                              {new Date(task.createdAt).toLocaleTimeString()}
                            </span>
                            {task.status === 'completed' && task.audio && (
                              <Space size={4}>
                                <Tooltip title="试听">
                                  <Button
                                    type="link"
                                    size="small"
                                    icon={<PlayCircleOutlined />}
                                    style={{ padding: 0 }}
                                    onClick={() => handlePlayAudio(task.audio!)}
                                  />
                                </Tooltip>
                                <Tooltip title="广播到无人机">
                                  <Button
                                    type="link"
                                    size="small"
                                    icon={<SoundOutlined />}
                                    style={{ padding: 0 }}
                                    onClick={() => handleBroadcastAudio(task.audio!.id)}
                                    disabled={!currentSpeaker || currentSpeaker.status !== 'online' && currentSpeaker.status !== 'active'}
                                  />
                                </Tooltip>
                                <Tooltip title="下载">
                                  <Button
                                    type="link"
                                    size="small"
                                    icon={<DownloadOutlined />}
                                    style={{ padding: 0 }}
                                    onClick={() => task.audio?.fileURL && window.open(task.audio.fileURL, '_blank')}
                                  />
                                </Tooltip>
                              </Space>
                            )}
                          </Space>
                        }
                      />
                    </AudioItem>
                  )}
                />
              )
            }
          ]}
        />
      </PanelCard>
    </Space>
  )

  const tabItems = [
    {
      key: 'camera',
      label: (
        <Space size={4}>
          <CameraOutlined />
          <span>相机 {cameras.length > 0 && `(${cameras.filter(c => c.status === 'online' || c.status === 'active').length}/${cameras.length})`}</span>
        </Space>
      ),
      children: cameras.length === 0 ? (
        <Alert
          type="info"
          showIcon
          message="暂无相机设备"
          description="请确保无人机已连接并识别到相机载荷"
          style={{ background: 'rgba(24, 144, 255, 0.05)', border: '1px solid rgba(24, 144, 255, 0.2)' }}
        />
      ) : cameraTab
    },
    {
      key: 'sprayer',
      label: (
        <Space size={4}>
          <RotateLeftOutlined />
          <span>喷药 {sprayers.length > 0 && `(${sprayers.filter(s => s.status === 'online' || s.status === 'active').length}/${sprayers.length})`}</span>
        </Space>
      ),
      children: sprayers.length === 0 ? (
        <Alert
          type="info"
          showIcon
          message="暂无喷药器设备"
          description="请确保无人机已连接并识别到喷药器载荷"
          style={{ background: 'rgba(82, 196, 26, 0.05)', border: '1px solid rgba(82, 196, 26, 0.2)' }}
        />
      ) : sprayerTab
    },
    {
      key: 'speaker',
      label: (
        <Space size={4}>
          <SoundOutlined />
          <span>喊话 {speakers.length > 0 && `(${speakers.filter(s => s.status === 'online' || s.status === 'active').length}/${speakers.length})`}</span>
        </Space>
      ),
      children: speakers.length === 0 ? (
        <Alert
          type="info"
          showIcon
          message="暂无喊话器设备"
          description="请确保无人机已连接并识别到喊话器载荷"
          style={{ background: 'rgba(250, 140, 22, 0.05)', border: '1px solid rgba(250, 140, 22, 0.2)' }}
        />
      ) : speakerTab
    }
  ]

  return (
    <Container>
      <Tabs
        size="small"
        items={tabItems}
        tabBarStyle={{ marginBottom: 4 }}
      />
      <audio ref={audioRef} onEnded={() => setPlayingAudio(null)} />
    </Container>
  )
}

export default PayloadControlPanel
