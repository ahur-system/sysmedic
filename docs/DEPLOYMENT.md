# SysMedic Deployment & Publishing Guide

## Overview

This guide covers the complete deployment and publishing process for SysMedic, including package building, GitHub publishing, and distribution across different Linux platforms.

## Table of Contents

1. [Package Building](#package-building)
2. [GitHub Publishing](#github-publishing)
3. [Installation Methods](#installation-methods)
4. [Distribution](#distribution)
5. [Release Management](#release-management)
6. [Troubleshooting](#troubleshooting)

## Package Building

### Prerequisites

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install golang-go build-essential fpm rpm

# RHEL/CentOS
sudo yum install golang gcc rpm-build
gem install fpm
```

### Build Process

```bash
# Clone and build
git clone https://github.com/ahur-system/sysmedic.git
cd sysmedic

# Build binary
make build

# Build all packages
make packages
```

### Package Outputs

The build process creates the following packages:

| Package Type | Filename | Size | Target Distribution |
|--------------|----------|------|-------------------|
| **Debian/Ubuntu** | `sysmedic-amd64.deb` | 2.1MB | Debian, Ubuntu, derivatives |
| **RHEL/CentOS** | `sysmedic-x86_64.rpm` | 2.4MB | RHEL, CentOS, Fedora, SUSE |
| **Generic Linux** | `sysmedic-v1.0.x-linux-amd64.tar.gz` | 2.4MB | Any Linux x86_64 |
| **Checksums** | `SHA256SUMS` | - | Verification file |

### Manual Package Building

#### DEB Package
```bash
fpm -s dir -t deb -n sysmedic -v 1.0.3 \
    --description "Cross-platform Linux server monitoring CLI tool" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --license "MIT" \
    --depends systemd \
    --after-install scripts/install.sh \
    --before-remove scripts/uninstall.sh \
    sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.service=/etc/systemd/system/sysmedic.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml
```

#### RPM Package
```bash
fpm -s dir -t rpm -n sysmedic -v 1.0.3 \
    --description "Cross-platform Linux server monitoring CLI tool" \
    --url "https://github.com/ahur-system/sysmedic" \
    --maintainer "SysMedic Team <team@sysmedic.dev>" \
    --license "MIT" \
    --depends systemd \
    --after-install scripts/install.sh \
    --before-remove scripts/uninstall.sh \
    sysmedic=/usr/local/bin/sysmedic \
    scripts/sysmedic.service=/etc/systemd/system/sysmedic.service \
    scripts/config.example.yaml=/etc/sysmedic/config.yaml
```

## GitHub Publishing

### Authentication Setup

Since GitHub no longer supports password authentication, you need to use either a Personal Access Token (PAT) or SSH keys.

#### Option 1: Personal Access Token (Recommended)

1. **Create a Personal Access Token:**
   - Go to GitHub.com → Settings → Developer settings → Personal access tokens → Tokens (classic)
   - Click "Generate new token (classic)"
   - Select scopes: `repo` (full repository access)
   - Copy the token (save it securely - you won't see it again!)

2. **Configure Git with Token:**
   ```bash
   # Set your GitHub credentials
   git config --global user.name "ahur-system"
   git config --global user.email "your-email@example.com"
   
   # When prompted for password, use your PAT instead of password
   ```

#### Option 2: SSH Keys

1. **Generate SSH Key:**
   ```bash
   ssh-keygen -t ed25519 -C "your-email@example.com"
   cat ~/.ssh/id_ed25519.pub  # Copy this to GitHub
   ```

2. **Add to GitHub:**
   - Go to GitHub.com → Settings → SSH and GPG keys
   - Click "New SSH key" and paste your public key

### Repository Publishing

1. **Create GitHub Repository:**
   - Go to https://github.com/new
   - Repository name: `sysmedic`
   - Description: "Cross-platform Linux server monitoring CLI tool with user-centric resource tracking"
   - Set to Public
   - Don't initialize with README (we have our own)

2. **Push to GitHub:**
   ```bash
   cd /path/to/sysmedic
   
   # Add remote origin
   git remote add origin https://github.com/ahur-system/sysmedic.git
   
   # Push main branch
   git push -u origin main
   
   # Push tags
   git push origin --tags
   ```

### Release Creation

#### Automated Release (GitHub Actions)

Create `.github/workflows/release.yml`:

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.18
    
    - name: Install dependencies
      run: |
        sudo apt-get update
        sudo apt-get install -y rpm
        gem install fpm
    
    - name: Build packages
      run: make packages
    
    - name: Create Release
      uses: actions/create-release@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        tag_name: ${{ github.ref }}
        release_name: Release ${{ github.ref }}
        draft: false
        prerelease: false
    
    - name: Upload Release Assets
      uses: actions/upload-release-asset@v1
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
      with:
        upload_url: ${{ steps.create_release.outputs.upload_url }}
        asset_path: ./dist/
        asset_name: sysmedic-packages
        asset_content_type: application/zip
```

#### Manual Release

1. **Create Release on GitHub:**
   - Go to repository → Releases → "Create a new release"
   - Tag version: `v1.0.3`
   - Release title: `SysMedic v1.0.3 - Alert Management`
   - Description: Copy from CHANGELOG.md

2. **Upload Package Assets:**
   - Drag and drop the following files:
     - `sysmedic-amd64.deb`
     - `sysmedic-x86_64.rpm`
     - `sysmedic-v1.0.3-linux-amd64.tar.gz`
     - `SHA256SUMS`

3. **Publish Release:**
   - Click "Publish release"

## Installation Methods

### Package Manager Installation

#### Ubuntu/Debian (.deb)
```bash
# Download and install
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
sudo dpkg -i sysmedic-amd64.deb

# Enable and start service
sudo systemctl enable sysmedic
sudo systemctl start sysmedic

# Verify installation
sysmedic --version
```

#### RHEL/CentOS/Fedora (.rpm)
```bash
# Download and install
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
sudo rpm -ivh sysmedic-x86_64.rpm

# Enable and start service
sudo systemctl enable sysmedic
sudo systemctl start sysmedic

# Verify installation
sysmedic --version
```

### Generic Linux Installation

```bash
# Download and extract
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v1.0.3-linux-amd64.tar.gz
tar -xzf sysmedic-v1.0.3-linux-amd64.tar.gz
cd sysmedic-v1.0.3

# Run install script
sudo ./scripts/install.sh

# Or manual installation
sudo cp sysmedic /usr/local/bin/
sudo cp scripts/sysmedic.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable sysmedic
sudo systemctl start sysmedic
```

### Quick Install Script

```bash
# One-liner installation
curl -sSL https://raw.githubusercontent.com/ahur-system/sysmedic/main/scripts/install.sh | sudo bash
```

### Package Verification

```bash
# Download checksums
wget https://github.com/ahur-system/sysmedic/releases/latest/download/SHA256SUMS

# Verify package integrity
sha256sum -c SHA256SUMS --ignore-missing
```

## Distribution

### Download Statistics

Monitor download statistics through:
- GitHub Insights → Traffic
- GitHub Releases page
- Third-party tools like [GitHub Release Stats](https://github.com/seladb/github-release-stats)

### Distribution Channels

#### Primary Distribution
- **GitHub Releases**: Main distribution channel
- **Direct Downloads**: Latest release links
- **Package Managers**: Future APT/YUM repository

#### Secondary Distribution
- **Docker Hub**: Container images (planned)
- **Homebrew**: macOS support (planned)
- **Snap Store**: Universal Linux packages (planned)

### Download Links

#### Latest Release
- **Main**: https://github.com/ahur-system/sysmedic/releases/latest

#### Direct Downloads
- **DEB**: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
- **RPM**: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-x86_64.rpm
- **TAR.GZ**: https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-v1.0.3-linux-amd64.tar.gz
- **Checksums**: https://github.com/ahur-system/sysmedic/releases/latest/download/SHA256SUMS

## Release Management

### Version Numbering

SysMedic follows [Semantic Versioning](https://semver.org/):
- **MAJOR**: Incompatible API changes
- **MINOR**: Backward-compatible functionality additions
- **PATCH**: Backward-compatible bug fixes

Examples:
- `v1.0.0`: Initial production release
- `v1.0.3`: Bug fixes and new features
- `v1.1.0`: New features with backward compatibility
- `v2.0.0`: Breaking changes

### Release Process

1. **Development**
   - Feature development on feature branches
   - Merge to `main` branch
   - Update CHANGELOG.md

2. **Pre-Release**
   - Run full test suite: `make test`
   - Update version numbers
   - Build and test packages locally

3. **Release**
   - Create git tag: `git tag v1.0.x`
   - Push tag: `git push origin v1.0.x`
   - Create GitHub release
   - Upload package assets

4. **Post-Release**
   - Update documentation
   - Announce release
   - Monitor for issues

### Release Checklist

#### Pre-Release
- [ ] All tests passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version numbers updated
- [ ] Packages build successfully
- [ ] Manual testing completed

#### Release
- [ ] Git tag created and pushed
- [ ] GitHub release created
- [ ] Package assets uploaded
- [ ] Checksums generated and uploaded
- [ ] Release notes published

#### Post-Release
- [ ] Installation tested on multiple platforms
- [ ] Documentation deployed
- [ ] Community notified
- [ ] Issues monitoring set up

## Troubleshooting

### Build Issues

#### Missing Dependencies
```bash
# Ubuntu/Debian
sudo apt install build-essential golang-go fpm rpm

# RHEL/CentOS
sudo yum groupinstall "Development Tools"
sudo yum install golang rpm-build
gem install fpm
```

#### Go Version Issues
```bash
# Check Go version
go version

# Update Go if needed (replace 1.19.3 with latest)
wget https://golang.org/dl/go1.19.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.19.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
```

### Package Installation Issues

#### Permission Denied
```bash
# Make sure you're using sudo
sudo dpkg -i sysmedic-amd64.deb
sudo rpm -ivh sysmedic-x86_64.rpm
```

#### Dependency Issues
```bash
# Ubuntu/Debian - fix dependencies
sudo apt --fix-broken install

# RHEL/CentOS - check dependencies
rpm -qpR sysmedic-x86_64.rpm
```

#### systemd Service Issues
```bash
# Check service status
sudo systemctl status sysmedic

# Check logs
sudo journalctl -u sysmedic -f

# Reload systemd if needed
sudo systemctl daemon-reload
```

### GitHub Issues

#### Authentication Failures
```bash
# Verify credentials
git config --global user.name
git config --global user.email

# Test SSH connection
ssh -T git@github.com

# Or test HTTPS with token
git remote -v
```

#### Push Failures
```bash
# Check remote URL
git remote -v

# Update remote URL if needed
git remote set-url origin https://github.com/ahur-system/sysmedic.git

# Force push if necessary (be careful!)
git push --force-with-lease origin main
```

### Package Verification Failures

#### Checksum Mismatches
```bash
# Re-download the file
rm sysmedic-*.deb
wget https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb

# Verify again
sha256sum sysmedic-amd64.deb
```

#### Corrupted Downloads
```bash
# Download with retry
wget --retry-connrefused --waitretry=1 --read-timeout=20 --timeout=15 -t 5 \
  https://github.com/ahur-system/sysmedic/releases/latest/download/sysmedic-amd64.deb
```

## Best Practices

### Security
- Always verify package checksums
- Use HTTPS for downloads
- Keep GPG keys secure
- Regular security updates

### Quality Assurance
- Test on multiple distributions
- Automated testing in CI/CD
- Community feedback integration
- Regular dependency updates

### Documentation
- Keep README.md updated
- Maintain comprehensive CHANGELOG.md
- Document breaking changes clearly
- Provide migration guides

### Community
- Respond to issues promptly
- Accept community contributions
- Maintain coding standards
- Regular communication

---

## Summary

This deployment guide provides comprehensive instructions for:

✅ **Building packages** for multiple Linux distributions  
✅ **Publishing to GitHub** with proper authentication  
✅ **Creating releases** with automated and manual processes  
✅ **Installing packages** across different platforms  
✅ **Managing releases** with semantic versioning  
✅ **Troubleshooting** common deployment issues  

For additional support or questions about deployment, please refer to the [GitHub Issues](https://github.com/ahur-system/sysmedic/issues) page.