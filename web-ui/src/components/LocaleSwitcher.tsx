"use client"
import { usePathname, useRouter } from 'next/navigation'
import { useI18n } from '../lib/i18n-context'

export default function LocaleSwitcher() {
    const { locale } = useI18n()
    const pathname = usePathname() || `/${locale}/login`
    const router = useRouter()
    function switchTo(target: 'tr' | 'en') {
        const parts = pathname.split('/').filter(Boolean)
        if (parts.length === 0) {
            router.push(`/${target}/login`)
            return
        }
        // First segment is locale
        parts[0] = target
        router.push('/' + parts.join('/'))
    }
    return (
        <div style={{ display: 'inline-flex', gap: 8 }}>
            <button disabled={locale === 'tr'} onClick={() => switchTo('tr')}>TR</button>
            <button disabled={locale === 'en'} onClick={() => switchTo('en')}>EN</button>
        </div>
    )
}
