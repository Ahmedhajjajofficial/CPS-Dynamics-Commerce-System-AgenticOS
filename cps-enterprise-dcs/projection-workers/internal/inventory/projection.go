package inventory

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

// InventoryProjection maintains the inventory materialized view
type InventoryProjection struct {
	pool   *pgxpool.Pool
	logger *zap.Logger
	mu     sync.Mutex
}

// NewInventoryProjection creates a new inventory projection
func NewInventoryProjection(pool *pgxpool.Pool, logger *zap.Logger) *InventoryProjection {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}
	return &InventoryProjection{pool: pool, logger: logger}
}

// ProcessEvent processes an inventory event and updates the projection
func (p *InventoryProjection) ProcessEvent(ctx context.Context, event *kafka.Event) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	eventType := event.Headers["event_type"]
	branchID := event.Headers["branch_id"]

	switch eventType {
	case "INVENTORY_RECEIVED":
		return p.handleInventoryReceived(ctx, branchID, event)
	case "INVENTORY_SOLD":
		return p.handleInventorySold(ctx, branchID, event)
	case "INVENTORY_ADJUSTMENT":
		return p.handleInventoryAdjustment(ctx, branchID, event)
	case "LOW_STOCK_ALERT", "STOCKOUT_DETECTED":
		return p.handleStockAlert(ctx, branchID, event)
	default:
		return nil
	}
}

func (p *InventoryProjection) handleInventoryReceived(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	productID, _ := payload["product_id"].(string)
	quantity, _ := payload["quantity"].(float64)

	_, err := p.pool.Exec(ctx, `
		INSERT INTO projection_inventory (branch_id, product_id, current_quantity, available_quantity, last_restocked_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		ON CONFLICT (branch_id, product_id)
		DO UPDATE SET
			current_quantity = projection_inventory.current_quantity + $3,
			available_quantity = projection_inventory.available_quantity + $4,
			last_restocked_at = NOW(),
			updated_at = NOW()
	`, branchID, productID, int64(quantity), int64(quantity))

	if err != nil {
		p.logger.Error("failed to update inventory projection", zap.Error(err))
		return err
	}

	return nil
}

func (p *InventoryProjection) handleInventorySold(ctx context.Context, branchID string, event *kafka.Event) error {
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

func (p *InventoryProjection) handleInventoryAdjustment(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	productID, _ := payload["product_id"].(string)
	adjustment, _ := payload["adjustment_quantity"].(float64)

	_, err := p.pool.Exec(ctx, `
		UPDATE projection_inventory
		SET current_quantity = current_quantity + $1,
		    available_quantity = available_quantity + $1,
		    updated_at = NOW()
		WHERE branch_id = $2 AND product_id = $3
	`, int64(adjustment), branchID, productID)

	if err != nil {
		p.logger.Error("failed to update inventory projection", zap.Error(err))
		return err
	}

	return nil
}

func (p *InventoryProjection) handleStockAlert(ctx context.Context, branchID string, event *kafka.Event) error {
	var payload map[string]interface{}
	if err := json.Unmarshal(event.Value, &payload); err != nil {
		return err
	}

	productID, _ := payload["product_id"].(string)
	alertType := event.Headers["event_type"]

	_, err := p.pool.Exec(ctx, `
		UPDATE projection_inventory
		SET low_stock_alert = $1,
		    updated_at = NOW()
		WHERE branch_id = $2 AND product_id = $3
	`, alertType, branchID, productID)

	if err != nil {
		p.logger.Error("failed to update inventory alert", zap.Error(err))
		return err
	}

	return nil
}

// Start starts the projection worker
func (p *InventoryProjection) Start(ctx context.Context, brokers string, topic string) error {
	consumer, err := kafka.NewConsumer(brokers, topic, "inventory-projection", p.logger)
	if err != nil {
		return fmt.Errorf("failed to create kafka consumer: %w", err)
	}

	p.logger.Info("Starting inventory projection worker",
		zap.String("topic", topic),
	)

	return consumer.Consume(ctx, func(event *kafka.Event) error {
		return p.ProcessEvent(ctx, event)
	})
}
