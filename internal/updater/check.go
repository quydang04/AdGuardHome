package updater

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghalg"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/ioutil"
	"github.com/AdguardTeam/golibs/validate"
	"github.com/c2h5oh/datasize"
	"golang.org/x/mod/semver"
)

// TODO(a.garipov): Make configurable.
const versionCheckPeriod = 8 * time.Hour

// VersionInfo contains information about a new version.
type VersionInfo struct {
	NewVersion      string `json:"new_version,omitempty"`
	Announcement    string `json:"announcement,omitempty"`
	AnnouncementURL string `json:"announcement_url,omitempty"`
	// TODO(a.garipov): See if the frontend actually still cares about
	// nullability.
	CanAutoUpdate aghalg.NullBool `json:"can_autoupdate,omitempty"`
}

// maxVersionRespSize is the maximum length in bytes for version information
// response.
const maxVersionRespSize datasize.ByteSize = 64 * datasize.KB

// maxDockerVersionRespSize is the maximum length in bytes for Docker Hub tags
// API responses, which embed per-architecture image digests for each tag.
const maxDockerVersionRespSize datasize.ByteSize = 512 * datasize.KB

// VersionInfo downloads the latest version information.  If forceRecheck is
// false and there are cached results, those results are returned.
func (u *Updater) VersionInfo(ctx context.Context, forceRecheck bool) (vi VersionInfo, err error) {
	u.mu.Lock()
	defer u.mu.Unlock()

	now := time.Now()
	recheckTime := u.prevCheckTime.Add(versionCheckPeriod)
	if !forceRecheck && now.Before(recheckTime) {
		u.logger.DebugContext(ctx, "version info recheck is not required yet")

		return u.prevCheckResult, u.prevCheckError
	}

	vcu := u.versionCheckURL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, vcu, nil)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("constructing request to %s: %w", vcu, err)
	}
	req.Header.Set("Accept", "application/json")

	u.logger.DebugContext(ctx, "requesting version data", "url", vcu)

	resp, err := u.client.Do(req)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("sending http request to %s: %w", vcu, err)
	}
	defer func() { err = errors.WithDeferred(err, resp.Body.Close()) }()

	if resp.StatusCode != http.StatusOK {
		return VersionInfo{}, fmt.Errorf(
			"got status code %d, want %d",
			resp.StatusCode,
			http.StatusOK,
		)
	}

	respLimit := maxVersionRespSize.Bytes()
	if u.useDockerHub {
		respLimit = maxDockerVersionRespSize.Bytes()
	}

	r := ioutil.LimitReader(resp.Body, respLimit)

	// This use of ReadAll is safe, because we just limited the appropriate
	// ReadCloser.
	body, err := io.ReadAll(r)
	if err != nil {
		return VersionInfo{}, fmt.Errorf("reading response from %s: %w", vcu, err)
	}

	u.prevCheckTime = now
	if u.useDockerHub {
		u.prevCheckResult, u.prevCheckError = u.parseDockerTagsResponse(ctx, body)
	} else {
		u.prevCheckResult, u.prevCheckError = u.parseVersionResponse(ctx, body)
	}

	return u.prevCheckResult, u.prevCheckError
}

// parseVersionResponse parses version-related data and unmarshals it into the
// [VersionInfo] structure.
func (u *Updater) parseVersionResponse(
	ctx context.Context,
	data []byte,
) (vi VersionInfo, err error) {
	info := VersionInfo{
		CanAutoUpdate: aghalg.NBFalse,
	}
	versionJSON := map[string]string{
		"version":          "",
		"announcement":     "",
		"announcement_url": "",
	}
	err = json.Unmarshal(data, &versionJSON)
	if err != nil {
		return info, fmt.Errorf("version.json: %w", err)
	}

	for k, v := range versionJSON {
		err = validate.NotEmpty("version_json_value", v)
		if err != nil {
			return info, fmt.Errorf("bad value for %q key: %w", k, err)
		}
	}

	info.NewVersion = versionJSON["version"]
	info.Announcement = versionJSON["announcement"]
	info.AnnouncementURL = versionJSON["announcement_url"]

	packageURL, key, found := u.downloadURL(ctx, versionJSON)
	if !found {
		return info, fmt.Errorf("version.json: bad key %q: %w", key, errors.ErrNoValue)
	}

	isNewVersion := isNewerVersion(u.version, info.NewVersion)
	if isNewVersion {
		u.logger.InfoContext(
			ctx,
			"a new version is available",
			"current_version", u.version,
			"new_version", info.NewVersion,
		)
	} else {
		u.logger.DebugContext(
			ctx,
			"the current version is up-to-date",
			"current_version", u.version,
		)
	}

	info.CanAutoUpdate = aghalg.BoolToNullBool(isNewVersion && !u.isDocker)

	if !u.isDocker {
		u.newVersion = info.NewVersion
		u.packageURL = packageURL
	}

	return info, nil
}

// dockerHubTagsResponse represents the subset of Docker Hub v2 tags API used
// for container update checks.
type dockerHubTagsResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
}

// parseDockerTagsResponse parses Docker Hub tags JSON and extracts the highest
// stable release tag for the configured Docker repository.
func (u *Updater) parseDockerTagsResponse(
	ctx context.Context,
	data []byte,
) (vi VersionInfo, err error) {
	info := VersionInfo{
		CanAutoUpdate: aghalg.NBFalse,
	}

	var parsed dockerHubTagsResponse
	err = json.Unmarshal(data, &parsed)
	if err != nil {
		return info, fmt.Errorf("docker hub tags: %w", err)
	}

	var bestTag string
	var bestParts [4]int
	hasBest := false

	for _, r := range parsed.Results {
		parts, ok := parseReleaseTagParts(r.Name)
		if !ok {
			continue
		}

		if !hasBest || compareVersionParts(parts, bestParts) > 0 {
			bestTag = r.Name
			bestParts = parts
			hasBest = true
		}
	}

	if !hasBest {
		return info, fmt.Errorf("docker hub tags: no release version tag found")
	}

	normalizedVer := strings.TrimPrefix(strings.TrimSpace(bestTag), "v")
	info.NewVersion = normalizedVer
	info.Announcement = fmt.Sprintf("AdGuard Home v%s is now available on Docker Hub!", normalizedVer)
	info.AnnouncementURL = fmt.Sprintf("https://hub.docker.com/r/%s/tags", u.dockerRepo)

	if isNewerVersion(u.version, info.NewVersion) {
		u.logger.InfoContext(
			ctx,
			"a new docker image version is available",
			"current_version", u.version,
			"new_version", info.NewVersion,
			"repo", u.dockerRepo,
		)
	} else {
		u.logger.DebugContext(
			ctx,
			"the current docker image version is up-to-date",
			"current_version", u.version,
			"repo", u.dockerRepo,
		)
	}

	// Self-update inside a container is intentionally disabled; users update by
	// pulling the newer image from Docker Hub.
	u.newVersion = ""
	u.packageURL = ""

	return info, nil
}

// downloadURL returns the download URL for current build as well as its key in
// versionObj.  If the key is not found, it additionally prints an informative
// log message.
func (u *Updater) downloadURL(
	ctx context.Context,
	versionObj map[string]string,
) (dlURL, key string, ok bool) {
	if u.goarch == "arm" && u.goarm != "" {
		key = fmt.Sprintf("download_%s_%sv%s", u.goos, u.goarch, u.goarm)
	} else if isMIPS(u.goarch) && u.gomips != "" {
		key = fmt.Sprintf("download_%s_%s_%s", u.goos, u.goarch, u.gomips)
	} else {
		key = fmt.Sprintf("download_%s_%s", u.goos, u.goarch)
	}

	dlURL, ok = versionObj[key]
	if ok {
		return dlURL, key, true
	}

	keys := slices.Sorted(maps.Keys(versionObj))

	u.logger.ErrorContext(ctx, "key not found", "missing", key, "got", keys)

	return "", key, false
}

// isMIPS returns true if arch is any MIPS architecture.
func isMIPS(arch string) (ok bool) {
	switch arch {
	case
		"mips",
		"mips64",
		"mips64le",
		"mipsle":
		return true
	default:
		return false
	}
}

// isNewerVersion reports whether candidate represents a newer release than
// current.  When both strings are valid semantic versions or 4-part dotted
// versions (with or without a leading "v"), it compares them numerically so an
// older published release never triggers a downgrade prompt.  For non-semver
// build strings, it falls back to inequality ignoring an optional leading "v".
func isNewerVersion(current, candidate string) (newer bool) {
	curV, curOK := normalizeSemver(current)
	candV, candOK := normalizeSemver(candidate)
	if curOK && candOK {
		return semver.Compare(candV, curV) > 0
	}

	curParts, curPartsOK := parseReleaseTagParts(current)
	candParts, candPartsOK := parseReleaseTagParts(candidate)
	if curPartsOK && candPartsOK {
		return compareVersionParts(candParts, curParts) > 0
	}

	curTrimmed := strings.TrimPrefix(strings.TrimSpace(current), "v")
	candTrimmed := strings.TrimPrefix(strings.TrimSpace(candidate), "v")

	return candTrimmed != "" && candTrimmed != curTrimmed
}

// parseReleaseTagParts parses a 3-part or 4-part numeric release tag such as
// "v1.0.36" or "v1.0.35.1", rejecting prerelease/branch tags like
// "v1.0.35.1-demo", "latest", or "sha-abc123".
func parseReleaseTagParts(raw string) (parts [4]int, ok bool) {
	s := strings.TrimSpace(raw)
	s = strings.TrimPrefix(s, "v")
	if s == "" {
		return parts, false
	}

	segs := strings.Split(s, ".")
	if len(segs) != 3 && len(segs) != 4 {
		return parts, false
	}

	for i, seg := range segs {
		if seg == "" {
			return parts, false
		}
		for j := 0; j < len(seg); j++ {
			if seg[j] < '0' || seg[j] > '9' {
				return parts, false
			}
		}

		n, err := strconv.Atoi(seg)
		if err != nil {
			return parts, false
		}
		parts[i] = n
	}

	return parts, true
}

func compareVersionParts(a, b [4]int) int {
	for i := range a {
		if a[i] > b[i] {
			return 1
		}
		if a[i] < b[i] {
			return -1
		}
	}

	return 0
}

// normalizeSemver trims whitespace, adds an optional "v" prefix expected by
// [semver], and validates the resulting semantic version string.
func normalizeSemver(raw string) (v string, ok bool) {
	v = strings.TrimSpace(raw)
	if v == "" {
		return "", false
	}

	if !strings.HasPrefix(v, "v") {
		v = "v" + v
	}

	if !semver.IsValid(v) {
		return "", false
	}

	return v, true
}
