import React, { useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import Card from '../ui/Card';

const GAFAM_DOMAINS: Record<string, string[]> = {
    Google: [
        'google.com', 'googleapis.com', 'gstatic.com', 'youtube.com', 'ytimg.com',
        'googlevideo.com', 'googleusercontent.com', 'google-analytics.com',
        'googleadservices.com', 'doubleclick.net', 'googlesyndication.com',
        'googletagmanager.com', 'google.co', 'gmail.com', 'goog',
        'android.com', 'chromium.org', 'withgoogle.com', 'blogger.com',
        'appspot.com', 'googledomains.com', 'ggpht.com',
    ],
    Amazon: [
        'amazon.com', 'amazonaws.com', 'amazontrust.com', 'cloudfront.net',
        'amazonvideo.com', 'primevideo.com', 'alexa.com', 'amazon.co',
        'amzn.to', 'amzn.com', 'media-amazon.com', 'ssl-images-amazon.com',
    ],
    Facebook: [
        'facebook.com', 'fbcdn.net', 'fb.com', 'instagram.com',
        'whatsapp.com', 'whatsapp.net', 'messenger.com', 'fbsbx.com',
        'facebook.net', 'oculus.com', 'threads.net', 'cdninstagram.com',
    ],
    Apple: [
        'apple.com', 'icloud.com', 'mzstatic.com', 'apple-dns.net',
        'cdn-apple.com', 'apple.news', 'itunes.com', 'me.com',
        'icloud-content.com', 'aaplimg.com',
    ],
    Microsoft: [
        'microsoft.com', 'msftncsi.com', 'msedge.net', 'windows.com',
        'windows.net', 'microsoftonline.com', 'office.com', 'office365.com',
        'live.com', 'outlook.com', 'skype.com', 'bing.com', 'msn.com',
        'azure.com', 'azureedge.net', 'windowsupdate.com', 'xbox.com',
        'linkedin.com', 'github.com', 'visualstudio.com', 'hotmail.com',
        'sharepoint.com', 'onedrive.com', 'aka.ms',
    ],
};

const GAFAM_COLORS: Record<string, string> = {
    Google: '#4285F4',
    Amazon: '#FF9800',
    Facebook: '#1877F2',
    Apple: '#555555',
    Microsoft: '#7CB342',
    Others: '#9e9e9e',
};

function matchesGafam(domain: string): string | null {
    const lower = domain.toLowerCase();
    for (const [company, domains] of Object.entries(GAFAM_DOMAINS)) {
        for (const gafamDomain of domains) {
            if (lower === gafamDomain || lower.endsWith(`.${gafamDomain}`)) {
                return company;
            }
        }
    }
    return null;
}

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
    topQueriedDomains: { name: string; count: number }[];
    numDnsQueries: number;
    subtitle: string;
    refreshButton: React.ReactNode;
}

const GafamDominance = ({ topQueriedDomains, numDnsQueries, subtitle, refreshButton }: GafamDominanceProps) => {
    const { t } = useTranslation();

    const gafamData = useMemo(() => {
        const counts: Record<string, number> = {
            Google: 0,
            Amazon: 0,
            Facebook: 0,
            Apple: 0,
            Microsoft: 0,
        };
        let gafamTotal = 0;

        for (const { name, count } of topQueriedDomains) {
            const company = matchesGafam(name);
            if (company) {
                counts[company] += count;
                gafamTotal += count;
            }
        }

        const othersCount = numDnsQueries - gafamTotal;
        const total = numDnsQueries;

        const segments = Object.entries(counts)
            .filter(([, count]) => count > 0)
            .sort(([, a], [, b]) => b - a)
            .map(([label, value]) => ({
                label,
                value,
                color: GAFAM_COLORS[label],
                pct: total > 0 ? ((value / total) * 100).toFixed(2) : '0',
            }));

        if (othersCount > 0) {
            segments.push({
                label: 'Others',
                value: othersCount,
                color: GAFAM_COLORS.Others,
                pct: total > 0 ? ((othersCount / total) * 100).toFixed(2) : '0',
            });
        }

        return { segments, total };
    }, [topQueriedDomains, numDnsQueries]);

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
