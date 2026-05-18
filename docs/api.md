# API Reference

## Overview

CPS Enterprise DCS uses **gRPC** with **Protocol Buffers** for all inter-service communication. The service definitions are in `cps-enterprise-dcs/proto/cps_enterprise_v4.proto`.

{/* TODO: Restore link to the proto file once its canonical location is confirmed (e.g. a GitHub permalink in the AgenticOS repo). The previous relative path `../cps-enterprise-dcs/proto/cps_enterprise_v4.proto` resolved outside the docs tree and 404'd. */}

---

## gRPC Services

### AccountingSwarmProtocol

The primary service for financial event management, reconciliation, and agent coordination.

**Port:** `50051` (Local Agent), `50052` (Regional Agent)

#### Unary RPCs

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `BroadcastFinancialEvent` | `SovereignFinancialEvent` | `AckResponse` | Submit a single financial event |
| `GetEventById` | `GetEventRequest` | `SovereignFinancialEvent` | Retrieve an event by ID |
| `GetEventsByRange` | `GetEventsRangeRequest` | `EventBatch` | Query events by time range with pagination |
| `RequestReconciliation` | `ReconciliationRequest` | `ReconciliationResponse` | Trigger reconciliation for a branch |
| `ResolveLedgerConflict` | `ConflictResolutionRequest` | `SovereignFinancialEvent` | Resolve conflicting events |

#### Server Streaming RPCs

| Method | Request | Response Stream | Description |
|--------|---------|-----------------|-------------|
| `SubscribeEvents` | `SubscribeRequest` | `SovereignFinancialEvent` | Subscribe to real-time event stream |
| `StreamReconciliation` | `ReconciliationRequest` | `ReconciliationResponse` | Stream reconciliation results |

#### Client Streaming RPCs

| Method | Request Stream | Response | Description |
|--------|---------------|----------|-------------|
| `StreamOfflineEvents` | `SovereignFinancialEvent` | `BatchAckResponse` | Batch sync offline events |
| `StreamCRDTUpdates` | `CRDTStateBundle` | `BatchAckResponse` | Stream CRDT state updates |

#### Bidirectional Streaming RPCs

| Method | Request Stream | Response Stream | Description |
|--------|---------------|-----------------|-------------|
| `SwarmEventExchange` | `SovereignFinancialEvent` | `SovereignFinancialEvent` | Real-time event exchange between agents |
| `SyncCRDTState` | `CRDTStateBundle` | `CRDTStateBundle` | Bidirectional CRDT synchronization |
| `AgentCommunication` | `AgentMessage` | `AgentMessage` | General agent-to-agent messaging |

---

### QueryProtocol

Read-model query service for dashboards and reporting.

#### Unary RPCs

| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `GetBranchSummary` | `BranchQuery` | `BranchSummary` | Get branch sales/session summary |
| `GetInventoryStatus` | `InventoryQuery` | `InventoryStatus` | Get product inventory levels |
| `GetSalesReport` | `SalesReportQuery` | `SalesReport` | Get sales report by period |
| `GetLedgerBalance` | `LedgerQuery` | `LedgerBalance` | Get account balance |

#### Server Streaming RPCs

| Method | Request | Response Stream | Description |
|--------|---------|-----------------|-------------|
| `SubscribeDashboard` | `DashboardSubscription` | `DashboardUpdate` | Real-time dashboard metrics |

---

## Core Messages

### SovereignFinancialEvent

The central event type in the system. Every financial transaction is captured as an immutable event.

```protobuf
message SovereignFinancialEvent {
  string event_id = 1;           // UUID v7 (time-ordered)
  string correlation_id = 2;     // Saga event linkage
  string idempotency_key = 3;    // Deterministic SHA-256
  uint32 stream_version = 4;     // Optimistic concurrency control
  string event_hash = 5;         // SHA-256 of event content

  string branch_id = 6;
  string agent_id = 7;
  string saga_id = 8;
  string tenant_id = 9;

  EventType type = 10;
  double amount = 11;
  string currency = 12;
  SovereignPayload payload = 13;
  bytes public_metadata = 14;

  CausalContext causal_context = 15;
  google.protobuf.Timestamp ts = 16;
  int64 received_at = 17;

  string created_by = 18;
  string approved_by = 19;
  bytes digital_signature = 20;
}
```

### SovereignPayload

Envelope-encrypted payload attached to each event.

```protobuf
message SovereignPayload {
  bytes encrypted_data = 1;          // AES-256-GCM ciphertext
  bytes encrypted_dek = 2;          // DEK encrypted by KEK via Vault
  string kms_key_id = 3;           // KEK identifier in HashiCorp Vault
  bytes iv = 4;                    // 12-byte IV, unique per event
  bytes auth_tag = 5;             // GCM authentication tag (16 bytes)
  string hmac_signature = 6;      // HMAC-SHA512 on visible metadata
  uint32 schema_version = 7;
  bytes encrypted_inner_layer = 8; // Additional encryption for sensitive data
  string inner_key_derivation = 9; // HKDF identifier for inner layer
  bytes compliance_proof = 10;     // Zero-knowledge proof for compliance
  bytes audit_trail_hash = 11;    // Hash chain for tamper detection
}
```

### CausalContext

Tracks causality and ordering across distributed nodes.

```protobuf
message CausalContext {
  HybridLogicalClock hlc = 1;
  VectorClock v_clock = 2;
  repeated string parent_event_ids = 3;
  int64 lamport_timestamp = 4;
}

message HybridLogicalClock {
  int64 physical_ms = 1;    // Physical time in milliseconds
  int32 logical = 2;        // Logical counter for concurrent events
  string node_id = 3;       // Node identifier
  int32 counter = 4;        // Monotonic counter per node
}
```

---

## Event Types

Events are organized into domain-specific flows:

### Sales Flow
| Type | Value | Description |
|------|-------|-------------|
| `SALE_INITIATED` | 1 | Sale transaction started |
| `SALE_COMPLETED` | 2 | Sale successfully completed |
| `SALE_REVERSED` | 3 | Sale fully reversed/voided |
| `SALE_PARTIALLY_REFUNDED` | 4 | Partial refund processed |

### Inventory Flow
| Type | Value | Description |
|------|-------|-------------|
| `INVENTORY_RECEIVED` | 10 | Stock received from supplier |
| `INVENTORY_RESERVED` | 11 | Stock reserved for a pending sale |
| `INVENTORY_RESERVATION_CANCELLED` | 12 | Reservation cancelled |
| `INVENTORY_SOLD` | 13 | Stock sold to customer |
| `INVENTORY_TRANSFERRED_OUT` | 14 | Stock transferred out of branch |
| `INVENTORY_TRANSFERRED_IN` | 15 | Stock transferred into branch |
| `INVENTORY_ADJUSTMENT` | 16 | Manual stock adjustment |
| `INVENTORY_COUNTED` | 17 | Physical stock count recorded |
| `LOW_STOCK_ALERT` | 18 | Stock below reorder point |
| `STOCKOUT_DETECTED` | 19 | Stock at zero |

### Payment Flow
| Type | Value | Description |
|------|-------|-------------|
| `PAYMENT_AUTHORIZED` | 30 | Payment authorized |
| `PAYMENT_CAPTURED` | 31 | Payment captured/settled |
| `PAYMENT_REFUNDED` | 32 | Payment refunded |
| `PAYMENT_FAILED` | 33 | Payment failed |
| `PAYMENT_CHARGEBACK` | 34 | Chargeback received |

### Accounting Flow
| Type | Value | Description |
|------|-------|-------------|
| `LEDGER_ENTRY_POSTED` | 20 | Ledger entry created |
| `LEDGER_REVERSAL_POSTED` | 21 | Ledger reversal |
| `END_OF_DAY_RECONCILIATION` | 22 | End-of-day reconciliation |
| `PERIOD_CLOSED` | 23 | Accounting period closed |
| `TRIAL_BALANCE_GENERATED` | 24 | Trial balance generated |

### Saga Flow
| Type | Value | Description |
|------|-------|-------------|
| `SAGA_INITIATED` | 60 | Distributed transaction started |
| `SAGA_STEP_COMPLETED` | 61 | Saga step completed |
| `SAGA_COMPENSATION_TRIGGERED` | 62 | Compensation (rollback) started |
| `SAGA_COMPENSATION_COMPLETED` | 63 | Compensation completed |
| `SAGA_FAILED` | 64 | Saga failed |
| `TRANSACTION_REJECTED` | 65 | Transaction rejected |

### Agent Flow
| Type | Value | Description |
|------|-------|-------------|
| `AGENT_DECISION_MADE` | 70 | Agent made an autonomous decision |
| `AGENT_ACTION_EXECUTED` | 71 | Agent executed an action |
| `AGENT_LEARNING_UPDATED` | 72 | Agent model updated |

---

## CRDT Messages

### GCounter (Grow-Only Counter)
```protobuf
message GCounter {
  string node_id = 1;
  map<string, uint64> increments = 2;
}
```

### PNCounter (Positive-Negative Counter)
```protobuf
message PNCounter {
  string node_id = 1;
  map<string, uint64> increments = 2;
  map<string, uint64> decrements = 3;
}
```

### ORSet (Observed-Remove Set)
```protobuf
message ORSet {
  string node_id = 1;
  repeated ORSetElement elements = 2;
}
```

### LWWRegister (Last-Write-Wins Register)
```protobuf
message LWWRegister {
  string value = 1;
  HybridLogicalClock timestamp = 2;
  string node_id = 3;
}
```

### CRDTStateBundle (Sync Envelope)
```protobuf
message CRDTStateBundle {
  string branch_id = 1;
  string crdt_type = 2;        // "GCounter", "PNCounter", "ORSet", "LWWRegister"
  string crdt_id = 3;
  bytes serialized_state = 4;
  HybridLogicalClock last_updated = 5;
  uint64 version = 6;
}
```

---

## Saga Orchestration

### Saga Types

| Type | Description |
|------|-------------|
| `STANDARD_SALE` | Normal point-of-sale transaction |
| `REFUND_PROCESSING` | Refund with inventory return |
| `INVENTORY_TRANSFER` | Cross-branch stock transfer |
| `BATCH_OPERATION` | Bulk operation (e.g., price update) |
| `CROSS_BRANCH_SALE` | Sale involving multiple branches |

### Saga States

```
PENDING → RUNNING → COMPLETED
                  ↘ COMPENSATING → COMPENSATED
                  ↘ FAILED
                  ↘ TIMED_OUT
```

---

## Conflict Resolution Strategies

| Strategy | Description |
|----------|-------------|
| `LAST_WRITER_WINS` | Most recent timestamp wins |
| `VECTOR_CLOCK_ORDER` | Causal ordering via vector clocks |
| `AGENT_VOTE` | Majority vote among agents |
| `MASTER_DECIDES` | Master agent arbitrates |
| `CRDT_MERGE` | Automatic CRDT merge |
| `CUSTOM_RESOLUTION` | Application-specific logic |

---

## Agent Types

| Type | Description |
|------|-------------|
| `LOCAL_AGENT` | Branch-level edge agent (Python) |
| `REGIONAL_AGENT` | Regional coordination agent (Go) |
| `MASTER_AGENT` | Global orchestration agent |
| `PREDICTIVE_AGENT` | ML-based forecasting agent |
| `COMPLIANCE_AGENT` | Regulatory compliance agent |
| `SECURITY_AGENT` | Security monitoring agent |
