'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { bridge } from '../../../../lib/bridge'
import { getUser, clearAuth } from '../../../../lib/auth'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

export default function Dashboard() {
    const router = useRouter()
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
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #dee2e6',
                            borderRadius: 8,
                            opacity: 0.6
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#666' }}>💬 Chat</h3>
                        <p style={{ margin: 0, color: '#999', fontSize: 14 }}>
                            Coming soon
                        </p>
                    </div>

                    <div
                        style={{
                            padding: 20,
                            backgroundColor: '#fff',
                            border: '2px solid #dee2e6',
                            borderRadius: 8,
                            opacity: 0.6
                        }}
                    >
                        <h3 style={{ margin: '0 0 8px 0', color: '#666' }}>⚙️ Settings</h3>
                        <p style={{ margin: 0, color: '#999', fontSize: 14 }}>
                            Coming soon
                        </p>
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
