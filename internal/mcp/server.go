// Package mcp implements the MCP server for ARTF
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/iabtechlab/agentic-rtb-framework/internal/handlers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Server wraps the MCP server with ARTF-specific functionality
type Server struct {
	mcpServer *server.MCPServer
	handlers  *handlers.MutationHandlers
	port      int
}

// RTBRequest represents the MCP tool input for extend_rtb
type RTBRequest struct {
	Lifecycle   string                 `json:"lifecycle,omitempty"`
	ID          string                 `json:"id"`
	Tmax        int                    `json:"tmax,omitempty"`
	BidRequest  map[string]interface{} `json:"bid_request"`
	BidResponse map[string]interface{} `json:"bid_response,omitempty"`
}

// RTBResponse represents the MCP tool output
type RTBResponse struct {
	ID        string     `json:"id"`
	Mutations []Mutation `json:"mutations"`
	Metadata  Metadata   `json:"metadata"`
}

// Mutation represents a single mutation in the response
type Mutation struct {
	Intent string                 `json:"intent"`
	Op     string                 `json:"op"`
	Path   string                 `json:"path"`
	Value  map[string]interface{} `json:"value,omitempty"`
}

// Metadata contains response metadata
type Metadata struct {
	APIVersion   string `json:"api_version"`
	ModelVersion string `json:"model_version"`
}

// NewServer creates a new MCP server instance
func NewServer(h *handlers.MutationHandlers, port int) *Server {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"ARTF Agent",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &Server{
		mcpServer: mcpServer,
		handlers:  h,
		port:      port,
	}

	// Register the extend_rtb tool
	s.registerTools()

	return s
}

// registerTools registers the ARTF tools with the MCP server
func (s *Server) registerTools() {
	// Define the extend_rtb tool
	extendRTBTool := mcp.NewTool("extend_rtb",
		mcp.WithDescription("Process an OpenRTB bid request/response and return proposed mutations for segment activation, deal management, bid shading, and metrics"),
		mcp.WithString("id",
			mcp.Required(),
			mcp.Description("Unique request ID"),
		),
		mcp.WithNumber("tmax",
			mcp.Description("Maximum response time in milliseconds"),
		),
		mcp.WithObject("bid_request",
			mcp.Required(),
			mcp.Description("OpenRTB v2.6 BidRequest object"),
		),
		mcp.WithObject("bid_response",
			mcp.Description("OpenRTB v2.6 BidResponse object (optional)"),
		),
		mcp.WithString("lifecycle",
			mcp.Description("Auction lifecycle stage"),
		),
	)

	// Register tool with handler
	s.mcpServer.AddTool(extendRTBTool, s.handleExtendRTB)
}

// handleExtendRTB processes the extend_rtb tool call
func (s *Server) handleExtendRTB(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	startTime := time.Now()

	// Extract parameters
	id, err := request.RequireString("id")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: id"), nil
	}

	// Get optional tmax
	tmax := 100 // default
	if tmaxVal, err := request.RequireFloat("tmax"); err == nil {
		tmax = int(tmaxVal)
	}

	// Get bid_request
	bidRequestRaw, err := request.RequireObject("bid_request")
	if err != nil {
		return mcp.NewToolResultError("missing required parameter: bid_request"), nil
	}

	// Get optional bid_response
	var bidResponseRaw map[string]interface{}
	if br, err := request.RequireObject("bid_response"); err == nil {
		bidResponseRaw = br
	}

	log.Printf("MCP: Processing extend_rtb request %s with tmax=%d", id, tmax)

	// Process the request using handlers
	response, err := s.processRequest(ctx, id, tmax, bidRequestRaw, bidResponseRaw)
	if err != nil {
		log.Printf("MCP: Error processing request: %v", err)
		return mcp.NewToolResultError(fmt.Sprintf("processing error: %v", err)), nil
	}

	// Serialize response to JSON
	responseJSON, err := json.Marshal(response)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("serialization error: %v", err)), nil
	}

	log.Printf("MCP: Request %s processed in %v, returning %d mutations",
		id, time.Since(startTime), len(response.Mutations))

	return mcp.NewToolResultText(string(responseJSON)), nil
}

// processRequest processes the RTB request and generates mutations
func (s *Server) processRequest(ctx context.Context, id string, tmax int, bidRequest, bidResponse map[string]interface{}) (*RTBResponse, error) {
	var mutations []Mutation

	// Extract user data for segment activation
	if user, ok := bidRequest["user"].(map[string]interface{}); ok {
		segments := s.determineUserSegments(user)
		if len(segments) > 0 {
			mutations = append(mutations, Mutation{
				Intent: "ACTIVATE_SEGMENTS",
				Op:     "OPERATION_ADD",
				Path:   "/user/data/segment",
				Value: map[string]interface{}{
					"ids": map[string]interface{}{
						"id": segments,
					},
				},
			})
		}
	}

	// Process impressions for deal activation
	if imps, ok := bidRequest["imp"].([]interface{}); ok {
		for _, impRaw := range imps {
			if imp, ok := impRaw.(map[string]interface{}); ok {
				impID, _ := imp["id"].(string)
				deals := s.determineDealActivations(imp)
				if len(deals) > 0 {
					mutations = append(mutations, Mutation{
						Intent: "ACTIVATE_DEALS",
						Op:     "OPERATION_ADD",
						Path:   "/imp/" + impID,
						Value: map[string]interface{}{
							"ids": map[string]interface{}{
								"id": deals,
							},
						},
					})
				}
			}
		}
	}

	// Process bid response for bid shading
	if bidResponse != nil {
		if seatbids, ok := bidResponse["seatbid"].([]interface{}); ok {
			for _, seatbidRaw := range seatbids {
				if seatbid, ok := seatbidRaw.(map[string]interface{}); ok {
					seat, _ := seatbid["seat"].(string)
					if bids, ok := seatbid["bid"].([]interface{}); ok {
						for _, bidRaw := range bids {
							if bid, ok := bidRaw.(map[string]interface{}); ok {
								bidID, _ := bid["id"].(string)
								price, _ := bid["price"].(float64)
								impID, _ := bid["impid"].(string)

								shadedPrice := s.calculateShadedPrice(bidRequest, impID, price)
								if shadedPrice != nil && *shadedPrice != price {
									mutations = append(mutations, Mutation{
										Intent: "BID_SHADE",
										Op:     "OPERATION_REPLACE",
										Path:   fmt.Sprintf("/seatbid/%s/bid/%s", seat, bidID),
										Value: map[string]interface{}{
											"adjust_bid": map[string]interface{}{
												"price": *shadedPrice,
											},
										},
									})
								}
							}
						}
					}
				}
			}
		}
	}

	return &RTBResponse{
		ID:        id,
		Mutations: mutations,
		Metadata: Metadata{
			APIVersion:   "1.0",
			ModelVersion: "v1.0.0",
		},
	}, nil
}

// determineUserSegments analyzes user data and returns applicable segment IDs
func (s *Server) determineUserSegments(user map[string]interface{}) []string {
	var segments []string

	// Re-activate existing segments from user.data
	if data, ok := user["data"].([]interface{}); ok {
		for _, dataRaw := range data {
			if dataObj, ok := dataRaw.(map[string]interface{}); ok {
				if segs, ok := dataObj["segment"].([]interface{}); ok {
					for _, segRaw := range segs {
						if seg, ok := segRaw.(map[string]interface{}); ok {
							if id, ok := seg["id"].(string); ok && id != "" {
								segments = append(segments, id)
							}
						}
					}
				}
			}
		}
	}

	// Add demographic segments based on year of birth
	if yob, ok := user["yob"].(float64); ok && yob > 0 {
		age := 2024 - int(yob)
		switch {
		case age >= 18 && age <= 24:
			segments = append(segments, "demo-18-24")
		case age >= 25 && age <= 34:
			segments = append(segments, "demo-25-34")
		case age >= 35 && age <= 44:
			segments = append(segments, "demo-35-44")
		case age >= 45:
			segments = append(segments, "demo-45-plus")
		}
	}

	// Add gender segment
	if gender, ok := user["gender"].(string); ok {
		switch gender {
		case "M":
			segments = append(segments, "gender-male")
		case "F":
			segments = append(segments, "gender-female")
		}
	}

	return segments
}

// determineDealActivations returns deal IDs to activate for an impression
func (s *Server) determineDealActivations(imp map[string]interface{}) []string {
	var deals []string

	// Check bidfloor for premium deal activation
	if bidfloor, ok := imp["bidfloor"].(float64); ok && bidfloor >= 5.0 {
		deals = append(deals, "premium-deal-001")
	}

	// Check for video impression
	if _, ok := imp["video"]; ok {
		deals = append(deals, "video-deal-001")
	}

	// Check for native impression
	if _, ok := imp["native"]; ok {
		deals = append(deals, "native-deal-001")
	}

	// Check for banner impression
	if _, ok := imp["banner"]; ok {
		deals = append(deals, "display-deal-001")
	}

	return deals
}

// calculateShadedPrice calculates the optimal shaded bid price
func (s *Server) calculateShadedPrice(bidRequest map[string]interface{}, impID string, originalPrice float64) *float64 {
	if originalPrice <= 0 {
		return nil
	}

	// Find the impression to get bidfloor
	var bidfloor float64
	if imps, ok := bidRequest["imp"].([]interface{}); ok {
		for _, impRaw := range imps {
			if imp, ok := impRaw.(map[string]interface{}); ok {
				if id, _ := imp["id"].(string); id == impID {
					bidfloor, _ = imp["bidfloor"].(float64)
					break
				}
			}
		}
	}

	if bidfloor <= 0 {
		return nil
	}

	// Calculate shade percentage based on margin above floor
	var shadePercent float64
	margin := originalPrice - bidfloor
	switch {
	case margin > bidfloor*0.5:
		shadePercent = 0.15 // 15% shade
	case margin > bidfloor*0.2:
		shadePercent = 0.10 // 10% shade
	default:
		shadePercent = 0.05 // 5% shade
	}

	shadedPrice := originalPrice * (1 - shadePercent)
	return &shadedPrice
}

// Start starts the MCP server using Streamable HTTP transport
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	log.Printf("MCP server starting on %s", addr)

	// Create streamable HTTP server
	httpServer := server.NewStreamableHTTPServer(s.mcpServer)

	return httpServer.Start(addr)
}

// GetMCPServer returns the underlying MCP server for custom configuration
func (s *Server) GetMCPServer() *server.MCPServer {
	return s.mcpServer
}
