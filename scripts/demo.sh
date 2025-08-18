#!/bin/bash

# SysMedic Demo Script
# Demonstrates the key features of SysMedic system monitoring

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
DEMO_USER="demo_user"
DEMO_CONFIG="/tmp/sysmedic_demo_config.yaml"
SYSMEDIC_BIN="./build/sysmedic"

# Functions
print_header() {
    echo -e "\n${BLUE}========================================${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}========================================${NC}\n"
}

print_step() {
    echo -e "${GREEN}[STEP]${NC} $1"
}

print_info() {
    echo -e "${CYAN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

wait_for_user() {
    echo -e "\n${YELLOW}Press Enter to continue...${NC}"
    read -r
}

check_prerequisites() {
    print_header "Checking Prerequisites"

    print_step "Checking if SysMedic binary exists..."
    if [[ ! -f "$SYSMEDIC_BIN" ]]; then
        print_error "SysMedic binary not found at $SYSMEDIC_BIN"
        print_info "Please run 'make build' first"
        exit 1
    fi
    print_info "✓ SysMedic binary found"

    print_step "Checking permissions..."
    if [[ $EUID -ne 0 ]]; then
        print_warning "Not running as root - some features may not work"
        print_info "For full functionality, run: sudo $0"
    else
        print_info "✓ Running as root"
    fi

    print_step "Checking system compatibility..."
    if [[ ! -f /proc/stat ]]; then
        print_error "/proc/stat not found - this demo requires a Linux system"
        exit 1
    fi
    print_info "✓ Linux system detected"
}

create_demo_config() {
    print_header "Creating Demo Configuration"

    print_step "Creating demo configuration file..."
    cat > "$DEMO_CONFIG" << EOF
monitoring:
  check_interval: 10          # Fast interval for demo
  cpu_threshold: 70           # Lower threshold for demo
  memory_threshold: 70
  persistent_time: 1          # 1 minute for demo

users:
  cpu_threshold: 60           # Lower threshold for demo
  memory_threshold: 60
  persistent_time: 1

reporting:
  period: "hourly"
  retain_days: 1              # Short retention for demo

email:
  enabled: false              # Disabled for demo

user_thresholds:
  $DEMO_USER:
    cpu_threshold: 50         # Very low threshold for demo
    memory_threshold: 50
    persistent_time: 1
EOF

    print_info "✓ Demo configuration created at $DEMO_CONFIG"
    print_info "Configuration highlights:"
    echo "  - Check interval: 10 seconds (faster than default)"
    echo "  - CPU threshold: 70% (lower than default 80%)"
    echo "  - User thresholds: 60% (lower than default)"
    echo "  - Demo user threshold: 50% (very sensitive)"
}

demo_dashboard() {
    print_header "Demo 1: System Dashboard"

    print_step "Displaying current system status..."
    print_info "This shows real-time system metrics and top resource users"

    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN${NC}"
    $SYSMEDIC_BIN || true

    wait_for_user
}

demo_config_management() {
    print_header "Demo 2: Configuration Management"

    print_step "Showing current configuration..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN config show${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN config show || true

    wait_for_user

    print_step "Updating system threshold..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN config set cpu-threshold 85${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN config set cpu-threshold 85 || true

    print_step "Setting user-specific threshold..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN config set-user $DEMO_USER cpu-threshold 90${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN config set-user $DEMO_USER cpu-threshold 90 || true

    wait_for_user
}

demo_daemon() {
    print_header "Demo 3: Daemon Management"

    print_step "Checking daemon status..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN daemon status${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN daemon status || true

    if [[ $EUID -eq 0 ]]; then
        print_step "Starting daemon (will run for 30 seconds)..."
        echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN daemon start (background)${NC}"

        # Start daemon in background with timeout
        timeout 30s env SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN daemon start &
        DAEMON_PID=$!

        sleep 5

        print_step "Checking daemon status again..."
        SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN daemon status || true

        print_info "Daemon is now monitoring system every 10 seconds..."
        print_info "Letting it run for 25 more seconds to collect data..."

        sleep 25

        print_step "Stopping daemon..."
        SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN daemon stop || true

        # Ensure background process is killed
        kill $DAEMON_PID 2>/dev/null || true
        wait $DAEMON_PID 2>/dev/null || true
    else
        print_warning "Daemon demo requires root privileges"
        print_info "Run as root to see daemon functionality"
    fi

    wait_for_user
}

demo_reports() {
    print_header "Demo 4: Reports and Analytics"

    print_step "Showing system reports..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN reports${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN reports || true

    wait_for_user

    print_step "Showing user activity reports..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN reports users${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN reports users || true

    wait_for_user

    print_step "Showing top 5 users..."
    echo -e "\n${PURPLE}Running: $SYSMEDIC_BIN reports users --top 5${NC}"
    SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN reports users --top 5 || true

    wait_for_user
}

simulate_load() {
    print_header "Demo 5: Load Simulation (Optional)"

    print_info "This demo can simulate system load to trigger alerts"
    echo -e "${YELLOW}Would you like to simulate CPU load? (y/N)${NC}"
    read -r response

    if [[ "$response" =~ ^[Yy]$ ]]; then
        print_step "Starting CPU load simulation..."
        print_warning "This will consume CPU for 60 seconds"

        # Start CPU-intensive process
        yes > /dev/null &
        LOAD_PID1=$!
        yes > /dev/null &
        LOAD_PID2=$!

        print_info "Load simulation started (PIDs: $LOAD_PID1, $LOAD_PID2)"
        print_info "Check system status in another terminal:"
        echo "  $SYSMEDIC_BIN"

        sleep 10
        print_step "Checking system status with load..."
        SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN || true

        print_step "Stopping load simulation..."
        kill $LOAD_PID1 $LOAD_PID2 2>/dev/null || true
        wait $LOAD_PID1 $LOAD_PID2 2>/dev/null || true

        print_info "✓ Load simulation stopped"
    else
        print_info "Skipping load simulation"
    fi

    wait_for_user
}

show_advanced_features() {
    print_header "Demo 6: Advanced Features"

    print_step "Available advanced features:"
    echo "1. Email Alerts (configure SMTP in config)"
    echo "2. Persistent User Detection (tracks sustained high usage)"
    echo "3. Custom Per-User Thresholds"
    echo "4. Historical Data Analysis"
    echo "5. Systemd Integration"

    print_info "Email Alert Configuration Example:"
    echo "email:"
    echo "  enabled: true"
    echo "  smtp_host: \"smtp.gmail.com\""
    echo "  smtp_port: 587"
    echo "  username: \"alerts@company.com\""
    echo "  password: \"app_password\""
    echo "  from: \"sysmedic@company.com\""
    echo "  to: \"admin@company.com\""
    echo "  tls: true"

    wait_for_user

    print_info "Systemd Integration Commands:"
    echo "sudo systemctl enable sysmedic    # Auto-start on boot"
    echo "sudo systemctl start sysmedic     # Start service"
    echo "sudo systemctl status sysmedic    # Check service status"
    echo "sudo journalctl -u sysmedic -f    # View logs"

    wait_for_user
}

cleanup() {
    print_header "Cleanup"

    print_step "Cleaning up demo files..."
    rm -f "$DEMO_CONFIG"
    print_info "✓ Demo configuration removed"

    # Kill any remaining processes
    pkill -f "sysmedic.*daemon" 2>/dev/null || true
    pkill -f "yes" 2>/dev/null || true

    print_info "✓ Cleanup completed"
}

show_summary() {
    print_header "Demo Summary"

    print_info "You've seen the following SysMedic features:"
    echo "  ✓ Real-time system dashboard"
    echo "  ✓ Configuration management"
    echo "  ✓ Daemon operation"
    echo "  ✓ Reporting and analytics"
    echo "  ✓ Load detection"
    echo "  ✓ Advanced features overview"

    echo -e "\n${GREEN}Next Steps:${NC}"
    echo "1. Install SysMedic: make install"
    echo "2. Configure: edit /etc/sysmedic/config.yaml"
    echo "3. Enable service: sudo systemctl enable sysmedic"
    echo "4. Start monitoring: sudo systemctl start sysmedic"

    echo -e "\n${CYAN}Resources:${NC}"
    echo "• Documentation: README.md"
    echo "• Configuration: scripts/config.example.yaml"
    echo "• GitHub: https://github.com/sysmedic/sysmedic"

    echo -e "\n${PURPLE}Thank you for trying SysMedic!${NC}"
}

# Main execution
main() {
    print_header "SysMedic Interactive Demo"
    print_info "This demo will showcase SysMedic's key features"
    print_info "Demo duration: ~10 minutes"

    wait_for_user

    # Set trap for cleanup
    trap cleanup EXIT

    # Run demo steps
    check_prerequisites
    create_demo_config
    demo_dashboard
    demo_config_management
    demo_daemon
    demo_reports
    simulate_load
    show_advanced_features
    show_summary
}

# Handle command line arguments
case "${1:-}" in
    --help|-h)
        echo "SysMedic Demo Script"
        echo ""
        echo "Usage: $0 [OPTIONS]"
        echo ""
        echo "Options:"
        echo "  --help, -h     Show this help message"
        echo "  --quick        Run quick demo (skip interactive parts)"
        echo ""
        echo "This script demonstrates SysMedic's features interactively."
        echo "For full functionality, run as root: sudo $0"
        exit 0
        ;;
    --quick)
        print_header "SysMedic Quick Demo"
        check_prerequisites
        create_demo_config
        print_info "Dashboard:"
        $SYSMEDIC_BIN || true
        print_info "Configuration:"
        SYSMEDIC_CONFIG="$DEMO_CONFIG" $SYSMEDIC_BIN config show || true
        cleanup
        print_info "Quick demo completed!"
        exit 0
        ;;
    "")
        # No arguments, run full demo
        main
        ;;
    *)
        print_error "Unknown option: $1"
        print_info "Use --help for usage information"
        exit 1
        ;;
esac
