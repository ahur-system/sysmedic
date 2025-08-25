/**
 * SysMedic WebSocket Client Example
 *
 * This example demonstrates how to create a robust WebSocket client
 * that connects to the SysMedic WebSocket server with proper:
 * - Connection management
 * - Automatic reconnection
 * - Ping/pong handling
 * - Error handling
 * - Connection state monitoring
 */

class SysMedicClient {
    constructor(url, secret, options = {}) {
        this.url = url;
        this.secret = secret;
        this.ws = null;
        this.isConnected = false;
        this.reconnectAttempts = 0;
        this.maxReconnectAttempts = options.maxReconnectAttempts || 10;
        this.reconnectInterval = options.reconnectInterval || 5000; // 5 seconds
        this.pingInterval = options.pingInterval || 30000; // 30 seconds
        this.pongTimeout = options.pongTimeout || 10000; // 10 seconds
        this.reconnectTimer = null;
        this.pingTimer = null;
        this.pongTimer = null;
        this.lastPongReceived = Date.now();

        // Event handlers
        this.onConnect = options.onConnect || (() => {});
        this.onDisconnect = options.onDisconnect || (() => {});
        this.onMessage = options.onMessage || (() => {});
        this.onError = options.onError || (() => {});
        this.onReconnecting = options.onReconnecting || (() => {});

        // Auto-connect if not disabled
        if (options.autoConnect !== false) {
            this.connect();
        }
    }

    connect() {
        try {
            console.log(`Connecting to ${this.url}...`);

            // Clear any existing connection
            this.cleanup();

            // Build WebSocket URL with secret
            const wsUrl = new URL(this.url);
            wsUrl.searchParams.set("secret", this.secret);

            this.ws = new WebSocket(wsUrl.toString());

            this.ws.onopen = this.handleOpen.bind(this);
            this.ws.onmessage = this.handleMessage.bind(this);
            this.ws.onclose = this.handleClose.bind(this);
            this.ws.onerror = this.handleError.bind(this);
        } catch (error) {
            console.error("Failed to create WebSocket connection:", error);
            this.scheduleReconnect();
        }
    }

    handleOpen(event) {
        console.log("WebSocket connection established");
        this.isConnected = true;
        this.reconnectAttempts = 0;
        this.lastPongReceived = Date.now();

        // Start ping/pong monitoring
        this.startPingMonitoring();

        // Clear reconnect timer
        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        this.onConnect(event);
    }

    handleMessage(event) {
        try {
            const data = JSON.parse(event.data);

            // Handle different message types
            switch (data.type) {
                case "welcome":
                    console.log("Received welcome message:", data.data.message);
                    console.log("Server version:", data.data.version);
                    console.log("System:", data.data.system);
                    console.log("Status:", data.data.status);
                    console.log("Daemon:", data.data.daemon);
                    break;

                case "config":
                    console.log("📋 Configuration received:");
                    console.log(data.data);
                    break;

                case "system_update":
                    console.log("System update received:", data.data);
                    console.log(
                        `CPU: ${data.data.cpu_usage}%, Memory: ${data.data.memory_usage}%, Disk: ${data.data.disk_usage}%, Uptime: ${data.data.uptime}`,
                    );
                    break;

                case "pong":
                    this.handlePong();
                    break;

                case "alert":
                    console.log("Alert received:", data.data);
                    break;

                default:
                    console.log("Unknown message type:", data.type);
            }

            this.onMessage(data);
        } catch (error) {
            console.error("Failed to parse message:", error);
        }
    }

    handleClose(event) {
        console.log(
            `WebSocket connection closed: ${event.code} ${event.reason}`,
        );
        this.isConnected = false;
        this.cleanup();

        this.onDisconnect(event);

        // Schedule reconnection unless it was a clean close
        if (event.code !== 1000) {
            this.scheduleReconnect();
        }
    }

    handleError(event) {
        console.error("WebSocket error:", event);
        this.onError(event);
    }

    startPingMonitoring() {
        // Send ping every 30 seconds
        this.pingTimer = setInterval(() => {
            if (this.isConnected) {
                this.sendPing();
            }
        }, this.pingInterval);

        // Check for pong timeout every 10 seconds
        this.pongTimer = setInterval(() => {
            const timeSinceLastPong = Date.now() - this.lastPongReceived;
            if (timeSinceLastPong > this.pongTimeout + this.pingInterval) {
                console.warn("Pong timeout detected, reconnecting...");
                this.reconnect();
            }
        }, this.pongTimeout);
    }

    sendPing() {
        if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(
                    JSON.stringify({
                        type: "ping",
                        request_id: `ping_${Date.now()}`,
                    }),
                );
                console.log("Ping sent");
            } catch (error) {
                console.error("Failed to send ping:", error);
            }
        }
    }

    handlePong() {
        this.lastPongReceived = Date.now();
        console.log("Pong received");
    }

    scheduleReconnect() {
        if (this.reconnectAttempts >= this.maxReconnectAttempts) {
            console.error("Max reconnection attempts reached");
            return;
        }

        this.reconnectAttempts++;
        const delay = Math.min(
            this.reconnectInterval * Math.pow(1.5, this.reconnectAttempts - 1),
            30000,
        );

        console.log(
            `Scheduling reconnection attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`,
        );

        this.onReconnecting(this.reconnectAttempts, delay);

        this.reconnectTimer = setTimeout(() => {
            this.connect();
        }, delay);
    }

    reconnect() {
        console.log("Manual reconnection triggered");
        this.cleanup();
        this.connect();
    }

    send(data) {
        if (this.isConnected && this.ws.readyState === WebSocket.OPEN) {
            try {
                this.ws.send(JSON.stringify(data));
                return true;
            } catch (error) {
                console.error("Failed to send message:", error);
                return false;
            }
        } else {
            console.warn("Cannot send message: not connected");
            return false;
        }
    }

    // Request specific data from server
    requestSystemInfo() {
        return this.send({
            type: "get_system_info",
            request_id: `sysinfo_${Date.now()}`,
        });
    }

    requestAlerts() {
        return this.send({
            type: "get_alerts",
            request_id: `alerts_${Date.now()}`,
        });
    }

    requestUserMetrics() {
        return this.send({
            type: "get_user_metrics",
            request_id: `usermetrics_${Date.now()}`,
        });
    }

    requestConfig() {
        return this.send({
            type: "get_config",
            request_id: `config_${Date.now()}`,
        });
    }

    cleanup() {
        if (this.pingTimer) {
            clearInterval(this.pingTimer);
            this.pingTimer = null;
        }

        if (this.pongTimer) {
            clearInterval(this.pongTimer);
            this.pongTimer = null;
        }

        if (this.ws) {
            this.ws.onopen = null;
            this.ws.onmessage = null;
            this.ws.onclose = null;
            this.ws.onerror = null;

            if (this.ws.readyState === WebSocket.OPEN) {
                this.ws.close(1000, "Client disconnecting");
            }
            this.ws = null;
        }
    }

    disconnect() {
        console.log("Disconnecting WebSocket client...");
        this.isConnected = false;

        if (this.reconnectTimer) {
            clearTimeout(this.reconnectTimer);
            this.reconnectTimer = null;
        }

        this.cleanup();
    }

    getConnectionState() {
        return {
            isConnected: this.isConnected,
            reconnectAttempts: this.reconnectAttempts,
            readyState: this.ws ? this.ws.readyState : WebSocket.CLOSED,
            lastPongReceived: this.lastPongReceived,
        };
    }
}

// Usage example
function createSysMedicClient() {
    const client = new SysMedicClient(
        "ws://45.95.186.208:8060/ws",
        "55625821f7a0a9db98707bae107e46a4",
        {
            maxReconnectAttempts: 10,
            reconnectInterval: 5000,
            pingInterval: 30000,
            pongTimeout: 10000,
            onConnect: () => {
                console.log("✅ Connected to SysMedic server");
                updateConnectionStatus("Connected");
            },
            onDisconnect: (event) => {
                console.log("❌ Disconnected from SysMedic server");
                updateConnectionStatus("Disconnected");
            },
            onMessage: (data) => {
                handleServerMessage(data);
            },
            onError: (error) => {
                console.error("❌ WebSocket error:", error);
                updateConnectionStatus("Error");
            },
            onReconnecting: (attempt, delay) => {
                console.log(`🔄 Reconnecting... (${attempt}/10) in ${delay}ms`);
                updateConnectionStatus(`Reconnecting (${attempt}/10)`);
            },
        },
    );

    return client;
}

function updateConnectionStatus(status) {
    // Update UI connection status
    const statusElement = document.getElementById("connection-status");
    if (statusElement) {
        statusElement.textContent = status;
        statusElement.className = `status ${status.toLowerCase().replace(/[^a-z]/g, "")}`;
    }
}

function handleServerMessage(data) {
    switch (data.type) {
        case "system_update":
            updateSystemMetrics(data.data);
            break;
        case "alert":
            displayAlert(data.data);
            break;
        case "welcome":
            displayWelcomeInfo(data.data);
            break;
        case "config":
            displayConfig(data.data);
            break;
    }
}

function updateSystemMetrics(metrics) {
    // Update UI with system metrics
    if (metrics.cpu_usage !== undefined) {
        updateElement("cpu-usage", `${metrics.cpu_usage.toFixed(1)}%`);
    }
    if (metrics.memory_usage !== undefined) {
        updateElement("memory-usage", `${metrics.memory_usage.toFixed(1)}%`);
    }
    if (metrics.disk_usage !== undefined) {
        updateElement("disk-usage", `${metrics.disk_usage.toFixed(1)}%`);
    }
    if (metrics.uptime) {
        updateElement("uptime", metrics.uptime);
    }
}

function displayAlert(alert) {
    console.log("🚨 Alert:", alert);
    // Add alert to UI
}

function displayWelcomeInfo(info) {
    console.log("Welcome info:", info);
    updateElement("server-version", info.version || "Unknown");
    updateElement("server-uptime", info.system || "Unknown");
    updateConnectionStatus(`Connected (${info.status || "Unknown"})`);
}

function displayConfig(configData) {
    console.log("📋 Configuration data:", configData);
    // Display config in a formatted way if there's a UI element for it
    const configElement = document.getElementById("config-display");
    if (configElement) {
        configElement.textContent = configData;
    }
}

function updateElement(id, value) {
    const element = document.getElementById(id);
    if (element) {
        element.textContent = value;
    }
}

// Initialize client when page loads
let sysMedicClient;

document.addEventListener("DOMContentLoaded", () => {
    sysMedicClient = createSysMedicClient();

    // Add manual control buttons
    document.getElementById("connect-btn")?.addEventListener("click", () => {
        sysMedicClient.connect();
    });

    document.getElementById("disconnect-btn")?.addEventListener("click", () => {
        sysMedicClient.disconnect();
    });

    document
        .getElementById("request-sysinfo-btn")
        ?.addEventListener("click", () => {
            sysMedicClient.requestSystemInfo();
        });

    document
        .getElementById("request-alerts-btn")
        ?.addEventListener("click", () => {
            sysMedicClient.requestAlerts();
        });
});

// Clean up on page unload
window.addEventListener("beforeunload", () => {
    if (sysMedicClient) {
        sysMedicClient.disconnect();
    }
});

// Export for use in other modules
if (typeof module !== "undefined" && module.exports) {
    module.exports = SysMedicClient;
}
