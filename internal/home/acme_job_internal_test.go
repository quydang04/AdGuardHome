package home

import (
	"bufio"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/acme"
	"github.com/AdguardTeam/AdGuardHome/internal/agh"
	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghtls"
	"github.com/AdguardTeam/AdGuardHome/internal/client"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/golibs/httphdr"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/netutil/httputil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/AdguardTeam/golibs/timeutil"
	"github.com/go-acme/lego/v4/challenge/http01"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAcmeJob_Replay(t *testing.T) {
	j := newAcmeJob()

	j.log("info", "line one")
	j.log("info", "line two")

	lines, _, done, result := j.snapshot()
	require.Len(t, lines, 2)
	assert.Equal(t, "line one", lines[0].Message)
	assert.Equal(t, "line two", lines[1].Message)
	assert.False(t, done)
	assert.Nil(t, result)
}

func TestAcmeJob_Finish(t *testing.T) {
	j := newAcmeJob()

	j.log("info", "starting")
	j.finish(&acmeJobResult{Success: true})

	lines, _, done, result := j.snapshot()
	require.Len(t, lines, 1)
	assert.True(t, done)
	require.NotNil(t, result)
	assert.True(t, result.Success)

	// Further log/finish calls after done must be no-ops.
	j.log("info", "should be ignored")
	j.finish(&acmeJobResult{Success: false})

	lines, _, done, result = j.snapshot()
	assert.Len(t, lines, 1)
	assert.True(t, done)
	assert.True(t, result.Success)
	assert.True(t, j.isDone())
}

func TestAcmeJob_NotifyWakesSubscriber(t *testing.T) {
	j := newAcmeJob()

	_, notify, _, _ := j.snapshot()

	var wg sync.WaitGroup
	wg.Add(1)

	woke := false
	go func() {
		defer wg.Done()

		select {
		case <-notify:
			woke = true
		case <-time.After(2 * time.Second):
		}
	}()

	j.log("info", "wake up")
	wg.Wait()

	assert.True(t, woke, "subscriber should have woken up on new log line")
}

func TestAcmeJob_NotifyWakesOnFinish(t *testing.T) {
	j := newAcmeJob()

	_, notify, _, _ := j.snapshot()

	var wg sync.WaitGroup
	wg.Add(1)

	woke := false
	go func() {
		defer wg.Done()

		select {
		case <-notify:
			woke = true
		case <-time.After(2 * time.Second):
		}
	}()

	j.finish(&acmeJobResult{Success: true})
	wg.Wait()

	assert.True(t, woke, "subscriber should have woken up on finish")
}

func TestAcmeJob_ConcurrentAccess(t *testing.T) {
	j := newAcmeJob()

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()

			j.log("info", "concurrent line")
		}()
	}

	wg.Wait()
	j.finish(&acmeJobResult{Success: true})

	lines, _, done, result := j.snapshot()
	assert.Len(t, lines, 20)
	assert.True(t, done)
	assert.True(t, result.Success)
}

func TestTLSManager_ApplyIssuedCertificate_NoDeadlock(t *testing.T) {
	storeGlobals(t)

	ctx := testutil.ContextWithTimeout(t, testTimeout)

	dotLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	dotPort := uint16(testutil.RequireTypeAssert[*net.TCPAddr](t, dotLn.Addr()).Port)
	require.NoError(t, dotLn.Close())

	httpsLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	httpsPort := uint16(testutil.RequireTypeAssert[*net.TCPAddr](t, httpsLn.Addr()).Port)
	require.NoError(t, httpsLn.Close())

	config.HTTPConfig.Address = netip.AddrPortFrom(netutil.IPv4Localhost(), 3000)
	config.DNS.BindHosts = []netip.Addr{netutil.IPv4Localhost()}
	config.DNS.Port = 0

	m, err := newTLSManager(ctx, &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
		tlsSettings: tlsConfigSettings{
			Enabled:        false,
			PortHTTPS:      httpsPort,
			PortDNSOverTLS: dotPort,
		},
		servePlainDNS: true,
	})
	require.NoError(t, err)

	globalContext.dnsServer, err = dnsforward.NewServer(dnsforward.DNSCreateParams{
		Logger:            testLogger,
		TLSConfigProvider: m,
	})
	require.NoError(t, err)

	err = globalContext.dnsServer.Prepare(
		ctx,
		&dnsforward.ServerConfig{
			TLSConf: &dnsforward.TLSConfig{},
			Config: dnsforward.Config{
				UpstreamMode:     dnsforward.UpstreamModeLoadBalance,
				EDNSClientSubnet: &dnsforward.EDNSClientSubnet{Enabled: false},
				ClientsContainer: dnsforward.EmptyClientsContainer{},
			},
			ServePlainDNS: true,
		},
	)
	require.NoError(t, err)

	globalContext.clients.storage, err = client.NewStorage(ctx, &client.StorageConfig{
		BaseLogger: testLogger,
		Logger:     testLogger,
		Clock:      timeutil.SystemClock{},
	})
	require.NoError(t, err)

	web := newTestWeb(t, &webConfig{tlsManager: m})
	m.setWebAPI(web)

	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	notAfter := time.Now().Add(90 * 24 * time.Hour).UTC().Truncate(time.Second)
	certTmpl := &x509.Certificate{
		SerialNumber: big.NewInt(42),
		DNSNames:     []string{"*.example.org", "adguard.example.org"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     notAfter,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, certTmpl, certTmpl, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	doneCh := make(chan struct{})
	var status *tlsConfigStatus
	var applyErr error
	go func() {
		defer close(doneCh)
		status, _, applyErr = m.applyIssuedCertificate(ctx, &acme.Result{
			CertificatePEM: certPEM,
			PrivateKeyPEM:  keyPEM,
			NotAfter:       notAfter,
		})
	}()

	select {
	case <-doneCh:
	case <-time.After(5 * time.Second):
		t.Fatal("applyIssuedCertificate deadlocked")
	}

	testutil.CleanupAndRequireSuccess(t, func() (stopErr error) {
		return globalContext.dnsServer.Stop(testutil.ContextWithTimeout(t, testTimeout))
	})

	require.NoError(t, applyErr)
	require.NotNil(t, status)
	assert.True(t, status.ValidCert)
	assert.True(t, status.ValidKey)
	assert.True(t, status.ValidPair)

	extConf := m.extendedTLSConfig()
	assert.True(t, extConf.Enabled)
	assert.Equal(t, "adguard.example.org", extConf.ServerName)
	require.NotNil(t, m.TLSConfig())
}

func TestTLSManager_HandleACMEIssueStream_RealTimeFlushWithGzip(t *testing.T) {
	storeGlobals(t)

	ctx := testutil.ContextWithTimeout(t, testTimeout)

	mux := http.NewServeMux()
	mw := &webMw{}
	httpReg := aghhttp.NewDefaultRegistrar(mux, mw.wrap)

	m, err := newTLSManager(ctx, &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
		httpReg:      httpReg,
	})
	require.NoError(t, err)

	web := newTestWeb(t, &webConfig{tlsManager: m})
	m.setWebAPI(web)
	mw.set(web)
	m.registerACMEWebHandlers()

	job := newAcmeJob()
	job.logKey("info", "acme_progress_starting", map[string]string{"domains": "example.org"})
	m.acmeJob = job

	logMw := httputil.NewLogMiddleware(testLogger, 0)
	srv := httptest.NewServer(logMw.Wrap(mux))
	t.Cleanup(srv.Close)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/control/tls/acme/issue/stream", nil)
	require.NoError(t, err)
	req.Header.Set(httphdr.Accept, "text/event-stream")
	req.Header.Set(httphdr.AcceptEncoding, "gzip")

	respCh := make(chan *http.Response, 1)
	errCh := make(chan error, 1)
	go func() {
		resp, reqErr := srv.Client().Do(req)
		if reqErr != nil {
			errCh <- reqErr

			return
		}
		respCh <- resp
	}()

	var resp *http.Response
	select {
	case reqErr := <-errCh:
		require.NoError(t, reqErr)
	case resp = <-respCh:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for SSE response headers (blocked by gzip buffering)")
	}
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Empty(t, resp.Header.Get(httphdr.ContentEncoding), "SSE stream must not be gzip-encoded")

	scanner := bufio.NewScanner(resp.Body)
	lineCh := make(chan string, 4)
	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "event: ") || strings.HasPrefix(line, "data: ") {
				lineCh <- line
			}
		}
	}()

	// Verify the initial "line" event arrives BEFORE job.finish is called!
	select {
	case ev := <-lineCh:
		assert.Equal(t, "event: line", ev)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for real-time SSE line event before job finished")
	}

	job.finish(&acmeJobResult{Success: true})

	foundDone := false
	deadline := time.After(2 * time.Second)
	for !foundDone {
		select {
		case ev := <-lineCh:
			if ev == "event: done" {
				foundDone = true
			}
		case <-deadline:
			t.Fatal("timed out waiting for SSE done event")
		}
	}
}

func TestIsPublicResource_ACMEChallenge(t *testing.T) {
	assert.True(t, isPublicResource(http01.PathPrefix+"sample-token-123"))
	assert.False(t, isPublicResource("/control/tls/acme/status"))
}

func TestTLSManager_HandleACMEIssue_FullFlow(t *testing.T) {
	storeGlobals(t)

	ctx := testutil.ContextWithTimeout(t, testTimeout)

	dotLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	dotPort := uint16(testutil.RequireTypeAssert[*net.TCPAddr](t, dotLn.Addr()).Port)
	require.NoError(t, dotLn.Close())

	httpsLn, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	httpsPort := uint16(testutil.RequireTypeAssert[*net.TCPAddr](t, httpsLn.Addr()).Port)
	require.NoError(t, httpsLn.Close())

	config.HTTPConfig.Address = netip.AddrPortFrom(netutil.IPv4Localhost(), 3000)
	config.DNS.BindHosts = []netip.Addr{netutil.IPv4Localhost()}
	config.DNS.Port = 0
	config.ACME = &acmeConfig{
		Enabled:         false,
		Email:           "admin@example.org",
		Domains:         []string{"adguard.example.org"},
		Challenge:       string(acme.ChallengeHTTP01),
		AutoRenew:       true,
		RenewBeforeDays: 14,
	}

	mux := http.NewServeMux()
	mw := &webMw{}
	httpReg := aghhttp.NewDefaultRegistrar(mux, mw.wrap)

	m, err := newTLSManager(ctx, &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
		httpReg:      httpReg,
		tlsSettings: tlsConfigSettings{
			Enabled:        false,
			PortHTTPS:      httpsPort,
			PortDNSOverTLS: dotPort,
		},
		servePlainDNS: true,
	})
	require.NoError(t, err)

	m.issueFunc = func(_ context.Context, req *acme.Request) (*acme.Result, error) {
		if req.Progress != nil {
			req.Progress(acme.ProgressPreparing, map[string]string{"domains": strings.Join(req.Domains, ", ")})
		}

		key, genErr := rsa.GenerateKey(rand.Reader, 2048)
		if genErr != nil {
			return nil, genErr
		}

		notAfter := time.Now().Add(90 * 24 * time.Hour).UTC().Truncate(time.Second)
		certTmpl := &x509.Certificate{
			SerialNumber: big.NewInt(99),
			DNSNames:     req.Domains,
			NotBefore:    time.Now().Add(-time.Hour),
			NotAfter:     notAfter,
		}
		certDER, genErr := x509.CreateCertificate(rand.Reader, certTmpl, certTmpl, &key.PublicKey, key)
		if genErr != nil {
			return nil, genErr
		}

		return &acme.Result{
			CertificatePEM: pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER}),
			PrivateKeyPEM: pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(key),
			}),
			AccountURI: "https://acme.example/acct/1",
			NotAfter:   notAfter,
		}, nil
	}

	globalContext.dnsServer, err = dnsforward.NewServer(dnsforward.DNSCreateParams{
		Logger:            testLogger,
		TLSConfigProvider: m,
	})
	require.NoError(t, err)

	err = globalContext.dnsServer.Prepare(
		ctx,
		&dnsforward.ServerConfig{
			TLSConf: &dnsforward.TLSConfig{},
			Config: dnsforward.Config{
				UpstreamMode:     dnsforward.UpstreamModeLoadBalance,
				EDNSClientSubnet: &dnsforward.EDNSClientSubnet{Enabled: false},
				ClientsContainer: dnsforward.EmptyClientsContainer{},
			},
			ServePlainDNS: true,
		},
	)
	require.NoError(t, err)
	testutil.CleanupAndRequireSuccess(t, func() (stopErr error) {
		return globalContext.dnsServer.Stop(testutil.ContextWithTimeout(t, testTimeout))
	})

	globalContext.clients.storage, err = client.NewStorage(ctx, &client.StorageConfig{
		BaseLogger: testLogger,
		Logger:     testLogger,
		Clock:      timeutil.SystemClock{},
	})
	require.NoError(t, err)

	web := newTestWeb(t, &webConfig{tlsManager: m})
	m.setWebAPI(web)
	mw.set(web)
	m.registerWebHandlers()
	m.registerACMEWebHandlers()

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	issueReq, err := http.NewRequestWithContext(ctx, http.MethodPost, srv.URL+"/control/tls/acme/issue", nil)
	require.NoError(t, err)

	issueResp, err := srv.Client().Do(issueReq)
	require.NoError(t, err)
	require.NoError(t, issueResp.Body.Close())
	assert.Equal(t, http.StatusAccepted, issueResp.StatusCode)

	streamReq, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL+"/control/tls/acme/issue/stream", nil)
	require.NoError(t, err)
	streamReq.Header.Set(httphdr.Accept, "text/event-stream")
	streamReq.Header.Set(httphdr.AcceptEncoding, "gzip")

	streamResp, err := srv.Client().Do(streamReq)
	require.NoError(t, err)
	defer func() { _ = streamResp.Body.Close() }()
	assert.Equal(t, http.StatusOK, streamResp.StatusCode)

	scanner := bufio.NewScanner(streamResp.Body)
	var doneLine string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") && strings.Contains(line, `"success":true`) {
			doneLine = line

			break
		}
	}
	assert.NotEmpty(t, doneLine)

	extConf := m.extendedTLSConfig()
	assert.True(t, extConf.Enabled)
	assert.Equal(t, "adguard.example.org", extConf.ServerName)
}

