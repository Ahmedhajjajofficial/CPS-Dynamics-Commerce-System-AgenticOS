package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/config"
	"github.com/cps-enterprise/dcs/regional-agent/internal/store"
	"go.uber.org/zap"
)

func TestHealthServer(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	regionalAgent, err := agent.New(cfg, logger)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	server := StartHealthServer("127.0.0.1:0", regionalAgent, logger, nil)
	defer StopHealthServer(server)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK && w.Code != http.StatusServiceUnavailable {
		t.Fatalf("unexpected status %d", w.Code)
	}

	var resp HealthResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode health response: %v", err)
	}
	if resp.Status != "healthy" && resp.Status != "degraded" {
		t.Fatalf("unexpected health status %s", resp.Status)
	}
}

func TestHealthServerReady(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	regionalAgent, err := agent.New(cfg, logger)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	server := StartHealthServer("127.0.0.1:0", regionalAgent, logger, nil)
	defer StopHealthServer(server)

	req := httptest.NewRequest(http.MethodGet, "/ready", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected service unavailable when state is not ACTIVE, got %d", w.Code)
	}
}

func TestBranchSummaryRequiresStore(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	regionalAgent, err := agent.New(cfg, logger)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	server := StartHealthServer("127.0.0.1:0", regionalAgent, logger, nil)
	defer StopHealthServer(server)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/branches/BR001", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected service unavailable when store is nil, got %d", w.Code)
	}
}

func TestInventoryRequiresStore(t *testing.T) {
	cfg := &config.Config{}
	logger, _ := zap.NewDevelopment()
	regionalAgent, err := agent.New(cfg, logger)
	if err != nil {
		t.Fatalf("agent.New() error = %v", err)
	}

	server := StartHealthServer("127.0.0.1:0", regionalAgent, logger, nil)
	defer StopHealthServer(server)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/inventory/P1", nil)
	w := httptest.NewRecorder()
	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected service unavailable when store is nil, got %d", w.Code)
	}
}
