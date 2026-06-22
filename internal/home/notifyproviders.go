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
}
