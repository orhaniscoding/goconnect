'use client'
import { useEffect, useState } from 'react'
import { bridge } from '../../../lib/bridge'
import Footer from '../../../components/Footer'

export default function Dashboard() {
  const [status, setStatus] = useState<any>(null)
  const [err, setErr] = useState<string | null>(null)
  useEffect(() => {
    bridge('/status')
      .then(setStatus)
      .catch((e) => setErr(String(e)))
  }, [])
  return (
    <div style={{ padding: 24 }}>
      <h1>Dashboard</h1>
      {err ? <p style={{color:'crimson'}}>Bridge error: {err}</p> :
        <pre>{JSON.stringify(status, null, 2)}</pre>}
      <Footer />
    </div>
  )
}
