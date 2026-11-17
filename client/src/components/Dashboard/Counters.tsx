import React from 'react';
import { useTranslation, Trans } from 'react-i18next';
import round from 'lodash/round';
import { shallowEqual, useSelector } from 'react-redux';

import Card from '../ui/Card';
import { formatNumber, getPercent } from '../../helpers/helpers';
import LogsSearchLink from '../ui/LogsSearchLink';
import { RESPONSE_FILTER } from '../../helpers/constants';
import Tooltip from '../ui/Tooltip';
import EmptyCardState from '../ui/EmptyCardState';
import { RootState } from '../../initialState';
import './Counters.css';

interface CountersProps {
    refreshButton: React.ReactNode;
    subtitle: string;
}

const DnsIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="m17 2 4 4-4 4" />
        <path d="M3 6h18" />
        <path d="m7 22-4-4 4-4" />
        <path d="M21 18H3" />
    </svg>
);

const AdsIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <line x1="4.93" y1="4.93" x2="19.07" y2="19.07" />
    </svg>
);

const ThreatsIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10" />
        <path d="M12 8v4" />
        <path d="M12 16h.01" />
    </svg>
);

const AdultIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <path d="M12 8v8" />
        <path d="M8 12h8" />
    </svg>
);

const SafeSearchIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="11" cy="11" r="8" />
        <line x1="21" y1="21" x2="16.65" y2="16.65" />
    </svg>
);

const ProcessingTimeIcon = () => (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <polyline points="12 6 12 12 16 14" />
    </svg>
);

interface StatItemProps {
    icon: React.ReactNode;
    iconColor: string;
    label: string;
    countText: string;
    valueColor?: string;
    response_status?: string;
    percent?: number;
    barColor?: string;
    tooltipTitle?: string;
}

const StatItem = ({
    icon,
    iconColor,
    label,
    countText,
    valueColor,
    response_status,
    percent,
    barColor,
    tooltipTitle,
}: StatItemProps) => {
    const valueElement = response_status ? (
        <LogsSearchLink response_status={response_status}>{countText}</LogsSearchLink>
    ) : (
        countText
    );

    return (
        <div className="general-stats__row">
            <div className="general-stats__left">
                <span className="general-stats__icon" style={{ color: iconColor }}>
                    {icon}
                </span>
                <span className="general-stats__label">
                    <Trans>{label}</Trans>
                </span>
                {tooltipTitle && (
                    <Tooltip
                        content={tooltipTitle}
                        placement="top"
                        className="tooltip-container tooltip-custom--narrow text-center">
                        <svg className="icons icon--16 icon--lightgray ml-1">
                            <use xlinkHref="#question" />
                        </svg>
                    </Tooltip>
                )}
            </div>

            <div className="general-stats__right">
                <span className="general-stats__value" style={{ color: valueColor }}>
                    {valueElement}
                </span>

                {typeof percent === 'number' && barColor && (
                    <div className="general-stats__bar-track">
                        <div
                            className="general-stats__bar-fill"
                            style={{
                                width: `${Math.min(percent, 100)}%`,
                                backgroundColor: barColor,
                            }}
                        />
                    </div>
                )}
            </div>
        </div>
    );
};

const Counters = ({ refreshButton, subtitle }: CountersProps) => {
    const {
        numDnsQueries,
        numBlockedFiltering,
        numReplacedSafebrowsing,
        numReplacedParental,
        numReplacedSafesearch,
        avgProcessingTime,
    } = useSelector<RootState, RootState['stats']>((state) => state.stats, shallowEqual);

    const { t } = useTranslation();

    const percentAds = getPercent(numDnsQueries, numBlockedFiltering);
    const percentThreats = getPercent(numDnsQueries, numReplacedSafebrowsing);
    const percentAdult = getPercent(numDnsQueries, numReplacedParental);
    const percentSafeSearch = getPercent(numDnsQueries, numReplacedSafesearch);

    const formatCountWithPercent = (count: number, percent: number) =>
        count > 0 ? `${formatNumber(count)} (${percent}%)` : '0';

    const hasData = numDnsQueries > 0;

    return (
        <Card
            title={t('general_statistics')}
            headerExtra={
                hasData ? (
                    <span className="card-header-badge--muted d-none d-md-inline-flex">
                        {formatNumber(numDnsQueries)} {t('queries_total', 'queries total')}
                    </span>
                ) : null
            }
            bodyType="card-table">
            {!hasData ? (
                <EmptyCardState />
            ) : (
                <div className="general-stats">
                    <StatItem
                        icon={<DnsIcon />}
                        iconColor="#64748b"
                        label="dns_query"
                        countText={`${formatNumber(numDnsQueries)} (${t('total_count', 'total')})`}
                        response_status={RESPONSE_FILTER.ALL.QUERY}
                        percent={100}
                        barColor="#6b7280"
                    />

                    <StatItem
                        icon={<AdsIcon />}
                        iconColor="#f97316"
                        label="ads_blocked"
                        countText={formatCountWithPercent(numBlockedFiltering, percentAds)}
                        valueColor="#f97316"
                        response_status={RESPONSE_FILTER.BLOCKED.QUERY}
                        percent={percentAds}
                        barColor="#f97316"
                    />

                    <StatItem
                        icon={<ThreatsIcon />}
                        iconColor="#f59e0b"
                        label="threats_blocked"
                        countText={formatCountWithPercent(numReplacedSafebrowsing, percentThreats)}
                        valueColor="#f59e0b"
                        response_status={RESPONSE_FILTER.BLOCKED_THREATS.QUERY}
                        percent={percentThreats}
                        barColor="#f59e0b"
                    />

                    <StatItem
                        icon={<AdultIcon />}
                        iconColor="#a855f7"
                        label="adult_blocked"
                        countText={formatCountWithPercent(numReplacedParental, percentAdult)}
                        valueColor="#a855f7"
                        response_status={RESPONSE_FILTER.BLOCKED_ADULT_WEBSITES.QUERY}
                        percent={percentAdult}
                        barColor="#a855f7"
                    />

                    <StatItem
                        icon={<SafeSearchIcon />}
                        iconColor="#10b981"
                        label="safe_search_used"
                        countText={formatCountWithPercent(numReplacedSafesearch, percentSafeSearch)}
                        valueColor="#10b981"
                        response_status={RESPONSE_FILTER.SAFE_SEARCH.QUERY}
                        percent={percentSafeSearch}
                        barColor="#10b981"
                    />

                    <div className="general-stats__divider" />

                    <div className="general-stats__row general-stats__row--bottom">
                        <div className="general-stats__left">
                            <span className="general-stats__icon" style={{ color: '#64748b' }}>
                                <ProcessingTimeIcon />
                            </span>
                            <span className="general-stats__label">
                                <Trans>average_processing_time</Trans>
                            </span>
                        </div>

                        <div className="general-stats__right">
                            <span className="general-stats__value general-stats__value--bold">
                                {avgProcessingTime ? round(avgProcessingTime) : 0}&nbsp;ms
                            </span>
                        </div>
                    </div>
                </div>
            )}
        </Card>
    );
};

export default Counters;
