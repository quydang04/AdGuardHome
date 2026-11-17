import { describe, expect, test } from 'vitest';

import { areEqualVersions, isNewerVersion } from '../helpers/version';

describe('version helpers', () => {
    test('areEqualVersions handles optional v prefix', () => {
        expect(areEqualVersions('v1.0.37', '1.0.37')).toBe(true);
        expect(areEqualVersions('1.0.37', '1.0.37')).toBe(true);
        expect(areEqualVersions('1.0.37', '1.0.36')).toBe(false);
        expect(areEqualVersions('', '1.0.37')).toBe(false);
    });

    test('isNewerVersion avoids downgrade prompts and compares semver accurately', () => {
        expect(isNewerVersion('v1.0.37', '1.0.36')).toBe(false);
        expect(isNewerVersion('1.0.37', 'v1.0.36')).toBe(false);
        expect(isNewerVersion('v1.0.37', '1.0.37')).toBe(false);
        expect(isNewerVersion('1.0.36', '1.0.37')).toBe(true);
        expect(isNewerVersion('1.0.36', 'v1.0.37')).toBe(true);
        expect(isNewerVersion('v1.0.37-beta.1', '1.0.37')).toBe(true);
        expect(isNewerVersion('v1.0.37', '1.0.37-beta.2')).toBe(false);
        expect(isNewerVersion('v1.0.37-beta.1', 'v1.0.37-beta.2')).toBe(true);
        expect(isNewerVersion('1.0.35.1', '1.0.36')).toBe(true);
        expect(isNewerVersion('1.0.36', '1.0.35.1')).toBe(false);
        expect(isNewerVersion('', '1.0.37')).toBe(false);
        expect(isNewerVersion('1.0.37', '')).toBe(false);
    });
});
