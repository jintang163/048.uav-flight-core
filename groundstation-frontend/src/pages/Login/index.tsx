import React, { useState, useEffect } from 'react'
import styled from 'styled-components'
import { Form, Input, Button, Card, message, Checkbox } from 'antd'
import { UserOutlined, LockOutlined, RocketOutlined } from '@ant-design/icons'
import { useNavigate, useLocation } from 'react-router-dom'
import { useAppDispatch, useAppSelector } from '@/store'
import { login } from '@/store/slices/auth'
import type { LoginRequest } from '@/types'

const Container = styled.div`
  width: 100vw;
  height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  background: linear-gradient(135deg, #001529 0%, #002140 50%, #003a70 100%);
  position: relative;
  overflow: hidden;

  &::before {
    content: '';
    position: absolute;
    top: -50%;
    left: -50%;
    width: 200%;
    height: 200%;
    background: radial-gradient(circle, rgba(24, 144, 255, 0.1) 0%, transparent 50%);
    animation: rotate 20s linear infinite;
  }

  @keyframes rotate {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }
`

const BackgroundDecoration = styled.div`
  position: absolute;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  overflow: hidden;
  pointer-events: none;

  .grid-line {
    position: absolute;
    background: rgba(24, 144, 255, 0.1);

    &.horizontal {
      height: 1px;
      left: 0;
      right: 0;
    }

    &.vertical {
      width: 1px;
      top: 0;
      bottom: 0;
    }
  }
`

const LoginCard = styled(Card)`
  width: 400px;
  background: rgba(255, 255, 255, 0.95);
  backdrop-filter: blur(10px);
  border-radius: 16px;
  box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
  position: relative;
  z-index: 10;

  .ant-card-body {
    padding: 40px;
  }
`

const Logo = styled.div`
  text-align: center;
  margin-bottom: 32px;
`

const LogoIcon = styled.div`
  width: 64px;
  height: 64px;
  margin: 0 auto 16px;
  background: linear-gradient(135deg, #1890ff 0%, #52c41a 100%);
  border-radius: 16px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 32px;
  color: #fff;
  box-shadow: 0 8px 24px rgba(24, 144, 255, 0.3);
`

const Title = styled.h1`
  font-size: 24px;
  font-weight: 700;
  color: #001529;
  margin: 0 0 8px;
`

const Subtitle = styled.p`
  font-size: 14px;
  color: rgba(0, 0, 0, 0.45);
  margin: 0;
`

const StyledForm = styled(Form)`
  .ant-form-item {
    margin-bottom: 20px;
  }

  .ant-input-affix-wrapper {
    border-radius: 8px;
    padding: 8px 12px;
  }

  .ant-input {
    height: 40px;
    font-size: 14px;
  }

  .ant-btn-primary {
    height: 44px;
    font-size: 16px;
    font-weight: 600;
    border-radius: 8px;
    background: linear-gradient(135deg, #1890ff 0%, #096dd9 100%);
    border: none;

    &:hover {
      background: linear-gradient(135deg, #40a9ff 0%, #1890ff 100%);
    }
  }
`

const Footer = styled.div`
  text-align: center;
  margin-top: 24px;
  font-size: 12px;
  color: rgba(0, 0, 0, 0.45);
`

const Login: React.FC = () => {
  const [form] = Form.useForm()
  const dispatch = useAppDispatch()
  const navigate = useNavigate()
  const location = useLocation()
  const { loading, isAuthenticated, error } = useAppSelector(state => state.auth)
  const [rememberMe, setRememberMe] = useState<boolean>(true)

  const from = (location.state as { from?: string })?.from || '/dashboard'

  useEffect(() => {
    if (isAuthenticated) {
      navigate(from, { replace: true })
    }
  }, [isAuthenticated, navigate, from])

  useEffect(() => {
    if (error) {
      message.error(error)
    }
  }, [error])

  const handleSubmit = async (values: LoginRequest) => {
    try {
      const result = await dispatch(login(values)).unwrap()
      if (result) {
        message.success('登录成功')
        if (rememberMe) {
          localStorage.setItem('rememberedUsername', values.username)
        } else {
          localStorage.removeItem('rememberedUsername')
        }
      }
    } catch (err) {
      message.error(err instanceof Error ? err.message : '登录失败')
    }
  }

  useEffect(() => {
    const rememberedUsername = localStorage.getItem('rememberedUsername')
    if (rememberedUsername) {
      form.setFieldsValue({ username: rememberedUsername })
    }
  }, [form])

  const renderGridLines = () => {
    const lines = []
    for (let i = 0; i < 20; i++) {
      lines.push(
        <div
          key={`h-${i}`}
          className="grid-line horizontal"
          style={{ top: `${i * 5}%` }}
        />
      )
      lines.push(
        <div
          key={`v-${i}`}
          className="grid-line vertical"
          style={{ left: `${i * 5}%` }}
        />
      )
    }
    return lines
  }

  return (
    <Container>
      <BackgroundDecoration>
        {renderGridLines()}
      </BackgroundDecoration>

      <LoginCard>
        <Logo>
          <LogoIcon>
            <RocketOutlined />
          </LogoIcon>
          <Title>无人机地面站</Title>
          <Subtitle>UAV Ground Control Station</Subtitle>
        </Logo>

        <StyledForm
          form={form}
          layout="vertical"
          onFinish={handleSubmit}
          initialValues={{ username: '', password: '' }}
        >
          <Form.Item
            name="username"
            label="用户名"
            rules={[
              { required: true, message: '请输入用户名' },
              { min: 3, message: '用户名至少3个字符' }
            ]}
          >
            <Input
              prefix={<UserOutlined style={{ color: 'rgba(0,0,0,0.25)' }} />}
              placeholder="请输入用户名"
              size="large"
            />
          </Form.Item>

          <Form.Item
            name="password"
            label="密码"
            rules={[
              { required: true, message: '请输入密码' },
              { min: 6, message: '密码至少6个字符' }
            ]}
          >
            <Input.Password
              prefix={<LockOutlined style={{ color: 'rgba(0,0,0,0.25)' }} />}
              placeholder="请输入密码"
              size="large"
            />
          </Form.Item>

          <Form.Item>
            <Checkbox
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
            >
              记住用户名
            </Checkbox>
          </Form.Item>

          <Form.Item style={{ marginBottom: 0 }}>
            <Button
              type="primary"
              htmlType="submit"
              loading={loading}
              block
            >
              {loading ? '登录中...' : '登 录'}
            </Button>
          </Form.Item>
        </StyledForm>

        <Footer>
          版本 v1.0.0 | © 2024 UAV Ground Station
        </Footer>
      </LoginCard>
    </Container>
  )
}

export default Login
