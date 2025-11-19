package systeminfo

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/netip"
	"os"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Info contains basic system metrics that are safe to query on every platform
// supported by AdGuard Home.
type Info struct {
	OS            string   `json:"os"`
	OSVersion     string   `json:"os_version"`
	Arch          string   `json:"arch"`
	Hostname      string   `json:"hostname"`
	NumCPU        int      `json:"num_cpu"`
	CPUModel      string   `json:"cpu_model"`
	CPUUsage      float64  `json:"cpu_usage"`
	MemoryTotal   uint64   `json:"memory_total"`
	MemoryUsed    uint64   `json:"memory_used"`
	MemoryUsage   float64  `json:"memory_usage"`
	MemoryFree    uint64   `json:"memory_free"`
	DiskPath      string   `json:"disk_path"`
	DiskTotal     uint64   `json:"disk_total"`
	DiskUsed      uint64   `json:"disk_used"`
	DiskUsage     float64  `json:"disk_usage"`
	DiskFree      uint64   `json:"disk_free"`
	LocalIPs      []string `json:"local_ips"`
	PublicIP      string   `json:"public_ip"`
	UptimeSeconds uint64   `json:"uptime_seconds"`
}

// Collect returns a snapshot of the host system metrics.  In case of errors,
// it falls back to zero values for the affected fields while still returning
// any other available information.
func Collect() Info {
	info := Info{
		OS:     runtime.GOOS,
		Arch:   runtime.GOARCH,
		NumCPU: runtime.NumCPU(),
	}

	if hi, err := host.Info(); err == nil {
		info.Hostname = hi.Hostname
		info.UptimeSeconds = hi.Uptime
		if hi.PlatformVersion != "" {
			info.OSVersion = strings.TrimSpace(strings.Join([]string{hi.Platform, hi.PlatformVersion}, " "))
		} else if hi.Platform != "" {
			info.OSVersion = hi.Platform
		}
	}

	if cpuInfos, err := cpu.Info(); err == nil && len(cpuInfos) > 0 {
		info.CPUModel = cpuInfos[0].ModelName
		if info.CPUModel == "" {
			info.CPUModel = fmt.Sprintf("CPU %d", cpuInfos[0].CPU)
		}
	}

	if usages, err := cpu.Percent(0, false); err == nil && len(usages) > 0 {
		info.CPUUsage = usages[0]
	}

	if vm, err := mem.VirtualMemory(); err == nil {
		info.MemoryTotal = vm.Total
		info.MemoryUsed = vm.Used
		info.MemoryUsage = vm.UsedPercent
		// Prefer Available since it accounts for cached memory on Linux.
		if vm.Available > 0 {
			info.MemoryFree = vm.Available
		} else {
			info.MemoryFree = vm.Free
		}
	}

	if du, err := disk.Usage(rootPath()); err == nil {
		info.DiskPath = du.Path
		info.DiskTotal = du.Total
		info.DiskUsed = du.Used
		info.DiskUsage = du.UsedPercent
		info.DiskFree = du.Free
	}

	info.LocalIPs = collectLocalIPs()
	info.PublicIP = lookupPublicIP()

	return info
}

func rootPath() string {
	if runtime.GOOS != "windows" {
		return "/"
	}

	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}

	if strings.HasSuffix(drive, "\\") {
		return drive
	}

	return drive + "\\"
}

func collectLocalIPs() []string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}

	private := make([]string, 0)
	global := make([]string, 0)

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			ip, ok := addrToIP(addr)
			if !ok {
				continue
			}

			if ip.IsLoopback() || ip.IsLinkLocalUnicast() {
				continue
			}

			if ip.IsPrivate() {
				private = append(private, ip.String())

				continue
			}

			if ip.IsGlobalUnicast() {
				global = append(global, ip.String())
			}
		}
	}

	if len(private) > 0 {
		return makeUniqueSorted(private)
	}

	if len(global) > 0 {
		return makeUniqueSorted(global)
	}

	return nil
}

func addrToIP(addr net.Addr) (netip.Addr, bool) {
	var ip net.IP

	switch v := addr.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	default:
		return netip.Addr{}, false
	}

	if ip == nil {
		return netip.Addr{}, false
	}

	parsed, ok := netip.AddrFromSlice(ip)
	if !ok {
		return netip.Addr{}, false
	}

	return parsed.Unmap(), true
}

func makeUniqueSorted(src []string) []string {
	if len(src) == 0 {
		return nil
	}

	sort.Strings(src)

	dst := make([]string, 0, len(src))
	prev := ""
	for _, s := range src {
		if s == prev {
			continue
		}

		dst = append(dst, s)
		prev = s
	}

	return dst
}

const (
	publicIPPrimaryURL   = "https://api64.ipify.org?format=text"
	publicIPSecondaryURL = "https://api.ipify.org?format=text"
	publicIPCacheTTL     = 30 * time.Minute
	publicIPReqTimeout   = 2 * time.Second
)

var (
	publicIPMu      sync.RWMutex
	publicIPValue   string
	publicIPFetched time.Time
)

func lookupPublicIP() string {
	publicIPMu.RLock()
	val := publicIPValue
	fresh := time.Since(publicIPFetched) < publicIPCacheTTL
	publicIPMu.RUnlock()

	if fresh && val != "" {
		return val
	}

	ip := fetchPublicIP(publicIPPrimaryURL)
	if ip == "" {
		ip = fetchPublicIP(publicIPSecondaryURL)
	}

	if ip == "" {
		return val
	}

	publicIPMu.Lock()
	publicIPValue = ip
	publicIPFetched = time.Now()
	publicIPMu.Unlock()

	return ip
}

func fetchPublicIP(url string) string {
	client := http.Client{Timeout: publicIPReqTimeout}

	resp, err := client.Get(url)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 128))
	if err != nil {
		return ""
	}

	text := strings.TrimSpace(string(body))
	if text == "" {
		return ""
	}

	if _, err = netip.ParseAddr(text); err != nil {
		return ""
	}

	return text
}
