'use client'
import { useState, useEffect } from 'react'
import { useRouter, useParams, usePathname } from 'next/navigation'
import { getUser, clearAuth } from '../lib/auth'
import { useT } from '../lib/i18n-context'

interface NavItem {
  key: string
  label: string
  icon: string
  href: string
  adminOnly?: boolean
}

export default function Navbar() {
  const router = useRouter()
  const params = useParams()
  const pathname = usePathname()
  const t = useT()
  const locale = (params.locale as string) || 'en'

  const [user, setUser] = useState<any>(null)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  useEffect(() => {
    const userData = getUser()
    setUser(userData)
  }, [])

  const navItems: NavItem[] = [
    { key: 'dashboard', label: t('nav.dashboard'), icon: '🏠', href: `/${locale}/dashboard` },
    { key: 'networks', label: t('nav.networks'), icon: '🌐', href: `/${locale}/networks` },
    { key: 'devices', label: t('nav.devices'), icon: '💻', href: `/${locale}/devices` },
    { key: 'tenants', label: t('nav.tenants'), icon: '👥', href: `/${locale}/tenants` },
    { key: 'profile', label: t('nav.profile'), icon: '👤', href: `/${locale}/profile` },
    { key: 'settings', label: t('nav.settings'), icon: '⚙️', href: `/${locale}/settings` },
    { key: 'admin', label: t('nav.admin'), icon: '👑', href: `/${locale}/admin`, adminOnly: true },
  ]

  const isActive = (href: string) => {
    return pathname.startsWith(href)
  }

  const handleLogout = () => {
    clearAuth()
    router.push(`/${locale}/login`)
  }

  const visibleNavItems = navItems.filter(item => !item.adminOnly || user?.is_admin)

  return (
    <nav style={{
      backgroundColor: '#1f2937',
      padding: '0 24px',
      position: 'sticky',
      top: 0,
      zIndex: 1000,
      boxShadow: '0 2px 4px rgba(0,0,0,0.1)'
    }}>
      <div style={{
        maxWidth: 1400,
        margin: '0 auto',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'space-between',
        height: 56
      }}>
        {/* Logo */}
        <div
          onClick={() => router.push(`/${locale}/dashboard`)}
          style={{
            display: 'flex',
            alignItems: 'center',
            gap: 8,
            cursor: 'pointer',
            color: 'white',
            fontWeight: 700,
            fontSize: 18
          }}
        >
          <span style={{ fontSize: 24 }}>🔐</span>
          <span>GoConnect</span>
        </div>

        {/* Desktop Navigation */}
        <div style={{
          display: 'flex',
          alignItems: 'center',
          gap: 4
        }}
          className="desktop-nav"
        >
          {visibleNavItems.map(item => (
            <button
              key={item.key}
              onClick={() => router.push(item.href)}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 6,
                padding: '8px 12px',
                backgroundColor: isActive(item.href) ? 'rgba(59, 130, 246, 0.2)' : 'transparent',
                color: isActive(item.href) ? '#60a5fa' : '#d1d5db',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 14,
                fontWeight: isActive(item.href) ? 600 : 400,
                transition: 'all 0.2s'
              }}
              onMouseEnter={(e) => {
                if (!isActive(item.href)) {
                  e.currentTarget.style.backgroundColor = 'rgba(255,255,255,0.1)'
                  e.currentTarget.style.color = 'white'
                }
              }}
              onMouseLeave={(e) => {
                if (!isActive(item.href)) {
                  e.currentTarget.style.backgroundColor = 'transparent'
                  e.currentTarget.style.color = '#d1d5db'
                }
              }}
            >
              <span>{item.icon}</span>
              <span className="nav-label">{item.label}</span>
            </button>
          ))}
        </div>

        {/* User Menu */}
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          {user && (
            <span style={{ color: '#9ca3af', fontSize: 13 }}>
              {user.name || user.email}
            </span>
          )}
          <button
            onClick={handleLogout}
            style={{
              padding: '6px 12px',
              backgroundColor: '#dc2626',
              color: 'white',
              border: 'none',
              borderRadius: 6,
              cursor: 'pointer',
              fontSize: 13,
              fontWeight: 500,
              display: 'flex',
              alignItems: 'center',
              gap: 4
            }}
          >
            <span>🚪</span>
            <span className="nav-label">{t('nav.logout')}</span>
          </button>

          {/* Mobile Menu Toggle */}
          <button
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            style={{
              display: 'none',
              padding: 8,
              backgroundColor: 'transparent',
              color: 'white',
              border: 'none',
              cursor: 'pointer',
              fontSize: 20
            }}
            className="mobile-menu-btn"
          >
            {mobileMenuOpen ? '✕' : '☰'}
          </button>
        </div>
      </div>

      {/* Mobile Menu */}
      {mobileMenuOpen && (
        <div style={{
          backgroundColor: '#374151',
          padding: 16,
          borderTop: '1px solid #4b5563'
        }}
          className="mobile-menu"
        >
          {visibleNavItems.map(item => (
            <button
              key={item.key}
              onClick={() => {
                router.push(item.href)
                setMobileMenuOpen(false)
              }}
              style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                width: '100%',
                padding: '12px 16px',
                backgroundColor: isActive(item.href) ? 'rgba(59, 130, 246, 0.2)' : 'transparent',
                color: isActive(item.href) ? '#60a5fa' : '#d1d5db',
                border: 'none',
                borderRadius: 6,
                cursor: 'pointer',
                fontSize: 15,
                fontWeight: isActive(item.href) ? 600 : 400,
                textAlign: 'left',
                marginBottom: 4
              }}
            >
              <span>{item.icon}</span>
              <span>{item.label}</span>
            </button>
          ))}
        </div>
      )}

      <style jsx global>{`
        @media (max-width: 768px) {
          .desktop-nav {
            display: none !important;
          }
          .mobile-menu-btn {
            display: block !important;
          }
          .nav-label {
            display: none;
          }
        }
        @media (min-width: 769px) {
          .mobile-menu {
            display: none !important;
          }
        }
      `}</style>
    </nav>
  )
}
