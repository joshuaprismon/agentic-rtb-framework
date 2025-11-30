.PHONY: all build test clean generate docker run lint help proto-deps fetch-openrtb

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=artf-agent
DOCKER_IMAGE=artf-agent

# Build information
VERSION?=0.10.0
BUILD_TIME=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X main.BuildTime=$(BUILD_TIME) -w -s"

# Agent manifest configuration (customizable via environment or make args)
AGENT_NAME?=artf-reference-agent
AGENT_VENDOR?=IAB Tech Lab
AGENT_OWNER?=artf@iabtechlab.com

# Default ports
GRPC_PORT?=50051
MCP_PORT?=50052
WEB_PORT?=8081
HEALTH_PORT?=8080

# OpenRTB Proto Configuration
OPENRTB_REPO=https://raw.githubusercontent.com/IABTechLab/openrtb-proto-v2/master
OPENRTB_PROTO_SRC=$(OPENRTB_REPO)/openrtb-core/src/main/protobuf/openrtb.proto
OPENRTB_PROTO_DIR=proto/com/iabtechlab/openrtb/v2
OPENRTB_PROTO_FILE=$(OPENRTB_PROTO_DIR)/openrtb.proto

all: generate build test

## help: Show this help message
help:
	@echo "ARTF Agent - Makefile commands:"
	@echo ""
	@echo "Build Commands:"
	@echo "  make build              Build the agent binary"
	@echo "  make build-all          Build with all features enabled"
	@echo "  make generate           Generate protobuf code"
	@echo "  make fetch-openrtb      Download OpenRTB 2.6 proto from IAB Tech Lab"
	@echo "  make clean              Remove build artifacts"
	@echo ""
	@echo "Run Commands:"
	@echo "  make run                Run with gRPC only (default)"
	@echo "  make run-grpc           Run with gRPC enabled"
	@echo "  make run-mcp            Run with MCP enabled"
	@echo "  make run-web            Run with Web UI enabled"
	@echo "  make run-all            Run with all services enabled"
	@echo "  make run-dev            Run in development mode (all services)"
	@echo ""
	@echo "Test Commands:"
	@echo "  make test               Run unit tests"
	@echo "  make test-coverage      Run tests with coverage report"
	@echo "  make grpc-test          Test gRPC endpoint"
	@echo "  make mcp-test           Test MCP endpoint"
	@echo "  make health-check       Check health endpoints"
	@echo ""
	@echo "Docker Commands:"
	@echo "  make docker-build         Build Docker image with agent-manifest"
	@echo "  make docker-run           Run Docker container"
	@echo "  make docker-run-all       Run Docker with all services"
	@echo "  make docker-inspect-manifest  Show agent-manifest label"
	@echo "  make docker-compose-up    Start with docker-compose"
	@echo ""
	@echo "  Custom agent manifest (example):"
	@echo "    make docker-build AGENT_NAME=my-agent AGENT_VENDOR=\"My Corp\" VERSION=1.0.0"
	@echo ""
	@echo "Sample Commands:"
	@echo "  make sample-banner      Send sample banner request via MCP"
	@echo "  make sample-video       Send sample video request via MCP"
	@echo "  make sample-bidshade    Send sample bid shading request via MCP"
	@echo ""

## fetch-openrtb: Download OpenRTB 2.6 proto from IAB Tech Lab repository
fetch-openrtb:
	@echo "Fetching OpenRTB 2.6 protobuf definition from IAB Tech Lab..."
	@mkdir -p $(OPENRTB_PROTO_DIR)
	@curl -sSL $(OPENRTB_PROTO_SRC) -o $(OPENRTB_PROTO_FILE).tmp
	@echo "Adding go_package option for ARTF compatibility..."
	@sed '/^package /a\option go_package = "github.com/iabtechlab/agentic-rtb-framework/pkg/pb/openrtb";' $(OPENRTB_PROTO_FILE).tmp > $(OPENRTB_PROTO_FILE)
	@rm -f $(OPENRTB_PROTO_FILE).tmp
	@echo "OpenRTB 2.6 proto downloaded to $(OPENRTB_PROTO_FILE)"

## generate: Generate Go code from protobuf definitions
generate: fetch-openrtb
	@echo "Generating protobuf code..."
	@./scripts/generate.sh

## build: Build the agent binary
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) ./cmd/agent

## build-all: Build with embedded static files
build-all: generate build

## test: Run unit tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -cover ./...

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## lint: Run linter
lint:
	@echo "Running linter..."
	golangci-lint run ./...

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf pkg/pb/artf/*.pb.go
	rm -rf pkg/pb/openrtb/*.pb.go

#
# Run Commands
#

## run: Run the agent with default settings (gRPC only)
run: build
	@echo "Starting agent (gRPC only)..."
	./$(BINARY_NAME) --enable-grpc --grpc-port=$(GRPC_PORT) --health-port=$(HEALTH_PORT)

## run-grpc: Run with gRPC enabled
run-grpc: build
	@echo "Starting agent with gRPC..."
	./$(BINARY_NAME) --enable-grpc --grpc-port=$(GRPC_PORT) --health-port=$(HEALTH_PORT)

## run-mcp: Run with MCP enabled
run-mcp: build
	@echo "Starting agent with MCP..."
	./$(BINARY_NAME) --enable-grpc=false --enable-mcp --mcp-port=$(MCP_PORT) --health-port=$(HEALTH_PORT)

## run-web: Run with Web UI enabled (requires MCP)
run-web: build
	@echo "Starting agent with Web UI and MCP..."
	./$(BINARY_NAME) --enable-grpc=false --enable-mcp --enable-web --mcp-port=$(MCP_PORT) --web-port=$(WEB_PORT) --health-port=$(HEALTH_PORT)

## run-all: Run with all services enabled
run-all: build
	@echo "Starting agent with all services..."
	./$(BINARY_NAME) --enable-grpc --enable-mcp --enable-web \
		--grpc-port=$(GRPC_PORT) --mcp-port=$(MCP_PORT) --web-port=$(WEB_PORT) --health-port=$(HEALTH_PORT)

## run-dev: Run in development mode (all services, verbose)
run-dev: build
	@echo "Starting agent in development mode..."
	@echo "Services:"
	@echo "  gRPC:   localhost:$(GRPC_PORT)"
	@echo "  MCP:    http://localhost:$(MCP_PORT)/mcp"
	@echo "  Web UI: http://localhost:$(WEB_PORT)/"
	@echo "  Health: http://localhost:$(HEALTH_PORT)/health/ready"
	@echo ""
	./$(BINARY_NAME) --enable-grpc --enable-mcp --enable-web \
		--grpc-port=$(GRPC_PORT) --mcp-port=$(MCP_PORT) --web-port=$(WEB_PORT) --health-port=$(HEALTH_PORT)

#
# Docker Commands
#

## docker-build: Build Docker image with agent-manifest label
docker-build:
	@echo "Building Docker image..."
	@echo "  Agent: $(AGENT_NAME) v$(VERSION)"
	@echo "  Vendor: $(AGENT_VENDOR)"
	docker build \
		--build-arg VERSION=$(VERSION) \
		--build-arg AGENT_NAME="$(AGENT_NAME)" \
		--build-arg AGENT_VENDOR="$(AGENT_VENDOR)" \
		--build-arg AGENT_OWNER="$(AGENT_OWNER)" \
		-t $(DOCKER_IMAGE):$(VERSION) \
		-t $(DOCKER_IMAGE):latest .

## docker-run: Run Docker container with gRPC
docker-run:
	@echo "Running Docker container (gRPC only)..."
	docker run -p $(GRPC_PORT):50051 -p $(HEALTH_PORT):8080 \
		$(DOCKER_IMAGE):latest --enable-grpc

## docker-run-all: Run Docker container with all services
docker-run-all:
	@echo "Running Docker container (all services)..."
	docker run -p $(GRPC_PORT):50051 -p $(MCP_PORT):50052 -p $(WEB_PORT):8081 -p $(HEALTH_PORT):8080 \
		$(DOCKER_IMAGE):latest --enable-grpc --enable-mcp --enable-web

## docker-compose-up: Start with docker-compose
docker-compose-up:
	docker-compose up --build

## docker-compose-down: Stop docker-compose services
docker-compose-down:
	docker-compose down

## docker-inspect-manifest: Show the agent-manifest label from the Docker image
docker-inspect-manifest:
	@echo "Agent Manifest for $(DOCKER_IMAGE):$(VERSION):"
	@docker inspect $(DOCKER_IMAGE):$(VERSION) --format '{{index .Config.Labels "agent-manifest"}}' | python3 -m json.tool 2>/dev/null || \
		docker inspect $(DOCKER_IMAGE):$(VERSION) --format '{{index .Config.Labels "agent-manifest"}}'

#
# Dependency Commands
#

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## proto-deps: Install protobuf tooling
proto-deps:
	@echo "Installing protobuf tools..."
	$(GOGET) google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GOGET) google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

#
# Test Commands
#

## grpc-test: Test gRPC endpoint with grpcurl
grpc-test:
	@echo "Testing gRPC endpoint..."
	grpcurl -plaintext localhost:$(GRPC_PORT) list
	grpcurl -plaintext localhost:$(GRPC_PORT) describe com.iabtechlab.bidstream.mutation.v1.RTBExtensionPoint

## mcp-test: Test MCP endpoint initialization
mcp-test:
	@echo "Testing MCP endpoint..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0"}}}' | jq .

## mcp-tools: List available MCP tools
mcp-tools:
	@echo "Listing MCP tools..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":2,"method":"tools/list","params":{}}' | jq .

## health-check: Check health endpoints
health-check:
	@echo "Checking health endpoints..."
	@echo "=== Liveness ==="
	@curl -s http://localhost:$(HEALTH_PORT)/health/live | jq .
	@echo "=== Readiness ==="
	@curl -s http://localhost:$(HEALTH_PORT)/health/ready | jq .
	@echo "=== Info ==="
	@curl -s http://localhost:$(HEALTH_PORT)/health/info | jq .

#
# Sample Request Commands
#

## sample-banner: Send sample banner request via MCP
sample-banner:
	@echo "Sending sample banner request..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extend_rtb","arguments":'"$$(cat samples/banner-basic.json)"'}}' | jq .

## sample-video: Send sample video request via MCP
sample-video:
	@echo "Sending sample video request..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extend_rtb","arguments":'"$$(cat samples/video-deals.json)"'}}' | jq .

## sample-bidshade: Send sample bid shading request via MCP
sample-bidshade:
	@echo "Sending sample bid shading request..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extend_rtb","arguments":'"$$(cat samples/bid-shading.json)"'}}' | jq .

## sample-native: Send sample native ad request via MCP
sample-native:
	@echo "Sending sample native ad request..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extend_rtb","arguments":'"$$(cat samples/native-ad.json)"'}}' | jq .

## sample-multi: Send sample multi-impression request via MCP
sample-multi:
	@echo "Sending sample multi-impression request..."
	@curl -s -X POST http://localhost:$(MCP_PORT)/mcp \
		-H "Content-Type: application/json" \
		-d '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"extend_rtb","arguments":'"$$(cat samples/multi-impression.json)"'}}' | jq .

#
# Version Info
#

## version: Show version information
version:
	@echo "ARTF Agent v$(VERSION)"
	@echo "Build time: $(BUILD_TIME)"
