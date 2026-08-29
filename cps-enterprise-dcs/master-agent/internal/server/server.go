package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
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
	var s *Server
	if os.Getenv("DCS_ENV") == "production" {
		s = newSecureServer(a, logger)
	} else {
		s = newInsecureServer(a, logger)
	}

	return s
}

func newInsecureServer(a *agent.MasterAgent, logger *zap.Logger) *Server {
	s := &Server{
		agent:      a,
		logger:     logger,
		grpcServer: grpc.NewServer(),
	}

	pb.RegisterGlobalReconciliationServiceServer(s.grpcServer, &ReconciliationHandler{agent: a})
	pb.RegisterDecisionEngineServiceServer(s.grpcServer, &DecisionEngineHandler{agent: a})
	pb.RegisterMasterCoordinationServiceServer(s.grpcServer, &CoordinationHandler{agent: a})
	pb.RegisterEventStreamingServiceServer(s.grpcServer, &EventStreamingHandler{agent: a})

	if os.Getenv("DCS_ENV") != "production" {
		reflection.Register(s.grpcServer)
	}

	return s
}

func newSecureServer(a *agent.MasterAgent, logger *zap.Logger) *Server {
	certFile := os.Getenv("DCS_TLS_CERT_FILE")
	keyFile := os.Getenv("DCS_TLS_KEY_FILE")
	caFile := os.Getenv("DCS_TLS_CA_FILE")

	if certFile == "" || keyFile == "" {
		logger.Warn("TLS cert/key not configured; falling back to insecure gRPC")
		return newInsecureServer(a, logger)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Error("failed to load TLS certificate", zap.Error(err))
		return newInsecureServer(a, logger)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}

	if caFile != "" {
		caCert, err := os.ReadFile(caFile)
		if err == nil {
			caPool := x509.NewCertPool()
			if caPool.AppendCertsFromPEM(caCert) {
				tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
				tlsConfig.ClientCAs = caPool
			}
		}
	}

	creds := credentials.NewTLS(tlsConfig)
	s := &Server{
		agent:      a,
		logger:     logger,
		grpcServer: grpc.NewServer(grpc.Creds(creds)),
	}

	pb.RegisterGlobalReconciliationServiceServer(s.grpcServer, &ReconciliationHandler{agent: a})
	pb.RegisterDecisionEngineServiceServer(s.grpcServer, &DecisionEngineHandler{agent: a})
	pb.RegisterMasterCoordinationServiceServer(s.grpcServer, &CoordinationHandler{agent: a})
	pb.RegisterEventStreamingServiceServer(s.grpcServer, &EventStreamingHandler{agent: a})

	logger.Info("gRPC server started with mTLS")
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
