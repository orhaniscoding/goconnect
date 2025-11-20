'use client'

import { useState, FormEvent } from 'react'
import { useRouter } from 'next/navigation'
import Footer from '../../../../components/Footer'
import { login } from '../../../../lib/api'
import { setTokens, setUser } from '../../../../lib/auth'

interface LoginPageProps {
  params: { locale: string }
}

export default function Login({ params }: LoginPageProps) {
  const router = useRouter()
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
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
      setError(t.errorRequired)
      return
    }

    if (!email.includes('@')) {
      setError(t.errorEmailFormat)
      return
    }

    setIsLoading(true)

    try {
      const response = await login({ email, password })
      
      // Store tokens and user
      setTokens(response.data.access_token, response.data.refresh_token)
      setUser(response.data.user)
      
      // Redirect to dashboard
      router.push(`/${params.locale}/dashboard`)
    } catch (err: any) {
      console.error('Login error:', err)
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

        <form onSubmit={handleSubmit}>
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
              autoFocus
            />
          </div>

          <div style={{ marginBottom: 24 }}>
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
              autoComplete="current-password"
            />
          </div>

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

