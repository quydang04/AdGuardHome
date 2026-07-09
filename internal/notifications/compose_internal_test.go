package notifications

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/systeminfo"
)

func TestComposeCertExpiryMessage(t *testing.T) {
	cfg := TelegramConfig{}
	info := systeminfo.Info{Hostname: "test-host"}

	ev := CertExpiryReminder{
		Domains:  []string{"example.com", "www.example.com"},
		NotAfter: time.Now().Add(5 * 24 * time.Hour),
		DaysLeft: 5,
	}

	msg := composeCertExpiryMessage(cfg, ev, info)
	if !strings.Contains(msg, "example.com") {
		t.Errorf("expected message to contain domain, got: %s", msg)
	}
	if !strings.Contains(msg, "5") {
		t.Errorf("expected message to contain days left, got: %s", msg)
	}
}

func TestComposeCertRenewalMessage(t *testing.T) {
	cfg := TelegramConfig{}
	info := systeminfo.Info{Hostname: "test-host"}

	t.Run("success", func(t *testing.T) {
		ev := CertRenewalResult{
			Domains:  []string{"example.com"},
			NotAfter: time.Now().Add(90 * 24 * time.Hour),
		}

		msg := composeCertRenewalMessage(cfg, ev, info)
		if !strings.Contains(msg, "AUTO-RENEWED") {
			t.Errorf("expected success message, got: %s", msg)
		}
	})

	t.Run("failure", func(t *testing.T) {
		ev := CertRenewalResult{
			Domains: []string{"example.com"},
			Err:     errors.New("acme: connection refused"),
		}

		msg := composeCertRenewalMessage(cfg, ev, info)
		if !strings.Contains(msg, "FAILED") {
			t.Errorf("expected failure message, got: %s", msg)
		}
		if !strings.Contains(msg, "connection refused") {
			t.Errorf("expected error text in message, got: %s", msg)
		}
	})
}

func TestComposeYouTubeStatusMessage(t *testing.T) {
	testCases := []struct {
		name   string
		status YouTubeStatus
		want   string
	}{{
		name:   "disabled",
		status: YouTubeStatus{Enabled: false},
		want:   "DISABLED",
	}, {
		name:   "enabled_not_active",
		status: YouTubeStatus{Enabled: true, Active: false},
		want:   "not started",
	}, {
		name:   "active",
		status: YouTubeStatus{Enabled: true, Active: true, HealthyIPs: 2, TotalIPs: 2},
		want:   "ACTIVE",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			msg := composeYouTubeStatusMessage(tc.status)
			if !strings.Contains(msg, tc.want) {
				t.Errorf("composeYouTubeStatusMessage(%+v) = %q, want substring %q", tc.status, msg, tc.want)
			}
		})
	}
}

func TestComposeYouTubeAlertMessage(t *testing.T) {
	cfg := TelegramConfig{}
	info := systeminfo.Info{Hostname: "test-host"}

	status := YouTubeStatus{
		Enabled:        true,
		Active:         true,
		HealthyIPs:     0,
		TotalIPs:       4,
		LastSyncStatus: "all ips unhealthy",
	}

	msg := composeYouTubeAlertMessage(cfg, status, info)
	if !strings.Contains(msg, "unreachable") {
		t.Errorf("expected alert message to mention unreachable route server, got: %s", msg)
	}
	if !strings.Contains(msg, "0") || !strings.Contains(msg, "4") {
		t.Errorf("expected message to contain healthy/total IP counts, got: %s", msg)
	}
}

func TestMetricDisplayName(t *testing.T) {
	testCases := []struct {
		metric string
		want   string
	}{
		{metric: "cpu", want: "CPU Usage"},
		{metric: "protection", want: "DNS Protection"},
		{metric: "youtube_health", want: "YouTube Blocking"},
		{metric: "", want: "Metric"},
	}

	for _, tc := range testCases {
		t.Run(tc.metric, func(t *testing.T) {
			got := metricDisplayName(tc.metric)
			if got != tc.want {
				t.Errorf("metricDisplayName(%q) = %q, want %q", tc.metric, got, tc.want)
			}
		})
	}
}

func TestComposeCertStatusMessage(t *testing.T) {
	t.Run("enabled_with_expiry", func(t *testing.T) {
		status := CertStatus{
			Enabled:      true,
			AutoRenew:    true,
			Domains:      []string{"example.com"},
			Challenge:    "http-01",
			NotAfter:     time.Now().Add(60 * 24 * time.Hour),
			LastIssuedAt: time.Now(),
		}

		msg := composeCertStatusMessage(status)
		if !strings.Contains(msg, "example.com") {
			t.Errorf("expected message to contain domain, got: %s", msg)
		}
		if !strings.Contains(msg, "ENABLED") {
			t.Errorf("expected message to show ENABLED, got: %s", msg)
		}
		if !strings.Contains(msg, "ON") {
			t.Errorf("expected message to show auto-renew ON, got: %s", msg)
		}
	})

	t.Run("disabled_with_error", func(t *testing.T) {
		status := CertStatus{
			Enabled:   false,
			AutoRenew: false,
			LastError: "cloudflare api token is required",
		}

		msg := composeCertStatusMessage(status)
		if !strings.Contains(msg, "DISABLED") {
			t.Errorf("expected message to show DISABLED, got: %s", msg)
		}
		if !strings.Contains(msg, "cloudflare api token is required") {
			t.Errorf("expected message to contain last error, got: %s", msg)
		}
	})
}

func TestCertKeyboard(t *testing.T) {
	t.Run("auto_renew_on", func(t *testing.T) {
		kb := certKeyboard(true)
		if !keyboardHasCallback(kb, "cmd:ssl_autorenew_off") {
			t.Errorf("expected keyboard to offer disabling auto-renew: %+v", kb)
		}
	})

	t.Run("auto_renew_off", func(t *testing.T) {
		kb := certKeyboard(false)
		if !keyboardHasCallback(kb, "cmd:ssl_autorenew_on") {
			t.Errorf("expected keyboard to offer enabling auto-renew: %+v", kb)
		}
	})

	kb := certKeyboard(true)
	if !keyboardHasCallback(kb, "cmd:ssl_issue_now") {
		t.Errorf("expected keyboard to offer issuing now: %+v", kb)
	}
}

// keyboardHasCallback reports whether kb has a button with the given
// callback data.
func keyboardHasCallback(kb *tgInlineKeyboardMarkup, callback string) bool {
	for _, row := range kb.InlineKeyboard {
		for _, btn := range row {
			if btn.CallbackData == callback {
				return true
			}
		}
	}

	return false
}
