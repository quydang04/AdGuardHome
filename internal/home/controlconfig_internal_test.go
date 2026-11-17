package home

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/agh"
	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/aghtls"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/golibs/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDecodeImport_Compatibility(t *testing.T) {
	storeGlobals(t)

	t.Run("bom_only_empty_rejected", func(t *testing.T) {
		bomOnly := []byte{0xEF, 0xBB, 0xBF}
		_, _, err := decodeImport(bomOnly)
		require.Error(t, err)
	})

	t.Run("bom_prefixed_legacy_v0", func(t *testing.T) {
		legacyJSON := append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{
			"dns": {
				"upstream_dns": ["https://dns.nextdns.io/54ffe5", "[/*.vn/] https://dns.nextdns.io/54ffe5"],
				"bootstrap_dns": ["1.0.0.1", "8.8.4.4"],
				"fallback_dns": ["8.8.8.8"],
				"bind_hosts": ["0.0.0.0"],
				"port": "53",
				"upstream_timeout": 10,
				"fastest_timeout": 1000,
				"edns_client_subnet": {
					"custom_ip": "",
					"enabled": true,
					"use_custom": false
				}
			},
			"query_log": {
				"enabled": true,
				"interval": "30d"
			},
			"statistics": {
				"enabled": true,
				"interval": "720h"
			},
			"safebrowsing": true,
			"parental": false
		}`)...)

		imp, warnings, err := decodeImport(legacyJSON)
		require.NoError(t, err)
		assert.Empty(t, warnings)
		require.NotNil(t, imp.DNS)
		assert.Equal(t, uint16(53), imp.DNS.Port)
		require.NotNil(t, imp.DNS.EDNSClientSubnet)
		assert.True(t, imp.DNS.EDNSClientSubnet.Enabled)
		require.NotNil(t, imp.QueryLog)
		assert.Equal(t, uint32(720), imp.QueryLog.Interval)
		require.NotNil(t, imp.Statistics)
		assert.Equal(t, uint32(720), imp.Statistics.Interval)
		require.NotNil(t, imp.SafeBrowsing)
		assert.True(t, *imp.SafeBrowsing)
		require.NotNil(t, imp.Parental)
		assert.False(t, *imp.Parental)
	})

	t.Run("v2_roundtrip_compat", func(t *testing.T) {
		config.Language = "vi"
		config.Theme = ThemeDark
		config.Filtering.SafeBrowsingEnabled = false
		config.Filtering.ParentalEnabled = false

		exported := buildExportConfig()
		assert.Equal(t, exportFormatVersion, exported.Version)
		raw, err := json.Marshal(exported)
		require.NoError(t, err)

		imp, warnings, err := decodeImport(raw)
		require.NoError(t, err)
		assert.Empty(t, warnings)
		require.NotNil(t, imp.General)
		assert.Equal(t, "vi", imp.General.Language)
		require.NotNil(t, imp.SafeBrowsing)
		assert.False(t, *imp.SafeBrowsing)
		require.NotNil(t, imp.Parental)
		assert.False(t, *imp.Parental)

		// Also verify an older v0/v1-style unmarshal directly into exportConfig
		// succeeds without error.
		var oldDecoderTarget exportConfig
		require.NoError(t, json.Unmarshal(raw, &oldDecoderTarget))
		assert.NotNil(t, oldDecoderTarget.DNS)
		assert.NotNil(t, oldDecoderTarget.SafeBrowsing)
		assert.False(t, *oldDecoderTarget.SafeBrowsing)
	})
}

func TestHandleImportSettings_KeepsDNSRunningOnBindFailure(t *testing.T) {
	storeGlobals(t)

	ctx := testutil.ContextWithTimeout(t, testTimeout)

	// Occupy one port so the imported config cannot bind to it.
	occupiedPC, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = occupiedPC.Close() })
	occupiedTCP, err := net.Listen("tcp", occupiedPC.LocalAddr().String())
	require.NoError(t, err)
	t.Cleanup(func() { _ = occupiedTCP.Close() })
	occupiedPort := uint16(occupiedPC.LocalAddr().(*net.UDPAddr).Port)

	// Find a free port for the running DNS server.
	freePC, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	freePort := uint16(freePC.LocalAddr().(*net.UDPAddr).Port)
	require.NoError(t, freePC.Close())

	config.DNS.BindHosts = []netip.Addr{netip.MustParseAddr("127.0.0.1")}
	config.DNS.Port = freePort
	config.DNS.ServePlainDNS = true
	config.DNS.PendingRequests = &pendingRequests{Enabled: true}
	config.DNS.UpstreamDNS = []string{"1.1.1.1"}
	config.DNS.BootstrapDNS = []string{"1.1.1.1"}

	config.Filtering.Logger = testLogger
	config.Filtering.DataDir = t.TempDir()
	globalContext.filters, err = filtering.New(config.Filtering, nil)
	require.NoError(t, err)

	globalContext.clients = *newClientsContainer(t)

	tlsMgr, err := newTLSManager(ctx, &tlsManagerConfig{
		logger:       testLogger,
		confModifier: agh.EmptyConfigModifier{},
		manager:      aghtls.EmptyManager{},
	})
	require.NoError(t, err)

	globalContext.dnsServer, err = dnsforward.NewServer(dnsforward.DNSCreateParams{
		Logger:            testLogger,
		DNSFilter:         globalContext.filters,
		TLSConfigProvider: tlsMgr,
	})
	require.NoError(t, err)
	t.Cleanup(func() {
		_ = globalContext.dnsServer.Stop(testutil.ContextWithTimeout(t, testTimeout))
	})

	dnsConf, err := newServerConfig(
		&config.DNS,
		config.Clients.Sources,
		tlsMgr.extendedTLSConfig(),
		config.HTTPConfig.DoH,
		tlsMgr,
		aghhttp.EmptyRegistrar{},
		globalContext.clients.storage,
		agh.EmptyConfigModifier{},
	)
	require.NoError(t, err)

	require.NoError(t, globalContext.dnsServer.Prepare(ctx, dnsConf))
	require.NoError(t, globalContext.dnsServer.Start(ctx))
	require.True(t, globalContext.dnsServer.IsRunning())

	web := newTestWeb(t, &webConfig{
		tlsManager:     tlsMgr,
		configModifier: agh.EmptyConfigModifier{},
		httpReg:        aghhttp.EmptyRegistrar{},
	})

	importBody, err := json.Marshal(map[string]any{
		"export_version": 2,
		"dns": map[string]any{
			"bind_hosts":      []string{"127.0.0.1"},
			"port":            occupiedPort,
			"upstream_dns":    []string{"https://dns.nextdns.io/54ffe5"},
			"bootstrap_dns":   []string{"1.0.0.1"},
			"serve_plain_dns": true,
			"cache_enabled":   true,
			"cache_size":      4194304,
		},
		"filtering": map[string]any{
			"enabled":           true,
			"filtering_enabled": true,
			"blocking_mode":     "null_ip",
			"blocked_services": map[string]any{
				"schedule": map[string]any{"time_zone": "UTC"},
				"ids":      []string{},
			},
		},
		"tls": map[string]any{
			"enabled":     true,
			"private_key": "",
		},
	})
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodPost, "/control/settings/import", bytes.NewReader(importBody))
	web.handleImportSettings(w, r)

	require.Equal(t, http.StatusOK, w.Code)

	// DNS server must remain running on the local working port instead of getting
	// stuck stopped ("DNS server is starting up").
	assert.True(t, globalContext.dnsServer.IsRunning())
	assert.Equal(t, freePort, config.DNS.Port)
	assert.Equal(t, []string{"https://dns.nextdns.io/54ffe5"}, config.DNS.UpstreamDNS)

	var resp struct {
		Status   string   `json:"status"`
		Warnings []string `json:"warnings"`
	}
	require.NoError(t, json.NewDecoder(w.Body).Decode(&resp))
	assert.Equal(t, "ok", resp.Status)
	assert.NotEmpty(t, resp.Warnings)
}
