package home

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/agh"
	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghtest"
	"github.com/AdguardTeam/AdGuardHome/internal/aghtls"
	"github.com/AdguardTeam/golibs/errors"
	"github.com/AdguardTeam/golibs/httphdr"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestWeb_HandleGetProfile(t *testing.T) {
	storeGlobals(t)

	const (
		testTTL = 60

		userName     = "name"
		userPassword = "password"

		path = "/control/profile"
	)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tempDir := t.TempDir()

	sessionsDB := filepath.Join(tempDir, "sessions.db")

	user := &webUser{
		Name:         userName,
		PasswordHash: string(passwordHash),
	}

	baseMux := http.NewServeMux()

	auth, err := newAuth(testutil.ContextWithTimeout(t, testTimeout), &authConfig{
		baseLogger:     testLogger,
		mux:            baseMux,
		rateLimiter:    emptyRateLimiter{},
		trustedProxies: testTrustedProxies,
		dbFilename:     sessionsDB,
		users:          nil,
		sessionTTL:     testTTL * time.Second,
		isGLiNet:       false,
	})
	require.NoError(t, err)

	t.Cleanup(func() { auth.close(testutil.ContextWithTimeout(t, testTimeout)) })

	tlsMgr, err := newTLSManager(testutil.ContextWithTimeout(t, testTimeout), &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
	})
	require.NoError(t, err)

	web := newTestWeb(t, &webConfig{
		tlsManager: tlsMgr,
		auth:       auth,
		mux:        baseMux,
	})
	require.NoError(t, err)

	mux := auth.middleware().Wrap(baseMux)

	require.True(t, t.Run("userless", func(t *testing.T) {
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)

		web.handleGetProfile(w, r)
		assert.Equal(t, http.StatusOK, w.Code)
	}))

	require.True(t, t.Run("add_user", func(t *testing.T) {
		ctx := testutil.ContextWithTimeout(t, testTimeout)
		err = auth.addUser(ctx, user, userPassword)
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodGet, path, nil)

		web.handleGetProfile(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, path, nil)

		loginCookie := generateAuthCookie(t, mux, userName, userPassword)
		r.AddCookie(loginCookie)

		web.handleGetProfile(w, r)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	}))
}

func TestWeb_HandlePutProfile(t *testing.T) {
	storeGlobals(t)

	mw := &webMw{}
	mux := http.NewServeMux()
	httpReg := aghhttp.NewDefaultRegistrar(mux, mw.wrap)

	isConfigChanged := false
	confModifier := &aghtest.ConfigModifier{
		OnApply: func(_ context.Context) { isConfigChanged = true },
	}

	web := newTestWeb(t, &webConfig{
		mux:            mux,
		configModifier: confModifier,
		httpReg:        httpReg,
	})

	mw.set(web)

	var (
		dataValid = errors.Must(json.Marshal(&profileJSON{
			Language: "en",
			Theme:    "auto",
		}))

		dataInvalidLang = errors.Must(json.Marshal(&profileJSON{
			Language: "invalid_lang",
			Theme:    "auto",
		}))

		dataInvalidTheme = errors.Must(json.Marshal(&profileJSON{
			Language: "en",
			Theme:    "invalid_theme",
		}))
	)

	testCases := []struct {
		req      *http.Request
		name     string
		wantBody string
		wantCode int
	}{{
		req:      newProfileUpdateRequest(http.MethodPut, dataValid, true),
		name:     "basic",
		wantBody: "OK\n",
		wantCode: http.StatusOK,
	}, {
		req:      newProfileUpdateRequest(http.MethodGet, dataValid, true),
		name:     "invalid_method",
		wantBody: "only method PUT is allowed\n",
		wantCode: http.StatusMethodNotAllowed,
	}, {
		req:      newProfileUpdateRequest(http.MethodPut, dataValid, false),
		name:     "invalid_content_type",
		wantBody: "only content-type application/json is allowed\n",
		wantCode: http.StatusUnsupportedMediaType,
	}, {
		req:      newProfileUpdateRequest(http.MethodPut, nil, false),
		name:     "empty_body",
		wantBody: "reading req: EOF\n",
		wantCode: http.StatusBadRequest,
	}, {
		req:      newProfileUpdateRequest(http.MethodPut, dataInvalidLang, true),
		name:     "invalid_language",
		wantBody: `unknown language: "invalid_lang"` + "\n",
		wantCode: http.StatusBadRequest,
	}, {
		req:  newProfileUpdateRequest(http.MethodPut, dataInvalidTheme, true),
		name: "invalid_theme",
		wantBody: `reading req: invalid theme "invalid_theme", ` +
			`supported: "auto", "dark", "light"` + "\n",
		wantCode: http.StatusBadRequest,
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			mux.ServeHTTP(w, tc.req)
			assert.Equal(t, tc.wantCode, w.Code)
			assert.Equal(t, tc.wantBody, w.Body.String())
		})
	}

	require.True(t, t.Run("single_config_update", func(t *testing.T) {
		isConfigChanged = false
		config.Language = ""
		config.Theme = ""

		w := httptest.NewRecorder()

		mux.ServeHTTP(w, newProfileUpdateRequest(http.MethodPut, dataValid, true))
		require.Equal(t, http.StatusOK, w.Code)

		assert.True(t, isConfigChanged)

		isConfigChanged = false

		mux.ServeHTTP(w, newProfileUpdateRequest(http.MethodPut, dataValid, true))
		require.Equal(t, http.StatusOK, w.Code)

		assert.False(t, isConfigChanged)
	}))
}

// newProfileUpdateRequest builds an *http.Request for the profile update
// endpoint.  If body is non-nil, it is used as the request body.  If setCT is
// true, the Content-Type header is set to application/json.
func newProfileUpdateRequest(method string, body []byte, setCT bool) (req *http.Request) {
	var r io.Reader
	if body != nil {
		r = bytes.NewReader(body)
	}

	req = httptest.NewRequest(method, "/control/profile/update", r)
	if setCT {
		req.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
	}

	return req
}

func TestWeb_HandleChangeAccountAndPort(t *testing.T) {
	storeGlobals(t)

	const (
		userName     = "admin"
		newUserName  = "new_admin"
		userPassword = "password123"
		newPassword  = "newpassword123"
	)

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(userPassword), bcrypt.MinCost)
	require.NoError(t, err)

	ctx := testutil.ContextWithTimeout(t, testTimeout)
	baseMux := http.NewServeMux()
	mw := &webMw{}
	httpReg := aghhttp.NewDefaultRegistrar(baseMux, mw.wrap)

	auth, err := newAuth(ctx, &authConfig{
		baseLogger:     testLogger,
		mux:            baseMux,
		rateLimiter:    emptyRateLimiter{},
		trustedProxies: testTrustedProxies,
		dbFilename:     filepath.Join(t.TempDir(), "sessions.db"),
		users:          []webUser{{Name: userName, PasswordHash: string(passwordHash)}},
		sessionTTL:     60 * time.Second,
	})
	require.NoError(t, err)
	t.Cleanup(func() { auth.close(ctx) })

	tlsMgr, err := newTLSManager(ctx, &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
	})
	require.NoError(t, err)

	config.HTTPConfig.Address = netip.MustParseAddrPort("127.0.0.1:3000")

	web := newTestWeb(t, &webConfig{
		tlsManager:     tlsMgr,
		auth:           auth,
		mux:            baseMux,
		httpReg:        httpReg,
		configModifier: agh.EmptyConfigModifier{},
	})
	web.conf.BindAddr = config.HTTPConfig.Address
	mw.set(web)

	mux := auth.middleware().Wrap(baseMux)

	loginCookie := generateAuthCookie(t, mux, userName, userPassword)

	t.Run("change_username_keeps_session_and_updates_login", func(t *testing.T) {
		body, err := json.Marshal(changeUsernameJSON{NewUsername: "  " + newUserName + "  "})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/control/profile/username", bytes.NewReader(body))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)

		mux.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)

		// Existing session cookie must stay valid and reflect the new username.
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodGet, "/control/profile", nil)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)

		var prof profileJSON
		require.NoError(t, json.NewDecoder(w.Body).Decode(&prof))
		assert.Equal(t, newUserName, prof.Name)

		// Logging in with the new username must succeed.
		newCookie := generateAuthCookie(t, mux, newUserName, userPassword)
		require.NotNil(t, newCookie)
	})

	t.Run("change_password_validation_and_update", func(t *testing.T) {
		// Wrong current password returns 400 Bad Request so the SPA does not treat
		// it as an expired session (403).
		wrongBody, err := json.Marshal(changePasswordJSON{
			CurrentPassword: "wrongpassword",
			NewPassword:     newPassword,
		})
		require.NoError(t, err)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/control/profile/password", bytes.NewReader(wrongBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Too long (> 72 bytes) returns 400 Bad Request, not 500.
		longBody, err := json.Marshal(changePasswordJSON{
			CurrentPassword: userPassword,
			NewPassword:     strings.Repeat("a", 73),
		})
		require.NoError(t, err)

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/control/profile/password", bytes.NewReader(longBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Valid password change updates password for future logins.
		validBody, err := json.Marshal(changePasswordJSON{
			CurrentPassword: userPassword,
			NewPassword:     newPassword,
		})
		require.NoError(t, err)

		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/control/profile/password", bytes.NewReader(validBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)

		newCookie := generateAuthCookie(t, mux, newUserName, newPassword)
		require.NotNil(t, newCookie)
	})

	t.Run("change_port", func(t *testing.T) {
		// Zero port is rejected.
		zeroBody, err := json.Marshal(changePortJSON{Port: 0})
		require.NoError(t, err)
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/control/change_port", bytes.NewReader(zeroBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Port already occupied by another listener is rejected.
		occupiedLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		t.Cleanup(func() { _ = occupiedLn.Close() })
		occupiedPort := uint16(occupiedLn.Addr().(*net.TCPAddr).Port)

		occBody, err := json.Marshal(changePortJSON{Port: occupiedPort})
		require.NoError(t, err)
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/control/change_port", bytes.NewReader(occBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		assert.Equal(t, http.StatusBadRequest, w.Code)

		// Find an ephemeral free port, release it, then switch to it.
		freeLn, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		freePort := uint16(freeLn.Addr().(*net.TCPAddr).Port)
		require.NoError(t, freeLn.Close())

		okBody, err := json.Marshal(changePortJSON{Port: freePort})
		require.NoError(t, err)
		w = httptest.NewRecorder()
		r = httptest.NewRequest(http.MethodPost, "/control/change_port", bytes.NewReader(okBody))
		r.Header.Set(httphdr.ContentType, aghhttp.HdrValApplicationJSON)
		r.AddCookie(loginCookie)
		mux.ServeHTTP(w, r)
		require.Equal(t, http.StatusOK, w.Code)

		assert.Equal(t, freePort, config.HTTPConfig.Address.Port())
		assert.Equal(t, freePort, web.conf.BindAddr.Port())
	})
}
