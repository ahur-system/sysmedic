package cli

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ahur-system/sysmedic/internal/config"
	"github.com/ahur-system/sysmedic/internal/daemon"
	"github.com/ahur-system/sysmedic/internal/monitor"
	"github.com/ahur-system/sysmedic/internal/storage"
	"github.com/ahur-system/sysmedic/internal/websocket"
)

// ShowDashboard displays the main system dashboard
func ShowDashboard() {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// Initialize monitor
	mon := monitor.NewMonitor(cfg)

	// Get current system metrics
	systemMetrics, err := mon.GetSystemMetrics()
	if err != nil {
		fmt.Printf("Error getting system metrics: %v\n", err)
		return
	}

	// Get current user metrics
	userMetrics, err := mon.GetUserMetrics()
	if err != nil {
		fmt.Printf("Error getting user metrics: %v\n", err)
		return
	}

	// Initialize storage to get persistent users
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	persistentUsers, err := store.GetActivePersistentUsers()
	if err != nil {
		fmt.Printf("Error getting persistent users: %v\n", err)
		persistentUsers = []storage.PersistentUserRecord{} // Continue with empty list
	}

	// Convert persistent users to monitor format
	var monitorPersistentUsers []monitor.PersistentUser
	for _, pu := range persistentUsers {
		monitorPersistentUsers = append(monitorPersistentUsers, monitor.PersistentUser{
			Username:     pu.Username,
			Metric:       pu.Metric,
			StartTime:    pu.StartTime,
			Duration:     pu.Duration,
			CurrentUsage: pu.PeakUsage,
		})
	}

	// Determine system status
	systemStatus := monitor.DetermineSystemStatus(
		systemMetrics,
		userMetrics,
		float64(cfg.Monitoring.CPUThreshold),
		float64(cfg.Monitoring.MemoryThreshold),
		monitorPersistentUsers,
	)

	// Check daemon status
	d, _ := daemon.NewDaemon(cfg)
	daemonStatus := "Stopped"
	if d.IsRunning() {
		daemonStatus = "Running"
	}

	// Get OS info
	osInfo := getOSInfo()

	// Print dashboard
	fmt.Printf("System: %s\n", osInfo)
	fmt.Printf("Status: %s\n", systemStatus)
	fmt.Printf("Daemon: %s\n\n", daemonStatus)

	// System metrics
	fmt.Printf("System Metrics:\n")
	fmt.Printf("- CPU: %.1f%% (threshold: %d%%)\n", systemMetrics.CPUPercent, cfg.Monitoring.CPUThreshold)
	fmt.Printf("- Memory: %.1f%% (threshold: %d%%)\n", systemMetrics.MemoryPercent, cfg.Monitoring.MemoryThreshold)
	fmt.Printf("- Network: %.1f MB/s\n", systemMetrics.NetworkMBps)
	fmt.Printf("- Load Average: %.2f, %.2f, %.2f\n\n", systemMetrics.LoadAvg1, systemMetrics.LoadAvg5, systemMetrics.LoadAvg15)

	// Top resource users
	if len(userMetrics) > 0 {
		fmt.Printf("Top Resource Users (Last Hour):\n")

		// Show top 5 users
		count := len(userMetrics)
		if count > 5 {
			count = 5
		}

		for i := 0; i < count; i++ {
			user := userMetrics[i]
			persistent := ""

			// Check if user is persistent
			for _, pu := range monitorPersistentUsers {
				if pu.Username == user.Username {
					persistent = fmt.Sprintf(" (⚠️ High %s %v)", strings.Title(pu.Metric), pu.Duration.Round(time.Minute))
					break
				}
			}

			fmt.Printf("- %s: CPU %.1f%%, Memory %.1f%%, Processes: %d%s\n",
				user.Username,
				user.CPUPercent,
				user.MemoryPercent,
				user.ProcessCount,
				persistent)
		}
	} else {
		fmt.Printf("No active users detected\n")
	}

	// Show recent alerts if any
	alerts, err := store.GetRecentAlerts(24*time.Hour, nil)
	if err == nil && len(alerts) > 0 {
		unresolved := 0
		for _, alert := range alerts {
			if !alert.Resolved {
				unresolved++
			}
		}
		if unresolved > 0 {
			fmt.Printf("\n⚠️  %d unresolved alert(s) in the last 24 hours. Run 'sysmedic reports' for details.\n", unresolved)
		}
	}
}

// StartDaemon starts the monitoring daemon
func StartDaemon() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	d, err := daemon.NewDaemon(cfg)
	if err != nil {
		fmt.Printf("Error creating daemon: %v\n", err)
		os.Exit(1)
	}

	if d.IsRunning() {
		fmt.Println("Daemon is already running")
		return
	}

	fmt.Println("Starting SysMedic daemon...")
	if err := d.Start(); err != nil {
		fmt.Printf("Error starting daemon: %v\n", err)
		os.Exit(1)
	}
}

// StopDaemon stops the monitoring daemon
func StopDaemon() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	d, err := daemon.NewDaemon(cfg)
	if err != nil {
		fmt.Printf("Error creating daemon: %v\n", err)
		os.Exit(1)
	}

	if !d.IsRunning() {
		fmt.Println("Daemon is not running")
		return
	}

	fmt.Println("Stopping SysMedic daemon...")
	if err := d.Stop(); err != nil {
		fmt.Printf("Error stopping daemon: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Daemon stopped successfully")
}

// DaemonStatus shows the daemon status
func DaemonStatus() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	d, err := daemon.NewDaemon(cfg)
	if err != nil {
		fmt.Printf("Error creating daemon: %v\n", err)
		os.Exit(1)
	}

	status, err := d.GetStatus()
	if err != nil {
		fmt.Printf("Error getting daemon status: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Daemon status: %s\n", status)
}

// ShowConfig displays the current configuration
func ShowConfig() {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("SysMedic Configuration:\n\n")

	fmt.Printf("Monitoring:\n")
	fmt.Printf("  Check Interval: %d seconds\n", cfg.Monitoring.CheckInterval)
	fmt.Printf("  CPU Threshold: %d%%\n", cfg.Monitoring.CPUThreshold)
	fmt.Printf("  Memory Threshold: %d%%\n", cfg.Monitoring.MemoryThreshold)
	fmt.Printf("  Persistent Time: %d minutes\n\n", cfg.Monitoring.PersistentTime)

	fmt.Printf("Users:\n")
	fmt.Printf("  Default CPU Threshold: %d%%\n", cfg.Users.CPUThreshold)
	fmt.Printf("  Default Memory Threshold: %d%%\n", cfg.Users.MemoryThreshold)
	fmt.Printf("  Default Persistent Time: %d minutes\n\n", cfg.Users.PersistentTime)

	fmt.Printf("Reporting:\n")
	fmt.Printf("  Period: %s\n", cfg.Reporting.Period)
	fmt.Printf("  Retain Days: %d\n\n", cfg.Reporting.RetainDays)

	fmt.Printf("Email:\n")
	fmt.Printf("  Enabled: %t\n", cfg.Email.Enabled)
	if cfg.Email.Enabled {
		fmt.Printf("  SMTP Host: %s\n", cfg.Email.SMTPHost)
		fmt.Printf("  SMTP Port: %d\n", cfg.Email.SMTPPort)
		fmt.Printf("  To: %s\n", cfg.Email.To)
		fmt.Printf("  TLS: %t\n", cfg.Email.TLS)
	}

	if len(cfg.UserThresholds) > 0 {
		fmt.Printf("\nUser-Specific Thresholds:\n")
		for username, threshold := range cfg.UserThresholds {
			fmt.Printf("  %s:\n", username)
			if threshold.CPUThreshold > 0 {
				fmt.Printf("    CPU Threshold: %d%%\n", threshold.CPUThreshold)
			}
			if threshold.MemoryThreshold > 0 {
				fmt.Printf("    Memory Threshold: %d%%\n", threshold.MemoryThreshold)
			}
			if threshold.PersistentTime > 0 {
				fmt.Printf("    Persistent Time: %d minutes\n", threshold.PersistentTime)
			}
		}
	}
}

// SetConfig sets a configuration value
func SetConfig(key, value string) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Error: value must be a number\n")
		os.Exit(1)
	}

	if err := cfg.SetSystemThreshold(key, intValue); err != nil {
		fmt.Printf("Error setting config: %v\n", err)
		os.Exit(1)
	}

	if err := config.SaveConfig(cfg, ""); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Configuration updated: %s = %s\n", key, value)
}

// SetUserConfig sets a user-specific configuration value
func SetUserConfig(username, key, value string) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		fmt.Printf("Error: value must be a number\n")
		os.Exit(1)
	}

	cfg.SetUserThreshold(username, key, intValue)

	if err := config.SaveConfig(cfg, ""); err != nil {
		fmt.Printf("Error saving config: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("User configuration updated: %s %s = %s\n", username, key, value)
}

// ShowReports displays system reports
func ShowReports(period string) {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Determine time period
	var duration time.Duration
	switch period {
	case "hourly":
		duration = time.Hour
	case "daily":
		duration = 24 * time.Hour
	case "weekly":
		duration = 7 * 24 * time.Hour
	default:
		duration = time.Hour
	}

	fmt.Printf("System Reports (Last %s):\n\n", period)

	// Get recent alerts
	alerts, err := store.GetRecentAlerts(duration, nil)
	if err != nil {
		fmt.Printf("Error getting alerts: %v\n", err)
		return
	}

	if len(alerts) > 0 {
		fmt.Printf("Recent Alerts:\n")
		for _, alert := range alerts {
			status := "✓"
			if !alert.Resolved {
				status = "⚠️"
			}
			fmt.Printf("%s [%s] %s - %s (%v)\n",
				status,
				alert.Timestamp.Format("15:04"),
				strings.Title(alert.Severity),
				alert.Message,
				alert.Duration.Round(time.Minute))
		}
		fmt.Println()
	} else {
		fmt.Printf("No alerts in the last %s\n\n", period)
	}

	// Get system metrics summary
	systemMetrics, err := store.GetRecentSystemMetrics(duration)
	if err != nil {
		fmt.Printf("Error getting system metrics: %v\n", err)
		return
	}

	if len(systemMetrics) > 0 {
		// Calculate averages
		var avgCPU, avgMemory, avgNetwork float64
		for _, metric := range systemMetrics {
			avgCPU += metric.CPUPercent
			avgMemory += metric.MemoryPercent
			avgNetwork += metric.NetworkMBps
		}
		count := float64(len(systemMetrics))
		avgCPU /= count
		avgMemory /= count
		avgNetwork /= count

		fmt.Printf("System Performance Summary:\n")
		fmt.Printf("- Average CPU: %.1f%%\n", avgCPU)
		fmt.Printf("- Average Memory: %.1f%%\n", avgMemory)
		fmt.Printf("- Average Network: %.1f MB/s\n", avgNetwork)
		fmt.Printf("- Data Points: %d\n\n", len(systemMetrics))
	}

	// Database statistics
	stats, err := store.GetDatabaseStats()
	if err == nil {
		fmt.Printf("Database Statistics:\n")
		for key, value := range stats {
			fmt.Printf("- %s: %d\n", strings.Title(strings.ReplaceAll(key, "_", " ")), value)
		}
	}
}

// ShowUserReports displays user activity reports
func ShowUserReports(top int, username, period string) {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Determine time period
	var duration time.Duration
	switch period {
	case "hourly":
		duration = time.Hour
	case "daily":
		duration = 24 * time.Hour
	case "weekly":
		duration = 7 * 24 * time.Hour
	default:
		duration = time.Hour
	}

	if username != "" {
		// Show specific user report
		fmt.Printf("User Report: %s (Last %s)\n\n", username, period)

		userMetrics, err := store.GetRecentUserMetrics(duration, username)
		if err != nil {
			fmt.Printf("Error getting user metrics: %v\n", err)
			return
		}

		if len(userMetrics) == 0 {
			fmt.Printf("No activity found for user %s in the last %s\n", username, period)
			return
		}

		// Calculate averages and peaks
		var avgCPU, avgMemory, avgProcesses float64
		var peakCPU, peakMemory float64
		var maxProcesses int

		for _, metric := range userMetrics {
			avgCPU += metric.CPUPercent
			avgMemory += metric.MemoryPercent
			avgProcesses += float64(metric.ProcessCount)

			if metric.CPUPercent > peakCPU {
				peakCPU = metric.CPUPercent
			}
			if metric.MemoryPercent > peakMemory {
				peakMemory = metric.MemoryPercent
			}
			if metric.ProcessCount > maxProcesses {
				maxProcesses = metric.ProcessCount
			}
		}

		count := float64(len(userMetrics))
		avgCPU /= count
		avgMemory /= count
		avgProcesses /= count

		fmt.Printf("Activity Summary:\n")
		fmt.Printf("- Average CPU: %.1f%% (Peak: %.1f%%)\n", avgCPU, peakCPU)
		fmt.Printf("- Average Memory: %.1f%% (Peak: %.1f%%)\n", avgMemory, peakMemory)
		fmt.Printf("- Average Processes: %.1f (Max: %d)\n", avgProcesses, maxProcesses)
		fmt.Printf("- Data Points: %d\n", len(userMetrics))

	} else if top > 0 {
		// Show top users
		fmt.Printf("Top %d Resource Users (Last %s):\n\n", top, period)

		topUsers, err := store.GetTopUsers(duration, top, "")
		if err != nil {
			fmt.Printf("Error getting top users: %v\n", err)
			return
		}

		if len(topUsers) == 0 {
			fmt.Printf("No user activity found in the last %s\n", period)
			return
		}

		fmt.Printf("%-15s %-10s %-12s %-10s\n", "Username", "Avg CPU", "Avg Memory", "Avg Procs")
		fmt.Printf("%-15s %-10s %-12s %-10s\n", "--------", "-------", "----------", "---------")

		for _, user := range topUsers {
			fmt.Printf("%-15s %-10.1f %-12.1f %-10.0f\n",
				user.Username,
				user.CPUPercent,
				user.MemoryPercent,
				float64(user.ProcessCount))
		}

	} else {
		// Show general user activity report
		fmt.Printf("User Activity Report (Last %s):\n\n", period)

		userMetrics, err := store.GetRecentUserMetrics(duration, "")
		if err != nil {
			fmt.Printf("Error getting user metrics: %v\n", err)
			return
		}

		if len(userMetrics) == 0 {
			fmt.Printf("No user activity found in the last %s\n", period)
			return
		}

		// Group by user and calculate stats
		userStats := make(map[string]struct {
			TotalCPU      float64
			TotalMemory   float64
			TotalProcs    int
			Count         int
			LastSeen      time.Time
		})

		for _, metric := range userMetrics {
			stats := userStats[metric.Username]
			stats.TotalCPU += metric.CPUPercent
			stats.TotalMemory += metric.MemoryPercent
			stats.TotalProcs += metric.ProcessCount
			stats.Count++
			if metric.Timestamp.After(stats.LastSeen) {
				stats.LastSeen = metric.Timestamp
			}
			userStats[metric.Username] = stats
		}

		// Convert to slice for sorting
		type UserStat struct {
			Username    string
			AvgCPU      float64
			AvgMemory   float64
			AvgProcs    float64
			LastSeen    time.Time
		}

		var users []UserStat
		for username, stats := range userStats {
			users = append(users, UserStat{
				Username:  username,
				AvgCPU:    stats.TotalCPU / float64(stats.Count),
				AvgMemory: stats.TotalMemory / float64(stats.Count),
				AvgProcs:  float64(stats.TotalProcs) / float64(stats.Count),
				LastSeen:  stats.LastSeen,
			})
		}

		// Sort by average CPU usage
		sort.Slice(users, func(i, j int) bool {
			return users[i].AvgCPU > users[j].AvgCPU
		})

		fmt.Printf("%-15s %-10s %-12s %-10s %-12s\n", "Username", "Avg CPU", "Avg Memory", "Avg Procs", "Last Seen")
		fmt.Printf("%-15s %-10s %-12s %-10s %-12s\n", "--------", "-------", "----------", "---------", "---------")

		for _, user := range users {
			fmt.Printf("%-15s %-10.1f %-12.1f %-10.1f %-12s\n",
				user.Username,
				user.AvgCPU,
				user.AvgMemory,
				user.AvgProcs,
				user.LastSeen.Format("15:04"))
		}
	}
}

// getOSInfo returns basic OS information
func getOSInfo() string {
	// Try to read OS release info
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				name := strings.TrimPrefix(line, "PRETTY_NAME=")
				name = strings.Trim(name, "\"")
				return name
			}
		}
	}

	// Fallback to hostname
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}

	return "Linux"
}

// ShowAlerts displays the alerts overview
func ShowAlerts() {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Get recent alerts
	alerts, err := store.GetRecentAlerts(24*time.Hour, nil)
	if err != nil {
		fmt.Printf("Error getting alerts: %v\n", err)
		return
	}

	if len(alerts) == 0 {
		fmt.Printf("No alerts found in the last 24 hours\n")
		return
	}

	// Count unresolved alerts
	unresolved := 0
	resolved := 0
	for _, alert := range alerts {
		if alert.Resolved {
			resolved++
		} else {
			unresolved++
		}
	}

	fmt.Printf("Alert Summary (Last 24 Hours):\n\n")
	fmt.Printf("Total Alerts: %d\n", len(alerts))
	fmt.Printf("Unresolved: %d\n", unresolved)
	fmt.Printf("Resolved: %d\n\n", resolved)

	if unresolved > 0 {
		fmt.Printf("Recent Unresolved Alerts:\n")
		count := 0
		for _, alert := range alerts {
			if !alert.Resolved && count < 5 {
				fmt.Printf("  [%d] %s - %s (%s, %v)\n",
					alert.ID,
					alert.Timestamp.Format("15:04"),
					strings.Title(alert.Severity),
					alert.Message,
					alert.Duration.Round(time.Minute))
				count++
			}
		}
		if unresolved > 5 {
			fmt.Printf("  ... and %d more\n", unresolved-5)
		}
		fmt.Printf("\nUse 'sysmedic alerts list -u' to see all unresolved alerts\n")
		fmt.Printf("Use 'sysmedic alerts resolve <id>' to resolve specific alerts\n")
		fmt.Printf("Use 'sysmedic alerts resolve-all' to resolve all alerts\n")
	}
}

// ListAlerts displays a list of alerts
func ListAlerts(unresolved bool, period string) {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Determine time period
	var duration time.Duration
	switch period {
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	// Get alerts based on filter
	var resolvedFilter *bool
	if unresolved {
		filter := false
		resolvedFilter = &filter
	}

	alerts, err := store.GetRecentAlerts(duration, resolvedFilter)
	if err != nil {
		fmt.Printf("Error getting alerts: %v\n", err)
		return
	}

	if len(alerts) == 0 {
		if unresolved {
			fmt.Printf("No unresolved alerts found in the last %s\n", period)
		} else {
			fmt.Printf("No alerts found in the last %s\n", period)
		}
		return
	}

	// Display header
	if unresolved {
		fmt.Printf("Unresolved Alerts (Last %s):\n\n", period)
	} else {
		fmt.Printf("All Alerts (Last %s):\n\n", period)
	}

	fmt.Printf("%-5s %-12s %-8s %-10s %-8s %-40s\n", "ID", "Time", "Type", "Severity", "Status", "Message")
	fmt.Printf("%-5s %-12s %-8s %-10s %-8s %-40s\n", "---", "----", "----", "--------", "------", "-------")

	// Display alerts
	for _, alert := range alerts {
		status := "Open"
		if alert.Resolved {
			status = "Resolved"
		}

		message := alert.Message
		if len(message) > 37 {
			message = message[:37] + "..."
		}

		fmt.Printf("%-5d %-12s %-8s %-10s %-8s %-40s\n",
			alert.ID,
			alert.Timestamp.Format("15:04 Jan 02"),
			alert.AlertType,
			strings.Title(alert.Severity),
			status,
			message)
	}

	fmt.Printf("\nTotal: %d alerts\n", len(alerts))
}

// ResolveAlert resolves a specific alert by ID
func ResolveAlert(alertIDStr string) {
	// Parse alert ID
	alertID, err := strconv.ParseInt(alertIDStr, 10, 64)
	if err != nil {
		fmt.Printf("Error: Invalid alert ID '%s'. Must be a number.\n", alertIDStr)
		return
	}

	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Check if alert exists and is unresolved
	alert, err := store.GetAlertByID(alertID)
	if err != nil {
		fmt.Printf("Error: Alert with ID %d not found\n", alertID)
		return
	}

	if alert.Resolved {
		fmt.Printf("Alert %d is already resolved\n", alertID)
		return
	}

	// Resolve the alert
	if err := store.ResolveAlert(alertID); err != nil {
		fmt.Printf("Error resolving alert: %v\n", err)
		return
	}

	fmt.Printf("✓ Alert %d resolved successfully\n", alertID)
	fmt.Printf("  %s - %s (%s)\n",
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		strings.Title(alert.Severity),
		alert.Message)
}

// ResolveAllAlerts resolves all unresolved alerts
func ResolveAllAlerts() {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		fmt.Printf("Error getting data path: %v\n", err)
		return
	}

	store, err := storage.NewStorage(dataPath)
	if err != nil {
		fmt.Printf("Error initializing storage: %v\n", err)
		return
	}
	defer store.Close()

	// Get count of unresolved alerts first
	alerts, err := store.GetRecentAlerts(365*24*time.Hour, func() *bool { f := false; return &f }())
	if err != nil {
		fmt.Printf("Error getting alerts: %v\n", err)
		return
	}

	unresolvedCount := len(alerts)
	if unresolvedCount == 0 {
		fmt.Printf("No unresolved alerts found\n")
		return
	}

	// Confirm action
	fmt.Printf("Found %d unresolved alerts. Are you sure you want to resolve all of them? (y/N): ", unresolvedCount)
	var response string
	fmt.Scanln(&response)

	if strings.ToLower(response) != "y" && strings.ToLower(response) != "yes" {
		fmt.Printf("Operation cancelled\n")
		return
	}

	// Resolve all alerts
	rowsAffected, err := store.ResolveAllAlerts()
	if err != nil {
		fmt.Printf("Error resolving alerts: %v\n", err)
		return
	}

	fmt.Printf("✓ Successfully resolved %d alerts\n", rowsAffected)
}

// ShowWebSocketStatus displays the current WebSocket server status
func ShowWebSocketStatus() {
	manager := websocket.GetManager()
	status := manager.GetStatus()

	fmt.Printf("WebSocket Status: ")
	if running, ok := status["running"].(bool); ok && running {
		fmt.Printf("🟢 Running\n")
		if port, ok := status["port"].(int); ok {
			fmt.Printf("Port: %d\n", port)
		}
		if clients, ok := status["clients"].(int); ok {
			fmt.Printf("Clients: %d\n", clients)
		}
		if url, ok := status["connection_url"].(string); ok && url != "" {
			fmt.Printf("Connection URL: %s\n", url)
		}
	} else {
		fmt.Printf("🔴 Not Running\n")
		if enabled, ok := status["enabled"].(bool); ok && enabled {
			if port, ok := status["port"].(int); ok {
				fmt.Printf("Port: %d (configured)\n", port)
			}
			fmt.Printf("Status: Daemon not running or WebSocket disabled\n")
		} else {
			fmt.Printf("Status: WebSocket not configured\n")
		}
	}

	// Note: Daemon status check removed for simplicity
	fmt.Printf("Daemon: Check with 'systemctl status sysmedic'\n")
}

// StartWebSocketServer configures and starts the WebSocket server
func StartWebSocketServer(port int) {
	if port < 1 || port > 65535 {
		fmt.Printf("Error: Invalid port number %d. Must be between 1-65535\n", port)
		return
	}

	manager := websocket.GetManager()
	err := manager.Configure(port, true)
	if err != nil {
		fmt.Printf("Error configuring WebSocket server: %v\n", err)
		return
	}

	fmt.Printf("✓ WebSocket server configured on port %d\n", port)
	fmt.Printf("  To start the server, restart the daemon:\n")
	fmt.Printf("  sudo systemctl restart sysmedic\n")
	fmt.Printf("\n")
	fmt.Printf("  After restart, check status with:\n")
	fmt.Printf("  sysmedic websocket status\n")
}

// StopWebSocketServer stops and disables the WebSocket server
func StopWebSocketServer() {
	manager := websocket.GetManager()
	err := manager.Configure(0, false)
	if err != nil {
		fmt.Printf("Error disabling WebSocket server: %v\n", err)
		return
	}

	fmt.Printf("✓ WebSocket server disabled\n")
	fmt.Printf("  To stop the server, restart the daemon:\n")
	fmt.Printf("  sudo systemctl restart sysmedic\n")
}

// GenerateNewWebSocketSecret generates a new authentication secret
func GenerateNewWebSocketSecret() {
	manager := websocket.GetManager()
	status := manager.GetStatus()

	if enabled, ok := status["enabled"].(bool); !ok || !enabled {
		fmt.Printf("Error: WebSocket server is not configured\n")
		fmt.Printf("Enable it first with: sysmedic websocket start\n")
		return
	}

	err := manager.GenerateNewSecret()
	if err != nil {
		fmt.Printf("Error generating new secret: %v\n", err)
		return
	}

	fmt.Printf("✓ New WebSocket secret generated\n")
	fmt.Printf("  All existing clients will be disconnected\n")
	fmt.Printf("  Get new connection URL with:\n")
	fmt.Printf("  sysmedic websocket status\n")
}
