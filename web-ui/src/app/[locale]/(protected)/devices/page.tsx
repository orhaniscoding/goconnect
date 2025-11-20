'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import {
  listDevices,
  registerDevice,
  deleteDevice,
  disableDevice,
  enableDevice,
  downloadWireGuardConfig,
  listNetworks,
  Device,
  RegisterDeviceRequest,
  Network
} from '../../../../lib/api'
import { getAccessToken } from '../../../../lib/auth'
import { bridge } from '../../../../lib/bridge'
import { useNotification } from '../../../../contexts/NotificationContext'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'
import { generateQRCode } from '../../../../lib/qrcode'

type PlatformType = 'windows' | 'macos' | 'linux' | 'android' | 'ios'

export default function DevicesPage() {
  const router = useRouter()
  const notification = useNotification()
  const [devices, setDevices] = useState<Device[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showAddForm, setShowAddForm] = useState(false)
  const [filterPlatform, setFilterPlatform] = useState<string>('')

  // Local Daemon State
  const [localDaemonStatus, setLocalDaemonStatus] = useState<any>(null)
  const [registeringLocal, setRegisteringLocal] = useState(false)

  // Form state
  const [formData, setFormData] = useState<RegisterDeviceRequest>({
    name: '',
    platform: 'linux',
    pubkey: '',
    hostname: '',
    os_version: '',
    daemon_ver: ''
  })
  const [formError, setFormError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  // Config download state
  const [showConfigModal, setShowConfigModal] = useState(false)
  const [selectedDevice, setSelectedDevice] = useState<Device | null>(null)
  const [networks, setNetworks] = useState<Network[]>([])
  const [selectedNetworkId, setSelectedNetworkId] = useState<string>('')
  const [privateKey, setPrivateKey] = useState<string>('')
  const [generatedConfig, setGeneratedConfig] = useState<string>('')
  const [configLoading, setConfigLoading] = useState(false)
  const [configError, setConfigError] = useState<string | null>(null)
  const [qrCodeUrl, setQrCodeUrl] = useState<string>('')
  const [showQRCode, setShowQRCode] = useState(false)

  useEffect(() => {
    loadDevices()
  }, [filterPlatform])

  useEffect(() => {
    loadNetworks()
    checkLocalDaemon()
  }, [])

  const checkLocalDaemon = async () => {
    try {
      const status = await bridge('/status')
      setLocalDaemonStatus(status)
    } catch (e) {
      // Daemon not running, ignore
      console.log('Local daemon not detected')
    }
  }

  const handleRegisterLocalDevice = async () => {
    setRegisteringLocal(true)
    try {
      const token = getAccessToken()
      if (!token) {
        notification.error('Authentication Error', 'Please log in again')
        return
      }

      // Call local bridge to register
      await bridge('/register', {
        method: 'POST',
        body: JSON.stringify({ token })
      })

      notification.success('Device Registered', 'This device has been successfully registered!')
      
      // Refresh status and list
      await checkLocalDaemon()
      await loadDevices()
    } catch (err: any) {
      console.error('Registration failed:', err)
      notification.error('Registration Failed', 'Could not register this device. Is the daemon running?')
    } finally {
      setRegisteringLocal(false)
    }
  }

  const loadDevices = async () => {
    setLoading(true)
    setError(null)
    try {
      const token = getAccessToken()
      if (!token) {
        router.push('/en/login')
        return
      }

      const response = await listDevices(
        token,
        filterPlatform || undefined
      )
      setDevices(response.devices || [])
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load devices')
    } finally {
      setLoading(false)
    }
  }

  const loadNetworks = async () => {
    try {
      const token = getAccessToken()
      if (!token) return

      const response = await listNetworks('mine', token)
      setNetworks(response.data || [])
    } catch (err) {
      console.error('Failed to load networks:', err)
    }
  }

  const handleGetConfig = (device: Device) => {
    setSelectedDevice(device)
    setShowConfigModal(true)
    setPrivateKey('')
    setGeneratedConfig('')
    setConfigError(null)
    setQrCodeUrl('')
    setShowQRCode(false)
  }

  const handleCloseConfigModal = () => {
    setShowConfigModal(false)
    setSelectedDevice(null)
    setSelectedNetworkId('')
    setPrivateKey('')
    setGeneratedConfig('')
    setConfigError(null)
    setQrCodeUrl('')
    setShowQRCode(false)
  }

  const handleGenerateConfig = async () => {
    if (!selectedDevice || !selectedNetworkId || !privateKey) {
      setConfigError('Please fill in all fields')
      return
    }

    // Validate private key (44 character base64)
    if (privateKey.length !== 44 || !/^[A-Za-z0-9+/]+=*$/.test(privateKey)) {
      setConfigError('Invalid private key format. Must be 44 character base64 string.')
      return
    }

    setConfigLoading(true)
    setConfigError(null)
    try {
      const token = getAccessToken()
      if (!token) return

      const config = await downloadWireGuardConfig(
        selectedNetworkId,
        selectedDevice.id,
        privateKey,
        token
      )
      setGeneratedConfig(config)

      // Generate QR code
      const qrUrl = await generateQRCode(config, 300)
      setQrCodeUrl(qrUrl)
    } catch (err) {
      setConfigError(err instanceof Error ? err.message : 'Failed to generate config')
    } finally {
      setConfigLoading(false)
    }
  }

  const handleToggleQRCode = () => {
    setShowQRCode(!showQRCode)
  }

  const handleCopyConfig = () => {
    if (generatedConfig) {
      navigator.clipboard.writeText(generatedConfig)
      alert('Config copied to clipboard!')
    }
  }

  const handleDownloadConfig = () => {
    if (!generatedConfig || !selectedDevice) return

    const blob = new Blob([generatedConfig], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `${selectedDevice.name}.conf`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  }

  const handleAddDevice = async (e: React.FormEvent) => {
    e.preventDefault()
    setFormError(null)
    setSubmitting(true)

    try {
      const token = getAccessToken()
      if (!token) return

      await registerDevice(formData, token)

      // Reset form and close
      setFormData({
        name: '',
        platform: 'linux',
        pubkey: '',
        hostname: '',
        os_version: '',
        daemon_ver: ''
      })
      setShowAddForm(false)
      loadDevices()
      alert('Device registered successfully!')
    } catch (err) {
      setFormError(err instanceof Error ? err.message : 'Failed to register device')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeleteDevice = async (deviceId: string, deviceName: string) => {
    if (!confirm(`Are you sure you want to delete "${deviceName}"?`)) return

    try {
      const token = getAccessToken()
      if (!token) return

      await deleteDevice(deviceId, token)
      alert('Device deleted successfully')
      loadDevices()
    } catch (err) {
      alert('Failed to delete device: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const handleToggleDevice = async (device: Device) => {
    try {
      const token = getAccessToken()
      if (!token) return

      if (device.disabled_at) {
        await enableDevice(device.id, token)
        alert('Device enabled successfully')
      } else {
        await disableDevice(device.id, token)
        alert('Device disabled successfully')
      }
      loadDevices()
    } catch (err) {
      alert('Failed to toggle device: ' + (err instanceof Error ? err.message : 'Unknown error'))
    }
  }

  const getPlatformIcon = (platform: string) => {
    const icons: Record<string, string> = {
      windows: '🪟',
      macos: '🍎',
      linux: '🐧',
      android: '🤖',
      ios: '📱'
    }
    return icons[platform] || '💻'
  }

  const getPlatformColor = (platform: string) => {
    const colors: Record<string, string> = {
      windows: '#0078d4',
      macos: '#000000',
      linux: '#fcc624',
      android: '#3ddc84',
      ios: '#000000'
    }
    return colors[platform] || '#666'
  }

  if (loading) {
    return (
      <AuthGuard>
        <div style={{ padding: 24, textAlign: 'center' }}>
          <p>Loading devices...</p>
        </div>
      </AuthGuard>
    )
  }

  return (
    <AuthGuard>
      <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1200, margin: '0 auto' }}>
        {/* Header */}
        <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
          <h1 style={{ margin: 0 }}>Devices</h1>
          <button
            onClick={() => router.push('/en/dashboard')}
            style={{
              padding: '8px 16px',
              backgroundColor: '#6b7280',
              color: 'white',
              border: 'none',
              borderRadius: 6,
              cursor: 'pointer'
            }}
          >
            ← Back to Dashboard
          </button>
        </div>

        {/* Local Device Registration Banner */}
        {localDaemonStatus && !localDaemonStatus.device.registered && (
          <div style={{
            marginBottom: 24,
            padding: 20,
            backgroundColor: '#eff6ff',
            borderRadius: 8,
            border: '1px solid #bfdbfe',
            display: 'flex',
            justifyContent: 'space-between',
            alignItems: 'center'
          }}>
            <div>
              <h3 style={{ margin: '0 0 8px 0', color: '#1e40af' }}>
                New Device Detected
              </h3>
              <p style={{ margin: 0, color: '#1e3a8a', fontSize: 14 }}>
                This device is running the GoConnect Daemon but is not registered yet.
                <br />
                Public Key: <code style={{ backgroundColor: 'rgba(255,255,255,0.5)', padding: '2px 4px', borderRadius: 4 }}>
                  {localDaemonStatus.device.public_key.substring(0, 12)}...
                </code>
              </p>
            </div>
            <button
              onClick={handleRegisterLocalDevice}
              disabled={registeringLocal}
              style={{
                padding: '10px 20px',
                backgroundColor: '#2563eb',
                color: 'white',
                border: 'none',
                borderRadius: 6,
                cursor: registeringLocal ? 'not-allowed' : 'pointer',
                fontWeight: 600,
                opacity: registeringLocal ? 0.7 : 1
              }}
            >
              {registeringLocal ? 'Registering...' : 'Register This Device'}
            </button>
          </div>
        )}

        {/* Filter and Add Button */}
        <div style={{ marginBottom: 24, display: 'flex', gap: 12, alignItems: 'center' }}>
          <select
            value={filterPlatform}
            onChange={(e) => setFilterPlatform(e.target.value)}
            style={{
              padding: '8px 12px',
              borderRadius: 6,
              border: '1px solid #d1d5db',
              fontSize: 14
            }}
          >
            <option value="">All Platforms</option>
            <option value="windows">Windows</option>
            <option value="macos">macOS</option>
            <option value="linux">Linux</option>
            <option value="android">Android</option>
            <option value="ios">iOS</option>
          </select>

          <button
            onClick={() => setShowAddForm(!showAddForm)}
            style={{
              padding: '8px 16px',
              backgroundColor: '#10b981',
              color: 'white',
              border: 'none',
              borderRadius: 6,
              cursor: 'pointer',
              fontWeight: 500
            }}
          >
            {showAddForm ? '✕ Cancel' : '+ Add Device'}
          </button>
        </div>

        {/* Add Device Form */}
        {showAddForm && (
          <div style={{
            marginBottom: 24,
            padding: 20,
            backgroundColor: '#f9fafb',
            borderRadius: 8,
            border: '1px solid #e5e7eb'
          }}>
            <h3 style={{ marginTop: 0, marginBottom: 16 }}>Register New Device</h3>
            <form onSubmit={handleAddDevice}>
              <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: 16, marginBottom: 16 }}>
                <div>
                  <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                    Device Name *
                  </label>
                  <input
                    type="text"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="e.g., My Laptop"
                    required
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: 6,
                      border: '1px solid #d1d5db',
                      fontSize: 14
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                    Platform *
                  </label>
                  <select
                    value={formData.platform}
                    onChange={(e) => setFormData({ ...formData, platform: e.target.value as PlatformType })}
                    required
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: 6,
                      border: '1px solid #d1d5db',
                      fontSize: 14
                    }}
                  >
                    <option value="windows">Windows</option>
                    <option value="macos">macOS</option>
                    <option value="linux">Linux</option>
                    <option value="android">Android</option>
                    <option value="ios">iOS</option>
                  </select>
                </div>

                <div style={{ gridColumn: '1 / -1' }}>
                  <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                    WireGuard Public Key * (44 characters base64)
                  </label>
                  <input
                    type="text"
                    value={formData.pubkey}
                    onChange={(e) => setFormData({ ...formData, pubkey: e.target.value })}
                    placeholder="e.g., cOvbNjH7xqkK7xKJGVz8M3bKhq8tZ6vS4r9pW3nA2aZ="
                    required
                    minLength={44}
                    maxLength={44}
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: 6,
                      border: '1px solid #d1d5db',
                      fontSize: 14,
                      fontFamily: 'monospace'
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                    Hostname (optional)
                  </label>
                  <input
                    type="text"
                    value={formData.hostname}
                    onChange={(e) => setFormData({ ...formData, hostname: e.target.value })}
                    placeholder="e.g., LAPTOP-123"
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: 6,
                      border: '1px solid #d1d5db',
                      fontSize: 14
                    }}
                  />
                </div>

                <div>
                  <label style={{ display: 'block', marginBottom: 4, fontSize: 14, fontWeight: 500 }}>
                    OS Version (optional)
                  </label>
                  <input
                    type="text"
                    value={formData.os_version}
                    onChange={(e) => setFormData({ ...formData, os_version: e.target.value })}
                    placeholder="e.g., Windows 11"
                    style={{
                      width: '100%',
                      padding: '8px 12px',
                      borderRadius: 6,
                      border: '1px solid #d1d5db',
                      fontSize: 14
                    }}
                  />
                </div>
              </div>

              {formError && (
                <div style={{ padding: 12, backgroundColor: '#fee2e2', color: '#991b1b', borderRadius: 6, marginBottom: 16 }}>
                  {formError}
                </div>
              )}

              <div style={{ display: 'flex', gap: 12 }}>
                <button
                  type="submit"
                  disabled={submitting}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: submitting ? '#9ca3af' : '#10b981',
                    color: 'white',
                    border: 'none',
                    borderRadius: 6,
                    cursor: submitting ? 'not-allowed' : 'pointer',
                    fontWeight: 500
                  }}
                >
                  {submitting ? 'Registering...' : 'Register Device'}
                </button>
                <button
                  type="button"
                  onClick={() => setShowAddForm(false)}
                  style={{
                    padding: '10px 20px',
                    backgroundColor: '#6b7280',
                    color: 'white',
                    border: 'none',
                    borderRadius: 6,
                    cursor: 'pointer'
                  }}
                >
                  Cancel
                </button>
              </div>
            </form>
          </div>
        )}

        {/* Error Message */}
        {error && (
          <div style={{ padding: 16, backgroundColor: '#fee2e2', color: '#991b1b', borderRadius: 6, marginBottom: 24 }}>
            {error}
          </div>
        )}

        {/* Devices List */}
        {devices.length === 0 ? (
          <div style={{ textAlign: 'center', padding: 40, color: '#6b7280' }}>
            <p style={{ fontSize: 18, marginBottom: 8 }}>No devices registered yet</p>
            <p style={{ fontSize: 14 }}>Register your first device to get started with GoConnect VPN</p>
          </div>
        ) : (
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(350px, 1fr))', gap: 16 }}>
            {devices.map(device => (
              <div
                key={device.id}
                style={{
                  padding: 20,
                  backgroundColor: 'white',
                  border: '1px solid #e5e7eb',
                  borderRadius: 8,
                  position: 'relative',
                  opacity: device.disabled_at ? 0.6 : 1
                }}
              >
                {/* Platform Badge */}
                <div style={{
                  position: 'absolute',
                  top: 12,
                  right: 12,
                  fontSize: 24
                }}>
                  {getPlatformIcon(device.platform)}
                </div>

                {/* Device Info */}
                <div style={{ marginBottom: 16 }}>
                  <h3 style={{ margin: 0, marginBottom: 8, fontSize: 18 }}>
                    {device.name}
                    {device.disabled_at && (
                      <span style={{
                        marginLeft: 8,
                        fontSize: 12,
                        padding: '2px 8px',
                        backgroundColor: '#fee2e2',
                        color: '#991b1b',
                        borderRadius: 4
                      }}>
                        DISABLED
                      </span>
                    )}
                  </h3>
                  <div style={{ fontSize: 14, color: '#6b7280', marginBottom: 4 }}>
                    <span style={{
                      padding: '2px 8px',
                      backgroundColor: getPlatformColor(device.platform) + '20',
                      color: getPlatformColor(device.platform),
                      borderRadius: 4,
                      fontWeight: 500
                    }}>
                      {device.platform}
                    </span>
                    {device.active && (
                      <span style={{ marginLeft: 8, color: '#10b981' }}>● Online</span>
                    )}
                  </div>
                  {device.hostname && (
                    <div style={{ fontSize: 13, color: '#6b7280' }}>
                      🖥️ {device.hostname}
                    </div>
                  )}
                  {device.os_version && (
                    <div style={{ fontSize: 13, color: '#6b7280' }}>
                      📦 {device.os_version}
                    </div>
                  )}
                </div>

                {/* Public Key */}
                <div style={{ marginBottom: 16 }}>
                  <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 4 }}>Public Key:</div>
                  <div style={{
                    fontSize: 11,
                    fontFamily: 'monospace',
                    backgroundColor: '#f3f4f6',
                    padding: '6px 8px',
                    borderRadius: 4,
                    wordBreak: 'break-all'
                  }}>
                    {device.pubkey}
                  </div>
                </div>

                {/* Last Seen */}
                {device.last_seen && (
                  <div style={{ fontSize: 12, color: '#6b7280', marginBottom: 12 }}>
                    Last seen: {new Date(device.last_seen).toLocaleString()}
                  </div>
                )}

                {/* Actions */}
                <div style={{ display: 'flex', gap: 8, marginTop: 16 }}>
                  <button
                    onClick={() => handleGetConfig(device)}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      backgroundColor: '#3b82f6',
                      color: 'white',
                      border: 'none',
                      borderRadius: 6,
                      cursor: 'pointer',
                      fontSize: 13
                    }}
                  >
                    📥 Get Config
                  </button>
                  <button
                    onClick={() => handleToggleDevice(device)}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      backgroundColor: device.disabled_at ? '#10b981' : '#f59e0b',
                      color: 'white',
                      border: 'none',
                      borderRadius: 6,
                      cursor: 'pointer',
                      fontSize: 13
                    }}
                  >
                    {device.disabled_at ? '✓ Enable' : '⏸ Disable'}
                  </button>
                  <button
                    onClick={() => handleDeleteDevice(device.id, device.name)}
                    style={{
                      flex: 1,
                      padding: '8px 12px',
                      backgroundColor: '#ef4444',
                      color: 'white',
                      border: 'none',
                      borderRadius: 6,
                      cursor: 'pointer',
                      fontSize: 13
                    }}
                  >
                    🗑️ Delete
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}

        {/* Config Download Modal */}
        {showConfigModal && selectedDevice && (
          <div style={{
            position: 'fixed',
            top: 0,
            left: 0,
            right: 0,
            bottom: 0,
            backgroundColor: 'rgba(0,0,0,0.5)',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            zIndex: 1000
          }}>
            <div style={{
              backgroundColor: 'white',
              padding: 32,
              borderRadius: 12,
              maxWidth: 600,
              width: '90%',
              maxHeight: '90vh',
              overflow: 'auto'
            }}>
              <h2 style={{ marginTop: 0, marginBottom: 24 }}>
                Get WireGuard Config for {selectedDevice.name}
              </h2>

              {!generatedConfig ? (
                <>
                  <div style={{ marginBottom: 20 }}>
                    <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                      Select Network
                    </label>
                    <select
                      value={selectedNetworkId}
                      onChange={(e) => setSelectedNetworkId(e.target.value)}
                      style={{
                        width: '100%',
                        padding: 10,
                        borderRadius: 6,
                        border: '1px solid #d1d5db',
                        fontSize: 14
                      }}
                    >
                      <option value="">-- Choose Network --</option>
                      {networks.map(net => (
                        <option key={net.id} value={net.id}>
                          {net.name} ({net.cidr})
                        </option>
                      ))}
                    </select>
                  </div>

                  <div style={{ marginBottom: 20 }}>
                    <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                      Device Private Key
                    </label>
                    <input
                      type="password"
                      value={privateKey}
                      onChange={(e) => setPrivateKey(e.target.value)}
                      placeholder="Enter your device's WireGuard private key"
                      style={{
                        width: '100%',
                        padding: 10,
                        borderRadius: 6,
                        border: '1px solid #d1d5db',
                        fontSize: 14,
                        fontFamily: 'monospace'
                      }}
                    />
                    <p style={{ fontSize: 12, color: '#6b7280', marginTop: 4 }}>
                      44-character base64 string. Your private key is never stored on the server.
                    </p>
                  </div>

                  {configError && (
                    <div style={{
                      padding: 12,
                      backgroundColor: '#fee2e2',
                      border: '1px solid #fecaca',
                      borderRadius: 6,
                      color: '#991b1b',
                      marginBottom: 20,
                      fontSize: 14
                    }}>
                      {configError}
                    </div>
                  )}

                  <div style={{ display: 'flex', gap: 12 }}>
                    <button
                      onClick={handleGenerateConfig}
                      disabled={configLoading}
                      style={{
                        flex: 1,
                        padding: '10px 16px',
                        backgroundColor: '#3b82f6',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: configLoading ? 'not-allowed' : 'pointer',
                        fontSize: 14,
                        fontWeight: 600,
                        opacity: configLoading ? 0.6 : 1
                      }}
                    >
                      {configLoading ? 'Generating...' : 'Generate Config'}
                    </button>
                    <button
                      onClick={handleCloseConfigModal}
                      style={{
                        flex: 1,
                        padding: '10px 16px',
                        backgroundColor: '#6b7280',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: 'pointer',
                        fontSize: 14,
                        fontWeight: 600
                      }}
                    >
                      Cancel
                    </button>
                  </div>
                </>
              ) : (
                <>
                  <div style={{ marginBottom: 20 }}>
                    <label style={{ display: 'block', marginBottom: 8, fontWeight: 600 }}>
                      Configuration File
                    </label>
                    <textarea
                      value={generatedConfig}
                      readOnly
                      style={{
                        width: '100%',
                        height: showQRCode ? 200 : 300,
                        padding: 12,
                        borderRadius: 6,
                        border: '1px solid #d1d5db',
                        fontSize: 12,
                        fontFamily: 'monospace',
                        resize: 'none',
                        backgroundColor: '#f9fafb'
                      }}
                    />
                  </div>

                  {/* QR Code Section */}
                  {showQRCode && qrCodeUrl && (
                    <div style={{
                      marginBottom: 20,
                      padding: 20,
                      backgroundColor: '#f9fafb',
                      borderRadius: 8,
                      textAlign: 'center'
                    }}>
                      <p style={{
                        fontSize: 14,
                        color: '#6b7280',
                        marginBottom: 12
                      }}>
                        Scan this QR code with WireGuard mobile app
                      </p>
                      <img
                        src={qrCodeUrl}
                        alt="WireGuard Config QR Code"
                        style={{
                          maxWidth: '100%',
                          height: 'auto',
                          border: '2px solid #e5e7eb',
                          borderRadius: 8
                        }}
                      />
                    </div>
                  )}

                  <div style={{ display: 'flex', gap: 12, flexWrap: 'wrap' }}>
                    <button
                      onClick={handleToggleQRCode}
                      style={{
                        flex: 1,
                        minWidth: 120,
                        padding: '10px 16px',
                        backgroundColor: '#8b5cf6',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: 'pointer',
                        fontSize: 14,
                        fontWeight: 600
                      }}
                    >
                      {showQRCode ? '📱 Hide QR' : '📱 Show QR'}
                    </button>
                    <button
                      onClick={handleCopyConfig}
                      style={{
                        flex: 1,
                        padding: '10px 16px',
                        backgroundColor: '#10b981',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: 'pointer',
                        fontSize: 14,
                        fontWeight: 600
                      }}
                    >
                      📋 Copy
                    </button>
                    <button
                      onClick={handleDownloadConfig}
                      style={{
                        flex: 1,
                        padding: '10px 16px',
                        backgroundColor: '#3b82f6',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: 'pointer',
                        fontSize: 14,
                        fontWeight: 600
                      }}
                    >
                      💾 Download
                    </button>
                    <button
                      onClick={handleCloseConfigModal}
                      style={{
                        flex: 1,
                        padding: '10px 16px',
                        backgroundColor: '#6b7280',
                        color: 'white',
                        border: 'none',
                        borderRadius: 6,
                        cursor: 'pointer',
                        fontSize: 14,
                        fontWeight: 600
                      }}
                    >
                      Close
                    </button>
                  </div>
                </>
              )}
            </div>
          </div>
        )}

        <div style={{ marginTop: 40 }}>
          <Footer />
        </div>
      </div>
    </AuthGuard>
  )
}
