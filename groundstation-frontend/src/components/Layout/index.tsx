import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import {
  Layout,
  Menu,
  Avatar,
  Dropdown,
  Button,
  Space,
  Tag,
  Badge,
  Drawer,
  List,
  Tooltip
} from 'antd'
import {
  DashboardOutlined,
  RocketOutlined,
  ApiOutlined,
  HistoryOutlined,
  SafetyOutlined,
  BellOutlined,
  CloudUploadOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
  MenuUnfoldOutlined,
  MenuFoldOutlined,
  WifiOutlined,
  WifiOffOutlined,
  NotificationOutlined,
  BulbOutlined,
  SunOutlined,
  TeamOutlined
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '@/store'
import { logout } from '@/store/slices/auth'
import { useTheme } from '@/hooks/useTheme'
import { useAlert } from '@/hooks/useAlert'
import { useUAV } from '@/hooks/useUAV'
import { useWebSocket } from '@/hooks/useWebSocket'
import { formatDateTime } from '@/utils'

const { Header, Sider, Content } = Layout

const LayoutContainer = styled(Layout)`
  width: 100%;
  height: 100vh;
  overflow: hidden;
  background: #0f172a;
`

const SiderContainer = styled(Sider)`
  background: #1e293b;
  border-right: 1px solid rgba(255, 255, 255, 0.05);

  .ant-layout-sider-trigger {
    background: #1e293b;
    border-top: 1px solid rgba(255, 255, 255, 0.05);
    color: rgba(255, 255, 255, 0.7);

    &:hover {
      background: #334155;
    }
  }
`

const Logo = styled.div`
  height: 64px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding: 0 16px;
`

const LogoIcon = styled.div`
  width: 36px;
  height: 36px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #1890ff 0%, #52c41a 100%);
  border-radius: 8px;
  font-size: 18px;
  color: #fff;
`

const LogoText = styled.div<{ collapsed: boolean }>`
  font-size: 16px;
  font-weight: 700;
  color: #fff;
  white-space: nowrap;
  overflow: hidden;
  opacity: ${props => props.collapsed ? 0 : 1};
  transition: opacity 0.3s;
`

const HeaderContainer = styled(Header)`
  background: #1e293b;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);
  padding: 0 24px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 64px;
`

const HeaderLeft = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
`

const HeaderRight = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
`

const ContentContainer = styled(Content)`
  background: #0f172a;
  overflow: hidden;
`

const MenuContainer = styled(Menu)`
  background: transparent;
  border-right: none;
  margin-top: 8px;

  .ant-menu-item {
    margin: 4px 8px;
    border-radius: 6px;

    &:hover {
      background: rgba(24, 144, 255, 0.1) !important;
    }

    &.ant-menu-item-selected {
      background: linear-gradient(135deg, rgba(24, 144, 255, 0.3) 0%, rgba(82, 196, 26, 0.3) 100%) !important;
    }
  }

  .ant-menu-title-content {
    color: rgba(255, 255, 255, 0.7);
  }

  .ant-menu-item-selected .ant-menu-title-content {
    color: #fff;
  }

  .ant-menu-item-icon {
    color: rgba(255, 255, 255, 0.7);
  }

  .ant-menu-item-selected .ant-menu-item-icon {
    color: #1890ff;
  }
`

const StatusIndicator = styled.div<{ connected: boolean }>`
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  background: ${props => props.connected ? 'rgba(82, 196, 26, 0.1)' : 'rgba(255, 77, 79, 0.1)'};
  border-radius: 16px;
  color: ${props => props.connected ? '#52c41a' : '#ff4d4f'};
  font-size: 12px;
`

const UserInfo = styled.div`
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 4px 12px 4px 4px;
  border-radius: 20px;
  cursor: pointer;
  transition: background 0.3s;

  &:hover {
    background: rgba(255, 255, 255, 0.05);
  }
`

const UserName = styled.div`
  font-size: 14px;
  font-weight: 500;
  color: rgba(255, 255, 255, 0.9);
`

const NotificationBadge = styled(Badge)`
  .ant-badge-count {
    background: #ff4d4f;
    box-shadow: 0 0 0 1px #0f172a;
  }
`

const PageTitle = styled.div`
  font-size: 18px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
`

const UAVSelector = styled.div`
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 6px 12px;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.3s;

  &:hover {
    background: rgba(255, 255, 255, 0.1);
  }
`

const MainLayout: React.FC = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const dispatch = useAppDispatch()
  const { user } = useAppSelector(state => state.auth)
  const { isDark, toggleTheme } = useTheme()
  const { alerts, unreadCount, resetUnreadCount: markAllAsRead } = useAlert()
  const { uavList, selectedUAVId, currentUAV, selectCurrentUAV: selectUAV } = useUAV()
  const { isConnected } = useWebSocket()
  
  const [collapsed, setCollapsed] = useState(false)
  const [notificationVisible, setNotificationVisible] = useState(false)
  const [selectedKey, setSelectedKey] = useState<string>('')

  useEffect(() => {
    const pathname = location.pathname
    const key = pathname.split('/')[1] || 'dashboard'
    setSelectedKey(key)
  }, [location.pathname])

  const menuItems = [
    {
      key: 'dashboard',
      icon: <DashboardOutlined />,
      label: '主控制台'
    },
    {
      key: 'mission',
      icon: <RocketOutlined />,
      label: '航线规划'
    },
    {
      key: 'uav-list',
      icon: <ApiOutlined />,
      label: '无人机列表'
    },
    {
      key: 'flight-history',
      icon: <HistoryOutlined />,
      label: '飞行记录'
    },
    {
      key: 'geofence',
      icon: <SafetyOutlined />,
      label: '电子围栏'
    },
    {
      key: 'formation',
      icon: <TeamOutlined />,
      label: '编队控制'
    },
    {
      key: 'alert-center',
      icon: <BellOutlined />,
      label: '告警中心'
    },
    {
      key: 'firmware',
      icon: <CloudUploadOutlined />,
      label: '固件管理'
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置'
    }
  ]

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(`/${key}`)
  }

  const handleLogout = () => {
    dispatch(logout())
    navigate('/login')
  }

  const userMenuItems = [
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '系统设置',
      onClick: () => navigate('/settings')
    },
    {
      type: 'divider' as const
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: handleLogout
    }
  ]

  const uavMenuItems = uavList.map((uav: any) => ({
    key: uav.id,
    label: (
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <span>{uav.name}</span>
        <Tag color={uav.status !== 'disconnected' && uav.status !== 'error' ? '#52c41a' : '#d9d9d9'}>
          {uav.status !== 'disconnected' && uav.status !== 'error' ? '在线' : '离线'}
        </Tag>
      </div>
    )
  }))

  const pageTitleMap: Record<string, string> = {
    'dashboard': '主控制台',
    'mission': '航线规划',
    'uav-list': '无人机列表',
    'flight-history': '飞行记录',
    'geofence': '电子围栏',
    'alert-center': '告警中心',
    'firmware': '固件管理',
    'settings': '系统设置'
  }

  return (
    <LayoutContainer>
      <SiderContainer
        trigger={null}
        collapsible
        collapsed={collapsed}
        width={240}
        collapsedWidth={64}
      >
        <Logo>
          <LogoIcon>
            <RocketOutlined />
          </LogoIcon>
          <LogoText collapsed={collapsed}>无人机地面站</LogoText>
        </Logo>

        <MenuContainer
          mode="inline"
          selectedKeys={[selectedKey]}
          items={menuItems}
          onClick={handleMenuClick}
        />
      </SiderContainer>

      <Layout>
        <HeaderContainer>
          <HeaderLeft>
            <Button
              type="text"
              icon={collapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setCollapsed(!collapsed)}
              style={{ color: 'rgba(255,255,255,0.7)', fontSize: '16px' }}
            />
            <PageTitle>{pageTitleMap[selectedKey] || '主控制台'}</PageTitle>
          </HeaderLeft>

          <HeaderRight>
            <StatusIndicator connected={isConnected}>
              {isConnected ? <WifiOutlined /> : <WifiOffOutlined />}
              {isConnected ? '已连接' : '未连接'}
            </StatusIndicator>

            <Dropdown
              menu={{ items: uavMenuItems, onClick: ({ key }) => selectUAV(key) }}
              placement="bottomRight"
            >
              <UAVSelector>
                <ApiOutlined style={{ color: '#1890ff' }} />
                <span style={{ color: 'rgba(255,255,255,0.9)' }}>
                  {currentUAV?.name || '选择无人机'}
                </span>
              </UAVSelector>
            </Dropdown>

            <Tooltip title="主题切换">
              <Button
                type="text"
                icon={isDark ? <BulbOutlined /> : <SunOutlined />}
                onClick={toggleTheme}
                style={{ color: 'rgba(255,255,255,0.7)' }}
              />
            </Tooltip>

            <Tooltip title="告警通知">
              <Button
                type="text"
                icon={<NotificationBadge count={unreadCount} size="small">
                  <BellOutlined style={{ fontSize: '18px', color: 'rgba(255,255,255,0.7)' }} />
                </NotificationBadge>}
                onClick={() => setNotificationVisible(true)}
                style={{ color: 'rgba(255,255,255,0.7)' }}
              />
            </Tooltip>

            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <UserInfo>
                <Avatar size={32} icon={<UserOutlined />} style={{ background: 'linear-gradient(135deg, #1890ff 0%, #52c41a 100%)' }} />
                <UserName>{user?.nickname || user?.username || '用户'}</UserName>
              </UserInfo>
            </Dropdown>
          </HeaderRight>
        </HeaderContainer>

        <ContentContainer>
          <Outlet />
        </ContentContainer>
      </Layout>

      <Drawer
        title="告警通知"
        placement="right"
        open={notificationVisible}
        onClose={() => setNotificationVisible(false)}
        width={360}
        extra={
          <Button
            size="small"
            type="link"
            onClick={() => {
              markAllAsRead()
              setNotificationVisible(false)
            }}
          >
            全部已读
          </Button>
        }
      >
        <List
          dataSource={alerts.slice(0, 20)}
          renderItem={(item) => (
            <List.Item
              style={{
                padding: '12px 0',
                borderBottom: '1px solid rgba(255,255,255,0.05)',
                background: item.status === 'active' ? 'rgba(24, 144, 255, 0.05)' : 'transparent'
              }}
            >
              <List.Item.Meta
                avatar={
                  <Badge status={item.severity === 'critical' ? 'error' : item.severity === 'warning' ? 'warning' : 'info'} />
                }
                title={
                  <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                    <span style={{ color: 'rgba(255,255,255,0.9)', fontWeight: item.status === 'active' ? 600 : 400 }}>
                      {item.title}
                    </span>
                    <span style={{ fontSize: '12px', color: 'rgba(255,255,255,0.5)' }}>
                      {formatDateTime(item.createdAt)}
                    </span>
                  </div>
                }
                description={
                  <span style={{ color: 'rgba(255,255,255,0.6)' }}>
                    {item.message}
                  </span>
                }
              />
            </List.Item>
          )}
          locale={{ emptyText: '暂无告警' }}
        />
      </Drawer>
    </LayoutContainer>
  )
}

export default MainLayout