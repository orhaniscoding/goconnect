"use client"
import React, { createContext, useContext } from 'react'
import type { Locale } from './i18n'

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Dictionary = Record<string, any>

type I18nValue = { locale: Locale; dict: Dictionary }

const Ctx = createContext<I18nValue | null>(null)

export function I18nProvider({ locale, dict, children }: { locale: Locale; dict: Dictionary; children: React.ReactNode }) {
    return <Ctx.Provider value={{ locale, dict }}>{children}</Ctx.Provider>
}

export function useI18n() {
    const v = useContext(Ctx)
    if (!v) throw new Error('I18nProvider missing in tree')
    return v
}

/**
 * Translation hook with variable interpolation support
 * Usage:
 *   t('key') - simple translation
 *   t('key', { name: 'value' }) - with interpolation: "Hello {name}" -> "Hello value"
 *   t('key', 'fallback') - with fallback string
 */
export function useT() {
    const { dict } = useI18n()
    return (key: string, varsOrFallback?: Record<string, string | number> | string): string => {
        // Get nested keys like 'networks.title' -> dict.networks.title
        const parts = key.split('.')
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        let value: any = dict
        for (const part of parts) {
            if (value && typeof value === 'object' && part in value) {
                value = value[part]
            } else {
                // If not found, try flat key
                value = dict[key]
                break
            }
        }

        // If value is not a string, use fallback or key
        if (typeof value !== 'string') {
            if (typeof varsOrFallback === 'string') {
                return varsOrFallback
            }
            return key
        }

        // If varsOrFallback is string, it's a fallback (no interpolation needed)
        if (typeof varsOrFallback === 'string') {
            return value || varsOrFallback
        }

        // Interpolate variables: {varName} -> value
        if (varsOrFallback && typeof varsOrFallback === 'object') {
            return value.replace(/\{(\w+)\}/g, (_, varName) => {
                return String(varsOrFallback[varName] ?? `{${varName}}`)
            })
        }

        return value
    }
}
