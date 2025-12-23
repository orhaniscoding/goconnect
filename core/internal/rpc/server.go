package rpc

import (
	"fmt"
	"net"

	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server and provides lifecycle management.
type Server struct {
	grpcServer *grpc.Server
}

// NewServer creates a new gRPC server instance.
func NewServer() *Server {
	s := grpc.NewServer()
	
	// Register reflection for debugging (e.g., with grpcurl)
	reflection.Register(s)
	
	return &Server{
		grpcServer: s,
	}
}

// Start begins listening for requests on the provided listener.
func (s *Server) Start(lis net.Listener) error {
	logger.Info("gRPC Server listening", "addr", lis.Addr().String())
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("gRPC server failed: %w", err)
	}
	return nil
}

// GracefulStop stops the gRPC server gracefully.
func (s *Server) GracefulStop() {
	logger.Info("Stopping gRPC Server")
	s.grpcServer.GracefulStop()
}

// GetGRPCServer returns the underlying gRPC server for service registration.
func (s *Server) GetGRPCServer() *grpc.Server {
	return s.grpcServer
}
