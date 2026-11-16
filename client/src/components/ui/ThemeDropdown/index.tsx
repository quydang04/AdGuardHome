import React, { useState, useRef, useEffect } from 'react';
import { useTranslation, Trans } from 'react-i18next';
import cn from 'classnames';

import { getTheme, setUITheme } from '../../../helpers/helpers';
import { THEMES } from '../../../helpers/constants';
import './ThemeDropdown.css';

const SunIcon = () => (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="5" />
        <line x1="12" y1="1" x2="12" y2="3" />
        <line x1="12" y1="21" x2="12" y2="23" />
        <line x1="4.22" y1="4.22" x2="5.64" y2="5.64" />
        <line x1="18.36" y1="18.36" x2="19.78" y2="19.78" />
        <line x1="1" y1="12" x2="3" y2="12" />
        <line x1="21" y1="12" x2="23" y2="12" />
        <line x1="4.22" y1="19.78" x2="5.64" y2="18.36" />
        <line x1="18.36" y1="5.64" x2="19.78" y2="4.22" />
    </svg>
);

const MoonIcon = () => (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M21 12.79A9 9 0 1 1 11.21 3 7 7 0 0 0 21 12.79z" />
    </svg>
);

const AutoIcon = () => (
    <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
);

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
        className={cn('theme-dropdown__chevron', { 'theme-dropdown__chevron--open': isOpen })}>
        <polyline points="6 9 12 15 18 9" />
    </svg>
);

export interface ThemeDropdownProps {
    direction?: 'up' | 'down';
    className?: string;
    currentTheme?: string;
    onSelect?: (theme: string) => void;
}

export const ThemeDropdown = ({ direction = 'down', className, currentTheme, onSelect }: ThemeDropdownProps) => {
    const { t } = useTranslation();
    const [isOpen, setIsOpen] = useState(false);
    const containerRef = useRef<HTMLDivElement>(null);

    const [activeTheme, setActiveTheme] = useState(() => currentTheme || getTheme());

    useEffect(() => {
        if (currentTheme) {
            setActiveTheme(currentTheme);
        }
    }, [currentTheme]);

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
                setIsOpen(false);
            }
        };

        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const handleSelect = (newTheme: string) => {
        setIsOpen(false);
        setActiveTheme(newTheme);
        setUITheme(newTheme);
        if (onSelect) {
            onSelect(newTheme);
        }
    };

    const getIcon = (th: string) => {
        if (th === THEMES.dark) return <MoonIcon />;
        if (th === THEMES.auto) return <AutoIcon />;
        return <SunIcon />;
    };

    const getLabel = (th: string) => {
        if (th === THEMES.dark) return t('theme_dark', 'Tối');
        if (th === THEMES.auto) return t('theme_auto', 'Tự động');
        return t('theme_light', 'Sáng');
    };

    return (
        <div
            ref={containerRef}
            className={cn('theme-dropdown', `theme-dropdown--${direction}`, className)}>
            <button
                type="button"
                className="theme-dropdown__btn"
                onClick={() => setIsOpen((prev) => !prev)}
                aria-label="Change theme"
                aria-expanded={isOpen}>
                <span className="theme-dropdown__icon">{getIcon(activeTheme)}</span>
                <span className="theme-dropdown__label">{getLabel(activeTheme)}</span>
                <ChevronDownIcon isOpen={isOpen} />
            </button>

            {isOpen && (
                <div className="theme-dropdown__menu" role="menu">
                    <button
                        type="button"
                        className={cn('theme-dropdown__item', {
                            'theme-dropdown__item--active': activeTheme === THEMES.light,
                        })}
                        onClick={() => handleSelect(THEMES.light)}
                        role="menuitem">
                        <SunIcon />
                        <span><Trans>theme_light</Trans></span>
                    </button>

                    <button
                        type="button"
                        className={cn('theme-dropdown__item', {
                            'theme-dropdown__item--active': activeTheme === THEMES.dark,
                        })}
                        onClick={() => handleSelect(THEMES.dark)}
                        role="menuitem">
                        <MoonIcon />
                        <span><Trans>theme_dark</Trans></span>
                    </button>

                    <button
                        type="button"
                        className={cn('theme-dropdown__item', {
                            'theme-dropdown__item--active': activeTheme === THEMES.auto,
                        })}
                        onClick={() => handleSelect(THEMES.auto)}
                        role="menuitem">
                        <AutoIcon />
                        <span><Trans>theme_auto</Trans></span>
                    </button>
                </div>
            )}
        </div>
    );
};

export default ThemeDropdown;
