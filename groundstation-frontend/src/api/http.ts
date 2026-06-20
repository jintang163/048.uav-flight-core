import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { message } from 'antd'
import type { ApiResponse } from '@/types'

const baseURL = import.meta.env.VITE_API_BASE_URL || '/api'

const http: AxiosInstance = axios.create({
  baseURL,
  timeout: 30000,
  headers: {
    'Content-Type': 'application/json'
  }
})

http.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = localStorage.getItem('accessToken')
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`
    }
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)

http.interceptors.response.use(
  (response: AxiosResponse) => {
    const res = response.data as ApiResponse
    if (res.code !== 200) {
      message.error(res.message || '请求失败')
      if (res.code === 401) {
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
      }
      return Promise.reject(new Error(res.message || '请求失败'))
    }
    return res.data as unknown as AxiosResponse
  },
  (error) => {
    if (error.response) {
      const { status, data } = error.response
      if (status === 401) {
        localStorage.removeItem('accessToken')
        localStorage.removeItem('refreshToken')
        window.location.href = '/login'
      } else if (status === 403) {
        message.error('没有权限访问')
      } else if (status === 404) {
        message.error('请求的资源不存在')
      } else if (status === 500) {
        message.error(data?.message || '服务器错误')
      } else {
        message.error(error.message || '网络错误')
      }
    } else if (error.request) {
      message.error('网络连接失败，请检查网络')
    } else {
      message.error('请求配置错误')
    }
    return Promise.reject(error)
  }
)

export const request = <T = unknown>(config: AxiosRequestConfig): Promise<T> => {
  return http.request<unknown, T>(config)
}

export const get = <T = unknown>(url: string, params?: Record<string, unknown>): Promise<T> => {
  return request<T>({ method: 'GET', url, params })
}

export const post = <T = unknown>(url: string, data?: unknown): Promise<T> => {
  return request<T>({ method: 'POST', url, data })
}

export const put = <T = unknown>(url: string, data?: unknown): Promise<T> => {
  return request<T>({ method: 'PUT', url, data })
}

export const del = <T = unknown>(url: string, params?: Record<string, unknown>): Promise<T> => {
  return request<T>({ method: 'DELETE', url, params })
}

export default http
