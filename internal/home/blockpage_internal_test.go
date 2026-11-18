package home

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockPageMiddleware(t *testing.T) {
	flt, err := filtering.New(&filtering.Config{
		Logger:            slogutil.NewDiscardLogger(),
		DataDir:           t.TempDir(),
		ProtectionEnabled: true,
		FilteringEnabled:  true,
		BlockingMode:      filtering.BlockingModeBlockPage,
		UserRules:         []string{"||hostdemo.site^"},
	}, nil)
	require.NoError(t, err)
	flt.EnableFilters(false)
	t.Cleanup(flt.Close)

	dnsSrv, err := dnsforward.NewServer(dnsforward.DNSCreateParams{
		DNSFilter:   flt,
		PrivateNets: netutil.SubnetSetFunc(netutil.IsLocallyServed),
		Logger:      slogutil.NewDiscardLogger(),
	})
	require.NoError(t, err)

	prevFlt := globalContext.filters
	prevDNS := globalContext.dnsServer
	prevLang := config.Language
	prevACME := config.ACME
	t.Cleanup(func() {
		globalContext.filters = prevFlt
		globalContext.dnsServer = prevDNS
		config.Language = prevLang
		config.ACME = prevACME
	})

	globalContext.filters = flt
	globalContext.dnsServer = dnsSrv
	config.Language = "vi"
	config.ACME = &acmeConfig{Domains: []string{"dns.quydang.name.vn"}}

	adminHit := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		adminHit = true
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ADMIN_UI"))
	})

	web := &webAPI{logger: slogutil.NewDiscardLogger()}
	mw := web.blockPageMiddleware(next)

	t.Run("blocked_domain_renders_vietnamese_block_page", func(t *testing.T) {
		adminHit = false
		req := httptest.NewRequest(http.MethodGet, "http://hostdemo.site/", nil)
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.False(t, adminHit)
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "hostdemo.site")
		assert.Contains(t, body, "Trang này đã bị chặn")
		assert.Contains(t, body, "Quy tắc lọc tùy chỉnh")
		assert.Contains(t, body, "Xem chi tiết quy tắc")
	})

	t.Run("blocked_domain_english_override", func(t *testing.T) {
		adminHit = false
		req := httptest.NewRequest(http.MethodGet, "http://hostdemo.site/?lang=en", nil)
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.False(t, adminHit)
		assert.Equal(t, http.StatusOK, rec.Code)
		body := rec.Body.String()
		assert.Contains(t, body, "This site has been blocked")
		assert.Contains(t, body, "Custom filtering rules")
		assert.Contains(t, body, "Rule details")
	})

	t.Run("blocked_iframe_returns_204", func(t *testing.T) {
		adminHit = false
		req := httptest.NewRequest(http.MethodGet, "http://hostdemo.site/banner.html", nil)
		req.Header.Set("Sec-Fetch-Dest", "iframe")
		req.Header.Set("Accept", "text/html")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.False(t, adminHit)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("blocked_subresource_script_returns_204", func(t *testing.T) {
		adminHit = false
		req := httptest.NewRequest(http.MethodGet, "http://hostdemo.site/ad.js", nil)
		req.Header.Set("Sec-Fetch-Dest", "script")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.False(t, adminHit)
		assert.Equal(t, http.StatusNoContent, rec.Code)
	})

	t.Run("blocked_subresource_image_returns_1x1_gif", func(t *testing.T) {
		adminHit = false
		req := httptest.NewRequest(http.MethodGet, "http://hostdemo.site/pixel.gif", nil)
		req.Header.Set("Sec-Fetch-Dest", "image")
		rec := httptest.NewRecorder()

		mw.ServeHTTP(rec, req)

		assert.False(t, adminHit)
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "image/gif", rec.Header().Get("Content-Type"))
	})

	t.Run("preview_scenarios_and_allowed_check", func(t *testing.T) {
		reqMalware := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/block-page-preview?scenario=malware", nil)
		recMalware := httptest.NewRecorder()
		web.serveBlockPagePreview(recMalware, reqMalware)

		assert.Equal(t, http.StatusOK, recMalware.Code)
		assert.Contains(t, recMalware.Body.String(), "Trang web nguy hiểm đã bị chặn")

		reqAllowed := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/block-page-preview?domain=unblocked.example.com", nil)
		recAllowed := httptest.NewRecorder()
		web.serveBlockPagePreview(recAllowed, reqAllowed)

		assert.Equal(t, http.StatusOK, recAllowed.Code)
		assert.Contains(t, recAllowed.Body.String(), "Tên miền này không bị chặn")
	})

	t.Run("admin_ip_and_acme_domain_pass_through", func(t *testing.T) {
		for _, h := range []string{"127.0.0.1", "localhost", "192.168.1.33:80", "dns.quydang.name.vn", "unblocked.example.com"} {
			adminHit = false
			req := httptest.NewRequest(http.MethodGet, "http://"+h+"/", nil)
			rec := httptest.NewRecorder()

			mw.ServeHTTP(rec, req)

			assert.True(t, adminHit, "host %s should pass through to admin UI", h)
		}
	})
}
