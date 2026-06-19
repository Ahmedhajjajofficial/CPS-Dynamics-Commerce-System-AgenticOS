"""
Unit tests for EventStore implementations.
"""

import os
import tempfile
import pytest
import pytest_asyncio
from datetime import datetime

from event_store import (
    EventMetadata, StoredEvent, SQLiteEventStore,
    InMemoryEventStore, ConcurrencyException, EventStoreSubscription,
)


# ── EventMetadata ─────────────────────────────────────────────────────────────


class TestEventMetadata:
    def test_to_dict_minimal(self):
        m = EventMetadata()
        d = m.to_dict()
        assert d["correlation_id"] is None
        assert d["agent_id"] is None

    def test_to_dict_full(self):
        m = EventMetadata(
            correlation_id="corr-1",
            agent_id="agent-1",
            saga_id="saga-1",
            tenant_id="tenant-1",
            causal_context={"key": "value"},
            public_metadata=b"\xde\xad",
        )
        d = m.to_dict()
        assert d["correlation_id"] == "corr-1"
        assert d["public_metadata"] == "dead"

    def test_roundtrip(self):
        m = EventMetadata(
            correlation_id="c1",
            agent_id="a1",
            public_metadata=b"\xab\xcd",
        )
        d = m.to_dict()
        restored = EventMetadata.from_dict(d)
        assert restored.correlation_id == "c1"
        assert restored.agent_id == "a1"
        assert restored.public_metadata == b"\xab\xcd"

    def test_from_dict_missing_fields(self):
        restored = EventMetadata.from_dict({})
        assert restored.correlation_id is None
        assert restored.public_metadata is None


# ── StoredEvent ───────────────────────────────────────────────────────────────


class TestStoredEvent:
    def test_to_dict(self):
        now = datetime(2024, 1, 1, 12, 0, 0)
        e = StoredEvent(
            event_id="ev-1",
            stream_id="stream-1",
            version=1,
            event_type="SALE",
            payload=b"\x01\x02",
            metadata=EventMetadata(agent_id="a1"),
            created_at=now,
            event_hash="abc123",
        )
        d = e.to_dict()
        assert d["event_id"] == "ev-1"
        assert d["stream_id"] == "stream-1"
        assert d["version"] == 1
        assert d["payload"] == "0102"
        assert d["event_hash"] == "abc123"
        assert d["metadata"]["agent_id"] == "a1"


# ── InMemoryEventStore ────────────────────────────────────────────────────────


class TestInMemoryEventStore:
    @pytest.fixture
    def store(self):
        return InMemoryEventStore()

    @pytest.fixture
    def metadata(self):
        return EventMetadata(agent_id="test-agent")

    @pytest.mark.asyncio
    async def test_append_and_read(self, store, metadata):
        ev = await store.append("s1", "SALE", b"payload", metadata)
        assert ev.stream_id == "s1"
        assert ev.version == 1
        assert ev.event_type == "SALE"

        events = await store.read_stream("s1")
        assert len(events) == 1
        assert events[0].event_id == ev.event_id

    @pytest.mark.asyncio
    async def test_version_auto_increments(self, store, metadata):
        e1 = await store.append("s1", "SALE", b"p1", metadata)
        e2 = await store.append("s1", "SALE", b"p2", metadata)
        assert e1.version == 1
        assert e2.version == 2

    @pytest.mark.asyncio
    async def test_concurrency_check_passes(self, store, metadata):
        await store.append("s1", "SALE", b"p1", metadata)
        e2 = await store.append("s1", "SALE", b"p2", metadata, expected_version=1)
        assert e2.version == 2

    @pytest.mark.asyncio
    async def test_concurrency_check_fails(self, store, metadata):
        await store.append("s1", "SALE", b"p1", metadata)
        with pytest.raises(ConcurrencyException):
            await store.append("s1", "SALE", b"p2", metadata, expected_version=0)

    @pytest.mark.asyncio
    async def test_get_event_by_id(self, store, metadata):
        ev = await store.append("s1", "SALE", b"p", metadata)
        found = await store.get_event(ev.event_id)
        assert found is not None
        assert found.event_id == ev.event_id

    @pytest.mark.asyncio
    async def test_get_event_missing(self, store):
        found = await store.get_event("no-such-id")
        assert found is None

    @pytest.mark.asyncio
    async def test_read_stream_empty(self, store):
        events = await store.read_stream("nonexistent")
        assert events == []

    @pytest.mark.asyncio
    async def test_read_stream_from_version(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s1", "B", b"2", metadata)
        await store.append("s1", "C", b"3", metadata)
        events = await store.read_stream("s1", from_version=2)
        assert len(events) == 2
        assert events[0].event_type == "B"

    @pytest.mark.asyncio
    async def test_read_stream_to_version(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s1", "B", b"2", metadata)
        await store.append("s1", "C", b"3", metadata)
        events = await store.read_stream("s1", from_version=1, to_version=2)
        assert len(events) == 2

    @pytest.mark.asyncio
    async def test_read_all(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s2", "B", b"2", metadata)
        events = await store.read_all()
        assert len(events) == 2

    @pytest.mark.asyncio
    async def test_read_all_with_limit(self, store, metadata):
        for i in range(5):
            await store.append("s1", "T", bytes([i]), metadata)
        events = await store.read_all(from_position=0, limit=3)
        assert len(events) == 3

    @pytest.mark.asyncio
    async def test_get_stream_version(self, store, metadata):
        assert await store.get_stream_version("s1") == 0
        await store.append("s1", "A", b"1", metadata)
        assert await store.get_stream_version("s1") == 1

    @pytest.mark.asyncio
    async def test_separate_streams(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s2", "B", b"2", metadata)
        s1_events = await store.read_stream("s1")
        s2_events = await store.read_stream("s2")
        assert len(s1_events) == 1
        assert len(s2_events) == 1
        assert s1_events[0].event_type == "A"
        assert s2_events[0].event_type == "B"


# ── SQLiteEventStore ──────────────────────────────────────────────────────────


class TestSQLiteEventStore:
    @pytest.fixture
    def db_path(self, tmp_path):
        return str(tmp_path / "test_events.db")

    @pytest.fixture
    def store(self, db_path):
        return SQLiteEventStore(db_path)

    @pytest.fixture
    def metadata(self):
        return EventMetadata(agent_id="test-agent")

    @pytest.mark.asyncio
    async def test_append_and_read(self, store, metadata):
        ev = await store.append("s1", "SALE", b"payload", metadata)
        assert ev.version == 1
        events = await store.read_stream("s1")
        assert len(events) == 1

    @pytest.mark.asyncio
    async def test_concurrency_check(self, store, metadata):
        await store.append("s1", "A", b"p1", metadata)
        with pytest.raises(ConcurrencyException):
            await store.append("s1", "B", b"p2", metadata, expected_version=0)

    @pytest.mark.asyncio
    async def test_get_event(self, store, metadata):
        ev = await store.append("s1", "A", b"p", metadata)
        found = await store.get_event(ev.event_id)
        assert found is not None
        assert found.event_type == "A"

    @pytest.mark.asyncio
    async def test_get_event_missing(self, store):
        assert await store.get_event("nonexistent") is None

    @pytest.mark.asyncio
    async def test_read_all(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s2", "B", b"2", metadata)
        events = await store.read_all()
        assert len(events) == 2

    @pytest.mark.asyncio
    async def test_get_stream_version(self, store, metadata):
        assert await store.get_stream_version("s1") == 0
        await store.append("s1", "A", b"1", metadata)
        assert await store.get_stream_version("s1") == 1

    @pytest.mark.asyncio
    async def test_event_hash_integrity(self, store, metadata):
        ev = await store.append("s1", "A", b"data", metadata)
        assert ev.event_hash is not None
        assert len(ev.event_hash) == 64  # SHA-256 hex

    @pytest.mark.asyncio
    async def test_save_and_get_snapshot(self, store, metadata):
        await store.save_snapshot("s1", 5, b"snapshot-data")
        snap = await store.get_snapshot("s1")
        assert snap is not None
        version, data = snap
        assert version == 5
        assert data == b"snapshot-data"

    @pytest.mark.asyncio
    async def test_get_snapshot_missing(self, store):
        assert await store.get_snapshot("nope") is None

    @pytest.mark.asyncio
    async def test_query_by_event_type(self, store, metadata):
        await store.append("s1", "SALE", b"1", metadata)
        await store.append("s1", "REFUND", b"2", metadata)
        await store.append("s2", "SALE", b"3", metadata)
        sales = await store.query_by_event_type("SALE")
        assert len(sales) == 2
        assert all(e.event_type == "SALE" for e in sales)

    @pytest.mark.asyncio
    async def test_query_by_event_type_with_time_range(self, store, metadata):
        await store.append("s1", "SALE", b"1", metadata)
        sales = await store.query_by_event_type(
            "SALE",
            from_time=datetime(2020, 1, 1),
            to_time=datetime(2030, 1, 1),
        )
        assert len(sales) == 1

    @pytest.mark.asyncio
    async def test_read_stream_version_range(self, store, metadata):
        await store.append("s1", "A", b"1", metadata)
        await store.append("s1", "B", b"2", metadata)
        await store.append("s1", "C", b"3", metadata)
        events = await store.read_stream("s1", from_version=2, to_version=2)
        assert len(events) == 1
        assert events[0].event_type == "B"

    @pytest.mark.asyncio
    async def test_snapshot_upsert(self, store, metadata):
        await store.save_snapshot("s1", 1, b"snap-v1")
        await store.save_snapshot("s1", 5, b"snap-v5")
        snap = await store.get_snapshot("s1")
        assert snap[0] == 5
        assert snap[1] == b"snap-v5"


# ── EventStoreSubscription ───────────────────────────────────────────────────


class TestEventStoreSubscription:
    def test_on_event_registers_handler(self):
        store = InMemoryEventStore()
        sub = EventStoreSubscription(store)
        handler = lambda e: None
        result = sub.on_event(handler)
        assert result is sub
        assert handler in sub._handlers

    def test_stop(self):
        store = InMemoryEventStore()
        sub = EventStoreSubscription(store)
        sub._running = True
        sub.stop()
        assert sub._running is False

    def test_initial_state(self):
        store = InMemoryEventStore()
        sub = EventStoreSubscription(store, event_types=["SALE"])
        assert sub.event_types == ["SALE"]
        assert sub._running is False
        assert sub._last_position == 0
