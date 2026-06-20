import React, { useState } from 'react'
import styled from 'styled-components'
import {
  Card,
  Form,
  Input,
  Select,
  Switch,
  Button,
  InputNumber,
  Tabs,
  message,
  Divider,
  Row,
  Col,
  Avatar,
  Space,
  Tag,
  Modal
} from 'antd'
import {
  SettingOutlined,
  UserOutlined,
  SafetyOutlined,
  MapOutlined,
  BellOutlined,
  GlobalOutlined,
  CloudOutlined,
  BulbOutlined,
  InfoCircleOutlined,
  LogoutOutlined,
  EditOutlined,
  SoundOutlined,
  AudioOutlined,
  DesktopOutlined,
  NotificationOutlined,
  ClockCircleOutlined
} from '@ant-design/icons'
import { useAppDispatch, useAppSelector } from '@/store'
import { logout } from '@/store/slices/auth'
import { useTheme } from '@/hooks/useTheme'
import { changePassword } from '@/api/auth'
import type { SystemSettings } from '@/types'

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
  font-size: 18px;
  font-weight: 600;
  display: flex;
  align-items: center;
  gap: 10px;
`

const Content = styled.div`
  flex: 1;
  overflow: auto;
  padding-right: 8px;

  &::-webkit-scrollbar {
    width: 6px;
  }

  &::-webkit-scrollbar-track {
    background: rgba(255, 255, 255, 0.05);
    border-radius: 3px;
  }

  &::-webkit-scrollbar-thumb {
    background: rgba(255, 255, 255, 0.2);
    border-radius: 3px;
  }
`

const SettingsCard = styled(Card)`
  background: rgba(255, 255, 255, 0.03);
  border: 1px solid rgba(255, 255, 255, 0.1);
  margin-bottom: 16px;

  .ant-card-head {
    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
  }

  .ant-card-body {
    padding: 24px;
  }
`

const SectionTitle = styled.div`
  font-size: 14px;
  font-weight: 600;
  margin-bottom: 16px;
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgba(255, 255, 255, 0.9);
`

const SettingRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 12px 0;
  border-bottom: 1px solid rgba(255, 255, 255, 0.05);

  &:last-child {
    border-bottom: none;
  }
`

const SettingLabel = styled.div`
  display: flex;
  align-items: center;
  gap: 12px;
`

const SettingIcon = styled.div`
  width: 32px;
  height: 32px;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.05);
  border-radius: 6px;
  color: #1890ff;
`

const SettingInfo = styled.div``

const SettingName = styled.div`
  font-weight: 500;
  margin-bottom: 2px;
`

const SettingDesc = styled.div`
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
`

const UserProfile = styled.div`
  display: flex;
  align-items: center;
  gap: 16px;
  padding: 24px;
  background: linear-gradient(135deg, rgba(24, 144, 255, 0.1) 0%, rgba(82, 196, 26, 0.1) 100%);
  border-radius: 12px;
  margin-bottom: 24px;
`

const UserAvatar = styled(Avatar)`
  width: 64px;
  height: 64px;
  background: linear-gradient(135deg, #1890ff 0%, #52c41a 100%);
`

const UserInfo = styled.div`
  flex: 1;
`

const UserName = styled.div`
  font-size: 20px;
  font-weight: 700;
  margin-bottom: 4px;
`

const UserRoles = styled.div`
  display: flex;
  gap: 8px;
  margin-top: 8px;
`

const Settings: React.FC = () => {
  const dispatch = useAppDispatch()
  const { user } = useAppSelector(state => state.auth)
  const { theme, toggleTheme } = useTheme()
  const [passwordForm] = Form.useForm()
  const [settings, setSettings] = useState<SystemSettings>({
    theme: 'dark',
    language: 'zh-CN',
    mapType: 'standard',
    unitSystem: 'metric',
    autoRefresh: true,
    refreshInterval: 1000,
    notificationSettings: {
      soundEnabled: true,
      voiceEnabled: true,
      notificationEnabled: true,
      desktopNotification: true
    }
  })
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)

  const handleSettingChange = (key: keyof SystemSettings, value: any) => {
    setSettings(prev => ({ ...prev, [key]: value }))
    
    if (key === 'theme') {
      toggleTheme()
    }
    
    message.success('设置已保存')
  }

  const handlePasswordChange = async (values: any) => {
    try {
      await changePassword({
        oldPassword: values.oldPassword,
        newPassword: values.newPassword
      })
      message.success('密码修改成功')
      setPasswordModalVisible(false)
      passwordForm.resetFields()
    } catch (error) {
      message.error('密码修改失败')
    }
  }

  const handleLogout = () => {
    Modal.confirm({
      title: '确认退出',
      content: '确定要退出登录吗？',
      okText: '确认退出',
      cancelText: '取消',
      onOk: () => {
        dispatch(logout())
        message.success('已退出登录')
      }
    })
  }

  const handleSaveSettings = () => {
    localStorage.setItem('systemSettings', JSON.stringify(settings))
    message.success('设置已保存')
  }

  const handleResetSettings = () => {
    Modal.confirm({
      title: '重置设置',
      content: '确定要重置所有设置为默认值吗？',
      onOk: () => {
        const defaultSettings: SystemSettings = {
          theme: 'dark',
          language: 'zh-CN',
          mapType: 'standard',
          unitSystem: 'metric',
          autoRefresh: true,
          refreshInterval: 1000,
          notificationSettings: {
            soundEnabled: true,
            voiceEnabled: true,
            notificationEnabled: true,
            desktopNotification: true
          }
        }
        setSettings(defaultSettings)
        localStorage.setItem('systemSettings', JSON.stringify(defaultSettings))
        message.success('设置已重置')
      }
    })
  }

  return (
    <Container>
      <Header>
        <SettingOutlined style={{ color: '#1890ff' }} />
        系统设置
      </Header>

      <Content>
        {user && (
          <UserProfile>
            <UserAvatar size={64} icon={<UserOutlined />} />
            <UserInfo>
              <UserName>{user.nickname || user.username}</UserName>
              <div style={{ color: 'rgba(255,255,255,0.6)' }}>
                @{user.username}
              </div>
              <UserRoles>
                {user.roles.map((role: string, index: number) => (
                  <Tag key={index} color="blue">{role}</Tag>
                ))}
              </UserRoles>
            </UserInfo>
            <Space direction="vertical">
              <Button
                icon={<EditOutlined />}
                onClick={() => setPasswordModalVisible(true)}
              >
                修改密码
              </Button>
              <Button
                danger
                icon={<LogoutOutlined />
                onClick={handleLogout}
              >
                退出登录
              </Button>
            </Space>
          </UserProfile>
        )}

        <Tabs
          defaultActiveKey="general"
          items={[
            {
              key: 'general',
            label: '通用设置',
              icon: <SettingOutlined />,
              children: (
                <SettingsCard>
                  <SectionTitle>
                    <BulbOutlined />
                    显示设置
                  </SectionTitle>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <BulbOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>主题模式</SettingName>
                        <SettingDesc>切换深色/浅色主题</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.theme === 'dark'}
                      checkedChildren="深色"
                      unCheckedChildren="浅色"
                      onChange={(checked) => handleSettingChange('theme', checked ? 'dark' : 'light')}
                    />
                  </SettingRow>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <GlobalOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>语言</SettingName>
                        <SettingDesc>选择界面语言</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Select
                      value={settings.language}
                      onChange={(value) => handleSettingChange('language', value)}
                      style={{ width: 140 }}
                    >
                      <Select.Option value="zh-CN">简体中文</Select.Option>
                      <Select.Option value="en-US">English</Select.Option>
                    </Select>
                  </SettingRow>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <SafetyOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>单位系统</SettingName>
                        <SettingDesc>选择度量单位</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Select
                      value={settings.unitSystem}
                      onChange={(value) => handleSettingChange('unitSystem', value)}
                      style={{ width: 140 }}
                    >
                      <Select.Option value="metric">公制 (米/公里)</Select.Option>
                      <Select.Option value="imperial">英制 (英尺/英里)</Select.Option>
                    </Select>
                  </SettingRow>

                  <Divider style={{ margin: '16px 0' }} />

                  <SectionTitle>
                    <MapOutlined />
                    地图设置
                  </SectionTitle>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <MapOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>地图类型</SettingName>
                        <SettingDesc>选择地图显示样式</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Select
                      value={settings.mapType}
                      onChange={(value) => handleSettingChange('mapType', value)}
                      style={{ width: 140 }}
                    >
                      <Select.Option value="standard">标准地图</Select.Option>
                      <Select.Option value="satellite">卫星地图</Select.Option>
                      <Select.Option value="hybrid">混合地图</Select.Option>
                    </Select>
                  </SettingRow>

                  <Divider style={{ margin: '16px 0' }} />

                  <SectionTitle>
                    <CloudOutlined />
                    数据更新
                  </SectionTitle>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <CloudOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>自动刷新</SettingName>
                        <SettingDesc>自动刷新数据</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.autoRefresh}
                      onChange={(checked) => handleSettingChange('autoRefresh', checked)}
                    />
                  </SettingRow>
                  {settings.autoRefresh && (
                    <SettingRow>
                      <SettingLabel>
                        <SettingIcon>
                          <ClockCircleOutlined />
                        </SettingIcon>
                        <SettingInfo>
                          <SettingName>刷新间隔</SettingName>
                          <SettingDesc>数据刷新间隔（毫秒）</SettingDesc>
                        </SettingInfo>
                      </SettingLabel>
                      <InputNumber
                        min={500}
                        max={5000}
                        step={100}
                        value={settings.refreshInterval}
                        onChange={(value) => handleSettingChange('refreshInterval', value)}
                      />
                    </SettingRow>
                  )}
                </SettingsCard>
              )
            },
            {
              key: 'notifications',
              label: '通知设置',
              icon: <BellOutlined />,
              children: (
                <SettingsCard>
                  <SectionTitle>
                    <BellOutlined />
                    告警通知
                  </SectionTitle>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <SoundOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>告警音效</SettingName>
                        <SettingDesc>播放告警提示音</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.notificationSettings.soundEnabled}
                      onChange={(checked) => handleSettingChange('notificationSettings', {
                        ...settings.notificationSettings,
                        soundEnabled: checked
                      })}
                    />
                  </SettingRow>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <AudioOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>语音播报</SettingName>
                        <SettingDesc>语音播报告警内容</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.notificationSettings.voiceEnabled}
                      onChange={(checked) => handleSettingChange('notificationSettings', {
                        ...settings.notificationSettings,
                        voiceEnabled: checked
                      })}
                    />
                  </SettingRow>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <DesktopOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>桌面通知</SettingName>
                        <SettingDesc>显示桌面通知</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.notificationSettings.desktopNotification}
                      onChange={(checked) => handleSettingChange('notificationSettings', {
                        ...settings.notificationSettings,
                        desktopNotification: checked
                      })}
                    />
                  </SettingRow>
                  <SettingRow>
                    <SettingLabel>
                      <SettingIcon>
                        <NotificationOutlined />
                      </SettingIcon>
                      <SettingInfo>
                        <SettingName>应用内通知</SettingName>
                        <SettingDesc>显示应用内通知弹窗</SettingDesc>
                      </SettingInfo>
                    </SettingLabel>
                    <Switch
                      checked={settings.notificationSettings.notificationEnabled}
                      onChange={(checked) => handleSettingChange('notificationSettings', {
                        ...settings.notificationSettings,
                        notificationEnabled: checked
                      })}
                    />
                  </SettingRow>
                </SettingsCard>
              )
            },
            {
              key: 'about',
            label: '关于',
              icon: <InfoCircleOutlined />,
              children: (
                <SettingsCard>
                  <SectionTitle>
                    <InfoCircleOutlined />
                    系统信息
                  </SectionTitle>
                  <Row gutter={24}>
                    <Col span={12}>
                      <SettingRow>
                        <SettingLabel>
                          <SettingInfo>
                            <SettingName>应用名称</SettingName>
                          </SettingInfo>
                        </SettingLabel>
                        <span style={{ color: 'rgba(255,255,255,0.7)' }}>
                          无人机地面站
                        </span>
                      </SettingRow>
                    </Col>
                    <Col span={12}>
                      <SettingRow>
                        <SettingLabel>
                          <SettingInfo>
                            <SettingName>版本号</SettingName>
                          </SettingInfo>
                        </SettingLabel>
                        <Tag color="blue">v1.0.0</Tag>
                      </SettingRow>
                    </Col>
                    <Col span={12}>
                      <SettingRow>
                        <SettingLabel>
                          <SettingInfo>
                            <SettingName>构建时间</SettingName>
                          </SettingInfo>
                        </SettingLabel>
                        <span style={{ color: 'rgba(255,255,255,0.7)' }}>
                          2024-01-01
                        </span>
                      </SettingRow>
                    </Col>
                    <Col span={12}>
                      <SettingRow>
                        <SettingLabel>
                          <SettingInfo>
                            <SettingName>技术支持</SettingName>
                          </SettingInfo>
                        </SettingLabel>
                        <span style={{ color: 'rgba(255,255,255,0.7)' }}>
                          UAV Team
                        </span>
                      </SettingRow>
                    </Col>
                  </Row>

                  <Divider style={{ margin: '16px 0' }} />

                  <SectionTitle>
                    <SafetyOutlined />
                    第三方依赖
                  </SectionTitle>
                  <Row gutter={16}>
                    <Col span={8}>
                      <Tag color="blue">React 18.2.0</Tag>
                    </Col>
                    <Col span={8}>
                      <Tag color="green">Ant Design 5.11.0</Tag>
                    </Col>
                    <Col span={8}>
                      <Tag color="purple">TypeScript 5.2.0</Tag>
                    </Col>
                    <Col span={8}>
                      <Tag color="cyan">ECharts 5.4.0</Tag>
                    </Col>
                    <Col span={8}>
                      <Tag color="geekblue">Redux Toolkit 1.9.7</Tag>
                    </Col>
                    <Col span={8}>
                      <Tag color="magenta">高德地图 JS API 2.0</Tag>
                    </Col>
                  </Row>
                </SettingsCard>
              )
            }
          ]}
        />

        <div style={{ display: 'flex', gap: 12, justifyContent: 'flex-end', marginTop: 24, paddingBottom: 24 }}>
          <Button onClick={handleResetSettings}>
            重置设置
          </Button>
          <Button type="primary" onClick={handleSaveSettings}>
            保存设置
          </Button>
        </div>
      </Content>

      <Modal
        title="修改密码"
        open={passwordModalVisible}
        onCancel={() => {
          setPasswordModalVisible(false)
          passwordForm.resetFields()
        }}
        footer={null}
        width={400}
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handlePasswordChange}
        >
          <Form.Item
            name="oldPassword"
            label="当前密码"
            rules={[{ required: true, message: '请输入当前密码' }]}
          >
            <Input.Password placeholder="请输入当前密码" />
          </Form.Item>
          <Form.Item
            name="newPassword"
            label="新密码"
            rules={[
              { required: true, message: '请输入新密码' },
              { min: 6, message: '密码至少6个字符' }
            ]}
          >
            <Input.Password placeholder="请输入新密码" />
          </Form.Item>
          <Form.Item
            name="confirmPassword"
            label="确认新密码"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: '请确认新密码' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('两次输入的密码不一致'))
                }
              })
            ]}
          >
            <Input.Password placeholder="请再次输入新密码" />
          </Form.Item>
          <Form.Item>
            <Space>
              <Button type="primary" htmlType="submit">
                确认修改
              </Button>
              <Button onClick={() => {
                setPasswordModalVisible(false)
                passwordForm.resetFields()
              }}>
                取消
              </Button>
            </Space>
          </Form.Item>
        </Form>
      </Modal>
    </Container>
  )
}

export default Settings