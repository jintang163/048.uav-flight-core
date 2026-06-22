import React, { useEffect, useMemo } from 'react'
import {
  Card,
  Table,
  Tag,
  Button,
  Space,
  Statistic,
  Row,
  Col,
  Switch,
  Divider,
  List,
  Tooltip,
  Alert,
  Typography,
  Popover,
} from 'antd'
import {
  WarningOutlined,
  ThunderboltOutlined,
  SafetyOutlined,
  RocketOutlined,
  PauseCircleOutlined,
  CheckCircleOutlined,
  ReloadOutlined,
  CarOutlined,
} from '@ant-design/icons'
import { useAppSelector, useAppDispatch } from '@/store'
import {
  fetchCollisionStatus,
  fetchActiveAlerts,
  fetchIntersections,
  setCollisionEnabled,
  dismissAlert,
} from '@/store/slices/collision'
import { initCollisionWebSocket } from '@/websocket/collision'
import type { CollisionAlert, CollisionRiskLevel, RouteIntersection } from '@/types'

const { Text, Title } = Typography

const riskLevelConfig: Record<CollisionRiskLevel, { color: string; label: string; icon: React.ReactNode }> = {
  safe: { color: 'green', label: '安全', icon: <SafetyOutlined /> },
  warning: { color: 'orange', label: '警告', icon: <WarningOutlined /> },
  critical: { color: 'red', label: '严重', icon: <ThunderboltOutlined /> },
  avoiding: { color: 'purple', label: '避让中', icon: <PauseCircleOutlined /> },
  resolved: { color: 'blue', label: '已解除', icon: <CheckCircleOutlined /> },
}

const actionLabel: Record<string, string> = {
  speed_reduce: '减速避让',
  speed_adjust: '速度调整',
  hold_position: '悬停等待',
  waypoint_hold: '航点等待',
  altitude_change: '高度调整',
  resume: '恢复正常',
}

const CollisionAvoidancePanel: React.FC = () => {
  const dispatch = useAppDispatch()
  const { status, alerts, intersections, enabled, loading } = useAppSelector(state => state.collision)

  useEffect(() => {
    initCollisionWebSocket()
    dispatch(fetchCollisionStatus())
    dispatch(fetchActiveAlerts())
    dispatch(fetchIntersections())
  }, [dispatch])

  const handleToggle = (checked: boolean) => {
    dispatch(setCollisionEnabled(checked))
  }

  const handleRefresh = () => {
    dispatch(fetchCollisionStatus())
    dispatch(fetchActiveAlerts())
    dispatch(fetchIntersections())
  }

  const handleDismiss = (id: number) => {
    dispatch(dismissAlert(id))
  }

  const criticalAlerts = useMemo(() =>
    alerts.filter(a => a.risk_level === 'critical' || a.risk_level === 'avoiding'),
    [alerts]
  )

  const alertColumns = [
    {
      title: '预警级别',
      dataIndex: 'risk_level',
      key: 'risk_level',
      width: 100,
      render: (level: CollisionRiskLevel) => {
        const cfg = riskLevelConfig[level] || riskLevelConfig.warning
        return (
          <Tag color={cfg.color} icon={cfg.icon}>
            {cfg.label}
          </Tag>
        )
      },
    },
    {
      title: '无人机',
      key: 'uavs',
      width: 160,
      render: (_: unknown, record: CollisionAlert) => (
        <Space direction="vertical" size={0}>
          <Text type="secondary">无人机 #{record.uav_id_1}</Text>
          <Text type="secondary">无人机 #{record.uav_id_2}</Text>
        </Space>
      ),
    },
    {
      title: '距离',
      dataIndex: 'current_distance',
      key: 'current_distance',
      width: 100,
      render: (dist: number) => `${dist.toFixed(1)}m`,
    },
    {
      title: 'TTC',
      dataIndex: 'time_to_collision',
      key: 'ttc',
      width: 80,
      render: (ttc: number) => (ttc > 0 ? `${ttc.toFixed(1)}s` : '-'),
    },
    {
      title: '避让动作',
      dataIndex: 'action_taken',
      key: 'action',
      width: 120,
      render: (action: string) => (
        <Tag color="purple">{actionLabel[action] || action}</Tag>
      ),
    },
    {
      title: '详情',
      dataIndex: 'action_detail',
      key: 'detail',
      ellipsis: true,
    },
    {
      title: '操作',
      key: 'action',
      width: 100,
      render: (_: unknown, record: CollisionAlert) => (
        <Button
          type="link"
          size="small"
          onClick={() => handleDismiss(record.id)}
        >
          解除
        </Button>
      ),
    },
  ]

  const intersectionColumns = [
    {
      title: '风险等级',
      dataIndex: 'risk_level',
      key: 'risk_level',
      width: 100,
      render: (level: CollisionRiskLevel) => {
        const cfg = riskLevelConfig[level] || riskLevelConfig.safe
        return (
          <Tag color={cfg.color} icon={cfg.icon}>
            {cfg.label}
          </Tag>
        )
      },
    },
    {
      title: '无人机',
      key: 'uavs',
      width: 140,
      render: (_: unknown, record: RouteIntersection) => (
        <Space direction="vertical" size={0}>
          <Text type="secondary">#{record.uav_id_1} 航点{record.waypoint_seq_1}</Text>
          <Text type="secondary">#{record.uav_id_2} 航点{record.waypoint_seq_2}</Text>
        </Space>
      ),
    },
    {
      title: '间距',
      dataIndex: 'distance_m',
      key: 'distance',
      width: 80,
      render: (d: number) => `${d.toFixed(1)}m`,
    },
    {
      title: '时间差',
      dataIndex: 'time_diff_sec',
      key: 'time_diff',
      width: 90,
      render: (t: number) => `${t.toFixed(1)}s`,
    },
    {
      title: '位置',
      key: 'pos',
      render: (_: unknown, record: RouteIntersection) => (
        <Text type="secondary" ellipsis>
          {record.latitude.toFixed(6)}, {record.longitude.toFixed(6)}
        </Text>
      ),
    },
  ]

  const alertContent = (
    <div style={{ width: 280 }}>
      <List
        size="small"
        dataSource={criticalAlerts.slice(0, 5)}
        renderItem={alert => (
          <List.Item>
            <List.Item.Meta
              avatar={
                <ThunderboltOutlined style={{ color: '#ff4d4f', fontSize: 18 }} />
              }
              title={<Text strong>碰撞风险 {riskLevelConfig[alert.risk_level]?.label}</Text>}
              description={
                <>
                  <div>#{alert.uav_id_1} ↔ #{alert.uav_id_2}: {alert.current_distance.toFixed(1)}m</div>
                  <div>{alert.action_detail}</div>
                </>
              }
            />
          </List.Item>
        )}
      />
    </div>
  )

  return (
    <div style={{ padding: 16 }}>
      <Row gutter={[16, 16]}>
        <Col span={24}>
          <Card
            title={
              <Space>
                <CarOutlined />
                <span>多机协同避让</span>
              </Space>
            }
            extra={
              <Space>
                <Switch
                  checked={enabled}
                  onChange={handleToggle}
                  checkedChildren="开启"
                  unCheckedChildren="关闭"
                />
                <Button
                  type="text"
                  icon={<ReloadOutlined />}
                  onClick={handleRefresh}
                />
              </Space>
            }
          >
            <Row gutter={16}>
              <Col span={6}>
                <Statistic
                  title="在线无人机"
                  value={status.active_uavs}
                  prefix={<RocketOutlined />}
                  valueStyle={{ color: '#1890ff' }}
                />
              </Col>
              <Col span={6}>
                <Popover
                  content={alertContent}
                  title="碰撞预警详情"
                  trigger="hover"
                  disabled={criticalAlerts.length === 0}
                >
                  <Statistic
                    title="活跃预警"
                    value={status.active_alerts}
                    prefix={<WarningOutlined />}
                    valueStyle={{ color: status.active_alerts > 0 ? '#fa8c16' : '#52c41a' }}
                  />
                </Popover>
              </Col>
              <Col span={6}>
                <Statistic
                  title="航路交叉"
                  value={status.intersections}
                  prefix={<CarOutlined />}
                  valueStyle={{ color: status.intersections > 0 ? '#1890ff' : '#52c41a' }}
                />
              </Col>
              <Col span={6}>
                <Statistic
                  title="安全距离"
                  value={status.safe_distance_m}
                  suffix="m"
                  prefix={<SafetyOutlined />}
                  valueStyle={{ color: '#52c41a' }}
                />
              </Col>
            </Row>
          </Card>
        </Col>

        {criticalAlerts.length > 0 && (
          <Col span={24}>
            <Alert
              type="error"
              showIcon
              icon={<ThunderboltOutlined />}
              message={`检测到 ${criticalAlerts.length} 个严重碰撞风险`}
              description="系统已自动下发避让指令，请密切关注无人机状态"
              banner
            />
          </Col>
        )}

        <Col span={24}>
          <Card title="碰撞预警" size="small">
            <Table
              size="small"
              dataSource={alerts}
              columns={alertColumns}
              rowKey="id"
              loading={loading}
              pagination={false}
              locale={{ emptyText: '暂无碰撞预警，无人机安全飞行中' }}
            />
          </Card>
        </Col>

        <Col span={24}>
          <Card
            title="航路交叉检测"
            size="small"
            extra={
              <Button
                type="primary"
                size="small"
                onClick={() => dispatch(fetchIntersections())}
              >
                重新检测
              </Button>
            }
          >
            <Table
              size="small"
              dataSource={intersections}
              columns={intersectionColumns}
              rowKey="id"
              pagination={false}
              locale={{ emptyText: '未检测到航路交叉点' }}
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default CollisionAvoidancePanel
