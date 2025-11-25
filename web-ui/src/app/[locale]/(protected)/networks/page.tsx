'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { listNetworks, createNetwork, deleteNetwork, Network, CreateNetworkRequest } from '../../../../lib/api'
import { getAccessToken, getUser } from '../../../../lib/auth'
import { useT } from '../../../../lib/i18n-context'
import AuthGuard from '../../../../components/AuthGuard'
import Footer from '../../../../components/Footer'

type ViewMode = 'public' | 'mine' | 'all'

export default function NetworksPage() {
    const router = useRouter()
    const t = useT()
    const [networks, setNetworks] = useState<Network[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [viewMode, setViewMode] = useState<ViewMode>('public')
    const [showCreateModal, setShowCreateModal] = useState(false)
    const [isAdmin, setIsAdmin] = useState(false)

    useEffect(() => {
        const user = getUser()
        if (user?.is_admin) {
            setIsAdmin(true)
        }
        loadNetworks()
    }, [viewMode])

    const loadNetworks = async () => {
        setLoading(true)
        setError(null)
        try {
            const token = getAccessToken()
            if (!token) {
                router.push('/en/login')
                return
            }

            const response = await listNetworks(viewMode, token)
            setNetworks(response.data || [])
        } catch (err) {
            setError(err instanceof Error ? err.message : t('networks.error.load'))
        } finally {
            setLoading(false)
        }
    }

    const handleCreateNetwork = async (req: CreateNetworkRequest) => {
        try {
            const token = getAccessToken()
            if (!token) return

            await createNetwork(req, token)
            setShowCreateModal(false)
            loadNetworks() // Reload list
        } catch (err) {
            throw err // Let modal handle the error
        }
    }

    const handleDeleteNetwork = async (id: string, name: string) => {
        if (!confirm(t('networks.confirm.delete', { name }))) return

        try {
            const token = getAccessToken()
            if (!token) return

            await deleteNetwork(id, token)
            loadNetworks() // Reload list
        } catch (err) {
            alert(t('networks.error.delete') + ': ' + (err instanceof Error ? err.message : t('networks.error.unknown')))
        }
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1200, margin: '0 auto' }}>
                {/* Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
                    <div>
                        <h1 style={{ margin: 0, marginBottom: 8 }}>{t('networks.title')}</h1>
                        <div style={{ display: 'flex', gap: 8 }}>
                            <button
                                onClick={() => setViewMode('public')}
                                style={{
                                    padding: '6px 12px',
                                    backgroundColor: viewMode === 'public' ? '#007bff' : '#fff',
                                    color: viewMode === 'public' ? '#fff' : '#333',
                                    border: '1px solid #dee2e6',
                                    borderRadius: 4,
                                    cursor: 'pointer',
                                    fontSize: 13
                                }}
                            >
                                {t('networks.view.public')}
                            </button>
                            <button
                                onClick={() => setViewMode('mine')}
                                style={{
                                    padding: '6px 12px',
                                    backgroundColor: viewMode === 'mine' ? '#007bff' : '#fff',
                                    color: viewMode === 'mine' ? '#fff' : '#333',
                                    border: '1px solid #dee2e6',
                                    borderRadius: 4,
                                    cursor: 'pointer',
                                    fontSize: 13
                                }}
                            >
                                {t('networks.view.mine')}
                            </button>
                            {isAdmin && (
                                <button
                                    onClick={() => setViewMode('all')}
                                    style={{
                                        padding: '6px 12px',
                                        backgroundColor: viewMode === 'all' ? '#007bff' : '#fff',
                                        color: viewMode === 'all' ? '#fff' : '#333',
                                        border: '1px solid #dee2e6',
                                        borderRadius: 4,
                                        cursor: 'pointer',
                                        fontSize: 13
                                    }}
                                >
                                    {t('networks.view.all')}
                                </button>
                            )}
                        </div>
                    </div>
                    <button
                        onClick={() => setShowCreateModal(true)}
                        style={{
                            padding: '10px 20px',
                            backgroundColor: '#28a745',
                            color: 'white',
                            border: 'none',
                            borderRadius: 4,
                            cursor: 'pointer',
                            fontSize: 14,
                            fontWeight: 500
                        }}
                    >
                        {t('networks.action.create')}
                    </button>
                </div>

                {/* Loading State */}
                {loading && (
                    <div style={{ textAlign: 'center', padding: 40, color: '#666' }}>
                        {t('networks.loading')}
                    </div>
                )}

                {/* Error State */}
                {error && (
                    <div style={{ padding: 16, backgroundColor: '#f8d7da', color: '#721c24', borderRadius: 4, marginBottom: 16 }}>
                        {error}
                    </div>
                )}

                {/* Networks List */}
                {!loading && !error && (
                    <div style={{ display: 'grid', gridTemplateColumns: 'repeat(auto-fill, minmax(300px, 1fr))', gap: 16 }}>
                        {networks.length === 0 ? (
                            <div style={{ gridColumn: '1 / -1', textAlign: 'center', padding: 40, color: '#666' }}>
                                {t('networks.empty')}
                            </div>
                        ) : (
                            networks.map(network => (
                                <NetworkCard
                                    key={network.id}
                                    network={network}
                                    onDelete={handleDeleteNetwork}
                                    isAdmin={isAdmin}
                                />
                            ))
                        )}
                    </div>
                )}

                {/* Create Network Modal */}
                {showCreateModal && (
                    <CreateNetworkModal
                        onClose={() => setShowCreateModal(false)}
                        onCreate={handleCreateNetwork}
                    />
                )}

                <div style={{ marginTop: 40 }}>
                    <Footer />
                </div>
            </div>
        </AuthGuard>
    )
}

// Network Card Component
function NetworkCard({ network, onDelete, isAdmin }: { network: Network; onDelete: (id: string, name: string) => void; isAdmin: boolean }) {
    const router = useRouter()
    const t = useT()

    return (
        <div
            onClick={() => router.push(`/en/networks/${network.id}`)}
            style={{
                padding: 16,
                backgroundColor: '#fff',
                border: '1px solid #dee2e6',
                borderRadius: 8,
                transition: 'box-shadow 0.2s',
                cursor: 'pointer'
            }}
            onMouseEnter={(e) => e.currentTarget.style.boxShadow = '0 4px 12px rgba(0,0,0,0.1)'}
            onMouseLeave={(e) => e.currentTarget.style.boxShadow = 'none'}
        >
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'start', marginBottom: 12 }}>
                <h3 style={{ margin: 0, fontSize: 18 }}>{network.name}</h3>
                {isAdmin && (
                    <button
                        onClick={(e) => {
                            e.stopPropagation()
                            onDelete(network.id, network.name)
                        }}
                        style={{
                            padding: '4px 8px',
                            backgroundColor: '#dc3545',
                            color: 'white',
                            border: 'none',
                            borderRadius: 4,
                            cursor: 'pointer',
                            fontSize: 12
                        }}
                    >
                        {t('networks.card.delete')}
                    </button>
                )}
            </div>
            <div style={{ marginBottom: 8, color: '#666', fontSize: 14 }}>
                <div><strong>{t('networks.card.cidr')}</strong> {network.cidr}</div>
                <div style={{ display: 'flex', gap: 8, marginTop: 4 }}>
                    <span style={{
                        padding: '2px 8px',
                        backgroundColor: network.visibility === 'public' ? '#d4edda' : '#f8d7da',
                        color: network.visibility === 'public' ? '#155724' : '#721c24',
                        borderRadius: 4,
                        fontSize: 12
                    }}>
                        {network.visibility === 'public' ? t('networks.card.visibility.public') : t('networks.card.visibility.private')}
                    </span>
                    <span style={{
                        padding: '2px 8px',
                        backgroundColor: '#d1ecf1',
                        color: '#0c5460',
                        borderRadius: 4,
                        fontSize: 12
                    }}>
                        {network.join_policy === 'open' ? t('networks.card.policy.open') : network.join_policy === 'approval' ? t('networks.card.policy.approval') : t('networks.card.policy.invite')}
                    </span>
                </div>
            </div>
            <div style={{ fontSize: 12, color: '#999' }}>
                {t('networks.card.created')} {new Date(network.created_at).toLocaleDateString()}
            </div>
        </div>
    )
}

// Create Network Modal Component
function CreateNetworkModal({ onClose, onCreate }: { onClose: () => void; onCreate: (req: CreateNetworkRequest) => Promise<void> }) {
    const t = useT()
    const [name, setName] = useState('')
    const [visibility, setVisibility] = useState<'public' | 'private'>('public')
    const [joinPolicy, setJoinPolicy] = useState<'open' | 'approval' | 'invite'>('open')
    const [cidr, setCidr] = useState('')
    const [dns, setDns] = useState('')
    const [mtu, setMtu] = useState('')
    const [splitTunnel, setSplitTunnel] = useState(false)
    const [loading, setLoading] = useState(false)
    const [error, setError] = useState<string | null>(null)

    const handleSubmit = async (e: React.FormEvent) => {
        e.preventDefault()
        setError(null)

        // Validation
        if (!name.trim()) {
            setError(t('networks.modal.error.nameRequired'))
            return
        }

        if (!cidr.trim()) {
            setError(t('networks.modal.error.cidrRequired'))
            return
        }

        // Basic CIDR validation
        const cidrRegex = /^(\d{1,3}\.){3}\d{1,3}\/\d{1,2}$/
        if (!cidrRegex.test(cidr)) {
            setError(t('networks.modal.error.cidrInvalid'))
            return
        }

        setLoading(true)

        try {
            const req: CreateNetworkRequest = {
                name: name.trim(),
                visibility,
                join_policy: joinPolicy,
                cidr: cidr.trim(),
            }

            if (dns.trim()) req.dns = dns.trim()
            if (mtu.trim()) req.mtu = parseInt(mtu)
            req.split_tunnel = splitTunnel

            await onCreate(req)
        } catch (err) {
            setError(err instanceof Error ? err.message : t('networks.modal.error.create'))
            setLoading(false)
        }
    }

    return (
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
                borderRadius: 8,
                padding: 24,
                maxWidth: 500,
                width: '90%',
                maxHeight: '90vh',
                overflow: 'auto'
            }}>
                <h2 style={{ marginTop: 0 }}>{t('networks.modal.title')}</h2>

                {error && (
                    <div style={{ padding: 12, backgroundColor: '#f8d7da', color: '#721c24', borderRadius: 4, marginBottom: 16, fontSize: 14 }}>
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit}>
                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.name')}
                        </label>
                        <input
                            type="text"
                            value={name}
                            onChange={(e) => setName(e.target.value)}
                            placeholder={t('networks.modal.placeholder.name')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        />
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.cidr')}
                        </label>
                        <input
                            type="text"
                            value={cidr}
                            onChange={(e) => setCidr(e.target.value)}
                            placeholder={t('networks.modal.placeholder.cidr')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        />
                        <div style={{ fontSize: 12, color: '#666', marginTop: 4 }}>
                            {t('networks.modal.help.cidr')}
                        </div>
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.visibility')}
                        </label>
                        <select
                            value={visibility}
                            onChange={(e) => setVisibility(e.target.value as 'public' | 'private')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        >
                            <option value="public">{t('networks.modal.option.public')}</option>
                            <option value="private">{t('networks.modal.option.private')}</option>
                        </select>
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.policy')}
                        </label>
                        <select
                            value={joinPolicy}
                            onChange={(e) => setJoinPolicy(e.target.value as 'open' | 'approval' | 'invite')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        >
                            <option value="open">{t('networks.modal.option.open')}</option>
                            <option value="approval">{t('networks.modal.option.approval')}</option>
                            <option value="invite">{t('networks.modal.option.invite')}</option>
                        </select>
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.dns')}
                        </label>
                        <input
                            type="text"
                            value={dns}
                            onChange={(e) => setDns(e.target.value)}
                            placeholder={t('networks.modal.placeholder.dns')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        />
                    </div>

                    <div style={{ marginBottom: 16 }}>
                        <label style={{ display: 'block', marginBottom: 4, fontWeight: 500, fontSize: 14 }}>
                            {t('networks.modal.label.mtu')}
                        </label>
                        <input
                            type="number"
                            value={mtu}
                            onChange={(e) => setMtu(e.target.value)}
                            placeholder={t('networks.modal.placeholder.mtu')}
                            style={{
                                width: '100%',
                                padding: 8,
                                border: '1px solid #dee2e6',
                                borderRadius: 4,
                                fontSize: 14,
                                boxSizing: 'border-box'
                            }}
                        />
                    </div>

                    <div style={{ marginBottom: 24 }}>
                        <label style={{ display: 'flex', alignItems: 'center', cursor: 'pointer' }}>
                            <input
                                type="checkbox"
                                checked={splitTunnel}
                                onChange={(e) => setSplitTunnel(e.target.checked)}
                                style={{ marginRight: 8 }}
                            />
                            <span style={{ fontSize: 14 }}>{t('networks.modal.label.splitTunnel')}</span>
                        </label>
                    </div>

                    <div style={{ display: 'flex', gap: 8, justifyContent: 'flex-end' }}>
                        <button
                            type="button"
                            onClick={onClose}
                            disabled={loading}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#6c757d',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: loading ? 'not-allowed' : 'pointer',
                                fontSize: 14,
                                opacity: loading ? 0.6 : 1
                            }}
                        >
                            {t('networks.modal.action.cancel')}
                        </button>
                        <button
                            type="submit"
                            disabled={loading}
                            style={{
                                padding: '10px 20px',
                                backgroundColor: '#28a745',
                                color: 'white',
                                border: 'none',
                                borderRadius: 4,
                                cursor: loading ? 'not-allowed' : 'pointer',
                                fontSize: 14,
                                opacity: loading ? 0.6 : 1
                            }}
                        >
                            {loading ? t('networks.modal.action.creating') : t('networks.modal.action.create')}
                        </button>
                    </div>
                </form>
            </div>
        </div>
    )
}
