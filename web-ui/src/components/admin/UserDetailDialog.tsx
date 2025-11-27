'use client'

import { useEffect, useState } from 'react'
import { getUserDetails, unsuspendUser, type User } from '@/lib/api'

interface UserDetailDialogProps {
    userId: string
    onClose: () => void
}

export default function UserDetailDialog({ userId, onClose }: UserDetailDialogProps) {
    const [user, setUser] = useState<User | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [unsuspending, setUnsuspending] = useState(false)

    useEffect(() => {
        loadUser()
    }, [userId])

    async function loadUser() {
        setLoading(true)
        setError(null)

        try {
            const res = await getUserDetails(userId)
            setUser(res.data)
        } catch (err: any) {
            setError(err.message || 'Failed to load user details')
        } finally {
            setLoading(false)
        }
    }

    async function handleUnsuspend() {
        if (!user || !confirm('Are you sure you want to unsuspend this user?')) return

        setUnsuspending(true)
        try {
            await unsuspendUser(userId)
            onClose()
        } catch (err: any) {
            alert(err.message || 'Failed to unsuspend user')
        } finally {
            setUnsuspending(false)
        }
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

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full mx-4 max-h-[90vh] overflow-y-auto">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b">
                    <h2 className="text-2xl font-bold">User Details</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-600"
                    >
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* Content */}
                <div className="p-6">
                    {loading && <div className="text-center py-8">Loading...</div>}

                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                            <p className="text-red-800">{error}</p>
                        </div>
                    )}

                    {user && (
                        <div className="space-y-6">
                            {/* Profile Section */}
                            <div>
                                <h3 className="text-lg font-semibold mb-4">Profile</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Email</label>
                                        <div className="mt-1 text-sm text-gray-900">{user.email}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Username</label>
                                        <div className="mt-1 text-sm text-gray-900">{user.username || 'Not set'}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Full Name</label>
                                        <div className="mt-1 text-sm text-gray-900">{user.full_name || 'Not set'}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">User ID</label>
                                        <div className="mt-1 text-sm text-gray-900 font-mono">{user.id}</div>
                                    </div>
                                </div>

                                {user.bio && (
                                    <div className="mt-4">
                                        <label className="block text-sm font-medium text-gray-700">Bio</label>
                                        <div className="mt-1 text-sm text-gray-900">{user.bio}</div>
                                    </div>
                                )}
                            </div>

                            {/* Roles Section */}
                            <div>
                                <h3 className="text-lg font-semibold mb-4">Roles & Permissions</h3>
                                <div className="flex gap-2">
                                    {user.is_admin && (
                                        <span className="px-3 py-1 text-sm font-semibold rounded-full bg-blue-100 text-blue-800">
                                            Admin
                                        </span>
                                    )}
                                    {user.is_moderator && (
                                        <span className="px-3 py-1 text-sm font-semibold rounded-full bg-green-100 text-green-800">
                                            Moderator
                                        </span>
                                    )}
                                    {!user.is_admin && !user.is_moderator && (
                                        <span className="px-3 py-1 text-sm font-semibold rounded-full bg-gray-100 text-gray-800">
                                            User
                                        </span>
                                    )}
                                </div>
                            </div>

                            {/* Status Section */}
                            <div>
                                <h3 className="text-lg font-semibold mb-4">Account Status</h3>
                                <div className="space-y-3">
                                    <div className="flex items-center justify-between">
                                        <span className="text-sm font-medium text-gray-700">Status</span>
                                        {user.suspended ? (
                                            <span className="px-3 py-1 text-sm font-semibold rounded-full bg-red-100 text-red-800">
                                                Suspended
                                            </span>
                                        ) : (
                                            <span className="px-3 py-1 text-sm font-semibold rounded-full bg-green-100 text-green-800">
                                                Active
                                            </span>
                                        )}
                                    </div>

                                    {user.suspended && (
                                        <>
                                            <div className="flex items-center justify-between">
                                                <span className="text-sm font-medium text-gray-700">Suspended At</span>
                                                <span className="text-sm text-gray-900">{formatDate(user.suspended_at)}</span>
                                            </div>
                                            <div className="flex items-center justify-between">
                                                <span className="text-sm font-medium text-gray-700">Suspended By</span>
                                                <span className="text-sm text-gray-900">{user.suspended_by || 'Unknown'}</span>
                                            </div>
                                            {user.suspended_reason && (
                                                <div>
                                                    <label className="block text-sm font-medium text-gray-700 mb-1">Suspension Reason</label>
                                                    <div className="p-3 bg-red-50 border border-red-200 rounded text-sm text-gray-900">
                                                        {user.suspended_reason}
                                                    </div>
                                                </div>
                                            )}
                                        </>
                                    )}
                                </div>
                            </div>

                            {/* Activity Section */}
                            <div>
                                <h3 className="text-lg font-semibold mb-4">Activity</h3>
                                <div className="grid grid-cols-2 gap-4">
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Created</label>
                                        <div className="mt-1 text-sm text-gray-900">{formatDate(user.created_at)}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Last Updated</label>
                                        <div className="mt-1 text-sm text-gray-900">{formatDate(user.updated_at)}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Last Seen</label>
                                        <div className="mt-1 text-sm text-gray-900">{formatDate(user.last_seen)}</div>
                                    </div>
                                    <div>
                                        <label className="block text-sm font-medium text-gray-700">Tenant ID</label>
                                        <div className="mt-1 text-sm text-gray-900 font-mono">{user.tenant_id}</div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    )}
                </div>

                {/* Footer */}
                <div className="flex items-center justify-end gap-3 p-6 border-t bg-gray-50">
                    {user?.suspended && (
                        <button
                            onClick={handleUnsuspend}
                            disabled={unsuspending}
                            className="px-4 py-2 bg-green-600 text-white rounded-md hover:bg-green-700 disabled:opacity-50"
                        >
                            {unsuspending ? 'Unsuspending...' : 'Unsuspend User'}
                        </button>
                    )}
                    <button
                        onClick={onClose}
                        className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-100"
                    >
                        Close
                    </button>
                </div>
            </div>
        </div>
    )
}
