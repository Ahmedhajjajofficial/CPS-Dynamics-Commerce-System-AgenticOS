-- ═══════════════════════════════════════════════════════════════════════════════
-- CP'S Enterprise DCS - Event Store Schema
-- PostgreSQL Database for Event Sourcing
-- ═══════════════════════════════════════════════════════════════════════════════
-- Author: Ahmed Hajjaj - Full-Spectrum Architect
-- Version: 4.0.0
-- ═══════════════════════════════════════════════════════════════════════════════

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ═══════════════════════════════════════════════════════════════════════════════
-- CORE EVENT STORE
-- ═══════════════════════════════════════════════════════════════════════════════

-- Main event store table (partitioned by stream_id hash)
CREATE TABLE IF NOT EXISTS event_store (
    stream_id UUID NOT NULL,
    version BIGINT NOT NULL,
    event_id UUID NOT NULL DEFAULT gen_random_uuid(),
    event_type TEXT NOT NULL,
    payload BYTEA NOT NULL,
    metadata JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    event_hash TEXT,
    
    PRIMARY KEY (stream_id, version),
    UNIQUE (event_id)
) PARTITION BY HASH (stream_id);

-- Create partitions for event_store
CREATE TABLE IF NOT EXISTS event_store_p0 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 0);
CREATE TABLE IF NOT EXISTS event_store_p1 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 1);
CREATE TABLE IF NOT EXISTS event_store_p2 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 2);
CREATE TABLE IF NOT EXISTS event_store_p3 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 3);
CREATE TABLE IF NOT EXISTS event_store_p4 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 4);
CREATE TABLE IF NOT EXISTS event_store_p5 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 5);
CREATE TABLE IF NOT EXISTS event_store_p6 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 6);
CREATE TABLE IF NOT EXISTS event_store_p7 PARTITION OF event_store
    FOR VALUES WITH (MODULUS 8, REMAINDER 7);

-- Indexes for event_store
CREATE INDEX IF NOT EXISTS idx_event_store_event_type 
    ON event_store(event_type);
CREATE INDEX IF NOT EXISTS idx_event_store_created_at 
    ON event_store(created_at);
CREATE INDEX IF NOT EXISTS idx_event_store_metadata 
    ON event_store USING GIN(metadata);

-- ═══════════════════════════════════════════════════════════════════════════════
-- STREAM METADATA
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS stream_metadata (
    stream_id UUID PRIMARY KEY,
    stream_type TEXT NOT NULL,
    current_version BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    branch_id TEXT,
    tenant_id TEXT,
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_stream_metadata_branch 
    ON stream_metadata(branch_id);
CREATE INDEX IF NOT EXISTS idx_stream_metadata_tenant 
    ON stream_metadata(tenant_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- SNAPSHOTS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS snapshots (
    stream_id UUID PRIMARY KEY,
    version BIGINT NOT NULL,
    snapshot_data BYTEA NOT NULL,
    snapshot_type TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_snapshots_created_at 
    ON snapshots(created_at);

-- ═══════════════════════════════════════════════════════════════════════════════
-- SAGA ORCHESTRATION
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS saga_instances (
    saga_id UUID PRIMARY KEY,
    correlation_id UUID NOT NULL UNIQUE,
    saga_type TEXT NOT NULL,
    current_state TEXT NOT NULL,
    context JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    timeout_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    
    CONSTRAINT valid_state CHECK (current_state IN (
        'PENDING', 'RUNNING', 'COMPLETED', 'COMPENSATING', 
        'COMPENSATED', 'FAILED', 'TIMED_OUT'
    ))
);

CREATE INDEX IF NOT EXISTS idx_saga_correlation 
    ON saga_instances(correlation_id);
CREATE INDEX IF NOT EXISTS idx_saga_state 
    ON saga_instances(current_state);
CREATE INDEX IF NOT EXISTS idx_saga_timeout 
    ON saga_instances(timeout_at) 
    WHERE timeout_at IS NOT NULL;

-- Saga steps
CREATE TABLE IF NOT EXISTS saga_steps (
    step_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    saga_id UUID NOT NULL REFERENCES saga_instances(saga_id) ON DELETE CASCADE,
    step_name TEXT NOT NULL,
    step_order INTEGER NOT NULL,
    step_data JSONB DEFAULT '{}',
    state TEXT NOT NULL DEFAULT 'PENDING',
    result JSONB,
    error_message TEXT,
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    compensation_step_id UUID,
    
    CONSTRAINT valid_step_state CHECK (state IN (
        'PENDING', 'EXECUTING', 'COMPLETED', 'FAILED', 
        'COMPENSATING', 'COMPENSATED'
    ))
);

CREATE INDEX IF NOT EXISTS idx_saga_steps_saga 
    ON saga_steps(saga_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- PROJECTIONS (READ MODELS)
-- ═══════════════════════════════════════════════════════════════════════════════

-- Sales summary projection
CREATE TABLE IF NOT EXISTS projection_sales_summary (
    id SERIAL PRIMARY KEY,
    branch_id TEXT NOT NULL,
    date DATE NOT NULL,
    total_sales DECIMAL(15, 2) NOT NULL DEFAULT 0,
    total_transactions INTEGER NOT NULL DEFAULT 0,
    total_items INTEGER NOT NULL DEFAULT 0,
    average_transaction DECIMAL(15, 2),
    by_category JSONB DEFAULT '{}',
    by_payment_method JSONB DEFAULT '{}',
    last_event_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    UNIQUE(branch_id, date)
);

CREATE INDEX IF NOT EXISTS idx_sales_summary_branch_date 
    ON projection_sales_summary(branch_id, date);

-- Inventory projection
CREATE TABLE IF NOT EXISTS projection_inventory (
    product_id TEXT NOT NULL,
    branch_id TEXT NOT NULL,
    current_quantity INTEGER NOT NULL DEFAULT 0,
    reserved_quantity INTEGER NOT NULL DEFAULT 0,
    available_quantity INTEGER NOT NULL DEFAULT 0,
    last_movement_at TIMESTAMPTZ,
    reorder_point INTEGER DEFAULT 10,
    last_event_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (product_id, branch_id)
);

CREATE INDEX IF NOT EXISTS idx_inventory_branch 
    ON projection_inventory(branch_id);
CREATE INDEX IF NOT EXISTS idx_inventory_low_stock 
    ON projection_inventory(branch_id, available_quantity) 
    WHERE available_quantity <= reorder_point;

-- Customer loyalty projection
CREATE TABLE IF NOT EXISTS projection_customer_loyalty (
    customer_id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    email TEXT,
    phone TEXT,
    total_points INTEGER NOT NULL DEFAULT 0,
    available_points INTEGER NOT NULL DEFAULT 0,
    lifetime_value DECIMAL(15, 2) NOT NULL DEFAULT 0,
    membership_tier TEXT DEFAULT 'bronze',
    last_purchase_at TIMESTAMPTZ,
    last_event_id UUID,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_loyalty_tier 
    ON projection_customer_loyalty(membership_tier);

-- Branch metrics projection
CREATE TABLE IF NOT EXISTS projection_branch_metrics (
    branch_id TEXT NOT NULL,
    date DATE NOT NULL,
    total_transactions INTEGER NOT NULL DEFAULT 0,
    total_amount DECIMAL(15, 2) NOT NULL DEFAULT 0,
    compliance_events INTEGER NOT NULL DEFAULT 0,
    decisions_made INTEGER NOT NULL DEFAULT 0,
    avg_decision_confidence DECIMAL(5, 4),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (branch_id, date)
);

CREATE INDEX IF NOT EXISTS idx_branch_metrics_date 
    ON projection_branch_metrics(date);
CREATE INDEX IF NOT EXISTS idx_branch_metrics_branch 
    ON projection_branch_metrics(branch_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- CRDT STATE STORAGE
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS crdt_state (
    crdt_id TEXT NOT NULL,
    node_id TEXT NOT NULL,
    crdt_type TEXT NOT NULL,
    state_data JSONB NOT NULL,
    version BIGINT NOT NULL DEFAULT 0,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    PRIMARY KEY (crdt_id, node_id)
);

CREATE INDEX IF NOT EXISTS idx_crdt_type 
    ON crdt_state(crdt_type);
CREATE INDEX IF NOT EXISTS idx_crdt_node 
    ON crdt_state(node_id);

-- ═══════════════════════════════════════════════════════════════════════════════
-- AGENT REGISTRY
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS agent_registry (
    agent_id TEXT PRIMARY KEY,
    agent_type TEXT NOT NULL,
    branch_id TEXT,
    region_id TEXT,
    address TEXT,
    capabilities TEXT[],
    status TEXT NOT NULL DEFAULT 'INITIALIZING',
    public_key TEXT,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_heartbeat TIMESTAMPTZ,
    metadata JSONB DEFAULT '{}',
    
    CONSTRAINT valid_agent_status CHECK (status IN (
        'INITIALIZING', 'ACTIVE', 'DEGRADED', 'OFFLINE', 'SHUTDOWN'
    ))
);

CREATE INDEX IF NOT EXISTS idx_agent_branch 
    ON agent_registry(branch_id);
CREATE INDEX IF NOT EXISTS idx_agent_region 
    ON agent_registry(region_id);
CREATE INDEX IF NOT EXISTS idx_agent_status 
    ON agent_registry(status);

-- ═══════════════════════════════════════════════════════════════════════════════
-- AUDIT LOG
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS audit_log (
    id BIGSERIAL PRIMARY KEY,
    event_id UUID,
    action TEXT NOT NULL,
    actor_id TEXT NOT NULL,
    actor_type TEXT NOT NULL DEFAULT 'USER',
    resource_type TEXT NOT NULL,
    resource_id TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_event 
    ON audit_log(event_id);
CREATE INDEX IF NOT EXISTS idx_audit_actor 
    ON audit_log(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_resource 
    ON audit_log(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_created 
    ON audit_log(created_at);

-- Partition audit_log by month
-- Static initial partitions for 2024
CREATE TABLE IF NOT EXISTS audit_log_2024_01 PARTITION OF audit_log
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE IF NOT EXISTS audit_log_2024_02 PARTITION OF audit_log
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE IF NOT EXISTS audit_log_2024_03 PARTITION OF audit_log
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- Auto-create future monthly partitions on demand
CREATE OR REPLACE FUNCTION create_audit_partition_if_not_exists()
RETURNS TRIGGER AS $$
DECLARE
    partition_name TEXT;
    partition_start TEXT;
    partition_end TEXT;
BEGIN
    partition_name := 'audit_log_' || to_char(NEW.created_at, 'YYYY_MM');
    partition_start := to_char(NEW.created_at, 'YYYY-MM-01');
    partition_end := to_char(
        (NEW.created_at + INTERVAL '1 month')::date,
        'YYYY-MM-01'
    );

    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF audit_log FOR VALUES FROM (%L) TO (%L)',
            partition_name, partition_start, partition_end
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_auto_audit_partition
    BEFORE INSERT ON audit_log
    FOR EACH ROW
    EXECUTE FUNCTION create_audit_partition_if_not_exists();

-- ═══════════════════════════════════════════════════════════════════════════════
-- SECURITY POLICIES
-- ═══════════════════════════════════════════════════════════════════════════════

-- Row Level Security for event_store
ALTER TABLE event_store ENABLE ROW LEVEL SECURITY;

CREATE POLICY event_store_tenant_isolation ON event_store
    USING (metadata->>'tenant_id' = current_setting('app.current_tenant', true));

-- Row Level Security for projections
ALTER TABLE projection_sales_summary ENABLE ROW LEVEL SECURITY;
ALTER TABLE projection_inventory ENABLE ROW LEVEL SECURITY;

CREATE POLICY sales_summary_branch_isolation ON projection_sales_summary
    USING (branch_id = current_setting('app.current_branch', true));

CREATE POLICY inventory_branch_isolation ON projection_inventory
    USING (branch_id = current_setting('app.current_branch', true));

-- ═══════════════════════════════════════════════════════════════════════════════
-- FUNCTIONS & TRIGGERS
-- ═══════════════════════════════════════════════════════════════════════════════

-- Function to update stream_metadata version
CREATE OR REPLACE FUNCTION update_stream_version()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO stream_metadata (stream_id, stream_type, current_version, branch_id, tenant_id)
    VALUES (
        NEW.stream_id, 
        NEW.metadata->>'stream_type',
        NEW.version,
        NEW.metadata->>'branch_id',
        NEW.metadata->>'tenant_id'
    )
    ON CONFLICT (stream_id) 
    DO UPDATE SET 
        current_version = EXCLUDED.current_version,
        updated_at = NOW();
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_stream_version
    AFTER INSERT ON event_store
    FOR EACH ROW
    EXECUTE FUNCTION update_stream_version();

-- Function to prevent event updates (append-only)
CREATE OR REPLACE FUNCTION prevent_event_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'Event store is append-only. Updates are not allowed.';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_event_update
    BEFORE UPDATE ON event_store
    FOR EACH ROW
    EXECUTE FUNCTION prevent_event_update();

CREATE TRIGGER trg_prevent_event_delete
    BEFORE DELETE ON event_store
    FOR EACH ROW
    EXECUTE FUNCTION prevent_event_update();

-- Function to update timestamps
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_stream_metadata_updated_at
    BEFORE UPDATE ON stream_metadata
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER trg_saga_instances_updated_at
    BEFORE UPDATE ON saga_instances
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

-- ═══════════════════════════════════════════════════════════════════════════════
-- VIEWS
-- ═══════════════════════════════════════════════════════════════════════════════

-- View for daily sales by branch
CREATE OR REPLACE VIEW v_daily_sales AS
SELECT 
    branch_id,
    date,
    total_sales,
    total_transactions,
    average_transaction,
    by_category,
    by_payment_method
FROM projection_sales_summary
ORDER BY date DESC, branch_id;

-- View for low stock alerts
CREATE OR REPLACE VIEW v_low_stock_alerts AS
SELECT 
    product_id,
    branch_id,
    current_quantity,
    available_quantity,
    reorder_point,
    (reorder_point - available_quantity) as deficit
FROM projection_inventory
WHERE available_quantity <= reorder_point
ORDER BY deficit DESC;

-- View for agent health
CREATE OR REPLACE VIEW v_agent_health AS
SELECT 
    agent_id,
    agent_type,
    branch_id,
    region_id,
    status,
    last_heartbeat,
    CASE 
        WHEN last_heartbeat IS NULL THEN 'UNKNOWN'
        WHEN last_heartbeat < NOW() - INTERVAL '5 minutes' THEN 'UNHEALTHY'
        ELSE 'HEALTHY'
    END as health_status
FROM agent_registry
ORDER BY last_heartbeat DESC NULLS LAST;

-- ═══════════════════════════════════════════════════════════════════════════════
-- END OF SCHEMA
-- ═══════════════════════════════════════════════════════════════════════════════
