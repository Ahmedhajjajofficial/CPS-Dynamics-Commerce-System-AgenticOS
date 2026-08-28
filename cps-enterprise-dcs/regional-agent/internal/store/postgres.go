package store

import (
	"context"
	"fmt"
	"time"

	"github.com/cps-enterprise/dcs/regional-agent/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Store provides access to PostgreSQL projections for the regional agent.
type Store struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
}

// NewStore creates a new Store and verifies connectivity.
func NewStore(ctx context.Context, cfg *config.Config, logger *zap.Logger) (*Store, error) {
	pool, err := pgxpool.New(ctx, cfg.PostgreSQLURL)
	if err != nil {
		return nil, err
	}

	store := &Store{pool: pool, logger: logger}
	if err := pool.Ping(ctx); err != nil {
		_ = pool.Close()
		return nil, err
	}

	logger.Info("PostgreSQL store connected")
	return store, nil
}

// Close closes the underlying connection pool.
func (s *Store) Close() {
	if s == nil || s.pool == nil {
		return
	}
	s.pool.Close()
}

// GetBranchSummary returns live projection data for a branch.
func (s *Store) GetBranchSummary(ctx context.Context, branchID string) (todaySales float64, todayTransactions int32, currentBalance float64, activeSessions int32, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(total_sales), 0), COALESCE(SUM(total_transactions), 0)
		FROM projection_sales_summary
		WHERE branch_id = $1 AND date = $2`,
		branchID, time.Now().UTC().Format("2006-01-02"),
	).Scan(&todaySales, &todayTransactions)
	if err != nil {
		return 0, 0, 0, 0, err
	}

	err = s.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		FROM stream_metadata
		WHERE branch_id = $1 AND current_version > 0`,
		branchID,
	).Scan(&activeSessions)
	if err != nil {
		return todaySales, todayTransactions, 0, 0, err
	}

	currentBalance = todaySales
	return todaySales, todayTransactions, currentBalance, activeSessions, nil
}

// GetInventoryStatus returns live inventory data for a product in a branch.
func (s *Store) GetInventoryStatus(ctx context.Context, branchID, productID string) (currentQty, availableQty int32, isLowStock bool, err error) {
	err = s.pool.QueryRow(ctx, `
		SELECT current_quantity, available_quantity, available_quantity <= reorder_point
		FROM projection_inventory
		WHERE branch_id = $1 AND product_id = $2`,
		branchID, productID,
	).Scan(&currentQty, &availableQty, &isLowStock)
	if err != nil {
		return 0, 0, false, err
	}
	return currentQty, availableQty, isLowStock, nil
}

// InsertAuditLog inserts an audit log entry. It is best-effort; failures are logged but not returned.
func (s *Store) InsertAuditLog(ctx context.Context, action, actorID, resourceType, resourceID string, details map[string]interface{}) {
	if s == nil || s.pool == nil {
		return
	}
	_, err := s.pool.Exec(ctx, `
		INSERT INTO audit_log (action, actor_id, actor_type, resource_type, resource_id, details)
		VALUES ($1, $2, 'SYSTEM', $3, $4, $5)
	`, action, actorID, resourceType, resourceID, details)
	if err != nil {
		s.logger.Error("failed to insert audit log", zap.Error(err))
	}
}

// Query executes a query and returns rows for scanning.
// The caller must close the returned rows.
func (s *Store) Query(ctx context.Context, query string, args ...interface{}) (pgxpool.Rows, error) {
	if s == nil || s.pool == nil {
		return nil, fmt.Errorf("store not configured")
	}
	return s.pool.Query(ctx, query, args...)
}
