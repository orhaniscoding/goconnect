'use client'

import { useEffect } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { setTokens, setUser } from '../../../../../lib/auth'
import { api } from '../../../../../lib/api'
import { useNotification } from '../../../../../contexts/NotificationContext'

interface CallbackPageProps {
    params: { locale: string }
}

export default function Callback({ params }: CallbackPageProps) {
    const router = useRouter()
    const searchParams = useSearchParams()
    const notification = useNotification()

    useEffect(() => {
        const accessToken = searchParams.get('access_token')
        const refreshToken = searchParams.get('refresh_token')

        if (accessToken && refreshToken) {
            setTokens(accessToken, refreshToken)
            
            // Fetch user info
            api('/v1/auth/me', {
                headers: {
                    'Authorization': `Bearer ${accessToken}`
                }
            })
            .then(response => {
                setUser(response.data)
                notification.success('Login Successful', 'Welcome back!')
                router.push(`/${params.locale}/dashboard`)
            })
            .catch(err => {
                console.error('Failed to fetch user:', err)
                notification.error('Login Failed', 'Failed to retrieve user info')
                router.push(`/${params.locale}/login`)
            })

        } else {
            notification.error('Login Failed', 'Missing tokens')
            router.push(`/${params.locale}/login`)
        }
    }, [searchParams, router, params.locale, notification])

    return (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
            Processing login...
        </div>
    )
}
