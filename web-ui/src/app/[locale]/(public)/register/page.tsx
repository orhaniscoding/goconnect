'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Footer from '../../../../components/Footer'
import { register } from '../../../../lib/api'
import { setTokens, setUser } from '../../../../lib/auth'

interface RegisterPageProps {
    params: { locale: string }
}

export default function Register({ params }: RegisterPageProps) {
    const router = useRouter()
    const [name, setName] = useState('')
    const [email, setEmail] = useState('')
    const [password, setPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [error, setError] = useState('')
    const [success, setSuccess] = useState(false)
    const [isLoading, setIsLoading] = useState(false)

    // TODO: Use i18n translations from getDictionary
    const t = {
        title: 'Create your GoConnect account',
        name: 'Full Name',
        email: 'Email',
        password: 'Password',
        confirmPassword: 'Confirm Password',
        submit: 'Create Account',
        submitting: 'Creating account...',
        haveAccount: 'Already have an account?',
        signIn: 'Sign in',
        errorRequired: 'This field is required',
        errorEmailFormat: 'Please enter a valid email address',
        errorPasswordTooShort: 'Password must be at least 8 characters',
        errorPasswordMismatch: 'Passwords do not match',
        errorNetwork: 'Network error. Please try again.',
        success: 'Account created successfully! Redirecting...',
    }

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
            setError(t.errorPasswordTooShort)
            return
        }

        if (password !== confirmPassword) {
            setError(t.errorPasswordMismatch)
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
            setError(err.message || t.errorNetwork)
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
                    {t.title}
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
                        {t.success}
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
                            {t.name}
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
                            {t.email}
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
                            {t.password}
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
                            {t.confirmPassword}
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
                        {isLoading ? t.submitting : t.submit}
                    </button>
                </form>

                <div style={{
                    marginTop: 24,
                    textAlign: 'center',
                    fontSize: 14,
                    color: '#666'
                }}>
                    {t.haveAccount}{' '}
                    <a
                        href={`/${params.locale}/login`}
                        style={{ color: '#007bff', textDecoration: 'none' }}
                    >
                        {t.signIn}
                    </a>
                </div>
            </div>

            <Footer />
        </div>
    )
}
