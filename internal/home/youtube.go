package home

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/golibs/logutil/slogutil"
)

// youtubeConfig is the YouTube ad blocking configuration stored in YAML.
type youtubeConfig struct {
	Enabled          bool     `yaml:"enabled" json:"enabled"`
	RouteServer      string   `yaml:"route_server" json:"route_server"`
	BlockAds         bool     `yaml:"block_ads" json:"block_ads"`
	BlockTracking    bool     `yaml:"block_tracking" json:"block_tracking"`
	CustomDomains    []string `yaml:"custom_domains" json:"custom_domains"`
	RemoteDomainsURL string   `yaml:"remote_domains_url" json:"remote_domains_url"`
}

func defaultYoutubeConfig() *youtubeConfig {
	return &youtubeConfig{
		Enabled:          false,
		RouteServer:      "",
		BlockAds:         true,
		BlockTracking:    true,
		CustomDomains:    []string{},
		RemoteDomainsURL: "",
	}
}

const (
	youtubeSyncInterval       = 60 * time.Second
	youtubeDomainSyncInterval = 30 * time.Minute
	youtubeFetchTimeout       = 15 * time.Second
	youtubeHealthTimeout      = 5 * time.Second
	youtubeResolveTimeout     = 10 * time.Second
	youtubeFailThreshold      = 2
	youtubeRulePrefix         = "||"
	youtubeRuleSuffix         = "^"
	youtubeRuleComment        = "! YouTube ad blocking (managed by AdGuard Home)"
	youtubeMaxResponseSize    = 1 << 20 // 1 MB
)

var youtubeSNITestNames = []string{
	"youtube.com",
	"rr1---sn-42u-nbozl.googlevideo.com",
}

func youtubeAdDomains() []string {
	return []string{
		"ads.youtube.com",
		"ad.doubleclick.net",
		"www.googleadservices.com",
		"pagead2.googlesyndication.com",
		"video-ad-stats.googlesyndication.com",
		"s0.2mdn.net",
		"s1.2mdn.net",
		"googleads.g.doubleclick.net",
		"googleads4.g.doubleclick.net",
		"www.google-analytics.com",
		"ssl.google-analytics.com",
		"google-analytics.com",
		"stats.g.doubleclick.net",
		"adservice.google.com",
		"adservice.google.com.vn",
		"pagead-googlehosted.l.google.com",
		"tpc.googlesyndication.com",
		"www.youtube-nocookie.com",
		"static.doubleclick.net",
		"m.doubleclick.net",
		"mediavisor.doubleclick.net",
		"yt3.ggpht.com",
	}
}

func youtubeTrackingDomains() []string {
	return []string{
		"www.google-analytics.com",
		"ssl.google-analytics.com",
		"google-analytics.com",
		"analytics.youtube.com",
		"stats.g.doubleclick.net",
		"clients1.google.com",
		"video-stats.l.google.com",
		"www.googletagmanager.com",
		"www.googletagservices.com",
		"googletagmanager.com",
		"googletagservices.com",
	}
}

func youtubeRewriteDomains() []string {
	return []string{
		"youtube.com",
		"*.youtube.com",
		"youtubei.googleapis.com",
		"*.youtubei.googleapis.com",
		"googlevideo.com",
		"*.googlevideo.com",
	}
}

// youtubeRemoteDomains is the JSON format for a remote domain list source.
type youtubeRemoteDomains struct {
	AdDomains      []string `json:"ad_domains"`
	TrackDomains   []string `json:"tracking_domains"`
	RewriteDomains []string `json:"rewrite_domains"`
}

// youtubeIPStatus tracks the health status of a single route server IP.
type youtubeIPStatus struct {
	IP        string `json:"ip"`
	Healthy   bool   `json:"healthy"`
	FailCount int    `json:"fail_count"`
	LastCheck string `json:"last_check"`
}

// youtubeManager handles the YouTube ad blocking integration: health checking
// the route server, managing DNS rewrites, and managing ad blocking rules.
type youtubeManager struct {
	logger *slog.Logger

	mu sync.Mutex

	// healthyIPs are the currently healthy route server IPs.
	healthyIPs []string

	// allIPs are all resolved IPs from the last sync.
	allIPs []string

	// ipStatuses tracks health status per IP.
	ipStatuses map[string]*youtubeIPStatus

	// failCounts tracks consecutive health check failures per IP.
	failCounts map[string]int

	// active indicates whether the manager is currently running.
	active bool

	// cancel stops the sync goroutine.
	cancel context.CancelFunc

	// stats
	startedAt      time.Time
	lastSyncTime   time.Time
	totalSyncs     int
	lastSyncStatus string
	blockedRules   int
	activeRewrites int

	// remoteDomains stores domains fetched from the remote URL.
	remoteDomains     *youtubeRemoteDomains
	lastDomainSync    time.Time
	lastDomainStatus  string
	remoteDomainCount int

	// queryStats tracks per-query YouTube statistics.
	queryStats *youtubeQueryStats
}

var ytManager *youtubeManager

func initYoutubeManager(logger *slog.Logger) {
	ytManager = &youtubeManager{
		logger:     logger.With(slogutil.KeyPrefix, "youtube"),
		healthyIPs: nil,
		failCounts: make(map[string]int),
		ipStatuses: make(map[string]*youtubeIPStatus),
		queryStats: newYoutubeQueryStats(),
	}
}

// RecordYoutubeQuery records a DNS query result in the YouTube statistics.
// queryType must be one of "ad", "tracking", or "rewrite".  It is safe to
// call when ytManager is nil.
func RecordYoutubeQuery(domain, queryType string) {
	if ytManager == nil || ytManager.queryStats == nil {
		return
	}

	ytManager.queryStats.RecordQuery(domain, queryType)
}

// start begins the YouTube ad blocking manager if the config is enabled.
func (m *youtubeManager) start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.active {
		return
	}

	cfg := getYoutubeConf()
	if !cfg.Enabled {
		return
	}

	m.logger.InfoContext(ctx, "starting youtube ad blocking")

	m.applyBlockingRules(ctx, cfg)
	m.active = true
	m.startedAt = time.Now()
	m.totalSyncs = 0
	m.lastSyncStatus = "starting"

	syncCtx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel

	go m.syncLoop(syncCtx, cfg)
}

// stop halts the YouTube ad blocking manager and removes all managed rules.
func (m *youtubeManager) stop(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return
	}

	m.logger.InfoContext(ctx, "stopping youtube ad blocking")

	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}

	m.removeAllRewrites(ctx)
	m.removeBlockingRules(ctx)
	m.healthyIPs = nil
	m.allIPs = nil
	m.failCounts = make(map[string]int)
	m.ipStatuses = make(map[string]*youtubeIPStatus)
	m.active = false
	m.lastSyncStatus = "stopped"
	m.blockedRules = 0
	m.activeRewrites = 0
}

// restart stops and starts the manager with the current config.
func (m *youtubeManager) restart(ctx context.Context) {
	m.stop(ctx)
	m.start(ctx)
}

func getYoutubeConf() youtubeConfig {
	config.RLock()
	defer config.RUnlock()

	cfg := config.YouTube
	if cfg == nil {
		return *defaultYoutubeConfig()
	}

	return *cfg
}

// syncLoop periodically resolves the route server and updates DNS rewrites.
func (m *youtubeManager) syncLoop(ctx context.Context, cfg youtubeConfig) {
	m.logger.InfoContext(ctx, "sync loop started", "route_server", cfg.RouteServer)

	m.fetchRemoteDomains(ctx, cfg)
	m.doSync(ctx, cfg)

	ipTicker := time.NewTicker(youtubeSyncInterval)
	defer ipTicker.Stop()

	domainTicker := time.NewTicker(youtubeDomainSyncInterval)
	defer domainTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			m.logger.InfoContext(ctx, "sync loop stopped")

			return
		case <-domainTicker.C:
			cfg = getYoutubeConf()
			if m.fetchRemoteDomains(ctx, cfg) {
				m.applyBlockingRules(ctx, cfg)
				m.forceRewriteUpdate(ctx, cfg)
			}
		case <-ipTicker.C:
			cfg = getYoutubeConf()
			m.doSync(ctx, cfg)
		}
	}
}

// forceRewriteUpdate re-applies DNS rewrites with current healthy IPs.
func (m *youtubeManager) forceRewriteUpdate(ctx context.Context, cfg youtubeConfig) {
	m.mu.Lock()
	healthy := m.healthyIPs
	m.mu.Unlock()

	if len(healthy) > 0 {
		m.updateRewrites(ctx, healthy, cfg.CustomDomains)
	}
}

// doSync resolves the route server, health-checks IPs, and updates rewrites.
func (m *youtubeManager) doSync(ctx context.Context, cfg youtubeConfig) {
	if cfg.RouteServer == "" {
		m.mu.Lock()
		m.lastSyncStatus = "no route server configured"
		m.lastSyncTime = time.Now()
		m.mu.Unlock()
		m.logger.DebugContext(ctx, "no route server configured, skipping sync")

		return
	}

	ips := m.resolveRouteServer(ctx, cfg.RouteServer)
	if len(ips) == 0 {
		m.mu.Lock()
		m.lastSyncStatus = "failed to resolve route server"
		m.lastSyncTime = time.Now()
		m.totalSyncs++
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "failed to resolve route server", "server", cfg.RouteServer)

		return
	}

	healthy := m.healthCheckIPs(ctx, ips)

	m.mu.Lock()
	oldIPs := m.healthyIPs
	m.healthyIPs = healthy
	m.allIPs = ips
	m.lastSyncTime = time.Now()
	m.totalSyncs++

	now := m.lastSyncTime.Format(time.RFC3339)
	for _, ip := range ips {
		isHealthy := containsStr(healthy, ip)
		m.ipStatuses[ip] = &youtubeIPStatus{
			IP:        ip,
			Healthy:   isHealthy,
			FailCount: m.failCounts[ip],
			LastCheck: now,
		}
	}

	if len(healthy) > 0 {
		m.lastSyncStatus = fmt.Sprintf("ok: %d/%d IPs healthy", len(healthy), len(ips))
	} else {
		m.lastSyncStatus = fmt.Sprintf("warning: 0/%d IPs healthy", len(ips))
	}
	m.mu.Unlock()

	if len(healthy) > 0 {
		if !ipsEqual(oldIPs, healthy) {
			m.logger.InfoContext(ctx, "updating dns rewrites", "healthy_ips", healthy)
			m.updateRewrites(ctx, healthy, cfg.CustomDomains)
		}
	} else {
		m.logger.WarnContext(ctx, "no healthy route server IPs, removing rewrites")
		m.removeAllRewrites(ctx)
	}
}

// resolveRouteServer resolves the route server hostname to IP addresses.
func (m *youtubeManager) resolveRouteServer(ctx context.Context, server string) []string {
	resolver := &net.Resolver{
		PreferGo: true,
	}

	resolveCtx, cancel := context.WithTimeout(ctx, youtubeResolveTimeout)
	defer cancel()

	addrs, err := resolver.LookupHost(resolveCtx, server)
	if err != nil {
		m.logger.WarnContext(ctx, "resolving route server", slogutil.KeyError, err)

		return nil
	}

	m.logger.DebugContext(ctx, "resolved route server", "server", server, "addrs", addrs)

	return addrs
}

// fetchRemoteDomains fetches domain lists from the configured remote URL and
// returns true if the lists changed.
func (m *youtubeManager) fetchRemoteDomains(ctx context.Context, cfg youtubeConfig) bool {
	if cfg.RemoteDomainsURL == "" {
		m.mu.Lock()
		changed := m.remoteDomains != nil
		m.remoteDomains = nil
		m.lastDomainStatus = "no remote URL configured"
		m.remoteDomainCount = 0
		m.mu.Unlock()

		return changed
	}

	m.logger.DebugContext(ctx, "fetching remote domain list", "url", cfg.RemoteDomainsURL)

	client := &http.Client{Timeout: youtubeFetchTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.RemoteDomainsURL, nil)
	if err != nil {
		m.mu.Lock()
		m.lastDomainStatus = fmt.Sprintf("bad URL: %s", err)
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "creating remote domain request", slogutil.KeyError, err)

		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		m.mu.Lock()
		m.lastDomainStatus = fmt.Sprintf("fetch failed: %s", err)
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "fetching remote domains", slogutil.KeyError, err)

		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		m.mu.Lock()
		m.lastDomainStatus = fmt.Sprintf("HTTP %d", resp.StatusCode)
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "remote domain list returned non-200", "status", resp.StatusCode)

		return false
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, youtubeMaxResponseSize))
	if err != nil {
		m.mu.Lock()
		m.lastDomainStatus = fmt.Sprintf("read failed: %s", err)
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "reading remote domain response", slogutil.KeyError, err)

		return false
	}

	var remote youtubeRemoteDomains
	if err = json.Unmarshal(body, &remote); err != nil {
		m.mu.Lock()
		m.lastDomainStatus = fmt.Sprintf("invalid JSON: %s", err)
		m.mu.Unlock()
		m.logger.WarnContext(ctx, "parsing remote domain list", slogutil.KeyError, err)

		return false
	}

	totalCount := len(remote.AdDomains) + len(remote.TrackDomains) + len(remote.RewriteDomains)

	m.mu.Lock()
	old := m.remoteDomains
	m.remoteDomains = &remote
	m.lastDomainSync = time.Now()
	m.lastDomainStatus = fmt.Sprintf("ok: %d domains fetched", totalCount)
	m.remoteDomainCount = totalCount
	m.mu.Unlock()

	changed := !remoteDomainsEqual(old, &remote)
	if changed {
		m.logger.InfoContext(ctx, "remote domain list updated",
			"ad", len(remote.AdDomains),
			"tracking", len(remote.TrackDomains),
			"rewrite", len(remote.RewriteDomains),
		)
	}

	return changed
}

func remoteDomainsEqual(a, b *youtubeRemoteDomains) bool {
	if a == nil || b == nil {
		return a == b
	}

	return slicesEqual(a.AdDomains, b.AdDomains) &&
		slicesEqual(a.TrackDomains, b.TrackDomains) &&
		slicesEqual(a.RewriteDomains, b.RewriteDomains)
}

func slicesEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

// mergedAdDomains returns hardcoded ad domains merged with remote ones.
func (m *youtubeManager) mergedAdDomains() []string {
	base := youtubeAdDomains()

	m.mu.Lock()
	remote := m.remoteDomains
	m.mu.Unlock()

	if remote == nil {
		return base
	}

	return mergeUnique(base, remote.AdDomains)
}

// mergedTrackingDomains returns hardcoded tracking domains merged with remote ones.
func (m *youtubeManager) mergedTrackingDomains() []string {
	base := youtubeTrackingDomains()

	m.mu.Lock()
	remote := m.remoteDomains
	m.mu.Unlock()

	if remote == nil {
		return base
	}

	return mergeUnique(base, remote.TrackDomains)
}

// mergedRewriteDomains returns hardcoded rewrite domains merged with remote ones.
func (m *youtubeManager) mergedRewriteDomains() []string {
	base := youtubeRewriteDomains()

	m.mu.Lock()
	remote := m.remoteDomains
	m.mu.Unlock()

	if remote == nil {
		return base
	}

	return mergeUnique(base, remote.RewriteDomains)
}

func mergeUnique(base, extra []string) []string {
	seen := make(map[string]bool, len(base))
	for _, s := range base {
		seen[strings.ToLower(s)] = true
	}

	merged := make([]string, len(base))
	copy(merged, base)

	for _, s := range extra {
		lower := strings.ToLower(strings.TrimSpace(s))
		if lower != "" && !seen[lower] {
			seen[lower] = true
			merged = append(merged, s)
		}
	}

	return merged
}

// healthCheckIPs probes each IP with TCP+TLS SNI and returns healthy ones.
func (m *youtubeManager) healthCheckIPs(ctx context.Context, ips []string) []string {
	var healthy []string

	for _, ip := range ips {
		if m.probeIP(ctx, ip) {
			m.failCounts[ip] = 0
			healthy = append(healthy, ip)
		} else {
			m.failCounts[ip]++
			if m.failCounts[ip] < youtubeFailThreshold {
				healthy = append(healthy, ip)
			} else {
				m.logger.WarnContext(ctx, "ip failed health check threshold", "ip", ip, "fails", m.failCounts[ip])
			}
		}
	}

	return healthy
}

// probeIP checks if the IP is reachable via TCP and responds to TLS SNI.
func (m *youtubeManager) probeIP(ctx context.Context, ip string) bool {
	addr := net.JoinHostPort(ip, "443")

	for _, sni := range youtubeSNITestNames {
		dialer := &net.Dialer{Timeout: youtubeHealthTimeout}
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, &tls.Config{
			ServerName:         sni,
			InsecureSkipVerify: true,
		})
		if err != nil {
			m.logger.DebugContext(ctx, "tls probe failed", "ip", ip, "sni", sni, slogutil.KeyError, err)

			return false
		}
		conn.Close()
	}

	return true
}

// updateRewrites sets DNS rewrites for all YouTube domains to point to healthy
// route server IPs.
func (m *youtubeManager) updateRewrites(ctx context.Context, ips []string, customDomains []string) {
	flt := globalContext.filters
	if flt == nil {
		return
	}

	m.removeAllRewrites(ctx)

	domains := m.mergedRewriteDomains()
	domains = append(domains, customDomains...)

	for _, domain := range domains {
		for _, ip := range ips {
			rw := &filtering.LegacyRewrite{
				Domain:  domain,
				Answer:  ip,
				Enabled: true,
			}

			if err := rw.Normalize(ctx, m.logger); err != nil {
				m.logger.WarnContext(ctx, "normalizing rewrite", "domain", domain, slogutil.KeyError, err)

				continue
			}

			flt.AddRewrite(ctx, rw)
		}
	}

	m.mu.Lock()
	m.activeRewrites = len(domains) * len(ips)
	m.mu.Unlock()
	m.logger.InfoContext(ctx, "dns rewrites updated", "domains", len(domains), "ips", len(ips))
}

// removeAllRewrites removes all YouTube-managed DNS rewrites.
func (m *youtubeManager) removeAllRewrites(ctx context.Context) {
	flt := globalContext.filters
	if flt == nil {
		return
	}

	domains := m.mergedRewriteDomains()
	cfg := getYoutubeConf()
	domains = append(domains, cfg.CustomDomains...)

	flt.RemoveRewritesByDomains(ctx, domains)
}

// applyBlockingRules adds user rules for blocking YouTube ad/tracking domains.
func (m *youtubeManager) applyBlockingRules(ctx context.Context, cfg youtubeConfig) {
	flt := globalContext.filters
	if flt == nil {
		return
	}

	var rules []string

	if cfg.BlockAds {
		for _, d := range m.mergedAdDomains() {
			rules = append(rules, youtubeRulePrefix+d+youtubeRuleSuffix)
		}
	}

	if cfg.BlockTracking {
		for _, d := range m.mergedTrackingDomains() {
			rule := youtubeRulePrefix + d + youtubeRuleSuffix
			if !containsStr(rules, rule) {
				rules = append(rules, rule)
			}
		}
	}

	if len(rules) == 0 {
		return
	}

	flt.AddYouTubeRules(ctx, rules)
	m.blockedRules = len(rules)
	m.logger.InfoContext(ctx, "blocking rules applied", "count", len(rules))
}

// removeBlockingRules removes all YouTube-managed user rules.
func (m *youtubeManager) removeBlockingRules(ctx context.Context) {
	flt := globalContext.filters
	if flt == nil {
		return
	}

	flt.RemoveYouTubeRules(ctx)
	m.logger.InfoContext(ctx, "blocking rules removed")
}

func ipsEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func containsStr(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}

	return false
}

// --- HTTP API handlers ---

type youtubeConfigJSON struct {
	Enabled          bool     `json:"enabled"`
	RouteServer      string   `json:"route_server"`
	BlockAds         bool     `json:"block_ads"`
	BlockTracking    bool     `json:"block_tracking"`
	CustomDomains    []string `json:"custom_domains"`
	RemoteDomainsURL string   `json:"remote_domains_url"`
	AdDomains        []string `json:"ad_domains"`
	TrackDomains     []string `json:"tracking_domains"`
	RewriteDomains   []string `json:"rewrite_domains"`
}

// youtubeStatusJSON is the response for the YouTube status API.
type youtubeStatusJSON struct {
	Active            bool               `json:"active"`
	LastSyncTime      string             `json:"last_sync_time"`
	LastSyncStatus    string             `json:"last_sync_status"`
	TotalSyncs        int                `json:"total_syncs"`
	Uptime            string             `json:"uptime"`
	HealthyIPs        int                `json:"healthy_ips"`
	TotalIPs          int                `json:"total_ips"`
	IPStatuses        []*youtubeIPStatus `json:"ip_statuses"`
	BlockedRules      int                `json:"blocked_rules"`
	ActiveRewrites    int                `json:"active_rewrites"`
	RouteServer       string             `json:"route_server"`
	SyncInterval      int                `json:"sync_interval"`
	LastDomainSync    string             `json:"last_domain_sync"`
	LastDomainStatus  string             `json:"last_domain_status"`
	RemoteDomainCount int                `json:"remote_domain_count"`
	DomainSyncInterval int              `json:"domain_sync_interval"`
}

func (web *webAPI) registerYouTubeHandlers() {
	web.httpReg.Register(http.MethodGet, "/control/youtube/config", web.handleGetYoutubeConfig)
	web.httpReg.Register(http.MethodPut, "/control/youtube/config/update", web.handlePutYoutubeConfig)
	web.httpReg.Register(http.MethodGet, "/control/youtube/status", web.handleGetYoutubeStatus)
	web.httpReg.Register(http.MethodGet, "/control/youtube/stats", web.handleGetYoutubeStats)
}

func (web *webAPI) handleGetYoutubeConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var resp youtubeConfigJSON
	func() {
		config.RLock()
		defer config.RUnlock()

		cfg := config.YouTube
		if cfg == nil {
			cfg = defaultYoutubeConfig()
		}

		resp = youtubeConfigJSON{
			Enabled:          cfg.Enabled,
			RouteServer:      cfg.RouteServer,
			BlockAds:         cfg.BlockAds,
			BlockTracking:    cfg.BlockTracking,
			CustomDomains:    cfg.CustomDomains,
			RemoteDomainsURL: cfg.RemoteDomainsURL,
		}
	}()

	if ytManager != nil {
		resp.AdDomains = ytManager.mergedAdDomains()
		resp.TrackDomains = ytManager.mergedTrackingDomains()
		resp.RewriteDomains = ytManager.mergedRewriteDomains()
	} else {
		resp.AdDomains = youtubeAdDomains()
		resp.TrackDomains = youtubeTrackingDomains()
		resp.RewriteDomains = youtubeRewriteDomains()
	}

	if resp.CustomDomains == nil {
		resp.CustomDomains = []string{}
	}

	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, resp)
}

func (web *webAPI) handlePutYoutubeConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req youtubeConfigJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusBadRequest, "json decode: %s", err)

		return
	}

	cleanDomains := make([]string, 0, len(req.CustomDomains))
	for _, d := range req.CustomDomains {
		d = strings.TrimSpace(d)
		if d != "" {
			cleanDomains = append(cleanDomains, d)
		}
	}

	func() {
		config.Lock()
		defer config.Unlock()

		if config.YouTube == nil {
			config.YouTube = defaultYoutubeConfig()
		}

		config.YouTube.Enabled = req.Enabled
		config.YouTube.RouteServer = strings.TrimSpace(req.RouteServer)
		config.YouTube.BlockAds = req.BlockAds
		config.YouTube.BlockTracking = req.BlockTracking
		config.YouTube.CustomDomains = cleanDomains
		config.YouTube.RemoteDomainsURL = strings.TrimSpace(req.RemoteDomainsURL)
	}()

	web.logger.InfoContext(ctx, "youtube config updated", "enabled", req.Enabled)
	web.confModifier.Apply(ctx)

	if ytManager != nil {
		ytManager.restart(ctx)
	}

	aghhttp.OK(ctx, web.logger, w)
}

func (web *webAPI) handleGetYoutubeStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if ytManager == nil || ytManager.queryStats == nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusServiceUnavailable, "youtube manager not initialized")

		return
	}

	stats := ytManager.queryStats.getStats()
	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, &stats)
}

func (web *webAPI) handleGetYoutubeStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	resp := youtubeStatusJSON{
		SyncInterval:       int(youtubeSyncInterval.Seconds()),
		DomainSyncInterval: int(youtubeDomainSyncInterval.Seconds()),
	}

	if ytManager != nil {
		ytManager.mu.Lock()
		resp.Active = ytManager.active
		resp.TotalSyncs = ytManager.totalSyncs
		resp.LastSyncStatus = ytManager.lastSyncStatus
		resp.HealthyIPs = len(ytManager.healthyIPs)
		resp.TotalIPs = len(ytManager.allIPs)
		resp.BlockedRules = ytManager.blockedRules
		resp.ActiveRewrites = ytManager.activeRewrites
		resp.LastDomainStatus = ytManager.lastDomainStatus
		resp.RemoteDomainCount = ytManager.remoteDomainCount

		if !ytManager.lastSyncTime.IsZero() {
			resp.LastSyncTime = ytManager.lastSyncTime.Format(time.RFC3339)
		}

		if !ytManager.lastDomainSync.IsZero() {
			resp.LastDomainSync = ytManager.lastDomainSync.Format(time.RFC3339)
		}

		if ytManager.active && !ytManager.startedAt.IsZero() {
			resp.Uptime = time.Since(ytManager.startedAt).Truncate(time.Second).String()
		}

		statuses := make([]*youtubeIPStatus, 0, len(ytManager.ipStatuses))
		for _, s := range ytManager.ipStatuses {
			statuses = append(statuses, s)
		}
		resp.IPStatuses = statuses
		ytManager.mu.Unlock()
	}

	cfg := getYoutubeConf()
	resp.RouteServer = cfg.RouteServer

	if resp.IPStatuses == nil {
		resp.IPStatuses = []*youtubeIPStatus{}
	}

	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, resp)
}

// startYoutubeManager initializes and starts the YouTube ad blocking manager.
// It waits for the DNS filter system to be ready before applying rules.
func startYoutubeManager(ctx context.Context, logger *slog.Logger) {
	initYoutubeManager(logger)

	for i := 0; i < 30; i++ {
		if globalContext.filters != nil && globalContext.filters.IsStarted() {
			break
		}

		time.Sleep(500 * time.Millisecond)
	}

	ytManager.start(ctx)
}

// stopYoutubeManager stops the YouTube ad blocking manager.
func stopYoutubeManager(ctx context.Context) {
	if ytManager != nil {
		ytManager.stop(ctx)
	}
}

// youtubeStatus returns a formatted status string for logging.
func youtubeStatus() string {
	if ytManager == nil {
		return "not initialized"
	}

	ytManager.mu.Lock()
	defer ytManager.mu.Unlock()

	if !ytManager.active {
		return "disabled"
	}

	return fmt.Sprintf("active, %d healthy IPs", len(ytManager.healthyIPs))
}
