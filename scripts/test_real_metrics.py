#!/usr/bin/env python3
"""
Simple WebSocket test to demonstrate real-time system metrics
This script connects to SysMedic WebSocket and shows that data is actually updating
"""

import sys
import json
import time
import websocket
import threading
from datetime import datetime

class MetricsTracker:
    def __init__(self):
        self.metrics_history = []
        self.connection_time = None
        self.last_update = None

    def add_metrics(self, data):
        timestamp = datetime.now()
        self.metrics_history.append({
            'timestamp': timestamp,
            'data': data
        })
        self.last_update = timestamp

        # Keep only last 10 entries
        if len(self.metrics_history) > 10:
            self.metrics_history.pop(0)

    def show_summary(self):
        if len(self.metrics_history) < 2:
            print("❌ Not enough data to show changes")
            return

        print(f"\n📊 Metrics Summary ({len(self.metrics_history)} updates):")
        print("=" * 60)

        first = self.metrics_history[0]['data']
        last = self.metrics_history[-1]['data']

        def safe_get(data, key, default=0):
            return data.get(key, default) if isinstance(data, dict) else default

        def format_change(old_val, new_val, suffix='%'):
            if old_val == new_val:
                return f"{new_val:.1f}{suffix} (no change)"
            else:
                change = new_val - old_val
                sign = '+' if change > 0 else ''
                return f"{new_val:.1f}{suffix} ({sign}{change:.1f})"

        print(f"CPU Usage:    {format_change(safe_get(first, 'cpu_usage'), safe_get(last, 'cpu_usage'))}")
        print(f"Memory Usage: {format_change(safe_get(first, 'memory_usage'), safe_get(last, 'memory_usage'))}")
        print(f"Disk Usage:   {format_change(safe_get(first, 'disk_usage'), safe_get(last, 'disk_usage'))}")

        if 'uptime' in first and 'uptime' in last:
            print(f"Uptime:       {first.get('uptime', 'unknown')} → {last.get('uptime', 'unknown')}")

        if 'load_avg' in last:
            print(f"Load Average: {last.get('load_avg', 'unknown')}")

        if 'network_io' in last:
            print(f"Network I/O:  {last.get('network_io', 'unknown')}")

        # Show if data is actually changing
        cpu_changed = safe_get(first, 'cpu_usage') != safe_get(last, 'cpu_usage')
        mem_changed = safe_get(first, 'memory_usage') != safe_get(last, 'memory_usage')
        disk_changed = safe_get(first, 'disk_usage') != safe_get(last, 'disk_usage')

        if cpu_changed or mem_changed or disk_changed:
            print(f"\n✅ REAL DATA: Metrics are changing over time!")
        else:
            print(f"\n⚠️  STATIC DATA: Metrics appear to be the same")

        print("=" * 60)

tracker = MetricsTracker()

def on_message(ws, message):
    try:
        data = json.loads(message)
        msg_type = data.get('type', 'unknown')
        msg_data = data.get('data', {})
        timestamp = datetime.now().strftime("%H:%M:%S")

        if msg_type == 'welcome':
            print(f"🎉 [{timestamp}] Connected! {msg_data.get('message', 'Welcome')}")
            print(f"    Version: {msg_data.get('version', 'unknown')}")
            tracker.connection_time = datetime.now()

        elif msg_type == 'system_update':
            tracker.add_metrics(msg_data)
            count = len(tracker.metrics_history)

            # Show detailed info for first few updates
            if count <= 3:
                print(f"📊 [{timestamp}] Update #{count}:")
                for key, value in msg_data.items():
                    if isinstance(value, (int, float)):
                        print(f"    {key}: {value:.1f}")
                    else:
                        print(f"    {key}: {value}")
            else:
                # Show compact info for subsequent updates
                cpu = msg_data.get('cpu_usage', 0)
                mem = msg_data.get('memory_usage', 0)
                disk = msg_data.get('disk_usage', 0)
                print(f"📊 [{timestamp}] Update #{count}: CPU {cpu:.1f}% | Memory {mem:.1f}% | Disk {disk:.1f}%")

        elif msg_type == 'alert':
            print(f"🚨 [{timestamp}] ALERT: {msg_data}")

        else:
            print(f"📦 [{timestamp}] {msg_type}: {msg_data}")

    except json.JSONDecodeError as e:
        print(f"❌ Error parsing message: {e}")
    except Exception as e:
        print(f"❌ Unexpected error: {e}")

def on_error(ws, error):
    print(f"❌ WebSocket error: {error}")

def on_close(ws, close_status_code, close_msg):
    print(f"\n🔌 Connection closed (Code: {close_status_code})")
    if close_msg:
        print(f"   Message: {close_msg}")

    # Show summary when disconnecting
    tracker.show_summary()

    if tracker.connection_time:
        duration = datetime.now() - tracker.connection_time
        print(f"\nConnection duration: {duration.seconds} seconds")

def on_open(ws):
    print("✅ WebSocket connection established!")
    print("   Waiting for metrics... (Press Ctrl+C to stop)")

def main():
    if len(sys.argv) != 2:
        print("Real-Time SysMedic Metrics Test")
        print("=" * 35)
        print("Usage: python3 test_real_metrics.py 'sysmedic://secret@host:port/'")
        print("\nThis script connects to SysMedic WebSocket and verifies that")
        print("system metrics are actually updating with real data.")
        print("\nExample:")
        print("  python3 test_real_metrics.py 'sysmedic://abc123@localhost:8060/'")
        print("\nGet your connection URL from: ./sysmedic websocket status")
        sys.exit(1)

    url = sys.argv[1]

    if not url.startswith('sysmedic://'):
        print("❌ Invalid URL format. Must start with 'sysmedic://'")
        sys.exit(1)

    try:
        # Parse connection URL
        secret = url.split('//')[1].split('@')[0]
        host_port = url.split('@')[1].split('/')[0]
        ws_url = f"ws://{host_port}/ws?secret={secret}"

        print("🔍 SysMedic Real Metrics Test")
        print("=" * 30)
        print(f"Connection URL: {url}")
        print(f"WebSocket URL:  {ws_url}")
        print(f"Test Purpose:   Verify real-time data updates")
        print()
        print("This test will:")
        print("  1. Connect to your SysMedic WebSocket server")
        print("  2. Collect several metric updates")
        print("  3. Show if the data is actually changing")
        print("  4. Display a summary when you disconnect")
        print()

        # Create WebSocket connection
        ws = websocket.WebSocketApp(
            ws_url,
            on_open=on_open,
            on_message=on_message,
            on_error=on_error,
            on_close=on_close
        )

        # Run the connection
        ws.run_forever()

    except KeyboardInterrupt:
        print("\n👋 Test stopped by user")
    except Exception as e:
        print(f"❌ Connection failed: {e}")
        print("\nTroubleshooting:")
        print("  1. Make sure SysMedic daemon is running: ./sysmedic daemon status")
        print("  2. Check WebSocket server status: ./sysmedic websocket status")
        print("  3. Verify the connection URL is correct")
        print("  4. Check if the port is accessible")

if __name__ == "__main__":
    main()
