package stats

import (
	"fmt"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghnet"
)

// UpdateConfig replaces the runtime configuration of the statistics with the
// given values.  It does not write the configuration file.
func (s *StatsCtx) UpdateConfig(
	enabled bool,
	limit time.Duration,
	ignored []string,
	ignoredEnabled bool,
) (err error) {
	err = validateIvl(limit)
	if err != nil {
		return fmt.Errorf("unsupported interval: %w", err)
	}

	engine, err := aghnet.NewIgnoreEngine(ignored, ignoredEnabled)
	if err != nil {
		return fmt.Errorf("ignored: %w", err)
	}

	s.confMu.Lock()
	defer s.confMu.Unlock()

	s.ignored = engine
	s.limit = limit
	s.enabled = enabled

	return nil
}
