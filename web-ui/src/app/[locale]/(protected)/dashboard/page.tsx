'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { bridge } from '../../../../lib/bridge'
import { getUser, clearAuth } from '../../../../lib/auth'
import { useNotification } from '../../../../contexts/NotificationContext'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

export default function Dashboard() {
    const router = useRouter()
    const notification = useNotification()
    const [status, setStatus] = useState<any>(null)
    const [err, setErr] = useState<string | null>(null)
    const [user, setUser] = useState<any>(null)

    useEffect(() => {
        // Get user info
        const userData = getUser()
        setUser(userData)

        // Fetch bridge status
        bridge('/status', undefined)
            .then(setStatus)
            .catch((e) => {
                const msg = String(e)
                setErr(msg)
                // Only show notification if it's not just "Failed to fetch" (common when daemon is off)
                if (!msg.includes('Failed to fetch')) {
                    notification.error('Bridge Connection', msg)
                }
            })
    }, [])

    const handleLogout = () => {
        clearAuth()
        router.push('/en/login')
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
                    <h1 style={{ margin: 0 }}>Dashboard</h1>
                    <button
                        onClick={handleLogout}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#dc3545',
                            color: 'white',
                            border: 'none',
                            borderRadius: 4,
                            cursor: 'pointer',
                            fontSize: 14,
                            fontWeight: 500
                        }}
                    >
                        Logout
                    </button>
                </div>

                {user && (
                    <div style={{
                        padding: 16,
                        backgroundColor: '#f8f9fa',
                        borderRadius: 8,
                        marginBottom: 24,
                        border: '1px solid #dee2e6'
                    }}>
                        <h3 style={{ marginTop: 0 }}>Welcome, {user.name}!</h3>
                        <p style={{ margin: '4px 0', color: '#666' }}>
                            <strong>Email:</strong> {user.email}
                        </p>
                        <p style={{ margin: '4px 0', color: '#666' }}>
                            <strong>Role:</strong> {user.is_admin ? 'Admin' : user.is_moderator ? 'Moderator' : 'User'}
                        </p>
                    </div>
                )}

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
                        <h3 style={{ margin: '0 0 8px 0', color: '#007bff' }}>🌐 Networks</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            Manage your VPN networks
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
                        <h3 style={{ margin: '0 0 8px 0', color: '#10b981' }}>💻 Devices</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            Manage your registered devices
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
                        <h3 style={{ margin: '0 0 8px 0', color: '#8b5cf6' }}>💬 Chat</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            Real-time messaging
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
                        <h3 style={{ margin: '0 0 8px 0', color: '#ffc107' }}>👤 Profile</h3>
                        <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                            Account settings
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
                            <h3 style={{ margin: '0 0 8px 0', color: '#dc3545' }}>👑 Admin Panel</h3>
                            <p style={{ margin: 0, color: '#666', fontSize: 14 }}>
                                System administration
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
                    <h3 style={{ marginTop: 0, marginBottom: 16 }}>Connection Status</h3>
                    {err ? (
                        <div style={{ color: '#dc3545', display: 'flex', alignItems: 'center', gap: 8 }}>
                            <span>🔴</span>
                            <span>Daemon not reachable ({err})</span>
                        </div>
                    ) : status ? (
                        <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))', gap: 20 }}>
                            <div>
                                <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>VPN State</div>
                                <div style={{ fontSize: 18, fontWeight: 600, color: status.wg?.active ? '#10b981' : '#6b7280' }}>
                                    {status.wg?.active ? '● Connected' : '○ Disconnected'}
                                </div>
                            </div>

                            {status.wg?.active && (
                                <>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>Peers</div>
                                        <div style={{ fontSize: 18, fontWeight: 600 }}>
                                            {status.wg?.peers || 0} devices
                                        </div>
                                    </div>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>Data Transfer</div>
                                        <div style={{ fontSize: 14 }}>
                                            ⬇️ {formatBytes(status.wg?.total_rx || 0)}<br />
                                            ⬆️ {formatBytes(status.wg?.total_tx || 0)}
                                        </div>
                                    </div>
                                    <div>
                                        <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>Last Handshake</div>
                                        <div style={{ fontSize: 14 }}>
                                            {status.wg?.last_handshake && new Date(status.wg.last_handshake).getFullYear() > 1970
                                                ? new Date(status.wg.last_handshake).toLocaleTimeString()
                                                : 'Never'}
                                        </div>
                                    </div>
                                </>
                            )}

                            <div>
                                <div style={{ fontSize: 13, color: '#666', marginBottom: 4 }}>Daemon Version</div>
                                <div style={{ fontSize: 14, fontFamily: 'monospace' }}>
                                    {status.version || 'Unknown'}
                                </div>
                            </div>
                        </div>
                    ) : (
                        <p style={{ color: '#666' }}>Loading status...</p>
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
