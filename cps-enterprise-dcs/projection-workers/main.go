package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cps-enterprise/dcs/projection-workers/internal/branch"
	"github.com/cps-enterprise/dcs/projection-workers/internal/inventory"
	"github.com/cps-enterprise/dcs/projection-workers/internal/sales"
	"github.com/cps-enterprise/dcs/projection-workers/pkg/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	var (
		postgresURL = flag.String("postgres-url", getEnv("DATABASE_URL", "postgres://dcs_admin:CHANGE_ME_POSTGRES@localhost:5432/dcs_eventstore?sslmode=require"), "PostgreSQL connection URL")
		kafkaBrokers = flag.String("kafka-brokers", getEnv("KAFKA_BROKERS", "localhost:9092"), "Kafka brokers")
	)
	flag.Parse()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Sync()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	pool, err := pgxpool.New(ctx, *postgresURL)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("Failed to ping PostgreSQL", zap.Error(err))
	}
	logger.Info("Connected to PostgreSQL")

	salesProjection := sales.NewSalesSummaryProjection(pool, logger)
	inventoryProjection := inventory.NewInventoryProjection(pool, logger)
	branchProjection := branch.NewBranchMetricsProjection(pool, logger)

	var wg sync.WaitGroup

	wg.Add(3)
	go func() {
		defer wg.Done()
		if err := salesProjection.Start(ctx, *kafkaBrokers, "regional.events.*"); err != nil {
			logger.Error("Sales projection failed", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := inventoryProjection.Start(ctx, *kafkaBrokers, "regional.events.*"); err != nil {
			logger.Error("Inventory projection failed", zap.Error(err))
		}
	}()

	go func() {
		defer wg.Done()
		if err := branchProjection.Start(ctx, *kafkaBrokers, "regional.events.*"); err != nil {
			logger.Error("Branch metrics projection failed", zap.Error(err))
		}
	}()

	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║     CP'S Enterprise DCS - Projection Workers               ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	logger.Info("Shutting down projection workers...")
	cancel()
	wg.Wait()
	logger.Info("Projection workers shutdown complete")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
