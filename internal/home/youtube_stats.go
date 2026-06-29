package home

import (
	"cmp"
	"slices"
	"sync"
	"time"
)

const (
	youtubeStatsMaxHours = 24
	youtubeStatsMaxTopN  = 10
)

// youtubeQueryStats tracks per-query statistics for the YouTube ad blocking
// feature.
type youtubeQueryStats struct {
	mu sync.Mutex

	totalQueries        int64
	blockedAdQueries    int64
	blockedTrackQueries int64
	rewrittenQueries    int64

	blockedDomainCounts map[string]int64
	rewriteDomainCounts map[string]int64

	hourlyBlocked   [youtubeStatsMaxHours]int64
	hourlyRewritten [youtubeStatsMaxHours]int64
	currentHourSlot int
	lastHourUpdate  time.Time

	startedAt time.Time
}

// youtubeTopDomain is a single domain entry in the top domains list.
type youtubeTopDomain struct {
	Domain string `json:"domain"`
	Count  int64  `json:"count"`
	Type   string `json:"type"`
}

// youtubeStatsResponse is the JSON response for the YouTube query stats API.
type youtubeStatsResponse struct {
	TotalQueries        int64              `json:"total_youtube_queries"`
	BlockedAdQueries    int64              `json:"blocked_ad_queries"`
	BlockedTrackQueries int64              `json:"blocked_tracking_queries"`
	RewrittenQueries    int64              `json:"rewritten_queries"`
	TopBlockedDomains   []youtubeTopDomain `json:"top_blocked_domains"`
	TopRewriteDomains   []youtubeTopDomain `json:"top_rewrite_domains"`
	HourlyBlocked       []int64            `json:"hourly_blocked"`
	HourlyRewritten     []int64            `json:"hourly_rewritten"`
	QueryRate           float64            `json:"query_rate_per_min"`
	BlockRate           float64            `json:"block_rate_percent"`
}

// newYoutubeQueryStats creates and initialises a new youtubeQueryStats.
func newYoutubeQueryStats() *youtubeQueryStats {
	return &youtubeQueryStats{
		blockedDomainCounts: make(map[string]int64),
		rewriteDomainCounts: make(map[string]int64),
		startedAt:           time.Now(),
		lastHourUpdate:      time.Now(),
	}
}

// RecordQuery records a single DNS query result.  queryType must be one of
// "ad", "tracking", or "rewrite".
func (s *youtubeQueryStats) RecordQuery(domain, queryType string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rotateHourlySlotLocked()

	s.totalQueries++
	switch queryType {
	case "ad":
		s.blockedAdQueries++
		s.blockedDomainCounts[domain]++
		s.hourlyBlocked[s.currentHourSlot]++
	case "tracking":
		s.blockedTrackQueries++
		s.blockedDomainCounts[domain]++
		s.hourlyBlocked[s.currentHourSlot]++
	case "rewrite":
		s.rewrittenQueries++
		s.rewriteDomainCounts[domain]++
		s.hourlyRewritten[s.currentHourSlot]++
	}
}

// rotateHourlySlotLocked advances the hourly ring-buffer slots if time has
// passed.  It must be called with s.mu held.
func (s *youtubeQueryStats) rotateHourlySlotLocked() {
	now := time.Now()
	hoursSince := int(now.Sub(s.lastHourUpdate).Hours())
	if hoursSince <= 0 {
		return
	}

	for i := 0; i < hoursSince && i < youtubeStatsMaxHours; i++ {
		s.currentHourSlot = (s.currentHourSlot + 1) % youtubeStatsMaxHours
		s.hourlyBlocked[s.currentHourSlot] = 0
		s.hourlyRewritten[s.currentHourSlot] = 0
	}

	s.lastHourUpdate = now
}

// topDomainsFromMap returns the top n domains from a count map, sorted
// descending by count.
func topDomainsFromMap(m map[string]int64, domType string, n int) []youtubeTopDomain {
	type kv struct {
		k string
		v int64
	}

	pairs := make([]kv, 0, len(m))
	for k, v := range m {
		pairs = append(pairs, kv{k, v})
	}

	slices.SortFunc(pairs, func(a, b kv) int { return cmp.Compare(b.v, a.v) })

	if len(pairs) > n {
		pairs = pairs[:n]
	}

	result := make([]youtubeTopDomain, len(pairs))
	for i, p := range pairs {
		result[i] = youtubeTopDomain{Domain: p.k, Count: p.v, Type: domType}
	}

	return result
}

// getStats returns a snapshot of the current statistics.
func (s *youtubeQueryStats) getStats() youtubeStatsResponse {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.rotateHourlySlotLocked()

	hourlyBlocked := make([]int64, youtubeStatsMaxHours)
	hourlyRewritten := make([]int64, youtubeStatsMaxHours)
	for i := 0; i < youtubeStatsMaxHours; i++ {
		idx := (s.currentHourSlot - i + youtubeStatsMaxHours) % youtubeStatsMaxHours
		hourlyBlocked[i] = s.hourlyBlocked[idx]
		hourlyRewritten[i] = s.hourlyRewritten[idx]
	}

	topBlocked := topDomainsFromMap(s.blockedDomainCounts, "blocked", youtubeStatsMaxTopN)
	topRewrite := topDomainsFromMap(s.rewriteDomainCounts, "rewrite", youtubeStatsMaxTopN)

	elapsed := time.Since(s.startedAt).Minutes()

	var queryRate float64
	if elapsed > 0 {
		queryRate = float64(s.totalQueries) / elapsed
	}

	var blockRate float64
	if s.totalQueries > 0 {
		blockRate = float64(s.blockedAdQueries+s.blockedTrackQueries) / float64(s.totalQueries) * 100
	}

	return youtubeStatsResponse{
		TotalQueries:        s.totalQueries,
		BlockedAdQueries:    s.blockedAdQueries,
		BlockedTrackQueries: s.blockedTrackQueries,
		RewrittenQueries:    s.rewrittenQueries,
		TopBlockedDomains:   topBlocked,
		TopRewriteDomains:   topRewrite,
		HourlyBlocked:       hourlyBlocked,
		HourlyRewritten:     hourlyRewritten,
		QueryRate:           queryRate,
		BlockRate:           blockRate,
	}
}

// resetStats zeroes all counters and restarts the timer.
func (s *youtubeQueryStats) resetStats() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.totalQueries = 0
	s.blockedAdQueries = 0
	s.blockedTrackQueries = 0
	s.rewrittenQueries = 0
	s.blockedDomainCounts = make(map[string]int64)
	s.rewriteDomainCounts = make(map[string]int64)
	s.hourlyBlocked = [youtubeStatsMaxHours]int64{}
	s.hourlyRewritten = [youtubeStatsMaxHours]int64{}
	s.currentHourSlot = 0
	s.lastHourUpdate = time.Now()
	s.startedAt = time.Now()
}
