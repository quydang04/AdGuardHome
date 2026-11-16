package filtering

import (
	"context"
	"slices"
)

// ApplyImported copies the user-editable settings from src, which is expected
// to be the configuration produced by a settings import, into the running
// filter, and re-initializes the filter lists.  It doesn't write the
// configuration file.
func (d *DNSFilter) ApplyImported(ctx context.Context, src *Config) {
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
		c.BlockedServices = src.BlockedServices
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

	d.conf.filtersMu.Lock()
	d.conf.Filters = slices.Clone(src.Filters)
	d.conf.WhitelistFilters = slices.Clone(src.WhitelistFilters)
	d.conf.UserRules = slices.Clone(src.UserRules)
	d.conf.filtersMu.Unlock()

	d.EnableFilters(true)
}
