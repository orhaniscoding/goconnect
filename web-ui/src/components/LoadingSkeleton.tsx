'use client'
import React from 'react'

type LoadingSkeletonProps = {
  width?: number | string
  height?: number
  radius?: number
  shimmer?: boolean
  className?: string
  style?: React.CSSProperties
}

/**
 * Lightweight skeleton block with optional shimmer for loading states.
 */
export default function LoadingSkeleton({
  width = '100%',
  height = 14,
  radius = 8,
  shimmer = true,
  className,
  style
}: LoadingSkeletonProps) {
  return (
    <div
      className={className}
      style={{
        position: 'relative',
        overflow: 'hidden',
        backgroundColor: '#e5e7eb',
        borderRadius: radius,
        width,
        height,
        ...style
      }}
    >
      {shimmer && (
        <div
          style={{
            position: 'absolute',
            inset: 0,
            backgroundImage: 'linear-gradient(90deg, #e5e7eb 0%, #f3f4f6 50%, #e5e7eb 100%)',
            transform: 'translateX(-100%)',
            animation: 'skeleton-shimmer 1.4s ease-in-out infinite'
          }}
        />
      )}

      <style jsx>{`
        @keyframes skeleton-shimmer {
          0% {
            transform: translateX(-100%);
          }
          100% {
            transform: translateX(100%);
          }
        }
      `}</style>
    </div>
  )
}
