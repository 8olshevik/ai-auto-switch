type EventCallback = (data: any) => void

export const WS_EVENTS = {
  PROXY_STATUS: 'proxy:status',
  REQUEST_LOG: 'request:log',
  HEALTH_RESULT: 'health:result',
  ASSISTANT_REPLY: 'assistant:reply',
} as const

class WSClient {
  private ws: WebSocket | null = null
  private listeners: Map<string, Set<EventCallback>> = new Map()
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private isConnecting = false

  connect() {
    if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) return

    const token = localStorage.getItem('token')
    if (!token) return

    this.isConnecting = true
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const wsUrl = `${protocol}//${window.location.host}/api/v1/ws/events?token=${token}`

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      this.isConnecting = false
      console.log('[WS] Connected')
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(event.data)
        this.emit(msg.type, msg.payload)
      } catch (e) {
        console.error('[WS] Failed to parse message:', e)
      }
    }

    this.ws.onclose = () => {
      this.isConnecting = false
      this.scheduleReconnect()
    }

    this.ws.onerror = () => {
      this.isConnecting = false
    }
  }

  disconnect() {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }

  on(event: string, callback: EventCallback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, new Set())
    }
    this.listeners.get(event)!.add(callback)
  }

  off(event: string, callback: EventCallback) {
    this.listeners.get(event)?.delete(callback)
  }

  private emit(event: string, data: any) {
    this.listeners.get(event)?.forEach(cb => cb(data))
  }

  private scheduleReconnect() {
    if (this.reconnectTimer) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, 3000)
  }
}

export const wsClient = new WSClient()
