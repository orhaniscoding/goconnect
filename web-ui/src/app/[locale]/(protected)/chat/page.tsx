'use client'
import { useEffect, useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { getAccessToken, getUser } from '../../../../lib/auth'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

interface ChatMessage {
  id: string
  scope: string
  user_id: string
  body: string
  redacted: boolean
  deleted_at?: string
  created_at: string
  updated_at?: string
}

interface WebSocketMessage {
  type: string
  op_id?: string
  data?: any
  error?: {
    code: string
    message: string
  }
}

export default function ChatPage() {
  const router = useRouter()
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [messageText, setMessageText] = useState('')
  const [ws, setWs] = useState<WebSocket | null>(null)
  const [connected, setConnected] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [currentUserId, setCurrentUserId] = useState<string>('')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const [scope, setScope] = useState<string>('host') // 'host' for global chat

  useEffect(() => {
    const user = getUser()
    if (user) {
      setCurrentUserId(user.id)
    }
    connectWebSocket()

    return () => {
      if (ws) {
        ws.close()
      }
    }
  }, [])

  useEffect(() => {
    // Auto-scroll to bottom when new messages arrive
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  const connectWebSocket = () => {
    const token = getAccessToken()
    if (!token) {
      router.push('/en/login')
      return
    }

    try {
      const wsUrl = `ws://localhost:8080/v1/ws?token=${token}`
      const websocket = new WebSocket(wsUrl)

      websocket.onopen = () => {
        console.log('WebSocket connected')
        setConnected(true)
        setError(null)
      }

      websocket.onmessage = (event) => {
        try {
          const msg: WebSocketMessage = JSON.parse(event.data)
          handleWebSocketMessage(msg)
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err)
        }
      }

      websocket.onerror = (event) => {
        console.error('WebSocket error:', event)
        setError('WebSocket connection error')
        setConnected(false)
      }

      websocket.onclose = () => {
        console.log('WebSocket disconnected')
        setConnected(false)
        // Auto-reconnect after 3 seconds
        setTimeout(() => {
          connectWebSocket()
        }, 3000)
      }

      setWs(websocket)
    } catch (err) {
      setError('Failed to connect to WebSocket')
      console.error(err)
    }
  }

  const handleWebSocketMessage = (msg: WebSocketMessage) => {
    switch (msg.type) {
      case 'chat.message':
        // New message received
        if (msg.data && msg.data.scope === scope) {
          setMessages((prev) => [...prev, msg.data as ChatMessage])
        }
        break

      case 'chat.edited':
        // Message edited
        if (msg.data) {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === msg.data.message_id
                ? { ...m, body: msg.data.new_body, updated_at: msg.data.edited_at }
                : m
            )
          )
        }
        break

      case 'chat.deleted':
        // Message deleted
        if (msg.data) {
          setMessages((prev) =>
            prev.filter((m) => m.id !== msg.data.message_id)
          )
        }
        break

      case 'chat.redacted':
        // Message redacted
        if (msg.data) {
          setMessages((prev) =>
            prev.map((m) =>
              m.id === msg.data.message_id
                ? { ...m, body: '[REDACTED]', redacted: true }
                : m
            )
          )
        }
        break

      case 'ack':
        // Acknowledgment received
        console.log('ACK received for op_id:', msg.op_id, msg.data)
        break

      case 'error':
        // Error received
        console.error('WebSocket error:', msg.error)
        setError(msg.error?.message || 'Unknown error')
        break

      default:
        console.log('Unknown message type:', msg.type)
    }
  }

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault()

    if (!messageText.trim()) {
      return
    }

    if (!ws || ws.readyState !== WebSocket.OPEN) {
      setError('Not connected to WebSocket')
      return
    }

    const opId = `op-${Date.now()}`
    const message: WebSocketMessage = {
      type: 'chat.send',
      op_id: opId,
      data: {
        scope,
        body: messageText.trim(),
      },
    }

    try {
      ws.send(JSON.stringify(message))
      setMessageText('')
      setError(null)
    } catch (err) {
      setError('Failed to send message')
      console.error(err)
    }
  }

  const formatTime = (timestamp: string) => {
    const date = new Date(timestamp)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)

    if (diffMins < 1) return 'Just now'
    if (diffMins < 60) return `${diffMins}m ago`
    if (diffMins < 1440) return `${Math.floor(diffMins / 60)}h ago`
    return date.toLocaleDateString()
  }

  return (
    <AuthGuard>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        height: '100vh',
        fontFamily: 'system-ui, -apple-system, sans-serif',
        backgroundColor: '#f8f9fa'
      }}>
        {/* Header */}
        <div style={{
          padding: '16px 24px',
          backgroundColor: 'white',
          borderBottom: '1px solid #dee2e6',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div>
            <button
              onClick={() => router.push('/en/dashboard')}
              style={{
                padding: '6px 12px',
                backgroundColor: '#6c757d',
                color: 'white',
                border: 'none',
                borderRadius: 4,
                cursor: 'pointer',
                fontSize: 14,
                marginRight: 16
              }}
            >
              ← Back
            </button>
            <span style={{ fontSize: 20, fontWeight: 600 }}>💬 Global Chat</span>
          </div>
          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <div style={{
              padding: '6px 12px',
              borderRadius: 12,
              backgroundColor: connected ? '#d1e7dd' : '#f8d7da',
              color: connected ? '#0f5132' : '#842029',
              fontSize: 13,
              fontWeight: 500
            }}>
              {connected ? '● Connected' : '○ Disconnected'}
            </div>
          </div>
        </div>

        {/* Error Banner */}
        {error && (
          <div style={{
            padding: 12,
            backgroundColor: '#f8d7da',
            color: '#842029',
            borderBottom: '1px solid #f5c2c7',
            fontSize: 14,
            textAlign: 'center'
          }}>
            {error}
          </div>
        )}

        {/* Messages Container */}
        <div style={{
          flex: 1,
          overflowY: 'auto',
          padding: 24,
          display: 'flex',
          flexDirection: 'column',
          gap: 16
        }}>
          {messages.length === 0 && (
            <div style={{
              textAlign: 'center',
              padding: 40,
              color: '#6c757d'
            }}>
              <div style={{ fontSize: 48, marginBottom: 16 }}>💬</div>
              <div style={{ fontSize: 16 }}>No messages yet</div>
              <div style={{ fontSize: 14, marginTop: 8 }}>Be the first to say hi!</div>
            </div>
          )}

          {messages.map((msg) => {
            const isOwnMessage = msg.user_id === currentUserId
            return (
              <div
                key={msg.id}
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: isOwnMessage ? 'flex-end' : 'flex-start',
                  maxWidth: '70%',
                  marginLeft: isOwnMessage ? 'auto' : 0,
                  marginRight: isOwnMessage ? 0 : 'auto'
                }}
              >
                <div style={{
                  fontSize: 12,
                  color: '#6c757d',
                  marginBottom: 4,
                  paddingLeft: 8,
                  paddingRight: 8
                }}>
                  {isOwnMessage ? 'You' : `User ${msg.user_id.substring(0, 8)}`} · {formatTime(msg.created_at)}
                </div>
                <div style={{
                  padding: '12px 16px',
                  borderRadius: 12,
                  backgroundColor: isOwnMessage ? '#007bff' : 'white',
                  color: isOwnMessage ? 'white' : '#212529',
                  boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                  wordBreak: 'break-word',
                  fontSize: 15,
                  fontStyle: msg.redacted ? 'italic' : 'normal'
                }}>
                  {msg.body}
                </div>
                {msg.updated_at && msg.updated_at !== msg.created_at && (
                  <div style={{
                    fontSize: 11,
                    color: '#6c757d',
                    marginTop: 4,
                    paddingLeft: 8,
                    paddingRight: 8,
                    fontStyle: 'italic'
                  }}>
                    (edited)
                  </div>
                )}
              </div>
            )
          })}
          <div ref={messagesEndRef} />
        </div>

        {/* Message Input */}
        <div style={{
          padding: 24,
          backgroundColor: 'white',
          borderTop: '1px solid #dee2e6'
        }}>
          <form onSubmit={handleSendMessage} style={{ display: 'flex', gap: 12 }}>
            <input
              type="text"
              value={messageText}
              onChange={(e) => setMessageText(e.target.value)}
              placeholder="Type a message..."
              disabled={!connected}
              style={{
                flex: 1,
                padding: '12px 16px',
                borderRadius: 8,
                border: '1px solid #dee2e6',
                fontSize: 15,
                outline: 'none',
                transition: 'border-color 0.2s',
              }}
              onFocus={(e) => {
                e.target.style.borderColor = '#007bff'
              }}
              onBlur={(e) => {
                e.target.style.borderColor = '#dee2e6'
              }}
            />
            <button
              type="submit"
              disabled={!connected || !messageText.trim()}
              style={{
                padding: '12px 32px',
                backgroundColor: connected && messageText.trim() ? '#007bff' : '#6c757d',
                color: 'white',
                border: 'none',
                borderRadius: 8,
                fontSize: 15,
                fontWeight: 600,
                cursor: connected && messageText.trim() ? 'pointer' : 'not-allowed',
                transition: 'background-color 0.2s'
              }}
            >
              Send
            </button>
          </form>
        </div>
      </div>
    </AuthGuard>
  )
}
