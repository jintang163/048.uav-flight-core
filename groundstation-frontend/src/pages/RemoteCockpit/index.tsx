import React, { useEffect, useState, useCallback } from 'react'
import styled from 'styled-components'
import {
  Row,
  Col,
  Card,
  Button,
  Space,
  Tag,
  Select,
  Modal,
  Alert,
  Tooltip,
  Divider,
  Statistic,
  Dropdown,
  Switch,
  message
} from 'antd'
import {
  ThunderboltOutlined,
  PlayCircleOutlined,
  StopOutlined,
  SafetyOutlined,
  SwapOutlined,
  RocketOutlined,
  ApiOutlined,
  EyeOutlined,
  ClockCircleOutlined,
  WarningOutlined,
  DashboardOutlined,
  CheckCircleOutlined,
  QuestionCircleOutlined,
  ReloadOutlined
} from '@ant-design/icons'
import useRemoteCockpit from '@/hooks/useRemoteCockpit'
import useHIDGamepad from '@/hooks/useHIDGamepad'
import { useUAV } from '@/hooks/useUAV'
import VideoStreamPlayer from '@/components/VideoStreamPlayer'
import VideoQualityControl from '@/components/VideoQualityControl'
import HIDStatusPanel from '@/components/HIDStatusPanel'
import DualLinkStatusPanel from '@/components/DualLinkStatusPanel'
import LinkStatusIndicator from '@/components/LinkStatusIndicator'
import TelemetryPanel from '@/components/TelemetryPanel'
import ArtificialHorizon from '@/components/ArtificialHorizon'
import BatteryIndicator from '@/components/BatteryIndicator'
import RCChannels from '@/components/RCChannels'
import type { VideoQualityPreset, CockpitMode } from '@/types'
import { CockpitModeText } from '@/types/remote-cockpit'
import { LinkType } from '@/types/link'

const PageContainer = styled.div`
  width: 100%;
  height: 100%;
  padding: 16px;
  overflow: auto;
  background: #0f172a;
`

const HeaderBar = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 12px 16px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.08);
`

const Title = styled.div`
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 10px;
`

const ModeBadge = styled(Tag)<{ $mode: CockpitMode }>`
  font-size: 13px;
  padding: 2px 10px;
  font-weight: 600;

  ${props => {
    switch (props.$mode) {
      case CockpitMode.FLYING:
        return 'background: rgba(82, 196, 26, 0.2) !important; color: #52c41a !important; border-color: rgba(82, 196, 26, 0.3) !important;'
      case CockpitMode.MISSION:
        return 'background: rgba(24, 144, 255, 0.2) !important; color: #1890ff !important; border-color: rgba(24, 144, 255, 0.3) !important;'
      case CockpitMode.CONNECTING:
        return 'background: rgba(250, 173, 20, 0.2) !important; color: #faad14 !important; border-color: rgba(250, 173, 20, 0.3) !important;'
      case CockpitMode.EMERGENCY:
        return 'background: rgba(255, 77, 79, 0.2) !important; color: #ff4d4f !important; border-color: rgba(255, 77, 79, 0.3) !important;'
      case CockpitMode.DISCONNECTED:
        return 'background: rgba(255, 255, 255, 0.1) !important; color: rgba(255,255,255,0.6) !important; border-color: rgba(255, 255, 255, 0.15) !important;'
      default:
        return ''
    }
  }}
`

const UAVSelector = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
`

const StatsContainer = styled.div`
  display: flex;
  gap: 16px;
`

const StatCard = styled.div`
  padding: 8px 16px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  border: 1px solid rgba(255, 255, 255, 0.06);
  text-align: center;
  min-width: 100px;
`

const StatLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 2px;
`

const StatValue = styled.div`
  font-size: 18px;
  font-weight: 700;
  color: #fff;
  font-family: 'Courier New', monospace;
`

const formatDuration = (ms: number): string => {
  const seconds = Math.floor(ms / 1000)
  const mins = Math.floor(seconds / 60)
  const secs = seconds % 60
  return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`
}

const RemoteCockpit: React.FC = () => {
  const {
    state,
    isActive,
    mode,
    currentUAVId,
    videoConfig,
    videoStatus,
    networkMetrics,
    linkStatus,
    hidState,
    session,
    startSession,
    endSession,
    startVideoStreaming,
    stopVideoStreaming,
    setVideoQualityPreset,
    enableAdaptiveQuality,
    fetchLinkStatus,
    enableLinkFailover,
    switchPrimaryLink,
    enableAutoMissionFallback,
    fetchAvailableUAVs,
    switchUAV,
    setHIDAxesState,
    setHIDButtonsState,
    toggleHIDEnabled,
    selectHIDDevice,
    setCockpitMode
  } = useRemoteCockpit()

  const { supported, devices, activeDeviceId, selectDevice, calibrateDevice, refreshDevices } = useHIDGamepad()
  const { uavList, currentUAV, loadUAVList, selectCurrentUAV } = useUAV()

  const [confirmStopVisible, setConfirmStopVisible] = useState(false)
  const [switchingUAV, setSwitchingUAV] = useState(false)

  useEffect(() => {
    loadUAVList({ pageSize: 100 })
    fetchAvailableUAVs()
  }, [loadUAVList, fetchAvailableUAVs])

  useEffect(() => {
    if (hidState.enabled && isActive) {
      setHIDAxesState(hidState.axes)
    }
  }, [hidState.axes, hidState.enabled, isActive, setHIDAxesState])

  useEffect(() => {
    if (hidState.enabled) {
      setHIDButtonsState(hidState.buttons)
    }
  }, [hidState.buttons, hidState.enabled, setHIDButtonsState])

  const handleStartSession = useCallback(async () => {
    if (!currentUAVId) {
      message.warning('请先选择无人机')
      return
    }
    try {
      await startSession(currentUAVId)
      await startVideoStreaming()
      setCockpitMode(CockpitMode.FLYING)
      message.success('远程驾驶舱已启动')
    } catch (error) {
      message.error('启动远程驾驶舱失败')
    }
  }, [currentUAVId, startSession, startVideoStreaming, setCockpitMode])

  const handleStopSession = useCallback(async () => {
    setConfirmStopVisible(true)
  }, [])

  const confirmStopSession = useCallback(async () => {
    try {
      await endSession()
      setConfirmStopVisible(false)
      message.success('远程驾驶舱已停止')
    } catch (error) {
      message.error('停止远程驾驶舱失败')
    }
  }, [endSession])

  const handleSwitchUAV = useCallback(async (uavId: string) => {
    if (uavId === currentUAVId) return
    if (!state.available_uav_ids.includes(uavId)) {
      message.warning('该无人机不可用于远程驾驶')
      return
    }
    try {
      setSwitchingUAV(true)
      if (isActive) {
        await switchUAV(uavId)
      }
      selectCurrentUAV(uavId)
      message.success(`已切换至无人机: ${uavList.find(u => u.id === uavId)?.name || uavId}`)
    } catch (error) {
      message.error('切换无人机失败')
    } finally {
      setSwitchingUAV(false)
    }
  }, [currentUAVId, state.available_uav_ids, isActive, switchUAV, selectCurrentUAV, uavList])

  const uavOptions = uavList
    .filter(uav => state.available_uav_ids.includes(uav.id) || !isActive)
    .map(uav => ({
      value: uav.id,
      label: (
        <Space>
          <ApiOutlined />
          {uav.name}
          <Tag color={uav.status !== 'disconnected' && uav.status !== 'error' ? 'green' : 'default'} style={{ fontSize: 10 }}>
            {uav.status !== 'disconnected' && uav.status !== 'error' ? '在线' : '离线'}
          </Tag>
          {state.available_uav_ids.includes(uav.id) && (
            <Tag color="cyan" style={{ fontSize: 10 }}>可用</Tag>
          )}
        </Space>
      )
    }))

  const flightTime = session?.total_flight_time_ms || 0

  return (
    <PageContainer>
      <HeaderBar>
        <Space>
          <Title>
            <ThunderboltOutlined style={{ color: '#1890ff' }} />
            远程驾驶舱
          </Title>
          <ModeBadge $mode={mode}>
            {CockpitModeText[mode]}
          </ModeBadge>
        </Space>

        <Space size={16}>
          <StatsContainer>
            <StatCard>
              <StatLabel>
                <ClockCircleOutlined /> 飞行时长
              </StatLabel>
              <StatValue>{formatDuration(flightTime)}</StatValue>
            </StatCard>
            <StatCard>
              <StatLabel>
                <DashboardOutlined /> 发送指令
              </StatLabel>
              <StatValue>{session?.commands_sent || 0}</StatValue>
            </StatCard>
            <StatCard>
              <StatLabel>
                <SwapOutlined /> 链路切换
              </StatLabel>
              <StatValue>{linkStatus?.failover_count || 0}</StatValue>
            </StatCard>
          </StatsContainer>

          <UAVSelector>
            <Select
              value={currentUAVId}
              onChange={handleSwitchUAV}
              options={uavOptions}
              style={{ width: 220 }}
              placeholder="选择无人机"
              loading={switchingUAV}
              disabled={switchingUAV}
              allowClear={false}
            />
            <Tooltip title="刷新无人机列表">
              <Button
                size="small"
                icon={<ReloadOutlined />}
                onClick={() => {
                  loadUAVList({ pageSize: 100 })
                  fetchAvailableUAVs()
                }}
              />
            </Tooltip>
          </UAVSelector>

          <Space>
            <Tooltip title={state.auto_mission_fallback ? '断连自动切换航线已启用' : '断连自动切换航线已禁用'}>
              <Space>
                <SafetyOutlined style={{ color: state.auto_mission_fallback ? '#52c41a' : 'rgba(255,255,255,0.4)' }} />
                <Switch
                  checked={state.auto_mission_fallback}
                  onChange={enableAutoMissionFallback}
                  size="small"
                />
              </Space>
            </Tooltip>

            {!isActive ? (
              <Button
                type="primary"
                icon={<PlayCircleOutlined />}
                onClick={handleStartSession}
                disabled={!currentUAVId}
                size="large"
              >
                启动驾驶舱
              </Button>
            ) : (
              <Button
                danger
                icon={<StopOutlined />}
                onClick={handleStopSession}
                size="large"
              >
                停止驾驶舱
              </Button>
            )}
          </Space>
        </Space>
      </HeaderBar>

      {!isActive && mode === CockpitMode.IDLE && (
        <Alert
          type="info"
          showIcon
          icon={<QuestionCircleOutlined />}
          message="准备开始远程驾驶"
          description="选择可用的无人机并点击「启动驾驶舱」按钮开始远程飞行。确保4G/5G网络和数传电台连接正常。"
          style={{ marginBottom: 16, background: 'rgba(24, 144, 255, 0.1)', border: 'none' }}
        />
      )}

      {state.last_video_disconnect_time && mode !== CockpitMode.MISSION && (
        <Alert
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          message="视频流连接不稳定"
          description={`视频流于 ${new Date(state.last_video_disconnect_time).toLocaleTimeString()} 断开${state.auto_mission_fallback ? '，已自动切换至航线飞行模式' : ''}`}
          style={{ marginBottom: 16 }}
        />
      )}

      <Row gutter={[16, 16]}>
        <Col xs={24} xl={16}>
          <VideoStreamPlayer
            uavId={currentUAVId || ''}
            videoStatus={videoStatus}
            videoConfig={videoConfig}
            onStart={startVideoStreaming}
            onStop={stopVideoStreaming}
            autoPlay={isActive}
            showControls
            showMetrics
          />
        </Col>

        <Col xs={24} xl={8}>
          <Row gutter={[0, 16]}>
            <Col span={24}>
              <Row gutter={16}>
                <Col span={12}>
                  <ArtificialHorizon
                    pitch={currentUAV?.attitude?.pitch || 0}
                    roll={currentUAV?.attitude?.roll || 0}
                    heading={currentUAV?.heading || currentUAV?.attitude?.yaw || 0}
                  />
                </Col>
                <Col span={12}>
                  <BatteryIndicator
                    voltage={currentUAV?.battery?.voltage || 0}
                    current={currentUAV?.battery?.current || 0}
                    remaining={currentUAV?.battery?.remaining || 0}
                    temperature={currentUAV?.battery?.temperature || 0}
                  />
                </Col>
              </Row>
            </Col>

            <Col span={24}>
              <DualLinkStatusPanel
                linkStatus={linkStatus}
                onFailoverToggle={enableLinkFailover}
                onSwitchPrimary={switchPrimaryLink}
                onRefresh={fetchLinkStatus}
                radioRSSI={linkStatus?.primary_link === LinkType.RADIO ? -70 : -75}
                lteRSSI={linkStatus?.primary_link === LinkType.LTE ? -75 : -80}
              />
            </Col>
          </Row>
        </Col>

        <Col xs={24} xl={12}>
          <VideoQualityControl
            videoConfig={videoConfig}
            videoStatus={videoStatus}
            networkMetrics={networkMetrics}
            onPresetChange={setVideoQualityPreset}
            onAdaptiveToggle={enableAdaptiveQuality}
            onBitrateChange={(bitrate) => {
              setVideoQualityPreset(VideoQualityPreset.MEDIUM)
            }}
            onResolutionChange={() => {}}
            qualityAdjustmentCount={state.quality_adjustment_count}
          />
        </Col>

        <Col xs={24} xl={12}>
          <HIDStatusPanel
            hidState={hidState}
            devices={devices}
            supported={supported}
            onToggleEnabled={toggleHIDEnabled}
            onSelectDevice={selectDevice}
            onCalibrate={calibrateDevice}
            onRefreshDevices={refreshDevices}
          />
        </Col>

        <Col xs={24} xl={12}>
          <TelemetryPanel uavId={currentUAVId || ''} />
        </Col>

        <Col xs={24} xl={12}>
          <RCChannels uavId={currentUAVId || ''} />
        </Col>
      </Row>

      <Modal
        title="确认停止远程驾驶舱"
        open={confirmStopVisible}
        onOk={confirmStopSession}
        onCancel={() => setConfirmStopVisible(false)}
        okText="确认停止"
        cancelText="取消"
        okButtonProps={{ danger: true }}
      >
        <p>确定要停止远程驾驶舱吗？停止后：</p>
        <ul>
          <li>视频传输将断开</li>
          <li>摇杆控制将失效</li>
          {state.auto_mission_fallback && (
            <li style={{ color: '#faad14' }}>无人机将自动切换至航线飞行模式</li>
          )}
        </ul>
      </Modal>
    </PageContainer>
  )
}

export default RemoteCockpit
