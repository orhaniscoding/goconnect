export type Locale = 'tr' | 'en';

export const locales: Locale[] = ['tr', 'en'];
export const defaultLocale: Locale = 'tr';

export async function getDictionary(locale: Locale) {
    switch (locale) {
        case 'en':
            return (await import('../locales/en/common.json')).default as Record<string, string>;
        case 'tr':
        default:
            return (await import('../locales/tr/common.json')).default as Record<string, string>;
    }
}
