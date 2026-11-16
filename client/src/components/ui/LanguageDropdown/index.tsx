import React, { useState, useRef, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import cn from 'classnames';

import { setHtmlLangAttr } from '../../../helpers/helpers';
import i18n from '../../../i18n';
import './LanguageDropdown.css';

const ChevronDownIcon = ({ isOpen }: { isOpen: boolean }) => (
    <svg
        width="12"
        height="12"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={cn('lang-dropdown__chevron', { 'lang-dropdown__chevron--open': isOpen })}>
        <polyline points="6 9 12 15 18 9" />
    </svg>
);

const LANGUAGES = [
    { code: 'en', label: 'English', codeUpper: 'EN' },
    { code: 'vi', label: 'Tiếng Việt', codeUpper: 'VI' },
];

export interface LanguageDropdownProps {
    direction?: 'up' | 'down';
    className?: string;
    onSelect?: (lang: string) => void;
}

export const LanguageDropdown = ({ direction = 'down', className, onSelect }: LanguageDropdownProps) => {
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);
    const { i18n: i18nInstance } = useTranslation();

    const currentLang = (i18nInstance.language || i18n.language || 'vi').toLowerCase();
    const activeLangObj = LANGUAGES.find((l) => l.code === currentLang) || (currentLang.startsWith('en') ? LANGUAGES[0] : LANGUAGES[1]);

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleSelect = (code: string) => {
        setIsOpen(false);
        i18n.changeLanguage(code);
        setHtmlLangAttr(code);
        try {
            localStorage.setItem('i18nextLng', code);
        } catch {}
        if (onSelect) {
            onSelect(code);
        }
    };

    return (
        <div
            ref={containerRef}
            className={cn('lang-dropdown', `lang-dropdown--${direction}`, className)}>
            <button
                type="button"
                className="lang-dropdown__btn"
                onClick={() => setIsOpen((prev) => !prev)}
                aria-label="Change language"
                aria-expanded={isOpen}>
                <span className="lang-dropdown__glyph">A文</span>
                <span className="lang-dropdown__code">{activeLangObj.codeUpper}</span>
                <ChevronDownIcon isOpen={isOpen} />
            </button>

            {isOpen && (
                <div className="lang-dropdown__menu" role="menu">
                    {LANGUAGES.map(({ code, label }) => {
                        const isActive = activeLangObj.code === code;
                        return (
                            <button
                                key={code}
                                type="button"
                                className={cn('lang-dropdown__item', {
                                    'lang-dropdown__item--active': isActive,
                                })}
                                onClick={() => handleSelect(code)}
                                role="menuitem">
                                {label}
                            </button>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default LanguageDropdown;
