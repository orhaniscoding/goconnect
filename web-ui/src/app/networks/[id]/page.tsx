'use client'

import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { Layout } from '../../../components/Layout'
import {
    Network,
    Membership,
    JoinRequest,
    Device,
    getNetwork,
    joinNetwork,
    listMembers,
    listDevices,
} from '../../../lib/api'

export default function NetworkDetailPage() {
    const params = useParams()
    const router = useRouter()
    const networkId = params?.id as string

    const [network, setNetwork] = useState<Network | null>(null)
    const [members, setMembers] = useState<Membership[]>([])
    const [devices, setDevices] = useState<Device[]>([])
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [isJoining, setIsJoining] = useState(false)
    const [isMember, setIsMember] = useState(false)

    useEffect(() => {
        if (networkId) {
            loadNetworkDetails()
        }
    }, [networkId])

    const loadNetworkDetails = async () => {
        setIsLoading(true)
        setError(null)

        const token = localStorage.getItem('access_token')
        if (!token) {
            router.push('/login')
            return
        }

        try {
            // Load network details
            const networkRes = await getNetwork(networkId, token)
            setNetwork(networkRes.data)

            // Load members
            const membersRes = await listMembers(networkId, token)
            setMembers(membersRes.data || [])

            // Check if current user is member
            const userId = localStorage.getItem('user_id')
            const membership = membersRes.data?.find((m) => m.user_id === userId)
            setIsMember(!!membership)

            // Load devices if member
            if (membership) {
                const devicesRes = await listDevices(token)
                // Filter devices for this network (need to check peer associations)
                setDevices(devicesRes.devices || [])
            }
        } catch (err: any) {
            console.error('Failed to load network details:', err)
            setError(err.message || 'Failed to load network details')
        } finally {
            setIsLoading(false)
        }
    }

    const handleJoinNetwork = async () => {
        setIsJoining(true)
        const token = localStorage.getItem('access_token')
        if (!token) {
            router.push('/login')
            return
        }

        try {
            await joinNetwork(networkId, token)
            // Reload network details
            await loadNetworkDetails()
        } catch (err: any) {
            console.error('Failed to join network:', err)
            setError(err.message || 'Failed to join network')
        } finally {
            setIsJoining(false)
        }
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
                        onClick={() => router.push('/networks')}
                        className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
                    >
                        Back to Networks
                    </button>
                </div>
            </Layout>
        )
    }

    if (!network) {
        return (
            <Layout>
                <div className="text-center text-gray-500 py-12">
                    <p>Network not found</p>
                </div>
            </Layout>
        )
    }

    return (
        <Layout>
            <div className="max-w-7xl mx-auto">
                {/* Network Header */}
                <div className="bg-white rounded-lg shadow-md p-6 mb-6">
                    <div className="flex items-center justify-between mb-4">
                        <div>
                            <h1 className="text-3xl font-bold text-gray-900">{network.name}</h1>
                            <p className="text-gray-600 mt-1">{network.cidr}</p>
                        </div>
                        <div className="flex items-center gap-3">
                            <span
                                className={`px-3 py-1 rounded-full text-sm font-medium ${network.visibility === 'public'
                                    ? 'bg-green-100 text-green-800'
                                    : 'bg-gray-100 text-gray-800'
                                    }`}
                            >
                                {network.visibility === 'public' ? '🌐 Public' : '🔒 Private'}
                            </span>
                            <span className="px-3 py-1 bg-blue-100 text-blue-800 rounded-full text-sm font-medium">
                                {network.join_policy === 'open' && '🔓 Open'}
                                {network.join_policy === 'approval' && '✋ Approval'}
                                {network.join_policy === 'invite' && '🔐 Invite Only'}
                            </span>
                        </div>
                    </div>

                    <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-4">
                        {network.dns && (
                            <div>
                                <p className="text-sm text-gray-600">DNS</p>
                                <p className="font-medium">{network.dns}</p>
                            </div>
                        )}
                        {network.mtu && (
                            <div>
                                <p className="text-sm text-gray-600">MTU</p>
                                <p className="font-medium">{network.mtu}</p>
                            </div>
                        )}
                        <div>
                            <p className="text-sm text-gray-600">Split Tunnel</p>
                            <p className="font-medium">{network.split_tunnel ? 'Yes' : 'No'}</p>
                        </div>
                        <div>
                            <p className="text-sm text-gray-600">Created</p>
                            <p className="font-medium">
                                {new Date(network.created_at).toLocaleDateString()}
                            </p>
                        </div>
                    </div>

                    {/* Join Button (if not member) */}
                    {!isMember && (
                        <button
                            onClick={handleJoinNetwork}
                            disabled={isJoining}
                            className="w-full md:w-auto px-6 py-3 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed"
                        >
                            {isJoining ? 'Joining...' : 'Join Network'}
                        </button>
                    )}
                </div>

                {/* Members Section (if member) */}
                {isMember && (
                    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
                        <h2 className="text-2xl font-bold text-gray-900 mb-4">
                            Members ({members.length})
                        </h2>
                        <div className="space-y-3">
                            {members.length === 0 ? (
                                <p className="text-gray-500">No members yet</p>
                            ) : (
                                members.map((member) => (
                                    <div
                                        key={member.id}
                                        className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                                    >
                                        <div>
                                            <p className="font-medium">User ID: {member.user_id}</p>
                                            <p className="text-sm text-gray-600">
                                                Joined {new Date(member.joined_at).toLocaleDateString()}
                                            </p>
                                        </div>
                                        <span
                                            className={`px-3 py-1 rounded-full text-sm font-medium ${member.role === 'owner'
                                                ? 'bg-purple-100 text-purple-800'
                                                : member.role === 'admin'
                                                    ? 'bg-blue-100 text-blue-800'
                                                    : 'bg-gray-100 text-gray-800'
                                                }`}
                                        >
                                            {member.role}
                                        </span>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}

                {/* Devices Section (if member) */}
                {isMember && (
                    <div className="bg-white rounded-lg shadow-md p-6">
                        <h2 className="text-2xl font-bold text-gray-900 mb-4">
                            My Devices ({devices.length})
                        </h2>
                        <div className="space-y-3">
                            {devices.length === 0 ? (
                                <div className="text-center py-8">
                                    <p className="text-gray-500 mb-4">No devices registered yet</p>
                                    <button
                                        onClick={() => router.push('/devices')}
                                        className="px-4 py-2 bg-blue-600 text-white rounded hover:bg-blue-700"
                                    >
                                        Register Device
                                    </button>
                                </div>
                            ) : (
                                devices.map((device) => (
                                    <div
                                        key={device.id}
                                        className="flex items-center justify-between p-3 border border-gray-200 rounded-lg"
                                    >
                                        <div>
                                            <p className="font-medium">{device.name}</p>
                                            <p className="text-sm text-gray-600">{device.platform}</p>
                                        </div>
                                        <button
                                            onClick={() => router.push('/devices')}
                                            className="px-3 py-1 bg-gray-100 text-gray-800 rounded hover:bg-gray-200"
                                        >
                                            View Config
                                        </button>
                                    </div>
                                ))
                            )}
                        </div>
                    </div>
                )}
            </div>
        </Layout>
    )
}
