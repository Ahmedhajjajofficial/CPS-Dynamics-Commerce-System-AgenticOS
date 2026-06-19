"""
Event Store Implementation
==========================
Append-only log for immutable financial events.

Design Principles:
- Events are immutable - never updated or deleted
- Optimistic concurrency control via stream versioning
- Temporal queries supported via timestamp indexing
- Encrypted payloads for data sovereignty

PostgreSQL Schema:
    CREATE TABLE event_store (
        stream_id UUID NOT NULL,
        version BIGINT NOT NULL,
        event_id UUID NOT NULL DEFAULT gen_random_uuid(),
        event_type TEXT NOT NULL,
        payload BYTEA NOT NULL,
        metadata JSONB NOT NULL,
        created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
        PRIMARY KEY (stream_id, version),
        UNIQUE (event_id)
    ) PARTITION BY HASH (stream_id);
"""

from __future__ import annotations

import json
import sqlite3
import uuid
from abc import ABC, abstractmethod
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, List, Optional, Any, Callable
from contextlib import contextmanager
import logging
import threading
import asyncio
from pathlib import Path

logger = logging.getLogger("EventStore")


@dataclass
class EventMetadata:
    """Metadata for a financial event."""
    correlation_id: Optional[str] = None
    agent_id: Optional[str] = None
    saga_id: Optional[str] = None
    tenant_id: Optional[str] = None
    causal_context: Optional[Dict[str, Any]] = None
    public_metadata: Optional[bytes] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "correlation_id": self.correlation_id,
            "agent_id": self.agent_id,
            "saga_id": self.saga_id,
            "tenant_id": self.tenant_id,
            "causal_context": self.causal_context,
            "public_metadata": self.public_metadata.hex() if self.public_metadata else None
        }
    
    @classmethod
    def from_dict(cls, data: Dict[str, Any]) -> 'EventMetadata':
        return cls(
            correlation_id=data.get("correlation_id"),
            agent_id=data.get("agent_id"),
            saga_id=data.get("saga_id"),
            tenant_id=data.get("tenant_id"),
            causal_context=data.get("causal_context"),
            public_metadata=bytes.fromhex(data["public_metadata"]) if data.get("public_metadata") else None
        )


@dataclass
class StoredEvent:
    """A stored financial event."""
    event_id: str
    stream_id: str
    version: int
    event_type: str
    payload: bytes
    metadata: EventMetadata
    created_at: datetime
    event_hash: Optional[str] = None
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "event_id": self.event_id,
            "stream_id": self.stream_id,
            "version": self.version,
            "event_type": self.event_type,
            "payload": self.payload.hex(),
            "metadata": self.metadata.to_dict(),
            "created_at": self.created_at.isoformat(),
            "event_hash": self.event_hash
        }


class EventStore(ABC):
    """Abstract base class for event stores."""
    
    @abstractmethod
    async def append(
        self,
        stream_id: str,
        event_type: str,
        payload: bytes,
        metadata: EventMetadata,
        expected_version: Optional[int] = None
    ) -> StoredEvent:
        """
        Append an event to a stream.
        
        Args:
            stream_id: Unique identifier for the event stream
            event_type: Type of event (from EventType enum)
            payload: Encrypted event payload
            metadata: Event metadata
            expected_version: Expected current version for optimistic concurrency
        
        Returns:
            The stored event with assigned version
        
        Raises:
            ConcurrencyException: If expected_version doesn't match actual version
        """
        pass
    
    @abstractmethod
    async def read_stream(
        self,
        stream_id: str,
        from_version: int = 0,
        to_version: Optional[int] = None
    ) -> List[StoredEvent]:
        """Read events from a stream."""
        pass
    
    @abstractmethod
    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        """Get a single event by ID."""
        pass
    
    @abstractmethod
    async def read_all(
        self,
        from_position: int = 0,
        limit: int = 100
    ) -> List[StoredEvent]:
        """Read all events (for projections and catch-up subscriptions)."""
        pass
    
    @abstractmethod
    async def get_stream_version(self, stream_id: str) -> int:
        """Get the current version of a stream."""
        pass


class ConcurrencyException(Exception):
    """Raised when optimistic concurrency check fails."""
    pass


class SQLiteEventStore(EventStore):
    """
    SQLite-based event store for local agent.
    Optimized for edge deployment with minimal resources.
    """
    
    def __init__(self, db_path: str = "events.db"):
        self.db_path = db_path
        self._lock = threading.RLock()
        self._local = threading.local()
        self._initialize_db()
    
    def _get_connection(self) -> sqlite3.Connection:
        """Get thread-local database connection."""
        if not hasattr(self._local, 'connection') or self._local.connection is None:
            self._local.connection = sqlite3.connect(self.db_path, check_same_thread=False)
            self._local.connection.row_factory = sqlite3.Row
        return self._local.connection
    
    @contextmanager
    def _cursor(self, *, commit: bool = False):
        """Acquire the lock, yield a cursor, and optionally commit."""
        with self._lock:
            conn = self._get_connection()
            cursor = conn.cursor()
            yield cursor
            if commit:
                conn.commit()
    
    def _fetch_events(self, query: str, params: tuple = ()) -> List[StoredEvent]:
        """Execute a SELECT and return a list of StoredEvents."""
        with self._cursor() as cursor:
            cursor.execute(query, params)
            return [self._row_to_event(row) for row in cursor.fetchall()]
    
    def _fetch_one(self, query: str, params: tuple = ()) -> Optional[sqlite3.Row]:
        """Execute a SELECT and return a single row or None."""
        with self._cursor() as cursor:
            cursor.execute(query, params)
            return cursor.fetchone()
    
    def _initialize_db(self):
        """Initialize the database schema."""
        with self._lock:
            conn = self._get_connection()
            cursor = conn.cursor()
            
            # Main event store table
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS event_store (
                    event_id TEXT PRIMARY KEY,
                    stream_id TEXT NOT NULL,
                    version INTEGER NOT NULL,
                    event_type TEXT NOT NULL,
                    payload BLOB NOT NULL,
                    metadata TEXT NOT NULL,
                    created_at TEXT NOT NULL,
                    event_hash TEXT,
                    UNIQUE(stream_id, version)
                )
            """)
            
            # Indexes for efficient queries
            cursor.execute("""
                CREATE INDEX IF NOT EXISTS idx_stream_version 
                ON event_store(stream_id, version)
            """)
            
            cursor.execute("""
                CREATE INDEX IF NOT EXISTS idx_event_type 
                ON event_store(event_type)
            """)
            
            cursor.execute("""
                CREATE INDEX IF NOT EXISTS idx_created_at 
                ON event_store(created_at)
            """)
            
            # Stream metadata table for tracking current versions
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS stream_metadata (
                    stream_id TEXT PRIMARY KEY,
                    current_version INTEGER DEFAULT 0,
                    created_at TEXT NOT NULL,
                    updated_at TEXT NOT NULL
                )
            """)
            
            # Snapshots table for performance
            cursor.execute("""
                CREATE TABLE IF NOT EXISTS snapshots (
                    stream_id TEXT PRIMARY KEY,
                    version INTEGER NOT NULL,
                    snapshot_data BLOB NOT NULL,
                    created_at TEXT NOT NULL
                )
            """)
            
            conn.commit()
    
    async def append(
        self,
        stream_id: str,
        event_type: str,
        payload: bytes,
        metadata: EventMetadata,
        expected_version: Optional[int] = None
    ) -> StoredEvent:
        """Append an event with optimistic concurrency control."""
        with self._lock:
            conn = self._get_connection()
            cursor = conn.cursor()
            
            # Get or create stream metadata
            cursor.execute(
                "SELECT current_version FROM stream_metadata WHERE stream_id = ?",
                (stream_id,)
            )
            row = cursor.fetchone()
            
            if row:
                current_version = row[0]
            else:
                current_version = 0
                now = datetime.utcnow().isoformat()
                cursor.execute(
                    "INSERT INTO stream_metadata (stream_id, current_version, created_at, updated_at) VALUES (?, ?, ?, ?)",
                    (stream_id, 0, now, now)
                )
            
            # Check optimistic concurrency
            if expected_version is not None and current_version != expected_version:
                raise ConcurrencyException(
                    f"Expected version {expected_version} but found {current_version}"
                )
            
            # Generate new version and event ID
            new_version = current_version + 1
            event_id = str(uuid.uuid4())
            created_at = datetime.utcnow()
            
            # Calculate event hash for integrity
            event_hash = self._calculate_hash(event_id, stream_id, new_version, payload)
            
            # Insert the event
            cursor.execute(
                """
                INSERT INTO event_store 
                (event_id, stream_id, version, event_type, payload, metadata, created_at, event_hash)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                (
                    event_id,
                    stream_id,
                    new_version,
                    event_type,
                    payload,
                    json.dumps(metadata.to_dict()),
                    created_at.isoformat(),
                    event_hash
                )
            )
            
            # Update stream metadata
            cursor.execute(
                "UPDATE stream_metadata SET current_version = ?, updated_at = ? WHERE stream_id = ?",
                (new_version, created_at.isoformat(), stream_id)
            )
            
            conn.commit()
            
            return StoredEvent(
                event_id=event_id,
                stream_id=stream_id,
                version=new_version,
                event_type=event_type,
                payload=payload,
                metadata=metadata,
                created_at=created_at,
                event_hash=event_hash
            )
    
    async def read_stream(
        self,
        stream_id: str,
        from_version: int = 0,
        to_version: Optional[int] = None
    ) -> List[StoredEvent]:
        """Read events from a stream."""
        if to_version is not None:
            return self._fetch_events(
                "SELECT * FROM event_store WHERE stream_id = ? AND version >= ? AND version <= ? ORDER BY version ASC",
                (stream_id, from_version, to_version)
            )
        return self._fetch_events(
            "SELECT * FROM event_store WHERE stream_id = ? AND version >= ? ORDER BY version ASC",
            (stream_id, from_version)
        )
    
    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        """Get a single event by ID."""
        row = self._fetch_one(
            "SELECT * FROM event_store WHERE event_id = ?", (event_id,)
        )
        return self._row_to_event(row) if row else None
    
    async def read_all(
        self,
        from_position: int = 0,
        limit: int = 100
    ) -> List[StoredEvent]:
        """Read all events (for projections)."""
        return self._fetch_events(
            "SELECT * FROM event_store ORDER BY created_at ASC LIMIT ? OFFSET ?",
            (limit, from_position)
        )
    
    async def get_stream_version(self, stream_id: str) -> int:
        """Get the current version of a stream."""
        row = self._fetch_one(
            "SELECT current_version FROM stream_metadata WHERE stream_id = ?",
            (stream_id,)
        )
        return row[0] if row else 0
    
    async def save_snapshot(
        self,
        stream_id: str,
        version: int,
        snapshot_data: bytes
    ) -> None:
        """Save a snapshot for faster aggregate loading."""
        with self._cursor(commit=True) as cursor:
            cursor.execute(
                "INSERT OR REPLACE INTO snapshots (stream_id, version, snapshot_data, created_at) VALUES (?, ?, ?, ?)",
                (stream_id, version, snapshot_data, datetime.utcnow().isoformat())
            )
    
    async def get_snapshot(
        self,
        stream_id: str
    ) -> Optional[tuple[int, bytes]]:
        """Get the latest snapshot for a stream."""
        row = self._fetch_one(
            "SELECT version, snapshot_data FROM snapshots WHERE stream_id = ?",
            (stream_id,)
        )
        return (row[0], row[1]) if row else None
    
    async def query_by_event_type(
        self,
        event_type: str,
        from_time: Optional[datetime] = None,
        to_time: Optional[datetime] = None,
        limit: int = 100
    ) -> List[StoredEvent]:
        """Query events by type and time range."""
        query = "SELECT * FROM event_store WHERE event_type = ?"
        params: List[Any] = [event_type]
        
        if from_time:
            query += " AND created_at >= ?"
            params.append(from_time.isoformat())
        
        if to_time:
            query += " AND created_at <= ?"
            params.append(to_time.isoformat())
        
        query += " ORDER BY created_at ASC LIMIT ?"
        params.append(limit)
        
        return self._fetch_events(query, tuple(params))
    
    def _row_to_event(self, row: sqlite3.Row) -> StoredEvent:
        """Convert a database row to a StoredEvent."""
        metadata_dict = json.loads(row["metadata"])
        metadata = EventMetadata.from_dict(metadata_dict)
        
        return StoredEvent(
            event_id=row["event_id"],
            stream_id=row["stream_id"],
            version=row["version"],
            event_type=row["event_type"],
            payload=row["payload"],
            metadata=metadata,
            created_at=datetime.fromisoformat(row["created_at"]),
            event_hash=row["event_hash"]
        )
    
    def _calculate_hash(
        self,
        event_id: str,
        stream_id: str,
        version: int,
        payload: bytes
    ) -> str:
        """Calculate integrity hash for an event."""
        import hashlib
        data = f"{event_id}:{stream_id}:{version}:{payload.hex()}"
        return hashlib.sha256(data.encode()).hexdigest()


class InMemoryEventStore(EventStore):
    """
    In-memory event store for testing and development.
    Not suitable for production use.
    """
    
    def __init__(self):
        self._events: Dict[str, List[StoredEvent]] = {}
        self._all_events: Dict[str, StoredEvent] = {}
        self._stream_versions: Dict[str, int] = {}
        self._lock = threading.RLock()
    
    async def append(
        self,
        stream_id: str,
        event_type: str,
        payload: bytes,
        metadata: EventMetadata,
        expected_version: Optional[int] = None
    ) -> StoredEvent:
        with self._lock:
            current_version = self._stream_versions.get(stream_id, 0)
            
            if expected_version is not None and current_version != expected_version:
                raise ConcurrencyException(
                    f"Expected version {expected_version} but found {current_version}"
                )
            
            new_version = current_version + 1
            event_id = str(uuid.uuid4())
            created_at = datetime.utcnow()
            
            event = StoredEvent(
                event_id=event_id,
                stream_id=stream_id,
                version=new_version,
                event_type=event_type,
                payload=payload,
                metadata=metadata,
                created_at=created_at
            )
            
            if stream_id not in self._events:
                self._events[stream_id] = []
            
            self._events[stream_id].append(event)
            self._all_events[event_id] = event
            self._stream_versions[stream_id] = new_version
            
            return event
    
    async def read_stream(
        self,
        stream_id: str,
        from_version: int = 0,
        to_version: Optional[int] = None
    ) -> List[StoredEvent]:
        with self._lock:
            events = self._events.get(stream_id, [])
            result = [e for e in events if e.version >= from_version]
            if to_version is not None:
                result = [e for e in result if e.version <= to_version]
            return result
    
    async def get_event(self, event_id: str) -> Optional[StoredEvent]:
        with self._lock:
            return self._all_events.get(event_id)
    
    async def read_all(
        self,
        from_position: int = 0,
        limit: int = 100
    ) -> List[StoredEvent]:
        with self._lock:
            all_events = sorted(
                self._all_events.values(),
                key=lambda e: e.created_at
            )
            return all_events[from_position:from_position + limit]
    
    async def get_stream_version(self, stream_id: str) -> int:
        with self._lock:
            return self._stream_versions.get(stream_id, 0)


class EventStoreSubscription:
    """
    Subscription to event store for real-time projections.
    """
    
    def __init__(self, event_store: EventStore, event_types: Optional[List[str]] = None):
        self.event_store = event_store
        self.event_types = event_types or []
        self._handlers: List[Callable[[StoredEvent], None]] = []
        self._running = False
        self._last_position = 0
    
    def on_event(self, handler: Callable[[StoredEvent], None]):
        """Register an event handler."""
        self._handlers.append(handler)
        return self
    
    async def start(self):
        """Start the subscription."""
        self._running = True
        while self._running:
            events = await self.event_store.read_all(
                from_position=self._last_position,
                limit=100
            )
            
            for event in events:
                if not self.event_types or event.event_type in self.event_types:
                    for handler in self._handlers:
                        try:
                            handler(event)
                        except Exception as e:
                            logger.error(
                                "Subscription handler %s failed for event %s: %s",
                                handler.__name__ if hasattr(handler, '__name__') else repr(handler),
                                event.event_id,
                                e,
                                exc_info=True,
                            )
                
                self._last_position += 1
            
            if not events:
                await asyncio.sleep(0.1)  # Small delay when no new events
    
    def stop(self):
        """Stop the subscription."""
        self._running = False
