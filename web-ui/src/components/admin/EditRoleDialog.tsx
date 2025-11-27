'use client'

import { useState } from 'react'
import { updateUserRole, type UserListItem } from '@/lib/api'

interface EditRoleDialogProps {
    user: UserListItem
    onClose: () => void
}

export default function EditRoleDialog({ user, onClose }: EditRoleDialogProps) {
    const [isAdmin, setIsAdmin] = useState(user.is_admin)
    const [isModerator, setIsModerator] = useState(user.is_moderator)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()

        if (!confirm(`Are you sure you want to update roles for ${user.email}?`)) return

        setLoading(true)
        setError(null)

        try {
            await updateUserRole(user.id, {
                is_admin: isAdmin,
                is_moderator: isModerator,
            })
            onClose()
        } catch (err: any) {
            setError(err.message || 'Failed to update user role')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b">
                    <h2 className="text-xl font-bold">Edit User Role</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-400 hover:text-gray-600"
                    >
                        <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                        </svg>
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit}>
                    <div className="p-6 space-y-4">
                        {/* User Info */}
                        <div className="bg-gray-50 rounded-lg p-4">
                            <div className="text-sm font-medium text-gray-700">Email</div>
                            <div className="text-base text-gray-900">{user.email}</div>
                            {user.username && (
                                <>
                                    <div className="text-sm font-medium text-gray-700 mt-2">Username</div>
                                    <div className="text-base text-gray-900">@{user.username}</div>
                                </>
                            )}
                        </div>

                        {error && (
                            <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                                <p className="text-sm text-red-800">{error}</p>
                            </div>
                        )}

                        {/* Role Checkboxes */}
                        <div className="space-y-3">
                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={isAdmin}
                                    onChange={(e) => setIsAdmin(e.target.checked)}
                                    className="w-4 h-4 text-blue-600 border-gray-300 rounded focus:ring-blue-500"
                                />
                                <span className="ml-2 text-sm font-medium text-gray-700">
                                    Administrator
                                </span>
                            </label>
                            <p className="ml-6 text-xs text-gray-500">
                                Full system access, can manage all users and settings
                            </p>

                            <label className="flex items-center">
                                <input
                                    type="checkbox"
                                    checked={isModerator}
                                    onChange={(e) => setIsModerator(e.target.checked)}
                                    className="w-4 h-4 text-green-600 border-gray-300 rounded focus:ring-green-500"
                                />
                                <span className="ml-2 text-sm font-medium text-gray-700">
                                    Moderator
                                </span>
                            </label>
                            <p className="ml-6 text-xs text-gray-500">
                                Can moderate content and manage community members
                            </p>
                        </div>

                        {/* Warning */}
                        {!isAdmin && user.is_admin && (
                            <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                                <p className="text-sm text-yellow-800">
                                    ⚠️ Warning: Removing admin privileges cannot be undone without another admin.
                                </p>
                            </div>
                        )}
                    </div>

                    {/* Footer */}
                    <div className="flex items-center justify-end gap-3 p-6 border-t bg-gray-50">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 border border-gray-300 rounded-md text-gray-700 hover:bg-gray-100"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={loading}
                            className="px-4 py-2 bg-blue-600 text-white rounded-md hover:bg-blue-700 disabled:opacity-50"
                        >
                            {loading ? 'Saving...' : 'Save Changes'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
