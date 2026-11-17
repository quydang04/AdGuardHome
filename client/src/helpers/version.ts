/**
 * Checks if versions are equal.
 * Please note, that this method strips the "v" prefix.
 *
 * @param left {string} - left version
 * @param right {string} - right version
 * @return {boolean} true if versions are equal
 */
export const areEqualVersions = (left: any, right: any) => {
    if (!left || !right) {
        return false;
    }

    const leftVersion = left.replace(/^v/, '');
    const rightVersion = right.replace(/^v/, '');
    return leftVersion === rightVersion;
};

type ParsedSemver = {
    major: number;
    minor: number;
    patch: number;
    extra: number;
    prerelease: string[];
};

const parseSemver = (raw: any): ParsedSemver | null => {
    if (typeof raw !== 'string') {
        return null;
    }

    const trimmed = raw.trim().replace(/^v/, '');
    const match = /^(\d+)\.(\d+)\.(\d+)(?:\.(\d+))?(?:-([0-9A-Za-z.-]+))?(?:\+[0-9A-Za-z.-]+)?$/.exec(trimmed);
    if (!match) {
        return null;
    }

    return {
        major: Number(match[1]),
        minor: Number(match[2]),
        patch: Number(match[3]),
        extra: match[4] ? Number(match[4]) : 0,
        prerelease: match[5] ? match[5].split('.') : [],
    };
};

const comparePrereleaseIdentifiers = (a: string, b: string): number => {
    const aIsNum = /^\d+$/.test(a);
    const bIsNum = /^\d+$/.test(b);

    if (aIsNum && bIsNum) {
        return Number(a) - Number(b);
    }
    if (aIsNum) {
        return -1;
    }
    if (bIsNum) {
        return 1;
    }

    return a < b ? -1 : a > b ? 1 : 0;
};

/**
 * Checks whether candidate version is strictly newer than current version.
 * Falls back to inequality (ignoring leading "v") for non-semver builds.
 */
export const isNewerVersion = (current: any, candidate: any): boolean => {
    if (typeof candidate !== 'string' || !candidate.trim()) {
        return false;
    }
    if (typeof current !== 'string' || !current.trim()) {
        return false;
    }

    const cur = parseSemver(current);
    const cand = parseSemver(candidate);
    if (!cur || !cand) {
        return !areEqualVersions(current, candidate);
    }

    if (cand.major !== cur.major) {
        return cand.major > cur.major;
    }
    if (cand.minor !== cur.minor) {
        return cand.minor > cur.minor;
    }
    if (cand.patch !== cur.patch) {
        return cand.patch > cur.patch;
    }
    if (cand.extra !== cur.extra) {
        return cand.extra > cur.extra;
    }

    const curPre = cur.prerelease;
    const candPre = cand.prerelease;
    if (candPre.length === 0 && curPre.length === 0) {
        return false;
    }
    if (candPre.length === 0) {
        return true;
    }
    if (curPre.length === 0) {
        return false;
    }

    const len = Math.max(candPre.length, curPre.length);
    for (let i = 0; i < len; i += 1) {
        if (i >= curPre.length) {
            return true;
        }
        if (i >= candPre.length) {
            return false;
        }
        const cmp = comparePrereleaseIdentifiers(candPre[i], curPre[i]);
        if (cmp !== 0) {
            return cmp > 0;
        }
    }

    return false;
};
