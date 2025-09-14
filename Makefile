# SysMedic Makefile

# Variables
BINARY_NAME=sysmedic
BUILD_DIR=build
DIST_DIR=dist
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.version=$(VERSION) -X main.commit=$(COMMIT) -X main.date=$(BUILD_DATE) -s -w"

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build targets
.PHONY: all build clean test deps lint install uninstall package help

all: clean deps build

build: deps
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@echo "Running pre-build checks..."
	@$(GOCMD) fmt ./...
	@$(GOCMD) vet ./...
	@echo "Compiling binary..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/sysmedic
	@echo "Build completed successfully"
	@ls -la $(BUILD_DIR)/$(BINARY_NAME)

build-static: deps
	@echo "Building static $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BUILD_DIR)/$(BINARY_NAME)-static ./cmd/sysmedic

build-linux-amd64: deps
	@echo "Building $(BINARY_NAME) for Linux AMD64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/sysmedic

build-linux-arm64: deps
	@echo "Building $(BINARY_NAME) for Linux ARM64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=arm64 CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/sysmedic

build-all: build-linux-amd64 build-linux-arm64

test: deps
	@echo "Running tests..."
	$(GOTEST) -v ./...

test-coverage: deps
	@echo "Running tests with coverage..."
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

lint:
	@echo "Running linter..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed. Install with: curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; exit 1)
	golangci-lint run

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

install: build
	@echo "Installing $(BINARY_NAME)..."
	@echo "Checking for existing SysMedic daemons..."
	@sudo systemctl stop sysmedic.doctor 2>/dev/null || true
	@sudo systemctl stop sysmedic.websocket 2>/dev/null || true
	@sudo pkill -f "sysmedic.*--doctor-daemon" 2>/dev/null || true
	@sudo pkill -f "sysmedic.*--websocket-daemon" 2>/dev/null || true
	@sleep 2
	@echo "Installing binary..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)
	@echo "Creating directories..."
	sudo mkdir -p /etc/sysmedic
	sudo mkdir -p /var/lib/sysmedic
	@echo "Installing systemd services..."
	sudo cp scripts/sysmedic.doctor.service /etc/systemd/system/ 2>/dev/null || echo "Note: sysmedic.doctor.service not found"
	sudo cp scripts/sysmedic.websocket.service /etc/systemd/system/ 2>/dev/null || echo "Note: sysmedic.websocket.service not found"
	sudo systemctl daemon-reload
	@echo "Installation complete. Enable services with:"
	@echo "  sudo systemctl enable sysmedic.doctor"
	@echo "  sudo systemctl enable sysmedic.websocket"

uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	sudo systemctl stop sysmedic.doctor 2>/dev/null || true
	sudo systemctl stop sysmedic.websocket 2>/dev/null || true
	sudo systemctl disable sysmedic.doctor 2>/dev/null || true
	sudo systemctl disable sysmedic.websocket 2>/dev/null || true
	sudo rm -f /etc/systemd/system/sysmedic.doctor.service
	sudo rm -f /etc/systemd/system/sysmedic.websocket.service
	sudo systemctl daemon-reload
	sudo rm -f /usr/local/bin/$(BINARY_NAME)
	@echo "Note: Configuration and data files in /etc/sysmedic and /var/lib/sysmedic are preserved"

package: clean build-all
	@echo "Creating packages..."
	@mkdir -p $(DIST_DIR)

	# Create tar.gz package
	@echo "Creating tar.gz package..."
	@mkdir -p $(DIST_DIR)/sysmedic-$(VERSION)
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/sysmedic-$(VERSION)/sysmedic
	cp README.md $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cp LICENSE $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cp -r scripts $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cd $(DIST_DIR) && tar -czf sysmedic-$(VERSION)-linux-amd64.tar.gz sysmedic-$(VERSION)
	rm -rf $(DIST_DIR)/sysmedic-$(VERSION)

	# Create ARM64 tar.gz package
	@mkdir -p $(DIST_DIR)/sysmedic-$(VERSION)
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(DIST_DIR)/sysmedic-$(VERSION)/sysmedic
	cp README.md $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cp LICENSE $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cp -r scripts $(DIST_DIR)/sysmedic-$(VERSION)/ 2>/dev/null || true
	cd $(DIST_DIR) && tar -czf sysmedic-$(VERSION)-linux-arm64.tar.gz sysmedic-$(VERSION)
	rm -rf $(DIST_DIR)/sysmedic-$(VERSION)

package-deb: build-linux-amd64
	@echo "Creating DEB package..."
	@mkdir -p $(DIST_DIR)/deb/DEBIAN
	@mkdir -p $(DIST_DIR)/deb/usr/local/bin
	@mkdir -p $(DIST_DIR)/deb/etc/systemd/system
	@mkdir -p $(DIST_DIR)/deb/etc/sysmedic
	@mkdir -p $(DIST_DIR)/deb/var/lib/sysmedic

	# Copy binary
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/deb/usr/local/bin/$(BINARY_NAME)

	# Copy systemd services
	cp scripts/sysmedic.doctor.service $(DIST_DIR)/deb/etc/systemd/system/ 2>/dev/null || true
	cp scripts/sysmedic.websocket.service $(DIST_DIR)/deb/etc/systemd/system/ 2>/dev/null || true

	# Create control file
	@echo "Package: sysmedic" > $(DIST_DIR)/deb/DEBIAN/control
	@echo "Version: $(VERSION)" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo "Section: admin" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo "Priority: optional" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo "Architecture: amd64" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo "Maintainer: SysMedic Team <team@sysmedic.dev>" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo "Description: Cross-platform Linux server monitoring CLI tool" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo " SysMedic is a comprehensive server monitoring tool that tracks system" >> $(DIST_DIR)/deb/DEBIAN/control
	@echo " and user resource usage spikes with daemon capabilities." >> $(DIST_DIR)/deb/DEBIAN/control

	# Create postinst script
	@echo "#!/bin/bash" > $(DIST_DIR)/deb/DEBIAN/postinst
	@echo "chmod +x /usr/local/bin/sysmedic" >> $(DIST_DIR)/deb/DEBIAN/postinst
	@echo "systemctl daemon-reload" >> $(DIST_DIR)/deb/DEBIAN/postinst
	@echo "echo 'SysMedic installed. Enable with: sudo systemctl enable sysmedic.doctor && sudo systemctl enable sysmedic.websocket'" >> $(DIST_DIR)/deb/DEBIAN/postinst
	chmod +x $(DIST_DIR)/deb/DEBIAN/postinst

	# Create prerm script
	@echo "#!/bin/bash" > $(DIST_DIR)/deb/DEBIAN/prerm
	@echo "systemctl stop sysmedic.doctor 2>/dev/null || true" >> $(DIST_DIR)/deb/DEBIAN/prerm
	@echo "systemctl stop sysmedic.websocket 2>/dev/null || true" >> $(DIST_DIR)/deb/DEBIAN/prerm
	@echo "systemctl disable sysmedic.doctor 2>/dev/null || true" >> $(DIST_DIR)/deb/DEBIAN/prerm
	@echo "systemctl disable sysmedic.websocket 2>/dev/null || true" >> $(DIST_DIR)/deb/DEBIAN/prerm
	chmod +x $(DIST_DIR)/deb/DEBIAN/prerm

	# Build package
	dpkg-deb --build $(DIST_DIR)/deb $(DIST_DIR)/sysmedic-$(VERSION)-amd64.deb
	rm -rf $(DIST_DIR)/deb

package-rpm: build-linux-amd64
	@echo "Creating RPM package..."
	@mkdir -p $(DIST_DIR)/rpm/{BUILD,RPMS,SOURCES,SPECS,SRPMS}
	@mkdir -p $(DIST_DIR)/rpm/BUILD/usr/local/bin
	@mkdir -p $(DIST_DIR)/rpm/BUILD/etc/systemd/system
	@mkdir -p $(DIST_DIR)/rpm/BUILD/etc/sysmedic
	@mkdir -p $(DIST_DIR)/rpm/BUILD/var/lib/sysmedic

	# Copy binary
	cp $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(DIST_DIR)/rpm/BUILD/usr/local/bin/$(BINARY_NAME)

	# Copy systemd services
	cp scripts/sysmedic.doctor.service $(DIST_DIR)/rpm/BUILD/etc/systemd/system/ 2>/dev/null || true
	cp scripts/sysmedic.websocket.service $(DIST_DIR)/rpm/BUILD/etc/systemd/system/ 2>/dev/null || true

	# Create spec file
	@echo "Name: sysmedic" > $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "Version: $(VERSION)" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "Release: 1" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "Summary: Cross-platform Linux server monitoring CLI tool" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "License: MIT" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "Group: Applications/System" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "BuildArch: x86_64" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%description" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "SysMedic is a comprehensive server monitoring tool that tracks system" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "and user resource usage spikes with daemon capabilities." >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%files" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "/usr/local/bin/sysmedic" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "/etc/systemd/system/sysmedic.doctor.service" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "/etc/systemd/system/sysmedic.websocket.service" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%dir /etc/sysmedic" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%dir /var/lib/sysmedic" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%post" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "chmod +x /usr/local/bin/sysmedic" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "systemctl daemon-reload" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "%preun" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "systemctl stop sysmedic.doctor 2>/dev/null || true" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "systemctl stop sysmedic.websocket 2>/dev/null || true" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "systemctl disable sysmedic.doctor 2>/dev/null || true" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	@echo "systemctl disable sysmedic.websocket 2>/dev/null || true" >> $(DIST_DIR)/rpm/SPECS/sysmedic.spec

	# Build RPM (requires rpmbuild)
	@which rpmbuild > /dev/null || (echo "rpmbuild not installed. Install with: sudo apt-get install rpm or sudo yum install rpm-build"; exit 1)
	rpmbuild --define "_topdir $(PWD)/$(DIST_DIR)/rpm" --define "_builddir $(PWD)/$(DIST_DIR)/rpm/BUILD" -bb $(DIST_DIR)/rpm/SPECS/sysmedic.spec
	cp $(DIST_DIR)/rpm/RPMS/x86_64/*.rpm $(DIST_DIR)/
	rm -rf $(DIST_DIR)/rpm

dev: build
	@echo "Starting development mode..."
	@echo "Building and running daemon in foreground..."
	sudo $(BUILD_DIR)/$(BINARY_NAME) daemon start

docker-build:
	@echo "Building Docker image..."
	docker build -t sysmedic:$(VERSION) .
	docker tag sysmedic:$(VERSION) sysmedic:latest

docker-run: docker-build
	@echo "Running SysMedic in Docker..."
	docker run --rm -it --privileged --pid=host --net=host -v /proc:/host/proc:ro -v /sys:/host/sys:ro sysmedic:latest

benchmark: deps
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

release: clean test package package-deb
	@echo "Release $(VERSION) created in $(DIST_DIR)/"
	@ls -la $(DIST_DIR)/

help:
	@echo "SysMedic Build System"
	@echo ""
	@echo "Available targets:"
	@echo "  build          - Build the binary"
	@echo "  build-static   - Build static binary"
	@echo "  build-all      - Build for all platforms"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage"
	@echo "  deps           - Download dependencies"
	@echo "  lint           - Run linter"
	@echo "  clean          - Clean build artifacts"
	@echo "  install        - Install to system"
	@echo "  uninstall      - Remove from system"
	@echo "  package        - Create tar.gz packages"
	@echo "  package-deb    - Create DEB package"
	@echo "  package-rpm    - Create RPM package"
	@echo "  dev            - Run in development mode"
	@echo "  docker-build   - Build Docker image"
	@echo "  docker-run     - Run in Docker"
	@echo "  benchmark      - Run benchmarks"
	@echo "  release        - Create full release"
	@echo "  help           - Show this help"
	@echo ""
	@echo "Variables:"
	@echo "  VERSION=$(VERSION)"
	@echo "  COMMIT=$(COMMIT)"
	@echo "  BUILD_DATE=$(BUILD_DATE)"
