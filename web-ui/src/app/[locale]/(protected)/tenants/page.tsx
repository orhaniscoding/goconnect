'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { useT } from '../../../../lib/i18n-context'
import { useNotification } from '../../../../contexts/NotificationContext'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'
import {
    discoverTenants,
    getMyTenants,
    joinTenant,
    useTenantInvite,
    type TenantWithMemberCount
} from '../../../../lib/api'

type TabType = 'discover' | 'my-tenants'

export default function TenantsPage() {
    const router = useRouter()
    const params = useParams()
    const t = useT()
    const notification = useNotification()

    const [activeTab, setActiveTab] = useState<TabType>('discover')
    const [search, setSearch] = useState('')
    const [tenants, setTenants] = useState<TenantWithMemberCount[]>([])
    const [myTenants, setMyTenants] = useState<TenantWithMemberCount[]>([])
    const [loading, setLoading] = useState(true)
    const [loadingMore, setLoadingMore] = useState(false)
    const [inviteCode, setInviteCode] = useState('')
    const [joiningTenant, setJoiningTenant] = useState<string | null>(null)
    const [showInviteModal, setShowInviteModal] = useState(false)
    const [nextCursor, setNextCursor] = useState<string | undefined>(undefined)

    const loadDiscoverTenants = useCallback(async () => {
        setLoading(true)
        setNextCursor(undefined)
        try {
            const result = await discoverTenants({
                search: search || undefined,
                limit: 50
            })
            setTenants(result.data)
            setNextCursor(result.next_cursor)
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoading(false)
        }
    }, [search, notification, t])

    const loadMoreTenants = useCallback(async () => {
        if (!nextCursor) return
        setLoadingMore(true)
        try {
            const result = await discoverTenants({
                search: search || undefined,
                limit: 50,
                cursor: nextCursor
            })
            setTenants(prev => [...prev, ...result.data])
            setNextCursor(result.next_cursor)
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoadingMore(false)
        }
    }, [search, nextCursor, notification, t])

    const loadMyTenants = useCallback(async () => {
        setLoading(true)
        try {
            const data = await getMyTenants()
            setMyTenants(data)
        } catch (err) {
            notification.error(t('error.generic'), String(err))
        } finally {
            setLoading(false)
        }
    }, [notification, t])

    useEffect(() => {
        if (activeTab === 'discover') {
            loadDiscoverTenants()
        } else {
            loadMyTenants()
        }
    }, [activeTab, loadDiscoverTenants, loadMyTenants])

    const handleJoinTenant = async (tenantId: string, accessType: string) => {
        if (accessType === 'invite_only') {
            notification.info(t('tenant.join.title'), t('tenant.join.error.invalidCode'))
            return
        }

        if (accessType === 'password') {
            // Navigate to tenant detail page for password entry
            router.push(`/${params.locale}/tenants/${tenantId}`)
            return
        }

        setJoiningTenant(tenantId)
        try {
            await joinTenant(tenantId)
            notification.success(t('tenant.join.title'), t('tenant.join.success'))
            // Reload lists
            await loadDiscoverTenants()
            await loadMyTenants()
        } catch (err) {
            notification.error(t('tenant.join.title'), String(err))
        } finally {
            setJoiningTenant(null)
        }
    }

    const handleUseInviteCode = async () => {
        if (!inviteCode.trim()) return

        setJoiningTenant('invite')
        try {
            await useTenantInvite(inviteCode.trim())
            notification.success(t('tenant.join.title'), t('tenant.join.success'))
            setInviteCode('')
            setShowInviteModal(false)
            // Reload lists
            await loadDiscoverTenants()
            await loadMyTenants()
        } catch (err) {
            notification.error(t('tenant.join.title'), String(err))
        } finally {
            setJoiningTenant(null)
        }
    }

    const renderTenantCard = (tenant: TenantWithMemberCount, isMember: boolean = false) => (
        <div
            key={tenant.id}
            style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 20,
                border: '1px solid #e5e7eb',
                cursor: 'pointer',
                transition: 'all 0.2s ease',
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
            }}
            onClick={() => router.push(`/${params.locale}/tenants/${tenant.id}`)}
            onMouseOver={(e) => {
                e.currentTarget.style.boxShadow = '0 4px 6px rgba(0,0,0,0.1)'
                e.currentTarget.style.transform = 'translateY(-2px)'
            }}
            onMouseOut={(e) => {
                e.currentTarget.style.boxShadow = '0 1px 3px rgba(0,0,0,0.1)'
                e.currentTarget.style.transform = 'translateY(0)'
            }}
        >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
                <div style={{ flex: 1 }}>
                    <h3 style={{ margin: 0, fontSize: 18, fontWeight: 600, color: '#1f2937' }}>
                        {tenant.name}
                    </h3>
                    {tenant.description && (
                        <p style={{ margin: '8px 0 0', color: '#6b7280', fontSize: 14, lineHeight: 1.5 }}>
                            {tenant.description.length > 100
                                ? `${tenant.description.substring(0, 100)}...`
                                : tenant.description}
                        </p>
                    )}
                </div>
                <div style={{ display: 'flex', gap: 8, marginLeft: 16, flexShrink: 0 }}>
                    <span style={{
                        padding: '4px 10px',
                        borderRadius: 20,
                        fontSize: 12,
                        fontWeight: 500,
                        backgroundColor: tenant.visibility === 'public' ? '#dcfce7' : '#f3f4f6',
                        color: tenant.visibility === 'public' ? '#166534' : '#6b7280'
                    }}>
                        {t(`tenant.card.visibility.${tenant.visibility}`)}
                    </span>
                    <span style={{
                        padding: '4px 10px',
                        borderRadius: 20,
                        fontSize: 12,
                        fontWeight: 500,
                        backgroundColor: tenant.access_type === 'open' ? '#dbeafe' :
                            tenant.access_type === 'password' ? '#fef3c7' : '#fce7f3',
                        color: tenant.access_type === 'open' ? '#1d4ed8' :
                            tenant.access_type === 'password' ? '#92400e' : '#be185d'
                    }}>
                        {t(`tenant.card.accessType.${tenant.access_type}`)}
                    </span>
                </div>
            </div>

            <div style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                marginTop: 16,
                paddingTop: 16,
                borderTop: '1px solid #f3f4f6'
            }}>
                <span style={{ color: '#6b7280', fontSize: 14 }}>
                    {t('tenant.card.members', { count: tenant.member_count || 0 })}
                </span>

                {!isMember && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation()
                            handleJoinTenant(tenant.id, tenant.access_type)
                        }}
                        disabled={joiningTenant === tenant.id}
                        style={{
                            padding: '8px 20px',
                            backgroundColor: joiningTenant === tenant.id ? '#9ca3af' : '#3b82f6',
                            color: 'white',
                            border: 'none',
                            borderRadius: 6,
                            cursor: joiningTenant === tenant.id ? 'not-allowed' : 'pointer',
                            fontSize: 14,
                            fontWeight: 500,
                            transition: 'background-color 0.2s'
                        }}
                    >
                        {joiningTenant === tenant.id ? '...' : t('tenant.join.button')}
                    </button>
                )}
            </div>
        </div>
    )

    return (
        <AuthGuard>
            <div style={{
                minHeight: '100vh',
                backgroundColor: '#f9fafb',
                fontFamily: 'system-ui, -apple-system, sans-serif'
            }}>
                <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px 24px 100px' }}>
                    {/* Header */}
                    <div style={{
                        display: 'flex',
                        justifyContent: 'space-between',
                        alignItems: 'center',
                        marginBottom: 32
                    }}>
                        <div>
                            <button
                                onClick={() => router.push(`/${params.locale}/dashboard`)}
                                style={{
                                    background: 'none',
                                    border: 'none',
                                    color: '#6b7280',
                                    cursor: 'pointer',
                                    fontSize: 14,
                                    marginBottom: 8,
                                    padding: 0,
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 4
                                }}
                            >
                                ← {t('chat.header.back')}
                            </button>
                            <h1 style={{ margin: 0, fontSize: 28, fontWeight: 700, color: '#1f2937' }}>
                                {t('tenant.title')}
                            </h1>
                            <p style={{ margin: '8px 0 0', color: '#6b7280', fontSize: 16 }}>
                                {t('tenant.discover.subtitle')}
                            </p>
                        </div>

                        <button
                            onClick={() => setShowInviteModal(true)}
                            style={{
                                padding: '12px 24px',
                                backgroundColor: '#8b5cf6',
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
                            🎟️ {t('tenant.join.useInvite')}
                        </button>
                    </div>

                    {/* Tabs */}
                    <div style={{
                        display: 'flex',
                        gap: 8,
                        marginBottom: 24,
                        borderBottom: '2px solid #e5e7eb',
                        paddingBottom: 0
                    }}>
                        <button
                            onClick={() => setActiveTab('discover')}
                            style={{
                                padding: '12px 24px',
                                background: 'none',
                                border: 'none',
                                borderBottom: activeTab === 'discover' ? '2px solid #3b82f6' : '2px solid transparent',
                                marginBottom: -2,
                                color: activeTab === 'discover' ? '#3b82f6' : '#6b7280',
                                cursor: 'pointer',
                                fontSize: 15,
                                fontWeight: activeTab === 'discover' ? 600 : 500,
                                transition: 'all 0.2s'
                            }}
                        >
                            🔍 {t('tenant.discover.title')}
                        </button>
                        <button
                            onClick={() => setActiveTab('my-tenants')}
                            style={{
                                padding: '12px 24px',
                                background: 'none',
                                border: 'none',
                                borderBottom: activeTab === 'my-tenants' ? '2px solid #3b82f6' : '2px solid transparent',
                                marginBottom: -2,
                                color: activeTab === 'my-tenants' ? '#3b82f6' : '#6b7280',
                                cursor: 'pointer',
                                fontSize: 15,
                                fontWeight: activeTab === 'my-tenants' ? 600 : 500,
                                transition: 'all 0.2s'
                            }}
                        >
                            👥 {t('tenant.myTenants.title')} ({myTenants.length})
                        </button>
                    </div>

                    {/* Search (only for discover tab) */}
                    {activeTab === 'discover' && (
                        <div style={{
                            display: 'flex',
                            gap: 16,
                            marginBottom: 24,
                            flexWrap: 'wrap'
                        }}>
                            <div style={{ flex: 1, minWidth: 250 }}>
                                <input
                                    type="text"
                                    placeholder={t('tenant.discover.searchPlaceholder')}
                                    value={search}
                                    onChange={(e) => setSearch(e.target.value)}
                                    style={{
                                        width: '100%',
                                        padding: '12px 16px',
                                        border: '1px solid #e5e7eb',
                                        borderRadius: 8,
                                        fontSize: 15,
                                        outline: 'none',
                                        boxSizing: 'border-box'
                                    }}
                                />
                            </div>
                        </div>
                    )}

                    {/* Content */}
                    {loading ? (
                        <div style={{
                            display: 'flex',
                            justifyContent: 'center',
                            alignItems: 'center',
                            padding: 80
                        }}>
                            <div style={{ color: '#6b7280', fontSize: 16 }}>Loading...</div>
                        </div>
                    ) : activeTab === 'discover' ? (
                        tenants.length === 0 ? (
                            <div style={{
                                textAlign: 'center',
                                padding: 80,
                                color: '#6b7280'
                            }}>
                                <div style={{ fontSize: 48, marginBottom: 16 }}>🔍</div>
                                <p style={{ fontSize: 16 }}>{t('tenant.discover.noResults')}</p>
                            </div>
                        ) : (
                            <>
                                <div style={{
                                    display: 'grid',
                                    gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))',
                                    gap: 20
                                }}>
                                    {tenants.map(tenant => renderTenantCard(tenant, false))}
                                </div>
                                {nextCursor && (
                                    <div style={{
                                        display: 'flex',
                                        justifyContent: 'center',
                                        marginTop: 24
                                    }}>
                                        <button
                                            onClick={loadMoreTenants}
                                            disabled={loadingMore}
                                            style={{
                                                minWidth: 200,
                                                padding: '12px 24px',
                                                backgroundColor: loadingMore ? '#9ca3af' : '#1f2937',
                                                color: 'white',
                                                border: 'none',
                                                borderRadius: 8,
                                                cursor: loadingMore ? 'not-allowed' : 'pointer',
                                                fontSize: 15,
                                                fontWeight: 600,
                                                transition: 'background-color 0.2s'
                                            }}
                                        >
                                            {loadingMore ? t('tenant.discover.loadingMore') : t('tenant.discover.loadMore')}
                                        </button>
                                    </div>
                                )}
                            </>
                        )
                    ) : (
                        myTenants.length === 0 ? (
                            <div style={{
                                textAlign: 'center',
                                padding: 80,
                                color: '#6b7280'
                            }}>
                                <div style={{ fontSize: 48, marginBottom: 16 }}>👥</div>
                                <p style={{ fontSize: 16 }}>{t('tenant.myTenants.empty')}</p>
                            </div>
                        ) : (
                            <div style={{
                                display: 'grid',
                                gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))',
                                gap: 20
                            }}>
                                {myTenants.map(tenant => renderTenantCard(tenant, true))}
                            </div>
                        )
                    )}
                </div>

                {/* Invite Code Modal */}
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
                                maxWidth: 400,
                                boxShadow: '0 25px 50px -12px rgba(0,0,0,0.25)'
                            }}
                            onClick={(e) => e.stopPropagation()}
                        >
                            <h2 style={{ margin: '0 0 8px', fontSize: 20, fontWeight: 600 }}>
                                🎟️ {t('tenant.join.useInvite')}
                            </h2>
                            <p style={{ margin: '0 0 24px', color: '#6b7280', fontSize: 14 }}>
                                {t('tenant.invites.code')}
                            </p>

                            <input
                                type="text"
                                placeholder={t('tenant.join.inviteCodePlaceholder')}
                                value={inviteCode}
                                onChange={(e) => setInviteCode(e.target.value.toUpperCase())}
                                style={{
                                    width: '100%',
                                    padding: '14px 16px',
                                    border: '1px solid #e5e7eb',
                                    borderRadius: 8,
                                    fontSize: 18,
                                    fontFamily: 'monospace',
                                    letterSpacing: 2,
                                    textAlign: 'center',
                                    outline: 'none',
                                    boxSizing: 'border-box',
                                    marginBottom: 24
                                }}
                                autoFocus
                            />

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
                                    onClick={handleUseInviteCode}
                                    disabled={!inviteCode.trim() || joiningTenant === 'invite'}
                                    style={{
                                        flex: 1,
                                        padding: '12px 24px',
                                        backgroundColor: !inviteCode.trim() ? '#9ca3af' : '#8b5cf6',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 8,
                                        cursor: !inviteCode.trim() ? 'not-allowed' : 'pointer',
                                        fontSize: 15,
                                        fontWeight: 600
                                    }}
                                >
                                    {joiningTenant === 'invite' ? '...' : t('tenant.join.button')}
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
