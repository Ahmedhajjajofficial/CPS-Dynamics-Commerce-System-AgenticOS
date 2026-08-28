package kafka

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
)

// Producer implements KafkaProducer using segmentio/kafka-go
type Producer struct {
	writer *kafka.Writer
	logger *zap.Logger
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers string, logger *zap.Logger) (*Producer, error) {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:   []string{brokers},
		Topic:     "master.events",
		Balancer:  &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll,
		Async:     false,
		BatchSize: 100,
		BatchTimeout: 10 * time.Millisecond,
	})

	return &Producer{
		writer: writer,
		logger: logger,
	}, nil
}

// Send sends a message to Kafka
func (p *Producer) Send(topic, key string, value []byte) error {
	msg := kafka.Message{
		Topic: topic,
		Key:   []byte(key),
		Value: value,
	}

	if err := p.writer.WriteMessages(nil, msg); err != nil {
		p.logger.Error("failed to send kafka message",
			zap.Error(err),
			zap.String("topic", topic),
			zap.String("key", key),
		)
		return err
	}

	p.logger.Debug("kafka message sent",
		zap.String("topic", topic),
		zap.String("key", key),
	)

	return nil
}

// Close closes the Kafka producer
func (p *Producer) Close() error {
	if p.writer != nil {
		return p.writer.Close()
	}
	return nil
}

// GenerateMessageID generates a unique message ID
func GenerateMessageID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%s-%s", time.Now().Format("20060102150405"), hex.EncodeToString(b))
}
