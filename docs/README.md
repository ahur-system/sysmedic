# SysMedic Documentation

This directory contains comprehensive documentation for SysMedic, organized by topic and functionality.

## 📚 Documentation Index

### Core Documentation
- **[WEBSOCKET.md](WEBSOCKET.md)** - Complete WebSocket remote monitoring guide
  - Setup and configuration
  - API documentation with request-response protocol
  - Client examples (JavaScript, Python, Node.js)
  - Testing guides and troubleshooting
  - Implementation details and security considerations

- **[IMPLEMENTATION.md](IMPLEMENTATION.md)** - Technical implementation details
  - Architecture overview
  - Code structure and design patterns
  - Database schema and storage
  - Performance considerations

- **[DEPLOYMENT.md](DEPLOYMENT.md)** - Deployment and publishing guide
  - Package building for multiple Linux distributions
  - GitHub publishing and release management
  - Installation methods (DEB, RPM, generic Linux)
  - Distribution channels and download links
  - Troubleshooting deployment issues

- **[PORT_CHANGE_SUMMARY.md](PORT_CHANGE_SUMMARY.md)** - Port configuration changes
  - Historical port changes and migration notes
  - Configuration updates and compatibility

## 📖 Documentation Overview

### WebSocket Documentation
The **WEBSOCKET.md** file provides everything needed for remote monitoring:
- Quick start guide for immediate setup
- Complete API reference with message types
- Bidirectional communication with request-response protocol
- Multiple client examples for different programming languages
- Comprehensive testing methodologies
- Security best practices and troubleshooting

### Implementation Documentation
The **IMPLEMENTATION.md** file covers technical details:
- System architecture and component interactions
- Code organization and design decisions
- Database design and data flow
- Performance characteristics and optimization

### Deployment Documentation
The **DEPLOYMENT.md** file covers the complete deployment process:
- Building packages for Debian/Ubuntu (DEB) and RHEL/CentOS (RPM)
- GitHub repository setup and authentication
- Release management with semantic versioning
- Multiple installation methods for different environments
- Distribution strategies and download statistics

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
| **Remote Monitoring** | [WEBSOCKET.md](WEBSOCKET.md) | WebSocket API, clients, testing |
| **Technical Details** | [IMPLEMENTATION.md](IMPLEMENTATION.md) | Architecture, code structure |
| **Deployment** | [DEPLOYMENT.md](DEPLOYMENT.md) | Building, publishing, installation |
| **Port Changes** | [PORT_CHANGE_SUMMARY.md](PORT_CHANGE_SUMMARY.md) | Configuration migration |

## 📋 Documentation Standards

### Format
- All documentation is written in GitHub Flavored Markdown
- Each file includes a table of contents for easy navigation
- Code examples include syntax highlighting
- Screenshots and diagrams where helpful

### Organization
- **Comprehensive**: Each topic is covered in a single, complete document
- **Self-contained**: Documents can be read independently
- **Cross-referenced**: Related topics link to each other
- **Up-to-date**: Documentation is maintained with each release

### Maintenance
- Documentation is updated with each feature release
- Examples are tested to ensure they work correctly
- Links are verified regularly
- Community feedback is incorporated

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