'use client'
import { useState, useEffect } from 'react'
import { useRouter, useParams, usePathname } from 'next/navigation'
import { getUser, clearAuth } from '../lib/auth'
import { useT } from '../lib/i18n-context'
import { useDaemon } from '@/contexts/DaemonContext'

interface NavItem {
  key: string
  label: string
  icon: string
  href: string
  adminOnly?: boolean
}

const ICONS = {
  dashboard: '📊',
  networks: '🌐',
  devices: '💻',
  tenants: '👥',
  profile: '🙍',
  settings: '⚙️',
  admin: '🛡️',
  brand: '🌀',
  logout: '↩️',
  menuOpen: '✖️',
  menuClosed: '☰',
}

export default function Navbar() {
  const router = useRouter()
  const params = useParams()
  const pathname = usePathname()
  const t = useT()
  const locale = (params.locale as string) || 'en'
  const { status: daemonStatus, loading: daemonLoading } = useDaemon()

  const [user, setUser] = useState<any>(null)
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false)

  useEffect(() => {
    const userData = getUser()
    setUser(userData)
  }, [])

  const navItems: NavItem[] = [
    { key: 'dashboard', label: t('nav.dashboard'), icon: ICONS.dashboard, href: `/${locale}/dashboard` },
    { key: 'networks', label: t('nav.networks'), icon: ICONS.networks, href: `/${locale}/networks` },
    { key: 'devices', label: t('nav.devices'), icon: ICONS.devices, href: `/${locale}/devices` },
    { key: 'tenants', label: t('nav.tenants'), icon: ICONS.tenants, href: `/${locale}/tenants` },
    { key: 'profile', label: t('nav.profile'), icon: ICONS.profile, href: `/${locale}/profile` },
    { key: 'settings', label: t('nav.settings'), icon: ICONS.settings, href: `/${locale}/settings` },
    { key: 'admin', label: t('nav.admin'), icon: ICONS.admin, href: `/${locale}/admin`, adminOnly: true },
  ]

  const isActive = (href: string) => pathname.startsWith(href)

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
          <span style={{ fontSize: 24 }}>{ICONS.brand}</span>
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
          {/* Daemon Status Indicator */}
          <div 
            title={daemonStatus ? `Daemon v${daemonStatus.version}` : "Daemon not detected"}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 6,
              padding: '4px 8px',
              backgroundColor: 'rgba(0,0,0,0.2)',
              borderRadius: 12,
              fontSize: 12
            }}
          >
            <div style={{
              width: 8,
              height: 8,
              borderRadius: '50%',
              backgroundColor: daemonLoading ? '#fca5a5' : (daemonStatus ? '#4ade80' : '#ef4444'),
              boxShadow: daemonStatus ? '0 0 8px #4ade80' : 'none'
            }} />
            <span style={{ color: '#d1d5db', fontSize: 11 }}>
              {daemonLoading ? t('footer.daemon.connecting') : (daemonStatus ? "Daemon Active" : t('footer.daemon.disconnected'))}
            </span>
          </div>

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
            <span>{ICONS.logout}</span>
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
            {mobileMenuOpen ? ICONS.menuOpen : ICONS.menuClosed}
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