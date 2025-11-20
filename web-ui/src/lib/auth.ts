// Token storage and authentication utilities

const ACCESS_TOKEN_KEY = 'goconnect_access_token'
const REFRESH_TOKEN_KEY = 'goconnect_refresh_token'
const USER_KEY = 'goconnect_user'

export interface User {
  id: string
  name: string
  email: string
  is_admin: boolean
  is_moderator: boolean
  tenant_id: string
}

/**
 * Store authentication tokens in localStorage
 */
export function setTokens(accessToken: string, refreshToken: string): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(ACCESS_TOKEN_KEY, accessToken)
  localStorage.setItem(REFRESH_TOKEN_KEY, refreshToken)
}

/**
 * Get access token from localStorage
 */
export function getAccessToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(ACCESS_TOKEN_KEY)
}

/**
 * Get refresh token from localStorage
 */
export function getRefreshToken(): string | null {
  if (typeof window === 'undefined') return null
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

/**
 * Store user information in localStorage
 */
export function setUser(user: User): void {
  if (typeof window === 'undefined') return
  localStorage.setItem(USER_KEY, JSON.stringify(user))
}

/**
 * Get user information from localStorage
 */
export function getUser(): User | null {
  if (typeof window === 'undefined') return null
  const userJson = localStorage.getItem(USER_KEY)
  if (!userJson) return null
  try {
    return JSON.parse(userJson)
  } catch {
    return null
  }
}

/**
 * Clear all authentication data (logout)
 */
export function clearAuth(): void {
  if (typeof window === 'undefined') return
  localStorage.removeItem(ACCESS_TOKEN_KEY)
  localStorage.removeItem(REFRESH_TOKEN_KEY)
  localStorage.removeItem(USER_KEY)
}

/**
 * Check if user is authenticated (has valid access token)
 */
export function isAuthenticated(): boolean {
  return getAccessToken() !== null
}

/**
 * Parse JWT token to check expiration (basic validation)
 * Note: This does NOT verify signature, only checks expiration
 */
export function isTokenExpired(token: string): boolean {
  try {
    const payload = JSON.parse(atob(token.split('.')[1]))
    const exp = payload.exp * 1000 // Convert to milliseconds
    return Date.now() >= exp
  } catch {
    return true // If parsing fails, consider expired
  }
}

/**
 * Check if access token is valid (exists and not expired)
 */
export function hasValidAccessToken(): boolean {
  const token = getAccessToken()
  if (!token) return false
  return !isTokenExpired(token)
}
