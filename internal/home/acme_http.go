package home

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	Email              string   `json:"email"`
	Challenge          string   `json:"challenge"`
	CloudflareAPIToken string   `json:"cloudflare_api_token"`
	Domains            []string `json:"domains"`
	DNSResolvers       []string `json:"dns_resolvers"`

	// LastIssuedAt is RFC 3339-formatted, or empty if a certificate has
	// never been issued via ACME.
	LastIssuedAt    string `json:"last_issued_at"`
	LastError       string `json:"last_error"`
	RenewBeforeDays int    `json:"renew_before_days"`
	Enabled         bool   `json:"enabled"`
	AutoRenew       bool   `json:"auto_renew"`
}

// acmeConfigSnapshot returns a copy of the current ACME configuration,
// substituting defaults if it hasn't been initialized yet (for example, on a
// config file predating this feature).  Safe for concurrent use.
func acmeConfigSnapshot() (snap *acmeConfig) {
	config.RLock()
	defer config.RUnlock()

	if config.ACME == nil {
		return defaultACMEConfig()
	}

	cp := *config.ACME

	return &cp
}

// toACMEConfigJSON converts c into its JSON representation.  If c is nil,
// the defaults are used instead.
func toACMEConfigJSON(c *acmeConfig) (j acmeConfigJSON) {
	if c == nil {
		c = defaultACMEConfig()
	}

	var lastIssuedAt string
	if !c.LastIssuedAt.IsZero() {
		lastIssuedAt = c.LastIssuedAt.Format(time.RFC3339)
	}

	return acmeConfigJSON{
		Enabled:            c.Enabled,
		Email:              c.Email,
		Domains:            c.Domains,
		Challenge:          c.Challenge,
		CloudflareAPIToken: c.CloudflareAPIToken,
		DNSResolvers:       c.DNSResolvers,
		AutoRenew:          c.AutoRenew,
		RenewBeforeDays:    c.RenewBeforeDays,
		LastIssuedAt:       lastIssuedAt,
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

	resp := toACMEConfigJSON(acmeConfigSnapshot())

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
	if config.ACME == nil {
		config.ACME = defaultACMEConfig()
	}
	config.ACME.Enabled = req.Enabled
	config.ACME.Email = req.Email
	config.ACME.Domains = req.Domains
	config.ACME.Challenge = req.Challenge
	if req.CloudflareAPIToken != "" {
		config.ACME.CloudflareAPIToken = req.CloudflareAPIToken
	}
	config.ACME.DNSResolvers = req.DNSResolvers
	config.ACME.AutoRenew = req.AutoRenew
	config.ACME.RenewBeforeDays = req.RenewBeforeDays
	config.ACME.applyDefaults()
	resp := toACMEConfigJSON(config.ACME)
	config.Unlock()

	m.confModifier.Apply(ctx)

	aghhttp.WriteJSONResponseOK(ctx, m.logger, w, r, resp)
}

// errACMEIssueInProgress is returned by [tlsManager.startACMEIssueJob] if an
// issuance is already running.
const errACMEIssueInProgress errors.Error = "a certificate issuance is already in progress"

// handleACMEIssue is the handler for the POST /control/tls/acme/issue HTTP
// API.  It starts a certificate issuance in the background and returns
// immediately; progress and the final result (including the issued
// certificate and key, base64-encoded like the regular TLS configuration
// API) are delivered over GET /control/tls/acme/issue/stream as Server-Sent
// Events.
func (m *tlsManager) handleACMEIssue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, err := m.startACMEIssueJob(ctx)
	if err != nil {
		status := http.StatusBadRequest
		if err == errACMEIssueInProgress {
			status = http.StatusConflict
		}

		aghhttp.ErrorAndLog(ctx, m.logger, r, w, status, "%s", err)

		return
	}

	w.WriteHeader(http.StatusAccepted)
}

// startACMEIssueJob validates the persisted ACME configuration and starts a
// new issuance job in the background, unless one is already running.  The
// returned job is also stored on m so that [tlsManager.handleACMEIssueStream]
// can find it.
func (m *tlsManager) startACMEIssueJob(ctx context.Context) (job *acmeJob, err error) {
	snap := acmeConfigSnapshot()
	cfgJSON := toACMEConfigJSON(snap)

	if err = cfgJSON.validate(); err != nil {
		return nil, err
	}

	m.acmeJobMu.Lock()
	defer m.acmeJobMu.Unlock()

	if m.acmeJob != nil && !m.acmeJob.isDone() {
		return nil, errACMEIssueInProgress
	}

	job = newAcmeJob()
	m.acmeJob = job

	go m.runACMEIssueJob(context.Background(), job, cfgJSON, snap.AccountKeyPEM, snap.AccountURI)

	return job, nil
}

// runACMEIssueJob performs the actual issuance, reporting progress to job,
// applying and persisting the result, and notifying Telegram of the outcome,
// the same as an automatic renewal.  It's intended to be run as a goroutine.
func (m *tlsManager) runACMEIssueJob(
	ctx context.Context,
	job *acmeJob,
	cfgJSON acmeConfigJSON,
	accountKeyPEM string,
	accountURI string,
) {
	defer slogutil.RecoverAndLog(ctx, m.logger)

	job.log("info", fmt.Sprintf("Starting certificate issuance for: %s", strings.Join(cfgJSON.Domains, ", ")))

	res, err := m.issueCertificate(ctx, &acme.Request{
		Email:              cfgJSON.Email,
		Domains:            cfgJSON.Domains,
		Challenge:          acme.ChallengeType(cfgJSON.Challenge),
		CloudflareAPIToken: cfgJSON.CloudflareAPIToken,
		DNSResolvers:       cfgJSON.DNSResolvers,
		AccountKeyPEM:      accountKeyPEM,
		AccountURI:         accountURI,
		Progress:           func(msg string) { job.log("info", msg) },
	})
	if err != nil {
		m.recordACMEError(ctx, err)
		job.log("error", err.Error())
		job.finish(&acmeJobResult{Error: err.Error()})
		m.notifyRenewalResult(ctx, cfgJSON.Domains, time.Time{}, err)

		return
	}

	job.log("info", "Applying certificate to AdGuard Home...")
	status, applyErr := m.applyIssuedCertificate(ctx, res)

	config.Lock()
	if config.ACME == nil {
		config.ACME = defaultACMEConfig()
	}
	config.ACME.AccountKeyPEM = res.AccountKeyPEM
	config.ACME.AccountURI = res.AccountURI
	config.ACME.Enabled = true
	if applyErr == nil {
		config.ACME.LastIssuedAt = time.Now()
		config.ACME.LastError = ""
	} else {
		config.ACME.LastError = applyErr.Error()
	}
	config.Unlock()

	m.confModifier.Apply(ctx)

	if applyErr != nil {
		job.log("error", applyErr.Error())
		job.finish(&acmeJobResult{Error: applyErr.Error()})
		m.notifyRenewalResult(ctx, cfgJSON.Domains, time.Time{}, applyErr)

		return
	}

	job.log("success", fmt.Sprintf("Certificate issued successfully, valid until %s", status.NotAfter.Format(time.RFC3339)))
	job.finish(&acmeJobResult{
		Success:          true,
		Status:           status,
		CertificateChain: base64.StdEncoding.EncodeToString(res.CertificatePEM),
		PrivateKey:       base64.StdEncoding.EncodeToString(res.PrivateKeyPEM),
	})
	m.notifyRenewalResult(ctx, cfgJSON.Domains, status.NotAfter, nil)
}

// handleACMEIssueStream is the handler for the GET
// /control/tls/acme/issue/stream HTTP API.  It streams the progress of the
// current (or most recently finished) ACME issuance job as Server-Sent
// Events, replaying any lines recorded before the client connected.
func (m *tlsManager) handleACMEIssueStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	m.acmeJobMu.Lock()
	job := m.acmeJob
	m.acmeJobMu.Unlock()

	if job == nil {
		http.Error(w, "no certificate issuance found", http.StatusNotFound)

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)

		return
	}

	h := w.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	h.Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	sent := 0
	for {
		lines, notify, done, result := job.snapshot()
		for ; sent < len(lines); sent++ {
			writeSSEEvent(w, "line", lines[sent])
		}
		flusher.Flush()

		if done {
			writeSSEEvent(w, "done", result)
			flusher.Flush()

			return
		}

		select {
		case <-notify:
			continue
		case <-ctx.Done():
			return
		case <-time.After(25 * time.Second):
			// Comment ping to keep the connection alive through proxies that
			// buffer or time out idle connections.
			_, _ = fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

// writeSSEEvent writes v as a single Server-Sent Events message of the given
// event type.  Errors are ignored, since the client may have disconnected.
func writeSSEEvent(w http.ResponseWriter, event string, v any) {
	data, err := json.Marshal(v)
	if err != nil {
		return
	}

	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data)
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
	if config.ACME == nil {
		config.ACME = defaultACMEConfig()
	}
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
	m.httpReg.Register(http.MethodGet, "/control/tls/acme/issue/stream", m.handleACMEIssueStream)
}
