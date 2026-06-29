import React from 'react';

type Props = {
    value: number;
    label: string;
    variant?: 'total' | 'ad' | 'tracking' | 'rewrite';
};

const variantConfig: Record<string, { color: string; rgb: string; icon: React.ReactNode }> = {
    total: {
        color: '#4361ee',
        rgb: '67, 97, 238',
        icon: (
            <svg className="yt-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <ellipse cx="12" cy="5" rx="9" ry="3"></ellipse>
                <path d="M3 5v14c0 1.66 4 3 9 3s9-1.34 9-3V5"></path>
                <path d="M3 12c0 1.66 4 3 9 3s9-1.34 9-3"></path>
            </svg>
        ),
    },
    ad: {
        color: '#e74c3c',
        rgb: '231, 76, 60',
        icon: (
            <svg className="yt-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"></path>
            </svg>
        ),
    },
    tracking: {
        color: '#f39c12',
        rgb: '243, 156, 18',
        icon: (
            <svg className="yt-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"></path>
                <line x1="1" y1="1" x2="23" y2="23"></line>
            </svg>
        ),
    },
    rewrite: {
        color: '#27ae60',
        rgb: '39, 174, 96',
        icon: (
            <svg className="yt-card-icon" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <polyline points="17 1 21 5 17 9"></polyline>
                <path d="M3 11V9a4 4 0 0 1 4-4h14"></path>
                <polyline points="7 23 3 19 7 15"></polyline>
                <path d="M21 13v2a4 4 0 0 1-4 4H3"></path>
            </svg>
        ),
    },
};

const YouTubeStatsCard = ({ value, label, variant = 'total' }: Props) => {
    const config = variantConfig[variant];

    return (
        <div
            className={`yt-query-stat yt-query-stat--${variant}`}
            style={{
                background: `linear-gradient(135deg, var(--card-bgcolor), rgba(${config.rgb}, 0.035))`,
                '--accent-rgb': config.rgb,
                '--accent-color': config.color,
            } as React.CSSProperties}>
            <div className="yt-query-stat__header">
                <span className="yt-query-stat__label">{label}</span>
                <span className="yt-query-stat__icon-wrapper" style={{ color: config.color }}>
                    {config.icon}
                </span>
            </div>
            <div className="yt-query-stat__value" style={{ color: config.color }}>
                {value.toLocaleString()}
            </div>
        </div>
    );
};

export default YouTubeStatsCard;

