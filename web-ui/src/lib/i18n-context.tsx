"use client"
import React, { createContext, useContext } from 'react'
import type { Locale } from './i18n'

type I18nValue = { locale: Locale; dict: Record<string, string> }

const Ctx = createContext<I18nValue | null>(null)

export function I18nProvider({ locale, dict, children }: { locale: Locale; dict: Record<string, string>; children: React.ReactNode }) {
    return <Ctx.Provider value={{ locale, dict }}>{children}</Ctx.Provider>
}

export function useI18n() {
    const v = useContext(Ctx)
    if (!v) throw new Error('I18nProvider missing in tree')
    return v
}

export function useT() {
    const { dict } = useI18n()
    return (key: string, fallback?: string) => dict[key] ?? fallback ?? key
}
