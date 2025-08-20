#!/bin/bash

# SysMedic Architecture Validation Script
# Validates the single binary multi-daemon architecture and updated scripts

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Functions
print_header() {
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}  SysMedic Architecture Validation${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo
}

print_section() {
    echo -e "${YELLOW}📋 $1${NC}"
    echo "----------------------------------------"
}

print_test() {
    local test_name="$1"
    local test_result="$2"
    local details="$3"

    TOTAL_TESTS=$((TOTAL_TESTS + 1))

    if [ "$test_result" = "PASS" ]; then
        echo -e "${GREEN}✅ $test_name${NC}"
        TESTS_PASSED=$((TESTS_PASSED + 1))
        if [ -n "$details" ]; then
            echo -e "   ${details}"
        fi
    else
        echo -e "${RED}❌ $test_name${NC}"
        TESTS_FAILED=$((TESTS_FAILED + 1))
        if [ -n "$details" ]; then
            echo -e "   ${RED}$details${NC}"
        fi
    fi
}

print_summary() {
    echo
    echo -e "${BLUE}============================================${NC}"
    echo -e "${BLUE}  Validation Summary${NC}"
    echo -e "${BLUE}============================================${NC}"
    echo "Total Tests: $TOTAL_TESTS"
    echo -e "Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Failed: ${RED}$TESTS_FAILED${NC}"

    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "\n${GREEN}🎉 All tests passed! Architecture is ready for release.${NC}"
        return 0
    else
        echo -e "\n${RED}⚠️  Some tests failed. Please review and fix issues.${NC}"
        return 1
    fi
}

# Test functions
test_project_structure() {
    print_section "Project Structure Tests"

    cd "$PROJECT_ROOT"

    # Test main binary exists
    if [ -f "sysmedic" ]; then
        print_test "Main binary exists" "PASS" "sysmedic binary found"
    else
        print_test "Main binary exists" "FAIL" "sysmedic binary not found - run 'make build' first"
    fi

    # Test service files exist
    if [ -f "scripts/sysmedic.doctor.service" ]; then
        print_test "Doctor service file exists" "PASS"
    else
        print_test "Doctor service file exists" "FAIL" "scripts/sysmedic.doctor.service missing"
    fi

    if [ -f "scripts/sysmedic.websocket.service" ]; then
        print_test "WebSocket service file exists" "PASS"
    else
        print_test "WebSocket service file exists" "FAIL" "scripts/sysmedic.websocket.service missing"
    fi

    # Test old service file is removed
    if [ ! -f "scripts/sysmedic.service" ]; then
        print_test "Old single service file removed" "PASS"
    else
        print_test "Old single service file removed" "FAIL" "scripts/sysmedic.service should be deleted"
    fi

    # Test test files are removed
    test_files=$(find scripts/ -name "test_*" 2>/dev/null | wc -l)
    if [ "$test_files" -eq 0 ]; then
        print_test "Test files removed" "PASS"
    else
        print_test "Test files removed" "FAIL" "Found $test_files test_* files in scripts/"
    fi
}

test_binary_functionality() {
    print_section "Binary Functionality Tests"

    cd "$PROJECT_ROOT"

    # Test binary is executable
    if [ -x "sysmedic" ]; then
        print_test "Binary is executable" "PASS"
    else
        print_test "Binary is executable" "FAIL" "sysmedic binary is not executable"
        return
    fi

    # Test version command
    if ./sysmedic --version >/dev/null 2>&1; then
        version_output=$(./sysmedic --version 2>/dev/null | head -1)
        print_test "Version command works" "PASS" "$version_output"
    else
        print_test "Version command works" "FAIL" "sysmedic --version failed"
    fi

    # Test help command
    if ./sysmedic --help >/dev/null 2>&1; then
        print_test "Help command works" "PASS"
    else
        print_test "Help command works" "FAIL" "sysmedic --help failed"
    fi

    # Test daemon modes are available by checking main.go source
    if grep -q "\-\-doctor-daemon" cmd/sysmedic/main.go; then
        print_test "Doctor daemon mode available" "PASS" "--doctor-daemon flag found in source"
    else
        print_test "Doctor daemon mode available" "FAIL" "--doctor-daemon flag not found in main.go"
    fi

    if grep -q "\-\-websocket-daemon" cmd/sysmedic/main.go; then
        print_test "WebSocket daemon mode available" "PASS" "--websocket-daemon flag found in source"
    else
        print_test "WebSocket daemon mode available" "FAIL" "--websocket-daemon flag not found in main.go"
    fi

    # Test URL parameter authentication feature
    if grep -q "getPublicIP\|ws.*secret=" pkg/cli/cli.go; then
        print_test "URL parameter authentication implemented" "PASS" "ws://[host]:[port]/ws?secret=[secret] format"
    else
        print_test "URL parameter authentication implemented" "FAIL" "URL parameter authentication not found in CLI"
    fi

    # Test quick connect URL feature
    if grep -q "sysmedic://.*@.*:" pkg/cli/cli.go; then
        print_test "Quick connect URL feature implemented" "PASS" "sysmedic://[secret]@[publicip]:[port]/ format"
    else
        print_test "Quick connect URL feature implemented" "FAIL" "Quick connect URL feature not found in CLI"
    fi
}

test_service_files() {
    print_section "Service File Tests"

    cd "$PROJECT_ROOT"

    # Test doctor service file content
    if grep -q "ExecStart=/usr/local/bin/sysmedic --doctor-daemon" scripts/sysmedic.doctor.service; then
        print_test "Doctor service has correct ExecStart" "PASS"
    else
        print_test "Doctor service has correct ExecStart" "FAIL" "Doctor service ExecStart is incorrect"
    fi

    # Test websocket service file content
    if grep -q "ExecStart=/usr/local/bin/sysmedic --websocket-daemon" scripts/sysmedic.websocket.service; then
        print_test "WebSocket service has correct ExecStart" "PASS"
    else
        print_test "WebSocket service has correct ExecStart" "FAIL" "WebSocket service ExecStart is incorrect"
    fi

    # Test both services have unique PID files
    doctor_pid=$(grep "PIDFile=" scripts/sysmedic.doctor.service | cut -d'=' -f2)
    websocket_pid=$(grep "PIDFile=" scripts/sysmedic.websocket.service | cut -d'=' -f2)

    if [ "$doctor_pid" != "$websocket_pid" ] && [ -n "$doctor_pid" ] && [ -n "$websocket_pid" ]; then
        print_test "Services have unique PID files" "PASS" "Doctor: $doctor_pid, WebSocket: $websocket_pid"
    else
        print_test "Services have unique PID files" "FAIL" "PID files are not unique or missing"
    fi
}

test_packaging_structure() {
    print_section "Packaging Structure Tests"

    cd "$PROJECT_ROOT"

    # Test DEB packaging structure
    if [ -f "packaging/deb/DEBIAN/control" ]; then
        print_test "DEB control file exists" "PASS"
    else
        print_test "DEB control file exists" "FAIL" "packaging/deb/DEBIAN/control missing"
    fi

    # Test DEB includes both service files
    deb_doctor_service="packaging/deb/etc/systemd/system/sysmedic.doctor.service"
    deb_websocket_service="packaging/deb/etc/systemd/system/sysmedic.websocket.service"

    if [ -f "$deb_doctor_service" ] && [ -f "$deb_websocket_service" ]; then
        print_test "DEB includes both service files" "PASS"
    else
        print_test "DEB includes both service files" "FAIL" "Missing service files in DEB structure"
    fi

    # Test RPM spec file
    if [ -f "packaging/rpm/sysmedic.spec" ]; then
        print_test "RPM spec file exists" "PASS"
    else
        print_test "RPM spec file exists" "FAIL" "packaging/rpm/sysmedic.spec missing"
    fi

    # Test RPM spec includes both services
    if grep -q "sysmedic.doctor.service" packaging/rpm/sysmedic.spec && \
       grep -q "sysmedic.websocket.service" packaging/rpm/sysmedic.spec; then
        print_test "RPM spec includes both services" "PASS"
    else
        print_test "RPM spec includes both services" "FAIL" "RPM spec missing service references"
    fi
}

test_build_script() {
    print_section "Build Script Tests"

    cd "$PROJECT_ROOT"

    # Test script syntax
    if bash -n scripts/build-packages.sh; then
        print_test "Build script syntax valid" "PASS"
    else
        print_test "Build script syntax valid" "FAIL" "build-packages.sh has syntax errors"
    fi

    # Test package naming
    if grep -q 'PACKAGE_NAME="sysmedic_${VERSION}_${ARCH}.deb"' scripts/build-packages.sh; then
        print_test "DEB package naming correct" "PASS" "Using Debian standard naming"
    else
        print_test "DEB package naming correct" "FAIL" "DEB naming should be sysmedic_VERSION_ARCH.deb"
    fi

    # Test includes both service files in tarball
    if grep -q "sysmedic.doctor.service" scripts/build-packages.sh && \
       grep -q "sysmedic.websocket.service" scripts/build-packages.sh; then
        print_test "Build script includes both services" "PASS"
    else
        print_test "Build script includes both services" "FAIL" "Build script missing service file references"
    fi
}

test_install_script() {
    print_section "Install Script Tests"

    cd "$PROJECT_ROOT"

    # Test script syntax
    if bash -n scripts/install.sh; then
        print_test "Install script syntax valid" "PASS"
    else
        print_test "Install script syntax valid" "FAIL" "install.sh has syntax errors"
    fi

    # Test handles both service files
    if grep -q "DOCTOR_SERVICE_FILE" scripts/install.sh && \
       grep -q "WEBSOCKET_SERVICE_FILE" scripts/install.sh; then
        print_test "Install script handles both services" "PASS"
    else
        print_test "Install script handles both services" "FAIL" "Install script missing service file handling"
    fi

    # Test uninstall removes both services
    if grep -A 10 "\-\-uninstall" scripts/install.sh | grep -q "sysmedic.doctor" && \
       grep -A 10 "\-\-uninstall" scripts/install.sh | grep -q "sysmedic.websocket"; then
        print_test "Uninstall removes both services" "PASS"
    else
        print_test "Uninstall removes both services" "FAIL" "Uninstall doesn't handle both services"
    fi
}

test_release_script() {
    print_section "Release Script Tests"

    cd "$PROJECT_ROOT"

    # Test script syntax
    if bash -n scripts/create-release.sh; then
        print_test "Release script syntax valid" "PASS"
    else
        print_test "Release script syntax valid" "FAIL" "create-release.sh has syntax errors"
    fi

    # Test mentions single binary multi-daemon architecture
    if grep -q "Single Binary Multi-Daemon" scripts/create-release.sh; then
        print_test "Release notes mention new architecture" "PASS"
    else
        print_test "Release notes mention new architecture" "FAIL" "Release notes should mention single binary multi-daemon"
    fi

    # Test package filename compatibility
    if grep -q "sysmedic_.*_amd64.deb" scripts/create-release.sh; then
        print_test "Release script uses correct DEB naming" "PASS"
    else
        print_test "Release script uses correct DEB naming" "FAIL" "Release script DEB naming incorrect"
    fi
}

test_version_consistency() {
    print_section "Version Consistency Tests"

    cd "$PROJECT_ROOT"

    # Get version from main.go
    main_version=$(grep 'version = "' cmd/sysmedic/main.go | cut -d'"' -f2)

    # Get version from build script
    build_version=$(grep 'VERSION=' scripts/build-packages.sh | head -1 | cut -d'"' -f2)

    # Get version from release script
    release_version=$(grep 'VERSION=' scripts/create-release.sh | head -1 | cut -d'"' -f2)

    # Get version from DEB control
    deb_version=$(grep 'Version:' packaging/deb/DEBIAN/control | awk '{print $2}')

    # Get version from RPM spec
    rpm_version=$(grep 'Version:' packaging/rpm/sysmedic.spec | awk '{print $2}')

    if [ "$main_version" = "$build_version" ] && \
       [ "$main_version" = "$release_version" ] && \
       [ "$main_version" = "$deb_version" ] && \
       [ "$main_version" = "$rpm_version" ]; then
        print_test "Version consistency across files" "PASS" "All files use version $main_version"
    else
        print_test "Version consistency across files" "FAIL" "Versions don't match: main=$main_version, build=$build_version, release=$release_version, deb=$deb_version, rpm=$rpm_version"
    fi
}

# Test URL parameter authentication and quick connect features
test_url_parameter_auth_feature() {
    print_section "URL Parameter Authentication & Quick Connect Tests"

    cd "$PROJECT_ROOT"

    # Test getPublicIP function exists
    if grep -q "func getPublicIP" pkg/cli/cli.go; then
        print_test "getPublicIP function exists" "PASS"
    else
        print_test "getPublicIP function exists" "FAIL" "getPublicIP function not found"
    fi

    # Test URL parameter format in status display
    if grep -q "ws.*secret=" pkg/cli/cli.go; then
        print_test "URL parameter format in status display" "PASS"
    else
        print_test "URL parameter format in status display" "FAIL" "URL parameter format not found"
    fi

    # Test URL parameter format in start command
    if grep -A 10 "WebSocket daemon started successfully" pkg/cli/cli.go | grep -q "ws.*secret="; then
        print_test "URL parameter format in start command" "PASS"
    else
        print_test "URL parameter format in start command" "FAIL" "Start command doesn't show URL parameter format"
    fi

    # Test quick connect URL in status display
    if grep -q "Quick Connect.*sysmedic://" pkg/cli/cli.go; then
        print_test "Quick connect URL in status display" "PASS"
    else
        print_test "Quick connect URL in status display" "FAIL" "Quick connect URL not found in status"
    fi

    # Test quick connect URL in start command
    if grep -A 10 "WebSocket daemon started successfully" pkg/cli/cli.go | grep -q "Quick Connect.*sysmedic://"; then
        print_test "Quick connect URL in start command" "PASS"
    else
        print_test "Quick connect URL in start command" "FAIL" "Start command doesn't show quick connect URL"
    fi

    # Test documentation updated for URL parameter auth
    if grep -q "ws.*secret=" docs/WEBSOCKET.md; then
        print_test "Documentation includes URL parameter auth" "PASS"
    else
        print_test "Documentation includes URL parameter auth" "FAIL" "WEBSOCKET.md missing URL parameter documentation"
    fi

    # Test documentation includes quick connect
    if grep -q "sysmedic://.*@.*:" docs/WEBSOCKET.md; then
        print_test "Documentation includes quick connect" "PASS"
    else
        print_test "Documentation includes quick connect" "FAIL" "WEBSOCKET.md missing quick connect documentation"
    fi

    # Test README mentions quick connect
    if grep -q "Quick Connect.*sysmedic://" README.md; then
        print_test "README mentions quick connect" "PASS"
    else
        print_test "README mentions quick connect" "FAIL" "README.md missing quick connect feature"
    fi
}

# Main execution
main() {
    print_header

    # Check if we're in the right directory
    if [ ! -f "go.mod" ] || [ ! -d "cmd/sysmedic" ]; then
        echo -e "${RED}❌ Not in SysMedic project root directory${NC}"
        exit 1
    fi

    # Run all test suites
    test_project_structure
    echo
    test_binary_functionality
    echo
    test_service_files
    echo
    test_packaging_structure
    echo
    test_build_script
    echo
    test_install_script
    echo
    test_release_script
    echo
    test_version_consistency
    echo
    test_url_parameter_auth_feature

    # Print summary and exit with appropriate code
    if print_summary; then
        exit 0
    else
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    "--help"|"-h")
        echo "SysMedic Architecture Validation Script"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --quick        Run only quick tests (skip binary tests)"
        echo
        echo "This script validates the single binary multi-daemon architecture"
        echo "and ensures all scripts are properly configured."
        exit 0
        ;;
    "--quick")
        print_header
        if [ ! -f "go.mod" ] || [ ! -d "cmd/sysmedic" ]; then
            echo -e "${RED}❌ Not in SysMedic project root directory${NC}"
            exit 1
        fi
        test_project_structure
        echo
        test_service_files
        echo
        test_packaging_structure
        echo
        test_version_consistency
        echo
        test_url_parameter_auth_feature
        if print_summary; then
            exit 0
        else
            exit 1
        fi
        ;;
    "")
        main
        ;;
    *)
        echo -e "${RED}Unknown option: $1${NC}"
        echo "Use --help for usage information"
        exit 1
        ;;
esac
