import React from 'react';
import { Link } from 'react-router-dom';

const HamburgerIcon = () => (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="3" y1="12" x2="21" y2="12" />
        <line x1="3" y1="6" x2="21" y2="6" />
        <line x1="3" y1="18" x2="21" y2="18" />
    </svg>
);

const ShieldLogoIcon = () => (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="#22c55e" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
    </svg>
);

interface MobileHeaderProps {
    onToggleNav: () => void;
}

export const MobileHeader = ({ onToggleNav }: MobileHeaderProps) => (
    <header className="app-mobile-header">
        <button
            type="button"
            className="app-mobile-header__hamburger"
            onClick={onToggleNav}
            aria-label="Toggle navigation">
            <HamburgerIcon />
        </button>

        <Link to="/" className="app-mobile-header__brand">
            <ShieldLogoIcon />
            <span className="app-mobile-header__brand-text">adguard home</span>
        </Link>
    </header>
);

export default MobileHeader;
