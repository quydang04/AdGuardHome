//go:build !windows

package home

// ensureWritableWorkDir returns the provided workDir unchanged on platforms
// where no special handling is required.
func ensureWritableWorkDir(workDir string) (string, error) {
	return workDir, nil
}
