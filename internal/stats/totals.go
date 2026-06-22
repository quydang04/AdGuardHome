package stats

// GetCurrentStats returns aggregate DNS query statistics over the configured
// retention period.
func (s *StatsCtx) GetCurrentStats() (numQueries, numBlocked, numSafeBrowsing, numParental uint64, avgProcessingTime float64) {
	var resp *StatsResp
	var ok bool

	func() {
		s.confMu.RLock()
		defer s.confMu.RUnlock()

		resp, ok = s.getData(uint32(s.limit.Hours()))
	}()

	if !ok || resp == nil {
		return 0, 0, 0, 0, 0
	}

	return resp.NumDNSQueries,
		resp.NumBlockedFiltering,
		resp.NumReplacedSafebrowsing,
		resp.NumReplacedParental,
		resp.AvgProcessingTime
}
