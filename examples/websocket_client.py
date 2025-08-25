#!/usr/bin/env python3
"""
SysMedic WebSocket Client - Python Example

This script demonstrates how to connect to a SysMedic WebSocket server
and receive real-time system monitoring data and alerts with robust
connection handling, automatic reconnection, and ping/pong monitoring.

Requirements:
    pip install websocket-client

Usage:
    python websocket_client.py ws://hostname:port/ws secret
    or
    python websocket_client.py ws://45.95.186.208:8060/ws 55625821f7a0a9db98707bae107e46a4
"""

import sys
import json
import time
import signal
import threading
import logging
from urllib.parse import urlparse, urlencode
import websocket

# Configure logging
logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger(__name__)

class SysMedicClient:
    def __init__(self, url, secret, **options):
        self.url = url
        self.secret = secret
        self.ws = None
        self.running = False
        self.connected = False
        self.reconnect_attempts = 0
        self.max_reconnect_attempts = options.get('max_reconnect_attempts', 10)
        self.reconnect_interval = options.get('reconnect_interval', 5)
        self.ping_interval = options.get('ping_interval', 30)
        self.pong_timeout = options.get('pong_timeout', 10)

        self.last_pong = time.time()
        self.ping_timer = None
        self.pong_timer = None
        self.reconnect_timer = None

        self.metrics = {}
        self.message_count = 0

        # Event callbacks
        self.on_connect = options.get('on_connect', lambda: None)
        self.on_disconnect = options.get('on_disconnect', lambda: None)
        self.on_message = options.get('on_message', lambda data: None)
        self.on_error = options.get('on_error', lambda error: None)
        self.on_reconnecting = options.get('on_reconnecting', lambda attempt, delay: None)

        # Build WebSocket URL with secret
        self.ws_url = f"{self.url}?secret={self.secret}"

        # Auto-connect if not disabled
        if options.get('auto_connect', True):
            self.connect()

    def connect(self):
        """Establish WebSocket connection"""
        try:
            logger.info(f"Connecting to {self.url}...")

            # Clean up any existing connection
            self.cleanup()

            # Enable trace for debugging (optional)
            websocket.enableTrace(False)

            # Create WebSocket connection
            self.ws = websocket.WebSocketApp(
                self.ws_url,
                on_open=self._on_open,
                on_message=self._on_message,
                on_error=self._on_error,
                on_close=self._on_close,
                on_ping=self._on_ping,
                on_pong=self._on_pong
            )

            # Start the WebSocket connection in a separate thread
            self.ws_thread = threading.Thread(target=self.ws.run_forever, daemon=True)
            self.ws_thread.start()

        except Exception as error:
            logger.error(f"Failed to create WebSocket connection: {error}")
            self.schedule_reconnect()

    def _on_open(self, ws):
        """Called when WebSocket connection is established"""
        logger.info("WebSocket connection established")
        self.connected = True
        self.reconnect_attempts = 0
        self.last_pong = time.time()

        # Start ping monitoring
        self.start_ping_monitoring()

        # Clear reconnect timer
        if self.reconnect_timer:
            self.reconnect_timer.cancel()
            self.reconnect_timer = None

        self.on_connect()

    def _on_message(self, ws, message):
        """Called when a message is received"""
        try:
            data = json.loads(message)
            self.message_count += 1

            # Handle different message types
            msg_type = data.get('type', 'unknown')

            if msg_type == 'welcome':
                welcome_data = data.get('data', {})
                logger.info(f"Received welcome: {welcome_data.get('message', 'N/A')}")
                logger.info(f"Server version: {welcome_data.get('version', 'Unknown')}")
                logger.info(f"System: {welcome_data.get('system', 'Unknown')}")
                logger.info(f"Status: {welcome_data.get('status', 'Unknown')}")
                logger.info(f"Daemon: {welcome_data.get('daemon', 'Unknown')}")

            elif msg_type == 'config':
                config_data = data.get('data', '')
                logger.info("📋 Configuration received:")
                logger.info(f"Config data length: {len(config_data)} characters")

                # Try to parse and display key configuration values
                try:
                    import yaml
                    config_parsed = yaml.safe_load(config_data)
                    if isinstance(config_parsed, dict):
                        logger.info("✅ Configuration is valid YAML")

                        # Display key monitoring settings
                        if 'monitoring' in config_parsed:
                            monitoring = config_parsed['monitoring']
                            logger.info(f"🔧 Monitoring Settings:")
                            logger.info(f"   CPU Threshold: {monitoring.get('cpu_threshold', 'N/A')}%")
                            logger.info(f"   Memory Threshold: {monitoring.get('memory_threshold', 'N/A')}%")
                            logger.info(f"   Check Interval: {monitoring.get('check_interval', 'N/A')}s")

                        # Display WebSocket settings
                        if 'websocket' in config_parsed:
                            websocket_config = config_parsed['websocket']
                            logger.info(f"🌐 WebSocket Settings:")
                            logger.info(f"   Port: {websocket_config.get('port', 'N/A')}")
                            logger.info(f"   Enabled: {websocket_config.get('enabled', 'N/A')}")

                        # Display user thresholds if any
                        if 'user_thresholds' in config_parsed and config_parsed['user_thresholds']:
                            logger.info(f"👥 Custom User Thresholds: {len(config_parsed['user_thresholds'])} users configured")
                    else:
                        logger.warning("❌ Config data is not a valid dictionary")

                except ImportError:
                    logger.warning("⚠️  PyYAML not available, displaying raw config")
                    logger.info(f"Raw config (first 300 chars): {config_data[:300]}...")
                except Exception as e:
                    logger.error(f"❌ Failed to parse config YAML: {e}")
                    logger.info(f"Raw config (first 300 chars): {config_data[:300]}...")

            elif msg_type == 'system_update':
                self.metrics = data.get('data', {})
                self._log_metrics()

            elif msg_type == 'pong':
                self.last_pong = time.time()
                logger.debug("Pong received")

            elif msg_type == 'alert':
                alert_data = data.get('data', {})
                logger.warning(f"Alert received: {alert_data}")

            else:
                logger.info(f"Unknown message type: {msg_type}")

            self.on_message(data)

        except json.JSONDecodeError as error:
            logger.error(f"Failed to parse message: {error}")
        except Exception as error:
            logger.error(f"Error handling message: {error}")

    def _on_error(self, ws, error):
        """Called when an error occurs"""
        logger.error(f"WebSocket error: {error}")
        self.on_error(error)

    def _on_close(self, ws, close_status_code, close_msg):
        """Called when WebSocket connection is closed"""
        logger.info(f"WebSocket connection closed: {close_status_code} {close_msg}")
        self.connected = False
        self.cleanup()
        self.on_disconnect()

        # Schedule reconnection unless it was a clean close
        if close_status_code != 1000:
            self.schedule_reconnect()

    def _on_ping(self, ws, message):
        """Called when a ping is received"""
        logger.debug("Ping received")

    def _on_pong(self, ws, message):
        """Called when a pong is received"""
        self.last_pong = time.time()
        logger.debug("Pong received")

    def start_ping_monitoring(self):
        """Start ping/pong monitoring timers"""
        # Send ping every ping_interval seconds
        def send_ping():
            if self.connected and self.ws:
                try:
                    self.send_message({
                        'type': 'ping',
                        'request_id': f'ping_{int(time.time())}'
                    })
                    logger.debug("Ping sent")
                except Exception as error:
                    logger.error(f"Failed to send ping: {error}")

            if self.connected:
                self.ping_timer = threading.Timer(self.ping_interval, send_ping)
                self.ping_timer.start()

        # Check for pong timeout
        def check_pong_timeout():
            if self.connected:
                time_since_pong = time.time() - self.last_pong
                if time_since_pong > self.pong_timeout + self.ping_interval:
                    logger.warning("Pong timeout detected, reconnecting...")
                    self.reconnect()
                else:
                    self.pong_timer = threading.Timer(self.pong_timeout, check_pong_timeout)
                    self.pong_timer.start()

        # Start timers
        send_ping()
        check_pong_timeout()

    def schedule_reconnect(self):
        """Schedule automatic reconnection"""
        if self.reconnect_attempts >= self.max_reconnect_attempts:
            logger.error("Max reconnection attempts reached")
            return

        self.reconnect_attempts += 1
        delay = min(self.reconnect_interval * (1.5 ** (self.reconnect_attempts - 1)), 30)

        logger.info(f"Scheduling reconnection attempt {self.reconnect_attempts}/{self.max_reconnect_attempts} in {delay:.1f}s")

        self.on_reconnecting(self.reconnect_attempts, delay)

        self.reconnect_timer = threading.Timer(delay, self.connect)
        self.reconnect_timer.start()

    def reconnect(self):
        """Manually trigger reconnection"""
        logger.info("Manual reconnection triggered")
        self.cleanup()
        self.connect()

    def send_message(self, data):
        """Send a message to the server"""
        if self.connected and self.ws:
            try:
                message = json.dumps(data)
                self.ws.send(message)
                return True
            except Exception as error:
                logger.error(f"Failed to send message: {error}")
                return False
        else:
            logger.warning("Cannot send message: not connected")
            return False

    def request_system_info(self):
        """Request system information from server"""
        return self.send_message({
            'type': 'get_system_info',
            'request_id': f'sysinfo_{int(time.time())}'
        })

    def request_alerts(self):
        """Request alerts from server"""
        return self.send_message({
            'type': 'get_alerts',
            'request_id': f'alerts_{int(time.time())}'
        })

    def request_user_metrics(self):
        """Request user metrics from server"""
        return self.send_message({
            'type': 'get_user_metrics',
            'request_id': f'usermetrics_{int(time.time())}'
        })

    def request_config(self):
        """Request configuration from server"""
        return self.send_message({
            'type': 'get_config',
            'request_id': f'config_{int(time.time())}'
        })

    def cleanup(self):
        """Clean up timers and connections"""
        if self.ping_timer:
            self.ping_timer.cancel()
            self.ping_timer = None

        if self.pong_timer:
            self.pong_timer.cancel()
            self.pong_timer = None

        if self.ws:
            try:
                if self.ws.sock and self.ws.sock.connected:
                    self.ws.close()
            except:
                pass
            self.ws = None

    def disconnect(self):
        """Disconnect from server"""
        logger.info("Disconnecting WebSocket client...")
        self.running = False
        self.connected = False

        if self.reconnect_timer:
            self.reconnect_timer.cancel()
            self.reconnect_timer = None

        self.cleanup()

    def get_connection_state(self):
        """Get current connection state"""
        return {
            'connected': self.connected,
            'reconnect_attempts': self.reconnect_attempts,
            'last_pong': self.last_pong,
            'message_count': self.message_count,
            'metrics': self.metrics
        }

    def _log_metrics(self):
        """Log current system metrics"""
        if self.metrics:
            cpu = self.metrics.get('cpu_usage', 0)
            memory = self.metrics.get('memory_usage', 0)
            disk = self.metrics.get('disk_usage', 0)
            uptime = self.metrics.get('uptime', 'Unknown')

            logger.info(f"📊 System Metrics - CPU: {cpu:.1f}%, Memory: {memory:.1f}%, Disk: {disk:.1f}%, Uptime: {uptime}")

    def run_interactive(self):
        """Run interactive command loop"""
        self.running = True
        logger.info("Interactive mode started. Type 'help' for commands.")

        try:
            while self.running:
                try:
                    command = input("> ").strip().lower()

                    if command in ['quit', 'exit', 'q']:
                        break
                    elif command == 'help':
                        self._show_help()
                    elif command == 'status':
                        self._show_status()
                    elif command == 'connect':
                        self.connect()
                    elif command == 'disconnect':
                        self.disconnect()
                    elif command == 'reconnect':
                        self.reconnect()
                    elif command == 'sysinfo':
                        self.request_system_info()
                    elif command == 'alerts':
                        self.request_alerts()
                    elif command == 'metrics':
                        self.request_user_metrics()
                    elif command == 'config':
                        self.request_config()
                    elif command == 'ping':
                        self.send_message({'type': 'ping', 'request_id': f'manual_ping_{int(time.time())}'})
                    elif command:
                        logger.warning(f"Unknown command: {command}")

                except (EOFError, KeyboardInterrupt):
                    break
                except Exception as error:
                    logger.error(f"Command error: {error}")

        finally:
            self.disconnect()

    def _show_help(self):
        """Show available commands"""
        commands = [
            "help      - Show this help",
            "status    - Show connection status",
            "connect   - Connect to server",
            "disconnect- Disconnect from server",
            "reconnect - Reconnect to server",
            "sysinfo   - Request system info",
            "alerts    - Request alerts",
            "metrics   - Request user metrics",
            "config    - Request config",
            "ping      - Send manual ping",
            "quit/exit - Exit the client"
        ]
        print("\nAvailable commands:")
        for cmd in commands:
            print(f"  {cmd}")
        print()

    def _show_status(self):
        """Show current connection status"""
        state = self.get_connection_state()
        print(f"\n📋 Connection Status:")
        print(f"  🔗 Connected: {state['connected']}")
        print(f"  🔄 Reconnect attempts: {state['reconnect_attempts']}")
        print(f"  📨 Message count: {state['message_count']}")
        print(f"  🏓 Last pong: {time.ctime(state['last_pong'])}")
        if state['metrics']:
            print(f"  📊 Latest metrics:")
            print(f"    CPU: {state['metrics'].get('cpu_usage', 0):.1f}%")
            print(f"    Memory: {state['metrics'].get('memory_usage', 0):.1f}%")
            print(f"    Disk: {state['metrics'].get('disk_usage', 0):.1f}%")
            print(f"    Uptime: {state['metrics'].get('uptime', 'Unknown')}")
        print()


def main():
    """Main function"""
    if len(sys.argv) < 3:
        print("Usage: python websocket_client.py <ws_url> <secret>")
        print("Example: python websocket_client.py ws://45.95.186.208:8060/ws 55625821f7a0a9db98707bae107e46a4")
        sys.exit(1)

    url = sys.argv[1]
    secret = sys.argv[2]

    # Create client with event handlers
    def on_connect():
        logger.info("✅ Connected to SysMedic server!")

    def on_disconnect():
        logger.info("❌ Disconnected from SysMedic server")

    def on_error(error):
        logger.error(f"💥 Connection error: {error}")

    def on_reconnecting(attempt, delay):
        logger.info(f"🔄 Reconnecting... ({attempt}/10) in {delay:.1f}s")

    client = SysMedicClient(
        url,
        secret,
        max_reconnect_attempts=10,
        reconnect_interval=5,
        ping_interval=30,
        pong_timeout=10,
        on_connect=on_connect,
        on_disconnect=on_disconnect,
        on_error=on_error,
        on_reconnecting=on_reconnecting,
        auto_connect=True
    )

    # Set up signal handlers for graceful shutdown
    def signal_handler(signum, frame):
        logger.info("\nShutdown signal received...")
        client.disconnect()
        sys.exit(0)

    signal.signal(signal.SIGINT, signal_handler)
    signal.signal(signal.SIGTERM, signal_handler)

    try:
        # Run interactive mode
        client.run_interactive()
    except KeyboardInterrupt:
        logger.info("\nKeyboard interrupt received")
    finally:
        client.disconnect()
        logger.info("Client shutdown complete")


if __name__ == "__main__":
    main()
