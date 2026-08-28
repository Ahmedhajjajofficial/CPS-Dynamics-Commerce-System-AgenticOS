package server

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/store"
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

// BranchSummaryResponse represents the JSON branch summary response.
type BranchSummaryResponse struct {
	BranchID         string  `json:"branch_id"`
	TodaySales       float64 `json:"today_sales"`
	TodayTransactions int32   `json:"today_transactions"`
	CurrentBalance   float64 `json:"current_balance"`
	ActiveSessions   int32   `json:"active_sessions"`
}

// InventoryStatusResponse represents the JSON inventory status response.
type InventoryStatusResponse struct {
	ProductID         string `json:"product_id"`
	BranchID          string `json:"branch_id"`
	CurrentQuantity   int32  `json:"current_quantity"`
	AvailableQuantity int32  `json:"available_quantity"`
	IsLowStock        bool   `json:"is_low_stock"`
}

// StartHealthServer starts a lightweight HTTP server for health and admin query endpoints.
// It returns the server so the caller can shut it down gracefully.
func StartHealthServer(addr string, regionalAgent *agent.RegionalAgent, logger *zap.Logger, store *store.Store) *http.Server {
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

	mux.HandleFunc("/api/v1/branches/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		branchID := r.URL.Path[len("/api/v1/branches/"):]
		if branchID == "" {
			http.Error(w, "branch_id is required", http.StatusBadRequest)
			return
		}

		if store == nil {
			http.Error(w, "store not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		todaySales, todayTransactions, currentBalance, activeSessions, err := store.GetBranchSummary(ctx, branchID)
		if err != nil {
			logger.Error("failed to load branch summary", zap.Error(err), zap.String("branch_id", branchID))
			http.Error(w, "failed to load branch summary", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(BranchSummaryResponse{
			BranchID:          branchID,
			TodaySales:        todaySales,
			TodayTransactions: todayTransactions,
			CurrentBalance:    currentBalance,
			ActiveSessions:    activeSessions,
		})
	})

	mux.HandleFunc("/api/v1/inventory/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		productID := r.URL.Path[len("/api/v1/inventory/"):]
		if productID == "" {
			http.Error(w, "product_id is required", http.StatusBadRequest)
			return
		}

		branchID := r.URL.Query().Get("branch_id")
		if branchID == "" {
			branchID = "default"
		}

		if store == nil {
			http.Error(w, "store not configured", http.StatusServiceUnavailable)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		currentQty, availableQty, isLowStock, err := store.GetInventoryStatus(ctx, branchID, productID)
		if err != nil {
			logger.Error("failed to load inventory status", zap.Error(err), zap.String("product_id", productID))
			http.Error(w, "failed to load inventory status", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(InventoryStatusResponse{
			ProductID:         productID,
			BranchID:          branchID,
			CurrentQuantity:   currentQty,
			AvailableQuantity: availableQty,
			IsLowStock:        isLowStock,
		})
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

	logger.Info("Health/admin server started", zap.String("addr", addr))
	return server
}
