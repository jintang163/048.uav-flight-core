import React, { useEffect, useState, useMemo } from 'react'
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
  Form,
  Input,
  InputNumber,
  Slider,
  Switch,
  Badge,
  Table,
  Tooltip,
  message
} from 'antd'
import {
  RocketOutlined,
  TeamOutlined,
  WarningOutlined,
  BulbOutlined,
  PlayCircleOutlined,
  PauseCircleOutlined,
  ReloadOutlined,
  StopOutlined,
  PlusOutlined
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import {
  fetchFormationList,
  fetchFormationDetail,
  startFormationById,
  pauseFormationById,
  resumeFormationById,
  stopFormationById,
  createNewFormation,
  selectFormation
} from '@/store/slices/formation'
import { getUAVList } from '@/api/uav'
import { setFormationLight } from '@/api/formation'
import { FormationType, FormationStatus, LightEffect } from '@/types'
import type { Formation, UAVListItem, FormationLightConfig } from '@/types'

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
  gap: 24px;
`

const Title = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  font-size: 18px;
  font-weight: 600;
  color: #fff;
`

const FormationSelector = styled(Select)`
  width: 250px;

  .ant-select-selector {
    background: rgba(255, 255, 255, 0.1) !important;
    border: 1px solid rgba(255, 255, 255, 0.2) !important;
  }

  .ant-select-selection-item {
    color: #fff !important;
  }
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`

const StatusBadge = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 20px;
  font-size: 13px;
`

const Content = styled.div`
  flex: 1;
  display: grid;
  grid-template-columns: 1fr 350px;
  gap: 16px;
  overflow: hidden;
`

const LeftPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow: hidden;
`

const RightPanel = styled.div`
  display: flex;
  flex-direction: column;
  gap: 16px;
  overflow-y: auto;
`

const VisualizationCard = styled(Card)`
  flex: 1;
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  overflow: hidden;

  .ant-card-head {
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
  }

  .ant-card-body {
    height: calc(100% - 57px);
    padding: 16px;
  }
`

const FormationCanvas = styled.div`
  width: 100%;
  height: 100%;
  position: relative;
  background: rgba(0, 0, 0, 0.3);
  border-radius: 8px;
  overflow: hidden;
`

const GridOverlay = styled.div`
  position: absolute;
  inset: 0;
  background-image:
    linear-gradient(rgba(255, 255, 255, 0.05) 1px, transparent 1px),
    linear-gradient(90deg, rgba(255, 255, 255, 0.05) 1px, transparent 1px);
  background-size: 40px 40px;
  pointer-events: none;
`

const UAVDot = styled.div<{ $x: number; $y: number; $isLeader: boolean; $active: boolean }>`
  position: absolute;
  width: 24px;
  height: 24px;
  border-radius: 50%;
  background: ${props =>
    props.$isLeader
      ? 'linear-gradient(135deg, #1890ff, #096dd9)'
      : 'linear-gradient(135deg, #52c41a, #389e0d)'};
  border: 2px solid ${props => (props.$active ? '#fff' : 'rgba(255,255,255,0.3)')};
  transform: translate(-50%, -50%);
  left: ${props => props.$x}%;
  top: ${props => props.$y}%;
  box-shadow: 0 0 12px
    ${props =>
      props.$isLeader ? 'rgba(24, 144, 255, 0.6)' : 'rgba(82, 196, 26, 0.6)'};
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 10px;
  font-weight: 600;
  color: #fff;
  opacity: ${props => (props.$active ? 1 : 0.5)};
`

const LeaderLabel = styled.div`
  position: absolute;
  top: -20px;
  left: 50%;
  transform: translateX(-50%);
  font-size: 10px;
  color: #1890ff;
  white-space: nowrap;
`

const InfoCard = styled(Card)`
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;

  .ant-card-head {
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    min-height: 48px;
  }

  .ant-card-body {
    padding: 16px;
  }
`

const InfoRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);

  &:last-child {
    border-bottom: none;
  }
`

const InfoLabel = styled.span`
  color: rgba(255, 255, 255, 0.6);
  font-size: 13px;
`

const InfoValue = styled.span`
  color: #fff;
  font-size: 13px;
  font-weight: 500;
`

const LightControlCard = styled(Card)`
  background: rgba(255, 255, 255, 0.05);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;

  .ant-card-head {
    background: rgba(255, 255, 255, 0.05);
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
    color: #fff;
    min-height: 48px;
  }

  .ant-card-body {
    padding: 16px;
  }
`

const ColorPreview = styled.div<{ $r: number; $g: number; $b: number }>`
  width: 100%;
  height: 60px;
  border-radius: 8px;
  background: rgb(${props => props.$r}, ${props => props.$g}, ${props => props.$b});
  margin-bottom: 16px;
  box-shadow: 0 4px 12px
    rgba(${props => props.$r}, ${props => props.$g}, ${props => props.$b}, 0.3);
`

const CollisionWarningCard = styled(Card)`
  background: rgba(255, 77, 79, 0.1);
  border: 1px solid rgba(255, 77, 79, 0.3);
  border-radius: 8px;

  .ant-card-head {
    background: rgba(255, 77, 79, 0.1);
    border-bottom: 1px solid rgba(255, 77, 79, 0.2);
    color: #ff4d4f;
    min-height: 48px;
  }

  .ant-card-body {
    padding: 12px 16px;
  }
`

const WarningItem = styled.div`
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 0;
  border-bottom: 1px solid rgba(255, 77, 79, 0.1);

  &:last-child {
    border-bottom: none;
  }
`

const WarningText = styled.span`
  color: #ff7875;
  font-size: 13px;
`

const getFormationTypeLabel = (type: FormationType): string => {
  const labels: Record<FormationType, string> = {
    [FormationType.LINE]: '一字型',
    [FormationType.TRIANGLE]: '三角形',
    [FormationType.CIRCLE]: '圆形'
  }
  return labels[type] || '未知'
}

const getFormationStatusColor = (status: FormationStatus): string => {
  const colors: Record<FormationStatus, string> = {
    [FormationStatus.IDLE]: 'default',
    [FormationStatus.READY]: 'processing',
    [FormationStatus.EXECUTING]: 'success',
    [FormationStatus.PAUSED]: 'warning',
    [FormationStatus.COMPLETED]: 'default'
  }
  return colors[status] || 'default'
}

const getFormationStatusText = (status: FormationStatus): string => {
  const texts: Record<FormationStatus, string> = {
    [FormationStatus.IDLE]: '待命',
    [FormationStatus.READY]: '就绪',
    [FormationStatus.EXECUTING]: '执行中',
    [FormationStatus.PAUSED]: '已暂停',
    [FormationStatus.COMPLETED]: '已完成'
  }
  return texts[status] || '未知'
}

const FormationMonitor: React.FC = () => {
  const dispatch = useAppDispatch()
  const { formations, currentFormation, selectedFormationId, loading, listLoading, total } =
    useAppSelector(state => state.formation)

  const [createModalVisible, setCreateModalVisible] = useState(false)
  const [createForm] = Form.useForm()
  const [uavList, setUavList] = useState<UAVListItem[]>([])
  const [lightConfig, setLightConfig] = useState<FormationLightConfig>({
    red: 0,
    green: 255,
    blue: 0,
    effect: LightEffect.STATIC
  })
  const [lightEnabled, setLightEnabled] = useState(false)

  useEffect(() => {
    dispatch(fetchFormationList({ page: 1, pageSize: 20 }))
    loadUAVList()
  }, [dispatch])

  const loadUAVList = async () => {
    try {
      const result = await getUAVList({ page: 1, pageSize: 50 })
      setUavList(result.list)
    } catch (error) {
      console.error('Failed to load UAV list:', error)
    }
  }

  const handleFormationSelect = (value: string) => {
    dispatch(selectFormation(value))
    if (value) {
      dispatch(fetchFormationDetail(value))
    }
  }

  const handleCreateFormation = async () => {
    try {
      const values = await createForm.validateFields()
      await dispatch(
        createNewFormation({
          name: values.name,
          type: values.type,
          spacing: values.spacing || 5,
          description: values.description,
          uavIds: values.uavIds || []
        })
      ).unwrap()
      message.success('编队创建成功')
      setCreateModalVisible(false)
      createForm.resetFields()
      dispatch(fetchFormationList({ page: 1, pageSize: 20 }))
    } catch (error) {
      message.error(error instanceof Error ? error.message : '创建编队失败')
    }
  }

  const handleStart = () => {
    if (selectedFormationId) {
      dispatch(startFormationById(selectedFormationId))
      message.success('编队已启动')
    }
  }

  const handlePause = () => {
    if (selectedFormationId) {
      dispatch(pauseFormationById(selectedFormationId))
      message.success('编队已暂停')
    }
  }

  const handleResume = () => {
    if (selectedFormationId) {
      dispatch(resumeFormationById(selectedFormationId))
      message.success('编队已恢复')
    }
  }

  const handleStop = () => {
    if (selectedFormationId) {
      dispatch(stopFormationById(selectedFormationId))
      message.success('编队已停止')
    }
  }

  const handleLightApply = async () => {
    if (!selectedFormationId) return
    try {
      await setFormationLight(selectedFormationId, lightConfig)
      message.success('灯光配置已下发')
    } catch (error) {
      message.error('灯光配置下发失败')
    }
  }

  const uavPositions = useMemo(() => {
    if (!currentFormation?.members?.length) return []

    const members = currentFormation.members
    const type = currentFormation.type
    const spacing = currentFormation.spacing || 5

    const positions: { x: number; y: number; isLeader: boolean; name: string; index: number }[] = []

    const scale = 0.8
    const centerX = 50
    const centerY = 50

    if (type === FormationType.LINE) {
      const totalWidth = (members.length - 1) * spacing * scale
      const startX = centerX - totalWidth / 2
      members.forEach((member, i) => {
        positions.push({
          x: startX + i * spacing * scale,
          y: centerY,
          isLeader: member.isLeader,
          name: member.uav?.name || `UAV-${i + 1}`,
          index: i
        })
      })
    } else if (type === FormationType.TRIANGLE) {
      let row = 0
      let countInRow = 1
      let placed = 0

      while (placed < members.length) {
        const rowWidth = (countInRow - 1) * spacing * scale
        const startX = centerX - rowWidth / 2
        const rowY = centerY - row * spacing * scale * 0.866

        for (let i = 0; i < countInRow && placed < members.length; i++) {
          positions.push({
            x: startX + i * spacing * scale,
            y: rowY,
            isLeader: members[placed].isLeader,
            name: members[placed].uav?.name || `UAV-${placed + 1}`,
            index: placed
          })
          placed++
        }
        row++
        countInRow++
      }
    } else if (type === FormationType.CIRCLE) {
      const radius = (members.length * spacing * scale) / (2 * Math.PI)
      const r = Math.min(radius, 35)
      members.forEach((member, i) => {
        const angle = (2 * Math.PI * i) / members.length - Math.PI / 2
        positions.push({
          x: centerX + r * Math.cos(angle),
          y: centerY + r * Math.sin(angle),
          isLeader: member.isLeader,
          name: member.uav?.name || `UAV-${i + 1}`,
          index: i
        })
      })
    }

    return positions
  }, [currentFormation])

  const collisionWarnings = useMemo(() => {
    if (!currentFormation?.members) return []
    const warnings: { id: string; uav1: string; uav2: string; distance: number; level: string }[] = []

    const members = currentFormation.members
    for (let i = 0; i < members.length; i++) {
      for (let j = i + 1; j < members.length; j++) {
        const dx = members[i].offsetX - members[j].offsetX
        const dy = members[i].offsetY - members[j].offsetY
        const dz = members[i].offsetZ - members[j].offsetZ
        const distance = Math.sqrt(dx * dx + dy * dy + dz * dz)

        if (distance < 5) {
          warnings.push({
            id: `${members[i].id}-${members[j].id}`,
            uav1: members[i].uav?.name || `UAV-${i + 1}`,
            uav2: members[j].uav?.name || `UAV-${j + 1}`,
            distance: Math.round(distance * 100) / 100,
            level: distance < 3 ? 'critical' : 'warning'
          })
        }
      }
    }
    return warnings
  }, [currentFormation])

  const memberColumns = [
    {
      title: '序号',
      dataIndex: 'positionIndex',
      key: 'positionIndex',
      width: 60,
      render: (_: unknown, __: unknown, index: number) => index + 1
    },
    {
      title: '无人机',
      dataIndex: ['uav', 'name'],
      key: 'name',
      render: (name: string, record: { isLeader: boolean }) => (
        <Space>
          {record.isLeader && <Tag color="blue">长机</Tag>}
          <span style={{ color: '#fff' }}>{name}</span>
        </Space>
      )
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => <Tag color="green">{status || '正常'}</Tag>
    },
    {
      title: '偏移 (X/Y/Z)',
      key: 'offset',
      render: (_: unknown, record: { offsetX: number; offsetY: number; offsetZ: number }) => (
        <span style={{ color: 'rgba(255,255,255,0.7)', fontSize: '12px' }}>
          {record.offsetX.toFixed(1)} / {record.offsetY.toFixed(1)} / {record.offsetZ.toFixed(1)} m
        </span>
      )
    }
  ]

  return (
    <Container>
      <Header>
        <HeaderLeft>
          <Title>
            <TeamOutlined style={{ color: '#1890ff' }} />
            编队控制中心
          </Title>
          <FormationSelector
            placeholder="选择编队"
            value={selectedFormationId}
            onChange={handleFormationSelect}
            allowClear
            options={formations.map(f => ({
              label: f.name,
              value: f.id
            }))}
          />
        </HeaderLeft>
        <HeaderRight>
          {currentFormation && (
            <StatusBadge>
              <Badge status={getFormationStatusColor(currentFormation.status) as any} />
              <span style={{ color: '#fff' }}>{getFormationStatusText(currentFormation.status)}</span>
            </StatusBadge>
          )}
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateModalVisible(true)}>
            创建编队
          </Button>
        </HeaderRight>
      </Header>

      <Content>
        <LeftPanel>
          <VisualizationCard
            title={
              <Space>
                <RocketOutlined />
                编队队形视图
              </Space>
            }
            extra={
              <Space>
                {currentFormation?.status === FormationStatus.IDLE && (
                  <Button type="primary" icon={<PlayCircleOutlined />} onClick={handleStart}>
                    启动
                  </Button>
                )}
                {currentFormation?.status === FormationStatus.EXECUTING && (
                  <Button icon={<PauseCircleOutlined />} onClick={handlePause}>
                    暂停
                  </Button>
                )}
                {currentFormation?.status === FormationStatus.PAUSED && (
                  <Button type="primary" icon={<ReloadOutlined />} onClick={handleResume}>
                    恢复
                  </Button>
                )}
                {(currentFormation?.status === FormationStatus.EXECUTING ||
                  currentFormation?.status === FormationStatus.PAUSED) && (
                  <Button danger icon={<StopOutlined />} onClick={handleStop}>
                    停止
                  </Button>
                )}
              </Space>
            }
          >
            <FormationCanvas>
              <GridOverlay />
              {uavPositions.map((pos, index) => (
                <UAVDot
                  key={index}
                  $x={pos.x}
                  $y={pos.y}
                  $isLeader={pos.isLeader}
                  $active={currentFormation?.status === FormationStatus.EXECUTING}
                >
                  {pos.index + 1}
                  {pos.isLeader && <LeaderLabel>长机</LeaderLabel>}
                </UAVDot>
              ))}
            </FormationCanvas>
          </VisualizationCard>

          <InfoCard title="编队成员" style={{ maxHeight: '300px' }}>
            <Table
              size="small"
              dataSource={currentFormation?.members || []}
              columns={memberColumns}
              pagination={false}
              rowKey="id"
              scroll={{ y: 180 }}
            />
          </InfoCard>
        </LeftPanel>

        <RightPanel>
          <InfoCard title="编队信息">
            <InfoRow>
              <InfoLabel>队形类型</InfoLabel>
              <InfoValue>
                {currentFormation ? getFormationTypeLabel(currentFormation.type) : '-'}
              </InfoValue>
            </InfoRow>
            <InfoRow>
              <InfoLabel>编队间距</InfoLabel>
              <InfoValue>{currentFormation?.spacing || '-'} m</InfoValue>
            </InfoRow>
            <InfoRow>
              <InfoLabel>成员数量</InfoLabel>
              <InfoValue>{currentFormation?.members?.length || 0} 架</InfoValue>
            </InfoRow>
            <InfoRow>
              <InfoLabel>长机</InfoLabel>
              <InfoValue>
                {currentFormation?.members?.find(m => m.isLeader)?.uav?.name || '-'}
              </InfoValue>
            </InfoRow>
            <InfoRow>
              <InfoLabel>位置误差</InfoLabel>
              <InfoValue>
                <Tag color="green">&lt; 0.3m</Tag>
              </InfoValue>
            </InfoRow>
          </InfoCard>

          <LightControlCard
            title={
              <Space>
                <BulbOutlined />
                灯光控制
              </Space>
            }
            extra={<Switch checked={lightEnabled} onChange={setLightEnabled} />}
          >
            <ColorPreview $r={lightConfig.red} $g={lightConfig.green} $b={lightConfig.blue} />
            <Form layout="vertical">
              <Form.Item label="红色 (R)">
                <Slider
                  min={0}
                  max={255}
                  value={lightConfig.red}
                  onChange={value => setLightConfig(prev => ({ ...prev, red: value }))}
                  disabled={!lightEnabled}
                  tooltip={{ formatter: value => `${value}` }}
                />
              </Form.Item>
              <Form.Item label="绿色 (G)">
                <Slider
                  min={0}
                  max={255}
                  value={lightConfig.green}
                  onChange={value => setLightConfig(prev => ({ ...prev, green: value }))}
                  disabled={!lightEnabled}
                  tooltip={{ formatter: value => `${value}` }}
                />
              </Form.Item>
              <Form.Item label="蓝色 (B)">
                <Slider
                  min={0}
                  max={255}
                  value={lightConfig.blue}
                  onChange={value => setLightConfig(prev => ({ ...prev, blue: value }))}
                  disabled={!lightEnabled}
                  tooltip={{ formatter: value => `${value}` }}
                />
              </Form.Item>
              <Form.Item label="灯效模式">
                <Select
                  value={lightConfig.effect}
                  onChange={value => setLightConfig(prev => ({ ...prev, effect: value }))}
                  disabled={!lightEnabled}
                >
                  <Select.Option value={LightEffect.STATIC}>常亮</Select.Option>
                  <Select.Option value={LightEffect.BLINK}>闪烁</Select.Option>
                  <Select.Option value={LightEffect.RAINBOW}>彩虹</Select.Option>
                  <Select.Option value={LightEffect.BREATHING}>呼吸</Select.Option>
                </Select>
              </Form.Item>
              <Button type="primary" block onClick={handleLightApply} disabled={!lightEnabled}>
                应用灯光效果
              </Button>
            </Form>
          </LightControlCard>

          <CollisionWarningCard
            title={
              <Space>
                <WarningOutlined />
                碰撞预警
                {collisionWarnings.length > 0 && (
                  <Badge count={collisionWarnings.length} size="small" />
                )}
              </Space>
            }
          >
            {collisionWarnings.length === 0 ? (
              <div style={{ textAlign: 'center', padding: '16px 0', color: 'rgba(255,255,255,0.5)' }}>
                暂无碰撞预警
              </div>
            ) : (
              collisionWarnings.map(warning => (
                <WarningItem key={warning.id}>
                  <WarningText>
                    {warning.uav1} ↔ {warning.uav2}
                  </WarningText>
                  <Space>
                    <Tag color={warning.level === 'critical' ? 'red' : 'orange'}>
                      {warning.distance}m
                    </Tag>
                    <Badge status={warning.level === 'critical' ? 'error' : 'warning'} />
                  </Space>
                </WarningItem>
              ))
            )}
          </CollisionWarningCard>
        </RightPanel>
      </Content>

      <Modal
        title="创建编队"
        open={createModalVisible}
        onCancel={() => setCreateModalVisible(false)}
        onOk={handleCreateFormation}
        okText="创建"
        cancelText="取消"
      >
        <Form form={createForm} layout="vertical">
          <Form.Item
            label="编队名称"
            name="name"
            rules={[{ required: true, message: '请输入编队名称' }]}
          >
            <Input placeholder="请输入编队名称" />
          </Form.Item>
          <Form.Item
            label="队形类型"
            name="type"
            rules={[{ required: true, message: '请选择队形类型' }]}
          >
            <Select placeholder="请选择队形类型">
              <Select.Option value={FormationType.LINE}>一字型</Select.Option>
              <Select.Option value={FormationType.TRIANGLE}>三角形</Select.Option>
              <Select.Option value={FormationType.CIRCLE}>圆形</Select.Option>
            </Select>
          </Form.Item>
          <Form.Item label="编队间距 (米)" name="spacing" initialValue={5}>
            <InputNumber min={1} max={50} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item label="选择无人机" name="uavIds">
            <Select
              mode="multiple"
              placeholder="请选择无人机"
              options={uavList.map(uav => ({
                label: uav.name,
                value: uav.id
              }))}
            />
          </Form.Item>
          <Form.Item label="描述" name="description">
            <Input.TextArea rows={3} placeholder="请输入编队描述" />
          </Form.Item>
        </Form>
      </Modal>
    </Container>
  )
}

export default FormationMonitor
