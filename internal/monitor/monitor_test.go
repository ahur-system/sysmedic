package monitor

import (
	"fmt"
	"testing"
	"time"

	"github.com/ahur-system/sysmedic/internal/config"
)

// testConfig implements Config interface for testing
type testConfig struct{}

func (tc *testConfig) GetUserFiltering() config.UserFilteringConfig {
	return config.UserFilteringConfig{
		MinUIDForRealUsers: 1000,
		IgnoreSystemUsers:  true,
		MinCPUPercent:      5.0,
		MinMemoryPercent:   5.0,
		MinProcessCount:    1,
		ExcludedUsers:      []string{"root", "daemon", "sshd"},
		IncludedUsers:      []string{},
	}
}

func TestNewMonitor(t *testing.T) {
	config := &testConfig{}
	monitor := NewMonitor(config)
	if monitor == nil {
		t.Fatal("NewMonitor should return a valid monitor instance")
	}

	if monitor.lastCPUTimes == nil {
		t.Error("lastCPUTimes map should be initialized")
	}
}

func TestDetermineSystemStatus(t *testing.T) {
	tests := []struct {
		name            string
		systemMetrics   *SystemMetrics
		userMetrics     []UserMetrics
		cpuThreshold    float64
		memoryThreshold float64
		persistentUsers []PersistentUser
		expectedStatus  string
	}{
		{
			name: "Light usage - low system resources",
			systemMetrics: &SystemMetrics{
				CPUPercent:    30.0,
				MemoryPercent: 40.0,
			},
			userMetrics:     []UserMetrics{},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{},
			expectedStatus:  "Light Usage",
		},
		{
			name: "Heavy load - high CPU",
			systemMetrics: &SystemMetrics{
				CPUPercent:    85.0,
				MemoryPercent: 50.0,
			},
			userMetrics:     []UserMetrics{},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{},
			expectedStatus:  "Heavy Load",
		},
		{
			name: "Heavy load - high memory",
			systemMetrics: &SystemMetrics{
				CPUPercent:    50.0,
				MemoryPercent: 85.0,
			},
			userMetrics:     []UserMetrics{},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{},
			expectedStatus:  "Heavy Load",
		},
		{
			name: "Heavy load - persistent user",
			systemMetrics: &SystemMetrics{
				CPUPercent:    50.0,
				MemoryPercent: 50.0,
			},
			userMetrics:     []UserMetrics{},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{
				{
					Username:     "test_user",
					Metric:       "cpu",
					Duration:     65 * time.Minute,
					CurrentUsage: 85.0,
				},
			},
			expectedStatus: "Heavy Load",
		},
		{
			name: "Medium usage - moderate system load",
			systemMetrics: &SystemMetrics{
				CPUPercent:    70.0,
				MemoryPercent: 50.0,
			},
			userMetrics:     []UserMetrics{},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{},
			expectedStatus:  "Medium Usage",
		},
		{
			name: "Medium usage - user spike",
			systemMetrics: &SystemMetrics{
				CPUPercent:    50.0,
				MemoryPercent: 50.0,
			},
			userMetrics: []UserMetrics{
				{
					Username:   "spike_user",
					CPUPercent: 85.0,
				},
			},
			cpuThreshold:    80.0,
			memoryThreshold: 80.0,
			persistentUsers: []PersistentUser{},
			expectedStatus:  "Medium Usage",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := DetermineSystemStatus(
				tt.systemMetrics,
				tt.userMetrics,
				tt.cpuThreshold,
				tt.memoryThreshold,
				tt.persistentUsers,
			)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}

func TestCPUTimes(t *testing.T) {
	times := CPUTimes{
		User:    1000,
		Nice:    100,
		System:  500,
		Idle:    8000,
		IOWait:  200,
		IRQ:     50,
		SoftIRQ: 50,
	}

	times.Total = times.User + times.Nice + times.System + times.Idle +
		times.IOWait + times.IRQ + times.SoftIRQ

	expectedTotal := uint64(9900)
	if times.Total != expectedTotal {
		t.Errorf("Expected total %d, got %d", expectedTotal, times.Total)
	}
}

func TestUserMetrics(t *testing.T) {
	user := UserMetrics{
		Username:      "test_user",
		CPUPercent:    75.5,
		MemoryPercent: 45.2,
		ProcessCount:  5,
		Timestamp:     time.Now(),
		PIDs:          []int{1234, 5678, 9012},
	}

	if user.Username != "test_user" {
		t.Errorf("Expected username 'test_user', got '%s'", user.Username)
	}

	if len(user.PIDs) != 3 {
		t.Errorf("Expected 3 PIDs, got %d", len(user.PIDs))
	}

	if user.ProcessCount != 5 {
		t.Errorf("Expected process count 5, got %d", user.ProcessCount)
	}
}

func TestPersistentUser(t *testing.T) {
	startTime := time.Now()
	duration := 75 * time.Minute

	persistent := PersistentUser{
		Username:     "heavy_user",
		Metric:       "cpu",
		StartTime:    startTime,
		Duration:     duration,
		CurrentUsage: 92.5,
	}

	if persistent.Username != "heavy_user" {
		t.Errorf("Expected username 'heavy_user', got '%s'", persistent.Username)
	}

	if persistent.Metric != "cpu" {
		t.Errorf("Expected metric 'cpu', got '%s'", persistent.Metric)
	}

	if persistent.Duration != duration {
		t.Errorf("Expected duration %v, got %v", duration, persistent.Duration)
	}

	if persistent.CurrentUsage != 92.5 {
		t.Errorf("Expected current usage 92.5, got %f", persistent.CurrentUsage)
	}
}

func TestSystemMetrics(t *testing.T) {
	timestamp := time.Now()
	metrics := SystemMetrics{
		Timestamp:     timestamp,
		CPUPercent:    78.5,
		MemoryPercent: 62.3,
		NetworkMBps:   15.7,
		LoadAvg1:      2.45,
		LoadAvg5:      1.89,
		LoadAvg15:     1.23,
	}

	if metrics.CPUPercent != 78.5 {
		t.Errorf("Expected CPU percent 78.5, got %f", metrics.CPUPercent)
	}

	if metrics.MemoryPercent != 62.3 {
		t.Errorf("Expected memory percent 62.3, got %f", metrics.MemoryPercent)
	}

	if metrics.NetworkMBps != 15.7 {
		t.Errorf("Expected network MBps 15.7, got %f", metrics.NetworkMBps)
	}

	if !metrics.Timestamp.Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, metrics.Timestamp)
	}
}

func TestProcessInfo(t *testing.T) {
	proc := ProcessInfo{
		PID:      1234,
		Username: "testuser",
		CPUTime:  5000,
		MemoryKB: 102400,
		Command:  "test_process",
	}

	if proc.PID != 1234 {
		t.Errorf("Expected PID 1234, got %d", proc.PID)
	}

	if proc.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", proc.Username)
	}

	if proc.MemoryKB != 102400 {
		t.Errorf("Expected memory 102400 KB, got %d", proc.MemoryKB)
	}
}

func TestNetworkStats(t *testing.T) {
	stats := NetworkStats{
		BytesReceived:    1048576,
		BytesTransmitted: 2097152,
	}

	if stats.BytesReceived != 1048576 {
		t.Errorf("Expected bytes received 1048576, got %d", stats.BytesReceived)
	}

	if stats.BytesTransmitted != 2097152 {
		t.Errorf("Expected bytes transmitted 2097152, got %d", stats.BytesTransmitted)
	}
}

// Benchmark tests
func BenchmarkDetermineSystemStatus(b *testing.B) {
	systemMetrics := &SystemMetrics{
		CPUPercent:    75.0,
		MemoryPercent: 60.0,
	}

	userMetrics := []UserMetrics{
		{Username: "user1", CPUPercent: 30.0, MemoryPercent: 20.0},
		{Username: "user2", CPUPercent: 45.0, MemoryPercent: 30.0},
	}

	persistentUsers := []PersistentUser{
		{Username: "user3", Metric: "cpu", Duration: 70 * time.Minute, CurrentUsage: 85.0},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DetermineSystemStatus(systemMetrics, userMetrics, 80.0, 80.0, persistentUsers)
	}
}

func BenchmarkNewMonitor(b *testing.B) {
	config := &testConfig{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		NewMonitor(config)
	}
}

// Helper function for testing
func createTestUserMetrics(count int) []UserMetrics {
	metrics := make([]UserMetrics, count)
	for i := 0; i < count; i++ {
		metrics[i] = UserMetrics{
			Username:      fmt.Sprintf("user%d", i),
			CPUPercent:    float64(i * 10),
			MemoryPercent: float64(i * 5),
			ProcessCount:  i + 1,
			Timestamp:     time.Now(),
			PIDs:          []int{1000 + i, 2000 + i},
		}
	}
	return metrics
}

func TestCreateTestUserMetrics(t *testing.T) {
	metrics := createTestUserMetrics(3)

	if len(metrics) != 3 {
		t.Errorf("Expected 3 user metrics, got %d", len(metrics))
	}

	if metrics[0].Username != "user0" {
		t.Errorf("Expected first user 'user0', got '%s'", metrics[0].Username)
	}

	if metrics[2].CPUPercent != 20.0 {
		t.Errorf("Expected third user CPU 20.0, got %f", metrics[2].CPUPercent)
	}
}

// Test edge cases
func TestDetermineSystemStatusEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		systemMetrics  *SystemMetrics
		expectedStatus string
	}{
		{
			name: "Exactly at threshold",
			systemMetrics: &SystemMetrics{
				CPUPercent:    80.0,
				MemoryPercent: 80.0,
			},
			expectedStatus: "Medium Usage",
		},
		{
			name: "Just over threshold",
			systemMetrics: &SystemMetrics{
				CPUPercent:    80.1,
				MemoryPercent: 50.0,
			},
			expectedStatus: "Heavy Load",
		},
		{
			name: "Zero usage",
			systemMetrics: &SystemMetrics{
				CPUPercent:    0.0,
				MemoryPercent: 0.0,
			},
			expectedStatus: "Light Usage",
		},
		{
			name: "Maximum usage",
			systemMetrics: &SystemMetrics{
				CPUPercent:    100.0,
				MemoryPercent: 100.0,
			},
			expectedStatus: "Heavy Load",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := DetermineSystemStatus(
				tt.systemMetrics,
				[]UserMetrics{},
				80.0,
				80.0,
				[]PersistentUser{},
			)

			if status != tt.expectedStatus {
				t.Errorf("Expected status %s, got %s", tt.expectedStatus, status)
			}
		})
	}
}
