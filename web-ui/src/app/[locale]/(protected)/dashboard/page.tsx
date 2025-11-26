'use client'
import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { getUser, getAccessToken } from '../../../../lib/auth'
import { bridge } from '../../../../lib/bridge'
import { useNotification } from '../../../../contexts/NotificationContext'
import { useDaemon } from '../../../../contexts/DaemonContext'
import { useT } from '../../../../lib/i18n-context'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'
import LoadingSkeleton from '../../../../components/LoadingSkeleton'

export default function Dashboard() {
    const router = useRouter()
    const params = useParams()
    const notification = useNotification()
    const t = useT()
    const { status, error: err, refresh, connect, disconnect } = useDaemon()
    const [user, setUser] = useState<any>(null)
    const [registering, setRegistering] = useState(false)
    const [toggling, setToggling] = useState(false)

    const isStatusLoading = !status && !err
    const connectionLabel = (() => {
        if (err) return t('dashboard.status.error')
        if (!status) return t('dashboard.connection.loading')
        if (!status.device?.registered) return t('dashboard.device.setupRequired')
        return status.paused ? t('dashboard.vpn.disconnected') : t('dashboard.vpn.connected')
    })()

    const connectionColor = (() => {
        if (!status) return '#6b7280'
        if (!status.device?.registered) return '#dc2626'
        return status.paused ? '#f59e0b' : '#10b981'
    })()

    const statsCards = [
        {
            key: 'state',
            label: t('dashboard.connection.vpnState'),
            value: connectionLabel,
            description: status?.device?.device_id ? `ID: ${status.device.device_id.substring(0, 8)}…` : undefined,
            accent: connectionColor
        },
        {
            key: 'peers',
            label: t('dashboard.connection.peers'),
            value: typeof status?.wg?.peers === 'number' ? status.wg.peers : '—',
            description: t('dashboard.connection.devices'),
            accent: '#2563eb'
        },
        {
            key: 'data',
            label: t('dashboard.connection.dataTransfer'),
            value: `${formatBytes(status?.wg?.total_rx || 0, 1)} / ${formatBytes(status?.wg?.total_tx || 0, 1)}`,
            description: `${t('dashboard.stats.rx')} / ${t('dashboard.stats.tx')}`,
            accent: '#8b5cf6'
        },
        {
            key: 'daemon',
            label: t('dashboard.connection.daemonVersion'),
            value: status?.version || t('dashboard.connection.unknown'),
            description: status?.wg?.last_handshake && new Date(status.wg.last_handshake).getFullYear() > 1970
                ? `${t('dashboard.connection.lastHandshake')}: ${new Date(status.wg.last_handshake).toLocaleTimeString()}`
                : undefined,
            accent: '#111827'
        }
    ]

    useEffect(() => {
        // Get user info
        const userData = getUser()
        setUser(userData)
    }, [])

    const handleToggleConnection = async () => {
        if (!status) return
        setToggling(true)
        try {
            if (status.paused) {
                await connect()
                notification.success(t('dashboard.vpn.connected'), t('dashboard.vpn.connectedMsg'))
            } else {
                await disconnect()
                notification.info(t('dashboard.vpn.disconnected'), t('dashboard.vpn.disconnectedMsg'))
            }
        } catch (e) {
            notification.error(t('dashboard.vpn.error'), String(e))
        } finally {
            setToggling(false)
        }
    }

    const handleRegister = async () => {
        setRegistering(true)
        try {
            const token = getAccessToken()
            if (!token) {
                notification.error(t('dashboard.register.failed'), t('dashboard.register.noToken'))
                return
            }

            await bridge('/register', {
                method: 'POST',
                body: JSON.stringify({ token })
            })

            notification.success(t('dashboard.register.success'), t('dashboard.register.successMsg'))
            await refresh()
        } catch (e) {
            notification.error(t('dashboard.register.failed'), String(e))
        } finally {
            setRegistering(false)
        }
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1400, margin: '0 auto' }}>
                <div style={{ marginBottom: 24 }}>
                    <h1 style={{ margin: 0 }}>{t('dashboard.title')}</h1>
                </div>

                {user && (
                    <div style={{
                        padding: 16,
                        backgroundColor: '#f8f9fa',
                        borderRadius: 8,
                        marginBottom: 24,
                        border: '1px solid #dee2e6'
                    }}>
                        <h3 style={{ marginTop: 0 }}>{t('dashboard.welcome')}, {user.name}!</h3>
                        <p style={{ margin: '4px 0', color: '#666' }}>
                            <strong>{t('dashboard.email')}:</strong> {user.email}
                        </p>
                        <p style={{ margin: '4px 0', color: '#666' }}>
                            <strong>{t('dashboard.role')}:</strong> {user.is_admin ? t('role.admin') : user.is_moderator ? t('role.moderator') : t('role.user')}
                        </p>
                    </div>
                )}

                {/* Stats Overview */}
                <div style={{ marginBottom: 24 }}>
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 12 }}>
                        <h3 style={{ margin: 0 }}>Dashboard stats</h3>
                        {isStatusLoading && (
                            <span style={{ color: '#6b7280', fontSize: 13 }}>{t('dashboard.connection.loading')}</span>
                        )}
                    </div>
                    {isStatusLoading ? (
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
                            {Array.from({ length: 4 }).map((_, idx) => (
                                <div key={idx} style={{ padding: 16, backgroundColor: '#fff', borderRadius: 10, border: '1px solid #e5e7eb' }}>
                                    <LoadingSkeleton width="45%" height={12} style={{ marginBottom: 12 }} />
                                    <LoadingSkeleton width="70%" height={22} style={{ marginBottom: 10 }} />
                                    <LoadingSkeleton width="50%" height={10} />
                                </div>
                            ))}
                        </div>
                    ) : (
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(220px, 1fr))', gap: 16 }}>
                            {statsCards.map(card => (
                                <div
                                    key={card.key}
                                    style={{
                                        padding: 16,
                                        backgroundColor: '#fff',
                                        borderRadius: 10,
                                        border: '1px solid #e5e7eb',
                                        boxShadow: '0 4px 10px rgba(0,0,0,0.02)'
                                    }}
                                >
                                    <div style={{ fontSize: 13, color: '#6b7280', marginBottom: 6 }}>{card.label}</div>
                                    <div style={{ fontSize: 24, fontWeight: 700, color: card.accent, marginBottom: card.description ? 6 : 0 }}>
                                        {card.value}
                                    </div>
                                    {card.description && (
                                        <div style={{ fontSize: 12, color: '#6b7280' }}>{card.description}</div>
                                    )}
                                </div>
                            ))}
                        </div>
                    )}
                </div>

                {/* Device Status Card */}
                <div style={{
                    padding: 16,
                    backgroundColor: status?.device?.registered ? (status.paused ? '#fff3cd' : '#d4edda') : '#f8d7da',
                    borderRadius: 8,
                    marginBottom: 24,
                    border: `1px solid ${status?.device?.registered ? (status.paused ? '#ffeeba' : '#c3e6cb') : '#f5c6cb'}`,
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                }}>
                    <div>
                        <h3 style={{ marginTop: 0, color: status?.device?.registered ? (status.paused ? '#856404' : '#155724') : '#721c24' }}>
                            {status?.device?.registered
                                ? (status.paused ? t('dashboard.vpn.disconnected') : t('dashboard.vpn.connected'))
                                : t('dashboard.device.setupRequired')}
                        </h3>
                        <p style={{ margin: '4px 0', color: status?.device?.registered ? (status.paused ? '#856404' : '#155724') : '#721c24' }}>
                            {status?.device?.registered
                                ? (status.paused
                                    ? t('dashboard.vpn.disconnectedMsg')
                                    : t('dashboard.vpn.connectedMsg', { id: status.device.device_id.substring(0, 8) }))
                                : t('dashboard.device.notRegistered')}
                        </p>
                        {status?.wg?.active && !status.paused && (
                            <div style={{ fontSize: 12, marginTop: 8, display: 'flex', gap: 16 }}>
                                <span>{t('dashboard.stats.peers')}: <strong>{status.wg.peers || 0}</strong></span>
                                <span>{t('dashboard.stats.rx')}: <strong>{Math.round((status.wg.total_rx || 0) / 1024)} KB</strong></span>
                                <span>{t('dashboard.stats.tx')}: <strong>{Math.round((status.wg.total_tx || 0) / 1024)} KB</strong></span>
                            </div>
                        )}
                        {!status && !err && <p>{t('dashboard.status.checking')}</p>}
                        {err && <p style={{ color: 'red' }}>{t('dashboard.status.error')}: {err}</p>}
                    </div>

                    {status && !status.device.registered && (
                        <button
                            onClick={handleRegister}
                            disabled={registering}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#28a745',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: registering ? 'not-allowed' : 'pointer',
                                opacity: registering ? 0.7 : 1
                            }}
                        >
                            {registering ? t('dashboard.register.button.registering') : t('dashboard.register.button.register')}
                        </button>
                    )}

                    {status && status.device.registered && (
                        <button
                            onClick={handleToggleConnection}
                            disabled={toggling}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: status.paused ? '#007bff' : '#dc3545',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: toggling ? 'not-allowed' : 'pointer',
                                opacity: toggling ? 0.7 : 1,
                                minWidth: 100
                            }}
                        >
                            {toggling ? t('dashboard.vpn.button.working') : (status.paused ? t('dashboard.vpn.button.connect') : t('dashboard.vpn.button.disconnect'))}
                        </button>
                    )}
                </div>

                {/* Quick Actions */}
                <div style={{
                    display: 'grid',
                    gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
                    gap: 16,
                    marginBottom: 24
                }}>
                    <div
                        onClick={() => router.push('/en/networks')}
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #007bff',
                            borderRadius: 8,
                            cursor: 'pointer',
                            transition: 'transform 0.2s, box-shadow 0.2s'
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.transform = 'translateY(-2px)'
                            e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,123,255,0.2)'
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.transform = 'translateY(0)'
                            e.currentTarget.style.boxShadow = 'none'
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#007bff' }}>🌐 {t('dashboard.actions.networks.title')}</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            {t('dashboard.actions.networks.desc')}
                        </p>
                    </div>

                    <div
                        onClick={() => router.push('/en/devices')}
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #10b981',
                            borderRadius: 8,
                            cursor: 'pointer',
                            transition: 'transform 0.2s, box-shadow 0.2s'
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.transform = 'translateY(-2px)'
                            e.currentTarget.style.boxShadow = '0 4px 12px rgba(16,185,129,0.2)'
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.transform = 'translateY(0)'
                            e.currentTarget.style.boxShadow = 'none'
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#10b981' }}>💻 {t('dashboard.actions.devices.title')}</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            {t('dashboard.actions.devices.desc')}
                        </p>
                    </div>

                    <div
                        onClick={() => router.push('/en/chat')}
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #dee2e6',
                            borderRadius: 8,
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.borderColor = '#8b5cf6'
                            e.currentTarget.style.transform = 'translateY(-2px)'
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.borderColor = '#dee2e6'
                            e.currentTarget.style.transform = 'translateY(0)'
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#8b5cf6' }}>💬 {t('dashboard.actions.chat.title')}</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            {t('dashboard.actions.chat.desc')}
                        </p>
                    </div>

                    <div
                        onClick={() => router.push(`/${params.locale}/tenants`)}
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #dee2e6',
                            borderRadius: 8,
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.borderColor = '#3b82f6'
                            e.currentTarget.style.transform = 'translateY(-2px)'
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.borderColor = '#dee2e6'
                            e.currentTarget.style.transform = 'translateY(0)'
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#3b82f6' }}>👥 {t('tenant.title')}</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            {t('tenant.discover.subtitle')}
                        </p>
                    </div>

                    <div
                        onClick={() => router.push('/en/profile')}
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #dee2e6',
                            borderRadius: 8,
                            cursor: 'pointer',
                            transition: 'all 0.2s',
                        }}
                        onMouseEnter={(e) => {
                            e.currentTarget.style.borderColor = '#ffc107'
                            e.currentTarget.style.transform = 'translateY(-2px)'
                        }}
                        onMouseLeave={(e) => {
                            e.currentTarget.style.borderColor = '#dee2e6'
                            e.currentTarget.style.transform = 'translateY(0)'
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#ffc107' }}>👤 {t('dashboard.actions.profile.title')}</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            {t('dashboard.actions.profile.desc')}
                        </p>
                    </div>

                    {user?.is_admin && (
                        <div
                            onClick={() => router.push('/en/admin')}
                            style={{
                                padding: 20,
                                backgroundColor: '#fff',
                                border: '2px solid #dee2e6',
                                borderRadius: 8,
                                cursor: 'pointer',
                                transition: 'all 0.2s',
                            }}
                            onMouseEnter={(e) => {
                                e.currentTarget.style.borderColor = '#dc3545'
                                e.currentTarget.style.transform = 'translateY(-2px)'
                            }}
                            onMouseLeave={(e) => {
                                e.currentTarget.style.borderColor = '#dee2e6'
                                e.currentTarget.style.transform = 'translateY(0)'
                            }}
                        >
                            <h3 style={{ margin: '0 0 8px 0', color: '#dc3545' }}>👑 {t('dashboard.actions.admin.title')}</h3>
                            <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                                {t('dashboard.actions.admin.desc')}
                            </p>
                        </div>
                    )}
                </div>

                {/* Connection Status */}
                <div style={{
                    padding: 20,
                    backgroundColor: '#fff',
                    borderRadius: 8,
                    border: '1px solid #dee2e6',
                    marginBottom: 24
                }}>
                    <h3 style={{ marginTop: 0, marginBottom: 16 }}>{t('dashboard.connection.title')}</h3>
                    {err ? (
                        <div style={{ color: '#dc3545', display: 'flex', alignItems: 'center', gap: 8 }}>
                            <span>🔴</span>
                            <span>{t('dashboard.connection.daemonError')} ({err})</span>
                        </div>
                    ) : status ? (
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 20 }}>
                            <div>
                                <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>{t('dashboard.connection.vpnState')}</div>
                                <div style={{ fontSize: 18, fontWeight: 600, color: status.wg?.active ? '#10b981' : '#6b7280' }}>
                                    {status.wg?.active ? `● ${t('dashboard.connection.connected')}` : `○ ${t('dashboard.connection.disconnected')}`}
                                </div>
                            </div>

                            {status.wg?.active && (
                                <>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>{t('dashboard.connection.peers')}</div>
                                        <div style={{ fontSize: 18, fontWeight: 600 }}>
                                            {status.wg?.peers || 0} {t('dashboard.connection.devices')}
                                        </div>
                                    </div>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>{t('dashboard.connection.dataTransfer')}</div>
                                        <div style={{ fontSize: 14 }}>
                                            ⬇️ {formatBytes(status.wg?.total_rx || 0)}<br />
                                            ⬆️ {formatBytes(status.wg?.total_tx || 0)}
                                        </div>
                                    </div>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>{t('dashboard.connection.lastHandshake')}</div>
                                        <div style={{ fontSize: 14 }}>
                                            {status.wg?.last_handshake && new Date(status.wg.last_handshake).getFullYear() > 1970
                                                ? new Date(status.wg.last_handshake).toLocaleTimeString()
                                                : t('dashboard.connection.never')}
                                        </div>
                                    </div>
                                </>
                            )}

                            <div>
                                <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>{t('dashboard.connection.daemonVersion')}</div>
                                <div style={{ fontSize: 14, fontFamily: 'monospace' }}>
                                    {status.version || t('dashboard.connection.unknown')}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <p style={{ color: '#666' }}>{t('dashboard.connection.loading')}</p>
                    )}
                </div>

                <Footer />
            </div>
        </AuthGuard>
    )
}

function formatBytes(bytes: number, decimals = 2) {
    if (!+bytes) return '0 B'
    const k = 1024
    const dm = decimals < 0 ? 0 : decimals
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(dm))} ${sizes[i]}`
}
