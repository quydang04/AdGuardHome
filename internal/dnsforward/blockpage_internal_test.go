package dnsforward

import (
	"context"
	"fmt"
	"net/netip"
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
	"github.com/miekg/dns"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBlockPageDNSResponses(t *testing.T) {
	flt, err := filtering.New(&filtering.Config{
		Logger:            slogutil.NewDiscardLogger(),
		DataDir:           t.TempDir(),
		ProtectionEnabled: true,
		FilteringEnabled:  true,
		BlockingMode:      filtering.BlockingModeBlockPage,
		UserRules: []string{
			"192.168.1.50 nas.lan",
			"0.0.0.0 clix.vn",
			"@@||cached-allow.example^",
		},
	}, nil)
	require.NoError(t, err)
	flt.EnableFilters(false)
	t.Cleanup(flt.Close)

	require.NoError(t, validateBlockingMode(filtering.BlockingModeBlockPage, netip.Addr{}, netip.Addr{}))

	srv := &Server{
		logger:    slogutil.NewDiscardLogger(),
		dnsFilter: flt,
		webBindAddr: func() netip.AddrPort {
			return netip.MustParseAddrPort("0.0.0.0:80")
		},
	}

	ctx := context.Background()

	t.Run("A_loopback_returns_127_0_0_1", func(t *testing.T) {
		req := (&dns.Msg{}).SetQuestion("hostdemo.site.", dns.TypeA)
		dctx := &proxy.DNSContext{
			Req:  req,
			Addr: netip.MustParseAddrPort("127.0.0.1:54321"),
		}

		resp := srv.genDNSFilterMessage(ctx, srv.logger, dctx, &filtering.Result{
			Reason:     filtering.FilteredBlockList,
			IsFiltered: true,
		})
		require.NotNil(t, resp)
		require.Len(t, resp.Answer, 1)

		aRec, ok := resp.Answer[0].(*dns.A)
		require.True(t, ok)
		assert.Equal(t, "127.0.0.1", aRec.A.String())
	})

	t.Run("AAAA_ipv4_web_bind_returns_nodata", func(t *testing.T) {
		req := (&dns.Msg{}).SetQuestion("hostdemo.site.", dns.TypeAAAA)
		dctx := &proxy.DNSContext{
			Req:  req,
			Addr: netip.MustParseAddrPort("127.0.0.1:54321"),
		}

		resp := srv.makeResponseBlockPage(ctx, req, dctx)
		require.NotNil(t, resp)
		assert.Equal(t, dns.RcodeSuccess, resp.Rcode)
		assert.Empty(t, resp.Answer)
	})

	t.Run("HTTPS_returns_nodata", func(t *testing.T) {
		req := (&dns.Msg{}).SetQuestion("hostdemo.site.", dns.TypeHTTPS)
		dctx := &proxy.DNSContext{
			Req:  req,
			Addr: netip.MustParseAddrPort("127.0.0.1:54321"),
		}

		resp := srv.makeResponseBlockPage(ctx, req, dctx)
		require.NotNil(t, resp)
		assert.Equal(t, dns.RcodeSuccess, resp.Rcode)
		assert.Empty(t, resp.Answer)
	})

	t.Run("bounded_ring_cache_evicts_oldest", func(t *testing.T) {
		for i := range maxBlockPageEntries + 20 {
			srv.recordBlockPageHit(fmt.Sprintf("d%d.example.test.", i), &filtering.Result{
				Reason:     filtering.FilteredBlockList,
				IsFiltered: true,
			})
		}

		srv.blockPage.mu.RLock()
		count := len(srv.blockPage.entries)
		srv.blockPage.mu.RUnlock()

		assert.LessOrEqual(t, count, maxBlockPageEntries)

		res, ok := srv.LookupBlockedHost(fmt.Sprintf("d%d.example.test", maxBlockPageEntries+19), netip.MustParseAddr("127.0.0.1"))
		assert.True(t, ok)
		assert.True(t, res.IsFiltered)
	})

	t.Run("allowlist_bypasses_stale_cache", func(t *testing.T) {
		srv.recordBlockPageHit("cached-allow.example.", &filtering.Result{
			Reason:     filtering.FilteredBlockList,
			IsFiltered: true,
		})

		res, blocked := srv.LookupBlockedHost("cached-allow.example", netip.MustParseAddr("127.0.0.1"))
		assert.False(t, blocked)
		assert.Equal(t, filtering.NotFilteredAllowList, res.Reason)
	})

	t.Run("custom_lan_hosts_ip_not_hijacked_while_zero_ip_uses_block_page", func(t *testing.T) {
		setts := flt.Settings()
		setts.ProtectionEnabled = true

		lanRes, checkErr := flt.CheckHost("nas.lan", dns.TypeA, setts)
		require.NoError(t, checkErr)
		require.True(t, lanRes.IsFiltered)

		reqLAN := (&dns.Msg{}).SetQuestion("nas.lan.", dns.TypeA)
		dctxLAN := &proxy.DNSContext{
			Req:  reqLAN,
			Addr: netip.MustParseAddrPort("127.0.0.1:54321"),
		}
		respLAN := srv.genDNSFilterMessage(ctx, srv.logger, dctxLAN, &lanRes)
		require.NotNil(t, respLAN)
		require.Len(t, respLAN.Answer, 1)
		aLAN, ok := respLAN.Answer[0].(*dns.A)
		require.True(t, ok)
		assert.Equal(t, "192.168.1.50", aLAN.A.String())

		_, blockedLAN := srv.LookupBlockedHost("nas.lan", netip.MustParseAddr("127.0.0.1"))
		assert.False(t, blockedLAN)

		zeroRes, checkErr := flt.CheckHost("clix.vn", dns.TypeA, setts)
		require.NoError(t, checkErr)
		require.True(t, zeroRes.IsFiltered)

		reqZero := (&dns.Msg{}).SetQuestion("clix.vn.", dns.TypeA)
		dctxZero := &proxy.DNSContext{
			Req:  reqZero,
			Addr: netip.MustParseAddrPort("127.0.0.1:54321"),
		}
		respZero := srv.genDNSFilterMessage(ctx, srv.logger, dctxZero, &zeroRes)
		require.NotNil(t, respZero)
		require.Len(t, respZero.Answer, 1)
		aZero, ok := respZero.Answer[0].(*dns.A)
		require.True(t, ok)
		assert.Equal(t, "127.0.0.1", aZero.A.String())

		_, blockedZero := srv.LookupBlockedHost("clix.vn", netip.MustParseAddr("127.0.0.1"))
		assert.True(t, blockedZero)
	})
}
