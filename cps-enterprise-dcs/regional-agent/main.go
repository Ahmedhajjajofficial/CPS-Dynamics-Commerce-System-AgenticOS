/*
Regional Agent - The Regional Brain
===================================
Coordinates multiple branch agents within a region.

Responsibilities:
- Aggregate data from branch agents
- CRDT merge operations
- Regional forecasting
- Cross-branch coordination
- Raft consensus for master election

Architecture:
    ┌─────────────────────────────────────────┐
    │         Regional Agent                  │
    │  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
    │  │  Raft   │  │  CRDT   │  │Forecast │ │
    │  │Consensus│  │  Merge  │  │ Engine  │ │
    │  └────┬────┘  └────┬────┘  └────┬────┘ │
    │       └─────────────┴─────────────┘     │
    │                   │                     │
    │            ┌─────────────┐              │
    │            │  gRPC Server │              │
    │            │  (Branches)  │              │
    │            └─────────────┘              │
    └─────────────────────────────────────────┘
*/

package main

import (
	"context"
	"crypto/rand"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/config"
	"github.com/cps-enterprise/dcs/regional-agent/internal/server"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	var (
		agentID   = flag.String("agent-id", getEnv("DCS_AGENT_ID", ""), "Unique agent identifier")
		regionID  = flag.String("region-id", getEnv("DCS_REGION_ID", ""), "Region identifier")
		rpcAddr   = flag.String("rpc-addr", getEnv("DCS_RPC_ADDR", ":12000"), "Raft RPC address")
		raftAddr  = flag.String("raft-addr", getEnv("DCS_RAFT_ADDR", ":12001"), "Raft consensus address")
		dataDir   = flag.String("data-dir", getEnv("DCS_DATA_DIR", "./data"), "Data directory")
		bootstrap = flag.Bool("bootstrap", getEnvBool("DCS_BOOTSTRAP", false), "Bootstrap the cluster")
	)
	flag.Parse()

	// Validate required parameters
	if *agentID == "" {
		*agentID = fmt.Sprintf("regional-%s", generateID())
		log.Printf("Generated agent ID: %s", *agentID)
	}

	if *regionID == "" {
		log.Fatal("Error: --region-id is required")
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	// Create configuration
	cfg := &config.Config{
		AgentID:         *agentID,
		RegionID:        *regionID,
		RPCAddress:      *rpcAddr,
		RaftAddress:     *raftAddr,
		DataDir:         *dataDir,
		Bootstrap:       *bootstrap,
		GRPCPort:        getEnvInt("DCS_GRPC_PORT", 50052),
		PostgreSQLURL:   getEnv("DCS_POSTGRESQL_URL", ""),
		RedisURL:        getEnv("DCS_REDIS_URL", "localhost:6379"),
	}

	logger.Info("Starting Regional Agent",
		zap.String("agent_id", cfg.AgentID),
		zap.String("region_id", cfg.RegionID),
		zap.String("rpc_addr", cfg.RPCAddress),
		zap.String("raft_addr", cfg.RaftAddress),
	)

	// Create and initialize agent
	regionalAgent, err := agent.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
	}

	// Initialize agent
	ctx := context.Background()
	if err := regionalAgent.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize agent", zap.Error(err))
	}

	// Start gRPC server
	grpcServer := server.New(regionalAgent, logger)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.GRPCPort)
		logger.Info("Starting gRPC server", zap.String("addr", addr))
		if err := grpcServer.Start(addr); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	// Print startup banner
	fmt.Printf(`
╔══════════════════════════════════════════════════════════════╗
║     CP'S Enterprise DCS - Regional Agent                     ║
║                                                              ║
║  Agent ID:  %-48s ║
║  Region:    %-48s ║
║  RPC Addr:  %-48s ║
║  Raft Addr: %-48s ║
║  gRPC Port: %-48d ║
║                                                              ║
║  Press Ctrl+C to shutdown                                    ║
╚══════════════════════════════════════════════════════════════╝
`, cfg.AgentID, cfg.RegionID, cfg.RPCAddress, cfg.RaftAddress, cfg.GRPCPort)

	// Wait for shutdown signal
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down Regional Agent...")

	// Graceful shutdown
	if err := regionalAgent.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}

	grpcServer.Stop()

	logger.Info("Regional Agent shutdown complete")
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1"
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		result, err := strconv.Atoi(value)
		if err != nil {
			log.Printf("Warning: invalid integer for %s=%q, using default %d: %v", key, value, defaultValue, err)
			return defaultValue
		}
		return result
	}
	return defaultValue
}

func generateID() string {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "unknown"
	}
	return fmt.Sprintf("%x", b)[:8]
}
