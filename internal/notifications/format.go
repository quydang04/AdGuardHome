package notifications

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func systemOverviewLines(info systeminfo.Info) []string {
	lines := []string{"<b>System Overview</b>"}
	lines = append(lines, fmt.Sprintf("  Hostname: %s", fallbackString(info.Hostname)))
	lines = append(lines, fmt.Sprintf("  OS: %s", formatOS(info)))
	if info.KernelVersion != "" {
		lines = append(lines, fmt.Sprintf("  Kernel: %s", info.KernelVersion))
	}
	lines = append(lines, fmt.Sprintf("  CPU: %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("  CPU Usage: <code>%s</code>", formatPercentage(info.CPUUsage)))
	lines = append(lines, fmt.Sprintf("  Memory: %s", formatUsage(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("  Memory Free: %s", formatCapacity(info.MemoryFree, info.MemoryTotal)))
	lines = append(lines, fmt.Sprintf("  Disk: %s", formatUsage(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	lines = append(lines, fmt.Sprintf("  Disk Free: %s", formatCapacity(info.DiskFree, info.DiskTotal)))
	lines = append(lines, fmt.Sprintf("  Disk Path: %s", fallbackString(info.DiskPath)))
	lines = append(lines, fmt.Sprintf("  Local IPs: %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("  Public IP: %s", fallbackString(info.PublicIP)))
	uptime := formatUptime(info.UptimeSeconds)
	if uptime == "" {
		uptime = "-"
	}
	lines = append(lines, fmt.Sprintf("  Uptime: %s", uptime))

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
		osLine = fmt.Sprintf("%s (%s)", osLine, arch)
	}

	return osLine
}

func formatCPU(info systeminfo.Info) string {
	name := strings.TrimSpace(info.CPUModel)
	if name == "" {
		name = "Unknown CPU"
	}
	if info.NumCPU > 0 {
		name = fmt.Sprintf("%s (%s cores)", name, formatInt64(int64(info.NumCPU)))
	}

	return name
}

func formatLocalIPs(ips []string) string {
	if len(ips) == 0 {
		return "-"
	}

	return strings.Join(ips, ", ")
}

func formatUsage(used, total uint64, usage float64) string {
	if total == 0 {
		return "-"
	}

	idx := chooseUnit(total)
	return fmt.Sprintf("%s / %s (%s)", formatBytesWithUnit(used, idx), formatBytesWithUnit(total, idx), formatPercentage(usage))
}

func formatCapacity(current, total uint64) string {
	if total == 0 {
		return "-"
	}

	idx := chooseUnit(total)
	return fmt.Sprintf("%s / %s", formatBytesWithUnit(current, idx), formatBytesWithUnit(total, idx))
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
		parts = append(parts, fmt.Sprintf("%dd", d))
	}

	if h > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}

	parts = append(parts, fmt.Sprintf("%dm", mn))

	return strings.Join(parts, " ")
}

// timestampLine returns a formatted timestamp line for message footers.
func timestampLine() string {
	now := time.Now()
	return fmt.Sprintf("Updated: %s", now.Format("15:04:05 02/01/2006"))
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
