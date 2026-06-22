package filtering

// GetFilterSummary returns summary counts of all filter rules and enabled lists.
func (d *DNSFilter) GetFilterSummary() (totalRules int, enabledBlockLists int, enabledAllowLists int) {
	d.conf.filtersMu.RLock()
	defer d.conf.filtersMu.RUnlock()

	for _, f := range d.conf.Filters {
		if f.Enabled {
			enabledBlockLists++
			totalRules += f.RulesCount
		}
	}

	for _, f := range d.conf.WhitelistFilters {
		if f.Enabled {
			enabledAllowLists++
			totalRules += f.RulesCount
		}
	}

	return totalRules, enabledBlockLists, enabledAllowLists
}
