'use client'
import { useEffect, useState } from 'react'
import { useRouter, useParams } from 'next/navigation'
import {
  getNetwork,
  joinNetwork,
  listMembers,
  approveMember,
  denyMember,
  kickMember,
  banMember,
  listJoinRequests,
  listIPAllocations,
  adminReleaseIP,
  downloadConfig,
  updateNetwork,
  Network,
  Membership,
  JoinRequest,
  IPAllocation
} from '../../../../../lib/api'
import { getAccessToken, getUser } from '../../../../../lib/auth'
import AuthGuard from '../../../../../components/AuthGuard'
import Footer from '../../../../../components/Footer'

type Tab = 'overview' | 'members' | 'requests' | 'ip-allocations' | 'settings'

export default function NetworkDetailPage() {
  const router = useRouter()
  const params = useParams()
  const networkId = params?.id as string

  const [network, setNetwork] = useState<Network | null>(null)
  const [members, setMembers] = useState<Membership[]>([])
  const [joinRequests, setJoinRequests] = useState<JoinRequest[]>([])
  const [ipAllocations, setIPAllocations] = useState<IPAllocation[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [activeTab, setActiveTab] = useState<Tab>('overview')
  const [joinStatus, setJoinStatus] = useState<'not-member' | 'member' | 'pending' | 'joining'>('not-member')
  const [currentUserId, setCurrentUserId] = useState<string>('')
  const [isAdmin, setIsAdmin] = useState(false)

  useEffect(() => {
    const user = getUser()
    if (user) {
      setCurrentUserId(user.id)
      setIsAdmin(user.is_admin)
    }
    loadNetworkData()
  }, [networkId])

  useEffect(() => {
    // Load join requests when switching to requests tab (only for admins)
    if (activeTab === 'requests' && isAdmin) {
      const token = getAccessToken()
      if (token) {
        loadJoinRequests(token)
      }
    }
    // Load IP allocations when switching to ip-allocations tab (for members)
    if (activeTab === 'ip-allocations' && joinStatus === 'member') {
      const token = getAccessToken()
      if (token) {
        loadIPAllocations(token)
      }
    }
  }, [activeTab, isAdmin, joinStatus])

  const handleDownloadConfig = async () => {
    try {
      const token = getAccessToken()
      if (!token) return

      const blob = await downloadConfig(networkId, token)
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${network?.name || 'network'}.conf`
      document.body.appendChild(a)
      a.click()
      window.URL.revokeObjectURL(url)
      document.body.removeChild(a)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to download config')
    }
  }

  const loadNetworkData = async () => {
    setLoading(true)
    setError(null)
    try {
      const token = getAccessToken()
      if (!token) {
        router.push('/en/login')
        return
      }

      // Load network details
      const networkResponse = await getNetwork(networkId, token)
      setNetwork(networkResponse.data)

      // Load members
      await loadMembers(token)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load network')
    } finally {
      setLoading(false)
    }
  }

  const loadMembers = async (token: string) => {
    try {
      const membersResponse = await listMembers(networkId, token, 'approved')
      setMembers(membersResponse.data || [])

      // Check if current user is a member
      const user = getUser()
      if (user) {
        const isMember = membersResponse.data.some((m: Membership) => m.user_id === user.id && m.status === 'approved')
        setJoinStatus(isMember ? 'member' : 'not-member')
      }
    } catch (err) {
      console.error('Failed to load members:', err)
    }
  }

  const loadJoinRequests = async (token: string) => {
    try {
      const response = await listJoinRequests(networkId, token)
      setJoinRequests(response.data || [])
    } catch (err) {
      console.error('Failed to load join requests:', err)
    }
  }

  const loadIPAllocations = async (token: string) => {
    try {
      const response = await listIPAllocations(networkId, token)
      setIPAllocations(response.data || [])
    } catch (err) {
      console.error('Failed to load IP allocations:', err)
    }
  }

  const handleReleaseIP = async (userId: string) => {
    if (!confirm(`Are you sure you want to release IP for user ${userId}?`)) {
      return
    }

    try {
      const token = getAccessToken()
      if (!token) return

      await adminReleaseIP(networkId, userId, token)
      alert('IP released successfully')
      loadIPAllocations(token)
    } catch (err) {
      alert('Failed to release IP: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleJoinNetwork = async () => {
    setJoinStatus('joining')
    try {
      const token = getAccessToken()
      if (!token) return

      const response = await joinNetwork(networkId, token)

      if (response.data) {
        // Joined immediately (open policy)
        setJoinStatus('member')
        await loadMembers(token)
        alert('Successfully joined network!')
      } else if (response.join_request) {
        // Join request created (approval required)
        setJoinStatus('pending')
        alert('Join request sent. Waiting for approval.')
      }
    } catch (err) {
      setJoinStatus('not-member')
      alert('Failed to join network: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleApprove = async (userId: string) => {
    try {
      const token = getAccessToken()
      if (!token) return

      await approveMember(networkId, userId, token)
      alert('Member approved successfully')
      loadNetworkData()
    } catch (err) {
      alert('Failed to approve member: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleDeny = async (userId: string) => {
    if (!confirm('Are you sure you want to deny this request?')) return

    try {
      const token = getAccessToken()
      if (!token) return

      await denyMember(networkId, userId, token)
      alert('Request denied')
      loadNetworkData()
    } catch (err) {
      alert('Failed to deny request: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleKick = async (userId: string) => {
    if (!confirm('Are you sure you want to kick this member?')) return

    try {
      const token = getAccessToken()
      if (!token) return

      await kickMember(networkId, userId, token)
      alert('Member kicked')
      loadNetworkData()
    } catch (err) {
      alert('Failed to kick member: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleBan = async (userId: string) => {
    if (!confirm('Are you sure you want to ban this member?')) return

    try {
      const token = getAccessToken()
      if (!token) return

      await banMember(networkId, userId, token)
      alert('Member banned')
      loadNetworkData()
    } catch (err) {
      alert('Failed to ban member: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  if (loading) {
    return (
      <AuthGuard>
        <div style={{ padding: 24, textAlign: 'center' }}>
          <p>Loading network details...</p>
        </div>
      </AuthGuard>
    )
  }

  if (error || !network) {
    return (
      <AuthGuard>
        <div style={{ padding: 24 }}>
          <p style={{ color: 'crimson' }}>{error || 'Network not found'}</p>
          <button onClick={() => router.push('/en/networks')}>Back to Networks</button>
        </div>
      </AuthGuard>
    )
  }

  const isOwnerOrAdmin = isAdmin || members.some(m => m.user_id === currentUserId && (m.role === 'owner' || m.role === 'admin'))

  return (
    <AuthGuard>
      <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1200, margin: '0 auto' }}>
        {/* Header */}
        <div style={{ marginBottom: 24 }}>
          <button
            onClick={() => router.push('/en/networks')}
            style={{
              padding: '6px 12px',
              backgroundColor: '#6c757d',
              color: 'white',
              border: 'none',
              borderRadius: 4,
              cursor: 'pointer',
              fontSize: 14,
              marginBottom: 16
            }}
          >
            ← Back to Networks
          </button>

          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div>
              <h1 style={{ margin: 0, marginBottom: 8 }}>{network.name}</h1>
              <div style={{ display: 'flex', gap: 8 }}>
                <span style={{
                  padding: '4px 12px',
                  backgroundColor: network.visibility === 'public' ? '#d4edda' : '#f8d7da',
                  color: network.visibility === 'public' ? '#155724' : '#721c24',
                  borderRadius: 4,
                  fontSize: 14
                }}>
                  {network.visibility === 'public' ? 'Public' : 'Private'}
                </span>
                <span style={{
                  padding: '4px 12px',
                  backgroundColor: '#d1ecf1',
                  color: '#0c5460',
                  borderRadius: 4,
                  fontSize: 14
                }}>
                  {network.join_policy === 'open' ? 'Open' : network.join_policy === 'approval' ? 'Approval Required' : 'Invite Only'}
                </span>
              </div>
            </div>

            {joinStatus === 'not-member' && (
              <button
                onClick={handleJoinNetwork}
                style={{
                  padding: '10px 20px',
                  backgroundColor: '#28a745',
                  color: 'white',
                  border: 'none',
                  borderRadius: 4,
                  cursor: 'pointer',
                  fontSize: 14,
                  fontWeight: 500
                }}
              >
                Join Network
              </button>
            )}

            {joinStatus === 'member' && (
              <div style={{ display: 'flex', gap: 10, alignItems: 'center' }}>
                <button
                  onClick={handleDownloadConfig}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: '#17a2b8',
                    color: 'white',
                    border: 'none',
                    borderRadius: 4,
                    cursor: 'pointer',
                    fontSize: 14,
                    fontWeight: 500
                  }}
                >
                  Download Config
                </button>
                <span style={{
                  padding: '10px 20px',
                  backgroundColor: '#d4edda',
                  color: '#155724',
                  borderRadius: 4,
                  fontSize: 14,
                  fontWeight: 500
                }}>
                  ✓ Member
                </span>
              </div>
            )}

            {joinStatus === 'pending' && (
              <span style={{
                padding: '10px 20px',
                backgroundColor: '#fff3cd',
                color: '#856404',
                borderRadius: 4,
                fontSize: 14,
                fontWeight: 500
              }}>
                ⏳ Pending Approval
              </span>
            )}
          </div>
        </div>

        {/* Tabs */}
        <div style={{ borderBottom: '2px solid #dee2e6', marginBottom: 24 }}>
          <button
            onClick={() => setActiveTab('overview')}
            style={{
              padding: '12px 24px',
              backgroundColor: 'transparent',
              color: activeTab === 'overview' ? '#007bff' : '#666',
              border: 'none',
              borderBottom: activeTab === 'overview' ? '2px solid #007bff' : '2px solid transparent',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: 500,
              marginBottom: -2
            }}
          >
            Overview
          </button>
          <button
            onClick={() => setActiveTab('members')}
            style={{
              padding: '12px 24px',
              backgroundColor: 'transparent',
              color: activeTab === 'members' ? '#007bff' : '#666',
              border: 'none',
              borderBottom: activeTab === 'members' ? '2px solid #007bff' : '2px solid transparent',
              cursor: 'pointer',
              fontSize: 14,
              fontWeight: 500,
              marginBottom: -2
            }}
          >
            Members ({members.length})
          </button>
          {isOwnerOrAdmin && (
            <button
              onClick={() => setActiveTab('requests')}
              style={{
                padding: '12px 24px',
                backgroundColor: 'transparent',
                color: activeTab === 'requests' ? '#007bff' : '#666',
                border: 'none',
                borderBottom: activeTab === 'requests' ? '2px solid #007bff' : '2px solid transparent',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500,
                marginBottom: -2
              }}
            >
              Join Requests
            </button>
          )}
          {joinStatus === 'member' && (
            <button
              onClick={() => setActiveTab('ip-allocations')}
              style={{
                padding: '12px 24px',
                backgroundColor: 'transparent',
                color: activeTab === 'ip-allocations' ? '#007bff' : '#666',
                border: 'none',
                borderBottom: activeTab === 'ip-allocations' ? '2px solid #007bff' : '2px solid transparent',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500,
                marginBottom: -2
              }}
            >
              IP Allocations ({ipAllocations.length})
            </button>
          )}
          {isOwnerOrAdmin && (
            <button
              onClick={() => setActiveTab('settings')}
              style={{
                padding: '12px 24px',
                backgroundColor: 'transparent',
                color: activeTab === 'settings' ? '#007bff' : '#666',
                border: 'none',
                borderBottom: activeTab === 'settings' ? '2px solid #007bff' : '2px solid transparent',
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500,
                marginBottom: -2
              }}
            >
              Settings
            </button>
          )}
        </div>

        {/* Tab Content */}
        {activeTab === 'overview' && (
          <OverviewTab network={network} />
        )}

        {activeTab === 'members' && (
          <MembersTab
            members={members}
            isOwnerOrAdmin={isOwnerOrAdmin}
            currentUserId={currentUserId}
            onKick={handleKick}
            onBan={handleBan}
          />
        )}

        {activeTab === 'requests' && isOwnerOrAdmin && (
          <JoinRequestsTab
            requests={joinRequests}
            onApprove={handleApprove}
            onDeny={handleDeny}
          />
        )}

        {activeTab === 'ip-allocations' && joinStatus === 'member' && (
          <IPAllocationsTab
            allocations={ipAllocations}
            network={network}
            isOwnerOrAdmin={isOwnerOrAdmin}
            onReleaseIP={handleReleaseIP}
          />
        )}

        {activeTab === 'settings' && isOwnerOrAdmin && (
          <SettingsTab network={network} onUpdate={loadNetworkData} />
        )}

        <div style={{ marginTop: 40 }}>
          <Footer />
        </div>
      </div>
    </AuthGuard>
  )
}

// Overview Tab Component
function OverviewTab({ network }: { network: Network }) {
  return (
    <div style={{ backgroundColor: '#fff', border: '1px solid #dee2e6', borderRadius: 8, padding: 24 }}>
      <h3 style={{ marginTop: 0 }}>Network Information</h3>
      <div style={{ display: 'grid', gap: 16 }}>
        <div>
          <div style={{ fontWeight: 500, color: '#666', fontSize: 14, marginBottom: 4 }}>CIDR Block</div>
          <div style={{ fontSize: 16 }}>{network.cidr}</div>
        </div>
        {network.dns && (
          <div>
            <div style={{ fontWeight: 500, color: '#666', fontSize: 14, marginBottom: 4 }}>DNS Server</div>
            <div style={{ fontSize: 16 }}>{network.dns}</div>
          </div>
        )}
        {network.mtu && (
          <div>
            <div style={{ fontWeight: 500, color: '#666', fontSize: 14, marginBottom: 4 }}>MTU</div>
            <div style={{ fontSize: 16 }}>{network.mtu}</div>
          </div>
        )}
        <div>
          <div style={{ fontWeight: 500, color: '#666', fontSize: 14, marginBottom: 4 }}>Split Tunnel</div>
          <div style={{ fontSize: 16 }}>{network.split_tunnel ? 'Enabled' : 'Disabled'}</div>
        </div>
        <div>
          <div style={{ fontWeight: 500, color: '#666', fontSize: 14, marginBottom: 4 }}>Created</div>
          <div style={{ fontSize: 16 }}>{new Date(network.created_at).toLocaleString()}</div>
        </div>
      </div>
    </div>
  )
}

// Members Tab Component
function MembersTab({
  members,
  isOwnerOrAdmin,
  currentUserId,
  onKick,
  onBan
}: {
  members: Membership[]
  isOwnerOrAdmin: boolean
  currentUserId: string
  onKick: (userId: string) => void
  onBan: (userId: string) => void
}) {
  if (members.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: 40, color: '#666' }}>
        No members yet
      </div>
    )
  }

  return (
    <div style={{ display: 'grid', gap: 12 }}>
      {members.map(member => (
        <div
          key={member.id}
          style={{
            padding: 16,
            backgroundColor: '#fff',
            border: '1px solid #dee2e6',
            borderRadius: 8,
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center'
          }}
        >
          <div>
            <div style={{ fontWeight: 500, fontSize: 16 }}>User {member.user_id.substring(0, 8)}...</div>
            <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
              <span style={{
                padding: '2px 8px',
                backgroundColor: '#e7f3ff',
                color: '#004085',
                borderRadius: 4,
                fontSize: 12,
                textTransform: 'capitalize'
              }}>
                {member.role}
              </span>
              <span style={{ fontSize: 12, color: '#666' }}>
                Joined {new Date(member.joined_at).toLocaleDateString()}
              </span>
            </div>
          </div>
          {/* Online Status Indicator */}
          {(member as any).online_device_count > 0 ? (
            <div title={`${(member as any).online_device_count} active device(s)`} style={{ width: 10, height: 10, borderRadius: '50%', backgroundColor: '#10b981', boxShadow: '0 0 0 2px #d1fae5' }}></div>
          ) : (
            <div title="Offline" style={{ width: 10, height: 10, borderRadius: '50%', backgroundColor: '#d1d5db' }}></div>
          )}
          {isOwnerOrAdmin && member.user_id !== currentUserId && member.role !== 'owner' && (
            <div style={{ display: 'flex', gap: 8, marginTop: 'auto' }}>
              <button
                onClick={() => onKick(member.user_id)}
                style={{
                  padding: '6px 12px',
                  backgroundColor: '#ffc107',
                  color: '#000',
                  border: 'none',
                  borderRadius: 4,
                  cursor: 'pointer',
                  fontSize: 12
                }}
              >
                Kick
              </button>
              <button
                onClick={() => onBan(member.user_id)}
                style={{
                  padding: '6px 12px',
                  backgroundColor: '#dc3545',
                  color: 'white',
                  border: 'none',
                  borderRadius: 4,
                  cursor: 'pointer',
                  fontSize: 12
                }}
              >
                Ban
              </button>
            </div>
          )}
        </div>
      ))}
    </div>
  )
}

// Join Requests Tab Component
function JoinRequestsTab({
  requests,
  onApprove,
  onDeny
}: {
  requests: JoinRequest[]
  onApprove: (userId: string) => void
  onDeny: (userId: string) => void
}) {
  if (requests.length === 0) {
    return (
      <div style={{ textAlign: 'center', padding: 40, color: '#666' }}>
        <p>No pending join requests</p>
      </div>
    )
  }

  return (
    <div style={{ padding: 24 }}>
      <h3 style={{ marginTop: 0, marginBottom: 16 }}>Pending Join Requests</h3>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
        {requests.map(request => (
          <div
            key={request.id}
            style={{
              padding: 16,
              border: '1px solid #e0e0e0',
              borderRadius: 8,
              display: 'flex',
              justifyContent: 'space-between',
              alignItems: 'center'
            }}
          >
            <div>
              <div style={{ fontWeight: 500, marginBottom: 4 }}>
                User ID: {request.user_id}
              </div>
              <div style={{ fontSize: 14, color: '#666' }}>
                Requested: {new Date(request.created_at).toLocaleString()}
              </div>
            </div>
            <div style={{ display: 'flex', gap: 8 }}>
              <button
                onClick={() => onApprove(request.user_id)}
                style={{
                  padding: '8px 16px',
                  backgroundColor: '#10b981',
                  color: 'white',
                  border: 'none',
                  borderRadius: 6,
                  cursor: 'pointer'
                }}
              >
                ✓ Approve
              </button>
              <button
                onClick={() => onDeny(request.user_id)}
                style={{
                  padding: '8px 16px',
                  backgroundColor: '#ef4444',
                  color: 'white',
                  border: 'none',
                  borderRadius: 6,
                  cursor: 'pointer'
                }}
              >
                ✗ Deny
              </button>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// IP Allocations Tab Component
function IPAllocationsTab({
  allocations,
  network,
  isOwnerOrAdmin,
  onReleaseIP
}: {
  allocations: IPAllocation[]
  network: Network
  isOwnerOrAdmin: boolean
  onReleaseIP: (userId: string) => void
}) {
  // Calculate CIDR stats
  const getCIDRStats = (cidr: string) => {
    try {
      const [, maskBits] = cidr.split('/')
      const mask = parseInt(maskBits)
      const totalIPs = Math.pow(2, 32 - mask)
      const usableIPs = totalIPs - 2 // Subtract network and broadcast addresses
      const allocatedIPs = allocations.length
      const availableIPs = usableIPs - allocatedIPs
      const usagePercentage = Math.round((allocatedIPs / usableIPs) * 100)

      return { totalIPs, usableIPs, allocatedIPs, availableIPs, usagePercentage }
    } catch {
      return { totalIPs: 0, usableIPs: 0, allocatedIPs: 0, availableIPs: 0, usagePercentage: 0 }
    }
  }

  const stats = getCIDRStats(network.cidr)

  return (
    <div style={{ backgroundColor: '#fff', border: '1px solid #dee2e6', borderRadius: 8, padding: 24 }}>
      <h3 style={{ marginTop: 0, marginBottom: 20 }}>IP Address Allocations</h3>

      {/* Statistics */}
      <div style={{
        backgroundColor: '#f8f9fa',
        padding: 20,
        borderRadius: 8,
        marginBottom: 24,
        display: 'grid',
        gridTemplateColumns: 'repeat(auto-fit, minmax(150px, 1fr))',
        gap: 16
      }}>
        <div>
          <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Network</div>
          <div style={{ fontSize: 18, fontWeight: 600 }}>{network.cidr}</div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Allocated</div>
          <div style={{ fontSize: 18, fontWeight: 600, color: '#3b82f6' }}>{stats.allocatedIPs}</div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Available</div>
          <div style={{ fontSize: 18, fontWeight: 600, color: '#10b981' }}>{stats.availableIPs}</div>
        </div>
        <div>
          <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Usage</div>
          <div style={{ fontSize: 18, fontWeight: 600 }}>{stats.usagePercentage}%</div>
        </div>
      </div>

      {/* Progress Bar */}
      <div style={{ marginBottom: 24 }}>
        <div style={{
          width: '100%',
          height: 8,
          backgroundColor: '#e5e7eb',
          borderRadius: 4,
          overflow: 'hidden'
        }}>
          <div style={{
            width: `${stats.usagePercentage}%`,
            height: '100%',
            backgroundColor: stats.usagePercentage > 80 ? '#ef4444' : stats.usagePercentage > 50 ? '#f59e0b' : '#10b981',
            transition: 'width 0.3s ease'
          }} />
        </div>
      </div>

      {/* Allocations List */}
      {allocations.length === 0 ? (
        <div style={{ textAlign: 'center', padding: 40, color: '#6b7280' }}>
          <p>No IP addresses allocated yet</p>
        </div>
      ) : (
        <div style={{ display: 'grid', gap: 12 }}>
          {allocations.map((allocation) => (
            <div
              key={`${allocation.user_id}-${allocation.ip}`}
              style={{
                display: 'flex',
                justifyContent: 'space-between',
                alignItems: 'center',
                padding: 16,
                backgroundColor: '#f9fafb',
                border: '1px solid #e5e7eb',
                borderRadius: 8
              }}
            >
              <div style={{ flex: 1 }}>
                <div style={{
                  fontFamily: 'monospace',
                  fontSize: 16,
                  fontWeight: 600,
                  color: '#111827',
                  marginBottom: 4
                }}>
                  {allocation.ip}
                </div>
                <div style={{ fontSize: 12, color: '#6b7280' }}>
                  User ID: {allocation.user_id}
                </div>
              </div>

              {isOwnerOrAdmin && (
                <button
                  onClick={() => onReleaseIP(allocation.user_id)}
                  style={{
                    padding: '8px 16px',
                    backgroundColor: '#ef4444',
                    color: 'white',
                    border: 'none',
                    borderRadius: 6,
                    cursor: 'pointer',
                    fontSize: 13,
                    fontWeight: 500
                  }}
                >
                  Release
                </button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

// Settings Tab Component
interface SettingsTabProps {
  network: Network
  onUpdate: () => void
}

function SettingsTab({ network, onUpdate }: SettingsTabProps) {
  const [name, setName] = useState(network.name)
  const [visibility, setVisibility] = useState<'public' | 'private'>(network.visibility)
  const [joinPolicy, setJoinPolicy] = useState<'open' | 'approval' | 'invite'>(network.join_policy)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setSuccess(false)

    // Validation
    if (!name.trim()) {
      setError('Network name is required')
      return
    }

    if (name.trim().length < 3 || name.trim().length > 64) {
      setError('Network name must be between 3 and 64 characters')
      return
    }

    setLoading(true)

    try {
      const token = getAccessToken()
      if (!token) {
        setError('You must be logged in')
        return
      }

      // Build patch object (only send changed fields)
      const patch: {
        name?: string
        visibility?: 'public' | 'private'
        join_policy?: 'open' | 'approval' | 'invite'
      } = {}

      if (name.trim() !== network.name) patch.name = name.trim()
      if (visibility !== network.visibility) patch.visibility = visibility
      if (joinPolicy !== network.join_policy) patch.join_policy = joinPolicy

      // Only send patch if something changed
      if (Object.keys(patch).length === 0) {
        setError('No changes detected')
        return
      }

      await updateNetwork(network.id, patch, token)
      setSuccess(true)
      onUpdate() // Reload network data
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to update network settings')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div style={{ backgroundColor: '#fff', border: '1px solid #dee2e6', borderRadius: 8, padding: 24 }}>
      <h3 style={{ marginTop: 0 }}>Network Settings</h3>

      {error && (
        <div style={{
          padding: 12,
          backgroundColor: '#f8d7da',
          color: '#842029',
          borderRadius: 6,
          marginBottom: 16,
          fontSize: 14
        }}>
          {error}
        </div>
      )}

      {success && (
        <div style={{
          padding: 12,
          backgroundColor: '#d1e7dd',
          color: '#0f5132',
          borderRadius: 6,
          marginBottom: 16,
          fontSize: 14
        }}>
          Settings updated successfully!
        </div>
      )}

      <form onSubmit={handleSubmit}>
        <div style={{ marginBottom: 20 }}>
          <label style={{ display: 'block', marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
            Network Name *
          </label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            disabled={loading}
            style={{
              width: '100%',
              padding: 10,
              border: '1px solid #dee2e6',
              borderRadius: 6,
              fontSize: 14,
              boxSizing: 'border-box',
              opacity: loading ? 0.6 : 1
            }}
          />
          <small style={{ color: '#666', fontSize: 12 }}>
            3-64 characters
          </small>
        </div>

        <div style={{ marginBottom: 20 }}>
          <label style={{ display: 'block', marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
            Visibility *
          </label>
          <select
            value={visibility}
            onChange={(e) => setVisibility(e.target.value as 'public' | 'private')}
            disabled={loading}
            style={{
              width: '100%',
              padding: 10,
              border: '1px solid #dee2e6',
              borderRadius: 6,
              fontSize: 14,
              boxSizing: 'border-box',
              opacity: loading ? 0.6 : 1
            }}
          >
            <option value="public">Public (anyone can see)</option>
            <option value="private">Private (invite-only)</option>
          </select>
        </div>

        <div style={{ marginBottom: 24 }}>
          <label style={{ display: 'block', marginBottom: 6, fontWeight: 500, fontSize: 14 }}>
            Join Policy *
          </label>
          <select
            value={joinPolicy}
            onChange={(e) => setJoinPolicy(e.target.value as 'open' | 'approval' | 'invite')}
            disabled={loading}
            style={{
              width: '100%',
              padding: 10,
              border: '1px solid #dee2e6',
              borderRadius: 6,
              fontSize: 14,
              boxSizing: 'border-box',
              opacity: loading ? 0.6 : 1
            }}
          >
            <option value="open">Open (auto-approve)</option>
            <option value="approval">Approval Required</option>
            <option value="invite">Invite Only</option>
          </select>
        </div>

        <div style={{
          padding: 16,
          backgroundColor: '#fff3cd',
          borderRadius: 6,
          marginBottom: 20,
          fontSize: 13,
          color: '#856404'
        }}>
          <strong>Note:</strong> CIDR block cannot be changed after network creation. DNS, MTU, and Split Tunnel settings are not yet editable via UI.
        </div>

        <div style={{ display: 'flex', gap: 12 }}>
          <button
            type="submit"
            disabled={loading}
            style={{
              padding: '10px 24px',
              backgroundColor: loading ? '#ccc' : '#007bff',
              color: 'white',
              border: 'none',
              borderRadius: 6,
              cursor: loading ? 'not-allowed' : 'pointer',
              fontSize: 14,
              fontWeight: 500
            }}
          >
            {loading ? 'Saving...' : 'Save Changes'}
          </button>
        </div>
      </form>
    </div>
  )
}
