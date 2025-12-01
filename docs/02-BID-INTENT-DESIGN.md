# BID Intent Design Document

**Status:** Draft
**Version:** 0.2.0
**Author:** Index Exchange
**Date:** December 2025

---

## 1. Executive Summary

This document specifies the design for a new `BID` intent in the Agentic RTB Framework that enables pre-registered bidding on inventory. The design introduces a two-phase workflow: bid registration (via the new `register_bid` MCP tool) and bid execution (via the existing `extend_rtb` tool during auction).

---

## 2. Goals

1. Enable agents to pre-register bids for specific deal inventory
2. Support programmatic guaranteed and preferred deal execution
3. Maintain separation between bid registration and auction-time execution
4. Use only standard OpenRTB bid fields (no extensions)

---

## 3. Architecture

### 3.1 Two-Tool Model

```
┌─────────────────────────────────────────────────────────────────────┐
│                        ARTF Agent                                   │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌─────────────────────┐         ┌─────────────────────┐           │
│  │    register_bid     │         │     extend_rtb      │           │
│  │    (MCP Tool)       │         │    (MCP Tool)       │           │
│  │                     │         │                     │           │
│  │  Pre-auction        │         │  Auction-time       │           │
│  │  Bid Registration   │         │  Bid Execution      │           │
│  └──────────┬──────────┘         └──────────┬──────────┘           │
│             │                               │                       │
│             ▼                               ▼                       │
│  ┌─────────────────────────────────────────────────────┐           │
│  │                  Bid Registry                        │           │
│  │                                                      │           │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐          │           │
│  │  │ Deal-001 │  │ Deal-002 │  │ Deal-003 │  ...     │           │
│  │  │ CPM: $12 │  │ CPM: $8  │  │ CPM: $15 │          │           │
│  │  │ ADM: ... │  │ ADM: ... │  │ ADM: ... │          │           │
│  │  └──────────┘  └──────────┘  └──────────┘          │           │
│  └─────────────────────────────────────────────────────┘           │
│                                                                     │
│  ┌─────────────────────┐                                           │
│  │       gRPC          │  Same bid registry, different interface   │
│  │  RTBExtensionPoint  │                                           │
│  └─────────────────────┘                                           │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### 3.2 Workflow

```
Phase 1: Registration (Pre-Auction)
┌──────────┐                    ┌──────────┐
│  Agent   │  register_bid      │  ARTF    │
│  (LLM)   │ ─────────────────► │  Agent   │
│          │                    │          │
│          │  confirmation      │          │
│          │ ◄───────────────── │          │
└──────────┘                    └──────────┘

Phase 2: Execution (Auction-Time)
┌──────────┐                    ┌──────────┐                    ┌──────────┐
│   Host   │  RTBRequest        │  ARTF    │  lookup            │   Bid    │
│ Platform │ ─────────────────► │  Agent   │ ────────────────►  │ Registry │
│          │                    │          │                    │          │
│          │  BID mutation      │          │  matched bid       │          │
│          │ ◄───────────────── │          │ ◄──────────────────│          │
└──────────┘                    └──────────┘                    └──────────┘
```

---

## 4. MCP Tool Definitions

### 4.1 Tool: `register_bid`

Registers a bid for future execution when matching inventory appears.

#### Input Schema

```json
{
  "name": "register_bid",
  "description": "Register a pre-configured bid for execution when matching deal inventory appears in the bidstream",
  "inputSchema": {
    "type": "object",
    "properties": {
      "dealid": {
        "type": "string",
        "description": "Deal ID to target. Bid will execute when this deal appears in a bid request."
      },
      "price": {
        "type": "number",
        "description": "Bid price in CPM. Must meet or exceed the deal floor."
      },
      "adm": {
        "type": "string",
        "description": "Ad markup (HTML for banner, VAST XML for video, Native JSON for native)"
      },
      "adomain": {
        "type": "array",
        "items": {"type": "string"},
        "description": "Advertiser domain(s) for the creative"
      },
      "crid": {
        "type": "string",
        "description": "Creative ID for tracking and reporting"
      },
      "cid": {
        "type": "string",
        "description": "Campaign ID"
      },
      "cat": {
        "type": "array",
        "items": {"type": "string"},
        "description": "IAB content categories of the creative"
      },
      "attr": {
        "type": "array",
        "items": {"type": "integer"},
        "description": "Creative attributes (OpenRTB Table 5.3)"
      },
      "w": {
        "type": "integer",
        "description": "Creative width in pixels"
      },
      "h": {
        "type": "integer",
        "description": "Creative height in pixels"
      },
      "nurl": {
        "type": "string",
        "description": "Win notice URL. Supports ${AUCTION_PRICE} macro."
      },
      "burl": {
        "type": "string",
        "description": "Billing notice URL. Supports ${AUCTION_PRICE} macro."
      },
      "lurl": {
        "type": "string",
        "description": "Loss notice URL."
      },
      "iurl": {
        "type": "string",
        "description": "URL to creative preview image"
      },
      "adid": {
        "type": "string",
        "description": "Pre-registered ad ID with the exchange"
      },
      "language": {
        "type": "string",
        "description": "Language of the creative (ISO-639-1-alpha-2)"
      },
      "exp": {
        "type": "integer",
        "description": "Registration expiration in seconds from now. Default: 86400 (24 hours)"
      }
    },
    "required": ["dealid", "price", "adm", "adomain", "crid"]
  }
}
```

#### Response

```json
{
  "content": [{
    "type": "text",
    "text": "{\"status\":\"registered\",\"bid_id\":\"bid-uuid-123\",\"dealid\":\"deal-001\",\"expires_at\":\"2025-12-02T12:00:00Z\"}"
  }]
}
```

### 4.2 Tool: `extend_rtb` (Existing - Updated)

The existing tool continues to process bid requests and now returns BID mutations for registered bids.

#### Behavior Change

When processing an `RTBRequest`:
1. Check if any registered bids match deals in `imp.pmp.deals`
2. If match found, include a `BID` mutation in the response
3. Continue processing other intents (segments, deal activation, etc.)

---

## 5. Intent Definition

### 5.1 BID Intent

| Field | Value |
|-------|-------|
| Name | `BID` |
| Value | `8` |
| Description | Submit a bid for matching inventory |
| Operation | `OPERATION_ADD` |
| Path | `/seatbid/-/bid` |

### 5.2 Protobuf Changes

```protobuf
// Add to Intent enum
enum Intent {
  INTENT_UNSPECIFIED = 0;
  ACTIVATE_SEGMENTS = 1;
  ACTIVATE_DEALS = 2;
  SUPPRESS_DEALS = 3;
  ADJUST_DEAL_FLOOR = 4;
  ADJUST_DEAL_MARGIN = 5;
  BID_SHADE = 6;
  ADD_METRICS = 7;
  BID = 8;  // NEW
}

// New BidPayload message - Standard OpenRTB Bid fields only
message BidPayload {
  // Required
  string id = 1;                  // Bid ID
  string impid = 2;               // Impression ID
  double price = 3;               // CPM price

  // Targeting
  string dealid = 4;              // Deal ID

  // Creative content
  string adm = 5;                 // Ad markup
  string nurl = 6;                // Win notice URL
  string burl = 7;                // Billing notice URL
  string lurl = 8;                // Loss notice URL

  // Creative metadata
  string adid = 9;                // Pre-registered ad ID
  repeated string adomain = 10;   // Advertiser domains
  string iurl = 11;               // Preview image URL
  string cid = 12;                // Campaign ID
  string crid = 13;               // Creative ID
  repeated string cat = 14;       // IAB categories
  repeated int32 attr = 15;       // Creative attributes

  // Dimensions
  int32 w = 16;                   // Width
  int32 h = 17;                   // Height

  // Additional standard fields
  string language = 18;           // Creative language
  int32 exp = 19;                 // Bid expiration (seconds)
  int32 api = 20;                 // API framework
  int32 protocol = 21;            // Video protocol
  int32 qagmediarating = 22;      // QAG media rating
}

// Update Mutation message
message Mutation {
  Intent intent = 1;
  Operation op = 2;
  string path = 3;

  oneof payload {
    IDsPayload ids = 4;
    AdjustDealPayload adjust_deal = 5;
    AdjustBidPayload adjust_bid = 6;
    AddMetricsPayload metrics = 7;
    BidPayload bid = 8;           // NEW
  }
}
```

---

## 6. Bid Registry

### 6.1 Data Model

```go
type RegisteredBid struct {
    ID        string            // Unique bid registration ID
    DealID    string            // Target deal ID
    Price     float64           // CPM price
    Bid       *BidPayload       // Full bid payload
    CreatedAt time.Time         // Registration time
    ExpiresAt time.Time         // Expiration time
    Metadata  map[string]string // Optional metadata
}

type BidRegistry interface {
    // Register a new bid
    Register(ctx context.Context, bid *RegisteredBid) error

    // Find bids matching deal IDs
    FindByDeals(ctx context.Context, dealIDs []string) ([]*RegisteredBid, error)

    // Remove a registered bid
    Remove(ctx context.Context, bidID string) error

    // List all registered bids
    List(ctx context.Context) ([]*RegisteredBid, error)

    // Clean expired bids
    Cleanup(ctx context.Context) error
}
```

### 6.2 Storage Options

| Option | Pros | Cons | Recommended For |
|--------|------|------|-----------------|
| In-Memory | Fast, simple | Lost on restart | Development, single instance |
| Redis | Fast, distributed | External dependency | Production clusters |
| SQLite | Persistent, simple | Single node | Small deployments |

For the reference implementation, in-memory storage with optional persistence.

---

## 7. Standard OpenRTB Bid Fields

Only the following standard OpenRTB 2.6 Bid object fields are supported:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `id` | string | Yes | Bidder-generated bid ID |
| `impid` | string | Yes | ID of the impression being bid on |
| `price` | float | Yes | Bid price in CPM |
| `dealid` | string | Conditional | Reference to deal from bid request |
| `adm` | string | Recommended | Ad markup |
| `nurl` | string | No | Win notice URL |
| `burl` | string | No | Billing notice URL |
| `lurl` | string | No | Loss notice URL |
| `adid` | string | No | Pre-registered ad ID |
| `adomain` | string[] | Recommended | Advertiser domains |
| `iurl` | string | No | Creative preview URL |
| `cid` | string | No | Campaign ID |
| `crid` | string | Recommended | Creative ID |
| `cat` | string[] | No | IAB categories |
| `attr` | int[] | No | Creative attributes |
| `w` | int | No | Width in pixels |
| `h` | int | No | Height in pixels |
| `language` | string | No | Creative language |
| `exp` | int | No | Bid expiration in seconds |
| `api` | int | No | API framework |
| `protocol` | int | No | Video protocol |
| `qagmediarating` | int | No | QAG media rating |

**Note:** Extension fields (`ext`) are explicitly not supported to maintain simplicity and interoperability.

---

## 8. Validation Rules

### 8.1 Registration Validation

| Rule | Description |
|------|-------------|
| V1 | `dealid` is required and non-empty |
| V2 | `price` must be > 0 |
| V3 | `adm` is required and non-empty |
| V4 | `adomain` must have at least one entry |
| V5 | `crid` is required for creative tracking |
| V6 | `exp` must be between 60 and 604800 seconds (1 min to 7 days) |

### 8.2 Execution Validation

| Rule | Description |
|------|-------------|
| E1 | `dealid` must exist in `imp.pmp.deals` |
| E2 | `price` must meet or exceed `deal.bidfloor` |
| E3 | Bid must not be expired |
| E4 | One bid per deal per impression |

---

## 9. Example Flows

### 9.1 Registration Flow

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "register_bid",
    "arguments": {
      "dealid": "deal-premium-video-001",
      "price": 15.00,
      "adm": "<VAST version=\"4.0\"><Ad>...</Ad></VAST>",
      "adomain": ["brand.com"],
      "crid": "creative-video-001",
      "cid": "campaign-q4-2025",
      "cat": ["IAB1-1"],
      "w": 640,
      "h": 480,
      "exp": 86400
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"status\":\"registered\",\"bid_id\":\"reg-abc-123\",\"dealid\":\"deal-premium-video-001\",\"expires_at\":\"2025-12-02T12:00:00Z\"}"
    }]
  }
}
```

### 9.2 Execution Flow

**Request (via extend_rtb):**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "extend_rtb",
    "arguments": {
      "id": "req-001",
      "bid_request": {
        "id": "auction-xyz",
        "imp": [{
          "id": "imp-1",
          "video": {
            "mimes": ["video/mp4"],
            "w": 640,
            "h": 480
          },
          "pmp": {
            "deals": [{
              "id": "deal-premium-video-001",
              "bidfloor": 10.00
            }]
          }
        }]
      }
    }
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "result": {
    "content": [{
      "type": "text",
      "text": "{\"id\":\"req-001\",\"mutations\":[{\"intent\":\"BID\",\"op\":\"OPERATION_ADD\",\"path\":\"/seatbid/-/bid\",\"bid\":{\"id\":\"bid-exec-001\",\"impid\":\"imp-1\",\"price\":15.00,\"dealid\":\"deal-premium-video-001\",\"adm\":\"<VAST>...</VAST>\",\"adomain\":[\"brand.com\"],\"crid\":\"creative-video-001\",\"w\":640,\"h\":480}}],\"metadata\":{\"registered_bid_id\":\"reg-abc-123\"}}"
    }]
  }
}
```

---

## 10. gRPC Interface

The gRPC interface uses the same bid registry. Registered bids are returned as BID mutations in `RTBResponse`.

```protobuf
// No changes to service definition
service RTBExtensionPoint {
  rpc GetMutations (RTBRequest) returns (RTBResponse);
}
```

The handler checks the bid registry when processing requests and includes matching bids in the mutation list.

---

## 11. Implementation Plan

### Phase 1: Core Infrastructure
- [ ] Add `BID` intent to proto enum
- [ ] Define `BidPayload` message
- [ ] Update `Mutation` message with bid payload
- [ ] Generate protobuf code

### Phase 2: Bid Registry
- [ ] Implement `BidRegistry` interface
- [ ] In-memory storage implementation
- [ ] Expiration cleanup goroutine
- [ ] Unit tests

### Phase 3: MCP Integration
- [ ] Implement `register_bid` tool
- [ ] Update `extend_rtb` to check registry
- [ ] Add bid registration endpoints
- [ ] Integration tests

### Phase 4: gRPC Integration
- [ ] Update handler to use bid registry
- [ ] Return BID mutations
- [ ] End-to-end tests

### Phase 5: Documentation & UI
- [ ] Update Web UI for bid registration
- [ ] Add sample payloads
- [ ] Update CLAUDE.md
- [ ] Update MCP documentation

---

## 12. Security Considerations

1. **Bid Integrity** - Registered bids cannot be modified after registration
2. **Expiration** - All registrations must have an expiration
3. **Rate Limiting** - Limit registrations per time window
4. **Authorization** - Future: restrict registration to authorized deals
5. **Audit Logging** - Log all registrations and executions

---

## 13. Future Enhancements

1. **Persistent Storage** - Redis/database backend for production
2. **Multi-Seat Support** - Seat-specific bid registrations
3. **Budget Pacing** - Daily/hourly budget limits
4. **Frequency Capping** - User-level frequency limits
5. **A/B Testing** - Creative rotation and testing
6. **Analytics** - Win rate, spend tracking

---

*Document Version: 0.2.0*
*Last Updated: December 2025*
