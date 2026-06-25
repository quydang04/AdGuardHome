import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import Card from '../ui/Card';

const GAFAM_COMPANIES = ['Google', 'Amazon', 'Meta', 'Apple', 'Microsoft'] as const;

const GAFAM_COLORS: Record<string, string> = {
    Google: '#EA4335',
    Amazon: '#FF9900',
    Meta: '#0668E1',
    Apple: '#A2AAAD',
    Microsoft: '#7FBA00',
    Others: '#6B7280',
};

interface DonutChartProps {
    data: { label: string; value: number; color: string }[];
    total: number;
}

const DonutChart = ({ data, total }: DonutChartProps) => {
    const size = 160;
    const strokeWidth = 28;
    const radius = (size - strokeWidth) / 2;
    const circumference = 2 * Math.PI * radius;
    const center = size / 2;

    let accumulatedLength = 0;

    return (
        <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`}>
            <circle
                cx={center}
                cy={center}
                r={radius}
                fill="none"
                stroke="var(--card-border-color, #e0e0e0)"
                strokeWidth={strokeWidth}
            />
            {data.map((segment) => {
                const pct = total > 0 ? segment.value / total : 0;
                const dashLength = pct * circumference;
                const dashGap = circumference - dashLength;
                const dashOffset = circumference / 4 - accumulatedLength;
                accumulatedLength += dashLength;

                return (
                    <circle
                        key={segment.label}
                        cx={center}
                        cy={center}
                        r={radius}
                        fill="none"
                        stroke={segment.color}
                        strokeWidth={strokeWidth}
                        strokeDasharray={`${dashLength} ${dashGap}`}
                        strokeDashoffset={dashOffset}
                        style={{ transition: 'stroke-dasharray 0.3s ease' }}
                    />
                );
            })}
        </svg>
    );
};

interface GafamDominanceProps {
    gafamStats: Record<string, number>;
    numDnsQueries: number;
    subtitle: string;
    refreshButton: React.ReactNode;
}

const GafamDominance = ({ gafamStats, numDnsQueries, subtitle, refreshButton }: GafamDominanceProps) => {
    const { t } = useTranslation();

    const gafamData = useMemo(() => {
        const total = numDnsQueries;
        let gafamTotal = 0;

        const segments: { label: string; value: number; color: string; pct: string }[] = GAFAM_COMPANIES
            .map((company) => {
                const value = gafamStats[company] || 0;
                gafamTotal += value;

                return {
                    label: company as string,
                    value,
                    color: GAFAM_COLORS[company],
                    pct: total > 0 ? ((value / total) * 100).toFixed(2) : '0',
                };
            })
            .sort((a, b) => b.value - a.value);

        const othersCount = total - gafamTotal;
        if (othersCount > 0) {
            segments.push({
                label: 'Others',
                value: othersCount,
                color: GAFAM_COLORS.Others,
                pct: total > 0 ? ((othersCount / total) * 100).toFixed(2) : '0',
            });
        }

        return { segments, total };
    }, [gafamStats, numDnsQueries]);

    return (
        <Card
            title={t('gafam_dominance')}
            subtitle={subtitle}
            refresh={refreshButton}>
            <div className="gafam-dominance">
                <p className="gafam-dominance__desc">{t('gafam_dominance_desc')}</p>
                <div className="gafam-dominance__content">
                    <div className="gafam-dominance__chart">
                        <DonutChart data={gafamData.segments} total={gafamData.total} />
                    </div>
                    <div className="gafam-dominance__legend">
                        {gafamData.segments.map((segment) => (
                            <div key={segment.label} className="gafam-dominance__legend-item">
                                <span
                                    className="gafam-dominance__legend-dot"
                                    style={{ backgroundColor: segment.color }}
                                />
                                <span className="gafam-dominance__legend-label">
                                    <strong>{segment.label === 'Others' ? t('gafam_others') : segment.label}</strong>
                                    {' '}
                                    {segment.pct}%
                                    {' '}
                                    <span className="gafam-dominance__legend-count">
                                        ({segment.value.toLocaleString()} {t('requests_count').toLowerCase()})
                                    </span>
                                </span>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </Card>
    );
};

export default GafamDominance;
