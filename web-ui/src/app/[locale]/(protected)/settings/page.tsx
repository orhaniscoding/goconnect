'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import QRCode from 'qrcode'
import { getUser, getAccessToken, setUser } from '../../../../lib/auth'
import { generate2FA, enable2FA, disable2FA } from '../../../../lib/api'
import { useNotification } from '../../../../contexts/NotificationContext'
import AuthGuard from '../../../../components/AuthGuard'

export default function Settings() {
    const router = useRouter()
    const notification = useNotification()
    const [user, setUserState] = useState<any>(null)
    const [secret, setSecret] = useState('')
    const [otpAuthUrl, setOtpAuthUrl] = useState('')
    const [qrCodeUrl, setQrCodeUrl] = useState('')
    const [code, setCode] = useState('')
    const [isEnabling, setIsEnabling] = useState(false)
    const [step, setStep] = useState<'initial' | 'scan' | 'verify' | 'disable'>('initial')

    useEffect(() => {
        const u = getUser()
        setUserState(u)
    }, [])

    useEffect(() => {
        if (otpAuthUrl) {
            QRCode.toDataURL(otpAuthUrl)
                .then(url => setQrCodeUrl(url))
                .catch(err => console.error(err))
        }
    }, [otpAuthUrl])

    const handleStart2FA = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            const data = await generate2FA(token)
            setSecret(data.secret)
            setOtpAuthUrl(data.url)
            setStep('scan')
        } catch (err: any) {
            notification.error('Error', err.message || 'Failed to generate 2FA')
        }
    }

    const handleVerify2FA = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            await enable2FA(token, secret, code)

            // Update local user state
            const updatedUser = { ...user, two_fa_enabled: true }
            setUser(updatedUser)
            setUserState(updatedUser)

            notification.success('Success', 'Two-factor authentication enabled!')
            setStep('initial')
            setCode('')
        } catch (err: any) {
            notification.error('Error', err.message || 'Failed to verify code')
        }
    }

    const handleDisable2FA = () => {
        setStep('disable')
    }

    const confirmDisable2FA = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            await disable2FA(token, code)

            // Update local user state
            const updatedUser = { ...user, two_fa_enabled: false }
            setUser(updatedUser)
            setUserState(updatedUser)

            notification.success('Success', 'Two-factor authentication disabled.')
            setStep('initial')
            setCode('')
        } catch (err: any) {
            notification.error('Error', err.message || 'Failed to disable 2FA')
        }
    }

    if (!user) return null

    return (
        <AuthGuard>
            <div style={{ padding: 24, maxWidth: 800, margin: '0 auto' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 32 }}>
                    <h1 style={{ fontSize: 24, fontWeight: 'bold' }}>Settings</h1>
                    <button
                        onClick={() => router.back()}
                        style={{ padding: '8px 16px', cursor: 'pointer' }}
                    >
                        Back
                    </button>
                </div>

                <div style={{ backgroundColor: 'white', padding: 24, borderRadius: 8, boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
                    <h2 style={{ fontSize: 20, marginBottom: 16 }}>Profile</h2>
                    <div style={{ marginBottom: 8 }}><strong>Name:</strong> {user.name}</div>
                    <div style={{ marginBottom: 8 }}><strong>Email:</strong> {user.email}</div>
                    <div style={{ marginBottom: 8 }}><strong>Tenant ID:</strong> {user.tenant_id}</div>
                    <div style={{ marginBottom: 8 }}><strong>Auth Provider:</strong> {user.auth_provider || 'local'}</div>
                </div>

                <div style={{ marginTop: 24, backgroundColor: 'white', padding: 24, borderRadius: 8, boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
                    <h2 style={{ fontSize: 20, marginBottom: 16 }}>Security</h2>

                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <div>
                            <strong>Two-Factor Authentication (2FA)</strong>
                            <p style={{ color: '#666', marginTop: 4 }}>
                                {user.two_fa_enabled
                                    ? 'Your account is secured with 2FA.'
                                    : 'Add an extra layer of security to your account.'}
                            </p>
                        </div>

                        {!user.two_fa_enabled && step === 'initial' && (
                            <button
                                onClick={handleStart2FA}
                                style={{
                                    backgroundColor: '#007bff',
                                    color: 'white',
                                    border: 'none',
                                    padding: '10px 20px',
                                    borderRadius: 4,
                                    cursor: 'pointer'
                                }}
                            >
                                Enable 2FA
                            </button>
                        )}

                        {user.two_fa_enabled && (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                                <span style={{ color: 'green', fontWeight: 'bold' }}>Enabled</span>
                                <button
                                    onClick={handleDisable2FA}
                                    style={{
                                        backgroundColor: '#dc3545',
                                        color: 'white',
                                        border: 'none',
                                        padding: '8px 16px',
                                        borderRadius: 4,
                                        cursor: 'pointer',
                                        fontSize: 14
                                    }}
                                >
                                    Disable
                                </button>
                            </div>
                        )}
                    </div>

                    {step === 'disable' && (
                        <div style={{ marginTop: 24, borderTop: '1px solid #eee', paddingTop: 24 }}>
                            <h3>Disable Two-Factor Authentication</h3>
                            <p>Please enter the code from your authenticator app to confirm.</p>

                            <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
                                <input
                                    type="text"
                                    value={code}
                                    onChange={(e) => setCode(e.target.value)}
                                    placeholder="123456"
                                    style={{
                                        padding: 8,
                                        borderRadius: 4,
                                        border: '1px solid #ccc',
                                        width: 150
                                    }}
                                />
                                <button
                                    onClick={confirmDisable2FA}
                                    style={{
                                        backgroundColor: '#dc3545',
                                        color: 'white',
                                        border: 'none',
                                        padding: '8px 16px',
                                        borderRadius: 4,
                                        cursor: 'pointer'
                                    }}
                                >
                                    Confirm Disable
                                </button>
                                <button
                                    onClick={() => {
                                        setStep('initial')
                                        setCode('')
                                    }}
                                    style={{
                                        backgroundColor: '#6c757d',
                                        color: 'white',
                                        border: 'none',
                                        padding: '8px 16px',
                                        borderRadius: 4,
                                        cursor: 'pointer'
                                    }}
                                >
                                    Cancel
                                </button>
                            </div>
                        </div>
                    )}

                    {step === 'scan' && (
                        <div style={{ marginTop: 24, borderTop: '1px solid #eee', paddingTop: 24 }}>
                            <h3>1. Scan QR Code</h3>
                            <p>Open your authenticator app (Google Authenticator, Authy, etc.) and scan this code:</p>

                            {qrCodeUrl && (
                                <div style={{ margin: '16px 0' }}>
                                    <img src={qrCodeUrl} alt="2FA QR Code" style={{ width: 200, height: 200 }} />
                                </div>
                            )}

                            <p style={{ fontSize: 14, color: '#666' }}>
                                Or enter this secret manually: <strong>{secret}</strong>
                            </p>

                            <div style={{ marginTop: 24 }}>
                                <h3>2. Enter Code</h3>
                                <div style={{ display: 'flex', gap: 8 }}>
                                    <input
                                        type="text"
                                        value={code}
                                        onChange={(e) => setCode(e.target.value)}
                                        placeholder="123456"
                                        style={{
                                            padding: 8,
                                            borderRadius: 4,
                                            border: '1px solid #ccc',
                                            width: 150
                                        }}
                                    />
                                    <button
                                        onClick={handleVerify2FA}
                                        style={{
                                            backgroundColor: '#28a745',
                                            color: 'white',
                                            border: 'none',
                                            padding: '8px 16px',
                                            borderRadius: 4,
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Verify & Enable
                                    </button>
                                    <button
                                        onClick={() => setStep('initial')}
                                        style={{
                                            backgroundColor: '#6c757d',
                                            color: 'white',
                                            border: 'none',
                                            padding: '8px 16px',
                                            borderRadius: 4,
                                            cursor: 'pointer'
                                        }}
                                    >
                                        Cancel
                                    </button>
                                </div>
                            </div>
                        </div>
                    )}
                </div>
            </div>
        </AuthGuard>
    )
}
