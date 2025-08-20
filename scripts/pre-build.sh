#!/bin/bash

# SysMedic Pre-Build Validation Script
# This script runs before building packages to ensure everything is ready

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if make build succeeds
check_make_build() {
    log_info "Running 'make build' to validate build process..."

    cd "$PROJECT_ROOT"

    # Clean any previous builds
    if [ -d "build" ]; then
        rm -rf build/*
    fi

    # Run make build
    if make build; then
        log_success "make build completed successfully"

        # Verify binary was created
        if [ -f "build/sysmedic" ]; then
            log_success "Binary created at build/sysmedic"

            # Check if binary is executable
            if [ -x "build/sysmedic" ]; then
                log_success "Binary is executable"

                # Test binary version
                if ./build/sysmedic version &> /dev/null; then
                    VERSION=$(./build/sysmedic version 2>/dev/null | head -1)
                    log_success "Binary test passed: $VERSION"
                else
                    log_warning "Binary version check failed (may still be functional)"
                fi
            else
                log_error "Binary is not executable"
                return 1
            fi
        else
            log_error "Binary not found after make build"
            return 1
        fi
    else
        log_error "make build failed"
        return 1
    fi

    return 0
}

# Check Go environment
check_go_environment() {
    log_info "Checking Go environment..."

    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        return 1
    fi

    GO_VERSION=$(go version | awk '{print $3}')
    log_success "Go version: $GO_VERSION"

    # Check Go modules
    cd "$PROJECT_ROOT"
    if [ -f "go.mod" ]; then
        log_success "go.mod found"

        # Check module name
        MODULE_NAME=$(grep "^module " go.mod | awk '{print $2}')
        log_info "Module: $MODULE_NAME"

        # Verify dependencies
        if go mod verify; then
            log_success "Module verification passed"
        else
            log_warning "Module verification failed, running go mod tidy..."
            go mod tidy
            if go mod verify; then
                log_success "Module verification passed after tidy"
            else
                log_error "Module verification still failing"
                return 1
            fi
        fi
    else
        log_error "go.mod not found"
        return 1
    fi

    return 0
}

# Check project structure
check_project_structure() {
    log_info "Checking project structure..."

    cd "$PROJECT_ROOT"

    # Check essential files
    ESSENTIAL_FILES=(
        "go.mod"
        "go.sum"
        "cmd/sysmedic/main.go"
        "Makefile"
        "scripts/install.sh"
    )

    for file in "${ESSENTIAL_FILES[@]}"; do
        if [ -f "$file" ]; then
            log_success "✓ $file exists"
        else
            log_error "✗ $file missing"
            return 1
        fi
    done

    # Check essential directories
    ESSENTIAL_DIRS=(
        "cmd/sysmedic"
        "internal"
        "pkg"
        "scripts"
    )

    for dir in "${ESSENTIAL_DIRS[@]}"; do
        if [ -d "$dir" ]; then
            log_success "✓ $dir/ directory exists"
        else
            log_error "✗ $dir/ directory missing"
            return 1
        fi
    done

    log_success "Project structure check completed"
    return 0
}

# Check build dependencies
check_build_dependencies() {
    log_info "Checking build dependencies..."

    # Check for make
    if command -v make &> /dev/null; then
        log_success "make is available"
    else
        log_error "make is not installed"
        return 1
    fi

    # Check for git (for version info)
    if command -v git &> /dev/null; then
        log_success "git is available"

        # Check if we're in a git repo
        cd "$PROJECT_ROOT"
        if git rev-parse --git-dir &> /dev/null; then
            COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
            log_info "Git commit: $COMMIT"
        else
            log_warning "Not in a git repository"
        fi
    else
        log_warning "git not available (version info may be limited)"
    fi

    # Check for CGO dependencies
    if command -v gcc &> /dev/null; then
        log_success "gcc is available (CGO support)"
    else
        log_warning "gcc not found (CGO may not work)"
    fi

    log_success "Go environment check completed"
    return 0
}

# Check source code quality
check_source_quality() {
    log_info "Checking source code quality..."

    cd "$PROJECT_ROOT"

    # Check Go formatting
    if go fmt ./...; then
        log_success "Go formatting check passed"
    else
        log_warning "Go formatting issues found and fixed"
    fi

    # Check Go vet
    if go vet ./...; then
        log_success "Go vet check passed"
    else
        log_error "Go vet found issues"
        return 1
    fi

    # Check for common issues
    if grep -r "TODO" --include="*.go" . &> /dev/null; then
        TODO_COUNT=$(grep -r "TODO" --include="*.go" . | wc -l)
        log_warning "Found $TODO_COUNT TODO comments in code"
    fi

    if grep -r "FIXME" --include="*.go" . &> /dev/null; then
        FIXME_COUNT=$(grep -r "FIXME" --include="*.go" . | wc -l)
        log_warning "Found $FIXME_COUNT FIXME comments in code"
    fi

    log_success "Build dependencies check completed"
    return 0
}

# Check for running daemons (important for build environment)
check_running_daemons() {
    log_info "Checking for running SysMedic daemons..."

    DOCTOR_RUNNING=false
    WEBSOCKET_RUNNING=false

    # Check systemctl services
    if systemctl is-active --quiet sysmedic.doctor 2>/dev/null; then
        DOCTOR_RUNNING=true
        log_warning "Doctor daemon is running"
    fi

    if systemctl is-active --quiet sysmedic.websocket 2>/dev/null; then
        WEBSOCKET_RUNNING=true
        log_warning "WebSocket daemon is running"
    fi

    # Check for processes
    if pgrep -f "sysmedic.*--doctor-daemon" &> /dev/null; then
        DOCTOR_RUNNING=true
        log_warning "Doctor daemon process is running"
    fi

    if pgrep -f "sysmedic.*--websocket-daemon" &> /dev/null; then
        WEBSOCKET_RUNNING=true
        log_warning "WebSocket daemon process is running"
    fi

    if [[ "$DOCTOR_RUNNING" == true ]] || [[ "$WEBSOCKET_RUNNING" == true ]]; then
        log_warning "SysMedic daemons are running during build"
        log_warning "Consider stopping them for a clean build environment"
        log_info "To stop: sudo systemctl stop sysmedic.doctor sysmedic.websocket"
    else
        log_success "No SysMedic daemons are running"
    fi

    log_success "Source quality check completed"
    log_success "Running daemons check completed"
    return 0
}

# Main validation function
main() {
    echo "================================================================="
    echo "                SysMedic Pre-Build Validation                   "
    echo "================================================================="
    echo "  Validating environment before building packages"
    echo "================================================================="
    echo

    CHECKS_PASSED=0
    TOTAL_CHECKS=6

    # Run all checks
    if check_project_structure; then
        ((CHECKS_PASSED++))
    fi
    echo

    if check_go_environment; then
        ((CHECKS_PASSED++))
    fi
    echo

    if check_build_dependencies; then
        ((CHECKS_PASSED++))
    fi
    echo

    if check_source_quality; then
        ((CHECKS_PASSED++))
    fi
    echo

    if check_running_daemons; then
        ((CHECKS_PASSED++))
    fi
    echo

    # The critical test - make build
    if check_make_build; then
        ((CHECKS_PASSED++))
    fi
    echo

    # Summary
    echo "================================================================="
    echo "                    Validation Summary                           "
    echo "================================================================="

    if [ $CHECKS_PASSED -eq $TOTAL_CHECKS ]; then
        log_success "🎉 All validation checks passed ($CHECKS_PASSED/$TOTAL_CHECKS)"
        log_success "Ready to build packages!"
        echo
        log_info "Next steps:"
        echo "  1. Run: ./scripts/build-packages.sh"
        echo "  2. Or run: make package"
        echo "  3. Or run specific: ./scripts/build-packages.sh deb"
        exit 0
    else
        log_error "❌ Validation failed ($CHECKS_PASSED/$TOTAL_CHECKS checks passed)"
        log_error "Please fix the issues above before building packages"
        exit 1
    fi
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "SysMedic Pre-Build Validation Script"
        echo
        echo "This script validates the build environment before package creation."
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --make-only    Only run 'make build' test"
        echo "  --quick        Skip source quality checks"
        echo
        echo "The script performs these validations:"
        echo "  1. Project structure check"
        echo "  2. Go environment validation"
        echo "  3. Build dependencies check"
        echo "  4. Source code quality check"
        echo "  5. Running daemons check"
        echo "  6. Make build test (critical)"
        exit 0
        ;;
    --make-only)
        log_info "Running make build test only..."
        if check_make_build; then
            log_success "make build test passed"
            exit 0
        else
            log_error "make build test failed"
            exit 1
        fi
        ;;
    --quick)
        log_info "Running quick validation (skipping source quality checks)..."

        CHECKS_PASSED=0
        TOTAL_CHECKS=4

        if check_project_structure; then
            ((CHECKS_PASSED++))
        fi

        if check_go_environment; then
            ((CHECKS_PASSED++))
        fi

        if check_build_dependencies; then
            ((CHECKS_PASSED++))
        fi

        if check_make_build; then
            ((CHECKS_PASSED++))
        fi

        if [ $CHECKS_PASSED -eq $TOTAL_CHECKS ]; then
            log_success "Quick validation passed ($CHECKS_PASSED/$TOTAL_CHECKS)"
            exit 0
        else
            log_error "Quick validation failed ($CHECKS_PASSED/$TOTAL_CHECKS)"
            exit 1
        fi
        ;;
    "")
        # No arguments, run full validation
        main
        ;;
    *)
        log_error "Unknown option: $1"
        log_info "Use --help for usage information"
        exit 1
        ;;
esac
