'use client'
import { useEffect, useState, useRef, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useT } from '../../../../../../lib/i18n-context'
import { useNotification } from '../../../../../../contexts/NotificationContext'
import { getUser } from '../../../../../../lib/auth'
import AuthGuard from '../../../../../../components/AuthGuard'
import {
    getTenantByID,
    getTenantChat,
    sendTenantChatMessage,
    editTenantChatMessage,
    deleteTenantChatMessageSimple,
    getTenantMembers,
    type TenantWithMemberCount,
    type TenantChatMessage,
    type TenantMember
} from '../../../../../../lib/api'

export default function TenantChatPage() {
    const router = useRouter()
    const params = useParams()
    const t = useT()
    const notification = useNotification()
    const tenantId = params.id as string

    const [tenant, setTenant] = useState<TenantWithMemberCount | null>(null)
    const [messages, setMessages] = useState<TenantChatMessage[]>([])
    const [members, setMembers] = useState<TenantMember[]>([])
    const [myMembership, setMyMembership] = useState<TenantMember | null>(null)
    const [loading, setLoading] = useState(true)
    const [sending, setSending] = useState(false)
    const [message, setMessage] = useState('')
    const [editingMessage, setEditingMessage] = useState<TenantChatMessage | null>(null)
    const [editContent, setEditContent] = useState('')
    const [showSidebar, setShowSidebar] = useState(true)

    const messagesEndRef = useRef<HTMLDivElement>(null)
    const inputRef = useRef<HTMLInputElement>(null)
    const user = getUser()

    const scrollToBottom = () => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
    }

    const loadData = useCallback(async () => {
        setLoading(true)
        try {
            const [tenantData, messagesData, membersData] = await Promise.all([
                getTenantByID(tenantId),
                getTenantChat(tenantId, { limit: 100 }),
                getTenantMembers(tenantId)
            ])

            setTenant(tenantData)
            setMessages(messagesData)
            setMembers(membersData)

            if (user) {
                const myMember = membersData.find((m: TenantMember) => m.user_id === user.id)
                setMyMembership(myMember || null)

                if (!myMember) {
                    notification.error(t('tenant.permissions.title'), t('tenant.permissions.required'))
                    router.push(`/${params.locale}/tenants/${tenantId}`)
                }
            }
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoading(false)
        }
    }, [tenantId, user, notification, t, router, params.locale])

    useEffect(() => {
        loadData()
    }, [loadData])

    useEffect(() => {
        scrollToBottom()
    }, [messages])

    // Poll for new messages every 3 seconds
    useEffect(() => {
        if (!myMembership) return

        const interval = setInterval(async () => {
            try {
                const newMessages = await getTenantChat(tenantId, { limit: 100 })
                setMessages(newMessages)
            } catch {
                // Silently fail
            }
        }, 3000)

        return () => clearInterval(interval)
    }, [tenantId, myMembership])

    const handleSend = async () => {
        if (!message.trim() || sending) return

        setSending(true)
        try {
            await sendTenantChatMessage(tenantId, { content: message.trim() })
            setMessage('')
            inputRef.current?.focus()

            // Reload messages
            const newMessages = await getTenantChat(tenantId, { limit: 100 })
            setMessages(newMessages)
        } catch (err) {
            notification.error(t('chat.error.send'), String(err))
        } finally {
            setSending(false)
        }
    }

    const handleKeyPress = (e: React.KeyboardEvent) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault()
            handleSend()
        }
    }

    const handleStartEdit = (msg: TenantChatMessage) => {
        setEditingMessage(msg)
        setEditContent(msg.content)
    }

    const handleCancelEdit = () => {
        setEditingMessage(null)
        setEditContent('')
    }

    const handleSaveEdit = async () => {
        if (!editingMessage || !editContent.trim()) return

        try {
            await editTenantChatMessage(tenantId, editingMessage.id, { content: editContent.trim() })
            setEditingMessage(null)
            setEditContent('')

            // Reload messages
            const newMessages = await getTenantChat(tenantId, { limit: 100 })
            setMessages(newMessages)
        } catch (err) {
            notification.error(t('chat.error.edit'), String(err))
        }
    }

    const handleDelete = async (messageId: string) => {
        if (!confirm(t('chat.confirm.delete'))) return

        try {
            await deleteTenantChatMessageSimple(tenantId, messageId)

            // Reload messages
            const newMessages = await getTenantChat(tenantId, { limit: 100 })
            setMessages(newMessages)
        } catch (err) {
            notification.error(t('chat.error.delete'), String(err))
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
        const member = members.find((m: TenantMember) => m.user_id === userId)
        return member?.user_name || 'Unknown User'
    }

    const canModerate = myMembership && ['owner', 'admin', 'moderator'].includes(myMembership.role)

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
                    Loading...
                </div>
            </AuthGuard>
        )
    }

    if (!tenant || !myMembership) {
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
                    <p>{t('tenant.permissions.required')}</p>
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
                            onClick={() => router.push(`/${params.locale}/tenants/${tenantId}`)}
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
                                💬 {tenant.name}
                            </h1>
                            <span style={{ fontSize: 12, color: '#6b7280' }}>
                                {t('tenant.card.members', { count: members.length })}
                            </span>
                        </div>
                    </div>

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
                                    <p style={{ fontSize: 16, margin: 0 }}>{t('tenant.chat.noMessages')}</p>
                                </div>
                            ) : (
                                <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                                    {messages.map((msg) => {
                                        const isOwn = msg.user_id === user?.id
                                        const isDeleted = msg.is_deleted

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

                                                {editingMessage?.id === msg.id ? (
                                                    <div style={{
                                                        backgroundColor: 'white',
                                                        borderRadius: 12,
                                                        padding: 16,
                                                        border: '2px solid #3b82f6',
                                                        width: '100%'
                                                    }}>
                                                        <input
                                                            type="text"
                                                            value={editContent}
                                                            onChange={(e) => setEditContent(e.target.value)}
                                                            style={{
                                                                width: '100%',
                                                                padding: '8px 12px',
                                                                border: '1px solid #e5e7eb',
                                                                borderRadius: 6,
                                                                fontSize: 14,
                                                                marginBottom: 12,
                                                                boxSizing: 'border-box'
                                                            }}
                                                            autoFocus
                                                        />
                                                        <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                                                            <button
                                                                onClick={handleCancelEdit}
                                                                style={{
                                                                    padding: '6px 12px',
                                                                    backgroundColor: '#f3f4f6',
                                                                    color: '#374151',
                                                                    border: 'none',
                                                                    borderRadius: 6,
                                                                    cursor: 'pointer',
                                                                    fontSize: 13
                                                                }}
                                                            >
                                                                {t('chat.modal.edit.cancel')}
                                                            </button>
                                                            <button
                                                                onClick={handleSaveEdit}
                                                                style={{
                                                                    padding: '6px 12px',
                                                                    backgroundColor: '#3b82f6',
                                                                    color: 'white',
                                                                    border: 'none',
                                                                    borderRadius: 6,
                                                                    cursor: 'pointer',
                                                                    fontSize: 13,
                                                                    fontWeight: 500
                                                                }}
                                                            >
                                                                {t('chat.modal.edit.save')}
                                                            </button>
                                                        </div>
                                                    </div>
                                                ) : (
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
                                                            {isDeleted ? t('chat.message.redacted') : msg.content}
                                                        </p>
                                                        {msg.is_edited && !isDeleted && (
                                                            <span style={{
                                                                fontSize: 10,
                                                                opacity: 0.7,
                                                                marginLeft: 8
                                                            }}>
                                                                {t('chat.edited')}
                                                            </span>
                                                        )}

                                                        {/* Actions */}
                                                        {!isDeleted && (isOwn || canModerate) && (
                                                            <div style={{
                                                                display: 'flex',
                                                                gap: 4,
                                                                marginTop: 8,
                                                                opacity: 0.7
                                                            }}>
                                                                {isOwn && (
                                                                    <button
                                                                        onClick={() => handleStartEdit(msg)}
                                                                        style={{
                                                                            background: 'none',
                                                                            border: 'none',
                                                                            color: isOwn ? 'white' : '#6b7280',
                                                                            cursor: 'pointer',
                                                                            fontSize: 11,
                                                                            padding: '2px 6px'
                                                                        }}
                                                                    >
                                                                        {t('chat.action.edit')}
                                                                    </button>
                                                                )}
                                                                {(isOwn || canModerate) && (
                                                                    <button
                                                                        onClick={() => handleDelete(msg.id)}
                                                                        style={{
                                                                            background: 'none',
                                                                            border: 'none',
                                                                            color: isOwn ? 'white' : '#dc2626',
                                                                            cursor: 'pointer',
                                                                            fontSize: 11,
                                                                            padding: '2px 6px'
                                                                        }}
                                                                    >
                                                                        {t('chat.action.delete')}
                                                                    </button>
                                                                )}
                                                            </div>
                                                        )}
                                                    </div>
                                                )}

                                                <span style={{
                                                    fontSize: 11,
                                                    color: '#9ca3af',
                                                    marginTop: 4,
                                                    marginLeft: isOwn ? 0 : 12,
                                                    marginRight: isOwn ? 12 : 0
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

                        {/* Input */}
                        <div style={{
                            padding: '16px 24px',
                            backgroundColor: 'white',
                            borderTop: '1px solid #e5e7eb'
                        }}>
                            <div style={{
                                display: 'flex',
                                gap: 12,
                                alignItems: 'center'
                            }}>
                                <input
                                    ref={inputRef}
                                    type="text"
                                    placeholder={t('tenant.chat.placeholder')}
                                    value={message}
                                    onChange={(e) => setMessage(e.target.value)}
                                    onKeyPress={handleKeyPress}
                                    disabled={sending}
                                    style={{
                                        flex: 1,
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 24,
                                        fontSize: 15,
                                        outline: 'none',
                                        backgroundColor: sending ? '#f9fafb' : 'white'
                                    }}
                                />
                                <button
                                    onClick={handleSend}
                                    disabled={!message.trim() || sending}
                                    style={{
                                        width: 48,
                                        height: 48,
                                        borderRadius: 24,
                                        backgroundColor: !message.trim() || sending ? '#e5e7eb' : '#3b82f6',
                                        color: 'white',
                                        border: 'none',
                                        cursor: !message.trim() || sending ? 'not-allowed' : 'pointer',
                                        fontSize: 20,
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'center'
                                    }}
                                >
                                    {sending ? '...' : '→'}
                                </button>
                            </div>
                        </div>
                    </div>

                    {/* Sidebar */}
                    {showSidebar && (
                        <div style={{
                            width: 280,
                            backgroundColor: 'white',
                            borderLeft: '1px solid #e5e7eb',
                            padding: 16,
                            overflowY: 'auto'
                        }}>
                            <h3 style={{
                                margin: '0 0 16px',
                                fontSize: 14,
                                fontWeight: 600,
                                color: '#6b7280',
                                textTransform: 'uppercase',
                                letterSpacing: 0.5
                            }}>
                                {t('tenant.members.title')} — {members.length}
                            </h3>

                            <div style={{ display: 'flex', flexDirection: 'column', gap: 8 }}>
                                {members.map((member) => (
                                    <div
                                        key={member.id}
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 12,
                                            padding: 8,
                                            borderRadius: 8,
                                            backgroundColor: member.user_id === user?.id ? '#eff6ff' : 'transparent'
                                        }}
                                    >
                                        <div style={{
                                            width: 32,
                                            height: 32,
                                            borderRadius: 16,
                                            backgroundColor: '#e5e7eb',
                                            display: 'flex',
                                            justifyContent: 'center',
                                            alignItems: 'center',
                                            fontSize: 14,
                                            fontWeight: 600,
                                            color: '#6b7280'
                                        }}>
                                            {(member.user_name || 'U')[0].toUpperCase()}
                                        </div>
                                        <div style={{ flex: 1, minWidth: 0 }}>
                                            <div style={{
                                                fontSize: 14,
                                                fontWeight: 500,
                                                color: '#1f2937',
                                                whiteSpace: 'nowrap',
                                                overflow: 'hidden',
                                                textOverflow: 'ellipsis'
                                            }}>
                                                {member.user_name || 'Unknown'}
                                                {member.user_id === user?.id && (
                                                    <span style={{ color: '#6b7280', fontWeight: 400 }}>
                                                        {' '}({t('chat.user.you')})
                                                    </span>
                                                )}
                                            </div>
                                            <div style={{
                                                fontSize: 11,
                                                color: member.role === 'owner' ? '#92400e' :
                                                    member.role === 'admin' ? '#991b1b' :
                                                        member.role === 'moderator' ? '#1d4ed8' :
                                                            member.role === 'vip' ? '#7c3aed' : '#6b7280'
                                            }}>
                                                {t(`tenant.members.roles.${member.role}`)}
                                            </div>
                                        </div>
                                    </div>
                                ))}
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </AuthGuard>
    )
}
