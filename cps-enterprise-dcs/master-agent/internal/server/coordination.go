package server

import (
	"context"
	"time"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// CoordinationHandler handles MasterCoordinationService requests
type CoordinationHandler struct {
	pb.UnimplementedMasterCoordinationServiceServer
	agent *agent.MasterAgent
}

// RegisterRegionalAgent registers a regional agent
func (h *CoordinationHandler) RegisterRegionalAgent(ctx context.Context, req *pb.RegionalAgentRegistration) (*pb.RegistrationResponse, error) {
	return h.agent.RegisterRegionalAgent(req)
}

// Heartbeat handles heartbeat from regional agents
func (h *CoordinationHandler) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	h.agent.Logger().Debug("Heartbeat received",
		zap.String("agent_id", req.AgentId),
		zap.String("region_id", req.RegionId),
	)

	return &pb.HeartbeatResponse{
		Acknowledged:           true,
		NextHeartbeatDeadline:  timestamppb.New(time.Now().Add(90 * time.Second)),
	}, nil
}

// GetGlobalState gets the global state
func (h *CoordinationHandler) GetGlobalState(ctx context.Context, req *pb.GlobalStateRequest) (*pb.GlobalStateResponse, error) {
	return &pb.GlobalStateResponse{
		MasterAgentId: h.agent.Config().AgentID,
	}, nil
}

// PropagateDecision propagates a decision to regional agents
func (h *CoordinationHandler) PropagateDecision(ctx context.Context, req *pb.MasterDecision) (*pb.PropagationResult, error) {
	return &pb.PropagationResult{
		DecisionId:     req.DecisionId,
		Acknowledged:   true,
		AcknowledgedAt: timestamppb.Now(),
	}, nil
}
