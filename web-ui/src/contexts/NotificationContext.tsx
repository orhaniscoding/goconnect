'use client'
import React, { createContext, useContext, useState, useCallback, ReactNode } from 'react'

export type NotificationType = 'success' | 'error' | 'info' | 'warning'

export interface Notification {
    id: string
    type: NotificationType
    title: string
    message?: string
    duration?: number
}

interface NotificationContextType {
    notifications: Notification[]
    addNotification: (notification: Omit<Notification, 'id'>) => void
    removeNotification: (id: string) => void
    success: (title: string, message?: string) => void
    error: (title: string, message?: string) => void
    info: (title: string, message?: string) => void
    warning: (title: string, message?: string) => void
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined)

export function NotificationProvider({ children }: { children: ReactNode }) {
    const [notifications, setNotifications] = useState<Notification[]>([])

    const removeNotification = useCallback((id: string) => {
        setNotifications((prev) => prev.filter((n) => n.id !== id))
    }, [])

    const addNotification = useCallback((notification: Omit<Notification, 'id'>) => {
        const id = `notification-${Date.now()}-${Math.random()}`
        const duration = notification.duration ?? 5000

        setNotifications((prev) => [...prev, { ...notification, id }])

        if (duration > 0) {
            setTimeout(() => {
                removeNotification(id)
            }, duration)
        }
    }, [removeNotification])

    const success = useCallback((title: string, message?: string) => {
        addNotification({ type: 'success', title, message })
    }, [addNotification])

    const error = useCallback((title: string, message?: string) => {
        addNotification({ type: 'error', title, message })
    }, [addNotification])

    const info = useCallback((title: string, message?: string) => {
        addNotification({ type: 'info', title, message })
    }, [addNotification])

    const warning = useCallback((title: string, message?: string) => {
        addNotification({ type: 'warning', title, message })
    }, [addNotification])

    return (
        <NotificationContext.Provider
            value={{
                notifications,
                addNotification,
                removeNotification,
                success,
                error,
                info,
                warning
            }}
        >
            {children}
        </NotificationContext.Provider>
    )
}

export function useNotification() {
    const context = useContext(NotificationContext)
    if (!context) {
        throw new Error('useNotification must be used within NotificationProvider')
    }
    return context
}
