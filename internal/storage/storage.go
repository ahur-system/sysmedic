package storage

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sysmedic/sysmedic/internal/monitor"
)

// Storage handles database operations for SysMedic
type Storage struct {
	db *sql.DB
}

// Alert represents a system alert record
type Alert struct {
	ID          int64
	Timestamp   time.Time
	AlertType   string // "system" or "user"
	Severity    string // "light", "medium", "heavy"
	Message     string
	Duration    time.Duration
	PrimaryCause string
	UserDetails  string
	Resolved    bool
	ResolvedAt  *time.Time
}

// UserActivity represents historical user activity
type UserActivity struct {
	ID           int64
	Username     string
	Timestamp    time.Time
	CPUPercent   float64
	MemoryPercent float64
	ProcessCount int
	PIDs         string // JSON encoded PID list
}

// SystemActivity represents historical system metrics
type SystemActivity struct {
	ID            int64
	Timestamp     time.Time
	CPUPercent    float64
	MemoryPercent float64
	NetworkMBps   float64
	LoadAvg1      float64
	LoadAvg5      float64
	LoadAvg15     float64
}

// PersistentUserRecord represents users with sustained high usage
type PersistentUserRecord struct {
	ID           int64
	Username     string
	Metric       string
	StartTime    time.Time
	EndTime      *time.Time
	Duration     time.Duration
	PeakUsage    float64
	AverageUsage float64
	Resolved     bool
}

// NewStorage creates a new storage instance
func NewStorage(dataPath string) (*Storage, error) {
	dbPath := filepath.Join(dataPath, "sysmedic.db")

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	storage := &Storage{db: db}

	if err := storage.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return storage, nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}

// initSchema creates the database tables if they don't exist
func (s *Storage) initSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS system_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		cpu_percent REAL NOT NULL,
		memory_percent REAL NOT NULL,
		network_mbps REAL NOT NULL,
		load_avg_1 REAL NOT NULL,
		load_avg_5 REAL NOT NULL,
		load_avg_15 REAL NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS user_metrics (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		cpu_percent REAL NOT NULL,
		memory_percent REAL NOT NULL,
		process_count INTEGER NOT NULL,
		pids TEXT, -- JSON encoded PID list
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		timestamp DATETIME NOT NULL,
		alert_type TEXT NOT NULL, -- "system" or "user"
		severity TEXT NOT NULL,   -- "light", "medium", "heavy"
		message TEXT NOT NULL,
		duration_minutes INTEGER DEFAULT 0,
		primary_cause TEXT,
		user_details TEXT, -- JSON encoded user breakdown
		resolved BOOLEAN DEFAULT FALSE,
		resolved_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS persistent_users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT NOT NULL,
		metric TEXT NOT NULL, -- "cpu" or "memory"
		start_time DATETIME NOT NULL,
		end_time DATETIME,
		duration_minutes INTEGER DEFAULT 0,
		peak_usage REAL NOT NULL,
		average_usage REAL DEFAULT 0,
		resolved BOOLEAN DEFAULT FALSE,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	-- Indexes for better query performance
	CREATE INDEX IF NOT EXISTS idx_system_metrics_timestamp ON system_metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_user_metrics_timestamp ON user_metrics(timestamp);
	CREATE INDEX IF NOT EXISTS idx_user_metrics_username ON user_metrics(username);
	CREATE INDEX IF NOT EXISTS idx_alerts_timestamp ON alerts(timestamp);
	CREATE INDEX IF NOT EXISTS idx_alerts_resolved ON alerts(resolved);
	CREATE INDEX IF NOT EXISTS idx_persistent_users_username ON persistent_users(username);
	CREATE INDEX IF NOT EXISTS idx_persistent_users_resolved ON persistent_users(resolved);
	`

	_, err := s.db.Exec(schema)
	return err
}

// StoreSystemMetrics stores system metrics in the database
func (s *Storage) StoreSystemMetrics(metrics *monitor.SystemMetrics) error {
	query := `
		INSERT INTO system_metrics
		(timestamp, cpu_percent, memory_percent, network_mbps, load_avg_1, load_avg_5, load_avg_15)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query,
		metrics.Timestamp,
		metrics.CPUPercent,
		metrics.MemoryPercent,
		metrics.NetworkMBps,
		metrics.LoadAvg1,
		metrics.LoadAvg5,
		metrics.LoadAvg15,
	)

	return err
}

// StoreUserMetrics stores user metrics in the database
func (s *Storage) StoreUserMetrics(userMetrics []monitor.UserMetrics) error {
	query := `
		INSERT INTO user_metrics
		(username, timestamp, cpu_percent, memory_percent, process_count, pids)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, user := range userMetrics {
		// Convert PIDs to JSON string (simple comma-separated for now)
		pidsStr := ""
		for i, pid := range user.PIDs {
			if i > 0 {
				pidsStr += ","
			}
			pidsStr += fmt.Sprintf("%d", pid)
		}

		_, err = stmt.Exec(
			user.Username,
			user.Timestamp,
			user.CPUPercent,
			user.MemoryPercent,
			user.ProcessCount,
			pidsStr,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// StoreAlert stores an alert in the database
func (s *Storage) StoreAlert(alert *Alert) error {
	query := `
		INSERT INTO alerts
		(timestamp, alert_type, severity, message, duration_minutes, primary_cause, user_details, resolved)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		alert.Timestamp,
		alert.AlertType,
		alert.Severity,
		alert.Message,
		int(alert.Duration.Minutes()),
		alert.PrimaryCause,
		alert.UserDetails,
		alert.Resolved,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	alert.ID = id

	return nil
}

// StorePersistentUser stores a persistent user record
func (s *Storage) StorePersistentUser(record *PersistentUserRecord) error {
	query := `
		INSERT INTO persistent_users
		(username, metric, start_time, end_time, duration_minutes, peak_usage, average_usage, resolved)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := s.db.Exec(query,
		record.Username,
		record.Metric,
		record.StartTime,
		record.EndTime,
		int(record.Duration.Minutes()),
		record.PeakUsage,
		record.AverageUsage,
		record.Resolved,
	)

	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	record.ID = id

	return nil
}

// GetRecentSystemMetrics retrieves recent system metrics
func (s *Storage) GetRecentSystemMetrics(duration time.Duration) ([]SystemActivity, error) {
	query := `
		SELECT id, timestamp, cpu_percent, memory_percent, network_mbps,
		       load_avg_1, load_avg_5, load_avg_15
		FROM system_metrics
		WHERE timestamp > ?
		ORDER BY timestamp DESC
	`

	since := time.Now().Add(-duration)
	rows, err := s.db.Query(query, since)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []SystemActivity
	for rows.Next() {
		var metric SystemActivity
		err := rows.Scan(
			&metric.ID,
			&metric.Timestamp,
			&metric.CPUPercent,
			&metric.MemoryPercent,
			&metric.NetworkMBps,
			&metric.LoadAvg1,
			&metric.LoadAvg5,
			&metric.LoadAvg15,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// GetRecentUserMetrics retrieves recent user metrics
func (s *Storage) GetRecentUserMetrics(duration time.Duration, username string) ([]UserActivity, error) {
	var query string
	var args []interface{}

	if username != "" {
		query = `
			SELECT id, username, timestamp, cpu_percent, memory_percent, process_count, pids
			FROM user_metrics
			WHERE timestamp > ? AND username = ?
			ORDER BY timestamp DESC
		`
		args = []interface{}{time.Now().Add(-duration), username}
	} else {
		query = `
			SELECT id, username, timestamp, cpu_percent, memory_percent, process_count, pids
			FROM user_metrics
			WHERE timestamp > ?
			ORDER BY timestamp DESC
		`
		args = []interface{}{time.Now().Add(-duration)}
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []UserActivity
	for rows.Next() {
		var metric UserActivity
		err := rows.Scan(
			&metric.ID,
			&metric.Username,
			&metric.Timestamp,
			&metric.CPUPercent,
			&metric.MemoryPercent,
			&metric.ProcessCount,
			&metric.PIDs,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, metric)
	}

	return metrics, rows.Err()
}

// GetTopUsers retrieves top resource-consuming users
func (s *Storage) GetTopUsers(duration time.Duration, limit int, metric string) ([]UserActivity, error) {
	var orderBy string
	switch metric {
	case "cpu":
		orderBy = "AVG(cpu_percent) DESC"
	case "memory":
		orderBy = "AVG(memory_percent) DESC"
	default:
		orderBy = "AVG(cpu_percent + memory_percent) DESC"
	}

	query := fmt.Sprintf(`
		SELECT username,
		       AVG(cpu_percent) as avg_cpu,
		       AVG(memory_percent) as avg_memory,
		       AVG(process_count) as avg_processes,
		       MAX(timestamp) as last_seen
		FROM user_metrics
		WHERE timestamp > ?
		GROUP BY username
		ORDER BY %s
		LIMIT ?
	`, orderBy)

	since := time.Now().Add(-duration)
	rows, err := s.db.Query(query, since, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []UserActivity
	for rows.Next() {
		var user UserActivity
		err := rows.Scan(
			&user.Username,
			&user.CPUPercent,
			&user.MemoryPercent,
			&user.ProcessCount,
			&user.Timestamp,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

// GetRecentAlerts retrieves recent alerts
func (s *Storage) GetRecentAlerts(duration time.Duration, resolved *bool) ([]Alert, error) {
	query := `
		SELECT id, timestamp, alert_type, severity, message, duration_minutes,
		       primary_cause, user_details, resolved, resolved_at
		FROM alerts
		WHERE timestamp > ?
	`
	args := []interface{}{time.Now().Add(-duration)}

	if resolved != nil {
		query += " AND resolved = ?"
		args = append(args, *resolved)
	}

	query += " ORDER BY timestamp DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var alert Alert
		var durationMinutes int
		err := rows.Scan(
			&alert.ID,
			&alert.Timestamp,
			&alert.AlertType,
			&alert.Severity,
			&alert.Message,
			&durationMinutes,
			&alert.PrimaryCause,
			&alert.UserDetails,
			&alert.Resolved,
			&alert.ResolvedAt,
		)
		if err != nil {
			return nil, err
		}
		alert.Duration = time.Duration(durationMinutes) * time.Minute
		alerts = append(alerts, alert)
	}

	return alerts, rows.Err()
}

// GetActivePersistentUsers retrieves currently active persistent users
func (s *Storage) GetActivePersistentUsers() ([]PersistentUserRecord, error) {
	query := `
		SELECT id, username, metric, start_time, end_time, duration_minutes,
		       peak_usage, average_usage, resolved
		FROM persistent_users
		WHERE resolved = FALSE
		ORDER BY start_time DESC
	`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []PersistentUserRecord
	for rows.Next() {
		var user PersistentUserRecord
		var durationMinutes int
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.Metric,
			&user.StartTime,
			&user.EndTime,
			&durationMinutes,
			&user.PeakUsage,
			&user.AverageUsage,
			&user.Resolved,
		)
		if err != nil {
			return nil, err
		}
		user.Duration = time.Duration(durationMinutes) * time.Minute
		users = append(users, user)
	}

	return users, rows.Err()
}

// ResolvePersistentUser marks a persistent user as resolved
func (s *Storage) ResolvePersistentUser(username, metric string) error {
	query := `
		UPDATE persistent_users
		SET resolved = TRUE, end_time = CURRENT_TIMESTAMP
		WHERE username = ? AND metric = ? AND resolved = FALSE
	`

	_, err := s.db.Exec(query, username, metric)
	return err
}

// ResolveAlert marks an alert as resolved
func (s *Storage) ResolveAlert(alertID int64) error {
	query := `
		UPDATE alerts
		SET resolved = TRUE, resolved_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`

	_, err := s.db.Exec(query, alertID)
	return err
}

// CleanupOldData removes data older than the specified retention period
func (s *Storage) CleanupOldData(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	queries := []string{
		"DELETE FROM system_metrics WHERE timestamp < ?",
		"DELETE FROM user_metrics WHERE timestamp < ?",
		"DELETE FROM alerts WHERE timestamp < ? AND resolved = TRUE",
		"DELETE FROM persistent_users WHERE start_time < ? AND resolved = TRUE",
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query, cutoff); err != nil {
			return fmt.Errorf("cleanup failed for query %s: %w", query, err)
		}
	}

	// Vacuum database to reclaim space
	if _, err := s.db.Exec("VACUUM"); err != nil {
		return fmt.Errorf("vacuum failed: %w", err)
	}

	return nil
}

// GetDatabaseStats returns basic database statistics
func (s *Storage) GetDatabaseStats() (map[string]int, error) {
	stats := make(map[string]int)

	queries := map[string]string{
		"system_metrics":   "SELECT COUNT(*) FROM system_metrics",
		"user_metrics":     "SELECT COUNT(*) FROM user_metrics",
		"alerts":           "SELECT COUNT(*) FROM alerts",
		"persistent_users": "SELECT COUNT(*) FROM persistent_users",
		"unresolved_alerts": "SELECT COUNT(*) FROM alerts WHERE resolved = FALSE",
	}

	for name, query := range queries {
		var count int
		err := s.db.QueryRow(query).Scan(&count)
		if err != nil {
			return nil, fmt.Errorf("failed to get %s count: %w", name, err)
		}
		stats[name] = count
	}

	return stats, nil
}
