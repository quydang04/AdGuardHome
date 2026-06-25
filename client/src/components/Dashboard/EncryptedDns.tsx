import React from 'react';
import { useTranslation } from 'react-i18next';

import Card from '../ui/Card';

interface EncryptedDnsProps {
    numEncryptedDns: number;
    numDnsQueries: number;
    subtitle: string;
    refreshButton: React.ReactNode;
}

const EncryptedDns = ({ numEncryptedDns, numDnsQueries, subtitle, refreshButton }: EncryptedDnsProps) => {
    const { t } = useTranslation();

    const percentage = numDnsQueries > 0
        ? ((numEncryptedDns / numDnsQueries) * 100).toFixed(2)
        : '0';

    return (
        <Card title={t('encrypted_dns')} subtitle={subtitle} refresh={refreshButton}>
            <div className="progress-stat">
                <p className="progress-stat__desc">{t('encrypted_dns_desc')}</p>
                <div className="progress-stat__bar-wrapper">
                    <div className="progress-stat__bar">
                        <div
                            className="progress-stat__bar-fill progress-stat__bar-fill--green"
                            style={{ width: `${percentage}%` }}
                        />
                    </div>
                </div>
                <div className="progress-stat__value progress-stat__value--green">{percentage}%</div>
            </div>
        </Card>
    );
};

export default EncryptedDns;
