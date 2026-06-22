import { formatNumber } from './helpers';

export const BYTE_UNITS = ['B', 'KB', 'MB', 'GB', 'TB', 'PB'];

export const getUnitIndex = (bytes: number) => {
    let value = Math.max(bytes, 0);
    let unitIndex = 0;

    while (value >= 1024 && unitIndex < BYTE_UNITS.length - 1) {
        value /= 1024;
        unitIndex += 1;
    }

    return unitIndex;
};

export const formatBytes = (bytes: number, unitIndex: number) => {
    const divisor = 1024 ** unitIndex;
    const rawValue = divisor ? bytes / divisor : bytes;
    const rounded = unitIndex === 0 ? Math.round(rawValue) : Number(rawValue.toFixed(1));

    return `${formatNumber(rounded)} ${BYTE_UNITS[unitIndex]}`;
};

export const formatPercentage = (value: number) => {
    if (!Number.isFinite(value)) {
        return '–';
    }

    return `${formatNumber(Number(value.toFixed(1)))}%`;
};

export const formatUptime = (seconds: number) => {
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

export const renderCapacity = (current: number, total: number) => {
    if (!total) {
        return '–';
    }

    const unitIndex = getUnitIndex(total);

    return `${formatBytes(current, unitIndex)} / ${formatBytes(total, unitIndex)}`;
};

export const renderUsage = (used: number, total: number, usagePercent: number) => {
    if (!total) {
        return '–';
    }

    const unitIndex = getUnitIndex(total);

    return `${formatBytes(used, unitIndex)} / ${formatBytes(total, unitIndex)} (${formatPercentage(usagePercent)})`;
};
