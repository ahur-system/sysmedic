package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetDefaultConfig(t *testing.T) {
	config := GetDefaultConfig()

	if config == nil {
		t.Fatal("GetDefaultConfig should return a valid config")
	}

	// Test monitoring defaults
	if config.Monitoring.CheckInterval != 60 {
		t.Errorf("Expected check interval 60, got %d", config.Monitoring.CheckInterval)
	}

	if config.Monitoring.CPUThreshold != 80 {
		t.Errorf("Expected CPU threshold 80, got %d", config.Monitoring.CPUThreshold)
	}

	if config.Monitoring.MemoryThreshold != 80 {
		t.Errorf("Expected memory threshold 80, got %d", config.Monitoring.MemoryThreshold)
	}

	if config.Monitoring.PersistentTime != 60 {
		t.Errorf("Expected persistent time 60, got %d", config.Monitoring.PersistentTime)
	}

	// Test user defaults
	if config.Users.CPUThreshold != 80 {
		t.Errorf("Expected user CPU threshold 80, got %d", config.Users.CPUThreshold)
	}

	if config.Users.MemoryThreshold != 80 {
		t.Errorf("Expected user memory threshold 80, got %d", config.Users.MemoryThreshold)
	}

	// Test reporting defaults
	if config.Reporting.Period != "hourly" {
		t.Errorf("Expected reporting period 'hourly', got '%s'", config.Reporting.Period)
	}

	if config.Reporting.RetainDays != 30 {
		t.Errorf("Expected retain days 30, got %d", config.Reporting.RetainDays)
	}

	// Test email defaults
	if config.Email.Enabled != false {
		t.Errorf("Expected email disabled by default, got %t", config.Email.Enabled)
	}

	if config.Email.SMTPHost != "smtp.gmail.com" {
		t.Errorf("Expected SMTP host 'smtp.gmail.com', got '%s'", config.Email.SMTPHost)
	}

	if config.Email.SMTPPort != 587 {
		t.Errorf("Expected SMTP port 587, got %d", config.Email.SMTPPort)
	}

	if config.Email.TLS != true {
		t.Errorf("Expected TLS enabled by default, got %t", config.Email.TLS)
	}

	// Test user thresholds map is initialized
	if config.UserThresholds == nil {
		t.Error("UserThresholds map should be initialized")
	}
}

func TestLoadConfigNonExistent(t *testing.T) {
	// Test loading non-existent config file should return default
	config, err := LoadConfig("/non/existent/path/config.yaml")
	if err != nil {
		t.Errorf("LoadConfig should not error on non-existent file, got: %v", err)
	}

	if config == nil {
		t.Fatal("LoadConfig should return default config for non-existent file")
	}

	// Should be same as default config
	defaultConfig := GetDefaultConfig()
	if config.Monitoring.CheckInterval != defaultConfig.Monitoring.CheckInterval {
		t.Error("Non-existent config should return default values")
	}
}

func TestSaveAndLoadConfig(t *testing.T) {
	// Create temporary directory
	tempDir, err := os.MkdirTemp("", "sysmedic_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a custom config
	config := GetDefaultConfig()
	config.Monitoring.CheckInterval = 120
	config.Monitoring.CPUThreshold = 90
	config.Email.Enabled = true
	config.Email.To = "test@example.com"

	// Add user threshold
	config.UserThresholds["testuser"] = UserThreshold{
		CPUThreshold:    95,
		MemoryThreshold: 85,
		PersistentTime:  30,
	}

	// Save config
	err = SaveConfigToPath(config, configPath)
	if err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Load config back
	loadedConfig, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	// Verify values
	if loadedConfig.Monitoring.CheckInterval != 120 {
		t.Errorf("Expected check interval 120, got %d", loadedConfig.Monitoring.CheckInterval)
	}

	if loadedConfig.Monitoring.CPUThreshold != 90 {
		t.Errorf("Expected CPU threshold 90, got %d", loadedConfig.Monitoring.CPUThreshold)
	}

	if !loadedConfig.Email.Enabled {
		t.Error("Expected email enabled")
	}

	if loadedConfig.Email.To != "test@example.com" {
		t.Errorf("Expected email to 'test@example.com', got '%s'", loadedConfig.Email.To)
	}

	// Verify user threshold
	userThreshold, exists := loadedConfig.UserThresholds["testuser"]
	if !exists {
		t.Error("User threshold should exist")
	}

	if userThreshold.CPUThreshold != 95 {
		t.Errorf("Expected user CPU threshold 95, got %d", userThreshold.CPUThreshold)
	}
}

func TestGetUserThreshold(t *testing.T) {
	config := GetDefaultConfig()
	config.Users.CPUThreshold = 80
	config.Users.MemoryThreshold = 75

	// Add specific user threshold
	config.UserThresholds["poweruser"] = UserThreshold{
		CPUThreshold:    95,
		MemoryThreshold: 90,
	}

	tests := []struct {
		username string
		metric   string
		expected int
	}{
		{"normaluser", "cpu", 80},    // Default
		{"normaluser", "memory", 75}, // Default
		{"poweruser", "cpu", 95},     // Custom
		{"poweruser", "memory", 90},  // Custom
		{"unknown", "unknown", 80},   // Fallback
	}

	for _, test := range tests {
		result := config.GetUserThreshold(test.username, test.metric)
		if result != test.expected {
			t.Errorf("GetUserThreshold(%s, %s) = %d, expected %d",
				test.username, test.metric, result, test.expected)
		}
	}
}

func TestGetUserPersistentTime(t *testing.T) {
	config := GetDefaultConfig()
	config.Users.PersistentTime = 60

	// Add specific user persistent time
	config.UserThresholds["fastuser"] = UserThreshold{
		PersistentTime: 30,
	}

	// Test default user
	duration := config.GetUserPersistentTime("normaluser")
	expected := 60 * time.Minute
	if duration != expected {
		t.Errorf("Expected persistent time %v, got %v", expected, duration)
	}

	// Test custom user
	duration = config.GetUserPersistentTime("fastuser")
	expected = 30 * time.Minute
	if duration != expected {
		t.Errorf("Expected persistent time %v, got %v", expected, duration)
	}
}

func TestSetUserThreshold(t *testing.T) {
	config := GetDefaultConfig()

	// Set CPU threshold
	config.SetUserThreshold("testuser", "cpu_threshold", 95)
	threshold := config.GetUserThreshold("testuser", "cpu")
	if threshold != 95 {
		t.Errorf("Expected CPU threshold 95, got %d", threshold)
	}

	// Set memory threshold
	config.SetUserThreshold("testuser", "memory-threshold", 85)
	threshold = config.GetUserThreshold("testuser", "memory")
	if threshold != 85 {
		t.Errorf("Expected memory threshold 85, got %d", threshold)
	}

	// Set persistent time
	config.SetUserThreshold("testuser", "persistent-time", 45)
	duration := config.GetUserPersistentTime("testuser")
	expected := 45 * time.Minute
	if duration != expected {
		t.Errorf("Expected persistent time %v, got %v", expected, duration)
	}
}

func TestSetSystemThreshold(t *testing.T) {
	config := GetDefaultConfig()

	tests := []struct {
		metric  string
		value   int
		checkFn func() int
	}{
		{
			"cpu-threshold", 85,
			func() int { return config.Monitoring.CPUThreshold },
		},
		{
			"memory_threshold", 90,
			func() int { return config.Monitoring.MemoryThreshold },
		},
		{
			"persistent-time", 45,
			func() int { return config.Monitoring.PersistentTime },
		},
		{
			"check_interval", 30,
			func() int { return config.Monitoring.CheckInterval },
		},
	}

	for _, test := range tests {
		err := config.SetSystemThreshold(test.metric, test.value)
		if err != nil {
			t.Errorf("SetSystemThreshold(%s, %d) returned error: %v",
				test.metric, test.value, err)
		}

		result := test.checkFn()
		if result != test.value {
			t.Errorf("SetSystemThreshold(%s, %d) = %d, expected %d",
				test.metric, test.value, result, test.value)
		}
	}

	// Test invalid metric
	err := config.SetSystemThreshold("invalid_metric", 100)
	if err == nil {
		t.Error("SetSystemThreshold should return error for invalid metric")
	}
}

func TestGetCheckInterval(t *testing.T) {
	config := GetDefaultConfig()
	config.Monitoring.CheckInterval = 90

	duration := config.GetCheckInterval()
	expected := 90 * time.Second
	if duration != expected {
		t.Errorf("Expected check interval %v, got %v", expected, duration)
	}
}

func TestGetSystemPersistentTime(t *testing.T) {
	config := GetDefaultConfig()
	config.Monitoring.PersistentTime = 120

	duration := config.GetSystemPersistentTime()
	expected := 120 * time.Minute
	if duration != expected {
		t.Errorf("Expected system persistent time %v, got %v", expected, duration)
	}
}

func TestIsEmailEnabled(t *testing.T) {
	config := GetDefaultConfig()

	// Test disabled by default
	if config.IsEmailEnabled() {
		t.Error("Email should be disabled by default")
	}

	// Test enabled but missing config
	config.Email.Enabled = true
	if config.IsEmailEnabled() {
		t.Error("Email should be disabled when missing SMTP host")
	}

	// Test enabled with SMTP host but no recipient
	config.Email.SMTPHost = "smtp.example.com"
	if config.IsEmailEnabled() {
		t.Error("Email should be disabled when missing recipient")
	}

	// Test fully configured
	config.Email.To = "admin@example.com"
	if !config.IsEmailEnabled() {
		t.Error("Email should be enabled when fully configured")
	}
}

func TestGetDataPath(t *testing.T) {
	// This test might need to run as root or be skipped in CI
	// depending on system permissions
	path, err := GetDataPath()
	if err != nil {
		t.Skipf("Skipping GetDataPath test (permission denied): %v", err)
	}

	if path == "" {
		t.Error("GetDataPath should return a valid path")
	}

	if path != DefaultDataPath {
		t.Errorf("Expected data path %s, got %s", DefaultDataPath, path)
	}
}

func TestUserThresholdDefaults(t *testing.T) {
	config := GetDefaultConfig()

	// Test that a user threshold with partial config uses defaults
	config.UserThresholds["partialuser"] = UserThreshold{
		CPUThreshold: 95,
		// MemoryThreshold and PersistentTime not set
	}

	// CPU should use custom value
	cpuThreshold := config.GetUserThreshold("partialuser", "cpu")
	if cpuThreshold != 95 {
		t.Errorf("Expected CPU threshold 95, got %d", cpuThreshold)
	}

	// Memory should use default
	memThreshold := config.GetUserThreshold("partialuser", "memory")
	if memThreshold != config.Users.MemoryThreshold {
		t.Errorf("Expected memory threshold %d, got %d",
			config.Users.MemoryThreshold, memThreshold)
	}

	// Persistent time should use default
	persistentTime := config.GetUserPersistentTime("partialuser")
	expected := time.Duration(config.Users.PersistentTime) * time.Minute
	if persistentTime != expected {
		t.Errorf("Expected persistent time %v, got %v", expected, persistentTime)
	}
}

func BenchmarkGetUserThreshold(b *testing.B) {
	config := GetDefaultConfig()
	config.UserThresholds["benchuser"] = UserThreshold{
		CPUThreshold: 95,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.GetUserThreshold("benchuser", "cpu")
	}
}

func BenchmarkSetUserThreshold(b *testing.B) {
	config := GetDefaultConfig()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.SetUserThreshold("benchuser", "cpu_threshold", 95)
	}
}
