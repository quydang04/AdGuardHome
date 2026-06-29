import React from 'react';

type Props = {
    value: number;
    label: string;
    variant?: 'total' | 'ad' | 'tracking' | 'rewrite';
};

const variantColors: Record<string, string> = {
    total: '#4361ee',
    ad: '#e74c3c',
    tracking: '#f39c12',
    rewrite: '#27ae60',
};

const YouTubeStatsCard = ({ value, label, variant = 'total' }: Props) => {
    const color = variantColors[variant];

    return (
        <div className="yt-query-stat" style={{ borderLeft: `4px solid ${color}` }}>
            <div className="yt-query-stat__value" style={{ color }}>
                {value.toLocaleString()}
            </div>
            <div className="yt-query-stat__label">{label}</div>
        </div>
    );
};

export default YouTubeStatsCard;
