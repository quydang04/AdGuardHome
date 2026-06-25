package home

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/netip"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/dhcpd"
	"github.com/AdguardTeam/AdGuardHome/internal/dnsforward"
	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/AdGuardHome/internal/querylog"
	"github.com/AdguardTeam/AdGuardHome/internal/stats"
	"github.com/AdguardTeam/golibs/timeutil"
)

// exportConfig is the JSON structure used for exporting and importing all
// AdGuard Home settings.
type exportConfig struct {
	DNS          *exportDNSConfig       `json:"dns"`
	DHCP         *dhcpd.ServerConfig    `json:"dhcp,omitempty"`
	Filtering    *exportFilteringConfig `json:"filtering,omitempty"`
	QueryLog     *exportQueryLogConfig  `json:"querylog,omitempty"`
	Statistics   *exportStatsConfig     `json:"statistics,omitempty"`
	Clients      *exportClientsConfig   `json:"clients,omitempty"`
	TLS          *tlsConfigSettings     `json:"tls,omitempty"`
	SafeBrowsing *bool                  `json:"safebrowsing_enabled,omitempty"`
	Parental     *bool                  `json:"parental_enabled,omitempty"`
	SafeSearch   *filtering.SafeSearchConfig `json:"safesearch,omitempty"`
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
}

// exportFilteringConfig is the filtering portion of the export.
type exportFilteringConfig struct {
	Enabled          bool                   `json:"enabled"`
	Interval         uint32                 `json:"interval,omitempty"`
	Filters          []filtering.FilterYAML `json:"filters,omitempty"`
	WhitelistFilters []filtering.FilterYAML `json:"whitelist_filters,omitempty"`
	UserRules        []string               `json:"user_rules,omitempty"`
	BlockedServices  []string               `json:"blocked_services,omitempty"`
}

// exportQueryLogConfig is the query log portion of the export.
type exportQueryLogConfig struct {
	Enabled            bool   `json:"enabled"`
	Interval           uint32 `json:"interval,omitempty"`
	AnonymizeClientIP  bool   `json:"anonymize_client_ip,omitempty"`
}

// exportStatsConfig is the statistics portion of the export.
type exportStatsConfig struct {
	Enabled  bool   `json:"enabled"`
	Interval uint32 `json:"interval,omitempty"`
}

// exportClientsConfig is the clients portion of the export.
type exportClientsConfig struct {
	Persistent []*clientObject `json:"persistent,omitempty"`
}

// handleExportSettings handles GET /control/settings/export.
func (web *webAPI) handleExportSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	l := web.logger

	export := &exportConfig{}

	config.RLock()
	defer config.RUnlock()

	// DNS config.
	dnsConf := &exportDNSConfig{
		BindHosts:        config.DNS.BindHosts,
		Port:             config.DNS.Port,
		HostsFileEnabled: config.DNS.HostsFileEnabled,
		ServePlainDNS:    config.DNS.ServePlainDNS,
		UsePrivateRDNS:   config.DNS.UsePrivateRDNS,
		LocalPTRUpstream: config.DNS.PrivateRDNSResolvers,
		UseDNS64:         config.DNS.UseDNS64,
		DNS64Prefixes:    config.DNS.DNS64Prefixes,
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
	}
	export.DNS = dnsConf

	// Filtering config.
	if globalContext.filters != nil {
		fltConf := config.Filtering
		export.Filtering = &exportFilteringConfig{
			Enabled:          fltConf.ProtectionEnabled,
			Interval:         fltConf.FiltersUpdateIntervalHours,
			Filters:          config.Filters,
			WhitelistFilters: config.WhitelistFilters,
			UserRules:        config.UserRules,
		}

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
	if globalContext.queryLog != nil {
		dc := querylog.Config{}
		globalContext.queryLog.WriteDiskConfig(&dc)
		export.QueryLog = &exportQueryLogConfig{
			Enabled:           dc.Enabled,
			Interval:          uint32(dc.RotationIvl.Hours()),
			AnonymizeClientIP: dc.AnonymizeClientIP,
		}
	}

	// Statistics config.
	if globalContext.stats != nil {
		sc := stats.Config{}
		globalContext.stats.WriteDiskConfig(&sc)
		export.Statistics = &exportStatsConfig{
			Enabled:  sc.Enabled,
			Interval: uint32(sc.Limit.Hours()),
		}
	}

	// DHCP config.
	if config.DHCP != nil {
		export.DHCP = config.DHCP
	}

	// Clients config.
	clients := globalContext.clients.forConfig()
	if len(clients) > 0 {
		export.Clients = &exportClientsConfig{
			Persistent: clients,
		}
	}

	// TLS config (sensitive fields cleared).
	tlsConf := config.TLS
	tlsConf.PrivateKey = ""
	export.TLS = &tlsConf

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

	return nil
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
}

// applyStatsImport applies the imported statistics settings.
func applyStatsImport(st *exportStatsConfig) {
	config.Stats.Enabled = st.Enabled
	config.Stats.Interval = timeutil.Duration(time.Duration(st.Interval) * time.Hour)
}

// applyClientsImport applies the imported client settings.
func applyClientsImport(cl *exportClientsConfig) {
	if cl.Persistent != nil {
		config.Clients.Persistent = cl.Persistent
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
