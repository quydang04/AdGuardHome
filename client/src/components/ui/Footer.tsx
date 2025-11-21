import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useDispatch, useSelector } from 'react-redux';
import cn from 'classnames';

import { REPOSITORY, PRIVACY_POLICY_LINK, THEMES } from '../../helpers/constants';
import { LANGUAGES } from '../../helpers/twosky';
import i18n from '../../i18n';

import Version from './Version';
import './Footer.css';
import './Select.css';

import { setHtmlLangAttr, setUITheme } from '../../helpers/helpers';

import { changeLanguage, changeTheme } from '../../actions';
import { RootState } from '../../initialState';

type ThemeName = keyof typeof THEMES;

const linksData = [
    {
        href: REPOSITORY.URL,
        name: 'homepage',
    },
    {
        href: PRIVACY_POLICY_LINK,
        name: 'privacy_policy',
    },
    {
        href: REPOSITORY.ISSUES,
        className: 'btn btn-outline-primary btn-sm footer__link--report',
        name: 'report_an_issue',
    },
];

const Footer = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();

    const currentTheme = useSelector((state: RootState) => (state.dashboard ? state.dashboard.theme : THEMES.auto));
    const profileName = useSelector((state: RootState) => (state.dashboard ? state.dashboard.name : ''));
    const isLoggedIn = profileName !== '';
    const [currentThemeLocal, setCurrentThemeLocal] = useState(THEMES.auto);
    const languageOptions = Object.keys(LANGUAGES);

    const getYear = () => {
        const today = new Date();
        return today.getFullYear();
    };

    const onLanguageChange = (language: string) => {
        i18n.changeLanguage(language);
        setHtmlLangAttr(language);

        if (isLoggedIn) {
            dispatch(changeLanguage(language));
        }
    };

    const onThemeChange = (value: any) => {
        if (isLoggedIn) {
            dispatch(changeTheme(value));
        } else {
            setUITheme(value);
            setCurrentThemeLocal(value);
        }
    };

    const renderCopyright = () => (
        <div className="footer__column">
            <div className="footer__copyright">
                {t('copyright')} &copy; {getYear()}{' '}
                <a
                    target="_blank"
                    rel="noopener noreferrer"
                    href="https://link.adtidy.org/forward.html?action=home&from=ui&app=home">
                    AdGuard
                </a>
                <span className="footer__custom"> - Customize by quydangnet</span>
            </div>
        </div>
    );

    const renderLinks = (linksData: any) =>
        linksData.map(({ name, href, className = '' }: any) => (
            <a
                key={name}
                href={href}
                className={cn('footer__link', className)}
                target="_blank"
                rel="noopener noreferrer">
                {t(name)}
            </a>
        ));

    const renderThemeButtons = () => {
        const currentValue = isLoggedIn ? currentTheme : currentThemeLocal;

        const icons: Record<ThemeName, (className: string) => JSX.Element> = {
            auto: (className) => (
                <svg
                    className={className}
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                    focusable="false">
                    <path
                        fillRule="evenodd"
                        clipRule="evenodd"
                        d="M12 3C16.9706 3 21 7.02944 21 12C21 16.9706 16.9706 21 12 21C7.02944 21 3 16.9706 3 12C3 7.02944 7.02944 3 12 3Z"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                    <path
                        fillRule="evenodd"
                        clipRule="evenodd"
                        d="M12 3V21C16.9706 21 21 16.9706 21 12C21 7.02944 16.9706 3 12 3Z"
                        fill="currentColor"
                        stroke="currentColor"
                        strokeWidth="1.5"
                    />
                </svg>
            ),
            dark: (className) => (
                <svg
                    className={className}
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                    focusable="false">
                    <path
                        d="M3.80737 15.731L3.9895 15.0034C3.71002 14.9335 3.41517 15.0298 3.23088 15.2512C3.0466 15.4727 3.00545 15.7801 3.12501 16.0422L3.80737 15.731ZM14.1926 3.26892L14.3747 2.54137C14.0953 2.47141 13.8004 2.56772 13.6161 2.78917C13.4318 3.01062 13.3907 3.31806 13.5102 3.58018L14.1926 3.26892ZM12 20.2499C8.66479 20.2499 5.79026 18.2708 4.48974 15.4197L3.12501 16.0422C4.66034 19.4081 8.05588 21.7499 12 21.7499V20.2499ZM20.25 11.9999C20.25 16.5563 16.5563 20.2499 12 20.2499V21.7499C17.3848 21.7499 21.75 17.3847 21.75 11.9999H20.25ZM14.0105 3.99647C17.5955 4.89391 20.25 8.13787 20.25 11.9999H21.75C21.75 7.43347 18.6114 3.60193 14.3747 2.54137L14.0105 3.99647ZM13.5102 3.58018C13.9851 4.6211 14.25 5.77857 14.25 6.99995H15.75C15.75 5.5595 15.4371 4.1901 14.875 2.95766L13.5102 3.58018ZM14.25 6.99995C14.25 11.5563 10.5563 15.2499 5.99999 15.2499V16.7499C11.3848 16.7499 15.75 12.3847 15.75 6.99995H14.25ZM5.99999 15.2499C5.30559 15.2499 4.63225 15.1643 3.9895 15.0034L3.62525 16.4585C4.38616 16.649 5.18181 16.7499 5.99999 16.7499V15.2499Z"
                        fill="currentColor"
                    />
                </svg>
            ),
            light: (className) => (
                <svg
                    className={className}
                    viewBox="0 0 24 24"
                    fill="none"
                    aria-hidden="true"
                    focusable="false">
                    <path
                        d="M12 3.75C16.5563 3.75 20.25 7.44365 20.25 12H21.75C21.75 6.61522 17.3848 2.25 12 2.25V3.75ZM20.25 12C20.25 16.5563 16.5563 20.25 12 20.25V21.75C17.3848 21.75 21.75 17.3848 21.75 12H20.25ZM12 20.25C7.44365 20.25 3.75 16.5563 3.75 12H2.25C2.25 17.3848 6.61522 21.75 12 21.75V20.25ZM3.75 12C3.75 7.44365 7.44365 3.75 12 3.75V2.25C6.61522 2.25 2.25 6.61522 2.25 12H3.75Z"
                        fill="currentColor"
                    />
                    <path
                        fillRule="evenodd"
                        clipRule="evenodd"
                        d="M12 10C10.8954 10 10 10.8954 10 12C10 13.1046 10.8954 14 12 14C13.1046 14 14 13.1046 14 12C13.9987 10.896 13.104 10.0013 12 10Z"
                        fill="currentColor"
                    />
                </svg>
            ),
        };

        const content: Record<ThemeName, { desc: string; testId: string }> = {
            auto: {
                desc: t('theme_auto_desc'),
                testId: 'theme_auto',
            },
            dark: {
                desc: t('theme_dark_desc'),
                testId: 'theme_dark',
            },
            light: {
                desc: t('theme_light_desc'),
                testId: 'theme_light',
            },
        };

        return (Object.values(THEMES) as ThemeName[]).map((theme) => (
            <button
                key={theme}
                type="button"
                className="btn btn-sm btn-secondary footer__theme-button"
                onClick={() => onThemeChange(theme)}
                title={content[theme].desc}
                data-testid={content[theme].testId}>
                {icons[theme](cn('footer__theme-icon', { 'footer__theme-icon--active': currentValue === theme }))}
            </button>
        ));
    };

    return (
        <footer className="footer">
            <div className="container">
                <div className="footer__row">
                    <div className="footer__column footer__column--links">{renderLinks(linksData)}</div>

                    <div className="footer__column footer__column--language">
                        <div className="footer__themes footer__language-buttons">
                            {renderThemeButtons()}
                            {languageOptions.length > 1 &&
                                languageOptions.map((lang) => {
                                    const active = i18n.language === lang;
                                    return (
                                        <button
                                            key={lang}
                                            type="button"
                                            className={cn('btn btn-sm', {
                                                'btn-secondary': !active,
                                                'btn-primary': active,
                                            })}
                                            onClick={() => onLanguageChange(lang)}>
                                            {LANGUAGES[lang]}
                                        </button>
                                    );
                                })}
                        </div>
                    </div>
                </div>

                <div className="footer__row footer__row--meta">
                    {renderCopyright()}

                    <div className="footer__column footer__column--language">
                        <Version />
                    </div>
                </div>
            </div>
        </footer>
    );
};

export default Footer;
