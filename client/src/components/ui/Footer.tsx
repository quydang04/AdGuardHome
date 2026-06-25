import React from 'react';
import { useTranslation } from 'react-i18next';

import i18n from '../../i18n';

import Version from './Version';
import './Footer.css';

const Footer = () => {
    const { t } = useTranslation();

    const getYear = () => {
        const today = new Date();
        return today.getFullYear();
    };

    const customizeText = i18n.language === 'vi' ? ' - Tùy chỉnh bởi ' : ' - Customize by ';

    return (
        <footer className="footer">
            <div className="container">
                <div className="footer__row">
                    <div className="footer__column">
                        <div className="footer__copyright">
                            {t('copyright')} &copy; {getYear()}{' '}
                            <a
                                target="_blank"
                                rel="noopener noreferrer"
                                href="https://link.adtidy.org/forward.html?action=home&from=ui&app=home">
                                AdGuard
                            </a>
                            {customizeText}
                            <a
                                target="_blank"
                                rel="noopener noreferrer"
                                href="http://quydang.name.vn">
                                quydangblog
                            </a>
                        </div>
                    </div>

                    <div className="footer__column footer__column--version">
                        <Version />
                    </div>
                </div>
            </div>
        </footer>
    );
};

export default Footer;
