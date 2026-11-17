import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { Link, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';
import { shallowEqual, useDispatch, useSelector } from 'react-redux';
import round from 'lodash/round';
import classNames from 'classnames';

import Loading from '../ui/Loading';
import EmptyCardState from '../ui/EmptyCardState';
import LogsSearchLink from '../ui/LogsSearchLink';
import Tooltip from '../ui/Tooltip';
import CustomDropdownSelect from '../ui/CustomDropdownSelect';
import { ClientBlockActionButton } from './Clients';
import { TrackerBadge, WifiClientIcon, getClientSubtitleText } from './MobileStatsList';
import { getTotalUpstreamRequests } from './UpstreamResponses';

import { getStats, getStatsConfig } from '../../actions/stats';
import { getAccessList } from '../../actions/access';
import { getFilteringStatus } from '../../actions/filtering';
import { DASHBOARD_STATS_URLS, MENU_URLS, STATUS_COLORS } from '../../helpers/constants';
import { Filter, formatNumber, getFilterName, getPercent } from '../../helpers/helpers';
import { RootState } from '../../initialState';
import './TopStatsPage.css';

type SortOrder = 'desc' | 'asc' | 'alpha_asc' | 'alpha_desc';

type StatsCategory =
    | 'clients'
    | 'queried-domains'
    | 'blocked-domains'
    | 'upstreams'
    | 'upstream-avg-time'
    | 'blocked-reasons';

const PATH_TO_CATEGORY: Record<string, StatsCategory> = {
    [DASHBOARD_STATS_URLS.topClients]: 'clients',
    [DASHBOARD_STATS_URLS.topQueriedDomains]: 'queried-domains',
    [DASHBOARD_STATS_URLS.topBlockedDomains]: 'blocked-domains',
    [DASHBOARD_STATS_URLS.topUpstreams]: 'upstreams',
    [DASHBOARD_STATS_URLS.upstreamAvgTime]: 'upstream-avg-time',
    [DASHBOARD_STATS_URLS.blockedReasons]: 'blocked-reasons',
};

interface NormalizedStatItem {
    key: string;
    title: string;
    subtitle?: string;
    rawValue: number;
    percent: number;
    searchTarget?: string;
    disallowed?: boolean;
    disallowedRule?: string;
}

const TopStatsPage = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const location = useLocation();
    const category: StatsCategory = PATH_TO_CATEGORY[location.pathname] || 'blocked-domains';

    const [searchQuery, setSearchQuery] = useState('');
    const [sortOrder, setSortOrder] = useState<SortOrder>('desc');
    const [isRefreshing, setIsRefreshing] = useState(false);
    const [initialLoadDone, setInitialLoadDone] = useState(false);
    const refreshTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    const stats = useSelector<RootState, RootState['stats']>((state) => state.stats, shallowEqual);
    const filters = useSelector<RootState, Filter[]>((state) => state.filtering?.filters || [], shallowEqual);
    const whitelistFilters = useSelector<RootState, Filter[]>(
        (state) => state.filtering?.whitelistFilters || [],
        shallowEqual,
    );

    const fetchPageData = useCallback(() => {
        dispatch(getStatsConfig() as any);
        dispatch(getStats() as any);
        dispatch(getAccessList() as any);
        if (category === 'blocked-reasons') {
            dispatch(getFilteringStatus() as any);
        }
    }, [category, dispatch]);

    useEffect(() => {
        setSearchQuery('');
        setSortOrder('desc');
        fetchPageData();
    }, [category, fetchPageData]);

    useEffect(() => {
        if (!stats.processingStats) {
            setInitialLoadDone(true);
        }
    }, [stats.processingStats]);

    useEffect(
        () => () => {
            if (refreshTimeoutRef.current) {
                clearTimeout(refreshTimeoutRef.current);
            }
        },
        [],
    );

    const handleRefresh = useCallback(() => {
        if (isRefreshing) {
            return;
        }
        setIsRefreshing(true);
        fetchPageData();
        refreshTimeoutRef.current = setTimeout(() => setIsRefreshing(false), 1000);
    }, [fetchPageData, isRefreshing]);

    const config = useMemo(() => {
        switch (category) {
            case 'clients':
                return {
                    title: t('top_clients'),
                    metricLabel: t('queries_label', 'Queries'),
                    accentColor: '#94a3b8',
                    valueColor: undefined,
                    isTimeMetric: false,
                    isDomain: false,
                };
            case 'queried-domains':
                return {
                    title: t('stats_query_domain'),
                    metricLabel: t('queries_label', 'Queries'),
                    accentColor: '#94a3b8',
                    valueColor: undefined,
                    isTimeMetric: false,
                    isDomain: true,
                };
            case 'blocked-domains':
                return {
                    title: t('top_blocked_domains'),
                    metricLabel: t('blocked_queries', 'Blocked queries'),
                    accentColor: '#f97316',
                    valueColor: '#f97316',
                    isTimeMetric: false,
                    isDomain: true,
                };
            case 'upstreams':
                return {
                    title: t('top_upstreams'),
                    metricLabel: t('queries_label', 'Queries'),
                    accentColor: '#67b279',
                    valueColor: undefined,
                    isTimeMetric: false,
                    isDomain: false,
                };
            case 'upstream-avg-time':
                return {
                    title: t('average_upstream_response_time'),
                    metricLabel: t('response_time_label', 'Response time'),
                    accentColor: '#67b279',
                    valueColor: undefined,
                    isTimeMetric: true,
                    isDomain: false,
                };
            case 'blocked-reasons':
                return {
                    title: t('blocked_reasons'),
                    metricLabel: t('blocked_queries', 'Blocked queries'),
                    accentColor: STATUS_COLORS.red,
                    valueColor: STATUS_COLORS.red,
                    isTimeMetric: false,
                    isDomain: false,
                };
        }
    }, [category, t]);

    const rawItems: NormalizedStatItem[] = useMemo(() => {
        if (category === 'clients') {
            const list = stats.topClients || [];
            return list.map(({ name: ip, count, info }: any) => ({
                key: ip,
                title: ip,
                subtitle: getClientSubtitleText(info),
                rawValue: Number(count || 0),
                percent: getPercent(stats.numDnsQueries, count),
                searchTarget: ip,
                disallowed: Boolean(info?.disallowed),
                disallowedRule: info?.disallowed_rule || '',
            }));
        }

        if (category === 'queried-domains') {
            const list = stats.topQueriedDomains || [];
            return list.map(({ name: domain, count }: any) => ({
                key: domain,
                title: domain,
                rawValue: Number(count || 0),
                percent: getPercent(stats.numDnsQueries, count),
                searchTarget: domain,
            }));
        }

        if (category === 'blocked-domains') {
            const totalBlocked =
                (stats.numBlockedFiltering || 0)
                + (stats.numReplacedSafebrowsing || 0)
                + (stats.numReplacedParental || 0)
                + (stats.numReplacedSafesearch || 0);
            const list = stats.topBlockedDomains || [];
            return list.map(({ name: domain, count }: any) => ({
                key: domain,
                title: domain,
                rawValue: Number(count || 0),
                percent: getPercent(totalBlocked, count),
                searchTarget: domain,
            }));
        }

        if (category === 'upstreams') {
            const list = stats.topUpstreamsResponses || [];
            const totalUpstream = getTotalUpstreamRequests(list);
            return list.map(({ name: upstream, count }: any) => ({
                key: upstream,
                title: upstream,
                rawValue: Number(count || 0),
                percent: getPercent(totalUpstream, count),
            }));
        }

        if (category === 'upstream-avg-time') {
            const list = stats.topUpstreamsAvgTime || [];
            const maxSec = list.reduce((max: number, item: any) => Math.max(max, Number(item.count || 0)), 0);
            return list.map(({ name: upstream, count }: any) => {
                const sec = Number(count || 0);
                return {
                    key: upstream,
                    title: upstream,
                    rawValue: sec,
                    percent: maxSec > 0 ? round((sec / maxSec) * 100, 2) : 0,
                };
            });
        }

        const list = stats.topBlockedFilterLists || [];
        return list.map(({ name: filterId, count }: any) => {
            const id = parseInt(filterId, 10);
            const filterName = isNaN(id) ? filterId : getFilterName(filters, whitelistFilters, id);
            return {
                key: String(filterId),
                title: filterName,
                rawValue: Number(count || 0),
                percent: getPercent(stats.numBlockedFiltering, count),
            };
        });
    }, [category, filters, stats, whitelistFilters]);

    const filteredAndSortedItems = useMemo(() => {
        const query = searchQuery.trim().toLowerCase();
        const filtered = query
            ? rawItems.filter(
                  (item) =>
                      item.title.toLowerCase().includes(query)
                      || (item.subtitle && item.subtitle.toLowerCase().includes(query)),
              )
            : [...rawItems];

        filtered.sort((a, b) => {
            switch (sortOrder) {
                case 'asc':
                    return a.rawValue - b.rawValue || a.title.localeCompare(b.title);
                case 'alpha_asc':
                    return a.title.localeCompare(b.title);
                case 'alpha_desc':
                    return b.title.localeCompare(a.title);
                case 'desc':
                default:
                    return b.rawValue - a.rawValue || a.title.localeCompare(b.title);
            }
        });

        return filtered;
    }, [rawItems, searchQuery, sortOrder]);

    const sortOptions = useMemo(() => {
        if (config.isTimeMetric) {
            return [
                { value: 'desc', label: t('sort_decreasing_response_time', 'Decreasing response time') },
                { value: 'asc', label: t('sort_increasing_response_time', 'Increasing response time') },
                { value: 'alpha_asc', label: t('sort_alphabetical_asc', 'Alphabetical (A–Z)') },
                { value: 'alpha_desc', label: t('sort_alphabetical_desc', 'Alphabetical (Z–A)') },
            ];
        }
        return [
            { value: 'desc', label: t('sort_decreasing_queries', 'Decreasing queries') },
            { value: 'asc', label: t('sort_increasing_queries', 'Increasing queries') },
            { value: 'alpha_asc', label: t('sort_alphabetical_asc', 'Alphabetical (A–Z)') },
            { value: 'alpha_desc', label: t('sort_alphabetical_desc', 'Alphabetical (Z–A)') },
        ];
    }, [config.isTimeMetric, t]);

    const isLoading = !initialLoadDone && stats.processingStats;

    return (
        <div className="top-stats-page">
            <nav className="top-stats-breadcrumb" aria-label="Breadcrumb">
                <Link to={MENU_URLS.root} className="top-stats-breadcrumb__link">
                    {t('dashboard')}
                </Link>
                <span className="top-stats-breadcrumb__sep" aria-hidden="true">
                    &gt;
                </span>
                <span className="top-stats-breadcrumb__current">{config.title}</span>
            </nav>

            <div className="top-stats-page__header">
                <h1 className="top-stats-page__title">{config.title}</h1>
                <button
                    type="button"
                    className={classNames('top-stats-page__refresh-btn', {
                        'btn-loading': isRefreshing,
                    })}
                    onClick={handleRefresh}
                    disabled={isRefreshing}
                    title={t('refresh_statics', 'Refresh statistics')}
                    aria-label={t('refresh_statics', 'Refresh statistics')}>
                    <svg
                        width="20"
                        height="20"
                        viewBox="0 0 24 24"
                        fill="none"
                        stroke="currentColor"
                        strokeWidth="2.3"
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        className={isRefreshing ? 'icon-spin' : ''}>
                        <polyline points="23 4 23 10 17 10" />
                        <polyline points="1 20 1 14 7 14" />
                        <path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15" />
                    </svg>
                </button>
            </div>

            <div className="top-stats-controls">
                <div className="top-stats-search">
                    <span className="top-stats-search__icon" aria-hidden="true">
                        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <circle cx="11" cy="11" r="8" />
                            <line x1="21" y1="21" x2="16.65" y2="16.65" />
                        </svg>
                    </span>

                    <input
                        type="text"
                        className="top-stats-search__input"
                        placeholder={t('stats_search_placeholder', 'Search')}
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                    />

                    {searchQuery && (
                        <button
                            type="button"
                            className="top-stats-search__clear"
                            onClick={() => setSearchQuery('')}
                            aria-label="Clear search">
                            ×
                        </button>
                    )}

                    <span className="top-stats-search__help">
                        <Tooltip
                            content={t(
                                'stats_search_tooltip',
                                'Filter items by domain, client IP or name, or upstream server',
                            )}
                            placement="bottom-end"
                            className="tooltip-container tooltip-custom--narrow">
                            <span className="top-stats-search__help-btn" aria-label="Search help">
                                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">
                                    <circle cx="12" cy="12" r="10" />
                                    <path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3" />
                                    <line x1="12" y1="17" x2="12.01" y2="17" />
                                </svg>
                            </span>
                        </Tooltip>
                    </span>
                </div>

                <div className="top-stats-sort">
                    <label htmlFor="top-stats-sort-select" className="top-stats-sort__label">
                        {t('sort_by', 'Sort by')}
                    </label>
                    <CustomDropdownSelect<SortOrder>
                        id="top-stats-sort-select"
                        value={sortOrder}
                        options={sortOptions as { value: SortOrder; label: string }[]}
                        onChange={(val) => setSortOrder(val)}
                    />
                </div>
            </div>

            {isLoading && <Loading />}

            {!isLoading && filteredAndSortedItems.length === 0 && (
                <div className="card">
                    <EmptyCardState />
                </div>
            )}

            {!isLoading && filteredAndSortedItems.length > 0 && (
                <div className="top-stats-list">
                    {filteredAndSortedItems.map((item) => {
                        const valueDisplay = config.isTimeMetric ? (
                            <span className="top-stats-item-card__count">
                                {item.rawValue ? `${formatNumber(round(item.rawValue * 1000))} ms` : '–'}
                            </span>
                        ) : (
                            <span className="top-stats-item-card__count">
                                <span style={config.valueColor ? { color: config.valueColor } : undefined}>
                                    {item.searchTarget ? (
                                        <LogsSearchLink search={item.searchTarget}>
                                            {formatNumber(item.rawValue)}
                                        </LogsSearchLink>
                                    ) : (
                                        formatNumber(item.rawValue)
                                    )}
                                </span>{' '}
                                <span className="top-stats-item-card__percent">({item.percent}%)</span>
                            </span>
                        );

                        return (
                            <div
                                key={item.key}
                                className={classNames('top-stats-item-card', {
                                    'top-stats-item-card--blocked': item.disallowed,
                                })}>
                                <div className="top-stats-item-card__title-row">
                                    {category === 'clients' && <WifiClientIcon />}
                                    {config.isDomain && <TrackerBadge domain={item.title} />}
                                    {category === 'clients' ? (
                                        <>
                                            <Link
                                                to={`/logs?search="${encodeURIComponent(item.title)}"`}
                                                className="top-stats-item-card__title top-stats-item-card__title--link">
                                                {item.title}
                                            </Link>
                                            {!item.disallowed && item.disallowedRule && (
                                                <span className="client-status-badge client-status-badge--allowed ml-auto">
                                                    {t('client_status_allowed', 'Allowed')}
                                                </span>
                                            )}
                                            {item.disallowed && (
                                                <span className="client-status-badge client-status-badge--blocked ml-auto">
                                                    {t('client_status_blocked', 'Blocked')}
                                                </span>
                                            )}
                                        </>
                                    ) : (
                                        <div className="top-stats-item-card__title">{item.title}</div>
                                    )}
                                </div>

                                {category === 'clients' && item.subtitle && (
                                    <div className="top-stats-item-card__subtitle">{item.subtitle}</div>
                                )}

                                <div className="top-stats-item-card__meta">
                                    <span className="top-stats-item-card__metric-label">{config.metricLabel}</span>
                                    {valueDisplay}
                                </div>

                                <div className="top-stats-item-card__bar-track">
                                    <div
                                        className="top-stats-item-card__bar-fill"
                                        style={{
                                            width: `${Math.min(Math.max(item.percent, 0), 100)}%`,
                                            backgroundColor: item.disallowed ? '#f97316' : config.accentColor,
                                        }}
                                    />
                                </div>

                                {category === 'clients' && (
                                    <div className="top-stats-item-card__actions">
                                        <ClientBlockActionButton
                                            ip={item.title}
                                            disallowed={Boolean(item.disallowed)}
                                            disallowedRule={item.disallowedRule || ''}
                                            variant="full"
                                        />
                                    </div>
                                )}
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
};

export default TopStatsPage;
