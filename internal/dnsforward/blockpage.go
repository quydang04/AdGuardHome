package dnsforward

import (
	"context"
	"net"
	"net/netip"
	"strings"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
	"github.com/AdguardTeam/dnsproxy/proxy"
	"github.com/miekg/dns"
)

const (
	// maxBlockPageEntries bounds the recent blocked-domain lookup cache to keep
	// memory usage tiny.
	maxBlockPageEntries = 2048

	// blockPageEntryTTL is how long a recent DNS block decision is kept for
	// attributing CNAME-blocked HTTP requests.
	blockPageEntryTTL = 30 * time.Minute

	// primaryIPCacheTTL is how long the fallback local IPv4 address is cached.
	primaryIPCacheTTL = time.Minute
)

type blockPageEntry struct {
	res     filtering.Result
	expires time.Time
}

// blockPageTracker stores a fixed-size ring cache of recently blocked hostnames
// and caches the primary outbound IPv4 address.  Its zero value is ready to use.
type blockPageTracker struct {
	mu      sync.RWMutex
	entries map[string]blockPageEntry
	ring    [maxBlockPageEntries]string
	ringIdx int

	primaryV4        netip.Addr
	primaryV4Fetched time.Time
}

func normalizeBlockHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

// recordBlockPageHit records a blocked domain and its filtering result so HTTP
// block-page requests (including CNAME-blocked hosts) can show the exact
// blocking filter list or service.
func (s *Server) recordBlockPageHit(qname string, res *filtering.Result) {
	if s == nil || res == nil {
		return
	}

	host := normalizeBlockHost(qname)
	if host == "" {
		return
	}

	// Copy only lightweight fields needed by DescribeBlock.
	copied := filtering.Result{
		CanonName:   res.CanonName,
		ServiceName: res.ServiceName,
		Reason:      res.Reason,
		IsFiltered:  true,
	}
	if len(res.Rules) > 0 {
		copied.Rules = make([]*filtering.ResultRule, len(res.Rules))
		copy(copied.Rules, res.Rules)
	}

	now := time.Now()
	s.blockPage.mu.Lock()
	defer s.blockPage.mu.Unlock()

	if s.blockPage.entries == nil {
		s.blockPage.entries = make(map[string]blockPageEntry, maxBlockPageEntries)
	}

	if existing, exists := s.blockPage.entries[host]; !exists {
		idx := s.blockPage.ringIdx
		if old := s.blockPage.ring[idx]; old != "" {
			delete(s.blockPage.entries, old)
		}
		s.blockPage.ring[idx] = host
		s.blockPage.ringIdx = (idx + 1) % maxBlockPageEntries
	} else if len(copied.Rules) == 0 && len(existing.res.Rules) > 0 && now.Before(existing.expires) {
		// Keep richer rule metadata from A query when followed by AAAA/HTTPS.
		copied.Rules = existing.res.Rules
		if copied.CanonName == "" {
			copied.CanonName = existing.res.CanonName
		}
	}

	s.blockPage.entries[host] = blockPageEntry{
		res:     copied,
		expires: now.Add(blockPageEntryTTL),
	}
}

// hasCustomRedirectAddr reports whether ips contains a non-zero, non-loopback
// target address (such as a LAN /etc/hosts override like 192.168.1.50).
func hasCustomRedirectAddr(ips []netip.Addr) bool {
	for _, ip := range ips {
		if ip.IsValid() && !ip.IsUnspecified() && !ip.IsLoopback() {
			return true
		}
	}

	return false
}

// hasCustomRedirectIP reports whether resRules redirects the domain to a
// specific non-zero, non-loopback IP address instead of blocking it.
func hasCustomRedirectIP(resRules []*filtering.ResultRule) bool {
	for _, r := range resRules {
		if r == nil {
			continue
		}
		if ip := r.IP; ip.IsValid() {
			if !ip.IsUnspecified() && !ip.IsLoopback() {
				return true
			}

			continue
		}
		if fields := strings.Fields(r.Text); len(fields) >= 2 {
			if ip, err := netip.ParseAddr(fields[0]); err == nil &&
				ip.IsValid() && !ip.IsUnspecified() && !ip.IsLoopback() {
				return true
			}
		}
	}

	return false
}

func (s *Server) genCustomRedirectMessage(
	ctx context.Context,
	req *dns.Msg,
	resRules []*filtering.ResultRule,
) (resp *dns.Msg) {
	ips := ipsFromRules(resRules)
	if len(ips) > 0 {
		return s.genResponseWithIPs(ctx, req, ips)
	}

	return s.NewMsgNODATA(req)
}

// LookupBlockedHost evaluates live filtering rules for host and clientIP first
// (so recent whitelist/unblock changes take effect immediately), then falls
// back to the recent DNS block cache for CNAME-cloaked or client-ID hits.
func (s *Server) LookupBlockedHost(host string, clientIP netip.Addr) (res filtering.Result, blocked bool) {
	if s == nil || s.dnsFilter == nil {
		return filtering.Result{}, false
	}

	host = normalizeBlockHost(host)
	if host == "" {
		return filtering.Result{}, false
	}

	protEnabled, _ := s.UpdatedProtectionStatus(context.Background())
	if !protEnabled {
		return filtering.Result{}, false
	}

	setts := s.dnsFilter.Settings()
	setts.ProtectionEnabled = true

	if clientIP.IsValid() {
		s.dnsFilter.ApplyAdditionalFiltering(clientIP.Unmap(), "", setts)
	} else {
		s.dnsFilter.ApplyBlockedServices(setts)
	}

	checked, err := s.dnsFilter.CheckHost(host, dns.TypeA, setts)
	if err == nil && checked.Reason == filtering.NotFilteredNotFound {
		if v6Res, v6Err := s.dnsFilter.CheckHostRules(host, dns.TypeAAAA, setts); v6Err == nil && v6Res.Reason.Matched() {
			checked = v6Res
		}
	}

	if err == nil {
		switch checked.Reason {
		case filtering.NotFilteredAllowList,
			filtering.Rewritten,
			filtering.RewrittenAutoHosts,
			filtering.RewrittenRule:
			return checked, false
		}

		if hasCustomRedirectIP(checked.Rules) {
			return checked, false
		}

		if checked.IsFiltered && checked.Reason != filtering.FilteredSafeSearch {
			s.recordBlockPageHit(host, &checked)

			return checked, true
		}
	}

	now := time.Now()
	s.blockPage.mu.RLock()
	var (
		entry blockPageEntry
		ok    bool
	)
	if s.blockPage.entries != nil {
		entry, ok = s.blockPage.entries[host]
	}
	s.blockPage.mu.RUnlock()

	if ok && now.Before(entry.expires) && !hasCustomRedirectIP(entry.res.Rules) {
		if cname := normalizeBlockHost(entry.res.CanonName); cname != "" {
			if cnameRes, cErr := s.dnsFilter.CheckHost(cname, dns.TypeA, setts); cErr == nil &&
				cnameRes.Reason == filtering.NotFilteredAllowList {
				return cnameRes, false
			}
		}

		return entry.res, true
	}

	return checked, false
}

// makeResponseBlockPage generates a DNS response pointing A (and AAAA when the
// web server listens on IPv6) to AdGuard Home's reachable IP address, while
// returning NODATA for HTTPS/SVCB records so browsers do not force HTTPS
// upgrades via DNS.
func (s *Server) makeResponseBlockPage(
	ctx context.Context,
	req *dns.Msg,
	dctx *proxy.DNSContext,
) (resp *dns.Msg) {
	switch qt := req.Question[0].Qtype; qt {
	case dns.TypeA:
		ip4 := s.resolveBlockPageIPv4(dctx)

		return s.genARecord(req, ip4)
	case dns.TypeAAAA:
		if ip6, ok := s.resolveBlockPageIPv6(dctx); ok {
			return s.genAAAARecord(req, ip6)
		}

		return s.NewMsgNODATA(req)
	case dns.TypeHTTPS:
		return s.NewMsgNODATA(req)
	default:
		s.logger.DebugContext(ctx, "unexpected qtype for block page", "dns_type", dns.Type(qt))

		return s.NewMsgNODATA(req)
	}
}

func (s *Server) resolveBlockPageIPv4(dctx *proxy.DNSContext) netip.Addr {
	if s.webBindAddr != nil {
		bind := s.webBindAddr()
		if bind.IsValid() {
			addr := bind.Addr().Unmap()
			if addr.Is4() && !addr.IsUnspecified() {
				return addr
			}
		}
	}

	var clientIP netip.Addr
	if dctx != nil && dctx.Addr.IsValid() {
		clientIP = dctx.Addr.Addr().Unmap()
	}

	if !clientIP.IsValid() || clientIP.IsLoopback() {
		return netip.AddrFrom4([4]byte{127, 0, 0, 1})
	}

	if dctx != nil && dctx.Conn != nil {
		if localIP, ok := extractConnIPv4(dctx.Conn.LocalAddr()); ok {
			if clientIP.IsPrivate() || !localIP.IsPrivate() {
				return localIP
			}
		}
	}

	if clientIP.Is4() {
		if localIP, ok := outboundIPv4For(clientIP); ok {
			// If a public internet client queries a server behind 1:1 Cloud NAT,
			// prefer the cached public IPv4 if available.
			if !clientIP.IsPrivate() && !clientIP.IsLinkLocalUnicast() && localIP.IsPrivate() {
				if pubStr := systeminfo.CachedPublicIP(); pubStr != "" {
					if pubIP, err := netip.ParseAddr(pubStr); err == nil && pubIP.Unmap().Is4() {
						return pubIP.Unmap()
					}
				}
			}

			return localIP
		}
	}

	return s.primaryIPv4()
}

func (s *Server) resolveBlockPageIPv6(dctx *proxy.DNSContext) (netip.Addr, bool) {
	if s.webBindAddr != nil {
		bind := s.webBindAddr()
		if bind.IsValid() {
			addr := bind.Addr()
			// When http.address is IPv4 (e.g. 0.0.0.0:80), Go listens on tcp4
			// only. Return NODATA for AAAA so browsers connect via IPv4
			// immediately without Happy Eyeballs delay.
			if addr.Unmap().Is4() {
				return netip.Addr{}, false
			}
			if addr.Is6() && !addr.IsUnspecified() {
				return addr, true
			}
		}
	} else {
		return netip.Addr{}, false
	}

	var clientIP netip.Addr
	if dctx != nil && dctx.Addr.IsValid() {
		clientIP = dctx.Addr.Addr()
	}

	if clientIP.IsLoopback() {
		return netip.IPv6Loopback(), true
	}

	if clientIP.Is6() && !clientIP.Is4In6() {
		if local6, ok := outboundIPv6For(clientIP); ok {
			return local6, true
		}
	}

	return netip.Addr{}, false
}

func (s *Server) primaryIPv4() netip.Addr {
	now := time.Now()
	s.blockPage.mu.RLock()
	cached := s.blockPage.primaryV4
	fresh := cached.IsValid() && now.Sub(s.blockPage.primaryV4Fetched) < primaryIPCacheTTL
	s.blockPage.mu.RUnlock()

	if fresh {
		return cached
	}

	ip, ok := outboundIPv4For(netip.AddrFrom4([4]byte{1, 1, 1, 1}))
	if !ok {
		ip = netip.AddrFrom4([4]byte{127, 0, 0, 1})
	}

	s.blockPage.mu.Lock()
	s.blockPage.primaryV4 = ip
	s.blockPage.primaryV4Fetched = now
	s.blockPage.mu.Unlock()

	return ip
}

func extractConnIPv4(addr net.Addr) (netip.Addr, bool) {
	if addr == nil {
		return netip.Addr{}, false
	}

	var raw net.IP
	switch v := addr.(type) {
	case *net.TCPAddr:
		raw = v.IP
	case *net.UDPAddr:
		raw = v.IP
	default:
		host, _, err := net.SplitHostPort(addr.String())
		if err == nil {
			if parsed, pErr := netip.ParseAddr(host); pErr == nil {
				u := parsed.Unmap()
				if u.Is4() && !u.IsUnspecified() {
					return u, true
				}
			}
		}

		return netip.Addr{}, false
	}

	ip, ok := netip.AddrFromSlice(raw)
	if !ok {
		return netip.Addr{}, false
	}

	u := ip.Unmap()
	if !u.Is4() || u.IsUnspecified() {
		return netip.Addr{}, false
	}

	return u, true
}

func outboundIPv4For(dst netip.Addr) (netip.Addr, bool) {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{
		IP:   dst.AsSlice(),
		Port: 80,
	})
	if err != nil {
		return netip.Addr{}, false
	}
	defer conn.Close()

	return extractConnIPv4(conn.LocalAddr())
}

func outboundIPv6For(dst netip.Addr) (netip.Addr, bool) {
	conn, err := net.DialUDP("udp6", nil, &net.UDPAddr{
		IP:   dst.AsSlice(),
		Port: 80,
	})
	if err != nil {
		return netip.Addr{}, false
	}
	defer conn.Close()

	ua, ok := conn.LocalAddr().(*net.UDPAddr)
	if !ok || ua.IP == nil {
		return netip.Addr{}, false
	}

	ip, ok := netip.AddrFromSlice(ua.IP)
	if !ok || !ip.Is6() || ip.Is4In6() || ip.IsUnspecified() {
		return netip.Addr{}, false
	}

	return ip, true
}
