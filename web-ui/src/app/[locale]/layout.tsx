import type { ReactNode } from 'react'
import { locales, defaultLocale, type Locale, getDictionary } from '../../lib/i18n'
import { I18nProvider } from '../../lib/i18n-context'

export const dynamic = 'force-static'

export async function generateStaticParams() {
    return locales.map((l) => ({ locale: l }))
}

export async function generateMetadata({ params }: { params: Promise<{ locale: string }> }) {
    const { locale } = await params
    const l: Locale = locale === 'en' ? 'en' : 'tr'
    return { title: 'GoConnect', description: 'Secure virtual network & chat — orhaniscoding' }
}

export default async function LocaleLayout({ children, params }: { children: ReactNode; params: Promise<{ locale: string }> }) {
    const { locale } = await params
    const l: Locale = locale === 'en' ? 'en' : 'tr'
    const dict = await getDictionary(l)
    return <I18nProvider locale={l} dict={dict}>{children}</I18nProvider>
}
