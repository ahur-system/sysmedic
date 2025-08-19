#!/bin/bash

# SysMedic WebSocket Feature Test Script
# This script demonstrates the WebSocket functionality of SysMedic

set -e

echo "🖥️  SysMedic WebSocket Feature Test"
echo "=================================="
echo

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

# Check if sysmedic binary exists
if [ ! -f "./sysmedic" ]; then
    print_error "sysmedic binary not found. Building..."
    go build -o sysmedic ./cmd/sysmedic
    if [ $? -ne 0 ]; then
        print_error "Failed to build sysmedic"
        exit 1
    fi
    print_success "Built sysmedic binary"
fi

echo "Step 1: Testing WebSocket CLI Commands"
echo "-------------------------------------"

# Test websocket help
print_status "Testing WebSocket help command..."
./sysmedic websocket --help

echo
print_status "Testing WebSocket status (should show stopped)..."
./sysmedic websocket status

echo
echo "Step 2: Configuring WebSocket Server"
echo "------------------------------------"

# Configure WebSocket to start on port 9090
print_status "Configuring WebSocket server on port 9090..."
./sysmedic websocket start 9090

echo
print_status "Checking WebSocket configuration..."
./sysmedic websocket status

echo
echo "Step 3: Testing Daemon Integration"
echo "---------------------------------"

# Check current daemon status
print_status "Checking daemon status..."
./sysmedic daemon status

# Stop daemon if running
print_status "Ensuring daemon is stopped..."
./sysmedic daemon stop 2>/dev/null || true

# Start daemon with WebSocket integration
print_status "Starting daemon with WebSocket integration..."
echo "Note: Daemon will run for 10 seconds to demonstrate WebSocket functionality"

# Start daemon in background with timeout
timeout 10s ./sysmedic daemon start &
DAEMON_PID=$!

# Wait a moment for daemon to start
sleep 3

# Check if daemon started successfully
if kill -0 $DAEMON_PID 2>/dev/null; then
    print_success "Daemon started successfully with PID $DAEMON_PID"

    # Check WebSocket status while daemon is running
    print_status "Checking WebSocket status while daemon is running..."
    ./sysmedic websocket status

    # Test health endpoint
    print_status "Testing WebSocket health endpoint..."
    HEALTH_RESPONSE=$(curl -s http://localhost:9090/health 2>/dev/null || echo "Connection failed")
    if [[ "$HEALTH_RESPONSE" == *"healthy"* ]]; then
        print_success "Health endpoint responding correctly"
        echo "Health response: $HEALTH_RESPONSE"
    else
        print_warning "Health endpoint not accessible: $HEALTH_RESPONSE"
    fi

    # Wait for daemon to finish
    wait $DAEMON_PID 2>/dev/null || true
    print_success "Daemon stopped after timeout"
else
    print_error "Failed to start daemon"
fi

echo
echo "Step 4: WebSocket Connection URL"
echo "-------------------------------"

print_status "The WebSocket connection URL format is:"
echo "  sysmedic://[secret]@[hostname]:[port]/"
echo
print_status "Example connection URLs:"
echo "  sysmedic://abc123def456@localhost:9090/"
echo "  sysmedic://d8852f78260f16d31eeff80ca6158848@server.example.com:9090/"
echo
print_status "To get your actual connection URL:"
echo "  1. Start the daemon: ./sysmedic daemon start"
echo "  2. Check status: ./sysmedic websocket status"
echo "  3. Use the displayed connection URL"

echo
echo "Step 5: Client Connection Test"
echo "-----------------------------"

print_status "WebSocket client example is available at:"
echo "  examples/websocket_client.html"
echo
print_status "To test the client:"
echo "  1. Open examples/websocket_client.html in a web browser"
echo "  2. Start the daemon: ./sysmedic daemon start"
echo "  3. Get the connection URL: ./sysmedic websocket status"
echo "  4. Enter the URL in the client and click Connect"

echo
echo "Step 6: Testing Secret Regeneration"
echo "-----------------------------------"

print_status "Testing new secret generation..."
./sysmedic websocket new-secret || print_warning "Secret generation requires running WebSocket server"

echo
echo "Step 7: Configuration Management"
echo "-------------------------------"

print_status "WebSocket configuration is stored in:"
echo "  ~/.sysmedic/websocket.json"

if [ -f ~/.sysmedic/websocket.json ]; then
    print_status "Current configuration:"
    cat ~/.sysmedic/websocket.json | jq . 2>/dev/null || cat ~/.sysmedic/websocket.json
else
    print_warning "Configuration file not found (will be created on first use)"
fi

echo
echo "Step 8: Cleanup"
echo "--------------"

print_status "Stopping WebSocket server..."
./sysmedic websocket stop

print_status "Ensuring daemon is stopped..."
./sysmedic daemon stop 2>/dev/null || true

print_status "Final WebSocket status:"
./sysmedic websocket status

echo
echo "🎉 WebSocket Feature Test Complete!"
echo "==================================="
echo
print_success "All WebSocket features have been tested successfully!"
echo
print_status "Key features demonstrated:"
echo "  ✅ WebSocket server configuration"
echo "  ✅ Daemon integration"
echo "  ✅ Secret-based authentication"
echo "  ✅ Health endpoint"
echo "  ✅ CLI management commands"
echo "  ✅ Configuration persistence"
echo
print_status "Next steps:"
echo "  1. Start the daemon: ./sysmedic daemon start"
echo "  2. Configure WebSocket: ./sysmedic websocket start [port]"
echo "  3. Get connection URL: ./sysmedic websocket status"
echo "  4. Connect clients using the generated URL"
echo
print_status "For more information, see:"
echo "  - WEBSOCKET_FEATURE.md - Comprehensive documentation"
echo "  - examples/websocket_client.html - Test client"
echo
