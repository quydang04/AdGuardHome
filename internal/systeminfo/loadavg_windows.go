//go:build windows

package systeminfo

// collectLoadAvg returns zeros on Windows as load average is not available.
func collectLoadAvg() (float64, float64, float64) {
	return 0, 0, 0
}
