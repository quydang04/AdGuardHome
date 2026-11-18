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
				{Text: "📊 Trạng thái", CallbackData: "cmd:status"},
				{Text: "🔍 Thống kê DNS", CallbackData: "cmd:stats"},
			},
			{
				{Text: "🛡 Bảo vệ", CallbackData: "cmd:protection"},
				{Text: "📋 Nhật ký", CallbackData: "cmd:logs"},
			},
			{
				{Text: "🔒 Bộ lọc", CallbackData: "cmd:filters"},
				{Text: "🗂 Quản lý", CallbackData: "cmd:filtermgr"},
			},
			{
				{Text: "⚙️ Tiến trình", CallbackData: "cmd:processes"},
				{Text: "🔐 SSL/TLS", CallbackData: "cmd:ssl"},
			},
		},
	}
}

func backToMenuKeyboard(refreshCmd ...string) *tgInlineKeyboardMarkup {
	row := []tgInlineKeyboardButton{}
	if len(refreshCmd) > 0 && refreshCmd[0] != "" {
		row = append(row, tgInlineKeyboardButton{Text: "🔄 Làm mới", CallbackData: refreshCmd[0]})
	}
	row = append(row, tgInlineKeyboardButton{Text: "« Quay lại", CallbackData: "cmd:menu"})
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{row},
	}
}

// protectionKeyboard returns a keyboard with a toggle button based on current
// protection state.
func protectionKeyboard(enabled bool) *tgInlineKeyboardMarkup {
	var toggleBtn tgInlineKeyboardButton
	if enabled {
		toggleBtn = tgInlineKeyboardButton{Text: "⏸ Tắt bảo vệ", CallbackData: "cmd:protection_off"}
	} else {
		toggleBtn = tgInlineKeyboardButton{Text: "▶️ Bật bảo vệ", CallbackData: "cmd:protection_on"}
	}

	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{toggleBtn},
			{{Text: "🔄 Làm mới", CallbackData: "cmd:protection"}, {Text: "« Quay lại", CallbackData: "cmd:menu"}},
		},
	}
}

// certKeyboard returns a keyboard for the SSL/TLS issue status message, with
// buttons to toggle auto-renew and trigger an immediate issuance.
func certKeyboard(autoRenew bool) *tgInlineKeyboardMarkup {
	var toggleBtn tgInlineKeyboardButton
	if autoRenew {
		toggleBtn = tgInlineKeyboardButton{Text: "⏸ Tắt tự động gia hạn", CallbackData: "cmd:ssl_autorenew_off"}
	} else {
		toggleBtn = tgInlineKeyboardButton{Text: "▶️ Bật tự động gia hạn", CallbackData: "cmd:ssl_autorenew_on"}
	}

	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{toggleBtn},
			{{Text: "📜 Cấp phát ngay", CallbackData: "cmd:ssl_issue_now"}, {Text: "« Quay lại", CallbackData: "cmd:menu"}},
		},
	}
}

func filterManageKeyboard() *tgInlineKeyboardMarkup {
	return &tgInlineKeyboardMarkup{
		InlineKeyboard: [][]tgInlineKeyboardButton{
			{
				{Text: "➕ Thêm chặn", CallbackData: "cmd:filtermgr_addblock"},
				{Text: "➕ Thêm cho phép", CallbackData: "cmd:filtermgr_addallow"},
			},
			{
				{Text: "🗑 Xóa chặn", CallbackData: "cmd:filtermgr_rmblock"},
				{Text: "🗑 Xóa cho phép", CallbackData: "cmd:filtermgr_rmallow"},
			},
			{
				{Text: "✅ Kích hoạt", CallbackData: "cmd:filtermgr_enable"},
				{Text: "⏸ Vô hiệu", CallbackData: "cmd:filtermgr_disable"},
			},
			{
				{Text: "🔄 Cập nhật tất cả", CallbackData: "cmd:filtermgr_update"},
			},
			{
				{Text: "« Quay lại", CallbackData: "cmd:menu"},
			},
		},
	}
}

func composeFilterManageMessage() string {
	lines := []string{
		"🗂 <b>Quản lý bộ lọc</b>",
		divider(),
		"",
		"<b>Thao tác:</b>",
		"   ➕ <b>Thêm</b> — Thêm bộ lọc mới bằng URL",
		"   🗑 <b>Xóa</b> — Chọn bộ lọc để xóa",
		"   ✅ <b>Kích hoạt</b> — Bật bộ lọc đang tắt",
		"   ⏸ <b>Vô hiệu</b> — Tắt bộ lọc đang bật",
		"   🔄 <b>Cập nhật</b> — Làm mới tất cả bộ lọc",
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterSelectionMessage(title, action, listType string, blockLists, allowLists []FilterListInfo) (string, *tgInlineKeyboardMarkup) {
	var lines []string
	lines = append(lines, fmt.Sprintf("🗂 <b>%s</b>", title))
	lines = append(lines, divider())
	lines = append(lines, "")
	lines = append(lines, "Chọn bộ lọc:")

	buttons := make([][]tgInlineKeyboardButton, 0)

	addListButtons := func(lists []FilterListInfo, lt string, filterEnabled *bool) {
		for i, fl := range lists {
			if filterEnabled != nil && fl.Enabled != *filterEnabled {
				continue
			}

			name := fl.Name
			if name == "" {
				name = "Chưa đặt tên"
			}
			if len(name) > 28 {
				name = name[:25] + "..."
			}

			status := ""
			if action == "en" || action == "dis" {
				if fl.Enabled {
					status = " ✅"
				} else {
					status = " ⏸"
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
				lines = append(lines, "ℹ️ Chưa có danh sách cho phép nào.")
			} else {
				addListButtons(allowLists, "a", nil)
			}
		} else {
			if len(blockLists) == 0 {
				lines = append(lines, "")
				lines = append(lines, "ℹ️ Chưa có danh sách chặn nào.")
			} else {
				addListButtons(blockLists, "b", nil)
			}
		}
	case "en":
		showDisabled := false
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "⏸ <b>DS chặn (đang tắt):</b>")
			addListButtons(blockLists, "b", &showDisabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "⏸ <b>DS cho phép (đang tắt):</b>")
			addListButtons(allowLists, "a", &showDisabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ Tất cả bộ lọc đã được kích hoạt.")
		}
	case "dis":
		showEnabled := true
		if len(blockLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ <b>DS chặn (đang bật):</b>")
			addListButtons(blockLists, "b", &showEnabled)
		}
		if len(allowLists) > 0 {
			lines = append(lines, "")
			lines = append(lines, "✅ <b>DS cho phép (đang bật):</b>")
			addListButtons(allowLists, "a", &showEnabled)
		}
		if len(buttons) == 0 {
			lines = append(lines, "")
			lines = append(lines, "⏸ Tất cả bộ lọc đã bị vô hiệu.")
		}
	}

	buttons = append(buttons, []tgInlineKeyboardButton{{Text: "« Quản lý bộ lọc", CallbackData: "cmd:filtermgr"}})

	text := strings.Join(lines, "\n")
	kb := &tgInlineKeyboardMarkup{InlineKeyboard: buttons}

	return text, kb
}

func composeFilterManageHelpMessage(action string) string {
	switch action {
	case "addblock":
		return "➕ <b>Thêm danh sách chặn</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/addlist &lt;url&gt;</code>\n\n<b>Nhiều URL:</b>\n<code>/addlist url1 url2 url3</code>\n\n<b>Đặt tên:</b>\n<code>/addlist url | tên</code>"
	case "addallow":
		return "➕ <b>Thêm danh sách cho phép</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/addallow &lt;url&gt;</code>\n\n<b>Nhiều URL:</b>\n<code>/addallow url1 url2 url3</code>\n\n<b>Đặt tên:</b>\n<code>/addallow url | tên</code>"
	case "rmblock":
		return "🗑 <b>Xóa danh sách chặn</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/removelist &lt;url&gt;</code>"
	case "rmallow":
		return "🗑 <b>Xóa danh sách cho phép</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/removeallow &lt;url&gt;</code>"
	case "enable":
		return "✅ <b>Kích hoạt bộ lọc</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/enablelist &lt;url&gt;</code>"
	case "disable":
		return "⏸ <b>Vô hiệu bộ lọc</b>\n" + divider() + "\n\nGửi lệnh:\n\n<code>/disablelist &lt;url&gt;</code>"
	default:
		return "❓ Thao tác không xác định"
	}
}

func composeMainMenuMessage(info systeminfo.Info, protectionOn bool) string {
	protIcon := statusDot(protectionOn)
	protLabel := "Đang bảo vệ"
	if !protectionOn {
		protLabel = "Đã tắt"
	}

	lines := []string{
		"🏠 <b>AdGuard Home</b>",
		divider(),
		"",
		fmt.Sprintf("   %s Bảo vệ: <b>%s</b>", protIcon, protLabel),
		fmt.Sprintf("   ⚙️ CPU:  %s", usageBar(info.CPUUsage)),
		fmt.Sprintf("   💾 RAM:  %s", formatUsageWithBar(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)),
	}

	// Swap info if available.
	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("   🔄 Swap: %s", formatUsageWithBar(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("   🏷 <code>%s</code>", fallbackString(info.Hostname)))
	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeSystemStatusMessage(info systeminfo.Info, ioStats IOStats) string {
	lines := []string{
		"📊 <b>Trạng thái hệ thống</b>",
		divider(),
	}

	// ── Tổng quan ──
	lines = append(lines, "")
	lines = append(lines, sectionHeader("🏷", "Tổng quan"))
	lines = append(lines, fmt.Sprintf("   ├ Máy chủ: <code>%s</code>", fallbackString(info.Hostname)))
	lines = append(lines, fmt.Sprintf("   ├ OS: %s", formatOS(info)))

	if info.KernelVersion != "" {
		lines = append(lines, fmt.Sprintf("   ├ Kernel: <code>%s</code>", info.KernelVersion))
	}

	if info.VirtPlatform != "" {
		lines = append(lines, fmt.Sprintf("   ├ Nền tảng: <code>%s</code>", info.VirtPlatform))
	}

	if info.SystemTime != "" {
		if t, err := time.Parse(time.RFC3339, info.SystemTime); err == nil {
			lines = append(lines, fmt.Sprintf("   ├ Thời gian: <code>%s</code>", toLocal(t).Format("15:04:05 02/01/2006")))
		}
	}

	uptime := formatUptime(info.UptimeSeconds)
	if uptime == "" {
		uptime = "-"
	}
	lines = append(lines, fmt.Sprintf("   └ Hoạt động: %s", uptime))

	if info.BootTime > 0 {
		bootT := toLocal(time.Unix(int64(info.BootTime), 0))
		lines = append(lines, fmt.Sprintf("     Khởi động: <code>%s</code>", bootT.Format("02/01/2006 15:04")))
	}

	// ── CPU ──
	lines = append(lines, "")
	lines = append(lines, sectionHeader("⚙️", "CPU"))
	lines = append(lines, fmt.Sprintf("   ├ %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("   └ Sử dụng: %s", usageBar(info.CPUUsage)))

	if info.LoadAvg1 > 0 || info.LoadAvg5 > 0 || info.LoadAvg15 > 0 {
		lines = append(lines, fmt.Sprintf("     Tải: <code>%.2f / %.2f / %.2f</code> <i>(1/5/15 phút)</i>", info.LoadAvg1, info.LoadAvg5, info.LoadAvg15))
	}

	// ── Bộ nhớ ──
	lines = append(lines, "")
	lines = append(lines, sectionHeader("💾", "Bộ nhớ"))
	lines = append(lines, fmt.Sprintf("   ├ RAM:    %s", formatUsageWithBar(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("   └ Trống:  %s", formatCapacity(info.MemoryFree, info.MemoryTotal)))

	if info.SwapTotal > 0 {
		lines = append(lines, fmt.Sprintf("     Swap:   %s", formatUsageWithBar(info.SwapUsed, info.SwapTotal, info.SwapUsage)))
	}

	// ── Ổ đĩa ──
	lines = append(lines, "")
	lines = append(lines, sectionHeader("💿", "Ổ đĩa"))
	if len(info.AllDisks) > 0 {
		for i, d := range info.AllDisks {
			prefix := "├"
			if i == len(info.AllDisks)-1 && ioStats.DiskReadBytesPerSec == 0 && ioStats.DiskWriteBytesPerSec == 0 {
				prefix = "└"
			}
			fsLabel := ""
			if d.Filesystem != "" {
				fsLabel = fmt.Sprintf(" [%s]", d.Filesystem)
			}
			lines = append(lines, fmt.Sprintf("   %s <code>%s</code>%s", prefix, d.Path, fsLabel))
			lines = append(lines, fmt.Sprintf("     %s", formatUsageWithBar(d.Used, d.Total, d.UsagePercent)))
		}
	} else {
		lines = append(lines, fmt.Sprintf("   └ <code>%s</code>: %s", fallbackString(info.DiskPath), formatUsageWithBar(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	}

	if ioStats.DiskReadBytesPerSec > 0 || ioStats.DiskWriteBytesPerSec > 0 {
		lines = append(lines, fmt.Sprintf("   └ I/O: ↑<code>%s</code> ↓<code>%s</code>", formatRate(ioStats.DiskWriteBytesPerSec), formatRate(ioStats.DiskReadBytesPerSec)))
	}

	// ── Mạng ──
	lines = append(lines, "")
	lines = append(lines, sectionHeader("🌐", "Mạng"))
	lines = append(lines, fmt.Sprintf("   ├ Nội bộ: %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("   └ Công cộng: <code>%s</code>", fallbackString(info.PublicIP)))

	if ioStats.NetBytesRecvPerSec > 0 || ioStats.NetBytesSentPerSec > 0 {
		lines = append(lines, fmt.Sprintf("     Lưu lượng: ↓<code>%s</code> ↑<code>%s</code>", formatRate(ioStats.NetBytesRecvPerSec), formatRate(ioStats.NetBytesSentPerSec)))
	}

	if info.ActiveConns > 0 {
		lines = append(lines, fmt.Sprintf("     TCP: <code>%d</code> kết nối", info.ActiveConns))
	}

	if info.NetErrorsIn > 0 || info.NetErrorsOut > 0 {
		lines = append(lines, fmt.Sprintf("     ⚠️ Lỗi: vào <code>%d</code> ra <code>%d</code>", info.NetErrorsIn, info.NetErrorsOut))
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeDNSStatsMessage(numQueries, numBlocked, numSafeBrowsing, numParental uint64, avgMs float64) string {
	avgMsStr := fmt.Sprintf("<code>%s ms</code>", formatFloat(avgMs*1000))

	blockPctStr := "-"
	var blockPct float64
	if numQueries > 0 {
		blockPct = float64(numBlocked) / float64(numQueries) * 100
		blockPctStr = fmt.Sprintf("<code>%s</code>", formatPercentage(blockPct))
	}

	lines := []string{
		"🔍 <b>Thống kê DNS</b>",
		divider(),
		"",
		fmt.Sprintf("   📨 Truy vấn:       <code>%s</code>", formatUint64(numQueries)),
		fmt.Sprintf("   🚫 Đã chặn:        <code>%s</code>", formatUint64(numBlocked)),
		fmt.Sprintf("   🛡 Duyệt an toàn:  <code>%s</code>", formatUint64(numSafeBrowsing)),
		fmt.Sprintf("   👨‍👩‍👧 Kiểm soát:     <code>%s</code>", formatUint64(numParental)),
		"",
		thinDivider(),
		"",
		fmt.Sprintf("   ⚡ Xử lý TB:        %s", avgMsStr),
		fmt.Sprintf("   📈 Tỷ lệ chặn:     %s", blockPctStr),
	}

	if numQueries > 0 {
		bar := formatProgressBar(blockPct, 16)
		lines = append(lines, fmt.Sprintf("      %s", bar))
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeFilterMessage(totalRules, enabledBlockLists, enabledAllowLists int) string {
	lines := []string{
		"🔒 <b>Bộ lọc</b>",
		divider(),
		"",
		fmt.Sprintf("   📏 Tổng quy tắc:  <code>%s</code>", formatInt64(int64(totalRules))),
		fmt.Sprintf("   🚫 DS chặn:        <code>%d</code> đang hoạt động", enabledBlockLists),
		fmt.Sprintf("   ✅ DS cho phép:     <code>%d</code> đang hoạt động", enabledAllowLists),
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}

func composeFilterDetailedMessage(totalRules int, blockLists, allowLists []FilterListInfo) string {
	lines := []string{
		"🔒 <b>Bộ lọc</b>",
		divider(),
		"",
		fmt.Sprintf("   📏 Tổng quy tắc: <code>%s</code>", formatInt64(int64(totalRules))),
	}

	// Block lists section.
	enabledBlock := 0
	for _, bl := range blockLists {
		if bl.Enabled {
			enabledBlock++
		}
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("🚫 <b>Danh sách chặn</b> · %d/%d hoạt động", enabledBlock, len(blockLists)))

	for _, bl := range blockLists {
		statusIcon := "🟢"
		if !bl.Enabled {
			statusIcon = "⏸"
		}
		name := bl.Name
		if name == "" {
			name = "Chưa đặt tên"
		}
		lines = append(lines, fmt.Sprintf("   %s %s · <i>%s quy tắc</i>", statusIcon, name, formatInt64(int64(bl.RulesCount))))
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
		lines = append(lines, fmt.Sprintf("✅ <b>Danh sách cho phép</b> · %d/%d hoạt động", enabledAllow, len(allowLists)))

		for _, al := range allowLists {
			statusIcon := "🟢"
			if !al.Enabled {
				statusIcon = "⏸"
			}
			name := al.Name
			if name == "" {
				name = "Chưa đặt tên"
			}
			lines = append(lines, fmt.Sprintf("   %s %s · <i>%s quy tắc</i>", statusIcon, name, formatInt64(int64(al.RulesCount))))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeRecentLogsMessage(entries []QueryLogEntry) string {
	lines := []string{
		"📋 <b>Truy vấn gần đây</b>",
		divider(),
		"",
	}

	if len(entries) == 0 {
		lines = append(lines, "ℹ️ Không có truy vấn nào gần đây.")
	} else {
		for _, e := range entries {
			statusIcon := "✅"
			if e.Blocked {
				statusIcon = "🚫"
			}

			timeStr := toLocal(e.Time).Format("15:04:05")
			durationStr := ""
			if e.Duration > 0 {
				durationStr = fmt.Sprintf(" · <i>%s</i>", formatDuration(e.Duration))
			}

			clientStr := ""
			if e.Client != "" {
				clientStr = fmt.Sprintf(" <code>[%s]</code>", e.Client)
			}

			lines = append(lines, fmt.Sprintf("%s <code>%s</code>%s %s%s",
				statusIcon, timeStr, clientStr, e.Domain, durationStr))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

func composeProtectionStatusMessage(enabled bool) string {
	icon := statusDot(enabled)
	var statusText, description string
	if enabled {
		statusText = "ĐANG BẬT"
		description = "Lọc DNS đang hoạt động. Mọi truy vấn đều được bảo vệ."
	} else {
		statusText = "ĐÃ TẮT"
		description = "⚠️ Lọc DNS đã tắt. Truy vấn sẽ đi thẳng không qua bộ lọc."
	}

	lines := []string{
		"🛡 <b>Trạng thái bảo vệ</b>",
		divider(),
		"",
		fmt.Sprintf("   %s Bảo vệ DNS: <b>%s</b>", icon, statusText),
		"",
		fmt.Sprintf("   <i>%s</i>", description),
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
		"⚙️ <b>Thông tin tiến trình</b>",
		divider(),
		"",
		sectionHeader("🔧", "Runtime"),
		fmt.Sprintf("   ├ PID:         <code>%s</code>", pid),
		fmt.Sprintf("   ├ Heap:        <code>%s</code>", formatBytesUint(mem.Alloc)),
		fmt.Sprintf("   ├ Bộ nhớ hệ:   <code>%s</code>", formatBytesUint(mem.Sys)),
		fmt.Sprintf("   ├ Goroutines:  <code>%d</code>", goroutines),
		fmt.Sprintf("   ├ Hoạt động:   %s", uptime.String()),
		fmt.Sprintf("   └ Go:          <code>%s</code>", runtime.Version()),
	}

	if info.TotalProcesses > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🖥", "Hệ thống"))
		lines = append(lines, fmt.Sprintf("   └ Tiến trình: <code>%d</code>", info.TotalProcesses))
	}

	if info.SelfCPUPercent > 0 || info.SelfMemBytes > 0 {
		lines = append(lines, "")
		lines = append(lines, sectionHeader("🏠", "AdGuard Home"))
		if info.SelfCPUPercent > 0 {
			lines = append(lines, fmt.Sprintf("   ├ CPU:      %s", usageBar(info.SelfCPUPercent)))
		}
		if info.SelfMemBytes > 0 {
			lines = append(lines, fmt.Sprintf("   ├ Bộ nhớ:   <code>%s</code>", formatBytesUint(info.SelfMemBytes)))
		}
		if info.SelfOpenFiles > 0 {
			lines = append(lines, fmt.Sprintf("   ├ Tệp mở:   <code>%d</code>", info.SelfOpenFiles))
		}
		if info.SelfThreads > 0 {
			lines = append(lines, fmt.Sprintf("   └ Threads:   <code>%d</code>", info.SelfThreads))
		}
	}

	lines = append(lines, "")
	lines = append(lines, divider())
	lines = append(lines, timestampLine())

	return strings.Join(lines, "\n")
}

// composeFilterRemoveSelectionMessage builds the multi-select checkbox view for
// bulk removal.
func composeFilterRemoveSelectionMessage(title, listType string, lists []FilterListInfo, selected map[int]bool) (string, *tgInlineKeyboardMarkup) {
	selectedCount := 0
	for _, v := range selected {
		if v {
			selectedCount++
		}
	}

	lines := []string{
		fmt.Sprintf("🗑 <b>%s</b>", title),
		divider(),
		"",
	}

	if len(lists) == 0 {
		noMsg := "ℹ️ Chưa có danh sách chặn nào."
		if listType == "a" {
			noMsg = "ℹ️ Chưa có danh sách cho phép nào."
		}
		lines = append(lines, noMsg)
		kb := &tgInlineKeyboardMarkup{
			InlineKeyboard: [][]tgInlineKeyboardButton{
				{{Text: "« Quản lý bộ lọc", CallbackData: "cmd:filtermgr"}},
			},
		}

		return strings.Join(lines, "\n"), kb
	}

	if selectedCount > 0 {
		lines = append(lines, fmt.Sprintf("☑️ <b>Đã chọn %d</b> — bấm lại để bỏ chọn", selectedCount))
	} else {
		lines = append(lines, "☑️ Chọn bộ lọc cần xóa, sau đó xác nhận:")
	}
	lines = append(lines, "")

	buttons := make([][]tgInlineKeyboardButton, 0, len(lists)+2)
	for i, fl := range lists {
		name := fl.Name
		if name == "" {
			name = "Chưa đặt tên"
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

	confirmLabel := fmt.Sprintf("🗑 Xóa (%d)", selectedCount)
	buttons = append(buttons, []tgInlineKeyboardButton{
		{Text: confirmLabel, CallbackData: fmt.Sprintf("flt:rmdo:%s", listType)},
	})
	buttons = append(buttons, []tgInlineKeyboardButton{
		{Text: "❌ Hủy", CallbackData: "flt:rmcancel"},
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
		fmt.Sprintf("🗑 <b>Đã xóa %d/%d</b>", successCount, len(results)),
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
			lines = append(lines, fmt.Sprintf("   ✅ %s", name))
		} else {
			lines = append(lines, fmt.Sprintf("   ❌ %s", name))
			lines = append(lines, fmt.Sprintf("      <i>%s</i>", r.err.Error()))
		}
	}

	return strings.Join(lines, "\n")
}

// composeHelpMessage builds a help message listing all available commands.
func composeHelpMessage() string {
	lines := []string{
		"❓ <b>Trợ giúp — Danh sách lệnh</b>",
		divider(),
		"",
		"<b>Điều hướng</b>",
		"   /menu — Menu chính",
		"   /help — Xem trợ giúp",
		"",
		"<b>Giám sát</b>",
		"   /status — Trạng thái hệ thống",
		"   /stats — Thống kê DNS",
		"   /processes — Thông tin tiến trình",
		"   /logs — Truy vấn DNS gần đây",
		"",
		"<b>Điều khiển</b>",
		"   /protection — Bật/tắt bảo vệ DNS",
		"   /ssl — Chứng chỉ SSL/TLS",
		"",
		"<b>Quản lý bộ lọc</b>",
		"   /filters — Xem danh sách bộ lọc",
		"   /filtermgr — Menu quản lý",
		"   /updatefilters — Cập nhật tất cả",
		"   /addlist <code>&lt;url&gt;</code> — Thêm DS chặn",
		"   /addallow <code>&lt;url&gt;</code> — Thêm DS cho phép",
		"   /removelist <code>&lt;url&gt;</code> — Xóa DS chặn",
		"   /removeallow <code>&lt;url&gt;</code> — Xóa DS cho phép",
		"   /enablelist <code>&lt;url&gt;</code> — Kích hoạt",
		"   /disablelist <code>&lt;url&gt;</code> — Vô hiệu",
		"",
		divider(),
		timestampLine(),
	}

	return strings.Join(lines, "\n")
}
