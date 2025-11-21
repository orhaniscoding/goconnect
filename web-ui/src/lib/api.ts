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
  accessToken: string
): Promise<ListNetworksResponse> {
  const query = new URLSearchParams({
    limit: limit.toString(),
    cursor: cursor || '',
  })
  return api(`/v1/admin/networks?${query}`, {
    headers: {
      Authorization: `Bearer ${accessToken}`,
    },
  })
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

export async function listUsers(limit: number, offset: number, accessToken: string): Promise<{ data: AdminUser[], meta: any }> {
  return api(`/v1/admin/users?limit=${limit}&offset=${offset}`, {
    headers: { Authorization: `Bearer ${accessToken}` }
  })
}

export async function listTenants(limit: number, offset: number, accessToken: string): Promise<{ data: Tenant[], meta: any }> {
  return api(`/v1/admin/tenants?limit=${limit}&offset=${offset}`, {
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


