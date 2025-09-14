# SysMedic Deployment & Installation Guide

## Overview

This comprehensive guide covers building, packaging, publishing, and installing SysMedic across different Linux platforms. It includes complete instructions for package creation, GitHub publishing, release management, and multiple installation methods.

## Table of Contents

1. [Build Environment Setup](#build-environment-setup)
2. [Package Building](#package-building)
3. [GitHub Publishing](#github-publishing)
4. [Installation Methods](#installation-methods)
5. [Release Management](#release-management)
6. [Distribution Channels](#distribution-channels)
7. [Automated Deployment](#automated-deployment)
8. [Troubleshooting](#troubleshooting)

## Build Environment Setup

### Prerequisites

#### Ubuntu/Debian Build Environment
```bash
# Update package list
sudo apt update

# Install build tools
sudo apt install -y \
    golang-go \
    build-essential \
    git \
    curl \
    wget \
    fpm \
    rpm \
    ruby-dev \
    gcc

# Install latest Go (if needed)
wget https://golang.org/dl/go1.21.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc
```

#### RHEL/CentOS/Fedora Build Environment
```bash
# Install development tools
sudo yum groupinstall -y "Development Tools"
sudo yum install -y golang git curl wget rpm-build

# Install FPM for cross-platform packaging
sudo yum install -y ruby-devel
gem install fpm

# Or on newer systems with dnf
sudo dnf groupinstall -y "Development Tools"
sudo dnf install -y golang git curl wget rpm-build ruby-devel
gem install fpm
```

#### Arch Linux Build Environment
```bash
# Update system
sudo pacman -Syu

# Install build tools
sudo pacman -S --needed base-devel go git curl wget ruby

# Install FPM for cross-platform packaging
sudo gem install fpm

# Install makepkg and AUR helper (optional)
sudo pacman -S --needed devtools namcap

# For AUR package development
sudo pacman -S --needed git openssh
```

#### Docker Build Environment
```dockerfile
FROM ubuntu:22.04

RUN apt-get update && apt-get install -y \
    golang-go \
    build-essential \
    git \
    curl \
    wget \
    fpm \
    rpm \
    ruby-dev \
    gcc \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace
COPY . .
RUN make packages
```

### Go Environment Configuration
```bash
# Check Go installation
go version

# Set Go environment variables
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
export GO111MODULE=on

# Verify environment
go env
```

## Package Building

### Project Structure
```
sysmedic/
├── cmd/sysmedic/main.go          # Main application entry point
├── internal/                     # Internal packages
├── pkg/                         # Public packages
├── scripts/                     # Build and deployment scripts
├── packaging/                   # Package configuration files
├── Makefile                     # Build automation
├── go.mod                       # Go module definition
└── go.sum                       # Go module checksums
```

### Build Process

#### Quick Build
```bash
# Clone repository
git clone https://github.com/ahur-system/sysmedic.git
cd sysmedic

# Build binary
make build

# Build all packages
make packages

# Clean build artifacts
make clean
```

#### Detailed Build Steps

##### 1. Build Binary
```bash
# Build for current platform
go build -ldflags="-s -w" -o sysmedic cmd/sysmedic/main.go

# Build with version information
VERSION=$(git describe --tags --always)
go build -ldflags="-s -w -X main.Version=${VERSION}" -o sysmedic cmd/sysmedic/main.go

# Cross-compile for different architectures
GOOS=linux GOARCH=amd64 go build -o sysmedic-linux-amd64 cmd/sysmedic/main.go
GOOS=linux GOARCH=arm64 go build -o sysmedic-linux-arm64 cmd/sysmedic/main.go
```

##### 2. Create Package Structure
```bash
# Create packaging directories
mkdir -p dist/
mkdir -p packaging/deb/
mkdir -p packaging/rpm/
mkdir -p packaging/generic/

# Copy files to package structure
cp sysmedic packaging/deb/usr/local/bin/
cp scripts/sysmedic.doctor.service packaging/deb/etc/systemd/system/
cp scripts/sysmedic.websocket.service packaging/deb/etc/systemd/system/
cp scripts/config.example.yaml packaging/deb/etc/sysmedic/config.yaml
```

##### 3. Build DEB Package
```bash
fpm -s dir -t deb -n sysmedic -v 1.0.5 \
    --description "Single binary multi-daemon Linux system monitoring tool with user-centric resource tracking" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --vendor "SysMedic" \
    --license "MIT" \
    --architecture amd64 \
    --depends "systemd >= 219" \
    --after-install scripts/post-install.sh \
    --before-remove scripts/pre-remove.sh \
    --after-remove scripts/post-remove.sh \
    --deb-systemd scripts/sysmedic.doctor.service \
    --deb-systemd scripts/sysmedic.websocket.service \
    --config-files /etc/sysmedic/config.yaml \
    sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.doctor.service=/etc/systemd/system/sysmedic.doctor.service \
    scripts/sysmedic.websocket.service=/etc/systemd/system/sysmedic.websocket.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml
```

##### 4. Build RPM Package
```bash
fpm -s dir -t rpm -n sysmedic -v 1.0.5 \
    --description "Single binary multi-daemon Linux system monitoring tool with user-centric resource tracking" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --vendor "SysMedic" \
    --license "MIT" \
    --architecture x86_64 \
    --depends "systemd >= 219" \
    --after-install scripts/post-install.sh \
    --before-remove scripts/pre-remove.sh \
    --after-remove scripts/post-remove.sh \
    --config-files /etc/sysmedic/config.yaml \
    sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.doctor.service=/etc/systemd/system/sysmedic.doctor.service \
    scripts/sysmedic.websocket.service=/etc/systemd/system/sysmedic.websocket.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml
```

##### 5. Build Arch Package
```bash
# Create PKGBUILD for Arch Linux
mkdir -p packaging/arch
cat > packaging/arch/PKGBUILD << 'EOF'
# Maintainer: SysMedic Team <team@sysmedic.dev>
pkgname=sysmedic
pkgver=1.0.5
pkgrel=1
pkgdesc="Single binary multi-daemon Linux system monitoring tool with user-centric resource tracking"
arch=('x86_64' 'aarch64')
url="https://github.com/ahur-system/sysmedic"
license=('MIT')
depends=('systemd')
makedepends=('go' 'git')
source=("$pkgname-$pkgver.tar.gz::https://github.com/ahur-system/sysmedic/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')  # Update with actual checksums

build() {
    cd "$pkgname-$pkgver"
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
    
    go build -ldflags="-s -w -X main.Version=${pkgver}" -o sysmedic cmd/sysmedic/main.go
}

package() {
    cd "$pkgname-$pkgver"
    
    # Install binary
    install -Dm755 sysmedic "$pkgdir/usr/bin/sysmedic"
    
    # Install systemd services
    install -Dm644 scripts/sysmedic.doctor.service "$pkgdir/usr/lib/systemd/system/sysmedic.doctor.service"
    install -Dm644 scripts/sysmedic.websocket.service "$pkgdir/usr/lib/systemd/system/sysmedic.websocket.service"
    
    # Install configuration
    install -Dm644 scripts/config.example.yaml "$pkgdir/etc/sysmedic/config.yaml"
    
    # Install documentation
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    
    # Create directories
    install -dm755 "$pkgdir/var/lib/sysmedic"
    install -dm755 "$pkgdir/var/log/sysmedic"
}
EOF

# Build Arch package
cd packaging/arch
makepkg -sf --noconfirm
```

##### 6. Create Generic Archive
```bash
# Create tar.gz archive
tar -czf sysmedic-v1.0.5-linux-amd64.tar.gz \
    sysmedic \
    scripts/ \
    README.md \
    LICENSE \
    CHANGELOG.md
```

### Package Output Summary

| Package Type | Filename | Size | Target Distribution |
|--------------|----------|------|-------------------|
| **Debian/Ubuntu** | `sysmedic_1.0.5_amd64.deb` | ~4.1MB | Debian, Ubuntu, derivatives |
| **RHEL/CentOS** | `sysmedic-1.0.5-1.x86_64.rpm` | ~4.7MB | RHEL, CentOS, Fedora, SUSE |
| **Arch Linux** | `sysmedic-1.0.5-1-x86_64.pkg.tar.zst` | ~4.1MB | Arch Linux, Manjaro |
| **Generic Linux** | `sysmedic-v1.0.5-linux-amd64.tar.gz` | ~4.7MB | Any Linux x86_64 |
| **ARM64 Linux** | `sysmedic-v1.0.5-linux-arm64.tar.gz` | ~4.5MB | ARM64 Linux systems |
| **Checksums** | `SHA256SUMS` | ~1KB | Verification file |

### Automated Build Script

```bash
#!/bin/bash
# scripts/build-packages.sh

set -e

# Configuration
VERSION=${VERSION:-$(git describe --tags --always)}
ARCH=${ARCH:-amd64}
OUTPUT_DIR="dist"

echo "Building SysMedic packages v${VERSION} for ${ARCH}"

# Clean previous builds
rm -rf ${OUTPUT_DIR}
mkdir -p ${OUTPUT_DIR}

# Build binary
echo "Building binary..."
GOOS=linux GOARCH=${ARCH} go build \
    -ldflags="-s -w -X main.Version=${VERSION}" \
    -o ${OUTPUT_DIR}/sysmedic \
    cmd/sysmedic/main.go

# Make binary executable
chmod +x ${OUTPUT_DIR}/sysmedic

# Build DEB package
echo "Building DEB package..."
fpm -s dir -t deb -n sysmedic -v ${VERSION} \
    --description "Single binary multi-daemon Linux system monitoring tool" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --license "MIT" \
    --architecture ${ARCH} \
    --depends "systemd >= 219" \
    --after-install scripts/post-install.sh \
    --before-remove scripts/pre-remove.sh \
    --config-files /etc/sysmedic/config.yaml \
    --package ${OUTPUT_DIR} \
    ${OUTPUT_DIR}/sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.doctor.service=/etc/systemd/system/sysmedic.doctor.service \
    scripts/sysmedic.websocket.service=/etc/systemd/system/sysmedic.websocket.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml

# Build RPM package
echo "Building RPM package..."
fpm -s dir -t rpm -n sysmedic -v ${VERSION} \
    --description "Single binary multi-daemon Linux system monitoring tool" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --license "MIT" \
    --architecture x86_64 \
    --depends "systemd >= 219" \
    --after-install scripts/post-install.sh \
    --before-remove scripts/pre-remove.sh \
    --config-files /etc/sysmedic/config.yaml \
    --package ${OUTPUT_DIR} \
    ${OUTPUT_DIR}/sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.doctor.service=/etc/systemd/system/sysmedic.doctor.service \
    scripts/sysmedic.websocket.service=/etc/systemd/system/sysmedic.websocket.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml

# Build Arch package (if on Arch system or with makepkg available)
if command -v makepkg >/dev/null 2>&1; then
    echo "Building Arch package..."
    mkdir -p ${OUTPUT_DIR}/arch
    
    # Create PKGBUILD
    cat > ${OUTPUT_DIR}/arch/PKGBUILD << EOF
pkgname=sysmedic-bin
pkgver=${VERSION}
pkgrel=1
pkgdesc="Single binary multi-daemon Linux system monitoring tool"
arch=('x86_64' 'aarch64')
url="https://github.com/ahur-system/sysmedic"
license=('MIT')
depends=('systemd')
provides=('sysmedic')
conflicts=('sysmedic')

package() {
    install -Dm755 "${OUTPUT_DIR}/sysmedic" "\$pkgdir/usr/bin/sysmedic"
    install -Dm644 "scripts/sysmedic.doctor.service" "\$pkgdir/usr/lib/systemd/system/sysmedic.doctor.service"
    install -Dm644 "scripts/sysmedic.websocket.service" "\$pkgdir/usr/lib/systemd/system/sysmedic.websocket.service"
    install -Dm644 "scripts/config.example.yaml" "\$pkgdir/etc/sysmedic/config.yaml"
    install -dm755 "\$pkgdir/var/lib/sysmedic"
    install -dm755 "\$pkgdir/var/log/sysmedic"
}
EOF
    
    # Set compression to zstd for proper .pkg.tar.zst creation
    export PKGEXT='.pkg.tar.zst'
    export COMPRESSZST=(zstd -c -T0 --ultra -20 -)
        
    # Build package
    cd ${OUTPUT_DIR}/arch
    makepkg -sf --noconfirm
    mv *.pkg.tar.zst ../ 2>/dev/null || true
    cd ../..
else
    echo "Skipping Arch package build (makepkg not available)"
fi

# Create generic archive
echo "Creating generic archive..."
cd ${OUTPUT_DIR}
tar -czf sysmedic-v${VERSION}-linux-${ARCH}.tar.gz \
    sysmedic \
    ../scripts/ \
    ../README.md \
    ../LICENSE \
    ../CHANGELOG.md
cd ..

# Generate checksums
echo "Generating checksums..."
cd ${OUTPUT_DIR}
sha256sum *.deb *.rpm *.tar.gz *.pkg.tar.zst > SHA256SUMS 2>/dev/null || sha256sum *.deb *.rpm *.tar.gz > SHA256SUMS
cd ..

echo "Build complete! Packages available in ${OUTPUT_DIR}/"
ls -la ${OUTPUT_DIR}/
```

## GitHub Publishing

### Authentication Setup

#### Personal Access Token (Recommended)
```bash
# 1. Create Personal Access Token on GitHub:
#    Settings → Developer settings → Personal access tokens → Tokens (classic)
#    Scopes: repo (full repository access)

# 2. Configure Git
git config --global user.name "your-username"
git config --global user.email "your-email@example.com"

# 3. Use token as password when prompted
# Store token securely for future use
```

#### SSH Keys
```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your-email@example.com" -f ~/.ssh/sysmedic_deploy

# Add to ssh-agent
eval "$(ssh-agent -s)"
ssh-add ~/.ssh/sysmedic_deploy

# Copy public key to GitHub
cat ~/.ssh/sysmedic_deploy.pub
# Add this to GitHub: Settings → SSH and GPG keys

# Test connection
ssh -T git@github.com
```

### Repository Setup

#### Create Repository
```bash
# On GitHub.com:
# 1. Go to https://github.com/new
# 2. Repository name: sysmedic
# 3. Description: "Cross-platform Linux server monitoring CLI tool with user-centric resource tracking"
# 4. Set to Public
# 5. Don't initialize with README (we have our own)
```

#### Initial Push
```bash
# Add remote origin
git remote add origin https://github.com/ahur-system/sysmedic.git

# Push main branch
git branch -M main
git push -u origin main

# Push tags
git push origin --tags
```

### Release Creation

#### Automated Release (GitHub Actions)

Create `.github/workflows/release.yml`:
```yaml
name: Build and Release

on:
  push:
    tags:
      - 'v*'
  workflow_dispatch:
    inputs:
      version:
        description: 'Release version'
        required: true
        default: 'v1.0.x'

env:
  GO_VERSION: '1.21'

jobs:
  build:
    name: Build Packages
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
      with:
        fetch-depth: 0
    
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: ${{ env.GO_VERSION }}
    
    - name: Install build dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y rpm ruby-dev libarchive-tools
        sudo gem install fpm
        
        # Install makepkg for Arch package building
        wget https://github.com/archlinux/pacman/archive/v6.0.2.tar.gz
        tar -xzf v6.0.2.tar.gz
        cd pacman-6.0.2
        ./configure --prefix=/usr --sysconfdir=/etc --localstatedir=/var
        make -j$(nproc)
        sudo make install
        cd ..
    
    - name: Get version
      id: version
      run: |
        if [[ "${{ github.ref }}" == refs/tags/* ]]; then
          VERSION=${GITHUB_REF#refs/tags/}
        else
          VERSION=${{ github.event.inputs.version }}
        fi
        echo "VERSION=${VERSION}" >> $GITHUB_OUTPUT
        echo "Building version: ${VERSION}"
    
    - name: Build packages
      run: |
        export VERSION=${{ steps.version.outputs.VERSION }}
        make packages
        
        # Build Arch package if possible
        if command -v makepkg >/dev/null 2>&1; then
          cd dist
          mkdir -p arch
          cat > arch/PKGBUILD << EOF
        pkgname=sysmedic-bin
        pkgver=${VERSION#v}
        pkgrel=1
        pkgdesc="Single binary multi-daemon Linux system monitoring tool"
        arch=('x86_64')
        url="https://github.com/ahur-system/sysmedic"
        license=('MIT')
        depends=('systemd')
        provides=('sysmedic')
        conflicts=('sysmedic')
        
        package() {
            install -Dm755 "../sysmedic" "\$pkgdir/usr/bin/sysmedic"
            install -Dm644 "../../scripts/sysmedic.doctor.service" "\$pkgdir/usr/lib/systemd/system/sysmedic.doctor.service"
            install -Dm644 "../../scripts/sysmedic.websocket.service" "\$pkgdir/usr/lib/systemd/system/sysmedic.websocket.service"
            install -Dm644 "../../scripts/config.example.yaml" "\$pkgdir/etc/sysmedic/config.yaml"
            install -dm755 "\$pkgdir/var/lib/sysmedic"
            install -dm755 "\$pkgdir/var/log/sysmedic"
        }
        EOF
          cd arch
          makepkg -sf --noconfirm || echo "Failed to build Arch package"
          mv *.pkg.tar.zst ../ 2>/dev/null || true
          cd ../..
        fi
    
    - name: Upload build artifacts
      uses: actions/upload-artifact@v3
      with:
        name: packages-${{ steps.version.outputs.VERSION }}
        path: dist/
        retention-days: 30
    
    - name: Create Release
      if: startsWith(github.ref, 'refs/tags/')
      uses: softprops/action-gh-release@v1
      with:
        files: |
          dist/*.deb
          dist/*.rpm
          dist/*.pkg.tar.zst
          dist/*.tar.gz
          dist/SHA256SUMS
        body_path: CHANGELOG.md
        draft: false
        prerelease: false
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}

  test:
    name: Test Packages
    needs: build
    runs-on: ubuntu-latest
    strategy:
      matrix:
        os: ['ubuntu:20.04', 'ubuntu:22.04', 'debian:11', 'debian:12']
    
    steps:
    - name: Download artifacts
      uses: actions/download-artifact@v3
      with:
        name: packages-${{ needs.build.outputs.VERSION }}
        path: dist/
    
    - name: Test DEB installation
      run: |
        docker run --rm -v $(pwd)/dist:/packages ${{ matrix.os }} bash -c "
          apt-get update
          apt-get install -y systemd
          dpkg -i /packages/*.deb || apt-get -f install -y
          sysmedic --version
          systemctl enable sysmedic.doctor sysmedic.websocket
        "
        
    - name: Test Arch installation
      if: matrix.os == 'archlinux:latest'
      run: |
        docker run --rm -v $(pwd)/dist:/packages archlinux:latest bash -c "
          pacman -Sy --noconfirm
          pacman -U --noconfirm /packages/*.pkg.tar.zst || echo 'No Arch package found'
          if command -v sysmedic >/dev/null; then
            sysmedic --version
            systemctl enable sysmedic.doctor sysmedic.websocket
          fi
        "

  security-scan:
    name: Security Scan
    runs-on: ubuntu-latest
    
    steps:
    - name: Checkout code
      uses: actions/checkout@v4
    
    - name: Run Gosec Security Scanner
      uses: securecodewarrior/github-action-gosec@master
      with:
        args: '-no-fail -fmt sarif -out results.sarif ./...'
    
    - name: Upload SARIF file
      uses: github/codeql-action/upload-sarif@v2
      with:
        sarif_file: results.sarif
```

#### Manual Release Process

```bash
# 1. Create and push tag
git tag v1.0.5
git push origin v1.0.5

# 2. Build packages
make packages

# 3. Create release on GitHub
gh release create v1.0.5 \
    --title "SysMedic v1.0.5 - Feature Release" \
    --notes-file CHANGELOG.md \
    dist/*.deb \
    dist/*.rpm \
    dist/*.pkg.tar.zst \
    dist/*.tar.gz \
    dist/SHA256SUMS

# Or via web interface:
# https://github.com/ahur-system/sysmedic/releases/new
```

## Installation Methods

### Package Manager Installation

#### Ubuntu/Debian (.deb)
```bash
# Method 1: Direct download and install
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic_1.0.5_amd64.deb
sudo dpkg -i sysmedic_1.0.5_amd64.deb
sudo apt-get -f install  # Fix dependencies if needed

# Method 2: One-liner with curl
curl -LO https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic_1.0.5_amd64.deb && sudo dpkg -i sysmedic_1.0.5_amd64.deb

# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket

# Verify installation
sysmedic --version
sysmedic daemon status
```

#### RHEL/CentOS/Fedora (.rpm)
```bash
# Method 1: Direct download and install
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-1.0.5-1.x86_64.rpm
sudo rpm -ivh sysmedic-1.0.5-1.x86_64.rpm

# Method 2: Using yum/dnf
sudo yum install https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-1.0.5-1.x86_64.rpm
# or
sudo dnf install https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-1.0.5-1.x86_64.rpm

# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket

# Verify installation
sysmedic --version
sysmedic daemon status
```

#### Arch Linux (.pkg.tar.zst)
```bash
# Method 1: Direct download and install (latest release)
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-arch.pkg.tar.zst
sudo pacman -U sysmedic-arch.pkg.tar.zst

# Method 2: Direct download and install (specific version)
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-1.0.5-1-x86_64.pkg.tar.zst
sudo pacman -U sysmedic-1.0.5-1-x86_64.pkg.tar.zst

# Verify package format (optional)
file sysmedic-arch.pkg.tar.zst
# Should show: Zstandard compressed data (v0.8+), Dictionary ID: None

# Method 3: Install from AUR (when available)
# Using yay AUR helper
yay -S sysmedic

# Using paru AUR helper  
paru -S sysmedic

# Method 4: Manual AUR build
git clone https://aur.archlinux.org/sysmedic.git
cd sysmedic
makepkg -si

# Enable and start services
sudo systemctl daemon-reload
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket

# Verify installation
sysmedic --version
sysmedic daemon status
```

### Generic Linux Installation

```bash
# Download and extract
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v1.0.5-linux-amd64.tar.gz
tar -xzf sysmedic-v1.0.5-linux-amd64.tar.gz
cd sysmedic-v1.0.5

# Install using provided script
sudo ./scripts/install.sh

# Manual installation
sudo cp sysmedic /usr/local/bin/
sudo chmod +x /usr/local/bin/sysmedic
sudo mkdir -p /etc/sysmedic /var/lib/sysmedic /var/log/sysmedic
sudo cp scripts/config.example.yaml /etc/sysmedic/config.yaml
sudo cp scripts/sysmedic.doctor.service /etc/systemd/system/
sudo cp scripts/sysmedic.websocket.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable sysmedic.doctor sysmedic.websocket
sudo systemctl start sysmedic.doctor sysmedic.websocket
```

### Automated Installation Scripts

#### Quick Install Script
```bash
#!/bin/bash
# scripts/install.sh - Quick installation script

set -e

# Configuration
GITHUB_REPO="ahur-system/sysmedic"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/sysmedic"
DATA_DIR="/var/lib/sysmedic"
LOG_DIR="/var/log/sysmedic"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root (use sudo)"
        exit 1
    fi
}

# Detect system architecture
detect_arch() {
    case $(uname -m) in
        x86_64) echo "amd64" ;;
        aarch64|arm64) echo "arm64" ;;
        *) 
            log_error "Unsupported architecture: $(uname -m)"
            exit 1
            ;;
    esac
}

# Detect Linux distribution
detect_distro() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        echo $ID
    else
        log_error "Cannot detect Linux distribution"
        exit 1
    fi
}

# Install via package manager if possible
install_package() {
    local distro=$1
    local arch=$2
    
    case $distro in
        ubuntu|debian)
            log_info "Installing DEB package..."
            local deb_url="https://github.com/${GITHUB_REPO}/releases/latest/download/sysmedic_${VERSION}_${arch}.deb"
            wget -q $deb_url -O /tmp/sysmedic.deb
            dpkg -i /tmp/sysmedic.deb || apt-get -f install -y
            rm /tmp/sysmedic.deb
            return 0
            ;;
        rhel|centos|fedora|opensuse*)
            log_info "Installing RPM package..."
            local rpm_url="https://github.com/${GITHUB_REPO}/releases/latest/download/sysmedic-${VERSION}-1.x86_64.rpm"
            if command -v yum >/dev/null; then
                yum install -y $rpm_url
            elif command -v dnf >/dev/null; then
                dnf install -y $rpm_url
            else
                wget -q $rpm_url -O /tmp/sysmedic.rpm
                rpm -ivh /tmp/sysmedic.rpm
                rm /tmp/sysmedic.rpm
            fi
            return 0
            ;;
        arch|manjaro)
            log_info "Installing Arch package..."
            local pkg_url="https://github.com/${GITHUB_REPO}/releases/latest/download/sysmedic-${VERSION}-1-x86_64.pkg.tar.zst"
            wget -q $pkg_url -O /tmp/sysmedic.pkg.tar.zst
            pacman -U --noconfirm /tmp/sysmedic.pkg.tar.zst
            rm /tmp/sysmedic.pkg.tar.zst
            return 0
            ;;
    esac
    return 1
}

# Generic installation
install_generic() {
    local arch=$1
    
    log_info "Installing via generic tarball..."
    
    # Download and extract
    local tarball_url="https://github.com/${GITHUB_REPO}/releases/latest/download/sysmedic-v${VERSION}-linux-${arch}.tar.gz"
    wget -q $tarball_url -O /tmp/sysmedic.tar.gz
    cd /tmp
    tar -xzf sysmedic.tar.gz
    
    # Install binary
    cp sysmedic-*/sysmedic ${INSTALL_DIR}/
    chmod +x ${INSTALL_DIR}/sysmedic
    
    # Create directories
    mkdir -p ${CONFIG_DIR} ${DATA_DIR} ${LOG_DIR}
    
    # Install configuration and services
    cp sysmedic-*/scripts/config.example.yaml ${CONFIG_DIR}/config.yaml
    cp sysmedic-*/scripts/sysmedic.doctor.service /etc/systemd/system/
    cp sysmedic-*/scripts/sysmedic.websocket.service /etc/systemd/system/
    
    # Set permissions
    chown -R root:root ${CONFIG_DIR} ${DATA_DIR} ${LOG_DIR}
    chmod 755 ${CONFIG_DIR} ${DATA_DIR} ${LOG_DIR}
    chmod 644 ${CONFIG_DIR}/config.yaml
    
    # Clean up
    rm -rf /tmp/sysmedic*
}

# Enable and start services
setup_services() {
    log_info "Setting up systemd services..."
    systemctl daemon-reload
    systemctl enable sysmedic.doctor sysmedic.websocket
    systemctl start sysmedic.doctor sysmedic.websocket
}

# Verify installation
verify_installation() {
    log_info "Verifying installation..."
    
    # Check binary
    if ! command -v sysmedic >/dev/null; then
        log_error "sysmedic binary not found in PATH"
        return 1
    fi
    
    # Check version
    local version=$(sysmedic --version 2>/dev/null | head -1 || echo "unknown")
    log_info "Installed version: $version"
    
    # Check services
    sleep 2  # Give services time to start
    
    if systemctl is-active --quiet sysmedic.doctor; then
        log_info "Doctor daemon: running"
    else
        log_warn "Doctor daemon: not running"
    fi
    
    if systemctl is-active --quiet sysmedic.websocket; then
        log_info "WebSocket daemon: running"
    else
        log_warn "WebSocket daemon: not running"
    fi
    
    # Display status
    echo
    log_info "Installation complete! Usage:"
    echo "  sysmedic               # Show dashboard"
    echo "  sysmedic daemon status # Check daemon status"
    echo "  sysmedic websocket status # Check WebSocket status"
    echo
    log_info "Configuration: ${CONFIG_DIR}/config.yaml"
    log_info "Data directory: ${DATA_DIR}"
    log_info "Log directory: ${LOG_DIR}"
}

# Main installation function
main() {
    log_info "SysMedic Installation Script"
    echo
    
    # Checks
    check_root
    
    # Get latest version
    VERSION=$(curl -s https://api.github.com/repos/${GITHUB_REPO}/releases/latest | grep -Po '"tag_name": "\K.*?(?=")' | sed 's/^v//')
    if [[ -z "$VERSION" ]]; then
        log_error "Failed to get latest version"
        exit 1
    fi
    
    log_info "Installing SysMedic v${VERSION}"
    
    # Detect system
    ARCH=$(detect_arch)
    DISTRO=$(detect_distro)
    
    log_info "Detected: $DISTRO on $ARCH"
    
    # Try package installation first, fall back to generic
    if ! install_package $DISTRO $ARCH; then
        install_generic $ARCH
    fi
    
    # Setup and verify
    setup_services
    verify_installation
}

# Run main function
main "$@"
```

#### One-liner Installation
```bash
# Quick install command
curl -sSL https://raw.githubusercontent.com/ahur-system/sysmedic/main/scripts/install.sh | sudo bash

# With specific version
curl -sSL https://raw.githubusercontent.com/ahur-system/sysmedic/main/scripts/install.sh | sudo VERSION=1.0.5 bash

# Preview install script before running
curl -s https://raw.githubusercontent.com/ahur-system/sysmedic/main/scripts/install.sh | less
```

### Container Installation

#### Docker
```bash
# Pull official image
docker pull ghcr.io/ahur-system/sysmedic:latest

# Run with host monitoring
docker run -d \
    --name sysmedic \
    --pid host \
    --network host \
    -v /proc:/host/proc:ro \
    -v /sys:/host/sys:ro \
    -v /var/lib/sysmedic:/var/lib/sysmedic \
    ghcr.io/ahur-system/sysmedic:latest

# Run specific services
docker run -d \
    --name sysmedic-doctor \
    --pid host \
    -v /proc:/host/proc:ro \
    -v /sys:/host/sys:ro \
    -v /var/lib/sysmedic:/var/lib/sysmedic \
    ghcr.io/ahur-system/sysmedic:latest \
    sysmedic --doctor-daemon

docker run -d \
    --name sysmedic-websocket \
    --pid host \
    -p 8060:8060 \
    -v /var/lib/sysmedic:/var/lib/sysmedic \
    ghcr.io/ahur-system/sysmedic:latest \
    sysmedic --websocket-daemon
```

#### Docker Compose
```yaml
# docker-compose.yml
version: '3.8'

services:
  sysmedic-doctor:
    image: ghcr.io/ahur-system/sysmedic:latest
    container_name: sysmedic-doctor
    command: sysmedic --doctor-daemon
    pid: host
    volumes:
      - /proc:/host/proc:ro
      - /sys:/host/sys:ro
      - sysmedic-data:/var/lib/sysmedic
    restart: unless-stopped

  sysmedic-websocket:
    image: ghcr.io/ahur-system/sysmedic:latest
    container_name: sysmedic-websocket
    command: sysmedic --websocket-daemon
    pid: host
    ports:
      - "8060:8060"
    volumes:
      - sysmedic-data:/var/lib/sysmedic
    restart: unless-stopped
    depends_on:
      - sysmedic-doctor

volumes:
  sysmedic-data:
```

### AUR (Arch User Repository) Support

#### Publishing to AUR
```bash
# Clone AUR repository (requires AUR account)
git clone ssh://aur@aur.archlinux.org/sysmedic.git aur-sysmedic
cd aur-sysmedic

# Create/Update PKGBUILD
cat > PKGBUILD << 'EOF'
# Maintainer: SysMedic Team <team@sysmedic.dev>
pkgname=sysmedic
pkgver=1.0.5
pkgrel=1
pkgdesc="Single binary multi-daemon Linux system monitoring tool with user-centric resource tracking"
arch=('x86_64' 'aarch64')
url="https://github.com/ahur-system/sysmedic"
license=('MIT')
depends=('systemd')
makedepends=('go' 'git')
backup=('etc/sysmedic/config.yaml')
source=("$pkgname-$pkgver.tar.gz::https://github.com/ahur-system/sysmedic/archive/v$pkgver.tar.gz")
sha256sums=('SKIP')  # Update with actual checksums

build() {
    cd "$pkgname-$pkgver"
    export CGO_CPPFLAGS="${CPPFLAGS}"
    export CGO_CFLAGS="${CFLAGS}"
    export CGO_CXXFLAGS="${CXXFLAGS}"
    export CGO_LDFLAGS="${LDFLAGS}"
    export GOFLAGS="-buildmode=pie -trimpath -ldflags=-linkmode=external -mod=readonly -modcacherw"
    
    go build -ldflags="-s -w -X main.Version=${pkgver}" -o sysmedic cmd/sysmedic/main.go
}

check() {
    cd "$pkgname-$pkgver"
    # Add tests when available
    # go test -v ./...
}

package() {
    cd "$pkgname-$pkgver"
    
    # Install binary
    install -Dm755 sysmedic "$pkgdir/usr/bin/sysmedic"
    
    # Install systemd services
    install -Dm644 scripts/sysmedic.doctor.service "$pkgdir/usr/lib/systemd/system/sysmedic.doctor.service"
    install -Dm644 scripts/sysmedic.websocket.service "$pkgdir/usr/lib/systemd/system/sysmedic.websocket.service"
    
    # Install configuration
    install -Dm644 scripts/config.example.yaml "$pkgdir/etc/sysmedic/config.yaml"
    
    # Install documentation
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    install -Dm644 CHANGELOG.md "$pkgdir/usr/share/doc/$pkgname/CHANGELOG.md"
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    
    # Create data and log directories
    install -dm755 "$pkgdir/var/lib/sysmedic"
    install -dm755 "$pkgdir/var/log/sysmedic"
}
EOF

# Create .SRCINFO
makepkg --printsrcinfo > .SRCINFO

# Commit and push to AUR
git add PKGBUILD .SRCINFO
git commit -m "Update to version $pkgver"
git push origin master
```

#### AUR Binary Package
```bash
# For users who prefer pre-built binaries, create sysmedic-bin package
cat > PKGBUILD << 'EOF'
# Maintainer: SysMedic Team <team@sysmedic.dev>
pkgname=sysmedic-bin
pkgver=1.0.5
pkgrel=1
pkgdesc="Single binary multi-daemon Linux system monitoring tool (binary release)"
arch=('x86_64' 'aarch64')
url="https://github.com/ahur-system/sysmedic"
license=('MIT')
depends=('systemd')
provides=('sysmedic')
conflicts=('sysmedic')
backup=('etc/sysmedic/config.yaml')
source_x86_64=("https://github.com/ahur-system/sysmedic/releases/download/v$pkgver/sysmedic-v$pkgver-linux-amd64.tar.gz")
source_aarch64=("https://github.com/ahur-system/sysmedic/releases/download/v$pkgver/sysmedic-v$pkgver-linux-arm64.tar.gz")
sha256sums_x86_64=('SKIP')
sha256sums_aarch64=('SKIP')

package() {
    # Extract and install binary
    install -Dm755 sysmedic "$pkgdir/usr/bin/sysmedic"
    
    # Install systemd services
    install -Dm644 scripts/sysmedic.doctor.service "$pkgdir/usr/lib/systemd/system/sysmedic.doctor.service"
    install -Dm644 scripts/sysmedic.websocket.service "$pkgdir/usr/lib/systemd/system/sysmedic.websocket.service"
    
    # Install configuration
    install -Dm644 scripts/config.example.yaml "$pkgdir/etc/sysmedic/config.yaml"
    
    # Install documentation
    install -Dm644 README.md "$pkgdir/usr/share/doc/$pkgname/README.md"
    install -Dm644 CHANGELOG.md "$pkgdir/usr/share/doc/$pkgname/CHANGELOG.md" 2>/dev/null || true
    install -Dm644 LICENSE "$pkgdir/usr/share/licenses/$pkgname/LICENSE"
    
    # Create data and log directories
    install -dm755 "$pkgdir/var/lib/sysmedic"
    install -dm755 "$pkgdir/var/log/sysmedic"
}
EOF
```

### Package Verification

#### Checksum Verification
```bash
# Download checksums
wget https://github.com/ahur-system/sysmedic/releases/latest/download/SHA256SUMS

# Verify specific package
sha256sum -c SHA256SUMS --ignore-missing

# Manual verification
sha256sum sysmedic_1.0.5_amd64.deb
# Compare with value in SHA256SUMS file

# Verify Arch package
sha256sum sysmedic-1.0.5-1-x86_64.pkg.tar.zst
```

#### GPG Signature Verification (Future)
```bash
# Import GPG key (when available)
curl -s https://github.com/ahur-system/sysmedic/releases/download/gpg-key/sysmedic.asc | gpg --import

# Verify signature
gpg --verify sysmedic_1.0.5_amd64.deb.sig sysmedic_1.0.5_amd64.deb
```

## Release Management

### Version Numbering

SysMedic follows [Semantic Versioning](https://semver.org/) (SemVer):

- **MAJOR.MINOR.PATCH** (e.g., 1.0.5)
- **MAJOR**: Breaking changes to API or core functionality
- **MINOR**: New features, backward compatible
- **PATCH**: Bug fixes, backward compatible
- **Pre-release**: alpha, beta, rc suffixes (e.g., 1.1.0-alpha.1)

### Release Process

#### 1. Pre-Release Preparation
```bash
# Update version in relevant files
VERSION="1.0.6"

# Update CHANGELOG.md
echo "## [$VERSION] - $(date +%Y-%m-%d)" >> CHANGELOG.md
echo "" >> CHANGELOG.md
echo "### Added" >> CHANGELOG.md
echo "### Changed" >> CHANGELOG.md
echo "### Fixed" >> CHANGELOG.md

# Update version in Go files (if using ldflags injection)
# No changes needed as version is injected during build
```

#### 2. Tag and Release
```bash
# Create and push tag
git add CHANGELOG.md
git commit -m "Release v$VERSION"
git tag -a "v$VERSION" -m "Release version $VERSION"
git push origin main --tags

# GitHub Actions will automatically create the release
```

#### 3. Post-Release Tasks
```bash
# Update AUR packages
cd aur-sysmedic
# Update PKGBUILD with new version and checksums
makepkg --printsrcinfo > .SRCINFO
git add PKGBUILD .SRCINFO
git commit -m "Update to version $VERSION"
git push

# Update documentation
# Update homebrew formula (if applicable)
# Notify package maintainers
```

## Distribution Channels

### Primary Distribution
- **GitHub Releases**: Official source for all packages
- **Package Repositories**: Distribution-specific packages
- **Container Registries**: Docker images on GitHub Container Registry

### Package Repositories

#### Debian/Ubuntu PPA (Future)
```bash
# Add PPA repository
sudo add-apt-repository ppa:sysmedic/stable
sudo apt update
sudo apt install sysmedic
```

#### RPM Repository (Future)
```bash
# Add RPM repository
sudo tee /etc/yum.repos.d/sysmedic.repo << EOF
[sysmedic]
name=SysMedic Repository
baseurl=https://repo.sysmedic.dev/rpm
enabled=1
gpgcheck=1
gpgkey=https://repo.sysmedic.dev/gpg-key/sysmedic.asc
EOF

sudo yum install sysmedic
```

#### Arch User Repository (AUR)
```bash
# Available packages
yay -S sysmedic        # Build from source
yay -S sysmedic-bin    # Pre-built binary
yay -S sysmedic-git    # Latest development version
```

## Automated Deployment

### CI/CD Pipeline Features
- **Multi-platform builds**: Linux x64, ARM64
- **Package creation**: DEB, RPM, Arch, tar.gz
- **Security scanning**: Vulnerability assessment
- **Automated testing**: Package installation verification
- **Release automation**: GitHub releases with assets
- **Container publishing**: Multi-arch Docker images

### Deployment Environments
- **Development**: Feature branches, PR builds
- **Staging**: Release candidate testing
- **Production**: Tagged releases, stable packages

## Troubleshooting

### Build Issues

#### Go Environment Problems
```bash
# Issue: Go version too old
go version  # Check current version
# Solution: Install Go 1.21+
wget https://golang.org/dl/go1.21.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.3.linux-amd64.tar.gz

# Issue: Module download failures
export GOPROXY=https://proxy.golang.org,direct
export GOSUMDB=sum.golang.org
go mod download

# Issue: CGO compilation errors
export CGO_ENABLED=1
sudo apt-get install build-essential  # Debian/Ubuntu
sudo yum groupinstall "Development Tools"  # RHEL/CentOS
```

#### Package Building Problems
```bash
# Issue: FPM not found
gem install fpm
# Or on system with restricted gem installation:
sudo apt-get install ruby-dev
sudo gem install fpm

# Issue: RPM building fails on Debian/Ubuntu
sudo apt-get install rpm

# Issue: Missing dependencies
sudo apt-get install build-essential git curl wget
```

#### Arch Linux Specific Issues
```bash
# Issue: makepkg fails with permissions
chmod +x PKGBUILD
makepkg -s --noconfirm

# Issue: Missing base-devel
sudo pacman -S --needed base-devel

# Issue: Go module verification fails
export GOPROXY=direct
export GOSUMDB=off

# Issue: Package installation conflicts
sudo pacman -R sysmedic-bin  # Remove conflicting package
sudo pacman -S sysmedic

# Issue: AUR package build fails
# Check .SRCINFO is up to date
makepkg --printsrcinfo > .SRCINFO

# Issue: systemd service files not found
sudo systemctl daemon-reload
sudo systemctl enable sysmedic.doctor sysmedic.websocket
```

### Installation Issues

#### Permission Problems
```bash
# Issue: Cannot create directories
sudo mkdir -p /etc/sysmedic /var/lib/sysmedic /var/log/sysmedic
sudo chown -R root:root /etc/sysmedic /var/lib/sysmedic /var/log/sysmedic

# Issue: Binary not executable
sudo chmod +x /usr/local/bin/sysmedic
# or
sudo chmod +x /usr/bin/sysmedic
```

#### Service Issues
```bash
# Issue: Services fail to start
sudo systemctl status sysmedic.doctor
sudo systemctl status sysmedic.websocket
sudo journalctl -u sysmedic.doctor -f

# Issue: Port conflicts
sudo netstat -tlnp | grep :8060
# Change port in /etc/sysmedic/config.yaml

# Issue: Missing systemd services
sudo systemctl daemon-reload
ls -la /etc/systemd/system/sysmedic.*
```

#### Arch Linux Installation Issues
```bash
# Issue: Package not found
pacman -Ss sysmedic
yay -Ss sysmedic

# Issue: Dependency conflicts
sudo pacman -S --needed systemd
sudo pacman -Syu  # Update system first

# Issue: AUR helper problems
# Install yay if not available
sudo pacman -S --needed git base-devel
git clone https://aur.archlinux.org/yay.git
cd yay
makepkg -si

# Issue: Package signature verification
sudo pacman -S archlinux-keyring
sudo pacman-key --refresh-keys
```

### Runtime Issues

#### Configuration Problems
```bash
# Issue: Config file not found
ls -la /etc/sysmedic/config.yaml
# Copy from example if missing
sudo cp /usr/share/doc/sysmedic/config.example.yaml /etc/sysmedic/config.yaml

# Issue: Invalid YAML syntax
sysmedic --config /etc/sysmedic/config.yaml --validate
```

#### Performance Issues
```bash
# Issue: High memory usage
# Check memory limits in systemd service
sudo systemctl edit sysmedic.doctor
# Add:
# [Service]
# MemoryLimit=512M

# Issue: High CPU usage
# Check monitoring intervals in config
sudo nano /etc/sysmedic/config.yaml
```

#### Network Issues
```bash
# Issue: WebSocket connection fails
sudo firewall-cmd --add-port=8060/tcp --permanent  # RHEL/CentOS
sudo ufw allow 8060/tcp  # Ubuntu
sudo iptables -A INPUT -p tcp --dport 8060 -j ACCEPT  # Generic

# Issue: Cannot bind to port
sudo lsof -i :8060
# Kill conflicting process or change port
```

### Getting Help

#### Log Files
```bash
# System logs
sudo journalctl -u sysmedic.doctor -f
sudo journalctl -u sysmedic.websocket -f

# Application logs
sudo tail -f /var/log/sysmedic/sysmedic.log
sudo tail -f /var/log/sysmedic/error.log
```

#### Debug Mode
```bash
# Enable debug logging
sysmedic --debug --doctor-daemon
sysmedic --debug --websocket-daemon

# Or edit config file
sudo nano /etc/sysmedic/config.yaml
# Set log_level: debug
```

#### Support Channels
- **GitHub Issues**: https://github.com/ahur-system/sysmedic/issues
- **Documentation**: https://docs.sysmedic.dev
- **Community**: GitHub Discussions
- **Security Issues**: security@sysmedic.dev

#### Diagnostic Information
```bash
# Collect system information for bug reports
sysmedic --version
uname -a
cat /etc/os-release
systemctl status sysmedic.doctor sysmedic.websocket
cat /etc/sysmedic/config.yaml
sudo journalctl -u sysmedic.doctor --since "1 hour ago" --no-pager
```

---

**End of Documentation**

For the latest updates and additional information, visit:
- **Project Homepage**: https://github.com/ahur-system/sysmedic
- **Documentation**: https://docs.sysmedic.dev
- **Releases**: https://github.com/ahur-system/sysmedic/releases