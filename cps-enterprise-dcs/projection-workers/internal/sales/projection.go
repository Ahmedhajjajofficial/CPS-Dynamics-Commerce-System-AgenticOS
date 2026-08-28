package sales

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

// SalesSummaryProjection maintains the sales summary materialized view
type SalesSummaryProjection struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	mu     sync.Mutex
}

// NewSalesSummaryProjection creates a new sales summary projection
func NewSalesSummaryProjection(pool *pgxpool.Pool, logger *zap.Logger) *SalesSummaryProjection {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}
	return &SalesSummaryProjection{pool: pool, logger: logger}
}

// ProcessEvent processes a sales event and updates the projection
func (p *SalesSummaryProjection) ProcessEvent(ctx context.Context, event *kafka.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	eventType := event.Headers["event_type"]
	branchID := event.Headers["branch_id"]

	switch eventType {
	case "SALE_COMPLETED", "SALE_PARTIALLY_REFUNDED", "SALE_REVERSED":
		return p.updateSalesSummary(ctx, branchID, event)
	case "INVENTORY_SOLD":
		return p.updateInventoryDeduction(ctx, branchID, event)
	default:
		return nil
	}
}

func (p *SalesSummaryProjection) updateSalesSummary(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	amount, _ := payload["amount"].(float64)
	quantity, _ := payload["quantity"].(float64)
	eventDate := time.Now().Format("2006-01-02")

	_, err := p.pool.Exec(ctx, `
		INSERT INTO projection_sales_summary (branch_id, date, total_sales, total_transactions, total_items_sold, updated_at)
		VALUES ($1, $2, $3, 1, $4, NOW())
		ON CONFLICT (branch_id, date)
		DO UPDATE SET
			total_sales = projection_sales_summary.total_sales + $3,
			total_transactions = projection_sales_summary.total_transactions + 1,
			total_items_sold = projection_sales_summary.total_items_sold + $4,
			updated_at = NOW()
	`, branchID, eventDate, amount, int64(quantity))

	if err != nil {
		p.logger.Error("failed to update sales summary", zap.Error(err))
		return err
	}

	p.logger.Debug("sales summary updated",
		zap.String("branch_id", branchID),
		zap.String("date", eventDate),
		zap.Float64("amount", amount),
	)

	return nil
}

func (p *SalesSummaryProjection) updateInventoryDeduction(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	productID, _ := payload["product_id"].(string)
	quantity, _ := payload["quantity"].(float64)

	_, err := p.pool.Exec(ctx, `
		UPDATE projection_inventory
		SET current_quantity = current_quantity - $1,
		    available_quantity = available_quantity - $1,
		    updated_at = NOW()
		WHERE branch_id = $2 AND product_id = $3
	`, int64(quantity), branchID, productID)

	if err != nil {
		p.logger.Error("failed to update inventory projection", zap.Error(err))
		return err
	}

	return nil
}

// Start starts the projection worker
func (p *SalesSummaryProjection) Start(ctx context.Context, brokers string, topic string) error {
	consumer, err := kafka.NewConsumer(brokers, topic, "sales-summary-projection", p.logger)
	if err != nil {
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	p.logger.Info("Starting sales summary projection worker",
		zap.String("topic", topic),
	)

	return consumer.Consume(ctx, func(event *kafka.Event) error {
		return p.ProcessEvent(ctx, event)
	})
}
