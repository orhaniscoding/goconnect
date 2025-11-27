import { Device } from '../lib/api'

interface DeviceCardProps {
    device: Device
}

export function DeviceCard({ device }: DeviceCardProps) {
    const platformIcons: Record<Device['platform'], string> = {
        windows: '🪟',
        macos: '🍎',
        linux: '🐧',
        android: '🤖',
        ios: '📱',
    }

    const isOnline = device.last_seen
        ? new Date().getTime() - new Date(device.last_seen).getTime() < 5 * 60 * 1000
        : false

    const handleDownloadConfig = async () => {
        // TODO: Implement config download per network
        alert('Config download will be available per network in network detail page')
    }

    return (
        <div className="bg-white rounded-lg shadow-md p-6 hover:shadow-lg transition-shadow">
            {/* Header */}
            <div className="flex items-start justify-between mb-4">
                <div className="flex items-center gap-3">
                    <div className="text-4xl">{platformIcons[device.platform]}</div>
                    <div>
                        <h3 className="text-xl font-bold text-gray-900">{device.name}</h3>
                        <p className="text-sm text-gray-600 capitalize">{device.platform}</p>
                    </div>
                </div>
                <span
                    className={`px-3 py-1 rounded-full text-sm font-medium ${isOnline
                            ? 'bg-green-100 text-green-800'
                            : 'bg-gray-100 text-gray-800'
                        }`}
                >
                    {isOnline ? '🟢 Online' : '⚪ Offline'}
                </span>
            </div>

            {/* Device Info */}
            <div className="space-y-2 mb-4">
                {device.hostname && (
                    <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-600">Hostname:</span>
                        <span className="font-medium text-gray-900">{device.hostname}</span>
                    </div>
                )}
                {device.os_version && (
                    <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-600">OS Version:</span>
                        <span className="font-medium text-gray-900">{device.os_version}</span>
                    </div>
                )}
                {device.daemon_ver && (
                    <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-600">Daemon:</span>
                        <span className="font-medium text-gray-900">{device.daemon_ver}</span>
                    </div>
                )}
                {device.last_ip && (
                    <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-600">Last IP:</span>
                        <span className="font-medium text-gray-900">{device.last_ip}</span>
                    </div>
                )}
                {device.last_seen && (
                    <div className="flex items-center gap-2 text-sm">
                        <span className="text-gray-600">Last Seen:</span>
                        <span className="font-medium text-gray-900">
                            {new Date(device.last_seen).toLocaleString()}
                        </span>
                    </div>
                )}
            </div>

            {/* Public Key (truncated) */}
            <div className="mb-4">
                <p className="text-xs text-gray-600 mb-1">Public Key</p>
                <p className="text-xs font-mono bg-gray-50 p-2 rounded truncate">
                    {device.pubkey}
                </p>
            </div>

            {/* Footer */}
            <div className="flex items-center justify-between pt-4 border-t border-gray-200">
                <div className="text-xs text-gray-500">
                    {device.active ? '✅ Active' : '❌ Disabled'}
                </div>
                <button
                    onClick={handleDownloadConfig}
                    className="px-3 py-1 bg-blue-600 text-white rounded hover:bg-blue-700 text-sm"
                >
                    Download Config
                </button>
            </div>
        </div>
    )
}
