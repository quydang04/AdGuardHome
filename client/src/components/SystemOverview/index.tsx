import React, { useCallback, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';

import PageTitle from '../ui/PageTitle';
import Loading from '../ui/Loading';
import { formatNumber } from '../../helpers/helpers';
import {
    formatBytes,
    formatPercentage,
    formatUptime,
    getUnitIndex,
    renderCapacity,
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

    return (
        <div className={`gauge-card gauge-card--${tier}`}>
            <div className="gauge-card__header">
                <span className="gauge-card__label">{label}</span>
            </div>
            <span className="gauge-card__value">{formatPercentage(safe)}</span>
            <div className="gauge-card__bar">
                <div className="gauge-card__bar-fill" style={{ width: `${safe}%` }} />
            </div>
            {detail && <span className="gauge-card__detail">{detail}</span>}
        </div>
    );
};

interface InfoRowData {
    label: string;
    value: string;
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
                    <span className="info-card__row-label">{r.label}</span>
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
                <PageTitle title={t('system_overview')}>{refreshButton}</PageTitle>
                <Loading />
            </>
        );
    }

    if (!systemInfo) {
        return (
            <>
                <PageTitle title={t('system_overview')}>{refreshButton}</PageTitle>
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

    const cpuDetail = systemInfo.cpuModel
        ? `${systemInfo.cpuModel} · ${formatNumber(systemInfo.numCpu)} cores`
        : systemInfo.numCpu
          ? `${formatNumber(systemInfo.numCpu)} cores`
          : undefined;

    const memDetail = systemInfo.memoryTotal
        ? `${formatBytes(systemInfo.memoryUsed, memUnitIdx)} / ${formatBytes(systemInfo.memoryTotal, memUnitIdx)}`
        : undefined;

    const diskDetail = systemInfo.diskTotal
        ? `Total: ${formatBytes(systemInfo.diskTotal, diskUnitIdx)} · Free: ${formatBytes(systemInfo.diskFree, diskUnitIdx)}`
        : undefined;

    return (
        <>
            <PageTitle title={t('system_overview')}>{refreshButton}</PageTitle>

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
                        { label: t('system_overview_os'), value: systemInfo.osVersion || systemInfo.os || '–' },
                        { label: t('system_overview_hostname'), value: systemInfo.hostname || '–' },
                        { label: t('system_overview_architecture'), value: systemInfo.arch || '–' },
                        { label: t('system_overview_cpu_model'), value: systemInfo.cpuModel || '–' },
                        { label: t('system_overview_cpu_cores'), value: systemInfo.numCpu ? formatNumber(systemInfo.numCpu) : '–' },
                    ]}
                />

                <InfoCard
                    title={t('system_info_storage') || 'Storage & Memory'}
                    rows={[
                        {
                            label: t('system_overview_memory_usage'),
                            value: renderUsage(systemInfo.memoryUsed, systemInfo.memoryTotal, systemInfo.memoryUsage),
                        },
                        {
                            label: t('system_overview_memory_free'),
                            value: renderCapacity(systemInfo.memoryFree, systemInfo.memoryTotal),
                        },
                        {
                            label: t('system_overview_disk_usage'),
                            value: renderUsage(systemInfo.diskUsed, systemInfo.diskTotal, systemInfo.diskUsage),
                        },
                        {
                            label: t('system_overview_disk_free'),
                            value: renderCapacity(systemInfo.diskFree, systemInfo.diskTotal),
                        },
                        { label: t('system_overview_disk_path'), value: systemInfo.diskPath || '–' },
                    ]}
                />

                <InfoCard
                    title={t('system_info_network') || 'Network'}
                    rows={[
                        {
                            label: t('system_overview_local_ip'),
                            value: systemInfo.localIps?.length ? systemInfo.localIps.join(', ') : '–',
                        },
                        { label: t('system_overview_public_ip'), value: systemInfo.publicIp || '–' },
                    ]}
                />

                <InfoCard
                    title={t('system_info_runtime') || 'Runtime'}
                    rows={[
                        { label: t('system_overview_uptime'), value: formatUptime(systemInfo.uptimeSeconds) },
                        { label: t('system_overview_cpu_usage'), value: formatPercentage(systemInfo.cpuUsage) },
                    ]}
                />
            </div>
        </>
    );
};

export default SystemOverview;
