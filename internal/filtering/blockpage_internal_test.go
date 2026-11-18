package filtering

import (
	"testing"

	"github.com/AdguardTeam/AdGuardHome/internal/filtering/rulelist"
	"github.com/stretchr/testify/assert"
)

func TestDescribeBlockAndAllowed(t *testing.T) {
	flt, _ := newForTest(t, &Config{
		Filters: []FilterYAML{
			{
				Filter: Filter{ID: 1},
				Name:   "AdGuard DNS filter",
				URL:    "https://example.org/ads.txt",
				Enabled: true,
			},
			{
				Filter: Filter{ID: 2},
				Name:   "uBlock filters – Badware risks",
				URL:    "https://example.org/badware.txt",
				Enabled: true,
			},
			{
				Filter: Filter{ID: 3},
				Name:   "HaGeZi Windows Telemetry",
				URL:    "https://example.org/telemetry.txt",
				Enabled: true,
			},
			{
				Filter: Filter{ID: 4},
				Name:   "Adult Content Filter",
				URL:    "https://example.org/adult.txt",
				Enabled: true,
			},
		},
		WhitelistFilters: []FilterYAML{
			{
				Filter: Filter{ID: 100},
				Name:   "Personal Allowlist",
				URL:    "https://example.org/allow.txt",
				Enabled: true,
			},
		},
	}, nil)

	t.Run("safebrowsing_vi", func(t *testing.T) {
		info := flt.DescribeBlock("evil-phish.example", &Result{
			Reason:     FilteredSafeBrowsing,
			IsFiltered: true,
		}, true)
		assert.True(t, info.IsWarning)
		assert.Equal(t, BlockCategoryMalware, info.Category)
		assert.Equal(t, "Trang web nguy hiểm đã bị chặn", info.Heading)
		assert.Equal(t, "Trang lừa đảo hoặc chứa mã độc", info.ReasonValue)
	})

	t.Run("blocked_service_vi_and_en", func(t *testing.T) {
		infoVi := flt.DescribeBlock("www.youtube.com", &Result{
			Reason:      FilteredBlockedService,
			ServiceName: "youtube",
			IsFiltered:  true,
		}, true)
		assert.Equal(t, BlockCategoryService, infoVi.Category)
		assert.Equal(t, "Dịch vụ YouTube đang bị chặn", infoVi.Heading)
		assert.Equal(t, "YouTube", infoVi.SourceValue)

		infoEn := flt.DescribeBlock("zalo.me", &Result{
			Reason:      FilteredBlockedService,
			ServiceName: "zalo",
			IsFiltered:  true,
		}, false)
		assert.Equal(t, "Zalo is currently blocked", infoEn.Heading)
		assert.Equal(t, "Zalo", infoEn.SourceValue)
	})

	t.Run("malware_list_and_multiple_sources", func(t *testing.T) {
		info := flt.DescribeBlock("flexytalk.net", &Result{
			Reason:     FilteredBlockList,
			IsFiltered: true,
			CanonName:  "cname.bad-tracker.example.",
			Rules: []*ResultRule{
				{FilterListID: 1, Text: "||bad-tracker.example^"},
				{FilterListID: 2, Text: "||flexytalk.net^$important"},
			},
		}, true)
		assert.True(t, info.IsWarning)
		assert.Equal(t, BlockCategoryMalware, info.Category)
		assert.Equal(t, "Trang web không an toàn", info.Heading)
		assert.Contains(t, info.ReasonValue, "Nguy cơ lừa đảo hoặc mã độc")
		assert.Contains(t, info.ReasonValue, "(ẩn qua CNAME)")
		assert.Equal(t, "AdGuard DNS filter (+1)", info.SourceValue)
		assert.Equal(t, []string{"AdGuard DNS filter", "uBlock filters – Badware risks"}, info.AllSources)
		assert.Equal(t, "cname.bad-tracker.example", info.CNAMEValue)
	})

	t.Run("tracker_domain_heuristic", func(t *testing.T) {
		info := flt.DescribeBlock("www.google-analytics.com", &Result{
			Reason:     FilteredBlockList,
			IsFiltered: true,
			Rules: []*ResultRule{
				{FilterListID: 1, Text: "||google-analytics.com^"},
			},
		}, true)
		assert.Equal(t, BlockCategoryTracker, info.Category)
		assert.Equal(t, "Đã chặn kết nối theo dõi", info.Heading)
		assert.Equal(t, "Theo dõi & thu thập dữ liệu", info.ReasonValue)
		assert.Equal(t, "Chặn theo tên miền gốc google-analytics.com và mọi tên miền phụ", info.ScopeValue)
	})

	t.Run("custom_rule_and_hosts_scope", func(t *testing.T) {
		info := flt.DescribeBlock("clix.vn", &Result{
			Reason:     FilteredBlockList,
			IsFiltered: true,
			Rules: []*ResultRule{
				{FilterListID: rulelist.APIIDCustom, Text: "0.0.0.0 clix.vn"},
			},
		}, true)
		assert.Equal(t, BlockCategoryCustom, info.Category)
		assert.Equal(t, "Quy tắc lọc tùy chỉnh", info.SourceValue)
		assert.Equal(t, "Chặn trực tiếp theo tên miền (quy tắc hosts)", info.ScopeValue)
		assert.Contains(t, info.ReasonValue, "quảng cáo / theo dõi")
	})

	t.Run("regex_and_important_scope", func(t *testing.T) {
		scopeVi := describeRuleScope("ad123.example.com", `/^ad[0-9]+\.example\.com$/$important`, true)
		assert.Equal(t, "Khớp theo mẫu biểu thức chính quy (Regex) · Quy tắc ưu tiên cao", scopeVi)

		scopeEn := describeRuleScope("ad123.example.com", `/^ad[0-9]+\.example\.com$/$important`, false)
		assert.Equal(t, "Matched via regular expression (Regex) · High-priority rule", scopeEn)
	})

	t.Run("describe_allowed", func(t *testing.T) {
		allowedPlain := flt.DescribeAllowed("example.org", &Result{Reason: NotFilteredNotFound}, true)
		assert.True(t, allowedPlain.IsAllowed)
		assert.Equal(t, "Tên miền này không bị chặn", allowedPlain.Heading)

		allowedList := flt.DescribeAllowed("tinhte.vn", &Result{
			Reason: NotFilteredAllowList,
			Rules: []*ResultRule{
				{FilterListID: 100, Text: "@@||tinhte.vn^$important"},
			},
		}, true)
		assert.True(t, allowedList.IsAllowed)
		assert.Equal(t, "Personal Allowlist", allowedList.SourceValue)
		assert.Equal(t, "@@||tinhte.vn^$important", allowedList.RuleValue)
		assert.Contains(t, allowedList.ScopeValue, "Cho phép tên miền tinhte.vn")
	})

	t.Run("describe_preview_scenarios", func(t *testing.T) {
		for _, sc := range []string{"ads", "tracker", "malware", "service", "parental", "custom"} {
			info := flt.DescribePreviewScenario(sc, "", true)
			assert.NotEmpty(t, info.Heading, "scenario %s heading", sc)
			assert.NotEmpty(t, info.ReasonValue, "scenario %s reason", sc)
			assert.NotEmpty(t, info.SourceValue, "scenario %s source", sc)
		}
	})
}
