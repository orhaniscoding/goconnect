'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Footer from '../../../../components/Footer'
import { login } from '../../../../lib/api'
import { setTokens, setUser } from '../../../../lib/auth'
import { useNotification } from '../../../../contexts/NotificationContext'

interface LoginPageProps {
    params: { locale: string }
}

export default function Login({ params }: LoginPageProps) {
    const router = useRouter()
    const notification = useNotification()
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [code, setCode] = useState('')
    const [show2FA, setShow2FA] = useState(false)
    const [error, setError] = useState('')
    const [isLoading, setIsLoading] = useState(false)

    // TODO: Use i18n translations from getDictionary
    const t = {
        title: 'Sign in to GoConnect',
        email: 'Email',
        password: 'Password',
        submit: 'Sign In',
        submitting: 'Signing in...',
        noAccount: "Don't have an account?",
        signUp: 'Sign up',
        errorInvalid: 'Invalid email or password',
        errorNetwork: 'Network error. Please try again.',
        errorRequired: 'This field is required',
        errorEmailFormat: 'Please enter a valid email address',
    }

    const handleSubmit = async (e: FormEvent) => {
        e.preventDefault()
        setError('')

        // Basic validation
        if (!email || !password) {
            const msg = t.errorRequired
            setError(msg)
            notification.warning('Validation Error', msg)
            return
        }

        if (!email.includes('@')) {
            const msg = t.errorEmailFormat
            setError(msg)
            notification.warning('Validation Error', msg)
            return
        }

        setIsLoading(true)

        try {
            const response = await login({ email, password, code: show2FA ? code : undefined })

            // Store tokens and user
            setTokens(response.data.access_token, response.data.refresh_token)
            setUser(response.data.user)

            notification.success('Welcome back!', 'Login successful')

            // Redirect to dashboard
            router.push(`/${params.locale}/dashboard`)
        } catch (err: any) {
            console.error('Login error:', err)
            if (err.message.includes('ERR_2FA_REQUIRED') || err.message.includes('Two-factor authentication required')) {
                setShow2FA(true)
                setError('')
                notification.info('2FA Required', 'Please enter your authentication code')
            } else {
                const msg = err.message || t.errorNetwork
                setError(msg)
                notification.error('Login Failed', msg)
            }
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
                    textAlign: 'center',
                    color: '#111827'
                }}>
                    {t.title}
                </h1>

                <form onSubmit={handleSubmit}>
                    {!show2FA ? (
                        <>
                            <div style={{ marginBottom: 16 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                                    {t.email}
                                </label>
                                <input
                                    type="email"
                                    value={email}
                                    onChange={(e) => setEmail(e.target.value)}
                                    placeholder="name@company.com"
                                    style={{
                                        width: '100%',
                                        padding: '10px 12px',
                                        borderRadius: 6,
                                        border: '1px solid #d1d5db',
                                        fontSize: 16
                                    }}
                                />
                            </div>

                            <div style={{ marginBottom: 24 }}>
                                <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                                    {t.password}
                                </label>
                                <input
                                    type="password"
                                    value={password}
                                    onChange={(e) => setPassword(e.target.value)}
                                    placeholder="••••••••"
                                    style={{
                                        width: '100%',
                                        padding: '10px 12px',
                                        borderRadius: 6,
                                        border: '1px solid #d1d5db',
                                        fontSize: 16
                                    }}
                                />
                            </div>
                        </>
                    ) : (
                        <div style={{ marginBottom: 24 }}>
                            <label style={{ display: 'block', marginBottom: 8, fontWeight: 500 }}>
                                Authentication Code
                            </label>
                            <input
                                type="text"
                                value={code}
                                onChange={(e) => setCode(e.target.value)}
                                placeholder="123456"
                                maxLength={6}
                                autoFocus
                                style={{
                                    width: '100%',
                                    padding: '10px 12px',
                                    borderRadius: 6,
                                    border: '1px solid #d1d5db',
                                    fontSize: 16,
                                    textAlign: 'center',
                                    letterSpacing: '4px'
                                }}
                            />
                            <div style={{ marginTop: 8, textAlign: 'center' }}>
                                <button
                                    type="button"
                                    onClick={() => { setShow2FA(false); setCode(''); }}
                                    style={{
                                        background: 'none',
                                        border: 'none',
                                        color: '#2563eb',
                                        cursor: 'pointer',
                                        fontSize: 14
                                    }}
                                >
                                    Back to Login
                                </button>
                            </div>
                        </div>
                    )}

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

                    <button
                        type="submit"
                        disabled={isLoading}
                        style={{
                            width: '100%',
                            padding: 12,
                            backgroundColor: isLoading ? '#999' : '#007bff',
                            color: 'white',
                            border: 'none',
                            borderRadius: 4,
                            fontSize: 16,
                            fontWeight: 500,
                            cursor: isLoading ? 'not-allowed' : 'pointer'
                        }}
                    >
                        {isLoading ? t.submitting : t.submit}
                    </button>
                </form>

                <div style={{
                    marginTop: 24,
                    textAlign: 'center',
                    fontSize: 14,
                    color: '#666'
                }}>
                    {t.noAccount}{' '}
                    <a
                        href={`/${params.locale}/register`}
                        style={{ color: '#007bff', textDecoration: 'none' }}
                    >
                        {t.signUp}
                    </a>
                </div>
            </div>

            <Footer />
        </div>
    )
}

