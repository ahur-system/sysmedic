package monitor

import (
	"fmt"
	"io/fs"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ahur-system/sysmedic/internal/config"
)

// SystemMetrics represents overall system resource usage
type SystemMetrics struct {
	Timestamp    time.Time
	CPUPercent   float64
	MemoryPercent float64
	NetworkMBps  float64
	LoadAvg1     float64
	LoadAvg5     float64
	LoadAvg15    float64
}

// UserMetrics represents per-user resource usage
type UserMetrics struct {
	Username      string
	CPUPercent    float64
	MemoryPercent float64
	ProcessCount  int
	Timestamp     time.Time
	PIDs          []int
}

// PersistentUser tracks users with sustained high usage
type PersistentUser struct {
	Username    string
	Metric      string // "cpu" or "memory"
	StartTime   time.Time
	Duration    time.Duration
	CurrentUsage float64
}

// Monitor handles system and user monitoring
type Monitor struct {
	lastCPUTimes map[string]CPUTimes
	lastNetStats NetworkStats
	lastSampleTime time.Time
	config Config
}

// Config interface for monitor configuration
type Config interface {
	GetUserFiltering() config.UserFilteringConfig
}

// CPUTimes represents CPU time statistics
type CPUTimes struct {
	User   uint64
	Nice   uint64
	System uint64
	Idle   uint64
	IOWait uint64
	IRQ    uint64
	SoftIRQ uint64
	Total  uint64
}

// NetworkStats represents network interface statistics
type NetworkStats struct {
	BytesReceived    uint64
	BytesTransmitted uint64
}

// ProcessInfo represents information about a running process
type ProcessInfo struct {
	PID      int
	Username string
	CPUTime  uint64
	MemoryKB uint64
	Command  string
}

// NewMonitor creates a new system monitor
func NewMonitor(config Config) *Monitor {
	return &Monitor{
		lastCPUTimes: make(map[string]CPUTimes),
		config:       config,
	}
}

// GetSystemMetrics collects current system metrics
func (m *Monitor) GetSystemMetrics() (*SystemMetrics, error) {
	metrics := &SystemMetrics{
		Timestamp: time.Now(),
	}

	// Get CPU usage
	cpuPercent, err := m.getCPUUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU usage: %w", err)
	}
	metrics.CPUPercent = cpuPercent

	// Get memory usage
	memPercent, err := m.getMemoryUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get memory usage: %w", err)
	}
	metrics.MemoryPercent = memPercent

	// Get network usage
	networkMBps, err := m.getNetworkUsage()
	if err != nil {
		return nil, fmt.Errorf("failed to get network usage: %w", err)
	}
	metrics.NetworkMBps = networkMBps

	// Get load averages
	loadAvg1, loadAvg5, loadAvg15, err := m.getLoadAverages()
	if err != nil {
		return nil, fmt.Errorf("failed to get load averages: %w", err)
	}
	metrics.LoadAvg1 = loadAvg1
	metrics.LoadAvg5 = loadAvg5
	metrics.LoadAvg15 = loadAvg15

	return metrics, nil
}

// GetUserMetrics collects per-user resource usage for real users only
func (m *Monitor) GetUserMetrics() ([]UserMetrics, error) {
	processes, err := m.getProcesses()
	if err != nil {
		return nil, fmt.Errorf("failed to get processes: %w", err)
	}

	// Group processes by user, excluding system users
	userProcesses := make(map[string][]ProcessInfo)
	for _, proc := range processes {
		if m.isRealUser(proc.Username) {
			userProcesses[proc.Username] = append(userProcesses[proc.Username], proc)
		}
	}

	// Get total system resources for percentage calculations
	_, err = m.getTotalMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get total memory: %w", err)
	}

	var userMetrics []UserMetrics
	for username, procs := range userProcesses {
		var totalCPUTime uint64
		var totalMemoryKB uint64
		var pids []int

		for _, proc := range procs {
			totalCPUTime += proc.CPUTime
			totalMemoryKB += proc.MemoryKB
			pids = append(pids, proc.PID)
		}

		// Calculate CPU percentage (simplified)
		cpuPercent := float64(totalCPUTime) / 1000.0 // Rough approximation
		if cpuPercent > 100 {
			cpuPercent = 100
		}

		// Calculate memory percentage (simplified for demo)
		memoryPercent := float64(totalMemoryKB) / 1048576.0 // Convert to rough percentage
		if memoryPercent > 100 {
			memoryPercent = 100
		}

		userMetrics = append(userMetrics, UserMetrics{
			Username:      username,
			CPUPercent:    cpuPercent,
			MemoryPercent: memoryPercent,
			ProcessCount:  len(procs),
			Timestamp:     time.Now(),
			PIDs:          pids,
		})
	}

	// Filter to only include users with significant activity
	var significantUsers []UserMetrics
	for _, user := range userMetrics {
		if m.shouldTrackUser(user) {
			significantUsers = append(significantUsers, user)
		}
	}

	// Sort by CPU usage descending
	sort.Slice(significantUsers, func(i, j int) bool {
		return significantUsers[i].CPUPercent > significantUsers[j].CPUPercent
	})

	return significantUsers, nil
}

// getCPUUsage calculates CPU usage percentage
func (m *Monitor) getCPUUsage() (float64, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	if len(lines) == 0 {
		return 0, fmt.Errorf("no CPU data found")
	}

	// Parse first line (overall CPU)
	fields := strings.Fields(lines[0])
	if len(fields) < 8 || fields[0] != "cpu" {
		return 0, fmt.Errorf("invalid CPU data format")
	}

	var times CPUTimes
	times.User, _ = strconv.ParseUint(fields[1], 10, 64)
	times.Nice, _ = strconv.ParseUint(fields[2], 10, 64)
	times.System, _ = strconv.ParseUint(fields[3], 10, 64)
	times.Idle, _ = strconv.ParseUint(fields[4], 10, 64)
	times.IOWait, _ = strconv.ParseUint(fields[5], 10, 64)
	times.IRQ, _ = strconv.ParseUint(fields[6], 10, 64)
	times.SoftIRQ, _ = strconv.ParseUint(fields[7], 10, 64)

	times.Total = times.User + times.Nice + times.System + times.Idle +
		times.IOWait + times.IRQ + times.SoftIRQ

	// Calculate usage percentage
	if lastTimes, exists := m.lastCPUTimes["cpu"]; exists {
		totalDiff := times.Total - lastTimes.Total
		idleDiff := times.Idle - lastTimes.Idle

		if totalDiff > 0 {
			usage := 100.0 * (1.0 - float64(idleDiff)/float64(totalDiff))
			m.lastCPUTimes["cpu"] = times
			return usage, nil
		}
	}

	m.lastCPUTimes["cpu"] = times
	return 0, nil
}

// getMemoryUsage calculates memory usage percentage
func (m *Monitor) getMemoryUsage() (float64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	var memTotal, memFree, buffers, cached uint64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "MemTotal:":
			memTotal, _ = strconv.ParseUint(fields[1], 10, 64)
		case "MemFree:":
			memFree, _ = strconv.ParseUint(fields[1], 10, 64)
		case "Buffers:":
			buffers, _ = strconv.ParseUint(fields[1], 10, 64)
		case "Cached:":
			cached, _ = strconv.ParseUint(fields[1], 10, 64)
		}
	}

	if memTotal == 0 {
		return 0, fmt.Errorf("could not read memory total")
	}

	memUsed := memTotal - memFree - buffers - cached
	return float64(memUsed) / float64(memTotal) * 100, nil
}

// getNetworkUsage calculates network usage in MB/s
func (m *Monitor) getNetworkUsage() (float64, error) {
	data, err := os.ReadFile("/proc/net/dev")
	if err != nil {
		return 0, err
	}

	var totalRx, totalTx uint64
	lines := strings.Split(string(data), "\n")

	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}

		// Skip loopback interface
		if strings.Contains(line, "lo:") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 10 {
			continue
		}

		rx, err1 := strconv.ParseUint(fields[1], 10, 64)
		tx, err2 := strconv.ParseUint(fields[9], 10, 64)

		if err1 == nil && err2 == nil {
			totalRx += rx
			totalTx += tx
		}
	}

	currentStats := NetworkStats{
		BytesReceived:    totalRx,
		BytesTransmitted: totalTx,
	}

	// Calculate rate if we have previous data
	now := time.Now()
	if !m.lastSampleTime.IsZero() {
		elapsed := now.Sub(m.lastSampleTime).Seconds()
		if elapsed > 0 {
			rxDiff := currentStats.BytesReceived - m.lastNetStats.BytesReceived
			txDiff := currentStats.BytesTransmitted - m.lastNetStats.BytesTransmitted

			totalBytes := rxDiff + txDiff
			mbps := float64(totalBytes) / (1024 * 1024) / elapsed

			m.lastNetStats = currentStats
			m.lastSampleTime = now
			return mbps, nil
		}
	}

	m.lastNetStats = currentStats
	m.lastSampleTime = now
	return 0, nil
}

// getLoadAverages reads system load averages
func (m *Monitor) getLoadAverages() (float64, float64, float64, error) {
	data, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return 0, 0, 0, err
	}

	fields := strings.Fields(string(data))
	if len(fields) < 3 {
		return 0, 0, 0, fmt.Errorf("invalid loadavg format")
	}

	load1, err1 := strconv.ParseFloat(fields[0], 64)
	load5, err2 := strconv.ParseFloat(fields[1], 64)
	load15, err3 := strconv.ParseFloat(fields[2], 64)

	if err1 != nil || err2 != nil || err3 != nil {
		return 0, 0, 0, fmt.Errorf("failed to parse load averages")
	}

	return load1, load5, load15, nil
}

// getTotalMemory returns total system memory in KB
func (m *Monitor) getTotalMemory() (uint64, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return 0, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "MemTotal:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return strconv.ParseUint(fields[1], 10, 64)
			}
		}
	}

	return 0, fmt.Errorf("MemTotal not found in /proc/meminfo")
}

// getProcesses returns all running processes with resource usage
func (m *Monitor) getProcesses() ([]ProcessInfo, error) {
	var processes []ProcessInfo

	err := filepath.WalkDir("/proc", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // Skip errors, continue walking
		}

		if !d.IsDir() {
			return nil
		}

		// Check if directory name is a PID
		name := d.Name()
		pid, err := strconv.Atoi(name)
		if err != nil {
			return nil // Not a PID directory
		}

		// Only process first level /proc/[pid] directories
		if filepath.Dir(path) != "/proc" {
			return nil
		}

		proc, err := m.getProcessInfo(pid)
		if err != nil {
			return nil // Skip processes we can't read
		}

		processes = append(processes, *proc)
		return nil
	})

	return processes, err
}

// getProcessInfo gets detailed information about a specific process
func (m *Monitor) getProcessInfo(pid int) (*ProcessInfo, error) {
	// Read /proc/[pid]/stat for basic info
	statPath := fmt.Sprintf("/proc/%d/stat", pid)
	statData, err := os.ReadFile(statPath)
	if err != nil {
		return nil, err
	}

	statFields := strings.Fields(string(statData))
	if len(statFields) < 24 {
		return nil, fmt.Errorf("invalid stat format for PID %d", pid)
	}

	// Parse CPU time (user + system time in clock ticks)
	utime, _ := strconv.ParseUint(statFields[13], 10, 64)
	stime, _ := strconv.ParseUint(statFields[14], 10, 64)
	cpuTime := utime + stime

	// Parse memory usage (RSS in pages)
	rss, _ := strconv.ParseUint(statFields[23], 10, 64)
	memoryKB := rss * 4 // Assuming 4KB pages

	// Get username from /proc/[pid]/status
	statusPath := fmt.Sprintf("/proc/%d/status", pid)
	statusData, err := os.ReadFile(statusPath)
	if err != nil {
		return nil, err
	}

	var username string = "unknown"
	statusLines := strings.Split(string(statusData), "\n")
	for _, line := range statusLines {
		if strings.HasPrefix(line, "Uid:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				uid := fields[1]
				// Proper UID to username mapping
				if uid == "0" {
					username = "root"
				} else {
					// Use proper user lookup
					if u, err := user.LookupId(uid); err == nil {
						username = u.Username
					} else {
						username = fmt.Sprintf("uid_%s", uid)
					}
				}
			}
			break
		}
	}

	// Get command from /proc/[pid]/comm
	commPath := fmt.Sprintf("/proc/%d/comm", pid)
	commData, err := os.ReadFile(commPath)
	command := "unknown"
	if err == nil {
		command = strings.TrimSpace(string(commData))
	}

	return &ProcessInfo{
		PID:      pid,
		Username: username,
		CPUTime:  cpuTime,
		MemoryKB: memoryKB,
		Command:  command,
	}, nil
}

// DetermineSystemStatus determines overall system status based on metrics
func DetermineSystemStatus(systemMetrics *SystemMetrics, userMetrics []UserMetrics,
	cpuThreshold, memoryThreshold float64, persistentUsers []PersistentUser) string {

	// Heavy load conditions
	if systemMetrics.CPUPercent > cpuThreshold || systemMetrics.MemoryPercent > memoryThreshold {
		return "Heavy Load"
	}

	// Check for persistent user issues
	if len(persistentUsers) > 0 {
		return "Heavy Load"
	}

	// Medium load conditions
	if systemMetrics.CPUPercent > 60 || systemMetrics.MemoryPercent > 60 {
		return "Medium Usage"
	}

	// Check for temporary user spikes
	spikeCount := 0
	for _, user := range userMetrics {
		if user.CPUPercent > 80 || user.MemoryPercent > 80 {
			spikeCount++
		}
	}

	if spikeCount >= 1 && spikeCount <= 2 {
		return "Medium Usage"
	}

	return "Light Usage"
}

// isRealUser determines if a username represents a real user vs system user
func (m *Monitor) isRealUser(username string) bool {
	if m.config == nil {
		return true // fallback to allowing all users if no config
	}
	filtering := m.config.GetUserFiltering()

	// Check explicitly included users first
	for _, included := range filtering.IncludedUsers {
		if username == included {
			return true
		}
	}

	// Check explicitly excluded users
	for _, excluded := range filtering.ExcludedUsers {
		if username == excluded {
			return false
		}
	}

	// Skip users with UID format (failed lookups)
	if strings.HasPrefix(username, "uid_") {
		return false
	}

	// Skip users starting with underscore (common system user pattern)
	if strings.HasPrefix(username, "_") && filtering.IgnoreSystemUsers {
		return false
	}

	// Check UID range - configurable minimum UID for real users
	if u, err := user.Lookup(username); err == nil {
		if uid, err := strconv.Atoi(u.Uid); err == nil {
			return uid >= filtering.MinUIDForRealUsers
		}
	}

	// Default to true for unknown users (let them through)
	return true
}

// shouldTrackUser determines if a user's activity is worth tracking
func (m *Monitor) shouldTrackUser(userMetric UserMetrics) bool {
	if m.config == nil {
		return true // fallback to allowing all users if no config
	}
	filtering := m.config.GetUserFiltering()

	// Only track users with significant resource usage
	if userMetric.CPUPercent < filtering.MinCPUPercent && userMetric.MemoryPercent < filtering.MinMemoryPercent {
		return false
	}

	// Only track users with enough processes or high single process usage
	if userMetric.ProcessCount < filtering.MinProcessCount && userMetric.CPUPercent < 20.0 {
		return false
	}

	return true
}
