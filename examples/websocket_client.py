#!/usr/bin/env python3
"""
SysMedic WebSocket Client - Python Example

This script demonstrates how to connect to a SysMedic WebSocket server
and receive real-time system monitoring data and alerts.

Requirements:
    pip install websocket-client

Usage:
    python websocket_client.py sysmedic://secret@hostname:port/
"""

import sys
import json
import time
import signal
import threading
from urllib.parse import urlparse
import websocket

class SysMedicClient:
    def __init__(self, connection_url):
        self.connection_url = connection_url
        self.ws = None
        self.running = False
        self.metrics = {}

        # Parse the connection URL
        self.parse_url(connection_url)

    def parse_url(self, url):
        """Parse sysmedic://secret@host:port/ URL format"""
        if not url.startswith('sysmedic://'):
            raise ValueError("URL must start with 'sysmedic://'")

        # Replace sysmedic:// with http:// for parsing
        http_url = url.replace('sysmedic://', 'http://')
        parsed = urlparse(http_url)

        if not parsed.username or not parsed.hostname or not parsed.port:
            raise ValueError("Invalid URL format. Expected: sysmedic://secret@host:port/")

        self.secret = parsed.username
        self.host = parsed.hostname
        self.port = parsed.port
        self.ws_url = f"ws://{self.host}:{self.port}/ws?secret={self.secret}"

        print(f"🔗 Parsed connection:")
        print(f"   Host: {self.host}")
        print(f"   Port: {self.port}")
        print(f"   Secret: {self.secret[:8]}...")

    def on_message(self, ws, message):
        """Handle incoming WebSocket messages"""
        try:
            data = json.loads(message)
            self.handle_message(data)
        except json.JSONDecodeError as e:
            print(f"❌ Error parsing message: {e}")

    def on_error(self, ws, error):
        """Handle WebSocket errors"""
        print(f"❌ WebSocket error: {error}")

    def on_close(self, ws, close_status_code, close_msg):
        """Handle WebSocket connection close"""
        print(f"🔌 Connection closed (Code: {close_status_code})")
        if close_msg:
            print(f"   Message: {close_msg}")
        self.running = False

    def on_open(self, ws):
        """Handle WebSocket connection open"""
        print(f"✅ Connected to SysMedic at {self.host}:{self.port}")
        print("   Waiting for data...")
        self.running = True

    def handle_message(self, message):
        """Process different types of messages from SysMedic"""
        msg_type = message.get('type', 'unknown')
        data = message.get('data', {})
        timestamp = message.get('timestamp', '')

        if msg_type == 'welcome':
            print(f"🎉 {data.get('message', 'Welcome')} (Version: {data.get('version', 'unknown')})")

        elif msg_type == 'system_update':
            self.metrics.update(data)
            self.display_metrics(data)

        elif msg_type == 'alert':
            self.display_alert(data)

        else:
            print(f"📦 Unknown message type: {msg_type}")
            print(f"   Data: {json.dumps(data, indent=2)}")

    def display_metrics(self, data):
        """Display system metrics in a formatted way"""
        # Clear previous line (simple terminal update)
        print("\r" + " " * 80, end="\r")

        cpu = data.get('cpu_usage', 0)
        memory = data.get('memory_usage', 0)
        disk = data.get('disk_usage', 0)
        uptime = data.get('uptime', 'unknown')

        # Create progress bars
        cpu_bar = self.create_progress_bar(cpu)
        memory_bar = self.create_progress_bar(memory)
        disk_bar = self.create_progress_bar(disk)

        timestamp = time.strftime("%H:%M:%S")

        print(f"📊 [{timestamp}] CPU: {cpu:5.1f}% {cpu_bar} | "
              f"Memory: {memory:5.1f}% {memory_bar} | "
              f"Disk: {disk:5.1f}% {disk_bar} | "
              f"Uptime: {uptime}")

    def create_progress_bar(self, percentage, width=10):
        """Create a simple ASCII progress bar"""
        filled = int((percentage / 100) * width)
        bar = "█" * filled + "░" * (width - filled)
        return f"[{bar}]"

    def display_alert(self, data):
        """Display system alerts"""
        status = data.get('status', 'Unknown')
        primary_cause = data.get('primary_cause', 'No specific cause identified')
        recommendations = data.get('recommendations', [])

        print(f"\n🚨 ALERT: {status}")
        print(f"   Cause: {primary_cause}")

        if recommendations:
            print("   Recommendations:")
            for i, rec in enumerate(recommendations, 1):
                print(f"     {i}. {rec}")
        print()

    def connect(self):
        """Connect to the WebSocket server"""
        print(f"🔄 Connecting to {self.ws_url}...")

        # Enable debugging if needed
        # websocket.enableTrace(True)

        self.ws = websocket.WebSocketApp(
            self.ws_url,
            on_open=self.on_open,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close
        )

        # Start the WebSocket connection
        self.ws.run_forever()

    def disconnect(self):
        """Disconnect from the WebSocket server"""
        if self.ws:
            print("\n🔌 Disconnecting...")
            self.running = False
            self.ws.close()

    def display_status(self):
        """Display current connection status"""
        print(f"\n📋 Connection Status:")
        print(f"   URL: {self.connection_url}")
        print(f"   Host: {self.host}:{self.port}")
        print(f"   Status: {'🟢 Connected' if self.running else '🔴 Disconnected'}")

        if self.metrics:
            print(f"   Last Metrics:")
            for key, value in self.metrics.items():
                if isinstance(value, (int, float)):
                    print(f"     {key}: {value}")
                else:
                    print(f"     {key}: {value}")


def signal_handler(signum, frame):
    """Handle Ctrl+C gracefully"""
    print("\n\n👋 Goodbye!")
    sys.exit(0)


def print_usage():
    """Print usage information"""
    print("SysMedic WebSocket Client")
    print("=" * 25)
    print()
    print("Usage:")
    print("  python websocket_client.py <connection_url>")
    print()
    print("Connection URL Format:")
    print("  sysmedic://secret@hostname:port/")
    print()
    print("Examples:")
    print("  python websocket_client.py sysmedic://abc123@localhost:9090/")
    print("  python websocket_client.py sysmedic://d8852f78260f16d31eeff80ca6158848@server.example.com:8060/")
    print()
    print("Get your connection URL from:")
    print("  sysmedic websocket status")
    print()


def main():
    # Set up signal handler for graceful exit
    signal.signal(signal.SIGINT, signal_handler)

    # Check command line arguments
    if len(sys.argv) != 2:
        print_usage()
        sys.exit(1)

    connection_url = sys.argv[1]

    # Validate URL format
    if not connection_url.startswith('sysmedic://'):
        print("❌ Error: Invalid URL format")
        print_usage()
        sys.exit(1)

    try:
        # Create and connect client
        client = SysMedicClient(connection_url)

        print(f"🖥️  SysMedic WebSocket Client")
        print(f"============================")
        print()

        # Try to connect
        client.connect()

    except ValueError as e:
        print(f"❌ URL Error: {e}")
        print_usage()
        sys.exit(1)

    except KeyboardInterrupt:
        print("\n👋 Disconnected by user")

    except Exception as e:
        print(f"❌ Connection Error: {e}")
        print()
        print("Troubleshooting:")
        print("  1. Check if SysMedic daemon is running: sysmedic daemon status")
        print("  2. Verify WebSocket server is enabled: sysmedic websocket status")
        print("  3. Ensure the secret and connection details are correct")
        print("  4. Check firewall settings for the specified port")
        sys.exit(1)


if __name__ == "__main__":
    main()
