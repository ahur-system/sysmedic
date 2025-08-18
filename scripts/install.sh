#!/bin/bash

# SysMedic Installation Script
# This script installs SysMedic system monitoring tool

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BINARY_NAME="sysmedic"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/sysmedic"
DATA_DIR="/var/lib/sysmedic"
SERVICE_DIR="/etc/systemd/system"
SERVICE_FILE="sysmedic.service"

# Functions
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

check_root() {
    if [[ $EUID -ne 0 ]]; then
        log_error "This script must be run as root"
        log_info "Please run: sudo $0"
        exit 1
    fi
}

check_system() {
    log_info "Checking system compatibility..."

    # Check if systemd is available
    if ! command -v systemctl &> /dev/null; then
        log_error "systemd is required but not found"
        exit 1
    fi

    # Check architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64)
            log_info "Architecture: x86_64 (supported)"
            ;;
        aarch64|arm64)
            log_info "Architecture: ARM64 (supported)"
            ;;
        *)
            log_warning "Architecture $ARCH may not be fully supported"
            ;;
    esac

    # Check OS
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        log_info "Operating System: $PRETTY_NAME"
    else
        log_warning "Could not detect operating system"
    fi
}

create_directories() {
    log_info "Creating directories..."

    # Create configuration directory
    if [[ ! -d "$CONFIG_DIR" ]]; then
        mkdir -p "$CONFIG_DIR"
        chmod 755 "$CONFIG_DIR"
        log_success "Created $CONFIG_DIR"
    fi

    # Create data directory
    if [[ ! -d "$DATA_DIR" ]]; then
        mkdir -p "$DATA_DIR"
        chmod 755 "$DATA_DIR"
        log_success "Created $DATA_DIR"
    fi
}

install_binary() {
    log_info "Installing SysMedic binary..."

    # Check if binary exists in current directory
    if [[ -f "./$BINARY_NAME" ]]; then
        cp "./$BINARY_NAME" "$INSTALL_DIR/$BINARY_NAME"
        chmod +x "$INSTALL_DIR/$BINARY_NAME"
        log_success "Installed $BINARY_NAME to $INSTALL_DIR"
    else
        log_error "Binary $BINARY_NAME not found in current directory"
        log_info "Please ensure you've extracted the SysMedic package correctly"
        exit 1
    fi
}

install_service() {
    log_info "Installing systemd service..."

    if [[ -f "./scripts/$SERVICE_FILE" ]]; then
        cp "./scripts/$SERVICE_FILE" "$SERVICE_DIR/$SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$SERVICE_FILE"
        systemctl daemon-reload
        log_success "Installed systemd service"
    elif [[ -f "./$SERVICE_FILE" ]]; then
        cp "./$SERVICE_FILE" "$SERVICE_DIR/$SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$SERVICE_FILE"
        systemctl daemon-reload
        log_success "Installed systemd service"
    else
        log_warning "Service file not found, creating basic service..."
        create_basic_service
    fi
}

create_basic_service() {
    cat > "$SERVICE_DIR/$SERVICE_FILE" << EOF
[Unit]
Description=SysMedic System Monitoring Daemon
Documentation=https://github.com/sysmedic/sysmedic
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=$INSTALL_DIR/$BINARY_NAME daemon start
ExecStop=$INSTALL_DIR/$BINARY_NAME daemon stop
PIDFile=/var/run/sysmedic.pid
Restart=on-failure
RestartSec=5
Environment=PATH=/usr/local/bin:/usr/bin:/bin

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    log_success "Created basic systemd service"
}

install_config() {
    log_info "Setting up configuration..."

    CONFIG_FILE="$CONFIG_DIR/config.yaml"

    if [[ ! -f "$CONFIG_FILE" ]]; then
        if [[ -f "./scripts/config.example.yaml" ]]; then
            cp "./scripts/config.example.yaml" "$CONFIG_FILE"
            log_success "Installed example configuration to $CONFIG_FILE"
        else
            # Create minimal default config
            cat > "$CONFIG_FILE" << EOF
monitoring:
  check_interval: 60
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

users:
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

reporting:
  period: "hourly"
  retain_days: 30

email:
  enabled: false
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  tls: true

user_thresholds: {}
EOF
            log_success "Created default configuration at $CONFIG_FILE"
        fi

        chmod 644 "$CONFIG_FILE"
        log_info "Please edit $CONFIG_FILE to customize your settings"
    else
        log_info "Configuration file already exists at $CONFIG_FILE"
        log_info "Backup created at $CONFIG_FILE.backup"
        cp "$CONFIG_FILE" "$CONFIG_FILE.backup"
    fi
}

verify_installation() {
    log_info "Verifying installation..."

    # Check binary
    if [[ -x "$INSTALL_DIR/$BINARY_NAME" ]]; then
        log_success "Binary installed and executable"

        # Test binary
        if "$INSTALL_DIR/$BINARY_NAME" --version &> /dev/null; then
            VERSION=$("$INSTALL_DIR/$BINARY_NAME" --version 2>/dev/null | head -1)
            log_success "Binary test passed: $VERSION"
        else
            log_warning "Binary installed but version check failed"
        fi
    else
        log_error "Binary installation failed"
        exit 1
    fi

    # Check service
    if systemctl list-unit-files | grep -q "$BINARY_NAME.service"; then
        log_success "Systemd service installed"
    else
        log_error "Systemd service installation failed"
        exit 1
    fi

    # Check directories
    for dir in "$CONFIG_DIR" "$DATA_DIR"; do
        if [[ -d "$dir" ]]; then
            log_success "Directory $dir exists"
        else
            log_error "Directory $dir missing"
            exit 1
        fi
    done
}

show_next_steps() {
    log_success "SysMedic installation completed successfully!"
    echo
    log_info "Next steps:"
    echo "  1. Review configuration: $CONFIG_DIR/config.yaml"
    echo "  2. Enable service: sudo systemctl enable $BINARY_NAME"
    echo "  3. Start service: sudo systemctl start $BINARY_NAME"
    echo "  4. Check status: $BINARY_NAME"
    echo
    log_info "Quick commands:"
    echo "  • View dashboard: $BINARY_NAME"
    echo "  • Start daemon: sudo $BINARY_NAME daemon start"
    echo "  • View reports: $BINARY_NAME reports"
    echo "  • Show config: $BINARY_NAME config show"
    echo
    log_info "Documentation: https://github.com/sysmedic/sysmedic"
}

cleanup_on_error() {
    log_error "Installation failed. Cleaning up..."

    # Remove binary
    rm -f "$INSTALL_DIR/$BINARY_NAME"

    # Remove service
    systemctl stop "$BINARY_NAME" 2>/dev/null || true
    systemctl disable "$BINARY_NAME" 2>/dev/null || true
    rm -f "$SERVICE_DIR/$SERVICE_FILE"
    systemctl daemon-reload

    log_info "Cleanup completed"
    exit 1
}

# Main installation function
main() {
    echo "=========================================="
    echo "       SysMedic Installation Script      "
    echo "=========================================="
    echo

    # Set trap for cleanup on error
    trap cleanup_on_error ERR

    # Run installation steps
    check_root
    check_system
    create_directories
    install_binary
    install_service
    install_config
    verify_installation
    show_next_steps
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "SysMedic Installation Script"
        echo
        echo "Usage: $0 [OPTIONS]"
        echo
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --uninstall    Uninstall SysMedic"
        echo "  --version      Show version information"
        echo
        echo "This script must be run as root (with sudo)."
        echo "Make sure you have extracted the SysMedic package before running."
        exit 0
        ;;
    --uninstall)
        check_root
        log_info "Uninstalling SysMedic..."

        # Stop and disable service
        systemctl stop "$BINARY_NAME" 2>/dev/null || true
        systemctl disable "$BINARY_NAME" 2>/dev/null || true

        # Remove files
        rm -f "$INSTALL_DIR/$BINARY_NAME"
        rm -f "$SERVICE_DIR/$SERVICE_FILE"
        systemctl daemon-reload

        log_success "SysMedic uninstalled successfully"
        log_info "Configuration and data files preserved in $CONFIG_DIR and $DATA_DIR"
        log_info "To remove completely: sudo rm -rf $CONFIG_DIR $DATA_DIR"
        exit 0
        ;;
    --version)
        echo "SysMedic Installation Script v1.0"
        exit 0
        ;;
    "")
        # No arguments, proceed with installation
        main
        ;;
    *)
        log_error "Unknown option: $1"
        log_info "Use --help for usage information"
        exit 1
        ;;
esac
