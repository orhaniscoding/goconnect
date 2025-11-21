'use client'
import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { listAuditLogs, AuditLog } from '../../../../../lib/api'
import { getAccessToken, getUser } from '../../../../../lib/auth'
import AuthGuard from '../../../../../components/AuthGuard'
import Footer from '../../../../../components/Footer'

export default function AuditLogsPage() {
    const router = useRouter()
    const [logs, setLogs] = useState<AuditLog[]>([])
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [page, setPage] = useState(1)
    const [total, setTotal] = useState(0)
    const limit = 20

    useEffect(() => {
        const user = getUser()
        if (user && !user.is_admin) {
            router.push('/en/dashboard')
            return
        }
        loadLogs()
    }, [page])

    const loadLogs = async () => {
        setLoading(true)
        try {
            const token = getAccessToken()
            if (!token) return
            const res = await listAuditLogs(page, limit, token)
            setLogs(res.data || [])
            setTotal(res.pagination.total)
        } catch (err) {
            setError(err instanceof Error ? err.message : 'Failed to load logs')
        } finally {
            setLoading(false)
        }
    }

    return (
        <AuthGuard>
            <div style={{ padding: 24, fontFamily: 'system-ui, -apple-system, sans-serif', maxWidth: 1200, margin: '0 auto' }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 24 }}>
                    <h1 style={{ margin: 0 }}>Audit Logs</h1>
                    <button onClick={loadLogs} style={{ padding: '8px 16px', cursor: 'pointer' }}>Refresh</button>
                </div>

                {error && <div style={{ color: 'red', marginBottom: 16 }}>{error}</div>}

                <div style={{ backgroundColor: 'white', borderRadius: 8, border: '1px solid #dee2e6', overflow: 'hidden' }}>
                    <table style={{ width: '100%', borderCollapse: 'collapse' }}>
                        <thead style={{ backgroundColor: '#f8f9fa', borderBottom: '1px solid #dee2e6' }}>
                            <tr>
                                <th style={{ padding: 12, textAlign: 'left' }}>Time</th>
                                <th style={{ padding: 12, textAlign: 'left' }}>Action</th>
                                <th style={{ padding: 12, textAlign: 'left' }}>Actor</th>
                                <th style={{ padding: 12, textAlign: 'left' }}>Object</th>
                                <th style={{ padding: 12, textAlign: 'left' }}>Details</th>
                            </tr>
                        </thead>
                        <tbody>
                            {loading ? (
                                <tr><td colSpan={5} style={{ padding: 24, textAlign: 'center' }}>Loading...</td></tr>
                            ) : logs.length === 0 ? (
                                <tr><td colSpan={5} style={{ padding: 24, textAlign: 'center' }}>No logs found</td></tr>
                            ) : (
                                logs.map((log) => (
                                    <tr key={log.seq} style={{ borderBottom: '1px solid #eee' }}>
                                        <td style={{ padding: 12, fontSize: 13 }}>{new Date(log.timestamp).toLocaleString()}</td>
                                        <td style={{ padding: 12, fontSize: 13 }}>
                                            <span style={{
                                                padding: '2px 6px',
                                                borderRadius: 4,
                                                backgroundColor: '#e9ecef',
                                                fontSize: 12,
                                                fontWeight: 500
                                            }}>
                                                {log.action}
                                            </span>
                                        </td>
                                        <td style={{ padding: 12, fontSize: 13 }}>{log.actor}</td>
                                        <td style={{ padding: 12, fontSize: 13 }}>{log.object}</td>
                                        <td style={{ padding: 12, fontSize: 13, fontFamily: 'monospace' }}>
                                            {JSON.stringify(log.details)}
                                        </td>
                                    </tr>
                                ))
                            )}
                        </tbody>
                    </table>
                </div>

                <div style={{ marginTop: 16, display: 'flex', justifyContent: 'center', gap: 8 }}>
                    <button
                        disabled={page === 1}
                        onClick={() => setPage(p => p - 1)}
                        style={{ padding: '6px 12px', cursor: page === 1 ? 'not-allowed' : 'pointer' }}
                    >
                        Previous
                    </button>
                    <span style={{ padding: '6px 12px' }}>Page {page}</span>
                    <button
                        disabled={page * limit >= total}
                        onClick={() => setPage(p => p + 1)}
                        style={{ padding: '6px 12px', cursor: page * limit >= total ? 'not-allowed' : 'pointer' }}
                    >
                        Next
                    </button>
                </div>
            </div>
            <Footer />
        </AuthGuard>
    )
}
