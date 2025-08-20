# SysMedic Documentation

This directory contains comprehensive documentation for SysMedic, organized by topic and functionality.

## 🎯 Single Binary Multi-Daemon Architecture

SysMedic uses a revolutionary single binary that can operate in multiple independent daemon modes, providing the best of both worlds: deployment simplicity and process modularity.

### Architecture Overview

#### Single Binary Design
- **One Binary**: `/usr/local/bin/sysmedic` (11MB) handles all functionality
- **Multiple Modes**: CLI tool, Doctor daemon, and WebSocket daemon
- **Process Independence**: Each daemon runs as a separate process with its own PID
- **Shared Configuration**: All modes use the same configuration file

#### Daemon Modes

##### Doctor Daemon Mode (`sysmedic --doctor-daemon`)
- **Purpose**: Core system and user monitoring functionality
- **Responsibilities**:
  - Monitor system metrics (CPU, memory, network, load averages)
  - Track user resource usage with smart filtering
  - Generate alerts based on configurable thresholds
  - Store monitoring data in SQLite database
  - Manage persistent user tracking
- **PID File**: `/var/lib/sysmedic/sysmedic.doctor.pid`
- **Service**: `sysmedic.doctor.service`

##### WebSocket Daemon Mode (`sysmedic --websocket-daemon`)
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

##### CLI Mode (default)
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

### Usage Examples

#### Managing the Doctor Daemon
```bash
# Start monitoring daemon
sysmedic daemon start

# Check daemon status
sysmedic daemon status

# Stop monitoring daemon
sysmedic daemon stop
```

#### Managing the WebSocket Server
```bash
# Start WebSocket server on default port (8060)
sysmedic websocket start

# Start WebSocket server on custom port
sysmedic websocket start 9090

# Check WebSocket status
sysmedic websocket status

# Generate new authentication secret
sysmedic websocket new-secret

# Stop WebSocket server
sysmedic websocket stop
```

#### Combined Status Check
```bash
sysmedic daemon status
# Output:
# SysMedic Doctor daemon: running (PID: 1234)
# SysMedic WebSocket daemon: running (PID: 5678)
```

### SystemD Integration

#### Service Files
- `sysmedic.doctor.service` - Monitoring daemon service
- `sysmedic.websocket.service` - WebSocket daemon service

#### Installation and Management
```bash
# Enable both services
sudo systemctl enable sysmedic.doctor sysmedic.websocket

# Start both services
sudo systemctl start sysmedic.doctor sysmedic.websocket

# Individual service control
sudo systemctl status sysmedic.doctor
sudo systemctl status sysmedic.websocket

# View logs separately
sudo journalctl -u sysmedic.doctor -f
sudo journalctl -u sysmedic.websocket -f
```

### File Locations

#### Binary
- `/usr/local/bin/sysmedic` - Single binary (CLI + all daemon modes)

#### Configuration
- `/etc/sysmedic/config.yaml` - Main configuration file (shared by all modes)

#### Data
- `/var/lib/sysmedic/` - Data directory
- `/var/lib/sysmedic/sysmedic.db` - SQLite database
- `/var/lib/sysmedic/sysmedic.doctor.pid` - Doctor daemon PID
- `/var/lib/sysmedic/sysmedic.websocket.pid` - WebSocket daemon PID

#### Services
- `/etc/systemd/system/sysmedic.doctor.service`
- `/etc/systemd/system/sysmedic.websocket.service`

### Configuration

#### Shared Configuration
Both daemons use the same configuration file but access different sections:

```yaml
monitoring:
  check_interval: 30
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

users:
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

websocket:
  enabled: true
  port: 8060
  secret: "generated-secret-token"

data_path: "/var/lib/sysmedic"

user_filtering:
  min_uid_for_real_users: 1000
  ignore_system_users: true
  min_cpu_percent: 5.0
  min_memory_percent: 5.0
  excluded_users: ["root", "daemon", "www-data"]
```

#### WebSocket Authentication
- Uses secret-based authentication via `X-SysMedic-Secret` header
- Secrets can be regenerated using `sysmedic websocket new-secret`
- Automatic restart of WebSocket daemon when secret changes

### Migration from Previous Versions

#### Automatic Migration
When upgrading from previous versions:

1. Old daemon processes will be stopped automatically
2. New single binary will be installed
3. Configuration will be automatically migrated
4. WebSocket settings will be preserved
5. Service files will use the new binary with daemon flags

#### Manual Steps (if needed)
```bash
# Stop old services
sudo systemctl stop sysmedic.doctor sysmedic.websocket

# Install new version (via package manager or manual)
sudo dpkg -i sysmedic_1.0.5_amd64.deb  # or
sudo rpm -U sysmedic-1.0.5-1.x86_64.rpm

# Enable and start new services
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket
```

### Troubleshooting

#### Common Issues

**Daemon Won't Start**
- Check configuration: `sysmedic config show`
- Verify permissions: `ls -la /var/lib/sysmedic/`
- Check logs: `sudo journalctl -u sysmedic.doctor -n 50`

**WebSocket Connection Failed**
- Verify daemon is running: `sysmedic websocket status`
- Check port availability: `netstat -tlnp | grep 8060`
- Verify secret: Check configuration or regenerate with `sysmedic websocket new-secret`

**Permission Denied**
- Ensure data directory has correct permissions: `sudo chown -R root:root /var/lib/sysmedic`
- Check service file permissions and paths

#### Debug Mode
Run daemon modes in foreground for debugging:
```bash
# Run doctor daemon in foreground
sudo sysmedic --doctor-daemon

# Run WebSocket daemon in foreground
sudo sysmedic --websocket-daemon
```

## 📚 Documentation Index

### Core Documentation
- **[ARCHITECTURE_AND_FEATURES.md](ARCHITECTURE_AND_FEATURES.md)** - Complete system architecture guide
  - Single binary multi-daemon architecture
  - Core components and features overview
  - System monitoring logic and algorithms
  - Data storage and alert management
  - Configuration system and security model

- **[WEBSOCKET_API_GUIDE.md](WEBSOCKET_API_GUIDE.md)** - Complete WebSocket remote monitoring guide
  - WebSocket server management and configuration
  - Real-time API with message format specification
  - Request-response protocol documentation
  - Client examples (JavaScript, Python, Node.js)
  - Connection management and troubleshooting

- **[DEPLOYMENT_INSTALLATION.md](DEPLOYMENT_INSTALLATION.md)** - Deployment and installation guide
  - Build environment setup and package creation
  - GitHub publishing and release management
  - Multiple installation methods (DEB, RPM, generic Linux)
  - Automated deployment and distribution channels
  - Package verification and troubleshooting

- **[CONFIGURATION_TROUBLESHOOTING.md](CONFIGURATION_TROUBLESHOOTING.md)** - Configuration and troubleshooting guide
  - Complete configuration reference and management
  - WebSocket, email, and user filtering configuration
  - Comprehensive troubleshooting procedures
  - Common issues and solutions
  - Performance optimization and security hardening

## 📖 Documentation Overview

### Architecture Documentation
The **ARCHITECTURE_AND_FEATURES.md** file provides comprehensive system design coverage:
- Revolutionary single binary multi-daemon architecture
- Complete feature set including user-centric monitoring
- System monitoring algorithms and data flow
- Security model and performance characteristics
- File locations and usage examples

### WebSocket API Documentation
The **WEBSOCKET_API_GUIDE.md** file covers everything needed for remote monitoring:
- Quick start guide and server management
- Complete WebSocket API with real-time message types
- Request-response protocol for bidirectional communication
- Multiple client examples for different programming languages
- Connection management and comprehensive troubleshooting

### Deployment Documentation
The **DEPLOYMENT_INSTALLATION.md** file covers the complete deployment process:
- Build environment setup and automated package creation
- GitHub repository setup, authentication, and release management
- Multiple installation methods for different Linux distributions
- Automated deployment with GitHub Actions and Docker
- Package verification and comprehensive troubleshooting

### Configuration Documentation
The **CONFIGURATION_TROUBLESHOOTING.md** file provides complete operational guidance:
- Comprehensive configuration reference with all options
- WebSocket, email, and user filtering configuration details
- Diagnostic procedures and troubleshooting methodologies
- Common issues with step-by-step solutions
- Performance optimization and security hardening recommendations

## 🗂️ Related Documentation

### Root Directory
- **[README.md](../README.md)** - Main project documentation and quick start
- **[CHANGELOG.md](../CHANGELOG.md)** - Version history and release notes
- **[LICENSE](../LICENSE)** - MIT License terms

### Scripts Directory
- **[scripts/](../scripts/)** - Installation, configuration, and testing scripts
  - `install.sh` - Automated installation script
  - `config.example.yaml` - Example configuration file
  - Test scripts for WebSocket, alerts, and package verification

### Examples Directory
- **[examples/](../examples/)** - Working client examples
  - WebSocket clients in multiple languages
  - Configuration examples
  - Integration samples

## 🚀 Quick Navigation

| Topic | File | Description |
|-------|------|-------------|
| **Architecture & Features** | [ARCHITECTURE_AND_FEATURES.md](ARCHITECTURE_AND_FEATURES.md) | System design, components, features |
| **WebSocket API** | [WEBSOCKET_API_GUIDE.md](WEBSOCKET_API_GUIDE.md) | Remote monitoring, API, clients |
| **Deployment & Installation** | [DEPLOYMENT_INSTALLATION.md](DEPLOYMENT_INSTALLATION.md) | Building, publishing, installation |
| **Configuration & Troubleshooting** | [CONFIGURATION_TROUBLESHOOTING.md](CONFIGURATION_TROUBLESHOOTING.md) | Settings, diagnostics, solutions |

## 📋 Documentation Standards

### Format
- All documentation is written in GitHub Flavored Markdown
- Each file includes a comprehensive table of contents for easy navigation
- Code examples include syntax highlighting and working examples
- Comprehensive coverage with practical examples and troubleshooting

### Organization
- **Consolidated**: Related topics are merged into comprehensive guides
- **Self-contained**: Each document covers its topic completely
- **Cross-referenced**: Documents link to related sections appropriately
- **Practical**: Focus on real-world usage and problem-solving
- **Up-to-date**: Documentation reflects the latest features and fixes

### Maintenance
- Documentation is consolidated to eliminate redundancy
- All examples are tested and verified to work correctly
- Troubleshooting sections cover real issues and solutions
- Regular updates incorporate new features and community feedback

## 🤝 Contributing to Documentation

### Guidelines
- Follow existing formatting and style
- Include working code examples
- Test all instructions before submitting
- Update the index when adding new files

### Process
1. Create or update documentation files
2. Test all examples and instructions
3. Update this README.md index if needed
4. Submit a pull request with clear description

## 📞 Support

For questions about documentation:
- **Issues**: [GitHub Issues](https://github.com/ahur-system/sysmedic/issues)
- **Discussions**: [GitHub Discussions](https://github.com/ahur-system/sysmedic/discussions)
- **Wiki**: [GitHub Wiki](https://github.com/ahur-system/sysmedic/wiki)

---

**Note**: This documentation covers SysMedic v1.0.x and later. For older versions, refer to the git history or archived documentation.