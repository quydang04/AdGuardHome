import React, { useState, useEffect } from 'react';
import { NavLink, Link, useLocation } from 'react-router-dom';
import { useTranslation, Trans } from 'react-i18next';
import ReactModal from 'react-modal';
import cn from 'clsx';

import { SETTINGS_URLS, FILTERS_URLS, MENU_URLS } from '../../helpers/constants';
import './Sidebar.css';

// SVG Icons matching v1.0 layout
const ShieldLogoIcon = () => (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg">
        <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z" />
    </svg>
);

const HomeIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" />
        <polyline points="9 22 9 12 15 12 15 22" />
    </svg>
);

const GearIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="3" />
        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
    </svg>
);

const FilterIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <line x1="4" y1="21" x2="4" y2="14" />
        <line x1="4" y1="10" x2="4" y2="3" />
        <line x1="12" y1="21" x2="12" y2="12" />
        <line x1="12" y1="8" x2="12" y2="3" />
        <line x1="20" y1="21" x2="20" y2="16" />
        <line x1="20" y1="12" x2="20" y2="3" />
        <line x1="1" y1="14" x2="7" y2="14" />
        <line x1="9" y1="8" x2="15" y2="8" />
        <line x1="17" y1="16" x2="23" y2="16" />
    </svg>
);

const LogIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" />
        <polyline points="14 2 14 8 20 8" />
        <line x1="16" y1="13" x2="8" y2="13" />
        <line x1="16" y1="17" x2="8" y2="17" />
        <polyline points="10 9 9 9 8 9" />
    </svg>
);

const GuideIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
        <line x1="12" y1="17" x2="12.01" y2="17" />
    </svg>
);

const CpuIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="4" width="16" height="16" rx="2" />
        <rect x="9" y="9" width="6" height="6" />
        <line x1="9" y1="1" x2="9" y2="4" />
        <line x1="15" y1="1" x2="15" y2="4" />
        <line x1="9" y1="20" x2="9" y2="23" />
        <line x1="15" y1="20" x2="15" y2="23" />
        <line x1="20" y1="9" x2="23" y2="9" />
        <line x1="20" y1="14" x2="23" y2="14" />
        <line x1="1" y1="9" x2="4" y2="9" />
        <line x1="1" y1="14" x2="4" y2="14" />
    </svg>
);

const LogoutIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4" />
        <polyline points="16 17 21 12 16 7" />
        <line x1="21" y1="12" x2="9" y2="12" />
    </svg>
);

const ChevronDownIcon = ({ isOpen }: { isOpen: boolean }) => (
    <svg
        width="14"
        height="14"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2.2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={cn('app-sidebar__chevron', { 'app-sidebar__chevron--open': isOpen })}>
        <polyline points="6 9 12 15 18 9" />
    </svg>
);

const CollapseSidebarIcon = ({ isCollapsed }: { isCollapsed: boolean }) => (
    <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className={cn('app-sidebar__collapse-icon', { 'app-sidebar__collapse-icon--collapsed': isCollapsed })}>
        <rect x="3" y="3" width="18" height="18" rx="2" ry="2" />
        <line x1="9" y1="3" x2="9" y2="21" />
        <path d="M14 9l-3 3 3 3" />
    </svg>
);

interface SidebarProps {
    isMobileOpen: boolean;
    closeMobile: () => void;
}

if (typeof document !== 'undefined') {
    try {
        ReactModal.setAppElement('#root');
    } catch {}
}

export const Sidebar = ({ isMobileOpen, closeMobile }: SidebarProps) => {
    const { t } = useTranslation();
    const location = useLocation();
    const currentPath = location.pathname;

    const [isCollapsed, setIsCollapsed] = useState(() => {
        try {
            return localStorage.getItem('adguard_sidebar_collapsed') === 'true';
        } catch {
            return false;
        }
    });

    const [isLogoutModalOpen, setIsLogoutModalOpen] = useState(false);

    const toggleCollapse = () => {
        setIsCollapsed((prev) => {
            const next = !prev;
            try {
                localStorage.setItem('adguard_sidebar_collapsed', String(next));
            } catch {}
            return next;
        });
    };

    const isSettingsPath =
        currentPath.startsWith('/settings') ||
        currentPath.startsWith('/dns') ||
        currentPath.startsWith('/encryption') ||
        currentPath.startsWith('/notifications') ||
        currentPath.startsWith('/clients') ||
        currentPath.startsWith('/dhcp');

    const isFiltersPath =
        currentPath.startsWith('/filters') ||
        currentPath.includes('blocklist') ||
        currentPath.includes('allowlist') ||
        currentPath.includes('rewrites') ||
        currentPath.includes('custom_rules') ||
        currentPath.includes('blocked_services');

    const [isSettingsOpen, setIsSettingsOpen] = useState(isSettingsPath);
    const [isFiltersOpen, setIsFiltersOpen] = useState(isFiltersPath);

    useEffect(() => {
        if (isSettingsPath) {
            setIsSettingsOpen(true);
        }
        if (isFiltersPath) {
            setIsFiltersOpen(true);
        }
    }, [isSettingsPath, isFiltersPath]);

    const handleLinkClick = () => {
        closeMobile();
    };

    const handleLogoutClick = (e: React.MouseEvent) => {
        e.preventDefault();
        setIsLogoutModalOpen(true);
    };

    const handleConfirmLogout = () => {
        window.location.href = 'control/logout';
    };

    return (
        <>
            <div
                className={cn('app-sidebar-backdrop', {
                    'app-sidebar-backdrop--visible': isMobileOpen,
                })}
                onClick={closeMobile}
            />

            <aside
                className={cn('app-sidebar', {
                    'app-sidebar--mobile-open': isMobileOpen,
                    'app-sidebar--collapsed': isCollapsed,
                })}>
                <div>
                    {/* Header with Brand Logo & Desktop Collapse Button directly beside it */}
                    <div className="app-sidebar__header">
                        <div className="app-sidebar__brand-row">
                            <Link to="/" className="app-sidebar__brand" onClick={handleLinkClick} title="AdGuard Home">
                                <div className="app-sidebar__brand-logo">
                                    <ShieldLogoIcon />
                                </div>
                                {!isCollapsed && <span className="app-sidebar__brand-text">adguard home</span>}
                            </Link>

                            <button
                                type="button"
                                className="app-sidebar__collapse-btn"
                                onClick={toggleCollapse}
                                title={isCollapsed ? t('expand_menu', 'Mở rộng menu') : t('collapse_menu', 'Thu gọn menu')}
                                aria-label="Toggle sidebar">
                                <CollapseSidebarIcon isCollapsed={isCollapsed} />
                            </button>
                        </div>
                    </div>

                    <nav className="app-sidebar__nav">
                        {/* 1. Tổng quan */}
                        <div className="app-sidebar__item-wrap">
                            <NavLink
                                exact
                                to={MENU_URLS.root}
                                className="app-sidebar__link"
                                activeClassName="app-sidebar__link--active"
                                onClick={handleLinkClick}
                                title={t('dashboard', 'Tổng quan')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <HomeIcon />
                                    </span>
                                    <span><Trans>dashboard</Trans></span>
                                </div>
                            </NavLink>
                        </div>

                        {/* 2. Tổng quan hệ thống */}
                        <div className="app-sidebar__item-wrap">
                            <NavLink
                                to={MENU_URLS.systemOverview}
                                className="app-sidebar__link"
                                activeClassName="app-sidebar__link--active"
                                onClick={handleLinkClick}
                                title={t('system_overview', 'Tổng quan hệ thống')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <CpuIcon />
                                    </span>
                                    <span><Trans>system_overview</Trans></span>
                                </div>
                            </NavLink>
                        </div>

                        {/* 3. Cài đặt */}
                        <div className="app-sidebar__item-wrap app-sidebar__item-wrap--dropdown">
                            <button
                                type="button"
                                className={cn('app-sidebar__link', {
                                    'app-sidebar__link--active': isSettingsPath,
                                })}
                                onClick={() => {
                                    if (isCollapsed) {
                                        toggleCollapse();
                                        setIsSettingsOpen(true);
                                    } else {
                                        setIsSettingsOpen((prev) => !prev);
                                    }
                                }}
                                title={t('settings', 'Cài đặt')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <GearIcon />
                                    </span>
                                    <span><Trans>settings</Trans></span>
                                </div>
                                {!isCollapsed && <ChevronDownIcon isOpen={isSettingsOpen} />}
                            </button>

                            {/* Collapsed Flyout */}
                            {isCollapsed ? (
                                <div className="app-sidebar__flyout">
                                    <div className="app-sidebar__flyout-title"><Trans>settings</Trans></div>
                                    <NavLink to={SETTINGS_URLS.settings} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>general_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.dns} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.encryption} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>encryption_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.notifications} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>notifications_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.dhcp} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dhcp_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.clients} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>client_settings</Trans></NavLink>
                                </div>
                            ) : isSettingsOpen && (
                                <div className="app-sidebar__submenu">
                                    <NavLink to={SETTINGS_URLS.settings} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>general_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.dns} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.encryption} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>encryption_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.notifications} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>notifications_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.dhcp} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dhcp_settings</Trans></NavLink>
                                    <NavLink to={SETTINGS_URLS.clients} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>client_settings</Trans></NavLink>
                                </div>
                            )}
                        </div>

                        {/* 4. Bộ lọc */}
                        <div className="app-sidebar__item-wrap app-sidebar__item-wrap--dropdown">
                            <button
                                type="button"
                                className={cn('app-sidebar__link', {
                                    'app-sidebar__link--active': isFiltersPath,
                                })}
                                onClick={() => {
                                    if (isCollapsed) {
                                        toggleCollapse();
                                        setIsFiltersOpen(true);
                                    } else {
                                        setIsFiltersOpen((prev) => !prev);
                                    }
                                }}
                                title={t('filters', 'Bộ lọc')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <FilterIcon />
                                    </span>
                                    <span><Trans>filters</Trans></span>
                                </div>
                                {!isCollapsed && <ChevronDownIcon isOpen={isFiltersOpen} />}
                            </button>

                            {/* Collapsed Flyout */}
                            {isCollapsed ? (
                                <div className="app-sidebar__flyout">
                                    <div className="app-sidebar__flyout-title"><Trans>filters</Trans></div>
                                    <NavLink to={FILTERS_URLS.dns_blocklists} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_blocklists</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.dns_allowlists} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_allowlists</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.dns_rewrites} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_rewrites</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.blocked_services} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>blocked_services</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.custom_rules} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>custom_filtering_rules</Trans></NavLink>
                                </div>
                            ) : isFiltersOpen && (
                                <div className="app-sidebar__submenu">
                                    <NavLink to={FILTERS_URLS.dns_blocklists} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_blocklists</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.dns_allowlists} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_allowlists</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.dns_rewrites} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>dns_rewrites</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.blocked_services} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>blocked_services</Trans></NavLink>
                                    <NavLink to={FILTERS_URLS.custom_rules} className="app-sidebar__sublink" activeClassName="app-sidebar__sublink--active" onClick={handleLinkClick}><Trans>custom_filtering_rules</Trans></NavLink>
                                </div>
                            )}
                        </div>

                        {/* 5. Nhật ký truy vấn */}
                        <div className="app-sidebar__item-wrap">
                            <NavLink
                                to={MENU_URLS.logs}
                                className="app-sidebar__link"
                                activeClassName="app-sidebar__link--active"
                                onClick={handleLinkClick}
                                title={t('query_log', 'Nhật ký truy vấn')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <LogIcon />
                                    </span>
                                    <span><Trans>query_log</Trans></span>
                                </div>
                            </NavLink>
                        </div>

                        {/* 6. Hướng dẫn thiết lập */}
                        <div className="app-sidebar__item-wrap">
                            <NavLink
                                to={MENU_URLS.guide}
                                className="app-sidebar__link"
                                activeClassName="app-sidebar__link--active"
                                onClick={handleLinkClick}
                                title={t('setup_guide', 'Hướng dẫn thiết lập')}>
                                <div className="app-sidebar__link-left">
                                    <span className="app-sidebar__icon">
                                        <GuideIcon />
                                    </span>
                                    <span><Trans>setup_guide</Trans></span>
                                </div>
                            </NavLink>
                        </div>
                    </nav>
                </div>

                {/* Bottom Sign Out with Confirmation Modal */}
                <div className="app-sidebar__bottom">
                    <button
                        type="button"
                        className="app-sidebar__logout"
                        data-testid="sign_out"
                        onClick={handleLogoutClick}
                        title={t('sign_out', 'Đăng xuất')}>
                        <span className="app-sidebar__icon">
                            <LogoutIcon />
                        </span>
                        <span><Trans>sign_out</Trans></span>
                    </button>
                </div>
            </aside>

            {/* Logout Confirmation Modal */}
            <ReactModal
                className="Modal__Bootstrap modal-dialog modal-dialog-centered"
                closeTimeoutMS={150}
                isOpen={isLogoutModalOpen}
                onRequestClose={() => setIsLogoutModalOpen(false)}>
                <div className="modal-content">
                    <div className="modal-header">
                        <h4 className="modal-title">
                            <Trans>sign_out_confirm_title</Trans>
                        </h4>
                        <button
                            type="button"
                            className="close"
                            onClick={() => setIsLogoutModalOpen(false)}>
                            <span className="sr-only">Close</span>
                        </button>
                    </div>
                    <div className="modal-body">
                        <p className="mb-0">
                            <Trans>sign_out_confirm_desc</Trans>
                        </p>
                    </div>
                    <div className="modal-footer">
                        <button
                            type="button"
                            className="btn btn-secondary"
                            onClick={() => setIsLogoutModalOpen(false)}>
                            <Trans>cancel_btn</Trans>
                        </button>
                        <button
                            type="button"
                            className="btn btn-danger"
                            onClick={handleConfirmLogout}
                            data-testid="confirm_sign_out">
                            <Trans>sign_out</Trans>
                        </button>
                    </div>
                </div>
            </ReactModal>
        </>
    );
};

export default Sidebar;
