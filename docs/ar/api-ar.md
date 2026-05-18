# مرجع واجهة البرمجة (API)

## نظرة عامة

يستخدم نظام CPS Enterprise DCS تقنية **gRPC** مع **Protocol Buffers** لجميع الاتصالات بين الخدمات. تعريفات الخدمات موجودة في `cps-enterprise-dcs/proto/cps_enterprise_v4.proto`.

{/* TODO: Restore link to the proto file once its canonical location is confirmed (e.g. a GitHub permalink in the AgenticOS repo). The previous relative path `../cps-enterprise-dcs/proto/cps_enterprise_v4.proto` resolved outside the docs tree and 404'd. */}

---

## خدمات gRPC

### AccountingSwarmProtocol

الخدمة الأساسية لإدارة الأحداث المالية والتسوية وتنسيق الوكلاء.

**المنفذ:** `50051` (الوكيل المحلي)، `50052` (الوكيل الإقليمي)

#### استدعاءات Unary RPC

| الطريقة | الطلب | الاستجابة | الوصف |
|--------|---------|----------|-------------|
| `BroadcastFinancialEvent` | `SovereignFinancialEvent` | `AckResponse` | إرسال حدث مالي واحد |
| `GetEventById` | `GetEventRequest` | `SovereignFinancialEvent` | استرجاع حدث بواسطة المعرّف |
| `GetEventsByRange` | `GetEventsRangeRequest` | `EventBatch` | استعلام الأحداث ضمن نطاق زمني مع التقسيم إلى صفحات |
| `RequestReconciliation` | `ReconciliationRequest` | `ReconciliationResponse` | تشغيل عملية التسوية لفرع معين |
| `ResolveLedgerConflict` | `ConflictResolutionRequest` | `SovereignFinancialEvent` | حل تعارض الأحداث |

#### استدعاءات Server Streaming RPC

| الطريقة | الطلب | تدفق الاستجابة | الوصف |
|--------|---------|-----------------|-------------|
| `SubscribeEvents` | `SubscribeRequest` | `SovereignFinancialEvent` | الاشتراك في تدفق الأحداث الفوري |
| `StreamReconciliation` | `ReconciliationRequest` | `ReconciliationResponse` | بث نتائج التسوية |

#### استدعاءات Client Streaming RPC

| الطريقة | تدفق الطلب | الاستجابة | الوصف |
|--------|---------------|----------|-------------|
| `StreamOfflineEvents` | `SovereignFinancialEvent` | `BatchAckResponse` | مزامنة دفعات أحداث العمل دون اتصال |
| `StreamCRDTUpdates` | `CRDTStateBundle` | `BatchAckResponse` | بث تحديثات حالة CRDT |

#### استدعاءات Bidirectional Streaming RPC

| الطريقة | تدفق الطلب | تدفق الاستجابة | الوصف |
|--------|---------------|-----------------|-------------|
| `SwarmEventExchange` | `SovereignFinancialEvent` | `SovereignFinancialEvent` | تبادل الأحداث الفوري بين الوكلاء |
| `SyncCRDTState` | `CRDTStateBundle` | `CRDTStateBundle` | مزامنة CRDT ثنائية الاتجاه |
| `AgentCommunication` | `AgentMessage` | `AgentMessage` | تواصل عام بين الوكلاء |

---

### QueryProtocol

خدمة استعلام نموذج القراءة للوحات التحكم والتقارير.

#### استدعاءات Unary RPC

| الطريقة | الطلب | الاستجابة | الوصف |
|--------|---------|----------|-------------|
| `GetBranchSummary` | `BranchQuery` | `BranchSummary` | الحصول على ملخص مبيعات/جلسات الفرع |
| `GetInventoryStatus` | `InventoryQuery` | `InventoryStatus` | الحصول على مستويات مخزون المنتجات |
| `GetSalesReport` | `SalesReportQuery` | `SalesReport` | الحصول على تقرير المبيعات حسب الفترة |
| `GetLedgerBalance` | `LedgerQuery` | `LedgerBalance` | الحصول على رصيد الحساب |

#### استدعاءات Server Streaming RPC

| الطريقة | الطلب | تدفق الاستجابة | الوصف |
|--------|---------|-----------------|-------------|
| `SubscribeDashboard` | `DashboardSubscription` | `DashboardUpdate` | مقاييس لوحة التحكم الفورية |

---

## الرسائل الأساسية

### SovereignFinancialEvent

النوع المركزي للأحداث في النظام. تُسجَّل كل معاملة مالية كحدث غير قابل للتعديل.

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

حمولة مشفّرة بأسلوب المغلفات مرفقة بكل حدث.

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

يتتبع السببية والترتيب عبر العقد الموزعة.

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

## أنواع الأحداث

تُنظَّم الأحداث في تدفقات خاصة بالنطاق:

### تدفق المبيعات
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `SALE_INITIATED` | 1 | بدء معاملة بيع |
| `SALE_COMPLETED` | 2 | اكتمال البيع بنجاح |
| `SALE_REVERSED` | 3 | إلغاء/عكس البيع بالكامل |
| `SALE_PARTIALLY_REFUNDED` | 4 | معالجة استرداد جزئي |

### تدفق المخزون
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `INVENTORY_RECEIVED` | 10 | استلام مخزون من المورد |
| `INVENTORY_RESERVED` | 11 | حجز مخزون لبيع معلق |
| `INVENTORY_RESERVATION_CANCELLED` | 12 | إلغاء الحجز |
| `INVENTORY_SOLD` | 13 | بيع مخزون للعميل |
| `INVENTORY_TRANSFERRED_OUT` | 14 | نقل مخزون من الفرع |
| `INVENTORY_TRANSFERRED_IN` | 15 | نقل مخزون إلى الفرع |
| `INVENTORY_ADJUSTMENT` | 16 | تعديل مخزون يدوي |
| `INVENTORY_COUNTED` | 17 | تسجيل جرد فعلي |
| `LOW_STOCK_ALERT` | 18 | المخزون أقل من نقطة إعادة الطلب |
| `STOCKOUT_DETECTED` | 19 | المخزون عند الصفر |

### تدفق المدفوعات
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `PAYMENT_AUTHORIZED` | 30 | تمت الموافقة على الدفع |
| `PAYMENT_CAPTURED` | 31 | تم التحصيل/التسوية |
| `PAYMENT_REFUNDED` | 32 | تم استرداد المبلغ |
| `PAYMENT_FAILED` | 33 | فشل الدفع |
| `PAYMENT_CHARGEBACK` | 34 | استلام رد مبلغ (Chargeback) |

### تدفق المحاسبة
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `LEDGER_ENTRY_POSTED` | 20 | تم إنشاء قيد دفتر الأستاذ |
| `LEDGER_REVERSAL_POSTED` | 21 | عكس قيد دفتر الأستاذ |
| `END_OF_DAY_RECONCILIATION` | 22 | تسوية نهاية اليوم |
| `PERIOD_CLOSED` | 23 | إغلاق فترة محاسبية |
| `TRIAL_BALANCE_GENERATED` | 24 | توليد ميزان المراجعة |

### تدفق Saga
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `SAGA_INITIATED` | 60 | بدء معاملة موزعة |
| `SAGA_STEP_COMPLETED` | 61 | اكتمال خطوة Saga |
| `SAGA_COMPENSATION_TRIGGERED` | 62 | بدء التعويض (التراجع) |
| `SAGA_COMPENSATION_COMPLETED` | 63 | اكتمال التعويض |
| `SAGA_FAILED` | 64 | فشل Saga |
| `TRANSACTION_REJECTED` | 65 | رفض المعاملة |

### تدفق الوكيل
| النوع | القيمة | الوصف |
|------|-------|-------------|
| `AGENT_DECISION_MADE` | 70 | اتخذ الوكيل قراراً ذاتياً |
| `AGENT_ACTION_EXECUTED` | 71 | نفذ الوكيل إجراءً |
| `AGENT_LEARNING_UPDATED` | 72 | تحديث نموذج الوكيل |

---

## رسائل CRDT

### GCounter (عدّاد تزايد فقط)
```protobuf
message GCounter {
  string node_id = 1;
  map<string, uint64> increments = 2;
}
```

### PNCounter (عدّاد موجب-سالب)
```protobuf
message PNCounter {
  string node_id = 1;
  map<string, uint64> increments = 2;
  map<string, uint64> decrements = 3;
}
```

### ORSet (مجموعة الإزالة المُلاحظة)
```protobuf
message ORSet {
  string node_id = 1;
  repeated ORSetElement elements = 2;
}
```

### LWWRegister (سجل آخر كتابة تفوز)
```protobuf
message LWWRegister {
  string value = 1;
  HybridLogicalClock timestamp = 2;
  string node_id = 3;
}
```

### CRDTStateBundle (مغلف المزامنة)
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

## تنسيق Saga

### أنواع Saga

| النوع | الوصف |
|------|-------------|
| `STANDARD_SALE` | معاملة بيع اعتيادية في نقطة البيع |
| `REFUND_PROCESSING` | استرداد مع إرجاع المخزون |
| `INVENTORY_TRANSFER` | نقل مخزون بين الفروع |
| `BATCH_OPERATION` | عملية مجمّعة (مثل تحديث الأسعار) |
| `CROSS_BRANCH_SALE` | بيع يشمل عدة فروع |

### حالات Saga

```
PENDING → RUNNING → COMPLETED
                  ↘ COMPENSATING → COMPENSATED
                  ↘ FAILED
                  ↘ TIMED_OUT
```

---

## استراتيجيات حل التعارض

| الاستراتيجية | الوصف |
|----------|-------------|
| `LAST_WRITER_WINS` | يفوز أحدث طابع زمني |
| `VECTOR_CLOCK_ORDER` | ترتيب سببي عبر الساعات المتجهية |
| `AGENT_VOTE` | تصويت أغلبية بين الوكلاء |
| `MASTER_DECIDES` | الوكيل الرئيسي يحسم القرار |
| `CRDT_MERGE` | دمج CRDT تلقائي |
| `CUSTOM_RESOLUTION` | منطق خاص بالتطبيق |

---

## أنواع الوكلاء

| النوع | الوصف |
|------|-------------|
| `LOCAL_AGENT` | وكيل حافة على مستوى الفرع (Python) |
| `REGIONAL_AGENT` | وكيل تنسيق إقليمي (Go) |
| `MASTER_AGENT` | وكيل التنسيق العالمي |
| `PREDICTIVE_AGENT` | وكيل تنبؤ مبني على تعلم الآلة |
| `COMPLIANCE_AGENT` | وكيل الامتثال التنظيمي |
| `SECURITY_AGENT` | وكيل مراقبة أمنية |
