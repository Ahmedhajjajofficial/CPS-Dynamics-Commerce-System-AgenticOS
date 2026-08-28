package server

import (
	"context"

	"github.com/cps-enterprise/dcs/master-agent/internal/agent"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"go.uber.org/zap"
)

// EventStreamingHandler handles EventStreamingService requests
type EventStreamingHandler struct {
	pb.UnimplementedEventStreamingServiceServer
	agent *agent.MasterAgent
}

// PublishRegionalEvent publishes a regional event to Kafka
func (h *EventStreamingHandler) PublishRegionalEvent(ctx context.Context, req *pb.RegionalEventPublication) (*pb.PublicationResult, error) {
	return h.agent.ProcessRegionalEvent(req)
}

// SubscribeToMasterEvents subscribes to master events
func (h *EventStreamingHandler) SubscribeToMasterEvents(req *pb.SubscriptionRequest, stream pb.EventStreamingService_SubscribeToMasterEventsServer) error {
	h.agent.Logger().Info("SubscribeToMasterEvents called")
	return nil
}

// GetEventStats gets event statistics
func (h *EventStreamingHandler) GetEventStats(ctx context.Context, req *pb.EventStatsRequest) (*pb.EventStatsResponse, error) {
	return &pb.EventStatsResponse{}, nil
}
