'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getUser } from '../../../../lib/auth'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

interface User {
    id: string
    name: string
    email: string
    is_admin: boolean
    is_moderator: boolean
    tenant_id: string
    created_at: string
}

interface Tenant {
    id: string
    name: string
    user_count: number
    network_count: number
    created_at: string
}

interface SystemStats {
    total_users: number
    total_tenants: number
    total_networks: number
    total_devices: number
    active_connections: number
    messages_today: number
}

export default function AdminPage() {
    const router = useRouter()
    const [currentUser, setCurrentUser] = useState<any>(null)
    const [activeTab, setActiveTab] = useState<'stats' | 'users' | 'tenants'>('stats')
    
    // Mock data - will be replaced with real API calls
    const [stats] = useState<SystemStats>({
        total_users: 42,
        total_tenants: 8,
        total_networks: 15,
        total_devices: 127,
        active_connections: 89,
        messages_today: 1543
    })

    const [users] = useState<User[]>([
        {
            id: '550e8400-e29b-41d4-a716-446655440000',
            name: 'Admin User',
            email: 'admin@goconnect.local',
            is_admin: true,
            is_moderator: false,
            tenant_id: 'tenant-001',
            created_at: '2025-11-01T10:00:00Z'
        },
        {
            id: '550e8400-e29b-41d4-a716-446655440001',
            name: 'John Moderator',
            email: 'mod@goconnect.local',
            is_admin: false,
            is_moderator: true,
            tenant_id: 'tenant-001',
            created_at: '2025-11-05T14:30:00Z'
        },
        {
            id: '550e8400-e29b-41d4-a716-446655440002',
            name: 'Jane User',
            email: 'user@goconnect.local',
            is_admin: false,
            is_moderator: false,
            tenant_id: 'tenant-002',
            created_at: '2025-11-10T09:15:00Z'
        }
    ])

    const [tenants] = useState<Tenant[]>([
        {
            id: 'tenant-001',
            name: 'Main Organization',
            user_count: 25,
            network_count: 8,
            created_at: '2025-10-15T08:00:00Z'
        },
        {
            id: 'tenant-002',
            name: 'Partner Company',
            user_count: 12,
            network_count: 4,
            created_at: '2025-10-20T10:30:00Z'
        },
        {
            id: 'tenant-003',
            name: 'Test Environment',
            user_count: 5,
            network_count: 3,
            created_at: '2025-11-01T14:00:00Z'
        }
    ])

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

    const formatDate = (dateString: string) => {
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric'
        })
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
                    
                    {/* Statistics Tab */}
                    {activeTab === 'stats' && (
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
                                                    Name
                                                </th>
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
                                                    <td style={{ padding: '12px', fontSize: 14 }}>
                                                        {user.name}
                                                    </td>
                                                    <td style={{ padding: '12px', fontSize: 14, color: '#6c757d' }}>
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

                            <div style={{
                                backgroundColor: '#fff3cd',
                                padding: 16,
                                borderRadius: 8,
                                color: '#856404',
                                fontSize: 14,
                                marginTop: 16
                            }}>
                                ⚠️ User management features are coming soon. Backend API integration is planned.
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
                                                    <span style={{ fontWeight: 500 }}>Users:</span> {tenant.user_count}
                                                </div>
                                                <div>
                                                    <span style={{ fontWeight: 500 }}>Networks:</span> {tenant.network_count}
                                                </div>
                                                <div>
                                                    <span style={{ fontWeight: 500 }}>Created:</span> {formatDate(tenant.created_at)}
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>

                            <div style={{
                                backgroundColor: '#fff3cd',
                                padding: 16,
                                borderRadius: 8,
                                color: '#856404',
                                fontSize: 14,
                                marginTop: 16
                            }}>
                                ⚠️ Tenant management features are coming soon. Backend API integration is planned.
                            </div>
                        </div>
                    )}
                </div>

                <Footer />
            </div>
        </AuthGuard>
    )
}
