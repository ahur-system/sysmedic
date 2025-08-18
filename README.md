# SysMedic

**Cross-platform Linux server monitoring CLI tool with daemon capabilities for tracking system and user resource usage spikes.**

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)](https://github.com/sysmedic/sysmedic)

## Features

- **Real-time System Monitoring**: Track CPU, Memory, Network usage, and Load averages
- **User-Centric Monitoring**: Identify which users are consuming system resources
- **Persistent Usage Detection**: Flag users with sustained high resource usage (60+ minutes)
- **Smart Alerting**: Email notifications with detailed user breakdowns and recommendations
- **Daemon Mode**: Background monitoring with configurable intervals
- **SQLite Storage**: Embedded database with configurable retention
- **Zero Dependencies**: Single static binary with no external requirements
- **systemd Integration**: Native Linux service integration

## Quick Start

### Installation

**Download Binary:**
```bash
# Download latest release
wget https://github.com/sysmedic/sysmedic/releases/latest/download/sysmedic-linux-amd64.tar.gz
tar -xzf sysmedic-linux-amd64.tar.gz
sudo mv sysmedic /usr/local/bin/
sudo chmod +x /usr/local/bin/sysmedic
```

**From Source:**
```bash
git clone https://github.com/sysmedic/sysmedic.git
cd sysmedic
make build
sudo make install
```

### Basic Usage

**View Dashboard:**
```bash
sysmedic
```

**Start Daemon:**
```bash
sudo sysmedic daemon start
sudo systemctl enable sysmedic  # Auto-start on boot
```

**View Reports:**
```bash
sysmedic reports
sysmedic reports users --top 10
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
```bash
sysmedic daemon start                 # Start monitoring daemon
sysmedic daemon stop                  # Stop daemon  
sysmedic daemon status                # Check if running
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
git clone https://github.com/sysmedic/sysmedic.git
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
wget https://github.com/sysmedic/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### RHEL/CentOS (.rpm)
```bash
wget https://github.com/sysmedic/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -i sysmedic-x86_64.rpm
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### Generic Linux (tar.gz)
```bash
wget https://github.com/sysmedic/sysmedic/releases/latest/download/sysmedic-linux-amd64.tar.gz
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
sudo journalctl -u sysmedic -f

# Check permissions
sudo ls -la /var/lib/sysmedic/
sudo ls -la /etc/sysmedic/

# Manual start
sudo /usr/local/bin/sysmedic daemon start
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

- **Documentation**: [GitHub Wiki](https://github.com/sysmedic/sysmedic/wiki)
- **Issues**: [GitHub Issues](https://github.com/sysmedic/sysmedic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/sysmedic/sysmedic/discussions)

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