package notifications

import (
	"fmt"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

func composeAlertMessage(cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) string {
	lines := make([]string, 0, 20)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	lines = append(lines, fmt.Sprintf("🚨 <b>ALERT: %s</b>", alertHeadline(metric)))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, sectionHeader("📈", "Metrics"))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Metric:</b>    %s", metricDisplayName(metric)))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Current:</b>   %s", usageBar(value)))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Threshold:</b> <code>%s</code>", formatPercentage(threshold)))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeRecoveryMessage formats a recovery notification.
func composeRecoveryMessage(cfg TelegramConfig, metric string, currentValue, threshold float64, duration time.Duration, info systeminfo.Info) string {
	lines := make([]string, 0, 24)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	lines = append(lines, fmt.Sprintf("✅ <b>RECOVERY: %s</b>", recoveryHeadline(metric)))
	lines = append(lines, divider())
	lines = append(lines, "")

	lines = append(lines, sectionHeader("📈", "Metrics"))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Metric:</b>         %s", metricDisplayName(metric)))
	if metric != "protection" {
		lines = append(lines, fmt.Sprintf("  ▸ <b>Current:</b>        %s", usageBar(currentValue)))
		if threshold > 0 {
			lines = append(lines, fmt.Sprintf("  ▸ <b>Threshold:</b>      <code>%s</code>", formatPercentage(threshold)))
		}
	}

	lines = append(lines, fmt.Sprintf("  ▸ <b>Alert Duration:</b> <code>%s</code>", duration.String()))
	lines = append(lines, "")

	if info.Hostname != "" {
		lines = append(lines, systemOverviewLines(info)...)
		lines = append(lines, "")
	}

	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func recoveryHeadline(metric string) string {
	return fmt.Sprintf("%s back to normal", metricDisplayName(metric))
}

func composeProtectionAlertMessage(cfg TelegramConfig, info systeminfo.Info) string {
	lines := make([]string, 0, 24)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	lines = append(lines, "🚨 <b>ALERT: DNS Protection is DISABLED!</b>")
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, "🔴 DNS filtering is currently <b>turned off</b>.")
	lines = append(lines, "<i>All queries pass through unfiltered.</i>")
	lines = append(lines, "")

	if info.Hostname != "" {
		lines = append(lines, systemOverviewLines(info)...)
		lines = append(lines, "")
	}

	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeYouTubeAlertMessage formats an alert that the YouTube ad-blocking
// route server has no healthy IPs left.
func composeYouTubeAlertMessage(cfg TelegramConfig, status YouTubeStatus, info systeminfo.Info) string {
	lines := make([]string, 0, 16)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	lines = append(lines, "🚨 <b>ALERT: YouTube Blocking route server unreachable!</b>")
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  ▸ <b>Healthy IPs:</b> <code>%d</code> / <code>%d</code>", status.HealthyIPs, status.TotalIPs))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Last sync:</b>   %s", fallbackString(status.LastSyncStatus)))
	lines = append(lines, "")
	lines = append(lines, "<i>DNS rewrites routing YouTube traffic may stop working until the route server recovers.</i>")
	lines = append(lines, "")

	if info.Hostname != "" {
		lines = append(lines, systemOverviewLines(info)...)
		lines = append(lines, "")
	}

	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeYouTubeStatusMessage(status YouTubeStatus) string {
	statusIcon, statusText := "🔴", "DISABLED"
	if status.Enabled {
		if status.Active {
			statusIcon, statusText = "🟢", "ACTIVE"
		} else {
			statusIcon, statusText = "🟡", "ENABLED (not started)"
		}
	}

	lines := []string{
		"📺 <b>YouTube Blocking</b>",
		divider(),
		"",
		fmt.Sprintf("  %s <b>Status:</b>          %s", statusIcon, statusText),
		fmt.Sprintf("  ▸ <b>Healthy IPs:</b>    <code>%d</code> / <code>%d</code>", status.HealthyIPs, status.TotalIPs),
		fmt.Sprintf("  ▸ <b>Blocked rules:</b>  <code>%d</code>", status.BlockedRules),
		fmt.Sprintf("  ▸ <b>Active rewrites:</b> <code>%d</code>", status.ActiveRewrites),
		fmt.Sprintf("  ▸ <b>Last sync:</b>      %s", fallbackString(status.LastSyncStatus)),
	}

	if !status.LastSyncTime.IsZero() {
		lines = append(lines, fmt.Sprintf("  ▸ <b>Last sync at:</b>   <code>%s</code>", toLocal(status.LastSyncTime).Format("15:04:05 02/01/2006")))
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeCertStatusMessage formats the current "SSL/TLS issue" (ACME) status
// for the bot menu.
func composeCertStatusMessage(status CertStatus) string {
	statusIcon, statusText := "🔴", "DISABLED"
	if status.Enabled {
		statusIcon, statusText = "🟢", "ENABLED"
	}

	autoRenewIcon, autoRenewText := "🔴", "OFF"
	if status.AutoRenew {
		autoRenewIcon, autoRenewText = "🟢", "ON"
	}

	lines := []string{
		"🔒 <b>SSL/TLS Issue</b>",
		divider(),
		"",
		fmt.Sprintf("  %s <b>Status:</b>      %s", statusIcon, statusText),
		fmt.Sprintf("  ▸ <b>Domains:</b>     %s", fallbackString(strings.Join(status.Domains, ", "))),
		fmt.Sprintf("  ▸ <b>Challenge:</b>   %s", fallbackString(status.Challenge)),
		fmt.Sprintf("  %s <b>Auto-renew:</b>  %s", autoRenewIcon, autoRenewText),
	}

	if !status.NotAfter.IsZero() {
		lines = append(lines, fmt.Sprintf("  ▸ <b>Expires:</b>     <code>%s</code>", status.NotAfter.Format(time.RFC1123)))
	}

	if !status.LastIssuedAt.IsZero() {
		lines = append(lines, fmt.Sprintf(
			"  ▸ <b>Last issued:</b> <code>%s</code>",
			toLocal(status.LastIssuedAt).Format("15:04:05 02/01/2006"),
		))
	}

	if status.LastError != "" {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("  ⚠️ <b>Last error:</b> <code>%s</code>", status.LastError))
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeFilterUpdateMessage(cfg TelegramConfig, update FilterUpdate, info systeminfo.Info) string {
	lines := make([]string, 0, 24)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	head := filterUpdateHeader(update.ListType)
	lines = append(lines, fmt.Sprintf("🔄 <b>%s</b>", head))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, sectionHeader("📋", "List Details"))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Name:</b>   %s", fallbackString(update.Name)))
	if update.ID != 0 {
		lines = append(lines, fmt.Sprintf("  ▸ <b>ID:</b>     <code>#%s</code>", formatUint64(update.ID)))
	}
	lines = append(lines, fmt.Sprintf("  ▸ <b>Type:</b>   %s", filterTypeLabel(update.ListType)))
	if update.URL != "" {
		lines = append(lines, fmt.Sprintf("  ▸ <b>Source:</b> <code>%s</code>", update.URL))
	}

	rules := update.RulesCount
	if rules < 0 {
		rules = 0
	}
	lines = append(lines, fmt.Sprintf("  ▸ <b>Rules:</b>  <code>%s</code> entries", formatInt64(int64(rules))))
	if update.BytesWritten > 0 {
		lines = append(lines, fmt.Sprintf("  ▸ <b>Size:</b>   <code>%s</code>", formatBytesUint(uint64(update.BytesWritten))))
	}

	statusIcon := "✅"
	statusLabel := "Enabled"
	if !update.Enabled {
		statusIcon = "🚫"
		statusLabel = "Disabled"
	}
	lines = append(lines, fmt.Sprintf("  ▸ <b>Status:</b> %s %s", statusIcon, statusLabel))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeCertExpiryMessage formats a reminder that a certificate is nearing
// expiration and should be renewed manually.
func composeCertExpiryMessage(cfg TelegramConfig, ev CertExpiryReminder, info systeminfo.Info) string {
	lines := make([]string, 0, 16)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	icon := "⏰"
	if ev.DaysLeft <= 0 {
		icon = "🚨"
	}

	lines = append(lines, fmt.Sprintf("%s <b>TLS CERTIFICATE EXPIRING</b>", icon))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, sectionHeader("🔐", "Certificate"))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Domains:</b>    %s", fallbackString(strings.Join(ev.Domains, ", "))))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Expires:</b>    <code>%s</code>", ev.NotAfter.Format(time.RFC1123)))
	lines = append(lines, fmt.Sprintf("  ▸ <b>Days left:</b>  <code>%d</code>", ev.DaysLeft))
	lines = append(lines, "")
	lines = append(lines, "Renew it manually in AdGuard Home's encryption settings.")
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeCertRenewalMessage formats a notification about the outcome of an
// automatic ACME certificate renewal.
func composeCertRenewalMessage(cfg TelegramConfig, ev CertRenewalResult, info systeminfo.Info) string {
	lines := make([]string, 0, 16)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	if ev.Err != nil {
		lines = append(lines, "🚨 <b>TLS CERTIFICATE AUTO-RENEWAL FAILED</b>")
		lines = append(lines, divider())
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🔐", "Certificate"))
		lines = append(lines, fmt.Sprintf("  ▸ <b>Domains:</b> %s", fallbackString(strings.Join(ev.Domains, ", "))))
		lines = append(lines, fmt.Sprintf("  ▸ <b>Error:</b>   <code>%s</code>", ev.Err.Error()))
		lines = append(lines, "")
		lines = append(lines, "Renew it manually in AdGuard Home's encryption settings.")
	} else {
		lines = append(lines, "✅ <b>TLS CERTIFICATE AUTO-RENEWED</b>")
		lines = append(lines, divider())
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🔐", "Certificate"))
		lines = append(lines, fmt.Sprintf("  ▸ <b>Domains:</b>     %s", fallbackString(strings.Join(ev.Domains, ", "))))
		lines = append(lines, fmt.Sprintf("  ▸ <b>New expiry:</b>  <code>%s</code>", ev.NotAfter.Format(time.RFC1123)))
	}

	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func alertHeadline(metric string) string {
	return fmt.Sprintf("%s exceeded threshold", metricDisplayName(metric))
}

func metricDisplayName(metric string) string {
	switch strings.ToLower(metric) {
	case "cpu":
		return "CPU Usage"
	case "memory":
		return "Memory Usage"
	case "disk":
		return "Disk Usage"
	case "protection":
		return "DNS Protection"
	case "youtube_health":
		return "YouTube Blocking"
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
		return "Allowlist Updated"
	case FilterListTypeBlock:
		return "Blocklist Updated"
	default:
		return "Filter Updated"
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
