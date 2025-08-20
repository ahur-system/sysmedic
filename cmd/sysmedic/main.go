package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/ahur-system/sysmedic/internal/config"
	"github.com/ahur-system/sysmedic/internal/daemon"
	"github.com/ahur-system/sysmedic/internal/monitor"
	"github.com/ahur-system/sysmedic/internal/storage"
	"github.com/ahur-system/sysmedic/pkg/cli"
	"github.com/spf13/cobra"

	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"strings"
)

var (
	version = "1.0.5"
	commit  = "dev"
	date    = "unknown"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

type WebSocketDaemon struct {
	config         *config.Config
	storage        *storage.Storage
	server         *http.Server
	updateInterval time.Duration
}

func main() {
	// Check if running as daemon
	if len(os.Args) >= 2 {
		switch os.Args[1] {
		case "--doctor-daemon":
			runDoctorDaemon()
			return
		case "--websocket-daemon":
			runWebSocketDaemon()
			return
		}
	}

	rootCmd := &cobra.Command{
		Use:   "sysmedic",
		Short: "Cross-platform Linux server monitoring CLI tool",
		Long: `SysMedic is a comprehensive server monitoring tool that tracks system
and user resource usage spikes with daemon capabilities and intelligent alerting.`,
		Version: fmt.Sprintf("%s (commit: %s, built: %s)", version, commit, date),
		Run: func(cmd *cobra.Command, args []string) {
			// Default command shows dashboard
			cli.ShowDashboard()
		},
	}

	// Daemon commands
	daemonCmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage the SysMedic monitoring daemon",
		Long:  "Start, stop, or check the status of the SysMedic background monitoring daemon",
	}

	daemonStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the monitoring daemon",
		Run: func(cmd *cobra.Command, args []string) {
			cli.StartDaemon()
		},
	}

	daemonStopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the monitoring daemon",
		Run: func(cmd *cobra.Command, args []string) {
			cli.StopDaemon()
		},
	}

	daemonStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check daemon status",
		Run: func(cmd *cobra.Command, args []string) {
			cli.DaemonStatus()
		},
	}

	daemonCmd.AddCommand(daemonStartCmd, daemonStopCmd, daemonStatusCmd)

	// Config commands
	configCmd := &cobra.Command{
		Use:   "config",
		Short: "Manage SysMedic configuration",
		Long:  "View and modify SysMedic configuration settings",
	}

	configShowCmd := &cobra.Command{
		Use:   "show",
		Short: "Display current configuration",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ShowConfig()
		},
	}

	configSetCmd := &cobra.Command{
		Use:   "set [key] [value]",
		Short: "Set configuration value",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			cli.SetConfig(args[0], args[1])
		},
	}

	configSetUserCmd := &cobra.Command{
		Use:   "set-user [username] [key] [value]",
		Short: "Set user-specific configuration",
		Args:  cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			cli.SetUserConfig(args[0], args[1], args[2])
		},
	}

	configCmd.AddCommand(configShowCmd, configSetCmd, configSetUserCmd)

	// Reports commands
	reportsCmd := &cobra.Command{
		Use:   "reports",
		Short: "View system and user activity reports",
		Long:  "Generate reports on system alerts, user activity, and resource usage patterns",
		Run: func(cmd *cobra.Command, args []string) {
			period, _ := cmd.Flags().GetString("period")
			cli.ShowReports(period)
		},
	}

	reportsCmd.Flags().StringP("period", "p", "hourly", "Report period (hourly, daily, weekly)")

	reportsUsersCmd := &cobra.Command{
		Use:   "users",
		Short: "Show detailed user activity reports",
		Run: func(cmd *cobra.Command, args []string) {
			top, _ := cmd.Flags().GetInt("top")
			user, _ := cmd.Flags().GetString("user")
			period, _ := cmd.Flags().GetString("period")
			cli.ShowUserReports(top, user, period)
		},
	}

	reportsUsersCmd.Flags().IntP("top", "t", 0, "Show top N users")
	reportsUsersCmd.Flags().StringP("user", "u", "", "Show specific user")
	reportsUsersCmd.Flags().StringP("period", "p", "hourly", "Report period")

	reportsCmd.AddCommand(reportsUsersCmd)

	// Alerts commands
	alertsCmd := &cobra.Command{
		Use:   "alerts",
		Short: "Manage system alerts",
		Long:  "View, resolve, and manage system and user alerts",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ShowAlerts()
		},
	}

	alertsListCmd := &cobra.Command{
		Use:   "list",
		Short: "List all alerts",
		Run: func(cmd *cobra.Command, args []string) {
			unresolved, _ := cmd.Flags().GetBool("unresolved")
			period, _ := cmd.Flags().GetString("period")
			cli.ListAlerts(unresolved, period)
		},
	}

	alertsListCmd.Flags().BoolP("unresolved", "u", false, "Show only unresolved alerts")
	alertsListCmd.Flags().StringP("period", "p", "24h", "Time period (24h, 7d, 30d)")

	alertsResolveCmd := &cobra.Command{
		Use:   "resolve [alert-id]",
		Short: "Resolve a specific alert by ID",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.ResolveAlert(args[0])
		},
	}

	alertsResolveAllCmd := &cobra.Command{
		Use:   "resolve-all",
		Short: "Resolve all unresolved alerts",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ResolveAllAlerts()
		},
	}

	alertsCmd.AddCommand(alertsListCmd, alertsResolveCmd, alertsResolveAllCmd)

	// WebSocket commands
	websocketCmd := &cobra.Command{
		Use:   "websocket",
		Short: "Manage WebSocket remote connection server",
		Long:  "Start, stop, or check the status of the WebSocket server for remote monitoring",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ShowWebSocketStatus()
		},
	}

	websocketStartCmd := &cobra.Command{
		Use:   "start [port]",
		Short: "Start the WebSocket server",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			port := 8060 // Default port
			if len(args) > 0 {
				if p, err := strconv.Atoi(args[0]); err == nil {
					port = p
				} else {
					fmt.Printf("Invalid port number: %s\n", args[0])
					return
				}
			}
			cli.StartWebSocketServer(port)
		},
	}

	websocketStopCmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the WebSocket server",
		Run: func(cmd *cobra.Command, args []string) {
			cli.StopWebSocketServer()
		},
	}

	websocketStatusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check WebSocket server status",
		Run: func(cmd *cobra.Command, args []string) {
			cli.ShowWebSocketStatus()
		},
	}

	websocketSecretCmd := &cobra.Command{
		Use:   "new-secret",
		Short: "Generate a new connection secret",
		Run: func(cmd *cobra.Command, args []string) {
			cli.GenerateNewWebSocketSecret()
		},
	}

	websocketCmd.AddCommand(websocketStartCmd, websocketStopCmd, websocketStatusCmd, websocketSecretCmd)

	// Add all commands to root
	rootCmd.AddCommand(daemonCmd, configCmd, reportsCmd, alertsCmd, websocketCmd)

	// Execute the CLI
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runDoctorDaemon runs the monitoring daemon
func runDoctorDaemon() {
	// Set up logging
	log.SetPrefix("[sysmedic.doctor] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create PID file
	pidFile := filepath.Join(cfg.DataPath, "sysmedic.doctor.pid")
	if err := daemon.CreatePIDFile(pidFile); err != nil {
		log.Fatalf("Failed to create PID file: %v", err)
	}
	defer os.Remove(pidFile)

	// Set up signal handling
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize monitoring system
	monitor := monitor.New(cfg, store)

	// Start monitoring in a goroutine
	go func() {
		log.Println("Starting system monitoring...")
		monitor.Start(ctx)
	}()

	// Log startup information
	log.Printf("SysMedic Doctor daemon v%s started (commit: %s, built: %s)", version, commit, date)
	log.Printf("PID: %d", os.Getpid())
	log.Printf("Data path: %s", cfg.DataPath)
	log.Printf("Monitoring interval: %v", time.Duration(cfg.Monitoring.CheckInterval)*time.Second)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal, stopping...")

	// Cancel context to stop monitoring
	cancel()

	// Give some time for graceful shutdown
	time.Sleep(2 * time.Second)

	log.Println("SysMedic Doctor daemon stopped")
}

// runWebSocketDaemon runs the WebSocket server daemon
func runWebSocketDaemon() {
	// Set up logging
	log.SetPrefix("[sysmedic.websocket] ")
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Ensure WebSocket secret is set
	if cfg.WebSocket.Secret == "" {
		secret, err := config.GenerateSecret()
		if err != nil {
			log.Fatalf("Failed to generate WebSocket secret: %v", err)
		}
		cfg.WebSocket.Secret = secret
		if err := config.SaveConfig(cfg); err != nil {
			log.Printf("Warning: Could not save generated secret: %v", err)
		}
	}

	// Initialize storage
	store, err := storage.NewStorage(cfg.DataPath)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Create PID file
	pidFile := filepath.Join(cfg.DataPath, "sysmedic.websocket.pid")
	if err := daemon.CreatePIDFile(pidFile); err != nil {
		log.Fatalf("Failed to create PID file: %v", err)
	}
	defer os.Remove(pidFile)

	// Create WebSocket daemon
	updateInterval := 3 * time.Second // Default 3 seconds
	if cfg.WebSocket.UpdateInterval > 0 {
		updateInterval = time.Duration(cfg.WebSocket.UpdateInterval) * time.Second
	}

	wsDaemon := &WebSocketDaemon{
		config:         cfg,
		storage:        store,
		updateInterval: updateInterval,
	}

	// Set up HTTP server
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", wsDaemon.handleWebSocket)
	mux.HandleFunc("/status", wsDaemon.handleStatus)
	mux.HandleFunc("/health", wsDaemon.handleHealth)

	wsDaemon.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.WebSocket.Port),
		Handler: mux,
	}

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	go func() {
		log.Printf("Starting WebSocket server on port %d", cfg.WebSocket.Port)
		if err := wsDaemon.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Log startup information
	log.Printf("SysMedic WebSocket daemon v%s started (commit: %s, built: %s)", version, commit, date)
	log.Printf("PID: %d", os.Getpid())
	log.Printf("WebSocket server listening on port %d", cfg.WebSocket.Port)
	log.Printf("Connection secret: %s", cfg.WebSocket.Secret)

	// Wait for shutdown signal
	<-sigChan
	log.Println("Received shutdown signal, stopping...")

	// Shutdown server gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := wsDaemon.server.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	log.Println("SysMedic WebSocket daemon stopped")
}

func (ws *WebSocketDaemon) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check for secret authentication via URL parameter
	secret := r.URL.Query().Get("secret")
	if secret != ws.config.WebSocket.Secret {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		log.Printf("Unauthorized WebSocket connection attempt from %s", r.RemoteAddr)
		return
	}

	// Upgrade connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("WebSocket client connected from %s", r.RemoteAddr)

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Send welcome message
	if err := ws.sendWelcomeMessage(conn); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Set up periodic data sending with configurable interval
	dataTicker := time.NewTicker(ws.updateInterval)
	defer dataTicker.Stop()

	// Set up ping ticker
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Create channels for coordination
	done := make(chan struct{})
	intervalUpdate := make(chan time.Duration, 1)

	// Start message reader goroutine
	go ws.readPump(conn, done, intervalUpdate)

	// Main loop for sending data and pings
	for {
		select {
		case <-dataTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := ws.sendSystemData(conn); err != nil {
				log.Printf("Failed to send system data: %v", err)
				return
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
		case newInterval := <-intervalUpdate:
			// Update the data ticker with new interval
			dataTicker.Stop()
			dataTicker = time.NewTicker(newInterval)
			ws.updateInterval = newInterval
			log.Printf("WebSocket update interval changed to %v", newInterval)
		case <-done:
			return
		}
	}
}

func (ws *WebSocketDaemon) readPump(conn *websocket.Conn, done chan struct{}, intervalUpdate chan time.Duration) {
	defer close(done)

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket connection closed: %v", err)
			}
			return
		}

		// Handle incoming messages for configuration
		messageStr := string(message)
		if strings.HasPrefix(messageStr, "websocket.options.sysupdate:") {
			// Parse interval setting: "websocket.options.sysupdate: 5"
			parts := strings.Split(messageStr, ":")
			if len(parts) == 2 {
				intervalStr := strings.TrimSpace(parts[1])
				if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
					newInterval := time.Duration(interval) * time.Second

					// Save to configuration persistently
					ws.config.WebSocket.UpdateInterval = interval
					if err := config.SaveConfig(ws.config); err != nil {
						log.Printf("Failed to save WebSocket interval to config: %v", err)
					} else {
						log.Printf("WebSocket interval saved to config: %d seconds", interval)
					}

					select {
					case intervalUpdate <- newInterval:
						log.Printf("Received interval update request: %d seconds", interval)
					default:
						// Channel is full, skip this update
					}
				}
			}
		}
	}
}

func (ws *WebSocketDaemon) sendWelcomeMessage(conn *websocket.Conn) error {
	// Get system information like CLI does
	mon := monitor.NewMonitor(ws.config)
	systemMetrics, err := mon.GetSystemMetrics()
	if err != nil {
		log.Printf("Failed to get system metrics for welcome: %v", err)
		systemMetrics = &monitor.SystemMetrics{
			Timestamp:     time.Now(),
			CPUPercent:    0,
			MemoryPercent: 0,
		}
	}

	// Get OS info
	osInfo := ws.getOSInfo()

	// Get system status
	systemStatus := ws.getSystemStatus(systemMetrics)

	// Get daemon status
	daemonStatus := "Running" // We know it's running since we're in the daemon

	// Prepare welcome message
	welcome := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"message": "Connected to SysMedic",
			"version": version,
			"system":  osInfo,
			"status":  systemStatus,
			"daemon":  daemonStatus,
		},
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
	}

	// Send welcome message with write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(welcome)
}

func (ws *WebSocketDaemon) sendSystemData(conn *websocket.Conn) error {
	// Get current system metrics like CLI does
	mon := monitor.NewMonitor(ws.config)
	systemMetrics, err := mon.GetSystemMetrics()
	if err != nil {
		log.Printf("Failed to get system metrics: %v", err)
		// Send error state in the expected format
		response := ws.createErrorSystemUpdate()
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteJSON(response)
	}

	// Get user metrics
	userMetrics, err := mon.GetUserMetrics()
	if err != nil {
		log.Printf("Failed to get user metrics: %v", err)
		userMetrics = []monitor.UserMetrics{} // Continue with empty list
	}

	// Get recent alerts
	alerts, err := ws.storage.GetRecentAlerts(24*time.Hour, nil)
	if err != nil {
		log.Printf("Failed to get recent alerts: %v", err)
		alerts = nil
	}

	// Get disk usage and uptime for compatibility
	diskUsage := ws.getDiskUsage()
	uptime := ws.getSystemUptime()

	// Convert user metrics to the expected format
	var users []map[string]interface{}
	for _, user := range userMetrics {
		users = append(users, map[string]interface{}{
			"Username":      user.Username,
			"CPUPercent":    user.CPUPercent,
			"MemoryPercent": user.MemoryPercent,
			"ProcessCount":  user.ProcessCount,
			"Timestamp":     user.Timestamp.Format("2006-01-02T15:04:05.000000000Z"),
			"PIDs":          user.PIDs,
		})
	}

	// Get CLI-style summary
	summary := ws.generateSystemSummary(systemMetrics, userMetrics, alerts)

	// Prepare complete system update message matching the working format
	response := map[string]interface{}{
		"type": "system_update",
		"data": map[string]interface{}{
			"cpu_usage":    systemMetrics.CPUPercent,
			"memory_usage": systemMetrics.MemoryPercent,
			"disk_usage":   diskUsage,
			"uptime":       uptime,
			"summary":      summary,
		},
		"alerts": alerts,
		"system": map[string]interface{}{
			"system": map[string]interface{}{
				"cpu":     systemMetrics.CPUPercent,
				"load1":   systemMetrics.LoadAvg1,
				"load15":  systemMetrics.LoadAvg15,
				"load5":   systemMetrics.LoadAvg5,
				"memory":  systemMetrics.MemoryPercent,
				"network": systemMetrics.NetworkMBps,
			},
			"timestamp": time.Now().Unix(),
			"users":     users,
		},
		"timestamp": time.Now().Unix(),
	}

	// Send data as JSON with write deadline
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(response)
}

func (ws *WebSocketDaemon) createErrorSystemUpdate() map[string]interface{} {
	return map[string]interface{}{
		"type": "system_update",
		"data": map[string]interface{}{
			"cpu_usage":    0.0,
			"memory_usage": 0.0,
			"disk_usage":   0.0,
			"uptime":       "unknown",
			"summary":      "Error retrieving system data",
		},
		"alerts": nil,
		"system": map[string]interface{}{
			"system": map[string]interface{}{
				"cpu":     0.0,
				"load1":   0.0,
				"load15":  0.0,
				"load5":   0.0,
				"memory":  0.0,
				"network": 0.0,
			},
			"timestamp": time.Now().Unix(),
			"users":     []interface{}{},
		},
		"timestamp": time.Now().Unix(),
	}
}

func (ws *WebSocketDaemon) generateSystemSummary(systemMetrics *monitor.SystemMetrics, userMetrics []monitor.UserMetrics, alerts []storage.Alert) string {
	var summary strings.Builder

	// Get OS info and system status
	osInfo := ws.getOSInfo()
	systemStatus := ws.getSystemStatus(systemMetrics)

	summary.WriteString(fmt.Sprintf("System: %s | Status: %s\n", osInfo, systemStatus))
	summary.WriteString(fmt.Sprintf("CPU: %.1f%% | Memory: %.1f%% | Load: %.2f, %.2f, %.2f\n",
		systemMetrics.CPUPercent, systemMetrics.MemoryPercent,
		systemMetrics.LoadAvg1, systemMetrics.LoadAvg5, systemMetrics.LoadAvg15))

	// Active users summary
	if len(userMetrics) > 0 {
		summary.WriteString(fmt.Sprintf("Active Users: %d | ", len(userMetrics)))
		if len(userMetrics) > 0 {
			topUser := userMetrics[0]
			summary.WriteString(fmt.Sprintf("Top: %s (%.1f%% CPU, %.1f%% Memory)",
				topUser.Username, topUser.CPUPercent, topUser.MemoryPercent))
		}
	} else {
		summary.WriteString("No active users detected")
	}

	// Alerts summary
	if alerts != nil && len(alerts) > 0 {
		unresolved := 0
		for _, alert := range alerts {
			if !alert.Resolved {
				unresolved++
			}
		}
		if unresolved > 0 {
			summary.WriteString(fmt.Sprintf(" | Alerts: %d unresolved", unresolved))
		}
	}

	return summary.String()
}

func (ws *WebSocketDaemon) getOSInfo() string {
	// Try to read from /etc/os-release
	data, err := os.ReadFile("/etc/os-release")
	if err != nil {
		return "Unknown Linux"
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "PRETTY_NAME=") {
			// Remove PRETTY_NAME= and quotes
			name := strings.TrimPrefix(line, "PRETTY_NAME=")
			name = strings.Trim(name, "\"")
			return name
		}
	}
	return "Unknown Linux"
}

func (ws *WebSocketDaemon) getSystemStatus(systemMetrics *monitor.SystemMetrics) string {
	// Determine system status like CLI does
	cpuThreshold := float64(ws.config.Monitoring.CPUThreshold)
	memoryThreshold := float64(ws.config.Monitoring.MemoryThreshold)

	if systemMetrics.CPUPercent > cpuThreshold || systemMetrics.MemoryPercent > memoryThreshold {
		return "High Usage"
	}

	// Check for moderate usage
	if systemMetrics.CPUPercent > cpuThreshold*0.7 || systemMetrics.MemoryPercent > memoryThreshold*0.7 {
		return "Moderate Usage"
	}

	return "Light Usage"
}

func (ws *WebSocketDaemon) getDiskUsage() float64 {
	var stat syscall.Statfs_t
	err := syscall.Statfs("/", &stat)
	if err != nil {
		return 0.0
	}

	// Calculate usage percentage
	total := stat.Blocks * uint64(stat.Bsize)
	free := stat.Bavail * uint64(stat.Bsize)
	used := total - free

	if total == 0 {
		return 0.0
	}

	usage := float64(used) / float64(total) * 100.0
	return usage
}

func (ws *WebSocketDaemon) getSystemUptime() string {
	// Read uptime from /proc/uptime
	data, err := os.ReadFile("/proc/uptime")
	if err != nil {
		return "unknown"
	}

	// Parse uptime (first number is total uptime in seconds)
	uptimeStr := strings.Fields(string(data))[0]
	uptimeSeconds, err := strconv.ParseFloat(uptimeStr, 64)
	if err != nil {
		return "unknown"
	}

	// Convert to human readable format
	days := int(uptimeSeconds) / 86400
	hours := (int(uptimeSeconds) % 86400) / 3600
	minutes := (int(uptimeSeconds) % 3600) / 60

	if days > 0 {
		return fmt.Sprintf("%dd %dh %dm", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, minutes)
	} else {
		return fmt.Sprintf("%dm", minutes)
	}
}

func (ws *WebSocketDaemon) handleStatus(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "running",
		"version":   version,
		"port":      ws.config.WebSocket.Port,
		"timestamp": time.Now().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(status); err != nil {
		log.Printf("Failed to encode status response: %v", err)
	}
}

func (ws *WebSocketDaemon) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}
