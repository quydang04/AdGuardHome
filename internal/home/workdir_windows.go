//go:build windows

package home

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"syscall"

	"github.com/AdguardTeam/AdGuardHome/internal/aghos"
)

// ensureWritableWorkDir verifies that workDir can be used for mutable data and
// falls back to a per-user writable directory when it is not.
func ensureWritableWorkDir(workDir string) (string, error) {
	writable, err := isDirWritable(workDir)
	if err != nil {
		return "", fmt.Errorf("checking writability of %q: %w", workDir, err)
	}

	if writable {
		return workDir, nil
	}

	fallback, err := resolveFallbackWorkDir()
	if err != nil {
		return "", fmt.Errorf("resolving fallback work dir: %w", err)
	}

	if err = os.MkdirAll(fallback, aghos.DefaultPermDir); err != nil {
		return "", fmt.Errorf("creating fallback work dir: %w", err)
	}

	fallbackWritable, err := isDirWritable(fallback)
	if err != nil {
		return "", fmt.Errorf("checking fallback writability: %w", err)
	}

	if !fallbackWritable {
		return "", fmt.Errorf("fallback work dir %q is not writable", fallback)
	}

	if err = migrateConfigToFallback(workDir, fallback); err != nil {
		return "", err
	}

	resolved, err := filepath.EvalSymlinks(fallback)
	if err != nil {
		return "", fmt.Errorf("resolving fallback symlinks: %w", err)
	}

	fallbackWorkDirUsed.Store(true)

	return resolved, nil
}

// resolveFallbackWorkDir returns the per-user directory used when the default
// workDir is not writable.
func resolveFallbackWorkDir() (string, error) {
	if custom := os.Getenv("ADGUARDHOME_WORKDIR"); custom != "" {
		return custom, nil
	}

	if local := os.Getenv("LOCALAPPDATA"); local != "" {
		return filepath.Join(local, "AdGuardHome"), nil
	}

	if roaming := os.Getenv("APPDATA"); roaming != "" {
		return filepath.Join(roaming, "AdGuardHome"), nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining user home: %w", err)
	}

	return filepath.Join(homeDir, "AppData", "Local", "AdGuardHome"), nil
}

// isDirWritable attempts to create and remove a temporary file to determine if
// dir can be written to by the current user.
func isDirWritable(dir string) (bool, error) {
	f, err := os.CreateTemp(dir, "agh-writetest-*")
	if err != nil {
		if errors.Is(err, fs.ErrPermission) {
			return false, nil
		}

		var pathErr *fs.PathError
		if errors.As(err, &pathErr) {
			switch {
			case errors.Is(pathErr.Err, syscall.ERROR_ACCESS_DENIED):
				return false, nil
			case errors.Is(pathErr.Err, syscall.ERROR_PATH_NOT_FOUND):
				return false, nil
			}
		}

		return false, err
	}

	name := f.Name()

	if closeErr := f.Close(); closeErr != nil {
		_ = os.Remove(name)

		return false, closeErr
	}

	if rmErr := os.Remove(name); rmErr != nil {
		return false, rmErr
	}

	return true, nil
}

// migrateConfigToFallback copies the configuration file from the original
// workDir to fallbackWorkDir if the latter does not have one yet.
func migrateConfigToFallback(workDir, fallbackWorkDir string) error {
	src := filepath.Join(workDir, "AdGuardHome.yaml")

	_, err := os.Stat(src)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}

		return fmt.Errorf("checking existing config at %q: %w", src, err)
	}

	dst := filepath.Join(fallbackWorkDir, "AdGuardHome.yaml")

	if _, err = os.Stat(dst); err == nil {
		return nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("checking fallback config at %q: %w", dst, err)
	}

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("reading existing config from %q: %w", src, err)
	}

	if err = os.WriteFile(dst, data, aghos.DefaultPermFile); err != nil {
		return fmt.Errorf("writing config to fallback %q: %w", dst, err)
	}

	return nil
}
