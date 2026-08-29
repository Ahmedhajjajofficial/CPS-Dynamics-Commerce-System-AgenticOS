-- ═══════════════════════════════════════════════════════════════════════════════
-- CP'S Enterprise DCS - PostgreSQL High Availability Configuration
-- Patroni + etcd Cluster Setup
-- ═══════════════════════════════════════════════════════════════════════════════

-- This script configures PostgreSQL for HA using Patroni
-- Run this on the primary node after initializing the cluster

-- ═══════════════════════════════════════════════════════════════════════════════
-- WAL ARCHIVING CONFIGURATION
-- ═══════════════════════════════════════════════════════════════════════════════

-- Enable WAL archiving for point-in-time recovery
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET archive_mode = on;
ALTER SYSTEM SET archive_command = 'test ! -f /var/lib/postgresql/wal_archive/%f && cp %p /var/lib/postgresql/wal_archive/%f';
ALTER SYSTEM SET max_wal_senders = 10;
ALTER SYSTEM SET wal_keep_size = '1GB';
ALTER SYSTEM SET hot_standby = on;
ALTER SYSTEM SET hot_standby_feedback = on;

-- ═══════════════════════════════════════════════════════════════════════════════
-- REPLICATION CONFIGURATION
-- ═══════════════════════════════════════════════════════════════════════════════

-- Create replication user
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'replicator') THEN
        CREATE ROLE replicator WITH REPLICATION LOGIN PASSWORD 'CHANGE_ME_REPLICATOR_PASSWORD';
    END IF;
END
$$;

-- Create publication for logical replication
CREATE PUBLICATION dcs_publication FOR ALL TABLES;

-- ═══════════════════════════════════════════════════════════════════════════════
-- CONNECTION POOLING CONFIGURATION (PgBouncer)
-- ═══════════════════════════════════════════════════════════════════════════════

-- Create role for application connections
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'dcs_app') THEN
        CREATE ROLE dcs_app WITH LOGIN PASSWORD 'CHANGE_ME_APP_PASSWORD';
    END IF;
END
$$;

-- Grant necessary permissions
GRANT CONNECT ON DATABASE dcs_eventstore TO dcs_app;
GRANT USAGE ON SCHEMA public TO dcs_app;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO dcs_app;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA public TO dcs_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO dcs_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT ON SEQUENCES TO dcs_app;

-- ═══════════════════════════════════════════════════════════════════════════════
-- BACKUP CONFIGURATION
-- ═══════════════════════════════════════════════════════════════════════════════

-- Create backup role
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'backup') THEN
        CREATE ROLE backup WITH LOGIN PASSWORD 'CHANGE_ME_BACKUP_PASSWORD';
        GRANT EXECUTE ON FUNCTION pg_start_backup(text, bool, bool) TO backup;
        GRANT EXECUTE ON FUNCTION pg_stop_backup() TO backup;
        GRANT EXECUTE ON FUNCTION pg_switch_wal() TO backup;
        GRANT pg_read_all_settings TO backup;
        GRANT pg_read_all_data TO backup;
    END IF;
END
$$;

-- ═══════════════════════════════════════════════════════════════════════════════
-- MONITORING CONFIGURATION
-- ═══════════════════════════════════════════════════════════════════════════════

-- Create monitoring role
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'monitor') THEN
        CREATE ROLE monitor WITH LOGIN PASSWORD 'CHANGE_ME_MONITOR_PASSWORD';
        GRANT pg_monitor TO monitor;
    END IF;
END
$$;

-- ═══════════════════════════════════════════════════════════════════════════════
-- PATRONI CONFIGURATION (to be placed in /etc/patroni.yml)
-- ═══════════════════════════════════════════════════════════════════════════════

-- Note: Patroni configuration should be stored in a separate patroni.yml file
-- Example configuration:
--
-- scope: dcs-cluster
-- name: postgresql-0
--
-- restapi:
--   listen: 0.0.0.0:8008
--   connect_address: localhost:8008
--
-- etcd:
--   hosts:
--   - etcd-0:2379
--   - etcd-1:2379
--   - etcd-2:2379
--
-- bootstrap:
--   dcs:
--     ttl: 30
--     loop_wait: 10
--     retry_timeout: 10
--     maximum_lag_on_failover: 1048576
--     postgresql:
--       use_pg_rewind: true
--       parameters:
--         wal_level: replica
--         hot_standby: "on"
--         max_wal_senders: 10
--         wal_keep_size: 1GB
--         archive_mode: "on"
--         archive_command: mkdir -p /var/lib/postgresql/wal_archive && test ! -f /var/lib/postgresql/wal_archive/%f && cp %p /var/lib/postgresql/wal_archive/%f
--
--   post_init:
--     - "psql -U postgres -c \"CREATE ROLE dcs_app WITH LOGIN PASSWORD 'CHANGE_ME_APP_PASSWORD';\""
--     - "psql -U postgres -c \"GRANT CONNECT ON DATABASE dcs_eventstore TO dcs_app;\""
--
-- postgresql:
--   listen: 0.0.0.0:5432
--   connect_address: localhost:5432
--   data_dir: /var/lib/postgresql/data
--   authentication:
--     replication:
--       username: replicator
--       password: CHANGE_ME_REPLICATOR_PASSWORD
--     superuser:
--       username: postgres
--       password: CHANGE_ME_POSTGRES_PASSWORD
--     rewind:
--       username: rewind_user
--       password: CHANGE_ME_REWIND_PASSWORD

SELECT 'PostgreSQL HA configuration applied successfully';
