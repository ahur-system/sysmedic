package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/ahur-system/sysmedic/internal/alerts"
	"github.com/ahur-system/sysmedic/internal/config"
	"github.com/ahur-system/sysmedic/internal/monitor"
	"github.com/ahur-system/sysmedic/internal/storage"
	"github.com/ahur-system/sysmedic/internal/websocket"
)

// CreatePIDFile creates a PID file with the current process ID
func CreatePIDFile(pidFile string) error {
	pid := os.Getpid()
	return os.WriteFile(pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// Daemon represents the SysMedic background monitoring daemon
type Daemon struct {
	config       *config.Config
	monitor      *monitor.Monitor
	storage      *storage.Storage
	alertManager *alerts.AlertManager
	ctx          context.Context
	cancel       context.CancelFunc
	pidFile      string
}

// PersistentTracker tracks users with sustained high usage
type PersistentTracker struct {
	users map[string]*UserTrackingState
}

// UserTrackingState tracks the state of a user's resource usage over time
type UserTrackingState struct {
	Username         string
	CPUStartTime     *time.Time
	MemoryStartTime  *time.Time
	CPUPeakUsage     float64
	MemoryPeakUsage  float64
	CPUTotalUsage    float64
	MemoryTotalUsage float64
	CPUSamples       int
	MemorySamples    int
}

// NewDaemon creates a new daemon instance
func NewDaemon(cfg *config.Config) (*Daemon, error) {
	// Initialize storage
	dataPath, err := config.GetDataPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get data path: %w", err)
	}

	storage, err := storage.NewStorage(dataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	// Initialize monitor
	mon := monitor.NewMonitor(cfg)

	// Initialize alert manager
	alertMgr := alerts.NewAlertManager(cfg, storage)

	ctx, cancel := context.WithCancel(context.Background())

	return &Daemon{
		config:       cfg,
		monitor:      mon,
		storage:      storage,
		alertManager: alertMgr,
		ctx:          ctx,
		cancel:       cancel,
		pidFile:      config.DefaultPIDPath,
	}, nil
}

// Start starts the daemon
func (d *Daemon) Start() error {
	// Check if already running
	if d.IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	// Write PID file
	if err := d.writePIDFile(); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Initialize persistent tracker
	tracker := &PersistentTracker{
		users: make(map[string]*UserTrackingState),
	}

	// Start monitoring loop
	ticker := time.NewTicker(d.config.GetCheckInterval())
	defer ticker.Stop()

	fmt.Printf("SysMedic daemon started (PID: %d)\n", os.Getpid())
	fmt.Printf("Monitoring interval: %v\n", d.config.GetCheckInterval())

	// Start WebSocket server if enabled
	d.startWebSocketIfEnabled()

	// Cleanup old data on startup
	go d.performMaintenance()

	for {
		select {
		case <-d.ctx.Done():
			fmt.Println("Daemon context cancelled")
			return d.cleanup()

		case sig := <-sigChan:
			fmt.Printf("Received signal: %v\n", sig)
			d.cancel()
			return d.cleanup()

		case <-ticker.C:
			if err := d.monitoringCycle(tracker); err != nil {
				fmt.Printf("Monitoring cycle error: %v\n", err)
				// Continue monitoring despite errors
			}
		}
	}
}

// Stop stops the daemon
func (d *Daemon) Stop() error {
	if !d.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	pid, err := d.readPIDFile()
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send termination signal: %w", err)
	}

	// Wait for process to exit
	for i := 0; i < 30; i++ { // Wait up to 30 seconds
		if !d.IsRunning() {
			break
		}
		time.Sleep(1 * time.Second)
	}

	if d.IsRunning() {
		// Force kill if still running
		if err := process.Kill(); err != nil {
			return fmt.Errorf("failed to kill process: %w", err)
		}
	}

	return d.removePIDFile()
}

// IsRunning checks if the daemon is currently running
func (d *Daemon) IsRunning() bool {
	pid, err := d.readPIDFile()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// Check if process is actually running
	err = process.Signal(syscall.Signal(0))
	return err == nil
}

// GetStatus returns the daemon status
func (d *Daemon) GetStatus() (string, error) {
	if !d.IsRunning() {
		return "stopped", nil
	}

	pid, err := d.readPIDFile()
	if err != nil {
		return "unknown", err
	}

	return fmt.Sprintf("running (PID: %d)", pid), nil
}

// monitoringCycle performs one monitoring cycle
func (d *Daemon) monitoringCycle(tracker *PersistentTracker) error {
	// Get system metrics
	systemMetrics, err := d.monitor.GetSystemMetrics()
	if err != nil {
		return fmt.Errorf("failed to get system metrics: %w", err)
	}

	// Get user metrics
	userMetrics, err := d.monitor.GetUserMetrics()
	if err != nil {
		return fmt.Errorf("failed to get user metrics: %w", err)
	}

	// Store metrics in database
	if err := d.storage.StoreSystemMetrics(systemMetrics); err != nil {
		return fmt.Errorf("failed to store system metrics: %w", err)
	}

	if err := d.storage.StoreUserMetrics(userMetrics); err != nil {
		return fmt.Errorf("failed to store user metrics: %w", err)
	}

	// Update persistent user tracking
	persistentUsers := d.updatePersistentTracking(tracker, userMetrics)

	// Determine system status
	systemStatus := monitor.DetermineSystemStatus(
		systemMetrics,
		userMetrics,
		float64(d.config.Monitoring.CPUThreshold),
		float64(d.config.Monitoring.MemoryThreshold),
		persistentUsers,
	)

	// Check for alerts
	alertCtx := &alerts.AlertContext{
		SystemMetrics:   systemMetrics,
		UserMetrics:     userMetrics,
		PersistentUsers: persistentUsers,
		SystemStatus:    systemStatus,
		PrimaryCause:    d.identifyPrimaryCause(userMetrics, persistentUsers),
		Recommendations: alerts.GenerateRecommendations(&alerts.AlertContext{
			SystemMetrics:   systemMetrics,
			UserMetrics:     userMetrics,
			PersistentUsers: persistentUsers,
			SystemStatus:    systemStatus,
		}, d.config),
	}

	// Send alerts if necessary
	if err := d.alertManager.CheckAndSendAlerts(alertCtx); err != nil {
		return fmt.Errorf("failed to check and send alerts: %w", err)
	}

	return nil
}

// updatePersistentTracking updates the tracking of users with sustained high usage
func (d *Daemon) updatePersistentTracking(tracker *PersistentTracker, userMetrics []monitor.UserMetrics) []monitor.PersistentUser {
	var persistentUsers []monitor.PersistentUser
	now := time.Now()

	// Update tracking for current users
	for _, user := range userMetrics {
		state, exists := tracker.users[user.Username]
		if !exists {
			state = &UserTrackingState{
				Username: user.Username,
			}
			tracker.users[user.Username] = state
		}

		// Check CPU threshold
		cpuThreshold := d.config.GetUserThreshold(user.Username, "cpu")
		if user.CPUPercent > float64(cpuThreshold) {
			if state.CPUStartTime == nil {
				startTime := now
				state.CPUStartTime = &startTime
				state.CPUPeakUsage = user.CPUPercent
			} else {
				// Update peak usage
				if user.CPUPercent > state.CPUPeakUsage {
					state.CPUPeakUsage = user.CPUPercent
				}
				// Update running average
				state.CPUTotalUsage += user.CPUPercent
				state.CPUSamples++

				// Check if persistent
				duration := now.Sub(*state.CPUStartTime)
				persistentTime := d.config.GetUserPersistentTime(user.Username)
				if duration >= persistentTime {
					persistentUsers = append(persistentUsers, monitor.PersistentUser{
						Username:     user.Username,
						Metric:       "cpu",
						StartTime:    *state.CPUStartTime,
						Duration:     duration,
						CurrentUsage: user.CPUPercent,
					})
				}
			}
		} else {
			// Reset CPU tracking if below threshold
			if state.CPUStartTime != nil {
				state.CPUStartTime = nil
				state.CPUPeakUsage = 0
				state.CPUTotalUsage = 0
				state.CPUSamples = 0
			}
		}

		// Check Memory threshold
		memThreshold := d.config.GetUserThreshold(user.Username, "memory")
		if user.MemoryPercent > float64(memThreshold) {
			if state.MemoryStartTime == nil {
				startTime := now
				state.MemoryStartTime = &startTime
				state.MemoryPeakUsage = user.MemoryPercent
			} else {
				// Update peak usage
				if user.MemoryPercent > state.MemoryPeakUsage {
					state.MemoryPeakUsage = user.MemoryPercent
				}
				// Update running average
				state.MemoryTotalUsage += user.MemoryPercent
				state.MemorySamples++

				// Check if persistent
				duration := now.Sub(*state.MemoryStartTime)
				persistentTime := d.config.GetUserPersistentTime(user.Username)
				if duration >= persistentTime {
					persistentUsers = append(persistentUsers, monitor.PersistentUser{
						Username:     user.Username,
						Metric:       "memory",
						StartTime:    *state.MemoryStartTime,
						Duration:     duration,
						CurrentUsage: user.MemoryPercent,
					})
				}
			}
		} else {
			// Reset memory tracking if below threshold
			if state.MemoryStartTime != nil {
				state.MemoryStartTime = nil
				state.MemoryPeakUsage = 0
				state.MemoryTotalUsage = 0
				state.MemorySamples = 0
			}
		}
	}

	// Clean up tracking for users no longer active
	activeUsers := make(map[string]bool)
	for _, user := range userMetrics {
		activeUsers[user.Username] = true
	}

	for username := range tracker.users {
		if !activeUsers[username] {
			delete(tracker.users, username)
		}
	}

	return persistentUsers
}

// identifyPrimaryCause identifies the primary cause of system issues
func (d *Daemon) identifyPrimaryCause(userMetrics []monitor.UserMetrics, persistentUsers []monitor.PersistentUser) string {
	// Check for persistent users first
	if len(persistentUsers) > 0 {
		user := persistentUsers[0] // Take the first persistent user
		return fmt.Sprintf("%s (%s: %.1f%% for %v)",
			user.Username,
			user.Metric,
			user.CurrentUsage,
			user.Duration.Round(time.Minute))
	}

	// Check for high resource users
	if len(userMetrics) > 0 {
		topUser := userMetrics[0]
		if topUser.CPUPercent > 80 || topUser.MemoryPercent > 80 {
			return fmt.Sprintf("%s (CPU: %.1f%%, Memory: %.1f%%)",
				topUser.Username,
				topUser.CPUPercent,
				topUser.MemoryPercent)
		}
	}

	return ""
}

// performMaintenance performs periodic maintenance tasks
func (d *Daemon) performMaintenance() {
	// Clean up old data daily
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-d.ctx.Done():
			return
		case <-ticker.C:
			if err := d.storage.CleanupOldData(d.config.Reporting.RetainDays); err != nil {
				fmt.Printf("Maintenance error: %v\n", err)
			}
		}
	}
}

// cleanup performs cleanup operations
func (d *Daemon) cleanup() error {
	fmt.Println("Cleaning up daemon resources...")

	// Stop WebSocket server if running
	d.stopWebSocketIfRunning()

	// Close storage
	if err := d.storage.Close(); err != nil {
		fmt.Printf("Error closing storage: %v\n", err)
	}

	// Remove PID file
	if err := d.removePIDFile(); err != nil {
		fmt.Printf("Error removing PID file: %v\n", err)
	}

	fmt.Println("Daemon stopped")
	return nil
}

// writePIDFile writes the current process PID to the PID file
func (d *Daemon) writePIDFile() error {
	pid := os.Getpid()
	return os.WriteFile(d.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

// readPIDFile reads the PID from the PID file
func (d *Daemon) readPIDFile() (int, error) {
	data, err := os.ReadFile(d.pidFile)
	if err != nil {
		return 0, err
	}

	pidStr := strings.TrimSpace(string(data))
	return strconv.Atoi(pidStr)
}

// removePIDFile removes the PID file
func (d *Daemon) removePIDFile() error {
	if _, err := os.Stat(d.pidFile); os.IsNotExist(err) {
		return nil // File doesn't exist, nothing to remove
	}
	return os.Remove(d.pidFile)
}

// startWebSocketIfEnabled starts the WebSocket server if it's enabled in config
func (d *Daemon) startWebSocketIfEnabled() {
	// Check main config for WebSocket settings
	if d.config.WebSocket.Enabled {
		manager := websocket.GetManager()
		if err := manager.StartServer(d.config.WebSocket.Port); err != nil {
			fmt.Printf("Warning: Could not start WebSocket server: %v\n", err)
		} else {
			fmt.Printf("WebSocket server started on port %d\n", d.config.WebSocket.Port)
		}
	}
}

// stopWebSocketIfRunning stops the WebSocket server if it's running
func (d *Daemon) stopWebSocketIfRunning() {
	manager := websocket.GetManager()
	if manager.IsRunning() {
		if err := manager.StopServer(); err != nil {
			fmt.Printf("Warning: Could not stop WebSocket server: %v\n", err)
		} else {
			fmt.Printf("WebSocket server stopped\n")
		}
	}
}

// RunInForeground runs the daemon in the foreground (for testing/debugging)
func (d *Daemon) RunInForeground() error {
	fmt.Println("Starting SysMedic daemon in foreground mode...")
	return d.Start()
}
