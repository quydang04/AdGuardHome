import React from 'react';
import { Link } from 'react-router-dom';
import { Trans, useTranslation } from 'react-i18next';

import Tooltip from '../ui/Tooltip';
import { getTrackerData, getSourceData } from '../../helpers/trackers/trackers';
import { captitalizeWords, normalizeWhois } from '../../helpers/helpers';
import { DASHBOARD_MOBILE_PREVIEW_LIMIT } from '../../helpers/constants';

const EyeTrackerIcon = () => (
    <svg
        width="16"
        height="16"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="1.8"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="mobile-stat-item__eye-icon">
        <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z" />
        <circle cx="12" cy="12" r="3" />
    </svg>
);

export const WifiClientIcon = () => (
    <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        className="mobile-client-card__icon">
        <path d="M5 12.55a11 11 0 0 1 14.08 0" />
        <path d="M1.42 9a16 16 0 0 1 21.16 0" />
        <path d="M8.53 16.11a6 6 0 0 1 6.95 0" />
        <line x1="12" y1="20" x2="12.01" y2="20" />
    </svg>
);

const renderTrackerLink = ({ url, name }: { url: string; name: string }) => (
    <a className="tooltip-custom__content-link" target="_blank" rel="noopener noreferrer" href={url}>
        <strong>{name}</strong>
    </a>
);

export const TrackerBadge = ({ domain }: { domain: string }) => {
    const trackerData = getTrackerData(domain);
    if (!trackerData) {
        return <span className="mobile-stat-item__dot" aria-hidden="true" />;
    }

    const trackerInfo = [
        { key: 'name_table_header', value: trackerData, render: renderTrackerLink },
        {
            key: 'category_label',
            value: captitalizeWords(trackerData.category),
            render: (val: string) => (
                <strong>
                    <Trans>{val}</Trans>
                </strong>
            ),
        },
        { key: 'source_label', value: getSourceData(trackerData), render: renderTrackerLink },
    ];

    const content = (
        <div className="popover__list">
            <div className="tooltip-custom__content-title mb-1">
                <Trans>found_in_known_domain_db</Trans>
            </div>
            {trackerInfo.map(({ key, value, render }) => (
                <div key={key} className="tooltip-custom__content-item">
                    <Trans>{key}</Trans>: {render(value)}
                </div>
            ))}
        </div>
    );

    return (
        <Tooltip content={content} placement="top" className="tooltip-container tooltip-custom--wide">
            <span className="mobile-stat-item__icon-wrap">
                <EyeTrackerIcon />
            </span>
        </Tooltip>
    );
};

export const getClientSubtitleText = (info: any): string => {
    if (!info || typeof info !== 'object') {
        return 'N/A';
    }

    if (info.name) {
        return info.name;
    }

    const whoisInfo = info.whois_info ? normalizeWhois(info.whois_info) : null;
    if (whoisInfo) {
        const parts = Object.values(whoisInfo).filter(Boolean);
        if (parts.length > 0) {
            return parts.join(' · ');
        }
    }

    return 'N/A';
};

export interface MobileStatRowItem {
    key: string;
    label: string;
    valueNode: React.ReactNode;
    percent?: number;
    barColor?: string;
    valueColor?: string;
    isTrackerDomain?: boolean;
}

interface MobileStatsListProps {
    items: MobileStatRowItem[];
    totalCount: number;
    showMoreTo?: string;
    previewLimit?: number;
}

export const MobileStatsList = ({
    items,
    totalCount,
    showMoreTo,
    previewLimit = DASHBOARD_MOBILE_PREVIEW_LIMIT,
}: MobileStatsListProps) => {
    const { t } = useTranslation();
    const visibleItems = items.slice(0, previewLimit);

    return (
        <div className="dashboard-mobile-only">
            <div className="mobile-stats-list">
                {visibleItems.map((item) => (
                    <div key={item.key} className="mobile-stat-item">
                        <div className="mobile-stat-item__row">
                            <div className="mobile-stat-item__left">
                                {item.isTrackerDomain && <TrackerBadge domain={item.label} />}
                                <span className="mobile-stat-item__label" title={item.label}>
                                    {item.label}
                                </span>
                            </div>
                            <div
                                className="mobile-stat-item__value"
                                style={item.valueColor ? { color: item.valueColor } : undefined}>
                                {item.valueNode}
                            </div>
                        </div>

                        {typeof item.percent === 'number' && (
                            <div className="mobile-stat-item__bar-track">
                                <div
                                    className="mobile-stat-item__bar-fill"
                                    style={{
                                        width: `${Math.min(Math.max(item.percent, 0), 100)}%`,
                                        backgroundColor: item.barColor || '#94a3b8',
                                    }}
                                />
                            </div>
                        )}
                    </div>
                ))}
            </div>

            {showMoreTo && totalCount > previewLimit && (
                <Link to={showMoreTo} className="dashboard-card-show-more">
                    {t('show_more', 'Show more')}
                </Link>
            )}
        </div>
    );
};

export default MobileStatsList;
