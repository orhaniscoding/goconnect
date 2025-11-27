'use client'

import { Network } from '../lib/api'

interface NetworkCardProps {
    network: Network
    onClick: () => void
}

export function NetworkCard({ network, onClick }: NetworkCardProps) {
    const visibilityColors = {
        public: 'bg-green-100 text-green-800',
        private: 'bg-gray-100 text-gray-800',
    }

    const joinPolicyIcons = {
        open: '🔓',
        approval: '✋',
        invite: '🔐',
    }

    const joinPolicyLabels = {
        open: 'Open',
        approval: 'Approval Required',
        invite: 'Invite Only',
    }

    return (
        <div
            onClick={onClick}
            className="bg-white rounded-lg shadow hover:shadow-lg transition-shadow cursor-pointer p-6 border border-gray-200"
        >
            {/* Header */}
            <div className="flex items-start justify-between mb-4">
                <div className="flex-1">
                    <h3 className="text-xl font-bold text-gray-900 mb-1">
                        {network.name}
                    </h3>
                    <p className="text-sm text-gray-600 font-mono">{network.cidr}</p>
                </div>
                <span
                    className={`px-3 py-1 rounded-full text-xs font-medium ${visibilityColors[network.visibility]
                        }`}
                >
                    {network.visibility}
                </span>
            </div>

            {/* Stats */}
            <div className="space-y-2 mb-4">
                <div className="flex items-center gap-2 text-sm text-gray-600">
                    <span>{joinPolicyIcons[network.join_policy]}</span>
                    <span>{joinPolicyLabels[network.join_policy]}</span>
                </div>

                {network.dns && (
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                        <span>🌐</span>
                        <span className="font-mono">{network.dns}</span>
                    </div>
                )}

                {network.mtu && (
                    <div className="flex items-center gap-2 text-sm text-gray-600">
                        <span>📏</span>
                        <span>MTU: {network.mtu}</span>
                    </div>
                )}
            </div>

            {/* Footer */}
            <div className="pt-4 border-t border-gray-200 flex items-center justify-between text-sm text-gray-500">
                <span>
                    Created {new Date(network.created_at).toLocaleDateString()}
                </span>
                <button
                    onClick={(e) => {
                        e.stopPropagation()
                        onClick()
                    }}
                    className="text-blue-600 hover:text-blue-800 font-medium"
                >
                    View Details →
                </button>
            </div>
        </div>
    )
}
