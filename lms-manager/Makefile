# LMS Manager Makefile
# Author: jejakawan007
# Company: K2NET
# Website: https://k2net.id

# Variables
APP_NAME = lms-manager
VERSION = 1.0.0
BUILD_DIR = build
DIST_DIR = dist
GO_VERSION = 1.21
LDFLAGS = -ldflags "-X main.version=$(VERSION) -X main.buildTime=$(shell date -u '+%Y-%m-%d_%H:%M:%S')"

# Colors
RED = \033[0;31m
GREEN = \033[0;32m
YELLOW = \033[1;33m
BLUE = \033[0;34m
NC = \033[0m # No Color

# Default target
.PHONY: all
all: clean build

# Help target
.PHONY: help
help:
	@echo "$(BLUE)LMS Manager - Build System$(NC)"
	@echo ""
	@echo "Available targets:"
	@echo "  $(GREEN)build$(NC)        - Build the application"
	@echo "  $(GREEN)clean$(NC)        - Clean build artifacts"
	@echo "  $(GREEN)test$(NC)         - Run tests"
	@echo "  $(GREEN)lint$(NC)         - Run linter"
	@echo "  $(GREEN)format$(NC)       - Format code"
	@echo "  $(GREEN)install$(NC)      - Install dependencies"
	@echo "  $(GREEN)run$(NC)          - Run the application"
	@echo "  $(GREEN)docker$(NC)       - Build Docker image"
	@echo "  $(GREEN)package$(NC)      - Create distribution package"
	@echo "  $(GREEN)release$(NC)      - Create release package"
	@echo "  $(GREEN)install-script$(NC) - Install systemd service"
	@echo "  $(GREEN)uninstall$(NC)    - Uninstall systemd service"
	@echo ""

# Clean build artifacts
.PHONY: clean
clean:
	@echo "$(YELLOW)Cleaning build artifacts...$(NC)"
	@rm -rf $(BUILD_DIR)
	@rm -rf $(DIST_DIR)
	@rm -f $(APP_NAME)
	@rm -f $(APP_NAME).exe
	@rm -f *.log
	@rm -f *.db
	@echo "$(GREEN)Clean completed$(NC)"

# Install dependencies
.PHONY: install
install:
	@echo "$(YELLOW)Installing dependencies...$(NC)"
	@go mod download
	@go mod tidy
	@echo "$(GREEN)Dependencies installed$(NC)"

# Format code
.PHONY: format
format:
	@echo "$(YELLOW)Formatting code...$(NC)"
	@go fmt ./...
	@echo "$(GREEN)Code formatted$(NC)"

# Run linter
.PHONY: lint
lint:
	@echo "$(YELLOW)Running linter...$(NC)"
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "$(YELLOW)golangci-lint not found, installing...$(NC)"; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.54.2; \
		golangci-lint run; \
	fi
	@echo "$(GREEN)Linting completed$(NC)"

# Run tests
.PHONY: test
test:
	@echo "$(YELLOW)Running tests...$(NC)"
	@go test -v ./...
	@echo "$(GREEN)Tests completed$(NC)"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "$(YELLOW)Running tests with coverage...$(NC)"
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)Coverage report generated: coverage.html$(NC)"

# Build the application
.PHONY: build
build: install
	@echo "$(YELLOW)Building $(APP_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME) main.go
	@echo "$(GREEN)Build completed: $(BUILD_DIR)/$(APP_NAME)$(NC)"

# Build for multiple platforms
.PHONY: build-all
build-all: install
	@echo "$(YELLOW)Building for multiple platforms...$(NC)"
	@mkdir -p $(BUILD_DIR)
	@GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-amd64 main.go
	@GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-linux-arm64 main.go
	@GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-windows-amd64.exe main.go
	@GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-amd64 main.go
	@GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(APP_NAME)-darwin-arm64 main.go
	@echo "$(GREEN)Multi-platform build completed$(NC)"

# Run the application
.PHONY: run
run: build
	@echo "$(YELLOW)Running $(APP_NAME)...$(NC)"
	@cd $(BUILD_DIR) && ./$(APP_NAME)

# Run in development mode
.PHONY: dev
dev:
	@echo "$(YELLOW)Running in development mode...$(NC)"
	@go run main.go

# Build Docker image
.PHONY: docker
docker:
	@echo "$(YELLOW)Building Docker image...$(NC)"
	@docker build -t $(APP_NAME):$(VERSION) -f docker/Dockerfile .
	@docker tag $(APP_NAME):$(VERSION) $(APP_NAME):latest
	@echo "$(GREEN)Docker image built: $(APP_NAME):$(VERSION)$(NC)"

# Run Docker container
.PHONY: docker-run
docker-run: docker
	@echo "$(YELLOW)Running Docker container...$(NC)"
	@docker run -d -p 8080:8080 --name $(APP_NAME) $(APP_NAME):latest
	@echo "$(GREEN)Docker container started$(NC)"

# Stop Docker container
.PHONY: docker-stop
docker-stop:
	@echo "$(YELLOW)Stopping Docker container...$(NC)"
	@docker stop $(APP_NAME) || true
	@docker rm $(APP_NAME) || true
	@echo "$(GREEN)Docker container stopped$(NC)"

# Create distribution package
.PHONY: package
package: build-all
	@echo "$(YELLOW)Creating distribution package...$(NC)"
	@mkdir -p $(DIST_DIR)
	@cp -r static templates config scripts $(BUILD_DIR)/
	@cp README.md LICENSE $(BUILD_DIR)/ 2>/dev/null || true
	@cd $(BUILD_DIR) && tar -czf ../$(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz *
	@echo "$(GREEN)Distribution package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz$(NC)"

# Create release package
.PHONY: release
release: clean test lint build-all package
	@echo "$(YELLOW)Creating release package...$(NC)"
	@mkdir -p $(DIST_DIR)/release
	@cp $(DIST_DIR)/$(APP_NAME)-$(VERSION).tar.gz $(DIST_DIR)/release/
	@cp scripts/install.sh $(DIST_DIR)/release/
	@cp scripts/uninstall.sh $(DIST_DIR)/release/
	@cp scripts/update.sh $(DIST_DIR)/release/
	@chmod +x $(DIST_DIR)/release/*.sh
	@cd $(DIST_DIR)/release && tar -czf ../$(APP_NAME)-$(VERSION)-release.tar.gz *
	@echo "$(GREEN)Release package created: $(DIST_DIR)/$(APP_NAME)-$(VERSION)-release.tar.gz$(NC)"

# Install systemd service
.PHONY: install-service
install-service: build
	@echo "$(YELLOW)Installing systemd service...$(NC)"
	@sudo cp $(BUILD_DIR)/$(APP_NAME) /usr/local/bin/
	@sudo chmod +x /usr/local/bin/$(APP_NAME)
	@sudo cp scripts/install.sh /tmp/lms-manager-install.sh
	@sudo chmod +x /tmp/lms-manager-install.sh
	@echo "$(GREEN)Service installation script ready: /tmp/lms-manager-install.sh$(NC)"

# Uninstall systemd service
.PHONY: uninstall-service
uninstall-service:
	@echo "$(YELLOW)Uninstalling systemd service...$(NC)"
	@sudo systemctl stop $(APP_NAME) || true
	@sudo systemctl disable $(APP_NAME) || true
	@sudo rm -f /etc/systemd/system/$(APP_NAME).service
	@sudo rm -f /usr/local/bin/$(APP_NAME)
	@sudo systemctl daemon-reload
	@echo "$(GREEN)Service uninstalled$(NC)"

# Generate documentation
.PHONY: docs
docs:
	@echo "$(YELLOW)Generating documentation...$(NC)"
	@if command -v godoc >/dev/null 2>&1; then \
		godoc -http=:6060 & \
		echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"; \
	else \
		echo "$(YELLOW)godoc not found, installing...$(NC)"; \
		go install golang.org/x/tools/cmd/godoc@latest; \
		godoc -http=:6060 & \
		echo "$(GREEN)Documentation server started at http://localhost:6060$(NC)"; \
	fi

# Security scan
.PHONY: security
security:
	@echo "$(YELLOW)Running security scan...$(NC)"
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "$(YELLOW)gosec not found, installing...$(NC)"; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi
	@echo "$(GREEN)Security scan completed$(NC)"

# Benchmark tests
.PHONY: benchmark
benchmark:
	@echo "$(YELLOW)Running benchmark tests...$(NC)"
	@go test -bench=. -benchmem ./...
	@echo "$(GREEN)Benchmark tests completed$(NC)"

# Check for updates
.PHONY: check-updates
check-updates:
	@echo "$(YELLOW)Checking for updates...$(NC)"
	@go list -u -m all
	@echo "$(GREEN)Update check completed$(NC)"

# Update dependencies
.PHONY: update-deps
update-deps:
	@echo "$(YELLOW)Updating dependencies...$(NC)"
	@go get -u ./...
	@go mod tidy
	@echo "$(GREEN)Dependencies updated$(NC)"

# Show version information
.PHONY: version
version:
	@echo "$(BLUE)LMS Manager$(NC)"
	@echo "Version: $(VERSION)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Build Time: $(shell date -u '+%Y-%m-%d %H:%M:%S UTC')"
	@echo "Git Commit: $(shell git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Git Branch: $(shell git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')"

# Show build information
.PHONY: info
info:
	@echo "$(BLUE)Build Information$(NC)"
	@echo "App Name: $(APP_NAME)"
	@echo "Version: $(VERSION)"
	@echo "Build Dir: $(BUILD_DIR)"
	@echo "Dist Dir: $(DIST_DIR)"
	@echo "Go Version: $(GO_VERSION)"
	@echo "LDFLAGS: $(LDFLAGS)"

# Development setup
.PHONY: dev-setup
dev-setup: install
	@echo "$(YELLOW)Setting up development environment...$(NC)"
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install golang.org/x/tools/cmd/godoc@latest
	@go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	@echo "$(GREEN)Development environment setup completed$(NC)"

# CI/CD pipeline
.PHONY: ci
ci: clean install format lint test security build
	@echo "$(GREEN)CI/CD pipeline completed successfully$(NC)"

# Production build
.PHONY: prod
prod: clean install test lint security build-all package
	@echo "$(GREEN)Production build completed$(NC)"

# Quick development cycle
.PHONY: quick
quick: format lint test build run
	@echo "$(GREEN)Quick development cycle completed$(NC)"
