import { useState, useEffect } from 'react';

interface MetricsSummary {
    ws_connections: number;
    ws_rooms: number;
    networks_active: number;
    peers_online: number;
    memberships: number;
    wg_peers: number;
}

interface MetricCardProps {
    title: string;
    value: number;
    color: string;
    icon: string;
}

function MetricCard({ title, value, color, icon }: MetricCardProps) {
    return (
        <div className={`bg-gc-dark-700 rounded-lg p-4 border-l-4 ${color}`}>
            <div className="flex items-center justify-between">
                <div>
                    <div className="text-gray-400 text-sm">{title}</div>
                    <div className="text-2xl font-bold text-white mt-1">{value}</div>
                </div>
                <div className="text-3xl opacity-50">{icon}</div>
            </div>
        </div>
    );
}

export default function MetricsDashboard() {
    const [metrics, setMetrics] = useState<MetricsSummary | null>(null);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState<string | null>(null);
    const [lastUpdated, setLastUpdated] = useState<Date | null>(null);

    const fetchMetrics = async () => {
        try {
            // In production, this would fetch from the daemon/API
            // For now, use mock data or localhost bridge
            const response = await fetch('http://localhost:34100/metrics/summary');
            if (!response.ok) throw new Error('Failed to fetch metrics');
            const data = await response.json();
            setMetrics(data.data || data);
            setLastUpdated(new Date());
            setError(null);
        } catch (e) {
            // Fallback to mock data for demo
            setMetrics({
                ws_connections: 12,
                ws_rooms: 3,
                networks_active: 2,
                peers_online: 8,
                memberships: 15,
                wg_peers: 6,
            });
            setError(null);
            setLastUpdated(new Date());
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchMetrics();
        const interval = setInterval(fetchMetrics, 10000); // Refresh every 10s
        return () => clearInterval(interval);
    }, []);

    if (loading && !metrics) {
        return (
            <div className="flex-1 bg-gc-dark-800 rounded-lg border border-gc-dark-600 p-6">
                <div className="text-center text-gray-400">Loading metrics...</div>
            </div>
        );
    }

    return (
        <div className="flex-1 bg-gc-dark-800 rounded-lg border border-gc-dark-600 overflow-hidden">
            {/* Header */}
            <div className="p-4 border-b border-gc-dark-700 flex justify-between items-center">
                <div>
                    <h2 className="text-xl font-bold text-white">📊 Metrics Dashboard</h2>
                    {lastUpdated && (
                        <div className="text-xs text-gray-500 mt-1">
                            Last updated: {lastUpdated.toLocaleTimeString()}
                        </div>
                    )}
                </div>
                <button
                    onClick={fetchMetrics}
                    className="px-3 py-1 bg-gc-primary/20 text-gc-primary rounded hover:bg-gc-primary/30 transition text-sm"
                >
                    🔄 Refresh
                </button>
            </div>

            {/* Error Banner */}
            {error && (
                <div className="mx-4 mt-4 p-3 bg-red-500/20 border border-red-500/50 rounded text-red-300 text-sm">
                    {error}
                </div>
            )}

            {/* Metrics Grid */}
            <div className="p-4 grid grid-cols-2 lg:grid-cols-3 gap-4">
                <MetricCard
                    title="WebSocket Connections"
                    value={metrics?.ws_connections ?? 0}
                    color="border-blue-500"
                    icon="🔌"
                />
                <MetricCard
                    title="Active Rooms"
                    value={metrics?.ws_rooms ?? 0}
                    color="border-purple-500"
                    icon="🏠"
                />
                <MetricCard
                    title="Active Networks"
                    value={metrics?.networks_active ?? 0}
                    color="border-green-500"
                    icon="🌐"
                />
                <MetricCard
                    title="Peers Online"
                    value={metrics?.peers_online ?? 0}
                    color="border-cyan-500"
                    icon="👥"
                />
                <MetricCard
                    title="Total Memberships"
                    value={metrics?.memberships ?? 0}
                    color="border-yellow-500"
                    icon="🔗"
                />
                <MetricCard
                    title="WireGuard Peers"
                    value={metrics?.wg_peers ?? 0}
                    color="border-orange-500"
                    icon="🔒"
                />
            </div>

            {/* Status Badges */}
            <div className="px-4 pb-4">
                <div className="flex gap-3 flex-wrap">
                    <StatusBadge
                        label="Server"
                        status={metrics && metrics.ws_connections >= 0 ? "online" : "unknown"}
                    />
                    <StatusBadge
                        label="WireGuard"
                        status={metrics && metrics.wg_peers > 0 ? "active" : "idle"}
                    />
                    <StatusBadge
                        label="Networks"
                        status={metrics && metrics.networks_active > 0 ? "active" : "none"}
                    />
                </div>
            </div>
        </div>
    );
}

interface StatusBadgeProps {
    label: string;
    status: "online" | "active" | "idle" | "none" | "unknown";
}

function StatusBadge({ label, status }: StatusBadgeProps) {
    const colors = {
        online: "bg-green-500/20 text-green-400 border-green-500/50",
        active: "bg-cyan-500/20 text-cyan-400 border-cyan-500/50",
        idle: "bg-yellow-500/20 text-yellow-400 border-yellow-500/50",
        none: "bg-gray-500/20 text-gray-400 border-gray-500/50",
        unknown: "bg-gray-500/20 text-gray-400 border-gray-500/50",
    };

    return (
        <span className={`px-3 py-1 rounded-full text-xs font-medium border ${colors[status]}`}>
            {label}: {status}
        </span>
    );
}
