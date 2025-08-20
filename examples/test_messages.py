#!/usr/bin/env python3
"""
Test script to verify WebSocket message format from SysMedic server.
This script connects to the WebSocket server and validates that messages
match the expected format.
"""

import json
import sys
import time
import websocket
import threading

class MessageFormatTester:
    def __init__(self, url, secret):
        self.url = url
        self.secret = secret
        self.ws_url = f"{url}?secret={secret}"
        self.ws = None
        self.connected = False
        self.messages_received = []
        self.welcome_received = False
        self.system_updates_received = 0

    def on_open(self, ws):
        print("✅ Connected to WebSocket server")
        self.connected = True

    def on_message(self, ws, message):
        try:
            data = json.loads(message)
            self.messages_received.append(data)

            print(f"\n📨 Message received:")
            print(f"Raw: {message}")
            print(f"Parsed: {json.dumps(data, indent=2)}")

            # Validate message format
            self.validate_message(data)

        except json.JSONDecodeError as e:
            print(f"❌ Invalid JSON: {e}")
            print(f"Raw message: {message}")

    def on_error(self, ws, error):
        print(f"❌ WebSocket error: {error}")

    def on_close(self, ws, close_status_code, close_msg):
        print(f"🔌 Connection closed: {close_status_code} {close_msg}")
        self.connected = False

    def validate_message(self, data):
        """Validate message format against expected structure"""

        # Check basic structure
        if not isinstance(data, dict):
            print("❌ Message is not a JSON object")
            return False

        if 'type' not in data:
            print("❌ Message missing 'type' field")
            return False

        if 'timestamp' not in data:
            print("❌ Message missing 'timestamp' field")
            return False

        msg_type = data['type']

        if msg_type == 'welcome':
            return self.validate_welcome_message(data)
        elif msg_type == 'system_update':
            return self.validate_system_update_message(data)
        else:
            print(f"ℹ️ Unknown message type: {msg_type}")
            return True

    def validate_welcome_message(self, data):
        """Validate welcome message format"""
        print("🎉 Validating welcome message...")

        expected_structure = {
            "type": "welcome",
            "data": {
                "message": "Connected to SysMedic",
                "version": "1.0.5",
                "system": str,  # e.g., "Ubuntu 22.04.5 LTS"
                "status": str,  # e.g., "Light Usage"
                "daemon": "Running"
            },
            "timestamp": str  # ISO format
        }

        # Check data field exists
        if 'data' not in data:
            print("❌ Welcome message missing 'data' field")
            return False

        welcome_data = data['data']

        # Check required fields
        required_fields = ['message', 'version', 'system', 'status', 'daemon']
        for field in required_fields:
            if field not in welcome_data:
                print(f"❌ Welcome data missing '{field}' field")
                return False

        # Validate specific values
        if welcome_data['message'] != "Connected to SysMedic":
            print(f"❌ Unexpected welcome message: {welcome_data['message']}")
            return False

        if welcome_data['daemon'] != "Running":
            print(f"❌ Unexpected daemon status: {welcome_data['daemon']}")
            return False

        # Check timestamp format (should be ISO format)
        timestamp = data['timestamp']
        if not isinstance(timestamp, str) or 'T' not in timestamp:
            print(f"❌ Invalid timestamp format: {timestamp}")
            return False

        print("✅ Welcome message format is correct!")
        print(f"   Message: {welcome_data['message']}")
        print(f"   Version: {welcome_data['version']}")
        print(f"   System: {welcome_data['system']}")
        print(f"   Status: {welcome_data['status']}")
        print(f"   Daemon: {welcome_data['daemon']}")
        print(f"   Timestamp: {timestamp}")

        self.welcome_received = True
        return True

    def validate_system_update_message(self, data):
        """Validate system update message format"""
        print("📊 Validating system update message...")

        # Check data field exists
        if 'data' not in data:
            print("❌ System update missing 'data' field")
            return False

        update_data = data['data']

        # Check required fields
        required_fields = ['cpu_usage', 'memory_usage', 'disk_usage', 'uptime']
        for field in required_fields:
            if field not in update_data:
                print(f"❌ System update missing '{field}' field")
                return False

        # Validate data types
        if not isinstance(update_data['cpu_usage'], (int, float)):
            print(f"❌ cpu_usage should be numeric, got: {type(update_data['cpu_usage'])}")
            return False

        if not isinstance(update_data['memory_usage'], (int, float)):
            print(f"❌ memory_usage should be numeric, got: {type(update_data['memory_usage'])}")
            return False

        if not isinstance(update_data['disk_usage'], (int, float)):
            print(f"❌ disk_usage should be numeric, got: {type(update_data['disk_usage'])}")
            return False

        if not isinstance(update_data['uptime'], str):
            print(f"❌ uptime should be string, got: {type(update_data['uptime'])}")
            return False

        # Validate ranges (should be 0-100 for percentages)
        for field in ['cpu_usage', 'memory_usage', 'disk_usage']:
            value = update_data[field]
            if not (0 <= value <= 100):
                print(f"⚠️ {field} seems unusual: {value}% (should be 0-100)")

        # Check timestamp format
        timestamp = data['timestamp']
        if not isinstance(timestamp, str) or 'T' not in timestamp:
            print(f"❌ Invalid timestamp format: {timestamp}")
            return False

        print("✅ System update message format is correct!")
        print(f"   CPU: {update_data['cpu_usage']}%")
        print(f"   Memory: {update_data['memory_usage']}%")
        print(f"   Disk: {update_data['disk_usage']}%")
        print(f"   Uptime: {update_data['uptime']}")
        print(f"   Timestamp: {timestamp}")

        self.system_updates_received += 1
        return True

    def connect_and_test(self, test_duration=10):
        """Connect and test for specified duration"""
        print(f"🔄 Connecting to {self.ws_url}")

        # Enable debug for troubleshooting
        # websocket.enableTrace(True)

        self.ws = websocket.WebSocketApp(
            self.ws_url,
            on_open=self.on_open,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close
        )

        # Start WebSocket in a separate thread
        ws_thread = threading.Thread(target=self.ws.run_forever, daemon=True)
        ws_thread.start()

        # Wait for connection and messages
        print(f"⏱️ Testing for {test_duration} seconds...")
        time.sleep(test_duration)

        # Close connection
        if self.ws:
            self.ws.close()

        # Wait for thread to finish
        time.sleep(1)

        # Report results
        self.print_test_results()

    def print_test_results(self):
        """Print test results summary"""
        print("\n" + "="*50)
        print("📋 TEST RESULTS SUMMARY")
        print("="*50)

        print(f"Total messages received: {len(self.messages_received)}")
        print(f"Welcome message received: {'✅ Yes' if self.welcome_received else '❌ No'}")
        print(f"System updates received: {self.system_updates_received}")

        if not self.welcome_received:
            print("❌ FAIL: No welcome message received")
            return False

        if self.system_updates_received == 0:
            print("❌ FAIL: No system updates received")
            return False

        if self.system_updates_received < 2:
            print("⚠️ WARNING: Expected more system updates (should be every 3 seconds)")

        print("\n✅ SUCCESS: All message formats are correct!")
        print(f"Expected: Welcome + System updates every 3 seconds")
        print(f"Received: Welcome + {self.system_updates_received} system updates")

        return True


def main():
    if len(sys.argv) != 3:
        print("Usage: python test_messages.py <ws_url> <secret>")
        print("Example: python test_messages.py ws://45.95.186.208:8060/ws 55625821f7a0a9db98707bae107e46a4")
        sys.exit(1)

    url = sys.argv[1]
    secret = sys.argv[2]

    print("🧪 SysMedic WebSocket Message Format Tester")
    print("=" * 45)
    print(f"URL: {url}")
    print(f"Secret: {secret[:8]}...")
    print()

    tester = MessageFormatTester(url, secret)

    try:
        # Test for 15 seconds to get multiple system updates
        tester.connect_and_test(test_duration=15)

    except KeyboardInterrupt:
        print("\n⏹️ Test interrupted by user")
    except Exception as e:
        print(f"❌ Test failed with error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
