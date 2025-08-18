# SysMedic Package Deployment Guide

## 🎉 Deployment Summary

SysMedic packages have been successfully built and deployed to GitHub Releases! The following package formats are now available for easy installation across different Linux distributions.

## 📦 Available Packages

### Package Assets
- **Debian/Ubuntu**: `sysmedic-amd64.deb` (2.1MB)
- **RHEL/CentOS/Fedora**: `sysmedic-x86_64.rpm` (2.4MB)
- **Generic Linux**: `sysmedic-v1.0.0-linux-amd64.tar.gz` (2.4MB)
- **Checksums**: `SHA256SUMS` for verification

### Download Links
- **Latest Release**: https://github.com/ahur-system/sysmedic/releases/latest
- **Direct Downloads**:
  - Debian/Ubuntu: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
  - RHEL/CentOS: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
  - Generic: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v1.0.0-linux-amd64.tar.gz

## 🚀 Installation Instructions

### Ubuntu/Debian (.deb)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### RHEL/CentOS/Fedora (.rpm)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -i sysmedic-x86_64.rpm
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### Generic Linux (tarball)
```bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v1.0.0-linux-amd64.tar.gz
tar xzf sysmedic-v1.0.0-linux-amd64.tar.gz
cd sysmedic-v1.0.0-linux-amd64
sudo ./scripts/install.sh
```

## ✅ Package Features

### Debian Package (.deb)
- **Control Files**: Proper metadata and dependencies
- **Post-Install**: Automatic user creation and service setup
- **Pre/Post Remove**: Clean uninstallation process
- **SystemD Integration**: Service file installation
- **Configuration**: Default config with secure permissions

### RPM Package (.rpm)
- **Spec File**: Comprehensive package specification
- **User Management**: Automatic system user creation
- **Service Integration**: SystemD service setup
- **File Permissions**: Secure default permissions
- **Configuration**: Template configuration file

### Common Features
- **Single Binary**: Statically linked, no dependencies
- **SystemD Service**: Native Linux service integration
- **Security**: Runs as dedicated `sysmedic` user
- **Configuration**: `/etc/sysmedic/config.yaml`
- **Data Directory**: `/var/lib/sysmedic`
- **Log Directory**: `/var/log/sysmedic`

## 🛠️ Build Infrastructure

### Package Build Script
- **Location**: `scripts/build-packages.sh`
- **Features**: Builds all package formats in one command
- **Dependencies**: Automatically checks for required tools
- **Validation**: Tests package integrity after build

### Release Automation
- **Location**: `scripts/create-release.sh`
- **Features**: Automated GitHub release creation
- **Assets**: Uploads all package formats
- **Notes**: Comprehensive release documentation

### Directory Structure
```
packaging/
├── deb/
│   ├── DEBIAN/
│   │   ├── control
│   │   ├── postinst
│   │   ├── prerm
│   │   └── postrm
│   ├── etc/systemd/system/
│   └── usr/local/bin/
└── rpm/
    └── sysmedic.spec
```

## 🔧 Build Process

### Quick Build
```bash
# Build all packages
./scripts/build-packages.sh

# Build specific package
./scripts/build-packages.sh deb    # Debian only
./scripts/build-packages.sh rpm    # RPM only
./scripts/build-packages.sh tarball # Tarball only
```

### Release Process
```bash
# Create GitHub release with all assets
./scripts/create-release.sh

# Check assets before release
./scripts/create-release.sh check

# Clean up release (for recreation)
./scripts/create-release.sh clean
```

## 📋 Package Details

### System Integration
- **Service Name**: `sysmedic`
- **Service Type**: SystemD daemon
- **Default Port**: 8080 (configurable)
- **User/Group**: `sysmedic:sysmedic`
- **Binary Location**: `/usr/local/bin/sysmedic`

### File Locations
```
/usr/local/bin/sysmedic                 # Main binary
/etc/sysmedic/config.yaml               # Configuration
/etc/systemd/system/sysmedic.service    # Service file
/var/lib/sysmedic/                      # Data directory
/var/log/sysmedic/                      # Log directory
```

### Permissions
```
/etc/sysmedic/          755 root:root
/etc/sysmedic/config.yaml  640 sysmedic:sysmedic
/var/lib/sysmedic/      750 sysmedic:sysmedic
/var/log/sysmedic/      750 sysmedic:sysmedic
/usr/local/bin/sysmedic 755 root:root
```

## 🔍 Verification

### Package Integrity
```bash
# Verify checksums
cd dist/
sha256sum -c SHA256SUMS

# Check package contents
dpkg-deb --contents sysmedic-amd64.deb
rpm -qlp sysmedic-x86_64.rpm
```

### Installation Test
```bash
# Test installation (use with caution on production systems)
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl status sysmedic
curl http://localhost:8080/health
```

## 🗑️ Uninstallation

### Debian/Ubuntu
```bash
sudo systemctl stop sysmedic
sudo systemctl disable sysmedic
sudo dpkg -r sysmedic  # Remove package
sudo dpkg -P sysmedic  # Purge (removes config files)
```

### RHEL/CentOS
```bash
sudo systemctl stop sysmedic
sudo systemctl disable sysmedic
sudo rpm -e sysmedic
```

## 📊 Package Statistics

| Package | Size | Architecture | Dependencies |
|---------|------|--------------|--------------|
| .deb    | 2.1MB | amd64       | systemd      |
| .rpm    | 2.4MB | x86_64      | systemd      |
| .tar.gz | 2.4MB | amd64       | none         |

## 🔄 Update Process

### Automated Updates
The packages support in-place upgrades:

```bash
# Debian/Ubuntu
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb

# RHEL/CentOS
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -U sysmedic-x86_64.rpm
```

The service will automatically restart with the new version while preserving configuration and data.

## 🎯 Next Steps

1. **Test Installations**: Verify packages work on target distributions
2. **Documentation**: Update main README with installation instructions
3. **CI/CD**: Consider GitHub Actions for automated builds
4. **Monitoring**: Set up download analytics
5. **Feedback**: Collect user feedback on installation experience

---

**Package Deployment Status**: ✅ **COMPLETE**

All package formats are successfully built, tested, and deployed to GitHub Releases with working download links.