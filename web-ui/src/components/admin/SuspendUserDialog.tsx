'use client'

import { useState } from 'react'
import { suspendUser, type UserListItem } from '@/lib/api'

interface SuspendUserDialogProps {
    user: UserListItem
    onClose: () => void
}

export default function SuspendUserDialog({ user, onClose }: SuspendUserDialogProps) {
    const [reason, setReason] = useState('')
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()

        if (reason.length < 10) {
            setError('Reason must be at least 10 characters')
            return
        }

        if (reason.length > 500) {
            setError('Reason cannot exceed 500 characters')
            return
        }

        if (!confirm(`Are you sure you want to suspend ${user.email}?`)) return

        setLoading(true)
        setError(null)

        try {
            await suspendUser(user.id, { reason })
            onClose()
        } catch (err: any) {
            setError(err.message || 'Failed to suspend user')
        } finally {
            setLoading(false)
        }
    }

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
            <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
                {/* Header */}
                <div className="flex items-center justify-between p-6 border-b">
                    <h2 className="text-xl font-bold text-red-600">Suspend User</h2>
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
                        <div className="bg-red-50 rounded-lg p-4 border border-red-200">
                            <p className="text-sm text-red-800 font-medium mb-2">
                                ⚠️ You are about to suspend this user:
                            </p>
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

                        {/* Reason Input */}
                        <div>
                            <label htmlFor="reason" className="block text-sm font-medium text-gray-700 mb-2">
                                Suspension Reason <span className="text-red-500">*</span>
                            </label>
                            <textarea
                                id="reason"
                                value={reason}
                                onChange={(e) => setReason(e.target.value)}
                                placeholder="Provide a clear reason for suspension (min 10 characters)..."
                                rows={4}
                                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:ring-red-500 focus:border-red-500"
                                required
                                minLength={10}
                                maxLength={500}
                            />
                            <div className="mt-1 text-xs text-gray-500">
                                {reason.length} / 500 characters (minimum 10)
                            </div>
                        </div>

                        {/* Warning */}
                        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3">
                            <p className="text-sm text-yellow-800">
                                <strong>Note:</strong> Suspended users will be immediately logged out and unable to access the system until unsuspended by an administrator.
                            </p>
                        </div>
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
                            disabled={loading || reason.length < 10}
                            className="px-4 py-2 bg-red-600 text-white rounded-md hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {loading ? 'Suspending...' : 'Suspend User'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
