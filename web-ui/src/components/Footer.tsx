"use client"
import { useI18n, useT } from '../lib/i18n-context'
import { useDaemon } from '../contexts/DaemonContext'
import LocaleSwitcher from './LocaleSwitcher'

export default function Footer() {
  const { dict } = useI18n()
  const t = useT()
  const { status, error } = useDaemon()

  return (
    <footer style={{ padding: 16, borderTop: '1px solid #eee', marginTop: 24, fontSize: 12, opacity: 0.8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
        <span>{dict['footer.brand']}</span>
        {status ? (
          <span style={{ color: 'green', display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: 'green' }}></span>
            Daemon v{status.version}
          </span>
        ) : (
          <span style={{ color: 'gray', display: 'flex', alignItems: 'center', gap: 4 }}>
            <span style={{ width: 8, height: 8, borderRadius: '50%', backgroundColor: 'gray' }}></span>
            Daemon: {error ? 'Disconnected' : 'Connecting...'}
          </span>
        )}
      </div>
      <div aria-label={t('footer.locale', 'Locale')}><LocaleSwitcher /></div>
    </footer>
  )
}
