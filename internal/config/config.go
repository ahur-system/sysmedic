package config

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Monitoring MonitoringConfig `yaml:"monitoring"`
	Users      UsersConfig      `yaml:"users"`
	Reporting  ReportingConfig  `yaml:"reporting"`
	Email      EmailConfig      `yaml:"email"`
	UserThresholds map[string]UserThreshold `yaml:"user_thresholds"`
}

// MonitoringConfig contains system-wide monitoring settings
type MonitoringConfig struct {
	CheckInterval    int `yaml:"check_interval"`    // seconds
	CPUThreshold     int `yaml:"cpu_threshold"`     // system-wide %
	MemoryThreshold  int `yaml:"memory_threshold"`  // system-wide %
	PersistentTime   int `yaml:"persistent_time"`   // minutes for persistent detection
}

// UsersConfig contains default user monitoring settings
type UsersConfig struct {
	CPUThreshold    int `yaml:"cpu_threshold"`    // default per-user %
	MemoryThreshold int `yaml:"memory_threshold"` // default per-user %
	PersistentTime  int `yaml:"persistent_time"`  // minutes before flagging user
}

// ReportingConfig contains reporting and data retention settings
type ReportingConfig struct {
	Period     string `yaml:"period"`      // hourly/daily/weekly
	RetainDays int    `yaml:"retain_days"` // days to keep data
}

// EmailConfig contains email notification settings
type EmailConfig struct {
	Enabled  bool   `yaml:"enabled"`
	SMTPHost string `yaml:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	From     string `yaml:"from"`
	To       string `yaml:"to"`
	TLS      bool   `yaml:"tls"`
}

// UserThreshold contains per-user custom thresholds
type UserThreshold struct {
	CPUThreshold    int `yaml:"cpu_threshold"`
	MemoryThreshold int `yaml:"memory_threshold"`
	PersistentTime  int `yaml:"persistent_time,omitempty"`
}

const (
	DefaultConfigPath = "/etc/sysmedic/config.yaml"
	DefaultDataPath   = "/var/lib/sysmedic"
	DefaultPIDPath    = "/var/run/sysmedic.pid"
)

// GetDefaultConfig returns the default configuration
func GetDefaultConfig() *Config {
	return &Config{
		Monitoring: MonitoringConfig{
			CheckInterval:   60, // 1 minute
			CPUThreshold:    80,
			MemoryThreshold: 80,
			PersistentTime:  60, // 1 hour
		},
		Users: UsersConfig{
			CPUThreshold:    80,
			MemoryThreshold: 80,
			PersistentTime:  60,
		},
		Reporting: ReportingConfig{
			Period:     "hourly",
			RetainDays: 30,
		},
		Email: EmailConfig{
			Enabled:  false,
			SMTPHost: "smtp.gmail.com",
			SMTPPort: 587,
			TLS:      true,
		},
		UserThresholds: make(map[string]UserThreshold),
	}
}

// LoadConfig loads configuration from file or returns default
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// If config file doesn't exist, return default config
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return GetDefaultConfig(), nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	config := GetDefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	if configPath == "" {
		configPath = DefaultConfigPath
	}

	// Ensure directory exists
	dir := filepath.Dir(configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// GetUserThreshold returns the threshold for a specific user or default
func (c *Config) GetUserThreshold(username string, metric string) int {
	if userThreshold, exists := c.UserThresholds[username]; exists {
		switch metric {
		case "cpu":
			if userThreshold.CPUThreshold > 0 {
				return userThreshold.CPUThreshold
			}
		case "memory":
			if userThreshold.MemoryThreshold > 0 {
				return userThreshold.MemoryThreshold
			}
		}
	}

	// Return default threshold
	switch metric {
	case "cpu":
		return c.Users.CPUThreshold
	case "memory":
		return c.Users.MemoryThreshold
	default:
		return 80 // fallback
	}
}

// GetUserPersistentTime returns the persistent time for a specific user or default
func (c *Config) GetUserPersistentTime(username string) time.Duration {
	if userThreshold, exists := c.UserThresholds[username]; exists {
		if userThreshold.PersistentTime > 0 {
			return time.Duration(userThreshold.PersistentTime) * time.Minute
		}
	}
	return time.Duration(c.Users.PersistentTime) * time.Minute
}

// SetUserThreshold sets a threshold for a specific user
func (c *Config) SetUserThreshold(username, metric string, value int) {
	if c.UserThresholds == nil {
		c.UserThresholds = make(map[string]UserThreshold)
	}

	threshold := c.UserThresholds[username]
	switch metric {
	case "cpu_threshold", "cpu-threshold":
		threshold.CPUThreshold = value
	case "memory_threshold", "memory-threshold":
		threshold.MemoryThreshold = value
	case "persistent_time", "persistent-time":
		threshold.PersistentTime = value
	}
	c.UserThresholds[username] = threshold
}

// SetSystemThreshold sets a system-wide threshold
func (c *Config) SetSystemThreshold(metric string, value int) error {
	switch metric {
	case "cpu_threshold", "cpu-threshold":
		c.Monitoring.CPUThreshold = value
	case "memory_threshold", "memory-threshold":
		c.Monitoring.MemoryThreshold = value
	case "persistent_time", "persistent-time":
		c.Monitoring.PersistentTime = value
	case "check_interval", "check-interval":
		c.Monitoring.CheckInterval = value
	default:
		return fmt.Errorf("unknown system threshold: %s", metric)
	}
	return nil
}

// GetCheckInterval returns the monitoring check interval as duration
func (c *Config) GetCheckInterval() time.Duration {
	return time.Duration(c.Monitoring.CheckInterval) * time.Second
}

// GetSystemPersistentTime returns the system persistent time as duration
func (c *Config) GetSystemPersistentTime() time.Duration {
	return time.Duration(c.Monitoring.PersistentTime) * time.Minute
}

// IsEmailEnabled returns whether email notifications are enabled
func (c *Config) IsEmailEnabled() bool {
	return c.Email.Enabled && c.Email.SMTPHost != "" && c.Email.To != ""
}

// GetDataPath returns the data directory path, creating it if needed
func GetDataPath() (string, error) {
	if err := os.MkdirAll(DefaultDataPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create data directory: %w", err)
	}
	return DefaultDataPath, nil
}
