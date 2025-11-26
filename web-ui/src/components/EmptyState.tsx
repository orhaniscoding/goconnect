'use client'
import React from 'react'

type Action = {
  label: string
  onClick: () => void
}

type EmptyStateProps = {
  title: string
  description?: string
  icon?: React.ReactNode
  primaryAction?: Action
  secondaryAction?: Action
  children?: React.ReactNode
}

/**
 * Simple empty state card used across pages when there is no data.
 */
export default function EmptyState({
  title,
  description,
  icon,
  primaryAction,
  secondaryAction,
  children
}: EmptyStateProps) {
  return (
    <div
      style={{
        border: '1px dashed #d1d5db',
        backgroundColor: '#f9fafb',
        borderRadius: 12,
        padding: 24,
        textAlign: 'center',
        color: '#4b5563'
      }}
    >
      {icon && (
        <div style={{ fontSize: 28, marginBottom: 8, color: '#9ca3af' }}>
          {icon}
        </div>
      )}
      <h3 style={{ margin: '0 0 8px 0', color: '#111827' }}>{title}</h3>
      {description && (
        <p style={{ margin: '0 0 12px 0', fontSize: 14 }}>{description}</p>
      )}

      {children}

      {(primaryAction || secondaryAction) && (
        <div style={{ marginTop: 12, display: 'flex', gap: 8, justifyContent: 'center', flexWrap: 'wrap' }}>
          {primaryAction && (
            <button
              onClick={primaryAction.onClick}
              style={{
                padding: '8px 14px',
                backgroundColor: '#2563eb',
                color: '#fff',
                border: 'none',
                borderRadius: 8,
                cursor: 'pointer',
                fontWeight: 600
              }}
            >
              {primaryAction.label}
            </button>
          )}
          {secondaryAction && (
            <button
              onClick={secondaryAction.onClick}
              style={{
                padding: '8px 14px',
                backgroundColor: '#e5e7eb',
                color: '#111827',
                border: '1px solid #d1d5db',
                borderRadius: 8,
                cursor: 'pointer',
                fontWeight: 600
              }}
            >
              {secondaryAction.label}
            </button>
          )}
        </div>
      )}
    </div>
  )
}
