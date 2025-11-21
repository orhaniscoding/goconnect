'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getUser, getAccessToken } from '../../../../lib/auth'
import {
    getSystemStats,
    listUsers,
    listTenants,
    SystemStats,
    AdminUser,
    Tenant
} from '../../../../lib/api'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

export default function AdminPage() {
    const router = useRouter()
    const [currentUser, setCurrentUser] = useState<any>(null)
    const [activeTab, setActiveTab] = useState<'stats' | 'users' | 'tenants'>('stats')

    const [stats, setStats] = useState<SystemStats | null>(null)
    const [users, setUsers] = useState<AdminUser[]>([])
    const [tenants, setTenants] = useState<Tenant[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        const user = getUser()
        if (user) {
            setCurrentUser(user)
            if (!user.is_admin) {
                router.push('/en/dashboard')
                return
            }
            loadData()
        }
    }, [])

    const loadData = async () => {
        try {
            setLoading(true)
            const token = getAccessToken()
            if (!token) return

            const [statsRes, usersRes, tenantsRes] = await Promise.all([
                getSystemStats(token),
                listUsers(50, 0, token),
                listTenants(50, 0, token)
            ])

            setStats(statsRes.data)
            setUsers(usersRes.data)
            setTenants(tenantsRes.data)
        } catch (err: any) {
            console.error('Failed to load admin data:', err)
            setError(err.message || 'Failed to load data')
        } finally {
            setLoading(false)
        }
    }

    useEffect(() => {
        const user = getUser()
        if (!user?.is_admin) {
            router.push('/en/dashboard')
            return
        }
        setCurrentUser(user)
    }, [router])

    if (!currentUser?.is_admin) {
        return (
            <AuthGuard>
                <div style={{ padding: 24, textAlign: 'center' }}>
                    <h2>Access Denied</h2>
                    <p>You need administrator privileges to access this page.</p>
                </div>
            </AuthGuard>
        )
    }

    function formatDate(dateString: string) {
        return new Date(dateString).toLocaleDateString()
    }

    return (
        <AuthGuard>
            <div style={{
                display: 'flex',
                flexDirection: 'column',
                minHeight: '100vh',
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
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 14,
                                fontWeight: 500
                            }}
                        >
                            ← Back
                        </button>
                        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>
                            👑 Admin Panel
                        </h1>
                    </div>
                    <div style={{
                        padding: '6px 12px',
                        backgroundColor: '#ffc107',
                        color: '#856404',
                        borderRadius: 6,
                        fontSize: 13,
                        fontWeight: 600
                    }}>
                        ADMIN ACCESS
                    </div>
                </div>

                {/* Tabs */}
                <div style={{
                    padding: '16px 24px',
                    backgroundColor: 'white',
                    borderBottom: '1px solid #dee2e6'
                }}>
                    <div style={{ display: 'flex', gap: 8 }}>
                        <button
                            onClick={() => setActiveTab('stats')}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: activeTab === 'stats' ? '#007bff' : 'transparent',
                                color: activeTab === 'stats' ? 'white' : '#6c757d',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 14,
                                fontWeight: 500
                            }}
                        >
                            📊 Statistics
                        </button>
                        <button
                            onClick={() => setActiveTab('users')}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: activeTab === 'users' ? '#007bff' : 'transparent',
                                color: activeTab === 'users' ? 'white' : '#6c757d',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 14,
                                fontWeight: 500
                            }}
                        >
                            👥 Users
                        </button>
                        <button
                            onClick={() => setActiveTab('tenants')}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: activeTab === 'tenants' ? '#007bff' : 'transparent',
                                color: activeTab === 'tenants' ? 'white' : '#6c757d',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 14,
                                fontWeight: 500
                            }}
                        >
                            🏢 Tenants
                        </button>
                    </div>
                </div>

                {/* Main Content */}
                <div style={{ flex: 1, padding: 24, maxWidth: 1400, margin: '0 auto', width: '100%' }}>

                    {loading && (
                        <div style={{ textAlign: 'center', padding: 40 }}>
                            <div style={{ fontSize: 24 }}>⏳ Loading...</div>
                        </div>
                    )}

                    {error && (
                        <div style={{
                            padding: 16,
                            backgroundColor: '#f8d7da',
                            color: '#842029',
                            borderRadius: 8,
                            marginBottom: 24
                        }}>
                            {error}
                        </div>
                    )}

                    {/* Statistics Tab */}
                    {activeTab === 'stats' && stats && (
                        <div>
                            <div style={{
                                display: 'grid',
                                gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
                                gap: 16,
                                marginBottom: 24
                            }}>
                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #007bff'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Total Users
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#007bff' }}>
                                        {stats.total_users}
                                    </div>
                                </div>

                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #28a745'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Total Tenants
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#28a745' }}>
                                        {stats.total_tenants}
                                    </div>
                                </div>

                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #ffc107'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Total Networks
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#ffc107' }}>
                                        {stats.total_networks}
                                    </div>
                                </div>

                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #17a2b8'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Total Devices
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#17a2b8' }}>
                                        {stats.total_devices}
                                    </div>
                                </div>

                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #6f42c1'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Active Connections
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#6f42c1' }}>
                                        {stats.active_connections}
                                    </div>
                                </div>

                                <div style={{
                                    backgroundColor: 'white',
                                    borderRadius: 12,
                                    padding: 24,
                                    boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                                    borderLeft: '4px solid #dc3545'
                                }}>
                                    <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                                        Messages Today
                                    </div>
                                    <div style={{ fontSize: 32, fontWeight: 600, color: '#dc3545' }}>
                                        {stats.messages_today}
                                    </div>
                                </div>
                            </div>

                            <div style={{
                                backgroundColor: '#fff3cd',
                                padding: 16,
                                borderRadius: 8,
                                color: '#856404',
                                fontSize: 14
                            }}>
                                ⚠️ These are mock statistics. Backend integration is planned for future release.
                            </div>
                        </div>
                    )}

                    {/* Users Tab */}
                    {activeTab === 'users' && (
                        <div>
                            <div style={{
                                backgroundColor: 'white',
                                borderRadius: 12,
                                padding: 24,
                                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                            }}>
                                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                                    User Management
                                </h2>

                                <div style={{ overflowX: 'auto' }}>
                                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                                        <thead>
                                            <tr style={{ borderBottom: '2px solid #dee2e6' }}>
                                                <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                                                    Email
                                                </th>
                                                <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                                                    Role
                                                </th>
                                                <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                                                    Tenant ID
                                                </th>
                                                <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                                                    Created
                                                </th>
                                                <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                                                    Actions
                                                </th>
                                            </tr>
                                        </thead>
                                        <tbody>
                                            {users.map((user) => (
                                                <tr key={user.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                                                    <td style={{ padding: '12px', fontSize: 14, color: '#212529' }}>
                                                        {user.email}
                                                    </td>
                                                    <td style={{ padding: '12px' }}>
                                                        <span style={{
                                                            padding: '4px 8px',
                                                            backgroundColor: user.is_admin ? '#d1e7dd' : user.is_moderator ? '#cfe2ff' : '#f8f9fa',
                                                            color: user.is_admin ? '#0f5132' : user.is_moderator ? '#084298' : '#6c757d',
                                                            borderRadius: 4,
                                                            fontSize: 12,
                                                            fontWeight: 500
                                                        }}>
                                                            {user.is_admin ? '👑 Admin' : user.is_moderator ? '🛡️ Moderator' : '👤 User'}
                                                        </span>
                                                    </td>
                                                    <td style={{ padding: '12px', fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                                                        {user.tenant_id.substring(0, 12)}...
                                                    </td>
                                                    <td style={{ padding: '12px', fontSize: 13, color: '#6c757d' }}>
                                                        {formatDate(user.created_at)}
                                                    </td>
                                                    <td style={{ padding: '12px' }}>
                                                        <button
                                                            disabled
                                                            style={{
                                                                padding: '4px 8px',
                                                                backgroundColor: '#6c757d',
                                                                color: 'white',
                                                                border: 'none',
                                                                borderRadius: 4,
                                                                fontSize: 12,
                                                                cursor: 'not-allowed',
                                                                opacity: 0.6
                                                            }}
                                                        >
                                                            Edit
                                                        </button>
                                                    </td>
                                                </tr>
                                            ))}
                                        </tbody>
                                    </table>
                                </div>
                            </div>
                        </div>
                    )}

                    {/* Tenants Tab */}
                    {activeTab === 'tenants' && (
                        <div>
                            <div style={{
                                backgroundColor: 'white',
                                borderRadius: 12,
                                padding: 24,
                                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                            }}>
                                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                                    Tenant Management
                                </h2>

                                <div style={{ display: 'grid', gap: 16 }}>
                                    {tenants.map((tenant) => (
                                        <div
                                            key={tenant.id}
                                            style={{
                                                padding: 20,
                                                border: '1px solid #dee2e6',
                                                borderRadius: 8,
                                                backgroundColor: '#f8f9fa'
                                            }}
                                        >
                                            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 12 }}>
                                                <div>
                                                    <h3 style={{ margin: '0 0 4px 0', fontSize: 16, fontWeight: 600 }}>
                                                        {tenant.name}
                                                    </h3>
                                                    <div style={{ fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                                                        ID: {tenant.id}
                                                    </div>
                                                </div>
                                                <button
                                                    disabled
                                                    style={{
                                                        padding: '6px 12px',
                                                        backgroundColor: '#6c757d',
                                                        color: 'white',
                                                        border: 'none',
                                                        borderRadius: 6,
                                                        fontSize: 13,
                                                        cursor: 'not-allowed',
                                                        opacity: 0.6
                                                    }}
                                                >
                                                    Manage
                                                </button>
                                            </div>

                                            <div style={{ display: 'flex', gap: 24, fontSize: 14, color: '#6c757d' }}>
                                                <div>
                                                    <span style={{ fontWeight: 500 }}>Created:</span> {formatDate(tenant.created_at)}
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                <Footer />
            </div>
        </AuthGuard>
    )
}
