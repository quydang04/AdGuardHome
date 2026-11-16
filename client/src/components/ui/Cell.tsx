import React from 'react';

import LogsSearchLink from './LogsSearchLink';
import { formatNumber } from '../../helpers/helpers';
import './Cell.css';

interface CellProps {
    value: number;
    percent: number;
    color: string;
    search?: string;
    onSearchRedirect?: (...args: unknown[]) => string;
}

const Cell = ({ value, percent, color, search }: CellProps) => (
    <div className="stats-cell">
        <div className="stats-cell__value">
            <strong>
                {search ? <LogsSearchLink search={search}>{formatNumber(value)}</LogsSearchLink> : formatNumber(value)}
            </strong>

            <span className="stats-cell__percent">({percent}%)</span>
        </div>

        <div className="stats-cell__bar">
            <div
                className="stats-cell__bar-fill"
                style={{
                    width: `${Math.min(percent, 100)}%`,
                    backgroundColor: color,
                }}
            />
        </div>
    </div>
);

export default Cell;
