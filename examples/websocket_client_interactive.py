#!/usr/bin/env python3
"""
SysMedic Interactive WebSocket Client - Python Example

This script demonstrates how to connect to a SysMedic WebSocket server
and request specific information from the server.

Requirements:
    pip install websocket-client

Usage:
    python websocket_client_interactive.py sysmedic://secret@hostname:port/
"""

import sys
import json
import time
import signal
import threading
import uuid
from urllib.parse import urlparse
import websocket

class SysMedicInteractiveClient:
    def __init__(self, connection_url):
        self.connection_url = connection_url
        self.ws = None
        self.running = False
        self.pending_requests = {}
        self.request_counter = 0

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

    def generate_request_id(self):
        """Generate unique request ID"""
        self.request_counter += 1
        return f"req_{self.request_counter}_{int(time.time())}"

    def send_request(self, request_type, data=None):
        """Send a request to the server"""
        if not self.ws or not self.running:
            print("❌ Cannot send request: Not connected to server")
            return None

        request_id = self.generate_request_id()
        request = {
            "type": request_type,
            "request_id": request_id,
            "data": data
        }

        try:
            self.ws.send(json.dumps(request))
            self.pending_requests[request_id] = {
                "type": request_type,
                "timestamp": time.time()
            }
            print(f"📤 Sent request: {request_type} (ID: {request_id})")
            return request_id
        except Exception as e:
            print(f"❌ Error sending request: {e}")
            return None

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
        print("   Interactive mode ready! Type 'help' for available commands.")
        self.running = True

    def handle_message(self, message):
        """Process different types of messages from SysMedic"""
        msg_type = message.get('type', 'unknown')
        data = message.get('data', {})
        timestamp = message.get('timestamp', '')
        request_id = message.get('request_id')

        # Handle responses to requests
        if request_id and request_id in self.pending_requests:
            request_info = self.pending_requests.pop(request_id)
            elapsed = time.time() - request_info['timestamp']

            print(f"\n📥 Response for {request_info['type']} (took {elapsed:.2f}s):")
            print("=" * 50)
            self.display_response_data(msg_type, data)
            print("=" * 50)
            return

        # Handle regular messages
        if msg_type == 'welcome':
            print(f"🎉 {data.get('message', 'Welcome')} (Version: {data.get('version', 'unknown')})")

        elif msg_type == 'system_update':
            self.display_system_update(data)

        elif msg_type == 'alert':
            self.display_alert(data)

        elif msg_type == 'error':
            print(f"❌ Server Error: {data.get('error', 'Unknown error')}")

        else:
            print(f"📦 Unknown message type: {msg_type}")
            print(f"   Data: {json.dumps(data, indent=2)}")

    def display_response_data(self, response_type, data):
        """Display response data in a formatted way"""
        if response_type == 'system_info_response':
            print("🖥️  SYSTEM INFORMATION:")
            for key, value in data.items():
                if isinstance(value, (int, float)):
                    if 'usage' in key:
                        print(f"   {key.replace('_', ' ').title()}: {value:.1f}%")
                    else:
                        print(f"   {key.replace('_', ' ').title()}: {value}")
                else:
                    print(f"   {key.replace('_', ' ').title()}: {value}")

        elif response_type == 'alerts_response':
            print("🚨 ALERTS INFORMATION:")
            print(f"   Unresolved: {data.get('unresolved_count', 0)}")
            print(f"   Total: {data.get('total_count', 0)}")
            print(f"   Status: {data.get('status', 'Unknown')}")
            recent = data.get('recent_alerts', [])
            if recent:
                print("   Recent Alerts:")
                for alert in recent:
                    print(f"     - {alert}")

        elif response_type == 'user_metrics_response':
            print("👥 USER METRICS:")
            if isinstance(data, list) and data:
                for user in data[:5]:  # Show top 5 users
                    username = user.get('username', 'unknown')
                    cpu = user.get('cpu_percent', 0)
                    memory = user.get('memory_percent', 0)
                    print(f"   {username}: CPU {cpu:.1f}%, Memory {memory:.1f}%")
                if len(data) > 5:
                    print(f"   ... and {len(data) - 5} more users")
            else:
                print("   No active users found")

        elif response_type == 'config_response':
            print("⚙️  CONFIGURATION:")
            for key, value in data.items():
                print(f"   {key.replace('_', ' ').title()}: {value}")

        elif response_type == 'uptime_response':
            print("⏰ UPTIME:")
            print(f"   System Uptime: {data.get('uptime', 'unknown')}")

        elif response_type == 'pong':
            print("🏓 PONG:")
            print(f"   Message: {data.get('message', 'pong')}")

        else:
            print(f"📄 {response_type.replace('_', ' ').upper()}:")
            print(json.dumps(data, indent=2))

    def display_system_update(self, data):
        """Display system metrics in a compact format"""
        timestamp = time.strftime("%H:%M:%S")
        cpu = data.get('cpu_usage', 0)
        memory = data.get('memory_usage', 0)
        disk = data.get('disk_usage', 0)
        uptime = data.get('uptime', 'unknown')

        # Create simple progress bars
        cpu_bar = self.create_progress_bar(cpu)
        memory_bar = self.create_progress_bar(memory)
        disk_bar = self.create_progress_bar(disk)

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

        self.ws = websocket.WebSocketApp(
            self.ws_url,
            on_open=self.on_open,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close
        )

        # Start WebSocket connection in a separate thread
        ws_thread = threading.Thread(target=self.ws.run_forever)
        ws_thread.daemon = True
        ws_thread.start()

        # Wait for connection
        time.sleep(2)

        if not self.running:
            raise Exception("Failed to connect to WebSocket server")

        return ws_thread

    def interactive_mode(self):
        """Run interactive command mode"""
        print("\n" + "=" * 60)
        print("🎮 INTERACTIVE MODE")
        print("=" * 60)
        self.show_help()

        while self.running:
            try:
                command = input("\nsysmedic> ").strip().lower()

                if command == 'help' or command == 'h':
                    self.show_help()
                elif command == 'system' or command == 's':
                    self.send_request('get_system_info')
                elif command == 'alerts' or command == 'a':
                    self.send_request('get_alerts')
                elif command == 'users' or command == 'u':
                    self.send_request('get_user_metrics')
                elif command == 'config' or command == 'c':
                    self.send_request('get_config')
                elif command == 'uptime' or command == 't':
                    self.send_request('get_uptime')
                elif command == 'ping' or command == 'p':
                    self.send_request('ping')
                elif command == 'status':
                    self.show_status()
                elif command == 'quit' or command == 'q' or command == 'exit':
                    break
                elif command == '':
                    continue
                else:
                    print(f"❌ Unknown command: {command}")
                    print("   Type 'help' for available commands")

            except KeyboardInterrupt:
                print("\n👋 Goodbye!")
                break
            except EOFError:
                print("\n👋 Goodbye!")
                break

        self.disconnect()

    def show_help(self):
        """Show available commands"""
        print("\n📋 Available Commands:")
        print("   system  (s) - Get current system information")
        print("   alerts  (a) - Get alerts information")
        print("   users   (u) - Get user metrics")
        print("   config  (c) - Get server configuration")
        print("   uptime  (t) - Get system uptime")
        print("   ping    (p) - Ping the server")
        print("   status      - Show connection status")
        print("   help    (h) - Show this help")
        print("   quit    (q) - Exit the client")

    def show_status(self):
        """Show connection status"""
        print(f"\n📊 Connection Status:")
        print(f"   Server: {self.host}:{self.port}")
        print(f"   Connected: {'🟢 Yes' if self.running else '🔴 No'}")
        print(f"   Pending Requests: {len(self.pending_requests)}")

        if self.pending_requests:
            print("   Waiting for responses:")
            for req_id, req_info in self.pending_requests.items():
                elapsed = time.time() - req_info['timestamp']
                print(f"     - {req_info['type']} ({elapsed:.1f}s ago)")

    def disconnect(self):
        """Disconnect from the WebSocket server"""
        if self.ws:
            print("\n🔌 Disconnecting...")
            self.running = False
            self.ws.close()


def signal_handler(signum, frame):
    """Handle Ctrl+C gracefully"""
    print("\n\n👋 Goodbye!")
    sys.exit(0)


def print_usage():
    """Print usage information"""
    print("SysMedic Interactive WebSocket Client")
    print("=" * 37)
    print()
    print("This client allows you to connect to SysMedic WebSocket server")
    print("and request specific information interactively.")
    print()
    print("Usage:")
    print("  python websocket_client_interactive.py <connection_url>")
    print()
    print("Connection URL Format:")
    print("  sysmedic://secret@hostname:port/")
    print()
    print("Examples:")
    print("  python websocket_client_interactive.py sysmedic://abc123@localhost:8060/")
    print("  python websocket_client_interactive.py sysmedic://d8852f78@server.com:8060/")
    print()
    print("Get your connection URL from:")
    print("  sysmedic websocket status")
    print()
    print("Available Requests:")
    print("  - System Information (CPU, Memory, Disk, etc.)")
    print("  - Alerts Information (Active alerts, status)")
    print("  - User Metrics (Per-user resource usage)")
    print("  - Server Configuration (Monitoring settings)")
    print("  - System Uptime")
    print("  - Ping/Pong test")
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
        client = SysMedicInteractiveClient(connection_url)

        print(f"🖥️  SysMedic Interactive WebSocket Client")
        print(f"=========================================")
        print()

        # Connect to server
        ws_thread = client.connect()

        # Run interactive mode
        client.interactive_mode()

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
