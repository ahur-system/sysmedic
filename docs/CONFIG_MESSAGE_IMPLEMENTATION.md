# WebSocket Configuration Message Implementation

## Overview

This document describes the implementation of automatic YAML configuration sending through the SysMedic WebSocket connection. After a client connects and receives the welcome message, the server now automatically sends the current working YAML configuration.

## Implementation Details

### Message Flow

1. **Client connects** to WebSocket server with authentication
2. **Welcome message** is sent with basic system information
3. **Config message** is sent immediately after welcome with full YAML configuration
4. **System updates** continue every 3 seconds as before

### Message Format

The config message follows this JSON structure:

```json
{
    "type": "config",
    "data": "yaml_configuration_string",
    "timestamp": "2025-08-20T17:37:23Z"
}
```

Where `data` contains the complete YAML configuration as a string.

### Code Changes

#### 1. WebSocket Daemon Handler (`cmd/sysmedic/main.go`)

**Added functionality:**
- New `sendConfigMessage()` method to WebSocketDaemon
- Modified `handleWebSocket()` to call config sending after welcome
- Added `gopkg.in/yaml.v2` import

**Key changes:**
```go
// In handleWebSocket function - added after welcome message
if err := ws.sendConfigMessage(conn); err != nil {
    log.Printf("Failed to send config message: %v", err)
    return
}

// New method added
func (ws *WebSocketDaemon) sendConfigMessage(conn *websocket.Conn) error {
    yamlData, err := yaml.Marshal(ws.config)
    if err != nil {
        return fmt.Errorf("failed to marshal config to YAML: %v", err)
    }

    configMessage := map[string]interface{}{
        "type":      "config",
        "data":      string(yamlData),
        "timestamp": time.Now().Format("2006-01-02T15:04:05Z"),
    }

    conn.SetWriteDeadline(time.Now().Add(writeWait))
    if err := conn.WriteJSON(configMessage); err != nil {
        return fmt.Errorf("failed to send config message: %v", err)
    }

    return nil
}
```

#### 2. Client Examples Updated

**Python Client (`examples/websocket_client.py`):**
- Enhanced config message handler with YAML parsing
- Displays key configuration sections (monitoring, websocket, user_thresholds)
- Shows formatted configuration values
- Graceful fallback if PyYAML unavailable

**JavaScript Client (`examples/websocket_client.js`):**
- Added config message case in message handler
- Added `displayConfig()` function for UI integration

## Configuration Data Sent

The complete configuration includes all sections:

- **monitoring**: System-wide thresholds and intervals
- **users**: Default user monitoring settings
- **reporting**: Data retention and reporting periods
- **email**: Email notification configuration
- **websocket**: WebSocket server settings
- **user_thresholds**: Per-user custom thresholds
- **user_filtering**: User filtering and inclusion rules
- **data_path**: Data storage location

## Testing

### Test Scripts Created

1. **`test_config_message.py`** - Comprehensive test verifying:
   - Welcome message received first
   - Config message received immediately after
   - Config data is valid YAML
   - Expected configuration sections present

2. **`simple_test.py`** - Simple connection test showing all messages
3. **`test_websocket_startup.sh`** - Automated end-to-end test script

### Test Results

All tests pass successfully:
- ✅ Welcome message received first
- ✅ Config message received after welcome
- ✅ Config data is valid YAML
- ✅ Config contains expected sections

## Usage Examples

### Python Client Handling
```python
elif msg_type == 'config':
    config_data = data.get('data', '')
    config_parsed = yaml.safe_load(config_data)
    
    # Access configuration values
    cpu_threshold = config_parsed['monitoring']['cpu_threshold']
    websocket_port = config_parsed['websocket']['port']
```

### JavaScript Client Handling
```javascript
case "config":
    console.log("📋 Configuration received:");
    console.log(data.data);
    displayConfig(data.data);
    break;
```

## Security Considerations

- Configuration is only sent to authenticated WebSocket connections
- Secret authentication required via URL parameter
- No sensitive data like passwords are included in the config transmission
- Configuration reflects the actual working configuration file

## Backward Compatibility

This implementation maintains full backward compatibility:
- Existing clients continue to work without modification
- New config message is simply additional data
- No changes to existing message types or timing
- Welcome and system update messages unchanged

## Benefits

1. **Real-time Configuration Access**: Clients can immediately see current server configuration
2. **Dynamic Configuration Display**: Web UIs can show live configuration values
3. **Troubleshooting**: Easier to verify configuration from remote connections
4. **Integration**: Clients can adapt behavior based on server configuration
5. **Transparency**: Complete visibility into server settings

## Future Enhancements

Potential improvements could include:
- Configuration change notifications
- Selective configuration sections
- Configuration validation feedback
- Remote configuration updates