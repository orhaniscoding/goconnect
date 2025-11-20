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

