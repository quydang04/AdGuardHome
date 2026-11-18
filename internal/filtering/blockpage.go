package filtering

import (
	"fmt"
	"html"
	"net/netip"
	"slices"
	"strings"
	"unicode"

	"github.com/AdguardTeam/AdGuardHome/internal/filtering/rulelist"
)

// BlockCategory is a semantic block type used for color-coding and layout
// treatment on the block page.
type BlockCategory uint8

const (
	// BlockCategoryAds is the default: advertising domains.
	BlockCategoryAds BlockCategory = iota
	// BlockCategoryTracker covers telemetry, analytics, and privacy trackers.
	BlockCategoryTracker
	// BlockCategoryMalware covers phishing, malware, badware, and scams.
	BlockCategoryMalware
	// BlockCategoryAdult covers parental control and restricted adult content.
	BlockCategoryAdult
	// BlockCategoryService covers blocked online applications and services.
	BlockCategoryService
	// BlockCategoryCustom covers custom user rules, hosts rules, and admin checks.
	BlockCategoryCustom
)

// BlockPageInfo contains localized, human-readable details explaining why a
// domain was blocked or allowed.
type BlockPageInfo struct {
	PageTitle   string
	Kicker      string
	Heading     string
	LeadHTML    string
	ReasonLabel string
	ReasonValue string
	SourceLabel string
	SourceValue string
	AllSources  []string
	ScopeLabel  string
	ScopeValue  string
	CNAMELabel  string
	CNAMEValue  string
	RuleLabel   string
	RuleValue   string
	IsWarning   bool
	IsAllowed   bool
	Category    BlockCategory
}

type blockListCategory uint8

const (
	catAdsAndTrackers blockListCategory = iota
	catMalwareThreat
	catTrackerTelemetry
	catRestrictedContent
)

var knownServiceDisplayNames = map[string]string{
	"4chan":                "4chan",
	"9gag":                 "9GAG",
	"activision_blizzard":  "Activision Blizzard",
	"aliexpress":           "AliExpress",
	"amazon":               "Amazon",
	"amino":                "Amino",
	"apple_streaming":      "Apple TV+ / Music",
	"battle_net":           "Battle.net",
	"betano":               "Betano",
	"betfair":              "Betfair",
	"betway":               "Betway",
	"bigo_live":            "Bigo Live",
	"bilibili":             "Bilibili",
	"blizzard":             "Blizzard",
	"bluesky":              "Bluesky",
	"box":                  "Box",
	"canva":                "Canva",
	"chatgpt":              "ChatGPT",
	"claude":               "Claude",
	"cloudflare":           "Cloudflare",
	"clubhouse":            "Clubhouse",
	"crunchyroll":          "Crunchyroll",
	"dailymotion":          "Dailymotion",
	"deezer":               "Deezer",
	"directvgo":            "DIRECTV GO",
	"discord":              "Discord",
	"discoveryplus":        "Discovery+",
	"disneyplus":           "Disney+",
	"douban":               "Douban",
	"dropbox":              "Dropbox",
	"ebay":                 "eBay",
	"electronic_arts":      "Electronic Arts",
	"epic_games":           "Epic Games",
	"espn":                 "ESPN",
	"facebook":             "Facebook",
	"flickr":               "Flickr",
	"fubo":                 "FuboTV",
	"github":               "GitHub",
	"glovo":                "Glovo",
	"gog":                  "GOG.com",
	"hbomax":               "HBO Max",
	"hulu":                 "Hulu",
	"icloud_private_relay": "iCloud Private Relay",
	"imgur":                "Imgur",
	"instagram":            "Instagram",
	"iqiyi":                "iQIYI",
	"kakao_talk":           "KakaoTalk",
	"kik":                  "Kik",
	"lazada":               "Lazada",
	"league_of_legends":    "League of Legends",
	"line":                 "LINE",
	"linkedin":             "LinkedIn",
	"lookfantastic":        "LOOKFANTASTIC",
	"mail_ru":              "Mail.ru",
	"mastodon":             "Mastodon",
	"mercadolibre":         "Mercado Libre",
	"minecraft":            "Minecraft",
	"nebula":               "Nebula",
	"netflix":              "Netflix",
	"nintendo":             "Nintendo",
	"nvidia":               "NVIDIA",
	"ok":                   "Odnoklassniki",
	"olx":                  "OLX",
	"onlyfans":             "OnlyFans",
	"openai":               "OpenAI",
	"origin":               "EA / Origin",
	"paramountplus":        "Paramount+",
	"peacock_tv":           "Peacock TV",
	"pinterest":            "Pinterest",
	"plenty_of_fish":       "Plenty of Fish",
	"playstation":          "PlayStation",
	"plex":                 "Plex",
	"privacy":              "Privacy",
	"primevideo":           "Prime Video",
	"qq":                   "QQ",
	"qobuz":                "Qobuz",
	"rakuten_viki":         "Rakuten Viki",
	"reddit":               "Reddit",
	"riot_games":           "Riot Games",
	"roblox":               "Roblox",
	"rockstar_games":       "Rockstar Games",
	"samsung_tv_plus":      "Samsung TV Plus",
	"shein":                "SHEIN",
	"shopee":               "Shopee",
	"signal":               "Signal",
	"skype":                "Skype",
	"slack":                "Slack",
	"sling_tv":             "Sling TV",
	"snapchat":             "Snapchat",
	"soundcloud":           "SoundCloud",
	"spotify":              "Spotify",
	"steam":                "Steam",
	"teams":                "Microsoft Teams",
	"telegram":             "Telegram",
	"temu":                 "Temu",
	"threads":              "Threads",
	"tidal":                "TIDAL",
	"tiktok":               "TikTok",
	"tinder":               "Tinder",
	"tumblr":               "Tumblr",
	"twitch":               "Twitch",
	"twitter":              "X (Twitter)",
	"uber":                 "Uber",
	"ubisoft":              "Ubisoft",
	"valorant":             "VALORANT",
	"viber":                "Viber",
	"vimeo":                "Vimeo",
	"vk":                   "VK",
	"voot":                 "Voot",
	"wargaming":            "Wargaming",
	"wechat":               "WeChat",
	"weibo":                "Weibo",
	"whatsapp":             "WhatsApp",
	"wizz":                 "Wizz",
	"world_of_tanks":       "World of Tanks",
	"xboxlive":             "Xbox Live",
	"x":                    "X (Twitter)",
	"xiaohongshu":          "Xiaohongshu",
	"youtube":              "YouTube",
	"zalo":                 "Zalo",
	"zoom":                 "Zoom",
}

// filterListMeta returns the human-readable name and inferred category of a
// configured blocklist by its API ID.
func (d *DNSFilter) filterListMeta(id rulelist.APIID) (name string, cat blockListCategory) {
	if d == nil || d.conf == nil || d.conf.filtersMu == nil {
		return "", catAdsAndTrackers
	}

	d.conf.filtersMu.RLock()
	defer d.conf.filtersMu.RUnlock()

	for i := range d.conf.Filters {
		f := &d.conf.Filters[i]
		if rulelist.APIID(f.ID) == id {
			if f.Name != "" {
				name = f.Name
			} else {
				name = f.URL
			}

			return name, classifyFilterList(f.Name, f.URL)
		}
	}

	return "", catAdsAndTrackers
}

// allowListName returns the human-readable name of a configured allowlist by
// its API ID, or empty if not found.
func (d *DNSFilter) allowListName(id rulelist.APIID) (name string) {
	if d == nil || d.conf == nil || d.conf.filtersMu == nil {
		return ""
	}

	d.conf.filtersMu.RLock()
	defer d.conf.filtersMu.RUnlock()

	for i := range d.conf.WhitelistFilters {
		f := &d.conf.WhitelistFilters[i]
		if rulelist.APIID(f.ID) == id {
			if f.Name != "" {
				return f.Name
			}

			return f.URL
		}
	}

	return ""
}

func classifyFilterList(name, url string) blockListCategory {
	combined := strings.ToLower(name + " " + url)
	switch {
	case containsAny(combined,
		"badware", "malware", "phish", "scam", "fraud", "ransom", "virus",
		"threat", "abuse", "botnet", "stalkerware", "trojan", "cryptojack",
		"coinminer", "nocoin", "spam404", "urlhaus", "phishtank", "openphish",
		"chongluadao", "luadao", "anti-scam", "dga", "typosquat", "exploit", "malicious",
	):
		return catMalwareThreat
	case containsAny(combined,
		"adult", "porn", "nsfw", "gambling", "betting", "casino", "18+", "xxx",
	):
		return catRestrictedContent
	case containsAny(combined,
		"tracker", "tracking", "telemetry", "easyprivacy", "privacy", "analytics",
		"metrics", "spyware", "spy", "smart-tv", "smarttv", "windows/office",
	):
		return catTrackerTelemetry
	default:
		return catAdsAndTrackers
	}
}

func isLikelyTrackerHost(domain, ruleText string) bool {
	combined := strings.ToLower(domain + " " + ruleText)

	return containsAny(combined,
		"google-analytics", "analytics.", ".analytics.", "telemetry.", ".telemetry.",
		"metrics.", ".metrics.", "pixel.", "px.", "beacon.", "ping.", "chartbeat",
		"hotjar", "mixpanel", "segment.", "sentry.", "app-measurement",
		"appsflyer", "adjust.com", "go-mpulse", "scorecardresearch", "quantserve",
	)
}

func isLikelyAdOrTrackerHost(domain, ruleText string) bool {
	if isLikelyTrackerHost(domain, ruleText) {
		return true
	}

	combined := strings.ToLower(domain + " " + ruleText)

	return containsAny(combined,
		"doubleclick", "googlesyndication", "googleadservices", "adservice.",
		"moatads", "ads.", "ad.", "adserver", "clix.vn", "adtima", "admicro",
		"eclick", "criteo", "taboola", "outbrain", "popads", "pubmatic", "rubiconproject",
	)
}

func containsAny(haystack string, needles ...string) bool {
	for _, n := range needles {
		if strings.Contains(haystack, n) {
			return true
		}
	}

	return false
}

// describeRuleScope converts raw Adblock/hosts filtering syntax into a plain,
// human-readable description of what the matched rule targets.
func describeRuleScope(domain, ruleText string, isVi bool) string {
	clean := strings.TrimSpace(ruleText)
	if clean == "" {
		return ""
	}

	hasImportant := false
	lower := strings.ToLower(clean)
	if strings.Contains(lower, "$important") || strings.Contains(lower, ",important") {
		hasImportant = true
	}

	isAllowRule := strings.HasPrefix(clean, "@@")
	pattern := clean
	if isAllowRule {
		pattern = strings.TrimPrefix(pattern, "@@")
	}

	var scope string

	// Check /etc/hosts syntax: "<ip> <host>..."
	fields := strings.Fields(pattern)
	if len(fields) >= 2 {
		if _, err := netip.ParseAddr(fields[0]); err == nil {
			if isAllowRule {
				scope = pickStr(isVi, "Cho phép trực tiếp theo tên miền (hosts)", "Exact domain allow (hosts rule)")
			} else {
				scope = pickStr(isVi, "Chặn trực tiếp theo tên miền (quy tắc hosts)", "Exact domain match (hosts rule)")
			}
			if hasImportant {
				scope += pickStr(isVi, " · Quy tắc ưu tiên cao", " · High-priority rule")
			}

			return scope
		}
	}

	switch {
	case strings.HasPrefix(pattern, "||"):
		rest := strings.TrimPrefix(pattern, "||")
		idx := strings.IndexAny(rest, "^$/|*")
		target := rest
		if idx >= 0 {
			target = rest[:idx]
		}
		target = strings.TrimSuffix(strings.TrimSpace(target), ".")
		if target != "" {
			if isAllowRule {
				if !strings.EqualFold(domain, target) && strings.HasSuffix(strings.ToLower(domain), "."+strings.ToLower(target)) {
					scope = pickStr(isVi,
						"Cho phép theo tên miền gốc "+target+" và mọi tên miền phụ",
						"Allowed via parent domain "+target+" and its subdomains",
					)
				} else {
					scope = pickStr(isVi,
						"Cho phép tên miền "+target+" và các tên miền phụ",
						"Allows "+target+" and its subdomains",
					)
				}
			} else if !strings.EqualFold(domain, target) && strings.HasSuffix(strings.ToLower(domain), "."+strings.ToLower(target)) {
				scope = pickStr(isVi,
					"Chặn theo tên miền gốc "+target+" và mọi tên miền phụ",
					"Blocked via parent domain "+target+" and its subdomains",
				)
			} else {
				scope = pickStr(isVi,
					"Chặn tên miền "+target+" và các tên miền phụ",
					"Blocks "+target+" and its subdomains",
				)
			}
		} else {
			scope = pickStr(isVi, "Khớp theo mẫu quy tắc lọc", "Matched filtering pattern")
		}

	case isRegexRule(pattern):
		scope = pickStr(isVi,
			"Khớp theo mẫu biểu thức chính quy (Regex)",
			"Matched via regular expression (Regex)",
		)

	default:
		if isAllowRule {
			scope = pickStr(isVi, "Khớp theo mẫu quy tắc ngoại lệ", "Matched exception rule pattern")
		} else {
			scope = pickStr(isVi, "Khớp theo mẫu quy tắc lọc", "Matched filtering pattern")
		}
	}

	if hasImportant {
		scope += pickStr(isVi, " · Quy tắc ưu tiên cao", " · High-priority rule")
	}

	return scope
}

func isRegexRule(pattern string) bool {
	if !strings.HasPrefix(pattern, "/") {
		return false
	}

	body := pattern
	if idx := strings.LastIndex(pattern, "$"); idx > 1 {
		body = pattern[:idx]
	}

	return len(body) > 2 && strings.HasSuffix(body, "/")
}

// FilterListName returns the human-readable name of a configured blocklist by
// its API ID, or an empty string if not found.
func (d *DNSFilter) FilterListName(id rulelist.APIID) string {
	name, _ := d.filterListMeta(id)

	return name
}

// DescribeBlock constructs accurate, localized human-readable block information
// for rendering on the HTTP block page.
func (d *DNSFilter) DescribeBlock(domain string, res *Result, isVi bool) BlockPageInfo {
	safeDomain := html.EscapeString(domain)
	boldDomain := "<strong>" + safeDomain + "</strong>"

	info := BlockPageInfo{
		ReasonLabel: pickStr(isVi, "Lý do chặn", "Reason"),
		SourceLabel: pickStr(isVi, "Danh sách lọc", "Filter list"),
		ScopeLabel:  pickStr(isVi, "Phạm vi chặn", "Match scope"),
		CNAMELabel:  pickStr(isVi, "Tên miền đích (CNAME)", "CNAME target"),
		RuleLabel:   pickStr(isVi, "Quy tắc khớp", "Matched rule"),
		Category:    BlockCategoryAds,
	}

	if res != nil {
		cname := strings.TrimSuffix(strings.TrimSpace(res.CanonName), ".")
		if cname != "" && !strings.EqualFold(cname, domain) {
			info.CNAMEValue = cname
		}
		for _, r := range res.Rules {
			if r != nil && strings.TrimSpace(r.Text) != "" {
				info.RuleValue = strings.TrimSpace(r.Text)
				info.ScopeValue = describeRuleScope(domain, info.RuleValue, isVi)
				break
			}
		}
	}

	if res == nil {
		info.PageTitle = pickStr(isVi, "Đã chặn truy cập — "+domain, domain+" is blocked")
		info.Kicker = pickStr(isVi, "Bảo vệ DNS", "DNS Protection")
		info.Heading = pickStr(isVi, "Trang này đã bị chặn", "This site has been blocked")
		info.LeadHTML = pickStr(
			isVi,
			"Kết nối tới "+boldDomain+" đã được chặn tự động bởi bộ lọc DNS trên mạng của bạn.",
			"Connection to "+boldDomain+" was automatically blocked by DNS filtering on your network.",
		)
		info.ReasonValue = pickStr(isVi, "Quảng cáo & theo dõi", "Advertising & tracking")
		info.SourceValue = pickStr(isVi, "Bộ lọc DNS", "DNS Filter")

		return info
	}

	switch res.Reason {
	case FilteredSafeBrowsing:
		info.IsWarning = true
		info.Category = BlockCategoryMalware
		info.PageTitle = pickStr(isVi, "Cảnh báo nguy hiểm — "+domain, "Security warning — "+domain)
		info.Kicker = pickStr(isVi, "Cảnh báo bảo mật", "Security Warning")
		info.Heading = pickStr(isVi, "Trang web nguy hiểm đã bị chặn", "Dangerous website blocked")
		info.LeadHTML = pickStr(
			isVi,
			"AdGuard Home đã ngắt kết nối tới "+boldDomain+" vì phát hiện dấu hiệu lừa đảo đánh cắp thông tin hoặc phát tán phần mềm độc hại.",
			"AdGuard Home blocked "+boldDomain+" because it is flagged for phishing or distributing malware.",
		)
		info.ReasonValue = pickStr(isVi, "Trang lừa đảo hoặc chứa mã độc", "Phishing or malware threat")
		info.SourceLabel = pickStr(isVi, "Tính năng bảo vệ", "Protected by")
		info.SourceValue = pickStr(isVi, "Duyệt web an toàn (Safe Browsing)", "AdGuard Safe Browsing")
		info.RuleValue = ""
		info.ScopeValue = ""

		return info

	case FilteredParental:
		info.IsWarning = true
		info.Category = BlockCategoryAdult
		info.PageTitle = pickStr(isVi, "Giới hạn nội dung — "+domain, domain+" is restricted")
		info.Kicker = pickStr(isVi, "Kiểm soát nội dung", "Parental Control")
		info.Heading = pickStr(isVi, "Trang này đang bị giới hạn", "Content is restricted")
		info.LeadHTML = pickStr(
			isVi,
			"Bạn không thể mở "+boldDomain+" vì mạng đang bật chế độ lọc nội dung người lớn và trang web không phù hợp lứa tuổi.",
			"Access to "+boldDomain+" is restricted because adult content filtering is active on this network.",
		)
		info.ReasonValue = pickStr(isVi, "Nội dung giới hạn độ tuổi", "Age-restricted content")
		info.SourceLabel = pickStr(isVi, "Tính năng bảo vệ", "Protected by")
		info.SourceValue = pickStr(isVi, "Kiểm soát của phụ huynh", "Parental Control")
		info.RuleValue = ""
		info.ScopeValue = ""

		return info

	case FilteredBlockedService:
		svc := formatServiceName(res.ServiceName)
		safeSvc := html.EscapeString(svc)
		info.Category = BlockCategoryService
		info.PageTitle = pickStr(isVi, "Dịch vụ tạm khóa — "+domain, domain+" is blocked")
		info.Kicker = pickStr(isVi, "Giới hạn dịch vụ", "Service Restriction")
		if svc != "" {
			info.Heading = pickStr(isVi, "Dịch vụ "+svc+" đang bị chặn", svc+" is currently blocked")
			info.LeadHTML = pickStr(
				isVi,
				"Quản trị viên mạng đang tạm khóa dịch vụ <strong>"+safeSvc+"</strong>, nên thiết bị không thể kết nối tới "+boldDomain+".",
				"Access to <strong>"+safeSvc+"</strong> is currently restricted on this network, so "+boldDomain+" cannot be opened.",
			)
		} else {
			info.Heading = pickStr(isVi, "Dịch vụ này đang bị chặn", "This service is currently blocked")
			info.LeadHTML = pickStr(
				isVi,
				"Truy cập tới "+boldDomain+" không khả dụng vì dịch vụ này đang bị tạm khóa trên mạng.",
				"Access to "+boldDomain+" is unavailable because this service is currently restricted on this network.",
			)
		}
		info.ReasonValue = pickStr(isVi, "Chặn theo ứng dụng / dịch vụ", "Blocked online service")
		info.SourceLabel = pickStr(isVi, "Ứng dụng / Dịch vụ", "Blocked service")
		if svc != "" {
			info.SourceValue = svc
		} else {
			info.SourceValue = pickStr(isVi, "Dịch vụ trực tuyến", "Online service")
		}

		return info

	case FilteredInvalid:
		info.Category = BlockCategoryCustom
		info.PageTitle = pickStr(isVi, "Tên miền không hợp lệ — "+domain, domain+" is invalid")
		info.Kicker = pickStr(isVi, "Bảo vệ DNS", "DNS Protection")
		info.Heading = pickStr(isVi, "Tên miền không hợp lệ", "Invalid domain request")
		info.LeadHTML = pickStr(
			isVi,
			"Không thể kết nối tới "+boldDomain+" do định dạng tên miền không đúng chuẩn.",
			"The request to "+boldDomain+" was rejected because the domain format is invalid.",
		)
		info.ReasonValue = pickStr(isVi, "Sai định dạng tên miền", "Invalid domain format")
		info.SourceLabel = pickStr(isVi, "Nguồn chặn", "Blocked by")
		info.SourceValue = "AdGuard Home"

		return info
	}

	// FilteredBlockList: inspect matched rules to distinguish custom rules,
	// system hosts, malware lists, tracker lists, and general ad blocklists.
	var (
		sources      []string
		onlyCustom   = true
		onlyEtcHosts = true
		hasRules     = false
		topCat       = catAdsAndTrackers
	)

	for _, r := range res.Rules {
		if r == nil {
			continue
		}
		hasRules = true

		var name string
		switch r.FilterListID {
		case rulelist.APIIDCustom:
			onlyEtcHosts = false
			name = pickStr(isVi, "Quy tắc lọc tùy chỉnh", "Custom filtering rules")
		case rulelist.APIIDEtcHosts:
			onlyCustom = false
			name = pickStr(isVi, "Tệp hosts hệ thống", "System hosts file")
		default:
			onlyCustom = false
			onlyEtcHosts = false
			if d != nil {
				var cat blockListCategory
				name, cat = d.filterListMeta(r.FilterListID)
				if cat == catMalwareThreat {
					topCat = catMalwareThreat
				} else if cat == catRestrictedContent && topCat == catAdsAndTrackers {
					topCat = catRestrictedContent
				} else if cat == catTrackerTelemetry && topCat == catAdsAndTrackers {
					topCat = catTrackerTelemetry
				}
			}
		}

		if name != "" && !slices.Contains(sources, name) {
			sources = append(sources, name)
		}
	}

	info.PageTitle = pickStr(isVi, "Đã chặn truy cập — "+domain, domain+" is blocked")
	info.Kicker = pickStr(isVi, "Bảo vệ DNS", "DNS Protection")
	info.Heading = pickStr(isVi, "Trang này đã bị chặn", "This site has been blocked")

	switch {
	case hasRules && onlyCustom:
		info.Category = BlockCategoryCustom
		info.Kicker = pickStr(isVi, "Quy tắc tùy chỉnh", "Custom Rule")
		info.LeadHTML = pickStr(
			isVi,
			"Tên miền "+boldDomain+" đã bị chặn theo quy tắc riêng do quản trị viên mạng thiết lập.",
			"Access to "+boldDomain+" was blocked by a custom rule configured by the network administrator.",
		)
		if isLikelyAdOrTrackerHost(domain, info.RuleValue) {
			info.ReasonValue = pickStr(isVi, "Chặn thủ công (quảng cáo / theo dõi)", "Custom rule (ads & tracking)")
		} else {
			info.ReasonValue = pickStr(isVi, "Chặn thủ công bởi quản trị viên", "Blocked by administrator rule")
		}
		info.SourceLabel = pickStr(isVi, "Nguồn chặn", "Blocked by")
		info.SourceValue = pickStr(isVi, "Quy tắc lọc tùy chỉnh", "Custom filtering rules")

	case hasRules && onlyEtcHosts:
		info.Category = BlockCategoryCustom
		info.Kicker = pickStr(isVi, "Cấu hình máy chủ", "Host Configuration")
		info.LeadHTML = pickStr(
			isVi,
			"Tên miền "+boldDomain+" đang bị chặn trực tiếp trong tệp hosts của máy chủ DNS.",
			"Access to "+boldDomain+" is blocked via the DNS server's system hosts file.",
		)
		info.ReasonValue = pickStr(isVi, "Chặn qua tệp hosts máy chủ", "Blocked via system hosts")
		info.SourceLabel = pickStr(isVi, "Nguồn chặn", "Blocked by")
		info.SourceValue = pickStr(isVi, "Tệp hosts hệ thống", "System hosts file")

	case topCat == catMalwareThreat:
		info.IsWarning = true
		info.Category = BlockCategoryMalware
		info.PageTitle = pickStr(isVi, "Cảnh báo nguy hiểm — "+domain, "Security warning — "+domain)
		info.Kicker = pickStr(isVi, "Cảnh báo bảo mật", "Security Warning")
		info.Heading = pickStr(isVi, "Trang web không an toàn", "Unsafe website blocked")
		info.LeadHTML = pickStr(
			isVi,
			"Kết nối tới "+boldDomain+" đã bị chặn vì tên miền này nằm trong danh sách cảnh báo lừa đảo, giả mạo hoặc phát tán mã độc.",
			"Connection to "+boldDomain+" was blocked because this domain is listed for phishing, scams, or malware risks.",
		)
		info.ReasonValue = pickStr(isVi, "Nguy cơ lừa đảo hoặc mã độc", "Phishing or malware risk")
		info.SourceLabel = pickStr(isVi, "Danh sách cảnh báo", "Threat list")
		setBlockSources(&info, sources, isVi)

	case topCat == catRestrictedContent:
		info.IsWarning = true
		info.Category = BlockCategoryAdult
		info.Kicker = pickStr(isVi, "Kiểm soát nội dung", "Content Filtering")
		info.Heading = pickStr(isVi, "Nội dung này đã bị chặn", "Restricted content blocked")
		info.LeadHTML = pickStr(
			isVi,
			"Trang "+boldDomain+" không thể mở vì thuộc danh sách chặn nội dung nhạy cảm hoặc hạn chế trên mạng này.",
			"Access to "+boldDomain+" was blocked because it belongs to a restricted content list on this network.",
		)
		info.ReasonValue = pickStr(isVi, "Nội dung nhạy cảm / hạn chế", "Restricted content")
		setBlockSources(&info, sources, isVi)

	case topCat == catTrackerTelemetry || isLikelyTrackerHost(domain, info.RuleValue):
		info.Category = BlockCategoryTracker
		info.Kicker = pickStr(isVi, "Bảo vệ quyền riêng tư", "Privacy Protection")
		info.Heading = pickStr(isVi, "Đã chặn kết nối theo dõi", "Tracking connection blocked")
		info.LeadHTML = pickStr(
			isVi,
			"AdGuard Home đã chặn "+boldDomain+" để ngăn trang web hoặc ứng dụng âm thầm thu thập dữ liệu hoạt động từ thiết bị của bạn.",
			"AdGuard Home blocked "+boldDomain+" to stop websites or apps from silently collecting telemetry and activity data.",
		)
		info.ReasonValue = pickStr(isVi, "Theo dõi & thu thập dữ liệu", "Tracking & telemetry")
		setBlockSources(&info, sources, isVi)

	default:
		info.Category = BlockCategoryAds
		info.LeadHTML = pickStr(
			isVi,
			"Kết nối tới "+boldDomain+" đã được chặn tự động vì tên miền này chuyên phân phối quảng cáo hoặc theo dõi người dùng.",
			"Connection to "+boldDomain+" was automatically blocked because this domain serves ads or tracks browsing activity.",
		)
		info.ReasonValue = pickStr(isVi, "Quảng cáo & theo dõi", "Advertising & tracking")
		setBlockSources(&info, sources, isVi)
	}

	if info.CNAMEValue != "" {
		info.ReasonValue += pickStr(isVi, " (ẩn qua CNAME)", " (via CNAME)")
	}

	return info
}

// DescribeAllowed constructs localized diagnostic information for the preview
// and domain-checker tool when a queried domain is NOT currently blocked.
func (d *DNSFilter) DescribeAllowed(domain string, res *Result, isVi bool) BlockPageInfo {
	safeDomain := html.EscapeString(domain)
	boldDomain := "<strong>" + safeDomain + "</strong>"

	info := BlockPageInfo{
		PageTitle:   pickStr(isVi, domain+" được phép truy cập", domain+" is allowed"),
		Kicker:      pickStr(isVi, "Kiểm tra tên miền", "Domain Check"),
		Heading:     pickStr(isVi, "Tên miền này không bị chặn", "This domain is allowed"),
		ReasonLabel: pickStr(isVi, "Trạng thái", "Status"),
		SourceLabel: pickStr(isVi, "Nguồn cho phép", "Allowed by"),
		ScopeLabel:  pickStr(isVi, "Phạm vi áp dụng", "Rule scope"),
		CNAMELabel:  pickStr(isVi, "Tên miền đích (CNAME)", "CNAME target"),
		RuleLabel:   pickStr(isVi, "Quy tắc ngoại lệ", "Exception rule"),
		IsAllowed:   true,
		Category:    BlockCategoryCustom,
	}

	if res != nil && res.Reason == NotFilteredAllowList {
		var sources []string
		for _, r := range res.Rules {
			if r == nil {
				continue
			}
			if info.RuleValue == "" && strings.TrimSpace(r.Text) != "" {
				info.RuleValue = strings.TrimSpace(r.Text)
				info.ScopeValue = describeRuleScope(domain, info.RuleValue, isVi)
			}

			var name string
			switch r.FilterListID {
			case rulelist.APIIDCustom:
				name = pickStr(isVi, "Quy tắc lọc tùy chỉnh", "Custom filtering rules")
			case rulelist.APIIDEtcHosts:
				name = pickStr(isVi, "Tệp hosts hệ thống", "System hosts file")
			default:
				if d != nil {
					name = d.allowListName(r.FilterListID)
				}
			}

			if name != "" && !slices.Contains(sources, name) {
				sources = append(sources, name)
			}
		}

		info.LeadHTML = pickStr(
			isVi,
			"Tên miền "+boldDomain+" hiện đang được cho phép đi qua vì khớp với quy tắc ngoại lệ (allowlist) trên hệ thống.",
			"Access to "+boldDomain+" is permitted because it matches an active allowlist exception rule.",
		)
		info.ReasonValue = pickStr(isVi, "Được mở khóa bởi danh sách cho phép", "Allowed by exception rule")
		if len(sources) > 0 {
			info.SourceValue = sources[0]
			if len(sources) > 1 {
				info.AllSources = sources
				info.SourceValue = fmt.Sprintf("%s (+%d)", sources[0], len(sources)-1)
			}
		} else {
			info.SourceValue = pickStr(isVi, "Danh sách cho phép", "Allowlist")
		}

		return info
	}

	info.LeadHTML = pickStr(
		isVi,
		"Tên miền "+boldDomain+" không trùng với bất kỳ bộ lọc quảng cáo, cảnh báo bảo mật hay dịch vụ nào đang bật. Thiết bị có thể truy cập bình thường.",
		boldDomain+" does not match any active blocklist, security filter, or service restriction. Traffic to this domain proceeds normally.",
	)
	info.ReasonValue = pickStr(isVi, "Không bị chặn — truy cập bình thường", "Not blocked — allowed normally")
	info.SourceLabel = pickStr(isVi, "Trạng thái bộ lọc", "Filter status")
	info.SourceValue = pickStr(isVi, "Không khớp quy tắc chặn nào", "No matching block rule")

	return info
}

// DescribePreviewScenario constructs a realistic preview BlockPageInfo for the
// admin scenario preview buttons (ads, tracker, malware, service, parental, custom).
func (d *DNSFilter) DescribePreviewScenario(scenario, domain string, isVi bool) BlockPageInfo {
	s := strings.ToLower(strings.TrimSpace(scenario))
	if domain == "" {
		switch s {
		case "malware":
			domain = "phishing.example.com"
		case "service":
			domain = "tiktok.com"
		case "parental":
			domain = "restricted.example.com"
		case "custom":
			domain = "blocked.example.com"
		case "tracker":
			domain = "analytics.example.com"
		default:
			domain = "ads.example.com"
		}
	}

	switch s {
	case "malware":
		info := d.DescribeBlock(domain, &Result{
			Reason:     FilteredSafeBrowsing,
			IsFiltered: true,
		}, isVi)
		return info

	case "service":
		return d.DescribeBlock(domain, &Result{
			Reason:      FilteredBlockedService,
			ServiceName: "tiktok",
			IsFiltered:  true,
		}, isVi)

	case "parental":
		return d.DescribeBlock(domain, &Result{
			Reason:     FilteredParental,
			IsFiltered: true,
		}, isVi)

	case "custom":
		return d.DescribeBlock(domain, &Result{
			Reason:     FilteredBlockList,
			IsFiltered: true,
			Rules: []*ResultRule{
				{
					FilterListID: rulelist.APIIDCustom,
					Text:         "||" + domain + "^",
				},
			},
		}, isVi)

	case "tracker":
		ruleText := "||" + domain + "^"
		info := d.DescribeBlock(domain, nil, isVi)
		info.Category = BlockCategoryTracker
		info.Kicker = pickStr(isVi, "Bảo vệ quyền riêng tư", "Privacy Protection")
		info.Heading = pickStr(isVi, "Đã chặn kết nối theo dõi", "Tracking connection blocked")
		safeDomain := html.EscapeString(domain)
		info.LeadHTML = pickStr(
			isVi,
			"AdGuard Home đã chặn <strong>"+safeDomain+"</strong> để ngăn trang web hoặc ứng dụng âm thầm thu thập dữ liệu hoạt động từ thiết bị của bạn.",
			"AdGuard Home blocked <strong>"+safeDomain+"</strong> to stop websites or apps from silently collecting telemetry and activity data.",
		)
		info.ReasonValue = pickStr(isVi, "Theo dõi & thu thập dữ liệu", "Tracking & telemetry")
		info.SourceValue = d.firstEnabledFilterName("tracker", "AdGuard DNS filter")
		info.RuleValue = ruleText
		info.ScopeValue = describeRuleScope(domain, ruleText, isVi)

		return info

	default:
		ruleText := "||" + domain + "^"
		info := d.DescribeBlock(domain, nil, isVi)
		info.SourceValue = d.firstEnabledFilterName("", "AdGuard DNS filter")
		info.RuleValue = ruleText
		info.ScopeValue = describeRuleScope(domain, ruleText, isVi)

		return info
	}
}

func (d *DNSFilter) firstEnabledFilterName(preferSubstr, fallback string) string {
	if d == nil || d.conf == nil || d.conf.filtersMu == nil {
		return fallback
	}

	d.conf.filtersMu.RLock()
	defer d.conf.filtersMu.RUnlock()

	var firstEnabled string
	for i := range d.conf.Filters {
		f := &d.conf.Filters[i]
		if !f.Enabled || f.Name == "" {
			continue
		}
		if firstEnabled == "" {
			firstEnabled = f.Name
		}
		if preferSubstr != "" && strings.Contains(strings.ToLower(f.Name), preferSubstr) {
			return f.Name
		}
	}

	if firstEnabled != "" {
		return firstEnabled
	}

	return fallback
}

func setBlockSources(info *BlockPageInfo, sources []string, isVi bool) {
	if len(sources) == 0 {
		info.SourceValue = pickStr(isVi, "Bộ lọc DNS", "DNS Filter")
		return
	}
	if len(sources) == 1 {
		info.SourceValue = sources[0]
		return
	}

	info.AllSources = slices.Clone(sources)
	info.SourceValue = fmt.Sprintf("%s (+%d)", sources[0], len(sources)-1)
}

func pickStr(isVi bool, vi, en string) string {
	if isVi {
		return vi
	}

	return en
}

func formatServiceName(s string) string {
	raw := strings.TrimSpace(strings.ToLower(s))
	if raw == "" {
		return ""
	}
	if known, ok := knownServiceDisplayNames[raw]; ok {
		return known
	}

	clean := strings.ReplaceAll(raw, "_", " ")
	runes := []rune(clean)
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}
