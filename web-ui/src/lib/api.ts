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

