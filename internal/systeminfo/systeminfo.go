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
	gopsNet "github.com/shirou/gopsutil/v4/net"
	"github.com/shirou/gopsutil/v4/process"
)

// DiskInfo describes usage information for a single disk partition.
type DiskInfo struct {
	Path         string  `json:"path"`
	Total        uint64  `json:"total"`
	Used         uint64  `json:"used"`
	Free         uint64  `json:"free"`
	UsagePercent float64 `json:"usage_percent"`
	Filesystem   string  `json:"filesystem"`
}

// Info contains system metrics that are safe to query on every platform
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

	// Swap memory.
	SwapTotal uint64  `json:"swap_total"`
	SwapUsed  uint64  `json:"swap_used"`
	SwapFree  uint64  `json:"swap_free"`
	SwapUsage float64 `json:"swap_usage"`

	// Host info.
	KernelVersion string `json:"kernel_version"`
	BootTime      uint64 `json:"boot_time"`
	VirtPlatform  string `json:"virt_platform"`

	// Container detection.
	IsContainer bool   `json:"is_container"`
	HostOS      string `json:"host_os,omitempty"`

	// Current server time (RFC 3339).
	SystemTime string `json:"system_time"`

	// Load average (zero on Windows).
	LoadAvg1  float64 `json:"load_avg_1"`
	LoadAvg5  float64 `json:"load_avg_5"`
	LoadAvg15 float64 `json:"load_avg_15"`

	// All disk partitions (filtered, excluding pseudo-fs).
	AllDisks []DiskInfo `json:"all_disks"`

	// Raw disk I/O counters (cumulative totals, Manager computes rates).
	DiskReadBytes  uint64 `json:"disk_read_bytes"`
	DiskWriteBytes uint64 `json:"disk_write_bytes"`
	DiskReadCount  uint64 `json:"disk_read_count"`
	DiskWriteCount uint64 `json:"disk_write_count"`

	// Raw network I/O counters (cumulative totals, Manager computes rates).
	NetBytesSent   uint64 `json:"net_bytes_sent"`
	NetBytesRecv   uint64 `json:"net_bytes_recv"`
	NetPacketsSent uint64 `json:"net_packets_sent"`
	NetPacketsRecv uint64 `json:"net_packets_recv"`
	NetErrorsIn    uint64 `json:"net_errors_in"`
	NetErrorsOut   uint64 `json:"net_errors_out"`
	NetDropsIn     uint64 `json:"net_drops_in"`
	ActiveConns    int    `json:"active_conns"`

	// Process info (self and total).
	TotalProcesses int     `json:"total_processes"`
	SelfCPUPercent float64 `json:"self_cpu_percent"`
	SelfMemBytes   uint64  `json:"self_mem_bytes"`
	SelfOpenFiles  int32   `json:"self_open_files"`
	SelfThreads    int32   `json:"self_threads"`
}

// skipFS contains pseudo-filesystem types that should always be excluded from
// disk enumeration.
var skipFS = map[string]bool{
	"tmpfs":       true,
	"devfs":       true,
	"sysfs":       true,
	"proc":        true,
	"devtmpfs":    true,
	"squashfs":    true,
	"nsfs":        true,
	"cgroup":      true,
	"cgroup2":     true,
	"fuse.lxcfs":  true,
}

// containerFS contains filesystem types used by container runtimes (overlay,
// aufs).  These are skipped unless they are mounted at "/" because in Docker
// the root filesystem is typically an overlay.
var containerFS = map[string]bool{
	"overlay": true,
	"aufs":    true,
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

	containerOS := ""
	if hi, err := host.Info(); err == nil {
		info.Hostname = hi.Hostname
		info.UptimeSeconds = hi.Uptime
		if hi.PlatformVersion != "" {
			containerOS = strings.TrimSpace(strings.Join([]string{hi.Platform, hi.PlatformVersion}, " "))
		} else if hi.Platform != "" {
			containerOS = hi.Platform
		}
		info.OSVersion = containerOS
	}

	// Container detection: check if running inside Docker/LXC/etc.
	info.IsContainer = isContainer()
	if info.IsContainer {
		if hostOS := readHostOSRelease(); hostOS != "" {
			info.HostOS = hostOS
			info.OSVersion = hostOS
		}
	}

	info.SystemTime = time.Now().Format(time.RFC3339)

	// Kernel version.
	if kv, err := host.KernelVersion(); err == nil {
		info.KernelVersion = kv
	}

	// Boot time.
	if bt, err := host.BootTime(); err == nil {
		info.BootTime = bt
	}

	// Virtualization platform.
	if virt, _, err := host.Virtualization(); err == nil && virt != "" {
		info.VirtPlatform = virt
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

	// Swap memory.
	if sw, err := mem.SwapMemory(); err == nil {
		info.SwapTotal = sw.Total
		info.SwapUsed = sw.Used
		info.SwapFree = sw.Free
		info.SwapUsage = sw.UsedPercent
	}

	if du, err := disk.Usage(rootPath()); err == nil {
		info.DiskPath = du.Path
		info.DiskTotal = du.Total
		info.DiskUsed = du.Used
		info.DiskUsage = du.UsedPercent
		info.DiskFree = du.Free
	}

	// All disk partitions.
	info.AllDisks = collectAllDisks()

	// Load average (platform-specific).
	info.LoadAvg1, info.LoadAvg5, info.LoadAvg15 = collectLoadAvg()

	// Disk I/O counters (cumulative).
	if counters, err := disk.IOCounters(); err == nil {
		for _, c := range counters {
			info.DiskReadBytes += c.ReadBytes
			info.DiskWriteBytes += c.WriteBytes
			info.DiskReadCount += c.ReadCount
			info.DiskWriteCount += c.WriteCount
		}
	}

	// Network I/O counters (cumulative, aggregated across all interfaces).
	if netCounters, err := gopsNet.IOCounters(false); err == nil && len(netCounters) > 0 {
		c := netCounters[0]
		info.NetBytesSent = c.BytesSent
		info.NetBytesRecv = c.BytesRecv
		info.NetPacketsSent = c.PacketsSent
		info.NetPacketsRecv = c.PacketsRecv
		info.NetErrorsIn = c.Errin
		info.NetErrorsOut = c.Errout
		info.NetDropsIn = c.Dropin
	}

	// Active TCP connections.
	if conns, err := gopsNet.Connections("tcp"); err == nil {
		info.ActiveConns = len(conns)
	}

	// Process info.
	collectProcessInfo(&info)

	info.LocalIPs = collectLocalIPs()
	info.PublicIP = lookupPublicIP()

	return info
}

// collectAllDisks enumerates physical disk partitions and returns usage info.
func collectAllDisks() []DiskInfo {
	parts, err := disk.Partitions(false)
	if err != nil {
		return nil
	}

	disks := make([]DiskInfo, 0, len(parts))
	seen := make(map[string]bool)

	for _, p := range parts {
		if skipFS[p.Fstype] {
			continue
		}

		if containerFS[p.Fstype] && p.Mountpoint != "/" {
			continue
		}

		if seen[p.Mountpoint] {
			continue
		}
		seen[p.Mountpoint] = true

		du, duErr := disk.Usage(p.Mountpoint)
		if duErr != nil || du.Total == 0 {
			continue
		}

		disks = append(disks, DiskInfo{
			Path:         du.Path,
			Total:        du.Total,
			Used:         du.Used,
			Free:         du.Free,
			UsagePercent: du.UsedPercent,
			Filesystem:   p.Fstype,
		})
	}

	return disks
}

// collectProcessInfo gathers total process count and self-process metrics.
func collectProcessInfo(info *Info) {
	if pids, err := process.Pids(); err == nil {
		info.TotalProcesses = len(pids)
	}

	pid := int32(os.Getpid())
	proc, err := process.NewProcess(pid)
	if err != nil {
		return
	}

	if cpuPct, cpuErr := proc.CPUPercent(); cpuErr == nil {
		info.SelfCPUPercent = cpuPct
	}

	if memInfo, memErr := proc.MemoryInfo(); memErr == nil && memInfo != nil {
		info.SelfMemBytes = memInfo.RSS
	}

	if fds, fdErr := proc.NumFDs(); fdErr == nil {
		info.SelfOpenFiles = fds
	}

	if threads, tErr := proc.NumThreads(); tErr == nil {
		info.SelfThreads = threads
	}
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
