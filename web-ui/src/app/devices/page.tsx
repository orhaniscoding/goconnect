'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Layout } from '../../components/Layout'
import { DeviceCard } from '../../components/DeviceCard'
import { RegisterDeviceDialog } from '../../components/RegisterDeviceDialog'
import { Device, listDevices } from '../../lib/api'

export default function DevicesPage() {
    const [devices, setDevices] = useState<Device[]>([])
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [showRegisterDialog, setShowRegisterDialog] = useState(false)
    const router = useRouter()

    useEffect(() => {
        loadDevices()
    }, [])

    const loadDevices = async () => {
        setIsLoading(true)
        setError(null)

        const token = localStorage.getItem('access_token')
        if (!token) {
            router.push('/login')
            return
        }

        try {
            const response = await listDevices(token)
            setDevices(response.devices || [])
        } catch (err: any) {
            console.error('Failed to load devices:', err)
            setError(err.message || 'Failed to load devices')
        } finally {
            setIsLoading(false)
        }
    }

    const handleDeviceRegistered = () => {
        setShowRegisterDialog(false)
        loadDevices()
    }

    if (isLoading) {
        return (
            <Layout>
                <div className="flex items-center justify-center h-64">
                    <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
                </div>
            </Layout>
        )
    }

    if (error) {
        return (
            <Layout>
                <div className="bg-red-50 border border-red-200 rounded-lg p-6">
                    <h2 className="text-red-800 font-semibold mb-2">Error</h2>
                    <p className="text-red-600">{error}</p>
                    <button
                        onClick={() => loadDevices()}
                        className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                    >
                        Retry
                    </button>
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="max-w-7xl mx-auto">
                {/* Header */}
                <div className="flex items-center justify-between mb-6">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900">My Devices</h1>
                        <p className="text-gray-600 mt-1">
                            Register and manage your WireGuard devices
                        </p>
                    </div>
                    <button
                        onClick={() => setShowRegisterDialog(true)}
                        className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 flex items-center gap-2"
                    >
                        <span>+</span>
                        <span>Register Device</span>
                    </button>
                </div>

                {/* Empty State */}
                {devices.length === 0 ? (
                    <div className="bg-white rounded-lg shadow-md p-12 text-center">
                        <div className="text-6xl mb-4">📱</div>
                        <h2 className="text-2xl font-semibold text-gray-900 mb-2">
                            No devices registered
                        </h2>
                        <p className="text-gray-600 mb-6">
                            Register your first device to connect to VPN networks
                        </p>
                        <button
                            onClick={() => setShowRegisterDialog(true)}
                            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700"
                        >
                            Register Your First Device
                        </button>
                    </div>
                ) : (
                    /* Device Grid */
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {devices.map((device) => (
                            <DeviceCard key={device.id} device={device} />
                        ))}
                    </div>
                )}

                {/* Register Device Dialog */}
                {showRegisterDialog && (
                    <RegisterDeviceDialog
                        onClose={() => setShowRegisterDialog(false)}
                        onRegistered={handleDeviceRegistered}
                    />
                )}
            </div>
        </Layout>
    )
}
