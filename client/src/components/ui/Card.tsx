import React from 'react';

import './Card.css';

interface CardProps {
    id?: string;
    title?: string;
    subtitle?: string;
    bodyType?: string;
    type?: string;
    refresh?: React.ReactNode;
    headerExtra?: React.ReactNode;
    children: React.ReactNode;
}

const Card = ({ type, id, title, subtitle, refresh, headerExtra, bodyType, children }: CardProps) => (
    <div className={type ? `card ${type}` : 'card'} id={id || ''}>
        {(title || subtitle || headerExtra || refresh) && (
            <div className="card-header">
                <div className="card-inner">
                    {title && <div className="card-title">{title}</div>}

                    {subtitle && <div className="card-subtitle" dangerouslySetInnerHTML={{ __html: subtitle }} />}
                </div>

                <div className="card-options">
                    {headerExtra && <div className="card-header-extra">{headerExtra}</div>}
                    {refresh}
                </div>
            </div>
        )}

        <div className={bodyType || 'card-body'}>{children}</div>
    </div>
);

export default Card;
