# SysMedic WebSocket API Guide

## Overview

This comprehensive guide covers the SysMedic WebSocket API for remote system monitoring. The WebSocket server provides real-time access to system metrics, user activity, and alert information through a standardized JSON-based protocol.

## Table of Contents

1. [Quick Start](#quick-start)
2. [WebSocket Server Management](#websocket-server-management)
3. [Connection & Authentication](#connection--authentication)
4. [Message Format Specification](#message-format-specification)
5. [Real-time Message Types](#real-time-message-types)
6. [Request-Response Protocol](#request-response-protocol)
7. [Client Examples](#client-examples)
8. [Testing & Validation](#testing--validation)
9. [Connection Management](#connection-management)
10. [Troubleshooting](#troubleshooting)
11. [Security Considerations](#security-considerations)
12. [Performance & Optimization](#performance--optimization)

## Quick Start

### 1. Start WebSocket Server

```bash
# Start with default settings (port 8060)
sysmedic websocket start

# Start on custom port
sysmedic websocket start 9090

# Check status and get connection details
sysmedic websocket status
```

### 2. Get Connection Information

```bash
$ sysmedic websocket status
WebSocket Daemon Status: 🟢 running (PID: 12345)
Quick Connect: sysmedic://d8852f78260f16d31eeff80ca6158848@203.0.113.45:8060/
Port: 8060
Secret: d8852f78260f16d31eeff80ca6158848
Local URL: ws://localhost:8060/ws
Public URL: ws://203.0.113.45:8060/ws
Auth Header: X-SysMedic-Secret: d8852f78260f16d31eeff80ca6158848
```

### 3. Basic Connection Test

```bash
# Test WebSocket connectivity
python examples/websocket_client.py ws://localhost:8060/ws d8852f78260f16d31eeff80ca6158848

# Test with curl (HTTP endpoints)
curl http://localhost:8060/health
curl http://localhost:8060/status
```

## WebSocket Server Management

### CLI Commands

#### Start WebSocket Daemon
```bash
# Default port (8060)
sysmedic websocket start

# Custom port
sysmedic websocket start 9090

# Background daemon mode
sysmedic --websocket-daemon

# Expected output:
Starting WebSocket daemon on port 8060...
✓ WebSocket daemon started successfully
Quick Connect: sysmedic://[secret]@[ip]:8060/
```

#### Check Status
```bash
sysmedic websocket status

# Detailed output includes:
# - Daemon status and PID
# - Quick Connect URL
# - Connection endpoints
# - Authentication details
# - Client connection count
```

#### Stop WebSocket Daemon
```bash
sysmedic websocket stop

# Expected output:
Stopping WebSocket daemon...
✓ WebSocket daemon stopped successfully
```

#### Generate New Secret
```bash
sysmedic websocket new-secret

# Automatically:
# - Generates new cryptographic secret
# - Updates configuration
# - Restarts WebSocket daemon
# - Displays new connection details
```

### SystemD Service Integration

```bash
# Enable WebSocket service
sudo systemctl enable sysmedic.websocket

# Start service
sudo systemctl start sysmedic.websocket

# Check service status
sudo systemctl status sysmedic.websocket

# View logs
sudo journalctl -u sysmedic.websocket -f
```

## Connection & Authentication

### Connection URL Format

#### Standard WebSocket URL
```
ws://[hostname]:[port]/ws?secret=[authentication_secret]
```

#### Quick Connect URL
```
sysmedic://[secret]@[hostname]:[port]/
```

**Example URLs:**
```bash
# Standard WebSocket
ws://192.168.1.100:8060/ws?secret=d8852f78260f16d31eeff80ca6158848

# Quick Connect
sysmedic://d8852f78260f16d31eeff80ca6158848@192.168.1.100:8060/
```

### Authentication Methods

#### Query Parameter (Recommended)
```javascript
const ws = new WebSocket('ws://host:8060/ws?secret=your_secret_here');
```

#### HTTP Header
```javascript
const ws = new WebSocket('ws://host:8060/ws', [], {
    headers: {
        'X-SysMedic-Secret': 'your_secret_here'
    }
});
```

### HTTP Endpoints

#### Health Check Endpoint
```bash
GET /health

# Response:
{
  "status": "healthy",
  "timestamp": "2025-01-19T00:08:42Z"
}
```

#### Status Information Endpoint
```bash
GET /status

# Response:
{
  "status": "running",
  "clients": 2,
  "port": 8060,
  "hostname": "server.example.com",
  "has_secret": true,
  "version": "1.0.5",
  "uptime": "2d 4h 32m"
}
```

## Message Format Specification

### Base Message Structure

All WebSocket messages follow this JSON structure:
```json
{
  "type": "message_type",
  "data": { ... },
  "timestamp": "2025-01-19T00:08:42Z"
}
```

**Fields:**
- `type`: String identifying the message type
- `data`: Object containing message-specific data
- `timestamp`: ISO 8601 formatted timestamp in UTC

### Message Flow

```
Client Connect → Authentication → Welcome Message → System Updates (every 3s)
                                                  ↓
                                               Request-Response (on-demand)
```

## Real-time Message Types

### 1. Welcome Message

**Sent**: Immediately upon successful connection  
**Frequency**: Once per connection

```json
{
  "type": "welcome",
  "data": {
    "message": "Connected to SysMedic",
    "version": "1.0.5",
    "system": "Ubuntu 22.04.5 LTS",
    "status": "Light Usage",
    "daemon": "Running"
  },
  "timestamp": "2025-01-19T00:08:42Z"
}
```

**Data Fields:**
- `message`: Always "Connected to SysMedic"
- `version`: SysMedic version string
- `system`: Operating system information from `/etc/os-release`
- `status`: Current system status classification
- `daemon`: Daemon status (always "Running" when connected)

### 2. System Update Messages

**Sent**: Every 3 seconds after connection  
**Frequency**: Continuous real-time updates

```json
{
  "type": "system_update",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 32.1,
    "uptime": "2d 4h 32m"
  },
  "timestamp": "2025-01-19T00:08:45Z"
}
```

**Data Fields:**
- `cpu_usage`: CPU usage percentage (0.0 - 100.0)
- `memory_usage`: Memory usage percentage (0.0 - 100.0)
- `disk_usage`: Root filesystem usage percentage (0.0 - 100.0)
- `uptime`: Human-readable system uptime

**Uptime Format Examples:**
- `"2d 4h 32m"` - 2 days, 4 hours, 32 minutes
- `"4h 32m"` - 4 hours, 32 minutes
- `"32m"` - 32 minutes

### 3. Alert Messages

**Sent**: When system status changes or alerts are triggered  
**Frequency**: Event-driven

```json
{
  "type": "alert",
  "data": {
    "status": "Heavy Load",
    "system_metrics": {
      "cpu_usage": 85.3,
      "memory_usage": 91.2,
      "disk_usage": 45.6
    },
    "user_metrics": [
      {
        "username": "dbuser",
        "cpu_percent": 82.4,
        "memory_percent": 45.2,
        "processes": 15
      }
    ],
    "persistent_users": ["dbuser", "webuser"],
    "primary_cause": "High system memory usage",
    "recommendations": [
      "Check memory-intensive processes",
      "Consider restarting services",
      "Monitor user: dbuser"
    ]
  },
  "timestamp": "2025-01-19T00:09:15Z"
}
```

## Request-Response Protocol

The WebSocket API supports bidirectional communication through a request-response protocol for on-demand data retrieval.

### Request Format

```json
{
  "type": "request_type",
  "request_id": "unique_identifier",
  "data": { ... }
}
```

**Fields:**
- `type`: Request type identifier
- `request_id`: Unique identifier for matching responses
- `data`: Request-specific parameters (optional)

### Response Format

```json
{
  "type": "response",
  "data": { ... },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "unique_identifier"
}
```

**Fields:**
- `type`: Always "response"
- `data`: Response data specific to request type
- `timestamp`: Response timestamp
- `request_id`: Matches the request identifier

### Available Request Types

#### 1. System Information (`get_system_info`)

**Request:**
```json
{
  "type": "get_system_info",
  "request_id": "req_001"
}
```

**Response:**
```json
{
  "type": "response",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 32.1,
    "uptime": "2d 4h 32m",
    "load_avg": [1.25, 1.45, 1.32],
    "network_io": {
      "bytes_sent": 1024000,
      "bytes_recv": 2048000
    }
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_001"
}
```

#### 2. Alerts Information (`get_alerts`)

**Request:**
```json
{
  "type": "get_alerts",
  "request_id": "req_002"
}
```

**Response:**
```json
{
  "type": "response",
  "data": {
    "unresolved_count": 2,
    "total_count": 15,
    "recent_alerts": [
      {
        "timestamp": "2025-01-19T00:05:00Z",
        "severity": "high",
        "message": "System CPU usage above threshold"
      }
    ],
    "status": "Heavy Load"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_002"
}
```

#### 3. User Metrics (`get_user_metrics`)

**Request:**
```json
{
  "type": "get_user_metrics",
  "request_id": "req_003"
}
```

**Response:**
```json
{
  "type": "response",
  "data": [
    {
      "username": "dbuser",
      "cpu_percent": 82.4,
      "memory_percent": 45.2,
      "processes": 15
    },
    {
      "username": "webuser",
      "cpu_percent": 25.1,
      "memory_percent": 30.8,
      "processes": 8
    }
  ],
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_003"
}
```

#### 4. Server Configuration (`get_config`)

**Request:**
```json
{
  "type": "get_config",
  "request_id": "req_004"
}
```

**Response:**
```json
{
  "type": "response",
  "data": {
    "monitoring_interval": 60,
    "cpu_threshold": 80,
    "memory_threshold": 80,
    "version": "1.0.5"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_004"
}
```

#### 5. Ping Test (`ping`)

**Request:**
```json
{
  "type": "ping",
  "request_id": "req_005"
}
```

**Response:**
```json
{
  "type": "response",
  "data": {
    "message": "pong"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_005"
}
```

### Error Responses

When a request fails, the server returns an error response:

```json
{
  "type": "error",
  "data": {
    "error": "Invalid request type"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_invalid"
}
```

## Client Examples

### JavaScript (Browser)

```html
<!DOCTYPE html>
<html>
<head>
    <title>SysMedic WebSocket Client</title>
    <style>
        body { font-family: monospace; margin: 20px; }
        .status { padding: 10px; margin: 10px 0; border-radius: 5px; }
        .connected { background: #d4edda; color: #155724; }
        .disconnected { background: #f8d7da; color: #721c24; }
        .metrics { display: grid; grid-template-columns: repeat(4, 1fr); gap: 10px; }
        .metric { padding: 10px; background: #f8f9fa; border-radius: 5px; }
    </style>
</head>
<body>
    <h1>SysMedic Monitor</h1>
    <div id="status" class="status disconnected">Connecting...</div>
    <div id="system-info"></div>
    <div id="metrics" class="metrics"></div>
    <div id="controls">
        <button onclick="requestSystemInfo()">Get System Info</button>
        <button onclick="requestUserMetrics()">Get User Metrics</button>
        <button onclick="requestAlerts()">Get Alerts</button>
        <button onclick="pingServer()">Ping Server</button>
    </div>
    <div id="responses"></div>

    <script>
        let ws;
        let requestCounter = 0;
        const responses = {};

        function connect() {
            // Replace with your actual connection details
            const wsUrl = 'ws://localhost:8060/ws?secret=your_secret_here';
            
            ws = new WebSocket(wsUrl);

            ws.onopen = () => {
                updateStatus('Connected to SysMedic', 'connected');
                console.log('WebSocket connected');
            };

            ws.onmessage = (event) => {
                const message = JSON.parse(event.data);
                handleMessage(message);
            };

            ws.onclose = () => {
                updateStatus('Disconnected from SysMedic', 'disconnected');
                console.log('WebSocket disconnected');
                // Auto-reconnect after 5 seconds
                setTimeout(connect, 5000);
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
                updateStatus('Connection error', 'disconnected');
            };
        }

        function handleMessage(message) {
            console.log('Received:', message);

            switch (message.type) {
                case 'welcome':
                    displaySystemInfo(message.data);
                    break;

                case 'system_update':
                    displayMetrics(message.data);
                    break;

                case 'alert':
                    displayAlert(message.data);
                    break;

                case 'response':
                    displayResponse(message);
                    break;

                case 'error':
                    displayError(message);
                    break;
            }
        }

        function sendRequest(type, data = {}) {
            if (ws.readyState !== WebSocket.OPEN) {
                console.error('WebSocket not connected');
                return;
            }

            const requestId = `req_${++requestCounter}`;
            const request = {
                type: type,
                request_id: requestId,
                data: data
            };

            ws.send(JSON.stringify(request));
            return requestId;
        }

        function updateStatus(message, className) {
            const statusDiv = document.getElementById('status');
            statusDiv.textContent = message;
            statusDiv.className = `status ${className}`;
        }

        function displaySystemInfo(data) {
            document.getElementById('system-info').innerHTML = `
                <h3>System Information</h3>
                <p><strong>System:</strong> ${data.system}</p>
                <p><strong>Status:</strong> ${data.status}</p>
                <p><strong>Version:</strong> ${data.version}</p>
            `;
        }

        function displayMetrics(data) {
            document.getElementById('metrics').innerHTML = `
                <div class="metric">
                    <h4>CPU Usage</h4>
                    <p>${data.cpu_usage.toFixed(1)}%</p>
                </div>
                <div class="metric">
                    <h4>Memory Usage</h4>
                    <p>${data.memory_usage.toFixed(1)}%</p>
                </div>
                <div class="metric">
                    <h4>Disk Usage</h4>
                    <p>${data.disk_usage.toFixed(1)}%</p>
                </div>
                <div class="metric">
                    <h4>Uptime</h4>
                    <p>${data.uptime}</p>
                </div>
            `;
        }

        function displayAlert(data) {
            const alertDiv = document.createElement('div');
            alertDiv.className = 'alert';
            alertDiv.innerHTML = `
                <h4>Alert: ${data.status}</h4>
                <p><strong>Primary Cause:</strong> ${data.primary_cause}</p>
                <p><strong>Recommendations:</strong></p>
                <ul>${data.recommendations.map(r => `<li>${r}</li>`).join('')}</ul>
            `;
            document.body.appendChild(alertDiv);
        }

        function displayResponse(message) {
            const responsesDiv = document.getElementById('responses');
            responsesDiv.innerHTML = `
                <h4>Response (${message.request_id})</h4>
                <pre>${JSON.stringify(message.data, null, 2)}</pre>
            `;
        }

        function displayError(message) {
            console.error('Server error:', message.data.error);
            alert(`Server error: ${message.data.error}`);
        }

        // Request functions
        function requestSystemInfo() {
            sendRequest('get_system_info');
        }

        function requestUserMetrics() {
            sendRequest('get_user_metrics');
        }

        function requestAlerts() {
            sendRequest('get_alerts');
        }

        function pingServer() {
            sendRequest('ping');
        }

        // Start connection when page loads
        window.onload = connect;
    </script>
</body>
</html>
```

### Python Client

```python
#!/usr/bin/env python3
"""
SysMedic WebSocket Python Client
Comprehensive example with error handling and reconnection
"""

import websocket
import json
import threading
import time
import uuid
import signal
import sys

class SysMedicClient:
    def __init__(self, ws_url, secret):
        self.ws_url = f"{ws_url}?secret={secret}"
        self.ws = None
        self.running = False
        self.pending_requests = {}
        
    def connect(self):
        """Connect to SysMedic WebSocket server"""
        print(f"🔌 Connecting to {self.ws_url}")
        
        # Configure WebSocket
        websocket.enableTrace(False)
        self.ws = websocket.WebSocketApp(
            self.ws_url,
            on_message=self.on_message,
            on_error=self.on_error,
            on_close=self.on_close,
            on_open=self.on_open
        )
        
        # Start WebSocket in a separate thread
        self.running = True
        wst = threading.Thread(target=self.ws.run_forever)
        wst.daemon = True
        wst.start()
        
    def disconnect(self):
        """Disconnect from WebSocket server"""
        self.running = False
        if self.ws:
            self.ws.close()
            
    def on_open(self, ws):
        """Handle WebSocket connection open"""
        print("✅ Connected to SysMedic")
        
    def on_message(self, ws, message):
        """Handle incoming WebSocket messages"""
        try:
            data = json.loads(message)
            self.handle_message(data)
        except json.JSONDecodeError as e:
            print(f"❌ Invalid JSON received: {e}")
            
    def on_error(self, ws, error):
        """Handle WebSocket errors"""
        print(f"❌ WebSocket error: {error}")
        
    def on_close(self, ws, close_status_code, close_msg):
        """Handle WebSocket connection close"""
        print(f"🔌 Connection closed: {close_status_code} - {close_msg}")
        
        # Auto-reconnect if running
        if self.running:
            print("🔄 Reconnecting in 5 seconds...")
            time.sleep(5)
            self.connect()
            
    def handle_message(self, message):
        """Handle different message types"""
        msg_type = message.get('type')
        
        if msg_type == 'welcome':
            self.handle_welcome(message['data'])
        elif msg_type == 'system_update':
            self.handle_system_update(message['data'])
        elif msg_type == 'alert':
            self.handle_alert(message['data'])
        elif msg_type == 'response':
            self.handle_response(message)
        elif msg_type == 'error':
            self.handle_error(message)
        else:
            print(f"❓ Unknown message type: {msg_type}")
            
    def handle_welcome(self, data):
        """Handle welcome message"""
        print(f"🖥️  Connected to {data['system']}")
        print(f"📊 Status: {data['status']}")
        print(f"🏷️  Version: {data['version']}")
        
    def handle_system_update(self, data):
        """Handle system update messages"""
        print(f"📊 CPU: {data['cpu_usage']:.1f}%, "
              f"Memory: {data['memory_usage']:.1f}%, "
              f"Disk: {data['disk_usage']:.1f}%, "
              f"Uptime: {data['uptime']}")
              
    def handle_alert(self, data):
        """Handle alert messages"""
        print(f"🚨 ALERT: {data['status']}")
        print(f"⚠️  Primary Cause: {data['primary_cause']}")
        if data.get('persistent_users'):
            print(f"👥 Persistent Users: {', '.join(data['persistent_users'])}")
        for rec in data.get('recommendations', []):
            print(f"💡 {rec}")
            
    def handle_response(self, message):
        """Handle response messages"""
        request_id = message.get('request_id')
        data = message.get('data')
        
        print(f"📨 Response for {request_id}:")
        print(json.dumps(data, indent=2))
        
        # Store response for pending requests
        if request_id in self.pending_requests:
            self.pending_requests[request_id] = data
            
    def handle_error(self, message):
        """Handle error messages"""
        error = message['data'].get('error')
        request_id = message.get('request_id')
        print(f"❌ Error for {request_id}: {error}")
        
    def send_request(self, request_type, data=None):
        """Send request to server"""
        if not self.ws or self.ws.sock is None:
            print("❌ WebSocket not connected")
            return None
            
        request_id = f"req_{uuid.uuid4().hex[:8]}"
        request = {
            'type': request_type,
            'request_id': request_id,
            'data': data or {}
        }
        
        try:
            self.ws.send(json.dumps(request))
            self.pending_requests[request_id] = None
            print(f"📤 Sent request: {request_type} ({request_id})")
            return request_id
        except Exception as e:
            print(f"❌ Failed to send request: {e}")
            return None
            
    def request_system_info(self):
        """Request system information"""
        return self.send_request('get_system_info')
        
    def request_user_metrics(self):
        """Request user metrics"""
        return self.send_request('get_user_metrics')
        
    def request_alerts(self):
        """Request alerts information"""
        return self.send_request('get_alerts')
        
    def request_config(self):
        """Request server configuration"""
        return self.send_request('get_config')
        
    def ping(self):
        """Send ping request"""
        return self.send_request('ping')

def interactive_mode(client):
    """Interactive command mode"""
    print("\n" + "="*50)
    print("SysMedic Interactive Mode")
    print("Commands: info, users, alerts, config, ping, quit")
    print("="*50)
    
    while client.running:
        try:
            cmd = input("\nsysmedic> ").strip().lower()
            
            if cmd == 'quit' or cmd == 'exit':
                break
            elif cmd == 'info':
                client.request_system_info()
            elif cmd == 'users':
                client.request_user_metrics()
            elif cmd == 'alerts':
                client.request_alerts()
            elif cmd == 'config':
                client.request_config()
            elif cmd == 'ping':
                client.ping()
            elif cmd == 'help':
                print("Commands: info, users, alerts, config, ping, quit")
            else:
                print(f"Unknown command: {cmd}")
                
        except KeyboardInterrupt:
            break
            
    print("\n👋 Goodbye!")

def signal_handler(sig, frame):
    """Handle Ctrl+C gracefully"""
    print('\n👋 Shutting down...')
    sys.exit(0)

def main():
    """Main function"""
    import argparse
    
    parser = argparse.ArgumentParser(description='SysMedic WebSocket Client')
    parser.add_argument('url', help='WebSocket URL (ws://host:port/ws)')
    parser.add_argument('secret', help='Authentication secret')
    parser.add_argument('--interactive', '-i', action='store_true', 
                       help='Enable interactive mode')
    
    args = parser.parse_args()
    
    # Set up signal handling
    signal.signal(signal.SIGINT, signal_handler)
    
    # Create and connect client
    client = SysMedicClient(args.url, args.secret)
    client.connect()
    
    # Wait for connection
    time.sleep(2)
    
    if args.interactive:
        interactive_mode(client)
    else:
        # Non-interactive: just listen for messages
        print("📡 Listening for messages... (Press Ctrl+C to quit)")
        try:
            while client.running:
                time.sleep(1)
        except KeyboardInterrupt:
            pass
    
    client.disconnect()

if __name__ == "__main__":
    main()
```

### Node.js Client

```javascript
const WebSocket = require('ws');
const { v4: uuidv4 } = require('uuid');

class SysMedicClient {
    constructor(wsUrl, secret) {
        this.wsUrl = `${wsUrl}?secret=${secret}`;
        this.ws = null;
        this.running = false;
        this.pendingRequests = new Map();
        this.reconnectDelay = 5000;
        this.maxReconnectDelay = 60000;
        this.currentReconnectDelay = this.reconnectDelay;
    }

    connect() {
        console.log(`🔌 Connecting to ${this.wsUrl}`);
        
        this.ws = new WebSocket(this.wsUrl);
        this.running = true;

        this.ws.on('open', () => {
            console.log('✅ Connected to SysMedic');
            this.currentReconnectDelay = this.reconnectDelay; // Reset delay on successful connection
        });

        this.ws.on('message', (data) => {
            try {
                const message = JSON.parse(data.toString());
                this.handleMessage(message);
            } catch (error) {
                console.error('❌ Invalid JSON received:', error);
            }
        });

        this.ws.on('close', (code, reason) => {
            console.log(`🔌 Connection closed: ${code} - ${reason}`);
            
            if (this.running) {
                console.log(`🔄 Reconnecting in ${this.currentReconnectDelay / 1000} seconds...`);
                setTimeout(() => {
                    if (this.running) {
                        this.connect();
                        // Exponential backoff with jitter
                        this.currentReconnectDelay = Math.min(
                            this.currentReconnectDelay * 1.5 + Math.random() * 1000,
                            this.maxReconnectDelay
                        );
                    }
                }, this.currentReconnectDelay);
            }
        });

        this.ws.on('error', (error) => {
            console.error('❌ WebSocket error:', error);
        });

        // Ping/pong handling
        this.ws.on('ping', () => {
            this.ws.pong();
        });
    }

    disconnect() {
        this.running = false;
        if (this.ws) {
            this.ws.close();
        }
    }

    handleMessage(message) {
        const { type, data, timestamp, request_id } = message;

        switch (type) {
            case 'welcome':
                this.handleWelcome(data);
                break;
            
            case 'system_update':
                this.handleSystemUpdate(data);
                break;
                
            case 'alert':
                this.handleAlert(data);
                break;
                
            case 'response':
                this.handleResponse(message);
                break;
                
            case 'error':
                this.handleError(message);
                break;
                
            default:
                console.log(`❓ Unknown message type: ${type}`);
        }
    }

    handleWelcome(data) {
        console.log(`🖥️  Connected to ${data.system}`);
        console.log(`📊 Status: ${data.status}`);
        console.log(`🏷️  Version: ${data.version}`);
    }

    handleSystemUpdate(data) {
        console.log(`📊 CPU: ${data.cpu_usage.toFixed(1)}%, ` +
                   `Memory: ${data.memory_usage.toFixed(1)}%, ` +
                   `Disk: ${data