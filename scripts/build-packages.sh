#!/bin/bash

# SysMedic Package Builder
# Builds .deb and .rpm packages for distribution

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
VERSION="1.0.1"
ARCH="amd64"
RPM_ARCH="x86_64"

echo "🏗️  Building SysMedic packages..."
echo "Version: $VERSION"
echo "Architecture: $ARCH"
echo "Project root: $PROJECT_ROOT"

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

# Check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."

    # Check for dpkg-deb (for .deb packages)
    if ! command -v dpkg-deb &> /dev/null; then
        print_error "dpkg-deb is required but not installed. Install with: sudo apt install dpkg-dev"
        exit 1
    fi

    # Check for rpmbuild (for .rpm packages)
    if ! command -v rpmbuild &> /dev/null; then
        print_warning "rpmbuild not found. RPM package will be skipped."
        print_warning "Install with: sudo apt install rpm (Ubuntu/Debian) or sudo yum install rpm-build (RHEL/CentOS)"
        SKIP_RPM=true
    else
        SKIP_RPM=false
    fi

    # Check for Go
    if ! command -v go &> /dev/null; then
        print_error "Go is required but not installed."
        exit 1
    fi

    print_success "Dependencies checked"
}

# Build the binary
build_binary() {
    print_status "Building SysMedic binary..."

    cd "$PROJECT_ROOT"

    # Clean previous builds
    rm -f sysmedic

    # Build for Linux AMD64
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$VERSION" -o sysmedic ./cmd/sysmedic

    if [ ! -f "sysmedic" ]; then
        print_error "Failed to build binary"
        exit 1
    fi

    # Make it executable
    chmod +x sysmedic

    print_success "Binary built successfully"
}

# Create Debian package
build_deb() {
    print_status "Building Debian package..."

    cd "$PROJECT_ROOT"

    # Create working directory
    DEB_DIR="packaging/deb-build"
    rm -rf "$DEB_DIR"
    mkdir -p "$DEB_DIR"

    # Copy package structure
    cp -r packaging/deb/* "$DEB_DIR/"

    # Copy binary
    cp sysmedic "$DEB_DIR/usr/local/bin/"
    chmod +x "$DEB_DIR/usr/local/bin/sysmedic"

    # Set correct permissions for package scripts
    chmod 755 "$DEB_DIR/DEBIAN/postinst"
    chmod 755 "$DEB_DIR/DEBIAN/prerm"
    chmod 755 "$DEB_DIR/DEBIAN/postrm"

    # Build the package
    PACKAGE_NAME="sysmedic-${VERSION}-${ARCH}.deb"
    dpkg-deb --build "$DEB_DIR" "dist/$PACKAGE_NAME"

    if [ -f "dist/$PACKAGE_NAME" ]; then
        print_success "Debian package created: dist/$PACKAGE_NAME"

        # Verify package
        dpkg-deb --info "dist/$PACKAGE_NAME"
        dpkg-deb --contents "dist/$PACKAGE_NAME"
    else
        print_error "Failed to create Debian package"
        exit 1
    fi

    # Cleanup
    rm -rf "$DEB_DIR"
}

# Create RPM package
build_rpm() {
    if [ "$SKIP_RPM" = true ]; then
        print_warning "Skipping RPM build (rpmbuild not available)"
        return
    fi

    print_status "Building RPM package..."

    cd "$PROJECT_ROOT"

    # Create RPM build directories
    RPM_ROOT="packaging/rpm-build"
    rm -rf "$RPM_ROOT"
    mkdir -p "$RPM_ROOT"/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

    # Create source tarball
    TAR_DIR="$RPM_ROOT/sysmedic-$VERSION"
    mkdir -p "$TAR_DIR"

    # Copy binary and files
    cp sysmedic "$TAR_DIR/"
    cp README.md "$TAR_DIR/"
    cp LICENSE "$TAR_DIR/"

    # Create tarball
    cd "$RPM_ROOT"
    tar czf "SOURCES/sysmedic-$VERSION.tar.gz" "sysmedic-$VERSION/"

    # Copy spec file
    cp "../rpm/sysmedic.spec" "SPECS/"

    # Build RPM
    rpmbuild --define "_topdir $(pwd)" -ba "SPECS/sysmedic.spec"

    # Copy built RPM to dist
    cd "$PROJECT_ROOT"
    RPM_FILE=$(find "$RPM_ROOT/RPMS" -name "*.rpm" | head -1)
    if [ -n "$RPM_FILE" ]; then
        PACKAGE_NAME="sysmedic-${VERSION}-1.${RPM_ARCH}.rpm"
        cp "$RPM_FILE" "dist/$PACKAGE_NAME"
        print_success "RPM package created: dist/$PACKAGE_NAME"

        # Verify package
        rpm -qip "dist/$PACKAGE_NAME"
    else
        print_error "Failed to create RPM package"
        exit 1
    fi

    # Cleanup
    rm -rf "$RPM_ROOT"
}

# Create generic tarball (fallback)
build_tarball() {
    print_status "Building generic tarball..."

    cd "$PROJECT_ROOT"

    # Create tarball directory
    TAR_DIR="dist/sysmedic-v$VERSION-linux-$ARCH"
    rm -rf "$TAR_DIR"
    mkdir -p "$TAR_DIR/scripts"

    # Copy files
    cp sysmedic "$TAR_DIR/"
    cp README.md "$TAR_DIR/"
    cp LICENSE "$TAR_DIR/"
    cp scripts/install.sh "$TAR_DIR/scripts/"
    cp scripts/config.example.yaml "$TAR_DIR/scripts/config.yaml"
    cp scripts/sysmedic.service "$TAR_DIR/scripts/"

    # Create demo script
    cat > "$TAR_DIR/scripts/demo.sh" << 'EOF'
#!/bin/bash
echo "Starting SysMedic demo..."
./sysmedic version
echo "Demo complete. Install with: sudo ./scripts/install.sh"
EOF
    chmod +x "$TAR_DIR/scripts/demo.sh"

    # Create tarball
    cd dist
    tar czf "sysmedic-v$VERSION-linux-$ARCH.tar.gz" "sysmedic-v$VERSION-linux-$ARCH/"

    print_success "Tarball created: dist/sysmedic-v$VERSION-linux-$ARCH.tar.gz"

    # Cleanup
    rm -rf "sysmedic-v$VERSION-linux-$ARCH"
}

# Create checksums
create_checksums() {
    print_status "Creating checksums..."

    cd "$PROJECT_ROOT/dist"

    # Create SHA256 checksums
    sha256sum *.deb *.rpm *.tar.gz > SHA256SUMS 2>/dev/null || true

    if [ -f "SHA256SUMS" ]; then
        print_success "Checksums created: dist/SHA256SUMS"
        cat SHA256SUMS
    fi
}

# Validate packages
validate_packages() {
    print_status "Validating packages..."

    cd "$PROJECT_ROOT/dist"

    # Check if files exist
    for package in *.deb *.rpm *.tar.gz; do
        if [ -f "$package" ]; then
            size=$(du -h "$package" | cut -f1)
            print_success "✓ $package ($size)"
        fi
    done

    # Test DEB package (if available)
    DEB_FILE=$(ls *.deb 2>/dev/null | head -1)
    if [ -n "$DEB_FILE" ]; then
        print_status "Testing Debian package structure..."
        dpkg-deb --contents "$DEB_FILE" | head -10
    fi

    # Test RPM package (if available)
    RPM_FILE=$(ls *.rpm 2>/dev/null | head -1)
    if [ -n "$RPM_FILE" ] && command -v rpm &> /dev/null; then
        print_status "Testing RPM package structure..."
        rpm -qlp "$RPM_FILE" | head -10
    fi
}

# Main execution
main() {
    print_status "Starting SysMedic package build process..."

    # Ensure we're in the right directory
    cd "$PROJECT_ROOT"

    # Create dist directory
    mkdir -p dist
    rm -f dist/*

    # Execute build steps
    check_dependencies
    build_binary
    build_deb
    build_rpm
    build_tarball
    create_checksums
    validate_packages

    print_success "🎉 Package build complete!"
    echo ""
    echo "📦 Built packages:"
    ls -la dist/
    echo ""
    echo "🚀 Ready for release!"
}

# Handle script arguments
case "${1:-}" in
    "deb")
        check_dependencies
        build_binary
        mkdir -p dist
        build_deb
        ;;
    "rpm")
        check_dependencies
        build_binary
        mkdir -p dist
        build_rpm
        ;;
    "tarball")
        check_dependencies
        build_binary
        mkdir -p dist
        build_tarball
        ;;
    "clean")
        print_status "Cleaning build artifacts..."
        rm -rf dist/ packaging/deb-build/ packaging/rpm-build/
        rm -f sysmedic
        print_success "Cleanup complete"
        ;;
    *)
        main
        ;;
esac
