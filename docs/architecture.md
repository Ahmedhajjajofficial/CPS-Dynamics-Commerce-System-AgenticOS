# Architecture

## System Overview

CPS Enterprise DCS follows a three-tier agentic architecture where autonomous agents operate at local, regional, and global levels. Each tier has distinct responsibilities and communicates via gRPC with Protocol Buffers.

```
                    ┌───────────────────────┐
                    │    Master Agent        │
                    │  (Global Orchestration)│
                    └───────────┬───────────┘
                                │
                ┌───────────────┼───────────────┐
                ▼                               ▼
    ┌───────────────────┐           ┌───────────────────┐
    │  Regional Agent   │           │  Regional Agent   │
    │  (Go / Raft)      │           │  (Go / Raft)      │
    └───────┬───────────┘           └───────┬───────────┘
            │                               │
    ┌───────┼───────┐               ┌───────┼───────┐
    ▼       ▼       ▼               ▼       ▼       ▼
  ┌───┐  ┌───┐  ┌───┐           ┌───┐  ┌───┐  ┌───┐
  │ L │  │ L │  │ L │           │ L │  │ L │  │ L │
  │ A │  │ A │  │ A │           │ A │  │ A │  │ A │
  └─┬─┘  └─┬─┘  └─┬─┘           └─┬─┘  └─┬─┘  └─┬─┘
    │       │       │               │       │       │
  ┌─┴─┐  ┌─┴─┐  ┌─┴─┐           ┌─┴─┐  ┌─┴─┐  ┌─┴─┐
  │POS│  │POS│  │POS│           │POS│  │POS│  │POS│
  └───┘  └───┘  └───┘           └───┘  └───┘  └───┘

  LA = Local Agent (Python)    POS = RockDeals POS (React)
```

---

## Component Details

### 1. POS Interface (React/TypeScript)

**Location:** `cps-enterprise-dcs/pos-interface/`

The point-of-sale frontend that cashiers interact with directly. Built with React 18, TypeScript, and Zustand for state management.

**Key features:**
- Product grid with category filtering and search
- Shopping cart with real-time tax calculation
- Session management (start/end with cash balance tracking)
- Offline-capable with sync status indicator
- Communicates with Local Agent via gRPC-Web

**State management:** Zustand stores handle session state, cart state, and sync status. The store is designed to queue transactions locally when offline and sync them when connectivity is restored.

### 2. Admin App (React/TypeScript)

**Location:** `app/`

The administration dashboard built with React 19, Vite 7, and shadcn/ui components. Provides management views for the commerce system.

**UI library:** Uses a comprehensive set of Radix UI primitives via shadcn/ui, including dialogs, dropdowns, tabs, accordions, and data visualization with Recharts.

### 3. Local Agent (Python)

**Location:** `cps-enterprise-dcs/local-agent/`

The edge computing agent that runs at each branch/store location. Operates autonomously even during complete network isolation.

**Responsibilities:**
- **Event Store**: Maintains a local SQLite event store for all financial events
- **CRDT Management**: Implements GCounter, PNCounter, ORSet, and LWWRegister for conflict-free data replication
- **gRPC Server**: Exposes services for the POS interface to submit transactions
- **Security**: Envelope encryption (AES-256-GCM) for all financial events
- **Sync**: Periodically syncs events and CRDT states with the Regional Agent

**Key modules:**
| Module | Purpose |
|--------|---------|
| `agent.py` | Core agent lifecycle and orchestration |
| `crdt.py` | CRDT implementations (GCounter, PNCounter, ORSet, LWWRegister) |
| `event_store.py` | SQLite-based append-only event store |
| `grpc_server.py` | gRPC service implementation |
| `security.py` | Envelope encryption and HMAC signing |

### 4. Regional Agent (Go)

**Location:** `cps-enterprise-dcs/regional-agent/`

The regional coordination agent that aggregates data from multiple Local Agents using Raft consensus.

**Responsibilities:**
- **Raft Consensus**: Uses HashiCorp Raft for leader election and log replication across regional nodes
- **CRDT Aggregation**: Merges CRDT states from all local agents in the region
- **Event Propagation**: Forwards events to the PostgreSQL event store
- **gRPC Server**: Accepts connections from Local Agents and provides coordination services

**Internal packages:**
| Package | Purpose |
|---------|---------|
| `internal/agent/` | Raft-based agent with bootstrap, join, and consensus logic |
| `internal/config/` | Configuration parsing from environment and flags |
| `internal/crdt/` | CRDT manager with merge operations |
| `internal/server/` | gRPC server for inter-agent communication |

---

## Event Sourcing

All state changes in DCS are captured as immutable events. The event store is the single source of truth.

### Event Flow

```
POS → Local Agent → SQLite (local) → Regional Agent → PostgreSQL (regional)
                                                     → Kafka (streaming)
```

### Event Structure

Each `SovereignFinancialEvent` contains:
- **Identifiers**: `event_id` (UUID v7), `correlation_id`, `idempotency_key`
- **Context**: `branch_id`, `agent_id`, `tenant_id`
- **Content**: `type` (EventType enum), `amount`, `currency`, encrypted `payload`
- **Causality**: `CausalContext` with Hybrid Logical Clock and Vector Clock
- **Audit**: `created_by`, `approved_by`, `digital_signature` (ECDSA)

### Event Types

Events are categorized into flows:
- **Sales**: `SALE_INITIATED`, `SALE_COMPLETED`, `SALE_REVERSED`, `SALE_PARTIALLY_REFUNDED`
- **Inventory**: `INVENTORY_RECEIVED`, `INVENTORY_SOLD`, `INVENTORY_TRANSFERRED_*`, `LOW_STOCK_ALERT`
- **Payments**: `PAYMENT_AUTHORIZED`, `PAYMENT_CAPTURED`, `PAYMENT_REFUNDED`, `PAYMENT_FAILED`
- **Accounting**: `LEDGER_ENTRY_POSTED`, `END_OF_DAY_RECONCILIATION`, `PERIOD_CLOSED`
- **Saga**: `SAGA_INITIATED`, `SAGA_STEP_COMPLETED`, `SAGA_COMPENSATION_TRIGGERED`
- **Agent**: `AGENT_DECISION_MADE`, `AGENT_ACTION_EXECUTED`, `AGENT_LEARNING_UPDATED`

---

## CRDTs (Conflict-Free Replicated Data Types)

CRDTs ensure that distributed state converges without coordination. DCS implements four CRDT types:

| CRDT | Use Case | Merge Strategy |
|------|----------|----------------|
| **GCounter** | Transaction counts, visit counts | Max per node |
| **PNCounter** | Inventory quantities (increment/decrement) | Max per node for both P and N |
| **ORSet** | Active product catalogs, user sessions | Union with tombstone tracking |
| **LWWRegister** | Product prices, configuration values | Latest timestamp wins |

### Synchronization

1. Local Agent maintains CRDT state locally
2. On sync interval (default 30s), Local Agent sends CRDT state bundles to Regional Agent
3. Regional Agent merges states from all Local Agents using type-specific merge functions
4. Merged state is propagated back to all Local Agents

---

## Consensus (Raft)

The Regional Agent uses the Raft consensus protocol for:

- **Leader Election**: One regional node becomes the leader; others are followers
- **Log Replication**: The leader replicates event logs to all followers
- **Fault Tolerance**: The cluster continues operating as long as a majority of nodes are alive

**Implementation**: Uses HashiCorp Raft library with BoltDB for stable storage.

---

## Security Model

### Envelope Encryption

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Plaintext   │────▶│  AES-256-GCM │────▶│  Ciphertext   │
│  Event Data  │     │  (DEK)       │     │  + Auth Tag   │
└──────────────┘     └──────┬───────┘     └──────────────┘
                            │
                     ┌──────▼───────┐
                     │  Vault KEK   │
                     │  (Transit)   │
                     └──────────────┘
```

1. A unique **Data Encryption Key (DEK)** is generated for each event
2. The DEK encrypts the event payload using **AES-256-GCM**
3. The DEK itself is encrypted by a **Key Encryption Key (KEK)** stored in HashiCorp Vault
4. An **HMAC-SHA512** signature is computed over visible metadata for integrity verification
5. An **ECDSA digital signature** is attached by the event creator for non-repudiation

### Zero-Knowledge Compliance

Compliance proofs can be generated without exposing underlying data, using zero-knowledge proof fields in the `SovereignPayload` message.

---

## Database Schema

The PostgreSQL event store (`cps-enterprise-dcs/event-store/schema.sql`) includes:

| Table | Purpose |
|-------|---------|
| `event_store` | Partitioned (8 hash partitions) append-only event log |
| `stream_metadata` | Stream version tracking for optimistic concurrency |
| `snapshots` | Aggregate snapshots for read optimization |
| `saga_instances` | Distributed transaction state machines |
| `saga_steps` | Individual saga step tracking |
| `projection_sales_summary` | Materialized daily sales view |
| `projection_inventory` | Real-time inventory levels |
| `projection_customer_loyalty` | Customer loyalty points and tiers |
| `crdt_state` | Persisted CRDT states per node |
| `agent_registry` | Agent discovery and health tracking |
| `audit_log` | Month-partitioned audit trail |

Row-level security (RLS) enforces tenant and branch isolation.

---

## Infrastructure

### Docker Services

The `docker-compose.yml` orchestrates the full stack:

| Service | Image | Port |
|---------|-------|------|
| PostgreSQL 16 | `postgres:16-alpine` | 5432 |
| Redis 7 | `redis:7-alpine` | 6379 |
| HashiCorp Vault | `hashicorp/vault:1.15` | 8200 |
| Zookeeper | `confluentinc/cp-zookeeper:7.5.0` | 2181 |
| Kafka | `confluentinc/cp-kafka:7.5.0` | 9092 |
| Prometheus | `prom/prometheus:v2.48.0` | 9090 |
| Grafana | `grafana/grafana:10.2.0` | 3001 |
| Regional Agent | Custom Go build | 12000, 12001, 50052 |
| Local Agent | Custom Python build | 50051 |
| POS Interface | Custom React build | 3000 |
| Nginx | `nginx:alpine` | 80, 443 |
