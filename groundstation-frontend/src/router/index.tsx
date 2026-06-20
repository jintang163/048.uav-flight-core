import React, { lazy, Suspense } from 'react'
import { Navigate, useRoutes } from 'react-router-dom'
import { Spin } from 'antd'
import type { RouteObject } from 'react-router-dom'
import MainLayout from '@/components/Layout'
import { useAppSelector } from '@/store'

const Loading = () => (
  <div style={{
    width: '100%',
    height: '100vh',
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    background: '#0f172a'
  }}>
    <Spin size="large" tip="加载中..." />
  </div>
)

const Login = lazy(() => import('@/pages/Login'))
const Dashboard = lazy(() => import('@/pages/Dashboard'))
const Mission = lazy(() => import('@/pages/Mission'))
const UAVList = lazy(() => import('@/pages/UAVList'))
const FlightHistory = lazy(() => import('@/pages/FlightHistory'))
const Geofence = lazy(() => import('@/pages/Geofence'))
const AlertCenter = lazy(() => import('@/pages/AlertCenter'))
const Firmware = lazy(() => import('@/pages/Firmware'))
const Settings = lazy(() => import('@/pages/Settings'))
const Formation = lazy(() => import('@/pages/Formation'))

const PrivateRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAppSelector(state => state.auth)
  const token = localStorage.getItem('accessToken')
  
  if (!isAuthenticated && !token) {
    return <Navigate to="/login" replace />
  }
  
  return <>{children}</>
}

const PublicRoute: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const { isAuthenticated } = useAppSelector(state => state.auth)
  const token = localStorage.getItem('accessToken')
  
  if (isAuthenticated || token) {
    return <Navigate to="/dashboard" replace />
  }
  
  return <>{children}</>
}

const routes: RouteObject[] = [
  {
    path: '/login',
    element: (
      <PublicRoute>
        <Suspense fallback={<Loading />}>
          <Login />
        </Suspense>
      </PublicRoute>
    )
  },
  {
    path: '/',
    element: (
      <PrivateRoute>
        <MainLayout />
      </PrivateRoute>
    ),
    children: [
      {
        index: true,
        element: <Navigate to="/dashboard" replace />
      },
      {
        path: 'dashboard',
        element: (
          <Suspense fallback={<Loading />}>
            <Dashboard />
          </Suspense>
        )
      },
      {
        path: 'mission',
        element: (
          <Suspense fallback={<Loading />}>
            <Mission />
          </Suspense>
        )
      },
      {
        path: 'uav-list',
        element: (
          <Suspense fallback={<Loading />}>
            <UAVList />
          </Suspense>
        )
      },
      {
        path: 'flight-history',
        element: (
          <Suspense fallback={<Loading />}>
            <FlightHistory />
          </Suspense>
        )
      },
      {
        path: 'geofence',
        element: (
          <Suspense fallback={<Loading />}>
            <Geofence />
          </Suspense>
        )
      },
      {
        path: 'formation',
        element: (
          <Suspense fallback={<Loading />}>
            <Formation />
          </Suspense>
        )
      },
      {
        path: 'alert-center',
        element: (
          <Suspense fallback={<Loading />}>
            <AlertCenter />
          </Suspense>
        )
      },
      {
        path: 'firmware',
        element: (
          <Suspense fallback={<Loading />}>
            <Firmware />
          </Suspense>
        )
      },
      {
        path: 'settings',
        element: (
          <Suspense fallback={<Loading />}>
            <Settings />
          </Suspense>
        )
      }
    ]
  },
  {
    path: '*',
    element: <Navigate to="/dashboard" replace />
  }
]

const Router: React.FC = () => {
  const element = useRoutes(routes)
  return element
}

export default Router