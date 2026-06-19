package config

import (
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.GRPCPort != 50052 {
		t.Fatalf("expected GRPCPort 50052, got %d", cfg.GRPCPort)
	}
	if cfg.MaxConnections != 1000 {
		t.Fatalf("expected MaxConnections 1000, got %d", cfg.MaxConnections)
	}
	if cfg.BatchSize != 100 {
		t.Fatalf("expected BatchSize 100, got %d", cfg.BatchSize)
	}
	if cfg.SyncIntervalSec != 30 {
		t.Fatalf("expected SyncIntervalSec 30, got %d", cfg.SyncIntervalSec)
	}
	if cfg.ForecastInterval != 3600 {
		t.Fatalf("expected ForecastInterval 3600, got %d", cfg.ForecastInterval)
	}
}

func TestValidate_MissingAgentID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.RegionID = "region-1"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing AgentID")
	}
	cfgErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("expected *ConfigError, got %T", err)
	}
	if cfgErr.Field != "AgentID" {
		t.Fatalf("expected field AgentID, got %q", cfgErr.Field)
	}
}

func TestValidate_MissingRegionID(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AgentID = "agent-1"
	err := cfg.Validate()
	if err == nil {
		t.Fatal("expected error for missing RegionID")
	}
	cfgErr, ok := err.(*ConfigError)
	if !ok {
		t.Fatalf("expected *ConfigError, got %T", err)
	}
	if cfgErr.Field != "RegionID" {
		t.Fatalf("expected field RegionID, got %q", cfgErr.Field)
	}
}

func TestValidate_Success(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AgentID = "agent-1"
	cfg.RegionID = "region-1"
	err := cfg.Validate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestValidate_DefaultsRPCAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AgentID = "a"
	cfg.RegionID = "r"
	cfg.RPCAddress = ""
	cfg.Validate()
	if cfg.RPCAddress != ":12000" {
		t.Fatalf("expected :12000, got %q", cfg.RPCAddress)
	}
}

func TestValidate_DefaultsRaftAddress(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AgentID = "a"
	cfg.RegionID = "r"
	cfg.RaftAddress = ""
	cfg.Validate()
	if cfg.RaftAddress != ":12001" {
		t.Fatalf("expected :12001, got %q", cfg.RaftAddress)
	}
}

func TestValidate_PreservesExistingAddresses(t *testing.T) {
	cfg := DefaultConfig()
	cfg.AgentID = "a"
	cfg.RegionID = "r"
	cfg.RPCAddress = ":9999"
	cfg.RaftAddress = ":8888"
	cfg.Validate()
	if cfg.RPCAddress != ":9999" {
		t.Fatalf("expected :9999, got %q", cfg.RPCAddress)
	}
	if cfg.RaftAddress != ":8888" {
		t.Fatalf("expected :8888, got %q", cfg.RaftAddress)
	}
}

func TestConfigError_Error(t *testing.T) {
	e := &ConfigError{Field: "X", Message: "missing X"}
	if e.Error() != "missing X" {
		t.Fatalf("expected 'missing X', got %q", e.Error())
	}
}

func TestErrMissingAgentID(t *testing.T) {
	if ErrMissingAgentID.Field != "AgentID" {
		t.Fatal("wrong field")
	}
}

func TestErrMissingRegionID(t *testing.T) {
	if ErrMissingRegionID.Field != "RegionID" {
		t.Fatal("wrong field")
	}
}
