package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

func composeAlertMessage(cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) string {
	lines := make([]string, 0, 16)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
	}

	lines = append(lines, fmt.Sprintf("<b>Alert: %s</b>", alertHeadline(metric)))
	lines = append(lines, "")
	lines = append(lines, "Metrics")
	lines = append(lines, fmt.Sprintf("  Metric: %s", metricDisplayName(metric)))
	lines = append(lines, fmt.Sprintf("  Current: <code>%s</code>", formatPercentage(value)))
	lines = append(lines, fmt.Sprintf("  Threshold: <code>%s</code>", formatPercentage(threshold)))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeRecoveryMessage formats a recovery notification.
func composeRecoveryMessage(metric string, currentValue, threshold float64, duration time.Duration) string {
	lines := []string{
		fmt.Sprintf("<b>%s has recovered</b>", metricDisplayName(metric)),
		"",
		fmt.Sprintf("  Current: <code>%s</code>", formatPercentage(currentValue)),
	}

	if threshold > 0 {
		lines = append(lines, fmt.Sprintf("  Threshold: <code>%s</code>", formatPercentage(threshold)))
	}

	lines = append(lines, fmt.Sprintf("  Alert duration: %s", duration.String()))
	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeFilterUpdateMessage(cfg TelegramConfig, update FilterUpdate, info systeminfo.Info) string {
	lines := make([]string, 0, 20)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
	}

	head := filterUpdateHeader(update.ListType)
	lines = append(lines, fmt.Sprintf("<b>%s</b>", head))
	lines = append(lines, fmt.Sprintf("  List: %s", fallbackString(update.Name)))
	if update.ID != 0 {
		lines = append(lines, fmt.Sprintf("  ID: #%s", formatUint64(update.ID)))
	}
	lines = append(lines, fmt.Sprintf("  Type: %s", filterTypeLabel(update.ListType)))
	if update.URL != "" {
		lines = append(lines, fmt.Sprintf("  Source: %s", update.URL))
	}
	rules := update.RulesCount
	if rules < 0 {
		rules = 0
	}
	lines = append(lines, fmt.Sprintf("  Rules: %s entries", formatInt64(int64(rules))))
	if update.BytesWritten > 0 {
		lines = append(lines, fmt.Sprintf("  Size: %s", formatBytesUint(uint64(update.BytesWritten))))
	}
	statusLabel := "Enabled"
	if !update.Enabled {
		statusLabel = "Disabled"
	}
	lines = append(lines, fmt.Sprintf("  Status: %s", statusLabel))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func alertHeadline(metric string) string {
	return fmt.Sprintf("%s exceeded threshold", metricDisplayName(metric))
}

func metricDisplayName(metric string) string {
	switch strings.ToLower(metric) {
	case "cpu":
		return "CPU usage"
	case "memory":
		return "Memory usage"
	case "disk":
		return "Disk usage"
	case "protection":
		return "DNS Protection"
	default:
		if metric == "" {
			return "Metric"
		}
		return strings.ToUpper(metric[:1]) + strings.ToLower(metric[1:])
	}
}

func filterUpdateHeader(listType FilterListType) string {
	switch listType {
	case FilterListTypeAllow:
		return "Allowlist Update"
	case FilterListTypeBlock:
		return "Blocklist Update"
	default:
		return "Filter Update"
	}
}

func filterTypeLabel(listType FilterListType) string {
	switch listType {
	case FilterListTypeAllow:
		return "Allowlist"
	case FilterListTypeBlock:
		return "Blocklist"
	default:
		return "Filter"
	}
}
