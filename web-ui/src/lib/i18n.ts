export type Locale = 'tr' | 'en';

export const locales: Locale[] = ['tr', 'en'];
export const defaultLocale: Locale = 'tr';

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type Dictionary = Record<string, any>;

export async function getDictionary(locale: Locale): Promise<Dictionary> {
    switch (locale) {
        case 'en':
            return (await import('../locales/en/common.json')).default as Dictionary;
        case 'tr':
        default:
            return (await import('../locales/tr/common.json')).default as Dictionary;
    }
}
