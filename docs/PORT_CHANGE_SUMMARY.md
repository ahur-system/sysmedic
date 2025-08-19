# WebSocket Port Change Summary

## Overview

Successfully changed the default WebSocket port from **8080** to **8060** to avoid conflicts with other services running on port 8080.

## Changes Made

### 1. Core Configuration Files

#### `internal/websocket/manager.go`
- Changed default port from 8080 to 8060 in `LoadConfig()` function
- Updated fallback configuration

#### `cmd/sysmedic/main.go`
- Updated default port assignment in WebSocket start command
- Changed from `port := 8080` to `port := 8060`

#### `pkg/cli/cli.go`
- Updated fallback configuration in `StopWebSocketServer()` function
- Fixed variable naming conflicts to prevent compilation errors
- Improved configuration flow to allow setting up WebSocket even when daemon is not running

### 2. Documentation Updates

#### `README.md`
- Updated WebSocket section to show 8060 as default port
- Changed help text: "Enable WebSocket server (default port 8060)"

#### `WEBSOCKET_FEATURE.md`
- Updated start command documentation to reflect 8060 as default
- Changed example from 8080 to 8060

### 3. Example Client Updates

#### `examples/websocket_client.html`
- Updated example URL from port 8080 to 8060
- Changed example connection string

#### `examples/websocket_client.py`
- Updated example URLs in help text and documentation
- Changed default port references

## Testing Results

### Before Changes
```bash
$ ./sysmedic websocket start
✓ WebSocket server configuration saved
Port: 8080  # OLD PORT
```

### After Changes
```bash
$ ./sysmedic websocket start
✓ WebSocket server configuration saved
Port: 8060  # NEW PORT
```

### Daemon Integration Test
```bash
$ timeout 5s ./sysmedic daemon start
Starting SysMedic daemon...
SysMedic daemon started (PID: 149172)
Monitoring interval: 1m0s
2025/08/19 00:27:29 WebSocket server starting on port 8060  # ✅ CORRECT PORT
WebSocket server started on port 8060
WebSocket connection URL: sysmedic://[secret]@automation.cli.ahur.ir:8060/
```

### Configuration File
```json
{
  "port": 8060,
  "enabled": true
}
```

## Backward Compatibility

- **Custom Ports**: Users can still specify custom ports with `sysmedic websocket start [port]`
- **Existing Configurations**: Existing WebSocket configurations will preserve their configured ports
- **Migration**: No migration needed - only the default changes for new configurations

## Commands Updated

All WebSocket commands now use 8060 as the default:

```bash
# Uses port 8060 by default
sysmedic websocket start

# Custom port still works
sysmedic websocket start 9090

# Status shows correct port
sysmedic websocket status

# Configuration persists correctly
sysmedic websocket stop
```

## Port Conflict Resolution

### Problem
- Port 8080 was already in use by another service
- WebSocket server couldn't bind to port 8080
- Error: `listen tcp :8080: bind: address already in use`

### Solution
- Changed default port to 8060
- Port 8060 is typically available on most systems
- Maintains same functionality with different port

## Files Modified

1. `internal/websocket/manager.go` - Default port configuration
2. `cmd/sysmedic/main.go` - CLI default port
3. `pkg/cli/cli.go` - CLI functions and fallbacks
4. `README.md` - Documentation update
5. `WEBSOCKET_FEATURE.md` - Feature documentation
6. `examples/websocket_client.html` - HTML client example
7. `examples/websocket_client.py` - Python client example

## Verification

✅ **Default Port**: Confirmed 8060 is used when no port specified  
✅ **Custom Ports**: Confirmed custom ports still work correctly  
✅ **Configuration**: Confirmed settings persist correctly  
✅ **Daemon Integration**: Confirmed WebSocket starts with daemon  
✅ **Documentation**: All docs updated to reflect new default  
✅ **Examples**: Client examples updated with correct port  

## Status: ✅ COMPLETE

The WebSocket default port has been successfully changed from 8080 to 8060. All functionality remains the same, with improved port availability and conflict resolution.