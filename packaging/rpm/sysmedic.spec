Name:           sysmedic
Version:        1.0.2
Release:        1%{?dist}
Summary:        Advanced Linux system monitoring and diagnostic tool
License:        MIT
URL:            https://github.com/ahur-system/sysmedic
Source0:        %{name}-%{version}.tar.gz
BuildArch:      x86_64

Requires:       systemd
Requires(post): systemd
Requires(preun): systemd
Requires(postun): systemd

%description
SysMedic is a comprehensive system monitoring tool that provides real-time
insights into system performance, resource usage, and health metrics.

Features:
- Real-time CPU, memory, disk, and network monitoring
- System health diagnostics and alerts
- Historical data collection and analysis
- Web-based dashboard interface
- RESTful API for integration
- Configurable monitoring intervals and thresholds
- Systemd service integration

SysMedic helps system administrators proactively monitor their Linux systems
and quickly identify performance bottlenecks or potential issues.

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
mkdir -p $RPM_BUILD_ROOT/var/log/sysmedic

# Install binary
cp sysmedic $RPM_BUILD_ROOT/usr/local/bin/

# Install systemd service
cat > $RPM_BUILD_ROOT/etc/systemd/system/sysmedic.service << 'EOF'
[Unit]
Description=SysMedic System Monitor
Documentation=https://github.com/ahur-system/sysmedic
After=network.target
Wants=network-online.target

[Service]
Type=simple
User=sysmedic
Group=sysmedic
ExecStart=/usr/local/bin/sysmedic --config=/etc/sysmedic/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=sysmedic
KillMode=mixed
KillSignal=SIGTERM
TimeoutStopSec=30

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/sysmedic /var/log/sysmedic
CapabilityBoundingSet=CAP_NET_BIND_SERVICE

[Install]
WantedBy=multi-user.target
EOF

# Install default configuration
cat > $RPM_BUILD_ROOT/etc/sysmedic/config.yaml << 'EOF'
# SysMedic Configuration
server:
  host: "127.0.0.1"
  port: 8080
  tls:
    enabled: false

monitoring:
  interval: 5s
  metrics:
    cpu: true
    memory: true
    disk: true
    network: true
    processes: true

storage:
  type: "file"
  path: "/var/lib/sysmedic/data"
  retention: "30d"

logging:
  level: "info"
  file: "/var/log/sysmedic/sysmedic.log"
  max_size: "100MB"
  max_files: 5

alerts:
  enabled: true
  thresholds:
    cpu_usage: 80
    memory_usage: 85
    disk_usage: 90
EOF

%clean
rm -rf $RPM_BUILD_ROOT

%pre
# Create sysmedic user and group if they don't exist
getent group sysmedic >/dev/null || groupadd -r sysmedic
getent passwd sysmedic >/dev/null || \
    useradd -r -g sysmedic -d /var/lib/sysmedic -s /sbin/nologin \
    -c "SysMedic System Monitor" sysmedic

%post
# Set proper ownership and permissions
chown -R sysmedic:sysmedic /var/lib/sysmedic
chown -R sysmedic:sysmedic /var/log/sysmedic
chown sysmedic:sysmedic /etc/sysmedic/config.yaml
chmod 755 /etc/sysmedic
chmod 750 /var/lib/sysmedic
chmod 750 /var/log/sysmedic
chmod 640 /etc/sysmedic/config.yaml

# Reload systemd and enable service
%systemd_post sysmedic.service

echo "SysMedic has been installed successfully!"
echo "Configuration file: /etc/sysmedic/config.yaml"
echo "Start the service with: sudo systemctl start sysmedic"
echo "View logs with: sudo journalctl -u sysmedic -f"

%preun
%systemd_preun sysmedic.service

%postun
%systemd_postun_with_restart sysmedic.service

if [ $1 -eq 0 ]; then
    # Complete removal
    rm -rf /etc/sysmedic
    rm -rf /var/lib/sysmedic
    rm -rf /var/log/sysmedic

    # Remove user and group
    userdel sysmedic 2>/dev/null || true
    groupdel sysmedic 2>/dev/null || true

    echo "SysMedic has been completely removed from the system."
fi

%files
%defattr(-,root,root,-)
%attr(755,root,root) /usr/local/bin/sysmedic
%attr(644,root,root) /etc/systemd/system/sysmedic.service
%config(noreplace) %attr(640,sysmedic,sysmedic) /etc/sysmedic/config.yaml
%dir %attr(755,root,root) /etc/sysmedic
%dir %attr(750,sysmedic,sysmedic) /var/lib/sysmedic
%dir %attr(750,sysmedic,sysmedic) /var/log/sysmedic

%changelog
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
