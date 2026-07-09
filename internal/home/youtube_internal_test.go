package home

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatUptimeDHMS(t *testing.T) {
	testCases := []struct {
		name string
		in   time.Duration
		want string
	}{{
		name: "seconds",
		in:   32 * time.Second,
		want: "32s",
	}, {
		name: "minutes_seconds",
		in:   17*time.Minute + 32*time.Second,
		want: "17m32s",
	}, {
		name: "hours_minutes_seconds",
		in:   2*time.Hour + 5*time.Minute + 1*time.Second,
		want: "2h5m1s",
	}, {
		name: "just_under_a_day",
		in:   23*time.Hour + 59*time.Minute + 59*time.Second,
		want: "23h59m59s",
	}, {
		name: "exactly_a_day",
		in:   24 * time.Hour,
		want: "1d0h0m0s",
	}, {
		name: "a_day_and_change",
		in:   25*time.Hour + 3*time.Minute + 2*time.Second,
		want: "1d1h3m2s",
	}, {
		name: "multiple_days",
		in:   50*time.Hour + 30*time.Minute,
		want: "2d2h30m0s",
	}}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, formatUptimeDHMS(tc.in))
		})
	}
}
