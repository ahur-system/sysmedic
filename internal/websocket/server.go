package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ahur-system/sysmedic/internal/config"
	"github.com/ahur-system/sysmedic/internal/monitor"
	"github.com/gorilla/websocket"
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

type Server struct {
	port     int
	secret   string
	hostname string
	mu       sync.RWMutex
	clients  map[*websocket.Conn]bool
	running  bool
	server   *http.Server
	listener net.Listener
}

type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	RequestID string      `json:"request_id,omitempty"`
}

type ClientRequest struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"request_id"`
}

func NewServer(port int) *Server {
	hostname, err := os.Hostname()
	if err != nil {
		// Fallback to IP if hostname fails
		hostname = getLocalIP()
	}

	return &Server{
		port:     port,
		hostname: hostname,
		clients:  make(map[*websocket.Conn]bool),
	}
}

func NewServerWithSecret(port int, secret string) *Server {
	hostname, err := os.Hostname()
	if err != nil {
		// Fallback to IP if hostname fails
		hostname = getLocalIP()
	}

	return &Server{
		port:     port,
		hostname: hostname,
		secret:   secret,
		clients:  make(map[*websocket.Conn]bool),
	}
}

func (s *Server) generateSecret() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based secret
		return fmt.Sprintf("sysmedic_%d", time.Now().Unix())
	}
	return hex.EncodeToString(bytes)
}

func (s *Server) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("WebSocket server is already running")
	}

	// Only generate secret if not already set
	if s.secret == "" {
		s.secret = s.generateSecret()
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", s.port),
		Handler: mux,
	}

	// Test if we can bind to the port first
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to bind to port %d: %v", s.port, err)
	}

	s.listener = listener
	s.running = true

	go func() {
		log.Printf("WebSocket server starting on port %d", s.port)
		if err := s.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Printf("WebSocket server error: %v", err)
			s.mu.Lock()
			s.running = false
			s.mu.Unlock()
		}
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	return nil
}

func (s *Server) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	authSecret := r.URL.Query().Get("secret")
	if authSecret != s.secret {
		log.Printf("Unauthorized WebSocket connection attempt")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		return
	}
	defer conn.Close()

	// Configure connection
	conn.SetReadLimit(maxMessageSize)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	s.mu.Lock()
	s.clients[conn] = true
	clientCount := len(s.clients)
	s.mu.Unlock()
	defer s.removeClient(conn)

	log.Printf("New WebSocket client connected. Total clients: %d", clientCount)

	// Send welcome message with system information
	systemInfo := s.getSystemInfo()
	welcome := map[string]interface{}{
		"type": "welcome",
		"data": map[string]interface{}{
			"message": "Connected to SysMedic",
			"version": "1.0.5",
			"system":  systemInfo["system"],
			"status":  systemInfo["status"],
			"daemon":  systemInfo["daemon"],
		},
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
	}
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	if err := conn.WriteJSON(welcome); err != nil {
		log.Printf("Error sending welcome message: %v", err)
		return
	}

	// Create channels for coordination
	done := make(chan struct{})

	// Start message reader goroutine
	go s.readPump(conn, done)

	// Start sending periodic system updates every 3 seconds
	dataTicker := time.NewTicker(3 * time.Second)
	defer dataTicker.Stop()

	// Set up ping ticker
	pingTicker := time.NewTicker(pingPeriod)
	defer pingTicker.Stop()

	// Main loop for sending data and pings
	for {
		select {
		case <-dataTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := s.sendSystemUpdate(conn); err != nil {
				log.Printf("Error sending system update: %v", err)
				return
			}
		case <-pingTicker.C:
			conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
		case <-done:
			return
		}
	}
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	clientCount := len(s.clients)
	running := s.running
	s.mu.RUnlock()

	status := map[string]interface{}{
		"status":     "healthy",
		"running":    running,
		"clients":    clientCount,
		"port":       s.port,
		"hostname":   s.hostname,
		"has_secret": s.secret != "",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, `{"status":"%s","running":%t,"clients":%d,"port":%d,"hostname":"%s","has_secret":%t}`,
		status["status"], status["running"], status["clients"], status["port"], status["hostname"], status["has_secret"])
}

func (s *Server) readPump(conn *websocket.Conn, done chan struct{}) {
	defer close(done)

	for {
		var request ClientRequest
		err := conn.ReadJSON(&request)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			} else {
				log.Printf("WebSocket connection closed: %v", err)
			}
			return
		}

		s.processClientRequest(conn, request)
	}
}

func (s *Server) processClientRequest(conn *websocket.Conn, request ClientRequest) {
	switch request.Type {
	case "get_system_info":
		s.sendSystemInfo(conn, request.RequestID)
	case "get_alerts":
		s.sendAlerts(conn, request.RequestID)
	case "get_user_metrics":
		s.sendUserMetrics(conn, request.RequestID)
	case "get_config":
		s.sendConfig(conn, request.RequestID)
	case "get_uptime":
		s.sendUptime(conn, request.RequestID)
	case "ping":
		s.sendPong(conn, request.RequestID)
	default:
		s.sendError(conn, request.RequestID, fmt.Sprintf("Unknown request type: %s", request.Type))
	}
}

func (s *Server) sendSystemInfo(conn *websocket.Conn, requestID string) {
	metrics, err := s.getSystemMetrics()
	if err != nil {
		s.sendError(conn, requestID, fmt.Sprintf("Failed to get system metrics: %v", err))
		return
	}

	response := Message{
		Type:      "system_info_response",
		Data:      metrics,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendAlerts(conn *websocket.Conn, requestID string) {
	// This would integrate with the alerts system
	alerts := map[string]interface{}{
		"unresolved_count": 0,
		"total_count":      0,
		"recent_alerts":    []string{},
		"status":           "No active alerts",
	}

	response := Message{
		Type:      "alerts_response",
		Data:      alerts,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendUserMetrics(conn *websocket.Conn, requestID string) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	mon := monitor.NewMonitor(cfg)
	userMetrics, err := mon.GetUserMetrics()
	if err != nil {
		s.sendError(conn, requestID, fmt.Sprintf("Failed to get user metrics: %v", err))
		return
	}

	response := Message{
		Type:      "user_metrics_response",
		Data:      userMetrics,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendConfig(conn *websocket.Conn, requestID string) {
	cfg, err := config.LoadConfig("")
	if err != nil {
		s.sendError(conn, requestID, fmt.Sprintf("Failed to load config: %v", err))
		return
	}

	// Send safe config info (no sensitive data)
	configInfo := map[string]interface{}{
		"monitoring_interval": cfg.GetCheckInterval().String(),
		"cpu_threshold":       cfg.Monitoring.CPUThreshold,
		"memory_threshold":    cfg.Monitoring.MemoryThreshold,
		"version":             "1.0.5",
	}

	response := Message{
		Type:      "config_response",
		Data:      configInfo,
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendUptime(conn *websocket.Conn, requestID string) {
	uptime := getSystemUptime()

	response := Message{
		Type:      "uptime_response",
		Data:      map[string]interface{}{"uptime": uptime},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendPong(conn *websocket.Conn, requestID string) {
	response := Message{
		Type:      "pong",
		Data:      map[string]interface{}{"message": "pong"},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.SetWriteDeadline(time.Now().Add(writeWait))
	conn.WriteJSON(response)
}

func (s *Server) sendError(conn *websocket.Conn, requestID string, errorMsg string) {
	response := Message{
		Type:      "error",
		Data:      map[string]interface{}{"error": errorMsg},
		Timestamp: time.Now(),
		RequestID: requestID,
	}
	conn.WriteJSON(response)
}

func (s *Server) sendSystemUpdate(conn *websocket.Conn) error {
	// Get current system stats from monitor
	metrics, err := s.getSystemMetrics()
	if err != nil {
		log.Printf("Error getting system metrics: %v", err)
		// Fall back to basic data
		update := map[string]interface{}{
			"type": "system_update",
			"data": map[string]interface{}{
				"cpu_usage":    0.0,
				"memory_usage": 0.0,
				"disk_usage":   0.0,
				"uptime":       "unknown",
			},
			"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
		}
		conn.SetWriteDeadline(time.Now().Add(writeWait))
		return conn.WriteJSON(update)
	}

	update := map[string]interface{}{
		"type": "system_update",
		"data": map[string]interface{}{
			"cpu_usage":    metrics["cpu_usage"],
			"memory_usage": metrics["memory_usage"],
			"disk_usage":   metrics["disk_usage"],
			"uptime":       metrics["uptime"],
		},
		"timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
	}

	conn.SetWriteDeadline(time.Now().Add(writeWait))
	return conn.WriteJSON(update)
}

func (s *Server) removeClient(conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.clients[conn] {
		delete(s.clients, conn)
		conn.Close()
		log.Printf("Client disconnected. Total clients: %d", len(s.clients))
	}
}

func (s *Server) BroadcastAlert(alertData interface{}) {
	s.mu.RLock()
	clients := make([]*websocket.Conn, 0, len(s.clients))
	for conn := range s.clients {
		clients = append(clients, conn)
	}
	s.mu.RUnlock()

	message := Message{
		Type:      "alert",
		Data:      alertData,
		Timestamp: time.Now(),
	}

	for _, conn := range clients {
		if err := conn.WriteJSON(message); err != nil {
			log.Printf("Error broadcasting alert: %v", err)
			s.removeClient(conn)
		}
	}
}

func (s *Server) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("WebSocket server is not running")
	}

	// Close all client connections
	for conn := range s.clients {
		conn.Close()
	}
	s.clients = make(map[*websocket.Conn]bool)

	// Close the server
	if s.server != nil {
		s.server.Close()
	}
	if s.listener != nil {
		s.listener.Close()
	}

	s.running = false
	s.secret = ""
	s.server = nil
	s.listener = nil

	log.Println("WebSocket server stopped")
	return nil
}

func (s *Server) GetConnectionURL() string {
	if s.secret == "" {
		return ""
	}
	return fmt.Sprintf("sysmedic://%s@%s:%d/", s.secret, s.hostname, s.port)
}

// getSystemStatusFromMetrics determines system status from metrics
func (s *Server) getSystemStatusFromMetrics(systemMetrics *monitor.SystemMetrics, cfg *config.Config) string {
	cpuThreshold := float64(cfg.Monitoring.CPUThreshold)
	memoryThreshold := float64(cfg.Monitoring.MemoryThreshold)

	if systemMetrics.CPUPercent > cpuThreshold || systemMetrics.MemoryPercent > memoryThreshold {
		return "High Usage"
	}

	// Check for moderate usage
	if systemMetrics.CPUPercent > cpuThreshold*0.7 || systemMetrics.MemoryPercent > memoryThreshold*0.7 {
		return "Moderate Usage"
	}

	return "Light Usage"
}

func (s *Server) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

func (s *Server) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// getSystemMetrics gets real system metrics using the monitor package
func (s *Server) getSystemMetrics() (map[string]interface{}, error) {
	// Load the system config
	cfg, err := config.LoadConfig("")
	if err != nil {
		// Fall back to default config if loading fails
		cfg = config.GetDefaultConfig()
	}

	// Create monitor instance with proper config
	mon := monitor.NewMonitor(cfg)
	systemMetrics, err := mon.GetSystemMetrics()
	if err != nil {
		return nil, err
	}

	// Get disk usage separately
	diskUsage := getDiskUsage()

	// Calculate uptime
	uptime := getSystemUptime()

	return map[string]interface{}{
		"cpu_usage":    systemMetrics.CPUPercent,
		"memory_usage": systemMetrics.MemoryPercent,
		"disk_usage":   diskUsage,
		"uptime":       uptime,
		"load_avg":     fmt.Sprintf("%.2f, %.2f, %.2f", systemMetrics.LoadAvg1, systemMetrics.LoadAvg5, systemMetrics.LoadAvg15),
		"network_io":   fmt.Sprintf("%.2f MB/s", systemMetrics.NetworkMBps),
	}, nil
}

// getDiskUsage gets disk usage for root filesystem using syscalls
func getDiskUsage() float64 {
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

// getSystemUptime gets the system uptime
func getSystemUptime() string {
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

func getLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

// getSystemInfo gets system information for the welcome message
func (s *Server) getSystemInfo() map[string]interface{} {
	systemInfo := map[string]interface{}{
		"system": "Unknown",
		"status": "Unknown",
		"daemon": "Running",
	}

	// Get OS information
	if osInfo := getOSInfo(); osInfo != "" {
		systemInfo["system"] = osInfo
	}

	// Get system status using SysMedic's monitor logic
	// Get system status using actual metrics
	cfg, err := config.LoadConfig("")
	if err == nil {
		mon := monitor.NewMonitor(cfg)
		if systemMetrics, err := mon.GetSystemMetrics(); err == nil {
			systemInfo["status"] = s.getSystemStatusFromMetrics(systemMetrics, cfg)
		} else {
			systemInfo["status"] = "Unknown"
		}
	} else {
		systemInfo["status"] = "Unknown"
	}

	return systemInfo
}

// getOSInfo gets the operating system information
func getOSInfo() string {
	// Try to read from /etc/os-release
	if data, err := os.ReadFile("/etc/os-release"); err == nil {
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				// Extract the value between quotes
				name := strings.TrimPrefix(line, "PRETTY_NAME=")
				name = strings.Trim(name, "\"")
				return name
			}
		}
	}

	// Fallback to lsb_release command
	if cmd := exec.Command("lsb_release", "-d"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			parts := strings.SplitN(string(output), ":", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[1])
			}
		}
	}

	// Final fallback to uname
	if cmd := exec.Command("uname", "-a"); cmd != nil {
		if output, err := cmd.Output(); err == nil {
			return strings.TrimSpace(string(output))
		}
	}

	return "Linux"
}

// getSystemStatus determines system usage status using SysMedic's monitor logic
func (s *Server) getSystemStatus() string {
	// Load configuration
	cfg, err := config.LoadConfig("")
	if err != nil {
		cfg = config.GetDefaultConfig()
	}

	// Create monitor instance
	mon := monitor.NewMonitor(cfg)

	// Get system metrics
	systemMetrics, err := mon.GetSystemMetrics()
	if err != nil {
		return "Unknown"
	}

	// Get user metrics
	userMetrics, err := mon.GetUserMetrics()
	if err != nil {
		userMetrics = []monitor.UserMetrics{} // Empty slice if we can't get user metrics
	}

	// Use SysMedic's DetermineSystemStatus function
	// Note: We pass empty persistent users slice since we don't track them in websocket context
	status := monitor.DetermineSystemStatus(
		systemMetrics,
		userMetrics,
		float64(cfg.Monitoring.CPUThreshold),
		float64(cfg.Monitoring.MemoryThreshold),
		[]monitor.PersistentUser{},
	)

	return status
}
