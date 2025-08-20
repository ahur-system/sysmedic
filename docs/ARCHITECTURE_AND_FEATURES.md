# SysMedic Architecture & Features Guide

## Overview

SysMedic is a revolutionary single binary multi-daemon Linux system monitoring tool that provides the best of both worlds: deployment simplicity and process modularity. This document covers the complete architecture, features, and design principles of SysMedic.

## Table of Contents

1. [Single Binary Multi-Daemon Architecture](#single-binary-multi-daemon-architecture)
2. [Core Components](#core-components)
3. [Features Overview](#features-overview)
4. [System Monitoring Logic](#system-monitoring-logic)
5. [Data Storage & Management](#data-storage--management)
6. [Alert System](#alert-system)
7. [Configuration System](#configuration-system)
8. [Security Model](#security-model)
9. [Performance Characteristics](#performance-characteristics)
10. [File Locations](#file-locations)

## Single Binary Multi-Daemon Architecture

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                        Single Binary (sysmedic)                             │
│                                  11MB                                       │
└─────────────────────────────────────────────────────────────────────────────┘
                                    │
            ┌───────────────────────┼───────────────────────┐
            │                       │                       │
            ▼                       ▼                       ▼
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Mode      │    │  Doctor Daemon   │    │ WebSocket       │
│   (Default)     │    │ --doctor-daemon  │    │ Daemon          │
│                 │    │                  │    │ --websocket-    │
│ • Dashboard     │    │ • System Monitor │    │ daemon          │
│ • Config Mgmt   │    │ • User Tracking  │    │                 │
│ • Reports       │    │ • Alert Engine   │    │ • Remote Access │
│ • Daemon Ctrl   │    │ • Data Storage   │    │ • Real-time API │
│ • Process Mgmt  │    │ • Persistence    │    │ • Auth Secrets  │
└─────────────────┘    └──────────────────┘    └─────────────────┘
            │                       │                       │
            │                       ▼                       │
            │              ┌─────────────────┐              │
            │              │   Storage       │              │
            │              │                 │              │
            │              │ • SQLite DB     │              │
            │              │ • Metrics       │              │
            │              │ • Alerts        │              │
            │              │ • User Activity │              │
            │              └─────────────────┘              │
            │                       │                       │
            │                       ▼                       │
            │              ┌─────────────────┐              │
            │              │  Email Alerts   │              │
            │              │                 │              │
            │              │ • SMTP/TLS      │              │
            │              │ • Contextual    │              │
            │              │ • Actionable    │              │
            │              └─────────────────┘              │
            │                                               │
            └───────────────── Management ──────────────────┘
```

### Daemon Modes

#### Doctor Daemon Mode (`sysmedic --doctor-daemon`)
- **Purpose**: Core system and user monitoring functionality
- **Responsibilities**:
  - Monitor system metrics (CPU, memory, network, load averages)
  - Track user resource usage with smart filtering
  - Generate alerts based on configurable thresholds
  - Store monitoring data in SQLite database
  - Manage persistent user tracking
- **PID File**: `/var/lib/sysmedic/sysmedic.doctor.pid`
- **Service**: `sysmedic.doctor.service`

#### WebSocket Daemon Mode (`sysmedic --websocket-daemon`)
- **Purpose**: Remote monitoring access via WebSocket connections
- **Responsibilities**:
  - Serve WebSocket connections for remote clients
  - Authenticate clients using secret tokens
  - Provide real-time system data updates
  - Serve HTTP status and health endpoints
- **PID File**: `/var/lib/sysmedic/sysmedic.websocket.pid`
- **Service**: `sysmedic.websocket.service`
- **Endpoints**:
  - `/ws` - WebSocket connection endpoint
  - `/status` - JSON status information
  - `/health` - Health check endpoint

#### CLI Mode (default)
- **Purpose**: Command-line interface for managing both daemon processes
- **Responsibilities**:
  - Start/stop/status management for both daemon processes
  - Spawn background daemon processes using the same binary
  - Configuration management and validation
  - Display dashboards and reports
  - Alert management and resolution
  - WebSocket configuration and secret generation

### Benefits of Single Binary Multi-Daemon Architecture

#### Deployment Simplicity
- Only one binary to distribute and manage
- Simpler packaging (DEB, RPM, tarball)
- Reduced disk space and memory overhead
- Single version management across all components
- Easier updates and deployments

#### Process Independence
- Daemon processes can be started, stopped, and restarted independently
- WebSocket server can be managed without affecting monitoring
- Better resource isolation and fault tolerance
- Each process has its own PID and lifecycle
- Independent crash recovery

#### Scalability & Performance
- Each daemon process optimized for its specific workload
- Better performance monitoring per component
- Independent resource allocation
- Easier horizontal scaling if needed

#### Security & Maintenance
- Reduced attack surface per daemon process
- Different security policies can be applied to each daemon
- Process isolation between monitoring and web services
- Easier debugging and troubleshooting
- Cleaner separation of concerns

## Core Components

### 1. Command Line Interface (`pkg/cli/`)
- **File**: `cli.go` (623 lines)
- **Purpose**: User-facing commands and daemon management
- **Features**:
  - Real-time dashboard display
  - Configuration management
  - Report generation
  - Dual daemon process control
  - WebSocket server management
  - Quick Connect URL generation

### 2. Doctor Daemon (`internal/monitor/`)
- **Files**: `monitor.go` (506 lines), `monitor_test.go` (388 lines)
- **Purpose**: Independent system monitoring daemon process
- **Process**: Runs as `--doctor-daemon` mode
- **Features**:
  - CPU usage calculation from `/proc/stat`
  - Memory usage from `/proc/meminfo`
  - Network statistics from `/proc/net/dev`
  - Per-user resource tracking via `/proc/[pid]/`
  - Load average monitoring
  - Process enumeration and analysis

### 3. WebSocket Daemon (`internal/websocket/`)
- **Files**: `server.go`, `manager.go`
- **Purpose**: Independent remote access daemon process
- **Process**: Runs as `--websocket-daemon` mode
- **Features**:
  - Real-time WebSocket connections
  - Secret-based authentication
  - Independent process lifecycle
  - HTTP health endpoints
  - Client connection management
  - Live system data streaming

### 4. Background Daemon (`internal/daemon/`)
- **File**: `daemon.go` (458 lines)
- **Purpose**: Doctor daemon process management and coordination
- **Features**:
  - Configurable monitoring intervals
  - Independent PID file management
  - Signal handling (SIGTERM, SIGINT)
  - Persistent user state tracking
  - Automatic data cleanup
  - Dual systemd service integration

### 5. Configuration Management (`internal/config/`)
- **Files**: `config.go` (235 lines), `config_test.go` (401 lines)
- **Purpose**: Shared YAML-based configuration for all daemon modes
- **Features**:
  - System-wide threshold settings
  - Per-user custom thresholds
  - Email notification configuration
  - WebSocket daemon configuration
  - Flexible threshold management
  - Hot configuration updates

### 6. Data Storage (`internal/storage/`)
- **File**: `storage.go` (591 lines)
- **Purpose**: SQLite-based data persistence (shared by both daemons)
- **Features**:
  - System metrics storage
  - User activity tracking
  - Alert history management
  - Persistent user records
  - Configurable data retention
  - Database statistics and maintenance

### 7. Alert Management (`internal/alerts/`)
- **File**: `alerts.go` (462 lines)
- **Purpose**: Intelligent alerting with email notifications
- **Features**:
  - Context-aware alert generation
  - SMTP/TLS email delivery
  - Alert severity classification
  - User breakdown in alerts
  - Actionable recommendations
  - Alert deduplication logic

### 8. Main Application (`cmd/sysmedic/`)
- **File**: `main.go` (136 lines)
- **Purpose**: Single binary entry point with multi-mode support
- **Features**:
  - Daemon mode detection (`--doctor-daemon`, `--websocket-daemon`)
  - Cobra CLI command structure
  - Version information
  - Process spawning and management

## Features Overview

### Core Monitoring Features

#### System Metrics
- **CPU Usage**: Real-time CPU utilization monitoring
- **Memory Usage**: Active memory consumption tracking
- **Disk Usage**: Root filesystem usage monitoring
- **Network I/O**: Network interface statistics
- **Load Averages**: 1, 5, and 15-minute load averages
- **Uptime Tracking**: System uptime with human-readable formatting

#### User-Centric Monitoring
- **Per-User Resource Tracking**: Individual user CPU and memory usage
- **Process Enumeration**: Detailed process information per user
- **Smart User Filtering**: Exclude system users and low-impact processes
- **Persistent User Detection**: Identify users with sustained high usage
- **User Activity History**: Track user resource patterns over time

#### Advanced Features
- **Configurable Thresholds**: System-wide and per-user threshold settings
- **Status Classification**: Intelligent system status determination
- **Real-time Dashboard**: Live CLI dashboard with color coding
- **Historical Data**: SQLite-based data retention and analysis
- **Email Alerting**: SMTP/TLS email notifications with context
- **WebSocket API**: Real-time remote monitoring capabilities

### WebSocket Features

#### Remote Access
- **Real-time Streaming**: Live system data via WebSocket connections
- **Multiple Clients**: Support for concurrent client connections
- **Authentication**: Secret-based authentication system
- **Quick Connect URLs**: Easy connection setup with `sysmedic://` URLs
- **HTTP Endpoints**: Health check and status endpoints

#### Message Types
- **Welcome Messages**: Connection confirmation with system info
- **System Updates**: Real-time metrics every 3 seconds
- **Request-Response**: Bidirectional communication protocol
- **Error Handling**: Proper error responses and connection management

### Management Features

#### Daemon Control
- **Independent Control**: Start/stop daemons independently
- **Status Monitoring**: Real-time daemon status checking
- **Process Management**: PID tracking and lifecycle management
- **Service Integration**: systemd service file support
- **Log Management**: Comprehensive logging with rotation

#### Configuration Management
- **YAML Configuration**: Human-readable configuration files
- **Hot Reloading**: Runtime configuration updates
- **Validation**: Configuration validation and error reporting
- **Defaults**: Sensible default settings for all options
- **Environment Support**: Environment variable overrides

## System Monitoring Logic

### CPU Usage Calculation
```go
func (m *Monitor) getCPUUsage() (float64, error) {
    // Reads /proc/stat
    // Calculates delta between measurements
    // Returns usage percentage
    totalDelta := (total2 - total1)
    idleDelta := (idle2 - idle1)
    cpuUsage := 100.0 * (1.0 - float64(idleDelta)/float64(totalDelta))
    return cpuUsage, nil
}
```

### Memory Usage Calculation
```go
func (m *Monitor) getMemoryUsage() (float64, error) {
    // Reads /proc/meminfo
    // MemTotal - MemFree - Buffers - Cached
    usedMemory := totalMemory - freeMemory - buffers - cached
    memoryUsage := 100.0 * float64(usedMemory) / float64(totalMemory)
    return memoryUsage, nil
}
```

### Per-User Process Tracking
```go
func (m *Monitor) getProcesses() ([]ProcessInfo, error) {
    // Walk /proc/[pid] directories
    // Parse stat, status, and comm files
    // Group by username
    // Filter based on user filtering criteria
}
```

### Status Classification Logic
```go
func DetermineSystemStatus(systemMetrics, userMetrics, thresholds, persistentUsers) string {
    // Heavy Load: System thresholds exceeded OR persistent users detected
    if systemCPU > threshold || systemMemory > threshold {
        return "Heavy Load"
    }
    if len(persistentUsers) > 0 {
        return "Heavy Load"
    }
    
    // Medium Usage: Above 60% of thresholds
    if systemCPU > (threshold * 0.6) || systemMemory > (threshold * 0.6) {
        return "Medium Usage"
    }
    
    // Check for temporary user spikes
    for _, user := range userMetrics {
        if user.CPUPercent > threshold || user.MemoryPercent > threshold {
            return "Medium Usage"
        }
    }
    
    return "Light Usage"
}
```

### Persistent User Detection
```go
type UserTrackingState struct {
    CPUStartTime    *time.Time
    MemoryStartTime *time.Time
    CPUPeakUsage    float64
    MemoryPeakUsage float64
    CPUSampleCount  int
    CPUTotalUsage   float64
    // Additional tracking fields...
}

// Users flagged as persistent when:
// - Above threshold for configurable time (default 60 minutes)
// - Tracked separately for CPU and memory usage
// - State persisted across daemon restarts
// - Automatic cleanup when usage drops below thresholds
```

## Data Storage & Management

### Database Schema

#### System Metrics Table
```sql
CREATE TABLE system_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    cpu_percent REAL NOT NULL,
    memory_percent REAL NOT NULL,
    disk_percent REAL NOT NULL,
    network_mbps REAL DEFAULT 0,
    load_avg_1 REAL DEFAULT 0,
    load_avg_5 REAL DEFAULT 0,
    load_avg_15 REAL DEFAULT 0
);
```

#### User Metrics Table
```sql
CREATE TABLE user_metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    username TEXT NOT NULL,
    cpu_percent REAL NOT NULL,
    memory_percent REAL NOT NULL,
    process_count INTEGER DEFAULT 0,
    pids TEXT DEFAULT ''
);
```

#### Alerts Table
```sql
CREATE TABLE alerts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
    alert_type TEXT NOT NULL,
    severity TEXT NOT NULL,
    message TEXT NOT NULL,
    system_cpu REAL DEFAULT 0,
    system_memory REAL DEFAULT 0,
    duration_minutes INTEGER DEFAULT 0,
    primary_cause TEXT DEFAULT '',
    user_details TEXT DEFAULT '',
    resolved BOOLEAN DEFAULT 0
);
```

#### Persistent Users Table
```sql
CREATE TABLE persistent_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    username TEXT NOT NULL,
    metric TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME,
    duration_minutes INTEGER DEFAULT 0,
    peak_usage REAL NOT NULL,
    average_usage REAL DEFAULT 0,
    sample_count INTEGER DEFAULT 0,
    resolved BOOLEAN DEFAULT 0
);
```

### Data Retention
- **Configurable Retention**: Automatic cleanup of old data
- **Size Management**: Database size monitoring and maintenance
- **Indexing**: Optimized queries for performance
- **Backup Support**: SQLite backup and restore capabilities

## Alert System

### Alert Types
- **System Alerts**: CPU/Memory threshold violations
- **User Alerts**: Per-user resource violations
- **Persistent User Alerts**: Sustained high usage detection
- **Recovery Alerts**: System recovery notifications

### Alert Severity Levels
- **Critical**: System thresholds exceeded with persistent users
- **High**: System thresholds exceeded
- **Medium**: Individual user thresholds exceeded
- **Low**: Warning levels approaching thresholds

### Email Alert Features
```go
type AlertManager struct {
    SMTPConfig SMTPConfig
    Templates  AlertTemplates
}

func (am *AlertManager) sendEmailWithTLS(auth, msg) error {
    // TLS connection establishment
    // SMTP authentication with secure protocols
    // Message delivery with proper MIME formatting
    // Error handling and retry logic
}
```

#### Alert Content Structure
- **System Status**: Current system metrics and status
- **User Breakdown**: Detailed user resource usage
- **Persistent Users**: List of users with sustained high usage
- **Duration Information**: How long the condition has persisted
- **Recommendations**: Actionable suggestions for resolution
- **System Context**: Load averages, uptime, and system info

### Alert Deduplication
- **Time-based Suppression**: Avoid duplicate alerts within time windows
- **Status Change Detection**: Only alert on status changes
- **Recovery Notifications**: Automatic recovery alerts when issues resolve

## Configuration System

### Configuration Structure
```yaml
# System monitoring settings
monitoring:
  check_interval: 60          # Monitoring frequency in seconds
  cpu_threshold: 80           # System CPU threshold percentage
  memory_threshold: 80        # System memory threshold percentage
  persistent_time: 60         # Minutes before flagging persistent users
  
# User monitoring settings  
users:
  cpu_threshold: 80           # Default user CPU threshold
  memory_threshold: 80        # Default user memory threshold
  persistent_time: 60         # User persistence detection time

# Per-user custom thresholds
user_thresholds:
  database_user:
    cpu_threshold: 90         # Higher tolerance for database users
    memory_threshold: 70      # But stricter memory limits
  webserver_user:
    cpu_threshold: 85
    memory_threshold: 75

# User filtering configuration
user_filtering:
  min_uid_for_real_users: 1000        # Minimum UID for real users
  ignore_system_users: true           # Ignore system accounts
  min_cpu_percent: 5.0                # Minimum CPU to track
  min_memory_percent: 5.0             # Minimum memory to track
  excluded_users: ["root", "daemon", "www-data"]

# WebSocket server configuration
websocket:
  enabled: true               # Enable WebSocket server
  port: 8060                 # WebSocket server port
  secret: "generated-secret" # Authentication secret

# Email alert configuration
email:
  enabled: false             # Enable email alerts
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  from_email: "alerts@example.com"
  to_emails: ["admin@example.com"]
  username: "smtp_username"
  password: "smtp_password"

# Storage configuration
data_path: "/var/lib/sysmedic"
database_retention_days: 30
```

### Configuration Management Features
- **Validation**: Comprehensive configuration validation
- **Defaults**: Sensible defaults for all settings
- **Environment Variables**: Override configuration via environment
- **Hot Reloading**: Runtime configuration updates
- **Migration**: Automatic configuration migration between versions

## Security Model

### Authentication
- **WebSocket Secrets**: Cryptographically secure authentication tokens
- **Secret Rotation**: Easy secret regeneration and rotation
- **Token-based Access**: No password-based authentication

### Authorization
- **Root Privileges**: Required for system monitoring access
- **File Permissions**: Secure configuration and data file permissions
- **Process Isolation**: Each daemon runs with appropriate privileges

### Data Protection
- **Local Storage**: All data stored locally in SQLite
- **Secure Defaults**: No remote data transmission by default
- **Configuration Security**: Secure handling of sensitive configuration

### systemd Security
```ini
[Service]
# Security restrictions
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=read-only
ReadWritePaths=/var/lib/sysmedic /var/log/sysmedic

# Resource limits
MemoryMax=256M
CPUQuota=50%
```

## Performance Characteristics

### Resource Usage
- **Memory Footprint**: Minimal memory usage (~50MB typical)
- **CPU Impact**: Configurable monitoring intervals (default 60s)
- **Disk I/O**: Efficient SQLite storage with batched writes
- **Network Usage**: Optional email alerts only

### Scalability
- **Concurrent Connections**: Support for multiple WebSocket clients
- **Data Volume**: Handles high-frequency monitoring data efficiently
- **Long-term Storage**: Automatic cleanup and maintenance
- **Process Limits**: Independent daemon resource limits

### Optimization Features
- **Configurable Intervals**: Adjust monitoring frequency based on needs
- **Smart Filtering**: Reduce noise from system processes
- **Efficient Queries**: Optimized database operations
- **Resource Limits**: Built-in resource usage controls

## File Locations

### Binary
- `/usr/local/bin/sysmedic` - Single binary (CLI + all daemon modes)

### Configuration
- `/etc/sysmedic/config.yaml` - Main configuration file (shared by all modes)

### Data Storage
- `/var/lib/sysmedic/` - Data directory
- `/var/lib/sysmedic/sysmedic.db` - SQLite database
- `/var/lib/sysmedic/sysmedic.doctor.pid` - Doctor daemon PID
- `/var/lib/sysmedic/sysmedic.websocket.pid` - WebSocket daemon PID

### Logging
- `/var/log/sysmedic/` - Log directory
- `/var/log/sysmedic/sysmedic.log` - Main application log
- `/var/log/sysmedic/websocket.log` - WebSocket server log

### SystemD Services
- `/etc/systemd/system/sysmedic.doctor.service` - Doctor daemon service
- `/etc/systemd/system/sysmedic.websocket.service` - WebSocket daemon service

### Runtime Files
- `/run/sysmedic/` - Runtime directory for temporary files
- `/tmp/sysmedic/` - Temporary files and sockets

## Usage Examples

### Managing the Doctor Daemon
```bash
# Start monitoring daemon
sysmedic daemon start

# Check daemon status
sysmedic daemon status
# Output: SysMedic Doctor daemon: running (PID: 1234)

# Stop monitoring daemon
sysmedic daemon stop
```

### Managing the WebSocket Server
```bash
# Start WebSocket server on default port (8060)
sysmedic websocket start

# Start WebSocket server on custom port
sysmedic websocket start 9090

# Check WebSocket status with Quick Connect URL
sysmedic websocket status
# Output includes: Quick Connect: sysmedic://[secret]@[ip]:[port]/

# Generate new authentication secret
sysmedic websocket new-secret

# Stop WebSocket server
sysmedic websocket stop
```

### SystemD Integration
```bash
# Enable both services
sudo systemctl enable sysmedic.doctor sysmedic.websocket

# Start both services
sudo systemctl start sysmedic.doctor sysmedic.websocket

# Check status
sudo systemctl status sysmedic.doctor
sudo systemctl status sysmedic.websocket

# View logs separately
sudo journalctl -u sysmedic.doctor -f
sudo journalctl -u sysmedic.websocket -f
```

### Configuration Management
```bash
# Show current configuration
sysmedic config show

# Update configuration settings
sysmedic config set monitoring.cpu_threshold 90
sysmedic config set websocket.port 8080

# Validate configuration
sysmedic config validate
```

## Real-World Usage Scenarios

### Multi-user Development Servers
- Monitor developer resource usage
- Identify resource-intensive development processes
- Set custom thresholds for different developer roles
- Email alerts for sustained high usage

### Database Servers
- Custom thresholds for database users
- Monitor query performance impact
- Track backup and maintenance job resources
- Alert on sustained database locks

### Web Servers
- Monitor application server resources
- Track user session impacts
- Identify problematic web processes
- Custom thresholds for web server users

### Critical Production Systems
- Strict monitoring with fast alerts
- Multiple alert channels (email, webhooks)
- Detailed user activity tracking
- Historical analysis for capacity planning

## Extensibility

The modular architecture supports easy extension:

### Custom Metrics
- Additional monitor implementations
- Custom data collection modules
- Plugin-based metric extensions

### Alert Channels
- Slack integration
- Discord notifications
- Webhook endpoints
- SMS alerts via APIs

### Storage Backends
- Alternative database engines
- Cloud storage integration
- Time-series database support

### Configuration Extensions
- Additional threshold types
- Custom user filtering rules
- Dynamic configuration sources

## Migration and Compatibility

### Version Compatibility
- Backward compatible configuration
- Automatic database schema migration
- Graceful handling of configuration changes

### Upgrade Process
- In-place binary replacement
- Configuration migration scripts
- Service restart coordination
- Data preservation during upgrades

### Legacy Support
- Support for older configuration formats
- Migration tools for previous versions
- Compatibility warnings and guidance

---

## Summary

SysMedic's single binary multi-daemon architecture provides a unique combination of deployment simplicity and operational flexibility. The system offers comprehensive monitoring capabilities with user-centric features that distinguish it from traditional monitoring tools.

**Key Architectural Benefits:**
- 🚀 **Simple Deployment**: Single 11MB binary for all functionality
- 🔧 **Process Independence**: Separate daemon processes for monitoring and WebSocket services
- 📊 **User-Centric**: Unique focus on per-user resource tracking and persistent user detection
- 🔒 **Security Focused**: Root privilege management with secure defaults
- ⚡ **High Performance**: Efficient monitoring with configurable impact
- 🎛️ **Highly Configurable**: Flexible thresholds, filtering, and alert settings
- 🌐 **Remote Access**: Real-time WebSocket API for remote monitoring
- 💾 **Data Persistence**: SQLite-based storage with automatic maintenance

This architecture makes SysMedic suitable for a wide range of environments, from small development servers to large multi-user production systems, while maintaining ease of deployment and management.