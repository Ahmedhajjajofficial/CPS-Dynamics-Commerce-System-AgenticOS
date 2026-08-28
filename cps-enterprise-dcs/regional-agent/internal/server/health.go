package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"go.uber.org/zap"
)

// HealthResponse represents the JSON health check response.
type HealthResponse struct {
	Status    string            `json:"status"`
	AgentID   string            `json:"agent_id"`
	RegionID  string            `json:"region_id"`
	State     string            `json:"agent_state"`
	RaftState string            `json:"raft_state"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// StartHealthServer starts a lightweight HTTP health server.
// It returns the server so the caller can shut it down gracefully.
func StartHealthServer(addr string, regionalAgent *agent.RegionalAgent, logger *zap.Logger) *http.Server {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		checks := make(map[string]string)
		status := "healthy"

		raftState := "unknown"
		if regionalAgent != nil && regionalAgent.Raft() != nil {
			raftState = regionalAgent.Raft().State().String()
			if raftState != "Leader" && raftState != "Follower" {
				status = "degraded"
				checks["raft"] = "not_leader_or_follower"
			} else {
				checks["raft"] = "ok"
			}
		}

		agentState := "unknown"
		if regionalAgent != nil {
			agentState = regionalAgent.GetState().String()
		}

		resp := HealthResponse{
			Status:    status,
			AgentID:   os.Getenv("DCS_AGENT_ID"),
			RegionID:  os.Getenv("DCS_REGION_ID"),
			State:     agentState,
			RaftState: raftState,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Checks:    checks,
		}

		w.Header().Set("Content-Type", "application/json")
		if status == "healthy" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}
		json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		if regionalAgent != nil && regionalAgent.GetState().String() == "ACTIVE" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ready"))
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
			w.Write([]byte("not_ready"))
		}
	})

	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Health server error", zap.Error(err))
		}
	}()

	logger.Info("Health server started", zap.String("addr", addr))
	return server
}

// StopHealthServer gracefully stops the health HTTP server.
func StopHealthServer(server *http.Server) {
	if server == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		// Log only if a real logger is available; ignore shutdown errors during process exit
		_ = err
	}
}
