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
            .catch((e) => setErr(String(e)))
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

                {/* Notification System Demo Section */}
                <div style={{
                    padding: 16,
                    backgroundColor: '#fff',
                    borderRadius: 8,
                    border: '1px solid #dee2e6',
                    marginBottom: 24
                }}>
                    <h3 style={{ marginTop: 0, marginBottom: 12 }}>🔔 Notification System Demo</h3>
                    <p style={{ marginBottom: 16, color: '#666', fontSize: 14 }}>
                        Test the notification system with different types of messages:
                    </p>
                    <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
                        <button
                            onClick={() => notification.success('Success!', 'Operation completed successfully')}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#28a745',
                                color: 'white',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontWeight: 500,
                                fontSize: 14
                            }}
                        >
                            ✓ Show Success
                        </button>
                        <button
                            onClick={() => notification.error('Error!', 'Something went wrong')}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#dc3545',
                                color: 'white',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontWeight: 500,
                                fontSize: 14
                            }}
                        >
                            ✕ Show Error
                        </button>
                        <button
                            onClick={() => notification.warning('Warning!', 'Please be careful with this action')}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#ffc107',
                                color: '#212529',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontWeight: 500,
                                fontSize: 14
                            }}
                        >
                            ⚠ Show Warning
                        </button>
                        <button
                            onClick={() => notification.info('Info', 'This is some useful information')}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#0d6efd',
                                color: 'white',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontWeight: 500,
                                fontSize: 14
                            }}
                        >
                            ℹ Show Info
                        </button>
                    </div>
                </div>

                <div style={{
                    padding: 16,
                    backgroundColor: '#fff',
                    borderRadius: 8,
                    border: '1px solid #dee2e6',
                    marginBottom: 24
                }}>
                    <h3 style={{ marginTop: 0 }}>Bridge Status</h3>
                    {err ? (
                        <p style={{ color: 'crimson' }}>Bridge error: {err}</p>
                    ) : status ? (
                        <pre style={{
                            backgroundColor: '#f8f9fa',
                            padding: 12,
                            borderRadius: 4,
                            overflow: 'auto',
                            fontSize: 13
                        }}>
                            {JSON.stringify(status, null, 2)}
                        </pre>
                    ) : (
                        <p style={{ color: '#666' }}>Loading...</p>
                    )}
                </div>

                <Footer />
            </div>
        </AuthGuard>
    )
}
