package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
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
	rateLimitWindow       = time.Minute
	rateLimitMax          = 10
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

// ioSnapshot holds cumulative I/O counters for delta computation.
type ioSnapshot struct {
	diskReadBytes  uint64
	diskWriteBytes uint64
	diskReadCount  uint64
	diskWriteCount uint64
	netBytesSent   uint64
	netBytesRecv   uint64
	netPacketsSent uint64
	netPacketsRecv uint64
}

// Manager orchestrates background checks and delivers alerts via Telegram.
type Manager struct {
	logger      *slog.Logger
	mu          sync.RWMutex
	telegram    TelegramConfig
	client      *http.Client
	pollClient  *http.Client
	stopCh      chan struct{}
	wg          sync.WaitGroup
	lastSent    map[string]time.Time
	alertActive map[string]bool
	startTime   time.Time
	stats       StatsProvider
	filters     FilterProvider
	filterMgr   FilterManager
	protection  ProtectionProvider

	logs LogsProvider

	// Recovery alert support.
	alertStartTime map[string]time.Time

	// I/O snapshot for delta computation.
	lastIOSnapshot   *ioSnapshot
	lastIOSnapshotAt time.Time

	// Computed I/O rates (updated each check cycle).
	diskReadBytesPerSec  uint64
	diskWriteBytesPerSec uint64
	diskReadIOPS         uint64
	diskWriteIOPS        uint64
	netBytesSentPerSec   uint64
	netBytesRecvPerSec   uint64
	netPacketsSentPerSec uint64
	netPacketsRecvPerSec uint64

	// Rate limiting: timestamps of recent commands per chat.
	cmdTimestamps []time.Time
}

// NewManager creates a new notifications manager instance.
func NewManager(l *slog.Logger, cfg TelegramConfig) *Manager {
	if l == nil {
		l = slog.Default()
	}

	cfg = normalizeTelegramConfig(cfg)

	return &Manager{
		logger:         l,
		telegram:       cfg,
		client:         &http.Client{Timeout: 10 * time.Second},
		pollClient:     &http.Client{Timeout: 35 * time.Second},
		lastSent:       map[string]time.Time{},
		alertActive:    map[string]bool{},
		alertStartTime: map[string]time.Time{},
		startTime:      time.Now(),
		cmdTimestamps:  make([]time.Time, 0, rateLimitMax),
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

	if m.telegram.BotToken != "" {
		m.wg.Add(1)
		go m.pollLoop(ctx, stopCh)
	}
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
	oldToken := m.telegram.BotToken
	m.telegram = cfg
	if !cfg.Enabled {
		m.alertActive = map[string]bool{}
		m.alertStartTime = map[string]time.Time{}
	}
	m.mu.Unlock()

	if cfg.Enabled && cfg.BotToken != "" && cfg.BotToken != oldToken {
		go m.registerBotCommands(context.Background(), cfg)
	}
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

	if err := m.sendTelegramWithRetry(ctx, cfg, msg); err != nil {
		m.logger.Error("telegram filter update failed",
			"list_type", string(update.ListType),
			"name", update.Name,
			slog.String("error", err.Error()),
		)
	}
}

// ValidateBotToken checks whether the given token is valid by calling getMe.
func (m *Manager) ValidateBotToken(ctx context.Context, token string) (string, error) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", fmt.Errorf("create getMe request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("getMe request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("getMe status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var result struct {
		OK     bool `json:"ok"`
		Result struct {
			Username string `json:"username"`
		} `json:"result"`
	}

	if err = json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("decode getMe response: %w", err)
	}

	if !result.OK {
		return "", fmt.Errorf("getMe returned ok=false")
	}

	return result.Result.Username, nil
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

	// Update I/O rates from delta.
	m.updateIOSnapshot(info)

	m.handleMetric(ctx, cfg, "cpu", info.CPUUsage, cfg.CPUThreshold, info)
	m.handleMetric(ctx, cfg, "memory", info.MemoryUsage, cfg.MemoryThreshold, info)
	m.handleMetric(ctx, cfg, "disk", info.DiskUsage, cfg.DiskThreshold, info)

	// Check protection status.
	m.checkProtectionAlert(ctx, cfg)
}

// checkProtectionAlert sends an alert if DNS protection is disabled.
func (m *Manager) checkProtectionAlert(ctx context.Context, cfg TelegramConfig) {
	m.mu.RLock()
	pp := m.protection
	m.mu.RUnlock()

	if pp == nil {
		return
	}

	if !pp.IsProtectionEnabled() {
		m.mu.RLock()
		alreadyAlerted := m.alertActive["protection"]
		m.mu.RUnlock()

		if !alreadyAlerted {
			msg := "🚨 DNS Protection is DISABLED!\n\nDNS filtering is currently turned off. All queries pass through unfiltered."
			if err := m.sendTelegramWithRetry(ctx, cfg, msg); err != nil {
				m.logger.Error("telegram protection alert failed", slog.String("error", err.Error()))
			} else {
				m.mu.Lock()
				m.alertActive["protection"] = true
				m.alertStartTime["protection"] = time.Now()
				m.mu.Unlock()
			}
		}
	} else {
		m.mu.RLock()
		wasActive := m.alertActive["protection"]
		m.mu.RUnlock()

		if wasActive {
			m.clearAlertWithRecovery(ctx, cfg, "protection", 0, 0)
		}
	}
}

// updateIOSnapshot computes I/O rates from the delta between current and
// previous snapshots.
func (m *Manager) updateIOSnapshot(info systeminfo.Info) {
	current := &ioSnapshot{
		diskReadBytes:  info.DiskReadBytes,
		diskWriteBytes: info.DiskWriteBytes,
		diskReadCount:  info.DiskReadCount,
		diskWriteCount: info.DiskWriteCount,
		netBytesSent:   info.NetBytesSent,
		netBytesRecv:   info.NetBytesRecv,
		netPacketsSent: info.NetPacketsSent,
		netPacketsRecv: info.NetPacketsRecv,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if m.lastIOSnapshot != nil && !m.lastIOSnapshotAt.IsZero() {
		elapsed := time.Since(m.lastIOSnapshotAt).Seconds()
		if elapsed > 0 {
			prev := m.lastIOSnapshot
			m.diskReadBytesPerSec = rateDelta(current.diskReadBytes, prev.diskReadBytes, elapsed)
			m.diskWriteBytesPerSec = rateDelta(current.diskWriteBytes, prev.diskWriteBytes, elapsed)
			m.diskReadIOPS = rateDelta(current.diskReadCount, prev.diskReadCount, elapsed)
			m.diskWriteIOPS = rateDelta(current.diskWriteCount, prev.diskWriteCount, elapsed)
			m.netBytesSentPerSec = rateDelta(current.netBytesSent, prev.netBytesSent, elapsed)
			m.netBytesRecvPerSec = rateDelta(current.netBytesRecv, prev.netBytesRecv, elapsed)
			m.netPacketsSentPerSec = rateDelta(current.netPacketsSent, prev.netPacketsSent, elapsed)
			m.netPacketsRecvPerSec = rateDelta(current.netPacketsRecv, prev.netPacketsRecv, elapsed)
		}
	}

	m.lastIOSnapshot = current
	m.lastIOSnapshotAt = time.Now()
}

// rateDelta computes a per-second rate from cumulative counter deltas.
func rateDelta(current, previous uint64, elapsedSec float64) uint64 {
	if current <= previous || elapsedSec <= 0 {
		return 0
	}

	return uint64(float64(current-previous) / elapsedSec)
}

// IOStats returns the most recently computed I/O rates.
type IOStats struct {
	DiskReadBytesPerSec  uint64
	DiskWriteBytesPerSec uint64
	DiskReadIOPS         uint64
	DiskWriteIOPS        uint64
	NetBytesSentPerSec   uint64
	NetBytesRecvPerSec   uint64
	NetPacketsSentPerSec uint64
	NetPacketsRecvPerSec uint64
}

// GetIOStats returns the latest computed I/O rates.
func (m *Manager) GetIOStats() IOStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return IOStats{
		DiskReadBytesPerSec:  m.diskReadBytesPerSec,
		DiskWriteBytesPerSec: m.diskWriteBytesPerSec,
		DiskReadIOPS:         m.diskReadIOPS,
		DiskWriteIOPS:        m.diskWriteIOPS,
		NetBytesSentPerSec:   m.netBytesSentPerSec,
		NetBytesRecvPerSec:   m.netBytesRecvPerSec,
		NetPacketsSentPerSec: m.netPacketsSentPerSec,
		NetPacketsRecvPerSec: m.netPacketsRecvPerSec,
	}
}

func (m *Manager) handleMetric(ctx context.Context, cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) {
	if threshold <= 0 || value <= 0 {
		m.clearAlertWithRecovery(ctx, cfg, metric, value, threshold)
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

				m.mu.Lock()
				m.alertStartTime[metric] = now
				m.mu.Unlock()
			}
		}

		return
	}

	if active && value < threshold*resetFactor {
		m.clearAlertWithRecovery(ctx, cfg, metric, value, threshold)
	}
}

func (m *Manager) sendAlert(ctx context.Context, cfg TelegramConfig, metric string, value, threshold float64, info systeminfo.Info) error {
	message := composeAlertMessage(cfg, metric, value, threshold, info)
	return m.sendTelegramWithRetry(ctx, cfg, message)
}

// sendTelegramWithRetry attempts to send a message with exponential backoff.
func (m *Manager) sendTelegramWithRetry(ctx context.Context, cfg TelegramConfig, msg string) error {
	delays := []time.Duration{1 * time.Second, 3 * time.Second, 10 * time.Second}
	var lastErr error

	// First attempt without delay.
	if lastErr = m.sendTelegram(ctx, cfg, msg); lastErr == nil {
		return nil
	}

	for _, delay := range delays {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		if lastErr = m.sendTelegram(ctx, cfg, msg); lastErr == nil {
			return nil
		}
	}

	return lastErr
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
	data.Set("parse_mode", "HTML")

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

// clearAlertWithRecovery clears the alert and sends a recovery notification
// if the alert was previously active.
func (m *Manager) clearAlertWithRecovery(ctx context.Context, cfg TelegramConfig, metric string, currentValue, threshold float64) {
	m.mu.RLock()
	wasActive := m.alertActive[metric]
	startTime := m.alertStartTime[metric]
	m.mu.RUnlock()

	if wasActive {
		duration := time.Since(startTime).Truncate(time.Second)
		msg := composeRecoveryMessage(metric, currentValue, threshold, duration)
		if err := m.sendTelegramWithRetry(ctx, cfg, msg); err != nil {
			m.logger.Debug("telegram recovery message failed", slog.String("error", err.Error()))
		}
	}

	m.mu.Lock()
	delete(m.alertActive, metric)
	delete(m.alertStartTime, metric)
	m.mu.Unlock()
}

func (m *Manager) clearAlert(metric string) {
	m.updateMetricState(metric, false, time.Time{})
}

func (m *Manager) getTelegramConfig() TelegramConfig {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.telegram
}

// isRateLimited checks whether the command rate limit has been exceeded.
func (m *Manager) isRateLimited() bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-rateLimitWindow)

	// Prune old timestamps.
	valid := m.cmdTimestamps[:0]
	for _, ts := range m.cmdTimestamps {
		if ts.After(cutoff) {
			valid = append(valid, ts)
		}
	}
	m.cmdTimestamps = valid

	if len(m.cmdTimestamps) >= rateLimitMax {
		return true
	}

	m.cmdTimestamps = append(m.cmdTimestamps, now)
	return false
}

func (m *Manager) pollLoop(ctx context.Context, stop <-chan struct{}) {
	defer m.wg.Done()

	var offset int

	for {
		cfg := m.getTelegramConfig()
		if !cfg.Enabled || cfg.BotToken == "" {
			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-time.After(10 * time.Second):
				continue
			}
		}

		updates, err := m.getUpdates(ctx, cfg, offset)
		if err != nil {
			m.logger.Debug("telegram poll error", slog.String("error", err.Error()))

			select {
			case <-stop:
				return
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}

			continue
		}

		for _, u := range updates {
			m.handleUpdate(ctx, cfg, u)
			offset = u.UpdateID + 1
		}
	}
}

func (m *Manager) handleUpdate(ctx context.Context, cfg TelegramConfig, u tgUpdate) {
	var chatID int64
	var command string
	var fullText string
	var messageID int64

	if u.CallbackQuery != nil {
		if u.CallbackQuery.Message != nil && u.CallbackQuery.Message.Chat != nil {
			chatID = u.CallbackQuery.Message.Chat.ID
			messageID = u.CallbackQuery.Message.MessageID
		}
		command = u.CallbackQuery.Data
		fullText = command

		_ = m.answerCallbackQuery(ctx, cfg, u.CallbackQuery.ID)
	} else if u.Message != nil && u.Message.Chat != nil {
		chatID = u.Message.Chat.ID
		fullText = strings.TrimSpace(u.Message.Text)
		command = extractCommand(fullText)
	} else {
		return
	}

	if fmt.Sprintf("%d", chatID) != cfg.ChatID {
		return
	}

	// Rate limiting.
	if m.isRateLimited() {
		_ = m.sendTelegram(ctx, cfg, "Warning: Too many requests, please try again later.")
		return
	}

	switch command {
	case "/start", "/menu", "cmd:menu":
		m.sendMainMenu(ctx, cfg, chatID, messageID)
	case "/status", "cmd:status":
		m.sendSystemStatus(ctx, cfg, chatID, messageID)
	case "/stats", "cmd:stats":
		m.sendDNSStats(ctx, cfg, chatID, messageID)
	case "/filters", "cmd:filters":
		m.sendFilterInfo(ctx, cfg, chatID, messageID)
	case "/protection", "cmd:protection":
		m.sendProtectionStatus(ctx, cfg, chatID, messageID)
	case "/processes", "cmd:processes":
		m.sendProcessInfo(ctx, cfg, chatID, messageID)
	case "/logs", "cmd:logs":
		m.sendRecentLogs(ctx, cfg, chatID, messageID)
	case "/filtermgr", "cmd:filtermgr":
		m.sendFilterManage(ctx, cfg, chatID, messageID)
	case "cmd:protection_on":
		m.toggleProtection(ctx, cfg, chatID, messageID, true)
	case "cmd:protection_off":
		m.toggleProtection(ctx, cfg, chatID, messageID, false)
	case "cmd:filtermgr_addblock":
		m.sendFilterManageHelp(ctx, cfg, chatID, messageID, "addblock")
	case "cmd:filtermgr_addallow":
		m.sendFilterManageHelp(ctx, cfg, chatID, messageID, "addallow")
	case "cmd:filtermgr_rmblock":
		m.sendFilterListSelection(ctx, cfg, chatID, messageID, "rm", "b")
	case "cmd:filtermgr_rmallow":
		m.sendFilterListSelection(ctx, cfg, chatID, messageID, "rm", "a")
	case "cmd:filtermgr_enable":
		m.sendFilterListSelection(ctx, cfg, chatID, messageID, "en", "")
	case "cmd:filtermgr_disable":
		m.sendFilterListSelection(ctx, cfg, chatID, messageID, "dis", "")
	case "/updatefilters", "cmd:filtermgr_update":
		m.handleRefreshFilters(ctx, cfg, chatID, messageID)
	default:
		if strings.HasPrefix(command, "flt:") {
			m.handleFilterAction(ctx, cfg, chatID, messageID, command)
		} else {
			m.handleTextCommand(ctx, cfg, chatID, fullText)
		}
	}
}

// extractCommand strips the @BotUsername suffix and any trailing arguments
// from a Telegram command. For example, "/status@MyBot extra" becomes "/status".
func extractCommand(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	if !strings.HasPrefix(text, "/") {
		return text
	}

	// Take only the first word (the command itself).
	cmd, _, _ := strings.Cut(text, " ")

	// Strip @BotUsername suffix from the command.
	cmd, _, _ = strings.Cut(cmd, "@")

	return cmd
}

func (m *Manager) sendMainMenu(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	info := systeminfo.Collect()

	protOn := true
	m.mu.RLock()
	pp := m.protection
	m.mu.RUnlock()

	if pp != nil {
		protOn = pp.IsProtectionEnabled()
	}

	text := composeMainMenuMessage(info, protOn)
	kb := mainMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			m.logger.Debug("edit main menu failed, sending new", slog.String("error", err.Error()))
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send main menu failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendSystemStatus(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	info := systeminfo.Collect()
	ioStats := m.GetIOStats()
	text := composeSystemStatusMessage(info, ioStats)
	kb := backToMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send system status failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendDNSStats(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	m.mu.RLock()
	sp := m.stats
	m.mu.RUnlock()

	var text string
	if sp == nil {
		text = "DNS Statistics\n---\nData not available"
	} else {
		nq, nb, nsb, np, avg := sp.GetCurrentStats()
		text = composeDNSStatsMessage(nq, nb, nsb, np, avg)
	}

	kb := backToMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send dns stats failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendFilterInfo(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	m.mu.RLock()
	fp := m.filters
	m.mu.RUnlock()

	var text string
	if fp == nil {
		text = "Filter Information\n---\nData not available"
	} else {
		totalRules, _, _ := fp.GetFilterSummary()
		blockLists, allowLists := fp.GetFilterDetails()
		text = composeFilterDetailedMessage(totalRules, blockLists, allowLists)
	}

	kb := backToMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send filter info failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendProtectionStatus(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	m.mu.RLock()
	pp := m.protection
	m.mu.RUnlock()

	var text string
	var kb *tgInlineKeyboardMarkup

	if pp == nil {
		text = "Protection Status\n---\nData not available"
		kb = backToMenuKeyboard()
	} else {
		enabled := pp.IsProtectionEnabled()
		text = composeProtectionStatusMessage(enabled)
		kb = protectionKeyboard(enabled)
	}

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send protection status failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendProcessInfo(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	info := systeminfo.Collect()
	text := composeProcessInfoMessage(m.startTime, info)
	kb := backToMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send process info failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendRecentLogs(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	m.mu.RLock()
	lp := m.logs
	m.mu.RUnlock()

	var text string
	if lp == nil {
		text = "<b>Recent Queries</b>\n---\nQuery log not available"
	} else {
		entries := lp.GetRecentQueries(10)
		text = composeRecentLogsMessage(entries)
	}

	kb := backToMenuKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send recent logs failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendFilterManage(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	text := composeFilterManageMessage()
	kb := filterManageKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send filter manage failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendFilterManageHelp(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, action string) {
	text := composeFilterManageHelpMessage(action)
	kb := filterManageKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send filter manage help failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) sendFilterListSelection(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, action string, listType string) {
	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		text := "Filter management not available."
		kb := filterManageKeyboard()
		if messageID > 0 {
			_ = m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb)
		} else {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}

		return
	}

	blockLists, allowLists := fm.GetFilterDetails()

	var title string
	switch action {
	case "rm":
		if listType == "a" {
			title = "Remove Allowlist"
		} else {
			title = "Remove Blocklist"
		}
	case "en":
		title = "Enable List"
	case "dis":
		title = "Disable List"
	}

	text, kb := composeFilterSelectionMessage(title, action, listType, blockLists, allowLists)

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send filter list selection failed", slog.String("error", err.Error()))
		}
	}
}

func (m *Manager) handleFilterAction(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, callback string) {
	// Format: flt:<action>:<listType>:<index>
	// e.g. flt:rm:b:0, flt:en:a:1, flt:dis:b:2
	parts := strings.Split(callback, ":")
	if len(parts) != 4 {
		return
	}

	action := parts[1]
	listType := parts[2]
	idx, err := strconv.Atoi(parts[3])
	if err != nil || idx < 0 {
		return
	}

	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		_ = m.sendTelegram(ctx, cfg, "Filter management not available.")

		return
	}

	blockLists, allowLists := fm.GetFilterDetails()

	whitelist := listType == "a"
	var lists []FilterListInfo
	if whitelist {
		lists = allowLists
	} else {
		lists = blockLists
	}

	if idx >= len(lists) {
		_ = m.sendTelegram(ctx, cfg, "List not found (index out of range). The list may have changed.")

		return
	}

	target := lists[idx]
	var resultMsg string

	switch action {
	case "rm":
		if err = fm.RemoveFilterList(target.URL, whitelist); err != nil {
			resultMsg = fmt.Sprintf("<b>Failed to remove</b>\n%s\n\nError: %s", target.Name, err.Error())
		} else {
			resultMsg = fmt.Sprintf("<b>List removed</b>\n\nName: %s\nURL: %s", target.Name, target.URL)
		}
	case "en":
		if err = fm.EnableFilterList(target.URL, true, whitelist); err != nil {
			resultMsg = fmt.Sprintf("<b>Failed to enable</b>\n%s\n\nError: %s", target.Name, err.Error())
		} else {
			resultMsg = fmt.Sprintf("<b>List enabled</b>\n\nName: %s\nURL: %s", target.Name, target.URL)
		}
	case "dis":
		if err = fm.EnableFilterList(target.URL, false, whitelist); err != nil {
			resultMsg = fmt.Sprintf("<b>Failed to disable</b>\n%s\n\nError: %s", target.Name, err.Error())
		} else {
			resultMsg = fmt.Sprintf("<b>List disabled</b>\n\nName: %s\nURL: %s", target.Name, target.URL)
		}
	default:
		return
	}

	resultMsg += "\n\n" + timestampLine()
	kb := filterManageKeyboard()

	if messageID > 0 {
		if err2 := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, resultMsg, kb); err2 != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, resultMsg, kb)
		}
	} else {
		_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, resultMsg, kb)
	}
}

func (m *Manager) handleRefreshFilters(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64) {
	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		text := "Filter management not available."
		kb := filterManageKeyboard()
		if messageID > 0 {
			_ = m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb)
		} else {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}

		return
	}

	updated, ok := fm.RefreshFilters()

	var text string
	if !ok {
		text = "<b>Filter Update</b>\n---\nAnother refresh is already in progress. Please try again later."
	} else {
		text = fmt.Sprintf("<b>Filter Update</b>\n---\nRefresh completed.\nUpdated lists: <code>%d</code>\n\n%s", updated, timestampLine())
	}

	kb := filterManageKeyboard()

	if messageID > 0 {
		if err := m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, kb); err != nil {
			_ = m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb)
		}
	} else {
		if err := m.sendMessageWithKeyboard(ctx, cfg, chatID, text, kb); err != nil {
			m.logger.Debug("send filter refresh result failed", slog.String("error", err.Error()))
		}
	}
}

// handleTextCommand handles parameterized text commands like /addlist <url>.
func (m *Manager) handleTextCommand(ctx context.Context, cfg TelegramConfig, chatID int64, text string) {
	parts := strings.Fields(text)
	if len(parts) == 0 {
		return
	}

	// Strip @BotName suffix from command.
	cmdRaw := parts[0]
	cmdRaw, _, _ = strings.Cut(cmdRaw, "@")
	cmd := strings.ToLower(cmdRaw)

	switch cmd {
	case "/addlist":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /addlist &lt;url1&gt; [url2] [url3] ...\nOr: /addlist &lt;url&gt; | &lt;name&gt;")
			return
		}
		m.handleAddLists(ctx, cfg, chatID, parts[1:], false)

	case "/addallow":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /addallow &lt;url1&gt; [url2] [url3] ...\nOr: /addallow &lt;url&gt; | &lt;name&gt;")
			return
		}
		m.handleAddLists(ctx, cfg, chatID, parts[1:], true)

	case "/removelist":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /removelist &lt;url&gt;")
			return
		}
		m.handleRemoveList(ctx, cfg, chatID, parts[1], false)

	case "/removeallow":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /removeallow &lt;url&gt;")
			return
		}
		m.handleRemoveList(ctx, cfg, chatID, parts[1], true)

	case "/enablelist":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /enablelist &lt;url&gt;")
			return
		}
		m.handleEnableList(ctx, cfg, chatID, parts[1], true, false)

	case "/disablelist":
		if len(parts) < 2 {
			_ = m.sendTelegram(ctx, cfg, "Usage: /disablelist &lt;url&gt;")
			return
		}
		m.handleEnableList(ctx, cfg, chatID, parts[1], false, false)

	}
}

// handleAddLists adds one or more filter lists. Supports formats:
//   - /addlist url1 url2 url3 (multiple URLs, no names)
//   - /addlist url | name (single URL with a custom name, pipe-separated)
func (m *Manager) handleAddLists(ctx context.Context, cfg TelegramConfig, chatID int64, args []string, whitelist bool) {
	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		_ = m.sendTelegram(ctx, cfg, "Filter management not available.")
		return
	}

	listType := "blocklist"
	if whitelist {
		listType = "allowlist"
	}

	// Check if using pipe syntax for name: "url | name"
	joined := strings.Join(args, " ")
	if strings.Contains(joined, "|") {
		parts := strings.SplitN(joined, "|", 2)
		listURL := strings.TrimSpace(parts[0])
		name := ""
		if len(parts) > 1 {
			name = strings.TrimSpace(parts[1])
		}

		if err := fm.AddFilterList(listURL, name, whitelist); err != nil {
			msg := fmt.Sprintf("Failed to add %s: %s", listType, err.Error())
			_ = m.sendTelegram(ctx, cfg, msg)
			return
		}

		displayName := name
		if displayName == "" {
			displayName = listURL
		}
		msg := fmt.Sprintf("<b>%s added</b>\n\nName: %s\nURL: %s", capitalizeFirst(listType), displayName, listURL)
		_ = m.sendTelegram(ctx, cfg, msg)
		return
	}

	// Multiple URLs mode: each arg is a URL.
	results := make([]string, 0, len(args))
	successCount := 0
	for _, listURL := range args {
		listURL = strings.TrimSpace(listURL)
		if listURL == "" {
			continue
		}

		if err := fm.AddFilterList(listURL, "", whitelist); err != nil {
			results = append(results, fmt.Sprintf("FAIL %s: %s", listURL, err.Error()))
		} else {
			results = append(results, fmt.Sprintf("OK %s", listURL))
			successCount++
		}
	}

	header := fmt.Sprintf("<b>Add %s results</b> (%d/%d succeeded)\n", capitalizeFirst(listType), successCount, len(args))
	msg := header + "\n" + strings.Join(results, "\n")
	_ = m.sendTelegram(ctx, cfg, msg)
}

func (m *Manager) handleRemoveList(ctx context.Context, cfg TelegramConfig, chatID int64, listURL string, whitelist bool) {
	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		_ = m.sendTelegram(ctx, cfg, "Filter management not available.")
		return
	}

	listType := "blocklist"
	if whitelist {
		listType = "allowlist"
	}

	if err := fm.RemoveFilterList(listURL, whitelist); err != nil {
		msg := fmt.Sprintf("Failed to remove %s: %s", listType, err.Error())
		_ = m.sendTelegram(ctx, cfg, msg)
		return
	}

	msg := fmt.Sprintf("<b>%s removed</b>\n\nURL: %s", capitalizeFirst(listType), listURL)
	_ = m.sendTelegram(ctx, cfg, msg)
}

func (m *Manager) handleEnableList(ctx context.Context, cfg TelegramConfig, chatID int64, listURL string, enabled, whitelist bool) {
	m.mu.RLock()
	fm := m.filterMgr
	m.mu.RUnlock()

	if fm == nil {
		_ = m.sendTelegram(ctx, cfg, "Filter management not available.")
		return
	}

	if err := fm.EnableFilterList(listURL, enabled, whitelist); err != nil {
		action := "enable"
		if !enabled {
			action = "disable"
		}
		msg := fmt.Sprintf("Failed to %s list: %s", action, err.Error())
		_ = m.sendTelegram(ctx, cfg, msg)
		return
	}

	action := "enabled"
	if !enabled {
		action = "disabled"
	}
	msg := fmt.Sprintf("<b>Filter list %s</b>\n\nURL: %s", action, listURL)
	_ = m.sendTelegram(ctx, cfg, msg)
}

// toggleProtection enables or disables DNS protection via the bot.
func (m *Manager) toggleProtection(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, enable bool) {
	m.mu.RLock()
	pp := m.protection
	m.mu.RUnlock()

	if pp == nil {
		_ = m.sendTelegram(ctx, cfg, "Protection provider not available.")
		return
	}

	if err := pp.SetProtectionEnabled(enable); err != nil {
		msg := fmt.Sprintf("Failed to change protection: %s", err.Error())
		_ = m.sendTelegram(ctx, cfg, msg)
		return
	}

	var msg string
	if enable {
		msg = "DNS Protection has been <b>ENABLED</b>"
	} else {
		msg = "DNS Protection has been <b>DISABLED</b>"
	}

	if messageID > 0 {
		_ = m.editMessageText(ctx, cfg, chatID, messageID, msg)
	} else {
		_ = m.sendTelegram(ctx, cfg, msg)
	}

	// Refresh protection status display.
	m.sendProtectionStatus(ctx, cfg, chatID, 0)
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
