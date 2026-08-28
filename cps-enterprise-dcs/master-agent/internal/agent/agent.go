package agent

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/cps-enterprise/dcs/master-agent/internal/config"
	pb "github.com/cps-enterprise/dcs/master-agent/internal/proto"
	"github.com/cps-enterprise/dcs/proto"
	"github.com/hashicorp/go-hclog"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// MasterAgent represents the global brain of the system
type MasterAgent struct {
	config        *config.Config
	logger        *zap.Logger
	state         AgentState
	stateMu       sync.RWMutex
	shutdownCh    chan struct{}
	tasks         sync.WaitGroup

	// Regional agents registry
	regions   map[string]*RegionalConnection
	regionsMu sync.RWMutex

	// Kafka producer for event streaming
	kafkaProducer KafkaProducer

	// PostgreSQL store for global state
	store *Store

	// Decision engine
	decisionEngine *DecisionEngine
}

// RegionalConnection represents a connected regional agent
type RegionalConnection struct {
	AgentID       string
	RegionID      string
	RPCAddress    string
	RaftAddress   string
	Status        proto.AgentStatus
	LastHeartbeat time.Time
	ActiveBranches int64
	PendingEvents  int64
	Capabilities  []string
}

// AgentState represents the master agent state
type AgentState int32

const (
	AgentStateInitializing AgentState = iota
	AgentStateActive
	AgentStateDegraded
	AgentStateOffline
	AgentStateShutdown
)

func (s AgentState) String() string {
	switch s {
	case AgentStateInitializing:
		return "INITIALIZING"
	case AgentStateActive:
		return "ACTIVE"
	case AgentStateDegraded:
		return "DEGRADED"
	case AgentStateOffline:
		return "OFFLINE"
	case AgentStateShutdown:
		return "SHUTDOWN"
	default:
		return "UNKNOWN"
	}
}

// New creates a new Master Agent
func New(cfg *config.Config, logger *zap.Logger) (*MasterAgent, error) {
	if logger == nil {
		hclogger := hclog.New(&hclog.LoggerOptions{Name: "master-agent", Level: hclog.Info})
		logger = zap.New(hclogger)
	}

	agent := &MasterAgent{
		config:     cfg,
		logger:     logger,
		state:      AgentStateInitializing,
		shutdownCh: make(chan struct{}),
		regions:    make(map[string]*RegionalConnection),
	}

	return agent, nil
}

// Initialize sets up the master agent
func (a *MasterAgent) Initialize(ctx context.Context) error {
	a.logger.Info("Initializing Master Agent")

	// Initialize Kafka producer
	if a.config.KafkaBrokers != "" {
		producer, err := NewKafkaProducer(a.config.KafkaBrokers, a.logger)
		if err != nil {
			return fmt.Errorf("failed to create Kafka producer: %w", err)
		}
		a.kafkaProducer = producer
	}

	// Initialize PostgreSQL store
	if a.config.PostgreSQLURL != "" {
		store, err := NewStore(ctx, a.config, a.logger)
		if err != nil {
			return fmt.Errorf("failed to connect to PostgreSQL: %w", err)
		}
		a.store = store
	}

	// Initialize decision engine
	a.decisionEngine = NewDecisionEngine(a.store, a.logger)

	// Start background tasks
	a.tasks.Add(1)
	go a.heartbeatMonitor()

	a.tasks.Add(1)
	go a.reconciliationScheduler()

	a.setState(AgentStateActive)
	a.logger.Info("Master Agent initialized successfully")

	return nil
}

// Shutdown gracefully stops the master agent
func (a *MasterAgent) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down Master Agent")
	a.setState(AgentStateShutdown)

	// Signal shutdown
	close(a.shutdownCh)

	// Close Kafka producer
	if a.kafkaProducer != nil {
		a.kafkaProducer.Close()
	}

	// Close PostgreSQL store
	if a.store != nil {
		a.store.Close()
	}

	// Wait for background tasks
	done := make(chan struct{})
	go func() {
		a.tasks.Wait()
		close(done)
	}()

	select {
	case <-done:
		a.logger.Info("All background tasks completed")
	case <-ctx.Done():
		a.logger.Warn("Shutdown timeout, some tasks may not have completed")
	}

	return nil
}

// GetState returns the current state of the agent
func (a *MasterAgent) GetState() AgentState {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

// Logger returns the logger instance
func (a *MasterAgent) Logger() *zap.Logger {
	return a.logger
}

// Config returns the agent configuration
func (a *MasterAgent) Config() *config.Config {
	return a.config
}

func (a *MasterAgent) setState(state AgentState) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	a.state = state
}

// RegisterRegionalAgent registers a regional agent with the master
func (a *MasterAgent) RegisterRegionalAgent(req *pb.RegionalAgentRegistration) (*pb.RegistrationResponse, error) {
	a.regionsMu.Lock()
	defer a.regionsMu.Unlock()

	conn := &RegionalConnection{
		AgentID:       req.AgentId,
		RegionID:      req.RegionId,
		RPCAddress:    req.RpcAddress,
		RaftAddress:   req.RaftAddress,
		Status:        proto.AgentStatus_AGENT_ACTIVE,
		LastHeartbeat: time.Now(),
		Capabilities:  req.Capabilities,
	}

	a.regions[req.RegionId] = conn

	a.logger.Info("Regional agent registered",
		zap.String("agent_id", req.AgentId),
		zap.String("region_id", req.RegionId),
	)

	sessionToken := generateSessionToken()

	return &pb.RegistrationResponse{
		Success:       true,
		MasterAgentId: a.config.AgentID,
		RegisteredAt:  timestamppb.Now(),
		SessionToken:  sessionToken,
	}, nil
}

// ProcessRegionalEvent processes an event from a regional agent
func (a *MasterAgent) ProcessRegionalEvent(event *pb.RegionalEventPublication) (*pb.PublicationResult, error) {
	if a.kafkaProducer == nil {
		return &pb.PublicationResult{
			Success:      false,
			ErrorMessage: "kafka producer not configured",
		}, fmt.Errorf("kafka producer not configured")
	}

	topic := fmt.Sprintf("regional.events.%s", event.RegionId)
	err := a.kafkaProducer.Send(topic, event.EventId, event.EventPayload)
	if err != nil {
		a.logger.Error("failed to publish event to kafka",
			zap.Error(err),
			zap.String("event_id", event.EventId),
		)
		return &pb.PublicationResult{
			Success:        false,
			ErrorMessage:  err.Error(),
		}, err
	}

	a.logger.Debug("event published to kafka",
		zap.String("topic", topic),
		zap.String("event_id", event.EventId),
	)

	return &pb.PublicationResult{
		Success: true,
		Topic:   topic,
	}, nil
}

// heartbeatMonitor monitors regional agent heartbeats
func (a *MasterAgent) heartbeatMonitor() {
	defer a.tasks.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.checkRegionHealth()
		case <-a.shutdownCh:
			return
		}
	}
}

func (a *MasterAgent) checkRegionHealth() {
	a.regionsMu.RLock()
	defer a.regionsMu.RUnlock()

	now := time.Now()
	for regionID, conn := range a.regions {
		if now.Sub(conn.LastHeartbeat) > 90*time.Second {
			a.logger.Warn("regional agent heartbeat missed",
				zap.String("region_id", regionID),
				zap.String("agent_id", conn.AgentID),
			)
			conn.Status = proto.AgentStatus_AGENT_DEGRADED
		}
	}
}

// reconciliationScheduler schedules periodic reconciliation jobs
func (a *MasterAgent) reconciliationScheduler() {
	defer a.tasks.Done()

	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			a.triggerDailyReconciliation()
		case <-a.shutdownCh:
			return
		}
	}
}

func (a *MasterAgent) triggerDailyReconciliation() {
	a.logger.Info("triggering daily reconciliation")

	a.regionsMu.RLock()
	regionIDs := make([]string, 0, len(a.regions))
	for regionID := range a.regions {
		regionIDs = append(regionIDs, regionID)
	}
	a.regionsMu.RUnlock()

	for _, regionID := range regionIDs {
		job := &pb.ReconciliationJob{
			RegionId:  regionID,
			Type:      pb.ReconciliationType_RECONCILIATION_DAILY,
			Status:    pb.ReconciliationStatus_RECONCILIATION_PENDING,
			CreatedAt: timestamppb.Now(),
		}

		go func(job *pb.ReconciliationJob) {
			if err := a.executeReconciliation(job); err != nil {
				a.logger.Error("reconciliation failed",
					zap.Error(err),
					zap.String("region_id", job.RegionId),
				)
			}
		}(job)
	}
}

func (a *MasterAgent) executeReconciliation(job *pb.ReconciliationJob) error {
	a.logger.Info("executing reconciliation",
		zap.String("region_id", job.RegionId),
		zap.String("job_id", job.JobId),
	)

	// TODO: Implement actual reconciliation logic
	// 1. Fetch events from regional agent
	// 2. Compare with master state
	// 3. Resolve discrepancies
	// 4. Update master state

	job.Status = pb.ReconciliationStatus_RECONCILIATION_COMPLETED
	now := timestamppb.Now()
	job.CompletedAt = now

	return nil
}

func generateSessionToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
