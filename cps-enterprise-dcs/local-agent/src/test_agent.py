"""
Unit tests for Local Agent core logic:
- CircuitBreaker
- Saga compensation helpers
- HTTP Gateway endpoints
"""

import os
import sys
import json
import time
import asyncio
import pytest
from datetime import datetime, timedelta
from unittest.mock import patch, MagicMock

# Ensure src package is importable when running from repo root
sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from agent import LocalAgent, AgentConfig, AgentState, CircuitBreaker


# ── CircuitBreaker ────────────────────────────────────────────────────────────


class TestCircuitBreaker:
    def test_initial_state_closed(self):
        cb = CircuitBreaker(failure_threshold=3, recovery_timeout=10.0)
        assert cb.state == "CLOSED"
        assert cb.allow_request() is True

    def test_opens_after_threshold(self):
        cb = CircuitBreaker(failure_threshold=2, recovery_timeout=10.0)
        cb.record_success()
        assert cb.allow_request() is True
        cb.record_failure()
        assert cb.state == "CLOSED"
        cb.record_failure()
        assert cb.state == "OPEN"
        assert cb.allow_request() is False

    def test_half_open_after_timeout(self):
        cb = CircuitBreaker(failure_threshold=2, recovery_timeout=0.1)
        cb.record_failure()
        cb.record_failure()
        assert cb.state == "OPEN"
        assert cb.allow_request() is False

        time.sleep(0.15)
        assert cb.allow_request() is True
        assert cb.state == "HALF_OPEN"

    def test_success_closes_circuit(self):
        cb = CircuitBreaker(failure_threshold=2, recovery_timeout=0.1)
        cb.record_failure()
        cb.record_failure()
        assert cb.state == "OPEN"

        time.sleep(0.15)
        assert cb.allow_request() is True
        cb.record_success()
        assert cb.state == "CLOSED"
        assert cb.allow_request() is True


# ── LocalAgent Saga helpers ───────────────────────────────────────────────────


class TestLocalAgentSaga:
    @pytest.fixture
    def agent(self, tmp_path):
        db_path = str(tmp_path / "events.db")
        cfg = AgentConfig(
            agent_id="agent-test",
            branch_id="BR001",
            region_id="R001",
            db_path=db_path,
            sync_interval_seconds=30,
            batch_size=100,
            enable_encryption=False,
            master_key=None,
            regional_agent_endpoint=None,
            pos_interface_port=50051,
        )
        agent = LocalAgent(cfg)
        return agent

    @pytest.mark.asyncio
    async def test_start_and_complete_saga(self, agent):
        await agent.initialize()
        saga_id = await agent.start_saga("STANDARD_SALE", {"product_id": "p1"})
        assert saga_id in agent._active_sagas

        await agent.complete_saga_step(saga_id, "record_sale", {"stream_id": "s1"})
        await agent.complete_saga(saga_id)
        assert agent._active_sagas[saga_id]["state"] == "COMPLETED"

    @pytest.mark.asyncio
    async def test_fail_saga_triggers_compensation(self, agent):
        await agent.initialize()
        saga_id = await agent.start_saga("STANDARD_SALE", {"product_id": "p1"})
        await agent.complete_saga_step(saga_id, "record_sale", {"stream_id": "s1"})

        await agent.fail_saga(saga_id, "payment_failed")
        saga = agent._active_sagas[saga_id]
        assert saga["state"] == "FAILED"
        assert saga["failure_reason"] == "payment_failed"

    @pytest.mark.asyncio
    async def test_timeout_saga_triggers_compensation(self, agent):
        await agent.initialize()
        saga_id = await agent.start_saga("STANDARD_SALE", {"product_id": "p1"})
        await agent.complete_saga_step(saga_id, "record_sale", {"stream_id": "s1"})

        await agent.timeout_saga(saga_id)
        saga = agent._active_sagas[saga_id]
        assert saga["state"] == "TIMED_OUT"

    @pytest.mark.asyncio
    async def test_compensation_records_reversal(self, agent):
        await agent.initialize()
        saga_id = await agent.start_saga("STANDARD_SALE", {"product_id": "p1"})
        await agent.complete_saga_step(
            saga_id,
            "record_sale",
            {
                "stream_id": "s1",
                "product_id": "p1",
                "quantity": 2,
                "total_amount": 100.0,
            },
        )

        await agent.fail_saga(saga_id, "payment_failed")
        # Allow background compensation task to run
        await asyncio.sleep(0.1)

        events = await agent.event_store.read_stream("s1")
        event_types = [e.event_type for e in events]
        assert "SALE_REVERSED" in event_types
