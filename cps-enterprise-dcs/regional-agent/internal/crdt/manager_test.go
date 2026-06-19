package crdt

import (
	"encoding/json"
	"testing"
)

// ── PNCounter ────────────────────────────────────────────────────────────────

func TestPNCounter_InitialValue(t *testing.T) {
	c := NewPNCounter("n1")
	if c.Value() != 0 {
		t.Fatalf("expected 0, got %d", c.Value())
	}
}

func TestPNCounter_Increment(t *testing.T) {
	c := NewPNCounter("n1")
	c.Increment(10)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %d", c.Value())
	}
}

func TestPNCounter_Decrement(t *testing.T) {
	c := NewPNCounter("n1")
	c.Increment(10)
	c.Decrement(3)
	if c.Value() != 7 {
		t.Fatalf("expected 7, got %d", c.Value())
	}
}

func TestPNCounter_MultipleIncrements(t *testing.T) {
	c := NewPNCounter("n1")
	c.Increment(5)
	c.Increment(3)
	c.Increment(2)
	if c.Value() != 10 {
		t.Fatalf("expected 10, got %d", c.Value())
	}
}

func TestPNCounter_Merge(t *testing.T) {
	a := NewPNCounter("n1")
	b := NewPNCounter("n2")
	a.Increment(10)
	a.Decrement(2)
	b.Increment(5)
	b.Decrement(1)

	a.Merge(b)
	// n1: inc=10, dec=2 | n2: inc=5, dec=1  → (10+5) - (2+1) = 12
	if a.Value() != 12 {
		t.Fatalf("expected 12, got %d", a.Value())
	}
}

func TestPNCounter_MergeCommutative(t *testing.T) {
	a := NewPNCounter("n1")
	b := NewPNCounter("n2")
	a.Increment(10)
	b.Increment(5)

	aCopy := NewPNCounter("n1")
	aCopy.Increment(10)
	bCopy := NewPNCounter("n2")
	bCopy.Increment(5)

	aCopy.Merge(bCopy)
	bMerged := NewPNCounter("n2")
	bMerged.Increment(5)
	bMerged.Merge(a)

	if aCopy.Value() != bMerged.Value() {
		t.Fatalf("merge not commutative: %d != %d", aCopy.Value(), bMerged.Value())
	}
}

func TestPNCounter_MergeIdempotent(t *testing.T) {
	a := NewPNCounter("n1")
	b := NewPNCounter("n2")
	a.Increment(10)
	b.Increment(5)

	a.Merge(b)
	v1 := a.Value()
	a.Merge(b)
	v2 := a.Value()

	if v1 != v2 {
		t.Fatalf("merge not idempotent: %d != %d", v1, v2)
	}
}

func TestPNCounter_MergeSameNode(t *testing.T) {
	a := NewPNCounter("n1")
	b := NewPNCounter("n1")
	a.Increment(10)
	b.Increment(5)

	a.Merge(b)
	if a.Value() != 10 { // max(10, 5) = 10
		t.Fatalf("expected 10, got %d", a.Value())
	}
}

func TestPNCounter_ToMap(t *testing.T) {
	c := NewPNCounter("n1")
	c.Increment(7)
	c.Decrement(2)
	m := c.ToMap()

	if m["node_id"] != "n1" {
		t.Fatal("missing node_id")
	}
}

// ── ORSet ────────────────────────────────────────────────────────────────────

func TestORSet_EmptySet(t *testing.T) {
	s := NewORSet("n1")
	active := s.ActiveElements()
	if active != nil && len(active) != 0 {
		t.Fatalf("expected empty, got %v", active)
	}
}

func TestORSet_Add(t *testing.T) {
	s := NewORSet("n1")
	id := s.Add("product_A")
	if id == "" {
		t.Fatal("expected non-empty id")
	}
	if !s.Contains("product_A") {
		t.Fatal("expected set to contain product_A")
	}
}

func TestORSet_AddMultiple(t *testing.T) {
	s := NewORSet("n1")
	s.Add("a")
	s.Add("b")
	s.Add("c")
	active := s.ActiveElements()
	if len(active) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(active))
	}
}

func TestORSet_Remove(t *testing.T) {
	s := NewORSet("n1")
	s.Add("a")
	s.Add("b")
	removed := s.Remove("a")
	if len(removed) != 1 {
		t.Fatalf("expected 1 removed, got %d", len(removed))
	}
	if s.Contains("a") {
		t.Fatal("a should be removed")
	}
	if !s.Contains("b") {
		t.Fatal("b should remain")
	}
}

func TestORSet_RemoveNonexistent(t *testing.T) {
	s := NewORSet("n1")
	removed := s.Remove("nope")
	if removed != nil && len(removed) != 0 {
		t.Fatalf("expected no removals, got %v", removed)
	}
}

func TestORSet_RemoveMultipleSameValue(t *testing.T) {
	s := NewORSet("n1")
	s.Add("x")
	s.Add("x")
	removed := s.Remove("x")
	if len(removed) != 2 {
		t.Fatalf("expected 2 removed, got %d", len(removed))
	}
	if s.Contains("x") {
		t.Fatal("x should be fully removed")
	}
}

func TestORSet_Merge(t *testing.T) {
	a := NewORSet("n1")
	b := NewORSet("n2")
	a.Add("x")
	b.Add("y")

	a.Merge(b)

	if !a.Contains("x") {
		t.Fatal("missing x after merge")
	}
	if !a.Contains("y") {
		t.Fatal("missing y after merge")
	}
}

func TestORSet_MergePreservesRemoval(t *testing.T) {
	a := NewORSet("n1")
	id := a.Add("x")

	b := NewORSet("n2")
	b.Elements[id] = &ORSetElement{
		ID: id, Value: "x", IsRemoved: true, RemovedAt: now(),
	}

	a.Merge(b)
	if a.Contains("x") {
		t.Fatal("x should be removed after merge")
	}
}

func TestORSet_ActiveElementsDedup(t *testing.T) {
	s := NewORSet("n1")
	s.Add("dup")
	s.Add("dup")
	active := s.ActiveElements()
	count := 0
	for _, v := range active {
		if v == "dup" {
			count++
		}
	}
	if count != 1 {
		t.Fatalf("ActiveElements should dedup; got %d occurrences", count)
	}
}

func TestORSet_ToMap(t *testing.T) {
	s := NewORSet("n1")
	s.Add("a")
	m := s.ToMap()
	if m["node_id"] != "n1" {
		t.Fatal("missing node_id")
	}
	elems, ok := m["elements"].([]map[string]interface{})
	if !ok {
		t.Fatal("elements wrong type")
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
}

// ── LWWRegister ──────────────────────────────────────────────────────────────

func TestLWWRegister_InitialEmpty(t *testing.T) {
	r := NewLWWRegister("n1")
	if r.Get() != "" {
		t.Fatalf("expected empty, got %q", r.Get())
	}
}

func TestLWWRegister_SetAndGet(t *testing.T) {
	r := NewLWWRegister("n1")
	r.Set("hello")
	if r.Get() != "hello" {
		t.Fatalf("expected hello, got %q", r.Get())
	}
}

func TestLWWRegister_Merge(t *testing.T) {
	a := NewLWWRegister("n1")
	b := NewLWWRegister("n2")

	a.mu.Lock()
	a.Value = "old"
	a.Timestamp = "2024-01-01T00:00:00Z"
	a.mu.Unlock()

	b.mu.Lock()
	b.Value = "new"
	b.Timestamp = "2024-01-02T00:00:00Z"
	b.mu.Unlock()

	a.Merge(b)
	if a.Get() != "new" {
		t.Fatalf("expected 'new', got %q", a.Get())
	}
}

func TestLWWRegister_MergeOlderDoesNotOverwrite(t *testing.T) {
	a := NewLWWRegister("n1")
	b := NewLWWRegister("n2")

	a.mu.Lock()
	a.Value = "newer"
	a.Timestamp = "2024-01-02T00:00:00Z"
	a.mu.Unlock()

	b.mu.Lock()
	b.Value = "older"
	b.Timestamp = "2024-01-01T00:00:00Z"
	b.mu.Unlock()

	a.Merge(b)
	if a.Get() != "newer" {
		t.Fatalf("expected 'newer', got %q", a.Get())
	}
}

func TestLWWRegister_ToMap(t *testing.T) {
	r := NewLWWRegister("n1")
	r.Set("val")
	m := r.ToMap()
	if m["node_id"] != "n1" {
		t.Fatal("missing node_id")
	}
	if m["value"] != "val" {
		t.Fatalf("expected val, got %v", m["value"])
	}
}

// ── Manager ──────────────────────────────────────────────────────────────────

func TestManager_GetOrCreateCounter(t *testing.T) {
	mgr := NewManager("n1")
	c := mgr.GetOrCreateCounter("inv_001")
	if c == nil {
		t.Fatal("expected counter")
	}
	c2 := mgr.GetOrCreateCounter("inv_001")
	if c != c2 {
		t.Fatal("expected same counter instance")
	}
}

func TestManager_GetCounter(t *testing.T) {
	mgr := NewManager("n1")
	_, ok := mgr.GetCounter("nope")
	if ok {
		t.Fatal("expected not found")
	}
	mgr.GetOrCreateCounter("inv_001")
	c, ok := mgr.GetCounter("inv_001")
	if !ok || c == nil {
		t.Fatal("expected found")
	}
}

func TestManager_MergeCounter(t *testing.T) {
	mgr := NewManager("n1")
	local := mgr.GetOrCreateCounter("inv")
	local.Increment(10)

	remote := NewPNCounter("n2")
	remote.Increment(5)

	mgr.MergeCounter("inv", remote)
	c, _ := mgr.GetCounter("inv")
	if c.Value() != 15 {
		t.Fatalf("expected 15, got %d", c.Value())
	}
}

func TestManager_MergeCounterNew(t *testing.T) {
	mgr := NewManager("n1")
	remote := NewPNCounter("n2")
	remote.Increment(42)

	mgr.MergeCounter("new_counter", remote)
	c, ok := mgr.GetCounter("new_counter")
	if !ok {
		t.Fatal("expected counter to be created")
	}
	if c.Value() != 42 {
		t.Fatalf("expected 42, got %d", c.Value())
	}
}

func TestManager_GetOrCreateSet(t *testing.T) {
	mgr := NewManager("n1")
	s := mgr.GetOrCreateSet("promos")
	if s == nil {
		t.Fatal("expected set")
	}
	s2 := mgr.GetOrCreateSet("promos")
	if s != s2 {
		t.Fatal("expected same set instance")
	}
}

func TestManager_MergeSet(t *testing.T) {
	mgr := NewManager("n1")
	local := mgr.GetOrCreateSet("promos")
	local.Add("promo_A")

	remote := NewORSet("n2")
	remote.Add("promo_B")

	mgr.MergeSet("promos", remote)
	if !local.Contains("promo_A") || !local.Contains("promo_B") {
		t.Fatal("merge should include both promos")
	}
}

func TestManager_MergeSetNew(t *testing.T) {
	mgr := NewManager("n1")
	remote := NewORSet("n2")
	remote.Add("item")

	mgr.MergeSet("new_set", remote)
	s := mgr.GetOrCreateSet("new_set")
	if !s.Contains("item") {
		t.Fatal("expected item in merged set")
	}
}

func TestManager_GetOrCreateRegister(t *testing.T) {
	mgr := NewManager("n1")
	r := mgr.GetOrCreateRegister("price_001")
	if r == nil {
		t.Fatal("expected register")
	}
	r2 := mgr.GetOrCreateRegister("price_001")
	if r != r2 {
		t.Fatal("expected same register instance")
	}
}

func TestManager_MergeRegister(t *testing.T) {
	mgr := NewManager("n1")
	local := mgr.GetOrCreateRegister("price")
	local.mu.Lock()
	local.Value = "old"
	local.Timestamp = "2024-01-01T00:00:00Z"
	local.mu.Unlock()

	remote := NewLWWRegister("n2")
	remote.mu.Lock()
	remote.Value = "new_price"
	remote.Timestamp = "2024-01-02T00:00:00Z"
	remote.mu.Unlock()

	mgr.MergeRegister("price", remote)
	if local.Get() != "new_price" {
		t.Fatalf("expected new_price, got %q", local.Get())
	}
}

func TestManager_MergeRegisterNew(t *testing.T) {
	mgr := NewManager("n1")
	remote := NewLWWRegister("n2")
	remote.Set("val")

	mgr.MergeRegister("new_reg", remote)
	r := mgr.GetOrCreateRegister("new_reg")
	if r.Get() != "val" {
		t.Fatalf("expected val, got %q", r.Get())
	}
}

func TestManager_GetAllCounters(t *testing.T) {
	mgr := NewManager("n1")
	mgr.GetOrCreateCounter("a")
	mgr.GetOrCreateCounter("b")
	all := mgr.GetAllCounters()
	if len(all) != 2 {
		t.Fatalf("expected 2 counters, got %d", len(all))
	}
}

func TestManager_SerializeState(t *testing.T) {
	mgr := NewManager("n1")
	c := mgr.GetOrCreateCounter("inv")
	c.Increment(5)
	s := mgr.GetOrCreateSet("promos")
	s.Add("deal")
	r := mgr.GetOrCreateRegister("price")
	r.Set("9.99")

	data, err := mgr.SerializeState()
	if err != nil {
		t.Fatalf("serialize failed: %v", err)
	}

	var state map[string]interface{}
	if err := json.Unmarshal(data, &state); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	if state["node_id"] != "n1" {
		t.Fatal("missing node_id")
	}
	counters := state["counters"].(map[string]interface{})
	if _, ok := counters["inv"]; !ok {
		t.Fatal("missing inv counter")
	}
}

// ── Helpers ──────────────────────────────────────────────────────────────────

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()
	if id1 == "" || id2 == "" {
		t.Fatal("IDs should not be empty")
	}
}

func TestNow(t *testing.T) {
	ts := now()
	if ts == "" {
		t.Fatal("timestamp should not be empty")
	}
}
