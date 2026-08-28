package config

// Config holds the master agent configuration
type Config struct {
	AgentID         string
	RegionID        string
	GRPCPort        int
	PostgreSQLURL   string
	KafkaBrokers    string
	DataDir         string
}
