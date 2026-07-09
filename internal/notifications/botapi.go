package notifications

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type tgUpdate struct {
	UpdateID      int              `json:"update_id"`
	Message       *tgMessage       `json:"message"`
	CallbackQuery *tgCallbackQuery `json:"callback_query"`
}

type tgMessage struct {
	MessageID int64   `json:"message_id"`
	Chat      *tgChat `json:"chat"`
	From      *tgUser `json:"from"`
	Text      string  `json:"text"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type tgCallbackQuery struct {
	ID      string     `json:"id"`
	From    *tgUser    `json:"from"`
	Message *tgMessage `json:"message"`
	Data    string     `json:"data"`
}

type tgUser struct {
	ID        int64  `json:"id"`
	FirstName string `json:"first_name"`
}

type tgInlineKeyboardMarkup struct {
	InlineKeyboard [][]tgInlineKeyboardButton `json:"inline_keyboard"`
}

type tgInlineKeyboardButton struct {
	Text         string `json:"text"`
	CallbackData string `json:"callback_data"`
}

type tgGetUpdatesResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

type tgSendMessageRequest struct {
	ChatID      int64                   `json:"chat_id"`
	Text        string                  `json:"text"`
	ParseMode   string                  `json:"parse_mode,omitempty"`
	ReplyMarkup *tgInlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type tgEditMessageTextRequest struct {
	ChatID      int64                   `json:"chat_id"`
	MessageID   int64                   `json:"message_id"`
	Text        string                  `json:"text"`
	ParseMode   string                  `json:"parse_mode,omitempty"`
	ReplyMarkup *tgInlineKeyboardMarkup `json:"reply_markup,omitempty"`
}

type tgSetMyCommandsRequest struct {
	Commands []tgBotCommand `json:"commands"`
}

type tgBotCommand struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

func (m *Manager) getUpdates(ctx context.Context, cfg TelegramConfig, offset int) ([]tgUpdate, error) {
	endpoint := fmt.Sprintf(
		"https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=30",
		cfg.BotToken,
		offset,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create getUpdates request: %w", err)
	}

	resp, err := m.pollClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("getUpdates request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read getUpdates response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getUpdates status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var apiResp tgGetUpdatesResponse
	if err = json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("decode getUpdates response: %w", err)
	}

	if !apiResp.OK {
		return nil, fmt.Errorf("getUpdates returned ok=false")
	}

	return apiResp.Result, nil
}

func (m *Manager) sendMessageWithKeyboard(ctx context.Context, cfg TelegramConfig, chatID int64, text string, kb *tgInlineKeyboardMarkup) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", cfg.BotToken)

	payload := tgSendMessageRequest{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal sendMessage payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create sendMessage request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("sendMessage request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, telegramMaxMessageLen))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("sendMessage status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

// editMessageWithKeyboard edits an existing message's text and inline keyboard.
func (m *Manager) editMessageWithKeyboard(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, text string, kb *tgInlineKeyboardMarkup) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/editMessageText", cfg.BotToken)

	payload := tgEditMessageTextRequest{
		ChatID:      chatID,
		MessageID:   messageID,
		Text:        text,
		ParseMode:   "HTML",
		ReplyMarkup: kb,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal editMessageText payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create editMessageText request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("editMessageText request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(io.LimitReader(resp.Body, telegramMaxMessageLen))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("editMessageText status %d: %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	return nil
}

// editMessageText edits only the text of an existing message without keyboard.
func (m *Manager) editMessageText(ctx context.Context, cfg TelegramConfig, chatID int64, messageID int64, text string) error {
	return m.editMessageWithKeyboard(ctx, cfg, chatID, messageID, text, nil)
}

func (m *Manager) answerCallbackQuery(ctx context.Context, cfg TelegramConfig, callbackID string) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/answerCallbackQuery", cfg.BotToken)

	payload := map[string]string{"callback_query_id": callbackID}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal answerCallbackQuery payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create answerCallbackQuery request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("answerCallbackQuery request: %w", err)
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	return nil
}

func (m *Manager) deleteWebhook(ctx context.Context, cfg TelegramConfig) error {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/deleteWebhook", cfg.BotToken)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create deleteWebhook request: %w", err)
	}

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("deleteWebhook request: %w", err)
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("deleteWebhook status %d", resp.StatusCode)
	}

	return nil
}

func (m *Manager) registerBotCommands(ctx context.Context, cfg TelegramConfig) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/setMyCommands", cfg.BotToken)

	payload := tgSetMyCommandsRequest{
		Commands: []tgBotCommand{
			{Command: "menu", Description: "Open main menu"},
			{Command: "status", Description: "System status"},
			{Command: "stats", Description: "DNS statistics"},
			{Command: "filters", Description: "Filter lists info"},
			{Command: "protection", Description: "Protection status"},
			{Command: "youtube", Description: "YouTube blocking status"},
			{Command: "processes", Description: "Process info"},
			{Command: "logs", Description: "Recent DNS queries"},
			{Command: "filtermgr", Description: "Manage filter lists"},
			{Command: "updatefilters", Description: "Check for filter updates"},
			{Command: "addlist", Description: "Add blocklist: /addlist <url>"},
			{Command: "addallow", Description: "Add allowlist: /addallow <url>"},
			{Command: "removelist", Description: "Remove blocklist: /removelist <url>"},
			{Command: "removeallow", Description: "Remove allowlist: /removeallow <url>"},
			{Command: "enablelist", Description: "Enable list: /enablelist <url>"},
			{Command: "disablelist", Description: "Disable list: /disablelist <url>"},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		m.logger.Debug("marshal setMyCommands failed", "error", err.Error())

		return
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		m.logger.Debug("create setMyCommands request failed", "error", err.Error())

		return
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req)
	if err != nil {
		m.logger.Debug("setMyCommands request failed", "error", err.Error())

		return
	}
	defer resp.Body.Close()

	io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
}
