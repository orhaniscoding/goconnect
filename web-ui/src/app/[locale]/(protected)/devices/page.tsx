'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import {
    listDevices,
    registerDevice,
    deleteDevice,
    disableDevice,
    enableDevice,
    Device,
    RegisterDeviceRequest
} from '../../../../lib/api'
import { getAccessToken } from '../../../../lib/auth'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

type PlatformType = 'windows' | 'macos' | 'linux' | 'android' | 'ios'

export default function DevicesPage() {
    const router = useRouter()
    const [devices, setDevices] = useState<Device[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [showAddForm, setShowAddForm] = useState(false)
    const [filterPlatform, setFilterPlatform] = useState<string>('')

    // Form state
    const [formData, setFormData] = useState<RegisterDeviceRequest>({
        name: '',
        platform: 'linux',
        pubkey: '',
        hostname: '',
        os_version: '',
        daemon_ver: ''
    })
    const [formError, setFormError] = useState<string | null>(null)
    const [submitting, setSubmitting] = useState(false)

    useEffect(() => {
        loadDevices()
    }, [filterPlatform])

    const loadDevices = async () => {
        setLoading(true)
        setError(null)
        try {
            const token = getAccessToken()
            if (!token) {
                router.push('/en/login')
                return
            }

            const response = await listDevices(
                token,
                filterPlatform || undefined
            )
            setDevices(response.devices || [])
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load devices')
        } finally {
            setLoading(false)
        }
    }

    const handleAddDevice = async (e: React.FormEvent) => {
        e.preventDefault()
        setFormError(null)
        setSubmitting(true)

        try {
            const token = getAccessToken()
            if (!token) return

            await registerDevice(formData, token)

            // Reset form and close
            setFormData({
                name: '',
                platform: 'linux',
                pubkey: '',
                hostname: '',
                os_version: '',
                daemon_ver: ''
            })
            setShowAddForm(false)
            loadDevices()
            alert('Device registered successfully!')
        } catch (err) {
            setFormError(err instanceof Error ? err.message : 'Failed to register device')
        } finally {
            setSubmitting(false)
        }
    }

    const handleDeleteDevice = async (deviceId: string, deviceName: string) => {
        if (!confirm(`Are you sure you want to delete "${deviceName}"?`)) return

        try {
            const token = getAccessToken()
            if (!token) return

            await deleteDevice(deviceId, token)
            alert('Device deleted successfully')
            loadDevices()
        } catch (err) {
            alert('Failed to delete device: ' + (err instanceof Error ? err.message : 'Unknown error'))
        }
    }

    const handleToggleDevice = async (device: Device) => {
        try {
            const token = getAccessToken()
            if (!token) return

            if (device.disabled_at) {
                await enableDevice(device.id, token)
                alert('Device enabled successfully')
            } else {
                await disableDevice(device.id, token)
                alert('Device disabled successfully')
            }
            loadDevices()
        } catch (err) {
            alert('Failed to toggle device: ' + (err instanceof Error ? err.message : 'Unknown error'))
        }
    }

    const getPlatformIcon = (platform: string) => {
        const icons: Record<string, string> = {
            windows: '🪟',
            macos: '🍎',
            linux: '🐧',
            android: '🤖',
            ios: '📱'
        }
        return icons[platform] || '💻'
    }

    const getPlatformColor = (platform: string) => {
        const colors: Record<string, string> = {
            windows: '#0078d4',
            macos: '#000000',
            linux: '#fcc624',
            android: '#3ddc84',
            ios: '#000000'
        }
        return colors[platform] || '#666'
    }

    if (loading) {
        return (
            <AuthGuard>
                <div style={{ padding: 24, textAlign: 'center' }}>
                    <p>Loading devices...</p>
                </div>
            </AuthGuard>
        )
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1200, margin: '0 auto' }}>
                {/* Header */}
                <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <h1 style={{ margin: 0 }}>Devices</h1>
                    <button
                        onClick={() => router.push('/en/dashboard')}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#6b7280',
                            color: 'white',
                            border: 'none',
                            borderRadius: 6,
                            cursor: 'pointer'
                        }}
                    >
                        ← Back to Dashboard
                    </button>
                </div>

                {/* Filter and Add Button */}
                <div style={{ marginBottom: 24, display: 'flex', gap: 12, alignItems: 'center' }}>
                    <select
                        value={filterPlatform}
                        onChange={(e) => setFilterPlatform(e.target.value)}
                        style={{
                            padding: '8px 12px',
                            borderRadius: 6,
                            border: '1px solid #d1d5db',
                            fontSize: 14
                        }}
                    >
                        <option value="">All Platforms</option>
                        <option value="windows">Windows</option>
                        <option value="macos">macOS</option>
                        <option value="linux">Linux</option>
                        <option value="android">Android</option>
                        <option value="ios">iOS</option>
                    </select>

                    <button
                        onClick={() => setShowAddForm(!showAddForm)}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: '#10b981',
                            color: 'white',
                            border: 'none',
                            borderRadius: 6,
                            cursor: 'pointer',
                            fontWeight: 500
                        }}
                    >
                        {showAddForm ? '✕ Cancel' : '+ Add Device'}
                    </button>
                </div>

                {/* Add Device Form */}
                {showAddForm && (
                    <div style={{
                        marginBottom: 24,
                        padding: 20,
                        backgroundColor: '#f9fafb',
                        borderRadius: 8,
                        border: '1px solid #e5e7eb'
                    }}>
                        <h3 style={{ marginTop: 0, marginBottom: 16 }}>Register New Device</h3>
                        <form onSubmit={handleAddDevice}>
                            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 16 }}>
                                <div>
                                    <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                                        Device Name *
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.name}
                                        onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                                        placeholder="e.g., My Laptop"
                                        required
                                        style={{
                                            width: '100%',
                                            padding: '8px 12px',
                                            borderRadius: 6,
                                            border: '1px solid #d1d5db',
                                            fontSize: 14
                                        }}
                                    />
                                </div>

                                <div>
                                    <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                                        Platform *
                                    </label>
                                    <select
                                        value={formData.platform}
                                        onChange={(e) => setFormData({ ...formData, platform: e.target.value as PlatformType })}
                                        required
                                        style={{
                                            width: '100%',
                                            padding: '8px 12px',
                                            borderRadius: 6,
                                            border: '1px solid #d1d5db',
                                            fontSize: 14
                                        }}
                                    >
                                        <option value="windows">Windows</option>
                                        <option value="macos">macOS</option>
                                        <option value="linux">Linux</option>
                                        <option value="android">Android</option>
                                        <option value="ios">iOS</option>
                                    </select>
                                </div>

                                <div style={{ gridColumn: '1 / -1' }}>
                                    <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                                        WireGuard Public Key * (44 characters base64)
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.pubkey}
                                        onChange={(e) => setFormData({ ...formData, pubkey: e.target.value })}
                                        placeholder="e.g., cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ="
                                        required
                                        minLength={44}
                                        maxLength={44}
                                        style={{
                                            width: '100%',
                                            padding: '8px 12px',
                                            borderRadius: 6,
                                            border: '1px solid #d1d5db',
                                            fontSize: 14,
                                            fontFamily: 'monospace'
                                        }}
                                    />
                                </div>

                                <div>
                                    <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                                        Hostname (optional)
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.hostname}
                                        onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                                        placeholder="e.g., LAPTOP-123"
                                        style={{
                                            width: '100%',
                                            padding: '8px 12px',
                                            borderRadius: 6,
                                            border: '1px solid #d1d5db',
                                            fontSize: 14
                                        }}
                                    />
                                </div>

                                <div>
                                    <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                                        OS Version (optional)
                                    </label>
                                    <input
                                        type="text"
                                        value={formData.os_version}
                                        onChange={(e) => setFormData({ ...formData, os_version: e.target.value })}
                                        placeholder="e.g., Windows 11"
                                        style={{
                                            width: '100%',
                                            padding: '8px 12px',
                                            borderRadius: 6,
                                            border: '1px solid #d1d5db',
                                            fontSize: 14
                                        }}
                                    />
                                </div>
                            </div>

                            {formError && (
                                <div style={{ padding: 12, backgroundColor: '#fee2e2', color: '#991b1b', borderRadius: 6, marginBottom: 16 }}>
                                    {formError}
                                </div>
                            )}

                            <div style={{ display: 'flex', gap: 12 }}>
                                <button
                                    type="submit"
                                    disabled={submitting}
                                    style={{
                                        padding: '10px 20px',
                                        backgroundColor: submitting ? '#9ca3af' : '#10b981',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 6,
                                        cursor: submitting ? 'not-allowed' : 'pointer',
                                        fontWeight: 500
                                    }}
                                >
                                    {submitting ? 'Registering...' : 'Register Device'}
                                </button>
                                <button
                                    type="button"
                                    onClick={() => setShowAddForm(false)}
                                    style={{
                                        padding: '10px 20px',
                                        backgroundColor: '#6b7280',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 6,
                                        cursor: 'pointer'
                                    }}
                                >
                                    Cancel
                                </button>
                            </div>
                        </form>
                    </div>
                )}

                {/* Error Message */}
                {error && (
                    <div style={{ padding: 16, backgroundColor: '#fee2e2', color: '#991b1b', borderRadius: 6, marginBottom: 24 }}>
                        {error}
                    </div>
                )}

                {/* Devices List */}
                {devices.length === 0 ? (
                    <div style={{ textAlign: 'center', padding: 40, color: '#6b7280' }}>
                        <p style={{ fontSize: 18, marginBottom: 8 }}>No devices registered yet</p>
                        <p style={{ fontSize: 14 }}>Register your first device to get started with GoConnect VPN</p>
                    </div>
                ) : (
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))', gap: 16 }}>
                        {devices.map(device => (
                            <div
                                key={device.id}
                                style={{
                                    padding: 20,
                                    backgroundColor: 'white',
                                    border: '1px solid #e5e7eb',
                                    borderRadius: 8,
                                    position: 'relative',
                                    opacity: device.disabled_at ? 0.6 : 1
                                }}
                            >
                                {/* Platform Badge */}
                                <div style={{
                                    position: 'absolute',
                                    top: 12,
                                    right: 12,
                                    fontSize: 24
                                }}>
                                    {getPlatformIcon(device.platform)}
                                </div>

                                {/* Device Info */}
                                <div style={{ marginBottom: 16 }}>
                                    <h3 style={{ margin: 0, marginBottom: 8, fontSize: 18 }}>
                                        {device.name}
                                        {device.disabled_at && (
                                            <span style={{
                                                marginLeft: 8,
                                                fontSize: 12,
                                                padding: '2px 8px',
                                                backgroundColor: '#fee2e2',
                                                color: '#991b1b',
                                                borderRadius: 4
                                            }}>
                                                DISABLED
                                            </span>
                                        )}
                                    </h3>
                                    <div style={{ fontSize: 14, color: '#6b7280', marginBottom: 4 }}>
                                        <span style={{
                                            padding: '2px 8px',
                                            backgroundColor: getPlatformColor(device.platform) + '20',
                                            color: getPlatformColor(device.platform),
                                            borderRadius: 4,
                                            fontWeight: 500
                                        }}>
                                            {device.platform}
                                        </span>
                                        {device.active && (
                                            <span style={{ marginLeft: 8, color: '#10b981' }}>● Online</span>
                                        )}
                                    </div>
                                    {device.hostname && (
                                        <div style={{ fontSize: 13, color: '#6b7280' }}>
                                            🖥️ {device.hostname}
                                        </div>
                                    )}
                                    {device.os_version && (
                                        <div style={{ fontSize: 13, color: '#6b7280' }}>
                                            📦 {device.os_version}
                                        </div>
                                    )}
                                </div>

                                {/* Public Key */}
                                <div style={{ marginBottom: 16 }}>
                                    <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Public Key:</div>
                                    <div style={{
                                        fontSize: 11,
                                        fontFamily: 'monospace',
                                        backgroundColor: '#f3f4f6',
                                        padding: '6px 8px',
                                        borderRadius: 4,
                                        wordBreak: 'break-all'
                                    }}>
                                        {device.pubkey}
                                    </div>
                                </div>

                                {/* Last Seen */}
                                {device.last_seen && (
                                    <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 12 }}>
                                        Last seen: {new Date(device.last_seen).toLocaleString()}
                                    </div>
                                )}

                                {/* Actions */}
                                <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
                                    <button
                                        onClick={() => handleToggleDevice(device)}
                                        style={{
                                            flex: 1,
                                            padding: '8px 12px',
                                            backgroundColor: device.disabled_at ? '#10b981' : '#f59e0b',
                                            color: 'white',
                                            border: 'none',
                                            borderRadius: 6,
                                            cursor: 'pointer',
                                            fontSize: 13
                                        }}
                                    >
                                        {device.disabled_at ? '✓ Enable' : '⏸ Disable'}
                                    </button>
                                    <button
                                        onClick={() => handleDeleteDevice(device.id, device.name)}
                                        style={{
                                            flex: 1,
                                            padding: '8px 12px',
                                            backgroundColor: '#ef4444',
                                            color: 'white',
                                            border: 'none',
                                            borderRadius: 6,
                                            cursor: 'pointer',
                                            fontSize: 13
                                        }}
                                    >
                                        🗑️ Delete
                                    </button>
                                </div>
                            </div>
                        ))}
                    </div>
                )}

                <div style={{ marginTop: 40 }}>
                    <Footer />
                </div>
            </div>
        </AuthGuard>
    )
}
