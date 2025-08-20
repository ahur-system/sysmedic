# SysMedic

**Single binary multi-daemon Linux system monitoring tool with independent doctor and WebSocket daemon processes for comprehensive system and user resource tracking.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)](https://github.com/sysmedic/sysmedic)

## Features

- **Single Binary Multi-Daemon Architecture**: One 11MB binary with independent doctor and WebSocket daemon processes
- **Real-time System Monitoring**: Track CPU, Memory, Network usage, and Load averages
- **User-Centric Monitoring**: Smart user filtering focusing on real problematic users
- **Persistent Usage Detection**: Flag users with sustained high resource usage (60+ minutes)
- **Remote Access**: WebSocket server for real-time remote monitoring
- **Smart Alerting**: Email notifications with detailed user breakdowns and recommendations
- **Alert Management**: View, filter, and resolve system alerts with CLI commands
- **Independent Daemon Processes**: Doctor (monitoring) and WebSocket (remote access) run separately
- **SQLite Storage**: Embedded database with configurable retention
- **Zero Dependencies**: Single static binary with no external requirements
- **Dual SystemD Integration**: Separate services for monitoring and WebSocket functionality

## Quick Start

### Installation

**Download Binary:**
```bash
# Download latest release
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-linux-amd64.tar.gz
tar -xzf sysmedic-linux-amd64.tar.gz
sudo mv sysmedic /usr/local/bin/
sudo chmod +x /usr/local/bin/sysmedic
```

**From Source:**
```bash
git clone https://github.com/ahur-system/sysmedic.git
cd sysmedic
make build
sudo make install
```

### Basic Usage

**View Dashboard:**
```bash
sysmedic
```

**Start Daemons:**
```bash
# Start both daemon processes
sudo sysmedic daemon start      # Doctor daemon (monitoring)
sudo sysmedic websocket start   # WebSocket daemon (remote access)

# Enable auto-start on boot
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket
```

**View Reports:**
```bash
sysmedic reports
sysmedic reports users --top 10
```

**Remote Access:**
```bash
sysmedic websocket start             # Start WebSocket daemon
sysmedic websocket status            # Get quick connect URL
# Output: Quick Connect: sysmedic://[secret]@[publicip]:[port]/
```

## Dashboard Output

```
System: Ubuntu 20.04.3 LTS
Status: Heavy Load
Daemon: Running

System Metrics:
- CPU: 89.2% (threshold: 80%)
- Memory: 67.1% (threshold: 80%)
- Network: 12.5 MB/s
- Load Average: 2.45, 1.89, 1.23

Top Resource Users (Last Hour):
- database_user: CPU 85.2%, Memory 45.1%, Processes: 12 (⚠️ High CPU 65min)
- web_server: CPU 35.4%, Memory 25.8%, Processes: 8
- backup_job: CPU 15.2%, Memory 60.3%, Processes: 3

⚠️ 2 unresolved alert(s) in the last 24 hours. Run 'sysmedic reports' for details.

Remote Access:
Quick Connect: sysmedic://d8852f78260f16d31eeff80ca6158848@192.168.1.100:8060/
```

## Configuration

**Default Configuration** (`/etc/sysmedic/config.yaml`):

```yaml
monitoring:
  check_interval: 60          # seconds
  cpu_threshold: 80           # system-wide %
  memory_threshold: 80        # system-wide %
  persistent_time: 60         # minutes for persistent detection

users:
  cpu_threshold: 80           # default per-user %
  memory_threshold: 80        # default per-user %
  persistent_time: 60         # minutes before flagging user

reporting:
  period: "hourly"            # hourly/daily/weekly
  retain_days: 30

email:
  enabled: false
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  username: "alerts@company.com"
  password: "app_password"
  from: "sysmedic@company.com"
  to: "admin@company.com"
  tls: true

user_thresholds:              # custom per-user limits
  database_user:
    cpu_threshold: 90
    memory_threshold: 70
  high_priority_user:
    cpu_threshold: 95
    persistent_time: 30
```

## Commands

### Dashboard
```bash
sysmedic                              # Show current status + top users
```

### Daemon Management
### Daemon Commands
```bash
# Doctor daemon (monitoring)
sysmedic daemon start                 # Start doctor daemon
sysmedic daemon stop                  # Stop doctor daemon  
sysmedic daemon status                # Check daemon status

# WebSocket daemon (remote access)
sysmedic websocket start              # Start WebSocket daemon
sysmedic websocket stop               # Stop WebSocket daemon
sysmedic websocket status             # Check status & get quick connect URL
sysmedic websocket new-secret         # Generate new auth secret
```

### Configuration
```bash
sysmedic config show                  # Display current config
sysmedic config set cpu-threshold 85  # Update system threshold
sysmedic config set-user john cpu-threshold 90  # Set user-specific limit
```

### Reports
```bash
sysmedic reports                      # Recent system alerts
sysmedic reports users                # Detailed user activity
sysmedic reports users --top 10       # Top 10 resource consumers
sysmedic reports --period daily       # Custom time period
sysmedic reports users --user database_user # Specific user history
```

### Alert Management
```bash
sysmedic alerts                       # Show alert summary and overview
sysmedic alerts list                  # List all alerts (24h default)
sysmedic alerts list -u               # Show only unresolved alerts
sysmedic alerts list -p 7d            # Show alerts from last 7 days
sysmedic alerts resolve 123           # Resolve specific alert by ID
sysmedic alerts resolve-all           # Resolve all unresolved alerts (with confirmation)
```

## Email Alerts

When enabled, SysMedic sends detailed email notifications:

**Sample Alert:**
```
Subject: SysMedic Alert - Heavy Load Detected on prod-web-01

Server: prod-web-01
Timestamp: 2024-01-15 14:30:45
Duration: 75 minutes
Severity: Heavy

System Status:
- CPU: 89% (threshold: 80%)
- Memory: 67%
- Network: 45MB/s
- Load Average: 2.45, 1.89, 1.23

Primary Cause: database_user (CPU: 95% for 75min)

User Breakdown:
- database_user: CPU 95%, Memory 45% (⚠️ PERSISTENT)
- web_server: CPU 25%, Memory 20%
- backup_job: CPU 12%, Memory 15%

Persistent Issues:
- database_user: cpu usage 95.2% for 1h 15m

Recommendations:
- Investigate database_user processes for user database_user
- Check for CPU-intensive processes
- Review database_user's processes - high CPU usage detected
```

## Status Logic

```
Light Usage:  System < 60% AND no persistent user issues
Medium Usage: System 60-80% OR 1-2 users with temporary spikes  
Heavy Load:   System > 80% OR any user persistent (60+ min) above threshold
```

## Data Storage

- **Location**: `/var/lib/sysmedic/sysmedic.db` (SQLite)
- **Structure**: System metrics, user activity, alerts, persistent user records
- **Retention**: Configurable (default 30 days)
- **Cleanup**: Automatic daily maintenance

## Building from Source

**Prerequisites:**
- Go 1.21+
- SQLite3 development libraries

**Build:**
```bash
git clone https://github.com/ahur-system/sysmedic.git
cd sysmedic
make deps
make build
```

**Development:**
```bash
make dev          # Run in development mode
make test         # Run tests
make lint         # Run linter
make benchmark    # Run benchmarks
```

**Packaging:**
```bash
make package      # Create tar.gz packages
make package-deb  # Create DEB package
make package-rpm  # Create RPM package
make release      # Full release build
```

## Installation Methods

### Ubuntu/Debian (.deb)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### RHEL/CentOS (.rpm)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -i sysmedic-x86_64.rpm
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### Generic Linux (tar.gz)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-linux-amd64.tar.gz
tar -xzf sysmedic-linux-amd64.tar.gz
cd sysmedic-*
sudo ./scripts/install.sh
```

## Docker Support

```bash
# Build image
make docker-build

# Run with host monitoring
docker run --rm -it --privileged --pid=host --net=host \
  -v /proc:/host/proc:ro \
  -v /sys:/host/sys:ro \
  sysmedic:latest
```

## Architecture

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Interface │    │  Background      │    │   Storage       │
│                 │    │  Daemon          │    │                 │
│ • Dashboard     │◄──►│                  │◄──►│ • SQLite DB     │
│ • Config Mgmt   │    │ • System Monitor │    │ • Metrics       │
│ • Reports       │    │ • User Tracking  │    │ • Alerts        │
│ • Daemon Ctrl   │    │ • Alert Engine   │    │ • User Activity │
└─────────────────┘    └──────────────────┘    └─────────────────┘
                                │
                                ▼
                       ┌─────────────────┐
                       │  Email Alerts   │
                       │                 │
                       │ • SMTP/TLS      │
                       │ • Contextual    │
                       │ • Actionable    │
                       └─────────────────┘
```

## Key Differentiators

- **User-Centric**: Identifies which users cause system issues
- **Persistent Detection**: Distinguishes temporary vs sustained problems  
- **Actionable Alerts**: Reports include specific user responsible for spikes
- **Zero Dependencies**: Single binary installation
- **Flexible Configuration**: System-wide and per-user thresholds
- **Production Ready**: Battle-tested monitoring logic

## Troubleshooting

**Daemon won't start:**
```bash
# Check logs
sudo journalctl -u sysmedic.doctor -f     # Doctor daemon
sudo journalctl -u sysmedic.websocket -f  # WebSocket daemon

# Check permissions
sudo ls -la /var/lib/sysmedic/
sudo ls -la /etc/sysmedic/

# Manual start (foreground for debugging)
sudo sysmedic --doctor-daemon
sudo sysmedic --websocket-daemon

# Or use CLI commands
sudo sysmedic daemon start
sudo sysmedic websocket start
```

**Email alerts not working:**
```bash
# Test configuration
sysmedic config show

# Check SMTP connectivity
telnet smtp.gmail.com 587
```

**High resource usage:**
```bash
# Check SysMedic itself
ps aux | grep sysmedic
top -p $(pgrep sysmedic)

# Reduce monitoring frequency
sysmedic config set check-interval 120  # 2 minutes
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Support

- **Documentation**: [GitHub Wiki](https://github.com/ahur-system/sysmedic/wiki)
- **Issues**: [GitHub Issues](https://github.com/ahur-system/sysmedic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ahur-system/sysmedic/discussions)

## Roadmap

- [ ] Web dashboard interface
- [ ] Slack/Discord integration
- [ ] Container monitoring support
- [ ] Multi-server monitoring
- [ ] Custom metric plugins
- [ ] Grafana integration
- [ ] Windows support

---

**Made with ❤️ for system administrators who need to identify resource-hungry users quickly and efficiently.**