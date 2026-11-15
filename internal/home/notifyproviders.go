package home

import (
	"context"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
	"github.com/AdguardTeam/golibs/errors"
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

// certAdapter implements [notifications.CertProvider] on top of the
// package-level ACME config and the TLS manager.
type certAdapter struct {
	web *webAPI
}

func (a certAdapter) GetCertStatus() (status notifications.CertStatus) {
	snap := acmeConfigSnapshot()
	status.Enabled = snap.Enabled
	status.Domains = snap.Domains
	status.Challenge = snap.Challenge
	status.AutoRenew = snap.AutoRenew
	status.LastIssuedAt = snap.LastIssuedAt
	status.LastError = snap.LastError

	if a.web != nil && a.web.tlsManager != nil {
		tlsMgr := a.web.tlsManager
		tlsMgr.mu.Lock()
		status.NotAfter = tlsMgr.status.NotAfter
		tlsMgr.mu.Unlock()
	}

	return status
}

func (a certAdapter) SetCertAutoRenew(enabled bool) error {
	func() {
		config.Lock()
		defer config.Unlock()

		if config.ACME == nil {
			config.ACME = defaultACMEConfig()
		}

		config.ACME.AutoRenew = enabled
	}()

	if a.web != nil && a.web.confModifier != nil {
		a.web.confModifier.Apply(context.Background())
	}

	return nil
}

func (a certAdapter) IssueCertificateNow() error {
	if a.web == nil || a.web.tlsManager == nil {
		return errors.Error("tls manager not available")
	}

	_, err := a.web.tlsManager.startACMEIssueJob(context.Background())

	return err
}

func injectNotificationProviders(web *webAPI) {
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

	n.SetCertProvider(certAdapter{web: web})
}
