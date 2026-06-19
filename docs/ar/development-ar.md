# دليل التطوير

## المتطلبات الأساسية

| الأداة | الإصدار | التثبيت |
|------|---------|-------------|
| Node.js | 18+ (يُوصى بـ 22) | [nodejs.org](https://nodejs.org/) |
| Python | 3.11+ | [python.org](https://www.python.org/) |
| Go | 1.21+ | [go.dev](https://go.dev/) |
| Docker و Compose | 24.0+ / 2.20+ | [docker.com](https://www.docker.com/) |
| Make | أي إصدار | مثبّت مسبقاً على macOS/Linux |

---

## الإعداد الأولي

### 1. استنساخ المستودع

```bash
git clone https://github.com/Ahmedhajjajofficial/CPS-Enterprise-Dynamics-Commerce-System-DCS.git
cd CPS-Enterprise-Dynamics-Commerce-System-DCS
```

### 2. تثبيت جميع التبعيات

```bash
make install
```

يقوم هذا بتشغيل ما يلي خلف الكواليس:

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

### 3. بناء جميع المكونات

```bash
make build
```

---

## تشغيل خوادم التطوير

### خوادم تطوير الواجهة الأمامية

```bash
# Admin app — http://localhost:5173
make dev-app

# POS interface — http://localhost:3000
make dev-pos

# Both simultaneously
make dev
```

### الوكلاء الخلفيون (يتطلب بنية Docker التحتية)

يتطلب الوكيل المحلي والوكيل الإقليمي PostgreSQL و Redis و Kafka و Vault. ابدأها عبر Docker:

```bash
# Copy environment template
cp cps-enterprise-dcs/.env.example cps-enterprise-dcs/.env

# Start infrastructure
make docker-up

# Verify services are healthy
make docker-ps
```

---

## مكونات المشروع

### تطبيق الإدارة (`app/`)

- **الإطار**: React 19 + TypeScript + Vite 7
- **مكتبة الواجهة**: shadcn/ui (عناصر Radix UI الأولية) + Tailwind CSS
- **النماذج**: React Hook Form + التحقق عبر Zod
- **الرسوم البيانية**: Recharts
- **نظام الوحدات**: ESM (`"type": "module"`)

```bash
cd app
npm run dev       # Start dev server
npm run build     # Production build (tsc -b && vite build)
npm run lint      # ESLint check
npm run preview   # Preview production build
```

### واجهة نقطة البيع (`cps-enterprise-dcs/pos-interface/`)

- **الإطار**: React 18 + TypeScript + Vite
- **الحالة**: Zustand
- **التنسيق**: Tailwind CSS + أيقونات Lucide
- **الرسوم البيانية**: Recharts
- **اسم الحزمة**: `rockdeals-pos` v4.0.0

```bash
cd cps-enterprise-dcs/pos-interface
npm run dev       # Start dev server (:3000)
npm run build     # Production build (tsc && vite build)
npm run lint      # ESLint check
```

### الوكيل المحلي (`cps-enterprise-dcs/local-agent/`)

- **اللغة**: Python 3.11+
- **RPC**: gRPC (grpcio, grpcio-tools, protobuf)
- **قاعدة البيانات**: SQLAlchemy + aiosqlite (SQLite محلي)، asyncpg (PostgreSQL)
- **الرسائل**: aiokafka، redis (مع hiredis)
- **تعلم الآلة**: scikit-learn، onnxruntime
- **المراقبة**: prometheus-client، opentelemetry

```bash
# Activate venv
source cps-enterprise-dcs/local-agent/.venv/bin/activate

# Run tests
pytest src/ -v

# Type check a module
python -c "import py_compile; py_compile.compile('src/agent.py')"
```

### الوكيل الإقليمي (`cps-enterprise-dcs/regional-agent/`)

- **اللغة**: Go 1.21+
- **الإجماع**: HashiCorp Raft + BoltDB
- **RPC**: gRPC (google.golang.org/grpc)
- **التسجيل**: go.uber.org/zap

```bash
cd cps-enterprise-dcs/regional-agent

go build ./...       # Build
go test ./... -v     # Run tests
go vet ./...         # Lint
```

---

## فحص الكود (Linting)

```bash
# Lint everything
make lint

# Individual components
make lint-app              # ESLint on admin app
make lint-pos              # ESLint on POS interface
make lint-regional-agent   # go vet on regional agent
```

### إعدادات ESLint

تستخدم كلتا الواجهتين الأماميتين ESLint 9+ مع التكوين المسطّح (`eslint.config.js`):
- `eslint-plugin-react-hooks` — قواعد خطافات React
- `eslint-plugin-react-refresh` — التحقق من التحديث السريع
- فحص مدرك لـ TypeScript عبر `typescript-eslint`

---

## الاختبار

```bash
# Run all tests
make test

# Individual components
make test-local-agent      # pytest
make test-regional-agent   # go test
```

### اختبارات Python

يستخدم الوكيل المحلي `pytest` مع `pytest-asyncio` لدعم الاختبارات غير المتزامنة:

```bash
source cps-enterprise-dcs/local-agent/.venv/bin/activate
pytest src/ -v --tb=short --cov=src
```

### اختبارات Go

```bash
cd cps-enterprise-dcs/regional-agent
go test ./... -v -race -coverprofile=coverage.out
go tool cover -html=coverage.out   # View coverage report
```

---

## التنظيف

```bash
# Remove build artifacts (dist/, bin/, __pycache__)
make clean

# Remove everything including node_modules and venv
make clean-all
```

---

## توليد كود Protobuf

إذا عدّلت `cps-enterprise-dcs/proto/cps_enterprise_v4.proto`، أعد توليد الروابط:

```bash
make proto
```

يتطلب هذا تثبيت `protoc` و `protoc-gen-go` و `protoc-gen-go-grpc` و `grpcio-tools` (Python).

---

## متغيرات البيئة

انسخ القالب واضبط الإعدادات:

```bash
cp cps-enterprise-dcs/.env.example cps-enterprise-dcs/.env
```

المتغيرات الرئيسية:

| المتغير | الافتراضي | الوصف |
|----------|---------|-------------|
| `POSTGRES_USER` | `dcs_admin` | اسم مستخدم PostgreSQL |
| `POSTGRES_PASSWORD` | (غيّر هذه) | كلمة مرور PostgreSQL |
| `POSTGRES_DB` | `dcs_eventstore` | اسم قاعدة البيانات |
| `REDIS_PORT` | `6379` | منفذ Redis |
| `VAULT_TOKEN` | `dcs-dev-token` | رمز تطوير Vault |
| `KAFKA_PORT` | `9092` | منفذ وسيط Kafka |
| `REGIONAL_AGENT_ID` | `regional-001` | معرّف الوكيل الإقليمي |
| `LOCAL_AGENT_ID` | `local-001` | معرّف الوكيل المحلي |
| `BRANCH_ID` | `BR001` | معرّف الفرع |
| `POS_PORT` | `3000` | منفذ واجهة نقطة البيع |
| `DCS_MASTER_KEY` | (ولّده) | المفتاح الرئيسي للتشفير (base64) |
| `ENABLE_ENCRYPTION` | `true` | تفعيل تشفير المغلفات |
| `ENABLE_AUDIT_LOG` | `true` | تفعيل تسجيل التدقيق |
| `SYNC_INTERVAL` | `30` | فاصل مزامنة CRDT بالثواني |

ولّد مفتاحاً رئيسياً:

```bash
openssl rand -base64 32
```

---

## استكشاف الأخطاء وإصلاحها

### فشل بناء Vite بسبب خطأ CommonJS

تستخدم الواجهات الأمامية ESM (`"type": "module"`). تأكد من استخدام صيغة `export default` في ملفات التكوين بدلاً من `module.exports`.

### فشل بناء Go بسبب تبعيات مفقودة

```bash
cd cps-enterprise-dcs/regional-agent
go mod tidy
go mod download
```

### أخطاء استيراد Python

تأكد من استخدام venv المشروع:

```bash
source cps-enterprise-dcs/local-agent/.venv/bin/activate
which python  # Should point to .venv/bin/python
```

### عدم بدء خدمات Docker

```bash
# Check service status and logs
make docker-ps
make docker-logs

# Restart with rebuild
make docker-rebuild
```

### تعارض المنافذ

المنافذ الافتراضية المستخدمة:
- `3000` — واجهة نقطة البيع
- `5173` — تطبيق الإدارة (Vite dev)
- `5432` — PostgreSQL
- `6379` — Redis
- `8200` — Vault
- `9090` — Prometheus
- `9092` — Kafka
- `3001` — Grafana
- `50051` — gRPC الوكيل المحلي
- `50052` — gRPC الوكيل الإقليمي
- `12000` — RPC الوكيل الإقليمي
- `12001` — Raft الوكيل الإقليمي

غيّر المنافذ عبر متغيرات البيئة في `.env`.
