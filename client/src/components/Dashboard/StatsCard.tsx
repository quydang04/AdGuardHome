import React from 'react';
import cn from 'clsx';
import { Link } from 'react-router-dom';

import { formatNumber } from '../../helpers/helpers';
import Line from '../ui/Line';
import './StatsCard.css';

export const STATS_CARD_VARIANTS = {
    QUERIES: 'queries',
    ADS: 'ads',
    THREATS: 'threats',
    ADULT: 'adult',
} as const;

export type StatsCardVariant = typeof STATS_CARD_VARIANTS[keyof typeof STATS_CARD_VARIANTS];

const CHART_COLORS: Record<StatsCardVariant, string> = {
    queries: '#94a3b8',
    ads: '#f97316',
    threats: '#f59e0b',
    adult: '#a855f7',
};

type Props = {
    total: number;
    lineData: number[];
    title: React.ReactNode;
    variant: StatsCardVariant;
    percent?: number;
    to?: string;
};

export const StatsCard = React.memo(({ total, lineData, percent, title, variant, to }: Props) => {
    const showPercent = typeof percent === 'number';
    const accentColor = CHART_COLORS[variant];

    const boxContent = (
        <>
            <div className="stats-card-v1__header">
                <span className="stats-card-v1__value">{formatNumber(total)}</span>
                {showPercent && (
                    <span className="stats-card-v1__percent">{Math.round(percent)}%</span>
                )}
            </div>

            <div className="stats-card-v1__chart">
                <Line data={lineData} color={accentColor} variant={variant} />
            </div>
        </>
    );

    return (
        <div className={cn('stats-card-v1', `stats-card-v1--${variant}`)}>
            {to ? (
                <Link to={to} className="stats-card-v1__box">
                    {boxContent}
                </Link>
            ) : (
                <div className="stats-card-v1__box">
                    {boxContent}
                </div>
            )}

            <div className="stats-card-v1__label">
                {to ? <Link to={to}>{title}</Link> : title}
            </div>
        </div>
    );
});

export default StatsCard;
