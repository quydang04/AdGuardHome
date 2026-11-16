package querylog

import (
	"fmt"
	"time"

	"github.com/AdguardTeam/AdGuardHome/internal/aghnet"
)

// ImportedConfig contains the query log settings that can be replaced at
// runtime, e.g. when importing a settings file.
type ImportedConfig struct {
	// Ignored is the list of ignored hosts.
	Ignored []string

	// RotationIvl is the log rotation interval.  It must pass [validateIvl].
	RotationIvl time.Duration

	// MemSize is the number of entries kept in memory.  Zero means unchanged.
	MemSize uint

	// Enabled enables the query log.
	Enabled bool

	// FileEnabled enables writing the log to a file.
	FileEnabled bool

	// AnonymizeClientIP enables anonymization of client IP addresses.
	AnonymizeClientIP bool

	// IgnoredEnabled enables the Ignored list.
	IgnoredEnabled bool
}

// UpdateConfig replaces the runtime configuration of the query log with c.  It
// does not write the configuration file.
func (l *queryLog) UpdateConfig(c ImportedConfig) (err error) {
	err = validateIvl(c.RotationIvl)
	if err != nil {
		return fmt.Errorf("unsupported interval: %w", err)
	}

	engine, err := aghnet.NewIgnoreEngine(c.Ignored, c.IgnoredEnabled)
	if err != nil {
		return fmt.Errorf("ignored: %w", err)
	}

	l.confMu.Lock()
	defer l.confMu.Unlock()

	conf := *l.conf
	conf.Ignored = engine
	conf.RotationIvl = c.RotationIvl
	conf.Enabled = c.Enabled
	conf.FileEnabled = c.FileEnabled
	conf.AnonymizeClientIP = c.AnonymizeClientIP
	if c.MemSize > 0 {
		conf.MemSize = c.MemSize
	}

	if conf.AnonymizeClientIP {
		l.anonymizer.Store(AnonymizeIP)
	} else {
		l.anonymizer.Store(nil)
	}

	l.conf = &conf

	return nil
}
