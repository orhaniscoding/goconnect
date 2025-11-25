'use client'

import { useState, useEffect } from 'react'
import { useRouter } from 'next/navigation'
import QRCode from 'qrcode'
import { getUser, getAccessToken, setUser } from '../../../../lib/auth'
import { generate2FA, enable2FA, disable2FA, generateRecoveryCodes, getRecoveryCodeCount } from '../../../../lib/api'
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

    // Recovery Codes state
    const [recoveryCodesCount, setRecoveryCodesCount] = useState<number | null>(null)
    const [recoveryCodes, setRecoveryCodes] = useState<string[]>([])
    const [showRecoveryCodes, setShowRecoveryCodes] = useState(false)
    const [recoveryCodeInput, setRecoveryCodeInput] = useState('')
    const [isGeneratingRecoveryCodes, setIsGeneratingRecoveryCodes] = useState(false)

    useEffect(() => {
        const u = getUser()
        setUserState(u)
    }, [])

    // Load recovery codes count when user has 2FA enabled
    useEffect(() => {
        const loadRecoveryCodesCount = async () => {
            const token = getAccessToken()
            if (!token || !user?.two_fa_enabled) return

            try {
                const data = await getRecoveryCodeCount(token)
                setRecoveryCodesCount(data.remaining_codes)
            } catch (err) {
                console.error('Failed to load recovery codes count:', err)
            }
        }

        loadRecoveryCodesCount()
    }, [user?.two_fa_enabled])

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

    const handleGenerateRecoveryCodes = async () => {
        if (!recoveryCodeInput || recoveryCodeInput.length !== 6) {
            notification.error(t('settings.notifications.error'), t('settings.recovery.enterCode'))
            return
        }

        setIsGeneratingRecoveryCodes(true)
        try {
            const token = getAccessToken()
            if (!token) return

            const data = await generateRecoveryCodes(token, recoveryCodeInput)
            setRecoveryCodes(data.codes)
            setShowRecoveryCodes(true)
            setRecoveryCodesCount(data.codes.length)
            setRecoveryCodeInput('')
            notification.success(t('settings.notifications.success'), t('settings.recovery.generated'))
        } catch (err: any) {
            notification.error(t('settings.notifications.error'), err.message || t('settings.recovery.generateFailed'))
        } finally {
            setIsGeneratingRecoveryCodes(false)
        }
    }

    const handleCopyRecoveryCodes = () => {
        const codesText = recoveryCodes.join('\n')
        navigator.clipboard.writeText(codesText)
        notification.success(t('settings.notifications.success'), t('settings.recovery.copied'))
    }

    const handleDownloadRecoveryCodes = () => {
        const codesText = `GoConnect Recovery Codes\n${'='.repeat(30)}\n\nGenerated: ${new Date().toISOString()}\n\n${recoveryCodes.map((c, i) => `${i + 1}. ${c}`).join('\n')}\n\n${'='.repeat(30)}\nKeep these codes safe!\nEach code can only be used once.`
        const blob = new Blob([codesText], { type: 'text/plain' })
        const url = URL.createObjectURL(blob)
        const a = document.createElement('a')
        a.href = url
        a.download = 'goconnect-recovery-codes.txt'
        a.click()
        URL.revokeObjectURL(url)
        notification.success(t('settings.notifications.success'), t('settings.recovery.downloaded'))
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

                    {/* Recovery Codes Section */}
                    {user.two_fa_enabled && (
                        <div style={{ marginTop: 24, borderTop: '1px solid #eee', paddingTop: 24 }}>
                            <h3 style={{ marginBottom: 8 }}>{t('settings.recovery.title')}</h3>
                            <p style={{ color: '#666', marginBottom: 16 }}>{t('settings.recovery.desc')}</p>

                            {/* Recovery codes status */}
                            {recoveryCodesCount !== null && !showRecoveryCodes && (
                                <div style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'space-between',
                                    backgroundColor: recoveryCodesCount <= 2 ? '#fff3cd' : '#e7f5e7',
                                    padding: 16,
                                    borderRadius: 8,
                                    marginBottom: 16
                                }}>
                                    <div>
                                        <strong>{t('settings.recovery.remaining')}: {recoveryCodesCount}/8</strong>
                                        {recoveryCodesCount <= 2 && (
                                            <p style={{ color: '#856404', marginTop: 4, fontSize: 14 }}>
                                                {t('settings.recovery.lowWarning')}
                                            </p>
                                        )}
                                    </div>
                                </div>
                            )}

                            {/* Generate new recovery codes */}
                            {!showRecoveryCodes && (
                                <div style={{ marginTop: 16 }}>
                                    <p style={{ marginBottom: 8, fontWeight: 500 }}>{t('settings.recovery.generateNew')}</p>
                                    <p style={{ color: '#666', fontSize: 14, marginBottom: 12 }}>
                                        {t('settings.recovery.generateWarning')}
                                    </p>
                                    <div style={{ display: 'flex', gap: 8 }}>
                                        <input
                                            type="text"
                                            value={recoveryCodeInput}
                                            onChange={(e) => setRecoveryCodeInput(e.target.value)}
                                            placeholder="123456"
                                            maxLength={6}
                                            style={{
                                                padding: 8,
                                                borderRadius: 4,
                                                border: '1px solid #ccc',
                                                width: 120
                                            }}
                                        />
                                        <button
                                            onClick={handleGenerateRecoveryCodes}
                                            disabled={isGeneratingRecoveryCodes}
                                            style={{
                                                backgroundColor: '#17a2b8',
                                                color: 'white',
                                                border: 'none',
                                                padding: '8px 16px',
                                                borderRadius: 4,
                                                cursor: isGeneratingRecoveryCodes ? 'not-allowed' : 'pointer',
                                                opacity: isGeneratingRecoveryCodes ? 0.7 : 1
                                            }}
                                        >
                                            {isGeneratingRecoveryCodes ? t('settings.recovery.generating') : t('settings.recovery.generate')}
                                        </button>
                                    </div>
                                </div>
                            )}

                            {/* Display recovery codes */}
                            {showRecoveryCodes && recoveryCodes.length > 0 && (
                                <div style={{
                                    backgroundColor: '#f8f9fa',
                                    padding: 20,
                                    borderRadius: 8,
                                    border: '1px solid #dee2e6'
                                }}>
                                    <div style={{
                                        display: 'flex',
                                        alignItems: 'center',
                                        justifyContent: 'space-between',
                                        marginBottom: 16
                                    }}>
                                        <strong style={{ color: '#dc3545' }}>⚠️ {t('settings.recovery.saveNow')}</strong>
                                        <div style={{ display: 'flex', gap: 8 }}>
                                            <button
                                                onClick={handleCopyRecoveryCodes}
                                                style={{
                                                    backgroundColor: '#6c757d',
                                                    color: 'white',
                                                    border: 'none',
                                                    padding: '6px 12px',
                                                    borderRadius: 4,
                                                    cursor: 'pointer',
                                                    fontSize: 14
                                                }}
                                            >
                                                📋 {t('settings.recovery.copy')}
                                            </button>
                                            <button
                                                onClick={handleDownloadRecoveryCodes}
                                                style={{
                                                    backgroundColor: '#28a745',
                                                    color: 'white',
                                                    border: 'none',
                                                    padding: '6px 12px',
                                                    borderRadius: 4,
                                                    cursor: 'pointer',
                                                    fontSize: 14
                                                }}
                                            >
                                                ⬇️ {t('settings.recovery.download')}
                                            </button>
                                        </div>
                                    </div>

                                    <div style={{
                                        display: 'grid',
                                        gridTemplateColumns: 'repeat(2, 1fr)',
                                        gap: 8,
                                        backgroundColor: 'white',
                                        padding: 16,
                                        borderRadius: 4,
                                        fontFamily: 'monospace',
                                        fontSize: 14
                                    }}>
                                        {recoveryCodes.map((code, idx) => (
                                            <div key={idx} style={{
                                                padding: 8,
                                                backgroundColor: '#f8f9fa',
                                                borderRadius: 4,
                                                textAlign: 'center'
                                            }}>
                                                {idx + 1}. {code}
                                            </div>
                                        ))}
                                    </div>

                                    <p style={{ marginTop: 12, color: '#666', fontSize: 13 }}>
                                        {t('settings.recovery.oneTimeUse')}
                                    </p>

                                    <button
                                        onClick={() => {
                                            setShowRecoveryCodes(false)
                                            setRecoveryCodes([])
                                        }}
                                        style={{
                                            marginTop: 16,
                                            backgroundColor: '#6c757d',
                                            color: 'white',
                                            border: 'none',
                                            padding: '8px 16px',
                                            borderRadius: 4,
                                            cursor: 'pointer'
                                        }}
                                    >
                                        {t('settings.recovery.close')}
                                    </button>
                                </div>
                            )}
                        </div>
                    )}
                </div>
            </div>
        </AuthGuard>
    )
}
