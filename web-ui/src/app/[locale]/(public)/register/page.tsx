'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Footer from '../../../../components/Footer'
import { register } from '../../../../lib/api'
import { setTokens, setUser } from '../../../../lib/auth'
import { useT } from '../../../../lib/i18n-context'

interface RegisterPageProps {
    params: { locale: string }
}

export default function Register({ params }: RegisterPageProps) {
    const router = useRouter()
    const t = useT()
    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [error, setError] = useState('')
    const [success, setSuccess] = useState(false)
    const [isLoading, setIsLoading] = useState(false)

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault()
        setError('')
        setSuccess(false)

        // Validation
        if (!name || !email || !password || !confirmPassword) {
            setError(t.errorRequired)
            return
        }

        if (!email.includes('@')) {
            setError(t.errorEmailFormat)
            return
        }

        if (password.length < 8) {
            setError(t('register.error.passwordTooShort'))
            return
        }

        if (password !== confirmPassword) {
            setError(t('register.error.passwordMismatch'))
            return
        }

        setIsLoading(true)

        try {
            const response = await register({ name, email, password })

            // Store tokens and user
            setTokens(response.data.access_token, response.data.refresh_token)
            setUser(response.data.user)

            // Show success message
            setSuccess(true)

            // Redirect to dashboard after 1 second
            setTimeout(() => {
                router.push(`/${params.locale}/dashboard`)
            }, 1000)
        } catch (err: any) {
            console.error('Registration error:', err)
            setError(err.message || t('register.error.network'))
        } finally {
            setIsLoading(false)
        }
    }

    return (
        <div style={{
            minHeight: '100vh',
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
            backgroundColor: '#f5f5f5',
            padding: 24
        }}>
            <div style={{
                width: '100%',
                maxWidth: 400,
                backgroundColor: 'white',
                borderRadius: 8,
                padding: 32,
                boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
            }}>
                <h1 style={{
                    fontSize: 24,
                    fontWeight: 600,
                    marginBottom: 24,
                    textAlign: 'center'
                }}>
                    {t('register.title')}
                </h1>

                {error && (
                    <div style={{
                        backgroundColor: '#fee',
                        color: '#c33',
                        padding: 12,
                        borderRadius: 4,
                        marginBottom: 16,
                        fontSize: 14
                    }}>
                        {error}
                    </div>
                )}

                {success && (
                    <div style={{
                        backgroundColor: '#efe',
                        color: '#2a2',
                        padding: 12,
                        borderRadius: 4,
                        marginBottom: 16,
                        fontSize: 14
                    }}>
                        {t('register.success')}
                    </div>
                )}

                <form onSubmit={handleSubmit}>
                    <div style={{ marginBottom: 16 }}>
                        <label style={{
                            display: 'block',
                            marginBottom: 8,
                            fontSize: 14,
                            fontWeight: 500
                        }}>
                            {t('register.name')}
                        </label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: 10,
                                border: '1px solid #ddd',
                                borderRadius: 4,
                                fontSize: 14
                            }}
                            autoComplete="name"
                            autoFocus
                        />
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{
                            display: 'block',
                            marginBottom: 8,
                            fontSize: 14,
                            fontWeight: 500
                        }}>
                            {t('register.email')}
                        </label>
                        <input
                            type="email"
                            value={email}
                            onChange={(e) => setEmail(e.target.value)}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: 10,
                                border: '1px solid #ddd',
                                borderRadius: 4,
                                fontSize: 14
                            }}
                            autoComplete="email"
                        />
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{
                            display: 'block',
                            marginBottom: 8,
                            fontSize: 14,
                            fontWeight: 500
                        }}>
                            {t('register.password')}
                        </label>
                        <input
                            type="password"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: 10,
                                border: '1px solid #ddd',
                                borderRadius: 4,
                                fontSize: 14
                            }}
                            autoComplete="new-password"
                        />
                    </div>

                    <div style={{ marginBottom: 24 }}>
                        <label style={{
                            display: 'block',
                            marginBottom: 8,
                            fontSize: 14,
                            fontWeight: 500
                        }}>
                            {t('register.confirmPassword')}
                        </label>
                        <input
                            type="password"
                            value={confirmPassword}
                            onChange={(e) => setConfirmPassword(e.target.value)}
                            disabled={isLoading}
                            style={{
                                width: '100%',
                                padding: 10,
                                border: '1px solid #ddd',
                                borderRadius: 4,
                                fontSize: 14
                            }}
                            autoComplete="new-password"
                        />
                    </div>

                    <button
                        type="submit"
                        disabled={isLoading || success}
                        style={{
                            width: '100%',
                            padding: 12,
                            backgroundColor: isLoading || success ? '#999' : '#28a745',
                            color: 'white',
                            border: 'none',
                            borderRadius: 4,
                            fontSize: 16,
                            fontWeight: 500,
                            cursor: isLoading || success ? 'not-allowed' : 'pointer'
                        }}
                    >
                        {isLoading ? t('register.submitting') : t('register.submit')}
                    </button>
                </form>

                <div style={{
                    marginTop: 24,
                    textAlign: 'center',
                    fontSize: 14,
                    color: '#666'
                }}>
                    {t('register.haveAccount')}{' '}
                    <a
                        href={`/${params.locale}/login`}
                        style={{ color: '#007bff', textDecoration: 'none' }}
                    >
                        {t('register.signIn')}
                    </a>
                </div>
            </div>

            <Footer />
        </div>
    )
}
