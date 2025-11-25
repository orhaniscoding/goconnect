/**
 * WebSocket client for real-time communication with GoConnect server.
 *
 * Supports:
 * - Automatic reconnection with exponential backoff
 * - Tenant chat rooms (tenant.chat.send, tenant.chat.message, etc.)
 * - Network chat rooms
 * - Presence and typing indicators
 * - Operation acknowledgment correlation
 */

import { getAccessToken } from './auth'

// Message types matching server/internal/websocket/message.go
export type MessageType =
  // Inbound (client -> server)
  | 'auth.refresh'
  | 'chat.send'
  | 'chat.edit'
  | 'chat.delete'
  | 'chat.typing'
  | 'room.join'
  | 'room.leave'
  | 'presence.set'
  | 'presence.ping'
  | 'tenant.chat.send'
  | 'tenant.chat.edit'
  | 'tenant.chat.delete'
  | 'tenant.chat.typing'
  | 'tenant.join'
  | 'tenant.leave'
  // Outbound (server -> client)
  | 'chat.message'
  | 'chat.edited'
  | 'chat.deleted'
  | 'chat.typing.user'
  | 'tenant.chat.message'
  | 'tenant.chat.edited'
  | 'tenant.chat.deleted'
  | 'tenant.chat.typing.user'
  | 'tenant.member.joined'
  | 'tenant.member.left'
  | 'tenant.member.kicked'
  | 'tenant.member.role_changed'
  | 'tenant.announcement'
  | 'member.joined'
  | 'member.left'
  | 'device.online'
  | 'device.offline'
  | 'presence.update'
  | 'presence.pong'
  | 'ack'
  | 'error'

export interface WSMessage<T = unknown> {
  type: MessageType
  op_id?: string
  data?: T
  error?: { code: string; message: string; details?: Record<string, string> }
}

export interface TenantChatMessage {
  id: string
  tenant_id: string
  user_id: string
  user?: {
    id: string
    display_name: string
    nickname?: string
  }
  content: string
  created_at: string
  edited_at?: string
}

export interface TenantTypingUser {
  tenant_id: string
  user_id: string
  typing: boolean
}

export interface TenantMemberEvent {
  tenant_id: string
  user_id: string
  role?: string
  old_role?: string
  new_role?: string
  by?: string
  reason?: string
}

// Payload sent by server for tenant.chat.message
export interface TenantChatMessagePayload {
  tenant_id: string
  message_id: string
  sender: {
    user_id: string
    display_name: string
    nickname?: string
  }
  content: string
  timestamp: string
}

// Payload for tenant.chat.edited
export interface TenantChatEditedPayload {
  tenant_id: string
  message_id: string
  content: string
  updated_at: string
}

// Payload for tenant.chat.deleted
export interface TenantChatDeletedPayload {
  tenant_id: string
  message_id: string
}

// Payload for tenant.chat.typing.user
export interface TenantChatTypingPayload {
  tenant_id: string
  user_id: string
  user_name: string
  is_typing: boolean
}

type MessageHandler<T = unknown> = (data: T, msg: WSMessage<T>) => void
type AckHandler = (data: unknown, error?: WSMessage['error']) => void

interface PendingAck {
  resolve: AckHandler
  timeout: ReturnType<typeof setTimeout>
}

const WS_URL = process.env.NEXT_PUBLIC_WS_URL || 'ws://localhost:8080/ws'
const RECONNECT_BASE_DELAY = 1000
const RECONNECT_MAX_DELAY = 30000
const ACK_TIMEOUT = 10000
const PING_INTERVAL = 25000

export class WebSocketClient {
  private ws: WebSocket | null = null
  private reconnectAttempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private pingTimer: ReturnType<typeof setInterval> | null = null
  private messageHandlers = new Map<MessageType, Set<MessageHandler>>()
  private pendingAcks = new Map<string, PendingAck>()
  private opIdCounter = 0
  private _isConnected = false
  private joinedRooms = new Set<string>()

  get isConnected(): boolean {
    return this._isConnected
  }

  /**
   * Connect to the WebSocket server.
   */
  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN) return

    const token = getAccessToken()
    if (!token) {
      console.warn('[WS] No access token, cannot connect')
      return
    }

    // Include token in URL for initial auth
    const url = `${WS_URL}?token=${encodeURIComponent(token)}`
    this.ws = new WebSocket(url)

    this.ws.onopen = () => {
      console.log('[WS] Connected')
      this._isConnected = true
      this.reconnectAttempts = 0
      this.startPing()
      // Rejoin previously joined rooms
      this.joinedRooms.forEach((room) => {
        if (room.startsWith('tenant:')) {
          this.joinTenantRoom(room.replace('tenant:', ''))
        } else {
          this.joinRoom(room)
        }
      })
    }

    this.ws.onclose = (event) => {
      console.log('[WS] Disconnected', event.code, event.reason)
      this._isConnected = false
      this.stopPing()
      this.scheduleReconnect()
    }

    this.ws.onerror = (error) => {
      console.error('[WS] Error', error)
    }

    this.ws.onmessage = (event) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.handleMessage(msg)
      } catch (err) {
        console.error('[WS] Failed to parse message', err)
      }
    }
  }

  /**
   * Disconnect from the WebSocket server.
   */
  disconnect(): void {
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.stopPing()
    if (this.ws) {
      this.ws.close(1000, 'Client disconnect')
      this.ws = null
    }
    this._isConnected = false
  }

  /**
   * Subscribe to a message type.
   * Handler receives the data payload directly.
   */
  on<T = unknown>(type: MessageType, handler: MessageHandler<T>): () => void {
    if (!this.messageHandlers.has(type)) {
      this.messageHandlers.set(type, new Set())
    }
    // Store wrapper that extracts data
    const wrapper: MessageHandler = (data, msg) => handler(data as T, msg as WSMessage<T>)
      ; (handler as unknown as { __wrapper: MessageHandler }).__wrapper = wrapper
    this.messageHandlers.get(type)!.add(wrapper)

    // Return unsubscribe function
    return () => {
      this.messageHandlers.get(type)?.delete(wrapper)
    }
  }

  /**
   * Unsubscribe from a message type.
   */
  off<T = unknown>(type: MessageType, handler: MessageHandler<T>): void {
    const wrapper = (handler as unknown as { __wrapper?: MessageHandler }).__wrapper
    if (wrapper) {
      this.messageHandlers.get(type)?.delete(wrapper)
    }
  }

  /**
   * Send a message and optionally wait for acknowledgment.
   */
  send<T = unknown>(type: MessageType, data?: unknown, waitForAck = false): Promise<T | void> {
    return new Promise((resolve, reject) => {
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
        reject(new Error('WebSocket not connected'))
        return
      }

      const opId = waitForAck ? `op_${++this.opIdCounter}_${Date.now()}` : undefined
      const msg: WSMessage = { type, op_id: opId, data }

      if (waitForAck && opId) {
        const timeout = setTimeout(() => {
          this.pendingAcks.delete(opId)
          reject(new Error('Acknowledgment timeout'))
        }, ACK_TIMEOUT)

        this.pendingAcks.set(opId, {
          resolve: (ackData, error) => {
            if (error) {
              reject(new Error(error.message || error.code))
            } else {
              resolve(ackData as T)
            }
          },
          timeout,
        })
      }

      this.ws.send(JSON.stringify(msg))

      if (!waitForAck) {
        resolve()
      }
    })
  }

  // ==================== ROOM MANAGEMENT ====================

  /**
   * Join a generic room (e.g., 'network:123').
   */
  joinRoom(room: string): void {
    this.joinedRooms.add(room)
    this.send('room.join', { room })
  }

  /**
   * Leave a generic room.
   */
  leaveRoom(room: string): void {
    this.joinedRooms.delete(room)
    this.send('room.leave', { room })
  }

  /**
   * Join a tenant room for real-time updates.
   */
  joinTenantRoom(tenantId: string): void {
    this.joinedRooms.add(`tenant:${tenantId}`)
    this.send('tenant.join', { tenant_id: tenantId })
  }

  /**
   * Leave a tenant room.
   */
  leaveTenantRoom(tenantId: string): void {
    this.joinedRooms.delete(`tenant:${tenantId}`)
    this.send('tenant.leave', { tenant_id: tenantId })
  }

  // ==================== TENANT CHAT ====================

  /**
   * Send a message to tenant chat.
   */
  sendTenantChatMessage(
    tenantId: string,
    content: string
  ): Promise<{ message_id: string } | void> {
    return this.send('tenant.chat.send', { tenant_id: tenantId, content }, true)
  }

  /**
   * Edit a tenant chat message.
   */
  editTenantChatMessage(tenantId: string, messageId: string, content: string): Promise<void> {
    return this.send(
      'tenant.chat.edit',
      { tenant_id: tenantId, message_id: messageId, content },
      true
    ) as Promise<void>
  }

  /**
   * Delete a tenant chat message.
   */
  deleteTenantChatMessage(tenantId: string, messageId: string): Promise<void> {
    return this.send(
      'tenant.chat.delete',
      { tenant_id: tenantId, message_id: messageId },
      true
    ) as Promise<void>
  }

  /**
   * Send typing indicator for tenant chat.
   */
  sendTenantTyping(tenantId: string, typing: boolean): void {
    this.send('tenant.chat.typing', { tenant_id: tenantId, typing })
  }

  // ==================== PRESENCE ====================

  /**
   * Set presence status.
   */
  setPresence(status: 'online' | 'away' | 'busy' | 'offline'): void {
    this.send('presence.set', { status })
  }

  // ==================== PRIVATE METHODS ====================

  private handleMessage(msg: WSMessage): void {
    // Handle acknowledgments
    if (msg.type === 'ack' && msg.op_id) {
      const pending = this.pendingAcks.get(msg.op_id)
      if (pending) {
        clearTimeout(pending.timeout)
        this.pendingAcks.delete(msg.op_id)
        pending.resolve(msg.data, msg.error)
      }
      return
    }

    // Handle errors
    if (msg.type === 'error' && msg.op_id) {
      const pending = this.pendingAcks.get(msg.op_id)
      if (pending) {
        clearTimeout(pending.timeout)
        this.pendingAcks.delete(msg.op_id)
        pending.resolve(null, msg.error)
      }
    }

    // Dispatch to handlers - pass data directly
    const handlers = this.messageHandlers.get(msg.type)
    if (handlers) {
      handlers.forEach((handler) => {
        try {
          handler(msg.data, msg)
        } catch (err) {
          console.error('[WS] Handler error', err)
        }
      })
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) return

    const delay = Math.min(
      RECONNECT_BASE_DELAY * Math.pow(2, this.reconnectAttempts),
      RECONNECT_MAX_DELAY
    )
    this.reconnectAttempts++

    console.log(`[WS] Reconnecting in ${delay}ms (attempt ${this.reconnectAttempts})`)
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, delay)
  }

  private startPing(): void {
    this.stopPing()
    this.pingTimer = setInterval(() => {
      this.send('presence.ping')
    }, PING_INTERVAL)
  }

  private stopPing(): void {
    if (this.pingTimer) {
      clearInterval(this.pingTimer)
      this.pingTimer = null
    }
  }
}

// Singleton instance
let wsClient: WebSocketClient | null = null

/**
 * Get the global WebSocket client instance.
 */
export function getWebSocketClient(): WebSocketClient {
  if (!wsClient) {
    wsClient = new WebSocketClient()
  }
  return wsClient
}

/**
 * React hook for using WebSocket with automatic cleanup.
 */
export function useWebSocket() {
  const client = getWebSocketClient()

  // Connect on first use (if not already connected)
  if (typeof window !== 'undefined' && !client.isConnected) {
    client.connect()
  }

  return client
}
