# MCP Server Integration

This document describes the Model Context Protocol (MCP) server integration for the Agentic RTB Framework, enabling AI agents and LLMs to interact with RTB extension points through the standardized MCP protocol.

---

## Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [MCP Tool Definition](#mcp-tool-definition)
4. [Transport Options](#transport-options)
5. [Web Interface](#web-interface)
6. [Sample Payloads](#sample-payloads)
7. [Configuration](#configuration)
8. [Usage Examples](#usage-examples)

---

## Overview

The ARTF implementation exposes the `RTBExtensionPoint` service via MCP as a tool called **"Extend RTB"**. This enables:

- **AI Agent Integration** - LLMs can invoke RTB mutations through MCP tool calls
- **Streamable HTTP** - Web-based clients can interact via HTTP streaming
- **Web UI** - Built-in web component for testing and demonstration
- **Dual Protocol Support** - Run gRPC and MCP simultaneously on different ports

### Why MCP?

Per the ARTF specification, while gRPC is mandated for service-to-service communication, MCP enables **model-to-agent orchestration** for autonomic agentic flows. MCP provides:

- Standardized tool definitions for LLM interaction
- JSON-RPC based communication (natural successor to REST)
- Support for both structured and streaming responses
- OAuth authentication support (MCP 2025-06-18+)

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                        ARTF Agent                                │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐           │
│  │   gRPC       │  │    MCP       │  │    Web       │           │
│  │  :50051      │  │   :50052     │  │   :8081      │           │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘           │
│         │                 │                 │                    │
│         └─────────────────┼─────────────────┘                    │
│                           │                                      │
│                    ┌──────▼───────┐                              │
│                    │   Handlers   │                              │
│                    │  (Shared)    │                              │
│                    └──────────────┘                              │
│                                                                  │
│  ┌──────────────┐                                                │
│  │   Health     │                                                │
│  │   :8080      │                                                │
│  └──────────────┘                                                │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### Port Assignments

| Port | Protocol | Service | Description |
|------|----------|---------|-------------|
| 50051 | gRPC | RTBExtensionPoint | Primary service endpoint |
| 50052 | HTTP/SSE | MCP Streamable HTTP | MCP tool endpoint |
| 8080 | HTTP | Health Checks | Liveness/readiness probes |
| 8081 | HTTP | Web UI | Testing interface |

---

## MCP Tool Definition

### Tool: "Extend RTB"

The MCP server exposes a single tool that wraps the `RTBExtensionPoint.GetMutations` RPC.

#### Tool Schema

```json
{
  "name": "extend_rtb",
  "description": "Process an OpenRTB bid request/response and return proposed mutations for segment activation, deal management, bid shading, and metrics",
  "inputSchema": {
    "type": "object",
    "properties": {
      "lifecycle": {
        "type": "string",
        "description": "Auction lifecycle stage",
        "enum": ["LIFECYCLE_UNSPECIFIED"],
        "default": "LIFECYCLE_UNSPECIFIED"
      },
      "id": {
        "type": "string",
        "description": "Unique request ID"
      },
      "tmax": {
        "type": "integer",
        "description": "Maximum response time in milliseconds",
        "default": 100
      },
      "bid_request": {
        "type": "object",
        "description": "OpenRTB v2.6 BidRequest object",
        "required": true
      },
      "bid_response": {
        "type": "object",
        "description": "OpenRTB v2.6 BidResponse object (optional)"
      }
    },
    "required": ["id", "bid_request"]
  }
}
```

#### Response Format

```json
{
  "content": [
    {
      "type": "text",
      "text": "{\"id\":\"req-123\",\"mutations\":[...],\"metadata\":{...}}"
    }
  ]
}
```

---

## Transport Options

### Streamable HTTP (Recommended)

The MCP server uses Streamable HTTP transport for web compatibility:

```
POST /mcp    - Send JSON-RPC requests
GET  /mcp    - Establish SSE stream for server messages
DELETE /mcp  - Terminate session
```

#### Request Flow

1. Client sends `initialize` request via POST
2. Server returns session ID in `Mcp-Session-Id` header
3. Client includes session ID in subsequent requests
4. Server streams responses via SSE (GET connection)

### SSE Transport (Alternative)

For clients requiring traditional SSE:

```
GET  /sse      - Establish SSE connection
POST /message  - Send JSON-RPC messages
```

---

## Web Interface

The built-in web interface provides:

- **Request Builder** - Visual form for constructing ORTB payloads
- **Sample Payloads** - Pre-built examples for common scenarios
- **Live Response** - Real-time mutation results display
- **MCP Inspector** - Debug MCP message flow

### Endpoints

| Path | Description |
|------|-------------|
| `/` | Main web interface |
| `/api/samples` | List available sample payloads |
| `/api/samples/{name}` | Get specific sample payload |
| `/mcp` | MCP streamable HTTP endpoint |

### Web Component

The interface uses a custom web component `<artf-tester>` that can be embedded:

```html
<script type="module" src="/static/artf-tester.js"></script>
<artf-tester mcp-endpoint="/mcp"></artf-tester>
```

---

## Sample Payloads

### Basic Banner Request

```json
{
  "id": "sample-banner-001",
  "tmax": 100,
  "bid_request": {
    "id": "auction-123",
    "imp": [
      {
        "id": "imp-1",
        "banner": {
          "w": 300,
          "h": 250,
          "pos": 1
        },
        "bidfloor": 1.50,
        "bidfloorcur": "USD"
      }
    ],
    "site": {
      "id": "site-456",
      "domain": "example.com",
      "cat": ["IAB1"],
      "page": "https://example.com/article"
    },
    "user": {
      "id": "user-789",
      "yob": 1990,
      "gender": "M",
      "data": [
        {
          "id": "data-provider-1",
          "name": "Example DMP",
          "segment": [
            {"id": "seg-sports", "name": "Sports Enthusiast"}
          ]
        }
      ]
    },
    "device": {
      "ua": "Mozilla/5.0...",
      "ip": "192.168.1.1",
      "geo": {
        "country": "USA",
        "region": "CA"
      }
    }
  }
}
```

### Video Request with Deals

```json
{
  "id": "sample-video-001",
  "tmax": 150,
  "bid_request": {
    "id": "auction-456",
    "imp": [
      {
        "id": "imp-1",
        "video": {
          "mimes": ["video/mp4"],
          "minduration": 15,
          "maxduration": 30,
          "w": 640,
          "h": 480
        },
        "bidfloor": 8.00,
        "pmp": {
          "private_auction": 1,
          "deals": [
            {
              "id": "deal-premium-video",
              "bidfloor": 10.00,
              "at": 1
            }
          ]
        }
      }
    ],
    "site": {
      "id": "site-789",
      "domain": "streaming.example.com",
      "cat": ["IAB1-6"]
    },
    "user": {
      "id": "user-456",
      "yob": 1985
    }
  }
}
```

### Bid Response with Shading

```json
{
  "id": "sample-bidshade-001",
  "tmax": 100,
  "bid_request": {
    "id": "auction-789",
    "imp": [
      {
        "id": "imp-1",
        "banner": {"w": 728, "h": 90},
        "bidfloor": 2.00
      }
    ],
    "user": {"id": "user-123"}
  },
  "bid_response": {
    "id": "auction-789",
    "seatbid": [
      {
        "seat": "dsp-001",
        "bid": [
          {
            "id": "bid-abc",
            "impid": "imp-1",
            "price": 5.50,
            "adomain": ["advertiser.com"]
          }
        ]
      }
    ]
  }
}
```

---

## Configuration

### Command-Line Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--enable-grpc` | true | Enable gRPC interface |
| `--enable-mcp` | false | Enable MCP interface |
| `--enable-web` | false | Enable web interface |
| `--grpc-port` | 50051 | gRPC port |
| `--mcp-port` | 50052 | MCP port |
| `--web-port` | 8081 | Web UI port |
| `--health-port` | 8080 | Health check port |

### Environment Variables

| Variable | Description |
|----------|-------------|
| `ARTF_ENABLE_GRPC` | Enable gRPC server |
| `ARTF_ENABLE_MCP` | Enable MCP server |
| `ARTF_ENABLE_WEB` | Enable web interface |
| `ARTF_GRPC_PORT` | gRPC port |
| `ARTF_MCP_PORT` | MCP port |
| `ARTF_WEB_PORT` | Web UI port |

### Example Configurations

**gRPC Only (Production)**
```bash
./artf-agent --enable-grpc --grpc-port=50051
```

**All Services (Development)**
```bash
./artf-agent --enable-grpc --enable-mcp --enable-web
```

**MCP Only (AI Integration)**
```bash
./artf-agent --enable-mcp --mcp-port=50052
```

---

## Usage Examples

### MCP Client (Python)

```python
import httpx
import json

MCP_ENDPOINT = "http://localhost:50052/mcp"

# Initialize session
init_request = {
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
        "protocolVersion": "2024-11-05",
        "capabilities": {},
        "clientInfo": {"name": "test-client", "version": "1.0"}
    }
}

response = httpx.post(MCP_ENDPOINT, json=init_request)
session_id = response.headers.get("Mcp-Session-Id")

# Call extend_rtb tool
tool_request = {
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
        "name": "extend_rtb",
        "arguments": {
            "id": "test-001",
            "tmax": 100,
            "bid_request": {
                "id": "auction-123",
                "imp": [{"id": "imp-1", "banner": {"w": 300, "h": 250}}],
                "user": {"id": "user-789", "yob": 1990}
            }
        }
    }
}

response = httpx.post(
    MCP_ENDPOINT,
    json=tool_request,
    headers={"Mcp-Session-Id": session_id}
)
print(json.dumps(response.json(), indent=2))
```

### cURL Example

```bash
# Initialize session
curl -X POST http://localhost:50052/mcp \
  -H "Content-Type: application/json" \
  -d '{
    "jsonrpc": "2.0",
    "id": 1,
    "method": "initialize",
    "params": {
      "protocolVersion": "2024-11-05",
      "capabilities": {},
      "clientInfo": {"name": "curl", "version": "1.0"}
    }
  }' -i

# Note the Mcp-Session-Id header from response, then:

# Call tool
curl -X POST http://localhost:50052/mcp \
  -H "Content-Type: application/json" \
  -H "Mcp-Session-Id: <session-id>" \
  -d '{
    "jsonrpc": "2.0",
    "id": 2,
    "method": "tools/call",
    "params": {
      "name": "extend_rtb",
      "arguments": {
        "id": "test-001",
        "bid_request": {
          "id": "auction-123",
          "imp": [{"id": "imp-1", "banner": {"w": 300, "h": 250}}],
          "user": {"id": "user-789", "yob": 1990}
        }
      }
    }
  }'
```

### Claude Desktop Configuration

Add to `claude_desktop_config.json`:

```json
{
  "mcpServers": {
    "artf": {
      "command": "/path/to/artf-agent",
      "args": ["--enable-mcp", "--mcp-port=50052"],
      "env": {}
    }
  }
}
```

---

## Default Implementation: Segment Activation

The default MCP handler implements segment activation based on user demographics from the RTB payload:

### Logic

1. **Extract User Data** - Parse `user.yob`, `user.gender`, existing segments
2. **Determine Segments** - Apply demographic rules:
   - Age 18-24 → `demo-18-24`
   - Age 25-34 → `demo-25-34`
   - Age 35-44 → `demo-35-44`
   - Age 45+ → `demo-45-plus`
3. **Enrich Segments** - Re-activate existing segments from `user.data`
4. **Return Mutations** - Generate `ACTIVATE_SEGMENTS` mutations

### Example Response

```json
{
  "id": "test-001",
  "mutations": [
    {
      "intent": "ACTIVATE_SEGMENTS",
      "op": "OPERATION_ADD",
      "path": "/user/data/segment",
      "ids": {
        "id": ["demo-25-34", "seg-sports"]
      }
    }
  ],
  "metadata": {
    "api_version": "1.0",
    "model_version": "v1.0.0"
  }
}
```

---

## CORS Support

The MCP interface includes full CORS (Cross-Origin Resource Sharing) support to enable web browsers to communicate with the MCP endpoint from different origins (e.g., the Web UI on port 8081 accessing MCP on port 50052).

### CORS Headers

| Header | Value | Description |
|--------|-------|-------------|
| `Access-Control-Allow-Origin` | `*` | Allow requests from any origin |
| `Access-Control-Allow-Methods` | `GET, POST, DELETE, OPTIONS` | Allowed HTTP methods |
| `Access-Control-Allow-Headers` | `Content-Type, Mcp-Session-Id, Last-Event-ID` | Allowed request headers |
| `Access-Control-Expose-Headers` | `Mcp-Session-Id` | Headers exposed to browser |

### Preflight Handling

The MCP server automatically handles `OPTIONS` preflight requests, returning a `200 OK` with appropriate CORS headers. This enables browsers to make cross-origin requests without additional configuration.

### Example Cross-Origin Request

```javascript
// From Web UI on localhost:8081 to MCP on localhost:50052
fetch('http://localhost:50052/mcp', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Mcp-Session-Id': sessionId
  },
  body: JSON.stringify({
    jsonrpc: '2.0',
    id: 1,
    method: 'tools/call',
    params: { name: 'extend_rtb', arguments: {...} }
  })
});
```

---

## Security Considerations

### MCP-Specific Security

- **Session Management** - Use stateful sessions in production
- **Authentication** - Implement OAuth for external access
- **Rate Limiting** - Apply per-session rate limits
- **Input Validation** - Validate all ORTB payloads
- **CORS Restrictions** - In production, restrict `Access-Control-Allow-Origin` to specific domains

### Network Isolation

In containerized deployments, the MCP port should be:
- Exposed only to authorized AI systems
- Protected by network policies
- Monitored for unusual patterns

---

## References

- [Model Context Protocol Specification](https://modelcontextprotocol.io/)
- [mcp-go Library](https://github.com/mark3labs/mcp-go)
- [MCP Streamable HTTP Transport](https://spec.modelcontextprotocol.io/specification/basic/transports/#streamable-http)
- [ARTF Specification](https://iabtechlab.com/standards/artf/)

---

*Document Version: 0.10.0*
*Last Updated: November 2025*
