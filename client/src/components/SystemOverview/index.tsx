import React, { useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Loading from '../ui/Loading';
import { formatNumber, formatTime } from '../../helpers/helpers';
import {
    formatBytes,
    formatPercentage,
    formatUptime,
    getUnitIndex,
    renderUsage,
} from '../../helpers/systemInfoHelpers';
import { SystemInfoData } from '../../initialState';
import './SystemOverview.css';

const RELOAD_INTERVAL_MS = 1000;

interface SystemOverviewProps {
    systemInfo: SystemInfoData | null;
    processing: boolean;
    getStats: () => void;
}

const getTier = (percent: number) => {
    if (percent >= 90) return 'danger';
    if (percent >= 70) return 'warning';
    return 'normal';
};

interface GaugeCardProps {
    label: string;
    percent: number;
    detail?: string;
}

const GaugeCard = ({ label, percent, detail }: GaugeCardProps) => {
    const safe = Math.min(Math.max(percent || 0, 0), 100);
    const tier = getTier(safe);
    const radius = 28;
    const circumference = 2 * Math.PI * radius;
    const strokeDashoffset = circumference - (safe / 100) * circumference;

    const labelLower = label.toLowerCase();
    let typeClass = '';
    let icon = null;

    if (labelLower.includes('cpu')) {
        typeClass = 'gauge-card--cpu';
        icon = (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <rect x="4" y="4" width="16" height="16" rx="2" />
                <rect x="9" y="9" width="6" height="6" />
                <line x1="9" y1="1" x2="9" y2="4" />
                <line x1="15" y1="1" x2="15" y2="4" />
                <line x1="9" y1="20" x2="9" y2="23" />
                <line x1="15" y1="20" x2="15" y2="23" />
                <line x1="20" y1="9" x2="23" y2="9" />
                <line x1="20" y1="15" x2="23" y2="15" />
                <line x1="1" y1="9" x2="4" y2="9" />
                <line x1="1" y1="15" x2="4" y2="15" />
            </svg>
        );
    } else if (labelLower.includes('ram')) {
        typeClass = 'gauge-card--ram';
        icon = (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M6 19v-3h12v3" />
                <path d="M6 16V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v11" />
                <line x1="10" y1="6" x2="14" y2="6" />
                <line x1="10" y1="10" x2="14" y2="10" />
                <line x1="6" y1="8" x2="7" y2="8" />
                <line x1="17" y1="8" x2="18" y2="8" />
                <line x1="6" y1="12" x2="7" y2="12" />
                <line x1="17" y1="12" x2="18" y2="12" />
            </svg>
        );
    } else {
        typeClass = 'gauge-card--disk';
        icon = (
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <ellipse cx="12" cy="5" rx="9" ry="3" />
                <path d="M3 5v6c0 1.66 4 3 9 3s9-1.34 9-3V5" />
                <path d="M3 11v6c0 1.66 4 3 9 3s9-1.34 9-3v-6" />
            </svg>
        );
    }

    return (
        <div className={`gauge-card gauge-card--${tier} ${typeClass}`}>
            <div className="gauge-card__left">
                <span className="gauge-card__label">{label}</span>
                <span className="gauge-card__value">{formatPercentage(safe)}</span>
                {detail && <span className="gauge-card__detail">{detail}</span>}
            </div>
            <div className="gauge-card__right">
                <div className="gauge-card__circle-container">
                    <svg className="gauge-card__circle-svg" width="68" height="68" viewBox="0 0 68 68">
                        <circle className="gauge-card__circle-bg" cx="34" cy="34" r={radius} />
                        <circle
                            className="gauge-card__circle-fill"
                            cx="34"
                            cy="34"
                            r={radius}
                            strokeDasharray={circumference}
                            strokeDashoffset={strokeDashoffset}
                        />
                    </svg>
                    <div className="gauge-card__icon-wrapper">
                        {icon}
                    </div>
                </div>
            </div>
        </div>
    );
};

const OS_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="2" y="3" width="20" height="14" rx="2" ry="2" />
        <line x1="8" y1="21" x2="16" y2="21" />
        <line x1="12" y1="17" x2="12" y2="21" />
    </svg>
);

const HOSTNAME_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M20.59 13.41l-7.17 7.17a2 2 0 0 1-2.83 0L2 12V2h10l8.59 8.59a2 2 0 0 1 0 2.82z" />
        <line x1="7" y1="7" x2="7.01" y2="7" />
    </svg>
);

const ARCH_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polygon points="12 2 2 7 12 12 22 7 12 2" />
        <polyline points="2 17 12 22 22 17" />
        <polyline points="2 12 12 17 22 12" />
    </svg>
);

const CPU_MODEL_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="4" y="4" width="16" height="16" rx="2" />
        <rect x="9" y="9" width="6" height="6" />
        <line x1="9" y1="1" x2="9" y2="4" />
        <line x1="15" y1="1" x2="15" y2="4" />
        <line x1="9" y1="20" x2="9" y2="23" />
        <line x1="15" y1="20" x2="15" y2="23" />
        <line x1="20" y1="9" x2="23" y2="9" />
        <line x1="20" y1="15" x2="23" y2="15" />
        <line x1="1" y1="9" x2="4" y2="9" />
        <line x1="1" y1="15" x2="4" y2="15" />
    </svg>
);

const CORES_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="3" y="3" width="7" height="7" />
        <rect x="14" y="3" width="7" height="7" />
        <rect x="14" y="14" width="7" height="7" />
        <rect x="3" y="14" width="7" height="7" />
    </svg>
);

const MEMORY_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M6 19v-3h12v3" />
        <path d="M6 16V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v11" />
        <line x1="10" y1="6" x2="14" y2="6" />
        <line x1="10" y1="10" x2="14" y2="10" />
        <line x1="6" y1="8" x2="7" y2="8" />
        <line x1="17" y1="8" x2="18" y2="8" />
    </svg>
);

const DISK_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <ellipse cx="12" cy="5" rx="9" ry="3" />
        <path d="M3 5v6c0 1.66 4 3 9 3s9-1.34 9-3V5" />
        <path d="M3 11v6c0 1.66 4 3 9 3s9-1.34 9-3v-6" />
    </svg>
);

const PATH_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M22 19a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h5l2 3h9a2 2 0 0 1 2 2z" />
    </svg>
);

const LOCAL_IP_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <rect x="16" y="16" width="6" height="6" rx="1" />
        <rect x="2" y="16" width="6" height="6" rx="1" />
        <rect x="9" y="2" width="6" height="6" rx="1" />
        <path d="M5 16v-3a1 1 0 0 1 1-1h12a1 1 0 0 1 1 1v3" />
        <path d="M12 12V8" />
    </svg>
);

const PUBLIC_IP_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <line x1="2" y1="12" x2="22" y2="12" />
        <path d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10z" />
    </svg>
);

const TIME_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="10" />
        <polyline points="12 6 12 12 16 14" />
    </svg>
);

const UPTIME_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
    </svg>
);

const CPU_USAGE_ICON = (
    <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <path d="M3 3v18h18" />
        <polyline points="18.7 8l-5.1 5.2-2.8-2.7L7 14.3" />
    </svg>
);

interface InfoRowData {
    label: string;
    value: string;
    icon?: React.ReactNode;
}

interface InfoCardProps {
    title: string;
    rows: InfoRowData[];
}

const InfoCard = ({ title, rows }: InfoCardProps) => (
    <div className="info-card">
        <div className="info-card__title">{title}</div>
        <ul className="info-card__rows">
            {rows.map((r) => (
                <li key={r.label} className="info-card__row">
                    <span className="info-card__row-label" style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
                        {r.icon && (
                            <span className="info-card__row-icon" style={{ display: 'flex', alignItems: 'center', opacity: 0.7 }}>
                                {r.icon}
                            </span>
                        )}
                        <span>{r.label}</span>
                    </span>
                    <span className="info-card__row-value">{r.value}</span>
                </li>
            ))}
        </ul>
    </div>
);

const SystemOverview = ({ systemInfo, processing, getStats }: SystemOverviewProps) => {
    const { t } = useTranslation();
    const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

    useEffect(() => {
        getStats();

        intervalRef.current = setInterval(() => {
            getStats();
        }, RELOAD_INTERVAL_MS);

        return () => {
            if (intervalRef.current) {
                clearInterval(intervalRef.current);
            }
        };
    }, []);

    const refreshButton = (
        <button
            type="button"
            className="btn btn-icon btn-outline-primary btn-sm"
            title={t('refresh_btn')}
            onClick={() => getStats()}>
            <svg className="icons icon12">
                <use xlinkHref="#refresh" />
            </svg>
        </button>
    );

    if (processing && !systemInfo) {
        return (
            <>
                <PageTitle title={t('system_overview')} containerClass="page-title--overview">{refreshButton}</PageTitle>
                <Loading />
            </>
        );
    }

    if (!systemInfo) {
        return (
            <>
                <PageTitle title={t('system_overview')} containerClass="page-title--overview">{refreshButton}</PageTitle>
                <div className="card">
                    <div className="card-body">
                        <p className="text-muted mb-0">{t('system_overview_unavailable')}</p>
                    </div>
                </div>
            </>
        );
    }

    const diskLabel = (() => {
        const path = systemInfo.diskPath || '';
        const match = path.match(/^([A-Z]:)/i);
        return match ? `Disk (${match[1]})` : 'Disk';
    })();

    const memUnitIdx = getUnitIndex(systemInfo.memoryTotal);
    const diskUnitIdx = getUnitIndex(systemInfo.diskTotal);

    let cpuDetail: string | undefined;
    if (systemInfo.cpuModel) {
        cpuDetail = `${systemInfo.cpuModel} · ${formatNumber(systemInfo.numCpu)} cores`;
    } else if (systemInfo.numCpu) {
        cpuDetail = `${formatNumber(systemInfo.numCpu)} cores`;
    }

    const memDetail = systemInfo.memoryTotal
        ? `${formatBytes(systemInfo.memoryUsed, memUnitIdx)} / ${formatBytes(systemInfo.memoryTotal, memUnitIdx)}`
        : undefined;

    const diskDetail = systemInfo.diskTotal
        ? `Total: ${formatBytes(systemInfo.diskTotal, diskUnitIdx)} · Free: ${formatBytes(systemInfo.diskFree, diskUnitIdx)}`
        : undefined;

    return (
        <>
            <PageTitle title={t('system_overview')} containerClass="page-title--overview">{refreshButton}</PageTitle>

            {/* Resource gauge cards */}
            <div className="system-overview__gauges">
                <GaugeCard label="CPU" percent={systemInfo.cpuUsage} detail={cpuDetail} />
                <GaugeCard label="RAM" percent={systemInfo.memoryUsage} detail={memDetail} />
                <GaugeCard label={diskLabel} percent={systemInfo.diskUsage} detail={diskDetail} />
            </div>

            {/* Info cards */}
            <div className="system-overview__info-grid">
                <InfoCard
                    title={t('system_info_hardware') || 'Hardware'}
                    rows={[
                        {
                            label: t('system_overview_os'),
                            value: (() => {
                                const os = systemInfo.osVersion || systemInfo.os || '–';
                                return systemInfo.isContainer ? `${os} (Docker)` : os;
                            })(),
                            icon: OS_ICON,
                        },
                        { label: t('system_overview_hostname'), value: systemInfo.hostname || '–', icon: HOSTNAME_ICON },
                        { label: t('system_overview_architecture'), value: systemInfo.arch || '–', icon: ARCH_ICON },
                        { label: t('system_overview_cpu_model'), value: systemInfo.cpuModel || '–', icon: CPU_MODEL_ICON },
                        { label: t('system_overview_cpu_cores'), value: systemInfo.numCpu ? formatNumber(systemInfo.numCpu) : '–', icon: CORES_ICON },
                    ]}
                />

                <InfoCard
                    title={t('system_info_storage') || 'Storage & Memory'}
                    rows={[
                        {
                            label: t('system_overview_memory_usage'),
                            value: renderUsage(systemInfo.memoryUsed, systemInfo.memoryTotal, systemInfo.memoryUsage),
                            icon: MEMORY_ICON,
                        },
                        {
                            label: t('system_overview_disk_usage'),
                            value: renderUsage(systemInfo.diskUsed, systemInfo.diskTotal, systemInfo.diskUsage),
                            icon: DISK_ICON,
                        },
                        { label: t('system_overview_disk_path'), value: systemInfo.diskPath || '–', icon: PATH_ICON },
                    ]}
                />

                <InfoCard
                    title={t('system_info_network') || 'Network'}
                    rows={[
                        {
                            label: t('system_overview_local_ip'),
                            value: systemInfo.localIps?.length ? systemInfo.localIps.join(', ') : '–',
                            icon: LOCAL_IP_ICON,
                        },
                        { label: t('system_overview_public_ip'), value: systemInfo.publicIp || '–', icon: PUBLIC_IP_ICON },
                    ]}
                />

                <InfoCard
                    title={t('system_info_runtime') || 'Runtime'}
                    rows={[
                        {
                            label: t('system_overview_system_time'),
                            value: systemInfo.systemTime
                                ? formatTime(systemInfo.systemTime, 'DD/MM/YYYY HH:mm:ss')
                                : '–',
                            icon: TIME_ICON,
                        },
                        { label: t('system_overview_uptime'), value: formatUptime(systemInfo.uptimeSeconds), icon: UPTIME_ICON },
                        { label: t('system_overview_cpu_usage'), value: formatPercentage(systemInfo.cpuUsage), icon: CPU_USAGE_ICON },
                    ]}
                />
            </div>
        </>
    );
};

export default SystemOverview;
