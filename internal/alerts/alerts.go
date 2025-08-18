package alerts

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"os"
	"strings"
	"time"

	"github.com/sysmedic/sysmedic/internal/config"
	"github.com/sysmedic/sysmedic/internal/monitor"
	"github.com/sysmedic/sysmedic/internal/storage"
)

// AlertManager handles alert generation and notification
type AlertManager struct {
	config  *config.Config
	storage *storage.Storage
}

// AlertContext contains information for generating alerts
type AlertContext struct {
	SystemMetrics    *monitor.SystemMetrics
	UserMetrics      []monitor.UserMetrics
	PersistentUsers  []monitor.PersistentUser
	SystemStatus     string
	Duration         time.Duration
	PrimaryCause     string
	Recommendations  []string
}

// NewAlertManager creates a new alert manager
func NewAlertManager(cfg *config.Config, storage *storage.Storage) *AlertManager {
	return &AlertManager{
		config:  cfg,
		storage: storage,
	}
}

// CheckAndSendAlerts checks system conditions and sends alerts if necessary
func (am *AlertManager) CheckAndSendAlerts(ctx *AlertContext) error {
	// Determine if we need to send an alert
	shouldAlert := am.shouldSendAlert(ctx)
	if !shouldAlert {
		return nil
	}

	// Create alert record
	alert := &storage.Alert{
		Timestamp:    time.Now(),
		Severity:     am.determineSeverity(ctx),
		Message:      am.generateAlertMessage(ctx),
		Duration:     ctx.Duration,
		PrimaryCause: ctx.PrimaryCause,
		UserDetails:  am.generateUserBreakdown(ctx.UserMetrics),
		Resolved:     false,
	}

	// Determine alert type
	if len(ctx.PersistentUsers) > 0 || ctx.PrimaryCause != "" {
		alert.AlertType = "user"
	} else {
		alert.AlertType = "system"
	}

	// Store alert in database
	if err := am.storage.StoreAlert(alert); err != nil {
		return fmt.Errorf("failed to store alert: %w", err)
	}

	// Send email notification if enabled
	if am.config.IsEmailEnabled() {
		if err := am.sendEmailAlert(alert, ctx); err != nil {
			return fmt.Errorf("failed to send email alert: %w", err)
		}
	}

	return nil
}

// shouldSendAlert determines if an alert should be sent
func (am *AlertManager) shouldSendAlert(ctx *AlertContext) bool {
	// Check for system-wide thresholds
	if ctx.SystemMetrics.CPUPercent > float64(am.config.Monitoring.CPUThreshold) {
		return true
	}
	if ctx.SystemMetrics.MemoryPercent > float64(am.config.Monitoring.MemoryThreshold) {
		return true
	}

	// Check for persistent user issues
	if len(ctx.PersistentUsers) > 0 {
		return true
	}

	// Check for user spikes above thresholds
	for _, user := range ctx.UserMetrics {
		cpuThreshold := am.config.GetUserThreshold(user.Username, "cpu")
		memThreshold := am.config.GetUserThreshold(user.Username, "memory")

		if user.CPUPercent > float64(cpuThreshold) || user.MemoryPercent > float64(memThreshold) {
			return true
		}
	}

	return false
}

// determineSeverity determines alert severity based on context
func (am *AlertManager) determineSeverity(ctx *AlertContext) string {
	switch ctx.SystemStatus {
	case "Heavy Load":
		return "heavy"
	case "Medium Usage":
		return "medium"
	default:
		return "light"
	}
}

// generateAlertMessage creates a human-readable alert message
func (am *AlertManager) generateAlertMessage(ctx *AlertContext) string {
	var parts []string

	// System status
	parts = append(parts, fmt.Sprintf("System Status: %s", ctx.SystemStatus))

	// System metrics
	parts = append(parts, fmt.Sprintf("CPU: %.1f%%, Memory: %.1f%%, Network: %.1f MB/s",
		ctx.SystemMetrics.CPUPercent,
		ctx.SystemMetrics.MemoryPercent,
		ctx.SystemMetrics.NetworkMBps))

	// Primary cause if identified
	if ctx.PrimaryCause != "" {
		parts = append(parts, fmt.Sprintf("Primary Cause: %s", ctx.PrimaryCause))
	}

	// Duration if significant
	if ctx.Duration > time.Minute {
		parts = append(parts, fmt.Sprintf("Duration: %v", ctx.Duration.Round(time.Minute)))
	}

	// Top resource users
	if len(ctx.UserMetrics) > 0 {
		parts = append(parts, "Top Users:")
		for i, user := range ctx.UserMetrics {
			if i >= 3 { // Show top 3 users
				break
			}
			persistent := ""
			for _, pu := range ctx.PersistentUsers {
				if pu.Username == user.Username {
					persistent = " (PERSISTENT)"
					break
				}
			}
			parts = append(parts, fmt.Sprintf("  - %s: CPU %.1f%%, Memory %.1f%%, Processes: %d%s",
				user.Username,
				user.CPUPercent,
				user.MemoryPercent,
				user.ProcessCount,
				persistent))
		}
	}

	return strings.Join(parts, "\n")
}

// generateUserBreakdown creates a detailed user breakdown for storage
func (am *AlertManager) generateUserBreakdown(userMetrics []monitor.UserMetrics) string {
	var parts []string

	for _, user := range userMetrics {
		parts = append(parts, fmt.Sprintf("%s:%.1f%%:%.1f%%:%d",
			user.Username,
			user.CPUPercent,
			user.MemoryPercent,
			user.ProcessCount))
	}

	return strings.Join(parts, ";")
}

// sendEmailAlert sends an email notification
func (am *AlertManager) sendEmailAlert(alert *storage.Alert, ctx *AlertContext) error {
	subject := am.generateEmailSubject(alert, ctx)
	body := am.generateEmailBody(alert, ctx)

	return am.sendEmail(subject, body)
}

// generateEmailSubject creates email subject line
func (am *AlertManager) generateEmailSubject(alert *storage.Alert, ctx *AlertContext) string {
	hostname := am.getHostname()
	return fmt.Sprintf("SysMedic Alert - %s Detected on %s",
		strings.Title(alert.Severity+" Load"), hostname)
}

// generateEmailBody creates email body content
func (am *AlertManager) generateEmailBody(alert *storage.Alert, ctx *AlertContext) string {
	hostname := am.getHostname()

	body := fmt.Sprintf(`SysMedic Alert Report

Server: %s
Timestamp: %s
Duration: %v
Severity: %s

System Status:
- CPU: %.1f%% (threshold: %d%%)
- Memory: %.1f%% (threshold: %d%%)
- Network: %.1f MB/s
- Load Average: %.2f, %.2f, %.2f

`,
		hostname,
		alert.Timestamp.Format("2006-01-02 15:04:05"),
		alert.Duration.Round(time.Minute),
		strings.Title(alert.Severity),
		ctx.SystemMetrics.CPUPercent,
		am.config.Monitoring.CPUThreshold,
		ctx.SystemMetrics.MemoryPercent,
		am.config.Monitoring.MemoryThreshold,
		ctx.SystemMetrics.NetworkMBps,
		ctx.SystemMetrics.LoadAvg1,
		ctx.SystemMetrics.LoadAvg5,
		ctx.SystemMetrics.LoadAvg15,
	)

	// Primary cause
	if alert.PrimaryCause != "" {
		body += fmt.Sprintf("Primary Cause: %s\n\n", alert.PrimaryCause)
	}

	// User breakdown
	if len(ctx.UserMetrics) > 0 {
		body += "User Breakdown:\n"
		for i, user := range ctx.UserMetrics {
			if i >= 5 { // Show top 5 users in email
				break
			}

			status := ""
			for _, pu := range ctx.PersistentUsers {
				if pu.Username == user.Username {
					status = " (⚠️ PERSISTENT)"
					break
				}
			}

			body += fmt.Sprintf("- %s: CPU %.1f%%, Memory %.1f%%, Processes: %d%s\n",
				user.Username,
				user.CPUPercent,
				user.MemoryPercent,
				user.ProcessCount,
				status,
			)
		}
		body += "\n"
	}

	// Persistent users details
	if len(ctx.PersistentUsers) > 0 {
		body += "Persistent Issues:\n"
		for _, pu := range ctx.PersistentUsers {
			body += fmt.Sprintf("- %s: %s usage %.1f%% for %v\n",
				pu.Username,
				pu.Metric,
				pu.CurrentUsage,
				pu.Duration.Round(time.Minute),
			)
		}
		body += "\n"
	}

	// Recommendations
	body += "Recommendations:\n"
	if len(ctx.Recommendations) > 0 {
		for _, rec := range ctx.Recommendations {
			body += fmt.Sprintf("- %s\n", rec)
		}
	} else {
		body += am.generateDefaultRecommendations(alert, ctx)
	}

	body += fmt.Sprintf("\n---\nGenerated by SysMedic at %s\n", time.Now().Format("2006-01-02 15:04:05"))

	return body
}

// generateDefaultRecommendations creates default recommendations based on alert context
func (am *AlertManager) generateDefaultRecommendations(alert *storage.Alert, ctx *AlertContext) string {
	var recommendations []string

	// Persistent user recommendations
	for _, pu := range ctx.PersistentUsers {
		recommendations = append(recommendations,
			fmt.Sprintf("Investigate %s processes for user %s", pu.Metric, pu.Username))
	}

	// High CPU recommendations
	if ctx.SystemMetrics.CPUPercent > float64(am.config.Monitoring.CPUThreshold) {
		recommendations = append(recommendations, "Check for CPU-intensive processes")
		recommendations = append(recommendations, "Consider process prioritization or resource limits")
	}

	// High memory recommendations
	if ctx.SystemMetrics.MemoryPercent > float64(am.config.Monitoring.MemoryThreshold) {
		recommendations = append(recommendations, "Check for memory leaks or excessive memory usage")
		recommendations = append(recommendations, "Consider adding more RAM or implementing memory limits")
	}

	// User-specific recommendations
	for _, user := range ctx.UserMetrics {
		cpuThreshold := am.config.GetUserThreshold(user.Username, "cpu")
		memThreshold := am.config.GetUserThreshold(user.Username, "memory")

		if user.CPUPercent > float64(cpuThreshold) {
			recommendations = append(recommendations,
				fmt.Sprintf("Review %s's processes - high CPU usage detected", user.Username))
		}
		if user.MemoryPercent > float64(memThreshold) {
			recommendations = append(recommendations,
				fmt.Sprintf("Review %s's processes - high memory usage detected", user.Username))
		}
	}

	if len(recommendations) == 0 {
		recommendations = append(recommendations, "Monitor system performance and user activity")
	}

	result := ""
	for _, rec := range recommendations {
		result += fmt.Sprintf("- %s\n", rec)
	}
	return result
}

// sendEmail sends an email using SMTP
func (am *AlertManager) sendEmail(subject, body string) error {
	emailConfig := am.config.Email

	// Create message
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s",
		emailConfig.To, subject, body)

	// Setup SMTP client
	auth := smtp.PlainAuth("", emailConfig.Username, emailConfig.Password, emailConfig.SMTPHost)

	// Setup TLS if enabled
	if emailConfig.TLS {
		return am.sendEmailWithTLS(auth, msg)
	}

	// Send without TLS
	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)
	return smtp.SendMail(addr, auth, emailConfig.From, []string{emailConfig.To}, []byte(msg))
}

// sendEmailWithTLS sends email with TLS encryption
func (am *AlertManager) sendEmailWithTLS(auth smtp.Auth, msg string) error {
	emailConfig := am.config.Email

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", emailConfig.SMTPHost, emailConfig.SMTPPort)

	// Setup TLS config
	tlsConfig := &tls.Config{
		ServerName: emailConfig.SMTPHost,
	}

	// Connect and start TLS
	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect with TLS: %w", err)
	}
	defer conn.Close()

	// Create SMTP client
	client, err := smtp.NewClient(conn, emailConfig.SMTPHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Quit()

	// Authenticate
	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Set sender
	if err := client.Mail(emailConfig.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	// Set recipient
	if err := client.Rcpt(emailConfig.To); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	// Send message
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = writer.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	return writer.Close()
}

// getHostname returns the system hostname
func (am *AlertManager) getHostname() string {
	hostname := "unknown"
	if data, err := os.ReadFile("/etc/hostname"); err == nil {
		hostname = strings.TrimSpace(string(data))
	}
	return hostname
}

// GenerateRecommendations creates contextual recommendations
func GenerateRecommendations(ctx *AlertContext, config *config.Config) []string {
	var recommendations []string

	// System-level recommendations
	if ctx.SystemMetrics.CPUPercent > float64(config.Monitoring.CPUThreshold) {
		recommendations = append(recommendations, "System CPU usage is high - check for runaway processes")
	}

	if ctx.SystemMetrics.MemoryPercent > float64(config.Monitoring.MemoryThreshold) {
		recommendations = append(recommendations, "System memory usage is high - check for memory leaks")
	}

	if ctx.SystemMetrics.LoadAvg1 > 2.0 {
		recommendations = append(recommendations, "High load average detected - system may be overloaded")
	}

	// User-specific recommendations
	for _, user := range ctx.UserMetrics {
		if user.ProcessCount > 20 {
			recommendations = append(recommendations,
				fmt.Sprintf("User %s has %d processes - check for process spawning issues",
					user.Username, user.ProcessCount))
		}
	}

	// Persistent user recommendations
	for _, pu := range ctx.PersistentUsers {
		recommendations = append(recommendations,
			fmt.Sprintf("User %s has sustained high %s usage - investigate immediately",
				pu.Username, pu.Metric))
	}

	return recommendations
}
