package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Event represents a Kafka event
type Event struct {
	Topic     string
	Partition int
	Offset    int64
	Key       []byte
	Value     []byte
	Headers   map[string]string
	Timestamp time.Time
}

// Consumer consumes events from Kafka
type Consumer struct {
	reader *kafka.Reader
	logger *zap.Logger
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(brokers, topic, groupID string, logger *zap.Logger) (*Consumer, error) {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{brokers},
		Topic:     topic,
		GroupID:   groupID,
		MinBytes:  1,
		MaxBytes:  10e6,
		CommitInterval: time.Second,
		StartOffset: kafka.FirstOffset,
	})

	return &Consumer{
		reader: reader,
		logger: logger,
	}, nil
}

// Consume consumes events from Kafka
func (c *Consumer) Consume(ctx context.Context, handler func(*Event) error) error {
	c.logger.Info("Starting Kafka consumer")

	for {
		select {
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping consumer")
			return nil
		default:
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				c.logger.Error("Failed to read message", zap.Error(err))
				continue
			}

			headers := make(map[string]string)
			for _, h := range msg.Headers {
				headers[h.Key] = string(h.Value)
			}

			event := &Event{
				Topic:     msg.Topic,
				Partition: msg.Partition,
				Offset:    msg.Offset,
				Key:       msg.Key,
				Value:     msg.Value,
				Headers:   headers,
				Timestamp: msg.Time,
			}

			if err := handler(event); err != nil {
				c.logger.Error("Failed to process event", zap.Error(err))
				continue
			}
		}
	}
}

// Close closes the consumer
func (c *Consumer) Close() error {
	return c.reader.Close()
}
