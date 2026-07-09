// Package acme implements automatic issuance and renewal of TLS certificates
// via the ACME protocol (RFC 8555), backing AdGuard Home's "SSL/TLS issue"
// settings.  It obtains certificates from Let's Encrypt without relying on
// any external tooling (such as acme.sh).
package acme

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AdguardTeam/golibs/errors"
	"github.com/go-acme/lego/v4/certcrypto"
	"github.com/go-acme/lego/v4/certificate"
	"github.com/go-acme/lego/v4/challenge/dns01"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/go-acme/lego/v4/lego"
	"github.com/go-acme/lego/v4/providers/dns/cloudflare"
	"github.com/go-acme/lego/v4/registration"
)

// ChallengeType is the ACME challenge method used to prove domain control.
type ChallengeType string

// Challenge types supported by [Manager.Issue].
const (
	// ChallengeHTTP01 proves control of a domain by serving a token from a
	// well-known path over plain HTTP.  It requires the domain to already
	// resolve to this server's public IP address, and does not support
	// wildcard domains.
	ChallengeHTTP01 ChallengeType = "http-01"

	// ChallengeCloudflareDNS01 proves control of a domain by creating a TXT
	// record via the Cloudflare API.  It supports wildcard domains and does
	// not require inbound HTTP access.
	ChallengeCloudflareDNS01 ChallengeType = "dns-01-cloudflare"
)

// Request describes a certificate to issue or renew.  Requesting a
// certificate for domains that already have a valid certificate from the
// same ACME account is how renewal works; ACME has no separate operation for
// it.
type Request struct {
	// Email is the contact address used for the ACME account and for expiry
	// notices sent by the CA.
	Email string

	// Domains are the domain names to request a certificate for.  The first
	// entry becomes the certificate's Common Name.  Must not be empty.
	Domains []string

	// Challenge is the validation method used to prove control of Domains.
	Challenge ChallengeType

	// CloudflareAPIToken is the Cloudflare API token used for
	// [ChallengeCloudflareDNS01].  It must have Zone:DNS:Edit permission for
	// the requested domains' zones.
	CloudflareAPIToken string

	// AccountKeyPEM is a previously persisted ACME account private key.  If
	// empty, a new key is generated and returned in [Result.AccountKeyPEM].
	AccountKeyPEM string

	// AccountURI is a previously persisted ACME account resource URI.  If
	// set together with AccountKeyPEM, Issue reuses the existing account
	// instead of registering a new one.
	AccountURI string

	// DNSResolvers are the nameservers (as "host" or "host:port", port
	// defaults to 53) used to check that a DNS-01 TXT record has propagated
	// before asking the CA to validate it.  Only used for
	// [ChallengeCloudflareDNS01].  If empty, the host's own system resolver
	// (e.g. /etc/resolv.conf) is used, same as the rest of AdGuard Home.
	DNSResolvers []string
}

// Result is the outcome of a successful certificate issuance or renewal.
type Result struct {
	// CertificatePEM is the PEM-encoded certificate chain (leaf followed by
	// intermediates), ready to use as AdGuard Home's certificate chain.
	CertificatePEM []byte

	// PrivateKeyPEM is the PEM-encoded private key matching CertificatePEM.
	PrivateKeyPEM []byte

	// AccountKeyPEM is the ACME account's private key.  Persist it and pass
	// it back in the next [Request] so that renewals reuse the same
	// account instead of registering a new one every time.
	AccountKeyPEM string

	// AccountURI is the ACME account resource URI.  Persist it alongside
	// AccountKeyPEM.
	AccountURI string

	// NotAfter is the new certificate's expiration time.
	NotAfter time.Time
}

// Manager issues and renews TLS certificates via ACME.
type Manager struct {
	logger *slog.Logger
	http   *httpProvider
}

// NewManager returns a new *Manager.  logger must not be nil.
func NewManager(logger *slog.Logger) (m *Manager) {
	return &Manager{
		logger: logger,
		http:   newHTTPProvider(),
	}
}

// ChallengeHandler returns the HTTP handler that serves ACME http-01 key
// authorizations.  Register it on AdGuard Home's main HTTP server at
// [http01.PathPrefix], without any authentication middleware, since Let's
// Encrypt's validation servers must be able to reach it over plain HTTP
// without a login session.  This lets AdGuard Home answer challenges on the
// port(s) it already listens on, instead of opening a dedicated listener
// that would conflict with them.
func (m *Manager) ChallengeHandler() (h http.Handler) {
	return m.http
}

// Issue obtains a certificate for req, registering a new ACME account first
// if req doesn't already carry one.
func (m *Manager) Issue(ctx context.Context, req *Request) (res *Result, err error) {
	if len(req.Domains) == 0 {
		return nil, errors.Error("acme: no domains specified")
	}

	m.logger.InfoContext(ctx, "issuing certificate", "domains", req.Domains, "challenge", req.Challenge)

	user := &acmeUser{email: req.Email}
	if req.AccountKeyPEM != "" {
		user.key, err = parseECPrivateKey(req.AccountKeyPEM)
		if err != nil {
			return nil, fmt.Errorf("acme: parsing account key: %w", err)
		}

		if req.AccountURI != "" {
			user.reg = &registration.Resource{URI: req.AccountURI}
		}
	} else {
		user.key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
		if err != nil {
			return nil, fmt.Errorf("acme: generating account key: %w", err)
		}
	}

	cfg := lego.NewConfig(user)
	cfg.Certificate.KeyType = certcrypto.RSA2048

	client, err := lego.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("acme: creating client: %w", err)
	}

	err = m.setChallengeProvider(client, req)
	if err != nil {
		return nil, err
	}

	if user.reg == nil {
		user.reg, err = client.Registration.Register(registration.RegisterOptions{
			TermsOfServiceAgreed: true,
		})
		if err != nil {
			return nil, fmt.Errorf("acme: registering account: %w", err)
		}
	}

	cert, err := client.Certificate.Obtain(certificate.ObtainRequest{
		Domains: req.Domains,
		Bundle:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("acme: obtaining certificate: %w", err)
	}

	notAfter, err := certNotAfter(cert.Certificate)
	if err != nil {
		return nil, fmt.Errorf("acme: parsing issued certificate: %w", err)
	}

	accountKeyPEM, err := marshalECPrivateKey(user.key)
	if err != nil {
		return nil, fmt.Errorf("acme: marshalling account key: %w", err)
	}

	m.logger.InfoContext(ctx, "issued certificate", "domains", req.Domains, "not_after", notAfter)

	return &Result{
		CertificatePEM: cert.Certificate,
		PrivateKeyPEM:  cert.PrivateKey,
		AccountKeyPEM:  accountKeyPEM,
		AccountURI:     user.reg.URI,
		NotAfter:       notAfter,
	}, nil
}

// setChallengeProvider configures client to solve challenges using the
// method requested in req.
func (m *Manager) setChallengeProvider(client *lego.Client, req *Request) (err error) {
	switch req.Challenge {
	case ChallengeHTTP01:
		return client.Challenge.SetHTTP01Provider(m.http)
	case ChallengeCloudflareDNS01:
		if req.CloudflareAPIToken == "" {
			return errors.Error("acme: cloudflare api token is required for dns-01-cloudflare challenge")
		}

		cfCfg := cloudflare.NewDefaultConfig()
		cfCfg.AuthToken = req.CloudflareAPIToken

		var provider *cloudflare.DNSProvider
		provider, err = cloudflare.NewDNSProviderConfig(cfCfg)
		if err != nil {
			return fmt.Errorf("acme: creating cloudflare dns provider: %w", err)
		}

		var opts []dns01.ChallengeOption
		if len(req.DNSResolvers) > 0 {
			opts = append(opts, dns01.AddRecursiveNameservers(req.DNSResolvers))
		}

		return client.Challenge.SetDNS01Provider(provider, opts...)
	default:
		return fmt.Errorf("acme: unsupported challenge type %q", req.Challenge)
	}
}

// acmeUser implements [registration.User], the account identity lego uses to
// interact with the ACME server.
type acmeUser struct {
	email string
	key   crypto.PrivateKey
	reg   *registration.Resource
}

// GetEmail implements the [registration.User] interface for *acmeUser.
func (u *acmeUser) GetEmail() (email string) { return u.email }

// GetRegistration implements the [registration.User] interface for
// *acmeUser.
func (u *acmeUser) GetRegistration() (reg *registration.Resource) { return u.reg }

// GetPrivateKey implements the [registration.User] interface for *acmeUser.
func (u *acmeUser) GetPrivateKey() (key crypto.PrivateKey) { return u.key }

// httpProvider implements [challenge.Provider] for the http-01 challenge by
// serving key authorizations from AdGuard Home's existing web server instead
// of opening a dedicated listener, since AdGuard Home already owns port
// 80/443 by default and a standalone ACME listener would conflict with it.
type httpProvider struct {
	mu     sync.Mutex
	tokens map[string]string
}

// newHTTPProvider returns a new *httpProvider.
func newHTTPProvider() (p *httpProvider) {
	return &httpProvider{tokens: map[string]string{}}
}

// Present implements the [challenge.Provider] interface for *httpProvider.
func (p *httpProvider) Present(_, token, keyAuth string) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.tokens[token] = keyAuth

	return nil
}

// CleanUp implements the [challenge.Provider] interface for *httpProvider.
func (p *httpProvider) CleanUp(_, token, _ string) (err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.tokens, token)

	return nil
}

// type check
var _ http.Handler = (*httpProvider)(nil)

// ServeHTTP implements the [http.Handler] interface for *httpProvider.
func (p *httpProvider) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := strings.TrimPrefix(r.URL.Path, http01.PathPrefix)

	p.mu.Lock()
	keyAuth, ok := p.tokens[token]
	p.mu.Unlock()

	if !ok {
		http.NotFound(w, r)

		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte(keyAuth))
}

// parseECPrivateKey decodes a PEM-encoded ECDSA private key.
func parseECPrivateKey(pemStr string) (key *ecdsa.PrivateKey, err error) {
	block, _ := pem.Decode([]byte(pemStr))
	if block == nil {
		return nil, errors.Error("acme: invalid account key pem")
	}

	return x509.ParseECPrivateKey(block.Bytes)
}

// marshalECPrivateKey encodes an ECDSA private key as PEM.
func marshalECPrivateKey(key crypto.PrivateKey) (pemStr string, err error) {
	ecKey, ok := key.(*ecdsa.PrivateKey)
	if !ok {
		return "", errors.Error("acme: account key is not ecdsa")
	}

	der, err := x509.MarshalECPrivateKey(ecKey)
	if err != nil {
		return "", err
	}

	block := &pem.Block{Type: "EC PRIVATE KEY", Bytes: der}

	return string(pem.EncodeToMemory(block)), nil
}

// certNotAfter parses the leaf certificate's expiration time out of a
// PEM-encoded certificate chain.
func certNotAfter(certPEM []byte) (notAfter time.Time, err error) {
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return time.Time{}, errors.Error("acme: invalid certificate pem")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return time.Time{}, err
	}

	return cert.NotAfter, nil
}
