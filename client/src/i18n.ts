import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import langDetect from 'i18next-browser-languagedetector';

import { LANGUAGES, BASE_LOCALE } from './helpers/twosky';
import vi from './__locales/vi.json';
import en from './__locales/en.json';

import { setHtmlLangAttr } from './helpers/helpers';

const resources = {
    vi: { translation: vi },
    en: { translation: en },
    'en-us': { translation: en },
};

const availableLanguages = Object.keys(LANGUAGES);

i18n.use(langDetect)
    .use(initReactI18next)
    .init(
        {
            resources,
            lowerCaseLng: true,
            fallbackLng: BASE_LOCALE,
            keySeparator: false,
            nsSeparator: false,
            returnEmptyString: false,
            interpolation: {
                escapeValue: false,
            },
            react: {
                wait: true,
            },
            whitelist: availableLanguages,
            supportedLngs: availableLanguages,
            detection: {
                order: ['localStorage', 'navigator'],
                lookupLocalStorage: 'i18nextLng',
                caches: ['localStorage'],
            },
        },
        () => {
            if (!availableLanguages.includes(i18n.language)) {
                i18n.changeLanguage(BASE_LOCALE);
            }
            setHtmlLangAttr(i18n.language);
        },
    );

export default i18n;
