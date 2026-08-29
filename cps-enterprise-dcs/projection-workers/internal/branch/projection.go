package branch

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/cps-enterprise/dcs/projection-workers/pkg/kafka"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// BranchMetricsProjection maintains branch-level metrics
type BranchMetricsProjection struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	mu     sync.Mutex
}

// NewBranchMetricsProjection creates a new branch metrics projection
func NewBranchMetricsProjection(pool *pgxpool.Pool, logger *zap.Logger) *BranchMetricsProjection {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}
	return &BranchMetricsProjection{pool: pool, logger: logger}
}

// ProcessEvent processes a branch event and updates metrics
func (p *BranchMetricsProjection) ProcessEvent(ctx context.Context, event *kafka.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	eventType := event.Headers["event_type"]
	branchID := event.Headers["branch_id"]

	switch eventType {
	case "SALE_COMPLETED", "SALE_REVERSED", "SALE_PARTIALLY_REFUNDED":
		return p.updateBranchTransactionMetrics(ctx, branchID, event)
	case "TEMPERATURE_WARNING", "TEMPERATURE_ANOMALY_DETECTED":
		return p.updateComplianceMetrics(ctx, branchID, event)
	case "AGENT_DECISION_MADE":
		return p.updateDecisionMetrics(ctx, branchID, event)
	default:
		return nil
	}
}

func (p *BranchMetricsProjection) updateBranchTransactionMetrics(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	amount, _ := payload["amount"].(float64)
	today := time.Now().Format("2006-01-02")

	_, err := p.pool.Exec(ctx, `
		INSERT INTO projection_branch_metrics (branch_id, date, total_transactions, total_amount, updated_at)
		VALUES ($1, $2, 1, $3, NOW())
		ON CONFLICT (branch_id, date)
		DO UPDATE SET
			total_transactions = projection_branch_metrics.total_transactions + 1,
			total_amount = projection_branch_metrics.total_amount + $3,
			updated_at = NOW()
	`, branchID, today, amount)

	if err != nil {
		p.logger.Error("failed to update branch metrics", zap.Error(err))
		return err
	}

	return nil
}

func (p *BranchMetricsProjection) updateComplianceMetrics(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	today := time.Now().Format("2006-01-02")
	_, err := p.pool.Exec(ctx, `
		INSERT INTO projection_branch_metrics (branch_id, date, compliance_events, updated_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (branch_id, date)
		DO UPDATE SET
			compliance_events = projection_branch_metrics.compliance_events + 1,
			updated_at = NOW()
	`, branchID, today)

	if err != nil {
		p.logger.Error("failed to update compliance metrics", zap.Error(err))
		return err
	}

	return nil
}

func (p *BranchMetricsProjection) updateDecisionMetrics(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	decisionType, _ := payload["decision_type"].(string)
	confidence, _ := payload["confidence_score"].(float64)
	today := time.Now().Format("2006-01-02")

	_, err := p.pool.Exec(ctx, `
		INSERT INTO projection_branch_metrics (branch_id, date, decisions_made, avg_decision_confidence, updated_at)
		VALUES ($1, $2, 1, $3, NOW())
		ON CONFLICT (branch_id, date)
		DO UPDATE SET
			decisions_made = projection_branch_metrics.decisions_made + 1,
			avg_decision_confidence = (projection_branch_metrics.avg_decision_confidence + $3) / 2,
			updated_at = NOW()
	`, branchID, today, confidence)

	if err != nil {
		p.logger.Error("failed to update decision metrics", zap.Error(err))
		return err
	}

	return nil
}

// Start starts the projection worker
func (p *BranchMetricsProjection) Start(ctx context.Context, brokers string, topic string) error {
	consumer, err := kafka.NewConsumer(brokers, topic, "branch-metrics-projection", p.logger)
	if err != nil {
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	p.logger.Info("Starting branch metrics projection worker",
		zap.String("topic", topic),
	)

	return consumer.Consume(ctx, func(event *kafka.Event) error {
		return p.ProcessEvent(ctx, event)
	})
}
