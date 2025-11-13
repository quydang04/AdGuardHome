import i18n from 'i18next';
import { initReactI18next } from 'react-i18next';
import langDetect from 'i18next-browser-languagedetector';

import { LANGUAGES, BASE_LOCALE } from './helpers/twosky';

import ar from './__locales/ar.json';
import be from './__locales/be.json';
import bg from './__locales/bg.json';
import cs from './__locales/cs.json';
import da from './__locales/da.json';
import de from './__locales/de.json';
import en from './__locales/en.json';
import es from './__locales/es.json';
import fa from './__locales/fa.json';
import fi from './__locales/fi.json';
import fr from './__locales/fr.json';
import hr from './__locales/hr.json';
import hu from './__locales/hu.json';
import id from './__locales/id.json';
import it from './__locales/it.json';
import ja from './__locales/ja.json';
import ko from './__locales/ko.json';
import nl from './__locales/nl.json';
import no from './__locales/no.json';
import pl from './__locales/pl.json';
import ptBR from './__locales/pt-br.json';
import ptPT from './__locales/pt-pt.json';
import ro from './__locales/ro.json';
import ru from './__locales/ru.json';
import siLk from './__locales/si-lk.json';
import sk from './__locales/sk.json';
import sl from './__locales/sl.json';
import srCS from './__locales/sr-cs.json';
import sv from './__locales/sv.json';
import th from './__locales/th.json';
import tr from './__locales/tr.json';
import uk from './__locales/uk.json';
import vi from './__locales/vi.json';
import zhCN from './__locales/zh-cn.json';
import zhHK from './__locales/zh-hk.json';
import zhTW from './__locales/zh-tw.json';

import { setHtmlLangAttr } from './helpers/helpers';

const resources = {
    ar: { translation: ar },
    be: { translation: be },
    bg: { translation: bg },
    cs: { translation: cs },
    da: { translation: da },
    de: { translation: de },
    en: { translation: en },
    'en-us': { translation: en },
    es: { translation: es },
    fa: { translation: fa },
    fi: { translation: fi },
    fr: { translation: fr },
    hr: { translation: hr },
    hu: { translation: hu },
    id: { translation: id },
    it: { translation: it },
    ja: { translation: ja },
    ko: { translation: ko },
    nl: { translation: nl },
    no: { translation: no },
    pl: { translation: pl },
    'pt-br': { translation: ptBR },
    'pt-pt': { translation: ptPT },
    ro: { translation: ro },
    ru: { translation: ru },
    'si-lk': { translation: siLk },
    sk: { translation: sk },
    sl: { translation: sl },
    'sr-cs': { translation: srCS },
    sv: { translation: sv },
    th: { translation: th },
    tr: { translation: tr },
    uk: { translation: uk },
    vi: { translation: vi },
    'zh-cn': { translation: zhCN },
    'zh-hk': { translation: zhHK },
    'zh-tw': { translation: zhTW },
};

const availableLanguages = Object.keys(LANGUAGES);

const loadServicesTranslations = async (lng: string) => {
    const currentLocale = lng.toLowerCase();
    const baseLocale = BASE_LOCALE.toLowerCase();

    const orderedLocales = ['en', baseLocale, currentLocale]
        .filter(Boolean)
        .filter((locale, idx, arr) => arr.indexOf(locale) === idx);

    const localeDictionaries = await Promise.all(
        orderedLocales.map(async (locale) => {
            try {
                const mod = await import(`./__locales-services/${locale}.json`);
                const dictionary = (mod as any).default ?? (mod as any);
                return dictionary && typeof dictionary === 'object'
                    ? dictionary
                    : null;
            } catch {
                return null;
            }
        })
    );

    const mergedDictionary = localeDictionaries.reduce<Record<string, any>>(
        (acc, dictionary) => {
            if (dictionary) Object.assign(acc, dictionary);
            return acc;
        },
        {}
    );

    if (Object.keys(mergedDictionary).length > 0) {
        i18n.addResourceBundle(currentLocale, 'translation', mergedDictionary, true, true);
        await i18n.reloadResources([currentLocale], ['translation']);
        (i18n as any).emit?.('loaded', { lng: currentLocale, ns: 'translation' });
    }
};

i18n
    .use(langDetect)
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
                bindI18n: 'languageChanged loaded',
            },
            whitelist: availableLanguages,
        },
        () => {
            if (!availableLanguages.includes(i18n.language)) {
                i18n.changeLanguage(BASE_LOCALE);
            }
            setHtmlLangAttr(i18n.language);

            loadServicesTranslations(i18n.language).catch(() => {});
        }
    );

i18n.on('languageChanged', (lng) => {
    loadServicesTranslations(lng).catch(() => {});
});

export default i18n;
