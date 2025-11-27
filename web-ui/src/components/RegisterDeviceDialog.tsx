import { useState } from 'react'
import { registerDevice, RegisterDeviceRequest, Device } from '../lib/api'

interface RegisterDeviceDialogProps {
    onClose: () => void
    onRegistered: (device: Device) => void
}

export function RegisterDeviceDialog({
    onClose,
    onRegistered,
}: RegisterDeviceDialogProps) {
    const [formData, setFormData] = useState({
        name: '',
        platform: 'linux' as Device['platform'],
        pubkey: '',
        hostname: '',
        os_version: '',
        daemon_ver: '',
    })
    const [error, setError] = useState<string | null>(null)
    const [isSubmitting, setIsSubmitting] = useState(false)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        // Validation
        if (!formData.name.trim()) {
            setError('Device name is required')
            return
        }

        if (!formData.pubkey.trim()) {
            setError('Public key is required')
            return
        }

        // Validate WireGuard public key format (44 characters base64)
        if (formData.pubkey.length !== 44) {
            setError('Invalid WireGuard public key format (must be 44 characters)')
            return
        }

        setIsSubmitting(true)

        const token = localStorage.getItem('access_token')
        if (!token) {
            setError('Not authenticated')
            setIsSubmitting(false)
            return
        }

        try {
            const request: RegisterDeviceRequest = {
                name: formData.name.trim(),
                platform: formData.platform,
                pubkey: formData.pubkey.trim(),
            }

            // Add optional fields if provided
            if (formData.hostname.trim()) {
                request.hostname = formData.hostname.trim()
            }
            if (formData.os_version.trim()) {
                request.os_version = formData.os_version.trim()
            }
            if (formData.daemon_ver.trim()) {
                request.daemon_ver = formData.daemon_ver.trim()
            }

            const device = await registerDevice(request, token)
            onRegistered(device)
        } catch (err: any) {
            console.error('Failed to register device:', err)
            setError(err.message || 'Failed to register device')
        } finally {
            setIsSubmitting(false)
        }
    }

    return (
        <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
            <div className="bg-white rounded-lg max-w-2xl w-full max-h-[90vh] overflow-y-auto">
                {/* Header */}
                <div className="p-6 border-b border-gray-200">
                    <h2 className="text-2xl font-bold text-gray-900">Register New Device</h2>
                    <p className="text-gray-600 mt-1">
                        Register a WireGuard device to connect to VPN networks
                    </p>
                </div>

                {/* Form */}
                <form onSubmit={handleSubmit} className="p-6 space-y-4">
                    {/* Error Display */}
                    {error && (
                        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
                            <p className="text-red-800 text-sm">{error}</p>
                        </div>
                    )}

                    {/* Device Name */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Device Name *
                        </label>
                        <input
                            type="text"
                            value={formData.name}
                            onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                            placeholder="My Laptop"
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">
                            A friendly name to identify this device
                        </p>
                    </div>

                    {/* Platform */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Platform *
                        </label>
                        <select
                            value={formData.platform}
                            onChange={(e) =>
                                setFormData({
                                    ...formData,
                                    platform: e.target.value as Device['platform'],
                                })
                            }
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required
                        >
                            <option value="windows">🪟 Windows</option>
                            <option value="macos">🍎 macOS</option>
                            <option value="linux">🐧 Linux</option>
                            <option value="android">🤖 Android</option>
                            <option value="ios">📱 iOS</option>
                        </select>
                    </div>

                    {/* Public Key */}
                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            WireGuard Public Key *
                        </label>
                        <input
                            type="text"
                            value={formData.pubkey}
                            onChange={(e) => setFormData({ ...formData, pubkey: e.target.value })}
                            placeholder="base64-encoded-public-key-44-characters-long="
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg font-mono text-sm focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            required
                        />
                        <p className="text-xs text-gray-500 mt-1">
                            Generate with: <code className="bg-gray-100 px-1 py-0.5 rounded">wg genkey | wg pubkey</code>
                        </p>
                    </div>

                    {/* Optional Fields */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                Hostname (Optional)
                            </label>
                            <input
                                type="text"
                                value={formData.hostname}
                                onChange={(e) =>
                                    setFormData({ ...formData, hostname: e.target.value })
                                }
                                placeholder="my-laptop"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            />
                        </div>

                        <div>
                            <label className="block text-sm font-medium text-gray-700 mb-2">
                                OS Version (Optional)
                            </label>
                            <input
                                type="text"
                                value={formData.os_version}
                                onChange={(e) =>
                                    setFormData({ ...formData, os_version: e.target.value })
                                }
                                placeholder="Ubuntu 22.04"
                                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                            />
                        </div>
                    </div>

                    <div>
                        <label className="block text-sm font-medium text-gray-700 mb-2">
                            Daemon Version (Optional)
                        </label>
                        <input
                            type="text"
                            value={formData.daemon_ver}
                            onChange={(e) =>
                                setFormData({ ...formData, daemon_ver: e.target.value })
                            }
                            placeholder="1.0.0"
                            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                        />
                    </div>

                    {/* Actions */}
                    <div className="flex items-center justify-end gap-3 pt-4 border-t border-gray-200">
                        <button
                            type="button"
                            onClick={onClose}
                            className="px-4 py-2 text-gray-700 hover:bg-gray-100 rounded-lg"
                            disabled={isSubmitting}
                        >
                            Cancel
                        </button>
                        <button
                            type="submit"
                            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
                            disabled={isSubmitting}
                        >
                            {isSubmitting ? 'Registering...' : 'Register Device'}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
