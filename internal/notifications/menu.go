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
				{Text: "🖥️ System Status", CallbackData: "cmd:status"},
				{Text: "📊 DNS Stats", CallbackData: "cmd:stats"},
			},
			{
				{Text: "🔒 Filters", CallbackData: "cmd:filters"},
				{Text: "🛡️ Protection", CallbackData: "cmd:protection"},
			},
			{
				{Text: "📋 Recent Logs", CallbackData: "cmd:logs"},
				{Text: "⚙️ Process Info", CallbackData: "cmd:processes"},
			},
			{
				{Text: "🗂️ Filter Manage", CallbackData: "cmd:filtermgr"},
			},
		},
	}
}

func backToMenuKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "🔙 Back to Menu", CallbackData: "cmd:menu"},
			},
		},
	}
}

// protectionKeyboard returns a keyboard with a toggle button based on current
// protection state.
func protectionKeyboard(enabled bool) *tgInlineKeyboardMarkup {
	var toggleBtn tgInlineKeyboardButton
	if enabled {
		toggleBtn = tgInlineKeyboardButton{Text: "🔴 Disable Protection", CallbackData: "cmd:protection_off"}
	} else {
		toggleBtn = tgInlineKeyboardButton{Text: "🟢 Enable Protection", CallbackData: "cmd:protection_on"}
	}

	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{toggleBtn},
			{{Text: "🔙 Back to Menu", CallbackData: "cmd:menu"}},
		},
	}
}

func filterManageKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "➕ Add Blocklist", CallbackData: "cmd:filtermgr_addblock"},
				{Text: "➕ Add Allowlist", CallbackData: "cmd:filtermgr_addallow"},
			},
			{
				{Text: "🗑️ Remove Blocklist", CallbackData: "cmd:filtermgr_rmblock"},
				{Text: "🗑️ Remove Allowlist", CallbackData: "cmd:filtermgr_rmallow"},
			},
			{
				{Text: "✅ Enable List", CallbackData: "cmd:filtermgr_enable"},
				{Text: "🚫 Disable List", CallbackData: "cmd:filtermgr_disable"},
			},
			{
				{Text: "🔄 Update All Lists", CallbackData: "cmd:filtermgr_update"},
			},
			{
				{Text: "🔙 Back to Menu", CallbackData: "cmd:menu"},
			},
		},
	}
}

func composeFilterManageMessage() string {
	lines := []string{
		"🗂️ <b>Filter Management</b>",
		divider(),
		"",
		"Manage your filter lists using the buttons below.",
		"",
		"  ➕ <b>Add</b>          — Add new blocklist or allowlist by URL",
		"  🗑️ <b>Remove</b>       — Select a list to remove",
		"  ✅ <b>Enable</b>       — Enable a disabled list",
		"  🚫 <b>Disable</b>      — Disable an active list",
		"  🔄 <b>Update All</b>   — Check all lists for updates",
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterSelectionMessage(title, action, listType string, blockLists, allowLists []FilterListInfo) (string, *tgInlineKeyboardMarkup) {
	var lines []string
	lines = append(lines, fmt.Sprintf("🗂️ <b>%s</b>", title))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, "Select a list from the buttons below:")

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
			if len(name) > 28 {
				name = name[:25] + "..."
			}

			status := ""
			if action == "en" || action == "dis" {
				if fl.Enabled {
					status = " ✅"
				} else {
					status = " 🚫"
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
				lines = append(lines, "ℹ️ No allowlists configured.")
			} else {
				addListButtons(allowLists, "a", nil)
			}
		} else {
			if len(blockLists) == 0 {
				lines = append(lines, "")
				lines = append(lines, "ℹ️ No blocklists configured.")
			} else {
				addListButtons(blockLists, "b", nil)
			}
		}
	case "en":
		showDisabled := false
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "🚫 <b>Blocklists (disabled):</b>")
			addListButtons(blockLists, "b", &showDisabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "🚫 <b>Allowlists (disabled):</b>")
			addListButtons(allowLists, "a", &showDisabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ All lists are already enabled.")
		}
	case "dis":
		showEnabled := true
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ <b>Blocklists (enabled):</b>")
			addListButtons(blockLists, "b", &showEnabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ <b>Allowlists (enabled):</b>")
			addListButtons(allowLists, "a", &showEnabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "🚫 All lists are already disabled.")
		}
	}

	buttons = append(buttons, []tgInlineKeyboardButton{{Text: "🔙 Back to Filter Manage", CallbackData: "cmd:filtermgr"}})

	text := strings.Join(lines, "\n")
	kb := &tgInlineKeyboardMarkup{InlineKeyboard: buttons}

	return text, kb
}

func composeFilterManageHelpMessage(action string) string {
	switch action {
	case "addblock":
		return "➕ <b>Add Blocklist</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/addlist &lt;url&gt;</code>\n\n<b>Multiple URLs:</b>\n<code>/addlist url1 url2 url3</code>\n\n<b>With custom name:</b>\n<code>/addlist url | name</code>"
	case "addallow":
		return "➕ <b>Add Allowlist</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/addallow &lt;url&gt;</code>\n\n<b>Multiple URLs:</b>\n<code>/addallow url1 url2 url3</code>\n\n<b>With custom name:</b>\n<code>/addallow url | name</code>"
	case "rmblock":
		return "🗑️ <b>Remove Blocklist</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/removelist &lt;url&gt;</code>"
	case "rmallow":
		return "🗑️ <b>Remove Allowlist</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/removeallow &lt;url&gt;</code>"
	case "enable":
		return "✅ <b>Enable List</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/enablelist &lt;url&gt;</code>"
	case "disable":
		return "🚫 <b>Disable List</b>\n" + divider() + "\n\nSend a message with the command:\n\n<code>/disablelist &lt;url&gt;</code>"
	default:
		return "❓ Unknown action"
	}
}

func composeMainMenuMessage(info systeminfo.Info, protectionOn bool) string {
	protIcon := "🟢"
	protStatus := "ON"
	if !protectionOn {
		protIcon = "🔴"
		protStatus = "OFF"
	}

	lines := []string{
		"🏠 <b>AdGuard Home</b>",
		divider(),
		"",
		sectionHeader("📡", "Quick Status"),
		fmt.Sprintf("  🏷️ <b>Host:</b>       <code>%s</code>", fallbackString(info.Hostname)),
		fmt.Sprintf("  🛡️ <b>Protection:</b> %s <b>%s</b>", protIcon, protStatus),
		fmt.Sprintf("  ⚙️ <b>CPU:</b>        %s", usageBar(info.CPUUsage)),
		fmt.Sprintf("  💾 <b>Memory:</b>     %s", formatUsageWithBar(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)),
	}

	// Swap info if available.
	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("  🔄 <b>Swap:</b>       %s", formatUsageWithBar(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	lines = append(lines, "")
	lines = append(lines, "📌 Select an option below:")
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeSystemStatusMessage(info systeminfo.Info, ioStats IOStats) string {
	lines := []string{
		"🖥️ <b>System Status</b>",
		divider(),
		"",
		sectionHeader("📋", "General"),
		fmt.Sprintf("  🏷️ <b>Hostname:</b> <code>%s</code>", fallbackString(info.Hostname)),
		fmt.Sprintf("  🐧 <b>OS:</b>       %s", formatOS(info)),
	}

	if info.KernelVersion != "" {
		lines = append(lines, fmt.Sprintf("  🔧 <b>Kernel:</b>   <code>%s</code>", info.KernelVersion))
	}

	if info.VirtPlatform != "" {
		lines = append(lines, fmt.Sprintf("  🖥️ <b>Platform:</b> <code>%s</code> (virtualized)", info.VirtPlatform))
	}

	if info.SystemTime != "" {
		if t, err := time.Parse(time.RFC3339, info.SystemTime); err == nil {
			lines = append(lines, fmt.Sprintf("  🕐 <b>Time:</b>     <code>%s</code>", toLocal(t).Format("15:04:05 02/01/2006")))
		}
	}

	uptime := formatUptime(info.UptimeSeconds)
	if uptime == "" {
		uptime = "-"
	}
	lines = append(lines, fmt.Sprintf("  ⏱️ <b>Uptime:</b>   %s", uptime))

	if info.BootTime > 0 {
		bootT := toLocal(time.Unix(int64(info.BootTime), 0))
		lines = append(lines, fmt.Sprintf("  🚀 <b>Boot:</b>     <code>%s</code>", bootT.Format("02/01/2006 15:04")))
	}

	lines = append(lines, "")

	// CPU section.
	lines = append(lines, sectionHeader("⚙️", "CPU"))
	lines = append(lines, fmt.Sprintf("  📌 <b>Model:</b> %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("  📊 <b>Usage:</b> %s", usageBar(info.CPUUsage)))

	// Load average (non-zero means non-Windows).
	if info.LoadAvg1 > 0 || info.LoadAvg5 > 0 || info.LoadAvg15 > 0 {
		lines = append(lines, fmt.Sprintf("  📈 <b>Load:</b>  <code>%.2f / %.2f / %.2f</code> <i>(1/5/15 min)</i>", info.LoadAvg1, info.LoadAvg5, info.LoadAvg15))
	}

	lines = append(lines, "")

	// Memory section.
	lines = append(lines, sectionHeader("💾", "Memory"))
	lines = append(lines, fmt.Sprintf("  📊 <b>RAM:</b>  %s", formatUsageWithBar(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("  ✅ <b>Free:</b> %s", formatCapacity(info.MemoryFree, info.MemoryTotal)))

	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("  🔄 <b>Swap:</b> %s", formatUsageWithBar(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	lines = append(lines, "")

	// Disk section.
	lines = append(lines, sectionHeader("💿", "Disk"))
	if len(info.AllDisks) > 0 {
		for _, d := range info.AllDisks {
			fsLabel := ""
			if d.Filesystem != "" {
				fsLabel = fmt.Sprintf(" <i>[%s]</i>", d.Filesystem)
			}
			lines = append(lines, fmt.Sprintf("  📁 <code>%s</code>: %s%s", d.Path, formatUsageWithBar(d.Used, d.Total, d.UsagePercent), fsLabel))
		}
	} else {
		lines = append(lines, fmt.Sprintf("  📁 <code>%s</code>: %s", fallbackString(info.DiskPath), formatUsageWithBar(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	}

	// Disk I/O rates.
	if ioStats.DiskReadBytesPerSec > 0 || ioStats.DiskWriteBytesPerSec > 0 {
		lines = append(lines, fmt.Sprintf("  ⬆️ <b>Write:</b> <code>%s</code>  ⬇️ <b>Read:</b> <code>%s</code>", formatRate(ioStats.DiskWriteBytesPerSec), formatRate(ioStats.DiskReadBytesPerSec)))
	}

	lines = append(lines, "")

	// Network section.
	lines = append(lines, sectionHeader("🌐", "Network"))
	lines = append(lines, fmt.Sprintf("  🔌 <b>Local IPs:</b>  %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("  🌍 <b>Public IP:</b>  <code>%s</code>", fallbackString(info.PublicIP)))

	if ioStats.NetBytesRecvPerSec > 0 || ioStats.NetBytesSentPerSec > 0 {
		lines = append(lines, fmt.Sprintf("  ⬇️ <b>Recv:</b> <code>%s</code>  ⬆️ <b>Send:</b> <code>%s</code>", formatRate(ioStats.NetBytesRecvPerSec), formatRate(ioStats.NetBytesSentPerSec)))
	}

	if info.ActiveConns > 0 {
		lines = append(lines, fmt.Sprintf("  🔗 <b>TCP connections:</b> <code>%d</code>", info.ActiveConns))
	}

	if info.NetErrorsIn > 0 || info.NetErrorsOut > 0 {
		lines = append(lines, fmt.Sprintf("  ⚠️ <b>Errors:</b> in <code>%d</code>  out <code>%d</code>", info.NetErrorsIn, info.NetErrorsOut))
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeDNSStatsMessage(numQueries, numBlocked, numSafeBrowsing, numParental uint64, avgMs float64) string {
	avgMsStr := fmt.Sprintf("<code>%s ms</code>", formatFloat(avgMs*1000))

	blockPctStr := "-"
	if numQueries > 0 {
		pct := float64(numBlocked) / float64(numQueries) * 100
		blockPctStr = usageBar(pct)
	}

	lines := []string{
		"📊 <b>DNS Statistics</b>",
		divider(),
		"",
		sectionHeader("🔍", "Query Summary"),
		fmt.Sprintf("  📨 <b>Total Queries:</b>   <code>%s</code>", formatUint64(numQueries)),
		fmt.Sprintf("  🚫 <b>Blocked:</b>         <code>%s</code>", formatUint64(numBlocked)),
		fmt.Sprintf("  🛡️ <b>Safe Browsing:</b>   <code>%s</code>", formatUint64(numSafeBrowsing)),
		fmt.Sprintf("  👨‍👩‍👧 <b>Parental:</b>        <code>%s</code>", formatUint64(numParental)),
		fmt.Sprintf("  ⚡ <b>Avg Response:</b>    %s", avgMsStr),
		fmt.Sprintf("  📈 <b>Block Rate:</b>      %s", blockPctStr),
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterMessage(totalRules, enabledBlockLists, enabledAllowLists int) string {
	lines := []string{
		"🔒 <b>Filter Information</b>",
		divider(),
		"",
		fmt.Sprintf("  📏 <b>Total Rules:</b>   <code>%s</code>", formatInt64(int64(totalRules))),
		fmt.Sprintf("  🚫 <b>Block Lists:</b>   <code>%d</code> enabled", enabledBlockLists),
		fmt.Sprintf("  ✅ <b>Allow Lists:</b>   <code>%d</code> enabled", enabledAllowLists),
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterDetailedMessage(totalRules int, blockLists, allowLists []FilterListInfo) string {
	lines := []string{
		"🔒 <b>Filter Information</b>",
		divider(),
		"",
		fmt.Sprintf("  📏 <b>Total Rules:</b> <code>%s</code>", formatInt64(int64(totalRules))),
	}

	// Block lists section.
	enabledBlock := 0
	for _, bl := range blockLists {
		if bl.Enabled {
			enabledBlock++
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("🚫 <b>Block Lists</b> (%d/%d enabled)", enabledBlock, len(blockLists)))

	for _, bl := range blockLists {
		statusIcon := "✅"
		if !bl.Enabled {
			statusIcon = "⏸️"
		}
		name := bl.Name
		if name == "" {
			name = "Unnamed"
		}
		lines = append(lines, fmt.Sprintf("  %s %s <i>(%s rules)</i>", statusIcon, name, formatInt64(int64(bl.RulesCount))))
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
		lines = append(lines, fmt.Sprintf("✅ <b>Allow Lists</b> (%d/%d enabled)", enabledAllow, len(allowLists)))

		for _, al := range allowLists {
			statusIcon := "✅"
			if !al.Enabled {
				statusIcon = "⏸️"
			}
			name := al.Name
			if name == "" {
				name = "Unnamed"
			}
			lines = append(lines, fmt.Sprintf("  %s %s <i>(%s rules)</i>", statusIcon, name, formatInt64(int64(al.RulesCount))))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeRecentLogsMessage(entries []QueryLogEntry) string {
	lines := []string{
		"📋 <b>Recent Queries</b>",
		divider(),
		"",
	}

	if len(entries) == 0 {
		lines = append(lines, "ℹ️ No recent queries found.")
	} else {
		for i, e := range entries {
			statusIcon := "✅"
			if e.Blocked {
				statusIcon = "🚫"
			}

			timeStr := toLocal(e.Time).Format("15:04:05")
			durationStr := ""
			if e.Duration > 0 {
				durationStr = fmt.Sprintf(" <i>(%s)</i>", e.Duration.Truncate(time.Microsecond).String())
			}

			clientStr := ""
			if e.Client != "" {
				clientStr = fmt.Sprintf(" <code>[%s]</code>", e.Client)
			}

			lines = append(lines, fmt.Sprintf("%s %d. <code>%s</code>%s %s%s",
				statusIcon, i+1, timeStr, clientStr, e.Domain, durationStr))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeProtectionStatusMessage(enabled bool) string {
	var statusIcon, statusText, description string
	if enabled {
		statusIcon = "🟢"
		statusText = "ENABLED"
		description = "DNS filtering is active. All queries are protected."
	} else {
		statusIcon = "🔴"
		statusText = "DISABLED"
		description = "⚠️ DNS filtering is OFF. Queries pass through unfiltered."
	}

	lines := []string{
		"🛡️ <b>Protection Status</b>",
		divider(),
		"",
		fmt.Sprintf("  %s DNS Protection: <b>%s</b>", statusIcon, statusText),
		"",
		fmt.Sprintf("  <i>%s</i>", description),
		"",
		divider(),
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
		"⚙️ <b>Process Information</b>",
		divider(),
		"",
		sectionHeader("🔧", "Runtime"),
		fmt.Sprintf("  🆔 <b>PID:</b>           <code>%s</code>", pid),
		fmt.Sprintf("  💾 <b>Heap Alloc:</b>    <code>%s</code>", formatBytesUint(mem.Alloc)),
		fmt.Sprintf("  🖥️ <b>Sys Memory:</b>    <code>%s</code>", formatBytesUint(mem.Sys)),
		fmt.Sprintf("  🔀 <b>Goroutines:</b>    <code>%d</code>", goroutines),
		fmt.Sprintf("  ⏱️ <b>Process Uptime:</b> %s", uptime.String()),
		fmt.Sprintf("  🐹 <b>Go Version:</b>    <code>%s</code>", runtime.Version()),
	}

	// Total processes on system.
	if info.TotalProcesses > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🖥️", "System Processes"))
		lines = append(lines, fmt.Sprintf("  📊 <b>Total:</b> <code>%d</code>", info.TotalProcesses))
	}

	// Self metrics from gopsutil.
	if info.SelfCPUPercent > 0 || info.SelfMemBytes > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🏠", "AdGuard Home Process"))
		if info.SelfCPUPercent > 0 {
			lines = append(lines, fmt.Sprintf("  ⚙️ <b>CPU:</b>        %s", usageBar(info.SelfCPUPercent)))
		}
		if info.SelfMemBytes > 0 {
			lines = append(lines, fmt.Sprintf("  💾 <b>Memory:</b>     <code>%s</code>", formatBytesUint(info.SelfMemBytes)))
		}
		if info.SelfOpenFiles > 0 {
			lines = append(lines, fmt.Sprintf("  📂 <b>Open Files:</b> <code>%d</code>", info.SelfOpenFiles))
		}
		if info.SelfThreads > 0 {
			lines = append(lines, fmt.Sprintf("  🔀 <b>Threads:</b>    <code>%d</code>", info.SelfThreads))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeFilterRemoveSelectionMessage builds the multi-select checkbox view for
// bulk removal. Each list item is a button that toggles a ☑️/⬜ marker.
// A "Delete Selected (N)" confirm button and a "Cancel" button are appended.
func composeFilterRemoveSelectionMessage(title, listType string, lists []FilterListInfo, selected map[int]bool) (string, *tgInlineKeyboardMarkup) {
	selectedCount := 0
	for _, v := range selected {
		if v {
			selectedCount++
		}
	}

	lines := []string{
		fmt.Sprintf("🗑️ <b>%s</b>", title),
		divider(),
		"",
	}

	if len(lists) == 0 {
		noMsg := "ℹ️ No blocklists configured."
		if listType == "a" {
			noMsg = "ℹ️ No allowlists configured."
		}
		lines = append(lines, noMsg)
		kb := &tgInlineKeyboardMarkup{
			InlineKeyboard: [][]tgInlineKeyboardButton{
				{{Text: "🔙 Back to Filter Manage", CallbackData: "cmd:filtermgr"}},
			},
		}

		return strings.Join(lines, "\n"), kb
	}

	if selectedCount > 0 {
		lines = append(lines, fmt.Sprintf("☑️ <b>%d list(s) selected</b> — tap again to deselect", selectedCount))
	} else {
		lines = append(lines, "☑️ Tap the lists you want to remove, then confirm:")
	}
	lines = append(lines, "")

	buttons := make([][]tgInlineKeyboardButton, 0, len(lists)+2)
	for i, fl := range lists {
		name := fl.Name
		if name == "" {
			name = "Unnamed"
		}
		if len(name) > 26 {
			name = name[:23] + "..."
		}

		checkIcon := "⬜"
		if selected[i] {
			checkIcon = "☑️"
		}
		label := fmt.Sprintf("%s %s", checkIcon, name)
		cb := fmt.Sprintf("flt:rmtog:%s:%d", listType, i)
		buttons = append(buttons, []tgInlineKeyboardButton{{Text: label, CallbackData: cb}})
	}

	// Confirm row.
	confirmLabel := fmt.Sprintf("🗑️ Delete Selected (%d)", selectedCount)
	buttons = append(buttons, []tgInlineKeyboardButton{
		{Text: confirmLabel, CallbackData: fmt.Sprintf("flt:rmdo:%s", listType)},
	})
	// Cancel row.
	buttons = append(buttons, []tgInlineKeyboardButton{
		{Text: "❌ Cancel", CallbackData: "flt:rmcancel"},
	})

	return strings.Join(lines, "\n"), &tgInlineKeyboardMarkup{InlineKeyboard: buttons}
}

// composeRemoveResultMessage builds a summary of a bulk-remove operation.
func composeRemoveResultMessage(results []removeResult) string {
	successCount := 0
	for _, r := range results {
		if r.err == nil {
			successCount++
		}
	}

	lines := []string{
		fmt.Sprintf("🗑️ <b>Bulk Remove — %d/%d succeeded</b>", successCount, len(results)),
		divider(),
		"",
	}

	for _, r := range results {
		name := r.name
		if name == "" {
			name = r.url
		}
		if len(name) > 40 {
			name = name[:37] + "..."
		}

		if r.err == nil {
			lines = append(lines, fmt.Sprintf("  ✅ %s", name))
		} else {
			lines = append(lines, fmt.Sprintf("  ❌ %s", name))
			lines = append(lines, fmt.Sprintf("     <i>%s</i>", r.err.Error()))
		}
	}

	return strings.Join(lines, "\n")
}
