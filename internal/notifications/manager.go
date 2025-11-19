package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

const (
	telegramMaxMessageLen = 4096
	defaultCheckInterval  = time.Minute
	defaultCooldown       = time.Minute
	resetFactor           = 0.9
)

// FilterListType specifies whether a list acts as a blocker or allowlist.
type FilterListType string

// Available filter list types.
const (
	FilterListTypeBlock FilterListType = "blocklist"
	FilterListTypeAllow FilterListType = "allowlist"
)

// FilterUpdate describes a freshly refreshed filter or allowlist.
type FilterUpdate struct {
	ID           uint64
	Name         string
	URL          string
	RulesCount   int
	BytesWritten int
	Enabled      bool
	ListType     FilterListType
}

// TelegramConfig contains runtime configuration for Telegram notifications.
type TelegramConfig struct {
	Enabled         bool
	BotToken        string
	ChatID          string
	CPUThreshold    float64
	MemoryThreshold float64
	DiskThreshold   float64
	CheckInterval   time.Duration
	Cooldown        time.Duration
	CustomMessage   string
}

// Manager orchestrates background checks and delivers alerts via Telegram.
type Manager struct {
	logger      *slog.Logger
	mu          sync.RWMutex
	telegram    TelegramConfig
	client      *http.Client
	stopCh      chan struct{}
	wg          sync.WaitGroup
	lastSent    map[string]time.Time
	alertActive map[string]bool
}

// NewManager creates a new notifications manager instance.
func NewManager(l *slog.Logger, cfg TelegramConfig) *Manager {
	if l == nil {
		l = slog.Default()
	}

	cfg = normalizeTelegramConfig(cfg)

	return &Manager{
		logger:      l,
		telegram:    cfg,
		client:      &http.Client{Timeout: 10 * time.Second},
		lastSent:    map[string]time.Time{},
		alertActive: map[string]bool{},
	}
}

// Start launches the monitoring loop once. Subsequent calls are no-ops.
func (m *Manager) Start(ctx context.Context) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopCh != nil {
		return
	}

	stopCh := make(chan struct{})
	m.stopCh = stopCh

	m.wg.Add(1)
	go m.loop(ctx, stopCh)
}

// Stop terminates the monitoring loop and waits for shutdown.
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.stopCh == nil {
		m.mu.Unlock()
		return
	}

	close(m.stopCh)
	m.stopCh = nil
	m.mu.Unlock()

	m.wg.Wait()
}

// UpdateTelegramConfig applies a new Telegram configuration at runtime.
func (m *Manager) UpdateTelegramConfig(cfg TelegramConfig) {
	cfg = normalizeTelegramConfig(cfg)

	m.mu.Lock()
	m.telegram = cfg
	if !cfg.Enabled {
		m.alertActive = map[string]bool{}
	}
	m.mu.Unlock()
}

// SendTelegramTest delivers a test message using the current configuration.
func (m *Manager) SendTelegramTest(ctx context.Context, message string) error {
	cfg := m.getTelegramConfig()
	if cfg.BotToken == "" || cfg.ChatID == "" {
		return fmt.Errorf("telegram configuration incomplete")
	}

	msg := strings.TrimSpace(message)
	if msg == "" {
		msg = "AdGuard Home test notification"
	}

	return m.sendTelegram(ctx, cfg, msg)
}

// NotifyFilterUpdate sends a formatted Telegram message describing a filter
// refresh event.
func (m *Manager) NotifyFilterUpdate(ctx context.Context, update FilterUpdate) {
	cfg := m.getTelegramConfig()
	if !cfg.Enabled || cfg.BotToken == "" || cfg.ChatID == "" {
		return
	}

	info := systeminfo.Collect()
	msg := composeFilterUpdateMessage(cfg, update, info)
	if msg == "" {
		return
	}

	if err := m.sendTelegram(ctx, cfg, msg); err != nil {
		m.logger.Error("telegram filter update failed",
			"list_type", string(update.ListType),
			"name", update.Name,
			slog.String("error", err.Error()),
		)
	}
}

func (m *Manager) loop(ctx context.Context, stop <-chan struct{}) {
	defer m.wg.Done()

	for {
		interval := m.getCheckInterval()
		timer := time.NewTimer(interval)

		select {
		case <-stop:
			timer.Stop()
			return
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			m.runCheck(ctx)
		}
	}
}

func (m *Manager) getCheckInterval() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()

	interval := m.telegram.CheckInterval
	if interval <= 0 {
		interval = defaultCheckInterval
	}

	return interval
}

func (m *Manager) runCheck(ctx context.Context) {
	cfg := m.getTelegramConfig()
	if !cfg.Enabled || cfg.BotToken == "" || cfg.ChatID == "" {
		return
	}

	info := systeminfo.Collect()

	m.handleMetric(ctx, cfg, "cpu", info.CPUUsage, cfg.CPUThreshold, info)
	m.handleMetric(ctx, cfg, "memory", info.MemoryUsage, cfg.MemoryThreshold, info)
	m.handleMetric(ctx, cfg, "disk", info.DiskUsage, cfg.DiskThreshold, info)
}

func (m *Manager) handleMetric(ctx context.Context, cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) {
	if threshold <= 0 || value <= 0 {
		m.clearAlert(metric)
		return
	}

	active, last := m.metricState(metric)
	cooldown := cfg.Cooldown
	if cooldown <= 0 {
		cooldown = defaultCooldown
	}

	if value >= threshold {
		if !active && time.Since(last) >= cooldown {
			if err := m.sendAlert(ctx, cfg, metric, value, threshold, info); err != nil {
				m.logger.Error("telegram alert failed",
					"metric", metric,
					slog.String("error", err.Error()),
				)
			} else {
				now := time.Now()
				m.updateMetricState(metric, true, now)
			}
		}

		return
	}

	if active && value < threshold*resetFactor {
		m.clearAlert(metric)
	}
}

func (m *Manager) sendAlert(ctx context.Context, cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) error {
	message := composeAlertMessage(cfg, metric, value, threshold, info)
	return m.sendTelegram(ctx, cfg, message)
}

func (m *Manager) sendTelegram(ctx context.Context, cfg TelegramConfig, message string) error {
	trimmed := strings.TrimSpace(message)
	if trimmed == "" {
		return nil
	}

	if len(trimmed) > telegramMaxMessageLen {
		trimmed = trimmed[:telegramMaxMessageLen]
	}

	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken)

	data := url.Values{}
	data.Set("chat_id", cfg.ChatID)
	data.Set("text", trimmed)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, telegramMaxMessageLen))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("telegram api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp struct {
		OK          bool   `json:"ok"`
		Description string `json:"description"`
	}

	if len(body) > 0 {
		if err = json.Unmarshal(body, &apiResp); err != nil {
			return fmt.Errorf("decode telegram response: %w", err)
		}
	}

	if !apiResp.OK {
		desc := strings.TrimSpace(apiResp.Description)
		if desc == "" {
			desc = strings.TrimSpace(string(body))
		}
		if desc == "" {
			desc = "unknown telegram error"
		}

		return fmt.Errorf("telegram api error: %s", desc)
	}

	return nil
}

func (m *Manager) metricState(metric string) (bool, time.Time) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.alertActive[metric], m.lastSent[metric]
}

func (m *Manager) updateMetricState(metric string, active bool, ts time.Time) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if active {
		m.alertActive[metric] = true
		m.lastSent[metric] = ts

		return
	}

	delete(m.alertActive, metric)
}

func (m *Manager) clearAlert(metric string) {
	m.updateMetricState(metric, false, time.Time{})
}

func (m *Manager) getTelegramConfig() TelegramConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.telegram
}

func normalizeTelegramConfig(cfg TelegramConfig) TelegramConfig {
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = defaultCheckInterval
	}

	if cfg.Cooldown <= 0 {
		cfg.Cooldown = defaultCooldown
	}

	return cfg
}

func composeAlertMessage(cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) string {
	lines := make([]string, 0, 16)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
	}

	lines = append(lines, fmt.Sprintf("🚨 Alert: %s", alertHeadline(metric)))
	lines = append(lines, "")
	lines = append(lines, "📈 Metrics")
	lines = append(lines, fmt.Sprintf("📏 Metric: %s", metricDisplayName(metric)))
	lines = append(lines, fmt.Sprintf("🔥 Current: %s", formatPercentage(value)))
	lines = append(lines, fmt.Sprintf("🎯 Threshold: %s", formatPercentage(threshold)))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)

	return strings.Join(lines, "\n")
}

func composeFilterUpdateMessage(cfg TelegramConfig, update FilterUpdate, info systeminfo.Info) string {
	lines := make([]string, 0, 20)
	if prefix := strings.TrimSpace(cfg.CustomMessage); prefix != "" {
		lines = append(lines, prefix)
	}

	head := filterUpdateHeader(update.ListType)
	lines = append(lines, head)
	lines = append(lines, fmt.Sprintf("📛 List: %s", fallbackString(update.Name)))
	if update.ID != 0 {
		lines = append(lines, fmt.Sprintf("🆔 ID: #%s", formatUint64(update.ID)))
	}
	lines = append(lines, fmt.Sprintf("🗂️ Type: %s", filterTypeLabel(update.ListType)))
	if update.URL != "" {
		lines = append(lines, fmt.Sprintf("🔗 Source: %s", update.URL))
	}
	rules := update.RulesCount
	if rules < 0 {
		rules = 0
	}
	lines = append(lines, fmt.Sprintf("📊 Rules: %s entries", formatInt64(int64(rules))))
	if update.BytesWritten > 0 {
		lines = append(lines, fmt.Sprintf("📦 Size: %s", formatBytesUint(uint64(update.BytesWritten))))
	}
	statusLabel := "Enabled"
	if !update.Enabled {
		statusLabel = "Disabled"
	}
	lines = append(lines, fmt.Sprintf("⚙️ Status: %s", statusLabel))
	lines = append(lines, "")
	lines = append(lines, systemOverviewLines(info)...)

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
		return "✅ Allowlist Update"
	case FilterListTypeBlock:
		return "🚫 Blocklist Update"
	default:
		return "🔄 Filter Update"
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

var byteUnits = []string{"B", "KB", "MB", "GB", "TB", "PB"}

func systemOverviewLines(info systeminfo.Info) []string {
	lines := []string{"🖥️ System Overview"}
	lines = append(lines, fmt.Sprintf("🏷️ Hostname: %s", fallbackString(info.Hostname)))
	lines = append(lines, fmt.Sprintf("💻 OS: %s", formatOS(info)))
	lines = append(lines, fmt.Sprintf("🧠 CPU: %s", formatCPU(info)))
	lines = append(lines, fmt.Sprintf("🔥 CPU Usage: %s", formatPercentage(info.CPUUsage)))
	lines = append(lines, fmt.Sprintf("🗃️ Memory Usage: %s", formatUsage(info.MemoryUsed, info.MemoryTotal, info.MemoryUsage)))
	lines = append(lines, fmt.Sprintf("📟 Memory Free: %s", formatCapacity(info.MemoryFree, info.MemoryTotal)))
	lines = append(lines, fmt.Sprintf("💽 Disk Usage: %s", formatUsage(info.DiskUsed, info.DiskTotal, info.DiskUsage)))
	lines = append(lines, fmt.Sprintf("📂 Disk Free: %s", formatCapacity(info.DiskFree, info.DiskTotal)))
	lines = append(lines, fmt.Sprintf("📁 Disk Path: %s", fallbackString(info.DiskPath)))
	lines = append(lines, fmt.Sprintf("🌐 Local IPs: %s", formatLocalIPs(info.LocalIPs)))
	lines = append(lines, fmt.Sprintf("🛰️ Public IP: %s", fallbackString(info.PublicIP)))
	uptime := formatUptime(info.UptimeSeconds)
	if uptime == "" {
		uptime = "-"
	}
	lines = append(lines, fmt.Sprintf("⏱️ Uptime: %s", uptime))

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
	m := (seconds % 3600) / 60

	parts := make([]string, 0, 3)
	if d > 0 {
		parts = append(parts, fmt.Sprintf("%dd", d))
	}

	if h > 0 || len(parts) > 0 {
		parts = append(parts, fmt.Sprintf("%dh", h))
	}

	parts = append(parts, fmt.Sprintf("%dm", m))

	return strings.Join(parts, " ")
}
