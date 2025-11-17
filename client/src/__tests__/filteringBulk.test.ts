import { describe, expect, it } from 'vitest';

import { parseBulkFiltersInput } from '../helpers/filteringBulk';

describe('parseBulkFiltersInput', () => {
    it('returns empty array for empty input', () => {
        expect(parseBulkFiltersInput('')).toStrictEqual([]);
    });

    it('ignores blank lines and trims values', () => {
        const input = '\n https://example.org/ads.txt \n\nhttps://example.org/tracking.txt  \n';
        expect(parseBulkFiltersInput(input)).toStrictEqual([
            { url: 'https://example.org/ads.txt', name: '' },
            { url: 'https://example.org/tracking.txt', name: '' },
        ]);
    });

    it('splits optional custom names by comma', () => {
        const input = 'https://example.org/ads.txt, Example Ads\nhttps://example.org/tracking.txt, Custom Name';
        expect(parseBulkFiltersInput(input)).toStrictEqual([
            { url: 'https://example.org/ads.txt', name: 'Example Ads' },
            { url: 'https://example.org/tracking.txt', name: 'Custom Name' },
        ]);
    });

    it('keeps commas inside custom names', () => {
        const input = 'https://example.org/ads.txt, Example, Ads';
        expect(parseBulkFiltersInput(input)).toStrictEqual([
            { url: 'https://example.org/ads.txt', name: 'Example, Ads' },
        ]);
    });
});
