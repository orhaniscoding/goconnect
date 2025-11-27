'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Layout } from '../../components/Layout'

interface AdminStats {
    total_users: number
    total_networks: number
    total_devices: number
    active_peers: number
    total_tenants: number
}

export default function AdminDashboardPage() {
    const [stats, setStats] = useState<AdminStats | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const router = useRouter()

    useEffect(() => {
        loadStats()
    }, [])

    const loadStats = async () => {
        setIsLoading(true)
        setError(null)

        const token = localStorage.getItem('access_token')
        if (!token) {
            router.push('/login')
            return
        }

        try {
            // TODO: Replace with actual API call to /v1/admin/stats when endpoint is ready
            // For now, using mock data
            const mockStats: AdminStats = {
                total_users: 150,
                total_networks: 25,
                total_devices: 380,
                active_peers: 245,
                total_tenants: 12,
            }

            setStats(mockStats)
        } catch (err: any) {
            console.error('Failed to load admin stats:', err)
            setError(err.message || 'Failed to load admin stats')
        } finally {
            setIsLoading(false)
        }
    }

    if (isLoading) {
        return (
            <Layout>
                <div className="flex items-center justify-center h-64">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                </div>
            </Layout>
        )
    }

    if (error) {
        return (
            <Layout>
                <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                    <h2 className="text-red-800 font-semibold mb-2">Error</h2>
                    <p className="text-red-600">{error}</p>
                    <button
                        onClick={() => loadStats()}
                        className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                    >
                        Retry
                    </button>
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="mb-6">
                    <h1 className="text-3xl font-bold text-gray-900">Admin Dashboard</h1>
                    <p className="text-gray-600 mt-1">System overview and management</p>
                </div>

                {/* Stats Grid */}
                {stats && (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-6 mb-8">
                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="text-4xl mb-2">👥</div>
                            <p className="text-sm text-gray-600 mb-1">Total Users</p>
                            <p className="text-3xl font-bold text-gray-900">{stats.total_users}</p>
                        </div>

                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="text-4xl mb-2">🌐</div>
                            <p className="text-sm text-gray-600 mb-1">Total Networks</p>
                            <p className="text-3xl font-bold text-gray-900">
                                {stats.total_networks}
                            </p>
                        </div>

                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="text-4xl mb-2">📱</div>
                            <p className="text-sm text-gray-600 mb-1">Total Devices</p>
                            <p className="text-3xl font-bold text-gray-900">
                                {stats.total_devices}
                            </p>
                        </div>

                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="text-4xl mb-2">🟢</div>
                            <p className="text-sm text-gray-600 mb-1">Active Peers</p>
                            <p className="text-3xl font-bold text-green-600">
                                {stats.active_peers}
                            </p>
                        </div>

                        <div className="bg-white rounded-lg shadow-md p-6">
                            <div className="text-4xl mb-2">🏢</div>
                            <p className="text-sm text-gray-600 mb-1">Total Tenants</p>
                            <p className="text-3xl font-bold text-gray-900">
                                {stats.total_tenants}
                            </p>
                        </div>
                    </div>
                )}

                {/* Quick Actions */}
                <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                    <button
                        onClick={() => router.push('/admin/users')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">👥</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            User Management
                        </h2>
                        <p className="text-gray-600">
                            View and manage all users, roles, and permissions
                        </p>
                    </button>

                    <button
                        onClick={() => router.push('/networks')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">🌐</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            Network Management
                        </h2>
                        <p className="text-gray-600">
                            View and manage all VPN networks across tenants
                        </p>
                    </button>

                    <button
                        onClick={() => router.push('/devices')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">📱</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            Device Management
                        </h2>
                        <p className="text-gray-600">
                            View and manage all registered devices
                        </p>
                    </button>

                    <button
                        onClick={() => router.push('/admin/audit')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">📋</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            Audit Logs
                        </h2>
                        <p className="text-gray-600">
                            View system audit logs and activity history
                        </p>
                    </button>

                    <button
                        onClick={() => router.push('/admin/tenants')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">🏢</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            Tenant Management
                        </h2>
                        <p className="text-gray-600">
                            View and manage all tenants and organizations
                        </p>
                    </button>

                    <button
                        onClick={() => router.push('/admin/settings')}
                        className="bg-white rounded-lg shadow-md p-6 text-left hover:shadow-lg transition-shadow"
                    >
                        <div className="text-3xl mb-3">⚙️</div>
                        <h2 className="text-xl font-bold text-gray-900 mb-2">
                            System Settings
                        </h2>
                        <p className="text-gray-600">
                            Configure system settings and preferences
                        </p>
                    </button>
                </div>

                {/* Notice */}
                <div className="mt-8 bg-blue-50 border border-blue-200 rounded-lg p-4">
                    <p className="text-sm text-blue-800">
                        <strong>Note:</strong> This is a placeholder admin dashboard. Backend
                        endpoints for admin statistics and detailed management pages need to be
                        implemented.
                    </p>
                </div>
            </div>
        </Layout>
    )
}
