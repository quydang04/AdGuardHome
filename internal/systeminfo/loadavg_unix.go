//go:build !windows

package systeminfo

import "github.com/shirou/gopsutil/v4/load"

// collectLoadAvg returns 1, 5, and 15 minute load averages on Unix systems.
func collectLoadAvg() (float64, float64, float64) {
	avg, err := load.Avg()
	if err != nil || avg == nil {
		return 0, 0, 0
	}

	return avg.Load1, avg.Load5, avg.Load15
}
