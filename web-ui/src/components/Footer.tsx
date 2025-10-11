"use client"
import { useI18n } from '../lib/i18n-context'

export default function Footer() {
  const { dict } = useI18n()
  return (
    <footer style={{ padding: 16, borderTop: '1px solid #eee', marginTop: 24, fontSize: 12, opacity: 0.8 }}>
      {dict['footer.brand']}
    </footer>
  )
}
