package server

import (
	"net"
	"os"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/regional-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server for the regional agent
type Server struct {
	agent      *agent.RegionalAgent
	logger     *zap.Logger
	grpcServer *grpc.Server
}

// New creates a new gRPC server for the regional agent
func New(a *agent.RegionalAgent, logger *zap.Logger) *Server {
	s := &Server{
		agent:      a,
		logger:     logger,
		grpcServer: grpc.NewServer(),
	}

	// Register handlers
	pb.RegisterAccountingSwarmProtocolServer(s.grpcServer, &SwarmHandler{server: s})
	pb.RegisterQueryProtocolServer(s.grpcServer, &QueryHandler{server: s})

	// Only enable gRPC reflection in development (exposes full API surface)
	if os.Getenv("DCS_ENV") != "production" {
		reflection.Register(s.grpcServer)
	}

	return s
}

// Start begins listening for gRPC connections on the given address
func (s *Server) Start(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	s.logger.Info("gRPC server listening", zap.String("addr", addr))
	return s.grpcServer.Serve(lis)
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
	s.logger.Info("gRPC server stopped")
}
