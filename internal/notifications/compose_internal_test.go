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
