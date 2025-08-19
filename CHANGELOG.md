# Changelog

All notable changes to SysMedic will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [1.0.5] - 2025-01-19

### Added
- Comprehensive documentation reorganization
- Created dedicated `docs/` directory with organized documentation
- Documentation index (`docs/README.md`) for easy navigation
- Combined WebSocket documentation into single comprehensive guide
- Combined deployment documentation into complete deployment guide
- Proper CHANGELOG.md following semantic versioning standards

### Changed
- **Breaking**: Moved all documentation from root to `docs/` directory
- Consolidated 4 WebSocket documentation files into single `docs/WEBSOCKET.md`
- Consolidated 3 deployment documentation files into single `docs/DEPLOYMENT.md`
- Consolidated 2 release files into proper `CHANGELOG.md`
- Moved all test scripts from root to `scripts/` directory
- Updated version references from 1.0.4 to 1.0.5

### Removed
- 11 scattered markdown files from root directory (moved to `docs/`)
- 5 test scripts from root directory (moved to `scripts/`)
- Redundant and fragmented documentation files

### Documentation
- `docs/WEBSOCKET.md`: Complete WebSocket guide (setup, API, clients, testing)
- `docs/DEPLOYMENT.md`: Complete deployment guide (building, publishing, installation)
- `docs/README.md`: Documentation index and navigation
- Root directory now contains only essential files (README, CHANGELOG, LICENSE)

### Project Organization
- 73% reduction in root directory markdown files (11 → 3)
- 100% cleanup of root directory scripts (5 → 0)
- All documentation now properly organized in `docs/` directory
- All scripts consolidated in `scripts/` directory
- Professional project structure following open-source best practices

## [1.0.4] - 2025-01-19

### Added
- WebSocket remote monitoring functionality
- Bidirectional communication support for WebSocket API
- Real-time system metrics streaming
- Secure authentication with generated secrets

## [1.0.3] - 2025-08-18

### Added
- Complete Alert Management system
- `sysmedic alerts` command for alert overview
- `sysmedic alerts list` with filtering options (`--unresolved`, `--period`)
- `sysmedic alerts resolve <id>` for individual alert resolution
- `sysmedic alerts resolve-all` for bulk alert resolution with confirmation
- Enhanced error handling for invalid/non-existent alert IDs
- New storage methods: `ResolveAllAlerts()` and `GetAlertByID()`
- Alert status tracking with `resolved_at` timestamps

### Fixed
- Dashboard alert count now updates after resolution
- Improved alert message truncation in list view
- Enhanced error messages for better user experience

### Technical
- Added 6 new CLI and storage functions
- 609 lines of code added with comprehensive error handling
- 19 test scenarios in comprehensive test suite
- Full backward compatibility maintained

## [1.0.0] - 2025-08-01

### Added
- Initial production release
- User-centric monitoring with specific user resource tracking
- Persistent detection of sustained high usage (60+ minutes)
- Real-time dashboard with detailed user breakdown
- Background daemon with configurable monitoring intervals
- systemd integration for service management
- Embedded SQLite storage with automatic cleanup
- Smart email alerting with actionable recommendations
- Flexible YAML configuration with per-user thresholds
- Security hardening with privilege restrictions

### Features
- **System Monitoring**: CPU, Memory, Disk, Network monitoring
- **User Tracking**: Individual user resource consumption
- **Alert System**: Email notifications with context
- **Dashboard**: Live system status display
- **Configuration**: YAML-based config with user-specific thresholds
- **Database**: SQLite with 30-day retention
- **Security**: systemd hardening and privilege separation

### Installation Options
- DEB packages for Debian/Ubuntu
- RPM packages for RHEL/CentOS
- Generic Linux binary
- Automated install script

### Technical Specifications
- Go 1.18+ support
- 7.4MB static binary with zero dependencies
- ~10MB typical memory usage
- Minimal CPU impact with configurable intervals
- Linux AMD64 platform support

### Commands
- `sysmedic` - View system dashboard
- `sysmedic daemon start/stop/status` - Daemon management
- `sysmedic config show/set` - Configuration management
- `sysmedic reports` - System reports and user activity

### Status Classifications
- **Light Usage**: System < 60% with no persistent user issues
- **Medium Usage**: System 60-80% or temporary user spikes
- **Heavy Load**: System > 80% or persistent user issues (60+ minutes)

### Configuration
- Default monitoring interval: 60 seconds
- Default thresholds: 80% CPU/Memory
- Configurable per-user thresholds
- Email SMTP configuration support
- Data retention: 30 days (configurable)

---

## Version History Summary

- **v1.0.0**: Initial production release with core monitoring functionality
- **v1.0.3**: Added comprehensive alert management system
- **v1.0.4**: WebSocket remote monitoring and real-time API
- **v1.0.5**: Documentation reorganization and project cleanup

## Migration Notes

### Upgrading to v1.0.5
- **Documentation moved**: All docs now in `docs/` directory
- **Scripts moved**: All test scripts now in `scripts/` directory  
- **No breaking changes**: All functionality preserved
- **Improved navigation**: New documentation index available

### Upgrading to v1.0.3
- No breaking changes
- Alert management commands are new additions
- Existing alerts can now be managed through CLI
- Database schema automatically updated

### Upgrading from Pre-1.0.0
- Configuration format may require updates
- Database migration handled automatically
- systemd service file may need reinstallation

## Support & Contributing

- **Issues**: [GitHub Issues](https://github.com/ahur-system/sysmedic/issues)
- **Documentation**: See README.md and docs/ directory
- **License**: MIT License
- **Contributing**: Fork, create feature branch, submit PR

---

**Note**: This changelog covers major releases. For detailed commit history and minor updates, see the [GitHub repository](https://github.com/ahur-system/sysmedic).