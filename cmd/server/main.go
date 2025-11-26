// Package main implements the ARTF server with gRPC, MCP, and Web interfaces
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/iabtechlab/agentic-rtb-framework/internal/handlers"
	"github.com/iabtechlab/agentic-rtb-framework/internal/health"
	"github.com/iabtechlab/agentic-rtb-framework/internal/mcp"
	"github.com/iabtechlab/agentic-rtb-framework/internal/server"
	"github.com/iabtechlab/agentic-rtb-framework/internal/web"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Version information (set at build time)
var (
	Version   = "dev"
	BuildTime = "unknown"
)

// Command-line flags
var (
	// Feature flags
	enableGRPC = flag.Bool("enable-grpc", true, "Enable gRPC server")
	enableMCP  = flag.Bool("enable-mcp", false, "Enable MCP server")
	enableWeb  = flag.Bool("enable-web", false, "Enable web interface")

	// Port configuration
	grpcPort   = flag.Int("grpc-port", 50051, "gRPC server port")
	mcpPort    = flag.Int("mcp-port", 50052, "MCP server port")
	webPort    = flag.Int("web-port", 8081, "Web interface port")
	healthPort = flag.Int("health-port", 8080, "Health check HTTP port")

	// Version flag
	showVersion = flag.Bool("version", false, "Show version information")
)

func main() {
	flag.Parse()

	// Show version and exit
	if *showVersion {
		fmt.Printf("ARTF Server v%s (built: %s)\n", Version, BuildTime)
		os.Exit(0)
	}

	log.Printf("Starting ARTF Agent Server v%s", Version)
	log.Printf("Features: gRPC=%v, MCP=%v, Web=%v", *enableGRPC, *enableMCP, *enableWeb)

	// Validate that at least one service is enabled
	if !*enableGRPC && !*enableMCP && !*enableWeb {
		log.Fatal("At least one service must be enabled (--enable-grpc, --enable-mcp, or --enable-web)")
	}

	// Create shared components
	healthChecker := health.NewChecker()
	mutationHandlers := handlers.NewMutationHandlers()

	// Track servers for shutdown
	var grpcServer *grpc.Server
	var mcpServer *mcp.Server
	var webServer *http.Server
	var healthServer *http.Server

	// Start gRPC server
	if *enableGRPC {
		grpcServer = grpc.NewServer(
			grpc.UnaryInterceptor(server.LoggingInterceptor),
		)

		artfServer := server.NewARTFServer(mutationHandlers)
		server.RegisterRTBExtensionPointServer(grpcServer, artfServer)
		reflection.Register(grpcServer)

		grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", *grpcPort))
		if err != nil {
			log.Fatalf("Failed to listen on gRPC port %d: %v", *grpcPort, err)
		}

		go func() {
			log.Printf("gRPC server listening on port %d", *grpcPort)
			if err := grpcServer.Serve(grpcListener); err != nil {
				log.Printf("gRPC server error: %v", err)
			}
		}()
	}

	// Start MCP server
	if *enableMCP {
		mcpServer = mcp.NewServer(mutationHandlers, *mcpPort)

		go func() {
			log.Printf("MCP server listening on port %d", *mcpPort)
			if err := mcpServer.Start(); err != nil {
				log.Printf("MCP server error: %v", err)
			}
		}()
	}

	// Start Web server
	if *enableWeb {
		mcpEndpoint := fmt.Sprintf("http://localhost:%d/mcp", *mcpPort)
		if !*enableMCP {
			// If MCP is disabled, web interface will show a warning
			mcpEndpoint = "/mcp-disabled"
		}

		webHandler, err := web.NewHandler(mcpEndpoint)
		if err != nil {
			log.Fatalf("Failed to create web handler: %v", err)
		}

		webMux := http.NewServeMux()
		webHandler.RegisterRoutes(webMux)

		webServer = &http.Server{
			Addr:         fmt.Sprintf(":%d", *webPort),
			Handler:      webMux,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}

		go func() {
			log.Printf("Web interface listening on port %d", *webPort)
			if err := webServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("Web server error: %v", err)
			}
		}()
	}

	// Start health check HTTP server (always enabled)
	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health/live", healthChecker.LivenessHandler)
	healthMux.HandleFunc("/health/ready", healthChecker.ReadinessHandler)
	healthMux.HandleFunc("/health/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"version":"%s","build_time":"%s","grpc":%v,"mcp":%v,"web":%v}`,
			Version, BuildTime, *enableGRPC, *enableMCP, *enableWeb)
	})

	healthServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", *healthPort),
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		log.Printf("Health check server listening on port %d", *healthPort)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Health server error: %v", err)
		}
	}()

	// Mark server as ready
	healthChecker.SetReady(true)
	log.Printf("Server is ready to accept requests")

	// Print service URLs
	printServiceURLs()

	// Wait for shutdown signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Printf("Shutting down server...")

	// Mark as not ready during shutdown
	healthChecker.SetReady(false)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown servers
	if grpcServer != nil {
		grpcServer.GracefulStop()
		log.Printf("gRPC server stopped")
	}

	if webServer != nil {
		if err := webServer.Shutdown(ctx); err != nil {
			log.Printf("Web server shutdown error: %v", err)
		} else {
			log.Printf("Web server stopped")
		}
	}

	if err := healthServer.Shutdown(ctx); err != nil {
		log.Printf("Health server shutdown error: %v", err)
	}

	log.Printf("Server stopped")
}

// printServiceURLs prints the URLs for enabled services
func printServiceURLs() {
	log.Println("=== Service URLs ===")
	if *enableGRPC {
		log.Printf("  gRPC:   localhost:%d", *grpcPort)
	}
	if *enableMCP {
		log.Printf("  MCP:    http://localhost:%d/mcp", *mcpPort)
	}
	if *enableWeb {
		log.Printf("  Web UI: http://localhost:%d/", *webPort)
	}
	log.Printf("  Health: http://localhost:%d/health/ready", *healthPort)
	log.Println("====================")
}
