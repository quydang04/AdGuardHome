package querylog

import (
	"github.com/AdguardTeam/AdGuardHome/internal/notifications"
)

// GetRecentQueries returns the most recent query log entries up to limit.
func (l *queryLog) GetRecentQueries(limit int) []notifications.QueryLogEntry {
	if limit <= 0 {
		limit = 10
	}

	l.bufferLock.RLock()
	defer l.bufferLock.RUnlock()

	entries := make([]notifications.QueryLogEntry, 0, limit)

	l.buffer.ReverseRange(func(entry *logEntry) (cont bool) {
		if len(entries) >= limit {
			return false
		}

		blocked := entry.Result.IsFiltered
		client := ""
		if entry.IP != nil {
			client = entry.IP.String()
		}
		if entry.ClientID != "" {
			client = entry.ClientID
		}

		entries = append(entries, notifications.QueryLogEntry{
			Domain:   entry.QHost,
			Client:   client,
			Blocked:  blocked,
			Duration: entry.Elapsed,
			Time:     entry.Time,
		})

		return true
	})

	return entries
}
