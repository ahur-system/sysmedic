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
DOCTOR_SERVICE_FILE="sysmedic.doctor.service"
WEBSOCKET_SERVICE_FILE="sysmedic.websocket.service"

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

stop_existing_daemons() {
    log_info "Checking for existing SysMedic installation..."

    # Check if SysMedic is already installed
    if command -v sysmedic &> /dev/null; then
        log_info "Existing SysMedic installation detected"
        log_warning "⚠️  Cannot update while daemons are active!"

        # Check if either daemon is running
        DOCTOR_RUNNING=false
        WEBSOCKET_RUNNING=false

        if systemctl is-active --quiet sysmedic.doctor 2>/dev/null; then
            DOCTOR_RUNNING=true
            log_info "Doctor daemon is currently running"
        elif pgrep -f "sysmedic.*--doctor-daemon" &> /dev/null; then
            DOCTOR_RUNNING=true
            log_info "Doctor daemon process is currently running"
        fi

        if systemctl is-active --quiet sysmedic.websocket 2>/dev/null; then
            WEBSOCKET_RUNNING=true
            log_info "WebSocket daemon is currently running"
        elif pgrep -f "sysmedic.*--websocket-daemon" &> /dev/null; then
            WEBSOCKET_RUNNING=true
            log_info "WebSocket daemon process is currently running"
        fi

        # Stop daemons if running
        if [[ "$DOCTOR_RUNNING" == true ]] || [[ "$WEBSOCKET_RUNNING" == true ]]; then
            log_info "🛑 Stopping running daemons before installation..."

            if [[ "$DOCTOR_RUNNING" == true ]]; then
                log_info "Stopping doctor daemon..."

                # Try systemctl first
                if systemctl is-active --quiet sysmedic.doctor 2>/dev/null; then
                    if systemctl stop sysmedic.doctor; then
                        log_success "Doctor daemon stopped successfully"
                    else
                        log_error "Failed to stop doctor daemon via systemctl"
                        exit 1
                    fi
                fi

                # Stop any remaining processes
                if pgrep -f "sysmedic.*--doctor-daemon" &> /dev/null; then
                    log_info "Stopping doctor daemon process..."
                    pkill -f "sysmedic.*--doctor-daemon" || true
                    sleep 2

                    # Force kill if still running
                    if pgrep -f "sysmedic.*--doctor-daemon" &> /dev/null; then
                        log_warning "Force stopping doctor daemon..."
                        pkill -9 -f "sysmedic.*--doctor-daemon" || true
                        sleep 1
                    fi
                fi
            fi

            if [[ "$WEBSOCKET_RUNNING" == true ]]; then
                log_info "Stopping WebSocket daemon..."

                # Try systemctl first
                if systemctl is-active --quiet sysmedic.websocket 2>/dev/null; then
                    if systemctl stop sysmedic.websocket; then
                        log_success "WebSocket daemon stopped successfully"
                    else
                        log_error "Failed to stop WebSocket daemon via systemctl"
                        exit 1
                    fi
                fi

                # Stop any remaining processes
                if pgrep -f "sysmedic.*--websocket-daemon" &> /dev/null; then
                    log_info "Stopping WebSocket daemon process..."
                    pkill -f "sysmedic.*--websocket-daemon" || true
                    sleep 2

                    # Force kill if still running
                    if pgrep -f "sysmedic.*--websocket-daemon" &> /dev/null; then
                        log_warning "Force stopping WebSocket daemon..."
                        pkill -9 -f "sysmedic.*--websocket-daemon" || true
                        sleep 1
                    fi
                fi
            fi

            # Final check - stop any remaining sysmedic processes
            if pgrep -f "sysmedic" &> /dev/null; then
                log_warning "Found additional SysMedic processes, stopping them..."
                pkill -f "sysmedic" || true
                sleep 3

                # Force kill if still running
                if pgrep -f "sysmedic" &> /dev/null; then
                    log_warning "Force stopping remaining SysMedic processes..."
                    pkill -9 -f "sysmedic" || true
                    sleep 1
                fi
            fi

            log_success "✅ All existing SysMedic daemons stopped"

            # Final verification
            if pgrep -f "sysmedic" &> /dev/null; then
                log_error "❌ Failed to stop all SysMedic processes"
                log_error "Manual intervention required. Please stop all SysMedic processes:"
                log_error "  sudo pkill -9 -f sysmedic"
                exit 1
            fi

            log_success "Installation can proceed safely"
        else
            log_info "No daemons are currently running"
        fi

    else
        log_info "No existing SysMedic installation found"
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
    log_info "Installing systemd services..."

    # Install doctor service
    if [[ -f "./scripts/$DOCTOR_SERVICE_FILE" ]]; then
        cp "./scripts/$DOCTOR_SERVICE_FILE" "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
        log_success "Installed doctor daemon service"
    elif [[ -f "./$DOCTOR_SERVICE_FILE" ]]; then
        cp "./$DOCTOR_SERVICE_FILE" "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
        log_success "Installed doctor daemon service"
    else
        log_warning "Doctor service file not found, creating basic service..."
        create_doctor_service
    fi

    # Install websocket service
    if [[ -f "./scripts/$WEBSOCKET_SERVICE_FILE" ]]; then
        cp "./scripts/$WEBSOCKET_SERVICE_FILE" "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"
        log_success "Installed WebSocket daemon service"
    elif [[ -f "./$WEBSOCKET_SERVICE_FILE" ]]; then
        cp "./$WEBSOCKET_SERVICE_FILE" "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"
        chmod 644 "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"
        log_success "Installed WebSocket daemon service"
    else
        log_warning "WebSocket service file not found, creating basic service..."
        create_websocket_service
    fi

    systemctl daemon-reload
}

create_doctor_service() {
    cat > "$SERVICE_DIR/$DOCTOR_SERVICE_FILE" << EOF
[Unit]
Description=SysMedic Doctor - System Monitoring Daemon
Documentation=https://github.com/sysmedic/sysmedic
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=$INSTALL_DIR/$BINARY_NAME --doctor-daemon
ExecReload=/bin/kill -HUP \$MAINPID
PIDFile=$DATA_DIR/sysmedic.doctor.pid
Restart=on-failure
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=SYSMEDIC_CONFIG=$CONFIG_DIR/config.yaml
Environment=SYSMEDIC_DATA=$DATA_DIR

[Install]
WantedBy=multi-user.target
EOF

    log_success "Created basic doctor daemon service"
}

create_websocket_service() {
    cat > "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE" << EOF
[Unit]
Description=SysMedic WebSocket - Remote Monitoring Server
Documentation=https://github.com/sysmedic/sysmedic
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=$INSTALL_DIR/$BINARY_NAME --websocket-daemon
ExecReload=/bin/kill -HUP \$MAINPID
PIDFile=$DATA_DIR/sysmedic.websocket.pid
Restart=on-failure
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=SYSMEDIC_CONFIG=$CONFIG_DIR/config.yaml
Environment=SYSMEDIC_DATA=$DATA_DIR

[Install]
WantedBy=multi-user.target
EOF

    log_success "Created basic WebSocket daemon service"
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

    # Check services
    if systemctl list-unit-files | grep -q "sysmedic.doctor.service"; then
        log_success "Doctor daemon service installed"
    else
        log_error "Doctor daemon service installation failed"
        exit 1
    fi

    if systemctl list-unit-files | grep -q "sysmedic.websocket.service"; then
        log_success "WebSocket daemon service installed"
    else
        log_error "WebSocket daemon service installation failed"
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
    log_info "🎯 Single Binary Multi-Daemon Architecture Installed"
    echo "   One binary (11MB) with independent daemon processes"
    echo
    log_info "Next steps:"
    echo "  1. Review configuration: $CONFIG_DIR/config.yaml"
    echo "  2. Enable both daemon services:"
    echo "     sudo systemctl enable sysmedic.doctor"
    echo "     sudo systemctl enable sysmedic.websocket"
    echo "  3. Start both daemon services:"
    echo "     sudo systemctl start sysmedic.doctor"
    echo "     sudo systemctl start sysmedic.websocket"
    echo "  4. Verify status: $BINARY_NAME daemon status"
    echo
    log_info "🚀 Quick commands:"
    echo "  • CLI interface: $BINARY_NAME"
    echo "  • Start doctor daemon: $BINARY_NAME daemon start"
    echo "  • Start WebSocket daemon: $BINARY_NAME websocket start"
    echo "  • Check both daemon status: $BINARY_NAME daemon status"
    echo "  • View monitoring reports: $BINARY_NAME reports"
    echo "  • Show configuration: $BINARY_NAME config show"
    echo "  • Test binary: $BINARY_NAME --version"
    echo
    log_info "🌐 Remote Access:"
    echo "  • WebSocket server: ws://localhost:8060/ws"
    echo "  • Access from CLI: $BINARY_NAME websocket status"
    echo
    log_info "📊 Process Architecture:"
    echo "  • Doctor daemon: System monitoring (independent process)"
    echo "  • WebSocket daemon: Remote access (independent process)"
    echo "  • CLI interface: Direct binary execution"
    echo
    log_info "📚 Documentation: https://github.com/ahur-system/sysmedic"
}

cleanup_on_error() {
    log_error "Installation failed. Cleaning up..."

    # Stop and disable both daemon services
    systemctl stop sysmedic.doctor 2>/dev/null || true
    systemctl stop sysmedic.websocket 2>/dev/null || true
    systemctl disable sysmedic.doctor 2>/dev/null || true
    systemctl disable sysmedic.websocket 2>/dev/null || true

    # Remove binary
    rm -f "$INSTALL_DIR/$BINARY_NAME"

    # Remove both service files
    rm -f "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
    rm -f "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"

    # Reload systemd to reflect changes
    systemctl daemon-reload

    # Remove directories (but keep config if it was customized)
    rmdir "$DATA_DIR" 2>/dev/null || true

    log_info "Cleanup completed - single binary multi-daemon installation removed"
    exit 1
}

# Main installation function
main() {
    echo "================================================="
    echo "    SysMedic Single Binary Multi-Daemon Setup   "
    echo "================================================="
    echo "  Installing single binary with independent"
    echo "  doctor (monitoring) and WebSocket daemons"
    echo "================================================="
    echo

    # Set trap for cleanup on error
    trap cleanup_on_error ERR

    # Run installation steps
    check_root
    check_system
    stop_existing_daemons
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
        log_info "Uninstalling SysMedic single binary multi-daemon architecture..."

        # Stop and disable both daemon services
        log_info "Stopping daemon services..."
        systemctl stop sysmedic.doctor 2>/dev/null || true
        systemctl stop sysmedic.websocket 2>/dev/null || true

        # Also stop any running processes
        pkill -f "sysmedic.*--doctor-daemon" 2>/dev/null || true
        pkill -f "sysmedic.*--websocket-daemon" 2>/dev/null || true
        sleep 2

        log_info "Disabling daemon services..."
        systemctl disable sysmedic.doctor 2>/dev/null || true
        systemctl disable sysmedic.websocket 2>/dev/null || true

        # Remove binary and service files
        log_info "Removing binary and service files..."
        rm -f "$INSTALL_DIR/$BINARY_NAME"
        rm -f "$SERVICE_DIR/$DOCTOR_SERVICE_FILE"
        rm -f "$SERVICE_DIR/$WEBSOCKET_SERVICE_FILE"

        # Reload systemd
        systemctl daemon-reload

        log_success "SysMedic single binary multi-daemon system uninstalled successfully"
        log_info "Configuration preserved: $CONFIG_DIR"
        log_info "Data preserved: $DATA_DIR"
        log_info "To remove all data: sudo rm -rf $CONFIG_DIR $DATA_DIR"
        log_info "Both doctor and WebSocket daemon services have been removed"
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
