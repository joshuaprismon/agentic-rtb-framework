# CLAUDE.md - Agentic RTB Framework

This repository implements an agent service based on the IAB Tech Lab's **Agentic RTB Framework (ARTF) v1.0** specification. The framework enables agent-driven containers to participate in OpenRTB bidstream processing.

## Overview

The Agentic RTB Framework defines a standard for implementing agent services that:
- Operate within a host platform's infrastructure
- Process bidstream data in real-time (sub-millisecond latency requirements)
- Propose mutations to OpenRTB bid requests/responses via the "OpenRTB Patch" protocol
- Communicate via gRPC with protobuf serialization

## Project Structure

```
.
├── CLAUDE.md                           # This file
├── README.md                           # Project readme
├── agenticrtbframework.proto           # gRPC service definition
└── Agentic RTB Framework Version 1.0   # IAB Tech Lab specification (PDF)
    for PUBLIC COMMENT.pdf
```

## Protocol Definition

### Service: RTBExtensionPoint

The core gRPC service that agents must implement:

```protobuf
service RTBExtensionPoint {
  rpc GetMutations (RTBRequest) returns (RTBResponse);
}
```

### Key Messages

| Message | Description |
|---------|-------------|
| `RTBRequest` | Contains lifecycle stage, request ID, tmax, bid_request, optional bid_response |
| `RTBResponse` | Returns request ID, list of mutations, and metadata |
| `Mutation` | Defines intent, operation (add/remove/replace), path, and value payload |

### Supported Intents

| Intent | Value | Description |
|--------|-------|-------------|
| `ACTIVATE_SEGMENTS` | 1 | Activate user segments by external segment IDs |
| `ACTIVATE_DEALS` | 2 | Activate deals by external deal IDs |
| `SUPPRESS_DEALS` | 3 | Suppress deals by external deal IDs |
| `ADJUST_DEAL_FLOOR` | 4 | Adjust the bid floor of a specific deal |
| `ADJUST_DEAL_MARGIN` | 5 | Adjust the deal margin |
| `BID_SHADE` | 6 | Adjust the bid price |
| `ADD_METRICS` | 7 | Add metrics to an impression |

### Operations

| Operation | Value | Description |
|-----------|-------|-------------|
| `OPERATION_ADD` | 1 | Add new data |
| `OPERATION_REMOVE` | 2 | Remove existing data |
| `OPERATION_REPLACE` | 3 | Replace existing data |

### Payload Types

- `IDsPayload` - List of string identifiers (for segments, deals)
- `AdjustDealPayload` - Bidfloor and margin adjustments
- `AdjustBidPayload` - Bid price adjustments
- `AddMetricsPayload` - OpenRTB Metric objects

## Development Commands

```bash
# Generate Go code from protobuf (requires protoc and go plugins)
protoc --go_out=. --go-grpc_out=. agenticrtbframework.proto

# Build the server
go build -o artf-server ./cmd/server

# Run the server
./artf-server

# Run tests
go test ./...

# Build Docker image
docker build -t artf-agent:latest .

# Run with Docker
docker run -p 50051:50051 artf-agent:latest
```

## Implementation Requirements

### Container Requirements (from ARTF spec)

1. **Must run as non-root user**
2. **Must implement Kubernetes health probes**:
   - Liveness probe: `GET /health/live`
   - Readiness probe: `GET /health/ready`
3. **Must follow least-privilege principle** - drop unnecessary capabilities
4. **Must handle graceful shutdowns**
5. **No external network access** - only communicate with orchestrator services
6. **Must support OpenTelemetry** for metrics and distributed tracing

### Performance Requirements

- Sub-millisecond response times expected
- Use efficient languages (Go, Rust, Java recommended)
- gRPC with protobuf for serialization efficiency
- Respect `tmax` timeout from requests

### Agent Manifest

Containers must include an `agent-manifest` label in image metadata with:

```json
{
  "name": "agent-name",
  "version": "1.0.0",
  "vendor": "vendor-name",
  "owner": "owner@example.com",
  "resources": {
    "cpu": "500m",
    "memory": "256Mi"
  },
  "intents": ["ACTIVATE_SEGMENTS", "ACTIVATE_DEALS"],
  "dependencies": {
    "serviceName": {
      "service": "svc-name",
      "port": 9000
    }
  },
  "health": {
    "livenessProbe": { "httpGet": { "path": "/health/live", "port": 8080 } },
    "readinessProbe": { "httpGet": { "path": "/health/ready", "port": 8080 } }
  }
}
```

## Code Style Guidelines

- Use Go for implementation (efficient, good gRPC support)
- Follow standard Go project layout
- Use `context.Context` for cancellation and timeouts
- Implement proper error handling with gRPC status codes
- Add structured logging with OpenTelemetry integration
- Write unit tests for all mutation handlers

## Example Mutation Response

```json
{
  "id": "request-123",
  "mutations": [
    {
      "intent": "ACTIVATE_SEGMENTS",
      "op": "OPERATION_ADD",
      "path": "/user/data/segment",
      "ids": {
        "id": ["18-35-age-segment", "soccer-watchers"]
      }
    }
  ],
  "metadata": {
    "api_version": "1.0",
    "model_version": "v2.1"
  }
}
```

## Dependencies

The protobuf imports OpenRTB v2.6 definitions:
```protobuf
import "com/iabtechlab/openrtb/v2.6/openrtb.proto";
```

You'll need the IAB Tech Lab OpenRTB protobuf definitions from:
https://github.com/InteractiveAdvertisingBureau/openrtb

## References

- [Agentic RTB Framework Specification](https://iabtechlab.com/standards/artf/)
- [OpenRTB Specification](https://iabtechlab.com/standards/openrtb/)
- [IAB Tech Lab GitHub](https://github.com/InteractiveAdvertisingBureau)
- [gRPC Go Documentation](https://grpc.io/docs/languages/go/)
- [Protocol Buffers](https://protobuf.dev/)

## License

The Agentic RTB Framework specification is licensed under Creative Commons Attribution 3.0.
