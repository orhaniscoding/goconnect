'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import QRCode from 'qrcode'
import { getUser, getAccessToken, setUser } from '../../../../lib/auth'
import { generate2FA, enable2FA, disable2FA } from '../../../../lib/api'
import { useNotification } from '../../../../contexts/NotificationContext'
import { useT } from '../../../../lib/i18n-context'
import AuthGuard from '../../../../components/AuthGuard'

export default function Settings() {
    const router = useRouter()
    const notification = useNotification()
    const t = useT()
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
            notification.error(t('settings.notifications.error'), err.message || t('settings.notifications.generateFailed'))
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

            notification.success(t('settings.notifications.success'), t('settings.notifications.enabled'))
            setStep('initial')
            setCode('')
        } catch (err: any) {
            notification.error(t('settings.notifications.error'), err.message || t('settings.notifications.verifyFailed'))
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

            notification.success(t('settings.notifications.success'), t('settings.notifications.disabled'))
            setStep('initial')
            setCode('')
        } catch (err: any) {
            notification.error(t('settings.notifications.error'), err.message || t('settings.notifications.disableFailed'))
        }
    }

    if (!user) return null

    return (
        <AuthGuard>
            <div style={{ padding: 24, maxWidth: 800, margin: '0 auto' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 32 }}>
                    <h1 style={{ fontSize: 24, fontWeight: 'bold' }}>{t('settings.title')}</h1>
                    <button
                        onClick={() => router.back()}
                        style={{ padding: '8px 16px', cursor: 'pointer' }}
                    >
                        {t('settings.back')}
                    </button>
                </div>

                <div style={{ backgroundColor: 'white', padding: 24, borderRadius: 8, boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
                    <h2 style={{ fontSize: 20, marginBottom: 16 }}>{t('settings.profile.title')}</h2>
                    <div style={{ marginBottom: 8 }}><strong>{t('settings.profile.name')}:</strong> {user.name}</div>
                    <div style={{ marginBottom: 8 }}><strong>{t('settings.profile.email')}:</strong> {user.email}</div>
                    <div style={{ marginBottom: 8 }}><strong>{t('settings.profile.tenant')}:</strong> {user.tenant_id}</div>
                    <div style={{ marginBottom: 8 }}><strong>{t('settings.profile.provider')}:</strong> {user.auth_provider || 'local'}</div>
                </div>

                <div style={{ marginTop: 24, backgroundColor: 'white', padding: 24, borderRadius: 8, boxShadow: '0 2px 4px rgba(0,0,0,0.1)' }}>
                    <h2 style={{ fontSize: 20, marginBottom: 16 }}>{t('settings.security.title')}</h2>

                    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between' }}>
                        <div>
                            <strong>{t('settings.2fa.title')}</strong>
                            <p style={{ color: '#666', marginTop: 4 }}>
                                {user.two_fa_enabled
                                    ? t('settings.2fa.enabled.desc')
                                    : t('settings.2fa.disabled.desc')}
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
                                {t('settings.2fa.enable')}
                            </button>
                        )}

                        {user.two_fa_enabled && (
                            <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                                <span style={{ color: 'green', fontWeight: 'bold' }}>{t('settings.2fa.enabled')}</span>
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
                                    {t('settings.2fa.disable')}
                                </button>
                            </div>
                        )}
                    </div>

                    {step === 'disable' && (
                        <div style={{ marginTop: 24, borderTop: '1px solid #eee', paddingTop: 24 }}>
                            <h3>{t('settings.2fa.disable.title')}</h3>
                            <p>{t('settings.2fa.disable.desc')}</p>

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
                                    {t('settings.2fa.confirm')}
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
                                    {t('settings.2fa.cancel')}
                                </button>
                            </div>
                        </div>
                    )}

                    {step === 'scan' && (
                        <div style={{ marginTop: 24, borderTop: '1px solid #eee', paddingTop: 24 }}>
                            <h3>{t('settings.2fa.scan.title')}</h3>
                            <p>{t('settings.2fa.scan.desc')}</p>

                            {qrCodeUrl && (
                                <div style={{ margin: '16px 0' }}>
                                    <img src={qrCodeUrl} alt="2FA QR Code" style={{ width: 200, height: 200 }} />
                                </div>
                            )}

                            <p style={{ fontSize: 14, color: '#666' }}>
                                {t('settings.2fa.manual')} <strong>{secret}</strong>
                            </p>

                            <div style={{ marginTop: 24 }}>
                                <h3>{t('settings.2fa.verify.title')}</h3>
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
                                        {t('settings.2fa.verify.button')}
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
                                        {t('settings.2fa.cancel')}
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
