'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { hasValidAccessToken, getUser } from '../lib/auth'

interface AuthGuardProps {
    children: React.ReactNode
    requireAdmin?: boolean
    requireModerator?: boolean
}

/**
 * AuthGuard component that protects routes from unauthorized access.
 * Checks for valid JWT token and optionally enforces role requirements.
 * 
 * @param children - React children to render if authenticated
 * @param requireAdmin - If true, only admin users can access
 * @param requireModerator - If true, only moderator or admin users can access
 */
export default function AuthGuard({ children, requireAdmin = false, requireModerator = false }: AuthGuardProps) {
    const router = useRouter()
    const [isAuthorized, setIsAuthorized] = useState(false)

    useEffect(() => {
        // Check authentication status
        if (!hasValidAccessToken()) {
            // No valid token, redirect to login
            router.replace('/en/login')
            return
        }

        const user = getUser()
        if (!user) {
            // Token exists but no user data, clear and redirect
            router.replace('/en/login')
            return
        }

        // Check role requirements
        if (requireAdmin && !user.is_admin) {
            // User is not admin but admin required
            router.replace('/en/dashboard')
            return
        }

        if (requireModerator && !user.is_moderator && !user.is_admin) {
            // User is not moderator/admin but moderator required
            router.replace('/en/dashboard')
            return
        }

        // All checks passed
        setIsAuthorized(true)
    }, [router, requireAdmin, requireModerator])

    // Show loading state while checking auth
    if (!isAuthorized) {
        return (
            <div style={{
                display: 'flex',
                justifyContent: 'center',
                alignItems: 'center',
                height: '100vh',
                fontFamily: 'system-ui, -apple-system, sans-serif',
                color: '#666'
            }}>
                <div style={{ textAlign: 'center' }}>
                    <div style={{
                        width: '40px',
                        height: '40px',
                        border: '4px solid #f3f3f3',
                        borderTop: '4px solid #007bff',
                        borderRadius: '50%',
                        animation: 'spin 1s linear infinite',
                        margin: '0 auto 16px'
                    }} />
                    <p>Checking authentication...</p>
                </div>
                <style jsx>{`
          @keyframes spin {
            0% { transform: rotate(0deg); }
            100% { transform: rotate(360deg); }
          }
        `}</style>
            </div>
        )
    }

    return <>{children}</>
}
