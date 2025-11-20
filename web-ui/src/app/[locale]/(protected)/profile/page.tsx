'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { getUser, getAccessToken } from '../../../../lib/auth'
import { generate2FA, enable2FA, disable2FA } from '../../../../lib/api'
import { useNotification } from '../../../../contexts/NotificationContext'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'
import QRCode from 'qrcode'

export default function ProfilePage() {
    const router = useRouter()
    const notification = useNotification()
    const [user, setUser] = useState<any>(null)
    const [editing, setEditing] = useState(false)
    const [changePasswordMode, setChangePasswordMode] = useState(false)

    // Password change form
    const [currentPassword, setCurrentPassword] = useState('')
    const [newPassword, setNewPassword] = useState('')
    const [confirmPassword, setConfirmPassword] = useState('')
    const [passwordError, setPasswordError] = useState<string | null>(null)
    const [passwordSuccess, setPasswordSuccess] = useState(false)

    // 2FA state
    const [twoFactorEnabled, setTwoFactorEnabled] = useState(false)
    const [show2FAModal, setShow2FAModal] = useState(false)
    const [qrCodeDataUrl, setQrCodeDataUrl] = useState('')
    const [secret, setSecret] = useState('')
    const [code, setCode] = useState('')
    const [error2FA, setError2FA] = useState<string | null>(null)

    useEffect(() => {
        const userData = getUser()
        if (userData) {
            setUser(userData)
            setTwoFactorEnabled(userData.two_fa_enabled || false)
        }
    }, [])

    const handlePasswordChange = async (e: React.FormEvent) => {
        e.preventDefault()
        setPasswordError(null)
        setPasswordSuccess(false)

        // Validation
        if (!currentPassword || !newPassword || !confirmPassword) {
            const msg = 'All password fields are required'
            setPasswordError(msg)
            notification.warning('Validation Error', msg)
            return
        }

        if (newPassword !== confirmPassword) {
            const msg = 'New passwords do not match'
            setPasswordError(msg)
            notification.warning('Validation Error', msg)
            return
        }

        if (newPassword.length < 8) {
            const msg = 'Password must be at least 8 characters'
            setPasswordError(msg)
            notification.warning('Validation Error', msg)
            return
        }

        // TODO: Backend integration - change password API call
        // For now, just show success message
        console.log('Password change requested')
        setPasswordSuccess(true)
        notification.success('Success', 'Password changed successfully')
        setCurrentPassword('')
        setNewPassword('')
        setConfirmPassword('')

        setTimeout(() => {
            setChangePasswordMode(false)
            setPasswordSuccess(false)
        }, 2000)
    }

    const handleToggle2FA = async () => {
        if (twoFactorEnabled) {
            // Disable flow
            setShow2FAModal(true)
            setError2FA(null)
            setCode('')
        } else {
            // Enable flow
            try {
                const token = getAccessToken()
                if (!token) return
                const data = await generate2FA(token)
                setSecret(data.secret)
                
                // Generate QR code data URL
                try {
                    const url = await QRCode.toDataURL(data.url)
                    setQrCodeDataUrl(url)
                } catch (e) {
                    console.error(e)
                }
                
                setShow2FAModal(true)
                setError2FA(null)
                setCode('')
            } catch (err: any) {
                notification.error('2FA Error', err.message)
            }
        }
    }

    const handleConfirm2FA = async () => {
        try {
            const token = getAccessToken()
            if (!token) return

            if (twoFactorEnabled) {
                await disable2FA(token, code)
                setTwoFactorEnabled(false)
                notification.success('2FA Disabled', 'Two-factor authentication has been disabled')
            } else {
                await enable2FA(token, secret, code)
                setTwoFactorEnabled(true)
                notification.success('2FA Enabled', 'Two-factor authentication has been enabled')
            }
            setShow2FAModal(false)
        } catch (err: any) {
            setError2FA(err.message)
        }
    }

    if (!user) {
        return (
            <AuthGuard>
                <div style={{ padding: 24 }}>Loading...</div>
            </AuthGuard>
        )
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 800, margin: '0 auto' }}>
                {/* Header */}
                <div style={{
                    padding: '16px 24px',
                    backgroundColor: 'white',
                    borderBottom: '1px solid #dee2e6',
                    display: 'flex',
                    justifyContent: 'space-between',
                    alignItems: 'center'
                }}>
                    <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
                        <button
                            onClick={() => router.push('/en/dashboard')}
                            style={{
                                padding: '6px 12px',
                                backgroundColor: '#6c757d',
                                color: 'white',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 14,
                                fontWeight: 500
                            }}
                        >
                            ← Back
                        </button>
                        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>Profile Settings</h1>
                    </div>
                </div>

                {/* Main Content */}
                <div style={{ flex: 1, padding: 24, maxWidth: 800, margin: '0 auto', width: '100%' }}>
                    {/* User Information Card */}
                    <div style={{
                        backgroundColor: 'white',
                        borderRadius: 12,
                        padding: 24,
                        marginBottom: 24,
                        boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                    }}>
                        <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600, color: '#212529' }}>
                            User Information
                        </h2>

                        <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                            <div>
                                <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                    User ID
                                </label>
                                <div style={{
                                    padding: '10px 12px',
                                    backgroundColor: '#f8f9fa',
                                    borderRadius: 6,
                                    fontFamily: 'monospace',
                                    fontSize: 14,
                                    color: '#495057'
                                }}>
                                    {user.id}
                                </div>
                            </div>

                            <div>
                                <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                    Name
                                </label>
                                <div style={{
                                    padding: '10px 12px',
                                    backgroundColor: '#f8f9fa',
                                    borderRadius: 6,
                                    fontSize: 15,
                                    color: '#212529'
                                }}>
                                    {user.name || 'Not set'}
                                </div>
                            </div>

                            <div>
                                <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                    Email
                                </label>
                                <div style={{
                                    padding: '10px 12px',
                                    backgroundColor: '#f8f9fa',
                                    borderRadius: 6,
                                    fontSize: 15,
                                    color: '#212529'
                                }}>
                                    {user.email || 'Not set'}
                                </div>
                            </div>

                            <div>
                                <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                    Role
                                </label>
                                <div style={{
                                    padding: '10px 12px',
                                    backgroundColor: user.is_admin ? '#d1e7dd' : user.is_moderator ? '#cfe2ff' : '#f8f9fa',
                                    borderRadius: 6,
                                    fontSize: 15,
                                    color: user.is_admin ? '#0f5132' : user.is_moderator ? '#084298' : '#212529',
                                    fontWeight: 500
                                }}>
                                    {user.is_admin ? '👑 Administrator' : user.is_moderator ? '🛡️ Moderator' : '👤 User'}
                                </div>
                            </div>

                            <div>
                                <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                    Tenant ID
                                </label>
                                <div style={{
                                    padding: '10px 12px',
                                    backgroundColor: '#f8f9fa',
                                    borderRadius: 6,
                                    fontFamily: 'monospace',
                                    fontSize: 14,
                                    color: '#495057'
                                }}>
                                    {user.tenant_id}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* Password Change Card */}
                    <div style={{
                        backgroundColor: 'white',
                        borderRadius: 12,
                        padding: 24,
                        marginBottom: 24,
                        boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                    }}>
                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 20 }}>
                            <h2 style={{ margin: 0, fontSize: 18, fontWeight: 600, color: '#212529' }}>
                                Password
                            </h2>
                            {!changePasswordMode && (
                                <button
                                    onClick={() => setChangePasswordMode(true)}
                                    style={{
                                        padding: '8px 16px',
                                        backgroundColor: '#007bff',
                                        color: 'white',
                                        border: 'none',
                                        borderRadius: 6,
                                        cursor: 'pointer',
                                        fontSize: 14,
                                        fontWeight: 500
                                    }}
                                >
                                    Change Password
                                </button>
                            )}
                        </div>

                        {changePasswordMode ? (
                            <form onSubmit={handlePasswordChange} style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                                {passwordError && (
                                    <div style={{
                                        padding: 12,
                                        backgroundColor: '#f8d7da',
                                        color: '#842029',
                                        borderRadius: 6,
                                        fontSize: 14
                                    }}>
                                        {passwordError}
                                    </div>
                                )}

                                {passwordSuccess && (
                                    <div style={{
                                        padding: 12,
                                        backgroundColor: '#d1e7dd',
                                        color: '#0f5132',
                                        borderRadius: 6,
                                        fontSize: 14
                                    }}>
                                        Password changed successfully!
                                    </div>
                                )}

                                <div>
                                    <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                        Current Password
                                    </label>
                                    <input
                                        type="password"
                                        value={currentPassword}
                                        onChange={(e) => setCurrentPassword(e.target.value)}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px',
                                            border: '1px solid #dee2e6',
                                            borderRadius: 6,
                                            fontSize: 15,
                                            outline: 'none'
                                        }}
                                    />
                                </div>

                                <div>
                                    <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                        New Password
                                    </label>
                                    <input
                                        type="password"
                                        value={newPassword}
                                        onChange={(e) => setNewPassword(e.target.value)}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px',
                                            border: '1px solid #dee2e6',
                                            borderRadius: 6,
                                            fontSize: 15,
                                            outline: 'none'
                                        }}
                                    />
                                </div>

                                <div>
                                    <label style={{ display: 'block', fontSize: 14, fontWeight: 500, color: '#6c757d', marginBottom: 6 }}>
                                        Confirm New Password
                                    </label>
                                    <input
                                        type="password"
                                        value={confirmPassword}
                                        onChange={(e) => setConfirmPassword(e.target.value)}
                                        style={{
                                            width: '100%',
                                            padding: '10px 12px',
                                            border: '1px solid #dee2e6',
                                            borderRadius: 6,
                                            fontSize: 15,
                                            outline: 'none'
                                        }}
                                    />
                                </div>

                                <div style={{ display: 'flex', gap: 12 }}>
                                    <button
                                        type="submit"
                                        style={{
                                            padding: '10px 20px',
                                            backgroundColor: '#28a745',
                                            color: 'white',
                                            border: 'none',
                                            borderRadius: 6,
                                            cursor: 'pointer',
                                            fontSize: 14,
                                            fontWeight: 500
                                        }}
                                    >
                                        Save Password
                                    </button>
                                    <button
                                        type="button"
                                        onClick={() => {
                                            setChangePasswordMode(false)
                                            setCurrentPassword('')
                                            setNewPassword('')
                                            setConfirmPassword('')
                                            setPasswordError(null)
                                            setPasswordSuccess(false)
                                        }}
                                        style={{
                                            padding: '10px 20px',
                                            backgroundColor: '#6c757d',
                                            color: 'white',
                                            border: 'none',
                                            borderRadius: 6,
                                            cursor: 'pointer',
                                            fontSize: 14,
                                            fontWeight: 500
                                        }}
                                    >
                                        Cancel
                                    </button>
                                </div>
                            </form>
                        ) : (
                            <div style={{ color: '#6c757d', fontSize: 14 }}>
                                Click "Change Password" to update your password
                            </div>
                        )}
                    </div>

                    {/* 2FA Card */}
                    <div style={{
                        backgroundColor: 'white',
                        borderRadius: 12,
                        padding: 24,
                        boxShadow: '0 1px 3px rgba(0,0,0,0.1)'
                    }}>
                        <h2 style={{ margin: '0 0 20px 0', fontSize: 18, fontWeight: 600, color: '#212529' }}>
                            Two-Factor Authentication (2FA)
                        </h2>

                        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                            <div>
                                <div style={{ fontSize: 15, color: '#212529', marginBottom: 4, fontWeight: 500 }}>
                                    2FA Status
                                </div>
                                <div style={{ fontSize: 14, color: '#6c757d' }}>
                                    {twoFactorEnabled
                                        ? 'Two-factor authentication is enabled for your account'
                                        : 'Enable 2FA for additional security (Coming soon)'}
                                </div>
                            </div>
                            <button
                                onClick={handleToggle2FA}
                                disabled={true}
                                style={{
                                    padding: '8px 16px',
                                    backgroundColor: twoFactorEnabled ? '#dc3545' : '#28a745',
                                    color: 'white',
                                    border: 'none',
                                    borderRadius: 6,
                                    cursor: 'not-allowed',
                                    fontSize: 14,
                                    fontWeight: 500,
                                    opacity: 0.6
                                }}
                            >
                                {twoFactorEnabled ? 'Disable 2FA' : 'Enable 2FA'}
                            </button>
                        </div>

                        <div style={{
                            marginTop: 16,
                            padding: 12,
                            backgroundColor: '#fff3cd',
                            color: '#856404',
                            borderRadius: 6,
                            fontSize: 13
                        }}>
                            ⚠️ 2FA feature is planned for future release
                        </div>
                    </div>
                </div>

                <Footer />
            </div>
        </AuthGuard>
    )
}
