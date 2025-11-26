'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useT } from '../../../../../lib/i18n-context'
import { useNotification } from '../../../../../contexts/NotificationContext'
import { getUser } from '../../../../../lib/auth'
import { useWebSocket } from '../../../../../lib/websocket'
import AuthGuard from '../../../../../components/AuthGuard'
import Footer from '../../../../../components/Footer'
import {
    getTenantByID,
    getTenantMembers,
    getTenantAnnouncements,
    getTenantInvites,
    joinTenant,
    leaveTenant,
    updateMemberRole,
    createTenantInvite,
    revokeTenantInvite,
    createTenantAnnouncement,
    deleteTenantAnnouncement,
    toggleTenantAnnouncementPin,
    removeTenantMember,
    banTenantMember,
    type TenantWithMemberCount,
    type TenantMember,
    type TenantAnnouncement,
    type TenantInvite,
    type TenantRole
} from '../../../../../lib/api'

type TabType = 'overview' | 'members' | 'announcements' | 'invites'

export default function TenantDetailPage() {
    const router = useRouter()
    const params = useParams()
    const t = useT()
    const notification = useNotification()
    const tenantId = params.id as string

    const [tenant, setTenant] = useState<TenantWithMemberCount | null>(null)
    const [members, setMembers] = useState<TenantMember[]>([])
    const [announcements, setAnnouncements] = useState<TenantAnnouncement[]>([])
    const [invites, setInvites] = useState<TenantInvite[]>([])
    const [myMembership, setMyMembership] = useState<TenantMember | null>(null)
    const [activeTab, setActiveTab] = useState<TabType>('overview')
    const [loading, setLoading] = useState(true)
    const [actionLoading, setActionLoading] = useState(false)

    // Modals
    const [showJoinModal, setShowJoinModal] = useState(false)
    const [showLeaveModal, setShowLeaveModal] = useState(false)
    const [showInviteModal, setShowInviteModal] = useState(false)
    const [showAnnouncementModal, setShowAnnouncementModal] = useState(false)
    const [password, setPassword] = useState('')

    // Forms
    const [inviteExpiry, setInviteExpiry] = useState('24h')
    const [inviteMaxUses, setInviteMaxUses] = useState('10')
    const [announcementTitle, setAnnouncementTitle] = useState('')
    const [announcementContent, setAnnouncementContent] = useState('')
    const [announcementPriority, setAnnouncementPriority] = useState<'low' | 'normal' | 'high' | 'urgent'>('normal')

    // Online status tracking
    const [onlineUsers, setOnlineUsers] = useState<Set<string>>(new Set())

    const user = getUser()
    const ws = useWebSocket()

    const loadTenantData = useCallback(async () => {
        setLoading(true)
        try {
            const tenantData = await getTenantByID(tenantId)
            setTenant(tenantData)

            const membersData = await getTenantMembers(tenantId)
            setMembers(membersData)

            // Find my membership
            if (user) {
                const myMember = membersData.find((m: TenantMember) => m.user_id === user.id)
                setMyMembership(myMember || null)
            }
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoading(false)
        }
    }, [tenantId, user, notification, t])

    const loadAnnouncements = useCallback(async () => {
        try {
            const data = await getTenantAnnouncements(tenantId)
            setAnnouncements(data)
        } catch (err) {
            console.error('Failed to load announcements:', err)
        }
    }, [tenantId])

    const loadInvites = useCallback(async () => {
        try {
            const data = await getTenantInvites(tenantId)
            setInvites(data)
        } catch (err) {
            console.error('Failed to load invites:', err)
        }
    }, [tenantId])

    useEffect(() => {
        loadTenantData()
    }, [loadTenantData])

    // WebSocket presence tracking for tenant members
    useEffect(() => {
        if (!ws.isConnected || !myMembership) return

        // Join tenant room for presence updates
        ws.joinTenantRoom(tenantId)

        // Listen for presence updates
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

        ws.on('presence.update', handlePresence)

        return () => {
            ws.off('presence.update', handlePresence)
            ws.leaveTenantRoom(tenantId)
        }
    }, [ws.isConnected, myMembership, tenantId, ws])

    useEffect(() => {
        if (myMembership && activeTab === 'announcements') {
            loadAnnouncements()
        }
        if (myMembership && canManageInvites && activeTab === 'invites') {
            loadInvites()
        }
    }, [activeTab, myMembership, loadAnnouncements, loadInvites])

    // Permission checks
    const canManageMembers = myMembership && ['owner', 'admin'].includes(myMembership.role)
    const canManageInvites = myMembership && ['owner', 'admin', 'moderator'].includes(myMembership.role)
    const canManageAnnouncements = myMembership && ['owner', 'admin', 'moderator'].includes(myMembership.role)

    const handleJoin = async () => {
        setActionLoading(true)
        try {
            await joinTenant(tenantId, tenant?.access_type === 'password' ? password : undefined)
            notification.success(t('tenant.join.title'), t('tenant.join.success'))
            setShowJoinModal(false)
            setPassword('')
            await loadTenantData()
        } catch (err) {
            notification.error(t('tenant.join.title'), String(err))
        } finally {
            setActionLoading(false)
        }
    }

    const handleLeave = async () => {
        setActionLoading(true)
        try {
            await leaveTenant(tenantId)
            notification.success(t('tenant.leave.title'), t('tenant.leave.success'))
            setShowLeaveModal(false)
            router.push(`/${params.locale}/tenants`)
        } catch (err) {
            notification.error(t('tenant.leave.title'), String(err))
        } finally {
            setActionLoading(false)
        }
    }

    const handleRoleChange = async (memberId: string, newRole: TenantRole) => {
        try {
            await updateMemberRole(tenantId, memberId, newRole)
            notification.success(t('tenant.members.title'), 'Role updated')
            await loadTenantData()
        } catch (err) {
            notification.error(t('tenant.members.title'), String(err))
        }
    }

    const handleKickMember = async (memberId: string) => {
        if (!confirm('Are you sure you want to remove this member?')) return
        try {
            await removeTenantMember(tenantId, memberId)
            notification.success(t('tenant.members.title'), 'Member removed')
            await loadTenantData()
        } catch (err) {
            notification.error(t('tenant.members.title'), String(err))
        }
    }

    const handleBanMember = async (memberId: string) => {
        if (!confirm(t('tenants.members.banConfirmMessage'))) return
        try {
            await banTenantMember(tenantId, memberId)
            notification.success(t('tenants.members.title'), t('tenants.members.banSuccess'))
            await loadTenantData()
        } catch (err) {
            notification.error(t('tenants.members.title'), t('tenants.members.banError'))
        }
    }

    const handleCreateInvite = async () => {
        setActionLoading(true)
        try {
            const expiresIn = inviteExpiry === 'never' ? undefined : parseInt(inviteExpiry)
            const maxUses = inviteMaxUses === 'unlimited' ? undefined : parseInt(inviteMaxUses)
            await createTenantInvite(tenantId, { expiresInHours: expiresIn, maxUses })
            notification.success(t('tenant.invites.title'), 'Invite created')
            setShowInviteModal(false)
            await loadInvites()
        } catch (err) {
            notification.error(t('tenant.invites.title'), String(err))
        } finally {
            setActionLoading(false)
        }
    }

    const handleRevokeInvite = async (inviteId: string) => {
        try {
            await revokeTenantInvite(tenantId, inviteId)
            notification.success(t('tenant.invites.title'), 'Invite revoked')
            await loadInvites()
        } catch (err) {
            notification.error(t('tenant.invites.title'), String(err))
        }
    }

    const handleCopyInviteCode = async (code: string) => {
        try {
            await navigator.clipboard.writeText(code)
            notification.success(t('tenant.invites.copyCode'), t('tenant.invites.copied'))
        } catch {
            notification.error(t('tenant.invites.copyCode'), 'Failed to copy')
        }
    }

    const handleCreateAnnouncement = async () => {
        if (!announcementTitle.trim() || !announcementContent.trim()) return
        setActionLoading(true)
        try {
            await createTenantAnnouncement(tenantId, {
                title: announcementTitle,
                content: announcementContent,
                priority: announcementPriority
            })
            notification.success(t('tenant.announcements.title'), 'Announcement created')
            setShowAnnouncementModal(false)
            setAnnouncementTitle('')
            setAnnouncementContent('')
            setAnnouncementPriority('normal')
            await loadAnnouncements()
        } catch (err) {
            notification.error(t('tenant.announcements.title'), String(err))
        } finally {
            setActionLoading(false)
        }
    }

    const handleDeleteAnnouncement = async (announcementId: string) => {
        if (!confirm('Are you sure you want to delete this announcement?')) return
        try {
            await deleteTenantAnnouncement(tenantId, announcementId)
            notification.success(t('tenant.announcements.title'), 'Announcement deleted')
            await loadAnnouncements()
        } catch (err) {
            notification.error(t('tenant.announcements.title'), String(err))
        }
    }

    const handleTogglePin = async (announcementId: string) => {
        try {
            await toggleTenantAnnouncementPin(tenantId, announcementId)
            await loadAnnouncements()
        } catch (err) {
            notification.error(t('tenant.announcements.title'), String(err))
        }
    }

    const getRoleBadgeStyle = (role: TenantRole) => {
        const colors: Record<TenantRole, { bg: string; text: string }> = {
            owner: { bg: '#fef3c7', text: '#92400e' },
            admin: { bg: '#fee2e2', text: '#991b1b' },
            moderator: { bg: '#dbeafe', text: '#1d4ed8' },
            vip: { bg: '#f3e8ff', text: '#7c3aed' },
            member: { bg: '#f3f4f6', text: '#374151' }
        }
        return colors[role] || colors.member
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
                    Loading...
                </div>
            </AuthGuard>
        )
    }

    if (!tenant) {
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
                    <div style={{ fontSize: 48, marginBottom: 16 }}>😕</div>
                    <p>Tenant not found</p>
                    <button
                        onClick={() => router.push(`/${params.locale}/tenants`)}
                        style={{
                            marginTop: 16,
                            padding: '12px 24px',
                            backgroundColor: '#3b82f6',
                            color: 'white',
                            border: 'none',
                            borderRadius: 8,
                            cursor: 'pointer'
                        }}
                    >
                        Back to Tenants
                    </button>
                </div>
            </AuthGuard>
        )
    }

    return (
        <AuthGuard>
            <div style={{
                minHeight: '100vh',
                backgroundColor: '#f9fafb',
                fontFamily: 'system-ui, -apple-system, sans-serif'
            }}>
                <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px 24px 100px' }}>
                    {/* Header */}
                    <div style={{ marginBottom: 32 }}>
                        <button
                            onClick={() => router.push(`/${params.locale}/tenants`)}
                            style={{
                                background: 'none',
                                border: 'none',
                                color: '#6b7280',
                                cursor: 'pointer',
                                fontSize: 14,
                                marginBottom: 16,
                                padding: 0,
                                display: 'flex',
                                alignItems: 'center',
                                gap: 4
                            }}
                        >
                            ← {t('chat.header.back')}
                        </button>

                        <div style={{
                            backgroundColor: 'white',
                            borderRadius: 16,
                            padding: 32,
                            border: '1px solid #e5e7eb',
                            boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                        }}>
                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                <div>
                                    <h1 style={{ margin: 0, fontSize: 28, fontWeight: 700, color: '#1f2937' }}>
                                        {tenant.name}
                                    </h1>
                                    {tenant.description && (
                                        <p style={{ margin: '12px 0 0', color: '#6b7280', fontSize: 16, maxWidth: 600 }}>
                                            {tenant.description}
                                        </p>
                                    )}
                                    <div style={{ display: 'flex', gap: 12, marginTop: 16, flexWrap: 'wrap' }}>
                                        <span style={{
                                            padding: '6px 14px',
                                            borderRadius: 20,
                                            fontSize: 13,
                                            fontWeight: 500,
                                            backgroundColor: tenant.visibility === 'public' ? '#dcfce7' : '#f3f4f6',
                                            color: tenant.visibility === 'public' ? '#166534' : '#6b7280'
                                        }}>
                                            {t(`tenant.card.visibility.${tenant.visibility}`)}
                                        </span>
                                        <span style={{
                                            padding: '6px 14px',
                                            borderRadius: 20,
                                            fontSize: 13,
                                            fontWeight: 500,
                                            backgroundColor: '#dbeafe',
                                            color: '#1d4ed8'
                                        }}>
                                            {t('tenant.card.members', { count: tenant.member_count || members.length })}
                                        </span>
                                        {myMembership && (
                                            <span style={{
                                                padding: '6px 14px',
                                                borderRadius: 20,
                                                fontSize: 13,
                                                fontWeight: 500,
                                                ...getRoleBadgeStyle(myMembership.role)
                                            }}>
                                                {t(`tenant.members.roles.${myMembership.role}`)}
                                            </span>
                                        )}
                                    </div>
                                </div>

                                <div style={{ display: 'flex', gap: 12 }}>
                                    {myMembership ? (
                                        <>
                                            <button
                                                onClick={() => router.push(`/${params.locale}/tenants/${tenantId}/chat`)}
                                                style={{
                                                    padding: '12px 24px',
                                                    backgroundColor: '#3b82f6',
                                                    color: 'white',
                                                    border: 'none',
                                                    borderRadius: 8,
                                                    cursor: 'pointer',
                                                    fontSize: 14,
                                                    fontWeight: 600,
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: 8
                                                }}
                                            >
                                                💬 {t('tenant.chat.title')}
                                            </button>
                                            {canManageMembers && (
                                                <button
                                                    onClick={() => router.push(`/${params.locale}/tenants/${tenantId}/settings`)}
                                                    style={{
                                                        padding: '12px 24px',
                                                        backgroundColor: '#f3f4f6',
                                                        color: '#374151',
                                                        border: '1px solid #e5e7eb',
                                                        borderRadius: 8,
                                                        cursor: 'pointer',
                                                        fontSize: 14,
                                                        fontWeight: 600,
                                                        display: 'flex',
                                                        alignItems: 'center',
                                                        gap: 8
                                                    }}
                                                >
                                                    ⚙️ {t('tenant.settings.title')}
                                                </button>
                                            )}
                                            {myMembership.role !== 'owner' && (
                                                <button
                                                    onClick={() => setShowLeaveModal(true)}
                                                    style={{
                                                        padding: '12px 24px',
                                                        backgroundColor: '#fee2e2',
                                                        color: '#991b1b',
                                                        border: 'none',
                                                        borderRadius: 8,
                                                        cursor: 'pointer',
                                                        fontSize: 14,
                                                        fontWeight: 600
                                                    }}
                                                >
                                                    {t('tenant.leave.button')}
                                                </button>
                                            )}
                                        </>
                                    ) : (
                                        <button
                                            onClick={() => tenant.access_type === 'open' ? handleJoin() : setShowJoinModal(true)}
                                            style={{
                                                padding: '12px 28px',
                                                backgroundColor: '#3b82f6',
                                                color: 'white',
                                                border: 'none',
                                                borderRadius: 8,
                                                cursor: 'pointer',
                                                fontSize: 15,
                                                fontWeight: 600
                                            }}
                                        >
                                            {t('tenant.join.button')}
                                        </button>
                                    )}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Tabs (only for members) */}
                    {myMembership && (
                        <>
                            <div style={{
                                display: 'flex',
                                gap: 8,
                                marginBottom: 24,
                                borderBottom: '2px solid #e5e7eb',
                                paddingBottom: 0
                            }}>
                                {(['overview', 'members', 'announcements', ...(canManageInvites ? ['invites'] : [])] as TabType[]).map(tab => (
                                    <button
                                        key={tab}
                                        onClick={() => setActiveTab(tab)}
                                        style={{
                                            padding: '12px 24px',
                                            background: 'none',
                                            border: 'none',
                                            borderBottom: activeTab === tab ? '2px solid #3b82f6' : '2px solid transparent',
                                            marginBottom: -2,
                                            color: activeTab === tab ? '#3b82f6' : '#6b7280',
                                            cursor: 'pointer',
                                            fontSize: 15,
                                            fontWeight: activeTab === tab ? 600 : 500
                                        }}
                                    >
                                        {tab === 'overview' && '📊'}
                                        {tab === 'members' && '👥'}
                                        {tab === 'announcements' && '📢'}
                                        {tab === 'invites' && '🎟️'}
                                        {' '}
                                        {tab === 'overview' ? 'Overview' :
                                            tab === 'members' ? t('tenant.members.title') :
                                                tab === 'announcements' ? t('tenant.announcements.title') :
                                                    t('tenant.invites.title')}
                                    </button>
                                ))}
                            </div>

                            {/* Tab Content */}
                            {activeTab === 'overview' && (
                                <div style={{
                                    display: 'grid',
                                    gridTemplateColumns: 'repeat(auto-fit, minmax(300px, 1fr))',
                                    gap: 20
                                }}>
                                    {/* Stats Card */}
                                    <div style={{
                                        backgroundColor: 'white',
                                        borderRadius: 12,
                                        padding: 24,
                                        border: '1px solid #e5e7eb'
                                    }}>
                                        <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 600, color: '#374151' }}>
                                            Statistics
                                        </h3>
                                        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                                            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                                <span style={{ color: '#6b7280' }}>Total Members</span>
                                                <span style={{ fontWeight: 600 }}>{members.length}</span>
                                            </div>
                                            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                                <span style={{ color: '#6b7280' }}>Your Role</span>
                                                <span style={{ fontWeight: 600 }}>{t(`tenant.members.roles.${myMembership.role}`)}</span>
                                            </div>
                                            <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                                                <span style={{ color: '#6b7280' }}>Joined</span>
                                                <span style={{ fontWeight: 600 }}>
                                                    {new Date(myMembership.created_at).toLocaleDateString()}
                                                </span>
                                            </div>
                                        </div>
                                    </div>

                                    {/* Quick Actions */}
                                    <div style={{
                                        backgroundColor: 'white',
                                        borderRadius: 12,
                                        padding: 24,
                                        border: '1px solid #e5e7eb'
                                    }}>
                                        <h3 style={{ margin: '0 0 16px', fontSize: 16, fontWeight: 600, color: '#374151' }}>
                                            Quick Actions
                                        </h3>
                                        <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
                                            <button
                                                onClick={() => router.push(`/${params.locale}/tenants/${tenantId}/chat`)}
                                                style={{
                                                    padding: '12px 16px',
                                                    backgroundColor: '#3b82f6',
                                                    color: 'white',
                                                    border: 'none',
                                                    borderRadius: 8,
                                                    cursor: 'pointer',
                                                    fontSize: 14,
                                                    fontWeight: 500,
                                                    textAlign: 'left'
                                                }}
                                            >
                                                💬 {t('tenant.chat.title')}
                                            </button>
                                            {canManageInvites && (
                                                <button
                                                    onClick={() => setShowInviteModal(true)}
                                                    style={{
                                                        padding: '12px 16px',
                                                        backgroundColor: '#8b5cf6',
                                                        color: 'white',
                                                        border: 'none',
                                                        borderRadius: 8,
                                                        cursor: 'pointer',
                                                        fontSize: 14,
                                                        fontWeight: 500,
                                                        textAlign: 'left'
                                                    }}
                                                >
                                                    🎟️ {t('tenant.invites.create')}
                                                </button>
                                            )}
                                        </div>
                                    </div>
                                </div>
                            )}

                            {activeTab === 'members' && (
                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    border: '1px solid #e5e7eb',
                                    overflow: 'hidden'
                                }}>
                                    {/* Online Members Summary */}
                                    <div style={{
                                        padding: '12px 16px',
                                        backgroundColor: '#f9fafb',
                                        borderBottom: '1px solid #e5e7eb',
                                        display: 'flex',
                                        alignItems: 'center',
                                        gap: 8
                                    }}>
                                        <span style={{
                                            width: 8,
                                            height: 8,
                                            borderRadius: '50%',
                                            backgroundColor: '#22c55e'
                                        }} />
                                        <span style={{ fontSize: 13, color: '#22c55e', fontWeight: 500 }}>
                                            {onlineUsers.size} {t('chat.status.online')}
                                        </span>
                                        <span style={{ color: '#9ca3af' }}>•</span>
                                        <span style={{ fontSize: 13, color: '#6b7280' }}>
                                            {members.length} {t('tenants.members')}
                                        </span>
                                    </div>
                                    {members.map((member, idx) => {
                                        const isOnline = onlineUsers.has(member.user_id)
                                        return (
                                            <div
                                                key={member.id}
                                                style={{
                                                    padding: 16,
                                                    display: 'flex',
                                                    justifyContent: 'space-between',
                                                    alignItems: 'center',
                                                    borderBottom: idx < members.length - 1 ? '1px solid #f3f4f6' : 'none'
                                                }}
                                            >
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                                                    <div style={{
                                                        width: 40,
                                                        height: 40,
                                                        borderRadius: 20,
                                                        backgroundColor: '#e5e7eb',
                                                        display: 'flex',
                                                        justifyContent: 'center',
                                                        alignItems: 'center',
                                                        fontSize: 16,
                                                        fontWeight: 600,
                                                        color: '#6b7280',
                                                        position: 'relative'
                                                    }}>
                                                        {(member.user_name || 'U')[0].toUpperCase()}
                                                        {/* Online Status Indicator */}
                                                        <span style={{
                                                            position: 'absolute',
                                                            bottom: 0,
                                                            right: 0,
                                                            width: 12,
                                                            height: 12,
                                                            borderRadius: '50%',
                                                            backgroundColor: isOnline ? '#22c55e' : '#9ca3af',
                                                            border: '2px solid white'
                                                        }} />
                                                    </div>
                                                    <div>
                                                        <div style={{ fontWeight: 500, color: '#1f2937', display: 'flex', alignItems: 'center', gap: 8 }}>
                                                            {member.user_name || 'Unknown User'}
                                                            {member.user_id === user?.id && (
                                                                <span style={{ color: '#6b7280', fontWeight: 400 }}>
                                                                    ({t('chat.user.you')})
                                                                </span>
                                                            )}
                                                        </div>
                                                        <div style={{ fontSize: 12, color: isOnline ? '#22c55e' : '#9ca3af', display: 'flex', alignItems: 'center', gap: 4 }}>
                                                            <span>{isOnline ? t('chat.status.online') : t('chat.status.offline')}</span>
                                                            <span>•</span>
                                                            <span style={{ color: '#9ca3af' }}>Joined {new Date(member.created_at).toLocaleDateString()}</span>
                                                        </div>
                                                    </div>
                                                </div>

                                                <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                                                    <span style={{
                                                        padding: '4px 12px',
                                                        borderRadius: 16,
                                                        fontSize: 12,
                                                        fontWeight: 500,
                                                        backgroundColor: getRoleBadgeStyle(member.role).bg,
                                                        color: getRoleBadgeStyle(member.role).text
                                                    }}>
                                                        {t(`tenant.members.roles.${member.role}`)}
                                                    </span>

                                                    {canManageMembers && member.user_id !== user?.id && member.role !== 'owner' && (
                                                        <div style={{ display: 'flex', gap: 8 }}>
                                                            <select
                                                                value={member.role}
                                                                onChange={(e) => handleRoleChange(member.id, e.target.value as TenantRole)}
                                                                style={{
                                                                    padding: '6px 10px',
                                                                    borderRadius: 6,
                                                                    border: '1px solid #e5e7eb',
                                                                    fontSize: 13,
                                                                    cursor: 'pointer'
                                                                }}
                                                            >
                                                                <option value="admin">{t('tenant.members.roles.admin')}</option>
                                                                <option value="moderator">{t('tenant.members.roles.moderator')}</option>
                                                                <option value="vip">{t('tenant.members.roles.vip')}</option>
                                                                <option value="member">{t('tenant.members.roles.member')}</option>
                                                            </select>
                                                            <button
                                                                onClick={() => handleKickMember(member.id)}
                                                                style={{
                                                                    padding: '6px 12px',
                                                                    backgroundColor: '#fee2e2',
                                                                    color: '#991b1b',
                                                                    border: 'none',
                                                                    borderRadius: 6,
                                                                    cursor: 'pointer',
                                                                    fontSize: 12,
                                                                    fontWeight: 500
                                                                }}
                                                            >
                                                                {t('tenant.members.actions.kick')}
                                                            </button>
                                                            {!member.banned_at && (
                                                                <button
                                                                    onClick={() => handleBanMember(member.id)}
                                                                    style={{
                                                                        padding: '6px 12px',
                                                                        backgroundColor: '#7f1d1d',
                                                                        color: 'white',
                                                                        border: 'none',
                                                                        borderRadius: 6,
                                                                        cursor: 'pointer',
                                                                        fontSize: 12,
                                                                        fontWeight: 500
                                                                    }}
                                                                >
                                                                    {t('tenants.members.ban')}
                                                                </button>
                                                            )}
                                                            {member.banned_at && (
                                                                <span style={{
                                                                    padding: '6px 12px',
                                                                    backgroundColor: '#7f1d1d',
                                                                    color: 'white',
                                                                    borderRadius: 6,
                                                                    fontSize: 12,
                                                                    fontWeight: 500
                                                                }}>
                                                                    {t('tenants.members.banned')}
                                                                </span>
                                                            )}
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        )
                                    })}
                                </div>
                            )}

                            {activeTab === 'announcements' && (
                                <div>
                                    {canManageAnnouncements && (
                                        <button
                                            onClick={() => setShowAnnouncementModal(true)}
                                            style={{
                                                marginBottom: 20,
                                                padding: '12px 24px',
                                                backgroundColor: '#3b82f6',
                                                color: 'white',
                                                border: 'none',
                                                borderRadius: 8,
                                                cursor: 'pointer',
                                                fontSize: 14,
                                                fontWeight: 600
                                            }}
                                        >
                                            + {t('tenant.announcements.create')}
                                        </button>
                                    )}

                                    {announcements.length === 0 ? (
                                        <div style={{
                                            textAlign: 'center',
                                            padding: 60,
                                            backgroundColor: 'white',
                                            borderRadius: 12,
                                            border: '1px solid #e5e7eb',
                                            color: '#6b7280'
                                        }}>
                                            <div style={{ fontSize: 40, marginBottom: 12 }}>📢</div>
                                            <p>{t('tenant.announcements.noAnnouncements')}</p>
                                        </div>
                                    ) : (
                                        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                                            {announcements.map(ann => (
                                                <div
                                                    key={ann.id}
                                                    style={{
                                                        backgroundColor: 'white',
                                                        borderRadius: 12,
                                                        padding: 20,
                                                        border: ann.is_pinned ? '2px solid #fbbf24' : '1px solid #e5e7eb'
                                                    }}
                                                >
                                                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                                                        <div>
                                                            <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                                                {ann.is_pinned && <span>📌</span>}
                                                                <h3 style={{ margin: 0, fontSize: 18, fontWeight: 600, color: '#1f2937' }}>
                                                                    {ann.title}
                                                                </h3>
                                                                <span style={{
                                                                    padding: '2px 8px',
                                                                    borderRadius: 12,
                                                                    fontSize: 11,
                                                                    fontWeight: 500,
                                                                    backgroundColor: ann.priority === 'urgent' ? '#fee2e2' :
                                                                        ann.priority === 'high' ? '#fef3c7' :
                                                                            ann.priority === 'low' ? '#f3f4f6' : '#dbeafe',
                                                                    color: ann.priority === 'urgent' ? '#991b1b' :
                                                                        ann.priority === 'high' ? '#92400e' :
                                                                            ann.priority === 'low' ? '#6b7280' : '#1d4ed8'
                                                                }}>
                                                                    {t(`tenant.announcements.form.priorities.${ann.priority}`)}
                                                                </span>
                                                            </div>
                                                            <p style={{ margin: '12px 0 0', color: '#4b5563', lineHeight: 1.6 }}>
                                                                {ann.content}
                                                            </p>
                                                            <div style={{ marginTop: 12, fontSize: 12, color: '#9ca3af' }}>
                                                                {new Date(ann.created_at).toLocaleString()}
                                                            </div>
                                                        </div>

                                                        {canManageAnnouncements && (
                                                            <div style={{ display: 'flex', gap: 8 }}>
                                                                <button
                                                                    onClick={() => handleTogglePin(ann.id)}
                                                                    style={{
                                                                        padding: '6px 12px',
                                                                        backgroundColor: '#f3f4f6',
                                                                        color: '#374151',
                                                                        border: 'none',
                                                                        borderRadius: 6,
                                                                        cursor: 'pointer',
                                                                        fontSize: 12
                                                                    }}
                                                                >
                                                                    {ann.is_pinned ? t('tenant.announcements.unpin') : t('tenant.announcements.pin')}
                                                                </button>
                                                                <button
                                                                    onClick={() => handleDeleteAnnouncement(ann.id)}
                                                                    style={{
                                                                        padding: '6px 12px',
                                                                        backgroundColor: '#fee2e2',
                                                                        color: '#991b1b',
                                                                        border: 'none',
                                                                        borderRadius: 6,
                                                                        cursor: 'pointer',
                                                                        fontSize: 12
                                                                    }}
                                                                >
                                                                    {t('tenant.announcements.delete')}
                                                                </button>
                                                            </div>
                                                        )}
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            )}

                            {activeTab === 'invites' && canManageInvites && (
                                <div>
                                    <button
                                        onClick={() => setShowInviteModal(true)}
                                        style={{
                                            marginBottom: 20,
                                            padding: '12px 24px',
                                            backgroundColor: '#8b5cf6',
                                            color: 'white',
                                            border: 'none',
                                            borderRadius: 8,
                                            cursor: 'pointer',
                                            fontSize: 14,
                                            fontWeight: 600
                                        }}
                                    >
                                        + {t('tenant.invites.create')}
                                    </button>

                                    {invites.length === 0 ? (
                                        <div style={{
                                            textAlign: 'center',
                                            padding: 60,
                                            backgroundColor: 'white',
                                            borderRadius: 12,
                                            border: '1px solid #e5e7eb',
                                            color: '#6b7280'
                                        }}>
                                            <div style={{ fontSize: 40, marginBottom: 12 }}>🎟️</div>
                                            <p>No active invites</p>
                                        </div>
                                    ) : (
                                        <div style={{
                                            backgroundColor: 'white',
                                            borderRadius: 12,
                                            border: '1px solid #e5e7eb',
                                            overflow: 'hidden'
                                        }}>
                                            {invites.map((invite, idx) => (
                                                <div
                                                    key={invite.id}
                                                    style={{
                                                        padding: 16,
                                                        display: 'flex',
                                                        justifyContent: 'space-between',
                                                        alignItems: 'center',
                                                        borderBottom: idx < invites.length - 1 ? '1px solid #f3f4f6' : 'none'
                                                    }}
                                                >
                                                    <div>
                                                        <div style={{
                                                            fontFamily: 'monospace',
                                                            fontSize: 18,
                                                            fontWeight: 600,
                                                            letterSpacing: 2,
                                                            color: '#1f2937'
                                                        }}>
                                                            {invite.code}
                                                        </div>
                                                        <div style={{ fontSize: 12, color: '#6b7280', marginTop: 4 }}>
                                                            {t('tenant.invites.uses', {
                                                                used: invite.use_count,
                                                                max: invite.max_uses || '∞'
                                                            })}
                                                            {invite.expires_at && (
                                                                <> • {t('tenant.invites.expiresAt', {
                                                                    date: new Date(invite.expires_at).toLocaleDateString()
                                                                })}</>
                                                            )}
                                                        </div>
                                                    </div>
                                                    <div style={{ display: 'flex', gap: 8 }}>
                                                        <button
                                                            onClick={() => handleCopyInviteCode(invite.code)}
                                                            style={{
                                                                padding: '8px 16px',
                                                                backgroundColor: '#3b82f6',
                                                                color: 'white',
                                                                border: 'none',
                                                                borderRadius: 6,
                                                                cursor: 'pointer',
                                                                fontSize: 13,
                                                                fontWeight: 500
                                                            }}
                                                        >
                                                            📋 {t('tenant.invites.copyCode')}
                                                        </button>
                                                        <button
                                                            onClick={() => handleRevokeInvite(invite.id)}
                                                            style={{
                                                                padding: '8px 16px',
                                                                backgroundColor: '#fee2e2',
                                                                color: '#991b1b',
                                                                border: 'none',
                                                                borderRadius: 6,
                                                                cursor: 'pointer',
                                                                fontSize: 13,
                                                                fontWeight: 500
                                                            }}
                                                        >
                                                            {t('tenant.invites.revoke')}
                                                        </button>
                                                    </div>
                                                </div>
                                            ))}
                                        </div>
                                    )}
                                </div>
                            )}
                        </>
                    )}
                </div>

                {/* Join Modal */}
                {showJoinModal && (
                    <div
                        style={{
                            position: 'fixed',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            backgroundColor: 'rgba(0,0,0,0.5)',
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            zIndex: 1000
                        }}
                        onClick={() => setShowJoinModal(false)}
                    >
                        <div
                            style={{
                                backgroundColor: 'white',
                                borderRadius: 16,
                                padding: 32,
                                width: '100%',
                                maxWidth: 400
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h2 style={{ margin: '0 0 16px', fontSize: 20, fontWeight: 600 }}>
                                {t('tenant.join.title')}
                            </h2>
                            {tenant.access_type === 'password' && (
                                <>
                                    <p style={{ margin: '0 0 16px', color: '#6b7280' }}>
                                        This tenant requires a password to join.
                                    </p>
                                    <input
                                        type="password"
                                        placeholder={t('tenant.join.passwordPlaceholder')}
                                        value={password}
                                        onChange={(e) => setPassword(e.target.value)}
                                        style={{
                                            width: '100%',
                                            padding: '12px 16px',
                                            border: '1px solid #e5e7eb',
                                            borderRadius: 8,
                                            fontSize: 15,
                                            marginBottom: 24,
                                            boxSizing: 'border-box'
                                        }}
                                    />
                                </>
                            )}
                            <div style={{ display: 'flex', gap: 12 }}>
                                <button
                                    onClick={() => setShowJoinModal(false)}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#f3f4f6',
                                        color: '#374151',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 500
                                    }}
                                >
                                    {t('chat.modal.edit.cancel')}
                                </button>
                                <button
                                    onClick={handleJoin}
                                    disabled={actionLoading || (tenant.access_type === 'password' && !password)}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#3b82f6',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 600
                                    }}
                                >
                                    {actionLoading ? '...' : t('tenant.join.button')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* Leave Modal */}
                {showLeaveModal && (
                    <div
                        style={{
                            position: 'fixed',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            backgroundColor: 'rgba(0,0,0,0.5)',
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            zIndex: 1000
                        }}
                        onClick={() => setShowLeaveModal(false)}
                    >
                        <div
                            style={{
                                backgroundColor: 'white',
                                borderRadius: 16,
                                padding: 32,
                                width: '100%',
                                maxWidth: 400
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h2 style={{ margin: '0 0 16px', fontSize: 20, fontWeight: 600 }}>
                                {t('tenant.leave.title')}
                            </h2>
                            <p style={{ margin: '0 0 24px', color: '#6b7280' }}>
                                {t('tenant.leave.confirm')}
                            </p>
                            <div style={{ display: 'flex', gap: 12 }}>
                                <button
                                    onClick={() => setShowLeaveModal(false)}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#f3f4f6',
                                        color: '#374151',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 500
                                    }}
                                >
                                    {t('chat.modal.edit.cancel')}
                                </button>
                                <button
                                    onClick={handleLeave}
                                    disabled={actionLoading}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#dc2626',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 600
                                    }}
                                >
                                    {actionLoading ? '...' : t('tenant.leave.button')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* Create Invite Modal */}
                {showInviteModal && (
                    <div
                        style={{
                            position: 'fixed',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            backgroundColor: 'rgba(0,0,0,0.5)',
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            zIndex: 1000
                        }}
                        onClick={() => setShowInviteModal(false)}
                    >
                        <div
                            style={{
                                backgroundColor: 'white',
                                borderRadius: 16,
                                padding: 32,
                                width: '100%',
                                maxWidth: 400
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h2 style={{ margin: '0 0 24px', fontSize: 20, fontWeight: 600 }}>
                                🎟️ {t('tenant.invites.create')}
                            </h2>

                            <div style={{ marginBottom: 20 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500, color: '#374151' }}>
                                    {t('tenant.invites.settings.expiry')}
                                </label>
                                <select
                                    value={inviteExpiry}
                                    onChange={(e) => setInviteExpiry(e.target.value)}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        boxSizing: 'border-box'
                                    }}
                                >
                                    <option value="1">{t('tenant.invites.settings.hours', { count: 1 })}</option>
                                    <option value="24">{t('tenant.invites.settings.days', { count: 1 })}</option>
                                    <option value="168">{t('tenant.invites.settings.days', { count: 7 })}</option>
                                    <option value="720">{t('tenant.invites.settings.days', { count: 30 })}</option>
                                    <option value="never">{t('tenant.invites.settings.noExpiry')}</option>
                                </select>
                            </div>

                            <div style={{ marginBottom: 24 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500, color: '#374151' }}>
                                    {t('tenant.invites.settings.maxUses')}
                                </label>
                                <select
                                    value={inviteMaxUses}
                                    onChange={(e) => setInviteMaxUses(e.target.value)}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        boxSizing: 'border-box'
                                    }}
                                >
                                    <option value="1">1</option>
                                    <option value="5">5</option>
                                    <option value="10">10</option>
                                    <option value="25">25</option>
                                    <option value="50">50</option>
                                    <option value="100">100</option>
                                    <option value="unlimited">{t('tenant.invites.settings.noLimit')}</option>
                                </select>
                            </div>

                            <div style={{ display: 'flex', gap: 12 }}>
                                <button
                                    onClick={() => setShowInviteModal(false)}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#f3f4f6',
                                        color: '#374151',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 500
                                    }}
                                >
                                    {t('chat.modal.edit.cancel')}
                                </button>
                                <button
                                    onClick={handleCreateInvite}
                                    disabled={actionLoading}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#8b5cf6',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 600
                                    }}
                                >
                                    {actionLoading ? '...' : t('tenant.invites.create')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                {/* Create Announcement Modal */}
                {showAnnouncementModal && (
                    <div
                        style={{
                            position: 'fixed',
                            top: 0,
                            left: 0,
                            right: 0,
                            bottom: 0,
                            backgroundColor: 'rgba(0,0,0,0.5)',
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            zIndex: 1000
                        }}
                        onClick={() => setShowAnnouncementModal(false)}
                    >
                        <div
                            style={{
                                backgroundColor: 'white',
                                borderRadius: 16,
                                padding: 32,
                                width: '100%',
                                maxWidth: 500
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h2 style={{ margin: '0 0 24px', fontSize: 20, fontWeight: 600 }}>
                                📢 {t('tenant.announcements.create')}
                            </h2>

                            <div style={{ marginBottom: 16 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500, color: '#374151' }}>
                                    {t('tenant.announcements.form.title')}
                                </label>
                                <input
                                    type="text"
                                    placeholder={t('tenant.announcements.form.titlePlaceholder')}
                                    value={announcementTitle}
                                    onChange={(e) => setAnnouncementTitle(e.target.value)}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        boxSizing: 'border-box'
                                    }}
                                />
                            </div>

                            <div style={{ marginBottom: 16 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500, color: '#374151' }}>
                                    {t('tenant.announcements.form.content')}
                                </label>
                                <textarea
                                    placeholder={t('tenant.announcements.form.contentPlaceholder')}
                                    value={announcementContent}
                                    onChange={(e) => setAnnouncementContent(e.target.value)}
                                    rows={5}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        boxSizing: 'border-box',
                                        resize: 'vertical'
                                    }}
                                />
                            </div>

                            <div style={{ marginBottom: 24 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500, color: '#374151' }}>
                                    {t('tenant.announcements.form.priority')}
                                </label>
                                <select
                                    value={announcementPriority}
                                    onChange={(e) => setAnnouncementPriority(e.target.value as any)}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        boxSizing: 'border-box'
                                    }}
                                >
                                    <option value="low">{t('tenant.announcements.form.priorities.low')}</option>
                                    <option value="normal">{t('tenant.announcements.form.priorities.normal')}</option>
                                    <option value="high">{t('tenant.announcements.form.priorities.high')}</option>
                                    <option value="urgent">{t('tenant.announcements.form.priorities.urgent')}</option>
                                </select>
                            </div>

                            <div style={{ display: 'flex', gap: 12 }}>
                                <button
                                    onClick={() => setShowAnnouncementModal(false)}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: '#f3f4f6',
                                        color: '#374151',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: 'pointer',
                                        fontSize: 15,
                                        fontWeight: 500
                                    }}
                                >
                                    {t('chat.modal.edit.cancel')}
                                </button>
                                <button
                                    onClick={handleCreateAnnouncement}
                                    disabled={actionLoading || !announcementTitle.trim() || !announcementContent.trim()}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: (!announcementTitle.trim() || !announcementContent.trim()) ? '#9ca3af' : '#3b82f6',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: (!announcementTitle.trim() || !announcementContent.trim()) ? 'not-allowed' : 'pointer',
                                        fontSize: 15,
                                        fontWeight: 600
                                    }}
                                >
                                    {actionLoading ? '...' : t('tenant.announcements.create')}
                                </button>
                            </div>
                        </div>
                    </div>
                )}

                <Footer />
            </div>
        </AuthGuard>
    )
}
