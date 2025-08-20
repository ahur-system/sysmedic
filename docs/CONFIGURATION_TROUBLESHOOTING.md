# SysMedic Configuration & Troubleshooting Guide

## Overview

This comprehensive guide covers SysMedic configuration management, troubleshooting procedures, and common issues with their solutions. It consolidates all configuration options, diagnostic procedures, and fixes from various system updates and improvements.

## Table of Contents

1. [Configuration Management](#configuration-management)
2. [System Configuration](#system-configuration)
3. [WebSocket Configuration](#websocket-configuration)
4. [Email Alert Configuration](#email-alert-configuration)
5. [User Filtering Configuration](#user-filtering-configuration)
6. [Troubleshooting Guide](#troubleshooting-guide)
7. [Common Issues & Solutions](#common-issues--solutions)
8. [Performance Optimization](#performance-optimization)
9. [Security Hardening](#security-hardening)
10. [Migration & Updates](#migration--updates)

## Configuration Management

### Configuration File Location
- **Primary**: `/etc/sysmedic/config.yaml`
- **User**: `~/.config/sysmedic/config.yaml`
- **Working Directory**: `./config.yaml`

### CLI Configuration Commands

```bash
# Display current configuration
sysmedic config show

# Display configuration with values and sources
sysmedic config show --verbose

# Validate configuration
sysmedic config validate

# Set configuration values
sysmedic config set monitoring.cpu_threshold 85
sysmedic config set websocket.port 8080
sysmedic config set email.enabled true

# Get specific configuration value
sysmedic config get monitoring.cpu_threshold

# Reset configuration to defaults
sysmedic config reset

# Backup current configuration
sysmedic config backup /path/to/backup.yaml

# Restore configuration from backup
sysmedic config restore /path/to/backup.yaml
```

### Environment Variable Overrides

```bash
# Override configuration via environment variables
export SYSMEDIC_MONITORING_CPU_THRESHOLD=90
export SYSMEDIC_WEBSOCKET_PORT=9090
export SYSMEDIC_EMAIL_ENABLED=true
export SYSMEDIC_DATA_PATH=/custom/data/path

# Run with environment overrides
sysmedic daemon start
```

## System Configuration

### Complete Configuration Schema

```yaml
# /etc/sysmedic/config.yaml

# System monitoring settings
monitoring:
  check_interval: 60                    # Monitoring frequency in seconds (30-300)
  cpu_threshold: 80                     # System CPU threshold percentage (50-95)
  memory_threshold: 80                  # System memory threshold percentage (50-95)
  disk_threshold: 90                    # Disk usage threshold percentage (70-99)
  load_threshold: 5.0                   # Load average threshold (1.0-20.0)
  persistent_time: 60                   # Minutes before flagging persistent users (15-240)
  
# User monitoring settings  
users:
  cpu_threshold: 80                     # Default user CPU threshold
  memory_threshold: 80                  # Default user memory threshold
  persistent_time: 60                   # User persistence detection time
  track_processes: true                 # Track individual processes
  max_processes_per_user: 100           # Limit processes tracked per user

# Per-user custom thresholds
user_thresholds:
  # Database users - higher CPU tolerance, strict memory
  postgres:
    cpu_threshold: 95
    memory_threshold: 70
    persistent_time: 120
  mysql:
    cpu_threshold: 90
    memory_threshold: 75
    persistent_time: 90
    
  # Web server users - balanced thresholds
  nginx:
    cpu_threshold: 85
    memory_threshold: 80
  apache:
    cpu_threshold: 85
    memory_threshold: 80
    
  # Development users - relaxed thresholds
  developer:
    cpu_threshold: 95
    memory_threshold: 90
    persistent_time: 180

# User filtering configuration
user_filtering:
  enabled: true                         # Enable user filtering
  min_uid_for_real_users: 1000         # Minimum UID for real users
  ignore_system_users: true            # Ignore system accounts
  min_cpu_percent: 5.0                 # Minimum CPU to track (0.1-20.0)
  min_memory_percent: 5.0              # Minimum memory to track (0.1-20.0)
  min_process_count: 2                 # Minimum processes to track user
  excluded_users:                      # Users to completely ignore
    - "root"
    - "daemon" 
    - "www-data"
    - "systemd-network"
    - "systemd-resolve"
    - "_apt"
  excluded_processes:                  # Process names to ignore
    - "kthreadd"
    - "ksoftirqd"
    - "systemd"
    - "dbus"

# WebSocket server configuration
websocket:
  enabled: true                        # Enable WebSocket server
  port: 8060                          # WebSocket server port (1024-65535)
  host: "0.0.0.0"                     # Bind address (0.0.0.0 for all interfaces)
  secret: ""                          # Authentication secret (auto-generated if empty)
  max_connections: 10                 # Maximum concurrent connections (1-100)
  ping_interval: 54                   # Ping interval in seconds (30-120)
  pong_timeout: 60                    # Pong timeout in seconds (30-180)
  write_timeout: 10                   # Write timeout in seconds (5-30)
  read_timeout: 60                    # Read timeout in seconds (30-300)
  buffer_size: 1024                   # Message buffer size in bytes
  
# Email alert configuration
email:
  enabled: false                      # Enable email alerts
  smtp_host: "smtp.gmail.com"         # SMTP server hostname
  smtp_port: 587                      # SMTP server port (25, 465, 587)
  use_tls: true                       # Use TLS encryption
  use_ssl: false                      # Use SSL encryption (alternative to TLS)
  from_email: "sysmedic@example.com"  # From email address
  from_name: "SysMedic Monitor"       # From display name
  to_emails:                          # List of recipient emails
    - "admin@example.com"
    - "ops-team@example.com"
  username: ""                        # SMTP username (if required)
  password: ""                        # SMTP password (if required)
  subject_prefix: "[SysMedic]"        # Email subject prefix
  send_recovery_alerts: true          # Send alerts when issues resolve
  alert_cooldown: 300                 # Seconds between duplicate alerts (60-3600)
  
# Storage configuration
storage:
  data_path: "/var/lib/sysmedic"      # Data directory path
  database_file: "sysmedic.db"       # SQLite database filename
  retention_days: 30                  # Data retention period (7-365)
  max_database_size: "100MB"          # Maximum database size
  backup_enabled: true                # Enable automatic backups
  backup_interval: "24h"              # Backup interval (1h-168h)
  backup_count: 7                     # Number of backups to keep (1-30)
  
# Logging configuration
logging:
  level: "info"                       # Log level (debug, info, warn, error)
  file: "/var/log/sysmedic/sysmedic.log" # Log file path
  max_size: "10MB"                    # Maximum log file size
  max_files: 5                        # Maximum number of log files
  console: false                      # Also log to console
  format: "json"                      # Log format (json, text)
  
# Performance tuning
performance:
  worker_count: 4                     # Number of worker goroutines (1-16)
  queue_size: 1000                    # Internal queue size (100-10000)
  batch_size: 100                     # Database batch insert size (10-1000)
  cache_ttl: 30                       # Cache TTL in seconds (10-300)
  gc_interval: 300                    # Garbage collection interval (60-3600)
```

### Configuration Validation Rules

```yaml
# Validation constraints applied to configuration
validation:
  monitoring:
    check_interval: { min: 30, max: 300 }
    cpu_threshold: { min: 50, max: 95 }
    memory_threshold: { min: 50, max: 95 }
    persistent_time: { min: 15, max: 240 }
    
  websocket:
    port: { min: 1024, max: 65535 }
    max_connections: { min: 1, max: 100 }
    ping_interval: { min: 30, max: 120 }
    
  storage:
    retention_days: { min: 7, max: 365 }
    backup_count: { min: 1, max: 30 }
```

## WebSocket Configuration

### Basic WebSocket Setup

```bash
# Enable WebSocket server
sysmedic config set websocket.enabled true

# Set custom port
sysmedic config set websocket.port 8080

# Generate new authentication secret
sysmedic websocket new-secret

# Start WebSocket daemon
sysmedic websocket start

# Check status with connection details
sysmedic websocket status
```

### Advanced WebSocket Configuration

```yaml
websocket:
  enabled: true
  port: 8060
  host: "0.0.0.0"                     # Bind to all interfaces
  secret: "your-secret-here"          # Set manually or use new-secret
  max_connections: 25                 # Allow more concurrent clients
  ping_interval: 45                   # More frequent pings
  pong_timeout: 90                    # Longer pong timeout
  write_timeout: 15                   # Longer write timeout
  read_timeout: 120                   # Longer read timeout
  buffer_size: 2048                   # Larger message buffer
  
  # CORS settings for web clients
  cors:
    enabled: true
    allowed_origins:
      - "https://monitor.example.com"
      - "http://localhost:3000"
    allowed_headers:
      - "X-SysMedic-Secret"
      - "Origin"
      - "Content-Type"
```

### WebSocket Security Configuration

```yaml
websocket:
  # Security settings
  secret_rotation_days: 30            # Rotate secret every 30 days
  require_secret: true                # Always require authentication
  rate_limiting:
    enabled: true
    requests_per_minute: 60           # Limit requests per client
    burst_size: 10                    # Allow burst of requests
  
  # TLS configuration (when behind proxy)
  tls:
    enabled: false                    # Enable direct TLS (not recommended)
    cert_file: "/path/to/cert.pem"
    key_file: "/path/to/key.pem"
```

### Quick Connect URL Configuration

```bash
# The Quick Connect URL feature provides easy client setup
# Format: sysmedic://[secret]@[public_ip]:[port]/

# Configure public IP detection
sysmedic config set websocket.public_ip_services '[
  "https://api.ipify.org",
  "https://ifconfig.me", 
  "https://icanhazip.com"
]'

# Set custom public IP (if auto-detection fails)
sysmedic config set websocket.public_ip "203.0.113.45"

# Disable public IP detection (use local IP)
sysmedic config set websocket.detect_public_ip false
```

## Email Alert Configuration

### Basic Email Setup

```yaml
email:
  enabled: true
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  use_tls: true
  from_email: "sysmedic@your-domain.com"
  from_name: "SysMedic Server Monitor"
  to_emails:
    - "admin@your-domain.com"
  username: "your-smtp-username"
  password: "your-app-password"      # Use app password for Gmail
```

### Gmail Configuration

```yaml
email:
  enabled: true
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  use_tls: true
  from_email: "your-account@gmail.com"
  username: "your-account@gmail.com"
  password: "your-16-char-app-password"  # Generate in Google Account settings
  to_emails:
    - "recipient@gmail.com"
```

### Corporate Email Configuration

```yaml
# Microsoft Exchange/Office 365
email:
  enabled: true
  smtp_host: "smtp-mail.outlook.com"
  smtp_port: 587
  use_tls: true
  from_email: "sysmedic@your-company.com"
  username: "sysmedic@your-company.com"
  password: "your-password"

# Custom SMTP server
email:
  enabled: true
  smtp_host: "mail.your-company.com"
  smtp_port: 25
  use_tls: false
  use_ssl: false
  from_email: "monitoring@your-company.com"
  # No authentication required for internal SMTP
```

### Advanced Email Settings

```yaml
email:
  enabled: true
  smtp_host: "smtp.example.com"
  smtp_port: 587
  use_tls: true
  
  # Advanced settings
  subject_prefix: "[ALERT-PROD]"      # Custom subject prefix
  send_recovery_alerts: true          # Send "resolved" notifications
  alert_cooldown: 600                 # 10 minutes between duplicate alerts
  max_alerts_per_hour: 12            # Rate limiting
  
  # Message formatting
  include_system_info: true           # Include system details
  include_recommendations: true       # Include actionable suggestions
  html_format: true                   # Send HTML emails
  
  # Multiple recipient groups
  to_emails:
    - "ops-team@example.com"
    - "on-call@example.com"
  cc_emails:
    - "manager@example.com"
  bcc_emails:
    - "monitoring-archive@example.com"
```

## User Filtering Configuration

### Basic User Filtering

```yaml
user_filtering:
  enabled: true
  min_uid_for_real_users: 1000        # Standard Linux user UID start
  ignore_system_users: true           # Skip system accounts
  min_cpu_percent: 10.0               # Only track users using >10% CPU
  min_memory_percent: 5.0             # Only track users using >5% memory
```

### Advanced User Filtering

```yaml
user_filtering:
  enabled: true
  min_uid_for_real_users: 1000
  ignore_system_users: true
  min_cpu_percent: 5.0
  min_memory_percent: 5.0
  min_process_count: 2                # User must have >2 processes
  
  # Comprehensive exclusion lists
  excluded_users:
    # System users
    - "root"
    - "daemon"
    - "bin"
    - "sys"
    - "sync"
    - "games"
    - "man"
    - "lp"
    - "mail"
    - "news"
    - "uucp"
    - "proxy"
    - "backup"
    - "list"
    - "irc"
    - "gnats"
    - "nobody"
    - "systemd-network"
    - "systemd-resolve"
    - "systemd-timesync"
    - "_apt"
    - "messagebus"
    - "avahi-autoipd"
    - "usbmux"
    - "dnsmasq"
    - "rtkit"
    - "cups-pk-helper"
    - "uuidd"
    - "nm-openvpn"
    - "nm-openconnect"
    - "pulse"
    - "avahi"
    - "colord"
    - "hplip"
    - "geoclue"
    - "gnome-initial-setup"
    
  # Process filtering
  excluded_processes:
    # Kernel threads
    - "kthreadd"
    - "ksoftirqd"
    - "migration"
    - "rcu_"
    - "watchdog"
    - "systemd"
    - "systemd-"
    - "dbus"
    - "NetworkManager"
    - "wpa_supplicant"
    - "avahi-daemon"
    - "cups"
    - "pulseaudio"
    
  # Include specific users even if they don't meet criteria
  always_include:
    - "database"
    - "webserver"
    - "application"
    
  # User group filtering
  excluded_groups:
    - "systemd-journal"
    - "systemd-network"
    - "systemd-resolve"
```

### Dynamic User Classification

```yaml
user_filtering:
  # Automatic user classification
  user_classes:
    system:
      uid_range: [0, 999]
      always_exclude: true
    
    service:
      uid_range: [1000, 1999]
      min_cpu_percent: 15.0           # Higher threshold for service users
      min_memory_percent: 10.0
      
    interactive:
      uid_range: [2000, 59999]
      min_cpu_percent: 5.0            # Lower threshold for interactive users
      min_memory_percent: 5.0
      track_login_sessions: true
      
    application:
      uid_range: [60000, 65533]
      min_cpu_percent: 20.0           # High threshold for app users
      min_memory_percent: 15.0
```

## Troubleshooting Guide

### Diagnostic Commands

```bash
# System status overview
sysmedic status

# Detailed daemon status
sysmedic daemon status --verbose

# WebSocket server diagnostics
sysmedic websocket status --verbose

# Configuration validation
sysmedic config validate --detailed

# System health check
sysmedic health-check

# View recent alerts
sysmedic alerts list --recent

# Database statistics
sysmedic db stats

# Performance metrics
sysmedic metrics --system --users --performance
```

### Log Analysis

```bash
# View daemon logs
sudo journalctl -u sysmedic.doctor -f

# View WebSocket logs
sudo journalctl -u sysmedic.websocket -f

# View application logs
tail -f /var/log/sysmedic/sysmedic.log

# Search for errors
grep -i error /var/log/sysmedic/*.log

# View logs with specific time range
journalctl -u sysmedic.doctor --since "1 hour ago"

# Export logs for analysis
journalctl -u sysmedic.doctor --since "1 day ago" > /tmp/sysmedic-logs.txt
```

### Network Diagnostics

```bash
# Check WebSocket port
netstat -tulpn | grep :8060
ss -tulpn | grep :8060

# Test WebSocket connectivity
curl -i http://localhost:8060/health
curl -i http://localhost:8060/status

# Test from remote host
curl -i http://your-server-ip:8060/health

# WebSocket connection test
websocat ws://localhost:8060/ws?secret=your-secret

# Firewall check
sudo ufw status
sudo iptables -L INPUT | grep 8060
```

### Database Diagnostics

```bash
# Check database file
ls -la /var/lib/sysmedic/sysmedic.db

# Database integrity check
sqlite3 /var/lib/sysmedic/sysmedic.db "PRAGMA integrity_check;"

# Database size and usage
du -h /var/lib/sysmedic/sysmedic.db
sqlite3 /var/lib/sysmedic/sysmedic.db ".dbinfo"

# View table statistics
sqlite3 /var/lib/sysmedic/sysmedic.db "
  SELECT 
    name,
    COUNT(*) as records 
  FROM (
    SELECT 'system_metrics' as name FROM system_metrics
    UNION ALL
    SELECT 'user_metrics' as name FROM user_metrics
    UNION ALL
    SELECT 'alerts' as name FROM alerts
    UNION ALL
    SELECT 'persistent_users' as name FROM persistent_users
  ) 
  GROUP BY name;
"
```

## Common Issues & Solutions

### Daemon Issues

#### Doctor Daemon Won't Start

**Symptoms:**
```bash
$ sysmedic daemon start
Error: Failed to start doctor daemon
```

**Diagnosis:**
```bash
# Check daemon status
sudo systemctl status sysmedic.doctor

# View detailed logs
sudo journalctl -u sysmedic.doctor -n 50

# Check configuration
sysmedic config validate
```

**Common Causes & Solutions:**

1. **Configuration Error:**
   ```bash
   # Fix configuration syntax
   sysmedic config validate --fix
   
   # Reset to defaults if corrupted
   sudo cp /etc/sysmedic/config.yaml /etc/sysmedic/config.yaml.backup
   sysmedic config reset
   ```

2. **Permission Issues:**
   ```bash
   # Fix data directory permissions
   sudo chown -R root:root /var/lib/sysmedic
   sudo chmod 755 /var/lib/sysmedic
   
   # Fix log directory permissions
   sudo mkdir -p /var/log/sysmedic
   sudo chown -R root:root /var/log/sysmedic
   sudo chmod 755 /var/log/sysmedic
   ```

3. **Database Lock:**
   ```bash
   # Check for lock file
   ls -la /var/lib/sysmedic/sysmedic.db-*
   
   # Remove stale locks (be careful!)
   sudo rm -f /var/lib/sysmedic/sysmedic.db-wal
   sudo rm -f /var/lib/sysmedic/sysmedic.db-shm
   ```

4. **Port Already in Use:**
   ```bash
   # Find process using port
   sudo lsof -i :8060
   sudo netstat -tulpn | grep :8060
   
   # Kill conflicting process or change port
   sysmedic config set websocket.port 8061
   ```

#### WebSocket Daemon Issues

**Connection Refused:**
```bash
# Check if daemon is running
sysmedic websocket status

# Check port binding
sudo netstat -tulpn | grep :8060

# Check firewall
sudo ufw status
sudo firewall-cmd --list-ports
```

**Authentication Failed:**
```bash
# Check secret configuration
sysmedic config get websocket.secret

# Generate new secret
sysmedic websocket new-secret

# Test with correct secret
curl "http://localhost:8060/ws?secret=$(sysmedic config get websocket.secret)"
```

**Random Disconnections:**
```bash
# Check WebSocket daemon logs
sudo journalctl -u sysmedic.websocket -f

# Adjust timeout settings
sysmedic config set websocket.ping_interval 30
sysmedic config set websocket.pong_timeout 90
sysmedic config set websocket.write_timeout 15

# Restart WebSocket daemon
sysmedic websocket stop
sysmedic websocket start
```

### Configuration Issues

#### Invalid Configuration

**Symptoms:**
```bash
$ sysmedic daemon start
Error: Invalid configuration: cpu_threshold must be between 50 and 95
```

**Solutions:**
```bash
# Validate configuration
sysmedic config validate --detailed

# Show configuration with validation
sysmedic config show --validate

# Fix specific values
sysmedic config set monitoring.cpu_threshold 80

# Reset invalid section
sysmedic config reset monitoring
```

#### Missing Configuration File

**Symptoms:**
```bash
$ sysmedic daemon start
Error: Configuration file not found: /etc/sysmedic/config.yaml
```

**Solutions:**
```bash
# Create default configuration
sudo mkdir -p /etc/sysmedic
sysmedic config init

# Copy example configuration
sudo cp /usr/share/doc/sysmedic/config.example.yaml /etc/sysmedic/config.yaml

# Generate minimal configuration
sysmedic config create-minimal > /tmp/config.yaml
sudo mv /tmp/config.yaml /etc/sysmedic/config.yaml
```

### Performance Issues

#### High CPU Usage

**Symptoms:**
```bash
# SysMedic itself using high CPU
top -p $(pgrep sysmedic)
```

**Solutions:**
```bash
# Increase monitoring interval
sysmedic config set monitoring.check_interval 120

# Reduce user filtering sensitivity
sysmedic config set user_filtering.min_cpu_percent 10.0
sysmedic config set user_filtering.min_memory_percent 10.0

# Limit process tracking
sysmedic config set users.max_processes_per_user 50

# Optimize database
sqlite3 /var/lib/sysmedic/sysmedic.db "VACUUM;"
sqlite3 /var/lib/sysmedic/sysmedic.db "REINDEX;"
```

#### Database Growing Too Large

**Symptoms:**
```bash
$ du -h /var/lib/sysmedic/sysmedic.db
500M    /var/lib/sysmedic/sysmedic.db
```

**Solutions:**
```bash
# Reduce retention period
sysmedic config set storage.retention_days 14

# Clean old data immediately
sysmedic db cleanup --days 14

# Enable automatic cleanup
sysmedic config set storage.auto_cleanup true

# Set database size limit
sysmedic config set storage.max_database_size "200MB"
```

#### Memory Usage Issues

**Solutions:**
```bash
# Reduce cache TTL
sysmedic config set performance.cache_ttl 15

# Reduce worker count
sysmedic config set performance.worker_count 2

# Reduce queue size
sysmedic config set performance.queue_size 500

# Enable more frequent garbage collection
sysmedic config set performance.gc_interval 180
```

### Email Alert Issues

#### Emails Not Sending

**Diagnosis:**
```bash
# Test email configuration
sysmedic email test

# Check email logs
grep -i email /var/log/sysmedic/sysmedic.log

# Verify SMTP settings
sysmedic config show email
```

**Common Solutions:**

1. **Gmail App Password:**
   ```bash
   # Use app password instead of regular password
   # Generate at: https://myaccount.google.com/apppasswords
   sysmedic config set email.password "your-16-char-app-password"
   ```

2. **SMTP Authentication:**
   ```bash
   # Enable authentication for servers that require it
   sysmedic config set email.username "your-smtp-username"
   sysmedic config set email.password "your-smtp-password"
   ```

3. **TLS/SSL Issues:**
   ```bash
   # Try different encryption settings
   sysmedic config set email.use_tls false
   sysmedic config set email.use_ssl true
   
   # Or disable encryption for internal SMTP
   sysmedic config set email.use_tls false
   sysmedic config set email.use_ssl false
   ```

#### Too Many Alert Emails

**Solutions:**
```bash
# Increase alert cooldown
sysmedic config set email.alert_cooldown 1800  # 30 minutes

# Set maximum alerts per hour
sysmedic config set email.max_alerts_per_hour 6

# Disable recovery alerts if too noisy
sysmedic config set email.send_recovery_alerts false

# Increase thresholds to reduce false positives
sysmedic config set monitoring.cpu_threshold 85
sysmedic config set monitoring.persistent_time 90
```

### Network and Firewall Issues

#### WebSocket Port Blocked

**Symptoms:**
```bash
$ curl http://server-ip:8060/health
curl: (7) Failed to connect to server-ip port 8060: Connection refused
```

**Solutions:**

1. **UFW Firewall:**
   ```bash
   sudo ufw allow 8060/tcp
   sudo ufw reload
   sudo ufw status
   ```

2. **iptables:**
   ```bash
   sudo iptables -A INPUT -p tcp --dport 8060 -j ACCEPT
   sudo iptables-save > /etc/iptables/rules.v4
   ```

3. **firewalld (CentOS/RHEL):**
   ```bash
   sudo firewall-cmd --permanent --add-port=8060/tcp
   sudo firewall-cmd --reload
   sudo firewall-cmd --list-ports
   ```

4. **Cloud Provider Security Groups:**
   ```bash
   # AWS: Add inbound rule for port 8060
   # GCP: Add firewall rule for port 8060
   # Azure: Add network security group rule
   ```

#### Network Interface Binding

**Issue:** WebSocket only accessible locally

**Solutions:**
```bash
# Bind to all interfaces
sysmedic config set websocket.host "0.0.0.0"

# Bind to specific interface
sysmedic config set websocket.host "192.168.1.100"

# Restart WebSocket daemon
sysmedic websocket stop
sysmedic websocket start
```

### Persistent User Detection Issues

#### Users Not Being Detected as Persistent

**Solutions:**
```bash
# Lower persistent time threshold
sysmedic config set monitoring.persistent_time 30

# Lower CPU/memory thresholds for detection
sysmedic config set monitoring.cpu_threshold 70
sysmedic config set users.cpu_threshold 70

# Check user filtering settings
sysmedic config show user_filtering

# Remove user from exclusion list
sysmedic config set user_filtering.excluded_users '["root", "daemon"]'
```

#### False Positive Persistent Users

**Solutions:**
```bash
# Increase persistent time threshold
sysmedic config set monitoring.persistent_time 120

# Increase thresholds
sysmedic config set users.cpu_threshold 85
sysmedic config set users.memory_threshold 85

# Add user to exclusion list
sysmedic config set user_filtering.excluded_users '["root", "daemon", "backup-user"]'

# Set custom threshold for specific user
sysmedic config set user_thresholds.backup-user.cpu_threshold 95
```

## Performance Optimization

### System Performance Tuning

```yaml
# Optimized configuration for high-performance systems
monitoring:
  check_interval: 30                  # More frequent monitoring
  
performance:
  worker_count: 8                     # More workers for multi-core systems
  queue_size: 2000                    # Larger queue for high throughput
  batch_size: 200                     # Larger batches for better DB performance
  cache_ttl: 60                       # Longer cache for stable systems
  gc_interval: 600                    # Less frequent GC for better performance

storage:
  retention_days: 15                  # Shorter retention for better performance
  max_database_size: "500MB"          # Allow larger database
  
user_filtering:
  min_cpu_percent: 15.0              # Higher threshold to reduce noise
  min_memory_percent: 10.0            # Higher threshold to reduce noise
  min_process_count: 3                # Require more processes to track user
```

### Memory Optimization

```yaml
# Memory-constrained system configuration
performance:
  worker_count: 2                     # Fewer workers to save memory
  queue_size: 500                     # Smaller queue
  batch_size: 50                      # Smaller batches
  cache_ttl: 15                       # Shorter cache TTL
  gc_interval: 120                    # More frequent garbage collection

storage:
  retention_days: 7                   # Short retention
  max_database_size: "50MB"           # Small database limit
  
monitoring:
  check_interval: 120                 # Less frequent monitoring
  
user_filtering:
  min_cpu_percent: 20.0              # Very high thresholds
  min_memory_percent: 15.0
```

### Database Optimization

```bash
# Regular database maintenance
# Add to cron: 0 2 * * * /usr/local/bin/sysmedic db maintenance

# Vacuum database weekly
sqlite3 /var/lib/sysmedic/sysme