package store

import (
	"context"
	"fmt"
	"time"

	"github.com/cps-enterprise/dcs/master-agent/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides access to PostgreSQL for the master agent
type Store struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewStore creates a new Store and verifies connectivity
func NewStore(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Store, error) {
	pool, err := pgxpool.New(ctx, cfg.PostgreSQLURL)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	store := &Store{pool: pool, logger: logger}
	if err := pool.Ping(ctx); err != nil {
		_ = pool.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	logger.Info("PostgreSQL store connected")
	return store, nil
}

// Close closes the underlying connection pool
func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

// InsertReconciliationJob inserts a reconciliation job
func (s *Store) InsertReconciliationJob(ctx context.Context, job *ReconciliationJob) error {
	if s == nil || s.pool == nil {
		return fmt.Errorf("store not configured")
	}

	_, err := s.pool.Exec(ctx, `
		INSERT INTO reconciliation_jobs (job_id, region_id, type, status, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, job.JobID, job.RegionID, job.Type, job.Status, time.Now())
	if err != nil {
		s.logger.Error("failed to insert reconciliation job", zap.Error(err))
		return err
	}

	return nil
}

// ReconciliationJob represents a reconciliation job
type ReconciliationJob struct {
	JobID     string
	RegionID  string
	Type      string
	Status    string
	CreatedAt time.Time
}
