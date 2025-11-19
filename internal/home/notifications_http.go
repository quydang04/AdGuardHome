package home

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghhttp"
	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
	"github.com/AdguardTeam/golibs/timeutil"
)

var registerNotificationHandlersOnce sync.Once

const (
	minTelegramInterval = time.Minute
	maxTelegramInterval = 24 * time.Hour
	minTelegramCooldown = time.Minute
	maxTelegramCooldown = 24 * time.Hour
)

type telegramConfigJSON struct {
	Enabled         bool    `json:"enabled"`
	BotToken        string  `json:"bot_token"`
	ChatID          string  `json:"chat_id"`
	CPUThreshold    float64 `json:"cpu_threshold"`
	MemoryThreshold float64 `json:"memory_threshold"`
	DiskThreshold   float64 `json:"disk_threshold"`
	CheckInterval   int64   `json:"check_interval"`
	Cooldown        int64   `json:"cooldown"`
	CustomMessage   string  `json:"custom_message"`
}

func (web *webAPI) registerNotificationHandlers() {
	registerNotificationHandlersOnce.Do(func() {
		web.httpReg.Register(http.MethodGet, "/control/notifications/telegram", web.handleGetTelegramConfig)
		web.httpReg.Register(http.MethodPut, "/control/notifications/telegram", web.handlePutTelegramConfig)
		web.httpReg.Register(http.MethodPost, "/control/notifications/telegram/test", web.handlePostTelegramTest)
	})
}

func (web *webAPI) handleGetTelegramConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var resp telegramConfigJSON
	func() {
		config.RLock()
		defer config.RUnlock()

		resp = telegramConfigToJSON(config.Notifications.Telegram)
	}()

	aghhttp.WriteJSONResponseOK(ctx, web.logger, w, r, resp)
}

func (web *webAPI) handlePutTelegramConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	req := telegramConfigJSON{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusBadRequest, "json decode: %s", err)

		return
	}

	cfg, err := telegramConfigFromJSON(&req)
	if err != nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusUnprocessableEntity, "%s", err)

		return
	}

	var (
		changed    bool
		runtimeCfg notifications.TelegramConfig
	)

	func() {
		config.Lock()
		defer config.Unlock()

		current := config.Notifications.Telegram
		if current == nil {
			current = defaultTelegramConfig()
			config.Notifications.Telegram = current
		}

		changed = !telegramConfigEqual(current, cfg)

		*current = *cfg
		current.applyDefaults()

		runtimeCfg = buildRuntimeTelegramConfig(current)
	}()

	if changed {
		web.logger.InfoContext(ctx, "telegram notifications updated", "enabled", runtimeCfg.Enabled)
		web.confModifier.Apply(ctx)
	}

	if globalContext.notifier != nil {
		globalContext.notifier.UpdateTelegramConfig(runtimeCfg)
	}

	aghhttp.OK(ctx, web.logger, w)
}

func (web *webAPI) handlePostTelegramTest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if globalContext.notifier == nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusServiceUnavailable, "notifications manager unavailable")

		return
	}

	var req struct {
		Message string `json:"message"`
	}

	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusBadRequest, "json decode: %s", err)

		return
	}

	if err := globalContext.notifier.SendTelegramTest(ctx, req.Message); err != nil {
		aghhttp.ErrorAndLog(ctx, web.logger, r, w, http.StatusBadGateway, "telegram test failed: %s", err)

		return
	}

	aghhttp.OK(ctx, web.logger, w)
}

func telegramConfigToJSON(cfg *telegramConfig) telegramConfigJSON {
	if cfg == nil {
		cfg = defaultTelegramConfig()
	}

	return telegramConfigJSON{
		Enabled:         cfg.Enabled,
		BotToken:        cfg.BotToken,
		ChatID:          cfg.ChatID,
		CPUThreshold:    cfg.CPUThreshold,
		MemoryThreshold: cfg.MemoryThreshold,
		DiskThreshold:   cfg.DiskThreshold,
		CheckInterval:   int64(time.Duration(cfg.CheckInterval) / time.Millisecond),
		Cooldown:        int64(time.Duration(cfg.Cooldown) / time.Millisecond),
		CustomMessage:   cfg.CustomMessage,
	}
}

func telegramConfigFromJSON(j *telegramConfigJSON) (*telegramConfig, error) {
	if j == nil {
		return nil, fmt.Errorf("empty payload")
	}

	check := time.Duration(j.CheckInterval) * time.Millisecond
	if check < minTelegramInterval || check > maxTelegramInterval {
		return nil, fmt.Errorf("check_interval must be between %s and %s", minTelegramInterval, maxTelegramInterval)
	}

	cooldown := time.Duration(j.Cooldown) * time.Millisecond
	if cooldown < minTelegramCooldown || cooldown > maxTelegramCooldown {
		return nil, fmt.Errorf("cooldown must be between %s and %s", minTelegramCooldown, maxTelegramCooldown)
	}

	for key, value := range map[string]float64{
		"cpu":    j.CPUThreshold,
		"memory": j.MemoryThreshold,
		"disk":   j.DiskThreshold,
	} {
		if value < 0 || value > 100 {
			return nil, fmt.Errorf("%s threshold must be between 0 and 100", key)
		}
	}

	cfg := &telegramConfig{
		Enabled:         j.Enabled,
		BotToken:        strings.TrimSpace(j.BotToken),
		ChatID:          strings.TrimSpace(j.ChatID),
		CPUThreshold:    j.CPUThreshold,
		MemoryThreshold: j.MemoryThreshold,
		DiskThreshold:   j.DiskThreshold,
		CheckInterval:   timeutil.Duration(check),
		Cooldown:        timeutil.Duration(cooldown),
		CustomMessage:   strings.TrimSpace(j.CustomMessage),
	}

	if cfg.Enabled && (cfg.BotToken == "" || cfg.ChatID == "") {
		return nil, fmt.Errorf("bot_token and chat_id are required when notifications are enabled")
	}

	cfg.applyDefaults()

	return cfg, nil
}

func telegramConfigEqual(a, b *telegramConfig) bool {
	return a.Enabled == b.Enabled &&
		a.BotToken == b.BotToken &&
		a.ChatID == b.ChatID &&
		a.CPUThreshold == b.CPUThreshold &&
		a.MemoryThreshold == b.MemoryThreshold &&
		a.DiskThreshold == b.DiskThreshold &&
		a.CheckInterval == b.CheckInterval &&
		a.Cooldown == b.Cooldown &&
		a.CustomMessage == b.CustomMessage
}

func buildRuntimeTelegramConfig(cfg *telegramConfig) notifications.TelegramConfig {
	return notifications.TelegramConfig{
		Enabled:         cfg.Enabled,
		BotToken:        cfg.BotToken,
		ChatID:          cfg.ChatID,
		CPUThreshold:    cfg.CPUThreshold,
		MemoryThreshold: cfg.MemoryThreshold,
		DiskThreshold:   cfg.DiskThreshold,
		CheckInterval:   time.Duration(cfg.CheckInterval),
		Cooldown:        time.Duration(cfg.Cooldown),
		CustomMessage:   cfg.CustomMessage,
	}
}
