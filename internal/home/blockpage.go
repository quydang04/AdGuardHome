package home

import (
	"bytes"
	"html"
	"net"
	"net/http"
	"net/netip"
	"net/url"
	"slices"
	"strings"

	"github.com/AdguardTeam/AdGuardHome/internal/filtering"
	"github.com/AdguardTeam/golibs/httphdr"
	"github.com/AdguardTeam/golibs/netutil"
	"github.com/go-acme/lego/v4/challenge/http01"
)

// transparentGIF1x1 is a minimal 43-byte 1x1 transparent GIF returned for
// blocked subresource <img> requests so webpages do not show broken-image icons.
var transparentGIF1x1 = []byte{
	0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00,
	0x01, 0x00, 0x80, 0x00, 0x00, 0x00, 0x00, 0x00,
	0xff, 0xff, 0xff, 0x21, 0xf9, 0x04, 0x01, 0x00,
	0x00, 0x00, 0x00, 0x2c, 0x00, 0x00, 0x00, 0x00,
	0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44,
	0x01, 0x00, 0x3b,
}

// adgBrandShieldSVG matches ShieldLogoIcon in Login/index.tsx and Sidebar/index.tsx.
const adgBrandShieldSVG = `<svg width="22" height="22" viewBox="0 0 24 24" fill="#67b279" xmlns="http://www.w3.org/2000/svg"><path d="M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z"/></svg>`

// blockPageMiddleware intercepts HTTP/HTTPS requests whose Host header matches
// a domain filtered by AdGuard Home when BlockingModeBlockPage is active,
// rendering an AdGuard Home block page without requiring authentication.
func (web *webAPI) blockPageMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := extractRequestHostname(r.Host)
		if isAdminOrInternalRequest(r, host) {
			next.ServeHTTP(w, r)

			return
		}

		flt := globalContext.filters
		dnsSrv := globalContext.dnsServer
		if flt == nil || dnsSrv == nil {
			next.ServeHTTP(w, r)

			return
		}

		mode, _, _ := flt.BlockingMode()
		if mode != filtering.BlockingModeBlockPage {
			next.ServeHTTP(w, r)

			return
		}

		clientIP := extractRemoteAddr(r.RemoteAddr)
		res, blocked := dnsSrv.LookupBlockedHost(host, clientIP)
		if !blocked {
			next.ServeHTTP(w, r)

			return
		}

		web.serveBlockedHTTPResponse(w, r, host, &res)
	})
}

func (web *webAPI) serveBlockPagePreview(w http.ResponseWriter, r *http.Request) {
	scenario := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("scenario")))
	domain := extractPreviewDomain(r.URL.Query().Get("domain"))
	isVi := isVietnameseBlockPage(r)

	flt := globalContext.filters
	var info filtering.BlockPageInfo

	switch {
	case scenario != "":
		if domain == "" {
			domain = previewDomainForScenario(scenario)
		}
		info = flt.DescribePreviewScenario(scenario, domain, isVi)

	case domain != "":
		if flt != nil && globalContext.dnsServer != nil {
			res, blocked := globalContext.dnsServer.LookupBlockedHost(domain, extractRemoteAddr(r.RemoteAddr))
			if blocked {
				info = flt.DescribeBlock(domain, &res, isVi)
			} else {
				info = flt.DescribeAllowed(domain, &res, isVi)
			}
		} else {
			info = flt.DescribeAllowed(domain, nil, isVi)
		}

	default:
		domain = previewDomainForScenario("ads")
		info = flt.DescribePreviewScenario("ads", domain, isVi)
	}

	writeBlockPageHTML(w, r, domain, info, isVi, currentBlockPageTheme(r))
}

func previewDomainForScenario(scenario string) string {
	switch scenario {
	case "malware":
		return "phishing.example.com"
	case "service":
		return "tiktok.com"
	case "parental":
		return "restricted.example.com"
	case "custom":
		return "blocked.example.com"
	case "tracker":
		return "analytics.example.com"
	default:
		return "ads.example.com"
	}
}

func extractPreviewDomain(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	if strings.Contains(s, "://") {
		if u, err := url.Parse(s); err == nil && u.Host != "" {
			s = u.Host
		}
	} else if slashIdx := strings.IndexByte(s, '/'); slashIdx >= 0 {
		s = s[:slashIdx]
	}

	return extractRequestHostname(s)
}

func (web *webAPI) serveBlockedHTTPResponse(
	w http.ResponseWriter,
	r *http.Request,
	host string,
	res *filtering.Result,
) {
	hdr := w.Header()
	hdr.Set(httphdr.CacheControl, "no-store, no-cache, must-revalidate")
	hdr.Set("X-Robots-Tag", "noindex, nofollow")

	if r.URL.Path == "/favicon.ico" {
		hdr.Set(httphdr.ContentType, "image/svg+xml")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(adgBrandShieldSVG))

		return
	}

	dest := strings.ToLower(r.Header.Get("Sec-Fetch-Dest"))
	accept := strings.ToLower(r.Header.Get(httphdr.Accept))

	if dest == "image" || (dest == "" && strings.HasPrefix(accept, "image/")) {
		hdr.Set(httphdr.ContentType, "image/gif")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(transparentGIF1x1)

		return
	}

	if !isDocumentNavigation(r, dest, accept) {
		hdr.Set(httphdr.AccessControlAllowOrigin, "*")
		w.WriteHeader(http.StatusNoContent)

		return
	}

	isVi := isVietnameseBlockPage(r)
	info := globalContext.filters.DescribeBlock(host, res, isVi)

	writeBlockPageHTML(w, r, host, info, isVi, currentBlockPageTheme(r))
}

func isDocumentNavigation(r *http.Request, secFetchDest, accept string) bool {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		return false
	}

	if strings.EqualFold(r.Header.Get("X-Requested-With"), "XMLHttpRequest") {
		return false
	}

	switch secFetchDest {
	case "document":
		return true
	case "iframe", "frame", "script", "style", "empty", "font", "audio", "video",
		"worker", "sharedworker", "serviceworker", "manifest", "object", "embed", "report":
		return false
	}

	secFetchMode := strings.ToLower(r.Header.Get("Sec-Fetch-Mode"))
	if secFetchMode == "cors" || secFetchMode == "no-cors" || secFetchMode == "websocket" {
		return false
	}

	if accept == "" || strings.Contains(accept, "text/html") || strings.Contains(accept, "application/xhtml+xml") {
		return true
	}

	return r.URL.Path == "/" && strings.Contains(accept, "*/*")
}

func extractRequestHostname(hostport string) string {
	hostport = strings.TrimSpace(hostport)
	if hostport == "" {
		return ""
	}

	h, _, err := net.SplitHostPort(hostport)
	if err != nil {
		h = strings.Trim(hostport, "[]")
	}

	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(h)), ".")
}

func extractRemoteAddr(remoteAddr string) netip.Addr {
	ipStr, err := netutil.SplitHost(remoteAddr)
	if err != nil {
		ipStr = remoteAddr
	}

	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		return netip.Addr{}
	}

	return ip.Unmap()
}

func isAdminOrInternalRequest(r *http.Request, host string) bool {
	if host == "" || host == "localhost" {
		return true
	}

	if _, err := netip.ParseAddr(host); err == nil {
		return true
	}

	p := r.URL.Path
	if strings.HasPrefix(p, http01.PathPrefix) ||
		strings.HasPrefix(p, "/dns-query") ||
		strings.HasPrefix(p, "/control/") {
		return true
	}

	config.RLock()
	srvName := config.TLS.ServerName
	var acmeDomains []string
	if config.ACME != nil {
		acmeDomains = config.ACME.Domains
	}
	ownServer := (srvName != "" && strings.EqualFold(host, srvName)) ||
		slices.ContainsFunc(acmeDomains, func(d string) bool {
			return strings.EqualFold(host, d)
		})
	config.RUnlock()

	return ownServer
}

func isVietnameseBlockPage(r *http.Request) bool {
	if qLang := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("lang"))); qLang != "" {
		return strings.HasPrefix(qLang, "vi")
	}

	config.RLock()
	cfgLang := strings.ToLower(strings.TrimSpace(config.Language))
	config.RUnlock()

	if cfgLang == "vi" {
		return true
	}
	if cfgLang != "" && cfgLang != "auto" {
		return false
	}

	return strings.Contains(strings.ToLower(r.Header.Get("Accept-Language")), "vi")
}

func currentBlockPageTheme(r *http.Request) string {
	if qTheme := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("theme"))); qTheme == "light" || qTheme == "dark" || qTheme == "auto" {
		return qTheme
	}

	config.RLock()
	th := strings.ToLower(strings.TrimSpace(string(config.Theme)))
	config.RUnlock()

	if th == "light" || th == "dark" {
		return th
	}

	return "auto"
}

func buildLangSwitchHref(r *http.Request, targetLang string) string {
	q := r.URL.Query()
	q.Set("lang", targetLang)
	p := r.URL.Path
	if p == "" {
		p = "/"
	}

	return p + "?" + q.Encode()
}

func panelCategoryClass(info filtering.BlockPageInfo) string {
	if info.IsAllowed {
		return ""
	}
	switch info.Category {
	case filtering.BlockCategoryMalware:
		return " cat-malware"
	case filtering.BlockCategoryAdult:
		return " cat-adult"
	case filtering.BlockCategoryService:
		return " cat-service"
	default:
		return ""
	}
}

func writeSpecRow(buf *bytes.Buffer, label, value string, mono bool) {
	if value == "" {
		return
	}
	buf.WriteString(`<div class="spec-row"><span class="spec-k">`)
	buf.WriteString(html.EscapeString(label))
	if mono {
		buf.WriteString(`</span><span class="spec-v spec-v--mono">`)
	} else {
		buf.WriteString(`</span><span class="spec-v">`)
	}
	buf.WriteString(html.EscapeString(value))
	buf.WriteString(`</span></div>`)
}

func writeBlockPageHTML(
	w http.ResponseWriter,
	r *http.Request,
	domain string,
	info filtering.BlockPageInfo,
	isVi bool,
	themeMode string,
) {
	var (
		langAttr      = "en"
		langBtnCode   = "EN"
		langTitle     = "Chuyển sang Tiếng Việt"
		nextLangHref  = html.EscapeString(buildLangSwitchHref(r, "vi"))
		lblDomain     = "Website"
		lblTech       = "Rule details"
		lblAllLists   = "All matched lists"
		btnBack       = "Go back"
		btnReload     = "Try again"
		themeAriaText = "Toggle theme"
	)
	if isVi {
		langAttr = "vi"
		langBtnCode = "VI"
		langTitle = "Switch to English"
		nextLangHref = html.EscapeString(buildLangSwitchHref(r, "en"))
		lblDomain = "Tên miền"
		lblTech = "Xem chi tiết quy tắc"
		lblAllLists = "Tất cả danh sách khớp"
		btnBack = "Quay lại"
		btnReload = "Thử tải lại"
		themeAriaText = "Đổi giao diện sáng/tối"
	}

	catClass := panelCategoryClass(info)
	hasTechDetails := info.RuleValue != "" || info.ScopeValue != "" || len(info.AllSources) > 1

	var buf bytes.Buffer
	buf.Grow(6800)

	buf.WriteString(`<!DOCTYPE html><html lang="`)
	buf.WriteString(langAttr)
	buf.WriteString(`" data-theme="`)
	buf.WriteString(themeMode)
	buf.WriteString(`"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1">`)
	buf.WriteString(`<title>`)
	buf.WriteString(html.EscapeString(info.PageTitle))
	buf.WriteString(`</title>`)
	buf.WriteString(`<link rel="icon" type="image/svg+xml" href="data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 24 24' fill='%2367b279'%3E%3Cpath d='M12 2L4 5v6.09c0 5.05 3.41 9.76 8 10.91 4.59-1.15 8-5.86 8-10.91V5l-8-3zm-1.06 14.54L6.7 12.3a.75.75 0 0 1 1.06-1.06l3.18 3.18 6.36-6.36a.75.75 0 0 1 1.06 1.06l-7.42 7.42a.75.75 0 0 1-1.06 0z'/%3E%3C/svg%3E">`)
	buf.WriteString(`<style>`)
	buf.WriteString(`:root{--bg:#1f1f1f;--header-border:#2d2d2d;--card:#282828;--card-border:rgba(255,255,255,0.08);--row-border:rgba(255,255,255,0.08);--text:#ffffff;--lead:#b5b5b5;--label:#8c8c8c;--val:#f0f0f0;--ctrl-bg:#2a2a2a;--ctrl-border:#3d3d3d;--ctrl-hover:#343434;--accent:#67b279;--reason-bg:rgba(103,178,121,0.09);--panel-border:var(--card-border);--tech-bg:rgba(255,255,255,0.025);--shadow:0 18px 44px rgba(0,0,0,0.34)}`)
	buf.WriteString(`[data-theme="light"]{--bg:#f5f6f8;--header-border:#e5e7eb;--card:#ffffff;--card-border:#e5e7eb;--row-border:#edf0f3;--text:#111827;--lead:#4b5563;--label:#6b7280;--val:#111827;--ctrl-bg:#ffffff;--ctrl-border:#d1d5db;--ctrl-hover:#f3f4f6;--reason-bg:rgba(103,178,121,0.07);--tech-bg:#f9fafb;--shadow:0 10px 30px rgba(15,23,42,0.06)}`)
	buf.WriteString(`@media(prefers-color-scheme:light){[data-theme="auto"]{--bg:#f5f6f8;--header-border:#e5e7eb;--card:#ffffff;--card-border:#e5e7eb;--row-border:#edf0f3;--text:#111827;--lead:#4b5563;--label:#6b7280;--val:#111827;--ctrl-bg:#ffffff;--ctrl-border:#d1d5db;--ctrl-hover:#f3f4f6;--reason-bg:rgba(103,178,121,0.07);--tech-bg:#f9fafb;--shadow:0 10px 30px rgba(15,23,42,0.06)}}`)
	buf.WriteString(`.cat-malware{--accent:#f87171;--reason-bg:rgba(248,113,113,0.09);--panel-border:rgba(248,113,113,0.28)}`)
	buf.WriteString(`[data-theme="light"] .cat-malware{--accent:#ef4444;--reason-bg:rgba(239,68,68,0.06);--panel-border:rgba(239,68,68,0.28)}`)
	buf.WriteString(`@media(prefers-color-scheme:light){[data-theme="auto"] .cat-malware{--accent:#ef4444;--reason-bg:rgba(239,68,68,0.06);--panel-border:rgba(239,68,68,0.28)}}`)
	buf.WriteString(`.cat-adult{--accent:#f59e0b;--reason-bg:rgba(245,158,11,0.08);--panel-border:rgba(245,158,11,0.24)}`)
	buf.WriteString(`.cat-service{--accent:#60a5fa;--reason-bg:rgba(96,165,250,0.08)}`)
	buf.WriteString(`[data-theme="light"] .cat-service{--accent:#2563eb;--reason-bg:rgba(37,99,235,0.06)}`)
	buf.WriteString(`@media(prefers-color-scheme:light){[data-theme="auto"] .cat-service{--accent:#2563eb;--reason-bg:rgba(37,99,235,0.06)}}`)
	buf.WriteString(`*,*::before,*::after{box-sizing:border-box;margin:0;padding:0}`)
	buf.WriteString(`html,body{min-height:100vh;background:var(--bg);color:var(--text);font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,Helvetica,Arial,sans-serif;-webkit-font-smoothing:antialiased;display:flex;flex-direction:column}`)
	buf.WriteString(`.hdr{height:64px;display:flex;align-items:center;justify-content:space-between;padding:0 40px;border-bottom:1px solid var(--header-border);background:var(--bg);flex-shrink:0}`)
	buf.WriteString(`.hdr-brand{display:flex;align-items:center;gap:12px}`)
	buf.WriteString(`.hdr-brand-text{font-size:16px;font-weight:700;color:var(--text);letter-spacing:-0.01em}`)
	buf.WriteString(`.hdr-actions{display:flex;align-items:center;gap:8px}`)
	buf.WriteString(`.hdr-btn{display:inline-flex;align-items:center;gap:6px;height:34px;padding:0 11px;border-radius:8px;background:var(--ctrl-bg);border:1px solid var(--ctrl-border);color:var(--text);font-family:inherit;font-size:12.5px;font-weight:600;text-decoration:none;cursor:pointer;transition:background-color 0.15s ease,border-color 0.15s ease}`)
	buf.WriteString(`.hdr-btn:hover{background:var(--ctrl-hover);border-color:#67b279}`)
	buf.WriteString(`.stage{flex:1;display:flex;align-items:center;justify-content:center;padding:36px 20px}`)
	buf.WriteString(`.panel{width:100%;max-width:560px;background:var(--card);border:1px solid var(--panel-border);border-radius:16px;box-shadow:var(--shadow);padding:36px 40px 32px}`)
	buf.WriteString(`.kicker{font-size:13px;font-weight:600;color:var(--accent);margin-bottom:8px}`)
	buf.WriteString(`h1{font-size:25px;font-weight:700;line-height:1.26;letter-spacing:-0.015em;color:var(--text);margin-bottom:10px}`)
	buf.WriteString(`.lead{font-size:14.5px;line-height:1.62;color:var(--lead);margin-bottom:20px;word-break:break-word}`)
	buf.WriteString(`.lead strong{font-weight:600;color:var(--text)}`)
	buf.WriteString(`.reason-box{border-left:3px solid var(--accent);background:var(--reason-bg);padding:12px 15px;border-radius:0 10px 10px 0;margin-bottom:20px}`)
	buf.WriteString(`.reason-k{font-size:12px;color:var(--label);margin-bottom:3px}`)
	buf.WriteString(`.reason-v{font-size:14.5px;font-weight:600;color:var(--text);line-height:1.4;word-break:break-word}`)
	buf.WriteString(`.specs{border-top:1px solid var(--row-border);margin-bottom:20px}`)
	buf.WriteString(`.spec-row{display:grid;grid-template-columns:148px 1fr;gap:14px;padding:11px 0;border-bottom:1px solid var(--row-border);align-items:baseline}`)
	buf.WriteString(`.spec-k{font-size:13.5px;color:var(--label)}`)
	buf.WriteString(`.spec-v{font-size:13.5px;font-weight:500;color:var(--val);word-break:break-word;text-align:right}`)
	buf.WriteString(`.spec-v--mono{font-family:ui-monospace,SFMono-Regular,Menlo,Consolas,monospace;font-size:12.5px}`)
	buf.WriteString(`.tech-box{margin-bottom:22px;border:1px solid var(--row-border);border-radius:10px;background:var(--tech-bg)}`)
	buf.WriteString(`.tech-sum{list-style:none;display:flex;align-items:center;justify-content:space-between;padding:10px 14px;font-size:13px;font-weight:600;color:var(--lead);cursor:pointer;user-select:none}`)
	buf.WriteString(`.tech-sum::-webkit-details-marker{display:none}`)
	buf.WriteString(`.tech-sum:hover{color:var(--text)}`)
	buf.WriteString(`.tech-chev{transition:transform 0.15s ease;flex-shrink:0}`)
	buf.WriteString(`.tech-box[open] .tech-chev{transform:rotate(180deg)}`)
	buf.WriteString(`.tech-body{padding:0 14px 4px;border-top:1px solid var(--row-border)}`)
	buf.WriteString(`.tech-body .spec-row:last-child{border-bottom:none}`)
	buf.WriteString(`.actions{display:flex;align-items:center;gap:10px;margin-top:4px}`)
	buf.WriteString(`.btn{display:inline-flex;align-items:center;justify-content:center;height:40px;padding:0 20px;border-radius:9px;font-family:inherit;font-size:14px;font-weight:600;cursor:pointer;border:1px solid transparent;transition:background-color 0.15s ease,border-color 0.15s ease}`)
	buf.WriteString(`.btn-green{background:#67b279;color:#ffffff;border-color:#67b279}`)
	buf.WriteString(`.btn-green:hover{background:#5aa36b;border-color:#5aa36b}`)
	buf.WriteString(`.btn-flat{background:var(--ctrl-bg);color:var(--text);border-color:var(--ctrl-border)}`)
	buf.WriteString(`.btn-flat:hover{background:var(--ctrl-hover)}`)
	buf.WriteString(`@media(max-width:560px){.hdr{padding:0 18px}.panel{padding:26px 20px 24px}h1{font-size:21px}.spec-row{grid-template-columns:1fr;gap:4px}.spec-v{text-align:left}.actions .btn{flex:1}}`)
	buf.WriteString(`</style></head><body>`)

	buf.WriteString(`<header class="hdr"><div class="hdr-brand">`)
	buf.WriteString(adgBrandShieldSVG)
	buf.WriteString(`<span class="hdr-brand-text">adguard home</span></div>`)
	buf.WriteString(`<div class="hdr-actions">`)
	buf.WriteString(`<button type="button" class="hdr-btn" id="thBtn" title="`)
	buf.WriteString(themeAriaText)
	buf.WriteString(`" aria-label="`)
	buf.WriteString(themeAriaText)
	buf.WriteString(`"><svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M4.93 4.93l1.41 1.41M17.66 17.66l1.41 1.41M2 12h2M20 12h2M6.34 17.66l-1.41 1.41M19.07 4.93l-1.41 1.41"/></svg></button>`)
	buf.WriteString(`<a class="hdr-btn" href="`)
	buf.WriteString(nextLangHref)
	buf.WriteString(`" title="`)
	buf.WriteString(langTitle)
	buf.WriteString(`">A文 `)
	buf.WriteString(langBtnCode)
	buf.WriteString(`</a></div></header>`)

	buf.WriteString(`<main class="stage"><section class="panel`)
	buf.WriteString(catClass)
	buf.WriteString(`">`)
	buf.WriteString(`<div class="kicker">`)
	buf.WriteString(html.EscapeString(info.Kicker))
	buf.WriteString(`</div>`)
	buf.WriteString(`<h1>`)
	buf.WriteString(html.EscapeString(info.Heading))
	buf.WriteString(`</h1>`)
	buf.WriteString(`<p class="lead">`)
	buf.WriteString(info.LeadHTML)
	buf.WriteString(`</p>`)

	buf.WriteString(`<div class="reason-box"><div class="reason-k">`)
	buf.WriteString(html.EscapeString(info.ReasonLabel))
	buf.WriteString(`</div><div class="reason-v">`)
	buf.WriteString(html.EscapeString(info.ReasonValue))
	buf.WriteString(`</div></div>`)

	buf.WriteString(`<div class="specs">`)
	writeSpecRow(&buf, lblDomain, domain, false)
	writeSpecRow(&buf, info.SourceLabel, info.SourceValue, false)
	writeSpecRow(&buf, info.CNAMELabel, info.CNAMEValue, true)
	buf.WriteString(`</div>`)

	if hasTechDetails {
		buf.WriteString(`<details class="tech-box"><summary class="tech-sum">`)
		buf.WriteString(html.EscapeString(lblTech))
		buf.WriteString(`<svg class="tech-chev" width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"/></svg></summary><div class="tech-body">`)
		writeSpecRow(&buf, info.ScopeLabel, info.ScopeValue, false)
		writeSpecRow(&buf, info.RuleLabel, info.RuleValue, true)
		if len(info.AllSources) > 1 {
			writeSpecRow(&buf, lblAllLists, strings.Join(info.AllSources, ", "), false)
		}
		buf.WriteString(`</div></details>`)
	}

	buf.WriteString(`<div class="actions">`)
	buf.WriteString(`<button type="button" class="btn btn-green" id="backBtn" onclick="history.back()">`)
	buf.WriteString(btnBack)
	buf.WriteString(`</button>`)
	buf.WriteString(`<button type="button" class="btn btn-flat" onclick="location.reload()">`)
	buf.WriteString(btnReload)
	buf.WriteString(`</button></div>`)
	buf.WriteString(`</section></main>`)

	buf.WriteString(`<script>(function(){var r=document.documentElement;if(window.history.length<=1){var b=document.getElementById('backBtn');if(b)b.style.display='none';}var t=document.getElementById('thBtn');if(t){t.onclick=function(){var c=r.getAttribute('data-theme');var dark=c==='dark'||(c==='auto'&&!window.matchMedia('(prefers-color-scheme:light)').matches);r.setAttribute('data-theme',dark?'light':'dark');};}})();</script>`)
	buf.WriteString(`</body></html>`)

	hdr := w.Header()
	hdr.Set(httphdr.ContentType, "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(buf.Bytes())
}
