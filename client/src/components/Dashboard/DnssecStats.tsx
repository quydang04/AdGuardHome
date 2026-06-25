import React from 'react';
import { useTranslation } from 'react-i18next';

import Card from '../ui/Card';

interface DnssecStatsProps {
    numDnssec: number;
    numDnsQueries: number;
    subtitle: string;
    refreshButton: React.ReactNode;
}

const DnssecStats = ({ numDnssec, numDnsQueries, subtitle, refreshButton }: DnssecStatsProps) => {
    const { t } = useTranslation();

    const percentage = numDnsQueries > 0
        ? ((numDnssec / numDnsQueries) * 100).toFixed(2)
        : '0';

    return (
        <Card title={t('dnssec')} subtitle={subtitle} refresh={refreshButton}>
            <div className="progress-stat">
                <p className="progress-stat__desc">{t('dnssec_desc')}</p>
                <div className="progress-stat__bar-wrapper">
                    <div className="progress-stat__bar">
                        <div
                            className="progress-stat__bar-fill progress-stat__bar-fill--blue"
                            style={{ width: `${percentage}%` }}
                        />
                    </div>
                </div>
                <div className="progress-stat__value progress-stat__value--blue">{percentage}%</div>
            </div>
        </Card>
    );
};

export default DnssecStats;
