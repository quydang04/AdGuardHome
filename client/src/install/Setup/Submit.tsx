import React from 'react';
import { Trans } from 'react-i18next';

import Controls from './Controls';
import { WebConfig } from './Settings';

type Props = {
    webConfig: WebConfig;
    openDashboard: (ip: string, port: number) => void;
};

export const Submit = ({ openDashboard, webConfig }: Props) => {
    return (
        <div className="setup__step">
            <div className="setup__group">
                <h1 className="setup__title">
                    <Trans>install_submit_title</Trans>
                </h1>

                <p className="setup__desc">
                    <Trans>install_submit_desc</Trans>
                </p>

                <div className="setup__mod-card">
                    <div className="setup__mod-icon">
                        <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#67b279" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                            <path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3z" />
                            <polyline points="9 12 11 14 15 10" />
                        </svg>
                    </div>
                    <div className="setup__mod-content">
                        <div className="setup__mod-title">
                            <span>AdGuard Home Mod</span>
                            <span className="setup__mod-badge">Flat Dark</span>
                        </div>
                        <div className="setup__mod-by">
                            <span>Tối ưu bởi <strong>quydangblog</strong></span>
                            <span className="setup__mod-dot">&bull;</span>
                            <a
                                href="http://quydang.name.vn"
                                target="_blank"
                                rel="noopener noreferrer"
                                className="setup__mod-link">
                                quydang.name.vn
                                <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round" strokeLinejoin="round">
                                    <path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6" />
                                    <polyline points="15 3 21 3 21 9" />
                                    <line x1="10" y1="14" x2="21" y2="3" />
                                </svg>
                            </a>
                        </div>
                    </div>
                </div>
            </div>

            <Controls openDashboard={openDashboard} ip={webConfig.ip} port={webConfig.port} />
        </div>
    );
};

export default Submit;
