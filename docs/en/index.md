---
icon: material/home
---

# CPS Enterprise Dynamics Commerce System (DCS)

<div class="hero-section" markdown>

## An Intelligent Distributed Event-Driven Commerce System

**CPS Enterprise DCS** is a distributed commerce platform built on an Agentic Architecture with full offline-first support.

</div>

---

## Key Features

<div class="grid cards" markdown>

-   :material-source-branch:{ .lg .middle } **Event Sourcing**

    ---

    All state changes are captured as immutable events. The single source of truth for the entire system.

-   :material-robot:{ .lg .middle } **Autonomous Intelligent Agents**

    ---

    Three-tier agent architecture: Local (Python), Regional (Go/Raft), and Master for global orchestration.

-   :material-sync:{ .lg .middle } **Conflict-Free CRDTs**

    ---

    Multiple data types (GCounter, PNCounter, ORSet, LWWRegister) ensure automatic convergence.

-   :material-shield-lock:{ .lg .middle } **Multi-Layer Encryption**

    ---

    Envelope encryption (AES-256-GCM) with key management via HashiCorp Vault.

-   :material-wifi-off:{ .lg .middle } **Offline-First Operation**

    ---

    Each branch operates fully independently and auto-syncs when connectivity is restored.

-   :material-chart-timeline-variant:{ .lg .middle } **Comprehensive Monitoring**

    ---

    Prometheus + Grafana for real-time monitoring with custom dashboards.

</div>

---

## Technical Architecture

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

## Technical Stack

| Component | Technology | Description |
|-----------|-----------|-------------|
| POS Interface | React 18 + TypeScript | Cashier interface with offline support |
| Admin Dashboard | React 19 + Vite 7 | Admin control panel with shadcn/ui |
| Local Agent | Python 3.11+ | Edge agent running at each branch |
| Regional Agent | Go 1.21+ | Regional coordination with Raft consensus |
| Database | PostgreSQL 16 | Partitioned event store (8 partitions) |
| Cache | Redis 7 | Caching and session management |
| Messaging | Apache Kafka | Cross-region event streaming |
| Security | HashiCorp Vault | Key management and encryption |
| Monitoring | Prometheus + Grafana | Performance monitoring and alerting |

---

## Quick Start

```bash
# Clone the repository
git clone https://gitlab.com/bumble-ideas-projects/niamimniaCmmPS-Enterprise-Dynamics-Commerce-System-DCS.git
cd CPS-Enterprise-Dynamics-Commerce-System-DCS

# Install all dependencies
make install

# Start infrastructure
make docker-up

# Run development servers
make dev
```

!!! tip "Tip"
    See the [Development Guide](development.md) for full setup instructions.

---

## Quick Links

- [Architecture](architecture.md) - Detailed technical architecture
- [API Reference](api.md) - Complete gRPC API reference
- [Development](development.md) - Development environment setup guide
- [Deployment](deployment.md) - Deployment and production options
