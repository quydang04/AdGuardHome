import React from 'react';
import { withTranslation } from 'react-i18next';
import { TFunction } from 'i18next';

import Card from '../ui/Card';
import { formatNumber } from '../../helpers/helpers';
import { SystemInfoData } from '../../initialState';

interface SystemInfoProps {
    info: SystemInfoData | null | undefined;
    refreshButton: React.ReactNode;
    t: TFunction;
}

const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

const getUnitIndex = (bytes: number) => {
    let value = Math.max(bytes, 0);
    let unitIndex = 0;

    while (value >= 1024 && unitIndex < BYTE_UNITS.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }

    return unitIndex;
};

const formatBytes = (bytes: number, unitIndex: number) => {
    const divisor = 1024 ** unitIndex;
    const rawValue = divisor ? bytes / divisor : bytes;
    const rounded = unitIndex === 0 ? Math.round(rawValue) : Number(rawValue.toFixed(1));

    return `${formatNumber(rounded)} ${BYTE_UNITS[unitIndex]}`;
};

const formatPercentage = (value: number) => {
    if (!Number.isFinite(value)) {
        return '–';
    }

    return `${formatNumber(Number(value.toFixed(1)))}%`;
};

const formatUptime = (seconds: number) => {
    if (!seconds) {
        return '–';
    }

    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    const parts: string[] = [];

    if (days) {
        parts.push(`${days}d`);
    }

    if (hours || parts.length) {
        parts.push(`${hours}h`);
    }

    parts.push(`${minutes}m`);

    return parts.join(' ');
};

const renderCapacity = (current: number, total: number) => {
    if (!total) {
        return '–';
    }

    const unitIndex = getUnitIndex(total);

    return `${formatBytes(current, unitIndex)} / ${formatBytes(total, unitIndex)}`;
};

const renderUsage = (used: number, total: number, usagePercent: number) => {
    if (!total) {
        return '–';
    }

    const unitIndex = getUnitIndex(total);

    return `${formatBytes(used, unitIndex)} / ${formatBytes(total, unitIndex)} (${formatPercentage(usagePercent)})`;
};

const SystemInfo = ({ info, refreshButton, t }: SystemInfoProps) => (
    <Card title={t('system_overview')} refresh={refreshButton} bodyType="card-table">
        {!info ? (
            <div className="card-body">
                <p className="text-muted mb-0">{t('system_overview_unavailable')}</p>
            </div>
        ) : (
            <div className="table-responsive system-info__table">
                <table className="table card-table">
                    <tbody>
                        <tr>
                            <th scope="row">{t('system_overview_os')}</th>
                            <td>{info.osVersion || info.os || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_hostname')}</th>
                            <td>{info.hostname || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_architecture')}</th>
                            <td>{info.arch || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_cpu_model')}</th>
                            <td>{info.cpuModel || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_cpu_usage')}</th>
                            <td>{formatPercentage(info.cpuUsage)}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_cpu_cores')}</th>
                            <td>{info.numCpu ? formatNumber(info.numCpu) : '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_memory_usage')}</th>
                            <td>{renderUsage(info.memoryUsed, info.memoryTotal, info.memoryUsage)}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_memory_free')}</th>
                            <td>{renderCapacity(info.memoryFree, info.memoryTotal)}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_disk_usage')}</th>
                            <td>{renderUsage(info.diskUsed, info.diskTotal, info.diskUsage)}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_disk_free')}</th>
                            <td>{renderCapacity(info.diskFree, info.diskTotal)}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_disk_path')}</th>
                            <td>{info.diskPath || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_local_ip')}</th>
                            <td>{info.localIps && info.localIps.length > 0 ? info.localIps.join(', ') : '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_public_ip')}</th>
                            <td>{info.publicIp || '–'}</td>
                        </tr>
                        <tr>
                            <th scope="row">{t('system_overview_uptime')}</th>
                            <td>{formatUptime(info.uptimeSeconds)}</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        )}
    </Card>
);

export default withTranslation()(SystemInfo);
