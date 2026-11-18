package notifications

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

// sectionHeader returns a styled section header with a line separator.
func sectionHeader(icon, title string) string {
	return fmt.Sprintf("%s <b>%s</b>", icon, title)
}

// divider returns a clean visual separator.
func divider() string {
	return "━━━━━━━━━━━━━━━━━━━━"
}

// thinDivider returns a thinner separator for sub-sections.
func thinDivider() string {
	return "┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄┄"
}

// formatProgressBar renders a modern gradient progress bar.
// width is the number of filled + empty cells, percentage 0–100.
func formatProgressBar(pct float64, width int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	filled := int(math.Round(pct / 100 * float64(width)))
	bar := strings.Repeat("▰", filled) + strings.Repeat("▱", width-filled)
	return bar
}

// usageBar returns a progress bar + percentage with color indicator.
func usageBar(pct float64) string {
	bar := formatProgressBar(pct, 12)
	indicator := levelIndicator(pct)
	return fmt.Sprintf("%s %s <code>%s</code>", indicator, bar, formatPercentage(pct))
}

// levelIndicator returns a colored circle based on the level.
func levelIndicator(pct float64) string {
	switch {
	case pct >= 90:
		return "🔴"
	case pct >= 70:
		return "🟡"
	default:
		return "🟢"
	}
}

func systemOverviewLines(info systeminfo.Info) []string {
	lines := []string{sectionHeader("💻", "Hệ thống")}
	lines = append(lines, fmt.Sprintf("   ├ Máy chủ: <code>%s</code>", fallbackString(info.Hostname)))
	lines = append(lines, fmt.Sprintf("   ├ OS: %s", formatOS(info)))
	if info.KernelVersion != "" {
		lines = append(lines, fmt.Sprintf("   ├ Kernel: <code>%s</code>", info.KernelVersion))
	}
	lines = append(lines, fmt.Sprintf("   ├ CPU: %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("   ├ CPU: %s", usageBar(info.CPUUsage)))
	lines = append(lines, fmt.Sprintf("   ├ RAM: %s", formatUsageWithBar(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("   ├ Ổ đĩa: %s", formatUsageWithBar(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	lines = append(lines, fmt.Sprintf("   ├ IP nội bộ: %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("   └ IP công cộng: <code>%s</code>", fallbackString(info.PublicIP)))
	if info.SystemTime != "" {
		if t, err := time.Parse(time.RFC3339, info.SystemTime); err == nil {
			lines = append(lines, fmt.Sprintf("   ⏱ <code>%s</code> · hoạt động %s", toLocal(t).Format("15:04:05 02/01/2006"), formatUptime(info.UptimeSeconds)))
		}
	}

	return lines
}

func formatOS(info systeminfo.Info) string {
	osLine := strings.TrimSpace(info.OSVersion)
	if osLine == "" {
		osLine = strings.TrimSpace(info.OS)
	}
	if osLine == "" {
		osLine = "-"
	}
	if arch := strings.TrimSpace(info.Arch); arch != "" {
		osLine = fmt.Sprintf("%s <code>(%s)</code>", osLine, arch)
	}

	return osLine
}

func formatCPU(info systeminfo.Info) string {
	name := strings.TrimSpace(info.CPUModel)
	if name == "" {
		name = "Không xác định"
	}
	if info.NumCPU > 0 {
		name = fmt.Sprintf("%s <code>(%s nhân)</code>", name, formatInt64(int64(info.NumCPU)))
	}

	return name
}

func formatLocalIPs(ips []string) string {
	if len(ips) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(ips))
	for _, ip := range ips {
		parts = append(parts, "<code>"+ip+"</code>")
	}

	return strings.Join(parts, ", ")
}

func formatUsage(used, total uint64, usage float64) string {
	if total == 0 {
		return "-"
	}

	idx := chooseUnit(total)
	return fmt.Sprintf("<code>%s / %s</code> (%s)", formatBytesWithUnit(used, idx), formatBytesWithUnit(total, idx), formatPercentage(usage))
}

// formatUsageWithBar adds a progress bar alongside usage numbers.
func formatUsageWithBar(used, total uint64, usage float64) string {
	if total == 0 {
		return "-"
	}

	idx := chooseUnit(total)
	bar := usageBar(usage)
	return fmt.Sprintf("%s <code>%s/%s</code>", bar, formatBytesWithUnit(used, idx), formatBytesWithUnit(total, idx))
}

func formatCapacity(current, total uint64) string {
	if total == 0 {
		return "-"
	}

	idx := chooseUnit(total)
	return fmt.Sprintf("<code>%s / %s</code>", formatBytesWithUnit(current, idx), formatBytesWithUnit(total, idx))
}

func formatBytesUint(value uint64) string {
	idx := chooseUnit(value)
	return formatBytesWithUnit(value, idx)
}

func formatBytesWithUnit(value uint64, idx int) string {
	if idx < 0 {
		idx = 0
	} else if idx >= len(byteUnits) {
		idx = len(byteUnits) - 1
	}

	unit := byteUnits[idx]
	if idx == 0 {
		return fmt.Sprintf("%s %s", formatInt64(int64(value)), unit)
	}

	div := math.Pow(1024, float64(idx))
	val := float64(value) / div
	return fmt.Sprintf("%s %s", formatFloat(val), unit)
}

func chooseUnit(value uint64) int {
	idx := 0
	for value >= 1024 && idx < len(byteUnits)-1 {
		value /= 1024
		idx++
	}

	return idx
}

func formatFloat(v float64) string {
	formatted := fmt.Sprintf("%.1f", v)
	formatted = strings.TrimRight(formatted, "0")
	formatted = strings.TrimSuffix(formatted, ".")
	if formatted == "" {
		return "0"
	}

	return formatted
}

func formatPercentage(value float64) string {
	if math.IsNaN(value) || math.IsInf(value, 0) {
		return "-"
	}
	if value < 0 {
		value = 0
	}

	return fmt.Sprintf("%s%%", formatFloat(value))
}

func formatInt64(val int64) string {
	neg := val < 0
	if neg {
		val = -val
	}

	return formatIntegerString(strconv.FormatInt(val, 10), neg)
}

func formatUint64(val uint64) string {
	return formatIntegerString(strconv.FormatUint(val, 10), false)
}

func formatIntegerString(s string, negative bool) string {
	if len(s) <= 3 {
		if negative {
			return "-" + s
		}

		return s
	}

	parts := make([]string, 0, (len(s)+2)/3)
	for len(s) > 3 {
		parts = append(parts, s[len(s)-3:])
		s = s[:len(s)-3]
	}
	if s != "" {
		parts = append(parts, s)
	}

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	result := strings.Join(parts, ",")
	if negative {
		return "-" + result
	}

	return result
}

func fallbackString(val string) string {
	val = strings.TrimSpace(val)
	if val == "" {
		return "-"
	}

	return val
}

func formatUptime(seconds uint64) string {
	if seconds == 0 {
		return ""
	}

	d := seconds / 86400
	h := (seconds % 86400) / 3600
	mn := (seconds % 3600) / 60

	parts := make([]string, 0, 3)
	if d > 0 {
		parts = append(parts, fmt.Sprintf("%d ngày", d))
	}

	if h > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%d giờ", h))
	}

	parts = append(parts, fmt.Sprintf("%d phút", mn))

	return strings.Join(parts, " ")
}

// systemLocation reads the current system timezone from OS configuration,
// bypassing Go's cached time.Local which is set once at process startup.
func systemLocation() *time.Location {
	if data, err := os.ReadFile("/etc/timezone"); err == nil {
		if name := strings.TrimSpace(string(data)); name != "" {
			if loc, err := time.LoadLocation(name); err == nil {
				return loc
			}
		}
	}

	if target, err := os.Readlink("/etc/localtime"); err == nil {
		if idx := strings.Index(target, "zoneinfo/"); idx >= 0 {
			name := target[idx+len("zoneinfo/"):]
			if loc, err := time.LoadLocation(name); err == nil {
				return loc
			}
		}
	}

	return time.Local
}

// localNow returns the current time in the system's configured timezone.
func localNow() time.Time {
	return time.Now().In(systemLocation())
}

// toLocal converts a time value to the system's configured timezone.
func toLocal(t time.Time) time.Time {
	return t.In(systemLocation())
}

// timestampLine returns a formatted timestamp line for message footers.
func timestampLine() string {
	now := localNow()
	return fmt.Sprintf("🕐 <i>Cập nhật: %s</i>", now.Format("15:04:05 · 02/01/2006"))
}

// capitalizeFirst returns s with its first letter uppercased.
func capitalizeFirst(s string) string {
	if s == "" {
		return ""
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// formatRate formats a per-second byte rate for display.
func formatRate(bytesPerSec uint64) string {
	if bytesPerSec == 0 {
		return "0 B/s"
	}

	return formatBytesUint(bytesPerSec) + "/s"
}

// formatDuration formats a duration into a human-friendly string.
func formatDuration(d time.Duration) string {
	if d < time.Millisecond {
		return fmt.Sprintf("%.0fµs", float64(d.Microseconds()))
	}
	if d < time.Second {
		return fmt.Sprintf("%.1fms", float64(d.Microseconds())/1000)
	}

	return d.Truncate(time.Millisecond).String()
}

// statusDot returns a colored status indicator.
func statusDot(ok bool) string {
	if ok {
		return "🟢"
	}

	return "🔴"
}
