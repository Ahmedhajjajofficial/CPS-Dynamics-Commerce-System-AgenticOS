package server

import (
	"context"
	"strings"

	"github.com/cps-enterprise/dcs/regional-agent/internal/agent"
	"github.com/cps-enterprise/dcs/regional-agent/internal/store"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor validates agent identity on every request.
// It expects the following gRPC metadata headers:
//   - x-agent-id: the caller agent ID
//   - x-branch-id: the caller branch ID
//   - x-region-id: the caller region ID
type AuthInterceptor struct {
	agent   *agent.RegionalAgent
	store   *store.Store
	logger  *zap.Logger
	allowed map[string]struct{}
}

// NewAuthInterceptor creates a new AuthInterceptor.
// In production, allowed MUST be non-empty; otherwise all requests are rejected.
func NewAuthInterceptor(a *agent.RegionalAgent, store *store.Store, logger *zap.Logger, allowed map[string]struct{}) *AuthInterceptor {
	if allowed == nil {
		allowed = make(map[string]struct{})
	}
	return &AuthInterceptor{
		agent:   a,
		store:   store,
		logger:  logger,
		allowed: allowed,
	}
}

// UnaryServerInterceptor returns a grpc.UnaryServerInterceptor.
func (a *AuthInterceptor) UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		agentID := strings.Join(md.Get("x-agent-id"), "")
		branchID := strings.Join(md.Get("x-branch-id"), "")
		regionID := strings.Join(md.Get("x-region-id"), "")

		if agentID == "" || branchID == "" || regionID == "" {
			return nil, status.Error(codes.InvalidArgument, "missing required metadata: x-agent-id, x-branch-id, x-region-id")
		}

		if len(a.allowed) > 0 {
			if _, ok := a.allowed[agentID]; !ok {
				return nil, status.Error(codes.PermissionDenied, "agent not authorized")
			}
		} else {
			return nil, status.Error(codes.PermissionDenied, "no authorized agents configured")
		}

		if a.store != nil {
			a.store.InsertAuditLog(ctx, "grpc_request", agentID, info.FullMethod, branchID, map[string]interface{}{
				"region_id": regionID,
			})
		}

		return handler(ctx, req)
	}
}

// StreamServerInterceptor returns a grpc.StreamServerInterceptor.
func (a *AuthInterceptor) StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		md, ok := metadata.FromIncomingContext(ss.Context())
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		agentID := strings.Join(md.Get("x-agent-id"), "")
		branchID := strings.Join(md.Get("x-branch-id"), "")
		regionID := strings.Join(md.Get("x-region-id"), "")

		if agentID == "" || branchID == "" || regionID == "" {
			return status.Error(codes.InvalidArgument, "missing required metadata: x-agent-id, x-branch-id, x-region-id")
		}

		if len(a.allowed) > 0 {
			if _, ok := a.allowed[agentID]; !ok {
				return status.Error(codes.PermissionDenied, "agent not authorized")
			}
		} else {
			return status.Error(codes.PermissionDenied, "no authorized agents configured")
		}

		if a.store != nil {
			a.store.InsertAuditLog(ss.Context(), "grpc_stream_request", agentID, info.FullMethod, branchID, map[string]interface{}{
				"region_id": regionID,
			})
		}

		return handler(srv, ss)
	}
}
