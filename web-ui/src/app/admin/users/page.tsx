'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getAccessToken } from '@/lib/auth'
import { listAllUsers, getUserStats, type UserListItem, type UserFilters, type UserStats } from '@/lib/api'
import UserDetailDialog from '@/components/admin/UserDetailDialog'
import EditRoleDialog from '@/components/admin/EditRoleDialog'
import SuspendUserDialog from '@/components/admin/SuspendUserDialog'

export default function UsersManagementPage() {
    const router = useRouter()
    const [users, setUsers] = useState<UserListItem[]>([])
    const [stats, setStats] = useState<UserStats | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    // Filters
    const [roleFilter, setRoleFilter] = useState<string>('')
    const [statusFilter, setStatusFilter] = useState<string>('')
    const [searchQuery, setSearchQuery] = useState('')

    // Pagination
    const [currentPage, setCurrentPage] = useState(1)
    const [totalPages, setTotalPages] = useState(1)
    const [totalCount, setTotalCount] = useState(0)
    const perPage = 50

    // Dialogs
    const [selectedUser, setSelectedUser] = useState<UserListItem | null>(null)
    const [showDetailDialog, setShowDetailDialog] = useState(false)
    const [showEditDialog, setShowEditDialog] = useState(false)
    const [showSuspendDialog, setShowSuspendDialog] = useState(false)

    useEffect(() => {
        const token = getAccessToken()
        if (!token) {
            router.push('/login')
            return
        }
        loadData()
    }, [currentPage, roleFilter, statusFilter, searchQuery])

    async function loadData() {
        setLoading(true)
        setError(null)

        try {
            const filters: UserFilters = {}
            if (roleFilter) filters.role = roleFilter as any
            if (statusFilter) filters.status = statusFilter as any
            if (searchQuery) filters.q = searchQuery

            const [usersRes, statsRes] = await Promise.all([
                listAllUsers(filters, { page: currentPage, per_page: perPage }),
                getUserStats()
            ])

            setUsers(usersRes.data || [])
            setTotalCount(usersRes.meta?.total || 0)
            setTotalPages(usersRes.meta?.pages || 1)
            setStats(statsRes.data)
        } catch (err: any) {
            setError(err.message || 'Failed to load users')
        } finally {
            setLoading(false)
        }
    }

    function handleUserClick(user: UserListItem) {
        setSelectedUser(user)
        setShowDetailDialog(true)
    }

    function handleEditRole(user: UserListItem) {
        setSelectedUser(user)
        setShowEditDialog(true)
    }

    function handleSuspend(user: UserListItem) {
        setSelectedUser(user)
        setShowSuspendDialog(true)
    }

    function handleDialogClose() {
        setShowDetailDialog(false)
        setShowEditDialog(false)
        setShowSuspendDialog(false)
        setSelectedUser(null)
        loadData() // Reload data after any change
    }

    function formatDate(dateString?: string) {
        if (!dateString) return 'Never'
        return new Date(dateString).toLocaleDateString('en-US', {
            year: 'numeric',
            month: 'short',
            day: 'numeric',
            hour: '2-digit',
            minute: '2-digit'
        })
    }

    if (loading && !users.length) {
        return (
            <div className="container mx-auto py-8">
                <div className="text-center">Loading...</div>
            </div>
        )
    }

    return (
        <div className="container mx-auto py-8">
            <h1 className="text-3xl font-bold mb-6">User Management</h1>

            {/* Stats Cards */}
            {stats && (
                <div className="grid grid-cols-1 md:grid-cols-4 gap-4 mb-8">
                    <div className="bg-white rounded-lg shadow p-6">
                        <div className="text-sm text-gray-600">Total Users</div>
                        <div className="text-2xl font-bold">{stats.total_users}</div>
                    </div>
                    <div className="bg-blue-50 rounded-lg shadow p-6">
                        <div className="text-sm text-blue-600">Admins</div>
                        <div className="text-2xl font-bold text-blue-700">{stats.admin_users}</div>
                    </div>
                    <div className="bg-green-50 rounded-lg shadow p-6">
                        <div className="text-sm text-green-600">Moderators</div>
                        <div className="text-2xl font-bold text-green-700">{stats.moderator_users}</div>
                    </div>
                    <div className="bg-red-50 rounded-lg shadow p-6">
                        <div className="text-sm text-red-600">Suspended</div>
                        <div className="text-2xl font-bold text-red-700">{stats.suspended_users}</div>
                    </div>
                </div>
            )}

            {/* Filters */}
            <div className="bg-white rounded-lg shadow p-6 mb-6">
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Role
                        </label>
                        <select
                            value={roleFilter}
                            onChange={(e) => {
                                setRoleFilter(e.target.value)
                                setCurrentPage(1)
                            }}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        >
                            <option value="">All Roles</option>
                            <option value="admin">Admin</option>
                            <option value="moderator">Moderator</option>
                            <option value="user">User</option>
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Status
                        </label>
                        <select
                            value={statusFilter}
                            onChange={(e) => {
                                setStatusFilter(e.target.value)
                                setCurrentPage(1)
                            }}
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        >
                            <option value="">All Status</option>
                            <option value="active">Active</option>
                            <option value="suspended">Suspended</option>
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Search
                        </label>
                        <input
                            type="text"
                            value={searchQuery}
                            onChange={(e) => {
                                setSearchQuery(e.target.value)
                                setCurrentPage(1)
                            }}
                            placeholder="Email or username..."
                            className="w-full px-3 py-2 border border-gray-300 rounded-md"
                        />
                    </div>
                </div>
            </div>

            {error && (
                <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                    <p className="text-red-800">{error}</p>
                </div>
            )}

            {/* Users Table */}
            <div className="bg-white rounded-lg shadow overflow-hidden">
                <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200">
                        <thead className="bg-gray-50">
                            <tr>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    User
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Role
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Status
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Created
                                </th>
                                <th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Last Seen
                                </th>
                                <th className="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase tracking-wider">
                                    Actions
                                </th>
                            </tr>
                        </thead>
                        <tbody className="bg-white divide-y divide-gray-200">
                            {users.map((user) => (
                                <tr key={user.id} className="hover:bg-gray-50">
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex items-center">
                                            <div className="flex-shrink-0 h-10 w-10 bg-blue-500 rounded-full flex items-center justify-center">
                                                <span className="text-white font-medium">
                                                    {user.email[0].toUpperCase()}
                                                </span>
                                            </div>
                                            <div className="ml-4">
                                                <div className="text-sm font-medium text-gray-900">{user.email}</div>
                                                {user.username && (
                                                    <div className="text-sm text-gray-500">@{user.username}</div>
                                                )}
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        <div className="flex gap-1">
                                            {user.is_admin && (
                                                <span className="px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800">
                                                    Admin
                                                </span>
                                            )}
                                            {user.is_moderator && (
                                                <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                                                    Moderator
                                                </span>
                                            )}
                                            {!user.is_admin && !user.is_moderator && (
                                                <span className="px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800">
                                                    User
                                                </span>
                                            )}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap">
                                        {user.suspended ? (
                                            <span className="px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800">
                                                Suspended
                                            </span>
                                        ) : (
                                            <span className="px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800">
                                                Active
                                            </span>
                                        )}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        {formatDate(user.created_at)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                                        {formatDate(user.last_seen)}
                                    </td>
                                    <td className="px-6 py-4 whitespace-nowrap text-right text-sm font-medium">
                                        <button
                                            onClick={() => handleUserClick(user)}
                                            className="text-blue-600 hover:text-blue-900 mr-3"
                                        >
                                            View
                                        </button>
                                        <button
                                            onClick={() => handleEditRole(user)}
                                            className="text-green-600 hover:text-green-900 mr-3"
                                        >
                                            Edit
                                        </button>
                                        {!user.suspended && (
                                            <button
                                                onClick={() => handleSuspend(user)}
                                                className="text-red-600 hover:text-red-900"
                                            >
                                                Suspend
                                            </button>
                                        )}
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                {totalPages > 1 && (
                    <div className="bg-gray-50 px-6 py-4 flex items-center justify-between border-t">
                        <div className="text-sm text-gray-700">
                            Showing {(currentPage - 1) * perPage + 1} to {Math.min(currentPage * perPage, totalCount)} of {totalCount} users
                        </div>
                        <div className="flex gap-2">
                            <button
                                onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                                disabled={currentPage === 1}
                                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Previous
                            </button>
                            <span className="px-4 py-2 text-sm text-gray-700">
                                Page {currentPage} of {totalPages}
                            </span>
                            <button
                                onClick={() => setCurrentPage(p => Math.min(totalPages, p + 1))}
                                disabled={currentPage === totalPages}
                                className="px-4 py-2 border border-gray-300 rounded-md text-sm font-medium text-gray-700 bg-white hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
                            >
                                Next
                            </button>
                        </div>
                    </div>
                )}
            </div>

            {/* Dialogs */}
            {selectedUser && showDetailDialog && (
                <UserDetailDialog
                    userId={selectedUser.id}
                    onClose={handleDialogClose}
                />
            )}

            {selectedUser && showEditDialog && (
                <EditRoleDialog
                    user={selectedUser}
                    onClose={handleDialogClose}
                />
            )}

            {selectedUser && showSuspendDialog && (
                <SuspendUserDialog
                    user={selectedUser}
                    onClose={handleDialogClose}
                />
            )}
        </div>
    )
}
