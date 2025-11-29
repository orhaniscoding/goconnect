'use client'

import { createContext, useContext, useEffect, useState, ReactNode } from 'react'
import { bridge, DaemonStatus } from '@/lib/bridge'

interface DaemonContextType {
  status: DaemonStatus | null
  loading: boolean
  error: string | null
  refresh: () => Promise<void>
  connect: () => Promise<void>
  disconnect: () => Promise<void>
  register: (token: string) => Promise<void>
}

const DaemonContext = createContext<DaemonContextType | undefined>(undefined)

export function DaemonProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<DaemonStatus | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const refresh = async () => {
    try {
      const s = await bridge.getStatus()
      setStatus(s)
      setError(null)
    } catch (e) {
      setStatus(null)
      // Don't set error string on polling fail to avoid UI spam, just set status null
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    refresh()
    // Poll every 5 seconds
    const interval = setInterval(refresh, 5000)
    return () => clearInterval(interval)
  }, [])

  const connect = async () => {
    try {
      await bridge.connect()
      await refresh()
    } catch (e: any) {
      setError(e.message)
      throw e
    }
  }

  const disconnect = async () => {
    try {
      await bridge.disconnect()
      await refresh()
    } catch (e: any) {
      setError(e.message)
      throw e
    }
  }

  const register = async (token: string) => {
    try {
      await bridge.register(token)
      await refresh()
    } catch (e: any) {
      setError(e.message)
      throw e
    }
  }

  return (
    <DaemonContext.Provider value={{ status, loading, error, refresh, connect, disconnect, register }}>
      {children}
    </DaemonContext.Provider>
  )
}

export function useDaemon() {
  const context = useContext(DaemonContext)
  if (context === undefined) {
    throw new Error('useDaemon must be used within a DaemonProvider')
  }
  return context
}