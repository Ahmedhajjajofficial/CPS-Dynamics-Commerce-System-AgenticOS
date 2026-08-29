package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	"github.com/cps-enterprise/dcs/master-agent/internal/config"
	"github.com/cps-enterprise/dcs/master-agent/internal/server"
	"go.uber.org/zap"
)

func main() {
	var (
		agentID   = flag.String("agent-id", getEnv("DCS_AGENT_ID", "master-001"), "Unique agent identifier")
		regionID  = flag.String("region-id", getEnv("DCS_REGION_ID", "global"), "Region identifier")
		grpcPort  = flag.Int("grpc-port", getEnvInt("DCS_GRPC_PORT", 50053), "gRPC server port")
		dataDir   = flag.String("data-dir", getEnv("DCS_DATA_DIR", "./data"), "Data directory")
	)

	flag.Parse()

	if *agentID == "" {
		log.Fatal("Error: --agent-id is required")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	cfg := &config.Config{
		AgentID:       *agentID,
		RegionID:      *regionID,
		GRPCPort:      *grpcPort,
		DataDir:       *dataDir,
	}

	logger.Info("Starting Master Agent",
		zap.String("agent_id", cfg.AgentID),
		zap.String("region_id", cfg.RegionID),
	)

	masterAgent, err := agent.New(cfg, logger)
	if err != nil {
		logger.Fatal("Failed to create agent", zap.Error(err))
	}

	ctx := context.Background()
	if err := masterAgent.Initialize(ctx); err != nil {
		logger.Fatal("Failed to initialize agent", zap.Error(err))
	}

	grpcServer := server.New(masterAgent, logger)
	go func() {
		addr := fmt.Sprintf(":%d", cfg.GRPCPort)
		logger.Info("Starting gRPC server", zap.String("addr", addr))
		if err := grpcServer.Start(addr); err != nil {
			logger.Fatal("Failed to start gRPC server", zap.Error(err))
		}
	}()

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":    "healthy",
			"agent_id":  cfg.AgentID,
			"region_id": cfg.RegionID,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	})
	healthMux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if masterAgent.GetState().String() == "ACTIVE" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("not_ready"))
		}
	})

	healthServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.GRPCPort+1),
		Handler:      healthMux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	go func() {
		logger.Info("Starting health server", zap.String("addr", healthServer.Addr))
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Health server error", zap.Error(err))
		}
	}()

	fmt.Printf("╔══════════════════════════════════════════════════════════════╗\n")
	fmt.Printf("║     CP'S Enterprise DCS - Master Agent                       ║\n")
	fmt.Printf("║                                                              ║\n")
	fmt.Printf("║  Agent ID:  %-48s ║\n", cfg.AgentID)
	fmt.Printf("║  Region:    %-48s ║\n", cfg.RegionID)
	fmt.Printf("║  gRPC Port: %-48d ║\n", cfg.GRPCPort)
	fmt.Printf("║                                                              ║\n")
	fmt.Printf("║  Press Ctrl+C to shutdown                                    ║\n")
	fmt.Printf("╚══════════════════════════════════════════════════════════════╝\n")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down Master Agent...")
	if err := masterAgent.Shutdown(ctx); err != nil {
		logger.Error("Error during shutdown", zap.Error(err))
	}
	grpcServer.Stop()
	
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error shutting down health server", zap.Error(err))
	}
	
	logger.Info("Master Agent shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		if _, err := fmt.Sscanf(value, "%d", &result); err == nil {
			return result
		}
	}
	return defaultValue
}
