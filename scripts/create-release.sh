#!/bin/bash

# SysMedic GitHub Release Creator
# Creates a new GitHub release and uploads all package assets

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.6"
TAG="v${VERSION}"
RELEASE_TITLE="SysMedic v${VERSION} - Arch Linux Package Support"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Release notes
RELEASE_NOTES="# SysMedic v${VERSION} - Arch Linux Package Support

## 🎯 MAJOR PACKAGING IMPROVEMENT: Complete Arch Linux Support

This release introduces comprehensive Arch Linux package support with proper `.pkg.tar.zst` format, making SysMedic available across all major Linux distributions with native package managers.

### ✨ Key Improvements
- **Arch Linux Support**: Proper `.pkg.tar.zst` packages with zstd compression
- **Standards Compliance**: Follows official Arch Linux packaging guidelines
- **Package Metadata**: Complete `.PKGINFO` and `.MTREE` files for integrity
- **Optimal Compression**: zstd ultra-20 compression for smaller downloads
- **Release Integration**: Arch packages automatically included in GitHub releases

### 🔧 What's New
- **Arch Package Format**: \`sysmedic-1.0.6-1-x86_64.pkg.tar.zst\` (~4.1MB)
- **User-Friendly Downloads**: \`sysmedic-arch.pkg.tar.zst\` for easy access
- **Package Validation**: Comprehensive testing script for format verification
- **Installation Methods**: Multiple installation options including AUR support
- **Complete Documentation**: Installation, upgrade, and uninstallation guides

### 📦 Package Distribution
- **Debian/Ubuntu**: `.deb` packages with apt integration
- **RHEL/CentOS/Fedora**: `.rpm` packages with yum/dnf support
- **Arch Linux**: `.pkg.tar.zst` packages with pacman support
- **Generic Linux**: `.tar.gz` archives for manual installation
- **Checksums**: SHA256SUMS for all packages

### 🚀 Enhanced Features (Single Binary Multi-Daemon Architecture)
- **Single Binary**: One 11MB binary with multiple daemon modes
- **Independent Daemons**: Doctor (monitoring) and WebSocket (remote access)
- **Smart User Filtering**: Focus on real users causing actual problems
- **Real Usernames**: Proper username display instead of uid_[id] format
- **Configurable Monitoring**: Control intervals, thresholds, and filtering
- **SystemD Integration**: Separate services using same binary

## ✨ Features (unchanged)

- 📊 **Real-time Monitoring**: CPU, memory, disk, and network monitoring
- 🔍 **System Diagnostics**: Advanced health checks and performance analysis
- 📈 **Historical Data**: Data collection and trend analysis
- 🌐 **Web Dashboard**: Interactive web-based interface
- 🔌 **REST API**: Programmatic access to system metrics
- ⚙️ **Configurable**: Customizable monitoring intervals and thresholds
- 🔧 **SystemD Integration**: Native Linux service support
- 🚨 **Smart Alerts**: Configurable alerting for system issues

## 📦 Installation

### Ubuntu/Debian (.deb)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl enable sysmedic.doctor
sudo systemctl enable sysmedic.websocket
sudo systemctl start sysmedic.doctor
sudo systemctl start sysmedic.websocket
\`\`\`

### RHEL/CentOS/Fedora (.rpm)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -i sysmedic-x86_64.rpm
sudo systemctl enable sysmedic.doctor
sudo systemctl enable sysmedic.websocket
sudo systemctl start sysmedic.doctor
sudo systemctl start sysmedic.websocket
\`\`\`

### Arch Linux (.pkg.tar.zst)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-arch.pkg.tar.zst
sudo pacman -U sysmedic-arch.pkg.tar.zst
sudo systemctl enable sysmedic.doctor
sudo systemctl enable sysmedic.websocket
sudo systemctl start sysmedic.doctor
sudo systemctl start sysmedic.websocket
\`\`\`

### Generic Linux (tarball)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v${VERSION}-linux-amd64.tar.gz
tar xzf sysmedic-v${VERSION}-linux-amd64.tar.gz
cd sysmedic-v${VERSION}-linux-amd64
sudo ./scripts/install.sh
\`\`\`

## 🚀 Quick Start

After installation, manage SysMedic daemons:

### Daemon Management
\`\`\`bash
# Start monitoring daemon
sysmedic daemon start

# Start WebSocket daemon for remote access
sysmedic websocket start

# Check status of both daemons
sysmedic daemon status

# WebSocket server available at:
# ws://localhost:8060/ws
\`\`\`

### Configuration

Edit the configuration file:
\`\`\`bash
sudo nano /etc/sysmedic/config.yaml
sudo systemctl restart sysmedic.doctor
sudo systemctl restart sysmedic.websocket
\`\`\`

### View Logs
\`\`\`bash
# Monitor doctor daemon logs
sudo journalctl -u sysmedic.doctor -f

# Monitor WebSocket daemon logs
sudo journalctl -u sysmedic.websocket -f
\`\`\`

## 📋 System Requirements

- **OS**: Linux (Ubuntu 18.04+, RHEL 7+, CentOS 7+, Debian 9+)
- **Architecture**: x86_64 (AMD64)
- **Memory**: 64MB RAM minimum
- **Disk**: 100MB free space
- **Network**: Port 8060 for WebSocket (configurable)

## 🔒 Security

- Runs as dedicated \`sysmedic\` user
- Minimal privileges and secure defaults
- No root access required for normal operation
- Configurable TLS encryption

## 📚 Documentation

- [Installation Guide](https://github.com/ahur-system/sysmedic#installation)
- [Configuration Reference](https://github.com/ahur-system/sysmedic/blob/main/docs/configuration.md)
- [API Documentation](https://github.com/ahur-system/sysmedic/blob/main/docs/api.md)
- [Troubleshooting](https://github.com/ahur-system/sysmedic/blob/main/docs/troubleshooting.md)

## 🛠️ What's Included

### Package Assets
- \`sysmedic-amd64.deb\` - Debian/Ubuntu package
- \`sysmedic-x86_64.rpm\` - RHEL/CentOS/Fedora package
- \`sysmedic-arch.pkg.tar.zst\` - Arch Linux package
- \`sysmedic-v${VERSION}-linux-amd64.tar.gz\` - Generic tarball
- \`SHA256SUMS\` - Checksums for verification

### Binary Features
- Single binary with multiple daemon modes (11MB)
- Built with Go for optimal performance
- Independent daemon processes for monitoring and WebSocket
- Complete process separation and isolation
- Production-ready with extensive testing

## 🔄 Upgrading

To upgrade from a previous version:

### Debian/Ubuntu
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl restart sysmedic.doctor
sudo systemctl restart sysmedic.websocket
\`\`\`

### RHEL/CentOS
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -U sysmedic-x86_64.rpm
sudo systemctl restart sysmedic.doctor
sudo systemctl restart sysmedic.websocket
\`\`\`

### Arch Linux
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-arch.pkg.tar.zst
sudo pacman -U sysmedic-arch.pkg.tar.zst
sudo systemctl restart sysmedic.doctor
sudo systemctl restart sysmedic.websocket
\`\`\`

## 🗑️ Uninstallation

### Debian/Ubuntu
\`\`\`bash
sudo systemctl stop sysmedic.doctor sysmedic.websocket
sudo systemctl disable sysmedic.doctor sysmedic.websocket
sudo dpkg -r sysmedic
\`\`\`

### RHEL/CentOS
\`\`\`bash
sudo systemctl stop sysmedic.doctor sysmedic.websocket
sudo systemctl disable sysmedic.doctor sysmedic.websocket
sudo rpm -e sysmedic
\`\`\`

### Arch Linux
\`\`\`bash
sudo systemctl stop sysmedic.doctor sysmedic.websocket
sudo systemctl disable sysmedic.doctor sysmedic.websocket
sudo pacman -R sysmedic
\`\`\`

## 🐛 Issues & Support

If you encounter any issues:

1. Check the [troubleshooting guide](https://github.com/ahur-system/sysmedic/blob/main/docs/troubleshooting.md)
2. View logs: \`sudo journalctl -u sysmedic.doctor -f\` or \`sudo journalctl -u sysmedic.websocket -f\`
3. Check daemon status: \`sysmedic daemon status\`
4. [Open an issue](https://github.com/ahur-system/sysmedic/issues/new)

## 📝 Changelog

### v${VERSION} (Arch Linux Package Support)
- 📦 **MAJOR**: Complete Arch Linux package support with .pkg.tar.zst format
- 🔧 **NEW**: Proper package metadata (.PKGINFO, .MTREE) and zstd compression
- 🚀 **ENHANCED**: Multi-distribution support (DEB, RPM, Arch, Generic)
- 📋 **IMPROVED**: Comprehensive installation documentation for all platforms
- ✅ **ADDED**: Package validation and testing infrastructure
- 🌐 **EXPANDED**: Release automation includes all package formats

### v1.0.5 (Single Binary Multi-Daemon Architecture)
- 🏗️ **MAJOR**: Single binary with multiple daemon modes
- 🔄 **NEW**: Independent doctor and WebSocket daemon processes
- 📦 **SIMPLIFIED**: One binary (11MB) handles all functionality
- 🎯 **ENHANCED**: Complete process separation and isolation
- ⚙️ **IMPROVED**: Better resource management and fault tolerance
- 🚀 **ADDED**: Independent daemon lifecycle management

### v1.0.1 (Critical Fix)
- 🚨 **CRITICAL FIX**: Replaced go-sqlite3 with pure Go SQLite driver
- 🔧 **FIXED**: \"CGO_ENABLED=0\" runtime error on all Linux distributions
- ✅ **VERIFIED**: Now works on AlmaLinux, RHEL, CentOS, Ubuntu, Debian, and more
- 📦 **IMPROVED**: True static compilation without CGO dependencies
- 🛡️ **ENHANCED**: Better portability and deployment reliability

### v1.0.0 (Initial Release)
- ✨ Initial public release
- 📊 Real-time system monitoring
- 🌐 Web dashboard interface
- 🔌 REST API for integration
- 📦 Native .deb and .rpm packages
- 🔧 SystemD service integration
- ⚙️ Comprehensive configuration options
- 🚨 Configurable alerting system

---

**Full Changelog**: https://github.com/ahur-system/sysmedic/commits/v${VERSION}"

# Check if we're in the right directory
check_directory() {
    if [ ! -f "go.mod" ] || [ ! -d "cmd/sysmedic" ]; then
        print_error "Not in SysMedic project root directory"
        exit 1
    fi
}

# Check if required files exist
check_assets() {
    print_status "Checking release assets..."

    local missing_files=()

    if [ ! -f "dist/sysmedic_${VERSION}_amd64.deb" ]; then
        missing_files+=("sysmedic_${VERSION}_amd64.deb")
    fi

    if [ ! -f "dist/sysmedic-${VERSION}-1.x86_64.rpm" ]; then
        missing_files+=("sysmedic-${VERSION}-1.x86_64.rpm")
    fi

    if [ ! -f "dist/sysmedic-${VERSION}-1-x86_64.pkg.tar.zst" ]; then
        missing_files+=("sysmedic-${VERSION}-1-x86_64.pkg.tar.zst")
    fi

    if [ ! -f "dist/sysmedic-v${VERSION}-linux-amd64.tar.gz" ]; then
        missing_files+=("sysmedic-v${VERSION}-linux-amd64.tar.gz")
    fi

    if [ ! -f "dist/SHA256SUMS" ]; then
        missing_files+=("SHA256SUMS")
    fi

    if [ ${#missing_files[@]} -gt 0 ]; then
        print_error "Missing release assets: ${missing_files[*]}"
        print_error "Run './scripts/build-packages.sh' first to build packages"
        exit 1
    fi

    print_success "All release assets found"
}

# Create or update release
create_release() {
    print_status "Creating GitHub release $TAG..."

    # Check if release already exists
    if gh release view "$TAG" >/dev/null 2>&1; then
        print_warning "Release $TAG already exists. Deleting and recreating..."
        gh release delete "$TAG" --yes

        # Also delete the tag locally and remotely
        git tag -d "$TAG" 2>/dev/null || true
        git push origin --delete "$TAG" 2>/dev/null || true
    fi

    # Create new tag if it doesn't exist
    if git rev-parse "$TAG" >/dev/null 2>&1; then
        print_warning "Tag $TAG already exists, skipping tag creation"
    else
        print_status "Creating tag $TAG..."
        git tag -a "$TAG" -m "Release $TAG"
        git push origin "$TAG"
    fi

    # Create release with notes
    print_status "Creating release with assets..."

    # Save release notes to temporary file
    NOTES_FILE=$(mktemp)
    echo "$RELEASE_NOTES" > "$NOTES_FILE"

    # Create generic named copies for user-friendly download URLs
    print_status "Creating generic named copies..."
    if [ -f "dist/sysmedic_${VERSION}_amd64.deb" ]; then
        cp "dist/sysmedic_${VERSION}_amd64.deb" "dist/sysmedic-amd64.deb"
    fi
    if [ -f "dist/sysmedic-${VERSION}-1.x86_64.rpm" ]; then
        cp "dist/sysmedic-${VERSION}-1.x86_64.rpm" "dist/sysmedic-x86_64.rpm"
    fi
    if [ -f "dist/sysmedic-${VERSION}-1-x86_64.pkg.tar.zst" ]; then
        cp "dist/sysmedic-${VERSION}-1-x86_64.pkg.tar.zst" "dist/sysmedic-arch.pkg.tar.zst"
    fi

    # Create release
    gh release create "$TAG" \
        --title "$RELEASE_TITLE" \
        --notes-file "$NOTES_FILE" \
        --latest \
        "dist/sysmedic_${VERSION}_amd64.deb" \
        "dist/sysmedic-${VERSION}-1.x86_64.rpm" \
        "dist/sysmedic-${VERSION}-1-x86_64.pkg.tar.zst" \
        "dist/sysmedic-v${VERSION}-linux-amd64.tar.gz" \
        dist/SHA256SUMS \
        "dist/sysmedic-amd64.deb" \
        "dist/sysmedic-x86_64.rpm" \
        "dist/sysmedic-arch.pkg.tar.zst"

    # Clean up
    rm -f "$NOTES_FILE"
    rm -f "dist/sysmedic-amd64.deb" "dist/sysmedic-x86_64.rpm" "dist/sysmedic-arch.pkg.tar.zst"

    print_success "Release $TAG created successfully!"
}

# Verify release
verify_release() {
    print_status "Verifying release..."

    # Check release exists
    if ! gh release view "$TAG" >/dev/null 2>&1; then
        print_error "Release verification failed - release not found"
        exit 1
    fi

    # List release assets
    print_status "Release assets:"
    gh release view "$TAG" --json assets --jq '.assets[].name' | while read asset; do
        print_success "✓ $asset"
    done

    # Get release URL
    RELEASE_URL=$(gh release view "$TAG" --json url --jq '.url')
    print_success "Release URL: $RELEASE_URL"
}

# Main execution
main() {
    print_status "🚀 Creating SysMedic release $TAG..."

    cd "$PROJECT_ROOT"

    check_directory
    check_assets
    create_release
    verify_release

    echo ""
    print_success "🎉 Release $TAG created successfully!"
    echo ""
    echo "📦 Download links:"
    echo "  • Debian/Ubuntu: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb"
    echo "  • RHEL/CentOS:   https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm"
    echo "  • Arch Linux:    https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-arch.pkg.tar.zst"
    echo "  • Generic:       https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v${VERSION}-linux-amd64.tar.gz"
    echo "  • View config: $BINARY_NAME config show"
    echo ""
    echo "🔍 View release: https://github.com/ahur-system/sysmedic/releases/tag/$TAG"
    echo ""
    echo "💡 Single Binary Multi-Daemon Architecture:"
    echo "  • One binary (11MB) with multiple daemon modes"
    echo "  • Independent doctor (monitoring) and WebSocket processes"
    echo "  • Complete process separation with single deployment"
    echo ""
    echo "✅ The installation commands in your documentation will now work correctly!"
}

# Handle script arguments
case "${1:-}" in
    "check")
        cd "$PROJECT_ROOT"
        check_directory
        check_assets
        print_success "All checks passed - ready for release"
        ;;
    "clean")
        print_status "Cleaning release artifacts..."
        gh release delete "$TAG" --yes 2>/dev/null || true
        git tag -d "$TAG" 2>/dev/null || true
        git push origin --delete "$TAG" 2>/dev/null || true
        print_success "Release cleanup complete"
        ;;
    *)
        main
        ;;
esac
