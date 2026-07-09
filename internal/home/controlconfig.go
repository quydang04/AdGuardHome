package home

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/netip"
	"net/http"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/dhcpd"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/querylog"
	"github.com/AdguardTeam/AdGuardHome/internal/stats"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/AdguardTeam/golibs/timeutil"
)

// exportConfig is the JSON structure used for exporting and importing all
// AdGuard Home settings.
type exportConfig struct {
	DNS           *exportDNSConfig            `json:"dns"`
	DHCP          *dhcpd.ServerConfig         `json:"dhcp,omitempty"`
	Filtering     *exportFilteringConfig      `json:"filtering,omitempty"`
	QueryLog      *exportQueryLogConfig       `json:"querylog,omitempty"`
	Statistics    *exportStatsConfig          `json:"statistics,omitempty"`
	Clients       *exportClientsConfig        `json:"clients,omitempty"`
	TLS           *tlsConfigSettings          `json:"tls,omitempty"`
	SafeBrowsing  *bool                       `json:"safebrowsing_enabled,omitempty"`
	Parental      *bool                       `json:"parental_enabled,omitempty"`
	SafeSearch    *filtering.SafeSearchConfig `json:"safesearch,omitempty"`
	Notifications *exportNotificationsConfig  `json:"notifications,omitempty"`
	General       *exportGeneralConfig        `json:"general,omitempty"`
	YouTube       *youtubeConfig              `json:"youtube,omitempty"`
	ACME          *exportACMEConfig           `json:"acme,omitempty"`
}

// exportGeneralConfig is the general settings portion of the export.
type exportGeneralConfig struct {
	Language string `json:"language,omitempty"`
	Theme    string `json:"theme,omitempty"`
}

// exportDNSConfig is the DNS portion of the export.
type exportDNSConfig struct {
	UpstreamDNS      []string       `json:"upstream_dns"`
	BootstrapDNS     []string       `json:"bootstrap_dns"`
	FallbackDNS      []string       `json:"fallback_dns"`
	BindHosts        []netip.Addr   `json:"bind_hosts,omitempty"`
	Port             uint16         `json:"port,omitempty"`
	UpstreamMode     string         `json:"upstream_mode,omitempty"`
	Ratelimit        uint32         `json:"ratelimit,omitempty"`
	EnableDNSSEC     bool           `json:"enable_dnssec,omitempty"`
	AAAADisabled     bool           `json:"aaaa_disabled,omitempty"`
	CacheEnabled     bool           `json:"cache_enabled,omitempty"`
	CacheSize        uint32         `json:"cache_size,omitempty"`
	CacheMinTTL      uint32         `json:"cache_ttl_min,omitempty"`
	CacheMaxTTL      uint32         `json:"cache_ttl_max,omitempty"`
	CacheOptimistic  bool           `json:"cache_optimistic,omitempty"`
	UsePrivateRDNS   bool           `json:"use_private_ptr_resolvers,omitempty"`
	LocalPTRUpstream []string       `json:"local_ptr_upstreams,omitempty"`
	UseDNS64         bool           `json:"use_dns64,omitempty"`
	DNS64Prefixes    []netip.Prefix `json:"dns64_prefixes,omitempty"`
	HostsFileEnabled bool           `json:"hostsfile_enabled,omitempty"`
	ServePlainDNS    bool           `json:"serve_plain_dns,omitempty"`

	// Additional DNS fields.
	AnonymizeClientIP  bool              `json:"anonymize_client_ip,omitempty"`
	UpstreamTimeout    timeutil.Duration `json:"upstream_timeout,omitempty"`
	PrivateNets        []netutil.Prefix  `json:"private_networks,omitempty"`
	ServeHTTP3         bool              `json:"serve_http3,omitempty"`
	UseHTTP3Upstreams  bool              `json:"use_http3_upstreams,omitempty"`
	PendingRequests    *bool             `json:"pending_requests,omitempty"`

	// Additional dnsforward.Config fields.
	RatelimitSubnetLenIPv4 uint             `json:"ratelimit_subnet_len_ipv4,omitempty"`
	RatelimitSubnetLenIPv6 uint             `json:"ratelimit_subnet_len_ipv6,omitempty"`
	RatelimitWhitelist     []netip.Addr     `json:"ratelimit_whitelist,omitempty"`
	RefuseAny              bool             `json:"refuse_any,omitempty"`
	FastestTimeout         timeutil.Duration `json:"fastest_timeout,omitempty"`
	AllowedClients         []string         `json:"allowed_clients,omitempty"`
	DisallowedClients      []string         `json:"disallowed_clients,omitempty"`
	BlockedHosts           []string         `json:"blocked_hosts,omitempty"`
	TrustedProxies         []netutil.Prefix `json:"trusted_proxies,omitempty"`
	CacheOptimisticTTL     timeutil.Duration `json:"cache_optimistic_answer_ttl,omitempty"`
	CacheOptimisticMaxAge  timeutil.Duration `json:"cache_optimistic_max_age,omitempty"`
	BogusNXDomain          []string         `json:"bogus_nxdomain,omitempty"`
	EDNSClientSubnet       *dnsforward.EDNSClientSubnet `json:"edns_client_subnet,omitempty"`
	MaxGoroutines          uint             `json:"max_goroutines,omitempty"`
	HandleDDR              bool             `json:"handle_ddr,omitempty"`
	IpsetList              []string         `json:"ipset,omitempty"`
	BootstrapPreferIPv6    bool             `json:"bootstrap_prefer_ipv6,omitempty"`
}

// exportFilteringConfig is the filtering portion of the export.
type exportFilteringConfig struct {
	Enabled          bool                   `json:"enabled"`
	Interval         uint32                 `json:"interval,omitempty"`
	Filters          []filtering.FilterYAML `json:"filters,omitempty"`
	WhitelistFilters []filtering.FilterYAML `json:"whitelist_filters,omitempty"`
	UserRules        []string               `json:"user_rules,omitempty"`

	// Additional filtering fields.
	FilteringEnabled      bool                     `json:"filtering_enabled,omitempty"`
	BlockingMode          string                   `json:"blocking_mode,omitempty"`
	BlockingIPv4          *netip.Addr              `json:"blocking_ipv4,omitempty"`
	BlockingIPv6          *netip.Addr              `json:"blocking_ipv6,omitempty"`
	BlockedServices       *filtering.BlockedServices `json:"blocked_services,omitempty"`
	ParentalBlockHost     string                   `json:"parental_block_host,omitempty"`
	SafeBrowsingBlockHost string                   `json:"safebrowsing_block_host,omitempty"`
	Rewrites              []*exportRewrite         `json:"rewrites,omitempty"`
	RewritesEnabled       bool                     `json:"rewrites_enabled,omitempty"`
	BlockedResponseTTL    uint32                   `json:"blocked_response_ttl,omitempty"`
	SafeBrowsingCacheSize uint                     `json:"safebrowsing_cache_size,omitempty"`
	SafeSearchCacheSize   uint                     `json:"safesearch_cache_size,omitempty"`
	ParentalCacheSize     uint                     `json:"parental_cache_size,omitempty"`
	CacheTime             uint                     `json:"cache_time,omitempty"`
}

// exportRewrite is a DNS rewrite for export/import.
type exportRewrite struct {
	Domain  string `json:"domain"`
	Answer  string `json:"answer"`
	Enabled bool   `json:"enabled"`
}

// exportQueryLogConfig is the query log portion of the export.
type exportQueryLogConfig struct {
	Enabled        bool     `json:"enabled"`
	Interval       uint32   `json:"interval,omitempty"`
	AnonymizeClientIP bool  `json:"anonymize_client_ip,omitempty"`
	Ignored        []string `json:"ignored,omitempty"`
	MemSize        uint     `json:"size_memory,omitempty"`
	IgnoredEnabled bool     `json:"ignored_enabled,omitempty"`
	FileEnabled    bool     `json:"file_enabled,omitempty"`
}

// exportStatsConfig is the statistics portion of the export.
type exportStatsConfig struct {
	Enabled        bool     `json:"enabled"`
	Interval       uint32   `json:"interval,omitempty"`
	Ignored        []string `json:"ignored,omitempty"`
	IgnoredEnabled bool     `json:"ignored_enabled,omitempty"`
}

// exportClientsConfig is the clients portion of the export.
type exportClientsConfig struct {
	Sources    *clientSourcesConfig `json:"runtime_sources,omitempty"`
	Persistent []*clientObject      `json:"persistent,omitempty"`
}

// exportNotificationsConfig is the notifications portion of the export.
type exportNotificationsConfig struct {
	Telegram *exportTelegramConfig `json:"telegram,omitempty"`
}

// exportTelegramConfig is the Telegram notification config for export.
type exportTelegramConfig struct {
	Enabled         bool              `json:"enabled"`
	BotToken        string            `json:"bot_token,omitempty"`
	ChatID          string            `json:"chat_id,omitempty"`
	CPUThreshold    float64           `json:"cpu_threshold,omitempty"`
	MemoryThreshold float64           `json:"memory_threshold,omitempty"`
	DiskThreshold   float64           `json:"disk_threshold,omitempty"`
	CheckInterval   timeutil.Duration `json:"check_interval,omitempty"`
	Cooldown        timeutil.Duration `json:"cooldown,omitempty"`
	CustomMessage   string            `json:"custom_message,omitempty"`
}

// exportACMEConfig is the ACME ("SSL/TLS issue") portion of the export.  The
// ACME account private key and status fields are intentionally excluded,
// mirroring how the main TLS private key is cleared before export.
type exportACMEConfig struct {
	Email              string   `json:"email,omitempty"`
	Domains            []string `json:"domains,omitempty"`
	Challenge          string   `json:"challenge,omitempty"`
	CloudflareAPIToken string   `json:"cloudflare_api_token,omitempty"`
	RenewBeforeDays    int      `json:"renew_before_days,omitempty"`
	Enabled            bool     `json:"enabled"`
	AutoRenew          bool     `json:"auto_renew"`
}

// handleExportSettings handles GET /control/settings/export.
func (web *webAPI) handleExportSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	export := &exportConfig{}

	config.RLock()
	defer config.RUnlock()

	// General config.
	export.General = &exportGeneralConfig{
		Language: config.Language,
		Theme:    string(config.Theme),
	}

	// DNS config.
	dnsConf := &exportDNSConfig{
		BindHosts:          config.DNS.BindHosts,
		Port:               config.DNS.Port,
		HostsFileEnabled:   config.DNS.HostsFileEnabled,
		ServePlainDNS:      config.DNS.ServePlainDNS,
		UsePrivateRDNS:     config.DNS.UsePrivateRDNS,
		LocalPTRUpstream:   config.DNS.PrivateRDNSResolvers,
		UseDNS64:           config.DNS.UseDNS64,
		DNS64Prefixes:      config.DNS.DNS64Prefixes,
		AnonymizeClientIP:  config.DNS.AnonymizeClientIP,
		UpstreamTimeout:    config.DNS.UpstreamTimeout,
		PrivateNets:        config.DNS.PrivateNets,
		ServeHTTP3:         config.DNS.ServeHTTP3,
		UseHTTP3Upstreams:  config.DNS.UseHTTP3Upstreams,
	}

	if pr := config.DNS.PendingRequests; pr != nil {
		dnsConf.PendingRequests = &pr.Enabled
	}

	if s := globalContext.dnsServer; s != nil {
		c := dnsforward.Config{}
		s.WriteDiskConfig(&c)

		dnsConf.UpstreamDNS = c.UpstreamDNS
		dnsConf.BootstrapDNS = c.BootstrapDNS
		dnsConf.FallbackDNS = c.FallbackDNS
		dnsConf.Ratelimit = c.Ratelimit
		dnsConf.EnableDNSSEC = c.EnableDNSSEC
		dnsConf.AAAADisabled = c.AAAADisabled
		dnsConf.CacheEnabled = c.CacheEnabled
		dnsConf.CacheSize = c.CacheSize
		dnsConf.CacheMinTTL = c.CacheMinTTL
		dnsConf.CacheMaxTTL = c.CacheMaxTTL
		dnsConf.CacheOptimistic = c.CacheOptimistic
		dnsConf.UpstreamMode = string(c.UpstreamMode)

		dnsConf.RatelimitSubnetLenIPv4 = c.RatelimitSubnetLenIPv4
		dnsConf.RatelimitSubnetLenIPv6 = c.RatelimitSubnetLenIPv6
		dnsConf.RatelimitWhitelist = c.RatelimitWhitelist
		dnsConf.RefuseAny = c.RefuseAny
		dnsConf.FastestTimeout = c.FastestTimeout
		dnsConf.AllowedClients = c.AllowedClients
		dnsConf.DisallowedClients = c.DisallowedClients
		dnsConf.BlockedHosts = c.BlockedHosts
		dnsConf.TrustedProxies = c.TrustedProxies
		dnsConf.CacheOptimisticTTL = c.CacheOptimisticAnswerTTL
		dnsConf.CacheOptimisticMaxAge = c.CacheOptimisticMaxAge
		dnsConf.BogusNXDomain = c.BogusNXDomain
		dnsConf.EDNSClientSubnet = c.EDNSClientSubnet
		dnsConf.MaxGoroutines = c.MaxGoroutines
		dnsConf.HandleDDR = c.HandleDDR
		dnsConf.IpsetList = c.IpsetList
		dnsConf.BootstrapPreferIPv6 = c.BootstrapPreferIPv6
	}
	export.DNS = dnsConf

	// Filtering config.
	if globalContext.filters != nil {
		fltConf := config.Filtering
		expFlt := &exportFilteringConfig{
			Enabled:          fltConf.ProtectionEnabled,
			Interval:         fltConf.FiltersUpdateIntervalHours,
			Filters:          config.Filters,
			WhitelistFilters: config.WhitelistFilters,
			UserRules:        config.UserRules,

			FilteringEnabled:      fltConf.FilteringEnabled,
			BlockingMode:          string(fltConf.BlockingMode),
			BlockedServices:       fltConf.BlockedServices,
			ParentalBlockHost:     fltConf.ParentalBlockHost,
			SafeBrowsingBlockHost: fltConf.SafeBrowsingBlockHost,
			RewritesEnabled:       fltConf.RewritesEnabled,
			BlockedResponseTTL:    fltConf.BlockedResponseTTL,
			SafeBrowsingCacheSize: fltConf.SafeBrowsingCacheSize,
			SafeSearchCacheSize:   fltConf.SafeSearchCacheSize,
			ParentalCacheSize:     fltConf.ParentalCacheSize,
			CacheTime:             fltConf.CacheTime,
		}

		if !fltConf.BlockingIPv4.IsUnspecified() {
			v4 := fltConf.BlockingIPv4
			expFlt.BlockingIPv4 = &v4
		}
		if !fltConf.BlockingIPv6.IsUnspecified() {
			v6 := fltConf.BlockingIPv6
			expFlt.BlockingIPv6 = &v6
		}

		if fltConf.Rewrites != nil {
			rewrites := make([]*exportRewrite, 0, len(fltConf.Rewrites))
			for _, rw := range fltConf.Rewrites {
				rewrites = append(rewrites, &exportRewrite{
					Domain:  rw.Domain,
					Answer:  rw.Answer,
					Enabled: rw.Enabled,
				})
			}
			expFlt.Rewrites = rewrites
		}

		export.Filtering = expFlt

		if fltConf.SafeBrowsingEnabled {
			t := true
			export.SafeBrowsing = &t
		}
		if fltConf.ParentalEnabled {
			t := true
			export.Parental = &t
		}
		ssConf := fltConf.SafeSearchConf
		export.SafeSearch = &ssConf
	}

	// Query log config.
	qlConf := &exportQueryLogConfig{
		Enabled:           config.QueryLog.Enabled,
		Interval:          uint32(time.Duration(config.QueryLog.Interval).Hours()),
		Ignored:           config.QueryLog.Ignored,
		MemSize:           config.QueryLog.MemSize,
		IgnoredEnabled:    config.QueryLog.IgnoredEnabled,
		FileEnabled:       config.QueryLog.FileEnabled,
		AnonymizeClientIP: config.DNS.AnonymizeClientIP,
	}
	if globalContext.queryLog != nil {
		dc := querylog.Config{}
		globalContext.queryLog.WriteDiskConfig(&dc)
		qlConf.Enabled = dc.Enabled
		qlConf.Interval = uint32(dc.RotationIvl.Hours())
		qlConf.AnonymizeClientIP = dc.AnonymizeClientIP
		qlConf.MemSize = dc.MemSize
		qlConf.FileEnabled = dc.FileEnabled
	}
	export.QueryLog = qlConf

	// Statistics config.
	stConf := &exportStatsConfig{
		Enabled:        config.Stats.Enabled,
		Interval:       uint32(time.Duration(config.Stats.Interval).Hours()),
		Ignored:        config.Stats.Ignored,
		IgnoredEnabled: config.Stats.IgnoredEnabled,
	}
	if globalContext.stats != nil {
		sc := stats.Config{}
		globalContext.stats.WriteDiskConfig(&sc)
		stConf.Enabled = sc.Enabled
		stConf.Interval = uint32(sc.Limit.Hours())
	}
	export.Statistics = stConf

	// DHCP config.
	if config.DHCP != nil {
		export.DHCP = config.DHCP
	}

	// Clients config.
	clients := globalContext.clients.forConfig()
	export.Clients = &exportClientsConfig{
		Persistent: clients,
		Sources:    config.Clients.Sources,
	}

	// TLS config (sensitive fields cleared).
	tlsConf := config.TLS
	tlsConf.PrivateKey = ""
	export.TLS = &tlsConf

	// ACME ("SSL/TLS issue") config.
	if a := config.ACME; a != nil {
		export.ACME = &exportACMEConfig{
			Enabled:            a.Enabled,
			Email:              a.Email,
			Domains:            a.Domains,
			Challenge:          a.Challenge,
			CloudflareAPIToken: a.CloudflareAPIToken,
			AutoRenew:          a.AutoRenew,
			RenewBeforeDays:    a.RenewBeforeDays,
		}
	}

	// Notifications config.
	if tg := config.Notifications.Telegram; tg != nil {
		export.Notifications = &exportNotificationsConfig{
			Telegram: &exportTelegramConfig{
				Enabled:         tg.Enabled,
				BotToken:        tg.BotToken,
				ChatID:          tg.ChatID,
				CPUThreshold:    tg.CPUThreshold,
				MemoryThreshold: tg.MemoryThreshold,
				DiskThreshold:   tg.DiskThreshold,
				CheckInterval:   tg.CheckInterval,
				Cooldown:        tg.Cooldown,
				CustomMessage:   tg.CustomMessage,
			},
		}
	}

	// YouTube config.
	if yt := config.YouTube; yt != nil {
		export.YouTube = yt
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", `attachment; filename="adguardhome-settings.json"`)
	aghhttp.WriteJSONResponseOK(ctx, l, w, r, export)
}

// handleImportSettings handles POST /control/settings/import.
func (web *webAPI) handleImportSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	body, err := io.ReadAll(io.LimitReader(r.Body, 8<<20))
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "reading body: %s", err)

		return
	}

	var imp exportConfig
	err = json.Unmarshal(body, &imp)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, l, r, w, http.StatusBadRequest, "parsing json: %s", err)

		return
	}

	l.InfoContext(ctx, "importing settings")

	func() {
		config.Lock()
		defer config.Unlock()

		err = applyImportedSettings(l, &imp)
	}()

	if err != nil {
		aghhttp.ErrorAndLog(
			ctx, l, r, w, http.StatusInternalServerError,
			"applying settings: %s", err,
		)

		return
	}

	web.confModifier.Apply(ctx)

	if imp.YouTube != nil && ytManager != nil {
		ytManager.restart(ctx)
	}

	if imp.Notifications != nil && globalContext.notifier != nil {
		func() {
			config.RLock()
			defer config.RUnlock()

			globalContext.notifier.UpdateTelegramConfig(buildRuntimeTelegramConfig(config.Notifications.Telegram))
		}()
	}

	l.InfoContext(ctx, "settings imported successfully")

	aghhttp.WriteJSONResponseOK(ctx, l, w, r, struct {
		Status string `json:"status"`
	}{
		Status: "ok",
	})
}

// applyImportedSettings applies all imported settings to the running
// configuration.
func applyImportedSettings(_ *slog.Logger, imp *exportConfig) (err error) {
	if imp.General != nil {
		applyGeneralImport(imp.General)
	}

	if imp.DNS != nil {
		applyDNSImport(imp.DNS)
	}

	if imp.Filtering != nil {
		applyFilteringImport(imp.Filtering)
	}

	if imp.SafeBrowsing != nil && config.Filtering != nil {
		config.Filtering.SafeBrowsingEnabled = *imp.SafeBrowsing
	}

	if imp.Parental != nil && config.Filtering != nil {
		config.Filtering.ParentalEnabled = *imp.Parental
	}

	if imp.SafeSearch != nil && config.Filtering != nil {
		config.Filtering.SafeSearchConf = *imp.SafeSearch
	}

	if imp.QueryLog != nil {
		applyQueryLogImport(imp.QueryLog)
	}

	if imp.Statistics != nil {
		applyStatsImport(imp.Statistics)
	}

	if imp.DHCP != nil {
		config.DHCP = imp.DHCP
	}

	if imp.Clients != nil {
		applyClientsImport(imp.Clients)
	}

	if imp.TLS != nil {
		applyTLSImport(imp.TLS)
	}

	if imp.Notifications != nil {
		applyNotificationsImport(imp.Notifications)
	}

	if imp.YouTube != nil {
		config.YouTube = imp.YouTube
	}

	if imp.ACME != nil {
		applyACMEImport(imp.ACME)
	}

	return nil
}

// applyGeneralImport applies the imported general settings.
func applyGeneralImport(gen *exportGeneralConfig) {
	if gen.Language != "" {
		config.Language = gen.Language
	}
	if gen.Theme != "" {
		config.Theme = Theme(gen.Theme)
	}
}

// applyDNSImport applies the imported DNS settings.
func applyDNSImport(dns *exportDNSConfig) {
	if len(dns.BindHosts) > 0 {
		config.DNS.BindHosts = dns.BindHosts
	}
	if dns.Port > 0 {
		config.DNS.Port = dns.Port
	}

	config.DNS.HostsFileEnabled = dns.HostsFileEnabled
	config.DNS.ServePlainDNS = dns.ServePlainDNS
	config.DNS.UsePrivateRDNS = dns.UsePrivateRDNS
	config.DNS.PrivateRDNSResolvers = dns.LocalPTRUpstream
	config.DNS.UseDNS64 = dns.UseDNS64
	config.DNS.DNS64Prefixes = dns.DNS64Prefixes
	config.DNS.AnonymizeClientIP = dns.AnonymizeClientIP
	config.DNS.ServeHTTP3 = dns.ServeHTTP3
	config.DNS.UseHTTP3Upstreams = dns.UseHTTP3Upstreams

	if dns.UpstreamTimeout != 0 {
		config.DNS.UpstreamTimeout = dns.UpstreamTimeout
	}
	if dns.PrivateNets != nil {
		config.DNS.PrivateNets = dns.PrivateNets
	}
	if dns.PendingRequests != nil {
		if config.DNS.PendingRequests == nil {
			config.DNS.PendingRequests = &pendingRequests{}
		}
		config.DNS.PendingRequests.Enabled = *dns.PendingRequests
	}

	if s := globalContext.dnsServer; s != nil {
		c := dnsforward.Config{}
		s.WriteDiskConfig(&c)

		if dns.UpstreamDNS != nil {
			c.UpstreamDNS = dns.UpstreamDNS
		}
		if dns.BootstrapDNS != nil {
			c.BootstrapDNS = dns.BootstrapDNS
		}
		if dns.FallbackDNS != nil {
			c.FallbackDNS = dns.FallbackDNS
		}

		c.Ratelimit = dns.Ratelimit
		c.EnableDNSSEC = dns.EnableDNSSEC
		c.AAAADisabled = dns.AAAADisabled
		c.CacheEnabled = dns.CacheEnabled
		c.CacheSize = dns.CacheSize
		c.CacheMinTTL = dns.CacheMinTTL
		c.CacheMaxTTL = dns.CacheMaxTTL
		c.CacheOptimistic = dns.CacheOptimistic

		if dns.UpstreamMode != "" {
			c.UpstreamMode = dnsforward.UpstreamMode(dns.UpstreamMode)
		}

		c.RatelimitSubnetLenIPv4 = dns.RatelimitSubnetLenIPv4
		c.RatelimitSubnetLenIPv6 = dns.RatelimitSubnetLenIPv6
		if dns.RatelimitWhitelist != nil {
			c.RatelimitWhitelist = dns.RatelimitWhitelist
		}
		c.RefuseAny = dns.RefuseAny
		c.FastestTimeout = dns.FastestTimeout
		if dns.AllowedClients != nil {
			c.AllowedClients = dns.AllowedClients
		}
		if dns.DisallowedClients != nil {
			c.DisallowedClients = dns.DisallowedClients
		}
		if dns.BlockedHosts != nil {
			c.BlockedHosts = dns.BlockedHosts
		}
		if dns.TrustedProxies != nil {
			c.TrustedProxies = dns.TrustedProxies
		}
		c.CacheOptimisticAnswerTTL = dns.CacheOptimisticTTL
		c.CacheOptimisticMaxAge = dns.CacheOptimisticMaxAge
		if dns.BogusNXDomain != nil {
			c.BogusNXDomain = dns.BogusNXDomain
		}
		if dns.EDNSClientSubnet != nil {
			c.EDNSClientSubnet = dns.EDNSClientSubnet
		}
		c.MaxGoroutines = dns.MaxGoroutines
		c.HandleDDR = dns.HandleDDR
		if dns.IpsetList != nil {
			c.IpsetList = dns.IpsetList
		}
		c.BootstrapPreferIPv6 = dns.BootstrapPreferIPv6

		config.DNS.Config = c
	}
}

// applyFilteringImport applies the imported filtering settings.
func applyFilteringImport(flt *exportFilteringConfig) {
	if config.Filtering == nil {
		return
	}

	config.Filtering.ProtectionEnabled = flt.Enabled
	config.Filtering.FiltersUpdateIntervalHours = flt.Interval
	config.Filtering.FilteringEnabled = flt.FilteringEnabled
	config.Filtering.RewritesEnabled = flt.RewritesEnabled

	if flt.BlockingMode != "" {
		config.Filtering.BlockingMode = filtering.BlockingMode(flt.BlockingMode)
	}
	if flt.BlockingIPv4 != nil {
		config.Filtering.BlockingIPv4 = *flt.BlockingIPv4
	}
	if flt.BlockingIPv6 != nil {
		config.Filtering.BlockingIPv6 = *flt.BlockingIPv6
	}
	if flt.BlockedServices != nil {
		config.Filtering.BlockedServices = flt.BlockedServices
	}
	if flt.ParentalBlockHost != "" {
		config.Filtering.ParentalBlockHost = flt.ParentalBlockHost
	}
	if flt.SafeBrowsingBlockHost != "" {
		config.Filtering.SafeBrowsingBlockHost = flt.SafeBrowsingBlockHost
	}
	if flt.BlockedResponseTTL > 0 {
		config.Filtering.BlockedResponseTTL = flt.BlockedResponseTTL
	}
	if flt.SafeBrowsingCacheSize > 0 {
		config.Filtering.SafeBrowsingCacheSize = flt.SafeBrowsingCacheSize
	}
	if flt.SafeSearchCacheSize > 0 {
		config.Filtering.SafeSearchCacheSize = flt.SafeSearchCacheSize
	}
	if flt.ParentalCacheSize > 0 {
		config.Filtering.ParentalCacheSize = flt.ParentalCacheSize
	}
	if flt.CacheTime > 0 {
		config.Filtering.CacheTime = flt.CacheTime
	}

	if flt.Rewrites != nil {
		rewrites := make([]*filtering.LegacyRewrite, 0, len(flt.Rewrites))
		for _, rw := range flt.Rewrites {
			rewrites = append(rewrites, &filtering.LegacyRewrite{
				Domain:  rw.Domain,
				Answer:  rw.Answer,
				Enabled: rw.Enabled,
			})
		}
		config.Filtering.Rewrites = rewrites
	}

	if flt.Filters != nil {
		config.Filters = flt.Filters
		config.Filtering.Filters = flt.Filters
	}
	if flt.WhitelistFilters != nil {
		config.WhitelistFilters = flt.WhitelistFilters
		config.Filtering.WhitelistFilters = flt.WhitelistFilters
	}
	if flt.UserRules != nil {
		config.UserRules = flt.UserRules
		config.Filtering.UserRules = flt.UserRules
	}
}

// applyQueryLogImport applies the imported query log settings.
func applyQueryLogImport(ql *exportQueryLogConfig) {
	config.QueryLog.Enabled = ql.Enabled
	config.QueryLog.Interval = timeutil.Duration(time.Duration(ql.Interval) * time.Hour)
	config.DNS.AnonymizeClientIP = ql.AnonymizeClientIP
	config.QueryLog.FileEnabled = ql.FileEnabled
	config.QueryLog.IgnoredEnabled = ql.IgnoredEnabled

	if ql.MemSize > 0 {
		config.QueryLog.MemSize = ql.MemSize
	}
	if ql.Ignored != nil {
		config.QueryLog.Ignored = ql.Ignored
	}
}

// applyStatsImport applies the imported statistics settings.
func applyStatsImport(st *exportStatsConfig) {
	config.Stats.Enabled = st.Enabled
	config.Stats.Interval = timeutil.Duration(time.Duration(st.Interval) * time.Hour)
	config.Stats.IgnoredEnabled = st.IgnoredEnabled

	if st.Ignored != nil {
		config.Stats.Ignored = st.Ignored
	}
}

// applyClientsImport applies the imported client settings.
func applyClientsImport(cl *exportClientsConfig) {
	if cl.Persistent != nil {
		config.Clients.Persistent = cl.Persistent
	}
	if cl.Sources != nil {
		config.Clients.Sources = cl.Sources
	}
}

// applyTLSImport applies the imported TLS settings, preserving the existing
// private key if not provided in the import.
func applyTLSImport(tls *tlsConfigSettings) {
	existingKey := config.TLS.PrivateKey
	config.TLS = *tls
	if tls.PrivateKey == "" {
		config.TLS.PrivateKey = existingKey
	}
}

// applyNotificationsImport applies the imported notification settings.
func applyNotificationsImport(notif *exportNotificationsConfig) {
	if notif.Telegram == nil {
		return
	}

	tg := notif.Telegram
	if config.Notifications.Telegram == nil {
		config.Notifications.Telegram = defaultTelegramConfig()
	}

	config.Notifications.Telegram.Enabled = tg.Enabled
	if tg.BotToken != "" {
		config.Notifications.Telegram.BotToken = tg.BotToken
	}
	if tg.ChatID != "" {
		config.Notifications.Telegram.ChatID = tg.ChatID
	}
	if tg.CPUThreshold > 0 {
		config.Notifications.Telegram.CPUThreshold = tg.CPUThreshold
	}
	if tg.MemoryThreshold > 0 {
		config.Notifications.Telegram.MemoryThreshold = tg.MemoryThreshold
	}
	if tg.DiskThreshold > 0 {
		config.Notifications.Telegram.DiskThreshold = tg.DiskThreshold
	}
	if tg.CheckInterval != 0 {
		config.Notifications.Telegram.CheckInterval = tg.CheckInterval
	}
	if tg.Cooldown != 0 {
		config.Notifications.Telegram.Cooldown = tg.Cooldown
	}
	config.Notifications.Telegram.CustomMessage = tg.CustomMessage
}

// applyACMEImport applies the imported ACME ("SSL/TLS issue") settings.
func applyACMEImport(a *exportACMEConfig) {
	if config.ACME == nil {
		config.ACME = defaultACMEConfig()
	}

	config.ACME.Enabled = a.Enabled
	config.ACME.AutoRenew = a.AutoRenew

	if a.Email != "" {
		config.ACME.Email = a.Email
	}
	if len(a.Domains) > 0 {
		config.ACME.Domains = a.Domains
	}
	if a.Challenge != "" {
		config.ACME.Challenge = a.Challenge
	}
	if a.CloudflareAPIToken != "" {
		config.ACME.CloudflareAPIToken = a.CloudflareAPIToken
	}
	if a.RenewBeforeDays > 0 {
		config.ACME.RenewBeforeDays = a.RenewBeforeDays
	}
}
