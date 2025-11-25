'use client'
import { useEffect, useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import { getUser, getAccessToken } from '../../../../lib/auth'
import { useT } from '../../../../lib/i18n-context'
import {
  getSystemStats,
  listUsers,
  listTenants,
  listAllNetworks,
  listAllDevices,
  listAuditLogs,
  toggleUserAdmin,
  deleteUser,
  deleteTenant,
  deleteDevice,
  SystemStats,
  AdminUser,
  Tenant,
  Network,
  Device,
  AuditLog
} from '../../../../lib/api'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

export default function AdminPage() {
  const router = useRouter()
  const t = useT()
  const [currentUser, setCurrentUser] = useState<any>(null)
  const [activeTab, setActiveTab] = useState<'stats' | 'users' | 'tenants' | 'networks' | 'devices' | 'audit'>('stats')

  const [stats, setStats] = useState<SystemStats | null>(null)
  const [users, setUsers] = useState<AdminUser[]>([])
  const [tenants, setTenants] = useState<Tenant[]>([])
  const [networks, setNetworks] = useState<Network[]>([])
  const [devices, setDevices] = useState<Device[]>([])
  const [auditLogs, setAuditLogs] = useState<AuditLog[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Pagination state
  const [usersOffset, setUsersOffset] = useState(0)
  const [tenantsOffset, setTenantsOffset] = useState(0)
  const [auditPage, setAuditPage] = useState(1)
  const [networksNextCursor, setNetworksNextCursor] = useState('')
  const [devicesNextCursor, setDevicesNextCursor] = useState('')

  // Search state
  const [userSearchQuery, setUserSearchQuery] = useState('')
  const [tenantSearchQuery, setTenantSearchQuery] = useState('')
  const [networkSearchQuery, setNetworkSearchQuery] = useState('')
  const [deviceSearchQuery, setDeviceSearchQuery] = useState('')

  useEffect(() => {
    const user = getUser()
    if (user) {
      setCurrentUser(user)
      if (!user.is_admin) {
        router.push('/en/dashboard')
        return
      }
      loadData()
    }
  }, [])

  const loadData = async () => {
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return

      const [statsRes, usersRes, tenantsRes, networksRes, devicesRes, auditRes] = await Promise.all([
        getSystemStats(token),
        listUsers(50, 0, token, userSearchQuery),
        listTenants(50, 0, token, tenantSearchQuery),
        listAllNetworks(50, '', token, networkSearchQuery),
        listAllDevices(50, '', token, deviceSearchQuery),
        listAuditLogs(1, 50, token)
      ])

      setStats(statsRes.data)
      setUsers(usersRes.data)
      setTenants(tenantsRes.data)
      setNetworks(networksRes.data)
      setDevices(devicesRes.devices)
      setAuditLogs(auditRes.data)

      // Set pagination data
      setNetworksNextCursor(networksRes.meta?.next_cursor || '')
      setDevicesNextCursor(devicesRes.next_cursor || '')
    } catch (err: any) {
      console.error('Failed to load admin data:', err)
      setError(err.message || t('admin.error.loadData'))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    const user = getUser()
    if (!user?.is_admin) {
      router.push('/en/dashboard')
      return
    }
    setCurrentUser(user)
  }, [router])

  if (!currentUser?.is_admin) {
    return (
      <AuthGuard>
        <div style={{ padding: 24, textAlign: 'center' }}>
          <h2>{t('admin.accessDenied')}</h2>
          <p>{t('admin.accessDeniedMsg')}</p>
        </div>
      </AuthGuard>
    )
  }

  function formatDate(dateString: string | undefined) {
    if (!dateString) return '-'
    return new Date(dateString).toLocaleDateString()
  }

  const handleToggleAdmin = async (userId: string) => {
    try {
      const token = getAccessToken()
      if (!token) return

      await toggleUserAdmin(userId, token)

      // Refresh user list
      const usersRes = await listUsers(50, 0, token)
      setUsers(usersRes.data)
    } catch (err: any) {
      console.error('Failed to toggle admin status:', err)
      alert(t('admin.error.toggleAdmin') + ': ' + (err.message || 'Unknown error'))
    }
  }

  const handleDeleteTenant = async (tenantId: string) => {
    if (!confirm(t('admin.confirm.deleteTenant'))) {
      return
    }

    try {
      const token = getAccessToken()
      if (!token) return

      await deleteTenant(tenantId, token)

      // Refresh tenant list
      const tenantsRes = await listTenants(50, 0, token)
      setTenants(tenantsRes.data)
    } catch (err: any) {
      console.error('Failed to delete tenant:', err)
      alert(t('admin.error.deleteTenant') + ': ' + (err.message || 'Unknown error'))
    }
  }

  const handleDeleteUser = async (userId: string) => {
    if (!confirm(t('admin.confirm.deleteUser'))) {
      return
    }

    try {
      const token = getAccessToken()
      if (!token) return

      await deleteUser(userId, token)

      // Refresh user list
      const usersRes = await listUsers(50, 0, token)
      setUsers(usersRes.data)
    } catch (err: any) {
      console.error('Failed to delete user:', err)
      alert(t('admin.error.deleteUser') + ': ' + (err.message || 'Unknown error'))
    }
  }

  const handleDeleteDevice = async (deviceId: string) => {
    if (!confirm(t('admin.confirm.deleteDevice'))) {
      return
    }

    try {
      const token = getAccessToken()
      if (!token) return

      await deleteDevice(deviceId, token)

      // Refresh device list
      const devicesRes = await listAllDevices(50, '', token)
      setDevices(devicesRes.devices)
    } catch (err: any) {
      console.error('Failed to delete device:', err)
      alert(t('admin.error.deleteDevice') + ': ' + (err.message || 'Unknown error'))
    }
  }

  const handleUsersPageChange = async (newOffset: number) => {
    if (newOffset < 0) return
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listUsers(50, newOffset, token, userSearchQuery)
      setUsers(res.data)
      setUsersOffset(newOffset)
    } catch (err: any) {
      console.error('Failed to load users:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleTenantsPageChange = async (newOffset: number) => {
    if (newOffset < 0) return
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listTenants(50, newOffset, token, tenantSearchQuery)
      setTenants(res.data)
      setTenantsOffset(newOffset)
    } catch (err: any) {
      console.error('Failed to load tenants:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleNetworksNextPage = async () => {
    if (!networksNextCursor) return
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listAllNetworks(50, networksNextCursor, token, networkSearchQuery)
      setNetworks(res.data)
      setNetworksNextCursor(res.meta?.next_cursor || '')
    } catch (err: any) {
      console.error('Failed to load networks:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDevicesNextPage = async () => {
    if (!devicesNextCursor) return
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listAllDevices(50, devicesNextCursor, token, deviceSearchQuery)
      setDevices(res.devices)
      setDevicesNextCursor(res.next_cursor || '')
    } catch (err: any) {
      console.error('Failed to load devices:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleAuditPageChange = async (newPage: number) => {
    if (newPage < 1) return
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listAuditLogs(newPage, 50, token)
      setAuditLogs(res.data)
      setAuditPage(newPage)
    } catch (err: any) {
      console.error('Failed to load audit logs:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleUserSearch = async (e: FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listUsers(50, 0, token, userSearchQuery)
      setUsers(res.data)
      setUsersOffset(0)
    } catch (err: any) {
      console.error('Failed to search users:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleTenantSearch = async (e: FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listTenants(50, 0, token, tenantSearchQuery)
      setTenants(res.data)
      setTenantsOffset(0)
    } catch (err: any) {
      console.error('Failed to search tenants:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleNetworkSearch = async (e: FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listAllNetworks(50, '', token, networkSearchQuery)
      setNetworks(res.data)
      setNetworksNextCursor(res.meta?.next_cursor || '')
    } catch (err: any) {
      console.error('Failed to search networks:', err)
    } finally {
      setLoading(false)
    }
  }

  const handleDeviceSearch = async (e: FormEvent) => {
    e.preventDefault()
    try {
      setLoading(true)
      const token = getAccessToken()
      if (!token) return
      const res = await listAllDevices(50, '', token, deviceSearchQuery)
      setDevices(res.devices)
      setDevicesNextCursor(res.next_cursor || '')
    } catch (err: any) {
      console.error('Failed to search devices:', err)
    } finally {
      setLoading(false)
    }
  }

  return (
    <AuthGuard>
      <div style={{
        display: 'flex',
        flexDirection: 'column',
        minHeight: '100vh',
        fontFamily: 'system-ui, -apple-system, sans-serif',
        backgroundColor: '#f8f9fa'
      }}>
        {/* Header */}
        <div style={{
          padding: '16px 24px',
          backgroundColor: 'white',
          borderBottom: '1px solid #dee2e6',
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center'
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            <button
              onClick={() => router.push('/en/dashboard')}
              style={{
                padding: '6px 12px',
                backgroundColor: '#6c757d',
                color: 'white',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              ← {t('admin.back')}
            </button>
            <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>
              👑 {t('admin.title')}
            </h1>
          </div>
          <div style={{
            padding: '6px 12px',
            backgroundColor: '#ffc107',
            color: '#856404',
            borderRadius: 6,
            fontSize: 13,
            fontWeight: 600
          }}>
            {t('admin.badge')}
          </div>
        </div>

        {/* Tabs */}
        <div style={{
          padding: '16px 24px',
          backgroundColor: 'white',
          borderBottom: '1px solid #dee2e6'
        }}>
          <div style={{ display: 'flex', gap: 8 }}>
            <button
              onClick={() => setActiveTab('stats')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'stats' ? '#007bff' : 'transparent',
                color: activeTab === 'stats' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              📊 {t('admin.tabs.stats')}
            </button>
            <button
              onClick={() => setActiveTab('users')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'users' ? '#007bff' : 'transparent',
                color: activeTab === 'users' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              👥 {t('admin.tabs.users')}
            </button>
            <button
              onClick={() => setActiveTab('tenants')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'tenants' ? '#007bff' : 'transparent',
                color: activeTab === 'tenants' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              🏢 {t('admin.tabs.tenants')}
            </button>
            <button
              onClick={() => setActiveTab('networks')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'networks' ? '#007bff' : 'transparent',
                color: activeTab === 'networks' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              🌐 {t('admin.tabs.networks')}
            </button>
            <button
              onClick={() => setActiveTab('devices')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'devices' ? '#007bff' : 'transparent',
                color: activeTab === 'devices' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              💻 {t('admin.tabs.devices')}
            </button>
            <button
              onClick={() => setActiveTab('audit')}
              style={{
                padding: '8px 16px',
                backgroundColor: activeTab === 'audit' ? '#007bff' : 'transparent',
                color: activeTab === 'audit' ? 'white' : '#6c757d',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: 500
              }}
            >
              📜 {t('admin.tabs.audit')}
            </button>
          </div>
        </div>

        {/* Main Content */}
        <div style={{ flex: 1, padding: 24, maxWidth: 1400, margin: '0 auto', width: '100%' }}>

          {loading && (
            <div style={{ textAlign: 'center', padding: 40 }}>
              <div style={{ fontSize: 24 }}>⏳ {t('admin.loading')}</div>
            </div>
          )}

          {error && (
            <div style={{
              padding: 16,
              backgroundColor: '#f8d7da',
              color: '#842029',
              borderRadius: 8,
              marginBottom: 24
            }}>
              {error}
            </div>
          )}

          {/* Statistics Tab */}
          {activeTab === 'stats' && stats && (
            <div>
              <div style={{
                display: 'grid',
                gridTemplateColumns: 'repeat(auto-fit, minmax(250px, 1fr))',
                gap: 16,
                marginBottom: 24
              }}>
                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #007bff'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.users')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#007bff' }}>
                    {stats.total_users}
                  </div>
                </div>

                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #28a745'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.tenants')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#28a745' }}>
                    {stats.total_tenants}
                  </div>
                </div>

                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #ffc107'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.networks')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#ffc107' }}>
                    {stats.total_networks}
                  </div>
                </div>

                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #17a2b8'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.devices')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#17a2b8' }}>
                    {stats.total_devices}
                  </div>
                </div>

                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #6f42c1'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.connections')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#6f42c1' }}>
                    {stats.active_connections}
                  </div>
                </div>

                <div style={{
                  backgroundColor: 'white',
                  borderRadius: 12,
                  padding: 24,
                  boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                  borderLeft: '4px solid #dc3545'
                }}>
                  <div style={{ fontSize: 14, color: '#6c757d', marginBottom: 8 }}>
                    {t('admin.stats.messages')}
                  </div>
                  <div style={{ fontSize: 32, fontWeight: 600, color: '#dc3545' }}>
                    {stats.messages_today}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Users Tab */}
          {activeTab === 'users' && (
            <div>
              <div style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 24,
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
              }}>
                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                  {t('admin.users.title')}
                </h2>

                <form onSubmit={handleUserSearch} style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
                  <input
                    type="text"
                    value={userSearchQuery}
                    onChange={(e) => setUserSearchQuery(e.target.value)}
                    placeholder={t('admin.users.searchPlaceholder')}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      border: '1px solid #ced4da',
                      borderRadius: 4,
                      fontSize: 14,
                      color: '#495057'
                    }}
                  />
                  <button
                    type="submit"
                    style={{
                      padding: '8px 16px',
                      backgroundColor: '#007bff',
                      color: 'white',
                      border: 'none',
                      borderRadius: 4,
                      cursor: 'pointer',
                      fontSize: 14,
                      fontWeight: 500
                    }}
                  >
                    🔍 {t('admin.users.search')}
                  </button>
                </form>

                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr style={{ borderBottom: '2px solid #dee2e6' }}>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.email')}
                        </th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.role')}
                        </th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.provider')}
                        </th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.tenantId')}
                        </th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.created')}
                        </th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>
                          {t('admin.users.table.actions')}
                        </th>
                      </tr>
                    </thead>
                    <tbody>
                      {users.map((user) => (
                        <tr key={user.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                          <td style={{ padding: '12px', fontSize: 14, color: '#212529' }}>
                            {user.email}
                          </td>
                          <td style={{ padding: '12px' }}>
                            <span style={{
                              padding: '4px 8px',
                              backgroundColor: user.is_admin ? '#d1e7dd' : user.is_moderator ? '#cfe2ff' : '#f8f9fa',
                              color: user.is_admin ? '#0f5132' : user.is_moderator ? '#084298' : '#6c757d',
                              borderRadius: 4,
                              fontSize: 12,
                              fontWeight: 500
                            }}>
                              {user.is_admin ? `👑 ${t('admin.users.role.admin')}` : user.is_moderator ? `🛡️ ${t('admin.users.role.moderator')}` : `👤 ${t('admin.users.role.user')}`}
                            </span>
                          </td>
                          <td style={{ padding: '12px' }}>
                            <span style={{
                              padding: '4px 8px',
                              backgroundColor: user.auth_provider === 'oidc' ? '#e2e3e5' : '#fff3cd',
                              color: user.auth_provider === 'oidc' ? '#383d41' : '#856404',
                              borderRadius: 4,
                              fontSize: 12,
                              fontWeight: 500
                            }}>
                              {user.auth_provider === 'oidc' ? t('admin.users.provider.sso') : t('admin.users.provider.local')}
                            </span>
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                            {user.tenant_id.substring(0, 12)}...
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, color: '#6c757d' }}>
                            {formatDate(user.created_at)}
                          </td>
                          <td style={{ padding: '12px' }}>
                            <button
                              onClick={() => handleToggleAdmin(user.id)}
                              disabled={user.id === currentUser?.id}
                              style={{
                                padding: '4px 8px',
                                backgroundColor: user.is_admin ? '#dc3545' : '#198754',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                fontSize: 12,
                                cursor: user.id === currentUser?.id ? 'not-allowed' : 'pointer',
                                opacity: user.id === currentUser?.id ? 0.6 : 1
                              }}
                              title={user.id === currentUser?.id ? t('admin.users.tooltip.selfAdmin') : (user.is_admin ? t('admin.users.action.revokeAdmin') : t('admin.users.action.makeAdmin'))}
                            >
                              {user.is_admin ? t('admin.users.action.revokeAdmin') : t('admin.users.action.makeAdmin')}
                            </button>
                            <button
                              onClick={() => handleDeleteUser(user.id)}
                              disabled={user.id === currentUser?.id}
                              style={{
                                marginLeft: 8,
                                padding: '4px 8px',
                                backgroundColor: '#dc3545',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                fontSize: 12,
                                cursor: user.id === currentUser?.id ? 'not-allowed' : 'pointer',
                                opacity: user.id === currentUser?.id ? 0.6 : 1
                              }}
                              title={user.id === currentUser?.id ? t('admin.users.tooltip.selfDelete') : t('admin.users.action.delete')}
                            >
                              {t('admin.users.action.delete')}
                            </button>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16, gap: 8 }}>
                  <button
                    onClick={() => handleUsersPageChange(usersOffset - 50)}
                    disabled={usersOffset === 0 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (usersOffset === 0 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (usersOffset === 0 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.previous')}
                  </button>
                  <button
                    onClick={() => handleUsersPageChange(usersOffset + 50)}
                    disabled={users.length < 50 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (users.length < 50 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (users.length < 50 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.next')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Tenants Tab */}
          {activeTab === 'tenants' && (
            <div>
              <div style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 24,
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
              }}>
                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                  {t('admin.tenants.title')}
                </h2>

                <form onSubmit={handleTenantSearch} style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
                  <input
                    type="text"
                    value={tenantSearchQuery}
                    onChange={(e) => setTenantSearchQuery(e.target.value)}
                    placeholder={t('admin.tenants.searchPlaceholder')}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      border: '1px solid #ced4da',
                      borderRadius: 4,
                      fontSize: 14,
                      color: '#495057'
                    }}
                  />
                  <button
                    type="submit"
                    style={{
                      padding: '8px 16px',
                      backgroundColor: '#007bff',
                      color: 'white',
                      border: 'none',
                      borderRadius: 4,
                      cursor: 'pointer',
                      fontSize: 14,
                      fontWeight: 500
                    }}
                  >
                    🔍 {t('admin.users.search')}
                  </button>
                </form>

                <div style={{ display: 'grid', gap: 16 }}>
                  {tenants.map((tenant) => (
                    <div
                      key={tenant.id}
                      style={{
                        padding: 20,
                        border: '1px solid #dee2e6',
                        borderRadius: 8,
                        backgroundColor: '#f8f9fa'
                      }}
                    >
                      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 12 }}>
                        <div>
                          <h3 style={{ margin: '0 0 4px 0', fontSize: 16, fontWeight: 600 }}>
                            {tenant.name}
                          </h3>
                          <div style={{ fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                            {t('admin.tenants.id')}: {tenant.id}
                          </div>
                        </div>
                        <button
                          onClick={() => handleDeleteTenant(tenant.id)}
                          style={{
                            padding: '6px 12px',
                            backgroundColor: '#dc3545',
                            color: 'white',
                            border: 'none',
                            borderRadius: 6,
                            fontSize: 13,
                            cursor: 'pointer'
                          }}
                        >
                          {t('admin.users.action.delete')}
                        </button>
                      </div>

                      <div style={{ display: 'flex', gap: 24, fontSize: 14, color: '#6c757d' }}>
                        <div>
                          <span style={{ fontWeight: 500 }}>{t('admin.tenants.created')}:</span> {formatDate(tenant.created_at)}
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16, gap: 8 }}>
                  <button
                    onClick={() => handleTenantsPageChange(tenantsOffset - 50)}
                    disabled={tenantsOffset === 0 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (tenantsOffset === 0 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (tenantsOffset === 0 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.previous')}
                  </button>
                  <button
                    onClick={() => handleTenantsPageChange(tenantsOffset + 50)}
                    disabled={tenants.length < 50 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (tenants.length < 50 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (tenants.length < 50 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.next')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Networks Tab */}
          {activeTab === 'networks' && (
            <div>
              <div style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 24,
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
              }}>
                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                  {t('admin.networks.title')}
                </h2>

                <form onSubmit={handleNetworkSearch} style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
                  <input
                    type="text"
                    value={networkSearchQuery}
                    onChange={(e) => setNetworkSearchQuery(e.target.value)}
                    placeholder={t('admin.networks.searchPlaceholder')}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      border: '1px solid #ced4da',
                      borderRadius: 4,
                      fontSize: 14,
                      color: '#495057'
                    }}
                  />
                  <button
                    type="submit"
                    style={{
                      padding: '8px 16px',
                      backgroundColor: '#007bff',
                      color: 'white',
                      border: 'none',
                      borderRadius: 4,
                      cursor: 'pointer',
                      fontSize: 14,
                      fontWeight: 500
                    }}
                  >
                    🔍 {t('admin.users.search')}
                  </button>
                </form>

                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr style={{ borderBottom: '2px solid #dee2e6' }}>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.networks.table.name')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.networks.table.cidr')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.users.table.tenantId')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.users.table.created')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {networks.map((network) => (
                        <tr key={network.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                          <td style={{ padding: '12px', fontSize: 14, color: '#212529', fontWeight: 500 }}>
                            {network.name}
                          </td>
                          <td style={{ padding: '12px', fontSize: 14, fontFamily: 'monospace', color: '#6c757d' }}>
                            {network.cidr}
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                            {network.tenant_id.substring(0, 12)}...
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, color: '#6c757d' }}>
                            {formatDate(network.created_at)}
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
                  <button
                    onClick={handleNetworksNextPage}
                    disabled={!networksNextCursor || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (!networksNextCursor || loading) ? 'not-allowed' : 'pointer',
                      opacity: (!networksNextCursor || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.nextPage')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Devices Tab */}
          {activeTab === 'devices' && (
            <div>
              <div style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 24,
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
              }}>
                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                  {t('admin.devices.title')}
                </h2>

                <form onSubmit={handleDeviceSearch} style={{ marginBottom: 16, display: 'flex', gap: 8 }}>
                  <input
                    type="text"
                    value={deviceSearchQuery}
                    onChange={(e) => setDeviceSearchQuery(e.target.value)}
                    placeholder={t('admin.devices.searchPlaceholder')}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      border: '1px solid #ced4da',
                      borderRadius: 4,
                      fontSize: 14,
                      color: '#495057'
                    }}
                  />
                  <button
                    type="submit"
                    style={{
                      padding: '8px 16px',
                      backgroundColor: '#007bff',
                      color: 'white',
                      border: 'none',
                      borderRadius: 4,
                      cursor: 'pointer',
                      fontSize: 14,
                      fontWeight: 500
                    }}
                  >
                    🔍 {t('admin.users.search')}
                  </button>
                </form>

                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                    <thead>
                      <tr style={{ borderBottom: '2px solid #dee2e6' }}>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.networks.table.name')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.devices.table.ip')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.devices.table.publicKey')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.users.table.tenantId')}</th>
                        <th style={{ padding: '12px', textAlign: 'left', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.devices.table.lastSeen')}</th>
                        <th style={{ padding: '12px', textAlign: 'right', fontSize: 14, fontWeight: 600, color: '#6c757d' }}>{t('admin.users.table.actions')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {devices.map((device) => (
                        <tr key={device.id} style={{ borderBottom: '1px solid #f0f0f0' }}>
                          <td style={{ padding: '12px', fontSize: 14, color: '#212529', fontWeight: 500 }}>
                            {device.name}
                          </td>
                          <td style={{ padding: '12px', fontSize: 14, fontFamily: 'monospace', color: '#6c757d' }}>
                            {device.last_ip || '-'}
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                            {device.pubkey.substring(0, 12)}...
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, fontFamily: 'monospace', color: '#6c757d' }}>
                            {device.tenant_id.substring(0, 12)}...
                          </td>
                          <td style={{ padding: '12px', fontSize: 13, color: '#6c757d' }}>
                            {formatDate(device.last_seen)}
                          </td>
                          <td style={{ padding: '12px', textAlign: 'right' }}>
                            <button
                              onClick={() => handleDeleteDevice(device.id)}
                              style={{
                                padding: '4px 8px',
                                backgroundColor: '#dc3545',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: 'pointer',
                                fontSize: 12
                              }}
                            >
                              {t('admin.users.action.delete')}
                            </button>
                          </td>
                        </tr>
                      ))}
                      {devices.length === 0 && (
                        <tr>
                          <td colSpan={6} style={{ padding: '24px', textAlign: 'center', color: '#6c757d' }}>
                            {t('admin.devices.noDevices')}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16 }}>
                  <button
                    onClick={handleDevicesNextPage}
                    disabled={!devicesNextCursor || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (!devicesNextCursor || loading) ? 'not-allowed' : 'pointer',
                      opacity: (!devicesNextCursor || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.nextPage')}
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* Audit Logs Tab */}
          {activeTab === 'audit' && (
            <div>
              <div style={{
                backgroundColor: 'white',
                borderRadius: 12,
                padding: 24,
                boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
              }}>
                <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600 }}>
                  {t('admin.audit.title')}
                </h2>

                <div style={{ overflowX: 'auto' }}>
                  <table style={{ width: '100%', borderCollapse: 'collapse', fontSize: 14 }}>
                    <thead>
                      <tr style={{ borderBottom: '2px solid #dee2e6', textAlign: 'left' }}>
                        <th style={{ padding: '12px', color: '#495057' }}>{t('admin.audit.table.time')}</th>
                        <th style={{ padding: '12px', color: '#495057' }}>{t('admin.audit.table.action')}</th>
                        <th style={{ padding: '12px', color: '#495057' }}>{t('admin.audit.table.actor')}</th>
                        <th style={{ padding: '12px', color: '#495057' }}>{t('admin.audit.table.object')}</th>
                        <th style={{ padding: '12px', color: '#495057' }}>{t('admin.audit.table.details')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {auditLogs.map((log) => (
                        <tr key={log.seq} style={{ borderBottom: '1px solid #dee2e6' }}>
                          <td style={{ padding: '12px' }}>
                            {new Date(log.timestamp).toLocaleString()}
                          </td>
                          <td style={{ padding: '12px' }}>
                            <span style={{
                              padding: '4px 8px',
                              backgroundColor: '#e9ecef',
                              borderRadius: 4,
                              fontSize: 12,
                              fontWeight: 600,
                              color: '#495057'
                            }}>
                              {log.action}
                            </span>
                          </td>
                          <td style={{ padding: '12px', fontFamily: 'monospace' }}>
                            {log.actor}
                          </td>
                          <td style={{ padding: '12px', fontFamily: 'monospace' }}>
                            {log.object}
                          </td>
                          <td style={{ padding: '12px' }}>
                            <pre style={{
                              margin: 0,
                              fontSize: 11,
                              maxWidth: '300px',
                              overflow: 'hidden',
                              textOverflow: 'ellipsis',
                              whiteSpace: 'nowrap'
                            }} title={JSON.stringify(log.details, null, 2)}>
                              {JSON.stringify(log.details)}
                            </pre>
                          </td>
                        </tr>
                      ))}
                      {auditLogs.length === 0 && (
                        <tr>
                          <td colSpan={5} style={{ padding: '24px', textAlign: 'center', color: '#6c757d' }}>
                            {t('admin.audit.noLogs')}
                          </td>
                        </tr>
                      )}
                    </tbody>
                  </table>
                </div>
                <div style={{ display: 'flex', justifyContent: 'flex-end', marginTop: 16, gap: 8 }}>
                  <button
                    onClick={() => handleAuditPageChange(auditPage - 1)}
                    disabled={auditPage === 1 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (auditPage === 1 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (auditPage === 1 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.previous')}
                  </button>
                  <button
                    onClick={() => handleAuditPageChange(auditPage + 1)}
                    disabled={auditLogs.length < 50 || loading}
                    style={{
                      padding: '6px 12px',
                      backgroundColor: '#f8f9fa',
                      border: '1px solid #dee2e6',
                      borderRadius: 4,
                      cursor: (auditLogs.length < 50 || loading) ? 'not-allowed' : 'pointer',
                      opacity: (auditLogs.length < 50 || loading) ? 0.6 : 1
                    }}
                  >
                    {t('admin.pagination.next')}
                  </button>
                </div>
              </div>
            </div>
          )}
        </div>

        <Footer />
      </div>
    </AuthGuard>
  )
}
