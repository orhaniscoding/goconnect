"use client"
import { useI18n, useT } from '../lib/i18n-context'
import LocaleSwitcher from './LocaleSwitcher'

export default function Footer() {
  const { dict } = useI18n()
  const t = useT()
  return (
    <footer style={{ padding: 16, borderTop: '1px solid #eee', marginTop: 24, fontSize: 12, opacity: 0.8, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
      <span>{dict['footer.brand']}</span>
      <div aria-label={t('footer.locale', 'Locale')}><LocaleSwitcher /></div>
    </footer>
  )
}
