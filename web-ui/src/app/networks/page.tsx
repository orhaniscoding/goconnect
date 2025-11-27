'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Layout } from '../../components/Layout'
import { NetworkCard } from '../../components/NetworkCard'
import { CreateNetworkDialog } from '../../components/CreateNetworkDialog'
import { Network, listNetworks } from '../../lib/api'

export default function NetworksPage() {
    const [networks, setNetworks] = useState<Network[]>([])
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [showCreateDialog, setShowCreateDialog] = useState(false)
    const router = useRouter()

    useEffect(() => {
        loadNetworks()
    }, [])

    const loadNetworks = async () => {
        try {
            setIsLoading(true)
            setError(null)

            const token = localStorage.getItem('accessToken')
            if (!token) {
                router.push('/login')
                return
            }

            const response = await listNetworks('mine', token)
            setNetworks(response.data || [])
        } catch (err) {
            console.error('Failed to load networks:', err)
            setError('Failed to load networks. Please try again.')
        } finally {
            setIsLoading(false)
        }
    }

    const handleNetworkCreated = (newNetwork: Network) => {
        setNetworks([newNetwork, ...networks])
        setShowCreateDialog(false)
    }

    if (isLoading) {
        return (
            <Layout>
                <div className="flex items-center justify-center min-h-[calc(100vh-4rem)]">
                    <div className="animate-spin h-12 w-12 border-4 border-blue-600 rounded-full border-t-transparent" />
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="container max-w-6xl mx-auto py-8 px-4">
                {/* Header */}
                <div className="flex items-center justify-between mb-8">
                    <div>
                        <h1 className="text-3xl font-bold text-gray-900">Networks</h1>
                        <p className="text-gray-600 mt-1">Manage your VPN networks</p>
                    </div>
                    <button
                        onClick={() => setShowCreateDialog(true)}
                        className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium flex items-center gap-2"
                    >
                        <span className="text-xl">+</span>
                        Create Network
                    </button>
                </div>

                {/* Error Message */}
                {error && (
                    <div className="bg-red-50 border border-red-200 rounded-lg p-4 mb-6">
                        <p className="text-red-800">{error}</p>
                        <button
                            onClick={loadNetworks}
                            className="mt-2 text-red-600 hover:text-red-800 font-medium"
                        >
                            Try Again
                        </button>
                    </div>
                )}

                {/* Networks Grid */}
                {networks.length === 0 ? (
                    <div className="text-center py-16 bg-white rounded-lg shadow">
                        <div className="text-6xl mb-4">🌐</div>
                        <h3 className="text-xl font-semibold text-gray-900 mb-2">
                            No networks yet
                        </h3>
                        <p className="text-gray-600 mb-6">
                            Create your first VPN network to get started
                        </p>
                        <button
                            onClick={() => setShowCreateDialog(true)}
                            className="px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 font-medium"
                        >
                            Create Network
                        </button>
                    </div>
                ) : (
                    <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
                        {networks.map((network) => (
                            <NetworkCard
                                key={network.id}
                                network={network}
                                onClick={() => router.push(`/networks/${network.id}`)}
                            />
                        ))}
                    </div>
                )}

                {/* Create Network Dialog */}
                {showCreateDialog && (
                    <CreateNetworkDialog
                        onClose={() => setShowCreateDialog(false)}
                        onCreated={handleNetworkCreated}
                    />
                )}
            </div>
        </Layout>
    )
}
