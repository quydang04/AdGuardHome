package geoip

import (
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"sync"

	"github.com/oschwald/maxminddb-golang/v2"
)

// CountryRecord is the MaxMind database record for country lookups.
type CountryRecord struct {
	Country struct {
		ISOCode string `maxminddb:"iso_code"`
	} `maxminddb:"country"`
}

// Database provides IP-to-country lookups using a MaxMind-format database.
type Database struct {
	logger *slog.Logger
	reader *maxminddb.Reader
	mu     sync.RWMutex
}

// Config is the configuration for the GeoIP database.
type Config struct {
	Logger *slog.Logger
	Path   string
}

// New creates a new GeoIP database from the specified MMDB file path.  Returns
// nil if the path is empty or the database cannot be opened.
func New(conf *Config) (db *Database) {
	if conf.Path == "" {
		conf.Logger.Info("geoip database path not configured, country tracking disabled")

		return nil
	}

	reader, err := maxminddb.Open(conf.Path)
	if err != nil {
		conf.Logger.Warn("failed to open geoip database, country tracking disabled",
			"path", conf.Path,
			"err", err,
		)

		return nil
	}

	conf.Logger.Info("geoip database loaded", "path", conf.Path)

	return &Database{
		logger: conf.Logger,
		reader: reader,
	}
}

// LookupIP returns the ISO 3166-1 alpha-2 country code for the given IP
// address.  Returns an empty string if the lookup fails.
func (db *Database) LookupIP(ip net.IP) (country string) {
	if db == nil {
		return ""
	}

	addr, ok := netip.AddrFromSlice(ip)
	if !ok {
		return ""
	}

	return db.LookupAddr(addr)
}

// LookupAddr returns the ISO 3166-1 alpha-2 country code for the given netip
// address.
func (db *Database) LookupAddr(addr netip.Addr) (country string) {
	if db == nil {
		return ""
	}

	db.mu.RLock()
	defer db.mu.RUnlock()

	if db.reader == nil {
		return ""
	}

	var record CountryRecord

	err := db.reader.Lookup(addr).Decode(&record)
	if err != nil {
		return ""
	}

	return record.Country.ISOCode
}

// Close closes the GeoIP database.
func (db *Database) Close() (err error) {
	if db == nil {
		return nil
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	if db.reader == nil {
		return nil
	}

	err = db.reader.Close()
	if err != nil {
		return fmt.Errorf("closing geoip database: %w", err)
	}

	db.reader = nil

	return nil
}
