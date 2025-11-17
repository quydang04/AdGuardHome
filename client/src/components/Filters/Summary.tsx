import React from 'react';

import { formatNumber } from '../../helpers/helpers';

type SummaryProps = {
    label: string;
    hint: string;
    total: number;
};

const Summary = ({ label, hint, total }: SummaryProps) => (
    <div className="mb-4">
        <div className="d-flex align-items-baseline">
            <h5 className="mb-0 mr-3">{label}</h5>
            <span className="h3 mb-0 font-weight-bold">{formatNumber(total)}</span>
        </div>
        <div className="text-muted small mt-1">{hint}</div>
    </div>
);

export default Summary;
