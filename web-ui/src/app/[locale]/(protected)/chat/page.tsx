'use client'
import { useEffect, useState, useRef } from 'react'
import { useRouter } from 'next/navigation'
import { getAccessToken, getUser } from '../../../../lib/auth'
import { useT } from '../../../../lib/i18n-context'
import { listNetworks, Network, listChatMessages, ChatMessage, uploadFile } from '../../../../lib/api'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

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
    const t = useT()
    const [messages, setMessages] = useState<ChatMessage[]>([])
    const [messageText, setMessageText] = useState('')
    const [ws, setWs] = useState<WebSocket | null>(null)
    const [connected, setConnected] = useState(false)
    const [error, setError] = useState<string | null>(null)
    const [currentUserId, setCurrentUserId] = useState<string>('')
    const messagesEndRef = useRef<HTMLDivElement>(null)
    const [scope, setScope] = useState<string>('host') // 'host' or 'network:<id>'
    const [networks, setNetworks] = useState<Network[]>([])
    const [loadingNetworks, setLoadingNetworks] = useState(true)
    const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())

    // Edit/Delete state
    const [editingMessageId, setEditingMessageId] = useState<string | null>(null)
    const [editText, setEditText] = useState('')
    const [deletingMessageId, setDeletingMessageId] = useState<string | null>(null)

    // File Upload state
    const fileInputRef = useRef<HTMLInputElement>(null)
    const [uploading, setUploading] = useState(false)
    const [attachments, setAttachments] = useState<string[]>([])

    // User permissions
    const [isModerator, setIsModerator] = useState(false)
    const [isAdmin, setIsAdmin] = useState(false)

    useEffect(() => {
        const user = getUser()
        if (user) {
            setCurrentUserId(user.id)
            setIsModerator(user.is_moderator || false)
            setIsAdmin(user.is_admin || false)
        }
        loadNetworks()
        connectWebSocket()

        return () => {
            if (ws) {
                ws.close()
            }
        }
    }, [])

    const loadNetworks = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            const response = await listNetworks('mine', token)
            setNetworks(response.data)
        } catch (err) {
            console.error('Failed to load networks:', err)
            setError(t('chat.error.loadNetworks'))
        } finally {
            setLoadingNetworks(false)
        }
    }

    useEffect(() => {
        // Auto-scroll to bottom when new messages arrive
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }, [messages])

    useEffect(() => {
        // Handle room join/leave when scope changes
        if (ws && ws.readyState === WebSocket.OPEN) {
            if (scope.startsWith('network:')) {
                // Join network room
                const roomName = scope
                const joinMessage: WebSocketMessage = {
                    type: 'room.join',
                    op_id: `join-${Date.now()}`,
                    data: { room: roomName }
                }
                ws.send(JSON.stringify(joinMessage))
                console.log(`Joining room: ${roomName}`)
            }
            // Clear messages when switching scope
            setMessages([])
            loadMessages()
        }
    }, [scope, ws])

    const loadMessages = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            const response = await listChatMessages(scope, token, 50)
            // Reverse to show oldest first (chat style)
            setMessages(response.messages.reverse())
        } catch (err) {
            console.error('Failed to load messages:', err)
            setError(t('chat.error.loadMessages'))
        }
    }

    const connectWebSocket = () => {
        const token = getAccessToken()
        if (!token) {
            router.push('/en/login')
            return
        }

        try {
            const apiBase = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
            const wsBase = apiBase.replace(/^http/, 'ws')
            const wsUrl = `${wsBase}/v1/ws?token=${token}`
            const websocket = new WebSocket(wsUrl)

            websocket.onopen = () => {
                console.log('WebSocket connected')
                setConnected(true)
                setError(null)

                // Join initial room if network scope
                if (scope.startsWith('network:')) {
                    const joinMessage: WebSocketMessage = {
                        type: 'room.join',
                        op_id: `join-${Date.now()}`,
                        data: { room: scope }
                    }
                    websocket.send(JSON.stringify(joinMessage))
                }
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
                setError(t('chat.error.wsConnection'))
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
            setError(t('chat.error.wsConnect'))
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
                                ? { ...m, body: t('chat.message.redacted'), redacted: true }
                                : m
                        )
                    )
                }
                break

            case 'presence.update':
                if (msg.data) {
                    const { user_id, status } = msg.data
                    setOnlineUsers(prev => {
                        const next = new Set(prev)
                        if (status === 'online') {
                            next.add(user_id)
                        } else {
                            next.delete(user_id)
                        }
                        return next
                    })
                }
                break

            case 'ack':
                // Acknowledgment received
                console.log('ACK received for op_id:', msg.op_id, msg.data)
                break

            case 'error':
                // Error received
                console.error('WebSocket error:', msg.error)
                setError(msg.error?.message || t('chat.status.unknown'))
                break

            default:
                console.log('Unknown message type:', msg.type)
        }
    }

    const handleSendMessage = (e: React.FormEvent) => {
        e.preventDefault()

        if (!messageText.trim() && attachments.length === 0) {
            return
        }

        if (!ws || ws.readyState !== WebSocket.OPEN) {
            setError(t('chat.error.notConnected'))
            return
        }

        const opId = `op-${Date.now()}`
        const message: WebSocketMessage = {
            type: 'chat.send',
            op_id: opId,
            data: {
                scope,
                body: messageText.trim(),
                attachments: attachments
            },
        }

        try {
            ws.send(JSON.stringify(message))
            setMessageText('')
            setAttachments([])
            setError(null)
        } catch (err) {
            setError(t('chat.error.send'))
            console.error(err)
        }
    }

    const handleFileSelect = async (e: React.ChangeEvent<HTMLInputElement>) => {
        if (!e.target.files || e.target.files.length === 0) return

        const file = e.target.files[0]
        setUploading(true)
        setError(null)

        try {
            const token = getAccessToken()
            if (!token) throw new Error(t('chat.error.notAuthenticated'))

            const result = await uploadFile(file, token)
            setAttachments(prev => [...prev, result.url])
        } catch (err: any) {
            console.error('Upload failed:', err)
            setError(err.message || t('chat.error.upload'))
        } finally {
            setUploading(false)
            // Reset input
            if (fileInputRef.current) {
                fileInputRef.current.value = ''
            }
        }
    }

    const removeAttachment = (index: number) => {
        setAttachments(prev => prev.filter((_, i) => i !== index))
    }

    const formatTime = (timestamp: string) => {
        const date = new Date(timestamp)
        const now = new Date()
        const diffMs = now.getTime() - date.getTime()
        const diffMins = Math.floor(diffMs / 60000)

        if (diffMins < 1) return t('chat.time.justNow')
        if (diffMins < 60) return `${diffMins}${t('chat.time.ago.m')}`
        if (diffMins < 1440) return `${Math.floor(diffMins / 60)}${t('chat.time.ago.h')}`
        return date.toLocaleDateString()
    }

    const canEditMessage = (message: ChatMessage) => {
        if (message.user_id !== currentUserId) return false
        if (message.redacted || message.deleted_at) return false

        // 15 minute edit window
        const createdAt = new Date(message.created_at)
        const now = new Date()
        const diffMins = Math.floor((now.getTime() - createdAt.getTime()) / 60000)
        return diffMins < 15
    }

    const handleStartEdit = (message: ChatMessage) => {
        setEditingMessageId(message.id)
        setEditText(message.body)
    }

    const handleCancelEdit = () => {
        setEditingMessageId(null)
        setEditText('')
    }

    const handleSaveEdit = () => {
        if (!editText.trim() || !editingMessageId || !ws || ws.readyState !== WebSocket.OPEN) {
            return
        }

        const message: WebSocketMessage = {
            type: 'chat.edit',
            op_id: `edit-${Date.now()}`,
            data: {
                message_id: editingMessageId,
                new_body: editText.trim()
            }
        }

        try {
            ws.send(JSON.stringify(message))
            setEditingMessageId(null)
            setEditText('')
            setError(null)
        } catch (err) {
            setError(t('chat.error.edit'))
            console.error(err)
        }
    }

    const handleDeleteMessage = (messageId: string) => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            setError(t('chat.error.notConnected'))
            return
        }

        if (!confirm(t('chat.confirm.delete'))) {
            return
        }

        const message: WebSocketMessage = {
            type: 'chat.delete',
            op_id: `delete-${Date.now()}`,
            data: {
                message_id: messageId,
                mode: 'soft' // soft delete by default
            }
        }

        try {
            ws.send(JSON.stringify(message))
            setError(null)
        } catch (err) {
            setError(t('chat.error.delete'))
            console.error(err)
        }
    }

    const handleRedactMessage = (messageId: string) => {
        if (!ws || ws.readyState !== WebSocket.OPEN) {
            setError(t('chat.error.notConnected'))
            return
        }

        if (!confirm(t('chat.confirm.redact'))) {
            return
        }

        const message: WebSocketMessage = {
            type: 'chat.redact',
            op_id: `redact-${Date.now()}`,
            data: {
                message_id: messageId,
                mask: t('chat.message.redactedByMod')
            }
        }

        try {
            ws.send(JSON.stringify(message))
            setError(null)
        } catch (err) {
            setError(t('chat.error.redact'))
            console.error(err)
        }
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
                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        <button
                            onClick={() => router.push('/en/dashboard')}
                            style={{
                                padding: '6px 12px',
                                backgroundColor: '#6c757d',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: 'pointer',
                                fontSize: 14
                            }}
                        >
                            ← {t('chat.header.back')}
                        </button>
                        <span style={{ fontSize: 20, fontWeight: 600 }}>
                            💬 {scope === 'host' ? t('chat.header.global') :
                                networks.find(n => `network:${n.id}` === scope)?.name || t('chat.header.network')}
                        </span>
                        <select
                            value={scope}
                            onChange={(e) => setScope(e.target.value)}
                            disabled={!connected || loadingNetworks}
                            style={{
                                padding: '8px 12px',
                                borderRadius: 6,
                                border: '1px solid #dee2e6',
                                fontSize: 14,
                                cursor: connected ? 'pointer' : 'not-allowed',
                                backgroundColor: 'white',
                                minWidth: 200
                            }}
                        >
                            <option value="host">🌐 {t('chat.header.global')}</option>
                            {networks.map((network) => (
                                <option key={network.id} value={`network:${network.id}`}>
                                    🔒 {network.name}
                                </option>
                            ))}
                        </select>
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
                            {connected ? `● ${t('chat.header.status.connected')}` : `○ ${t('chat.header.status.disconnected')}`}
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

                <div style={{ display: 'flex', flex: 1, overflow: 'hidden' }}>
                    <div style={{ flex: 1, display: 'flex', flexDirection: 'column' }}>
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
                                    <div style={{ fontSize: 16 }}>{t('chat.empty.title')}</div>
                                    <div style={{ fontSize: 14, marginTop: 8 }}>{t('chat.empty.subtitle')}</div>
                                </div>
                            )}

                            {messages.map((msg) => {
                                const isOwnMessage = msg.user_id === currentUserId
                                const canEdit = canEditMessage(msg)
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
                                            {isOwnMessage ? t('chat.user.you') : `${t('chat.user.other')} ${msg.user_id.substring(0, 8)}`} · {formatTime(msg.created_at)}
                                        </div>
                                        <div style={{
                                            padding: '12px 16px',
                                            borderRadius: 12,
                                            backgroundColor: isOwnMessage ? '#007bff' : 'white',
                                            color: isOwnMessage ? 'white' : '#212529',
                                            boxShadow: '0 1px 2px rgba(0,0,0,0.1)',
                                            wordBreak: 'break-word',
                                            fontSize: 15,
                                            fontStyle: msg.redacted ? 'italic' : 'normal',
                                            ...(msg.redacted && {
                                                backgroundColor: '#f8d7da',
                                                color: '#842029',
                                                border: '1px solid #f5c2c7'
                                            })
                                        }}>
                                            {msg.body}
                                        </div>
                                        {msg.attachments && msg.attachments.length > 0 && (
                                            <div style={{ marginTop: 8, display: 'flex', flexDirection: 'column', gap: 4 }}>
                                                {msg.attachments.map((url, i) => (
                                                    <a
                                                        key={i}
                                                        href={url}
                                                        target="_blank"
                                                        rel="noopener noreferrer"
                                                        style={{
                                                            fontSize: 13,
                                                            color: isOwnMessage ? 'white' : '#007bff',
                                                            textDecoration: 'underline',
                                                            display: 'flex',
                                                            alignItems: 'center',
                                                            gap: 4
                                                        }}
                                                    >
                                                        📎 {t('chat.attachment')} {i + 1}
                                                    </a>
                                                ))}
                                            </div>
                                        )}
                                        {msg.updated_at && msg.updated_at !== msg.created_at && (
                                            <div style={{
                                                fontSize: 11,
                                                color: '#6c757d',
                                                marginTop: 4,
                                                paddingLeft: 8,
                                                paddingRight: 8,
                                                fontStyle: 'italic'
                                            }}>
                                                {t('chat.edited')}
                                            </div>
                                        )}
                                        {isOwnMessage && !msg.deleted_at && !msg.redacted && (
                                            <div style={{
                                                display: 'flex',
                                                gap: 8,
                                                marginTop: 6,
                                                paddingLeft: 8,
                                                paddingRight: 8
                                            }}>
                                                {canEdit && (
                                                    <button
                                                        onClick={() => handleStartEdit(msg)}
                                                        style={{
                                                            padding: '4px 12px',
                                                            fontSize: 12,
                                                            backgroundColor: '#e9ecef',
                                                            border: '1px solid #dee2e6',
                                                            borderRadius: 6,
                                                            cursor: 'pointer',
                                                            color: '#495057',
                                                            fontWeight: 500
                                                        }}
                                                        onMouseEnter={(e) => {
                                                            e.currentTarget.style.backgroundColor = '#dee2e6'
                                                        }}
                                                        onMouseLeave={(e) => {
                                                            e.currentTarget.style.backgroundColor = '#e9ecef'
                                                        }}
                                                    >
                                                        ✏️ {t('chat.action.edit')}
                                                    </button>
                                                )}
                                                <button
                                                    onClick={() => handleDeleteMessage(msg.id)}
                                                    style={{
                                                        padding: '4px 12px',
                                                        fontSize: 12,
                                                        backgroundColor: '#f8d7da',
                                                        border: '1px solid #f5c2c7',
                                                        borderRadius: 6,
                                                        cursor: 'pointer',
                                                        color: '#842029',
                                                        fontWeight: 500
                                                    }}
                                                    onMouseEnter={(e) => {
                                                        e.currentTarget.style.backgroundColor = '#f5c2c7'
                                                    }}
                                                    onMouseLeave={(e) => {
                                                        e.currentTarget.style.backgroundColor = '#f8d7da'
                                                    }}
                                                >
                                                    🗑️ {t('chat.action.delete')}
                                                </button>
                                            </div>
                                        )}
                                        {!isOwnMessage && (isModerator || isAdmin) && !msg.deleted_at && !msg.redacted && (
                                            <div style={{
                                                display: 'flex',
                                                gap: 8,
                                                marginTop: 6,
                                                paddingLeft: 8,
                                                paddingRight: 8
                                            }}>
                                                <button
                                                    onClick={() => handleRedactMessage(msg.id)}
                                                    style={{
                                                        padding: '4px 12px',
                                                        fontSize: 12,
                                                        backgroundColor: '#fff3cd',
                                                        border: '1px solid #ffc107',
                                                        borderRadius: 6,
                                                        cursor: 'pointer',
                                                        color: '#856404',
                                                        fontWeight: 500
                                                    }}
                                                    onMouseEnter={(e) => {
                                                        e.currentTarget.style.backgroundColor = '#ffc107'
                                                    }}
                                                    onMouseLeave={(e) => {
                                                        e.currentTarget.style.backgroundColor = '#fff3cd'
                                                    }}
                                                >
                                                    🚫 {t('chat.action.redact')}
                                                </button>
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
                            {attachments.length > 0 && (
                                <div style={{ marginBottom: 12, display: 'flex', gap: 8, flexWrap: 'wrap' }}>
                                    {attachments.map((url, i) => (
                                        <div key={i} style={{
                                            padding: '4px 8px',
                                            backgroundColor: '#e9ecef',
                                            borderRadius: 4,
                                            fontSize: 12,
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 6
                                        }}>
                                            <span>📎 {t('chat.attachment')} {i + 1}</span>
                                            <button
                                                onClick={() => removeAttachment(i)}
                                                style={{ border: 'none', background: 'none', cursor: 'pointer', padding: 0, color: '#dc3545' }}
                                            >
                                                ✕
                                            </button>
                                        </div>
                                    ))}
                                </div>
                            )}
                            <form onSubmit={handleSendMessage} style={{ display: 'flex', gap: 12 }}>
                                <input
                                    type="file"
                                    ref={fileInputRef}
                                    onChange={handleFileSelect}
                                    style={{ display: 'none' }}
                                />
                                <button
                                    type="button"
                                    onClick={() => fileInputRef.current?.click()}
                                    disabled={!connected || uploading}
                                    style={{
                                        padding: '12px',
                                        backgroundColor: '#f8f9fa',
                                        border: '1px solid #dee2e6',
                                        borderRadius: 8,
                                        cursor: connected && !uploading ? 'pointer' : 'not-allowed',
                                        fontSize: 20
                                    }}
                                    title={t('chat.input.attach')}
                                >
                                    {uploading ? '⏳' : '📎'}
                                </button>
                                <input
                                    type="text"
                                    value={messageText}
                                    onChange={(e) => setMessageText(e.target.value)}
                                    placeholder={t('chat.input.placeholder')}
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
                                    disabled={!connected || (!messageText.trim() && attachments.length === 0)}
                                    style={{
                                        padding: '12px 32px',
                                        backgroundColor: connected && (messageText.trim() || attachments.length > 0) ? '#007bff' : '#6c757d',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        fontWeight: 600,
                                        cursor: connected && (messageText.trim() || attachments.length > 0) ? 'pointer' : 'not-allowed',
                                        transition: 'background-color 0.2s'
                                    }}
                                >
                                    {t('chat.input.send')}
                                </button>
                            </form>
                        </div>
                    </div>

                    {/* Online Users Sidebar */}
                    <div style={{
                        width: 240,
                        borderLeft: '1px solid #dee2e6',
                        backgroundColor: 'white',
                        display: 'flex',
                        flexDirection: 'column'
                    }}>
                        <div style={{
                            padding: '16px',
                            borderBottom: '1px solid #dee2e6',
                            fontWeight: 600,
                            fontSize: 14,
                            color: '#495057'
                        }}>
                            {t('chat.sidebar.online')} ({onlineUsers.size})
                        </div>
                        <div style={{ flex: 1, overflowY: 'auto', padding: '8px 0' }}>
                            {Array.from(onlineUsers).map(userId => (
                                <div key={userId} style={{
                                    padding: '8px 16px',
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 8,
                                    fontSize: 14
                                }}>
                                    <div style={{
                                        width: 8,
                                        height: 8,
                                        borderRadius: '50%',
                                        backgroundColor: '#28a745'
                                    }} />
                                    <span style={{
                                        whiteSpace: 'nowrap',
                                        overflow: 'hidden',
                                        textOverflow: 'ellipsis'
                                    }}>
                                        {userId === currentUserId ? t('chat.user.you') : `${t('chat.user.other')} ${userId.substring(0, 8)}`}
                                    </span>
                                </div>
                            ))}
                            {onlineUsers.size === 0 && (
                                <div style={{ padding: 16, color: '#6c757d', fontSize: 13, textAlign: 'center' }}>
                                    {t('chat.sidebar.empty')}
                                </div>
                            )}
                        </div>
                    </div>
                </div>

                {/* Edit Message Modal */}
                {editingMessageId && (
                    <div style={{
                        position: 'fixed',
                        top: 0,
                        left: 0,
                        right: 0,
                        bottom: 0,
                        backgroundColor: 'rgba(0,0,0,0.5)',
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        zIndex: 1000
                    }}>
                        <div style={{
                            backgroundColor: 'white',
                            borderRadius: 12,
                            padding: 24,
                            maxWidth: 500,
                            width: '90%',
                            boxShadow: '0 4px 16px rgba(0,0,0,0.2)'
                        }}>
                            <h3 style={{
                                margin: '0 0 16px 0',
                                fontSize: 18,
                                fontWeight: 600,
                                color: '#212529'
                            }}>
                                {t('chat.modal.edit.title')}
                            </h3>
                            <textarea
                                value={editText}
                                onChange={(e) => setEditText(e.target.value)}
                                autoFocus
                                style={{
                                    width: '100%',
                                    minHeight: 120,
                                    padding: '12px 16px',
                                    border: '1px solid #dee2e6',
                                    borderRadius: 8,
                                    fontSize: 15,
                                    fontFamily: 'inherit',
                                    resize: 'vertical',
                                    outline: 'none',
                                    marginBottom: 16
                                }}
                            />
                            <div style={{
                                display: 'flex',
                                gap: 12,
                                justifyContent: 'flex-end'
                            }}>
                                <button
                                    onClick={handleCancelEdit}
                                    style={{
                                        padding: '10px 20px',
                                        backgroundColor: '#6c757d',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        fontWeight: 600,
                                        cursor: 'pointer'
                                    }}
                                >
                                    {t('chat.modal.edit.cancel')}
                                </button>
                                <button
                                    onClick={handleSaveEdit}
                                    disabled={!editText.trim()}
                                    style={{
                                        padding: '10px 20px',
                                        backgroundColor: editText.trim() ? '#007bff' : '#6c757d',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        fontWeight: 600,
                                        cursor: editText.trim() ? 'pointer' : 'not-allowed'
                                    }}
                                >
                                    {t('chat.modal.edit.save')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}
            </div>
        </AuthGuard>
    )
}
