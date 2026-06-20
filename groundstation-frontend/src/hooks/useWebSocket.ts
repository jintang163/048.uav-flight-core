import { useEffect, useCallback, useRef, useState } from 'react'
import { createWebSocketClient, getWebSocketClient, destroyWebSocketClient } from '@/websocket/client'
import { setupTelemetryHandlers } from '@/websocket/telemetry'
import { useAppDispatch } from '@/store'

interface UseWebSocketOptions {
  url?: string
  autoConnect?: boolean
}

export const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const { url = import.meta.env.VITE_WS_URL || '/ws', autoConnect = true } = options
  const dispatch = useAppDispatch()
  const [isConnected, setIsConnected] = useState(false)
  const [reconnectAttempts, setReconnectAttempts] = useState(0)
  const clientRef = useRef(getWebSocketClient())

  const connect = useCallback(() => {
    let client = clientRef.current
    if (!client) {
      client = createWebSocketClient({
        url,
        onOpen: () => {
          setIsConnected(true)
          setReconnectAttempts(0)
        },
        onClose: () => {
          setIsConnected(false)
          if (client) {
            setReconnectAttempts(client.attempts)
          }
        }
      })
      clientRef.current = client
      setupTelemetryHandlers(client, dispatch)
    }
    client.connect()
  }, [dispatch, url])

  const disconnect = useCallback(() => {
    destroyWebSocketClient()
    clientRef.current = null
    setIsConnected(false)
    setReconnectAttempts(0)
  }, [])

  const send = useCallback((type: string, payload: unknown) => {
    const client = clientRef.current
    if (client?.connected) {
      client.send(type, payload)
    }
  }, [])

  const on = useCallback((type: string, handler: (data: unknown) => void) => {
    const client = clientRef.current
    if (client) {
      client.on(type, handler)
    }
  }, [])

  const off = useCallback((type: string, handler: (data: unknown) => void) => {
    const client = clientRef.current
    if (client) {
      client.off(type, handler)
    }
  }, [])

  useEffect(() => {
    if (autoConnect) {
      connect()
    }
    return () => {
      if (autoConnect) {
        disconnect()
      }
    }
  }, [autoConnect, connect, disconnect])

  return {
    isConnected,
    reconnectAttempts,
    connect,
    disconnect,
    send,
    on,
    off,
    client: clientRef.current
  }
}
