'use client'
import { useEffect, useState, useCallback } from 'react'
import { useRouter, useParams } from 'next/navigation'
import { listAuditLogs, AuditLog } from '../../../../../lib/api'
import { getAccessToken, getUser } from '../../../../../lib/auth'
import { useT } from '../../../../../lib/i18n-context'
import AuthGuard from '../../../../../components/AuthGuard'
import Footer from '../../../../../components/Footer'

// Action type categories for filtering
const ACTION_CATEGORIES = {
    network: [
        'NETWORK_CREATED',
        'NETWORK_UPDATED',
        'NETWORK_DELETED',
        'NETWORK_LIST',
        'NETWORK_GET'
    ],
    membership: [
        'NETWORK_JOIN',
        'NETWORK_JOIN_REQUEST',
        'NETWORK_JOIN_APPROVE',
        'NETWORK_JOIN_DENY',
        'NETWORK_MEMBER_KICK',
        'NETWORK_MEMBER_BAN'
    ],
    ip: [
        'IP_ALLOCATED',
        'IP_RELEASED',
        'PEER_PROVISION_FAILED',
        'PEER_DEPROVISION_FAILED'
    ]
}

// Get badge color based on action type
function getActionBadgeStyle(action: string): { backgroundColor: string; color: string } {
    if (action.includes('CREATED') || action.includes('JOIN_APPROVE') || action.includes('ALLOCATED')) {
        return { backgroundColor: '#d4edda', color: '#155724' } // green
    }
    if (action.includes('DELETED') || action.includes('BAN') || action.includes('KICK') || action.includes('DENY')) {
        return { backgroundColor: '#f8d7da', color: '#721c24' } // red
    }
    if (action.includes('UPDATED') || action.includes('RELEASED')) {
        return { backgroundColor: '#fff3cd', color: '#856404' } // yellow
    }
    if (action.includes('REQUEST') || action.includes('LIST') || action.includes('GET')) {
        return { backgroundColor: '#d1ecf1', color: '#0c5460' } // blue
    }
    if (action.includes('FAILED')) {
        return { backgroundColor: '#f5c6cb', color: '#721c24' } // pink-red
    }
    return { backgroundColor: '#e9ecef', color: '#495057' } // gray
}

export default function AuditLogsPage() {
    const router = useRouter()
    const params = useParams()
    const locale = params.locale as string
    const t = useT()

    const [logs, setLogs] = useState<AuditLog[]>([])
    const [filteredLogs, setFilteredLogs] = useState<AuditLog[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [page, setPage] = useState(1)
    const [total, setTotal] = useState(0)
    const limit = 50

    // Filter states
    const [categoryFilter, setCategoryFilter] = useState<string>('all')
    const [actionFilter, setActionFilter] = useState<string>('all')
    const [searchQuery, setSearchQuery] = useState<string>('')
    const [dateFrom, setDateFrom] = useState<string>('')
    const [dateTo, setDateTo] = useState<string>('')

    useEffect(() => {
        const user = getUser()
        if (user && !user.is_admin) {
            router.push(`/${locale}/dashboard`)
            return
        }
        loadLogs()
    }, [page, locale, router])

    const loadLogs = async () => {
        setLoading(true)
        try {
            const token = getAccessToken()
            if (!token) return
            const res = await listAuditLogs(page, limit, token)
            setLogs(res.data || [])
            setTotal(res.pagination.total)
        } catch (err) {
            setError(err instanceof Error ? err.message : t('audit.error.load'))
        } finally {
            setLoading(false)
        }
    }

    // Apply filters
    const applyFilters = useCallback(() => {
        let result = [...logs]

        // Category filter
        if (categoryFilter !== 'all') {
            const actions = ACTION_CATEGORIES[categoryFilter as keyof typeof ACTION_CATEGORIES] || []
            result = result.filter(log => actions.some(a => log.action.includes(a) || log.action === a))
        }

        // Specific action filter
        if (actionFilter !== 'all') {
            result = result.filter(log => log.action === actionFilter)
        }

        // Search query (actor or object)
        if (searchQuery) {
            const query = searchQuery.toLowerCase()
            result = result.filter(log =>
                log.actor.toLowerCase().includes(query) ||
                log.object.toLowerCase().includes(query) ||
                JSON.stringify(log.details).toLowerCase().includes(query)
            )
        }

        // Date range filter
        if (dateFrom) {
            const from = new Date(dateFrom)
            result = result.filter(log => new Date(log.timestamp) >= from)
        }
        if (dateTo) {
            const to = new Date(dateTo)
            to.setHours(23, 59, 59, 999)
            result = result.filter(log => new Date(log.timestamp) <= to)
        }

        setFilteredLogs(result)
    }, [logs, categoryFilter, actionFilter, searchQuery, dateFrom, dateTo])

    useEffect(() => {
        applyFilters()
    }, [applyFilters])

    // Get unique actions from logs for dropdown
    const uniqueActions = [...new Set(logs.map(log => log.action))].sort()

    const clearFilters = () => {
        setCategoryFilter('all')
        setActionFilter('all')
        setSearchQuery('')
        setDateFrom('')
        setDateTo('')
    }

    const formatDetails = (details: any): string => {
        if (!details) return '-'
        if (typeof details === 'string') return details
        const entries = Object.entries(details)
        if (entries.length === 0) return '-'
        return entries.map(([k, v]) => `${k}: ${v}`).join(', ')
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1400, margin: '0 auto' }}>
                {/* Header */}
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
                    <div>
                        <h1 style={{ margin: 0, fontSize: 24, fontWeight: 600 }}>{t('audit.title')}</h1>
                        <p style={{ margin: '4px 0 0', color: '#6b7280', fontSize: 14 }}>
                            {t('audit.subtitle', { total: total.toString() })}
                        </p>
                    </div>
                    <div style={{ display: 'flex', gap: 8 }}>
                        <button
                            onClick={() => router.push(`/${locale}/admin`)}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#f3f4f6',
                                color: '#374151',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 13
                            }}
                        >
                            ← {t('audit.backToAdmin')}
                        </button>
                        <button
                            onClick={loadLogs}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#3b82f6',
                                color: 'white',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 13
                            }}
                        >
                            🔄 {t('audit.refresh')}
                        </button>
                    </div>
                </div>

                {/* Filters */}
                <div style={{
                    backgroundColor: 'white',
                    borderRadius: 8,
                    border: '1px solid #e5e7eb',
                    padding: 16,
                    marginBottom: 16
                }}>
                    <div style={{ display: 'flex', gap: 16, flexWrap: 'wrap', alignItems: 'flex-end' }}>
                        {/* Category Filter */}
                        <div>
                            <label style={{ display: 'block', fontSize: 12, color: '#6b7280', marginBottom: 4 }}>
                                {t('audit.filter.category')}
                            </label>
                            <select
                                value={categoryFilter}
                                onChange={(e) => {
                                    setCategoryFilter(e.target.value)
                                    setActionFilter('all')
                                }}
                                style={{
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: 6,
                                    fontSize: 13,
                                    minWidth: 140
                                }}
                            >
                                <option value="all">{t('audit.filter.allCategories')}</option>
                                <option value="network">{t('audit.filter.network')}</option>
                                <option value="membership">{t('audit.filter.membership')}</option>
                                <option value="ip">{t('audit.filter.ip')}</option>
                            </select>
                        </div>

                        {/* Action Filter */}
                        <div>
                            <label style={{ display: 'block', fontSize: 12, color: '#6b7280', marginBottom: 4 }}>
                                {t('audit.filter.action')}
                            </label>
                            <select
                                value={actionFilter}
                                onChange={(e) => setActionFilter(e.target.value)}
                                style={{
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: 6,
                                    fontSize: 13,
                                    minWidth: 180
                                }}
                            >
                                <option value="all">{t('audit.filter.allActions')}</option>
                                {uniqueActions.map(action => (
                                    <option key={action} value={action}>{action}</option>
                                ))}
                            </select>
                        </div>

                        {/* Search */}
                        <div style={{ flex: 1, minWidth: 200 }}>
                            <label style={{ display: 'block', fontSize: 12, color: '#6b7280', marginBottom: 4 }}>
                                {t('audit.filter.search')}
                            </label>
                            <input
                                type="text"
                                value={searchQuery}
                                onChange={(e) => setSearchQuery(e.target.value)}
                                placeholder={t('audit.filter.searchPlaceholder')}
                                style={{
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: 6,
                                    fontSize: 13,
                                    width: '100%',
                                    boxSizing: 'border-box'
                                }}
                            />
                        </div>

                        {/* Date From */}
                        <div>
                            <label style={{ display: 'block', fontSize: 12, color: '#6b7280', marginBottom: 4 }}>
                                {t('audit.filter.dateFrom')}
                            </label>
                            <input
                                type="date"
                                value={dateFrom}
                                onChange={(e) => setDateFrom(e.target.value)}
                                style={{
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: 6,
                                    fontSize: 13
                                }}
                            />
                        </div>

                        {/* Date To */}
                        <div>
                            <label style={{ display: 'block', fontSize: 12, color: '#6b7280', marginBottom: 4 }}>
                                {t('audit.filter.dateTo')}
                            </label>
                            <input
                                type="date"
                                value={dateTo}
                                onChange={(e) => setDateTo(e.target.value)}
                                style={{
                                    padding: '8px 12px',
                                    border: '1px solid #d1d5db',
                                    borderRadius: 6,
                                    fontSize: 13
                                }}
                            />
                        </div>

                        {/* Clear Filters */}
                        <button
                            onClick={clearFilters}
                            style={{
                                padding: '8px 16px',
                                backgroundColor: '#f3f4f6',
                                color: '#374151',
                                border: 'none',
                                borderRadius: 6,
                                cursor: 'pointer',
                                fontSize: 13
                            }}
                        >
                            {t('audit.filter.clear')}
                        </button>
                    </div>

                    {/* Filter Status */}
                    <div style={{ marginTop: 12, fontSize: 13, color: '#6b7280' }}>
                        {t('audit.showing', { count: filteredLogs.length.toString(), total: logs.length.toString() })}
                    </div>
                </div>

                {error && <div style={{ color: 'red', marginBottom: 16 }}>{error}</div>}

                {/* Table */}
                <div style={{ backgroundColor: 'white', borderRadius: 8, border: '1px solid #e5e7eb', overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead style={{ backgroundColor: '#f9fafb', borderBottom: '1px solid #e5e7eb' }}>
                            <tr>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: 12, fontWeight: 600, color: '#374151' }}>
                                    {t('audit.table.time')}
                                </th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: 12, fontWeight: 600, color: '#374151' }}>
                                    {t('audit.table.action')}
                                </th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: 12, fontWeight: 600, color: '#374151' }}>
                                    {t('audit.table.actor')}
                                </th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: 12, fontWeight: 600, color: '#374151' }}>
                                    {t('audit.table.object')}
                                </th>
                                <th style={{ padding: '12px 16px', textAlign: 'left', fontSize: 12, fontWeight: 600, color: '#374151' }}>
                                    {t('audit.table.details')}
                                </th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr>
                                    <td colSpan={5} style={{ padding: 48, textAlign: 'center', color: '#6b7280' }}>
                                        {t('audit.loading')}
                                    </td>
                                </tr>
                            ) : filteredLogs.length === 0 ? (
                                <tr>
                                    <td colSpan={5} style={{ padding: 48, textAlign: 'center', color: '#6b7280' }}>
                                        {t('audit.noLogs')}
                                    </td>
                                </tr>
                            ) : (
                                filteredLogs.map((log) => {
                                    const badgeStyle = getActionBadgeStyle(log.action)
                                    return (
                                        <tr key={log.seq} style={{ borderBottom: '1px solid #f3f4f6' }}>
                                            <td style={{ padding: '12px 16px', fontSize: 13, color: '#374151', whiteSpace: 'nowrap' }}>
                                                {new Date(log.timestamp).toLocaleString()}
                                            </td>
                                            <td style={{ padding: '12px 16px' }}>
                                                <span style={{
                                                    padding: '4px 8px',
                                                    borderRadius: 4,
                                                    fontSize: 11,
                                                    fontWeight: 600,
                                                    ...badgeStyle
                                                }}>
                                                    {log.action}
                                                </span>
                                            </td>
                                            <td style={{ padding: '12px 16px', fontSize: 13, color: '#374151' }}>
                                                <code style={{
                                                    backgroundColor: '#f3f4f6',
                                                    padding: '2px 6px',
                                                    borderRadius: 4,
                                                    fontSize: 12
                                                }}>
                                                    {log.actor.substring(0, 12)}...
                                                </code>
                                            </td>
                                            <td style={{ padding: '12px 16px', fontSize: 13, color: '#374151' }}>
                                                {log.object ? (
                                                    <code style={{
                                                        backgroundColor: '#f3f4f6',
                                                        padding: '2px 6px',
                                                        borderRadius: 4,
                                                        fontSize: 12
                                                    }}>
                                                        {log.object.substring(0, 12)}...
                                                    </code>
                                                ) : '-'}
                                            </td>
                                            <td style={{ padding: '12px 16px', fontSize: 12, color: '#6b7280', fontFamily: 'monospace' }}>
                                                {formatDetails(log.details)}
                                            </td>
                                        </tr>
                                    )
                                })
                            )}
                        </tbody>
                    </table>
                </div>

                {/* Pagination */}
                <div style={{ marginTop: 16, display: 'flex', justifyContent: 'center', alignItems: 'center', gap: 12 }}>
                    <button
                        disabled={page === 1}
                        onClick={() => setPage(p => p - 1)}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: page === 1 ? '#f3f4f6' : 'white',
                            color: page === 1 ? '#9ca3af' : '#374151',
                            border: '1px solid #d1d5db',
                            borderRadius: 6,
                            cursor: page === 1 ? 'not-allowed' : 'pointer',
                            fontSize: 13
                        }}
                    >
                        ← {t('audit.pagination.previous')}
                    </button>
                    <span style={{ fontSize: 13, color: '#6b7280' }}>
                        {t('audit.pagination.page', { page: page.toString(), total: Math.ceil(total / limit).toString() })}
                    </span>
                    <button
                        disabled={page * limit >= total}
                        onClick={() => setPage(p => p + 1)}
                        style={{
                            padding: '8px 16px',
                            backgroundColor: page * limit >= total ? '#f3f4f6' : 'white',
                            color: page * limit >= total ? '#9ca3af' : '#374151',
                            border: '1px solid #d1d5db',
                            borderRadius: 6,
                            cursor: page * limit >= total ? 'not-allowed' : 'pointer',
                            fontSize: 13
                        }}
                    >
                        {t('audit.pagination.next')} →
                    </button>
                </div>
            </div>
            <Footer />
        </AuthGuard>
    )
}
