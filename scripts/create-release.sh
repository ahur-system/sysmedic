#!/bin/bash

# SysMedic GitHub Release Creator
# Creates a new GitHub release and uploads all package assets

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.1"
TAG="v${VERSION}"
RELEASE_TITLE="SysMedic v${VERSION}"

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
RELEASE_NOTES="# SysMedic v${VERSION} - Critical SQLite Fix

## 🚨 CRITICAL BUGFIX RELEASE

This release fixes a critical runtime error that prevented SysMedic from starting on most Linux systems.

### 🔧 Fixed Issues
- **CRITICAL**: Fixed \"Binary was compiled with 'CGO_ENABLED=0', go-sqlite3 requires cgo to work\" error
- **Database**: Replaced go-sqlite3 with pure Go SQLite driver for true static compilation
- **Compatibility**: Now works on all Linux distributions without CGO dependencies
- **Runtime**: Eliminates SQLite-related startup failures on AlmaLinux, RHEL, CentOS, and other distributions

### 📦 What Changed
- Switched from github.com/mattn/go-sqlite3 to github.com/glebarez/go-sqlite
- Updated SQLite driver name from sqlite3 to sqlite
- Maintained full database compatibility and functionality
- Binary size increased slightly (~9.5MB) due to embedded SQLite implementation

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
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
\`\`\`

### RHEL/CentOS/Fedora (.rpm)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -i sysmedic-x86_64.rpm
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
\`\`\`

### Generic Linux (tarball)
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v${VERSION}-linux-amd64.tar.gz
tar xzf sysmedic-v${VERSION}-linux-amd64.tar.gz
cd sysmedic-v${VERSION}-linux-amd64
sudo ./scripts/install.sh
\`\`\`

## 🚀 Quick Start

After installation, SysMedic will be available at:
- **Web Dashboard**: http://localhost:8080
- **API Endpoint**: http://localhost:8080/api/v1
- **Health Check**: http://localhost:8080/health

### Configuration

Edit the configuration file:
\`\`\`bash
sudo nano /etc/sysmedic/config.yaml
sudo systemctl restart sysmedic
\`\`\`

### View Logs
\`\`\`bash
sudo journalctl -u sysmedic -f
\`\`\`

## 📋 System Requirements

- **OS**: Linux (Ubuntu 18.04+, RHEL 7+, CentOS 7+, Debian 9+)
- **Architecture**: x86_64 (AMD64)
- **Memory**: 64MB RAM minimum
- **Disk**: 100MB free space
- **Network**: Port 8080 (configurable)

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
- \`sysmedic-v${VERSION}-linux-amd64.tar.gz\` - Generic tarball
- \`SHA256SUMS\` - Checksums for verification

### Binary Features
- Single statically-linked binary (no dependencies)
- Built with Go for optimal performance
- Comprehensive system monitoring capabilities
- Production-ready with extensive testing

## 🔄 Upgrading

To upgrade from a previous version:

### Debian/Ubuntu
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb
sudo systemctl restart sysmedic
\`\`\`

### RHEL/CentOS
\`\`\`bash
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -U sysmedic-x86_64.rpm
sudo systemctl restart sysmedic
\`\`\`

## 🗑️ Uninstallation

### Debian/Ubuntu
\`\`\`bash
sudo systemctl stop sysmedic
sudo systemctl disable sysmedic
sudo dpkg -r sysmedic
\`\`\`

### RHEL/CentOS
\`\`\`bash
sudo systemctl stop sysmedic
sudo systemctl disable sysmedic
sudo rpm -e sysmedic
\`\`\`

## 🐛 Issues & Support

If you encounter any issues:

1. Check the [troubleshooting guide](https://github.com/ahur-system/sysmedic/blob/main/docs/troubleshooting.md)
2. View logs: \`sudo journalctl -u sysmedic -f\`
3. [Open an issue](https://github.com/ahur-system/sysmedic/issues/new)

## 📝 Changelog

### v${VERSION} (Critical Fix)
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

    if [ ! -f "dist/sysmedic-amd64.deb" ]; then
        missing_files+=("sysmedic-amd64.deb")
    fi

    if [ ! -f "dist/sysmedic-x86_64.rpm" ]; then
        missing_files+=("sysmedic-x86_64.rpm")
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

    # Create new tag
    print_status "Creating tag $TAG..."
    git tag -a "$TAG" -m "Release $TAG"
    git push origin "$TAG"

    # Create release with notes
    print_status "Creating release with assets..."

    # Save release notes to temporary file
    NOTES_FILE=$(mktemp)
    echo "$RELEASE_NOTES" > "$NOTES_FILE"

    # Create release
    gh release create "$TAG" \
        --title "$RELEASE_TITLE" \
        --notes-file "$NOTES_FILE" \
        --latest \
        dist/sysmedic-amd64.deb \
        dist/sysmedic-x86_64.rpm \
        "dist/sysmedic-v${VERSION}-linux-amd64.tar.gz" \
        dist/SHA256SUMS

    # Clean up
    rm -f "$NOTES_FILE"

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
    echo "  • Generic:       https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v${VERSION}-linux-amd64.tar.gz"
    echo ""
    echo "🔍 View release: https://github.com/ahur-system/sysmedic/releases/tag/$TAG"
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
