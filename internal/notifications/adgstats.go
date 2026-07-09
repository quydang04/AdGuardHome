package notifications

import "time"

// StatsProvider exposes DNS query statistics for the bot menu.
type StatsProvider interface {
	GetCurrentStats() (numQueries, numBlocked, numSafeBrowsing, numParental uint64, avgProcessingTime float64)
}

// FilterListInfo describes a single filter list for display in bot messages.
type FilterListInfo struct {
	ID         uint64
	Name       string
	URL        string
	RulesCount int
	Enabled    bool
}

// FilterProvider exposes filter list summary data for the bot menu.
type FilterProvider interface {
	GetFilterSummary() (totalRules int, enabledBlockLists int, enabledAllowLists int)
	GetFilterDetails() (blockLists []FilterListInfo, allowLists []FilterListInfo)
}

// FilterManager extends FilterProvider with list management capabilities.
type FilterManager interface {
	FilterProvider
	AddFilterList(url string, name string, whitelist bool) error
	RemoveFilterList(url string, whitelist bool) error
	EnableFilterList(url string, enabled bool, whitelist bool) error
	RefreshFilters() (updated int, ok bool)
}

// QueryLogEntry represents a single recent DNS query for display.
type QueryLogEntry struct {
	Domain   string
	Client   string
	Blocked  bool
	Duration time.Duration
	Time     time.Time
}

// LogsProvider exposes recent DNS query log entries for the bot menu.
type LogsProvider interface {
	GetRecentQueries(limit int) []QueryLogEntry
}

// ProtectionProvider exposes the DNS protection toggle state for the bot menu
// and allows toggling it.
type ProtectionProvider interface {
	IsProtectionEnabled() bool
	SetProtectionEnabled(enabled bool) error
}

// YouTubeStatus summarizes the current state of the YouTube ad-blocking
// integration for display in bot messages.
type YouTubeStatus struct {
	// LastSyncStatus is a human-readable description of the last route
	// server sync attempt.
	LastSyncStatus string

	// LastSyncTime is when the last sync attempt happened.  The zero value
	// means no sync has run yet.
	LastSyncTime time.Time

	// HealthyIPs is the number of currently healthy route server IPs.
	HealthyIPs int

	// TotalIPs is the number of resolved route server IPs.
	TotalIPs int

	// BlockedRules is the number of active ad/tracking blocking rules.
	BlockedRules int

	// ActiveRewrites is the number of active DNS rewrites routing YouTube
	// traffic through the configured route server.
	ActiveRewrites int

	// Enabled indicates whether YouTube ad blocking is enabled in the
	// configuration.
	Enabled bool

	// Active indicates whether the manager is currently running.
	Active bool
}

// YouTubeProvider exposes the YouTube ad-blocking integration's status for
// the bot menu and allows toggling it on or off.
type YouTubeProvider interface {
	IsYouTubeBlockEnabled() bool
	SetYouTubeBlockEnabled(enabled bool) error
	GetYouTubeStatus() YouTubeStatus
}

// SetYouTubeProvider injects the YouTube ad-blocking provider.
func (m *Manager) SetYouTubeProvider(yp YouTubeProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.youtube = yp
}

// SetProviders injects data providers that are initialized after the Manager.
// Any provider may be nil if the corresponding module is not available.
func (m *Manager) SetProviders(sp StatsProvider, fp FilterProvider, pp ProtectionProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = sp
	m.filters = fp
	m.protection = pp

	// If the FilterProvider also implements FilterManager, store it.
	if fm, ok := fp.(FilterManager); ok {
		m.filterMgr = fm
	}
}

// SetFilterManager injects a filter manager for list management commands.
func (m *Manager) SetFilterManager(fm FilterManager) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.filterMgr = fm
	if m.filters == nil {
		m.filters = fm
	}
}

// SetLogsProvider injects the query log provider.
func (m *Manager) SetLogsProvider(lp LogsProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.logs = lp
}
