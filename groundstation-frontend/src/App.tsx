import React, { useEffect } from 'react'
import { ConfigProvider, App as AntdApp } from 'antd'
import zhCN from 'antd/locale/zh_CN'
import { Provider } from 'react-redux'
import { BrowserRouter } from 'react-router-dom'
import Router from '@/router'
import { store } from '@/store'
import { useTheme } from '@/hooks/useTheme'
import { useAlert } from '@/hooks/useAlert'
import { useWebSocket } from '@/hooks/useWebSocket'
import { useUAV } from '@/hooks/useUAV'
import { requestNotificationPermission, initAudioContext } from '@/utils'
import { initTheme } from '@/hooks/useTheme'

const darkTheme = {
  token: {
    colorPrimary: '#1890ff',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1890ff',
    colorBgContainer: '#1e293b',
    colorBgElevated: '#1e293b',
    colorBgLayout: '#0f172a',
    colorBgSpotlight: '#334155',
    colorBorder: 'rgba(255, 255, 255, 0.1)',
    colorBorderSecondary: 'rgba(255, 255, 255, 0.05)',
    colorText: 'rgba(255, 255, 255, 0.9)',
    colorTextSecondary: 'rgba(255, 255, 255, 0.7)',
    colorTextTertiary: 'rgba(255, 255, 255, 0.5)',
    colorTextQuaternary: 'rgba(255, 255, 255, 0.3)',
    borderRadius: 6,
    borderRadiusLG: 8,
    borderRadiusXS: 4,
    borderRadiusSM: 4,
    controlHeight: 32,
    controlHeightLG: 40,
    controlHeightSM: 24
  },
  components: {
    Layout: {
      bodyBg: '#0f172a',
      headerBg: '#1e293b',
      siderBg: '#1e293b'
    },
    Menu: {
      itemBg: 'transparent',
      itemSelectedBg: 'rgba(24, 144, 255, 0.2)',
      itemSelectedColor: '#fff',
      itemColor: 'rgba(255, 255, 255, 0.7)',
      itemHoverBg: 'rgba(24, 144, 255, 0.1)',
      itemHoverColor: '#fff'
    },
    Table: {
      headerBg: 'rgba(255, 255, 255, 0.03)',
      headerColor: 'rgba(255, 255, 255, 0.7)',
      rowHoverBg: 'rgba(255, 255, 255, 0.03)',
      borderColor: 'rgba(255, 255, 255, 0.1)',
      bodySortBg: 'rgba(24, 144, 255, 0.05)'
    },
    Card: {
      headerBg: 'rgba(255, 255, 255, 0.03)',
      headerBorderBottom: '1px solid rgba(255, 255, 255, 0.1)'
    },
    Modal: {
      headerBg: '#1e293b',
      contentBg: '#1e293b',
      maskBg: 'rgba(0, 0, 0, 0.7)'
    },
    Drawer: {
      headerBg: '#1e293b',
      bodyBg: '#1e293b',
      contentBg: '#1e293b'
    },
    Input: {
      activeBorderColor: '#1890ff',
      hoverBorderColor: '#40a9ff',
      colorBgContainer: 'rgba(255, 255, 255, 0.05)'
    },
    Select: {
      selectorBg: 'rgba(255, 255, 255, 0.05)',
      optionActiveBg: 'rgba(24, 144, 255, 0.1)',
      optionSelectedBg: 'rgba(24, 144, 255, 0.2)'
    },
    Button: {
      colorBgContainer: 'rgba(255, 255, 255, 0.05)',
      defaultBorderColor: 'rgba(255, 255, 255, 0.2)'
    },
    Tabs: {
      itemColor: 'rgba(255, 255, 255, 0.5)',
      itemSelectedColor: '#1890ff',
      itemHoverColor: 'rgba(255, 255, 255, 0.7)',
      inkBarColor: '#1890ff'
    },
    Form: {
      labelColor: 'rgba(255, 255, 255, 0.9)'
    }
  }
}

const lightTheme = {
  token: {
    colorPrimary: '#1890ff',
    colorSuccess: '#52c41a',
    colorWarning: '#faad14',
    colorError: '#ff4d4f',
    colorInfo: '#1890ff',
    borderRadius: 6
  }
}

const AppContent: React.FC = () => {
  const { isDark } = useTheme()
  const { initAlerts } = useAlert()
  const { loadUAVList } = useUAV()
  const { connect } = useWebSocket()

  useEffect(() => {
    initTheme()
    requestNotificationPermission()
    initAudioContext()
    initAlerts()
    loadUAVList()
    
    const token = localStorage.getItem('accessToken')
    if (token) {
      connect()
    }
  }, [])

  return (
    <ConfigProvider
      locale={zhCN}
      theme={isDark ? darkTheme : lightTheme}
    >
      <AntdApp>
        <Router />
      </AntdApp>
    </ConfigProvider>
  )
}

const App: React.FC = () => {
  return (
    <Provider store={store}>
      <BrowserRouter>
        <AppContent />
      </BrowserRouter>
    </Provider>
  )
}

export default App