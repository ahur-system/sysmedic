# SysMedic WebSocket Documentation

## Overview

SysMedic includes a WebSocket server that enables real-time remote monitoring of your system through secure connections. This comprehensive guide covers setup, usage, API documentation, testing, and implementation details.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Features](#features)
3. [CLI Commands](#cli-commands)
4. [WebSocket API](#websocket-api)
5. [Request-Response Protocol](#request-response-protocol)
6. [Testing Guide](#testing-guide)
7. [Client Examples](#client-examples)
8. [Implementation Details](#implementation-details)
9. [Security](#security)
10. [Troubleshooting](#troubleshooting)

## Quick Start

### 1. Configure and Start WebSocket Server

```bash
# Configure WebSocket server
sysmedic websocket start 8060

# Start the daemon (this will also start the WebSocket server)
sysmedic daemon start

# Get the connection URL
sysmedic websocket status
```

### 2. Basic Test

```bash
# Test health endpoint
curl http://localhost:8060/health

# Should return: {"status":"healthy","running":true,"clients":0,...}
```

### 3. Connect with Python Client

```bash
# Install dependencies
pip3 install websocket-client

# Connect using the URL from status command
python3 examples/websocket_client.py "sysmedic://secret@hostname:8060/"
```

## Features

- **Secure Authentication**: Each WebSocket connection uses a unique secret token
- **Real-time Monitoring**: Live system metrics streamed every 5 seconds
- **Alert Broadcasting**: Automatic alert notifications to all connected clients
- **Bidirectional Communication**: Request specific information on-demand
- **Daemon Integration**: WebSocket server runs as part of the SysMedic daemon
- **Health Monitoring**: HTTP health check endpoint
- **Multiple Clients**: Support for concurrent client connections

## CLI Commands

### Start WebSocket Server

```bash
# Start with default port (8060)
sysmedic websocket start

# Start with custom port
sysmedic websocket start 9090
```

**Note**: The WebSocket server runs as part of the SysMedic daemon. If the daemon is not running, you'll need to start it.

### Check Status

```bash
sysmedic websocket status
```

Shows:
- Daemon status (running/stopped)
- WebSocket configuration (enabled/disabled)
- Current connection status
- Active client count
- Connection URL (if running)

### Stop WebSocket Server

```bash
sysmedic websocket stop
```

Disables the WebSocket server in configuration. The server will stop when the daemon restarts.

### Generate New Secret

```bash
sysmedic websocket new-secret
```

Generates a new authentication secret and restarts the WebSocket server. All existing clients will be disconnected.

## WebSocket API

### Connection URL Format

```
sysmedic://[secret]@[hostname]:[port]/
```

**Example:**
```
sysmedic://d8852f78260f16d31eeff80ca6158848@server.example.com:8060/
```

### WebSocket Endpoint

```
ws://hostname:port/ws?secret=your_secret_here
```

### Health Check Endpoint

```
GET http://hostname:port/health
```

Returns:
```json
{
  "status": "healthy",
  "running": true,
  "clients": 2,
  "port": 8060,
  "hostname": "server.example.com",
  "has_secret": true
}
```

### Message Format

All messages follow this JSON structure:

```json
{
  "type": "message_type",
  "data": { ... },
  "timestamp": "2025-01-19T00:08:42Z"
}
```

### Real-time Message Types

#### Welcome Message
Sent immediately after successful connection:

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

#### System Updates
Sent every 5 seconds with current system metrics:

```json
{
  "type": "system_update",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 32.1,
    "uptime": "2d 4h 32m"
  },
  "timestamp": "2025-01-19T00:08:42Z"
}
```

#### Alerts
Sent when system alerts are triggered:

```json
{
  "type": "alert",
  "data": {
    "status": "Heavy Load",
    "system_metrics": { ... },
    "user_metrics": [ ... ],
    "persistent_users": [ ... ],
    "primary_cause": "user123 (cpu: 95.2% for 5m)",
    "recommendations": [ ... ]
  },
  "timestamp": "2025-01-19T00:08:42Z"
}
```

## Request-Response Protocol

SysMedic supports bidirectional communication, allowing clients to request specific information.

### Request Format

```json
{
  "type": "request_type",
  "request_id": "unique_identifier",
  "data": { /* optional request data */ }
}
```

### Response Format

```json
{
  "type": "response_type",
  "data": { /* response data */ },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "unique_identifier"
}
```

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
  "type": "system_info_response",
  "data": {
    "cpu_usage": 45.2,
    "memory_usage": 67.8,
    "disk_usage": 32.1,
    "uptime": "2d 4h 32m",
    "load_avg": "0.85, 0.92, 1.15",
    "network_io": "2.34 MB/s"
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
  "type": "alerts_response",
  "data": {
    "unresolved_count": 3,
    "total_count": 15,
    "recent_alerts": [
      "High CPU usage detected",
      "Memory threshold exceeded"
    ],
    "status": "3 unresolved alerts"
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
  "type": "user_metrics_response",
  "data": [
    {
      "username": "database_user",
      "cpu_percent": 85.4,
      "memory_percent": 45.2,
      "processes": 12
    },
    {
      "username": "web_server",
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
  "type": "config_response",
  "data": {
    "monitoring_interval": "1m0s",
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
  "type": "pong",
  "data": {
    "message": "pong"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_005"
}
```

### Error Responses

If a request fails, the server returns an error response:

```json
{
  "type": "error",
  "data": {
    "error": "Unknown request type: invalid_request"
  },
  "timestamp": "2025-01-19T00:08:42Z",
  "request_id": "req_001"
}
```

## Testing Guide

### Method 1: Health Check (Basic Test)
```bash
# Test if the server is responding
curl http://localhost:8060/health
```

### Method 2: Python Client (Recommended)
```bash
# Install WebSocket client
pip3 install websocket-client

# Run the provided Python client
python3 examples/websocket_client.py "sysmedic://secret@hostname:8060/"
```

### Method 3: HTML Browser Client
```bash
# Start a local web server
python3 -m http.server 3000

# Open in browser: http://localhost:3000/examples/websocket_client.html
# Enter your connection URL and click "Connect"
```

### Method 4: Command Line WebSocket Tools
```bash
# Using websocat
websocat "ws://localhost:8060/ws?secret=your_secret"

# Using wscat
wscat -c "ws://localhost:8060/ws?secret=your_secret"
```

### Advanced Testing

#### Test with Multiple Clients:
```bash
# Terminal 1
python3 examples/websocket_client.py "sysmedic://secret@host:8060/" &

# Terminal 2  
python3 examples/websocket_client.py "sysmedic://secret@host:8060/" &

# Check client count
curl http://localhost:8060/health | grep clients
```

#### Test Secret Rotation:
```bash
# Connect client with current secret
python3 examples/websocket_client.py "sysmedic://secret@host:8060/" &

# Generate new secret (disconnects old clients)
sysmedic websocket new-secret

# Connect with new secret
sysmedic websocket status  # Get new URL
python3 examples/websocket_client.py "sysmedic://newsecret@host:8060/"
```

## Client Examples

### JavaScript (Browser)

```javascript
let ws = new WebSocket('ws://localhost:8060/ws?secret=your_secret');
let requestCounter = 0;

function sendRequest(type, data = null) {
    const requestId = `req_${++requestCounter}_${Date.now()}`;
    const request = {
        type: type,
        request_id: requestId,
        data: data
    };
    
    ws.send(JSON.stringify(request));
    return requestId;
}

// Handle responses
ws.onmessage = function(event) {
    const message = JSON.parse(event.data);
    
    if (message.request_id) {
        console.log(`Response for ${message.request_id}:`, message.data);
    } else {
        console.log(`Real-time update:`, message);
    }
};

// Example requests
sendRequest('get_system_info');
sendRequest('get_alerts');
sendRequest('ping');
```

### Python

```python
import json
import websocket
import time

class SysMedicClient:
    def __init__(self, ws_url):
        self.ws = websocket.WebSocket()
        self.ws.connect(ws_url)
        self.request_counter = 0
    
    def send_request(self, request_type, data=None):
        self.request_counter += 1
        request_id = f"req_{self.request_counter}_{int(time.time())}"
        
        request = {
            "type": request_type,
            "request_id": request_id,
            "data": data
        }
        
        self.ws.send(json.dumps(request))
        return request_id
    
    def receive_response(self):
        response = self.ws.recv()
        return json.loads(response)

# Usage
client = SysMedicClient('ws://localhost:8060/ws?secret=your_secret')
client.send_request('get_system_info')

while True:
    message = client.receive_response()
    print(f"Received: {message}")
```

### Node.js

```javascript
const WebSocket = require('ws');

class SysMedicClient {
    constructor(wsUrl) {
        this.ws = new WebSocket(wsUrl);
        this.requestCounter = 0;
        this.pendingRequests = new Map();
        
        this.ws.on('message', (data) => {
            this.handleMessage(JSON.parse(data));
        });
    }
    
    sendRequest(type, data = null) {
        const requestId = `req_${++this.requestCounter}_${Date.now()}`;
        const request = { type, request_id: requestId, data };
        
        this.ws.send(JSON.stringify(request));
        
        return new Promise((resolve) => {
            this.pendingRequests.set(requestId, resolve);
        });
    }
    
    handleMessage(message) {
        if (message.request_id && this.pendingRequests.has(message.request_id)) {
            const resolve = this.pendingRequests.get(message.request_id);
            this.pendingRequests.delete(message.request_id);
            resolve(message);
        } else {
            console.log('Real-time update:', message);
        }
    }
}

// Usage
const client = new SysMedicClient('ws://localhost:8060/ws?secret=your_secret');

(async () => {
    const systemInfo = await client.sendRequest('get_system_info');
    console.log('System Info:', systemInfo.data);
})();
```

## Implementation Details

### Architecture

#### Core Components

1. **WebSocket Server** (`internal/websocket/server.go`)
   - Handles WebSocket connections with secret-based authentication
   - Broadcasts real-time system metrics every 5 seconds
   - Sends alerts when system status changes
   - Manages client connections and cleanup

2. **WebSocket Manager** (`internal/websocket/manager.go`)
   - Manages server lifecycle and configuration
   - Handles configuration persistence in `~/.sysmedic/websocket.json`
   - Provides singleton pattern for system-wide access

3. **Daemon Integration** (`internal/daemon/daemon.go`)
   - Automatically starts WebSocket server when daemon starts (if enabled)
   - Broadcasts alerts to connected WebSocket clients
   - Cleanly shuts down WebSocket server with daemon

4. **CLI Interface** (`pkg/cli/cli.go`, `cmd/sysmedic/main.go`)
   - Provides commands: `start`, `stop`, `status`, `new-secret`
   - Integrates with existing daemon management
   - Shows comprehensive status information

### Configuration

WebSocket configuration is stored in `~/.sysmedic/websocket.json`:

```json
{
  "port": 8060,
  "enabled": true
}
```

### Dependencies

- `github.com/gorilla/websocket v1.5.0` - WebSocket protocol implementation

### Performance Characteristics

- **Memory**: Minimal overhead per client connection
- **CPU**: Periodic broadcasts every 5 seconds
- **Network**: JSON messages typically <1KB each
- **Scalability**: Tested with multiple concurrent connections

## Security

### Security Considerations

1. **Secret Management**: Authentication secrets are randomly generated and stored only in memory
2. **Network Security**: Consider using a reverse proxy with SSL/TLS for production deployments
3. **Firewall**: Ensure the WebSocket port is properly secured in your firewall rules
4. **Secret Rotation**: Regularly generate new secrets using `sysmedic websocket new-secret`

### Production Security Recommendations

- Use HTTPS/WSS in production environments
- Implement firewall rules for WebSocket port
- Regular secret rotation
- Monitor connection logs
- Consider rate limiting for high-traffic scenarios

## Troubleshooting

### WebSocket Server Not Starting

1. **Check daemon status:**
   ```bash
   sysmedic daemon status
   ```

2. **Verify configuration:**
   ```bash
   sysmedic websocket status
   ```

3. **Check port availability:**
   ```bash
   netstat -tlnp | grep :8060
   # or
   ss -tlnp | grep :8060
   ```

4. **Review daemon logs:**
   ```bash
   journalctl -u sysmedic -f
   ```

### Connection Issues

1. **Verify the secret:** Check the connection URL for correct secret
2. **Check firewall rules:** Ensure WebSocket port is accessible
3. **Test hostname/IP:** Verify hostname is accessible from client
4. **Try health endpoint:**
   ```bash
   curl http://hostname:8060/health
   ```

### Client Disconnections

1. **Secret regeneration:** Clients disconnect when secrets are regenerated
2. **Server restarts:** Server restarts will disconnect all clients
3. **Network interruptions:** May cause temporary disconnections

### "Connection refused" Error

```bash
# Check if daemon is running
sysmedic daemon status

# Check WebSocket configuration
sysmedic websocket status

# Restart if needed
sysmedic daemon stop
sysmedic daemon start
```

### "Invalid secret" Error

```bash
# Generate new secret
sysmedic websocket new-secret

# Get new connection URL
sysmedic websocket status
```

### Port Already in Use

```bash
# Check what's using the port
netstat -tlnp | grep :8060

# Try a different port
sysmedic websocket start 9090
sysmedic daemon stop
sysmedic daemon start
```

## Example Usage Workflow

1. **Setup**: Start the SysMedic daemon
   ```bash
   sysmedic daemon start
   ```

2. **Enable WebSocket**: Configure WebSocket server
   ```bash
   sysmedic websocket start 8060
   ```

3. **Restart Daemon**: Apply configuration
   ```bash
   sysmedic daemon stop
   sysmedic daemon start
   ```

4. **Get Connection URL**: Check status for connection details
   ```bash
   sysmedic websocket status
   ```

5. **Connect Client**: Use the generated URL to connect your remote client

## Version History

- **v1.0.3**: Initial WebSocket feature implementation
  - Secure authentication with generated secrets
  - Real-time system monitoring  
  - Alert broadcasting
  - Daemon integration

- **v1.0.4**: Enhanced WebSocket functionality
  - Bidirectional communication support
  - Request-response protocol
  - Additional client examples
  - Improved error handling

- **v1.0.5**: Documentation reorganization and project cleanup
  - Consolidated WebSocket documentation into single comprehensive guide
  - Improved project structure and organization
  - Enhanced documentation navigation

## Integration with SysMedic

The WebSocket server is fully integrated with the SysMedic daemon:

1. **Automatic Startup**: If WebSocket is enabled in configuration, it starts automatically with the daemon
2. **Alert Broadcasting**: System alerts are automatically broadcast to connected clients
3. **Lifecycle Management**: Server stops cleanly when daemon stops
4. **Resource Sharing**: Uses the same monitoring data as the daemon
5. **Configuration Persistence**: Settings persist across daemon restarts

## API Summary

**WebSocket Endpoints:**
- `ws://host:port/ws?secret=<secret>` - WebSocket connection
- `http://host:port/health` - Health check endpoint

**Real-time Messages:**
- `welcome` - Connection confirmation
- `system_update` - Periodic system metrics (every 5s)
- `alert` - System alert notifications

**Request Types:**
- `get_system_info` - Current system metrics
- `get_alerts` - Alert information
- `get_user_metrics` - User resource usage
- `get_config` - Server configuration
- `ping` - Connection test

**CLI Commands:**
- `sysmedic websocket start [port]` - Configure WebSocket server
- `sysmedic websocket stop` - Disable WebSocket server
- `sysmedic websocket status` - Show comprehensive status
- `sysmedic websocket new-secret` - Generate new authentication secret

This comprehensive WebSocket implementation enables powerful real-time remote monitoring capabilities while maintaining security and integration with the existing SysMedic architecture.