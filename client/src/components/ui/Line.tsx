import React, { useMemo } from 'react';
import { Area, AreaChart, ResponsiveContainer, Tooltip, TooltipProps, YAxis } from 'recharts';
import { shallowEqual, useSelector } from 'react-redux';
import './Line.css';

import { RootState } from '../../initialState';
import { buildChartData, HistoryPoint } from './lineUtils';

interface LineProps {
    data: number[];
    color?: string;
    width?: number;
    height?: number;
    variant?: string;
}

const gradientId = (color: string) => `line-gradient-${color.replace('#', '')}`;

const CustomTooltip = React.memo(({ active, payload }: TooltipProps<number, string>) => {
    if (!active || !payload || payload.length === 0) {
        return null;
    }

    const point = payload[0].payload as HistoryPoint;

    return (
        <div className="line__tooltip">
            <strong className="line__tooltip-value">{point.value}</strong>
            <small className="line__tooltip-label">{point.label}</small>
        </div>
    );
});

CustomTooltip.displayName = 'CustomTooltip';

// Distinct organic baseline waves matching AdGuard v1.0 beta screenshots
const DEMO_WAVES: Record<string, number[]> = {
    queries: [2, 2.5, 3.2, 2.8, 4.0, 3.5, 5.0, 4.2, 4.8, 3.6, 2.8, 2.2, 2.6, 2.0],
    ads: [1.5, 2.0, 2.8, 3.8, 4.5, 3.8, 3.0, 2.4, 2.0, 1.8, 2.2, 2.5, 1.8, 1.5],
    threats: [1.8, 2.2, 2.6, 3.0, 3.5, 4.0, 3.6, 3.0, 2.5, 2.0, 1.8, 2.0, 1.6, 1.4],
    adult: [1.2, 1.5, 1.8, 2.2, 2.0, 2.8, 3.5, 4.2, 3.8, 3.0, 2.5, 2.0, 1.8, 1.5],
};

const Line = React.memo(({ data, color = 'black', variant = 'queries' }: LineProps) => {
    const { interval, timeUnits } = useSelector(
        (state: RootState) => ({
            interval: state.stats.viewInterval || state.stats.interval,
            timeUnits: state.stats.timeUnits,
        }),
        shallowEqual,
    );

    const hasData = useMemo(() => data && data.some((v) => v > 0), [data]);

    const chartData = useMemo(() => {
        if (hasData) {
            return buildChartData(data, interval, timeUnits);
        }
        const wave = DEMO_WAVES[variant] || DEMO_WAVES.queries;
        return wave.map((v) => ({
            value: v,
            label: '',
        }));
    }, [data, interval, timeUnits, hasData, variant]);

    const gradId = gradientId(color);

    return (
        <ResponsiveContainer width="100%" height="100%" minHeight={16}>
            <AreaChart data={chartData} margin={{ top: 8, right: 0, bottom: 0, left: 0 }}>
                <defs>
                    <linearGradient id={gradId} x1="0" y1="0" x2="0" y2="1">
                        <stop offset="0%" stopColor={color} stopOpacity={hasData ? 0.35 : 0.16} />
                        <stop offset="65%" stopColor={color} stopOpacity={hasData ? 0.12 : 0.05} />
                        <stop offset="100%" stopColor={color} stopOpacity={0.01} />
                    </linearGradient>
                </defs>

                <YAxis domain={[0, 'auto']} hide />

                {hasData && (
                    <Tooltip content={<CustomTooltip />} cursor={{ stroke: color, strokeWidth: 1 }} />
                )}

                <Area
                    type="natural"
                    dataKey="value"
                    strokeWidth={hasData ? 1.5 : 1.2}
                    stroke={color}
                    strokeOpacity={hasData ? 0.85 : 0.4}
                    fill={`url(#${gradId})`}
                    isAnimationActive={false}
                    activeDot={hasData ? { r: 3, fill: color } : false}
                />
            </AreaChart>
        </ResponsiveContainer>
    );
});

Line.displayName = 'Line';

export default Line;
