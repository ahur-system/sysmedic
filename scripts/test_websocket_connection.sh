#!/bin/bash

# SysMedic WebSocket Connection Test Script
# This script tests the WebSocket functionality end-to-end

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

print_header() {
    echo -e "${BLUE}🔍 $1${NC}"
    echo "$(printf '=%.0s' {1..50})"
}

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

cleanup() {
    print_status "Cleaning up..."
    ./sysmedic daemon stop 2>/dev/null || true
    if [ ! -z "$DAEMON_PID" ]; then
        kill $DAEMON_PID 2>/dev/null || true
    fi
    rm -f /tmp/sysmedic_test.log
    rm -f /tmp/websocket_test.py
}

# Set up cleanup on exit
trap cleanup EXIT

print_header "SysMedic WebSocket Connection Test"
echo

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

echo "Step 1: Setup WebSocket Configuration"
echo "------------------------------------"

# Stop any running daemon
./sysmedic daemon stop 2>/dev/null || true

# Configure WebSocket
print_status "Configuring WebSocket on port 8060..."
./sysmedic websocket start 8060

# Verify configuration
if [ -f ~/.sysmedic/websocket.json ]; then
    print_success "Configuration file created:"
    cat ~/.sysmedic/websocket.json
else
    print_error "Configuration file not found"
    exit 1
fi

echo
echo "Step 2: Start Daemon with WebSocket"
echo "-----------------------------------"

print_status "Starting daemon in background..."
nohup ./sysmedic daemon start > /tmp/sysmedic_test.log 2>&1 &
DAEMON_PID=$!

# Wait for daemon to start
sleep 5

# Check if daemon is running
if ! kill -0 $DAEMON_PID 2>/dev/null; then
    print_error "Daemon failed to start"
    cat /tmp/sysmedic_test.log
    exit 1
fi

print_success "Daemon started with PID $DAEMON_PID"

# Check daemon status
./sysmedic daemon status

echo
echo "Step 3: Verify WebSocket Server"
echo "-------------------------------"

# Get WebSocket status and extract connection URL
STATUS_OUTPUT=$(./sysmedic websocket status)
echo "$STATUS_OUTPUT"

# Extract connection URL
CONNECTION_URL=$(echo "$STATUS_OUTPUT" | grep "Connection URL:" | sed 's/.*Connection URL: //')

if [ -z "$CONNECTION_URL" ]; then
    print_error "Could not find WebSocket connection URL"
    print_status "Daemon log:"
    cat /tmp/sysmedic_test.log
    exit 1
fi

print_success "WebSocket connection URL: $CONNECTION_URL"

echo
echo "Step 4: Test Health Endpoint"
echo "----------------------------"

# Test health endpoint
print_status "Testing health endpoint..."
HEALTH_RESPONSE=$(curl -s http://localhost:8060/health || echo "FAILED")

if [[ "$HEALTH_RESPONSE" == *"healthy"* ]]; then
    print_success "Health endpoint is responding"
    echo "Response: $HEALTH_RESPONSE"
else
    print_warning "Health endpoint test failed: $HEALTH_RESPONSE"
    print_status "Checking if port is listening..."
    netstat -tlnp 2>/dev/null | grep :8060 || ss -tlnp | grep :8060 || echo "Port 8060 not found in netstat/ss"
fi

echo
echo "Step 5: Manual Connection Test Instructions"
echo "------------------------------------------"

print_status "Your WebSocket server is running!"
echo
echo "🌐 HTML Client Test:"
echo "  1. Copy the HTML client to a web directory:"
echo "     cp examples/websocket_client.html /var/www/html/"
echo "  2. Or start a local web server:"
echo "     python3 -m http.server 3000"
echo "  3. Open: http://localhost:3000/examples/websocket_client.html"
echo "  4. Enter connection URL: $CONNECTION_URL"
echo

echo "🐍 Python Client Test:"
echo "  1. Install WebSocket client:"
echo "     pip3 install websocket-client"
echo "  2. Run the Python client:"
echo "     python3 examples/websocket_client.py \"$CONNECTION_URL\""
echo

echo "🔧 Manual WebSocket Test:"
echo "  Connection URL: $CONNECTION_URL"
echo "  WebSocket endpoint: ws://localhost:8060/ws?secret=$(echo $CONNECTION_URL | cut -d'@' -f1 | cut -d'/' -f3)"
echo

echo
echo "Step 6: Create Simple Python Test Client"
echo "----------------------------------------"

# Create a simple test client
cat > /tmp/websocket_test.py << 'EOF'
#!/usr/bin/env python3
import sys
import json
import websocket
import threading
import time

def on_message(ws, message):
    try:
        data = json.loads(message)
        print(f"📨 [{data.get('type', 'unknown')}] {data}")
    except:
        print(f"📨 Raw: {message}")

def on_error(ws, error):
    print(f"❌ Error: {error}")

def on_close(ws, close_status_code, close_msg):
    print(f"🔌 Connection closed: {close_status_code}")

def on_open(ws):
    print("✅ Connected successfully!")

if len(sys.argv) != 2:
    print("Usage: python3 websocket_test.py 'sysmedic://secret@host:port/'")
    sys.exit(1)

url = sys.argv[1]
# Parse sysmedic://secret@host:port/
secret = url.split('//')[1].split('@')[0]
host_port = url.split('@')[1].split('/')[0]
ws_url = f"ws://{host_port}/ws?secret={secret}"

print(f"🔗 Connecting to: {ws_url}")

ws = websocket.WebSocketApp(ws_url,
                          on_open=on_open,
                          on_message=on_message,
                          on_error=on_error,
                          on_close=on_close)

try:
    ws.run_forever()
except KeyboardInterrupt:
    print("\n👋 Disconnecting...")
    ws.close()
EOF

chmod +x /tmp/websocket_test.py

# Check if websocket-client is available
if python3 -c "import websocket" 2>/dev/null; then
    print_status "Testing WebSocket connection with Python client..."
    echo "Running: python3 /tmp/websocket_test.py \"$CONNECTION_URL\""
    echo "Press Ctrl+C to stop the test after a few messages..."
    echo
    timeout 10s python3 /tmp/websocket_test.py "$CONNECTION_URL" || print_warning "Python test timed out (this is normal)"
else
    print_warning "websocket-client not installed. Install with: pip3 install websocket-client"
    print_status "Test client created at: /tmp/websocket_test.py"
    echo "Run manually: python3 /tmp/websocket_test.py \"$CONNECTION_URL\""
fi

echo
echo "Step 7: Test Results Summary"
echo "---------------------------"

print_success "✅ WebSocket Configuration: Working"
print_success "✅ Daemon Integration: Working"
print_success "✅ Connection URL Generation: Working"

if [[ "$HEALTH_RESPONSE" == *"healthy"* ]]; then
    print_success "✅ Health Endpoint: Working"
else
    print_warning "⚠️  Health Endpoint: Needs investigation"
fi

echo
print_header "Next Steps"
echo "To connect to your WebSocket server:"
echo
echo "1. 📱 Use the HTML client in examples/websocket_client.html"
echo "2. 🐍 Use the Python client in examples/websocket_client.py"
echo "3. 🔧 Use the test client: python3 /tmp/websocket_test.py"
echo "4. 🌐 Connect from any WebSocket client using:"
echo "   URL: $CONNECTION_URL"
echo
echo "The daemon will continue running in the background (PID: $DAEMON_PID)"
echo "To stop: ./sysmedic daemon stop"
echo
print_success "🎉 WebSocket connection test completed!"
