package home

import (
	"context"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
)

type protectionAdapter struct {
	srv     *dnsforward.Server
	filters *filtering.DNSFilter
}

func (a *protectionAdapter) IsProtectionEnabled() bool {
	if a.srv == nil {
		return false
	}

	enabled, _ := a.srv.UpdatedProtectionStatus(context.Background())

	return enabled
}

func (a *protectionAdapter) SetProtectionEnabled(enabled bool) error {
	if a.filters == nil {
		return nil
	}

	a.filters.SetProtectionEnabled(enabled)

	return nil
}

// youtubeAdapter implements [notifications.YouTubeProvider] on top of the
// package-level YouTube ad-blocking config and manager.
type youtubeAdapter struct{}

func (youtubeAdapter) IsYouTubeBlockEnabled() bool {
	return getYoutubeConf().Enabled
}

func (youtubeAdapter) SetYouTubeBlockEnabled(enabled bool) error {
	func() {
		config.Lock()
		defer config.Unlock()

		if config.YouTube == nil {
			config.YouTube = defaultYoutubeConfig()
		}

		config.YouTube.Enabled = enabled
	}()

	globalContext.web.confModifier.Apply(context.Background())

	if ytManager != nil {
		ytManager.restart(context.Background())
	}

	return nil
}

func (youtubeAdapter) GetYouTubeStatus() (status notifications.YouTubeStatus) {
	cfg := getYoutubeConf()
	status.Enabled = cfg.Enabled

	if ytManager == nil {
		return status
	}

	ytManager.mu.Lock()
	defer ytManager.mu.Unlock()

	status.Active = ytManager.active
	status.HealthyIPs = len(ytManager.healthyIPs)
	status.TotalIPs = len(ytManager.allIPs)
	status.BlockedRules = ytManager.blockedRules
	status.ActiveRewrites = ytManager.activeRewrites
	status.LastSyncStatus = ytManager.lastSyncStatus
	status.LastSyncTime = ytManager.lastSyncTime

	return status
}

func injectNotificationProviders() {
	n := globalContext.notifier
	if n == nil {
		return
	}

	var sp notifications.StatsProvider
	if globalContext.stats != nil {
		sp = globalContext.stats
	}

	var fp notifications.FilterProvider
	if globalContext.filters != nil {
		fp = globalContext.filters
	}

	var pp notifications.ProtectionProvider
	if globalContext.dnsServer != nil {
		pp = &protectionAdapter{srv: globalContext.dnsServer, filters: globalContext.filters}
	}

	n.SetProviders(sp, fp, pp)

	// If filtering supports management operations, inject it.
	if globalContext.filters != nil {
		n.SetFilterManager(globalContext.filters)
	}

	// Inject query log provider.
	if globalContext.queryLog != nil {
		n.SetLogsProvider(globalContext.queryLog)
	}

	n.SetYouTubeProvider(youtubeAdapter{})
}
