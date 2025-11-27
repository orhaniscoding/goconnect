import { useEffect, useState } from 'react'

export interface PeerStats {
    peer_id: string
    status: 'connected' | 'disconnected'
    rx_bytes: number
    tx_bytes: number
    last_handshake?: string
    endpoint?: string
}

interface PeerStatsWidgetProps {
    peerId: string
    deviceName?: string
}

export function PeerStatsWidget({ peerId, deviceName }: PeerStatsWidgetProps) {
    const [stats, setStats] = useState<PeerStats | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    useEffect(() => {
        // Initial load
        loadStats()

        // Poll every 30 seconds
        const interval = setInterval(loadStats, 30000)

        return () => clearInterval(interval)
    }, [peerId])

    const loadStats = async () => {
        setError(null)

        const token = localStorage.getItem('access_token')
        if (!token) {
            setError('Not authenticated')
            setIsLoading(false)
            return
        }

        try {
            // TODO: Implement actual API call when /v1/peers/:id/stats is ready
            // For now, using mock data
            const mockStats: PeerStats = {
                peer_id: peerId,
                status: Math.random() > 0.5 ? 'connected' : 'disconnected',
                rx_bytes: Math.floor(Math.random() * 1000000000),
                tx_bytes: Math.floor(Math.random() * 1000000000),
                last_handshake: new Date(
                    Date.now() - Math.random() * 300000
                ).toISOString(),
                endpoint: '192.168.1.100:51820',
            }

            setStats(mockStats)
        } catch (err: any) {
            console.error('Failed to load peer stats:', err)
            setError(err.message || 'Failed to load stats')
        } finally {
            setIsLoading(false)
        }
    }

    const formatBytes = (bytes: number): string => {
        if (bytes === 0) return '0 B'
        const k = 1024
        const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
        const i = Math.floor(Math.log(bytes) / Math.log(k))
        return Math.round(bytes / Math.pow(k, i) * 100) / 100 + ' ' + sizes[i]
    }

    const formatTimeSince = (timestamp: string): string => {
        const seconds = Math.floor(
            (Date.now() - new Date(timestamp).getTime()) / 1000
        )
        if (seconds < 60) return `${seconds}s ago`
        if (seconds < 3600) return `${Math.floor(seconds / 60)}m ago`
        if (seconds < 86400) return `${Math.floor(seconds / 3600)}h ago`
        return `${Math.floor(seconds / 86400)}d ago`
    }

    if (isLoading) {
        return (
            <div className="bg-white rounded-lg shadow-md p-4">
                <div className="animate-pulse">
                    <div className="h-4 bg-gray-200 rounded w-3/4 mb-2"></div>
                    <div className="h-4 bg-gray-200 rounded w-1/2"></div>
                </div>
            </div>
        )
    }

    if (error) {
        return (
            <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                <p className="text-red-800 text-sm">{error}</p>
            </div>
        )
    }

    if (!stats) {
        return (
            <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
                <p className="text-gray-600 text-sm">No stats available</p>
            </div>
        )
    }

    return (
        <div className="bg-white rounded-lg shadow-md p-4">
            {/* Header */}
            <div className="flex items-center justify-between mb-3">
                <h3 className="text-lg font-semibold text-gray-900">
                    {deviceName || 'Peer Stats'}
                </h3>
                <span
                    className={`px-2 py-1 rounded-full text-xs font-medium ${stats.status === 'connected'
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                >
                    {stats.status === 'connected' ? '🟢 Connected' : '⚪ Disconnected'}
                </span>
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-2 gap-3 mb-3">
                <div className="bg-blue-50 rounded-lg p-3">
                    <p className="text-xs text-blue-600 mb-1">Received</p>
                    <p className="text-lg font-bold text-blue-900">
                        {formatBytes(stats.rx_bytes)}
                    </p>
                </div>
                <div className="bg-purple-50 rounded-lg p-3">
                    <p className="text-xs text-purple-600 mb-1">Sent</p>
                    <p className="text-lg font-bold text-purple-900">
                        {formatBytes(stats.tx_bytes)}
                    </p>
                </div>
            </div>

            {/* Additional Info */}
            <div className="space-y-2 text-sm">
                {stats.last_handshake && (
                    <div className="flex items-center justify-between">
                        <span className="text-gray-600">Last Handshake:</span>
                        <span className="font-medium text-gray-900">
                            {formatTimeSince(stats.last_handshake)}
                        </span>
                    </div>
                )}
                {stats.endpoint && (
                    <div className="flex items-center justify-between">
                        <span className="text-gray-600">Endpoint:</span>
                        <span className="font-medium text-gray-900 font-mono text-xs">
                            {stats.endpoint}
                        </span>
                    </div>
                )}
            </div>

            {/* Refresh Indicator */}
            <div className="mt-3 pt-3 border-t border-gray-200 text-xs text-gray-500 text-center">
                Auto-refreshes every 30 seconds
            </div>
        </div>
    )
}
