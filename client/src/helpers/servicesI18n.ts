import i18next from 'i18next';
import { BASE_LOCALE } from './twosky';

type ServicesDict = Record<string, string>;

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
    if (cache[l] === undefined) {
        cache[l] = await loadLocaleFile(l);
    }
    const base = BASE_LOCALE.toLowerCase();
    if (cache[base] === undefined) {
        cache[base] = await loadLocaleFile(base);
    }
    if (cache.en === undefined) {
        cache.en = await loadLocaleFile('en');
    }
};
