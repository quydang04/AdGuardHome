import i18next from 'i18next';
import { BASE_LOCALE } from './twosky';

type ServicesDict = Record<string, string>;
type TranslationObject = { message: string } | string;

const cache: Record<string, ServicesDict | null> = {};

const loadLocaleFile = async (lang: string): Promise<ServicesDict | null> => {
    try {
        const mod = await import(`../__locales-services/${lang}.json`);
        return (mod as any).default ?? (mod as any);
    } catch (_) {
        return null;
    }
};

export const preloadServicesLocale = async (lang?: string): Promise<void> => {
    const l = (lang || i18next.language || BASE_LOCALE).toLowerCase();
    if (!cache[l]) {
        const result = await loadLocaleFile(l);
        if (result) cache[l] = result;
    }
    const base = BASE_LOCALE.toLowerCase();
    if (!cache[base]) {
        const result = await loadLocaleFile(base);
        if (result) cache[base] = result;
    }
    if (!cache.en) {
        const result = await loadLocaleFile('en');
        if (result) cache.en = result;
    }
};

/**
 * Helper function to extract the message string from service translation objects.
 */
export const getServiceTranslation = (t: (key: string) => any, key: string): string => {
    const currentLang = i18next.language?.toLowerCase() || 'en';
    const fallbackLangs = ['en', BASE_LOCALE.toLowerCase(), currentLang].filter((lang, idx, arr) => arr.indexOf(lang) === idx);
        
    for (const lang of fallbackLangs) {
        const servicesDict = cache[lang];
        console.log('Checking cache for lang:', { lang, hasDict: !!servicesDict, hasKey: servicesDict?.[key] });
        if (servicesDict && servicesDict[key]) {
            const translation = servicesDict[key] as TranslationObject;
            console.log('Found in cache:', { lang, key, translation });
            if (typeof translation === 'string') {
                return translation;
            }
            if (translation && typeof translation === 'object' && 'message' in translation) {
                return translation.message;
            }
        }
    }
    
    const translation = t(key) as TranslationObject;
    
    if (typeof translation === 'string') {
        return translation;
    }
    
    if (translation && typeof translation === 'object' && 'message' in translation) {
        return translation.message;
    }
    
    return key;
};
