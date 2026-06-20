import { useState, useEffect, useCallback } from 'react'
import { theme as antdTheme } from 'antd'

export const useTheme = () => {
  const [theme, setTheme] = useState<'light' | 'dark'>(() => {
    const saved = localStorage.getItem('theme')
    return (saved as 'light' | 'dark') || 'dark'
  })

  const toggleTheme = useCallback(() => {
    setTheme(prev => {
      const newTheme = prev === 'light' ? 'dark' : 'light'
      localStorage.setItem('theme', newTheme)
      return newTheme
    })
  }, [])

  const setThemeMode = useCallback((mode: 'light' | 'dark') => {
    setTheme(mode)
    localStorage.setItem('theme', mode)
  }, [])

  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme)
  }, [theme])

  const themeConfig = {
    algorithm: theme === 'dark' ? antdTheme.darkAlgorithm : antdTheme.defaultAlgorithm,
    token: {
      colorPrimary: '#1890ff',
      borderRadius: 6
    },
    components: {
      Layout: {
        headerBg: theme === 'dark' ? '#141414' : '#ffffff',
        siderBg: theme === 'dark' ? '#141414' : '#ffffff',
        bodyBg: theme === 'dark' ? '#000000' : '#f5f5f5'
      }
    }
  }

  return {
    theme,
    isDark: theme === 'dark',
    toggleTheme,
    setTheme: setThemeMode,
    themeConfig
  }
}
