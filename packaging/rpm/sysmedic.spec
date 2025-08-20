Name:           sysmedic
Version:        1.0.5
Release:        1%{?dist}
Summary:        Single binary multi-daemon Linux system monitoring tool
License:        MIT
URL:            https://github.com/ahur-system/sysmedic
Source0:        %{name}-%{version}.tar.gz
BuildArch:      x86_64

Requires:       systemd
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
SysMedic is a comprehensive system monitoring tool that uses a single binary
with multiple daemon modes for modular system monitoring and remote access.

Features:
- Single binary with independent doctor and WebSocket daemon processes
- Real-time CPU, memory, disk, and network monitoring
- Smart user filtering focusing on real problematic users
- WebSocket server for remote monitoring access
- Historical data collection and trend analysis
- Configurable monitoring intervals and thresholds
- Independent daemon process management
- Dual systemd service integration

The new architecture provides process separation while maintaining deployment
simplicity with a single 11MB binary that can operate in multiple modes.

%prep
%setup -q

%build
# Binary is pre-built

%install
rm -rf $RPM_BUILD_ROOT

# Create directories
mkdir -p $RPM_BUILD_ROOT/usr/local/bin
mkdir -p $RPM_BUILD_ROOT/etc/systemd/system
mkdir -p $RPM_BUILD_ROOT/etc/sysmedic
mkdir -p $RPM_BUILD_ROOT/var/lib/sysmedic


# Install binary
cp sysmedic $RPM_BUILD_ROOT/usr/local/bin/

# Install systemd services
cat > $RPM_BUILD_ROOT/etc/systemd/system/sysmedic.doctor.service << 'EOF'
[Unit]
Description=SysMedic Doctor - System Monitoring Daemon
Documentation=https://github.com/ahur-system/sysmedic
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/sysmedic --doctor-daemon
ExecReload=/bin/kill -HUP $MAINPID
PIDFile=/var/lib/sysmedic/sysmedic.doctor.pid
Restart=on-failure
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sysmedic /etc/sysmedic /var/run
PrivateTmp=true
PrivateDevices=false
ProtectHostname=true
ProtectClock=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectKernelLogs=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true
RestrictNamespaces=true

# Allow access to system monitoring files
ReadOnlyPaths=/proc /sys

# Environment
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=SYSMEDIC_CONFIG=/etc/sysmedic/config.yaml
Environment=SYSMEDIC_DATA=/var/lib/sysmedic

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sysmedic.doctor

[Install]
WantedBy=multi-user.target
EOF

cat > $RPM_BUILD_ROOT/etc/systemd/system/sysmedic.websocket.service << 'EOF'
[Unit]
Description=SysMedic WebSocket - Remote Monitoring Server
Documentation=https://github.com/ahur-system/sysmedic
After=network.target
Wants=network.target

[Service]
Type=simple
User=root
Group=root
ExecStart=/usr/local/bin/sysmedic --websocket-daemon
ExecReload=/bin/kill -HUP $MAINPID
PIDFile=/var/lib/sysmedic/sysmedic.websocket.pid
Restart=on-failure
RestartSec=5
StartLimitInterval=60
StartLimitBurst=3

# Security settings
NoNewPrivileges=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/sysmedic /etc/sysmedic /var/run
PrivateTmp=true
PrivateDevices=true
ProtectHostname=true
ProtectClock=true
ProtectKernelTunables=true
ProtectKernelModules=true
ProtectKernelLogs=true
ProtectControlGroups=true
RestrictRealtime=true
RestrictSUIDSGID=true
RemoveIPC=true
RestrictNamespaces=true

# Allow access to system monitoring files for data collection
ReadOnlyPaths=/proc /sys

# Environment
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=SYSMEDIC_CONFIG=/etc/sysmedic/config.yaml
Environment=SYSMEDIC_DATA=/var/lib/sysmedic

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sysmedic.websocket

[Install]
WantedBy=multi-user.target
EOF

# Install default configuration
cat > $RPM_BUILD_ROOT/etc/sysmedic/config.yaml << 'EOF'
monitoring:
  check_interval: 60
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

users:
  cpu_threshold: 80
  memory_threshold: 80
  persistent_time: 60

reporting:
  period: "hourly"
  retain_days: 30

email:
  enabled: false
  smtp_host: "smtp.gmail.com"
  smtp_port: 587
  tls: true

websocket:
  enabled: true
  port: 8060
  secret: ""

user_filtering:
  min_uid_for_real_users: 1000
  ignore_system_users: true
  min_cpu_percent: 5.0
  min_memory_percent: 5.0
  min_process_count: 1
  excluded_users: ["root", "daemon", "bin", "sys", "sync", "games", "man", "lp", "mail", "news", "uucp", "proxy", "www-data", "backup", "list", "irc", "gnats", "nobody", "systemd+", "syslog", "_apt"]
  included_users: []

data_path: "/var/lib/sysmedic"

user_thresholds: {}
EOF

%clean
rm -rf $RPM_BUILD_ROOT

%pre

%post
# Set proper permissions
chmod 755 /etc/sysmedic
chmod 755 /var/lib/sysmedic
chmod 644 /etc/sysmedic/config.yaml

# Reload systemd and enable services
%systemd_post sysmedic.doctor.service
%systemd_post sysmedic.websocket.service

echo "SysMedic has been installed successfully!"
echo "Configuration file: /etc/sysmedic/config.yaml"
echo "Enable services with:"
echo "  sudo systemctl enable sysmedic.doctor"
echo "  sudo systemctl enable sysmedic.websocket"
echo "Start services with:"
echo "  sudo systemctl start sysmedic.doctor"
echo "  sudo systemctl start sysmedic.websocket"
echo "Or use the CLI:"
echo "  sysmedic daemon start"
echo "  sysmedic websocket start"
echo "View logs with:"
echo "  sudo journalctl -u sysmedic.doctor -f"
echo "  sudo journalctl -u sysmedic.websocket -f"

%preun
%systemd_preun sysmedic.doctor.service
%systemd_preun sysmedic.websocket.service

%postun
%systemd_postun_with_restart sysmedic.doctor.service
%systemd_postun_with_restart sysmedic.websocket.service

if [ $1 -eq 0 ]; then
    # Complete removal
    rm -rf /etc/sysmedic
    rm -rf /var/lib/sysmedic

    echo "SysMedic has been completely removed from the system."
fi

%files
%defattr(-,root,root,-)
%attr(755,root,root) /usr/local/bin/sysmedic
%attr(644,root,root) /etc/systemd/system/sysmedic.doctor.service
%attr(644,root,root) /etc/systemd/system/sysmedic.websocket.service
%config(noreplace) %attr(644,root,root) /etc/sysmedic/config.yaml
%dir %attr(755,root,root) /etc/sysmedic
%dir %attr(755,root,root) /var/lib/sysmedic

%changelog
* Wed Dec 11 2024 SysMedic Team <support@sysmedic.io> - 1.0.5-1
- Major architectural change: Single binary multi-daemon design
- Single 11MB binary with independent doctor and WebSocket daemon processes
- Complete process separation while maintaining deployment simplicity
- Independent daemon lifecycle management
- Enhanced systemd integration with dual services
- Better resource isolation and fault tolerance

* Wed Dec 11 2024 SysMedic Team <support@sysmedic.io> - 1.0.2-1
- Major improvement: Enhanced user filtering and monitoring
- Real usernames instead of uid_[id] format
- Smart filtering excludes system users automatically
- Only tracks users with significant resource usage
- Configurable user filtering options
- Better focus on actual problematic users

* Wed Dec 11 2024 SysMedic Team <support@sysmedic.io> - 1.0.1-1
- Critical fix: Replace go-sqlite3 with pure Go SQLite driver
- Resolves CGO dependency issues that caused runtime failures
- Ensures compatibility across all Linux distributions
- Fixes "Binary was compiled with CGO_ENABLED=0" error

* Wed Dec 11 2024 SysMedic Team <support@sysmedic.io> - 1.0.0-1
- Initial release
- Real-time system monitoring capabilities
- Web dashboard and REST API
- Systemd service integration
- Configurable alerts and thresholds
