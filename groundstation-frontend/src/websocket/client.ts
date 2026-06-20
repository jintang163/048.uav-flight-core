type MessageHandler = (data: unknown) => void

interface WebSocketOptions {
  url: string
  reconnectInterval?: number
  maxReconnectAttempts?: number
  heartbeatInterval?: number
  onOpen?: () => void
  onClose?: () => void
  onError?: (error: Event) => void
  onMessage?: (data: unknown) => void
}

class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectInterval: number
  private maxReconnectAttempts: number
  private heartbeatInterval: number
  private reconnectAttempts: number = 0
  private heartbeatTimer: ReturnType<typeof setInterval> | null = null
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private shouldReconnect: boolean = true
  private messageHandlers: Map<string, Set<MessageHandler>> = new Map()
  private onOpen?: () => void
  private onClose?: () => void
  private onError?: (error: Event) => void
  private onMessage?: (data: unknown) => void
  private isConnected: boolean = false

  constructor(options: WebSocketOptions) {
    this.url = options.url
    this.reconnectInterval = options.reconnectInterval || 3000
    this.maxReconnectAttempts = options.maxReconnectAttempts || 10
    this.heartbeatInterval = options.heartbeatInterval || 30000
    this.onOpen = options.onOpen
    this.onClose = options.onClose
    this.onError = options.onError
    this.onMessage = options.onMessage
  }

  connect(): void {
    if (this.ws && (this.ws.readyState === WebSocket.OPEN || this.ws.readyState === WebSocket.CONNECTING)) {
      return
    }

    const token = localStorage.getItem('accessToken')
    const wsUrl = token ? `${this.url}?token=${token}` : this.url

    try {
      this.ws = new WebSocket(wsUrl)

      this.ws.onopen = this.handleOpen.bind(this)
      this.ws.onclose = this.handleClose.bind(this)
      this.ws.onerror = this.handleError.bind(this)
      this.ws.onmessage = this.handleMessage.bind(this)
    } catch (error) {
      console.error('WebSocket connection error:', error)
      this.scheduleReconnect()
    }
  }

  private handleOpen(): void {
    this.isConnected = true
    this.reconnectAttempts = 0
    this.startHeartbeat()
    this.onOpen?.()
  }

  private handleClose(event: CloseEvent): void {
    this.isConnected = false
    this.stopHeartbeat()
    this.onClose?.()

    if (this.shouldReconnect) {
      this.scheduleReconnect()
    }
  }

  private handleError(error: Event): void {
    console.error('WebSocket error:', error)
    this.onError?.(error)
  }

  private handleMessage(event: MessageEvent): void {
    try {
      const data = JSON.parse(event.data)
      const msgType = data.type
      const payload = data.payload !== undefined ? data.payload : data.data

      if (msgType && this.messageHandlers.has(msgType)) {
        this.messageHandlers.get(msgType)?.forEach(handler => handler(payload))
      }

      this.onMessage?.(data)
    } catch (error) {
      console.error('Failed to parse WebSocket message:', error)
    }
  }

  private startHeartbeat(): void {
    this.heartbeatTimer = setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'heartbeat', timestamp: Date.now() }))
      }
    }, this.heartbeatInterval)
  }

  private stopHeartbeat(): void {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      console.error('Max reconnect attempts reached')
      return
    }

    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }

    const delay = Math.min(this.reconnectInterval * Math.pow(1.5, this.reconnectAttempts), 30000)
    
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++
      this.connect()
    }, delay)
  }

  send(type: string, payload: unknown): void {
    const sendPayload = (): void => {
      if (this.ws?.readyState !== WebSocket.OPEN) {
        console.warn('WebSocket is not connected. Message queued.')
        setTimeout(sendPayload, 1000)
        return
      }

      const payloadObj = payload as Record<string, unknown>
      const msg: Record<string, unknown> = {
        type,
        action: type,
        payload,
        data: payload,
        timestamp: Date.now()
      }

      if (payloadObj && typeof payloadObj === 'object') {
        if (payloadObj.uavId !== undefined) {
          msg.uavId = payloadObj.uavId
          msg.uav_id = payloadObj.uavId
        }
        if (payloadObj.uav_id !== undefined) {
          msg.uav_id = payloadObj.uav_id
          msg.uavId = payloadObj.uav_id
        }
      }

      this.ws.send(JSON.stringify(msg))
    }

    sendPayload()
  }

  on(type: string, handler: MessageHandler): void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set())
    }
    this.messageHandlers.get(type)?.add(handler)
  }

  off(type: string, handler: MessageHandler): void {
    this.messageHandlers.get(type)?.delete(handler)
  }

  disconnect(): void {
    this.shouldReconnect = false
    
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    
    this.stopHeartbeat()
    
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
    
    this.isConnected = false
    this.reconnectAttempts = 0
  }

  reconnect(): void {
    this.shouldReconnect = true
    this.reconnectAttempts = 0
    this.connect()
  }

  get readyState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED
  }

  get connected(): boolean {
    return this.isConnected
  }

  get attempts(): number {
    return this.reconnectAttempts
  }
}

let wsClient: WebSocketClient | null = null

export const createWebSocketClient = (options: WebSocketOptions): WebSocketClient => {
  if (!wsClient) {
    wsClient = new WebSocketClient(options)
  }
  return wsClient
}

export const getWebSocketClient = (): WebSocketClient | null => {
  return wsClient
}

export const destroyWebSocketClient = (): void => {
  if (wsClient) {
    wsClient.disconnect()
    wsClient = null
  }
}

export default WebSocketClient
