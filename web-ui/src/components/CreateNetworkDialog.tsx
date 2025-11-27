'use client'

import { useState } from 'react'
import { Network, CreateNetworkRequest, createNetwork } from '../lib/api'

interface CreateNetworkDialogProps {
    onClose: () => void
    onCreated: (network: Network) => void
}

export function CreateNetworkDialog({ onClose, onCreated }: CreateNetworkDialogProps) {
    const [formData, setFormData] = useState<CreateNetworkRequest>({
        name: '',
        cidr: '10.0.0.0/24',
        visibility: 'private',
        join_policy: 'approval',
    })
    const [isLoading, setIsLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        if (!formData.name.trim()) {
            setError('Network name is required')
            return
        }

        if (!formData.cidr.match(/^\d+\.\d+\.\d+\.\d+\/\d+$/)) {
            setError('Invalid CIDR format (e.g., 10.0.0.0/24)')
            return
        }

        setIsLoading(true)
        try {
            const token = localStorage.getItem('accessToken')
            if (!token) {
                setError('Not authenticated')
                return
            }

            const response = await createNetwork(formData, token)
            onCreated(response.data)
        } catch (err: any) {
            console.error('Failed to create network:', err)
            setError(err.message || 'Failed to create network')
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-md w-full p-6">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <h2 className="text-2xl font-bold text-gray-900">Create Network</h2>
                    <button
                        onClick={onClose}
                        className="text-gray-500 hover:text-gray-700 text-2xl"
                    >
                        ×
                    </button>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="space-y-4">
                    {/* Name */}
                    <div>
                        <label htmlFor="name" className="block text-sm font-medium text-gray-700 mb-1">
                            Network Name *
                        </label>
                        <input
                            id="name"
                            type="text"
                            required
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            placeholder="My VPN Network"
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        />
                    </div>

                    {/* CIDR */}
                    <div>
                        <label htmlFor="cidr" className="block text-sm font-medium text-gray-700 mb-1">
                            CIDR Range *
                        </label>
                        <input
                            id="cidr"
                            type="text"
                            required
                            value={formData.cidr}
                            onChange={(e) => setFormData({ ...formData, cidr: e.target.value })}
                            placeholder="10.0.0.0/24"
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                        />
                        <p className="text-xs text-gray-500 mt-1">
                            IP range for this network (e.g., 10.0.0.0/24)
                        </p>
                    </div>

                    {/* Visibility */}
                    <div>
                        <label htmlFor="visibility" className="block text-sm font-medium text-gray-700 mb-1">
                            Visibility
                        </label>
                        <select
                            id="visibility"
                            value={formData.visibility}
                            onChange={(e) =>
                                setFormData({ ...formData, visibility: e.target.value as 'public' | 'private' })
                            }
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="private">Private</option>
                            <option value="public">Public</option>
                        </select>
                        <p className="text-xs text-gray-500 mt-1">
                            Public networks are discoverable by all users
                        </p>
                    </div>

                    {/* Join Policy */}
                    <div>
                        <label htmlFor="join_policy" className="block text-sm font-medium text-gray-700 mb-1">
                            Join Policy
                        </label>
                        <select
                            id="join_policy"
                            value={formData.join_policy}
                            onChange={(e) =>
                                setFormData({
                                    ...formData,
                                    join_policy: e.target.value as 'open' | 'approval' | 'invite',
                                })
                            }
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                        >
                            <option value="open">Open - Anyone can join</option>
                            <option value="approval">Approval - Admin must approve</option>
                            <option value="invite">Invite Only - Requires invite code</option>
                        </select>
                    </div>

                    {/* DNS (Optional) */}
                    <div>
                        <label htmlFor="dns" className="block text-sm font-medium text-gray-700 mb-1">
                            DNS Server (Optional)
                        </label>
                        <input
                            id="dns"
                            type="text"
                            value={formData.dns || ''}
                            onChange={(e) => setFormData({ ...formData, dns: e.target.value || undefined })}
                            placeholder="1.1.1.1"
                            className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 font-mono"
                        />
                    </div>

                    {/* Error */}
                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-3">
                            <p className="text-sm text-red-800">{error}</p>
                        </div>
                    )}

                    {/* Actions */}
                    <div className="flex gap-3 pt-4">
                        <button
                            type="button"
                            onClick={onClose}
                            className="flex-1 px-4 py-2 border border-gray-300 rounded-lg hover:bg-gray-50 font-medium"
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            disabled={isLoading}
                            className="flex-1 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                            {isLoading ? 'Creating...' : 'Create Network'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
