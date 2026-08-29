<p align="center">
  <img src="assets/brand/logo-light.png" alt="CPS Logo" width="200"/>
</p>

# CP'S Enterprise Dynamics Commerce System (DCS)

<p align="center">
  <strong>The Sovereign Commerce Platform</strong><br/>
  Built with Logic of Sovereignty, Not Dependency
</p>

<p align="center">
  <a href="#architecture">Architecture</a> &bull;
  <a href="#quick-start">Quick Start</a> &bull;
  <a href="#development">Development</a> &bull;
  <a href="#documentation">Documentation</a> &bull;
  <a href="#deployment">Deployment</a>
  &bull; <a href="SECURITY.md">Security</a>
</p>

---

## Overview

CP'S Enterprise DCS is a distributed commerce platform built on **event sourcing**, **CRDTs**, **CQRS**, and an **agentic architecture**. It is designed for offline-first retail operations with sovereign data ownership, meaning your business data stays under your control at all times.

### Core Principles

- **Sovereign Data** — Full data ownership with envelope encryption (AES-256-GCM + HashiCorp Vault)
- **Offline-First** — Business operations continue during complete network isolation (up to 72 hours)
- **Event Sourcing** — Immutable, append-only audit trail for compliance and forensics
- **CRDT Consistency** — Conflict-free replicated data types guarantee convergence without coordination
- **CQRS Projections** — Real-time materialized views for high-performance queries
- **Agentic Intelligence** — Autonomous agents manage operations across local, regional, and global levels
- **Observability** — Distributed tracing, metrics, and health monitoring across all services

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────────────────────────┐
│                        CP'S Enterprise DCS v5.0 - Enterprise Grade                      │
├─────────────────────────────────────────────────────────────────────────────────────────┤
│                                                                                         │
│  ┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐                    │
│  │   RockDeals POS │    │   RockDeals POS │    │   RockDeals POS │                    │
│  │   (React/TS)    │    │   (React/TS)    │    │   (React/TS)    │                    │
│  └────────┬────────┘    └────────┬────────┘    └────────┬────────┘                    │
│           │                      │                      │                              │
│           ▼                      ▼                      ▼                              │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                         Local Agent (Python)                                │      │
│  │   • SQLite Event Store   • CRDT State   • gRPC/HTTP Gateway                │      │
│  │   • Circuit Breaker      • Saga Compensation   • Offline Recovery           │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│           │                      │                      │                              │
│           │                      │                      │                              │
│           ▼                      ▼                      ▼                              │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                      Regional Agent (Go)                                    │      │
│  │   • Raft Consensus   • CRDT Aggregation   • PostgreSQL Store                │      │
│  │   • Auth Interceptor • mTLS Security      • HTTP Admin API                  │      │
│  │   • Forecasting      • Health Checks       • Event Streaming                │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│           │                      │                      │                              │
│           │                      │                      │                              │
│           ▼                      ▼                      ▼                              │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                        Master Agent (Go)                                    │      │
│  │   • Global Orchestration   • Cross-Region Reconciliation                    │      │
│  │   • Decision Engine        • Kafka Event Streaming                           │      │
│  │   • ML Model Management    • mTLS Security                                   │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│           │                      │                      │                              │
│           │                      │                      │                              │
│           ▼                      ▼                      ▼                              │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                   Projection Workers (Go)                                   │      │
│  │   • Sales Summary   • Inventory Status   • Branch Metrics                    │      │
│  │   • CQRS Materialized Views   • Real-time Updates                            │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│                                                                                         │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                   ML Platform                                                │      │
│  │   • Feature Store   • Anomaly Detection   • Model Training Pipeline          │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│                                                                                         │
│  ┌──────────────────────────────────────────────────────────────────────────────┐      │
│  │                   Observability Stack                                        │      │
│  │   • Jaeger (Distributed Tracing)   • Prometheus (Metrics)                   │      │
│  │   • Grafana (Dashboards)           • OpenTelemetry                           │      │
│  └──────────────────────────────────────────────────────────────────────────────┘      │
│                                                                                         │
└─────────────────────────────────────────────────────────────────────────────────────────┘
```

### Agent Hierarchy

| Agent | Technology | Responsibility | Deployment |
|-------|-----------|----------------|------------|
| **Local Agent** | Python 3.11+, SQLite, gRPC | Edge computing, offline ops, POS integration | Per branch/device |
| **Regional Agent** | Go 1.23+, Raft, PostgreSQL | Regional consensus, CRDT aggregation, forecasting | Per region |
| **Master Agent** | Go 1.23+, Kafka, PostgreSQL | Global orchestration, reconciliation, ML coordination | Global (HA) |

### Repository Structure

```
.
├── app/                                    # Main admin frontend (React 19 + Vite)
│   ├── src/
│   │   ├── components/ui/                  # shadcn/ui component library
│   │   ├── components/glass/               # Glass UI components
│   │   ├── services/adminApi.ts            # Admin API service
│   │   └── App.tsx                         # Admin dashboard
│   ├── package.json
│   └── vite.config.ts
│
├── cps-enterprise-dcs/                     # Core platform modules
│   ├── pos-interface/                      # RockDeals POS frontend (React 18 + Vite)
│   │   ├── src/
│   │   │   ├── components/                 # ProductGrid, Cart, etc.
│   │   │   ├── store/                      # Zustand state management
│   │   │   ├── services/api.ts             # Local Agent HTTP API
│   │   │   └── App.tsx
│   │   └── package.json
│   │
│   ├── local-agent/                        # Edge computing agent (Python 3.11+)
│   │   ├── src/
│   │   │   ├── agent.py                    # Core agent logic + circuit breaker
│   │   │   ├── crdt.py                     # CRDT implementations
│   │   │   ├── event_store.py              # SQLite event store
│   │   │   ├── grpc_server.py              # gRPC service
│   │   │   ├── http_gateway.py             # REST API for POS
│   │   │   ├── security.py                 # Envelope encryption
│   │   │   └── main.py                     # Entry point
│   │   ├── requirements.txt
│   │   └── Dockerfile
│   │
│   ├── regional-agent/                     # Regional coordination agent (Go 1.23+)
│   │   ├── internal/
│   │   │   ├── agent/                      # Raft-based agent + forecasting
│   │   │   ├── config/                     # Configuration
│   │   │   ├── crdt/                       # CRDT manager
│   │   │   ├── proto/                      # Generated protobuf stubs
│   │   │   ├── server/                     # gRPC server + auth + health
│   │   │   └── store/                      # PostgreSQL store
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── master-agent/                       # Global orchestrator (Go 1.23+)
│   │   ├── internal/
│   │   │   ├── agent/                      # Master agent core + Kafka producer
│   │   │   ├── config/                     # Configuration
│   │   │   ├── proto/                      # Master agent protobuf stubs
│   │   │   ├── server/                     # gRPC handlers + mTLS
│   │   │   └── store/                      # PostgreSQL store
│   │   ├── pkg/kafka/                      # Kafka producer
│   │   ├── main.go
│   │   ├── go.mod
│   │   └── Dockerfile
│   │
│   ├── projection-workers/                 # CQRS projection workers (Go 1.23+)
│   │   ├── internal/
│   │   │   ├── sales/                      # Sales summary projection
│   │   │   ├── inventory/                  # Inventory projection
│   │   │   └── branch/                     # Branch metrics projection
│   │   ├── pkg/kafka/                      # Kafka consumer
│   │   ├── main.go
│   │   └── go.mod
│   │
│   ├── ml-platform/                        # Machine Learning platform
│   │   ├── feature-store/                  # Feature store (PostgreSQL + Redis)
│   │   ├── anomaly-detection/              # Anomaly detection engine
│   │   │   └── pkg/stats/                  # Statistical utilities
│   │
│   ├── event-store/                        # PostgreSQL event store schema
│   │   ├── schema.sql                      # Main schema + projections + partitions
│   │   ├── master-schema.sql               # Master agent tables
│   │   └── postgres-ha.sql                 # PostgreSQL HA configuration
│   │
│   ├── proto/                              # Protocol Buffers definitions
│   │   ├── cps_enterprise_v4.proto         # Core gRPC services
│   │   └── master_agent_v1.proto           # Master agent services
│   │
│   ├── infrastructure/                     # Deployment configuration
│   │   ├── docker-compose.yml              # Full stack deployment
│   │   ├── prometheus.yml                  # Prometheus configuration
│   │   ├── nginx.conf                      # Reverse proxy
│   │   ├── tls/                            # mTLS certificate generation
│   │   │   ├── generate-certs.sh           # Bash cert generator
│   │   │   └── generate.go                 # Go cert generator
│   │   ├── tracing/                        # OpenTelemetry configuration
│   │   │   ├── tracing.go                  # Go tracing utilities
│   │   │   └── otel.py                     # Python tracing utilities
│   │   └── kubernetes/                     # K8s manifests
│   │       ├── 00-namespace-config.yaml
│   │       ├── 01-postgres.yaml
│   │       ├── 02-redis.yaml
│   │       ├── 03-regional-agent.yaml
│   │       ├── 04-local-agent.yaml
│   │       ├── 05-master-agent.yaml
│   │       ├── 06-projection-workers.yaml
│   │       └── 07-jaeger.yaml
│   │
│   └── .env.example                        # Environment variable template
│
├── docs/                                   # Documentation
│   ├── architecture.md                     # System architecture deep-dive
│   ├── development.md                      # Developer setup guide
│   ├── api.md                              # API reference
│   └── deployment.md                       # Deployment guide
│
├── .github/workflows/                      # CI/CD pipelines
│   └── workflows-fix.yml                   # GitHub Actions workflows
│
├── Makefile                                # Build, test, lint, run commands
└── README.md                               # This file
```

### Technology Stack

| Component | Technology | Purpose |
|-----------|------------|---------|
| **Admin App** | React 19, TypeScript, Vite 7, shadcn/ui | Administration dashboard |
| **POS Interface** | React 18, TypeScript, Vite, Zustand, Tailwind | Cashier point-of-sale |
| **Local Agent** | Python 3.11+, gRPC, SQLite, asyncio | Edge computing & offline ops |
| **Regional Agent** | Go 1.23+, Raft, gRPC, PostgreSQL | Regional consensus & coordination |
| **Master Agent** | Go 1.23+, Kafka, gRPC, PostgreSQL | Global orchestration & ML |
| **Projection Workers** | Go 1.23+, Kafka, PostgreSQL | CQRS materialized views |
| **ML Platform** | Go, Redis, PostgreSQL | Feature store & anomaly detection |
| **Event Store** | PostgreSQL 16, partitioned tables | Immutable event log |
| **Message Bus** | Apache Kafka 7.5 | Event streaming |
| **Cache** | Redis 7 | Session & idempotency cache |
| **Secrets** | HashiCorp Vault 1.15 | Envelope encryption key management |
| **Observability** | Prometheus, Grafana, Jaeger, OpenTelemetry | Tracing, metrics, dashboards |
| **Security** | mTLS, PgBouncer, Patroni | Service mesh & HA |

---

## Architecture Components

### 1. Event Sourcing & CQRS

All state changes are captured as immutable events in PostgreSQL. Projection workers consume these events to maintain materialized views for high-performance reads.

**Key Features:**
- Hash-partitioned event store (monthly partitions)
- Automatic partition creation via triggers
- Row-level security for multi-tenancy
- CQRS projections: sales, inventory, branch metrics, customer loyalty

### 2. CRDT Consistency

Conflict-free Replicated Data Types ensure convergence across distributed nodes without coordination:

| CRDT Type | Use Case |
|-----------|----------|
| **GCounter** | Sales counters, inventory counts |
| **PNCounter** | Increment/decrement operations |
| **ORSet** | Product sets, customer sets |
| **LWWRegister** | Last-write-wins fields (prices, names) |

### 3. Agentic Intelligence

Three-tier agent hierarchy with autonomous decision-making:

- **Local Agent**: Circuit breaker, saga compensation, offline recovery
- **Regional Agent**: Raft consensus, forecasting, auth interceptor, health API
- **Master Agent**: Global reconciliation, decision engine, ML coordination

### 4. Observability

- **Distributed Tracing**: OpenTelemetry with Jaeger backend
- **Metrics**: Prometheus scraping all services
- **Dashboards**: Grafana with pre-built DCS dashboards
- **Health Checks**: `/health` and `/ready` endpoints on all services

### 5. Security

- **mTLS**: Mutual TLS between all gRPC services
- **Envelope Encryption**: AES-256-GCM + Vault KEK
- **Auth Interceptor**: x-agent-id, x-branch-id, x-region-id validation
- **Row-Level Security**: PostgreSQL RLS for branch isolation
- **Audit Logging**: Append-only audit trail with partition automation

---

## Quick Start

### Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Node.js | 18+ | Frontend builds |
| Python | 3.11+ | Local agent |
| Go | 1.23+ | Regional, Master, Projection workers |
| Docker & Compose | 24.0+ / 2.20+ | Infrastructure services |
| OpenSSL | 1.1+ | mTLS certificate generation |

### 1. Clone and Install

```bash
git clone https://github.com/Ahmedhajjajofficial/CPS-Dynamics-Commerce-System-AgenticOS.git
cd CPS-Dynamics-Commerce-System-AgenticOS

# Install everything at once
make install

# Or install components individually:
make install-app           # Admin frontend
make install-pos           # POS interface
make install-local-agent   # Python local agent
make install-regional-agent # Go regional agent
make install-master-agent  # Go master agent
make install-projections   # Go projection workers
```

### 2. Generate mTLS Certificates

```bash
cd cps-enterprise-dcs/infrastructure/tls
chmod +x generate-certs.sh
./generate-certs.sh
```

### 3. Build

```bash
# Build all components
make build

# Build individually
make build-app
make build-pos
make build-regional-agent
make build-master-agent
make build-projections
```

### 4. Run Development Servers

```bash
# Start the admin app (localhost:5173)
make dev-app

# Start the POS interface (localhost:3000)
make dev-pos

# Start both frontends
make dev
```

### 5. Run with Docker (Full Stack)

```bash
# Copy environment config
cp cps-enterprise-dcs/.env.example cps-enterprise-dcs/.env

# Generate TLS certificates
make generate-certs

# Start all infrastructure + services
make docker-up

# View logs
make docker-logs

# Stop everything
make docker-down
```

### Access Points (Docker)

| Service | URL | Credentials |
|---------|-----|-------------|
| POS Interface | http://localhost:3000 | Demo: any/any |
| Admin Dashboard | http://localhost:5173 | — |
| Grafana | http://localhost:3001 | admin/admin |
| Prometheus | http://localhost:9090 | — |
| Jaeger | http://localhost:16686 | — |
| Vault | http://localhost:8200 | dcs-dev-token |
| PostgreSQL | localhost:5432 | dcs_admin/dcs_secure_password |
| PgBouncer | localhost:6432 | dcs_admin/dcs_secure_password |
| Redis | localhost:6379 | — |
| Kafka | localhost:9092 | — |

---

## Development

### Available Make Targets

Run `make help` for a full list. Key targets:

```
make install              # Install all dependencies
make build                # Build all components
make dev                  # Start all dev servers
make lint                 # Lint all components
make test                 # Run all tests
make clean                # Remove build artifacts
make docker-up            # Start Docker infrastructure
make docker-down          # Stop Docker infrastructure
make generate-certs       # Generate mTLS certificates
```

### Linting

```bash
# Lint everything
make lint

# Lint individually
make lint-app              # ESLint for admin app
make lint-pos              # ESLint for POS interface
make lint-regional-agent   # go vet for regional agent
make lint-master-agent     # go vet for master agent
make lint-projections      # go vet for projection workers
```

### Testing

```bash
# Run all tests
make test

# Test individually
make test-local-agent      # pytest for Python agent
make test-regional-agent   # go test for Go agent
make test-master-agent     # go test for master agent
make test-projections      # go test for projection workers
```

### Code Style

- **TypeScript/React**: ESLint with react-hooks and react-refresh plugins
- **Python**: Type hints encouraged, async/await patterns
- **Go**: Standard `gofmt`, `go vet`

---

## Documentation

Detailed documentation is available in the [`docs/`](docs/) directory:

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System architecture, event sourcing, CRDTs, agent hierarchy |
| [Development](docs/development.md) | Developer setup, debugging, testing guide |
| [API Reference](docs/api.md) | gRPC services, Protocol Buffer schemas, event types |
| [Deployment](docs/deployment.md) | Docker, Kubernetes, production deployment guide |

### Protocol Buffers

The gRPC API is defined in [`cps-enterprise-dcs/proto/`](cps-enterprise-dcs/proto/). Key services:

- **AccountingSwarmProtocol** — Financial event broadcasting, reconciliation, conflict resolution
- **AgentCoordination** — Agent registration, heartbeat, CRDT synchronization
- **SagaOrchestration** — Distributed transaction management
- **GlobalReconciliationService** — Cross-region reconciliation (Master Agent)
- **DecisionEngineService** — ML-based decision evaluation (Master Agent)
- **MasterCoordinationService** — Regional agent registration and heartbeat
- **EventStreamingService** — Kafka event streaming integration

### Event Store Schema

The PostgreSQL schema is in [`cps-enterprise-dcs/event-store/schema.sql`](cps-enterprise-dcs/event-store/schema.sql). Features:

- Hash-partitioned event store (monthly partitions)
- Append-only enforcement via triggers
- Row-level security for multi-tenancy
- Saga orchestration tables
- CRDT state storage
- Read model projections (sales, inventory, loyalty, branch metrics)
- Master agent tables (reconciliation, decisions, regional agents, global events)

---

## Deployment

### Docker Compose (Development / Staging)

```bash
cd cps-enterprise-dcs
cp .env.example .env
# Edit .env with your configuration

# Generate TLS certificates
make generate-certs

# Start all services
docker-compose -f infrastructure/docker-compose.yml up -d
docker-compose -f infrastructure/docker-compose.yml ps
```

### Kubernetes (Production)

```bash
# Apply namespace and config
kubectl apply -f infrastructure/kubernetes/00-namespace-config.yaml

# Deploy infrastructure
kubectl apply -f infrastructure/kubernetes/01-postgres.yaml
kubectl apply -f infrastructure/kubernetes/02-redis.yaml

# Deploy agents
kubectl apply -f infrastructure/kubernetes/03-regional-agent.yaml
kubectl apply -f infrastructure/kubernetes/04-local-agent.yaml
kubectl apply -f infrastructure/kubernetes/05-master-agent.yaml
kubectl apply -f infrastructure/kubernetes/06-projection-workers.yaml
kubectl apply -f infrastructure/kubernetes/07-jaeger.yaml
```

### Production Considerations

- Replace Vault dev mode with production seal/unseal
- Configure TLS certificates for all services (use `infrastructure/tls/generate-certs.sh`)
- Set strong passwords in `.env` (especially `POSTGRES_PASSWORD`, `DCS_MASTER_KEY`)
- Enable `ENABLE_ENCRYPTION=true` and `ENABLE_AUDIT_LOG=true`
- Configure Kafka replication factor > 1
- Set up backup strategy for PostgreSQL event store (WAL archiving + pg_basebackup)
- Use PgBouncer for connection pooling
- Deploy Master Agent with replicas >= 2 for HA
- Deploy Projection Workers with replicas >= 3 for throughput
- Configure Jaeger with persistent storage (Elasticsearch/Cassandra)
- Set up Prometheus remote write for long-term metrics storage

See [docs/deployment.md](docs/deployment.md) for the full production checklist.

---

## Security

- [Security Policy](SECURITY.md) — Protocols for vulnerability disclosure and system integrity.

### Envelope Encryption

Every financial event is encrypted using a two-layer envelope encryption scheme:

1. **Data Encryption Key (DEK)** — Unique per event, AES-256-GCM
2. **Key Encryption Key (KEK)** — Managed by HashiCorp Vault Transit engine
3. **HMAC-SHA512** — Integrity verification on all visible metadata
4. **Digital Signatures** — ECDSA signatures by event creators

### mTLS Security

All gRPC communications between services are secured with mutual TLS:

- Internal CA for certificate issuance
- Per-service certificates with 1-year validity
- Client certificate verification in production
- Certificate rotation support

### Threat Model

DCS maintains security guarantees even when:
- Cloud provider infrastructure is compromised
- Network connectivity is completely offline
- Insider threats exist within the organization
- Targeted attacks occur against specific endpoints

---

## License

This project is licensed under the **CP'S Enterprise License**. See [LICENSE](LICENSE) for details.

> **Note**: This is proprietary software. Unauthorized distribution is prohibited.

---

## Contact

- **Author**: Ahmed Hajjaj — Full-Spectrum Architect
- **Email**: info.cpsfortechnology@gmail.com

---

<p align="center">
  <strong>CP'S Enterprise DCS</strong><br/>
  <em>Built with Logic of Sovereignty, Not Dependency</em>
</p>
