package filtering

import (
	"context"
	"os"
	"slices"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/schedule"
)

// resetMissingFilterTimestamps clears LastUpdated on enabled filters whose
// local cache files do not exist so tryRefreshFilters fetches them right away.
func (d *DNSFilter) resetMissingFilterTimestamps(list []FilterYAML) {
	if d.conf.DataDir == "" {
		return
	}

	for i := range list {
		if !list[i].Enabled {
			continue
		}

		if _, err := os.Stat(list[i].Path(d.conf.DataDir)); err != nil {
			list[i].LastUpdated = time.Time{}
		}
	}
}

// ApplyImported copies the user-editable settings from src, which is expected
// to be the configuration produced by a settings import, into the running
// filter, and re-initializes the filter lists.  It doesn't write the
// configuration file.
func (d *DNSFilter) ApplyImported(ctx context.Context, src *Config) {
	blockedServices := src.BlockedServices
	if blockedServices != nil && blockedServices.Schedule == nil {
		blockedServices = &BlockedServices{
			Schedule: schedule.EmptyWeekly(),
			IDs:      slices.Clone(blockedServices.IDs),
		}
	}

	func() {
		d.confMu.Lock()
		defer d.confMu.Unlock()

		c := d.conf
		c.ProtectionEnabled = src.ProtectionEnabled
		c.FiltersUpdateIntervalHours = src.FiltersUpdateIntervalHours
		c.FilteringEnabled = src.FilteringEnabled
		c.RewritesEnabled = src.RewritesEnabled
		c.SafeBrowsingEnabled = src.SafeBrowsingEnabled
		c.ParentalEnabled = src.ParentalEnabled
		c.BlockingMode = src.BlockingMode
		c.BlockingIPv4 = src.BlockingIPv4
		c.BlockingIPv6 = src.BlockingIPv6
		c.BlockedServices = blockedServices
		c.ParentalBlockHost = src.ParentalBlockHost
		c.SafeBrowsingBlockHost = src.SafeBrowsingBlockHost
		c.BlockedResponseTTL = src.BlockedResponseTTL
		c.SafeBrowsingCacheSize = src.SafeBrowsingCacheSize
		c.SafeSearchCacheSize = src.SafeSearchCacheSize
		c.ParentalCacheSize = src.ParentalCacheSize
		c.CacheTime = src.CacheTime
		c.Rewrites = cloneRewrites(src.Rewrites)
		c.SafeSearchConf = src.SafeSearchConf
	}()

	filters := slices.Clone(src.Filters)
	for i := range filters {
		filters[i].white = false
	}
	d.resetMissingFilterTimestamps(filters)

	whitelistFilters := slices.Clone(src.WhitelistFilters)
	for i := range whitelistFilters {
		whitelistFilters[i].white = true
	}
	d.resetMissingFilterTimestamps(whitelistFilters)

	d.conf.filtersMu.Lock()
	d.conf.Filters = filters
	d.conf.WhitelistFilters = whitelistFilters
	d.conf.UserRules = slices.Clone(src.UserRules)
	d.conf.filtersMu.Unlock()

	d.EnableFilters(true)

	// Download any enabled filter lists whose cache files are not yet on disk.
	go d.tryRefreshFilters(true, true, false)
}
