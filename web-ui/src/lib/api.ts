import { getAccessToken } from './auth'

// Base API client
export async function api(path: string, init?: RequestInit): Promise<any> {
  const base = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
  const res = await fetch(base + path, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers || {}),
    },
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Unknown error' }))
    throw new Error(error.message || `HTTP ${res.status}`)
  }

  return res.json()
}

// Helper to get auth header
function getAuthHeader(): { Authorization: string } | {} {
  const token = getAccessToken()
  return token ? { Authorization: `Bearer ${token}` } : {}
}

// Authentication API client
export interface LoginRequest {
  email: string
  password: string
  code?: string
}

export interface RegisterRequest {
  name: string
  email: string
  password: string
}

export interface AuthResponse {
  data: {
    access_token: string
    refresh_token: string
    expires_in: number
    token_type: string
    user: {
      id: string
      name: string
      email: string
      is_admin: boolean
      is_moderator: boolean
      tenant_id: string
    }
  }
}

export async function login(req: LoginRequest): Promise<AuthResponse> {
  return api('/v1/auth/login', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function register(req: RegisterRequest): Promise<AuthResponse> {
  return api('/v1/auth/register', {
    method: 'POST',
    body: JSON.stringify(req),
  })
}

export async function refreshToken(refreshToken: string): Promise<AuthResponse> {
  return api('/v1/auth/refresh', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken }),
  })
}

export async function logout(accessToken: string): Promise<void> {
  await api('/v1/auth/logout', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export async function generate2FA(accessToken: string): Promise<{ secret: string; url: string }> {
  const res = await api('/v1/auth/2fa/generate', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
  return res.data
}

export async function enable2FA(accessToken: string, secret: string, code: string): Promise<void> {
  await api('/v1/auth/2fa/enable', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ secret, code }),
  })
}

export async function disable2FA(accessToken: string, code: string): Promise<void> {
  await api('/v1/auth/2fa/disable', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ code }),
  })
}

/**
 * Generate recovery codes (requires 2FA to be enabled)
 * @param accessToken - JWT access token
 * @param code - Current TOTP code to verify
 * @returns Array of 8 one-time recovery codes
 */
export async function generateRecoveryCodes(accessToken: string, code: string): Promise<{ codes: string[] }> {
  const res = await api('/v1/auth/2fa/recovery-codes', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ code }),
  })
  return res.data
}

/**
 * Get remaining recovery code count
 * @param accessToken - JWT access token
 */
export async function getRecoveryCodeCount(accessToken: string): Promise<{ remaining_codes: number }> {
  const res = await api('/v1/auth/2fa/recovery-codes/count', {
    method: 'GET',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
  return res.data
}

/**
 * Login using a recovery code (when 2FA device is lost)
 * @param email - User email
 * @param password - User password
 * @param recoveryCode - One-time recovery code (format: XXXXX-XXXXX)
 */
export async function loginWithRecoveryCode(email: string, password: string, recoveryCode: string): Promise<AuthResponse> {
  return api('/v1/auth/2fa/recovery', {
    method: 'POST',
    body: JSON.stringify({ email, password, recovery_code: recoveryCode }),
  })
}

/**
 * Change user password
 * @param accessToken - JWT access token
 * @param oldPassword - Current password
 * @param newPassword - New password
 */
export async function changePassword(accessToken: string, oldPassword: string, newPassword: string): Promise<void> {
  await api('/v1/auth/password', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword }),
  })
}

// Network Management API
export interface Network {
  id: string
  tenant_id: string
  name: string
  visibility: 'public' | 'private'
  join_policy: 'open' | 'approval' | 'invite'
  cidr: string
  dns?: string
  mtu?: number
  split_tunnel?: boolean
  created_by: string
  created_at: string
  updated_at: string
}

export interface CreateNetworkRequest {
  name: string
  visibility: 'public' | 'private'
  join_policy: 'open' | 'approval' | 'invite'
  cidr: string
  dns?: string
  mtu?: number
  split_tunnel?: boolean
}

export interface ListNetworksResponse {
  data: Network[]
  pagination: {
    limit: number
    next_cursor?: string
  }
}

/**
 * List networks with optional filtering
 * @param visibility - 'public', 'mine', or 'all' (admin only)
 * @param accessToken - JWT access token
 * @param cursor - Pagination cursor (optional)
 * @param search - Search by name (optional)
 */
export async function listNetworks(
  visibility: 'public' | 'mine' | 'all' = 'public',
  accessToken: string,
  cursor?: string,
  search?: string
): Promise<ListNetworksResponse> {
  const params = new URLSearchParams({ visibility })
  if (cursor) params.append('cursor', cursor)
  if (search) params.append('search', search)

  return api(`/v1/networks?${params.toString()}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Create a new network
 * @param req - Network creation request
 * @param accessToken - JWT access token
 */
export async function createNetwork(
  req: CreateNetworkRequest,
  accessToken: string
): Promise<{ data: Network }> {
  // Generate idempotency key (required for mutations)
  const idempotencyKey = `create-network-${Date.now()}-${Math.random().toString(36).substring(7)}`

  return api('/v1/networks', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify(req),
  })
}

/**
 * Get network by ID
 * @param id - Network ID
 * @param accessToken - JWT access token
 */
export async function getNetwork(
  id: string,
  accessToken: string
): Promise<{ data: Network }> {
  return api(`/v1/networks/${id}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Update network (admin only)
 * @param id - Network ID
 * @param patch - Partial network update
 * @param accessToken - JWT access token
 */
export async function updateNetwork(
  id: string,
  patch: {
    name?: string
    visibility?: 'public' | 'private'
    join_policy?: 'open' | 'approval' | 'invite'
    dns?: string | null
    mtu?: number | null
    split_tunnel?: boolean | null
  },
  accessToken: string
): Promise<{ data: Network }> {
  const idempotencyKey = `update-network-${id}-${Date.now()}`

  return api(`/v1/networks/${id}`, {
    method: 'PATCH',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify(patch),
  })
}

/**
 * Delete network (admin only)
 * @param id - Network ID
 * @param accessToken - JWT access token
 */
export async function deleteNetwork(
  id: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `delete-network-${id}-${Date.now()}`

  await api(`/v1/networks/${id}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
  })
}

// Membership Management API
export interface Membership {
  id: string
  network_id: string
  user_id: string
  role: 'owner' | 'admin' | 'moderator' | 'member'
  status: 'approved' | 'banned'
  joined_at: string
  updated_at: string
}

export interface JoinRequest {
  id: string
  network_id: string
  user_id: string
  status: 'pending' | 'approved' | 'denied'
  created_at: string
  decided_at?: string
}

export interface ListMembersResponse {
  data: Membership[]
  pagination: {
    limit: number
    next_cursor?: string
  }
}

/**
 * Join a network
 * @param networkId - Network ID to join
 * @param accessToken - JWT access token
 */
export async function joinNetwork(
  networkId: string,
  accessToken: string
): Promise<{ data?: Membership; join_request?: JoinRequest }> {
  const idempotencyKey = `join-network-${networkId}-${Date.now()}`

  return api(`/v1/networks/${networkId}/join`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
  })
}

/**
 * List network members
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 * @param status - Filter by status (approved/banned)
 * @param cursor - Pagination cursor
 */
export async function listMembers(
  networkId: string,
  accessToken: string,
  status?: string,
  cursor?: string
): Promise<ListMembersResponse> {
  const params = new URLSearchParams()
  if (status) params.append('status', status)
  if (cursor) params.append('cursor', cursor)

  const queryString = params.toString()
  const url = `/v1/networks/${networkId}/members${queryString ? '?' + queryString : ''}`

  return api(url, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Approve a join request (admin/owner only)
 * @param networkId - Network ID
 * @param userId - User ID to approve
 * @param accessToken - JWT access token
 */
export async function approveMember(
  networkId: string,
  userId: string,
  accessToken: string
): Promise<{ data: Membership }> {
  const idempotencyKey = `approve-${networkId}-${userId}-${Date.now()}`

  return api(`/v1/networks/${networkId}/approve`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({ user_id: userId }),
  })
}

/**
 * Deny a join request (admin/owner only)
 * @param networkId - Network ID
 * @param userId - User ID to deny
 * @param accessToken - JWT access token
 */
export async function denyMember(
  networkId: string,
  userId: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `deny-${networkId}-${userId}-${Date.now()}`

  await api(`/v1/networks/${networkId}/deny`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({ user_id: userId }),
  })
}

/**
 * Kick a member from network (admin/owner only)
 * @param networkId - Network ID
 * @param userId - User ID to kick
 * @param accessToken - JWT access token
 */
export async function kickMember(
  networkId: string,
  userId: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `kick-${networkId}-${userId}-${Date.now()}`

  await api(`/v1/networks/${networkId}/kick`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({ user_id: userId }),
  })
}

/**
 * Ban a member from network (admin/owner only)
 * @param networkId - Network ID
 * @param userId - User ID to ban
 * @param accessToken - JWT access token
 */
export async function banMember(
  networkId: string,
  userId: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `ban-${networkId}-${userId}-${Date.now()}`

  await api(`/v1/networks/${networkId}/ban`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({ user_id: userId }),
  })
}

/**
 * List join requests for a network (admin/owner only)
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 */
export async function listJoinRequests(
  networkId: string,
  accessToken: string
): Promise<{ data: JoinRequest[] }> {
  return api(`/v1/networks/${networkId}/join-requests`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

// IP Allocation API

export interface IPAllocation {
  network_id: string
  user_id: string
  ip: string
}

export interface IPAllocationListResponse {
  data: IPAllocation[]
}

/**
 * List IP allocations for a network
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 */
export async function listIPAllocations(
  networkId: string,
  accessToken: string
): Promise<IPAllocationListResponse> {
  return api(`/v1/networks/${networkId}/ip-allocations`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Allocate IP for current user
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 */
export async function allocateIP(
  networkId: string,
  accessToken: string
): Promise<{ data: IPAllocation }> {
  const idempotencyKey = `allocate-ip-${networkId}-${Date.now()}-${Math.random().toString(36).substring(7)}`

  return api(`/v1/networks/${networkId}/ip-allocations`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
    body: JSON.stringify({}),
  })
}

/**
 * Release own IP allocation
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 */
export async function releaseIP(
  networkId: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `release-ip-${networkId}-${Date.now()}-${Math.random().toString(36).substring(7)}`

  await api(`/v1/networks/${networkId}/ip-allocation`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
  })
}

/**
 * Admin: Release IP allocation for another user
 * @param networkId - Network ID
 * @param userId - User ID whose IP to release
 * @param accessToken - JWT access token
 */
export async function adminReleaseIP(
  networkId: string,
  userId: string,
  accessToken: string
): Promise<void> {
  const idempotencyKey = `admin-release-ip-${networkId}-${userId}-${Date.now()}`

  await api(`/v1/networks/${networkId}/ip-allocations/${userId}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${accessToken}`,
      'Idempotency-Key': idempotencyKey,
    },
  })
}

/**
 * Download WireGuard configuration for a network
 * @param networkId - Network ID
 * @param accessToken - JWT access token
 */
export async function downloadConfig(
  networkId: string,
  accessToken: string
): Promise<Blob> {
  const base = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
  const response = await fetch(`${base}/v1/networks/${networkId}/config`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })

  if (!response.ok) {
    const errorData = await response.json().catch(() => ({}))
    throw new Error(errorData.message || 'Failed to download config')
  }

  return response.blob()
}

// Device Management API

export interface Device {
  id: string
  user_id: string
  tenant_id: string
  name: string
  platform: 'windows' | 'macos' | 'linux' | 'android' | 'ios'
  pubkey: string
  hostname?: string
  os_version?: string
  daemon_ver?: string
  last_seen?: string
  last_ip?: string
  active: boolean
  disabled_at?: string
  created_at: string
  updated_at: string
}

export interface RegisterDeviceRequest {
  name: string
  platform: 'windows' | 'macos' | 'linux' | 'android' | 'ios'
  pubkey: string
  hostname?: string
  os_version?: string
  daemon_ver?: string
}

export interface DeviceListResponse {
  devices: Device[]
  next_cursor?: string
  has_more: boolean
}

/**
 * Register a new device
 * @param request - Device registration data
 * @param accessToken - JWT access token
 */
export async function registerDevice(
  request: RegisterDeviceRequest,
  accessToken: string
): Promise<Device> {
  return api('/v1/devices', {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: JSON.stringify(request),
  })
}

/**
 * List user's devices
 * @param accessToken - JWT access token
 * @param platform - Filter by platform (optional)
 * @param cursor - Pagination cursor (optional)
 */
export async function listDevices(
  accessToken: string,
  platform?: string,
  cursor?: string
): Promise<DeviceListResponse> {
  const params = new URLSearchParams()
  if (platform) params.append('platform', platform)
  if (cursor) params.append('cursor', cursor)

  const queryString = params.toString()
  const url = `/v1/devices${queryString ? '?' + queryString : ''}`

  return api(url, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Get device by ID
 * @param deviceId - Device ID
 * @param accessToken - JWT access token
 */
export async function getDevice(
  deviceId: string,
  accessToken: string
): Promise<Device> {
  return api(`/v1/devices/${deviceId}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Delete a device
 * @param deviceId - Device ID
 * @param accessToken - JWT access token
 */
export async function deleteDevice(
  deviceId: string,
  accessToken: string
): Promise<void> {
  await api(`/v1/devices/${deviceId}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Disable a device
 * @param deviceId - Device ID
 * @param accessToken - JWT access token
 */
export async function disableDevice(
  deviceId: string,
  accessToken: string
): Promise<void> {
  await api(`/v1/devices/${deviceId}/disable`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Enable a device
 * @param deviceId - Device ID
 * @param accessToken - JWT access token
 */
export async function enableDevice(
  deviceId: string,
  accessToken: string
): Promise<void> {
  await api(`/v1/devices/${deviceId}/enable`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Download WireGuard config for a device in a network
 * @param networkId - Network ID
 * @param deviceId - Device ID
 * @param privateKey - Device's WireGuard private key
 * @param accessToken - JWT access token
 * @returns Config file content as text
 */
export async function downloadWireGuardConfig(
  networkId: string,
  deviceId: string,
  privateKey: string,
  accessToken: string
): Promise<string> {
  const base = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
  const params = new URLSearchParams({
    device_id: deviceId,
    private_key: privateKey
  })

  const res = await fetch(`${base}/v1/networks/${networkId}/wg/profile?${params}`, {
    headers: {
      'Authorization': `Bearer ${accessToken}`,
    },
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Failed to download config' }))
    throw new Error(error.message || `HTTP ${res.status}`)
  }

  return res.text()
}

// Chat API

export interface ChatMessage {
  id: string
  scope: string
  user_id: string
  body: string
  redacted: boolean
  deleted_at?: string
  created_at: string
  updated_at?: string
  attachments?: string[]
}

export interface ListChatMessagesResponse {
  messages: ChatMessage[]
  next_cursor: string
  has_more: boolean
}

export async function listChatMessages(
  scope: string,
  accessToken: string,
  limit: number = 50,
  cursor?: string
): Promise<ListChatMessagesResponse> {
  const params = new URLSearchParams({
    scope,
    limit: limit.toString(),
  })
  if (cursor) params.append('cursor', cursor)

  return api(`/v1/chat?${params}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

export async function uploadFile(
  file: File,
  accessToken: string
): Promise<{ url: string; filename: string; size: number }> {
  const formData = new FormData()
  formData.append('file', file)

  const base = process.env.NEXT_PUBLIC_API_BASE || 'http://localhost:8080'
  const res = await fetch(`${base}/v1/uploads`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
    body: formData,
  })

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: 'Upload failed' }))
    throw new Error(error.message || `HTTP ${res.status}`)
  }

  return res.json()
}

// Audit API
export interface AuditLog {
  seq: number
  timestamp: string
  tenant_id: string
  action: string
  actor: string
  object: string
  details: any
  request_id: string
}

export interface ListAuditLogsResponse {
  data: AuditLog[]
  pagination: {
    page: number
    limit: number
    total: number
  }
}

export interface ListNetworksResponse {
  data: Network[]
  meta: {
    limit: number
    next_cursor: string
    has_more: boolean
  }
}

/**
 * List all networks (admin only)
 * @param limit - Page limit
 * @param cursor - Pagination cursor
 * @param accessToken - JWT access token
 */
export async function listAllNetworks(
  limit: number,
  cursor: string,
  accessToken: string,
  searchQuery?: string
): Promise<ListNetworksResponse> {
  const params: Record<string, string> = {
    limit: limit.toString(),
    cursor: cursor || '',
  }
  if (searchQuery) {
    params.q = searchQuery
  }
  const query = new URLSearchParams(params)
  return api(`/v1/admin/networks?${query}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * List all devices (admin only)
 * @param limit - Page limit
 * @param cursor - Pagination cursor
 * @param accessToken - JWT access token
 */
export async function listAllDevices(
  limit: number,
  cursor: string,
  accessToken: string,
  searchQuery?: string
): Promise<DeviceListResponse> {
  const params: Record<string, string> = {
    limit: limit.toString(),
    cursor: cursor || '',
  }
  if (searchQuery) {
    params.q = searchQuery
  }
  const query = new URLSearchParams(params)
  const res = await api(`/v1/admin/devices?${query}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })

  // Map backend response to DeviceListResponse
  return {
    devices: res.data,
    next_cursor: res.meta.next_cursor,
    has_more: res.meta.has_more
  }
}

export async function listAuditLogs(
  page: number,
  limit: number,
  accessToken: string
): Promise<ListAuditLogsResponse> {
  return api(`/v1/audit/logs?page=${page}&limit=${limit}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

// Admin API
export interface SystemStats {
  total_users: number
  total_tenants: number
  total_networks: number
  total_devices: number
  active_connections: number
  messages_today: number
}

export interface AdminUser {
  id: string
  tenant_id: string
  email: string
  locale: string
  is_admin: boolean
  is_moderator: boolean
  auth_provider?: string
  created_at: string
  updated_at: string
}

export interface Tenant {
  id: string
  name: string
  created_at: string
  updated_at: string
}

export async function getSystemStats(accessToken: string): Promise<{ data: SystemStats }> {
  return api('/v1/admin/stats', {
    headers: { Authorization: `Bearer ${accessToken}` }
  })
}

export async function listUsers(limit: number, offset: number, accessToken: string, query?: string): Promise<{ data: AdminUser[], meta: any }> {
  let url = `/v1/admin/users?limit=${limit}&offset=${offset}`
  if (query) {
    url += `&q=${encodeURIComponent(query)}`
  }
  return api(url, {
    headers: { Authorization: `Bearer ${accessToken}` }
  })
}

export async function listTenants(limit: number, offset: number, accessToken: string, query?: string): Promise<{ data: Tenant[], meta: any }> {
  let url = `/v1/admin/tenants?limit=${limit}&offset=${offset}`
  if (query) {
    url += `&q=${encodeURIComponent(query)}`
  }
  return api(url, {
    headers: { Authorization: `Bearer ${accessToken}` }
  })
}

export async function toggleUserAdmin(
  userId: string,
  accessToken: string
): Promise<{ data: AdminUser }> {
  return api(`/v1/admin/users/${userId}/toggle-admin`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Delete user (admin only)
 * @param userId - User ID
 * @param accessToken - JWT access token
 */
export async function deleteUser(
  userId: string,
  accessToken: string
): Promise<void> {
  await api(`/v1/admin/users/${userId}`, {
    method: 'DELETE',
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
}

/**
 * Delete tenant (admin only)
 * @param tenantId - Tenant ID
 * @param accessToken - JWT access token
 */
export async function deleteTenant(tenantId: string, accessToken: string): Promise<void> {
  return api(`/v1/admin/tenants/${tenantId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` }
  })
}

// ==================== TENANT MULTI-MEMBERSHIP API ====================

// Types
export type TenantRole = 'owner' | 'admin' | 'moderator' | 'vip' | 'member'
export type TenantVisibility = 'public' | 'unlisted' | 'private'
export type TenantAccessType = 'open' | 'password' | 'invite_only'

export interface TenantExtended {
  id: string
  name: string
  description: string
  icon_url?: string
  visibility: TenantVisibility
  access_type: TenantAccessType
  max_members: number
  owner_id: string
  member_count: number
  created_at: string
  updated_at: string
}

export interface TenantMember {
  id: string
  tenant_id: string
  user_id: string
  user_name?: string
  role: TenantRole
  nickname?: string
  created_at: string
  joined_at: string
  updated_at: string
  banned_at?: string | null
  banned_by?: string | null
  user?: {
    id: string
    email: string
    name?: string
    locale: string
  }
}

export interface TenantInvite {
  id: string
  tenant_id: string
  code: string
  max_uses: number
  use_count: number
  expires_at?: string
  created_by: string
  created_at: string
  revoked_at?: string
}

export interface TenantAnnouncement {
  id: string
  tenant_id: string
  title: string
  content: string
  priority: 'low' | 'normal' | 'high' | 'urgent'
  author_id: string
  is_pinned: boolean
  created_at: string
  updated_at: string
  author?: {
    id: string
    email: string
    name?: string
    locale: string
  }
}

export interface TenantChatMessage {
  id: string
  tenant_id: string
  user_id: string
  user_name?: string
  content: string
  is_deleted?: boolean
  is_edited?: boolean
  created_at: string
  updated_at?: string
  edited_at?: string
  user?: {
    id: string
    email: string
    name?: string
    locale: string
  }
}

// Request types
export interface CreateTenantRequest {
  name: string
  description?: string
  visibility: TenantVisibility
  access_type: TenantAccessType
  password?: string
  max_members?: number
}

export interface JoinTenantRequest {
  password?: string
}

export interface JoinByCodeRequest {
  code: string
}

export interface UpdateMemberRoleRequest {
  role: TenantRole
}

export interface CreateTenantInviteRequest {
  max_uses?: number
  expires_in?: number // seconds
}

export interface CreateAnnouncementRequest {
  title: string
  content: string
  is_pinned?: boolean
}

export interface UpdateAnnouncementRequest {
  title?: string
  content?: string
  is_pinned?: boolean
}

export interface SendChatMessageRequest {
  content: string
}

// ==================== TENANT OPERATIONS ====================

/**
 * Create a new tenant (user becomes owner)
 */
export async function createTenant(accessToken: string, req: CreateTenantRequest): Promise<{ data: TenantExtended }> {
  return api('/v1/tenants', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(req),
  })
}

/**
 * Get tenant by ID
 */
export async function getTenantById(accessToken: string, tenantId: string): Promise<{ data: TenantExtended }> {
  return api(`/v1/tenants/${tenantId}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

// Request type for updating tenant
export interface UpdateTenantRequest {
  name?: string
  description?: string
  visibility?: TenantVisibility
  access_type?: TenantAccessType
  password?: string // Required if changing to password access_type
  max_members?: number
}

/**
 * Update tenant settings (owner/admin only)
 */
export async function updateTenantWithToken(
  accessToken: string,
  tenantId: string,
  req: UpdateTenantRequest
): Promise<{ data: TenantExtended }> {
  return api(`/v1/tenants/${tenantId}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(req),
  })
}

/**
 * Update tenant settings (uses stored token) - convenience wrapper
 */
export async function updateTenant(
  tenantId: string,
  req: UpdateTenantRequest
): Promise<TenantExtended> {
  const token = getAccessToken()
  if (!token) {
    throw new Error('No access token available')
  }
  const response = await updateTenantWithToken(token, tenantId, req)
  return response.data
}

/**
 * Delete tenant (owner only) - permanently deletes tenant and all associated data
 * Uses stored token
 */
export async function deleteTenantAsOwner(tenantId: string): Promise<{ message: string }> {
  const token = getAccessToken()
  if (!token) {
    throw new Error('No access token available')
  }
  return api(`/v1/tenants/${tenantId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${token}` },
  })
}

// ==================== MEMBERSHIP OPERATIONS (WITH TOKEN) ====================

/**
 * Join a tenant (for open or password-protected tenants) - explicit token version
 */
export async function joinTenantWithToken(accessToken: string, tenantId: string, req?: JoinTenantRequest): Promise<{ data: TenantMember; message: string }> {
  return api(`/v1/tenants/${tenantId}/join`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: req ? JSON.stringify(req) : undefined,
  })
}

/**
 * Join a tenant using an invite code (Steam-like)
 */
export async function joinTenantByCode(accessToken: string, code: string): Promise<{ data: TenantMember; message: string }> {
  return api('/v1/tenants/join-by-code', {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify({ code }),
  })
}

/**
 * Leave a tenant (owner cannot leave) - explicit token version
 */
export async function leaveTenantWithToken(accessToken: string, tenantId: string): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/leave`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Get all tenants the current user is a member of
 */
export async function getUserTenants(accessToken: string): Promise<{ data: TenantMember[] }> {
  return api('/v1/users/me/tenants', {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Get members of a tenant - explicit token version
 */
export async function getTenantMembersWithToken(
  accessToken: string,
  tenantId: string,
  options?: { role?: string; limit?: number; cursor?: string }
): Promise<{ data: TenantMember[]; next_cursor?: string }> {
  const params = new URLSearchParams()
  if (options?.role) params.set('role', options.role)
  if (options?.limit) params.set('limit', String(options.limit))
  if (options?.cursor) params.set('cursor', options.cursor)
  const query = params.toString() ? `?${params.toString()}` : ''
  return api(`/v1/tenants/${tenantId}/members${query}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Update a member's role in a tenant (admin+ only) - explicit token version
 */
export async function updateMemberRoleWithToken(
  accessToken: string,
  tenantId: string,
  memberId: string,
  role: TenantRole
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/members/${memberId}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify({ role }),
  })
}

/**
 * Remove/kick a member from a tenant (moderator+ only) - explicit token version
 */
export async function removeTenantMemberWithToken(
  accessToken: string,
  tenantId: string,
  memberId: string
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/members/${memberId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

// ==================== INVITE OPERATIONS (WITH TOKEN) ====================

/**
 * Create a tenant invite code (admin+ only) - explicit token version
 */
export async function createTenantInviteWithToken(
  accessToken: string,
  tenantId: string,
  req?: CreateTenantInviteRequest
): Promise<{ data: TenantInvite }> {
  return api(`/v1/tenants/${tenantId}/invites`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: req ? JSON.stringify(req) : undefined,
  })
}

/**
 * List all invites for a tenant (admin+ only)
 */
export async function listTenantInvites(accessToken: string, tenantId: string): Promise<{ data: TenantInvite[] }> {
  return api(`/v1/tenants/${tenantId}/invites`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Revoke a tenant invite (admin+ only) - explicit token version
 */
export async function revokeTenantInviteWithToken(
  accessToken: string,
  tenantId: string,
  inviteId: string
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/invites/${inviteId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

// ==================== ANNOUNCEMENT OPERATIONS ====================

/**
 * Create an announcement (moderator+ only)
 */
export async function createAnnouncement(
  accessToken: string,
  tenantId: string,
  req: CreateAnnouncementRequest
): Promise<{ data: TenantAnnouncement }> {
  return api(`/v1/tenants/${tenantId}/announcements`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(req),
  })
}

/**
 * List announcements for a tenant
 */
export async function listAnnouncements(
  accessToken: string,
  tenantId: string,
  options?: { pinned?: boolean; limit?: number; cursor?: string }
): Promise<{ data: TenantAnnouncement[]; next_cursor?: string }> {
  const params = new URLSearchParams()
  if (options?.pinned !== undefined) params.set('pinned', String(options.pinned))
  if (options?.limit) params.set('limit', String(options.limit))
  if (options?.cursor) params.set('cursor', options.cursor)
  const query = params.toString() ? `?${params.toString()}` : ''
  return api(`/v1/tenants/${tenantId}/announcements${query}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Update an announcement (moderator+ only)
 */
export async function updateAnnouncement(
  accessToken: string,
  tenantId: string,
  announcementId: string,
  req: UpdateAnnouncementRequest
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/announcements/${announcementId}`, {
    method: 'PATCH',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify(req),
  })
}

/**
 * Delete an announcement (moderator+ only)
 */
export async function deleteAnnouncement(
  accessToken: string,
  tenantId: string,
  announcementId: string
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/announcements/${announcementId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

// ==================== CHAT OPERATIONS (WITH TOKEN) ====================

/**
 * Send a chat message to tenant general chat - explicit token version
 */
export async function sendTenantChatMessageWithToken(
  accessToken: string,
  tenantId: string,
  content: string
): Promise<{ data: TenantChatMessage }> {
  return api(`/v1/tenants/${tenantId}/chat/messages`, {
    method: 'POST',
    headers: { Authorization: `Bearer ${accessToken}` },
    body: JSON.stringify({ content }),
  })
}

/**
 * Get chat history for a tenant
 */
export async function getTenantChatHistory(
  accessToken: string,
  tenantId: string,
  options?: { before?: string; limit?: number }
): Promise<{ data: TenantChatMessage[] }> {
  const params = new URLSearchParams()
  if (options?.before) params.set('before', options.before)
  if (options?.limit) params.set('limit', String(options.limit))
  const query = params.toString() ? `?${params.toString()}` : ''
  return api(`/v1/tenants/${tenantId}/chat/messages${query}`, {
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

/**
 * Delete a chat message (author or moderator+ only) - explicit token version
 */
export async function deleteTenantChatMessageWithToken(
  accessToken: string,
  tenantId: string,
  messageId: string
): Promise<{ message: string }> {
  return api(`/v1/tenants/${tenantId}/chat/messages/${messageId}`, {
    method: 'DELETE',
    headers: { Authorization: `Bearer ${accessToken}` },
  })
}

// ==================== SIMPLIFIED API (AUTO TOKEN) ====================
// These functions automatically get the access token from local storage

export interface TenantWithMemberCount {
  id: string
  name: string
  description?: string
  visibility: 'public' | 'private'
  access_type: 'open' | 'password' | 'invite_only'
  member_count?: number
  max_members?: number
  owner_id: string
  created_at: string
  updated_at: string
}

/**
 * Discover public tenants (auto token)
 */
export interface TenantListResult {
  data: TenantWithMemberCount[]
  next_cursor?: string
}

export async function discoverTenants(options?: {
  search?: string
  limit?: number
  cursor?: string
}): Promise<TenantListResult> {
  const params = new URLSearchParams()
  if (options?.limit) params.set('limit', String(options.limit))
  if (options?.cursor) params.set('cursor', options.cursor)

  let path = '/v1/tenants/public'
  const query = options?.search?.trim()
  if (query) {
    params.set('q', query)
    path = '/v1/tenants/search'
  }

  const queryString = params.toString() ? `?${params.toString()}` : ''
  const res = await api(`${path}${queryString}`)
  return {
    data: res.data || [],
    next_cursor: res.next_cursor || res.nextCursor || undefined,
  }
}

/**
 * Get current user's tenants (auto token)
 */
export async function getMyTenants(): Promise<TenantWithMemberCount[]> {
  const res = await api('/v1/users/me/tenants', {
    headers: getAuthHeader(),
  })
  return res.data || []
}

/**
 * Get tenant by ID (auto token)
 */
export async function getTenantByID(tenantId: string): Promise<TenantWithMemberCount> {
  const res = await api(`/v1/tenants/${tenantId}`, {
    headers: getAuthHeader(),
  })
  return res.data
}

/**
 * Join a tenant (auto token)
 */
export async function joinTenant(tenantId: string, password?: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/join`, {
    method: 'POST',
    headers: getAuthHeader(),
    body: password ? JSON.stringify({ password }) : undefined,
  })
}

/**
 * Use invite code to join tenant (auto token)
 */
export async function useTenantInvite(code: string): Promise<void> {
  await api('/v1/tenants/join-by-code', {
    method: 'POST',
    headers: getAuthHeader(),
    body: JSON.stringify({ code }),
  })
}

/**
 * Leave a tenant (auto token)
 */
export async function leaveTenant(tenantId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/leave`, {
    method: 'POST',
    headers: getAuthHeader(),
  })
}

/**
 * Get tenant members (auto token)
 */
export async function getTenantMembers(tenantId: string): Promise<TenantMember[]> {
  const res = await api(`/v1/tenants/${tenantId}/members`, {
    headers: getAuthHeader(),
  })
  return res.data || []
}

/**
 * Update member role (auto token)
 */
export async function updateMemberRole(tenantId: string, memberId: string, role: TenantRole): Promise<void> {
  await api(`/v1/tenants/${tenantId}/members/${memberId}`, {
    method: 'PATCH',
    headers: getAuthHeader(),
    body: JSON.stringify({ role }),
  })
}

/**
 * Remove tenant member (auto token)
 */
export async function removeTenantMember(tenantId: string, memberId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/members/${memberId}`, {
    method: 'DELETE',
    headers: getAuthHeader(),
  })
}

/**
 * Ban tenant member (auto token)
 * Banned users cannot rejoin the tenant even with invite codes
 */
export async function banTenantMember(tenantId: string, memberId: string, reason?: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/members/${memberId}/ban`, {
    method: 'POST',
    headers: getAuthHeader(),
    body: reason ? JSON.stringify({ reason }) : undefined,
  })
}

/**
 * Unban tenant member (auto token)
 * Removes ban status, allowing the user to rejoin the tenant
 */
export async function unbanTenantMember(tenantId: string, memberId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/members/${memberId}/ban`, {
    method: 'DELETE',
    headers: getAuthHeader(),
  })
}

/**
 * Get banned tenant members (auto token)
 * Returns list of all banned members in the tenant
 */
export async function getBannedTenantMembers(tenantId: string): Promise<TenantMember[]> {
  const res = await api(`/v1/tenants/${tenantId}/members/banned`, {
    headers: getAuthHeader(),
  })
  return res.members || []
}

/**
 * Get tenant invites (auto token)
 */
export async function getTenantInvites(tenantId: string): Promise<TenantInvite[]> {
  const res = await api(`/v1/tenants/${tenantId}/invites`, {
    headers: getAuthHeader(),
  })
  return res.data || []
}

/**
 * Create tenant invite (auto token)
 */
export async function createTenantInvite(tenantId: string, options?: {
  expiresInHours?: number
  maxUses?: number
}): Promise<TenantInvite> {
  const res = await api(`/v1/tenants/${tenantId}/invites`, {
    method: 'POST',
    headers: getAuthHeader(),
    body: JSON.stringify({
      expires_in: options?.expiresInHours ? options.expiresInHours * 3600 : undefined,
      max_uses: options?.maxUses,
    }),
  })
  return res.data
}

/**
 * Revoke tenant invite (auto token)
 */
export async function revokeTenantInvite(tenantId: string, inviteId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/invites/${inviteId}`, {
    method: 'DELETE',
    headers: getAuthHeader(),
  })
}

/**
 * Get tenant announcements (auto token)
 */
export async function getTenantAnnouncements(tenantId: string): Promise<TenantAnnouncement[]> {
  const res = await api(`/v1/tenants/${tenantId}/announcements`, {
    headers: getAuthHeader(),
  })
  return res.data || []
}

/**
 * Create tenant announcement (auto token)
 */
export async function createTenantAnnouncement(tenantId: string, data: {
  title: string
  content: string
  priority?: 'low' | 'normal' | 'high' | 'urgent'
}): Promise<TenantAnnouncement> {
  const res = await api(`/v1/tenants/${tenantId}/announcements`, {
    method: 'POST',
    headers: getAuthHeader(),
    body: JSON.stringify(data),
  })
  return res.data
}

/**
 * Delete tenant announcement (auto token)
 */
export async function deleteTenantAnnouncement(tenantId: string, announcementId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/announcements/${announcementId}`, {
    method: 'DELETE',
    headers: getAuthHeader(),
  })
}

/**
 * Toggle announcement pin (auto token)
 */
export async function toggleTenantAnnouncementPin(tenantId: string, announcementId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/announcements/${announcementId}/pin`, {
    method: 'POST',
    headers: getAuthHeader(),
  })
}

/**
 * Get tenant chat messages (auto token)
 */
export async function getTenantChat(tenantId: string, options?: {
  limit?: number
  before?: string
}): Promise<TenantChatMessage[]> {
  const params = new URLSearchParams()
  if (options?.limit) params.set('limit', String(options.limit))
  if (options?.before) params.set('before', options.before)
  const query = params.toString() ? `?${params.toString()}` : ''
  const res = await api(`/v1/tenants/${tenantId}/chat/messages${query}`, {
    headers: getAuthHeader(),
  })
  return res.data || []
}

/**
 * Send tenant chat message (auto token)
 */
export async function sendTenantChatMessage(tenantId: string, data: {
  content: string
}): Promise<TenantChatMessage> {
  const res = await api(`/v1/tenants/${tenantId}/chat/messages`, {
    method: 'POST',
    headers: getAuthHeader(),
    body: JSON.stringify(data),
  })
  return res.data
}

/**
 * Edit tenant chat message (auto token)
 */
export async function editTenantChatMessage(tenantId: string, messageId: string, data: {
  content: string
}): Promise<void> {
  await api(`/v1/tenants/${tenantId}/chat/messages/${messageId}`, {
    method: 'PATCH',
    headers: getAuthHeader(),
    body: JSON.stringify(data),
  })
}

/**
 * Delete tenant chat message (soft delete) (auto token)
 */
export async function deleteTenantChatMessageSimple(tenantId: string, messageId: string): Promise<void> {
  await api(`/v1/tenants/${tenantId}/chat/messages/${messageId}`, {
    method: 'DELETE',
    headers: getAuthHeader(),
  })
}
