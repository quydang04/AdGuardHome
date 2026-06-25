//go:build windows

package systeminfo

// isContainer returns false on Windows as container detection is not supported.
func isContainer() bool {
	return false
}

// readHostOSRelease returns an empty string on Windows.
func readHostOSRelease() string {
	return ""
}

// resolveHostHostname returns an empty string on Windows.
func resolveHostHostname() string {
	return ""
}
