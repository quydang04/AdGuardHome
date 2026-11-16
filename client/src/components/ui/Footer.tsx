import React from 'react';
import { useTranslation, Trans } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';

import { getVersion, changeTheme, changeLanguage } from '../../actions';
import { RootState } from '../../initialState';
import { LanguageDropdown } from './LanguageDropdown';
import { ThemeDropdown } from './ThemeDropdown';
import './Footer.css';

const RefreshIcon = ({ isSpinning }: { isSpinning?: boolean }) => (
    <svg
        width="13"
        height="13"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.4"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={isSpinning ? 'icon-spin' : ''}>
        <polyline points="23 4 23 10 17 10" />
        <polyline points="1 20 1 14 7 14" />
        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
    </svg>
);

const Footer = () => {
    const dispatch = useDispatch();
    const { t } = useTranslation();

    const dashboard = useSelector((state: RootState) => state?.dashboard) || ({} as any);
    const { theme, dnsVersion, processingVersion, checkUpdateFlag } = dashboard;
    const install = useSelector((state: RootState) => state?.install, shallowEqual);
    const version = dnsVersion || install?.dnsVersion;

    const handleSelectTheme = (newTheme: string) => {
        dispatch(changeTheme(newTheme));
    };

    const handleCheckUpdate = () => {
        dispatch(getVersion(true));
    };

    const currentYear = new Date().getFullYear();

    return (
        <footer className="app-footer">
            <div className="app-footer__container">
                {/* Left: Copyright */}
                <div className="app-footer__left">
                    <span className="app-footer__copyright">
                        Copyright &copy; {currentYear}{' '}
                        <a
                            target="_blank"
                            rel="noopener noreferrer"
                            href="https://link.adtidy.org/forward.html?action=home&from=ui&app=home">
                            AdGuard
                        </a>
                        {' '}&ndash; Customize by{' '}
                        <a
                            target="_blank"
                            rel="noopener noreferrer"
                            href="http://quydang.name.vn">
                            quydangblog
                        </a>
                    </span>
                </div>

                {/* Center: Version */}
                <div className="app-footer__center">
                    {version && (
                        <div className="app-footer__version-wrap">
                            <span>
                                <Trans>version</Trans> {version}
                            </span>
                            {checkUpdateFlag && (
                                <button
                                    type="button"
                                    className="app-footer__version-btn"
                                    onClick={handleCheckUpdate}
                                    disabled={processingVersion}
                                    title={t('check_updates_now', 'Check for updates')}>
                                    <RefreshIcon isSpinning={processingVersion} />
                                </button>
                            )}
                        </div>
                    )}
                </div>

                {/* Right: Theme & Language Switchers side-by-side */}
                <div className="app-footer__right">
                    <ThemeDropdown direction="up" currentTheme={theme} onSelect={handleSelectTheme} />
                    <LanguageDropdown direction="up" onSelect={(lang) => dispatch(changeLanguage(lang))} />
                </div>
            </div>
        </footer>
    );
};

export default Footer;
