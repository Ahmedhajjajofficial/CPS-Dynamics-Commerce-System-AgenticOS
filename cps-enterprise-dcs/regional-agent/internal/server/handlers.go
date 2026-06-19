package server

import (
	"context"
	"time"

	pb "github.com/cps-enterprise/dcs/regional-agent/internal/proto"
)

// SwarmHandler handles AccountingSwarmProtocol requests
type SwarmHandler struct {
	pb.UnimplementedAccountingSwarmProtocolServer
	server *Server
}

// BroadcastFinancialEvent handles single event broadcast
func (h *SwarmHandler) BroadcastFinancialEvent(ctx context.Context, req *pb.SovereignFinancialEvent) (*pb.AckResponse, error) {
	// In a real implementation, we would pass this to the agent for Raft replication
	// For now, we return a success ack
	return &pb.AckResponse{
		Success:        true,
		Message:        "Event received at Regional Agent",
		ReceiptHash:    req.EventHash,
		ProcessingNode: h.server.agent.GetID(),
	}, nil
}

// RequestReconciliation handles reconciliation requests
func (h *SwarmHandler) RequestReconciliation(ctx context.Context, req *pb.ReconciliationRequest) (*pb.ReconciliationResponse, error) {
	return &pb.ReconciliationResponse{
		IsBalanced: true,
		ActualBalance: req.ExpectedBalance,
		ReconciliationTimestamp: &pb.HybridLogicalClock{
			PhysicalMs: time.Now().UnixMilli(),
		},
	}, nil
}

// QueryHandler handles QueryProtocol requests
type QueryHandler struct {
	pb.UnimplementedQueryProtocolServer
	server *Server
}

// GetBranchSummary returns summary for a branch
func (h *QueryHandler) GetBranchSummary(ctx context.Context, req *pb.BranchQuery) (*pb.BranchSummary, error) {
	return &pb.BranchSummary{
		BranchId: req.BranchId,
		TodaySales: 15000.50, // Mock data
		TodayTransactions: 120,
		CurrentBalance: 4500.00,
		ActiveSessions: 3,
	}, nil
}

// GetInventoryStatus returns inventory status
func (h *QueryHandler) GetInventoryStatus(ctx context.Context, req *pb.InventoryQuery) (*pb.InventoryStatus, error) {
	return &pb.InventoryStatus{
		ProductId: req.ProductId,
		BranchId:  req.BranchId,
		CurrentQuantity: 500,
		AvailableQuantity: 480,
		IsLowStock: false,
	}, nil
}
