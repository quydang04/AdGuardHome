import { trimLinesAndRemoveEmpty } from './helpers';

export type BulkFilterEntry = {
    url: string;
    name?: string;
};

const splitLine = (line: string) => line.split(',').map((part) => part.trim());

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

    return normalized.split('\n').reduce<BulkFilterEntry[]>((acc, rawLine) => {
        if (!rawLine) {
            return acc;
        }

        const [urlPart = '', ...nameParts] = splitLine(rawLine);
        const url = urlPart.trim();

        if (!url) {
            return acc;
        }

        const name = nameParts.join(',').trim();
        acc.push({ url, name });

        return acc;
    }, []);
};
