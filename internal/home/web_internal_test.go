package home

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"
	"testing"
	"testing/fstest"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghos"
	"github.com/AdguardTeam/AdGuardHome/internal/aghuser"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/netutil/urlutil"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// h2c upgrade headers.
//
// TODO(a.garipov): Add to httphdr.
const (
	headerConnection    = "Connection"
	headerUpgrade       = "Upgrade"
	headerHTTP2Settings = "HTTP2-Settings"
)

// h2c upgrade header values for tests.
const (
	testHeaderValueConnection = "Upgrade, HTTP2-Settings"
	testHeaderValueUpgrade    = "h2c"
	testSettings              = "AAEAABAAAAIAAAABAAQAAP__AAUAAEAAAAgAAAAAAAMAAABkAAYAAQAA"
)

// TestWebAPI_h2cVulnerability makes sure that AdGuard Home no longer
// establishes unencrypted HTTP/2 connections via the HTTP/1.1 upgrade
// mechanism, which was discontinued per RFC 9113.  See GHSA-p5f5-3p5g-rfjw.
func TestWebAPI_h2cVulnerability(t *testing.T) {
	storeGlobals(t)

	stop := make(chan struct{})
	t.Cleanup(func() {
		testutil.RequireReceive(t, stop, testTimeout)
	})

	password := "password"
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost)
	require.NoError(t, err)

	fs := fstest.MapFS{
		"build/static/login.html": &fstest.MapFile{
			Data: []byte("foo"),
			Mode: aghos.DefaultPermFile,
		},
	}

	user := webUser{
		Name:         "foo",
		PasswordHash: string(passwordHash),
		UserID:       aghuser.MustNewUserID(),
	}

	mux := http.NewServeMux()
	auth, err := newAuth(testutil.ContextWithTimeout(t, testTimeout), &authConfig{
		baseLogger:     testLogger,
		rateLimiter:    emptyRateLimiter{},
		trustedProxies: testTrustedProxies,
		dbFilename:     path.Join(t.TempDir(), "sessions.db"),
		users:          []webUser{user},
		sessionTTL:     testTimeout,
		isGLiNet:       false,
		mux:            mux,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		ctx := testutil.ContextWithTimeout(t, testTimeout)
		auth.close(ctx)
	})

	mw := &webMw{}
	registrar := aghhttp.NewDefaultRegistrar(mux, mw.wrap)
	web := newTestWeb(t, &webConfig{
		baseLogger:    testLogger,
		auth:          auth,
		mux:           mux,
		httpReg:       registrar,
		clientBuildFS: fs,
	})

	mw.set(web)
	globalContext.web = web

	port := config.HTTPConfig.Address.Port()
	host := fmt.Sprintf("%s:%d", netutil.IPv4Localhost(), port)

	go func() {
		ctx := testutil.ContextWithTimeout(t, testTimeout)
		web.start(ctx)
		close(stop)
	}()

	t.Cleanup(func() {
		ctx := testutil.ContextWithTimeout(t, testTimeout)
		web.close(ctx)
	})

	waitForWebAPIReady(t, host)
	requireNoH2CUpgrade(t, host)
}

// waitForWebAPIReady waits until the [webAPI] server has started and is ready
// to accept connections.
func waitForWebAPIReady(tb testing.TB, host string) {
	tb.Helper()

	u := (&url.URL{
		Scheme: urlutil.SchemeHTTP,
		Host:   host,
		Path:   "/login.html",
	}).String()

	require.EventuallyWithT(tb, func(c *assert.CollectT) {
		ctx := testutil.ContextWithTimeout(tb, testTimeout)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
		require.NoError(c, err)

		resp, err := http.DefaultClient.Do(req)
		require.NoError(c, err)
		assert.Equal(c, http.StatusOK, resp.StatusCode)
	}, testTimeout, testTimeout/10)
}

// requireNoH2CUpgrade establishes a TCP connection to the specified host and
// attempts to perform an HTTP/1.1-to-h2c protocol upgrade, verifying that the
// server does not honor it, i.e. does not respond with
// [http.StatusSwitchingProtocols].  host must not be empty.
func requireNoH2CUpgrade(tb testing.TB, host string) {
	tb.Helper()

	dialer := &net.Dialer{}
	ctx := testutil.ContextWithTimeout(tb, testTimeout)

	conn, err := dialer.DialContext(ctx, "tcp", host)
	require.NoError(tb, err)
	testutil.CleanupAndRequireSuccess(tb, conn.Close)

	writer := bufio.NewWriter(conn)
	reader := bufio.NewReader(conn)

	u := &url.URL{
		Scheme: urlutil.SchemeHTTP,
		Host:   host,
		Path:   "/control/login",
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	require.NoError(tb, err)

	req.Header.Set(headerConnection, testHeaderValueConnection)
	req.Header.Set(headerUpgrade, testHeaderValueUpgrade)
	req.Header.Set(headerHTTP2Settings, testSettings)

	err = req.Write(writer)
	require.NoError(tb, err)
	require.NoError(tb, writer.Flush())

	resp, err := http.ReadResponse(reader, req)
	require.NoError(tb, err)
	testutil.CleanupAndRequireSuccess(tb, resp.Body.Close)

	assert.NotEqual(tb, http.StatusSwitchingProtocols, resp.StatusCode)
}
