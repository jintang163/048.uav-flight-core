import React from 'react'
import styled from 'styled-components'
import { Card, Switch, Slider, Select, Tag, Space, Divider, Tooltip, Progress } from 'antd'
import {
  ThunderboltOutlined,
  SettingOutlined,
  RocketOutlined,
  ExperimentOutlined,
  BulbOutlined,
  WifiOutlined
} from '@ant-design/icons'
import type { VideoStreamConfig, VideoStreamStatus, NetworkMetrics, VideoQualityPreset } from '@/types'
import {
  VideoQualityPresetText,
  VideoQualityPresetConfig,
  VideoCodecText,
  ResolutionOrder,
  getResolutionDimensions
} from '@/types/remote-cockpit'

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
  margin-bottom: 6px;
  display: flex;
  align-items: center;
  gap: 4px;
`

const Value = styled.div`
  font-size: 13px;
  color: rgba(255, 255, 255, 0.9);
  font-family: 'Courier New', monospace;
  margin-bottom: 8px;
`

const Row = styled.div`
  margin-bottom: 14px;

  &:last-child {
    margin-bottom: 0;
  }
`

const MetricCard = styled.div<{ $color?: string }>`
  background: rgba(255, 255, 255, 0.03);
  border-radius: 6px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.05);
`

const MetricLabel = styled.div`
  font-size: 10px;
  color: rgba(255, 255, 255, 0.4);
  margin-bottom: 2px;
`

const MetricValue = styled.div<{ $color?: string }>`
  font-size: 15px;
  font-weight: 600;
  font-family: 'Courier New', monospace;
  color: ${props => props.$color || '#fff'};
`

const Grid = styled.div`
  display: grid;
  grid-template-columns: repeat(3, 1fr);
  gap: 8px;
  margin-bottom: 12px;
`

interface VideoQualityControlProps {
  videoConfig: VideoStreamConfig
  videoStatus: VideoStreamStatus
  networkMetrics: NetworkMetrics | null
  onPresetChange: (preset: VideoQualityPreset) => void
  onAdaptiveToggle: (enabled: boolean) => void
  onBitrateChange: (bitrate: number) => void
  onResolutionChange: (resolution: string) => void
  qualityAdjustmentCount: number
}

const VideoQualityControl: React.FC<VideoQualityControlProps> = ({
  videoConfig,
  videoStatus,
  networkMetrics,
  onPresetChange,
  onAdaptiveToggle,
  onBitrateChange,
  onResolutionChange,
  qualityAdjustmentCount
}) => {
  const presetOptions = Object.values(VideoQualityPreset).map(preset => ({
    value: preset,
    label: (
      <Space>
        {VideoQualityPresetText[preset]}
        <Tag color="blue" style={{ fontSize: 10, padding: '0 4px' }}>
          {VideoQualityPresetConfig[preset].bitrate_kbps}kbps
        </Tag>
      </Space>
    )
  }))

  const resolutionOptions = ResolutionOrder.map(res => ({
    value: res,
    label: res
  }))

  const currentPreset = (() => {
    for (const preset of Object.values(VideoQualityPreset)) {
      const presetCfg = VideoQualityPresetConfig[preset]
      if (Math.abs(presetCfg.bitrate_kbps - videoConfig.bitrate_kbps) < 500 &&
          presetCfg.resolution === videoConfig.resolution) {
        return preset
      }
    }
    return undefined
  })()

  const bandwidthUtilization = networkMetrics?.bandwidth_estimate_kbps
    ? Math.min(100, (videoStatus.current_bitrate_kbps / networkMetrics.bandwidth_estimate_kbps) * 100)
    : 0

  const dimensions = getResolutionDimensions(videoConfig.resolution)

  return (
    <Container>
      <Header>
        <Title>
          <SettingOutlined style={{ color: '#1890ff' }} />
          画质控制
          {qualityAdjustmentCount > 0 && (
            <Tag color="purple" style={{ fontSize: 10 }}>
              已自适应调整 {qualityAdjustmentCount} 次
            </Tag>
          )}
        </Title>
        <Space>
          <Tag color={videoConfig.adaptive_enabled ? 'green' : 'default'}>
            {videoConfig.adaptive_enabled ? <BulbOutlined /> : <ExperimentOutlined />}
            {videoConfig.adaptive_enabled ? '自动' : '手动'}
          </Tag>
        </Space>
      </Header>

      <Grid>
        <MetricCard>
          <MetricLabel>当前码率</MetricLabel>
          <MetricValue $color="#1890ff">
            {videoStatus.current_bitrate_kbps.toFixed(0)} kbps
          </MetricValue>
        </MetricCard>
        <MetricCard>
          <MetricLabel>目标码率</MetricLabel>
          <MetricValue $color="#52c41a">
            {videoConfig.bitrate_kbps} kbps
          </MetricValue>
        </MetricCard>
        <MetricCard>
          <MetricLabel>分辨率</MetricLabel>
          <MetricValue $color="#722ed1">
            {dimensions.width}x{dimensions.height}
          </MetricValue>
        </MetricCard>
      </Grid>

      {networkMetrics && (
        <>
          <Row>
            <Label>
              <WifiOutlined />
              网络带宽估计: {networkMetrics.bandwidth_estimate_kbps.toFixed(0)} kbps
            </Label>
            <Progress
              percent={Math.round(bandwidthUtilization)}
              showInfo
              format={() => `带宽占用 ${bandwidthUtilization.toFixed(0)}%`}
              strokeColor={{
                '0%': '#52c41a',
                '100%': bandwidthUtilization > 90 ? '#ff4d4f' : '#1890ff'
              }}
              size="small"
            />
          </Row>
          <Divider style={{ margin: '8px 0', borderColor: 'rgba(255,255,255,0.05)' }} />
        </>
      )}

      <Row>
        <Label>
          <RocketOutlined />
          画质预设
        </Label>
        <Select
          value={currentPreset}
          onChange={onPresetChange}
          options={presetOptions}
          style={{ width: '100%' }}
          size="small"
          placeholder="选择预设"
        />
      </Row>

      <Row>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 6 }}>
          <Label style={{ marginBottom: 0 }}>
            <ThunderboltOutlined />
            画质自适应
          </Label>
          <Switch
            checked={videoConfig.adaptive_enabled}
            onChange={onAdaptiveToggle}
            size="small"
          />
        </div>
        <div style={{ fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>
          {videoConfig.adaptive_enabled
            ? '根据网络状况自动调整码率和分辨率'
            : '使用固定的视频参数'}
        </div>
      </Row>

      <Row>
        <Label>
          码率 (kbps)
          <Tooltip title="数值越高画质越好，但对网络要求更高">
            <Tag color="blue" style={{ fontSize: 10, marginLeft: 4 }}>?</Tag>
          </Tooltip>
        </Label>
        <Value>{videoConfig.bitrate_kbps} kbps (范围: {videoConfig.min_bitrate_kbps} - {videoConfig.max_bitrate_kbps})</Value>
        <Slider
          min={videoConfig.min_bitrate_kbps}
          max={videoConfig.max_bitrate_kbps}
          step={100}
          value={videoConfig.bitrate_kbps}
          onChange={onBitrateChange}
          disabled={videoConfig.adaptive_enabled}
          tooltip={{ formatter: (v) => `${v} kbps` }}
        />
      </Row>

      <Row>
        <Label>分辨率</Label>
        <Value>{videoConfig.resolution}</Value>
        <Select
          value={videoConfig.resolution}
          onChange={onResolutionChange}
          options={resolutionOptions}
          style={{ width: '100%' }}
          size="small"
          disabled={videoConfig.adaptive_enabled}
        />
      </Row>

      <Row>
        <Label>编码格式</Label>
        <Value>
          <Tag color="cyan">{VideoCodecText[videoConfig.codec]}</Tag>
          <span style={{ marginLeft: 8, fontSize: 11, color: 'rgba(255,255,255,0.4)' }}>
            关键帧间隔: {videoConfig.keyframe_interval} · FPS: {videoConfig.fps}
          </span>
        </Value>
      </Row>
    </Container>
  )
}

export default VideoQualityControl
