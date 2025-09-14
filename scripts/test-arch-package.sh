#!/bin/bash

# Test script for Arch Linux package format verification
# Ensures that Arch packages are properly created as .pkg.tar.zst

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_info() {
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

# Test Arch package format
test_arch_package_format() {
    print_info "Testing Arch Linux package format..."

    cd "$PROJECT_ROOT"

    # Check if dist directory exists
    if [ ! -d "dist" ]; then
        print_error "dist directory not found. Run build-packages.sh first."
        exit 1
    fi

    cd dist

    # Look for Arch packages
    ARCH_PACKAGES=($(ls *.pkg.tar.zst 2>/dev/null || true))
    ARCH_PACKAGES_XZ=($(ls *.pkg.tar.xz 2>/dev/null || true))

    if [ ${#ARCH_PACKAGES[@]} -eq 0 ] && [ ${#ARCH_PACKAGES_XZ[@]} -eq 0 ]; then
        print_error "No Arch packages found (.pkg.tar.zst or .pkg.tar.xz)"
        print_info "Available files:"
        ls -la
        exit 1
    fi

    # Test .pkg.tar.zst packages (preferred)
    if [ ${#ARCH_PACKAGES[@]} -gt 0 ]; then
        for package in "${ARCH_PACKAGES[@]}"; do
            print_info "Testing package: $package"

            # Check file format
            if [[ "$package" == *.pkg.tar.zst ]]; then
                print_success "✓ Correct Arch package format: $package"
            else
                print_error "✗ Incorrect format: $package (should be .pkg.tar.zst)"
                exit 1
            fi

            # Check if file is actually zstd compressed
            if command -v zstd &> /dev/null; then
                if zstd -t "$package" &> /dev/null; then
                    print_success "✓ Valid zstd compression"
                else
                    print_error "✗ Invalid zstd compression"
                    exit 1
                fi
            else
                print_warning "zstd not available for compression verification"
            fi

            # Extract and check package contents
            TEMP_DIR=$(mktemp -d)
            cd "$TEMP_DIR"

            if command -v zstd &> /dev/null; then
                zstd -dc "$PROJECT_ROOT/dist/$package" | tar -x
            elif command -v tar &> /dev/null && tar --help | grep -q zstd; then
                tar --zstd -xf "$PROJECT_ROOT/dist/$package"
            else
                print_warning "Cannot extract zstd package for content verification"
                cd "$PROJECT_ROOT/dist"
                rm -rf "$TEMP_DIR"
                continue
            fi

            # Check required files in package
            if [ -f ".PKGINFO" ]; then
                print_success "✓ .PKGINFO file found"

                # Check package info content
                if grep -q "pkgname = sysmedic" .PKGINFO; then
                    print_success "✓ Correct package name in .PKGINFO"
                else
                    print_error "✗ Incorrect package name in .PKGINFO"
                fi

                if grep -q "arch = x86_64" .PKGINFO; then
                    print_success "✓ Correct architecture in .PKGINFO"
                else
                    print_warning "⚠ Architecture not found or incorrect in .PKGINFO"
                fi

                if grep -q "depend = systemd" .PKGINFO; then
                    print_success "✓ Systemd dependency found in .PKGINFO"
                else
                    print_warning "⚠ Systemd dependency not found in .PKGINFO"
                fi
            else
                print_error "✗ .PKGINFO file missing"
            fi

            # Check binary
            if [ -f "usr/bin/sysmedic" ]; then
                print_success "✓ sysmedic binary found in correct location"
                if [ -x "usr/bin/sysmedic" ]; then
                    print_success "✓ sysmedic binary is executable"
                else
                    print_error "✗ sysmedic binary is not executable"
                fi
            else
                print_error "✗ sysmedic binary not found in usr/bin/"
            fi

            # Check systemd services
            if [ -f "usr/lib/systemd/system/sysmedic.doctor.service" ]; then
                print_success "✓ Doctor service file found"
            else
                print_error "✗ Doctor service file missing"
            fi

            if [ -f "usr/lib/systemd/system/sysmedic.websocket.service" ]; then
                print_success "✓ WebSocket service file found"
            else
                print_error "✗ WebSocket service file missing"
            fi

            # Check configuration
            if [ -f "etc/sysmedic/config.yaml" ]; then
                print_success "✓ Configuration file found"
            else
                print_error "✗ Configuration file missing"
            fi

            # Check documentation
            if [ -f "usr/share/doc/sysmedic/README.md" ]; then
                print_success "✓ README documentation found"
            else
                print_warning "⚠ README documentation missing"
            fi

            if [ -f "usr/share/licenses/sysmedic/LICENSE" ]; then
                print_success "✓ License file found"
            else
                print_warning "⚠ License file missing"
            fi

            # Check directories
            if [ -d "var/lib/sysmedic" ]; then
                print_success "✓ Data directory structure created"
            else
                print_error "✗ Data directory missing"
            fi

            if [ -d "var/log/sysmedic" ]; then
                print_success "✓ Log directory structure created"
            else
                print_error "✗ Log directory missing"
            fi

            cd "$PROJECT_ROOT/dist"
            rm -rf "$TEMP_DIR"

            print_success "Package $package validation complete"
            echo ""
        done
    fi

    # Test .pkg.tar.xz packages (fallback)
    if [ ${#ARCH_PACKAGES_XZ[@]} -gt 0 ]; then
        print_warning "Found .pkg.tar.xz packages (fallback format):"
        for package in "${ARCH_PACKAGES_XZ[@]}"; do
            print_warning "  $package"

            # Basic xz format test
            if command -v xz &> /dev/null; then
                if xz -t "$package" &> /dev/null; then
                    print_success "✓ Valid xz compression"
                else
                    print_error "✗ Invalid xz compression"
                fi
            fi
        done
    fi
}

# Test makepkg availability
test_makepkg_availability() {
    print_info "Checking makepkg availability..."

    if command -v makepkg &> /dev/null; then
        print_success "✓ makepkg is available"
        makepkg_version=$(makepkg --version | head -1)
        print_info "Version: $makepkg_version"

        # Check PKGEXT setting
        if makepkg --help | grep -q "PKGEXT"; then
            print_success "✓ PKGEXT configuration supported"
        fi
    else
        print_warning "⚠ makepkg not available (will use manual packaging)"
    fi
}

# Test zstd availability
test_zstd_availability() {
    print_info "Checking zstd availability..."

    if command -v zstd &> /dev/null; then
        print_success "✓ zstd is available"
        zstd_version=$(zstd --version | head -1)
        print_info "Version: $zstd_version"
    else
        print_warning "⚠ zstd not available (will fallback to xz)"
    fi
}

# Main test function
main() {
    print_info "🧪 Arch Linux Package Format Test"
    echo ""

    test_makepkg_availability
    echo ""

    test_zstd_availability
    echo ""

    test_arch_package_format

    print_success "🎉 Arch package format test complete!"
}

# Handle arguments
case "${1:-}" in
    "format")
        test_arch_package_format
        ;;
    "tools")
        test_makepkg_availability
        test_zstd_availability
        ;;
    "--help"|"-h")
        echo "Arch Linux Package Format Test"
        echo ""
        echo "Usage: $0 [OPTION]"
        echo ""
        echo "Options:"
        echo "  format     Test package format only"
        echo "  tools      Test tool availability only"
        echo "  --help     Show this help"
        echo ""
        echo "Default: Run all tests"
        ;;
    *)
        main
        ;;
esac
