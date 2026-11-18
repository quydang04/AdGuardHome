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

	lines = append(lines, fmt.Sprintf("🚨 <b>CẢNH BÁO · %s</b>", alertHeadline(metric)))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("   ├ Chỉ số:     %s", metricDisplayName(metric)))
	lines = append(lines, fmt.Sprintf("   ├ Hiện tại:   %s", usageBar(value)))
	lines = append(lines, fmt.Sprintf("   └ Ngưỡng:     <code>%s</code>", formatPercentage(threshold)))
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

	lines = append(lines, fmt.Sprintf("✅ <b>PHỤC HỒI · %s</b>", recoveryHeadline(metric)))
	lines = append(lines, divider())
	lines = append(lines, "")

	lines = append(lines, fmt.Sprintf("   ├ Chỉ số:      %s", metricDisplayName(metric)))
	if metric != "protection" {
		lines = append(lines, fmt.Sprintf("   ├ Hiện tại:    %s", usageBar(currentValue)))
		if threshold > 0 {
			lines = append(lines, fmt.Sprintf("   ├ Ngưỡng:      <code>%s</code>", formatPercentage(threshold)))
		}
	}

	lines = append(lines, fmt.Sprintf("   └ Thời gian:   <code>%s</code>", duration.String()))
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
	return fmt.Sprintf("%s đã trở lại bình thường", metricDisplayName(metric))
}

func composeProtectionAlertMessage(cfg TelegramConfig, info systeminfo.Info) string {
	lines := make([]string, 0, 24)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
		lines = append(lines, "")
	}

	lines = append(lines, "🚨 <b>CẢNH BÁO · Bảo vệ DNS đã bị TẮT!</b>")
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, "   🔴 Lọc DNS đang <b>tắt</b>.")
	lines = append(lines, "   <i>Mọi truy vấn sẽ đi thẳng không qua bộ lọc.</i>")
	lines = append(lines, "")

	if info.Hostname != "" {
		lines = append(lines, systemOverviewLines(info)...)
		lines = append(lines, "")
	}

	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeCertStatusMessage formats the current "SSL/TLS issue" (ACME) status
// for the bot menu.
func composeCertStatusMessage(status CertStatus) string {
	statusIcon := statusDot(status.Enabled)
	statusText := "TẮT"
	if status.Enabled {
		statusText = "BẬT"
	}

	autoRenewIcon := statusDot(status.AutoRenew)
	autoRenewText := "TẮT"
	if status.AutoRenew {
		autoRenewText = "BẬT"
	}

	lines := []string{
		"🔐 <b>Chứng chỉ SSL/TLS</b>",
		divider(),
		"",
		fmt.Sprintf("   %s Trạng thái:    <b>%s</b>", statusIcon, statusText),
		fmt.Sprintf("   ├ Tên miền:      %s", fallbackString(strings.Join(status.Domains, ", "))),
		fmt.Sprintf("   ├ Thử thách:     %s", fallbackString(status.Challenge)),
		fmt.Sprintf("   %s Tự động gia hạn: <b>%s</b>", autoRenewIcon, autoRenewText),
	}

	if !status.NotAfter.IsZero() {
		daysLeft := int(time.Until(status.NotAfter).Hours() / 24)
		expiryStr := fmt.Sprintf("<code>%s</code>", status.NotAfter.Format("02/01/2006 15:04"))
		if daysLeft > 0 {
			expiryStr += fmt.Sprintf(" · <i>còn %d ngày</i>", daysLeft)
		} else if daysLeft == 0 {
			expiryStr += " · <b>⚠️ hết hạn hôm nay</b>"
		} else {
			expiryStr += " · <b>🚨 ĐÃ HẾT HẠN</b>"
		}
		lines = append(lines, fmt.Sprintf("   ├ Hết hạn:       %s", expiryStr))
	}

	if !status.LastIssuedAt.IsZero() {
		lines = append(lines, fmt.Sprintf(
			"   └ Cấp phát:      <code>%s</code>",
			toLocal(status.LastIssuedAt).Format("02/01/2006 15:04"),
		))
	}

	if status.LastError != "" {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("   ⚠️ Lỗi: <code>%s</code>", status.LastError))
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
	lines = append(lines, fmt.Sprintf("   ├ Tên:      %s", fallbackString(update.Name)))
	if update.ID != 0 {
		lines = append(lines, fmt.Sprintf("   ├ ID:       <code>#%s</code>", formatUint64(update.ID)))
	}
	lines = append(lines, fmt.Sprintf("   ├ Loại:     %s", filterTypeLabel(update.ListType)))
	if update.URL != "" {
		lines = append(lines, fmt.Sprintf("   ├ Nguồn:    <code>%s</code>", update.URL))
	}

	rules := update.RulesCount
	if rules < 0 {
		rules = 0
	}
	lines = append(lines, fmt.Sprintf("   ├ Quy tắc:  <code>%s</code>", formatInt64(int64(rules))))
	if update.BytesWritten > 0 {
		lines = append(lines, fmt.Sprintf("   ├ Dung lượng: <code>%s</code>", formatBytesUint(uint64(update.BytesWritten))))
	}

	statusIcon := statusDot(update.Enabled)
	statusLabel := "Đang hoạt động"
	if !update.Enabled {
		statusLabel = "Đã tắt"
	}
	lines = append(lines, fmt.Sprintf("   └ Trạng thái: %s %s", statusIcon, statusLabel))
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

	lines = append(lines, fmt.Sprintf("%s <b>CHỨNG CHỈ SẮP HẾT HẠN</b>", icon))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("   ├ Tên miền:   %s", fallbackString(strings.Join(ev.Domains, ", "))))
	lines = append(lines, fmt.Sprintf("   ├ Hết hạn:    <code>%s</code>", ev.NotAfter.Format("02/01/2006 15:04")))
	lines = append(lines, fmt.Sprintf("   └ Còn lại:    <code>%d ngày</code>", ev.DaysLeft))
	lines = append(lines, "")
	lines = append(lines, "<i>Vui lòng gia hạn thủ công trong cài đặt mã hóa.</i>")
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
		lines = append(lines, "🚨 <b>GIA HẠN TỰ ĐỘNG THẤT BẠI</b>")
		lines = append(lines, divider())
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("   ├ Tên miền: %s", fallbackString(strings.Join(ev.Domains, ", "))))
		lines = append(lines, fmt.Sprintf("   └ Lỗi:      <code>%s</code>", ev.Err.Error()))
		lines = append(lines, "")
		lines = append(lines, "<i>Vui lòng gia hạn thủ công trong cài đặt mã hóa.</i>")
	} else {
		lines = append(lines, "✅ <b>ĐÃ GIA HẠN TỰ ĐỘNG</b>")
		lines = append(lines, divider())
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("   ├ Tên miền: %s", fallbackString(strings.Join(ev.Domains, ", "))))
		lines = append(lines, fmt.Sprintf("   └ Hết hạn:  <code>%s</code>", ev.NotAfter.Format("02/01/2006 15:04")))
	}

	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func alertHeadline(metric string) string {
	return fmt.Sprintf("%s vượt ngưỡng", metricDisplayName(metric))
}

func metricDisplayName(metric string) string {
	switch strings.ToLower(metric) {
	case "cpu":
		return "CPU"
	case "memory":
		return "Bộ nhớ"
	case "disk":
		return "Ổ đĩa"
	case "protection":
		return "Bảo vệ DNS"
	default:
		if metric == "" {
			return "Chỉ số"
		}
		return strings.ToUpper(metric[:1]) + strings.ToLower(metric[1:])
	}
}

func filterUpdateHeader(listType FilterListType) string {
	switch listType {
	case FilterListTypeAllow:
		return "Đã cập nhật DS cho phép"
	case FilterListTypeBlock:
		return "Đã cập nhật DS chặn"
	default:
		return "Đã cập nhật bộ lọc"
	}
}

func filterTypeLabel(listType FilterListType) string {
	switch listType {
	case FilterListTypeAllow:
		return "Cho phép"
	case FilterListTypeBlock:
		return "Chặn"
	default:
		return "Bộ lọc"
	}
}
