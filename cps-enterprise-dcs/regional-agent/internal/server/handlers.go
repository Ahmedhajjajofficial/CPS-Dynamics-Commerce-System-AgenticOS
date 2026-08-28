package server

import (
	"context"
	"strings"
	"time"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/store"
	pb "github.com/cps-enterprise/dcs/regional-agent/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SwarmHandler handles AccountingSwarmProtocol requests
type SwarmHandler struct {
	pb.UnimplementedAccountingSwarmProtocolServer
	server *Server
}

// BroadcastFinancialEvent handles single event broadcast
func (h *SwarmHandler) BroadcastFinancialEvent(ctx context.Context, req *pb.SovereignFinancialEvent) (*pb.AckResponse, error) {
	if h.server.agent.Store() != nil {
		h.server.agent.Store().InsertAuditLog(ctx, "broadcast_financial_event", req.AgentId, "event", req.EventId, map[string]interface{}{
			"branch_id": req.BranchId,
			"event_type": req.Type.String(),
		})
	}

	return &pb.AckResponse{
		Success:        true,
		Message:        "Event received at Regional Agent",
		ReceiptHash:    req.EventHash,
		ProcessingNode: h.server.agent.GetID(),
	}, nil
}

// RequestReconciliation handles reconciliation requests
func (h *SwarmHandler) RequestReconciliation(ctx context.Context, req *pb.ReconciliationRequest) (*pb.ReconciliationResponse, error) {
	if h.server.agent.Store() != nil {
		h.server.agent.Store().InsertAuditLog(ctx, "request_reconciliation", req.RequestingAgentId, "reconciliation", req.BranchId, map[string]interface{}{
			"start_timestamp": req.StartTimestamp,
			"end_timestamp": req.EndTimestamp,
		})
	}

	return &pb.ReconciliationResponse{
		IsBalanced: true,
		ActualBalance: req.ExpectedBalance,
		ReconciliationTimestamp: &pb.HybridLogicalClock{
			PhysicalMs: time.Now().UnixMilli(),
		},
	}, nil
}

// StreamCRDTState handles bidirectional CRDT state synchronization.
func (h *SwarmHandler) StreamCRDTState(stream pb.AccountingSwarmProtocol_StreamCRDTStateServer) error {
	ctx := stream.Context()
	agentID := ""
	branchID := ""

	if md, ok := metadata.FromIncomingContext(ctx); ok {
		agentID = strings.Join(md.Get("x-agent-id"), "")
		branchID = strings.Join(md.Get("x-branch-id"), "")
	}

	if agentID == "" || branchID == "" {
		return status.Error(codes.InvalidArgument, "missing x-agent-id or x-branch-id metadata")
	}

	var bundles []*pb.CRDTStateBundle
	for {
		bundle, err := stream.Recv()
		if err != nil {
			return err
		}

		bundles = append(bundles, bundle)

		if h.server.agent.Store() != nil {
			h.server.agent.Store().InsertAuditLog(ctx, "crdt_state_received", agentID, "crdt", branchID, map[string]interface{}{
				"bundle_count": len(bundles),
			})
		}
	}

	if err := h.server.agent.MergeCRDTState(ctx, branchID, bundles); err != nil {
		return status.Errorf(codes.Internal, "failed to merge CRDT state: %v", err)
	}

	return stream.Send(&pb.BatchAckResponse{
		TotalProcessed: int32(len(bundles)),
		TotalFailed:    0,
	})
}

// QueryHandler handles QueryProtocol requests
type QueryHandler struct {
	pb.UnimplementedQueryProtocolServer
	server *Server
}

// GetBranchSummary returns summary for a branch
func (h *QueryHandler) GetBranchSummary(ctx context.Context, req *pb.BranchQuery) (*pb.BranchSummary, error) {
	if h.server.agent.Store() == nil {
		return &pb.BranchSummary{
			BranchId: req.BranchId,
		}, nil
	}

	todaySales, todayTransactions, currentBalance, activeSessions, err := h.server.agent.Store().GetBranchSummary(ctx, req.BranchId)
	if err != nil {
		return nil, err
	}

	return &pb.BranchSummary{
		BranchId: req.BranchId,
		TodaySales: todaySales,
		TodayTransactions: todayTransactions,
		CurrentBalance: currentBalance,
		ActiveSessions: activeSessions,
	}, nil
}

// GetInventoryStatus returns inventory status
func (h *QueryHandler) GetInventoryStatus(ctx context.Context, req *pb.InventoryQuery) (*pb.InventoryStatus, error) {
	if h.server.agent.Store() == nil {
		return &pb.InventoryStatus{
			ProductId: req.ProductId,
			BranchId:  req.BranchId,
		}, nil
	}

	currentQty, availableQty, isLowStock, err := h.server.agent.Store().GetInventoryStatus(ctx, req.BranchId, req.ProductId)
	if err != nil {
		return nil, err
	}

	return &pb.InventoryStatus{
		ProductId: req.ProductId,
		BranchId:  req.BranchId,
		CurrentQuantity: currentQty,
		AvailableQuantity: availableQty,
		IsLowStock: isLowStock,
	}, nil
}
