package home

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
)

// youtubeConfig is the YouTube ad blocking configuration stored in YAML.
type youtubeConfig struct {
	Enabled       bool     `yaml:"enabled" json:"enabled"`
	RouteServer   string   `yaml:"route_server" json:"route_server"`
	BlockAds      bool     `yaml:"block_ads" json:"block_ads"`
	BlockTracking bool     `yaml:"block_tracking" json:"block_tracking"`
	CustomDomains []string `yaml:"custom_domains" json:"custom_domains"`
}

func defaultYoutubeConfig() *youtubeConfig {
	return &youtubeConfig{
		Enabled:       false,
		RouteServer:   "",
		BlockAds:      true,
		BlockTracking: true,
		CustomDomains: []string{},
	}
}

// youtubeAdDomains returns the list of known YouTube/Google ad-serving domains.
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

// youtubeTrackingDomains returns tracking domains related to YouTube.
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

// youtubeRewriteDomains returns the YouTube domains used for DNS rewrite routing.
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

type youtubeConfigJSON struct {
	Enabled        bool     `json:"enabled"`
	RouteServer    string   `json:"route_server"`
	BlockAds       bool     `json:"block_ads"`
	BlockTracking  bool     `json:"block_tracking"`
	CustomDomains  []string `json:"custom_domains"`
	AdDomains      []string `json:"ad_domains"`
	TrackDomains   []string `json:"tracking_domains"`
	RewriteDomains []string `json:"rewrite_domains"`
}

func (web *webAPI) registerYouTubeHandlers() {
	web.httpReg.Register(http.MethodGet, "/control/youtube/config", web.handleGetYoutubeConfig)
	web.httpReg.Register(http.MethodPut, "/control/youtube/config/update", web.handlePutYoutubeConfig)
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
			Enabled:        cfg.Enabled,
			RouteServer:    cfg.RouteServer,
			BlockAds:       cfg.BlockAds,
			BlockTracking:  cfg.BlockTracking,
			CustomDomains:  cfg.CustomDomains,
			AdDomains:      youtubeAdDomains(),
			TrackDomains:   youtubeTrackingDomains(),
			RewriteDomains: youtubeRewriteDomains(),
		}
	}()

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
	}()

	web.logger.InfoContext(ctx, "youtube config updated", "enabled", req.Enabled)
	web.confModifier.Apply(ctx)

	aghhttp.OK(ctx, web.logger, w)
}
