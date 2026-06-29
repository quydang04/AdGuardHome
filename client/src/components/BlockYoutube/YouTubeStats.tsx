import React, { useEffect, useRef, useCallback, useState, useMemo } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import { useTranslation } from 'react-i18next';

import { getYoutubeStats } from '../../actions/youtube';
import { RootState } from '../../initialState';

import YouTubeStatsCard from './YouTubeStatsCard';
import YouTubeTopDomains from './YouTubeTopDomains';
import YouTubeQueryChart from './YouTubeQueryChart';

const EMPTY_STATS = {
    total_youtube_queries: 0,
    blocked_ad_queries: 0,
    blocked_tracking_queries: 0,
    rewritten_queries: 0,
    top_blocked_domains: [] as { domain: string; count: number; type: string }[],
    top_rewrite_domains: [] as { domain: string; count: number; type: string }[],
    hourly_blocked: new Array(24).fill(0) as number[],
    hourly_rewritten: new Array(24).fill(0) as number[],
    query_rate_per_min: 0,
    block_rate_percent: 0,
};

const YouTubeStats = () => {
    const { t } = useTranslation();
    const dispatch = useDispatch();
    const queryStats = useSelector((state: RootState) => state.youtube?.queryStats ?? null);
    const [isRefreshing, setIsRefreshing] = useState(false);
    const refreshTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

    useEffect(() => {
        return () => {
            if (refreshTimeoutRef.current) {
                clearTimeout(refreshTimeoutRef.current);
            }
        };
    }, []);

    const handleRefresh = useCallback(() => {
        if (isRefreshing) {
            return;
        }

        setIsRefreshing(true);
        dispatch(getYoutubeStats());
        refreshTimeoutRef.current = setTimeout(() => setIsRefreshing(false), 1000);
    }, [isRefreshing, dispatch]);

    const refreshButton = useMemo(
        () => (
            <button
                type="button"
                className={`btn btn-icon btn-outline-primary btn-sm${isRefreshing ? ' btn-loading' : ''}`}
                title={t('refresh_btn')}
                disabled={isRefreshing}
                onClick={handleRefresh}>
                <svg className={`icons icon12${isRefreshing ? ' icon-spin' : ''}`}>
                    <use xlinkHref="#refresh" />
                </svg>
            </button>
        ),
        [isRefreshing, handleRefresh, t],
    );

    const stats = queryStats ?? EMPTY_STATS;

    return (
        <div className="yt-stats-section">
            <div className="yt-stats-header">
                <span className="yt-stats-header__title">{t('youtube_query_stats')}</span>
                {refreshButton}
            </div>

            <div className="yt-query-stats-grid">
                <YouTubeStatsCard
                    value={stats.total_youtube_queries}
                    label={t('youtube_total_queries')}
                    variant="total"
                />
                <YouTubeStatsCard
                    value={stats.blocked_ad_queries}
                    label={t('youtube_blocked_ad_queries')}
                    variant="ad"
                />
                <YouTubeStatsCard
                    value={stats.blocked_tracking_queries}
                    label={t('youtube_blocked_tracking_queries')}
                    variant="tracking"
                />
                <YouTubeStatsCard
                    value={stats.rewritten_queries}
                    label={t('youtube_rewritten_queries')}
                    variant="rewrite"
                />
            </div>

            <YouTubeQueryChart
                hourlyBlocked={stats.hourly_blocked}
                hourlyRewritten={stats.hourly_rewritten}
            />

            <YouTubeTopDomains domains={stats.top_blocked_domains} title={t('youtube_top_blocked')} />

            <YouTubeTopDomains domains={stats.top_rewrite_domains} title={t('youtube_top_rewritten')} />
        </div>
    );
};

export default YouTubeStats;
