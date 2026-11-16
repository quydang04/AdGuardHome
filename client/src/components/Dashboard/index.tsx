import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';

import { HashLink as Link } from 'react-router-hash-link';
import { Trans, useTranslation } from 'react-i18next';
import classNames from 'classnames';

import Statistics from './Statistics';
import Counters from './Counters';
import Clients from './Clients';
import QueriedDomains from './QueriedDomains';
import BlockedDomains from './BlockedDomains';
import { DAY, HOUR, DISABLE_PROTECTION_TIMINGS, ONE_SECOND_IN_MS, SETTINGS_URLS, TIME_UNITS } from '../../helpers/constants';
import { msToSeconds, msToMinutes, msToHours, msToDays } from '../../helpers/helpers';

import PageTitle from '../ui/PageTitle';

import Loading from '../ui/Loading';
import './Dashboard.css';

import Dropdown from '../ui/Dropdown';
import UpstreamResponses from './UpstreamResponses';

import UpstreamAvgTime from './UpstreamAvgTime';
import BlockedReasons from './BlockedReasons';
import GafamDominance from './GafamDominance';
import EncryptedDns from './EncryptedDns';
import DnssecStats from './DnssecStats';
import { AccessData, DashboardData, StatsData } from '../../initialState';

const STATS_POLLING_INTERVAL_MS = 5000;

interface DashboardProps {
    dashboard: DashboardData;
    stats: StatsData;
    access: AccessData;
    getStats: (...args: unknown[]) => unknown;
    getStatsConfig: (...args: unknown[]) => unknown;
    setStatsConfig: (...args: unknown[]) => unknown;
    getFilteringStatus: (...args: unknown[]) => unknown;
    toggleProtection: (...args: unknown[]) => unknown;
    getClients: (...args: unknown[]) => unknown;
    getAccessList: () => (dispatch: any) => void;
}

const Dashboard = ({
    getAccessList,
    getStats,
    getStatsConfig,
    setStatsConfig,
    getFilteringStatus,
    dashboard: { protectionEnabled, processingProtection, protectionDisabledDuration },
    toggleProtection,
    stats,
    access,
}: DashboardProps) => {
    const { t } = useTranslation();
    const [initialLoadDone, setInitialLoadDone] = useState(false);
    const [isRefreshing, setIsRefreshing] = useState(false);
    const [isProtectionOptionsOpen, setIsProtectionOptionsOpen] = useState(false);
    const [isMobileProtectionOptionsOpen, setIsMobileProtectionOptionsOpen] = useState(false);
    const [isIntervalOpen, setIsIntervalOpen] = useState(false);
    const [isMobileIntervalOpen, setIsMobileIntervalOpen] = useState(false);
    const protectionOptionsRef = useRef<HTMLDivElement>(null);
    const mobileProtectionOptionsRef = useRef<HTMLDivElement>(null);
    const intervalRef = useRef<HTMLDivElement>(null);
    const mobileIntervalRef = useRef<HTMLDivElement>(null);
    const refreshTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        const handleClickOutside = (e: MouseEvent) => {
            const target = e.target as Node;
            if (protectionOptionsRef.current && !protectionOptionsRef.current.contains(target)) {
                setIsProtectionOptionsOpen(false);
            }
            if (mobileProtectionOptionsRef.current && !mobileProtectionOptionsRef.current.contains(target)) {
                setIsMobileProtectionOptionsOpen(false);
            }
            if (intervalRef.current && !intervalRef.current.contains(target)) {
                setIsIntervalOpen(false);
            }
            if (mobileIntervalRef.current && !mobileIntervalRef.current.contains(target)) {
                setIsMobileIntervalOpen(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const getAllStats = useCallback(() => {
        getAccessList();
        getStats();
        getFilteringStatus();
        getStatsConfig();
    }, [getAccessList, getStats, getFilteringStatus, getStatsConfig]);

    const handleManualRefresh = useCallback(() => {
        if (isRefreshing) {
            return;
        }
        setIsRefreshing(true);
        getAllStats();
        refreshTimeoutRef.current = setTimeout(() => setIsRefreshing(false), 1000);
    }, [isRefreshing, getAllStats]);

    useEffect(() => {
        getAllStats();
    }, [getAllStats]);

    const prevProcessing = useRef(true);
    useEffect(() => {
        if (prevProcessing.current && !stats.processingStats) {
            setInitialLoadDone(true);
        }
        prevProcessing.current = stats.processingStats;
    }, [stats.processingStats]);

    const getStatsRef = useRef(getStats);
    getStatsRef.current = getStats;

    useEffect(() => {
        const intervalId = setInterval(() => {
            if (document.visibilityState === 'visible') {
                getStatsRef.current();
            }
        }, STATS_POLLING_INTERVAL_MS);

        const handleVisibilityChange = () => {
            if (document.visibilityState === 'visible') {
                getStatsRef.current();
            }
        };

        document.addEventListener('visibilitychange', handleVisibilityChange);

        return () => {
            clearInterval(intervalId);
            document.removeEventListener('visibilitychange', handleVisibilityChange);
            if (refreshTimeoutRef.current) {
                clearTimeout(refreshTimeoutRef.current);
            }
        };
    }, []);

    const subtitle = useMemo(() => {
        if (!stats.enabled) {
            return t('stats_disabled_short');
        }

        const msIn7Days = 604800000;

        if (stats.timeUnits === TIME_UNITS.HOURS && stats.interval === msIn7Days) {
            return t('for_last_days', { count: msToDays(stats.interval) });
        }

        return stats.timeUnits === TIME_UNITS.HOURS
            ? t('for_last_hours', { count: msToHours(stats.interval) })
            : t('for_last_days', { count: msToDays(stats.interval) });
    }, [stats.enabled, stats.timeUnits, stats.interval, t]);

    const INTERVAL_OPTIONS = useMemo(() => {
        const maxLimit = stats.interval || DAY;
        const all = [
            { label: t('for_last_hours', { count: 6 }), value: HOUR * 6 },
            { label: t('for_last_hours', { count: 24 }), value: DAY },
            { label: t('for_last_days', { count: 7 }), value: DAY * 7 },
            { label: t('for_last_days', { count: 30 }), value: DAY * 30 },
            { label: t('for_last_days', { count: 90 }), value: DAY * 90 },
        ];
        return all.filter((opt) => opt.value <= Math.max(maxLimit, DAY));
    }, [t, stats.interval]);

    const handleSelectInterval = useCallback(
        async (val: number) => {
            setIsIntervalOpen(false);
            setIsMobileIntervalOpen(false);
            if (val === stats.interval) {
                return;
            }
            await setStatsConfig({
                enabled: stats.enabled !== false,
                interval: val,
                ignored: stats.ignored || [],
            });
            getStats();
        },
        [setStatsConfig, getStats, stats.enabled, stats.interval, stats.ignored],
    );

    const buttonClass = classNames('btn btn-sm dashboard-protection-button', {
        'btn-gray': protectionEnabled,
        'btn-success': !protectionEnabled,
    });

    const refreshButton = useMemo(() => (
        <button
            type="button"
            className={`btn btn-icon btn-outline-primary btn-sm${isRefreshing ? ' btn-loading' : ''}`}
            title={t('refresh_btn')}
            disabled={isRefreshing}
            onClick={handleManualRefresh}>
            <svg className={`icons icon12${isRefreshing ? ' icon-spin' : ''}`}>
                <use xlinkHref="#refresh" />
            </svg>
        </button>
    ), [isRefreshing, handleManualRefresh, t]);

    const statsProcessing = !initialLoadDone && (stats.processingStats || stats.processingGetConfig || access.processing);

    const getTomorrowTimeText = () => {
        const d = new Date();
        let hh = d.getHours();
        const mm = `0${d.getMinutes()}`.slice(-2);
        const ampm = hh >= 12 ? 'PM' : 'AM';
        hh = hh % 12 || 12;
        return `${`0${hh}`.slice(-2)}:${mm} ${ampm}`;
    };

    const DISABLE_PROTECTION_ITEMS = useMemo(() => [
        {
            text: t('disable_for_seconds', { count: msToSeconds(DISABLE_PROTECTION_TIMINGS.HALF_MINUTE), defaultValue: `Tạm dừng trong ${msToSeconds(DISABLE_PROTECTION_TIMINGS.HALF_MINUTE)} giây` }),
            disableTime: DISABLE_PROTECTION_TIMINGS.HALF_MINUTE,
        },
        {
            text: t('disable_for_minutes', { count: msToMinutes(DISABLE_PROTECTION_TIMINGS.MINUTE), defaultValue: `Tạm dừng trong ${msToMinutes(DISABLE_PROTECTION_TIMINGS.MINUTE)} phút` }),
            disableTime: DISABLE_PROTECTION_TIMINGS.MINUTE,
        },
        {
            text: t('disable_for_minutes', { count: msToMinutes(DISABLE_PROTECTION_TIMINGS.TEN_MINUTES), defaultValue: `Tạm dừng trong ${msToMinutes(DISABLE_PROTECTION_TIMINGS.TEN_MINUTES)} phút` }),
            disableTime: DISABLE_PROTECTION_TIMINGS.TEN_MINUTES,
        },
        {
            text: t('disable_for_hours', { count: msToHours(DISABLE_PROTECTION_TIMINGS.HOUR), defaultValue: `Tạm dừng trong ${msToHours(DISABLE_PROTECTION_TIMINGS.HOUR)} giờ` }),
            disableTime: DISABLE_PROTECTION_TIMINGS.HOUR,
        },
        {
            text: t('disable_until_tomorrow', { time: getTomorrowTimeText(), defaultValue: `Tạm dừng đến ngày mai, ${getTomorrowTimeText()}` }),
            disableTime: DISABLE_PROTECTION_TIMINGS.TOMORROW,
        },
    ], [t]);

    const getDisableProtectionItems = (isMobile = false) =>
        Object.values(DISABLE_PROTECTION_ITEMS).map((item: any, index: any) => (
            <button
                type="button"
                key={`disable_timings_${index}`}
                className="dropdown-item dropdown-item--protection"
                onMouseDown={(e) => {
                    e.stopPropagation();
                }}
                onClick={(e) => {
                    e.stopPropagation();
                    if (isMobile) {
                        setIsMobileProtectionOptionsOpen(false);
                    } else {
                        setIsProtectionOptionsOpen(false);
                    }
                    toggleProtection(protectionEnabled, item.disableTime - ONE_SECOND_IN_MS);
                }}>
                <span className="dropdown-item__dot" />
                <span className="dropdown-item__text">{item.text}</span>
            </button>
        ));

    const getRemaningTimeText = (milliseconds: any) => {
        if (!milliseconds) {
            return '';
        }

        const date = new Date(milliseconds);
        const hh = date.getUTCHours();
        const mm = `0${date.getUTCMinutes()}`.slice(-2);
        const ss = `0${date.getUTCSeconds()}`.slice(-2);
        const formattedHH = `0${hh}`.slice(-2);

        return hh ? `${formattedHH}:${mm}:${ss}` : `${mm}:${ss}`;
    };

    const getProtectionBtnText = (status: any) => (status ? t('disable_protection') : t('enable_protection'));

    return (
        <>
            {/* Desktop Topbar */}
            <div className="dashboard-topbar dashboard-topbar--desktop">
                <div className="dashboard-topbar__left">
                    <h1 className="dashboard-topbar__title">
                        {t('protection_section_label', 'Bảo vệ')}
                    </h1>

                    <label
                        className="dashboard-switch"
                        title={t(protectionEnabled ? 'disable_protection' : 'enable_protection')}>
                        <input
                            type="checkbox"
                            checked={protectionEnabled}
                            onChange={() => toggleProtection(protectionEnabled)}
                            disabled={processingProtection}
                        />
                        <span className="dashboard-switch__slider" />
                    </label>

                    {protectionDisabledDuration > 0 && (
                        <span className="dashboard-protection-timer badge badge-warning">
                            {getRemaningTimeText(protectionDisabledDuration)}
                        </span>
                    )}

                    {protectionEnabled && (
                        <div className="dashboard-more-dropdown" ref={protectionOptionsRef}>
                            <button
                                type="button"
                                className="dashboard-more-btn"
                                onClick={() => setIsProtectionOptionsOpen((prev) => !prev)}
                                title={t('more_options', 'More options')}>
                                <svg width="18" height="18" viewBox="0 0 24 24" fill="currentColor">
                                    <circle cx="12" cy="5" r="2.2" />
                                    <circle cx="12" cy="12" r="2.2" />
                                    <circle cx="12" cy="19" r="2.2" />
                                </svg>
                            </button>
                            {isProtectionOptionsOpen && (
                                <div className="dropdown-menu dropdown-menu--protection show">
                                    {getDisableProtectionItems(false)}
                                </div>
                            )}
                        </div>
                    )}
                </div>

                <div className="dashboard-topbar__right">
                    <button
                        type="button"
                        className={`dashboard-refresh-btn${isRefreshing ? ' btn-loading' : ''}`}
                        disabled={isRefreshing}
                        onClick={handleManualRefresh}>
                        <span><Trans>refresh_statics</Trans></span>
                        <svg
                            width="14"
                            height="14"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2.2"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className={`dashboard-refresh-icon ml-2${isRefreshing ? ' icon-spin' : ''}`}>
                            <polyline points="23 4 23 10 17 10" />
                            <polyline points="1 20 1 14 7 14" />
                            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
                        </svg>
                    </button>

                    <div className="dashboard-interval-dropdown-wrap" ref={intervalRef}>
                        <button
                            type="button"
                            className="dashboard-interval-btn"
                            onClick={() => setIsIntervalOpen((prev) => !prev)}
                            title={t('stats_config_title', 'Configure stats retention')}>
                            <span>{subtitle}</span>
                            <svg
                                width="14"
                                height="14"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2.2"
                                strokeLinecap="round"
                                strokeLinejoin="round"
                                className="dashboard-interval-icon ml-1">
                                <polyline points="6 9 12 15 18 9" />
                            </svg>
                        </button>

                        {isIntervalOpen && (
                            <div className="dashboard-interval-menu">
                                {INTERVAL_OPTIONS.map((opt) => (
                                    <button
                                        key={opt.value}
                                        type="button"
                                        className={classNames('dashboard-interval-menu-item', {
                                            'dashboard-interval-menu-item--active': stats.interval === opt.value,
                                        })}
                                        onClick={() => handleSelectInterval(opt.value)}>
                                        <span className="dashboard-interval-menu-item__icon">
                                            {stats.interval === opt.value ? (
                                                <span className="dashboard-interval-menu-item__check">✓</span>
                                            ) : (
                                                <span className="dashboard-interval-menu-item__dot">•</span>
                                            )}
                                        </span>
                                        <span className="dashboard-interval-menu-item__label">{opt.label}</span>
                                    </button>
                                ))}
                                <div className="dashboard-interval-menu__footer">
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dashboard-interval-menu__gear">
                                        <circle cx="12" cy="12" r="3" />
                                        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-2 2 2 2 0 0 1-2-2v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83 0 2 2 0 0 1 0-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
                                    </svg>
                                    <div className="dashboard-interval-menu__help">
                                        <Trans i18nKey="stats_interval_help">
                                            Để thay đổi các khoảng thời gian hiển thị, hãy chỉnh sửa <em>Thời gian lưu trữ thống kê</em> trong <Link to={`${SETTINGS_URLS.settings}#stats-config`} onClick={() => { setIsIntervalOpen(false); setIsMobileIntervalOpen(false); }}>Cài đặt chung</Link>
                                        </Trans>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {/* Mobile Topbar */}
            <div className="dashboard-topbar--mobile">
                <div className="dashboard-mobile-row dashboard-mobile-row--title">
                    <h1 className="dashboard-mobile-title">
                        {t('overview', 'Tổng quan')}
                    </h1>
                    <button
                        type="button"
                        className={`dashboard-mobile-refresh-btn${isRefreshing ? ' btn-loading' : ''}`}
                        disabled={isRefreshing}
                        onClick={handleManualRefresh}
                        title={t('refresh_statics', 'Làm mới thông kê')}>
                        <svg
                            width="20"
                            height="20"
                            viewBox="0 0 24 24"
                            fill="none"
                            stroke="currentColor"
                            strokeWidth="2.4"
                            strokeLinecap="round"
                            strokeLinejoin="round"
                            className={isRefreshing ? 'icon-spin' : ''}>
                            <polyline points="23 4 23 10 17 10" />
                            <polyline points="1 20 1 14 7 14" />
                            <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
                        </svg>
                    </button>
                </div>

                <div className="dashboard-mobile-row dashboard-mobile-row--switch">
                    <div className="dashboard-mobile-switch-wrap">
                        <label
                            className="dashboard-switch"
                            title={t(protectionEnabled ? 'disable_protection' : 'enable_protection')}>
                            <input
                                type="checkbox"
                                checked={protectionEnabled}
                                onChange={() => toggleProtection(protectionEnabled)}
                                disabled={processingProtection}
                            />
                            <span className="dashboard-switch__slider" />
                        </label>
                        <span className="dashboard-mobile-switch-label">
                            {t('protection_section_label', 'Bảo vệ')}
                        </span>
                    </div>

                    <div className="dashboard-more-dropdown" ref={mobileProtectionOptionsRef}>
                        <button
                            type="button"
                            className="dashboard-more-btn"
                            onClick={() => setIsMobileProtectionOptionsOpen((prev) => !prev)}
                            title={t('more_options', 'More options')}>
                            <svg width="20" height="20" viewBox="0 0 24 24" fill="currentColor">
                                <circle cx="12" cy="5" r="2.2" />
                                <circle cx="12" cy="12" r="2.2" />
                                <circle cx="12" cy="19" r="2.2" />
                            </svg>
                        </button>
                        {isMobileProtectionOptionsOpen && (
                            <div className="dropdown-menu dropdown-menu--protection show">
                                {getDisableProtectionItems(true)}
                            </div>
                        )}
                    </div>
                </div>

                <div className="dashboard-mobile-row dashboard-mobile-row--interval">
                    <div className="dashboard-interval-dropdown-wrap w-100" ref={mobileIntervalRef}>
                        <button
                            type="button"
                            className="dashboard-mobile-interval-btn"
                            onClick={() => setIsMobileIntervalOpen((prev) => !prev)}
                            title={t('stats_config_title', 'Configure stats retention')}>
                            <span>{subtitle}</span>
                            <svg
                                width="16"
                                height="16"
                                viewBox="0 0 24 24"
                                fill="none"
                                stroke="currentColor"
                                strokeWidth="2.2"
                                strokeLinecap="round"
                                strokeLinejoin="round">
                                <polyline points="6 9 12 15 18 9" />
                            </svg>
                        </button>

                        {isMobileIntervalOpen && (
                            <div className="dashboard-interval-menu dashboard-interval-menu--mobile">
                                {INTERVAL_OPTIONS.map((opt) => (
                                    <button
                                        key={opt.value}
                                        type="button"
                                        className={classNames('dashboard-interval-menu-item', {
                                            'dashboard-interval-menu-item--active': stats.interval === opt.value,
                                        })}
                                        onClick={() => handleSelectInterval(opt.value)}>
                                        <span className="dashboard-interval-menu-item__icon">
                                            {stats.interval === opt.value ? (
                                                <span className="dashboard-interval-menu-item__check">✓</span>
                                            ) : (
                                                <span className="dashboard-interval-menu-item__dot">•</span>
                                            )}
                                        </span>
                                        <span className="dashboard-interval-menu-item__label">{opt.label}</span>
                                    </button>
                                ))}
                                <div className="dashboard-interval-menu__footer">
                                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="dashboard-interval-menu__gear">
                                        <circle cx="12" cy="12" r="3" />
                                        <path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1 0 2.83 2 2 0 0 1-2.83 0l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1-2-2 2 2 0 0 1 2-2h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 0-2.83 2 2 0 0 1 2.83 0l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 2-2 2 2 0 0 1 2 2v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 0 2 2 0 0 1 0 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 2 2 2 2 0 0 1-2 2h-.09a1.65 1.65 0 0 0-1.51 1z" />
                                    </svg>
                                    <div className="dashboard-interval-menu__help">
                                        <Trans i18nKey="stats_interval_help">
                                            Để thay đổi các khoảng thời gian hiển thị, hãy chỉnh sửa <em>Thời gian lưu trữ thống kê</em> trong <Link to={`${SETTINGS_URLS.settings}#stats-config`} onClick={() => { setIsIntervalOpen(false); setIsMobileIntervalOpen(false); }}>Cài đặt chung</Link>
                                        </Trans>
                                    </div>
                                </div>
                            </div>
                        )}
                    </div>
                </div>
            </div>

            {statsProcessing && <Loading />}

            {!statsProcessing && (
                <div className="row row-cards dashboard">
                    <div className="col-lg-12">
                        {stats.interval === 0 && (
                            <div className="alert alert-warning" role="alert">
                                <Trans
                                    components={[
                                        <Link to={`${SETTINGS_URLS.settings}#stats-config`} key="0">
                                            link
                                        </Link>,
                                    ]}>
                                    stats_disabled
                                </Trans>
                            </div>
                        )}

                        <Statistics
                            dnsQueries={stats.dnsQueries}
                            blockedFiltering={stats.blockedFiltering}
                            replacedSafebrowsing={stats.replacedSafebrowsing}
                            replacedParental={stats.replacedParental}
                            numDnsQueries={stats.numDnsQueries}
                            numBlockedFiltering={stats.numBlockedFiltering}
                            numReplacedSafebrowsing={stats.numReplacedSafebrowsing}
                            numReplacedParental={stats.numReplacedParental}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <Counters subtitle="" refreshButton={null} />
                    </div>

                    <div className="col-lg-6">
                        <Clients subtitle="" refreshButton={null} />
                    </div>

                    <div className="col-lg-6">
                        <QueriedDomains
                            subtitle=""
                            dnsQueries={stats.numDnsQueries}
                            topQueriedDomains={stats.topQueriedDomains}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <BlockedDomains
                            subtitle=""
                            topBlockedDomains={stats.topBlockedDomains}
                            blockedFiltering={stats.numBlockedFiltering}
                            replacedSafebrowsing={stats.numReplacedSafebrowsing}
                            replacedSafesearch={stats.numReplacedSafesearch}
                            replacedParental={stats.numReplacedParental}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <UpstreamResponses
                            subtitle=""
                            topUpstreamsResponses={stats.topUpstreamsResponses}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <UpstreamAvgTime
                            subtitle=""
                            topUpstreamsAvgTime={stats.topUpstreamsAvgTime}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <BlockedReasons
                            subtitle=""
                            topBlockedFilterLists={stats.topBlockedFilterLists}
                            numBlockedFiltering={stats.numBlockedFiltering}
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-12">
                        <GafamDominance
                            gafamStats={stats.gafamStats}
                            numDnsQueries={stats.numDnsQueries}
                            subtitle=""
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <EncryptedDns
                            numEncryptedDns={stats.numEncryptedDns}
                            numDnsQueries={stats.numDnsQueries}
                            subtitle=""
                            refreshButton={null}
                        />
                    </div>

                    <div className="col-lg-6">
                        <DnssecStats
                            numDnssec={stats.numDnssec}
                            numDnsQueries={stats.numDnsQueries}
                            numEncryptedDns={stats.numEncryptedDns}
                            subtitle=""
                            refreshButton={null}
                        />
                    </div>

                </div>
            )}
        </>
    );
};

export default Dashboard;
