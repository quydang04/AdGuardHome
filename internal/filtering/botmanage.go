package filtering

import (
	"context"
	"fmt"
	"slices"

	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
)

// AddFilterList adds a new filter list by URL. If whitelist is true, it is
// added as an allowlist. The list is downloaded and validated before adding.
func (d *DNSFilter) AddFilterList(url string, name string, whitelist bool) error {
	if err := d.validateFilterURL(url); err != nil {
		return err
	}

	if d.filterExists(url) {
		return fmt.Errorf("filter with URL %q already exists", url)
	}

	filt := FilterYAML{
		Enabled: true,
		URL:     url,
		Name:    name,
		white:   whitelist,
		Filter: Filter{
			ID: d.idGen.next(),
		},
	}

	ok, err := d.update(&filt)
	if err != nil {
		return fmt.Errorf("fetch filter from %q: %w", url, err)
	}

	if !ok {
		return fmt.Errorf("filter at %q is invalid or empty", url)
	}

	if err = d.filterAdd(filt); err != nil {
		return fmt.Errorf("add filter %q: %w", url, err)
	}

	d.conf.ConfModifier.Apply(context.Background())
	d.EnableFilters(true)

	return nil
}

// RemoveFilterList removes a filter list by URL. If whitelist is true, it
// removes from the allowlist instead of the blocklist.
func (d *DNSFilter) RemoveFilterList(url string, whitelist bool) error {
	d.conf.filtersMu.Lock()

	filters := &d.conf.Filters
	if whitelist {
		filters = &d.conf.WhitelistFilters
	}

	delIdx := slices.IndexFunc(*filters, func(flt FilterYAML) bool {
		return flt.URL == url
	})

	if delIdx == -1 {
		d.conf.filtersMu.Unlock()
		return fmt.Errorf("filter with URL %q not found", url)
	}

	*filters = slices.Delete(*filters, delIdx, delIdx+1)
	d.conf.filtersMu.Unlock()

	d.conf.ConfModifier.Apply(context.Background())
	d.EnableFilters(true)

	return nil
}

// EnableFilterList enables or disables a filter list by URL.
func (d *DNSFilter) EnableFilterList(url string, enabled bool, whitelist bool) error {
	d.conf.filtersMu.Lock()

	filters := &d.conf.Filters
	if whitelist {
		filters = &d.conf.WhitelistFilters
	}

	idx := slices.IndexFunc(*filters, func(flt FilterYAML) bool {
		return flt.URL == url
	})

	if idx == -1 {
		d.conf.filtersMu.Unlock()
		return fmt.Errorf("filter with URL %q not found", url)
	}

	(*filters)[idx].Enabled = enabled
	d.conf.filtersMu.Unlock()

	d.conf.ConfModifier.Apply(context.Background())
	d.EnableFilters(true)

	return nil
}

// RefreshFilters triggers a forced refresh of all filter lists (both blocklists
// and allowlists). Returns the number of updated lists and whether the operation
// was able to acquire the refresh lock.
func (d *DNSFilter) RefreshFilters() (updated int, ok bool) {
	updated, _, ok = d.tryRefreshFilters(true, true, true)

	return updated, ok
}

// GetFilterDetails returns detailed information about all configured filter
// lists. This satisfies the notifications.FilterProvider interface extension.
func (d *DNSFilter) GetFilterDetails() (blockLists []notifications.FilterListInfo, allowLists []notifications.FilterListInfo) {
	d.conf.filtersMu.RLock()
	defer d.conf.filtersMu.RUnlock()

	blockLists = make([]notifications.FilterListInfo, 0, len(d.conf.Filters))
	for _, f := range d.conf.Filters {
		blockLists = append(blockLists, notifications.FilterListInfo{
			ID:         uint64(f.ID),
			Name:       f.Name,
			URL:        f.URL,
			RulesCount: f.RulesCount,
			Enabled:    f.Enabled,
		})
	}

	allowLists = make([]notifications.FilterListInfo, 0, len(d.conf.WhitelistFilters))
	for _, f := range d.conf.WhitelistFilters {
		allowLists = append(allowLists, notifications.FilterListInfo{
			ID:         uint64(f.ID),
			Name:       f.Name,
			URL:        f.URL,
			RulesCount: f.RulesCount,
			Enabled:    f.Enabled,
		})
	}

	return blockLists, allowLists
}
