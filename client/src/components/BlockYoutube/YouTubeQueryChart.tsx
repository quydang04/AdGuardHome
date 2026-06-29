import React from 'react';
import { useTranslation } from 'react-i18next';

type Props = {
    hourlyBlocked: number[];
    hourlyRewritten: number[];
};

const YouTubeQueryChart = ({ hourlyBlocked, hourlyRewritten }: Props) => {
    const { t } = useTranslation();

    const labels = Array.from({ length: 24 }, (_, i) => {
        const h = new Date();
        h.setHours(h.getHours() - (23 - i));
        return `${h.getHours()}:00`;
    });

    const blockedOrdered = [...hourlyBlocked].reverse();
    const rewrittenOrdered = [...hourlyRewritten].reverse();

    const maxVal = Math.max(...blockedOrdered, ...rewrittenOrdered, 1);

    return (
        <div className="yt-query-chart">
            <h6 className="yt-query-chart__title">{t('youtube_query_history')}</h6>
            <div className="yt-query-chart__legend">
                <span className="yt-legend-blocked">{t('youtube_blocked_ad_queries')}</span>
                <span className="yt-legend-rewritten">{t('youtube_rewritten_queries')}</span>
            </div>
            <div className="yt-query-chart__bars">
                {blockedOrdered.map((val, i) => (
                    // eslint-disable-next-line react/no-array-index-key
                    <div key={i} className="yt-bar-group">
                        <div
                            className="yt-bar yt-bar--blocked"
                            style={{ height: `${(val / maxVal) * 60}px` }}
                            title={`${labels[i]}: ${val} blocked`}
                        />
                        <div
                            className="yt-bar yt-bar--rewritten"
                            style={{ height: `${(rewrittenOrdered[i] / maxVal) * 60}px` }}
                            title={`${labels[i]}: ${rewrittenOrdered[i]} rewritten`}
                        />
                    </div>
                ))}
            </div>
        </div>
    );
};

export default YouTubeQueryChart;
