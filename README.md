# Agentic RTB Framework - Go Reference Implementation

A Go implementation of the IAB Tech Lab's **Agentic RTB Framework (ARTF) v1.0** specification for agent-driven containers in OpenRTB and Digital Advertising.

## Overview

This project implements a gRPC server that conforms to the ARTF specification, enabling:

- **Segment Activation** - Activate user segments based on bid request data
- **Deal Management** - Activate, suppress, and adjust deals dynamically
- **Bid Shading** - Optimize bid prices using intelligent pricing strategies
- **Metrics Addition** - Add viewability and other metrics to impressions

## Quick Start

### Prerequisites

- Go 1.22+
- Protocol Buffers compiler (`protoc`)
- Docker (optional, for containerized deployment)

### Build and Run

```bash
# Install dependencies
make deps

# Generate protobuf code
make generate

# Build the server
make build

# Run locally
make run
```

### Docker Deployment

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Or use docker-compose
make docker-compose-up
```

## Architecture

```
.
├── cmd/server/          # Main server entry point
├── internal/
│   ├── handlers/        # Mutation handlers for different intents
│   ├── health/          # Kubernetes health check endpoints
│   └── server/          # gRPC server implementation
├── pkg/pb/              # Generated protobuf Go code
├── proto/               # Protocol buffer definitions
│   ├── agenticrtbframework.proto  # ARTF service definition
│   └── com/iabtechlab/openrtb/    # OpenRTB v2.6 definitions
├── scripts/             # Build and utility scripts
├── Dockerfile           # Container build definition
└── docker-compose.yml   # Local development setup
```

## API

### RTBExtensionPoint Service

```protobuf
service RTBExtensionPoint {
  rpc GetMutations (RTBRequest) returns (RTBResponse);
}
```

### Supported Intents

| Intent | Description |
|--------|-------------|
| `ACTIVATE_SEGMENTS` | Activate user segments by external segment IDs |
| `ACTIVATE_DEALS` | Activate deals by external deal IDs |
| `SUPPRESS_DEALS` | Suppress deals by external deal IDs |
| `ADJUST_DEAL_FLOOR` | Adjust the bid floor of a specific deal |
| `ADJUST_DEAL_MARGIN` | Adjust the deal margin |
| `BID_SHADE` | Adjust the bid price |
| `ADD_METRICS` | Add metrics to an impression |

## Endpoints

| Port | Protocol | Endpoint | Description |
|------|----------|----------|-------------|
| 50051 | gRPC | RTBExtensionPoint | Main service endpoint |
| 8080 | HTTP | /health/live | Liveness probe |
| 8080 | HTTP | /health/ready | Readiness probe |

## Testing

```bash
# Run unit tests
make test

# Run with coverage
make test-coverage

# Test gRPC endpoint (requires grpcurl)
make grpc-test

# Check health endpoints
make health-check
```

## Configuration

| Flag | Default | Description |
|------|---------|-------------|
| `-grpc-port` | 50051 | gRPC server port |
| `-health-port` | 8080 | Health check HTTP port |

## Security

This implementation follows ARTF security requirements:

- Runs as non-root user
- Drops unnecessary capabilities
- Read-only filesystem
- No external network access (configurable)
- Health probes for Kubernetes integration

## Specification

This implementation is based on the [IAB Tech Lab Agentic RTB Framework v1.0](https://iabtechlab.com/standards/artf/) specification.

## License

Apache 2.0

## Contributing

Contributions welcome! Please read the ARTF specification before submitting changes.
