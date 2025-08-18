# 🚀 SysMedic v1.0.0 - Production Release

**Cross-platform Linux server monitoring CLI tool with user-centric resource tracking**

[![Release](https://img.shields.io/badge/Release-v1.0.0-blue.svg)](https://github.com/ahur-system/sysmedic/releases/tag/v1.0.0)
[![Go Version](https://img.shields.io/badge/Go-1.18+-00ADD8.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-24%2F24%20Passing-brightgreen.svg)](#testing)

## 🎉 What's New in v1.0.0

This is the **initial production release** of SysMedic, delivering a complete implementation of user-centric server monitoring with unique capabilities not found in traditional monitoring tools.

### ✨ Key Features

- **👥 User-Centric Monitoring**: Identifies which specific users consume system resources
- **⏱️ Persistent Detection**: Tracks sustained high usage (60+ minutes) vs temporary spikes  
- **📊 Real-time Dashboard**: Live system status with detailed user breakdown
- **🔧 Background Daemon**: Configurable monitoring intervals with systemd integration
- **💾 Embedded Storage**: SQLite database with automatic cleanup and retention
- **📧 Smart Alerting**: Email notifications with context and actionable recommendations
- **⚙️ Flexible Configuration**: YAML config with system-wide and per-user thresholds
- **🔒 Security Hardened**: systemd service with privilege restrictions

### 🎯 Perfect For

- **Multi-user servers** - Identify resource-heavy users quickly
- **Database servers** - Monitor DB users with custom thresholds
- **Web servers** - Track application users separately  
- **Development environments** - User accountability in shared systems
- **System administrators** - Rapid troubleshooting of performance issues

## 📦 Download & Installation

### Quick Install (Recommended)

```bash
# Download latest release
wget https://github.com/ahur-system/sysmedic/releases/download/v1.0.0/sysmedic-v1.0.0-linux-amd64.tar.gz

# Extract and install
tar -xzf sysmedic-v1.0.0-linux-amd64.tar.gz
cd sysmedic-v1.0.0
sudo ./scripts/install.sh

# Enable and start service
sudo systemctl enable sysmedic
sudo systemctl start sysmedic

# View dashboard
sysmedic
```

### Manual Installation

```bash
# Extract package
tar -xzf sysmedic-v1.0.0-linux-amd64.tar.gz
cd sysmedic-v1.0.0

# Copy binary
sudo cp sysmedic /usr/local/bin/
sudo chmod +x /usr/local/bin/sysmedic

# Create directories
sudo mkdir -p /etc/sysmedic /var/lib/sysmedic

# Install systemd service
sudo cp scripts/sysmedic.service /etc/systemd/system/
sudo systemctl daemon-reload
```

### Package Managers

**Ubuntu/Debian:**
```bash
wget https://github.com/ahur-system/sysmedic/releases/download/v1.0.0/sysmedic-v1.0.0-amd64.deb
sudo dpkg -i sysmedic-v1.0.0-amd64.deb
```

**RHEL/CentOS:**
```bash
wget https://github.com/ahur-system/sysmedic/releases/download/v1.0.0/sysmedic-v1.0.0-x86_64.rpm
sudo rpm -i sysmedic-v1.0.0-x86_64.rpm
```

## 🚀 Quick Start

### 1. View System Dashboard
```bash
sysmedic
```

**Example Output:**
```
System: Ubuntu 22.04.5 LTS
Status: Medium Usage
Daemon: Running

System Metrics:
- CPU: 45.2% (threshold: 80%)
- Memory: 67.1% (threshold: 80%)
- Network: 12.5 MB/s
- Load Average: 1.45, 1.89, 1.23

Top Resource Users (Last Hour):
- database_user: CPU 85.2%, Memory 45.1%, Processes: 12 (⚠️ High CPU 65min)
- web_server: CPU 35.4%, Memory 25.8%, Processes: 8
- backup_job: CPU 15.2%, Memory 60.3%, Processes: 3
```

### 2. Start Background Monitoring
```bash
sudo systemctl start sysmedic
sudo systemctl enable sysmedic  # Auto-start on boot
```

### 3. Configure Thresholds
```bash
# View current configuration
sysmedic config show

# Update system-wide threshold
sysmedic config set cpu-threshold 85

# Set user-specific threshold
sysmedic config set-user database_user cpu-threshold 90
```

### 4. View Reports
```bash
# Recent alerts and system summary
sysmedic reports

# Top resource users
sysmedic reports users --top 10

# Specific user activity
sysmedic reports users --user database_user
```

## ⚙️ Configuration

Default configuration is created at `/etc/sysmedic/config.yaml`:

```yaml
monitoring:
  check_interval: 60          # Monitor every 60 seconds
  cpu_threshold: 80           # System CPU threshold (%)
  memory_threshold: 80        # System memory threshold (%)
  persistent_time: 60         # Minutes before flagging persistent usage

users:
  cpu_threshold: 80           # Default per-user CPU threshold (%)
  memory_threshold: 80        # Default per-user memory threshold (%)
  persistent_time: 60         # Minutes before flagging user as persistent

user_thresholds:              # Custom per-user limits
  database_user:
    cpu_threshold: 90         # Allow higher CPU for DB user
    memory_threshold: 70      # But stricter memory limits
```

### Email Alerts Setup

```yaml
email:
  enabled: true
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "alerts@company.com"
  password: "app_password"     # Use app password for Gmail
  from: "sysmedic@company.com"
  to: "admin@company.com"
  tls: true
```

## 📊 Technical Specifications

| Specification | Details |
|---------------|---------|
| **Language** | Go 1.18+ |
| **Binary Size** | 7.4MB (static, zero dependencies) |
| **Database** | SQLite (embedded) |
| **Memory Usage** | ~10MB typical |
| **CPU Impact** | Minimal (configurable intervals) |
| **Platforms** | Linux AMD64 |
| **Dependencies** | None (single binary) |

## 🔒 Security Features

- **Privilege Separation**: Runs as root only for system monitoring
- **systemd Hardening**: Security restrictions in service file
- **Data Protection**: Local SQLite storage with secure defaults
- **Configuration Security**: No hardcoded credentials
- **Process Isolation**: Protected from other system processes

## 📈 Performance & Monitoring

### Default Behavior
- **Monitoring Interval**: 60 seconds (configurable)
- **Data Retention**: 30 days (configurable)
- **Automatic Cleanup**: Daily maintenance of old data
- **Alert Thresholds**: 80% CPU/Memory (configurable per-user)

### Status Classification
- **Light Usage**: System < 60% AND no persistent user issues
- **Medium Usage**: System 60-80% OR temporary user spikes
- **Heavy Load**: System > 80% OR persistent user issues (60+ minutes)

## 🛠️ Development & Building

### Prerequisites
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go build-essential sqlite3 libsqlite3-dev

# RHEL/CentOS
sudo yum install golang gcc sqlite-devel
```

### Build from Source
```bash
git clone https://github.com/ahur-system/sysmedic.git
cd sysmedic
make build
```

### Run Tests
```bash
make test          # All tests (24/24 passing)
make test-coverage # With coverage report
make benchmark     # Performance benchmarks
```

## 🔧 Troubleshooting

### Common Issues

**Daemon won't start:**
```bash
# Check logs
sudo journalctl -u sysmedic -f

# Check permissions
sudo ls -la /var/lib/sysmedic/
sudo ls -la /etc/sysmedic/

# Manual start for debugging
sudo /usr/local/bin/sysmedic daemon start
```

**High resource usage by SysMedic itself:**
```bash
# Check monitoring frequency
sysmedic config show

# Reduce frequency if needed
sysmedic config set check-interval 120  # 2 minutes instead of 1
```

**Email alerts not working:**
```bash
# Verify configuration
sysmedic config show

# Test SMTP connectivity
telnet smtp.gmail.com 587
```

### Log Locations
- **systemd logs**: `journalctl -u sysmedic`
- **Application logs**: Integrated with system journal
- **Database**: `/var/lib/sysmedic/sysmedic.db`
- **Configuration**: `/etc/sysmedic/config.yaml`

## 🗑️ Uninstallation & Cleanup

### Complete Removal

**1. Stop and disable service:**
```bash
sudo systemctl stop sysmedic
sudo systemctl disable sysmedic
```

**2. Remove binary and service:**
```bash
sudo rm -f /usr/local/bin/sysmedic
sudo rm -f /etc/systemd/system/sysmedic.service
sudo systemctl daemon-reload
```

**3. Remove configuration (optional):**
```bash
sudo rm -rf /etc/sysmedic/
```

**4. Remove data and logs (optional):**
```bash
# Remove database and cached data
sudo rm -rf /var/lib/sysmedic/

# Remove PID file if exists
sudo rm -f /var/run/sysmedic.pid
```

**5. Clean package manager (if installed via package):**
```bash
# Ubuntu/Debian
sudo apt remove sysmedic
sudo apt autoremove

# RHEL/CentOS  
sudo yum remove sysmedic
```

### Selective Cleanup

**Keep configuration, remove only data:**
```bash
sudo systemctl stop sysmedic
sudo rm -rf /var/lib/sysmedic/
sudo systemctl start sysmedic  # Will recreate with existing config
```

**Reset to defaults:**
```bash
sudo systemctl stop sysmedic
sudo rm -f /etc/sysmedic/config.yaml
sudo systemctl start sysmedic  # Will create default config
```

**Clear old alerts and data:**
```bash
# SysMedic has built-in cleanup, but manual cleanup:
sudo systemctl stop sysmedic
sudo rm -f /var/lib/sysmedic/sysmedic.db
sudo systemctl start sysmedic  # Will recreate database
```

### Cache and Temporary Files

SysMedic creates minimal temporary files:
```bash
# Check for any remaining files
sudo find /tmp -name "*sysmedic*" -delete
sudo find /var/tmp -name "*sysmedic*" -delete

# Check system logs retention (handled by systemd)
sudo journalctl --vacuum-time=1d  # Keep only 1 day of logs
```

## 🤝 Support & Community

### Documentation
- **README**: [Complete setup guide](README.md)
- **Implementation**: [Technical details](IMPLEMENTATION.md)
- **Configuration**: [Example configs](scripts/config.example.yaml)

### Getting Help
- **Issues**: [GitHub Issues](https://github.com/ahur-system/sysmedic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ahur-system/sysmedic/discussions)
- **Wiki**: [GitHub Wiki](https://github.com/ahur-system/sysmedic/wiki)

### Contributing
1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit changes: `git commit -m 'Add amazing feature'`
4. Push branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

## 📜 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- Built with Go and the excellent ecosystem of Go libraries
- SQLite for reliable embedded database storage
- systemd for robust service management
- The Linux /proc filesystem for system monitoring capabilities

---

**SysMedic v1.0.0** - Made with ❤️ for system administrators who need to identify resource-hungry users quickly and efficiently.

**Download**: [Latest Release](https://github.com/ahur-system/sysmedic/releases/latest) | **Source**: [GitHub Repository](https://github.com/ahur-system/sysmedic)