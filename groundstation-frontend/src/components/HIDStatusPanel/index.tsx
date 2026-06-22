import React from 'react'
import styled from 'styled-components'
import { Card, Tag, Switch, Space, Progress, Select, Button, Tooltip, Alert } from 'antd'
import {
  ThunderboltOutlined,
  GamepadOutlined,
  ControlOutlined,
  VerticalAlignTopOutlined,
  VerticalAlignBottomOutlined,
  ArrowUpOutlined,
  ArrowDownOutlined,
  ArrowLeftOutlined,
  ArrowRightOutlined,
  RotateLeftOutlined,
  RotateRightOutlined,
  SafetyCertificateOutlined,
  PauseCircleOutlined,
  HomeOutlined,
  WarningOutlined,
  ReloadOutlined
} from '@ant-design/icons'
import type { HIDState, HIDDeviceInfo } from '@/types'
import { HIDDeviceTypeText } from '@/types/remote-cockpit'

const Container = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);

  .ant-card-body {
    padding: 16px;
  }
`

const Header = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
`

const Title = styled.div`
  font-size: 14px;
  font-weight: 600;
  color: #fff;
  display: flex;
  align-items: center;
  gap: 8px;
`

const Label = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 4px;
`

const AxisContainer = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 16px;
`

const AxisCard = styled.div`
  background: rgba(255, 255, 255, 0.02);
  border-radius: 8px;
  padding: 10px 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
`

const AxisHeader = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 6px;
`

const AxisName = styled.div`
  font-size: 12px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
  gap: 4px;
`

const AxisValue = styled.div`
  font-size: 12px;
  font-family: 'Courier New', monospace;
  color: #fff;
`

const StickVisualization = styled.div`
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
  margin-bottom: 16px;
`

const StickContainer = styled.div`
  display: flex;
  flex-direction: column;
  align-items: center;
`

const StickLabel = styled.div`
  font-size: 11px;
  color: rgba(255, 255, 255, 0.5);
  margin-bottom: 6px;
`

const StickBase = styled.div`
  width: 80px;
  height: 80px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.05);
  border: 2px solid rgba(255, 255, 255, 0.1);
  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
`

const StickKnob = styled.div<{ $x: number; $y: number }>`
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: linear-gradient(135deg, #1890ff 0%, #52c41a 100%);
  box-shadow: 0 2px 8px rgba(24, 144, 255, 0.4);
  position: absolute;
  transform: translate(
    calc(${props => props.$x * 28}px),
    calc(${props => props.$y * -28}px)
  );
  transition: transform 0.05s linear;
`

const Crosshair = styled.div`
  position: absolute;
  width: 40px;
  height: 40px;
  border: 1px dashed rgba(255, 255, 255, 0.1);
  border-radius: 50%;
`

const ButtonsGrid = styled.div`
  display: grid;
  grid-template-columns: repeat(4, 1fr);
  gap: 6px;
  margin-bottom: 12px;
`

const ButtonIndicator = styled.div<{ $active: boolean; $color?: string }>`
  padding: 6px 4px;
  border-radius: 4px;
  text-align: center;
  font-size: 10px;
  background: ${props => props.$active
    ? props.$color ? `${props.$color}30` : 'rgba(24, 144, 255, 0.2)'
    : 'rgba(255, 255, 255, 0.03)'};
  border: 1px solid ${props => props.$active
    ? props.$color || 'rgba(24, 144, 255, 0.4)'
    : 'rgba(255, 255, 255, 0.05)'};
  color: ${props => props.$active ? props.$color || '#1890ff' : 'rgba(255, 255, 255, 0.5)'};
  font-weight: ${props => props.$active ? 600 : 400};
  transition: all 0.1s;
`

interface HIDStatusPanelProps {
  hidState: HIDState
  devices: HIDDeviceInfo[]
  supported: boolean
  onToggleEnabled: (enabled: boolean) => void
  onSelectDevice: (deviceId: string | null) => void
  onCalibrate: () => void
  onRefreshDevices: () => void
}

const HIDStatusPanel: React.FC<HIDStatusPanelProps> = ({
  hidState,
  devices,
  supported,
  onToggleEnabled,
  onSelectDevice,
  onCalibrate,
  onRefreshDevices
}) => {
  const activeDevice = devices.find(d => d.id === hidState.active_device_id)
  const deviceOptions = devices.map(d => ({
    value: d.id,
    label: `${d.name} (${HIDDeviceTypeText[d.type]})`
  }))

  const normalizedPitch = Math.max(-1, Math.min(1, hidState.axes.pitch))
  const normalizedRoll = Math.max(-1, Math.min(1, hidState.axes.roll))
  const normalizedYaw = Math.max(-1, Math.min(1, hidState.axes.yaw))
  const normalizedThrottle = Math.max(0, Math.min(1, hidState.axes.throttle))

  if (!supported) {
    return (
      <Container>
        <Alert
          type="warning"
          showIcon
          icon={<WarningOutlined />}
          message="HID 设备不支持"
          description="当前浏览器不支持 Gamepad API，请使用支持的浏览器（Chrome、Edge、Firefox 等）。"
          style={{ background: 'transparent', border: 'none', padding: 0 }}
        />
      </Container>
    )
  }

  return (
    <Container>
      <Header>
        <Title>
          <GamepadOutlined style={{ color: '#52c41a' }} />
          摇杆控制
        </Title>
        <Space>
          <Tag color={hidState.enabled ? 'green' : 'default'}>
            {hidState.enabled ? '已启用' : '已禁用'}
          </Tag>
          <Switch
            checked={hidState.enabled}
            onChange={onToggleEnabled}
            size="small"
          />
        </Space>
      </Header>

      {hidState.calibration_needed && (
        <Alert
          type="warning"
          showIcon
          message="需要校准"
          description="摇杆设备需要校准以获得更好的控制精度。"
          action={
            <Button size="small" type="primary" onClick={onCalibrate}>
              校准
            </Button>
          }
          style={{ marginBottom: 12, background: 'rgba(250, 173, 20, 0.1)', border: 'none' }}
        />
      )}

      <div style={{ marginBottom: 12 }}>
        <Label>输入设备</Label>
        <Space.Compact style={{ width: '100%' }}>
          <Select
            value={hidState.active_device_id}
            onChange={onSelectDevice}
            options={deviceOptions}
            style={{ flex: 1 }}
            size="small"
            placeholder={devices.length > 0 ? '选择设备' : '未检测到设备'}
            allowClear
          />
          <Tooltip title="刷新设备列表">
            <Button
              size="small"
              icon={<ReloadOutlined />}
              onClick={onRefreshDevices}
            />
          </Tooltip>
        </Space.Compact>
        {activeDevice && (
          <div style={{ marginTop: 4, fontSize: 11, color: 'rgba(255,255,255,0.5)' }}>
            类型: {HIDDeviceTypeText[activeDevice.type]}
          </div>
        )}
      </div>

      <StickVisualization>
        <StickContainer>
          <StickLabel>俯仰/横滚</StickLabel>
          <StickBase>
            <Crosshair />
            <StickKnob $x={normalizedRoll} $y={normalizedPitch} />
          </StickBase>
        </StickContainer>
        <StickContainer>
          <StickLabel>油门/偏航</StickLabel>
          <StickBase>
            <Crosshair />
            <StickKnob $x={normalizedYaw} $y={normalizedThrottle * 2 - 1} />
          </StickBase>
        </StickContainer>
      </StickVisualization>

      <AxisContainer>
        <AxisCard>
          <AxisHeader>
            <AxisName>
              <ArrowUpOutlined />
              <ArrowDownOutlined />
              俯仰 (Pitch)
            </AxisName>
            <AxisValue>{(normalizedPitch * 100).toFixed(0)}%</AxisValue>
          </AxisHeader>
          <Progress
            percent={Math.round(Math.abs(normalizedPitch) * 100)}
            showInfo={false}
            strokeColor={normalizedPitch >= 0 ? '#52c41a' : '#faad14'}
            size="small"
          />
        </AxisCard>

        <AxisCard>
          <AxisHeader>
            <AxisName>
              <ArrowLeftOutlined />
              <ArrowRightOutlined />
              横滚 (Roll)
            </AxisName>
            <AxisValue>{(normalizedRoll * 100).toFixed(0)}%</AxisValue>
          </AxisHeader>
          <Progress
            percent={Math.round(Math.abs(normalizedRoll) * 100)}
            showInfo={false}
            strokeColor={normalizedRoll >= 0 ? '#1890ff' : '#722ed1'}
            size="small"
          />
        </AxisCard>

        <AxisCard>
          <AxisHeader>
            <AxisName>
              <RotateLeftOutlined />
              <RotateRightOutlined />
              偏航 (Yaw)
            </AxisName>
            <AxisValue>{(normalizedYaw * 100).toFixed(0)}%</AxisValue>
          </AxisHeader>
          <Progress
            percent={Math.round(Math.abs(normalizedYaw) * 100)}
            showInfo={false}
            strokeColor={normalizedYaw >= 0 ? '#eb2f96' : '#13c2c2'}
            size="small"
          />
        </AxisCard>

        <AxisCard>
          <AxisHeader>
            <AxisName>
              <VerticalAlignTopOutlined />
              <VerticalAlignBottomOutlined />
              油门 (Throttle)
            </AxisName>
            <AxisValue>{(normalizedThrottle * 100).toFixed(0)}%</AxisValue>
          </AxisHeader>
          <Progress
            percent={Math.round(normalizedThrottle * 100)}
            showInfo={false}
            strokeColor="#fa8c16"
            size="small"
          />
        </AxisCard>
      </AxisContainer>

      <Label>
        <ControlOutlined />
        按钮状态
      </Label>
      <ButtonsGrid>
        <ButtonIndicator $active={hidState.buttons.arm} $color="#52c41a">
          <ThunderboltOutlined /><br />解锁
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.disarm} $color="#ff4d4f">
          <SafetyCertificateOutlined /><br />上锁
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.takeoff} $color="#1890ff">
          <ArrowUpOutlined /><br />起飞
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.land} $color="#13c2c2">
          <ArrowDownOutlined /><br />降落
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.rtl} $color="#722ed1">
          <HomeOutlined /><br />返航
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.pause} $color="#faad14">
          <PauseCircleOutlined /><br />暂停
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.mode_switch} $color="#fa8c16">
          <ControlOutlined /><br />模式
        </ButtonIndicator>
        <ButtonIndicator $active={hidState.buttons.emergency_stop} $color="#ff4d4f">
          <WarningOutlined /><br />急停
        </ButtonIndicator>
      </ButtonsGrid>
    </Container>
  )
}

export default HIDStatusPanel
