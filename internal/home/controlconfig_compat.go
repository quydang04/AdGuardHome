package home

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghalg"
	"github.com/AdguardTeam/AdGuardHome/internal/querylog"
)

// exportFormatVersion is the current version of the settings export format.
//
// Version history:
//   - 0 (no "export_version" key): legacy files.  Missing keys keep their
//     current value, and loosely typed values (strings instead of lists,
//     "30d" instead of hours, etc.) are normalized.
//   - 2: current format.  The file is a full snapshot, so missing keys mean
//     "off" or "empty", exactly as written by the exporter.
const exportFormatVersion = 2

// importSection describes one top-level section of the import file.
type importSection struct {
	// key is the JSON key of the section.
	key string

	// dst is a pointer to the corresponding field of [exportConfig].
	dst any

	// clear resets the field, so that the section is skipped.
	clear func()
}

// legacySectionAliases maps the alternative top-level keys found in older
// files to the current ones.
var legacySectionAliases = map[string]string{
	"query_log":    "querylog",
	"stats":        "statistics",
	"safe_search":  "safesearch",
	"safebrowsing": "safebrowsing_enabled",
	"safe_browsing": "safebrowsing_enabled",
	"parental":     "parental_enabled",
}

// decodeImport parses and normalizes an import file.  Sections that cannot be
// parsed are skipped and reported in warnings, so that one broken section
// doesn't prevent the rest of the file from being imported.
func decodeImport(body []byte) (imp *exportConfig, warnings []string, err error) {
	var root map[string]any
	err = json.Unmarshal(body, &root)
	if err != nil {
		return nil, nil, err
	}

	if root == nil {
		return nil, nil, fmt.Errorf("file is empty")
	}

	for from, to := range legacySectionAliases {
		if v, ok := root[from]; ok {
			if _, exists := root[to]; !exists {
				root[to] = v
			}
		}
	}

	version := 0
	if v, ok := root["export_version"].(float64); ok {
		version = int(v)
	}

	if version > exportFormatVersion {
		warnings = append(warnings, fmt.Sprintf(
			"file was exported by a newer version (format %d), some settings may be ignored",
			version,
		))
	}

	normalizeImport(root)

	if version >= 2 {
		// A full snapshot: absent keys are meaningful.
		imp = &exportConfig{}
	} else {
		// A legacy file: start from the current settings, so that keys that
		// didn't exist in the old version keep their value instead of being
		// reset to zero.
		imp, err = cloneCurrentExport()
		if err != nil {
			return nil, nil, fmt.Errorf("preparing defaults: %w", err)
		}
	}

	sections := []importSection{
		{"dns", &imp.DNS, func() { imp.DNS = nil }},
		{"dhcp", &imp.DHCP, func() { imp.DHCP = nil }},
		{"filtering", &imp.Filtering, func() { imp.Filtering = nil }},
		{"querylog", &imp.QueryLog, func() { imp.QueryLog = nil }},
		{"statistics", &imp.Statistics, func() { imp.Statistics = nil }},
		{"clients", &imp.Clients, func() { imp.Clients = nil }},
		{"tls", &imp.TLS, func() { imp.TLS = nil }},
		{"safebrowsing_enabled", &imp.SafeBrowsing, func() { imp.SafeBrowsing = nil }},
		{"parental_enabled", &imp.Parental, func() { imp.Parental = nil }},
		{"safesearch", &imp.SafeSearch, func() { imp.SafeSearch = nil }},
		{"notifications", &imp.Notifications, func() { imp.Notifications = nil }},
		{"general", &imp.General, func() { imp.General = nil }},
		{"acme", &imp.ACME, func() { imp.ACME = nil }},
	}

	for _, s := range sections {
		val, ok := root[s.key]
		if !ok || val == nil {
			s.clear()

			continue
		}

		raw, mErr := json.Marshal(val)
		if mErr == nil {
			mErr = json.Unmarshal(raw, s.dst)
		}

		if mErr != nil {
			warnings = append(warnings, fmt.Sprintf("section %q skipped: %s", s.key, mErr))
			s.clear()
		}
	}

	return imp, warnings, nil
}

// cloneCurrentExport returns a deep copy of the current settings.
func cloneCurrentExport() (cur *exportConfig, err error) {
	raw, err := json.Marshal(buildExportConfig())
	if err != nil {
		return nil, err
	}

	cur = &exportConfig{}
	err = json.Unmarshal(raw, cur)

	return cur, err
}

// normalizeImport converts loosely typed values of legacy files to the types
// expected by [exportConfig].  It modifies root in place.
func normalizeImport(root map[string]any) {
	if dns, ok := root["dns"].(map[string]any); ok {
		for _, k := range []string{
			"upstream_dns", "bootstrap_dns", "fallback_dns", "local_ptr_upstreams",
		} {
			dns[k] = toStringList(dns[k], "\n")
		}

		for _, k := range []string{
			"allowed_clients", "disallowed_clients", "blocked_hosts", "bogus_nxdomain",
			"ipset", "ratelimit_whitelist", "dns64_prefixes",
		} {
			dns[k] = toStringList(dns[k], "\n,; ")
		}

		for _, k := range []string{
			"upstream_timeout", "cache_optimistic_answer_ttl", "cache_optimistic_max_age",
		} {
			dns[k] = toDuration(dns[k], time.Second)
		}

		// The legacy API used milliseconds for this one.
		dns["fastest_timeout"] = toDuration(dns["fastest_timeout"], time.Millisecond)
		dns["port"] = toNumber(dns["port"])
		dropNil(dns)
	}

	for _, sec := range []string{"querylog", "statistics"} {
		if m, ok := root[sec].(map[string]any); ok {
			m["interval"] = toHours(m["interval"])
			m["ignored"] = toStringList(m["ignored"], "\n,; ")
			dropNil(m)
		}
	}

	if flt, ok := root["filtering"].(map[string]any); ok {
		flt["user_rules"] = toStringList(flt["user_rules"], "\n")
		flt["interval"] = toNumber(flt["interval"])
		dropNil(flt)
	}

	if acme, ok := root["acme"].(map[string]any); ok {
		acme["domains"] = toStringList(acme["domains"], "\n,; ")
		acme["dns_resolvers"] = toStringList(acme["dns_resolvers"], "\n,; ")
		dropNil(acme)
	}

	if n, ok := root["notifications"].(map[string]any); ok {
		if tg, tgOK := n["telegram"].(map[string]any); tgOK {
			for _, k := range []string{"check_interval", "cooldown"} {
				tg[k] = toDuration(tg[k], time.Second)
			}

			dropNil(tg)
		}
	}
}

// dropNil removes the keys with nil values from m.
func dropNil(m map[string]any) {
	for k, v := range m {
		if v == nil {
			delete(m, k)
		}
	}
}

// toStringList converts a string with items separated by any of seps into a
// list.  Other values are returned as is.
func toStringList(v any, seps string) (res any) {
	s, ok := v.(string)
	if !ok {
		return v
	}

	items := strings.FieldsFunc(s, func(r rune) bool { return strings.ContainsRune(seps, r) })
	list := make([]any, 0, len(items))
	for _, it := range items {
		it = strings.TrimSpace(it)
		if it != "" {
			list = append(list, it)
		}
	}

	return list
}

// toNumber converts a numeric string into a number.
func toNumber(v any) (res any) {
	s, ok := v.(string)
	if !ok {
		return v
	}

	f, err := strconv.ParseFloat(strings.TrimSpace(s), 64)
	if err != nil {
		return v
	}

	return f
}

// toDuration converts a bare number (or numeric string) into a duration
// string using unit.  Strings like "10s" are kept.
func toDuration(v any, unit time.Duration) (res any) {
	switch v := v.(type) {
	case float64:
		return time.Duration(v * float64(unit)).String()
	case string:
		f, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		if err != nil {
			return v
		}

		return time.Duration(f * float64(unit)).String()
	default:
		return v
	}
}

// toHours converts an interval into hours.  Numbers are considered to be
// hours already; strings may contain a unit, such as "720h", "30d" or "4w".
func toHours(v any) (res any) {
	s, ok := v.(string)
	if !ok {
		return v
	}

	s = strings.ToLower(strings.TrimSpace(s))
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	for suffix, mult := range map[string]float64{"w": 24 * 7, "d": 24} {
		if n, found := strings.CutSuffix(s, suffix); found {
			if f, err := strconv.ParseFloat(n, 64); err == nil {
				return f * mult
			}
		}
	}

	d, err := time.ParseDuration(s)
	if err != nil {
		return v
	}

	return d.Hours()
}

// applyImportedRuntime pushes the imported settings into the running
// services, since config.write reads the live state of those back into config.
// It returns warnings for the parts that couldn't be applied.
func (web *webAPI) applyImportedRuntime(ctx context.Context, imp *exportConfig) (warnings []string) {
	if imp.TLS != nil && web.tlsManager != nil {
		if w := web.applyImportedTLS(ctx, imp.TLS); w != "" {
			warnings = append(warnings, w)
		}
	}

	if imp.QueryLog != nil && globalContext.queryLog != nil {
		config.RLock()
		qc := config.QueryLog
		ignored := append([]string(nil), qc.Ignored...)
		anonymize := config.DNS.AnonymizeClientIP
		config.RUnlock()

		err := globalContext.queryLog.UpdateConfig(querylog.ImportedConfig{
			Ignored:           ignored,
			RotationIvl:       time.Duration(qc.Interval),
			MemSize:           qc.MemSize,
			Enabled:           qc.Enabled,
			FileEnabled:       qc.FileEnabled,
			AnonymizeClientIP: anonymize,
			IgnoredEnabled:    qc.IgnoredEnabled,
		})
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("query log settings not applied: %s", err))
		}
	}

	if imp.Statistics != nil && globalContext.stats != nil {
		config.RLock()
		sc := config.Stats
		ignored := append([]string(nil), sc.Ignored...)
		config.RUnlock()

		err := globalContext.stats.UpdateConfig(
			sc.Enabled,
			time.Duration(sc.Interval),
			ignored,
			sc.IgnoredEnabled,
		)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("statistics settings not applied: %s", err))
		}
	}

	filteringChanged := imp.Filtering != nil || imp.SafeBrowsing != nil ||
		imp.Parental != nil || imp.SafeSearch != nil
	if filteringChanged && globalContext.filters != nil {
		config.RLock()
		var fc = *config.Filtering
		config.RUnlock()

		globalContext.filters.ApplyImported(ctx, &fc)
	}

	if imp.Clients != nil && len(imp.Clients.Persistent) > 0 {
		warnings = append(warnings, applyImportedClients(ctx, imp.Clients.Persistent)...)
	}

	if imp.DNS != nil && globalContext.dnsServer != nil && web.tlsManager != nil {
		config.RLock()
		err := web.tlsManager.reconfigureDNSServer(ctx)
		config.RUnlock()
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("DNS server not reconfigured: %s", err))
		}
	}

	return warnings
}

// applyImportedTLS validates and applies the imported TLS settings.  If the
// private key is missing, which is always the case for exported files, the
// current key is kept.  It returns a non-empty warning if the settings
// couldn't be applied; the current TLS settings are kept in that case.
func (web *webAPI) applyImportedTLS(ctx context.Context, imp *tlsConfigSettings) (warning string) {
	m := web.tlsManager

	conf := *imp
	if conf.PrivateKey == "" && conf.PrivateKeyPath == "" {
		conf.PrivateKey = m.extendedTLSConfig().PrivateKey
	}

	req := tlsConfigSettingsExt{
		tlsConfigSettings: conf,
		ServePlainDNS:     aghalg.NBNull,
	}
	status := &tlsConfigStatus{}

	var restart bool
	var err error
	func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		err = m.validateTLSSettings(req)
		if err != nil {
			return
		}

		err = m.loadTLSConfig(ctx, &req.tlsConfigSettings, status)
		if err != nil {
			return
		}

		restart, err = m.setConfig(ctx, req.tlsConfigSettings, status, aghalg.NBNull)
		if err == nil {
			m.setCertFileTime(ctx)
		}
	}()
	if err != nil {
		return fmt.Sprintf("TLS settings not applied (the private key is not part of exports): %s", err)
	}

	if restart {
		go m.web.tlsConfigChanged(context.Background(), &req.tlsConfigSettings)
	}

	return ""
}

// applyImportedClients adds the imported persistent clients to the storage,
// updating the ones with the same name.  Existing clients that are not in the
// file are kept.
func applyImportedClients(ctx context.Context, objs []*clientObject) (warnings []string) {
	cc := &globalContext.clients
	if cc.storage == nil {
		return nil
	}

	cc.lock.Lock()
	defer cc.lock.Unlock()

	for _, o := range objs {
		p, err := o.toPersistent(ctx, cc.baseLogger, cc.safeSearchCacheSize, cc.safeSearchCacheTTL)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("client %q skipped: %s", o.Name, err))

			continue
		}

		err = cc.storage.Add(ctx, p)
		if err == nil {
			continue
		}

		err = cc.storage.Update(ctx, o.Name, p)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("client %q skipped: %s", o.Name, err))
		}
	}

	return warnings
}
