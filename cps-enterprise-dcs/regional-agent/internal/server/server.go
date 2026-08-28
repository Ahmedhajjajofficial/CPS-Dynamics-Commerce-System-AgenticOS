package server

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"os"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/store"
	pb "github.com/cps-enterprise/dcs/regional-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/reflection"
)

// Server wraps the gRPC server for the regional agent
type Server struct {
	agent      *agent.RegionalAgent
	logger     *zap.Logger
	grpcServer *grpc.Server
	store      *store.Store
	auth       *AuthInterceptor
}

// New creates a new gRPC server for the regional agent
func New(a *agent.RegionalAgent, logger *zap.Logger, store *store.Store) *Server {
	var s *Server
	if os.Getenv("DCS_ENV") == "production" {
		s = newSecureServer(a, logger, store)
	} else {
		s = newInsecureServer(a, logger, store)
	}

	// Only enable gRPC reflection in development (exposes full API surface)
	if os.Getenv("DCS_ENV") != "production" {
		reflection.Register(s.grpcServer)
	}

	return s
}

func newInsecureServer(a *agent.RegionalAgent, logger *zap.Logger, store *store.Store) *Server {
	s := &Server{
		agent:      a,
		logger:     logger,
		store:      store,
		grpcServer: grpc.NewServer(grpc.ChainUnaryInterceptor(authInterceptor(a, store, logger).UnaryServerInterceptor()), grpc.ChainStreamInterceptor(authInterceptor(a, store, logger).StreamServerInterceptor())),
	}

	// Register handlers
	pb.RegisterAccountingSwarmProtocolServer(s.grpcServer, &SwarmHandler{server: s})
	pb.RegisterQueryProtocolServer(s.grpcServer, &QueryHandler{server: s})

	return s
}

func newSecureServer(a *agent.RegionalAgent, logger *zap.Logger, store *store.Store) *Server {
	certFile := os.Getenv("DCS_TLS_CERT_FILE")
	keyFile := os.Getenv("DCS_TLS_KEY_FILE")
	caFile := os.Getenv("DCS_TLS_CA_FILE")

	if certFile == "" || keyFile == "" {
		logger.Warn("TLS cert/key not configured; falling back to insecure gRPC")
		return newInsecureServer(a, logger, store)
	}

	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		logger.Error("failed to load TLS certificate", zap.Error(err))
		return newInsecureServer(a, logger, store)
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
		store:      store,
		grpcServer: grpc.NewServer(grpc.ChainUnaryInterceptor(authInterceptor(a, store, logger).UnaryServerInterceptor()), grpc.ChainStreamInterceptor(authInterceptor(a, store, logger).StreamServerInterceptor()), grpc.Creds(creds)),
	}

	pb.RegisterAccountingSwarmProtocolServer(s.grpcServer, &SwarmHandler{server: s})
	pb.RegisterQueryProtocolServer(s.grpcServer, &QueryHandler{server: s})

	logger.Info("gRPC server started with mTLS")
	return s
}

// authInterceptor builds an AuthInterceptor with a default allow-all set when DCS_ENV is not production.
func authInterceptor(a *agent.RegionalAgent, store *store.Store, logger *zap.Logger) *AuthInterceptor {
	allowed := make(map[string]struct{})
	if os.Getenv("DCS_ENV") != "production" {
		allowed["*"] = struct{}{}
	}
	return NewAuthInterceptor(a, store, logger, allowed)
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
