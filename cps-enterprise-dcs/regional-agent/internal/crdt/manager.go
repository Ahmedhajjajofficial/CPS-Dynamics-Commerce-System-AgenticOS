/*
CRDT Manager - Regional Coordination
====================================
Manages CRDT state across multiple branches.

Responsibilities:
- Aggregate CRDT state from all branches
- Merge conflicting states
- Distribute updated state back to branches
- Maintain regional totals
*/

package crdt

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Manager manages CRDT state for the region
type Manager struct {
	nodeID string

	// CRDT storage
	counters   map[string]*PNCounter
	sets       map[string]*ORSet
	registers  map[string]*LWWRegister
	countersMu sync.RWMutex
	setsMu     sync.RWMutex
	registersMu sync.RWMutex
}

// NewManager creates a new CRDT manager
func NewManager(nodeID string) *Manager {
	return &Manager{
		nodeID:    nodeID,
		counters:  make(map[string]*PNCounter),
		sets:      make(map[string]*ORSet),
		registers: make(map[string]*LWWRegister),
	}
}

// GetOrCreateCounter gets or creates a PNCounter
func (m *Manager) GetOrCreateCounter(id string) *PNCounter {
	m.countersMu.Lock()
	defer m.countersMu.Unlock()

	if counter, exists := m.counters[id]; exists {
		return counter
	}

	counter := NewPNCounter(m.nodeID)
	m.counters[id] = counter
	return counter
}

// GetCounter gets a counter by ID
func (m *Manager) GetCounter(id string) (*PNCounter, bool) {
	m.countersMu.RLock()
	defer m.countersMu.RUnlock()

	counter, exists := m.counters[id]
	return counter, exists
}

// MergeCounter merges a counter from another node
func (m *Manager) MergeCounter(id string, other *PNCounter) {
	m.countersMu.Lock()
	defer m.countersMu.Unlock()

	if counter, exists := m.counters[id]; exists {
		counter.Merge(other)
	} else {
		m.counters[id] = other
	}
}

// GetOrCreateSet gets or creates an ORSet
func (m *Manager) GetOrCreateSet(id string) *ORSet {
	m.setsMu.Lock()
	defer m.setsMu.Unlock()

	if set, exists := m.sets[id]; exists {
		return set
	}

	set := NewORSet(m.nodeID)
	m.sets[id] = set
	return set
}

// MergeSet merges a set from another node
func (m *Manager) MergeSet(id string, other *ORSet) {
	m.setsMu.Lock()
	defer m.setsMu.Unlock()

	if set, exists := m.sets[id]; exists {
		set.Merge(other)
	} else {
		m.sets[id] = other
	}
}

// GetOrCreateRegister gets or creates an LWWRegister
func (m *Manager) GetOrCreateRegister(id string) *LWWRegister {
	m.registersMu.Lock()
	defer m.registersMu.Unlock()

	if reg, exists := m.registers[id]; exists {
		return reg
	}

	reg := NewLWWRegister(m.nodeID)
	m.registers[id] = reg
	return reg
}

// MergeRegister merges a register from another node
func (m *Manager) MergeRegister(id string, other *LWWRegister) {
	m.registersMu.Lock()
	defer m.registersMu.Unlock()

	if reg, exists := m.registers[id]; exists {
		reg.Merge(other)
	} else {
		m.registers[id] = other
	}
}

// GetAllCounters returns all counters
func (m *Manager) GetAllCounters() map[string]*PNCounter {
	m.countersMu.RLock()
	defer m.countersMu.RUnlock()

	result := make(map[string]*PNCounter, len(m.counters))
	for k, v := range m.counters {
		result[k] = v
	}
	return result
}

// SerializeState serializes all CRDT state
func (m *Manager) SerializeState() ([]byte, error) {
	state := map[string]interface{}{
		"node_id":   m.nodeID,
		"counters":  make(map[string]interface{}),
		"sets":      make(map[string]interface{}),
		"registers": make(map[string]interface{}),
	}

	m.countersMu.RLock()
	for id, counter := range m.counters {
		state["counters"].(map[string]interface{})[id] = counter.ToMap()
	}
	m.countersMu.RUnlock()

	m.setsMu.RLock()
	for id, set := range m.sets {
		state["sets"].(map[string]interface{})[id] = set.ToMap()
	}
	m.setsMu.RUnlock()

	m.registersMu.RLock()
	for id, reg := range m.registers {
		state["registers"].(map[string]interface{})[id] = reg.ToMap()
	}
	m.registersMu.RUnlock()

	return json.Marshal(state)
}

// PNCounter is a Positive-Negative Counter CRDT
type PNCounter struct {
	NodeID     string
	Increments map[string]int64
	Decrements map[string]int64
	mu         sync.RWMutex
}

// NewPNCounter creates a new PNCounter
func NewPNCounter(nodeID string) *PNCounter {
	return &PNCounter{
		NodeID:     nodeID,
		Increments: make(map[string]int64),
		Decrements: make(map[string]int64),
	}
}

// Increment adds to the positive counter
func (c *PNCounter) Increment(amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Increments[c.NodeID] += amount
}

// Decrement adds to the negative counter
func (c *PNCounter) Decrement(amount int64) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Decrements[c.NodeID] += amount
}

// Value returns the net value
func (c *PNCounter) Value() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var pos, neg int64
	for _, v := range c.Increments {
		pos += v
	}
	for _, v := range c.Decrements {
		neg += v
	}
	return pos - neg
}

// Merge merges another PNCounter
func (c *PNCounter) Merge(other *PNCounter) {
	c.mu.Lock()
	defer c.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	// Merge increments
	for node, val := range other.Increments {
		if current, exists := c.Increments[node]; !exists || val > current {
			c.Increments[node] = val
		}
	}

	// Merge decrements
	for node, val := range other.Decrements {
		if current, exists := c.Decrements[node]; !exists || val > current {
			c.Decrements[node] = val
		}
	}
}

// ToMap converts to a map
func (c *PNCounter) ToMap() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var pos, neg int64
	for _, v := range c.Increments {
		pos += v
	}
	for _, v := range c.Decrements {
		neg += v
	}

	return map[string]interface{}{
		"node_id":    c.NodeID,
		"increments": c.Increments,
		"decrements": c.Decrements,
		"value":      pos - neg,
	}
}

// ORSet is an Observed-Remove Set CRDT
type ORSet struct {
	NodeID   string
	Elements map[string]*ORSetElement
	mu       sync.RWMutex
}

// ORSetElement represents an element in the set
type ORSetElement struct {
	ID        string
	Value     string
	IsRemoved bool
	AddedAt   string
	RemovedAt string
}

// NewORSet creates a new ORSet
func NewORSet(nodeID string) *ORSet {
	return &ORSet{
		NodeID:   nodeID,
		Elements: make(map[string]*ORSetElement),
	}
}

// Add adds an element to the set
func (s *ORSet) Add(value string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := generateID()
	s.Elements[id] = &ORSetElement{
		ID:        id,
		Value:     value,
		IsRemoved: false,
		AddedAt:   now(),
	}
	return id
}

// Remove removes an element from the set
func (s *ORSet) Remove(value string) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	var removed []string
	for id, elem := range s.Elements {
		if elem.Value == value && !elem.IsRemoved {
			elem.IsRemoved = true
			elem.RemovedAt = now()
			removed = append(removed, id)
		}
	}
	return removed
}

// Contains checks if a value is in the set
func (s *ORSet) Contains(value string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, elem := range s.Elements {
		if elem.Value == value && !elem.IsRemoved {
			return true
		}
	}
	return false
}

// ActiveElements returns all active (non-removed) elements
func (s *ORSet) ActiveElements() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []string
	seen := make(map[string]bool)
	for _, elem := range s.Elements {
		if !elem.IsRemoved && !seen[elem.Value] {
			result = append(result, elem.Value)
			seen[elem.Value] = true
		}
	}
	return result
}

// Merge merges another ORSet
func (s *ORSet) Merge(other *ORSet) {
	s.mu.Lock()
	defer s.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	for id, elem := range other.Elements {
		if existing, exists := s.Elements[id]; exists {
			// Merge removal status
			if elem.IsRemoved || existing.IsRemoved {
				existing.IsRemoved = true
				if existing.RemovedAt == "" || elem.RemovedAt > existing.RemovedAt {
					existing.RemovedAt = elem.RemovedAt
				}
			}
		} else {
			// Add new element
			s.Elements[id] = elem
		}
	}
}

// ToMap converts to a map
func (s *ORSet) ToMap() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	elements := make([]map[string]interface{}, 0, len(s.Elements))
	for _, elem := range s.Elements {
		elements = append(elements, map[string]interface{}{
			"id":         elem.ID,
			"value":      elem.Value,
			"is_removed": elem.IsRemoved,
			"added_at":   elem.AddedAt,
			"removed_at": elem.RemovedAt,
		})
	}

	return map[string]interface{}{
		"node_id":  s.NodeID,
		"elements": elements,
	}
}

// LWWRegister is a Last-Write-Wins Register CRDT
type LWWRegister struct {
	NodeID    string
	Value     string
	Timestamp string
	mu        sync.RWMutex
}

// NewLWWRegister creates a new LWWRegister
func NewLWWRegister(nodeID string) *LWWRegister {
	return &LWWRegister{
		NodeID:    nodeID,
		Timestamp: "0",
	}
}

// Set sets the register value
func (r *LWWRegister) Set(value string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Value = value
	r.Timestamp = now()
}

// Get gets the register value
func (r *LWWRegister) Get() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.Value
}

// Merge merges another LWWRegister
func (r *LWWRegister) Merge(other *LWWRegister) {
	r.mu.Lock()
	defer r.mu.Unlock()

	other.mu.RLock()
	defer other.mu.RUnlock()

	if other.Timestamp > r.Timestamp {
		r.Value = other.Value
		r.Timestamp = other.Timestamp
	}
}

// ToMap converts to a map
func (r *LWWRegister) ToMap() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return map[string]interface{}{
		"node_id":   r.NodeID,
		"value":     r.Value,
		"timestamp": r.Timestamp,
	}
}

// Helper functions
func generateID() string {
	// Simple ID generation - replace with proper UUID in production
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func now() string {
	return time.Now().UTC().Format(time.RFC3339Nano)
}
