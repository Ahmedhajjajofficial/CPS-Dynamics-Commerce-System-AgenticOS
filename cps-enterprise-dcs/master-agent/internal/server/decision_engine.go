package server

import (
	"context"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// DecisionEngineHandler handles DecisionEngineService requests
type DecisionEngineHandler struct {
	pb.UnimplementedDecisionEngineServiceServer
	agent *agent.MasterAgent
}

// SubmitDecision submits a decision for evaluation
func (h *DecisionEngineHandler) SubmitDecision(ctx context.Context, req *pb.DecisionRequest) (*pb.DecisionResponse, error) {
	h.agent.Logger().Info("SubmitDecision called",
		zap.String("agent_id", req.AgentId),
		zap.String("decision_type", req.Type.String()),
	)

	return &pb.DecisionResponse{
		DecisionId:    req.DecisionId,
		Approved:      true,
		Decision:      "approved",
		Reasoning:     "Decision approved by master agent",
		ConfidenceScore: 0.95,
		DecidedAt:     timestamppb.Now(),
		DecidedBy:     h.agent.Config().AgentID,
	}, nil
}

// GetDecision gets a decision by ID
func (h *DecisionEngineHandler) GetDecision(ctx context.Context, req *pb.DecisionRequest) (*pb.AgentDecision, error) {
	return &pb.AgentDecision{}, nil
}

// ListDecisions lists decisions
func (h *DecisionEngineHandler) ListDecisions(ctx context.Context, req *pb.DecisionHistoryRequest) (*pb.DecisionHistoryResponse, error) {
	return &pb.DecisionHistoryResponse{}, nil
}

// EvaluateForecast evaluates a forecast model
func (h *DecisionEngineHandler) EvaluateForecast(ctx context.Context, req *pb.ForecastEvaluationRequest) (*pb.ForecastEvaluationResponse, error) {
	return &pb.ForecastEvaluationResponse{}, nil
}
