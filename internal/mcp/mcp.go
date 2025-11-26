// Copyright (c) 2025 Index Exchange Inc.
//
// This file is part of the Agentic RTB Framework reference implementation.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

// Package mcp implements the MCP interface for ARTF
package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/iabtechlab/agentic-rtb-framework/internal/handlers"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// Agent wraps the MCP server with ARTF-specific functionality
type Agent struct {
	mcpServer *server.MCPServer
	handlers  *handlers.MutationHandlers
	addr      string
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

// NewAgent creates a new MCP agent instance
func NewAgent(h *handlers.MutationHandlers, addr string, port int) *Agent {
	// Create MCP server
	mcpServer := server.NewMCPServer(
		"ARTF Agent",
		"0.10.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	a := &Agent{
		mcpServer: mcpServer,
		handlers:  h,
		addr:      addr,
		port:      port,
	}

	// Register the extend_rtb tool
	a.registerTools()

	return a
}

// registerTools registers the ARTF tools with the MCP server
func (a *Agent) registerTools() {
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
	a.mcpServer.AddTool(extendRTBTool, a.handleExtendRTB)
}

// handleExtendRTB processes the extend_rtb tool call
func (a *Agent) handleExtendRTB(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	args := request.GetArguments()
	bidRequestRaw, ok := args["bid_request"].(map[string]interface{})
	if !ok {
		return mcp.NewToolResultError("missing required parameter: bid_request"), nil
	}

	// Get optional bid_response
	var bidResponseRaw map[string]interface{}
	if br, ok := args["bid_response"].(map[string]interface{}); ok {
		bidResponseRaw = br
	}

	log.Printf("MCP: Processing extend_rtb request %s with tmax=%d", id, tmax)

	// Process the request using handlers
	response, err := a.processRequest(ctx, id, tmax, bidRequestRaw, bidResponseRaw)
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
func (a *Agent) processRequest(ctx context.Context, id string, tmax int, bidRequest, bidResponse map[string]interface{}) (*RTBResponse, error) {
	var mutations []Mutation

	// Extract user data for segment activation
	if user, ok := bidRequest["user"].(map[string]interface{}); ok {
		segments := a.determineUserSegments(user)
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
				deals := a.determineDealActivations(imp)
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

								shadedPrice := a.calculateShadedPrice(bidRequest, impID, price)
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
			ModelVersion: "v0.10.0",
		},
	}, nil
}

// determineUserSegments analyzes user data and returns applicable segment IDs
func (a *Agent) determineUserSegments(user map[string]interface{}) []string {
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
func (a *Agent) determineDealActivations(imp map[string]interface{}) []string {
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
func (a *Agent) calculateShadedPrice(bidRequest map[string]interface{}, impID string, originalPrice float64) *float64 {
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

// corsMiddleware wraps an http.Handler with CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Mcp-Session-Id, Last-Event-ID")
		w.Header().Set("Access-Control-Expose-Headers", "Mcp-Session-Id")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// Start starts the MCP interface using Streamable HTTP transport
func (a *Agent) Start() error {
	listenAddr := fmt.Sprintf("%s:%d", a.addr, a.port)
	log.Printf("MCP interface starting on %s", listenAddr)

	// Create streamable HTTP server
	streamableServer := server.NewStreamableHTTPServer(a.mcpServer)

	// Create HTTP mux and wrap with CORS middleware
	mux := http.NewServeMux()
	mux.Handle("/mcp", corsMiddleware(streamableServer))

	// Create and start HTTP server
	httpServer := &http.Server{
		Addr:         listenAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer.ListenAndServe()
}

// GetMCPServer returns the underlying MCP server for custom configuration
func (a *Agent) GetMCPServer() *server.MCPServer {
	return a.mcpServer
}

// Handler returns an HTTP handler for the MCP endpoint with CORS support.
// This can be mounted on an existing mux to serve MCP alongside other routes.
func (a *Agent) Handler() http.Handler {
	streamableServer := server.NewStreamableHTTPServer(a.mcpServer)
	return corsMiddleware(streamableServer)
}
