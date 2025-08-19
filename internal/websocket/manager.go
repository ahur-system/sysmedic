package websocket

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	globalManager *Manager
	managerOnce   sync.Once
)

type Manager struct {
	server     *Server
	configPath string
	mu         sync.RWMutex
}

type Config struct {
	Port    int    `json:"port"`
	Enabled bool   `json:"enabled"`
	Secret  string `json:"secret,omitempty"`
}

func GetManager() *Manager {
	managerOnce.Do(func() {
		homeDir, _ := os.UserHomeDir()
		configPath := filepath.Join(homeDir, ".sysmedic", "websocket.json")
		globalManager = &Manager{
			configPath: configPath,
		}
	})
	return globalManager
}

func (m *Manager) LoadConfig() (*Config, error) {
	config := &Config{
		Port:    8060, // Default port
		Enabled: false,
		Secret:  "", // Will be generated if empty
	}

	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		// Generate initial secret and save config
		config.Secret = m.generateSecret()
		return config, m.SaveConfig(config)
	}

	data, err := ioutil.ReadFile(m.configPath)
	if err != nil {
		return config, err
	}

	if err := json.Unmarshal(data, config); err != nil {
		return config, err
	}

	return config, nil
}

func (m *Manager) SaveConfig(config *Config) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.configPath), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) StartServer(port int) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server != nil && m.server.IsRunning() {
		return fmt.Errorf("WebSocket server is already running on port %d", port)
	}

	// Load existing config to preserve secret
	config, err := m.LoadConfig()
	if err != nil {
		config = &Config{
			Port:    port,
			Enabled: true,
			Secret:  m.generateSecret(),
		}
	} else {
		// Update port and enabled status, keep existing secret
		config.Port = port
		config.Enabled = true
		// Generate secret if it doesn't exist
		if config.Secret == "" {
			config.Secret = m.generateSecret()
		}
	}

	m.server = NewServerWithSecret(port, config.Secret)
	if err := m.server.Start(); err != nil {
		m.server = nil
		return err
	}

	// Save updated configuration
	if err := m.SaveConfig(config); err != nil {
		// Log error but don't fail the start
		fmt.Printf("Warning: Could not save WebSocket configuration: %v\n", err)
	}

	return nil
}

func (m *Manager) StopServer() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.server == nil || !m.server.IsRunning() {
		return fmt.Errorf("WebSocket server is not running")
	}

	if err := m.server.Stop(); err != nil {
		return err
	}

	m.server = nil

	// Update configuration
	config, _ := m.LoadConfig()
	config.Enabled = false
	if err := m.SaveConfig(config); err != nil {
		fmt.Printf("Warning: Could not save WebSocket configuration: %v\n", err)
	}

	return nil
}

func (m *Manager) GetStatus() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	status := map[string]interface{}{
		"running":      false,
		"port":         0,
		"clients":      0,
		"connection_url": "",
	}

	config, _ := m.LoadConfig()
	if m.server != nil && m.server.IsRunning() {
		status["running"] = true
		status["port"] = m.server.port
		status["clients"] = m.server.GetClientCount()
		status["connection_url"] = m.server.GetConnectionURL()
	} else if config.Enabled {
		status["port"] = config.Port
	}

	return status
}

func (m *Manager) GetConnectionURL() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.server == nil || !m.server.IsRunning() {
		return ""
	}

	return m.server.GetConnectionURL()
}

func (m *Manager) BroadcastAlert(alertData interface{}) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.server != nil && m.server.IsRunning() {
		m.server.BroadcastAlert(alertData)
	}
}

func (m *Manager) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.server != nil && m.server.IsRunning()
}

func (m *Manager) GetClientCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.server == nil {
		return 0
	}

	return m.server.GetClientCount()
}

func (m *Manager) generateSecret() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		// Fallback to timestamp-based secret
		return fmt.Sprintf("sysmedic_%d", time.Now().Unix())
	}
	return hex.EncodeToString(bytes)
}

func (m *Manager) RegenerateSecret() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Load current config
	config, err := m.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	// Generate new secret
	config.Secret = m.generateSecret()

	// If server is running, restart it with new secret
	if m.server != nil && m.server.IsRunning() {
		port := m.server.port
		if err := m.server.Stop(); err != nil {
			return fmt.Errorf("failed to stop server: %v", err)
		}

		m.server = NewServerWithSecret(port, config.Secret)
		if err := m.server.Start(); err != nil {
			m.server = nil
			return fmt.Errorf("failed to restart server: %v", err)
		}
	}

	// Save config with new secret
	if err := m.SaveConfig(config); err != nil {
		return fmt.Errorf("failed to save config: %v", err)
	}

	return nil
}
