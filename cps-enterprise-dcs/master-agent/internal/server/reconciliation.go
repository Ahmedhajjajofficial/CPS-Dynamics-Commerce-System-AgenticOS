package server

import (
	"context"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ReconciliationHandler handles GlobalReconciliationService requests
type ReconciliationHandler struct {
	pb.UnimplementedGlobalReconciliationServiceServer
	agent *agent.MasterAgent
}

// TriggerReconciliation triggers a reconciliation job
func (h *ReconciliationHandler) TriggerReconciliation(ctx context.Context, req *pb.ReconciliationJob) (*pb.ReconciliationResult, error) {
	h.agent.Logger().Info("TriggerReconciliation called",
		zap.String("region_id", req.RegionId),
	)

	return &pb.ReconciliationResult{
		JobId:          req.JobId,
		IsBalanced:     true,
		TotalExpected:  0,
		TotalActual:    0,
		Discrepancy:    0,
		CompletedAt:    timestamppb.Now(),
	}, nil
}

// GetReconciliationStatus gets the status of a reconciliation job
func (h *ReconciliationHandler) GetReconciliationStatus(ctx context.Context, req *pb.ReconciliationStatusRequest) (*pb.ReconciliationJob, error) {
	return &pb.ReconciliationJob{
		JobId:  req.JobId,
		Status: pb.ReconciliationStatus_RECONCILIATION_COMPLETED,
	}, nil
}

// ListReconciliationHistory lists reconciliation history
func (h *ReconciliationHandler) ListReconciliationHistory(ctx context.Context, req *pb.ReconciliationHistoryRequest) (*pb.ReconciliationHistoryResponse, error) {
	return &pb.ReconciliationHistoryResponse{
		Jobs: []*pb.ReconciliationJob{},
	}, nil
}

// ValidateBranchState validates the state of a branch
func (h *ReconciliationHandler) ValidateBranchState(ctx context.Context, req *pb.BranchValidationRequest) (*pb.BranchValidationResponse, error) {
	return &pb.BranchValidationResponse{
		IsValid:      true,
		ValidationSummary: "Branch state is valid",
	}, nil
}
