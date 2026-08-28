package agent

// KafkaProducer defines the interface for Kafka event production
type KafkaProducer interface {
	Send(topic, key string, value []byte) error
	Close() error
}
