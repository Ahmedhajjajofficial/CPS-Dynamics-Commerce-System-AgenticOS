package server

import (
	"net"
	"os"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server for the master agent
type Server struct {
	agent      *agent.MasterAgent
	logger     *zap.Logger
	grpcServer *grpc.Server
}

// New creates a new gRPC server for the master agent
func New(a *agent.MasterAgent, logger *zap.Logger) *Server {
	s := &Server{
		agent:      a,
		logger:     logger,
		grpcServer: grpc.NewServer(),
	}

	// Register handlers
	pb.RegisterGlobalReconciliationServiceServer(s.grpcServer, &ReconciliationHandler{agent: a})
	pb.RegisterDecisionEngineServiceServer(s.grpcServer, &DecisionEngineHandler{agent: a})
	pb.RegisterMasterCoordinationServiceServer(s.grpcServer, &CoordinationHandler{agent: a})
	pb.RegisterEventStreamingServiceServer(s.grpcServer, &EventStreamingHandler{agent: a})

	// Only enable gRPC reflection in development
	if os.Getenv("DCS_ENV") != "production" {
		reflection.Register(s.grpcServer)
	}

	return s
}

// Start starts the gRPC server
func (s *Server) Start(addr string) error {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s.logger.Info("Starting gRPC server", zap.String("addr", addr))
	return s.grpcServer.Serve(listener)
}

// Stop stops the gRPC server
func (s *Server) Stop() {
	s.grpcServer.GracefulStop()
}
