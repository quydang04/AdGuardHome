package home

import (
	"context"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/acme"
	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// certExpiryCheckInterval is how often [tlsManager.certExpiryLoop] checks the
// active certificate's expiration.
const certExpiryCheckInterval = 24 * time.Hour

// defaultRenewBeforeDays is used if the configured threshold is invalid.
const defaultRenewBeforeDays = 14

// certExpiryLoop periodically checks the active certificate's expiration and
// either auto-renews it via ACME or sends a Telegram reminder, depending on
// configuration.  It's intended to be run as a goroutine.
func (m *tlsManager) certExpiryLoop(ctx context.Context) {
	defer slogutil.RecoverAndLog(ctx, m.logger)

	ticker := time.NewTicker(certExpiryCheckInterval)
	defer ticker.Stop()

	m.checkCertExpiry(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.checkCertExpiry(ctx)
		}
	}
}

// checkCertExpiry inspects the active certificate's expiration and acts on
// it: auto-renewing via ACME if configured to do so, or sending a Telegram
// reminder otherwise.  It logs any errors it encounters.
func (m *tlsManager) checkCertExpiry(ctx context.Context) {
	m.mu.Lock()
	enabled := m.extTLSConf.Enabled
	notAfter := m.status.NotAfter
	m.mu.Unlock()

	if !enabled || notAfter.IsZero() {
		return
	}

	snap := acmeConfigSnapshot()
	acmeCfg := toACMEConfigJSON(snap)
	accountKeyPEM := snap.AccountKeyPEM
	accountURI := snap.AccountURI

	renewBeforeDays := acmeCfg.RenewBeforeDays
	if renewBeforeDays <= 0 {
		renewBeforeDays = defaultRenewBeforeDays
	}

	daysLeft := int(time.Until(notAfter).Hours() / 24)
	if daysLeft > renewBeforeDays {
		return
	}

	if !acmeCfg.Enabled || !acmeCfg.AutoRenew {
		m.notifyCertReminder(ctx, acmeCfg.Domains, notAfter, daysLeft)

		return
	}

	m.logger.InfoContext(ctx, "certificate nearing expiry, auto-renewing", "days_left", daysLeft)

	res, err := m.issueCertificate(ctx, &acme.Request{
		Email:              acmeCfg.Email,
		Domains:            acmeCfg.Domains,
		Challenge:          acme.ChallengeType(acmeCfg.Challenge),
		CloudflareAPIToken: acmeCfg.CloudflareAPIToken,
		DNSResolvers:       acmeCfg.DNSResolvers,
		AccountKeyPEM:      accountKeyPEM,
		AccountURI:         accountURI,
	})
	if err != nil {
		m.recordACMEError(ctx, err)
		m.notifyRenewalResult(ctx, acmeCfg.Domains, time.Time{}, err)

		return
	}

	status, err := m.applyIssuedCertificate(ctx, res)

	config.Lock()
	if config.ACME == nil {
		config.ACME = defaultACMEConfig()
	}
	config.ACME.AccountKeyPEM = res.AccountKeyPEM
	config.ACME.AccountURI = res.AccountURI
	if err == nil {
		config.ACME.LastIssuedAt = time.Now()
		config.ACME.LastError = ""
	} else {
		config.ACME.LastError = err.Error()
	}
	config.Unlock()

	m.confModifier.Apply(ctx)

	if err != nil {
		m.logger.ErrorContext(ctx, "applying auto-renewed certificate", slogutil.KeyError, err)
		m.notifyRenewalResult(ctx, acmeCfg.Domains, time.Time{}, err)

		return
	}

	m.notifyRenewalResult(ctx, acmeCfg.Domains, status.NotAfter, nil)
}

// notifyCertReminder sends a Telegram reminder for a certificate that is
// nearing expiration but not configured for auto-renewal.  It's a no-op if
// no notifier is configured.
func (m *tlsManager) notifyCertReminder(
	ctx context.Context,
	domains []string,
	notAfter time.Time,
	daysLeft int,
) {
	n := globalContext.notifier
	if n == nil {
		return
	}

	n.NotifyCertExpiry(ctx, notifications.CertExpiryReminder{
		Domains:  domains,
		NotAfter: notAfter,
		DaysLeft: daysLeft,
	})
}

// notifyRenewalResult sends a Telegram notification about the outcome of an
// automatic ACME certificate renewal.  It's a no-op if no notifier is
// configured.
func (m *tlsManager) notifyRenewalResult(
	ctx context.Context,
	domains []string,
	notAfter time.Time,
	renewErr error,
) {
	n := globalContext.notifier
	if n == nil {
		return
	}

	n.NotifyCertRenewal(ctx, notifications.CertRenewalResult{
		Domains:  domains,
		NotAfter: notAfter,
		Err:      renewErr,
	})
}
