# دليل النشر

## خيارات النشر

| الطريقة | حالة الاستخدام | التعقيد |
|--------|----------|------------|
| Docker Compose | التطوير، البيئة التجريبية | منخفض |
| Kubernetes | الإنتاج، عدة مناطق | مرتفع |
| يدوي | عقدة واحدة، اختبار | متوسط |

---

## النشر باستخدام Docker Compose

### المتطلبات الأساسية

- Docker 24.0+
- Docker Compose 2.20+
- 8 جيجابايت RAM كحد أدنى
- 20 جيجابايت مساحة قرص حرة

### 1. تكوين البيئة

```bash
cd cps-enterprise-dcs
cp .env.example .env
```

عدّل `.env` وحدد قيماً آمنة لـ:

```bash
POSTGRES_PASSWORD=<strong-password>
DCS_MASTER_KEY=$(openssl rand -base64 32)
VAULT_TOKEN=<vault-token>
```

### 2. بدء الخدمات

```bash
# From the repo root
make docker-up

# Or directly:
docker-compose -f cps-enterprise-dcs/infrastructure/docker-compose.yml up -d
```

### 3. التحقق من السلامة

```bash
make docker-ps

# Check individual service logs
make docker-logs
```

### 4. الوصول إلى الخدمات

| الخدمة | الرابط | بيانات الاعتماد الافتراضية |
|---------|-----|---------------------|
| واجهة نقطة البيع | http://localhost:3000 | تجريبي: any/any |
| Grafana | http://localhost:3001 | admin/admin |
| Prometheus | http://localhost:9090 | — |
| Vault UI | http://localhost:8200 | الرمز من `.env` |
| PostgreSQL | localhost:5432 | من `.env` |

### الخدمات التي تم تشغيلها

تشغّل حزمة Docker Compose هذه الخدمات:

| الخدمة | الصورة | المنافذ |
|---------|-------|-------|
| PostgreSQL 16 | `postgres:16-alpine` | 5432 |
| Redis 7 | `redis:7-alpine` | 6379 |
| HashiCorp Vault | `hashicorp/vault:1.15` | 8200 |
| Zookeeper | `confluentinc/cp-zookeeper:7.5.0` | 2181 |
| Kafka | `confluentinc/cp-kafka:7.5.0` | 9092 |
| Prometheus | `prom/prometheus:v2.48.0` | 9090 |
| Grafana | `grafana/grafana:10.2.0` | 3001 |
| Regional Agent | بناء Go مخصص | 12000, 12001, 50052 |
| Local Agent | بناء Python مخصص | 50051 |
| POS Interface | بناء React مخصص | 3000 |
| Nginx | `nginx:alpine` | 80, 443 |

### إيقاف الخدمات

```bash
make docker-down

# Or to stop and remove volumes:
docker-compose -f cps-enterprise-dcs/infrastructure/docker-compose.yml down -v
```

---

## النشر اليدوي

### مخزن أحداث PostgreSQL

1. ثبّت PostgreSQL 16+
2. أنشئ قاعدة البيانات وطبّق المخطط:

```bash
createdb dcs_eventstore
psql -d dcs_eventstore -f cps-enterprise-dcs/event-store/schema.sql
```

### Redis

```bash
redis-server --appendonly yes --maxmemory 256mb --maxmemory-policy allkeys-lru
```

### الوكيل الإقليمي

```bash
cd cps-enterprise-dcs/regional-agent
go build -o bin/regional-agent ./...

# Configure via environment variables
export DCS_AGENT_ID=regional-001
export DCS_REGION_ID=region-001
export DCS_RPC_ADDR=:12000
export DCS_RAFT_ADDR=:12001
export DCS_GRPC_PORT=50052
export DCS_POSTGRESQL_URL=postgres://dcs_admin:password@localhost:5432/dcs_eventstore?sslmode=disable
export DCS_REDIS_URL=localhost:6379
export DCS_BOOTSTRAP=true

./bin/regional-agent
```

### الوكيل المحلي

```bash
cd cps-enterprise-dcs/local-agent
python3 -m venv .venv
source .venv/bin/activate
pip install -r requirements.txt

export DCS_AGENT_ID=local-001
export DCS_BRANCH_ID=BR001
export DCS_REGION_ID=region-001
export DCS_GRPC_PORT=50051
export DCS_REGIONAL_ENDPOINT=localhost:50052

python -m src.main
```

### واجهة نقطة البيع

```bash
cd cps-enterprise-dcs/pos-interface
npm install
npm run build

# Serve with any static file server
npx serve dist -l 3000
```

---

## قائمة التحقق للإنتاج

### الأمان

- [ ] استبدل وضع تطوير Vault بإعدادات إنتاج للختم/فك الختم
- [ ] أنشئ شهادات TLS واضبطها لجميع الخدمات
- [ ] حدّد كلمات مرور قوية لـ PostgreSQL (`POSTGRES_PASSWORD`)
- [ ] أنشئ مفتاحاً رئيسياً للإنتاج (`DCS_MASTER_KEY`)
- [ ] فعّل التشفير (`ENABLE_ENCRYPTION=true`)
- [ ] فعّل تسجيل التدقيق (`ENABLE_AUDIT_LOG=true`)
- [ ] اضبط سياسات الشبكة لتقييد الوصول بين الخدمات
- [ ] إعداد تدوير الشهادات

### قاعدة البيانات

- [ ] اضبط تجميع اتصالات PostgreSQL (PgBouncer)
- [ ] أعدّ نسخاً احتياطياً تلقائياً لمخزن الأحداث
- [ ] اضبط أرشفة WAL للاسترداد عند نقطة زمنية
- [ ] راجع واختبر استراتيجية تقسيم event_store
- [ ] أنشئ أقسام audit_log إضافية بعد الأشهر الأولى

### Kafka

- [ ] حدّد عامل النسخ > 1 لمواضيع الإنتاج
- [ ] اضبط سياسات الاحتفاظ لمواضيع الأحداث
- [ ] أعدّ مراقبة Kafka (تأخر المستهلك، صحة الأقسام)
- [ ] اضبط `KAFKA_ADVERTISED_LISTENERS` لشبكتك

### المراقبة

- [ ] اضبط أهداف Prometheus لجميع الوكلاء
- [ ] استورد لوحات Grafana لمقاييس DCS
- [ ] إعداد قواعد التنبيه (PagerDuty, Slack, البريد الإلكتروني)
- [ ] اضبط تجميع السجلات (ELK, Loki)
- [ ] إعداد التتبع الموزع (OpenTelemetry → Jaeger)

### التوفر العالي

- [ ] انشر 3 عقد أو أكثر من الوكيل الإقليمي لنصاب Raft
- [ ] اضبط النسخ المتدفق لـ PostgreSQL
- [ ] إعداد Redis Sentinel أو Redis Cluster
- [ ] انشر عدة وسطاء Kafka
- [ ] اضبط موازنة الأحمال لواجهات نقاط البيع

### الأداء

- [ ] اضبط `SYNC_INTERVAL` بناءً على ظروف الشبكة
- [ ] اضبط `BATCH_SIZE` لإنتاجية بث الأحداث
- [ ] حدّد `MAX_CONNECTIONS` بناءً على التزامن المتوقع
- [ ] اضبط `maxmemory` لـ Redis بناءً على الذاكرة المتاحة
- [ ] راجع `shared_buffers` و `work_mem` في PostgreSQL

---

## التوسع

### التوسع الأفقي

- **واجهة نقطة البيع**: عديمة الحالة — انشرها خلف موازن أحمال
- **الوكيل المحلي**: واحد لكل فرع — يتوسع مع عدد الفروع
- **الوكيل الإقليمي**: عنقود Raft — 3 عقد كحد أدنى لكل منطقة، يُفضّل العدد الفردي
- **PostgreSQL**: نسخ قراءة لتفريغ الاستعلامات؛ تقسيم لإنتاجية الكتابة

### تعدد المناطق

```
Region A                    Region B
┌──────────────────┐       ┌──────────────────┐
│ Regional Agent   │◄─────►│ Regional Agent   │
│ (Raft Cluster)   │ gRPC  │ (Raft Cluster)   │
├──────────────────┤       ├──────────────────┤
│ Local Agent x N  │       │ Local Agent x N  │
│ POS x N          │       │ POS x N          │
│ PostgreSQL       │       │ PostgreSQL       │
│ Redis            │       │ Redis            │
│ Kafka            │       │ Kafka            │
└──────────────────┘       └──────────────────┘
```

تشغّل كل منطقة مكدسها الكامل. يحدث التواصل بين المناطق على مستوى الوكيل الإقليمي عبر gRPC.

---

## النسخ الاحتياطي والاسترداد

### النسخ الاحتياطي لمخزن الأحداث

```bash
# Full backup
pg_dump -Fc dcs_eventstore > dcs_eventstore_$(date +%Y%m%d).dump

# Restore
pg_restore -d dcs_eventstore dcs_eventstore_20240101.dump
```

### النسخ الاحتياطي لوحدات تخزين Docker

```bash
# Backup PostgreSQL data
docker run --rm -v dcs_postgres_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/postgres_data.tar.gz -C /data .

# Backup Redis data
docker run --rm -v dcs_redis_data:/data -v $(pwd):/backup \
  alpine tar czf /backup/redis_data.tar.gz -C /data .
```

### استراتيجية الاسترداد

نظراً لاستخدام DCS لمصدر الأحداث، فإن مخزن الأحداث هو الهدف الرئيسي للاسترداد. يمكن إعادة بناء جميع الحالات الأخرى (الإسقاطات، حالات CRDT، حالات Saga) بإعادة تشغيل الأحداث من مخزن الأحداث.

1. استعد مخزن أحداث PostgreSQL من النسخة الاحتياطية
2. أعد تشغيل جميع الوكلاء — سيعيدون بناء الحالة المحلية من مخزن الأحداث
3. ستتقارب حالات CRDT تلقائياً مع مزامنة الوكلاء
