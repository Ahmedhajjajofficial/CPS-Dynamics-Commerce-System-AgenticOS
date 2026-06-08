/*
Regional Agent - Core Implementation
====================================
The regional coordinator for branch agents.

Responsibilities:
1. Raft consensus for leader election
2. CRDT state aggregation from branches
3. Regional forecasting and analytics
4. Event propagation between branches
5. Master agent communication
*/

package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/cps-enterprise/dcs/regional-agent/internal/config"
	"github.com/cps-enterprise/dcs/regional-agent/internal/crdt"
	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/raft"
	raftboltdb "github.com/hashicorp/raft-boltdb/v2"
	"go.uber.org/zap"
)

// State represents the operational state of the regional agent
type State int

const (
	StateInitializing State = iota
	StateFollower
	StateCandidate
	StateLeader
	StateShutdown
)

func (s State) String() string {
	switch s {
	case StateInitializing:
		return "INITIALIZING"
	case StateFollower:
		return "FOLLOWER"
	case StateCandidate:
		return "CANDIDATE"
	case StateLeader:
		return "LEADER"
	case StateShutdown:
		return "SHUTDOWN"
	default:
		return "UNKNOWN"
	}
}

// RegionalAgent is the core regional coordinator
type RegionalAgent struct {
	config *config.Config
	logger *zap.Logger
	state  State
	stateMu sync.RWMutex

	// Raft consensus
	raft          *raft.Raft
	raftTransport *raft.NetworkTransport
	fsm           *FSM

	// CRDT management
	crdtManager *crdt.Manager

	// Connected branches
	branches   map[string]*BranchConnection
	branchesMu sync.RWMutex

	// Channels
	eventCh    chan *Event
	shutdownCh chan struct{}

	// Background tasks
	tasks sync.WaitGroup
}

// BranchConnection represents a connected branch agent
type BranchConnection struct {
	BranchID      string
	AgentID       string
	Address       string
	LastHeartbeat time.Time
	CRDTState     map[string]interface{}
	IsOnline      bool
}

// Event represents an event in the system
type Event struct {
	ID            string
	Type          string
	BranchID      string
	Payload       []byte
	Timestamp     time.Time
	CausalContext map[string]interface{}
}

// New creates a new regional agent
func New(cfg *config.Config, logger *zap.Logger) (*RegionalAgent, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	agent := &RegionalAgent{
		config:     cfg,
		logger:     logger,
		state:      StateInitializing,
		branches:   make(map[string]*BranchConnection),
		eventCh:    make(chan *Event, 10000),
		shutdownCh: make(chan struct{}),
	}

	return agent, nil
}

// Initialize sets up the agent and starts background tasks
func (a *RegionalAgent) Initialize(ctx context.Context) error {
	a.logger.Info("Initializing regional agent")

	// Initialize CRDT manager
	a.crdtManager = crdt.NewManager(a.config.AgentID)

	// Initialize Raft consensus
	if err := a.initRaft(); err != nil {
		return fmt.Errorf("failed to initialize raft: %w", err)
	}

	// Start background tasks
	a.tasks.Add(1)
	go a.eventProcessor()

	a.tasks.Add(1)
	go a.heartbeatMonitor()

	a.tasks.Add(1)
	go a.forecastEngine()

	a.setState(StateFollower)
	a.logger.Info("Regional agent initialized successfully")

	return nil
}

// Shutdown gracefully stops the agent
func (a *RegionalAgent) Shutdown(ctx context.Context) error {
	a.logger.Info("Shutting down regional agent")
	a.setState(StateShutdown)

	// Signal shutdown
	close(a.shutdownCh)

	// Shutdown Raft
	if a.raft != nil {
		if err := a.raft.Shutdown().Error(); err != nil {
			a.logger.Error("Error shutting down raft", zap.Error(err))
		}
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

// initRaft initializes the Raft consensus module
func (a *RegionalAgent) initRaft() error {
	// Create data directory
	raftDir := filepath.Join(a.config.DataDir, "raft")
	if err := os.MkdirAll(raftDir, 0755); err != nil {
		return fmt.Errorf("failed to create raft directory: %w", err)
	}

	// Create FSM
	a.fsm = NewFSM(a.logger)

	// Create Raft configuration
	raftConfig := raft.DefaultConfig()
	raftConfig.LocalID = raft.ServerID(a.config.AgentID)
	raftConfig.Logger = hclog.New(&hclog.LoggerOptions{
		Name:   "raft",
		Output: os.Stderr,
		Level:  hclog.Info,
	})

	// Create transport
	addr, err := net.ResolveTCPAddr("tcp", a.config.RaftAddress)
	if err != nil {
		return fmt.Errorf("failed to resolve raft address: %w", err)
	}

	transport, err := raft.NewTCPTransport(
		a.config.RaftAddress,
		addr,
		3,
		10*time.Second,
		os.Stderr,
	)
	if err != nil {
		return fmt.Errorf("failed to create raft transport: %w", err)
	}
	a.raftTransport = transport

	// Create log store and stable store
	logStore, err := raftboltdb.NewBoltStore(
		filepath.Join(raftDir, "raft.db"),
	)
	if err != nil {
		return fmt.Errorf("failed to create log store: %w", err)
	}

	// Create snapshot store
	snapStore, err := raft.NewFileSnapshotStore(
		raftDir,
		2,
		os.Stderr,
	)
	if err != nil {
		return fmt.Errorf("failed to create snapshot store: %w", err)
	}

	// Create Raft instance
	r, err := raft.NewRaft(
		raftConfig,
		a.fsm,
		logStore,
		logStore,
		snapStore,
		transport,
	)
	if err != nil {
		return fmt.Errorf("failed to create raft: %w", err)
	}
	a.raft = r

	// Bootstrap if needed
	if a.config.Bootstrap {
		configuration := raft.Configuration{
			Servers: []raft.Server{
				{
					ID:      raft.ServerID(a.config.AgentID),
					Address: transport.LocalAddr(),
				},
			},
		}
		if future := r.BootstrapCluster(configuration); future.Error() != nil {
			a.logger.Warn("Raft bootstrap returned error (may already be bootstrapped)",
				zap.Error(future.Error()),
			)
		} else {
			a.logger.Info("Bootstrapped Raft cluster")
		}
	}

	return nil
}

// GetID returns the agent's identifier
func (a *RegionalAgent) GetID() string {
	return a.config.AgentID
}

// IsLeader returns true if this agent is the Raft leader
func (a *RegionalAgent) IsLeader() bool {
	return a.raft.State() == raft.Leader
}

// GetState returns the current state of the agent
func (a *RegionalAgent) GetState() State {
	a.stateMu.RLock()
	defer a.stateMu.RUnlock()
	return a.state
}

func (a *RegionalAgent) setState(state State) {
	a.stateMu.Lock()
	defer a.stateMu.Unlock()
	oldState := a.state
	a.state = state
	if oldState != state {
		a.logger.Info("State changed",
			zap.String("from", oldState.String()),
			zap.String("to", state.String()),
		)
	}
}

// RegisterBranch registers a new branch connection
func (a *RegionalAgent) RegisterBranch(branchID, agentID, address string) error {
	a.branchesMu.Lock()
	defer a.branchesMu.Unlock()

	a.branches[branchID] = &BranchConnection{
		BranchID:      branchID,
		AgentID:       agentID,
		Address:       address,
		LastHeartbeat: time.Now(),
		IsOnline:      true,
	}

	a.logger.Info("Branch registered",
		zap.String("branch_id", branchID),
		zap.String("agent_id", agentID),
		zap.String("address", address),
	)

	return nil
}

// UnregisterBranch removes a branch connection
func (a *RegionalAgent) UnregisterBranch(branchID string) {
	a.branchesMu.Lock()
	defer a.branchesMu.Unlock()

	if branch, exists := a.branches[branchID]; exists {
		branch.IsOnline = false
		a.logger.Info("Branch unregistered", zap.String("branch_id", branchID))
	}
}

// GetBranch returns a branch connection
func (a *RegionalAgent) GetBranch(branchID string) (*BranchConnection, bool) {
	a.branchesMu.RLock()
	defer a.branchesMu.RUnlock()

	branch, exists := a.branches[branchID]
	return branch, exists
}

// GetAllBranches returns all registered branches
func (a *RegionalAgent) GetAllBranches() []*BranchConnection {
	a.branchesMu.RLock()
	defer a.branchesMu.RUnlock()

	branches := make([]*BranchConnection, 0, len(a.branches))
	for _, branch := range a.branches {
		branches = append(branches, branch)
	}
	return branches
}

// ProcessEvent processes an event from a branch
func (a *RegionalAgent) ProcessEvent(event *Event) error {
	select {
	case a.eventCh <- event:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("event queue full")
	}
}

// Background tasks

func (a *RegionalAgent) eventProcessor() {
	defer a.tasks.Done()

	a.logger.Info("Event processor started")

	for {
		select {
		case event := <-a.eventCh:
			a.handleEvent(event)

		case <-a.shutdownCh:
			a.logger.Info("Event processor stopping")
			return
		}
	}
}

func (a *RegionalAgent) handleEvent(event *Event) {
	// Apply event to FSM if leader
	if a.IsLeader() {
		// TODO: Apply to Raft log
	}

	// Update CRDT state
	// TODO: Merge CRDT state from event

	a.logger.Debug("Event processed",
		zap.String("event_id", event.ID),
		zap.String("event_type", event.Type),
		zap.String("branch_id", event.BranchID),
	)
}

func (a *RegionalAgent) heartbeatMonitor() {
	defer a.tasks.Done()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	a.logger.Info("Heartbeat monitor started")

	for {
		select {
		case <-ticker.C:
			a.checkHeartbeats()

		case <-a.shutdownCh:
			a.logger.Info("Heartbeat monitor stopping")
			return
		}
	}
}

func (a *RegionalAgent) checkHeartbeats() {
	a.branchesMu.Lock()
	defer a.branchesMu.Unlock()

	timeout := 2 * time.Minute
	now := time.Now()

	for branchID, branch := range a.branches {
		if branch.IsOnline && now.Sub(branch.LastHeartbeat) > timeout {
			branch.IsOnline = false
			a.logger.Warn("Branch heartbeat timeout",
				zap.String("branch_id", branchID),
				zap.Time("last_heartbeat", branch.LastHeartbeat),
			)
		}
	}
}

func (a *RegionalAgent) forecastEngine() {
	defer a.tasks.Done()

	ticker := time.NewTicker(time.Duration(a.config.ForecastInterval) * time.Second)
	defer ticker.Stop()

	a.logger.Info("Forecast engine started",
		zap.Int("interval_sec", a.config.ForecastInterval),
	)

	for {
		select {
		case <-ticker.C:
			a.runForecast()

		case <-a.shutdownCh:
			a.logger.Info("Forecast engine stopping")
			return
		}
	}
}

func (a *RegionalAgent) runForecast() {
	// TODO: Implement time-series forecasting
	a.logger.Debug("Running forecast")
}

// FSM implements the Raft finite state machine
type FSM struct {
	logger *zap.Logger
	data   map[string]interface{}
	mu     sync.RWMutex
}

// NewFSM creates a new FSM
func NewFSM(logger *zap.Logger) *FSM {
	return &FSM{
		logger: logger,
		data:   make(map[string]interface{}),
	}
}

// Apply applies a Raft log entry to the FSM
func (f *FSM) Apply(log *raft.Log) interface{} {
	f.mu.Lock()
	defer f.mu.Unlock()

	// TODO: Deserialize and apply command
	f.logger.Debug("Applying log entry", zap.Uint64("index", log.Index))

	return nil
}

// Snapshot returns a snapshot of the FSM
func (f *FSM) Snapshot() (raft.FSMSnapshot, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	data := make(map[string]interface{}, len(f.data))
	for k, v := range f.data {
		data[k] = v
	}

	buf, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal FSM snapshot: %w", err)
	}

	return &fsmSnapshot{data: buf}, nil
}

// Restore restores the FSM from a snapshot
func (f *FSM) Restore(rc io.ReadCloser) error {
	defer rc.Close()
	f.mu.Lock()
	defer f.mu.Unlock()

	var data map[string]interface{}
	if err := json.NewDecoder(rc).Decode(&data); err != nil {
		return fmt.Errorf("failed to decode FSM snapshot: %w", err)
	}
	f.data = data
	f.logger.Info("FSM restored from snapshot", zap.Int("keys", len(data)))
	return nil
}

// fsmSnapshot implements raft.FSMSnapshot
type fsmSnapshot struct {
	data []byte
}

func (s *fsmSnapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(s.data); err != nil {
		sink.Cancel()
		return fmt.Errorf("failed to write snapshot: %w", err)
	}
	return sink.Close()
}

func (s *fsmSnapshot) Release() {}
