'use client'
import React, { createContext, useContext, useEffect, useState } from 'react'
import { bridge } from '../lib/bridge'

interface DaemonStatus {
    running: boolean
    version: string
    paused?: boolean
    wg: {
        active: boolean
        error?: string
        [key: string]: any
    }
    device: {
        registered: boolean
        public_key: string
        device_id: string
    }
}

interface DaemonContextType {
    status: DaemonStatus | null
    error: string | null
    isLoading: boolean
    refresh: () => Promise<void>
    connect: () => Promise<void>
    disconnect: () => Promise<void>
}

const DaemonContext = createContext<DaemonContextType | undefined>(undefined)

export function DaemonProvider({ children }: { children: React.ReactNode }) {
    const [status, setStatus] = useState<DaemonStatus | null>(null)
    const [error, setError] = useState<string | null>(null)
    const [isLoading, setIsLoading] = useState(true)

    const fetchStatus = async () => {
        try {
            const data = await bridge('/status')
            setStatus(data)
            setError(null)
        } catch (err) {
            // console.error('Failed to fetch daemon status:', err)
            setError('Failed to connect to local daemon')
            setStatus(null)
        } finally {
            setIsLoading(false)
        }
    }

    const connect = async () => {
        await bridge('/connect', { method: 'POST' })
        await fetchStatus()
    }

    const disconnect = async () => {
        await bridge('/disconnect', { method: 'POST' })
        await fetchStatus()
    }

    useEffect(() => {
        fetchStatus()
        const interval = setInterval(fetchStatus, 5000)
        return () => clearInterval(interval)
    }, [])

    return (
        <DaemonContext.Provider value={{ status, error, isLoading, refresh: fetchStatus, connect, disconnect }}>
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
