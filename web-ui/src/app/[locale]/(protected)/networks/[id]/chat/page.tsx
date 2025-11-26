'use client'
import { useEffect, useState, useRef, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useT } from '../../../../../../lib/i18n-context'
import { useNotification } from '../../../../../../contexts/NotificationContext'
import { getUser, getAccessToken } from '../../../../../../lib/auth'
import AuthGuard from '../../../../../../components/AuthGuard'
import {
  getNetwork,
  listMembers,
  listChatMessages,
  uploadFile,
  type Network,
  type Membership,
  type ChatMessage
} from '../../../../../../lib/api'
import {
  useWebSocket,
} from '../../../../../../lib/websocket'

export default function NetworkChatPage() {
  const router = useRouter()
  const params = useParams()
  const t = useT()
  const notification = useNotification()
  const networkId = params.id as string
  const locale = params.locale as string

  const [network, setNetwork] = useState<Network | null>(null)
  const [messages, setMessages] = useState<ChatMessage[]>([])
  const [members, setMembers] = useState<Membership[]>([])
  const [myMembership, setMyMembership] = useState<Membership | null>(null)
  const [loading, setLoading] = useState(true)
  const [sending, setSending] = useState(false)
  const [message, setMessage] = useState('')
  const [showSidebar, setShowSidebar] = useState(true)
  const [typingUsers, setTypingUsers] = useState<string[]>([])
  const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())

  // File Upload state
  const fileInputRef = useRef<HTMLInputElement>(null)
  const [uploading, setUploading] = useState(false)
  const [attachments, setAttachments] = useState<string[]>([])

  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const typingTimeoutRef = useRef<NodeJS.Timeout | null>(null)
  const user = getUser()

  // WebSocket connection
  const ws = useWebSocket()

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }

  const loadData = useCallback(async () => {
    setLoading(true)
    try {
      const token = getAccessToken()
      if (!token) {
        router.push(`/${locale}/login`)
        return
      }

      const [networkData, messagesData, membersData] = await Promise.all([
        getNetwork(networkId, token),
        listChatMessages(`network:${networkId}`, token, 100),
        listMembers(networkId, token, 'approved')
      ])

      setNetwork(networkData.data)
      // Reverse messages to show oldest first
      setMessages((messagesData.messages || []).slice().reverse())
      setMembers(membersData.data || [])

      if (user) {
        const myMember = (membersData.data || []).find((m: Membership) => m.user_id === user.id && m.status === 'approved')
        setMyMembership(myMember || null)

        if (!myMember) {
          notification.error(t('network.chat.error.title'), t('network.chat.error.notMember'))
          router.push(`/${locale}/networks/${networkId}`)
        }
      }
    } catch (err) {
      notification.error(t('error.generic'), String(err))
    } finally {
      setLoading(false)
    }
  }, [networkId, user, notification, t, router, locale])

  useEffect(() => {
    loadData()
  }, [loadData])

  useEffect(() => {
    scrollToBottom()
  }, [messages])

  // WebSocket: Join network room and setup listeners
  useEffect(() => {
    if (!myMembership || !ws.isConnected) return

    const roomName = `network:${networkId}`

    // Join network chat room
    ws.joinRoom(roomName)

    // Listen for new messages
    const handleNewMessage = (msg: ChatMessage) => {
      if (msg.scope !== roomName) return

      setMessages(prev => {
        // Check if message already exists
        if (prev.find(m => m.id === msg.id)) return prev

        // Add new message
        return [...prev, msg]
      })
    }

    // Listen for edited messages
    const handleEditedMessage = (data: { id: string; body: string; updated_at: string }) => {
      setMessages(prev => prev.map(m =>
        m.id === data.id
          ? { ...m, body: data.body, updated_at: data.updated_at }
          : m
      ))
    }

    // Listen for deleted messages
    const handleDeletedMessage = (data: { id: string }) => {
      setMessages(prev => prev.filter(m => m.id !== data.id))
    }

    // Listen for typing users
    const handleTyping = (data: { user_id: string; user_name: string; is_typing: boolean }) => {
      if (data.user_id === user?.id) return

      setTypingUsers(prev => {
        if (data.is_typing) {
          if (!prev.includes(data.user_name)) {
            return [...prev, data.user_name]
          }
        } else {
          return prev.filter(name => name !== data.user_name)
        }
        return prev
      })

      // Auto-remove typing after 5 seconds
      if (data.is_typing) {
        setTimeout(() => {
          setTypingUsers(prev => prev.filter(name => name !== data.user_name))
        }, 5000)
      }
    }

    // Listen for presence updates (online/offline status)
    const handlePresence = (data: { user_id: string; status: 'online' | 'offline' }) => {
      setOnlineUsers(prev => {
        const next = new Set(prev)
        if (data.status === 'online') {
          next.add(data.user_id)
        } else {
          next.delete(data.user_id)
        }
        return next
      })
    }

    // Register listeners
    ws.on('chat.message', handleNewMessage)
    ws.on('chat.edited', handleEditedMessage)
    ws.on('chat.deleted', handleDeletedMessage)
    ws.on('chat.typing.user', handleTyping)
    ws.on('presence.update', handlePresence)

    // Cleanup
    return () => {
      ws.off('chat.message', handleNewMessage)
      ws.off('chat.edited', handleEditedMessage)
      ws.off('chat.deleted', handleDeletedMessage)
      ws.off('chat.typing.user', handleTyping)
      ws.off('presence.update', handlePresence)
      ws.leaveRoom(roomName)
    }
  }, [networkId, myMembership, ws.isConnected, ws, user?.id])

  const handleSend = async () => {
    if ((!message.trim() && attachments.length === 0) || sending) return

    setSending(true)
    try {
      // Use REST API with chat.send scope
      if (ws.isConnected) {
        ws.send('chat.send', {
          scope: `network:${networkId}`,
          body: message.trim(),
          attachments: attachments
        })
        setMessage('')
        setAttachments([])
        inputRef.current?.focus()
      } else {
        notification.error(t('chat.error.notConnected'), '')
      }
    } catch (err) {
      notification.error(t('chat.error.send'), String(err))
    } finally {
      setSending(false)
    }
  }

  // File upload handler
  const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
    if (!e.target.files || e.target.files.length === 0) return

    const file = e.target.files[0]
    setUploading(true)

    try {
      const token = getAccessToken()
      if (!token) throw new Error(t('chat.error.notAuthenticated'))

      const result = await uploadFile(file, token)
      setAttachments(prev => [...prev, result.url])
    } catch (err: any) {
      console.error('Upload failed:', err)
      notification.error(t('chat.error.upload'), err.message || '')
    } finally {
      setUploading(false)
      // Reset input
      if (fileInputRef.current) {
        fileInputRef.current.value = ''
      }
    }
  }

  // Remove attachment from list
  const removeAttachment = (index: number) => {
    setAttachments(prev => prev.filter((_, i) => i !== index))
  }

  // Send typing indicator
  const handleTyping = () => {
    if (!ws.isConnected) return

    // Debounce typing indicator
    if (typingTimeoutRef.current) {
      clearTimeout(typingTimeoutRef.current)
    }

    ws.send('chat.typing', { scope: `network:${networkId}`, typing: true })

    typingTimeoutRef.current = setTimeout(() => {
      ws.send('chat.typing', { scope: `network:${networkId}`, typing: false })
    }, 2000)
  }

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    } else {
      handleTyping()
    }
  }

  const formatTime = (dateString: string) => {
    const date = new Date(dateString)
    const now = new Date()
    const diffMs = now.getTime() - date.getTime()
    const diffMins = Math.floor(diffMs / 60000)
    const diffHours = Math.floor(diffMs / 3600000)

    if (diffMins < 1) return t('chat.time.justNow')
    if (diffMins < 60) return `${diffMins} ${t('chat.time.ago.m')}`
    if (diffHours < 24) return `${diffHours} ${t('chat.time.ago.h')}`

    return date.toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    })
  }

  const getMemberName = (userId: string) => {
    const member = members.find((m: Membership) => m.user_id === userId)
    // Membership doesn't have user_email, show truncated user_id
    return member ? `User ${member.user_id.substring(0, 8)}...` : 'Unknown User'
  }

  if (loading) {
    return (
      <AuthGuard>
        <div style={{
          minHeight: '100vh',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          backgroundColor: '#f9fafb'
        }}>
          {t('loading')}...
        </div>
      </AuthGuard>
    )
  }

  if (!network || !myMembership) {
    return (
      <AuthGuard>
        <div style={{
          minHeight: '100vh',
          display: 'flex',
          justifyContent: 'center',
          alignItems: 'center',
          backgroundColor: '#f9fafb',
          flexDirection: 'column'
        }}>
          <div style={{ fontSize: 48, marginBottom: 16 }}>🔒</div>
          <p>{t('network.chat.error.notMember')}</p>
        </div>
      </AuthGuard>
    )
  }

  return (
    <AuthGuard>
      <div style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        backgroundColor: '#f9fafb',
        fontFamily: 'system-ui, -apple-system, sans-serif'
      }}>
        {/* Header */}
        <div style={{
          padding: '12px 24px',
          backgroundColor: 'white',
          borderBottom: '1px solid #e5e7eb',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <button
              onClick={() => router.push(`/${locale}/networks/${networkId}`)}
              style={{
                background: 'none',
                border: 'none',
                color: '#6b7280',
                cursor: 'pointer',
                fontSize: 14,
                padding: 0,
                display: 'flex',
                alignItems: 'center',
                gap: 4
              }}
            >
              ← {t('chat.header.back')}
            </button>
            <div>
              <h1 style={{ margin: 0, fontSize: 18, fontWeight: 600, color: '#1f2937' }}>
                💬 {network.name}
              </h1>
              <span style={{ fontSize: 12, color: '#6b7280' }}>
                {t('network.chat.members', { count: members.length })}
              </span>
            </div>
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
            <span style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              fontSize: 12,
              color: ws.isConnected ? '#22c55e' : '#ef4444'
            }}>
              <span style={{
                width: 8,
                height: 8,
                borderRadius: '50%',
                backgroundColor: ws.isConnected ? '#22c55e' : '#ef4444'
              }} />
              {ws.isConnected ? t('chat.connection.connected') : t('chat.connection.disconnected')}
            </span>

            <button
              onClick={() => setShowSidebar(!showSidebar)}
              style={{
                padding: '8px 16px',
                backgroundColor: showSidebar ? '#3b82f6' : '#f3f4f6',
                color: showSidebar ? 'white' : '#374151',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 13,
                fontWeight: 500
              }}
            >
              👥 {t('chat.sidebar.online')}
            </button>
          </div>
        </div>

        {/* Main Content */}
        <div style={{ flex: 1, display: 'flex', overflow: 'hidden' }}>
          {/* Messages Area */}
          <div style={{
            flex: 1,
            display: 'flex',
            flexDirection: 'column',
            overflow: 'hidden'
          }}>
            {/* Messages */}
            <div style={{
              flex: 1,
              overflowY: 'auto',
              padding: '16px 24px'
            }}>
              {messages.length === 0 ? (
                <div style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                  height: '100%',
                  color: '#6b7280'
                }}>
                  <div style={{ fontSize: 48, marginBottom: 16 }}>💬</div>
                  <p style={{ fontSize: 16, margin: 0 }}>{t('chat.empty.title')}</p>
                  <p style={{ fontSize: 14, margin: '8px 0 0', opacity: 0.7 }}>{t('chat.empty.subtitle')}</p>
                </div>
              ) : (
                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                  {messages.map((msg) => {
                    const isOwn = msg.user_id === user?.id
                    const isDeleted = msg.deleted_at !== undefined

                    return (
                      <div
                        key={msg.id}
                        style={{
                          display: 'flex',
                          flexDirection: 'column',
                          alignItems: isOwn ? 'flex-end' : 'flex-start',
                          maxWidth: '70%',
                          alignSelf: isOwn ? 'flex-end' : 'flex-start'
                        }}
                      >
                        {!isOwn && (
                          <span style={{
                            fontSize: 12,
                            color: '#6b7280',
                            marginBottom: 4,
                            marginLeft: 12
                          }}>
                            {getMemberName(msg.user_id)}
                          </span>
                        )}

                        <div
                          style={{
                            backgroundColor: isDeleted ? '#f3f4f6' :
                              isOwn ? '#3b82f6' : 'white',
                            color: isDeleted ? '#9ca3af' :
                              isOwn ? 'white' : '#1f2937',
                            borderRadius: 16,
                            borderTopRightRadius: isOwn ? 4 : 16,
                            borderTopLeftRadius: isOwn ? 16 : 4,
                            padding: '12px 16px',
                            fontStyle: isDeleted ? 'italic' : 'normal',
                            boxShadow: '0 1px 2px rgba(0,0,0,0.05)',
                            position: 'relative'
                          }}
                        >
                          <p style={{ margin: 0, fontSize: 14, lineHeight: 1.5 }}>
                            {isDeleted ? t('chat.message.redacted') : msg.body}
                          </p>
                          {/* Attachments Display */}
                          {msg.attachments && msg.attachments.length > 0 && !isDeleted && (
                            <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 4 }}>
                              {msg.attachments.map((url, i) => {
                                const isImage = /\.(jpg|jpeg|png|gif|webp)$/i.test(url)
                                return (
                                  <a
                                    key={i}
                                    href={url}
                                    target="_blank"
                                    rel="noopener noreferrer"
                                    style={{
                                      color: isOwn ? 'white' : '#3b82f6',
                                      fontSize: 13,
                                      textDecoration: 'underline',
                                      display: 'flex',
                                      alignItems: 'center',
                                      gap: 4
                                    }}
                                  >
                                    {isImage ? '🖼️' : '📎'} {t('chat.attachment')} {i + 1}
                                  </a>
                                )
                              })}
                            </div>
                          )}
                        </div>

                        <span style={{
                          fontSize: 11,
                          color: '#9ca3af',
                          marginTop: 4,
                          marginRight: isOwn ? 0 : 'auto',
                          marginLeft: isOwn ? 'auto' : 12
                        }}>
                          {formatTime(msg.created_at)}
                        </span>
                      </div>
                    )
                  })}
                  <div ref={messagesEndRef} />
                </div>
              )}
            </div>

            {/* Typing Indicator */}
            {typingUsers.length > 0 && (
              <div style={{
                padding: '8px 24px',
                fontSize: 13,
                color: '#6b7280',
                fontStyle: 'italic'
              }}>
                {typingUsers.length === 1
                  ? `${typingUsers[0]} ${t('chat.typing.single')}`
                  : `${typingUsers.join(', ')} ${t('chat.typing.plural')}`
                }
              </div>
            )}

            {/* Input Area */}
            <div style={{
              padding: 16,
              backgroundColor: 'white',
              borderTop: '1px solid #e5e7eb'
            }}>
              {/* Attachment Preview */}
              {attachments.length > 0 && (
                <div style={{ marginBottom: 12, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                  {attachments.map((url, i) => (
                    <div key={i} style={{
                      padding: '4px 8px',
                      backgroundColor: '#e5e7eb',
                      borderRadius: 4,
                      fontSize: 12,
                      display: 'flex',
                      alignItems: 'center',
                      gap: 6
                    }}>
                      <span>📎 {t('chat.attachment')} {i + 1}</span>
                      <button
                        onClick={() => removeAttachment(i)}
                        style={{ border: 'none', background: 'none', cursor: 'pointer', padding: 0, color: '#ef4444' }}
                      >
                        ✕
                      </button>
                    </div>
                  ))}
                </div>
              )}
              <div style={{
                display: 'flex',
                gap: 12,
                alignItems: 'center'
              }}>
                {/* Hidden File Input */}
                <input
                  type="file"
                  ref={fileInputRef}
                  onChange={handleFileSelect}
                  style={{ display: 'none' }}
                />
                {/* Attachment Button */}
                <button
                  type="button"
                  onClick={() => fileInputRef.current?.click()}
                  disabled={!ws.isConnected || uploading}
                  style={{
                    padding: '10px 12px',
                    backgroundColor: '#f3f4f6',
                    border: '1px solid #e5e7eb',
                    borderRadius: 24,
                    cursor: ws.isConnected && !uploading ? 'pointer' : 'not-allowed',
                    fontSize: 18,
                    opacity: ws.isConnected && !uploading ? 1 : 0.5
                  }}
                  title={t('chat.input.attach')}
                >
                  {uploading ? '⏳' : '📎'}
                </button>
                <input
                  ref={inputRef}
                  type="text"
                  value={message}
                  onChange={(e) => setMessage(e.target.value)}
                  onKeyDown={handleKeyPress}
                  placeholder={t('chat.input.placeholder')}
                  style={{
                    flex: 1,
                    padding: '12px 16px',
                    border: '1px solid #e5e7eb',
                    borderRadius: 24,
                    fontSize: 14,
                    outline: 'none'
                  }}
                />
                <button
                  onClick={handleSend}
                  disabled={(!message.trim() && attachments.length === 0) || sending}
                  style={{
                    padding: '12px 24px',
                    backgroundColor: '#3b82f6',
                    color: 'white',
                    border: 'none',
                    borderRadius: 24,
                    cursor: (message.trim() || attachments.length > 0) && !sending ? 'pointer' : 'not-allowed',
                    fontSize: 14,
                    fontWeight: 500,
                    opacity: (message.trim() || attachments.length > 0) && !sending ? 1 : 0.5
                  }}
                >
                  {sending ? '...' : t('chat.input.send')}
                </button>
              </div>
            </div>
          </div>

          {/* Sidebar */}
          {showSidebar && (
            <div style={{
              width: 260,
              borderLeft: '1px solid #e5e7eb',
              backgroundColor: 'white',
              overflow: 'hidden',
              display: 'flex',
              flexDirection: 'column'
            }}>
              <div style={{
                padding: '16px',
                borderBottom: '1px solid #e5e7eb'
              }}>
                <h3 style={{
                  margin: 0,
                  fontSize: 14,
                  fontWeight: 600,
                  color: '#374151'
                }}>
                  👥 {t('network.chat.memberList')} ({members.length})
                </h3>
                <p style={{
                  margin: '4px 0 0',
                  fontSize: 12,
                  color: '#22c55e',
                  display: 'flex',
                  alignItems: 'center',
                  gap: 4
                }}>
                  <span style={{
                    width: 8,
                    height: 8,
                    borderRadius: '50%',
                    backgroundColor: '#22c55e'
                  }} />
                  {onlineUsers.size} {t('chat.status.online')}
                </p>
              </div>

              <div style={{
                flex: 1,
                overflowY: 'auto',
                padding: '8px 0'
              }}>
                {members.map((member) => {
                  const isOnline = onlineUsers.has(member.user_id)
                  return (
                    <div
                      key={member.id}
                      style={{
                        display: 'flex',
                        alignItems: 'center',
                        gap: 12,
                        padding: '8px 16px',
                        backgroundColor: member.user_id === user?.id ? '#f0f9ff' : 'transparent'
                      }}
                    >
                      <div style={{
                        width: 32,
                        height: 32,
                        borderRadius: '50%',
                        backgroundColor: '#e5e7eb',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        fontSize: 14,
                        position: 'relative'
                      }}>
                        👤
                        {/* Online Status Indicator */}
                        <span style={{
                          position: 'absolute',
                          bottom: 0,
                          right: 0,
                          width: 10,
                          height: 10,
                          borderRadius: '50%',
                          backgroundColor: isOnline ? '#22c55e' : '#9ca3af',
                          border: '2px solid white'
                        }} />
                      </div>
                      <div style={{ flex: 1, minWidth: 0 }}>
                        <div style={{
                          fontSize: 13,
                          fontWeight: 500,
                          color: '#1f2937',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 6
                        }}>
                          User {member.user_id.substring(0, 8)}...
                          {member.user_id === user?.id && (
                            <span style={{ color: '#6b7280', fontWeight: 400 }}> ({t('chat.user.you')})</span>
                          )}
                        </div>
                        <div style={{
                          fontSize: 11,
                          color: isOnline ? '#22c55e' : '#6b7280',
                          textTransform: 'capitalize',
                          display: 'flex',
                          alignItems: 'center',
                          gap: 4
                        }}>
                          {member.role}
                          <span style={{ color: '#9ca3af' }}>•</span>
                          <span>{isOnline ? t('chat.status.online') : t('chat.status.offline')}</span>
                        </div>
                      </div>
                    </div>
                  )
                })}
              </div>
            </div>
          )}
        </div>
      </div>
    </AuthGuard>
  )
}
