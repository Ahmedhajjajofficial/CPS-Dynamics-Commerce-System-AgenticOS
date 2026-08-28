-- ═══════════════════════════════════════════════════════════════════════════════
-- CP'S Enterprise DCS - Master Agent Schema
-- Global Orchestration Tables
-- ═══════════════════════════════════════════════════════════════════════════════

-- ═══════════════════════════════════════════════════════════════════════════════
-- RECONCILIATION
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS reconciliation_jobs (
    job_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    region_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('daily', 'weekly', 'monthly', 'adhoc', 'cross_region')),
    status VARCHAR(50) NOT NULL CHECK (status IN ('pending', 'in_progress', 'completed', 'failed', 'disputed')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    requested_by VARCHAR(255),
    result_summary JSONB
);

CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_region ON reconciliation_jobs(region_id);
CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_status ON reconciliation_jobs(status);
CREATE INDEX IF NOT EXISTS idx_reconciliation_jobs_created_at ON reconciliation_jobs(created_at);

-- ═══════════════════════════════════════════════════════════════════════════════
-- DECISIONS LOG
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS decisions_log (
    decision_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id VARCHAR(255) NOT NULL,
    branch_id VARCHAR(255),
    decision_type VARCHAR(100) NOT NULL,
    decision_payload JSONB NOT NULL,
    reasoning TEXT,
    confidence_score FLOAT CHECK (confidence_score >= 0 AND confidence_score <= 1),
    decided_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    decided_by VARCHAR(255) NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_decisions_log_agent ON decisions_log(agent_id);
CREATE INDEX IF NOT EXISTS idx_decisions_log_branch ON decisions_log(branch_id);
CREATE INDEX IF NOT EXISTS idx_decisions_log_created_at ON decisions_log(decided_at);

-- ═══════════════════════════════════════════════════════════════════════════════
-- REGIONAL AGENTS REGISTRY
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS regional_agents (
    agent_id VARCHAR(255) PRIMARY KEY,
    region_id VARCHAR(255) NOT NULL,
    rpc_address VARCHAR(255) NOT NULL,
    raft_address VARCHAR(255) NOT NULL,
    status VARCHAR(50) NOT NULL CHECK (status IN ('initializing', 'active', 'degraded', 'offline', 'shutdown')),
    capabilities TEXT[],
    public_key TEXT,
    last_heartbeat TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS idx_regional_agents_region ON regional_agents(region_id);
CREATE INDEX IF NOT EXISTS idx_regional_agents_status ON regional_agents(status);
CREATE INDEX IF NOT EXISTS idx_regional_agents_last_heartbeat ON regional_agents(last_heartbeat);

-- ═══════════════════════════════════════════════════════════════════════════════
-- GLOBAL EVENTS STREAM
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS global_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_region_id VARCHAR(255) NOT NULL,
    source_branch_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    event_payload JSONB NOT NULL,
    occurred_at TIMESTAMPTZ NOT NULL,
    received_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    kafka_topic VARCHAR(255),
    kafka_partition INT,
    kafka_offset BIGINT
);

CREATE INDEX IF NOT EXISTS idx_global_events_region ON global_events(source_region_id);
CREATE INDEX IF NOT EXISTS idx_global_events_branch ON global_events(source_branch_id);
CREATE INDEX IF NOT EXISTS idx_global_events_type ON global_events(event_type);
CREATE INDEX IF NOT EXISTS idx_global_events_occurred_at ON global_events(occurred_at);

-- Partition global_events by month
CREATE TABLE IF NOT EXISTS global_events_2024_01 PARTITION OF global_events
    FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');
CREATE TABLE IF NOT EXISTS global_events_2024_02 PARTITION OF global_events
    FOR VALUES FROM ('2024-02-01') TO ('2024-03-01');
CREATE TABLE IF NOT EXISTS global_events_2024_03 PARTITION OF global_events
    FOR VALUES FROM ('2024-03-01') TO ('2024-04-01');

-- Auto-create future monthly partitions
CREATE OR REPLACE FUNCTION create_global_events_partition_if_not_exists()
RETURNS TRIGGER AS $$
DECLARE
    partition_name TEXT;
    partition_start TEXT;
    partition_end TEXT;
BEGIN
    partition_name := 'global_events_' || to_char(NEW.occurred_at, 'YYYY_MM');
    partition_start := to_char(NEW.occurred_at, 'YYYY-MM-01');
    partition_end := to_char(
        (NEW.occurred_at + INTERVAL '1 month')::date,
        'YYYY-MM-01'
    );

    IF NOT EXISTS (
        SELECT 1 FROM pg_class WHERE relname = partition_name
    ) THEN
        EXECUTE format(
            'CREATE TABLE IF NOT EXISTS %I PARTITION OF global_events FOR VALUES FROM (%L) TO (%L)',
            partition_name, partition_start, partition_end
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_auto_global_events_partition
    BEFORE INSERT ON global_events
    FOR EACH ROW
    EXECUTE FUNCTION create_global_events_partition_if_not_exists();

-- ═══════════════════════════════════════════════════════════════════════════════
-- FORECAST EVALUATIONS
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS forecast_evaluations (
    evaluation_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    branch_id VARCHAR(255) NOT NULL,
    model_version VARCHAR(100) NOT NULL,
    evaluation_start TIMESTAMPTZ NOT NULL,
    evaluation_end TIMESTAMPTZ NOT NULL,
    mae FLOAT,
    rmse FLOAT,
    mape FLOAT,
    sample_count INT,
    evaluated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_forecast_evaluations_branch ON forecast_evaluations(branch_id);
CREATE INDEX IF NOT EXISTS idx_forecast_evaluations_model ON forecast_evaluations(model_version);

-- ═══════════════════════════════════════════════════════════════════════════════
-- AUDIT LOG FOR MASTER AGENT
-- ═══════════════════════════════════════════════════════════════════════════════

CREATE TABLE IF NOT EXISTS master_audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action VARCHAR(255) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL DEFAULT 'MASTER_AGENT',
    resource_type VARCHAR(100),
    resource_id VARCHAR(255),
    details JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_master_audit_log_actor ON master_audit_log(actor_id);
CREATE INDEX IF NOT EXISTS idx_master_audit_log_action ON master_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_master_audit_log_created_at ON master_audit_log(created_at);

-- Append-only trigger
CREATE OR REPLACE FUNCTION prevent_master_audit_update()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'master_audit_log is append-only';
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_prevent_master_audit_update
    BEFORE UPDATE ON master_audit_log
    FOR EACH ROW
    EXECUTE FUNCTION prevent_master_audit_update();

CREATE TRIGGER trg_prevent_master_audit_delete
    BEFORE DELETE ON master_audit_log
    FOR EACH ROW
    EXECUTE FUNCTION prevent_master_audit_delete();
