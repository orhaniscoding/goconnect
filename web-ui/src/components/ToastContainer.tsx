'use client'
import { useEffect, useState } from 'react'
import { useNotification, Notification, NotificationType } from '../contexts/NotificationContext'

const getNotificationStyles = (type: NotificationType) => {
    const baseStyles = {
        padding: '16px 20px',
        borderRadius: '8px',
        boxShadow: '0 4px 12px rgba(0,0,0,0.15)',
        display: 'flex',
        alignItems: 'start',
        gap: '12px',
        minWidth: '320px',
        maxWidth: '420px',
        animation: 'slideIn 0.3s ease-out',
        marginBottom: '12px',
        position: 'relative' as const,
        overflow: 'hidden'
    }

    switch (type) {
        case 'success':
            return {
                ...baseStyles,
                backgroundColor: '#d1e7dd',
                color: '#0f5132',
                borderLeft: '4px solid #28a745'
            }
        case 'error':
            return {
                ...baseStyles,
                backgroundColor: '#f8d7da',
                color: '#842029',
                borderLeft: '4px solid #dc3545'
            }
        case 'warning':
            return {
                ...baseStyles,
                backgroundColor: '#fff3cd',
                color: '#856404',
                borderLeft: '4px solid #ffc107'
            }
        case 'info':
            return {
                ...baseStyles,
                backgroundColor: '#cfe2ff',
                color: '#084298',
                borderLeft: '4px solid #0d6efd'
            }
    }
}

const getIcon = (type: NotificationType) => {
    switch (type) {
        case 'success':
            return '✓'
        case 'error':
            return '✕'
        case 'warning':
            return '⚠'
        case 'info':
            return 'ℹ'
    }
}

function ToastItem({ notification }: { notification: Notification }) {
    const { removeNotification } = useNotification()
    const [isExiting, setIsExiting] = useState(false)

    const handleClose = () => {
        setIsExiting(true)
        setTimeout(() => {
            removeNotification(notification.id)
        }, 300)
    }

    return (
        <div
            style={{
                ...getNotificationStyles(notification.type),
                animation: isExiting ? 'slideOut 0.3s ease-in' : 'slideIn 0.3s ease-out'
            }}
        >
            <div style={{
                fontSize: '20px',
                fontWeight: 'bold',
                lineHeight: 1,
                flexShrink: 0
            }}>
                {getIcon(notification.type)}
            </div>

            <div style={{ flex: 1 }}>
                <div style={{
                    fontWeight: 600,
                    fontSize: '15px',
                    marginBottom: notification.message ? '4px' : 0
                }}>
                    {notification.title}
                </div>
                {notification.message && (
                    <div style={{
                        fontSize: '14px',
                        opacity: 0.9
                    }}>
                        {notification.message}
                    </div>
                )}
            </div>

            <button
                onClick={handleClose}
                style={{
                    background: 'none',
                    border: 'none',
                    fontSize: '20px',
                    cursor: 'pointer',
                    padding: 0,
                    color: 'inherit',
                    opacity: 0.6,
                    lineHeight: 1,
                    width: '20px',
                    height: '20px',
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center'
                }}
                onMouseEnter={(e) => {
                    e.currentTarget.style.opacity = '1'
                }}
                onMouseLeave={(e) => {
                    e.currentTarget.style.opacity = '0.6'
                }}
            >
                ×
            </button>
        </div>
    )
}

export default function ToastContainer() {
    const { notifications } = useNotification()

    return (
        <>
            <style>{`
        @keyframes slideIn {
          from {
            transform: translateX(400px);
            opacity: 0;
          }
          to {
            transform: translateX(0);
            opacity: 1;
          }
        }
        
        @keyframes slideOut {
          from {
            transform: translateX(0);
            opacity: 1;
          }
          to {
            transform: translateX(400px);
            opacity: 0;
          }
        }
      `}</style>

            <div style={{
                position: 'fixed',
                top: '24px',
                right: '24px',
                zIndex: 9999,
                display: 'flex',
                flexDirection: 'column',
                alignItems: 'flex-end',
                fontFamily: 'system-ui, -apple-system, sans-serif'
            }}>
                {notifications.map((notification) => (
                    <ToastItem key={notification.id} notification={notification} />
                ))}
            </div>
        </>
    )
}
