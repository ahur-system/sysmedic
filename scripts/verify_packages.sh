#!/bin/bash

# SysMedic v1.0.3 Package Verification Script
# Tests DEB, RPM, and TAR.GZ packages to ensure they work correctly

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_TOTAL=0

echo -e "${BLUE}🔍 SysMedic v1.0.3 Package Verification${NC}"
echo "========================================"
echo

# Function to run a test
run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_pattern="$3"

    TESTS_TOTAL=$((TESTS_TOTAL + 1))
    echo -e "${BLUE}Test $TESTS_TOTAL: $test_name${NC}"

    if output=$(eval "$test_command" 2>&1); then
        if [[ -z "$expected_pattern" ]] || echo "$output" | grep -q "$expected_pattern"; then
            echo -e "  ${GREEN}✓ PASSED${NC}"
            TESTS_PASSED=$((TESTS_PASSED + 1))
        else
            echo -e "  ${RED}✗ FAILED${NC}"
            echo "    Expected pattern: $expected_pattern"
            echo "    Actual output:"
            echo "$output" | sed 's/^/    /'
        fi
    else
        echo -e "  ${RED}✗ FAILED (Command failed)${NC}"
        echo "    Output:"
        echo "$output" | sed 's/^/    /'
    fi
    echo
}

# Function to check if packages exist
check_packages() {
    echo -e "${YELLOW}Checking package files...${NC}"

    local missing_packages=0

    if [[ ! -f "dist/sysmedic-1.0.3-amd64.deb" ]]; then
        echo -e "${RED}❌ DEB package not found: dist/sysmedic-1.0.3-amd64.deb${NC}"
        missing_packages=$((missing_packages + 1))
    else
        echo -e "${GREEN}✅ DEB package found${NC}"
    fi

    if [[ ! -f "dist/sysmedic-1.0.3-x86_64.rpm" ]]; then
        echo -e "${RED}❌ RPM package not found: dist/sysmedic-1.0.3-x86_64.rpm${NC}"
        missing_packages=$((missing_packages + 1))
    else
        echo -e "${GREEN}✅ RPM package found${NC}"
    fi

    if [[ ! -f "dist/sysmedic-v1.0.3-linux-amd64.tar.gz" ]]; then
        echo -e "${RED}❌ TAR.GZ package not found: dist/sysmedic-v1.0.3-linux-amd64.tar.gz${NC}"
        missing_packages=$((missing_packages + 1))
    else
        echo -e "${GREEN}✅ TAR.GZ package found${NC}"
    fi

    if [[ $missing_packages -gt 0 ]]; then
        echo -e "${RED}Error: $missing_packages package(s) missing. Run 'make package' first.${NC}"
        exit 1
    fi

    echo
}

# Function to test DEB package
test_deb_package() {
    echo -e "${YELLOW}Testing DEB Package...${NC}"
    echo "====================="

    run_test "DEB Package Info" "dpkg-deb --info dist/sysmedic-1.0.3-amd64.deb" "Package: sysmedic"
    run_test "DEB Package Version" "dpkg-deb --info dist/sysmedic-1.0.3-amd64.deb" "Version: 1.0.3"
    run_test "DEB Package Contents" "dpkg-deb --contents dist/sysmedic-1.0.3-amd64.deb" "/usr/local/bin/sysmedic"
    run_test "DEB Package Systemd Service" "dpkg-deb --contents dist/sysmedic-1.0.3-amd64.deb" "/etc/systemd/system/sysmedic.service"
    run_test "DEB Package Control Scripts" "dpkg-deb --contents dist/sysmedic-1.0.3-amd64.deb" "DEBIAN"

    # Test file permissions
    run_test "DEB Binary Permissions" "dpkg-deb --contents dist/sysmedic-1.0.3-amd64.deb | grep 'usr/local/bin/sysmedic'" "-rwxr-xr-x"
}

# Function to test RPM package
test_rpm_package() {
    echo -e "${YELLOW}Testing RPM Package...${NC}"
    echo "====================="

    run_test "RPM Package Info" "rpm -qip dist/sysmedic-1.0.3-x86_64.rpm" "Name.*: sysmedic"
    run_test "RPM Package Version" "rpm -qip dist/sysmedic-1.0.3-x86_64.rpm" "Version.*: 1.0.3"
    run_test "RPM Package Contents" "rpm -qlp dist/sysmedic-1.0.3-x86_64.rpm" "/usr/local/bin/sysmedic"
    run_test "RPM Package Systemd Service" "rpm -qlp dist/sysmedic-1.0.3-x86_64.rpm" "/etc/systemd/system/sysmedic.service"
    run_test "RPM Package Scripts" "rpm -qp --scripts dist/sysmedic-1.0.3-x86_64.rpm" "%pre\\|%post\\|%preun\\|%postun"

    # Test package dependencies
    run_test "RPM Dependencies" "rpm -qp --requires dist/sysmedic-1.0.3-x86_64.rpm" "libc.so.6"
}

# Function to test TAR.GZ package
test_tarball_package() {
    echo -e "${YELLOW}Testing TAR.GZ Package...${NC}"
    echo "========================="

    run_test "TAR.GZ Package Contents" "tar -tzf dist/sysmedic-v1.0.3-linux-amd64.tar.gz" "sysmedic-v1.0.3/sysmedic"
    run_test "TAR.GZ README" "tar -tzf dist/sysmedic-v1.0.3-linux-amd64.tar.gz" "sysmedic-v1.0.3/README.md"
    run_test "TAR.GZ License" "tar -tzf dist/sysmedic-v1.0.3-linux-amd64.tar.gz" "sysmedic-v1.0.3/LICENSE"

    # Extract and test binary
    local temp_dir=$(mktemp -d)
    tar -xzf dist/sysmedic-v1.0.3-linux-amd64.tar.gz -C "$temp_dir"

    run_test "TAR.GZ Binary Executable" "test -x $temp_dir/sysmedic-v1.0.3/sysmedic && echo 'executable'" "executable"
    run_test "TAR.GZ Binary Version" "$temp_dir/sysmedic-v1.0.3/sysmedic --version" "v1.0.3"
    run_test "TAR.GZ Alert Commands" "$temp_dir/sysmedic-v1.0.3/sysmedic alerts --help" "View, resolve, and manage"

    # Clean up
    rm -rf "$temp_dir"
}

# Function to test binary functionality
test_binary_functionality() {
    echo -e "${YELLOW}Testing Binary Functionality...${NC}"
    echo "=============================="

    # Use the built binary for testing
    if [[ -f "build/sysmedic-linux-amd64" ]]; then
        BINARY="build/sysmedic-linux-amd64"
    elif [[ -f "build/sysmedic" ]]; then
        BINARY="build/sysmedic"
    elif [[ -f "./sysmedic" ]]; then
        BINARY="./sysmedic"
    else
        echo -e "${RED}❌ No sysmedic binary found for testing${NC}"
        return
    fi

    run_test "Binary Version Check" "$BINARY --version" "v1.0.3"
    run_test "Binary Help Command" "$BINARY --help" "Cross-platform Linux server monitoring"
    run_test "Alert Management Help" "$BINARY alerts --help" "View, resolve, and manage"
    run_test "Alert List Help" "$BINARY alerts list --help" "Show only unresolved alerts"
    run_test "Alert Resolve Help" "$BINARY alerts resolve --help" "Resolve a specific alert"
    run_test "Alert Resolve All Help" "$BINARY alerts resolve-all --help" "Resolve all unresolved alerts"

    # Test error handling
    run_test "Invalid Alert ID Error" "$BINARY alerts resolve abc 2>&1" "Invalid alert ID"
    run_test "Non-existent Alert Error" "$BINARY alerts resolve 99999 2>&1" "not found"

    # Test basic functionality (these might show "no alerts" which is fine)
    run_test "Alert Overview" "$BINARY alerts" "Alert Summary\\|No alerts"
    run_test "Alert List" "$BINARY alerts list" "Total:\\|No alerts"
    run_test "Dashboard Display" "$BINARY" "System:"
}

# Function to verify package sizes
verify_package_sizes() {
    echo -e "${YELLOW}Verifying Package Sizes...${NC}"
    echo "========================="

    # Check DEB package size (should be reasonable, not empty)
    local deb_size=$(stat -c%s "dist/sysmedic-1.0.3-amd64.deb" 2>/dev/null || echo "0")
    if [[ $deb_size -gt 1000000 ]]; then  # > 1MB
        echo -e "${GREEN}✅ DEB package size: $(( deb_size / 1024 / 1024 ))MB${NC}"
    else
        echo -e "${RED}❌ DEB package too small: ${deb_size} bytes${NC}"
    fi

    # Check RPM package size
    local rpm_size=$(stat -c%s "dist/sysmedic-1.0.3-x86_64.rpm" 2>/dev/null || echo "0")
    if [[ $rpm_size -gt 1000000 ]]; then  # > 1MB
        echo -e "${GREEN}✅ RPM package size: $(( rpm_size / 1024 / 1024 ))MB${NC}"
    else
        echo -e "${RED}❌ RPM package too small: ${rpm_size} bytes${NC}"
    fi

    # Check TAR.GZ package size
    local tar_size=$(stat -c%s "dist/sysmedic-v1.0.3-linux-amd64.tar.gz" 2>/dev/null || echo "0")
    if [[ $tar_size -gt 1000000 ]]; then  # > 1MB
        echo -e "${GREEN}✅ TAR.GZ package size: $(( tar_size / 1024 / 1024 ))MB${NC}"
    else
        echo -e "${RED}❌ TAR.GZ package too small: ${tar_size} bytes${NC}"
    fi

    echo
}

# Main execution
main() {
    echo "Starting package verification tests..."
    echo

    # Pre-flight checks
    check_packages
    verify_package_sizes

    # Test each package type
    test_deb_package
    test_rpm_package
    test_tarball_package
    test_binary_functionality

    # Final summary
    echo "========================================"
    echo -e "${BLUE}Verification Summary${NC}"
    echo "========================================"
    echo "Tests Passed: $TESTS_PASSED"
    echo "Tests Total:  $TESTS_TOTAL"
    echo

    if [[ $TESTS_PASSED -eq $TESTS_TOTAL ]]; then
        echo -e "${GREEN}🎉 All package verification tests passed!${NC}"
        echo
        echo -e "${GREEN}✅ Packages Ready for Distribution:${NC}"
        echo "   📦 dist/sysmedic-1.0.3-amd64.deb"
        echo "   📦 dist/sysmedic-1.0.3-x86_64.rpm"
        echo "   📦 dist/sysmedic-v1.0.3-linux-amd64.tar.gz"
        echo
        echo -e "${GREEN}Release v1.0.3 is ready! 🚀${NC}"
        exit 0
    else
        failed=$((TESTS_TOTAL - TESTS_PASSED))
        echo -e "${RED}❌ $failed verification test(s) failed${NC}"
        echo "Please review the failures above before distributing packages."
        exit 1
    fi
}

# Display usage information
usage() {
    echo "Usage: $0 [options]"
    echo
    echo "Options:"
    echo "  -h, --help     Show this help message"
    echo "  -q, --quiet    Run tests quietly (minimal output)"
    echo
    echo "This script verifies all SysMedic v1.0.3 packages are correctly built."
    echo "Run from the sysmedic project root directory."
    echo
}

# Parse command line arguments
case "${1:-}" in
    -h|--help)
        usage
        exit 0
        ;;
    -q|--quiet)
        # Redirect output for quiet mode, but keep summary
        exec 3>&1
        exec > /dev/null 2>&1
        main
        exec 1>&3 3>&-
        ;;
    *)
        main
        ;;
esac
