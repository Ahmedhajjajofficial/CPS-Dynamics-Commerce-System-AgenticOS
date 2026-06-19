# Development Guide

## Prerequisites

| Tool | Version | Installation |
|------|---------|-------------|
| Node.js | 18+ (22 recommended) | [nodejs.org](https://nodejs.org/) |
| Python | 3.11+ | [python.org](https://www.python.org/) |
| Go | 1.21+ | [go.dev](https://go.dev/) |
| Docker & Compose | 24.0+ / 2.20+ | [docker.com](https://www.docker.com/) |
| Make | Any | Pre-installed on macOS/Linux |

---

## Initial Setup

### 1. Clone the Repository

```bash
git clone https://github.com/Ahmedhajjajofficial/CPS-Dynamics-Commerce-System-AgenticOS.git
cd CPS-Dynamics-Commerce-System-AgenticOS
```

### 2. Install All Dependencies

```bash
make install
```

This runs the following under the hood:

```bash
# Admin app (React 19 + Vite)
cd app && npm install

# POS interface (React 18 + Vite)
cd cps-enterprise-dcs/pos-interface && npm install

# Local agent (Python)
python3 -m venv cps-enterprise-dcs/local-agent/.venv
source cps-enterprise-dcs/local-agent/.venv/bin/activate
pip install -r cps-enterprise-dcs/local-agent/requirements.txt

# Regional agent (Go)
cd cps-enterprise-dcs/regional-agent && go mod download
```

### 3. Build All Components

```bash
make build
```

---

## Running Development Servers

### Frontend Dev Servers

```bash
# Admin app — http://localhost:5173
make dev-app

# POS interface — http://localhost:3000
make dev-pos

# Both simultaneously
make dev
```

### Backend Agents (requires Docker infrastructure)

The Local Agent and Regional Agent require PostgreSQL, Redis, Kafka, and Vault. Start them via Docker:

```bash
# Copy environment template
cp cps-enterprise-dcs/.env.example cps-enterprise-dcs/.env

# Start infrastructure
make docker-up

# Verify services are healthy
make docker-ps
```

---

## Project Components

### Admin App (`app/`)

- **Framework**: React 19 + TypeScript + Vite 7
- **UI Library**: shadcn/ui (Radix UI primitives) + Tailwind CSS
- **Forms**: React Hook Form + Zod validation
- **Charts**: Recharts
- **Module System**: ESM (`"type": "module"`)

```bash
cd app
npm run dev       # Start dev server
npm run build     # Production build (tsc -b && vite build)
npm run lint      # ESLint check
npm run preview   # Preview production build
```

### POS Interface (`cps-enterprise-dcs/pos-interface/`)

- **Framework**: React 18 + TypeScript + Vite
- **State**: Zustand
- **Styling**: Tailwind CSS + Lucide icons
- **Charts**: Recharts
- **Package name**: `rockdeals-pos` v4.0.0

```bash
cd cps-enterprise-dcs/pos-interface
npm run dev       # Start dev server (:3000)
npm run build     # Production build (tsc && vite build)
npm run lint      # ESLint check
```

### Local Agent (`cps-enterprise-dcs/local-agent/`)

- **Language**: Python 3.11+
- **RPC**: gRPC (grpcio, grpcio-tools, protobuf)
- **Database**: SQLAlchemy + aiosqlite (local SQLite), asyncpg (PostgreSQL)
- **Messaging**: aiokafka, redis (with hiredis)
- **ML**: scikit-learn, onnxruntime
- **Observability**: prometheus-client, opentelemetry

```bash
# Activate venv
source cps-enterprise-dcs/local-agent/.venv/bin/activate

# Run tests
pytest src/ -v

# Type check a module
python -c "import py_compile; py_compile.compile('src/agent.py')"
```

### Regional Agent (`cps-enterprise-dcs/regional-agent/`)

- **Language**: Go 1.21+
- **Consensus**: HashiCorp Raft + BoltDB
- **RPC**: gRPC (google.golang.org/grpc)
- **Logging**: go.uber.org/zap

```bash
cd cps-enterprise-dcs/regional-agent

go build ./...       # Build
go test ./... -v     # Run tests
go vet ./...         # Lint
```

---

## Linting

```bash
# Lint everything
make lint

# Individual components
make lint-app              # ESLint on admin app
make lint-pos              # ESLint on POS interface
make lint-regional-agent   # go vet on regional agent
```

### ESLint Configuration

Both frontends use ESLint 9+ with flat config (`eslint.config.js`):
- `eslint-plugin-react-hooks` — React hooks rules
- `eslint-plugin-react-refresh` — Fast refresh validation
- TypeScript-aware linting via `typescript-eslint`

---

## Testing

```bash
# Run all tests
make test

# Individual components
make test-local-agent      # pytest
make test-regional-agent   # go test
```

### Python Tests

The local agent uses `pytest` with `pytest-asyncio` for async test support:

```bash
source cps-enterprise-dcs/local-agent/.venv/bin/activate
pytest src/ -v --tb=short --cov=src
```

### Go Tests

```bash
cd cps-enterprise-dcs/regional-agent
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out   # View coverage report
```

---

## Cleaning Up

```bash
# Remove build artifacts (dist/, bin/, __pycache__)
make clean

# Remove everything including node_modules and venv
make clean-all
```

---

## Protobuf Code Generation

If you modify `cps-enterprise-dcs/proto/cps_enterprise_v4.proto`, regenerate the bindings:

```bash
make proto
```

This requires `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`, and `grpcio-tools` (Python) to be installed.

---

## Environment Variables

Copy the template and configure:

```bash
cp cps-enterprise-dcs/.env.example cps-enterprise-dcs/.env
```

Key variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_USER` | `dcs_admin` | PostgreSQL username |
| `POSTGRES_PASSWORD` | (change this) | PostgreSQL password |
| `POSTGRES_DB` | `dcs_eventstore` | Database name |
| `REDIS_PORT` | `6379` | Redis port |
| `VAULT_TOKEN` | `dcs-dev-token` | Vault dev token |
| `KAFKA_PORT` | `9092` | Kafka broker port |
| `REGIONAL_AGENT_ID` | `regional-001` | Regional agent identifier |
| `LOCAL_AGENT_ID` | `local-001` | Local agent identifier |
| `BRANCH_ID` | `BR001` | Branch identifier |
| `POS_PORT` | `3000` | POS interface port |
| `DCS_MASTER_KEY` | (generate) | Master encryption key (base64) |
| `ENABLE_ENCRYPTION` | `true` | Enable envelope encryption |
| `ENABLE_AUDIT_LOG` | `true` | Enable audit logging |
| `SYNC_INTERVAL` | `30` | CRDT sync interval in seconds |

Generate a master key:

```bash
openssl rand -base64 32
```

---

## Troubleshooting

### Vite Build Fails with CommonJS Error

The frontends use ESM (`"type": "module"`). Ensure config files use `export default` syntax instead of `module.exports`.

### Go Build Fails with Missing Dependencies

```bash
cd cps-enterprise-dcs/regional-agent
go mod tidy
go mod download
```

### Python Import Errors

Ensure you're using the project venv:

```bash
source cps-enterprise-dcs/local-agent/.venv/bin/activate
which python  # Should point to .venv/bin/python
```

### Docker Services Not Starting

```bash
# Check service status and logs
make docker-ps
make docker-logs

# Restart with rebuild
make docker-rebuild
```

### Port Conflicts

Default ports used:
- `3000` — POS Interface
- `5173` — Admin App (Vite dev)
- `5432` — PostgreSQL
- `6379` — Redis
- `8200` — Vault
- `9090` — Prometheus
- `9092` — Kafka
- `3001` — Grafana
- `50051` — Local Agent gRPC
- `50052` — Regional Agent gRPC
- `12000` — Regional Agent RPC
- `12001` — Regional Agent Raft

Change ports via environment variables in `.env`.
