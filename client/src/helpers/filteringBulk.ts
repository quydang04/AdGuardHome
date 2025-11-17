import { trimLinesAndRemoveEmpty } from './helpers';

export type BulkFilterEntry = {
    url: string;
    name?: string;
};

/**
 * Parses user-provided bulk filter input. Each line should contain a URL and an optional
 * name separated by a comma. Lines are trimmed, empty entries get removed.
 */
export const parseBulkFiltersInput = (input?: string): BulkFilterEntry[] => {
    if (!input) {
        return [];
    }

    const normalized = trimLinesAndRemoveEmpty(input);

    if (!normalized) {
        return [];
    }

    return normalized.split('\n').reduce((acc, rawLine) => {
        if (!rawLine) {
            return acc;
        }

        const commaIndex = rawLine.indexOf(',');

        const urlPart = commaIndex >= 0 ? rawLine.slice(0, commaIndex) : rawLine;
        const namePart = commaIndex >= 0 ? rawLine.slice(commaIndex + 1) : '';

        const url = urlPart.trim();

        if (!url) {
            return acc;
        }

        const name = namePart.trim();
        acc.push({ url, name });

        return acc;
    }, [] as BulkFilterEntry[]);
};
