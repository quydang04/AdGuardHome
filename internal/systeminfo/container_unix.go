//go:build !windows

package systeminfo

import (
	"bufio"
	"os"
	"strings"
)

// isContainer reports whether the process is running inside a container
// (Docker, LXC, Podman, etc.).
func isContainer() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	f, err := os.Open("/proc/1/cgroup")
	if err != nil {
		return false
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "docker") ||
			strings.Contains(line, "lxc") ||
			strings.Contains(line, "kubepods") ||
			strings.Contains(line, "containerd") {
			return true
		}
	}

	return false
}

// resolveHostHostname attempts to determine the real host machine's hostname
// when running inside a container.  It checks mounted host files and
// environment variables, since os.Hostname inside a container returns the
// container ID.
func resolveHostHostname() string {
	if v := os.Getenv("HOST_HOSTNAME"); v != "" {
		return strings.TrimSpace(v)
	}

	if v := os.Getenv("HOSTNAME_OVERRIDE"); v != "" {
		return strings.TrimSpace(v)
	}

	for _, path := range []string{"/host/etc/hostname", "/etc/host_hostname"} {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		if name := strings.TrimSpace(string(data)); name != "" {
			return name
		}
	}

	return ""
}

// hostOSReleasePaths lists paths where the host's os-release may be mounted
// into a container.
var hostOSReleasePaths = []string{
	"/host/etc/os-release",
	"/host/usr/lib/os-release",
}

// readHostOSRelease tries to read the host OS identification from a mounted
// os-release file.  Returns an empty string when unavailable.
func readHostOSRelease() string {
	for _, path := range hostOSReleasePaths {
		if name := parseOSRelease(path); name != "" {
			return name
		}
	}

	return ""
}

// parseOSRelease reads an os-release file and returns the PRETTY_NAME value,
// falling back to NAME + VERSION if PRETTY_NAME is absent.
func parseOSRelease(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	var prettyName, name, version string

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}

		v = strings.Trim(v, `"`)

		switch k {
		case "PRETTY_NAME":
			prettyName = v
		case "NAME":
			name = v
		case "VERSION":
			version = v
		}
	}

	if prettyName != "" {
		return prettyName
	}

	if name != "" {
		if version != "" {
			return name + " " + version
		}

		return name
	}

	return ""
}
