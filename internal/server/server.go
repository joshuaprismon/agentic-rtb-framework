// Package server implements the ARTF gRPC service
package server

import (
	"context"
	"log"
	"time"

	"github.com/iabtechlab/agentic-rtb-framework/internal/handlers"
	pb "github.com/iabtechlab/agentic-rtb-framework/pkg/pb/artf"
	"google.golang.org/grpc"
)

// ARTFServer implements the RTBExtensionPoint gRPC service
type ARTFServer struct {
	pb.UnimplementedRTBExtensionPointServer
	handlers *handlers.MutationHandlers
}

// NewARTFServer creates a new ARTF server instance
func NewARTFServer(h *handlers.MutationHandlers) *ARTFServer {
	return &ARTFServer{
		handlers: h,
	}
}

// GetMutations processes an RTB request and returns proposed mutations
func (s *ARTFServer) GetMutations(ctx context.Context, req *pb.RTBRequest) (*pb.RTBResponse, error) {
	startTime := time.Now()

	// Check timeout budget
	tmax := req.GetTmax()
	if tmax > 0 {
		deadline := time.Now().Add(time.Duration(tmax) * time.Millisecond)
		var cancel context.CancelFunc
		ctx, cancel = context.WithDeadline(ctx, deadline)
		defer cancel()
	}

	// Collect mutations from all registered handlers
	var mutations []*pb.Mutation

	// Process based on lifecycle stage
	lifecycle := req.GetLifecycle()
	bidRequest := req.GetBidRequest()
	bidResponse := req.GetBidResponse()

	log.Printf("Processing request %s at lifecycle stage %v", req.GetId(), lifecycle)

	// Run segment activation handler
	if segmentMutations, err := s.handlers.ProcessSegments(ctx, bidRequest); err == nil {
		mutations = append(mutations, segmentMutations...)
	} else {
		log.Printf("Segment processing error: %v", err)
	}

	// Run deal activation handler
	if dealMutations, err := s.handlers.ProcessDeals(ctx, bidRequest); err == nil {
		mutations = append(mutations, dealMutations...)
	} else {
		log.Printf("Deal processing error: %v", err)
	}

	// Run bid shading handler (if bid response is present)
	if bidResponse != nil {
		if bidMutations, err := s.handlers.ProcessBidShading(ctx, bidRequest, bidResponse); err == nil {
			mutations = append(mutations, bidMutations...)
		} else {
			log.Printf("Bid shading error: %v", err)
		}
	}

	// Build response
	response := &pb.RTBResponse{
		Id:        req.Id,
		Mutations: mutations,
		Metadata: &pb.Metadata{
			ApiVersion:   stringPtr("1.0"),
			ModelVersion: stringPtr("v1.0.0"),
		},
	}

	log.Printf("Request %s processed in %v, returning %d mutations",
		req.GetId(), time.Since(startTime), len(mutations))

	return response, nil
}

// LoggingInterceptor logs gRPC requests
func LoggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	log.Printf("gRPC %s took %v, error: %v", info.FullMethod, time.Since(start), err)
	return resp, err
}

// RegisterRTBExtensionPointServer registers the service with a gRPC server
func RegisterRTBExtensionPointServer(s *grpc.Server, srv *ARTFServer) {
	pb.RegisterRTBExtensionPointServer(s, srv)
}

func stringPtr(s string) *string {
	return &s
}
