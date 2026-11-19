import React from 'react';
import { useTranslation } from 'react-i18next';
import './EmptyCardState.css';

export const EmptyCardState = React.memo(() => {
    const { t } = useTranslation();

    return (
        <div className="card-empty-state">
            <svg
                className="card-empty-state__icon"
                viewBox="0 0 48 48"
                fill="none"
                stroke="currentColor"
                xmlns="http://www.w3.org/2000/svg">
                <circle cx="21" cy="21" r="14" strokeWidth="2.5" />
                <line x1="31" y1="31" x2="42" y2="42" strokeWidth="3" strokeLinecap="round" />
                <line x1="16" y1="16" x2="26" y2="26" strokeWidth="2" strokeLinecap="round" />
                <line x1="26" y1="16" x2="16" y2="26" strokeWidth="2" strokeLinecap="round" />
            </svg>
            <div className="card-empty-state__text">
                {t('no_stats_yet')}
            </div>
        </div>
    );
});

EmptyCardState.displayName = 'EmptyCardState';

export default EmptyCardState;
