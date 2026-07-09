package home

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/acme"
	"github.com/AdguardTeam/AdGuardHome/internal/aghalg"
	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// acmeConfigJSON is the JSON representation of the ACME ("SSL/TLS issue")
// configuration used by the settings HTTP API.
type acmeConfigJSON struct {
	Email              string    `json:"email"`
	Challenge          string    `json:"challenge"`
	CloudflareAPIToken string    `json:"cloudflare_api_token"`
	Domains            []string  `json:"domains"`
	LastIssuedAt       time.Time `json:"last_issued_at"`
	LastError          string    `json:"last_error"`
	RenewBeforeDays    int       `json:"renew_before_days"`
	Enabled            bool      `json:"enabled"`
	AutoRenew          bool      `json:"auto_renew"`
}

// toACMEConfigJSON converts c into its JSON representation.  c must not be
// nil.
func toACMEConfigJSON(c *acmeConfig) (j acmeConfigJSON) {
	return acmeConfigJSON{
		Enabled:            c.Enabled,
		Email:              c.Email,
		Domains:            c.Domains,
		Challenge:          c.Challenge,
		CloudflareAPIToken: c.CloudflareAPIToken,
		AutoRenew:          c.AutoRenew,
		RenewBeforeDays:    c.RenewBeforeDays,
		LastIssuedAt:       c.LastIssuedAt,
		LastError:          c.LastError,
	}
}

// validate returns an error if the ACME configuration in j is not valid
// enough to attempt an issuance.
func (j *acmeConfigJSON) validate() (err error) {
	if len(j.Domains) == 0 {
		return errors.Error("at least one domain is required")
	}

	switch acme.ChallengeType(j.Challenge) {
	case acme.ChallengeHTTP01:
		// Go on.
	case acme.ChallengeCloudflareDNS01:
		if j.CloudflareAPIToken == "" {
			return errors.Error("cloudflare api token is required for the dns-01-cloudflare challenge")
		}
	default:
		return fmt.Errorf("unsupported challenge type %q", j.Challenge)
	}

	if j.RenewBeforeDays <= 0 {
		return errors.Error("renew_before_days must be positive")
	}

	return nil
}

// handleACMEStatus is the handler for the GET /control/tls/acme/status HTTP
// API.
func (m *tlsManager) handleACMEStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config.RLock()
	resp := toACMEConfigJSON(config.ACME)
	config.RUnlock()

	aghhttp.WriteJSONResponseOK(ctx, m.logger, w, r, resp)
}

// handleACMEConfigure is the handler for the POST /control/tls/acme/configure
// HTTP API.  It only persists settings; [tlsManager.handleACMEIssue] is what
// actually obtains a certificate.
func (m *tlsManager) handleACMEConfigure(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := acmeConfigJSON{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, m.logger, r, w, http.StatusBadRequest, "decoding request: %s", err)

		return
	}

	if req.Enabled {
		if err = req.validate(); err != nil {
			aghhttp.ErrorAndLog(ctx, m.logger, r, w, http.StatusBadRequest, "%s", err)

			return
		}
	}

	config.Lock()
	config.ACME.Enabled = req.Enabled
	config.ACME.Email = req.Email
	config.ACME.Domains = req.Domains
	config.ACME.Challenge = req.Challenge
	if req.CloudflareAPIToken != "" {
		config.ACME.CloudflareAPIToken = req.CloudflareAPIToken
	}
	config.ACME.AutoRenew = req.AutoRenew
	config.ACME.RenewBeforeDays = req.RenewBeforeDays
	config.ACME.applyDefaults()
	resp := toACMEConfigJSON(config.ACME)
	config.Unlock()

	m.confModifier.Apply(ctx)

	aghhttp.WriteJSONResponseOK(ctx, m.logger, w, r, resp)
}

// acmeIssueResponse is the response for the POST /control/tls/acme/issue
// HTTP API.  Unlike the regular TLS status endpoint, it deliberately
// includes the freshly issued private key, base64-encoded like the regular
// TLS configuration API, so that the frontend can paste it straight into the
// encryption settings form.
type acmeIssueResponse struct {
	Status           *tlsConfigStatus `json:"status"`
	CertificateChain string           `json:"certificate_chain"`
	PrivateKey       string           `json:"private_key"`
}

// handleACMEIssue is the handler for the POST /control/tls/acme/issue HTTP
// API.  It obtains (or renews) a certificate via ACME using the persisted
// configuration, applies it as the active TLS certificate, persists the
// result, and returns the PEM data so that the frontend can populate the
// certificate/private key fields.
func (m *tlsManager) handleACMEIssue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	config.RLock()
	cfgJSON := toACMEConfigJSON(config.ACME)
	accountKeyPEM := config.ACME.AccountKeyPEM
	accountURI := config.ACME.AccountURI
	config.RUnlock()

	if err := cfgJSON.validate(); err != nil {
		aghhttp.ErrorAndLog(ctx, m.logger, r, w, http.StatusBadRequest, "%s", err)

		return
	}

	res, err := m.issueCertificate(ctx, &acme.Request{
		Email:              cfgJSON.Email,
		Domains:            cfgJSON.Domains,
		Challenge:          acme.ChallengeType(cfgJSON.Challenge),
		CloudflareAPIToken: cfgJSON.CloudflareAPIToken,
		AccountKeyPEM:      accountKeyPEM,
		AccountURI:         accountURI,
	})
	if err != nil {
		m.recordACMEError(ctx, err)
		aghhttp.ErrorAndLog(ctx, m.logger, r, w, http.StatusBadGateway, "issuing certificate: %s", err)

		return
	}

	status, err := m.applyIssuedCertificate(ctx, res)

	config.Lock()
	config.ACME.AccountKeyPEM = res.AccountKeyPEM
	config.ACME.AccountURI = res.AccountURI
	config.ACME.Enabled = true
	if err == nil {
		config.ACME.LastIssuedAt = time.Now()
		config.ACME.LastError = ""
	} else {
		config.ACME.LastError = err.Error()
	}
	config.Unlock()

	m.confModifier.Apply(ctx)

	if err != nil {
		aghhttp.ErrorAndLog(ctx, m.logger, r, w, http.StatusInternalServerError, "applying certificate: %s", err)

		return
	}

	aghhttp.WriteJSONResponseOK(ctx, m.logger, w, r, &acmeIssueResponse{
		CertificateChain: base64.StdEncoding.EncodeToString(res.CertificatePEM),
		PrivateKey:       base64.StdEncoding.EncodeToString(res.PrivateKeyPEM),
		Status:           status,
	})
}

// issueCertificate calls into m.acme, if configured.
func (m *tlsManager) issueCertificate(ctx context.Context, req *acme.Request) (res *acme.Result, err error) {
	if m.acme == nil {
		return nil, errors.Error("acme: not initialized")
	}

	return m.acme.Issue(ctx, req)
}

// recordACMEError persists err as the last ACME error.
func (m *tlsManager) recordACMEError(ctx context.Context, err error) {
	m.logger.ErrorContext(ctx, "issuing acme certificate", slogutil.KeyError, err)

	config.Lock()
	config.ACME.LastError = err.Error()
	config.Unlock()

	m.confModifier.Apply(ctx)
}

// applyIssuedCertificate installs res as the active TLS certificate,
// following the same validate-then-apply flow as
// [tlsManager.handleTLSConfigure].
func (m *tlsManager) applyIssuedCertificate(
	ctx context.Context,
	res *acme.Result,
) (status *tlsConfigStatus, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	newConf := m.extTLSConf.clone()
	newConf.Enabled = true
	newConf.CertificateChain = string(res.CertificatePEM)
	newConf.PrivateKey = string(res.PrivateKeyPEM)
	newConf.CertificatePath = ""
	newConf.PrivateKeyPath = ""

	err = m.validateTLSSettings(tlsConfigSettingsExt{tlsConfigSettings: *newConf})
	if err != nil {
		return nil, fmt.Errorf("validating issued certificate: %w", err)
	}

	status = &tlsConfigStatus{}
	err = m.loadTLSConfig(ctx, newConf, status)
	if err != nil {
		return status, fmt.Errorf("loading issued certificate: %w", err)
	}

	restartHTTPS := m.setConfig(ctx, *newConf, status, aghalg.NBNull)
	m.setCertFileTime(ctx)

	if restartHTTPS {
		go m.web.tlsConfigChanged(context.Background(), newConf)
	}

	m.logger.InfoContext(ctx, "applied acme-issued certificate", "not_after", status.NotAfter)

	return status, nil
}

// registerACMEWebHandlers registers HTTP handlers for the ACME/"SSL/TLS
// issue" configuration.
func (m *tlsManager) registerACMEWebHandlers() {
	m.httpReg.Register(http.MethodGet, "/control/tls/acme/status", m.handleACMEStatus)
	m.httpReg.Register(http.MethodPost, "/control/tls/acme/configure", m.handleACMEConfigure)
	m.httpReg.Register(http.MethodPost, "/control/tls/acme/issue", m.handleACMEIssue)
}
