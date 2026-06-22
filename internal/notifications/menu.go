package notifications

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

func mainMenuKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "System Status", CallbackData: "cmd:status"},
				{Text: "DNS Stats", CallbackData: "cmd:stats"},
			},
			{
				{Text: "Filters", CallbackData: "cmd:filters"},
				{Text: "Protection", CallbackData: "cmd:protection"},
			},
			{
				{Text: "Recent Logs", CallbackData: "cmd:logs"},
				{Text: "Process Info", CallbackData: "cmd:processes"},
			},
			{
				{Text: "Filter Manage", CallbackData: "cmd:filtermgr"},
			},
		},
	}
}

func backToMenuKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "Back to Menu", CallbackData: "cmd:menu"},
			},
		},
	}
}

// protectionKeyboard returns a keyboard with a toggle button based on current
// protection state.
func protectionKeyboard(enabled bool) *tgInlineKeyboardMarkup {
	var toggleBtn tgInlineKeyboardButton
	if enabled {
		toggleBtn = tgInlineKeyboardButton{Text: "Disable Protection", CallbackData: "cmd:protection_off"}
	} else {
		toggleBtn = tgInlineKeyboardButton{Text: "Enable Protection", CallbackData: "cmd:protection_on"}
	}

	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{toggleBtn},
			{{Text: "Back to Menu", CallbackData: "cmd:menu"}},
		},
	}
}

func filterManageKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "Add Blocklist", CallbackData: "cmd:filtermgr_addblock"},
				{Text: "Add Allowlist", CallbackData: "cmd:filtermgr_addallow"},
			},
			{
				{Text: "Remove Blocklist", CallbackData: "cmd:filtermgr_rmblock"},
				{Text: "Remove Allowlist", CallbackData: "cmd:filtermgr_rmallow"},
			},
			{
				{Text: "Enable List", CallbackData: "cmd:filtermgr_enable"},
				{Text: "Disable List", CallbackData: "cmd:filtermgr_disable"},
			},
			{
				{Text: "Update All Lists", CallbackData: "cmd:filtermgr_update"},
			},
			{
				{Text: "Back to Menu", CallbackData: "cmd:menu"},
			},
		},
	}
}

func composeFilterManageMessage() string {
	lines := []string{
		"<b>Filter Management</b>",
		"---",
		"Manage your filter lists using the buttons below.",
		"",
		"<b>Add:</b> Add new blocklist or allowlist by URL",
		"<b>Remove:</b> Select a list to remove",
		"<b>Enable/Disable:</b> Toggle lists on/off",
		"<b>Update:</b> Check all lists for updates",
		"",
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterSelectionMessage(title, action, listType string, blockLists, allowLists []FilterListInfo) (string, *tgInlineKeyboardMarkup) {
	var lines []string
	lines = append(lines, fmt.Sprintf("<b>%s</b>", title))
	lines = append(lines, "---")
	lines = append(lines, "Select a list:")

	buttons := make([][]tgInlineKeyboardButton, 0)

	addListButtons := func(lists []FilterListInfo, lt string, filterEnabled *bool) {
		for i, fl := range lists {
			if filterEnabled != nil && fl.Enabled != *filterEnabled {
				continue
			}

			name := fl.Name
			if name == "" {
				name = "Unnamed"
			}
			if len(name) > 30 {
				name = name[:27] + "..."
			}

			status := ""
			if action == "en" || action == "dis" {
				if fl.Enabled {
					status = " [ON]"
				} else {
					status = " [OFF]"
				}
			}

			label := fmt.Sprintf("%s%s", name, status)
			cb := fmt.Sprintf("flt:%s:%s:%d", action, lt, i)
			buttons = append(buttons, []tgInlineKeyboardButton{{Text: label, CallbackData: cb}})
		}
	}

	switch action {
	case "rm":
		if listType == "a" {
			if len(allowLists) == 0 {
				lines = append(lines, "")
				lines = append(lines, "No allowlists configured.")
			} else {
				addListButtons(allowLists, "a", nil)
			}
		} else {
			if len(blockLists) == 0 {
				lines = append(lines, "")
				lines = append(lines, "No blocklists configured.")
			} else {
				addListButtons(blockLists, "b", nil)
			}
		}
	case "en":
		showDisabled := false
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "<b>Blocklists:</b>")
			addListButtons(blockLists, "b", &showDisabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "<b>Allowlists:</b>")
			addListButtons(allowLists, "a", &showDisabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "All lists are already enabled.")
		}
	case "dis":
		showEnabled := true
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "<b>Blocklists:</b>")
			addListButtons(blockLists, "b", &showEnabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "<b>Allowlists:</b>")
			addListButtons(allowLists, "a", &showEnabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "All lists are already disabled.")
		}
	}

	buttons = append(buttons, []tgInlineKeyboardButton{{Text: "Back to Filter Manage", CallbackData: "cmd:filtermgr"}})

	text := strings.Join(lines, "\n")
	kb := &tgInlineKeyboardMarkup{InlineKeyboard: buttons}

	return text, kb
}

func composeFilterManageHelpMessage(action string) string {
	switch action {
	case "addblock":
		return "<b>Add Blocklist</b>\n---\nSend a message with the command:\n\n<code>/addlist &lt;url&gt;</code>\n\nMultiple URLs:\n<code>/addlist url1 url2 url3</code>\n\nWith custom name:\n<code>/addlist url | name</code>"
	case "addallow":
		return "<b>Add Allowlist</b>\n---\nSend a message with the command:\n\n<code>/addallow &lt;url&gt;</code>\n\nMultiple URLs:\n<code>/addallow url1 url2 url3</code>\n\nWith custom name:\n<code>/addallow url | name</code>"
	case "rmblock":
		return "<b>Remove Blocklist</b>\n---\nSend a message with the command:\n\n<code>/removelist &lt;url&gt;</code>"
	case "rmallow":
		return "<b>Remove Allowlist</b>\n---\nSend a message with the command:\n\n<code>/removeallow &lt;url&gt;</code>"
	case "enable":
		return "<b>Enable List</b>\n---\nSend a message with the command:\n\n<code>/enablelist &lt;url&gt;</code>"
	case "disable":
		return "<b>Disable List</b>\n---\nSend a message with the command:\n\n<code>/disablelist &lt;url&gt;</code>"
	default:
		return "Unknown action"
	}
}

func composeMainMenuMessage(info systeminfo.Info, protectionOn bool) string {
	protStatus := "ON"
	if !protectionOn {
		protStatus = "OFF"
	}

	lines := []string{
		"<b>AdGuard Home Menu</b>",
		"---",
		fmt.Sprintf("Host: %s", fallbackString(info.Hostname)),
		fmt.Sprintf("Protection: <b>%s</b>", protStatus),
		fmt.Sprintf("CPU: <code>%s</code>", formatPercentage(info.CPUUsage)),
		fmt.Sprintf("Memory: %s", formatUsage(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)),
	}

	// Swap info if available.
	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("Swap: %s", formatUsage(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	lines = append(lines, "")
	lines = append(lines, "Select an option below:")
	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeSystemStatusMessage(info systeminfo.Info, ioStats IOStats) string {
	lines := []string{
		"<b>System Status</b>",
		"---",
		fmt.Sprintf("Hostname: %s", fallbackString(info.Hostname)),
		fmt.Sprintf("OS: %s", formatOS(info)),
	}

	if info.KernelVersion != "" {
		lines = append(lines, fmt.Sprintf("Kernel: %s", info.KernelVersion))
	}

	if info.VirtPlatform != "" {
		lines = append(lines, fmt.Sprintf("Platform: %s (virtualized)", info.VirtPlatform))
	}

	uptime := formatUptime(info.UptimeSeconds)
	if uptime == "" {
		uptime = "-"
	}
	lines = append(lines, fmt.Sprintf("Uptime: %s", uptime))

	if info.BootTime > 0 {
		bootT := time.Unix(int64(info.BootTime), 0)
		lines = append(lines, fmt.Sprintf("Boot time: %s", bootT.Format("2006-01-02 15:04")))
	}

	lines = append(lines, "")

	// CPU section.
	lines = append(lines, fmt.Sprintf("<b>CPU</b>: %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("  Usage: <code>%s</code>", formatPercentage(info.CPUUsage)))

	// Load average (non-zero means non-Windows).
	if info.LoadAvg1 > 0 || info.LoadAvg5 > 0 || info.LoadAvg15 > 0 {
		lines = append(lines, fmt.Sprintf("  Load: %.2f / %.2f / %.2f (1/5/15 min)", info.LoadAvg1, info.LoadAvg5, info.LoadAvg15))
	}

	lines = append(lines, "")

	// Memory section.
	lines = append(lines, "<b>Memory</b>")
	lines = append(lines, fmt.Sprintf("  RAM: %s", formatUsage(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("  Free: %s", formatCapacity(info.MemoryFree, info.MemoryTotal)))

	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("  Swap: %s", formatUsage(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	lines = append(lines, "")

	// Disk section.
	lines = append(lines, "<b>Disk</b>")
	if len(info.AllDisks) > 0 {
		for _, d := range info.AllDisks {
			fsLabel := ""
			if d.Filesystem != "" {
				fsLabel = fmt.Sprintf(" [%s]", d.Filesystem)
			}
			lines = append(lines, fmt.Sprintf("  %s: %s%s", d.Path, formatUsage(d.Used, d.Total, d.UsagePercent), fsLabel))
		}
	} else {
		lines = append(lines, fmt.Sprintf("  %s: %s", fallbackString(info.DiskPath), formatUsage(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	}

	// Disk I/O rates.
	if ioStats.DiskReadBytesPerSec > 0 || ioStats.DiskWriteBytesPerSec > 0 {
		lines = append(lines, fmt.Sprintf("  Read: %s  Write: %s", formatRate(ioStats.DiskReadBytesPerSec), formatRate(ioStats.DiskWriteBytesPerSec)))
	}

	lines = append(lines, "")

	// Network section.
	lines = append(lines, "<b>Network</b>")
	lines = append(lines, fmt.Sprintf("  Local IPs: %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("  Public IP: %s", fallbackString(info.PublicIP)))

	if ioStats.NetBytesRecvPerSec > 0 || ioStats.NetBytesSentPerSec > 0 {
		lines = append(lines, fmt.Sprintf("  Recv: %s  Send: %s", formatRate(ioStats.NetBytesRecvPerSec), formatRate(ioStats.NetBytesSentPerSec)))
	}

	if info.ActiveConns > 0 {
		lines = append(lines, fmt.Sprintf("  TCP connections: %d", info.ActiveConns))
	}

	if info.NetErrorsIn > 0 || info.NetErrorsOut > 0 {
		lines = append(lines, fmt.Sprintf("  Errors in: %d  out: %d", info.NetErrorsIn, info.NetErrorsOut))
	}

	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeDNSStatsMessage(numQueries, numBlocked, numSafeBrowsing, numParental uint64, avgMs float64) string {
	avgMsStr := fmt.Sprintf("%s ms", formatFloat(avgMs*1000))

	lines := []string{
		"<b>DNS Statistics</b>",
		"---",
		fmt.Sprintf("Total Queries: <code>%s</code>", formatUint64(numQueries)),
		fmt.Sprintf("Blocked: <code>%s</code>", formatUint64(numBlocked)),
		fmt.Sprintf("Safe Browsing: %s", formatUint64(numSafeBrowsing)),
		fmt.Sprintf("Parental: %s", formatUint64(numParental)),
		fmt.Sprintf("Avg Response: %s", avgMsStr),
	}

	if numQueries > 0 {
		pct := float64(numBlocked) / float64(numQueries) * 100
		lines = append(lines, fmt.Sprintf("Block Rate: <code>%s</code>", formatPercentage(pct)))
	}

	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeFilterMessage(totalRules, enabledBlockLists, enabledAllowLists int) string {
	lines := []string{
		"<b>Filter Information</b>",
		"---",
		fmt.Sprintf("Total Rules: <code>%s</code>", formatInt64(int64(totalRules))),
		fmt.Sprintf("Block Lists: %d enabled", enabledBlockLists),
		fmt.Sprintf("Allow Lists: %d enabled", enabledAllowLists),
		"",
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterDetailedMessage(totalRules int, blockLists, allowLists []FilterListInfo) string {
	lines := []string{
		"<b>Filter Information</b>",
		"---",
		fmt.Sprintf("Total Rules: <code>%s</code>", formatInt64(int64(totalRules))),
	}

	// Block lists section.
	enabledBlock := 0
	for _, bl := range blockLists {
		if bl.Enabled {
			enabledBlock++
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("<b>Block Lists</b> (%d/%d enabled)", enabledBlock, len(blockLists)))

	for _, bl := range blockLists {
		status := "ON"
		if !bl.Enabled {
			status = "OFF"
		}
		name := bl.Name
		if name == "" {
			name = "Unnamed"
		}
		lines = append(lines, fmt.Sprintf("  [%s] %s (%s rules)", status, name, formatInt64(int64(bl.RulesCount))))
	}

	// Allow lists section.
	enabledAllow := 0
	for _, al := range allowLists {
		if al.Enabled {
			enabledAllow++
		}
	}
	if len(allowLists) > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("<b>Allow Lists</b> (%d/%d enabled)", enabledAllow, len(allowLists)))

		for _, al := range allowLists {
			status := "ON"
			if !al.Enabled {
				status = "OFF"
			}
			name := al.Name
			if name == "" {
				name = "Unnamed"
			}
			lines = append(lines, fmt.Sprintf("  [%s] %s (%s rules)", status, name, formatInt64(int64(al.RulesCount))))
		}
	}

	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeRecentLogsMessage(entries []QueryLogEntry) string {
	lines := []string{
		"<b>Recent Queries</b>",
		"---",
	}

	if len(entries) == 0 {
		lines = append(lines, "No recent queries")
	} else {
		for i, e := range entries {
			status := "PASS"
			if e.Blocked {
				status = "BLOCK"
			}

			timeStr := e.Time.Format("15:04:05")
			durationStr := ""
			if e.Duration > 0 {
				durationStr = fmt.Sprintf(" (%s)", e.Duration.Truncate(time.Microsecond).String())
			}

			clientStr := ""
			if e.Client != "" {
				clientStr = fmt.Sprintf(" [%s]", e.Client)
			}

			lines = append(lines, fmt.Sprintf("%d. [%s] %s %s%s%s",
				i+1, status, timeStr, e.Domain, durationStr, clientStr))
		}
	}

	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeProtectionStatusMessage(enabled bool) string {
	status := "ENABLED"
	icon := "ON"
	if !enabled {
		status = "DISABLED"
		icon = "OFF"
	}

	lines := []string{
		"<b>Protection Status</b>",
		"---",
		fmt.Sprintf("DNS Protection: <b>%s</b> (%s)", status, icon),
		"",
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeProcessInfoMessage(startTime time.Time, info systeminfo.Info) string {
	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	pid := fmt.Sprintf("%d", os.Getpid())
	goroutines := runtime.NumGoroutine()
	uptime := time.Since(startTime).Truncate(time.Second)

	lines := []string{
		"<b>Process Information</b>",
		"---",
		fmt.Sprintf("PID: %s", pid),
		fmt.Sprintf("Alloc: %s", formatBytesUint(mem.Alloc)),
		fmt.Sprintf("Sys: %s", formatBytesUint(mem.Sys)),
		fmt.Sprintf("Goroutines: %d", goroutines),
		fmt.Sprintf("Process Uptime: %s", uptime.String()),
		fmt.Sprintf("Go Version: %s", runtime.Version()),
	}

	// Total processes on system.
	if info.TotalProcesses > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf("Total system processes: %d", info.TotalProcesses))
	}

	// Self metrics from gopsutil.
	if info.SelfCPUPercent > 0 || info.SelfMemBytes > 0 {
		lines = append(lines, "")
		lines = append(lines, "<b>AdGuard Home Process</b>")
		if info.SelfCPUPercent > 0 {
			lines = append(lines, fmt.Sprintf("  CPU: <code>%s</code>", formatPercentage(info.SelfCPUPercent)))
		}
		if info.SelfMemBytes > 0 {
			lines = append(lines, fmt.Sprintf("  Memory: %s", formatBytesUint(info.SelfMemBytes)))
		}
		if info.SelfOpenFiles > 0 {
			lines = append(lines, fmt.Sprintf("  Open files: %d", info.SelfOpenFiles))
		}
		if info.SelfThreads > 0 {
			lines = append(lines, fmt.Sprintf("  Threads: %d", info.SelfThreads))
		}
	}

	lines = append(lines, "")
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}
