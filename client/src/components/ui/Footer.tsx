import React from 'react';
import { useDispatch, useSelector } from 'react-redux';

import { changeTheme, changeLanguage } from '../../actions';
import { RootState } from '../../initialState';
import { LanguageDropdown } from './LanguageDropdown';
import { ThemeDropdown } from './ThemeDropdown';
import './Footer.css';

const Footer = () => {
    const dispatch = useDispatch();

    const dashboard = useSelector((state: RootState) => state?.dashboard) || ({} as any);
    const { theme } = dashboard;

    const handleSelectTheme = (newTheme: string) => {
        dispatch(changeTheme(newTheme));
    };

    return (
        <footer className="app-footer">
            <div className="app-footer__container">
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
