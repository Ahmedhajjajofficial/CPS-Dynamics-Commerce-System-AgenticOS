"""
Local Agent - The Sovereign Edge
================================
Autonomous agent operating at the retail branch level.

Capabilities:
- Offline-first operation (works without network)
- Event sourcing with local SQLite store
- CRDT-based state synchronization
- Sovereign encryption for data privacy
- Real-time POS integration
- Automatic sync when network available

Architecture:
    ┌─────────────────────────────────────────┐
    │           Local Agent                   │
    │  ┌─────────┐  ┌─────────┐  ┌─────────┐ │
    │  │  POS    │  │  Event  │  │  CRDT   │ │
    │  │ Interface│  │  Store  │  │ Manager │ │
    │  └────┬────┘  └────┬────┘  └────┬────┘ │
    │       └─────────────┴─────────────┘     │
    │                   │                     │
    │            ┌─────────────┐              │
    │            │  Sync Engine │              │
    │            │  (gRPC/HTTP) │              │
    │            └─────────────┘              │
    └─────────────────────────────────────────┘
"""

from __future__ import annotations

import asyncio
import json
import uuid
from dataclasses import dataclass, field
from datetime import datetime
from typing import Dict, List, Optional, Callable, Any, Set
from enum import Enum
import logging

from .crdt import CRDTManager, PNCounter, GCounter, ORSet, LWWRegister
from .event_store import (
    EventStore, SQLiteEventStore, StoredEvent, 
    EventMetadata, EventStoreSubscription
)
from .security import CryptoManager, SovereignPayload, EncryptedPayload


# Configure logging
logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("LocalAgent")


class AgentState(Enum):
    """Operational states of the local agent."""
    INITIALIZING = "initializing"
    ACTIVE = "active"
    OFFLINE = "offline"
    DEGRADED = "degraded"
    SHUTDOWN = "shutdown"


@dataclass
class AgentConfig:
    """Configuration for the local agent."""
    agent_id: str
    branch_id: str
    region_id: str
    db_path: str = "events.db"
    sync_interval_seconds: int = 30
    batch_size: int = 100
    enable_encryption: bool = True
    master_key: Optional[bytes] = None
    regional_agent_endpoint: Optional[str] = None
    pos_interface_port: int = 50051


@dataclass
class SyncStatus:
    """Status of synchronization with regional agent."""
    last_sync_at: Optional[datetime] = None
    pending_events: int = 0
    synced_events: int = 0
    failed_events: int = 0
    is_connected: bool = False
    latency_ms: Optional[float] = None


class LocalAgent:
    """
    The sovereign local agent for retail branch operations.
    
    This agent ensures business continuity by operating independently
    of network connectivity, synchronizing when possible.
    """
    
    def __init__(self, config: AgentConfig):
        self.config = config
        self.state = AgentState.INITIALIZING
        
        # Core components
        self.event_store: EventStore = SQLiteEventStore(config.db_path)
        self.crdt_manager = CRDTManager(config.agent_id)
        self.crypto_manager = CryptoManager(config.master_key)
        self.sovereign_payload = SovereignPayload(self.crypto_manager)
        
        # State management
        self._inventory_counters: Dict[str, PNCounter] = {}
        self._sales_counters: Dict[str, GCounter] = {}
        self._price_registers: Dict[str, LWWRegister] = {}
        self._active_promotions: ORSet = self.crdt_manager.create_orset("active_promotions")
        
        # Sync management
        self.sync_status = SyncStatus()
        self._sync_task: Optional[asyncio.Task] = None
        self._running = False
        
        # Event handlers
        self._event_handlers: Dict[str, List[Callable[[StoredEvent], None]]] = {}
        self._subscription: Optional[EventStoreSubscription] = None
        
        # Saga management
        self._active_sagas: Dict[str, Dict[str, Any]] = {}
        
        logger.info(f"LocalAgent initialized: {config.agent_id} @ {config.branch_id}")
    
    async def initialize(self):
        """Initialize the agent and load state."""
        logger.info("Initializing LocalAgent...")
        
        # Initialize CRDTs from persisted state
        await self._load_crdt_state()
        
        # Start event subscription for projections
        self._subscription = EventStoreSubscription(self.event_store)
        self._subscription.on_event(self._on_event)
        
        # Set state to active
        self.state = AgentState.ACTIVE
        self._running = True
        
        # Start sync task if endpoint configured
        if self.config.regional_agent_endpoint:
            self._sync_task = asyncio.create_task(self._sync_loop())
        
        logger.info("LocalAgent initialized successfully")
    
    async def shutdown(self):
        """Gracefully shutdown the agent."""
        logger.info("Shutting down LocalAgent...")
        self.state = AgentState.SHUTDOWN
        self._running = False
        
        if self._sync_task:
            self._sync_task.cancel()
            try:
                await self._sync_task
            except asyncio.CancelledError:
                pass
        
        # Persist CRDT state
        await self._save_crdt_state()
        
        logger.info("LocalAgent shutdown complete")
    
    # ═══════════════════════════════════════════════════════════════════════════
    # SHARED EVENT RECORDING HELPER
    # ═══════════════════════════════════════════════════════════════════════════
    
    async def _record_event(
        self,
        stream_id: str,
        event_type: str,
        event_data: Dict[str, Any],
        correlation_id: Optional[str] = None,
        tenant_id: Optional[str] = None,
        public_metadata: Optional[Dict[str, Any]] = None,
        sensitive_fields: Optional[list] = None
    ) -> StoredEvent:
        """
        Record an event to the store with optional encryption.
        
        Centralizes the repeated encrypt-or-serialize / build-metadata / append
        pattern used by every business operation.
        """
        if self.config.enable_encryption and public_metadata is not None:
            encrypted = self.sovereign_payload.encrypt_event(
                event_data=event_data,
                metadata=public_metadata,
                sensitive_fields=sensitive_fields
            )
            payload = encrypted.serialize()
        else:
            payload = json.dumps(event_data).encode()
        
        event_metadata = EventMetadata(
            correlation_id=correlation_id or str(uuid.uuid4()),
            agent_id=self.config.agent_id,
            tenant_id=tenant_id
        )
        
        event = await self.event_store.append(
            stream_id=stream_id,
            event_type=event_type,
            payload=payload,
            metadata=event_metadata
        )
        
        logger.info(f"Event recorded [{event_type}]: {event.event_id}")
        return event
    
    # ═══════════════════════════════════════════════════════════════════════════
    # EVENT SOURCING OPERATIONS
    # ═══════════════════════════════════════════════════════════════════════════
    
    async def record_sale(
        self,
        product_id: str,
        quantity: int,
        unit_price: float,
        total_amount: float,
        cashier_id: str,
        session_id: str,
        customer_id: Optional[str] = None,
        payment_method: str = "cash",
        metadata: Optional[Dict[str, Any]] = None
    ) -> StoredEvent:
        """
        Record a completed sale.
        
        This is the primary business operation - fast, reliable, and
        works even when completely offline.
        """
        event_data = {
            "product_id": product_id,
            "quantity": quantity,
            "unit_price": unit_price,
            "total_amount": total_amount,
            "cashier_id": cashier_id,
            "session_id": session_id,
            "customer_id": customer_id,
            "payment_method": payment_method,
            "metadata": metadata or {}
        }
        
        event = await self._record_event(
            stream_id=f"{self.config.branch_id}:sales:{session_id}",
            event_type="SALE_COMPLETED",
            event_data=event_data,
            tenant_id=self.config.branch_id,
            public_metadata={
                "event_type": "SALE_COMPLETED",
                "branch_id": self.config.branch_id,
                "product_id": product_id,
                "total_amount": total_amount,
                "timestamp": datetime.utcnow().isoformat()
            },
            sensitive_fields=["customer_id"]
        )
        
        await self._update_sales_counter(total_amount)
        await self._update_inventory_counter(product_id, -quantity)
        
        return event
    
    async def record_inventory_receipt(
        self,
        product_id: str,
        quantity: int,
        supplier_id: str,
        purchase_order_id: str,
        unit_cost: float,
        metadata: Optional[Dict[str, Any]] = None
    ) -> StoredEvent:
        """Record inventory receipt from supplier."""
        event_data = {
            "product_id": product_id,
            "quantity": quantity,
            "supplier_id": supplier_id,
            "purchase_order_id": purchase_order_id,
            "unit_cost": unit_cost,
            "metadata": metadata or {}
        }
        
        event = await self._record_event(
            stream_id=f"{self.config.branch_id}:inventory:{product_id}",
            event_type="INVENTORY_RECEIVED",
            event_data=event_data,
            public_metadata={
                "event_type": "INVENTORY_RECEIVED",
                "branch_id": self.config.branch_id,
                "product_id": product_id,
                "quantity": quantity
            }
        )
        
        await self._update_inventory_counter(product_id, quantity)
        
        return event
    
    async def start_sales_session(
        self,
        cashier_id: str,
        register_id: str,
        opening_balance: float
    ) -> str:
        """Start a new sales session."""
        session_id = str(uuid.uuid4())
        
        await self._record_event(
            stream_id=f"{self.config.branch_id}:sessions:{session_id}",
            event_type="SESSION_OPENED",
            event_data={
                "session_id": session_id,
                "cashier_id": cashier_id,
                "register_id": register_id,
                "opening_balance": opening_balance,
                "started_at": datetime.utcnow().isoformat()
            },
            correlation_id=session_id
        )
        
        return session_id
    
    async def close_sales_session(
        self,
        session_id: str,
        closing_balance: float,
        total_sales: float,
        transaction_count: int
    ) -> StoredEvent:
        """Close a sales session with reconciliation."""
        return await self._record_event(
            stream_id=f"{self.config.branch_id}:sessions:{session_id}",
            event_type="SESSION_CLOSED",
            event_data={
                "session_id": session_id,
                "closing_balance": closing_balance,
                "total_sales": total_sales,
                "transaction_count": transaction_count,
                "closed_at": datetime.utcnow().isoformat()
            },
            correlation_id=session_id
        )
    
    # ═══════════════════════════════════════════════════════════════════════════
    # CRDT OPERATIONS
    # ═══════════════════════════════════════════════════════════════════════════
    
    def _get_or_create_counter(
        self,
        registry: Dict,
        counter_id: str,
        counter_type: str = "PN"
    ):
        """Get an existing CRDT counter or create one if missing."""
        if counter_id not in registry:
            registry[counter_id] = self.crdt_manager.create_counter(
                counter_id, counter_type=counter_type
            )
        return registry[counter_id]
    
    async def _update_inventory_counter(self, product_id: str, delta: int):
        """Update inventory counter for a product."""
        counter = self._get_or_create_counter(
            self._inventory_counters, f"inventory:{product_id}", "PN"
        )
        if delta > 0:
            counter.increment(delta)
        else:
            counter.decrement(abs(delta))
    
    async def _update_sales_counter(self, amount: float):
        """Update daily sales counter."""
        today = datetime.utcnow().strftime("%Y-%m-%d")
        counter = self._get_or_create_counter(
            self._sales_counters, f"sales:{today}", "G"
        )
        counter.increment(int(amount * 100))  # Store as cents
    
    def get_inventory_level(self, product_id: str) -> int:
        """Get current inventory level for a product."""
        counter = self._inventory_counters.get(f"inventory:{product_id}")
        return counter.value if counter else 0
    
    def get_daily_sales(self, date: Optional[str] = None) -> float:
        """Get total sales for a date (default: today)."""
        if date is None:
            date = datetime.utcnow().strftime("%Y-%m-%d")
        counter = self._sales_counters.get(f"sales:{date}")
        return counter.value / 100.0 if counter else 0.0
    
    # ═══════════════════════════════════════════════════════════════════════════
    # QUERY OPERATIONS
    # ═══════════════════════════════════════════════════════════════════════════
    
    async def get_branch_summary(self) -> Dict[str, Any]:
        """Get summary of branch operations."""
        today = datetime.utcnow().strftime("%Y-%m-%d")
        
        return {
            "branch_id": self.config.branch_id,
            "agent_id": self.config.agent_id,
            "state": self.state.value,
            "today_sales": self.get_daily_sales(today),
            "sync_status": {
                "last_sync": self.sync_status.last_sync_at.isoformat() if self.sync_status.last_sync_at else None,
                "pending_events": self.sync_status.pending_events,
                "is_connected": self.sync_status.is_connected
            }
        }
    
    async def get_sales_history(
        self,
        from_date: Optional[str] = None,
        to_date: Optional[str] = None,
        limit: int = 100
    ) -> List[StoredEvent]:
        """Get sales history."""
        # Query by event type
        from datetime import datetime as dt
        
        from_dt = dt.fromisoformat(from_date) if from_date else None
        to_dt = dt.fromisoformat(to_date) if to_date else None
        
        if hasattr(self.event_store, 'query_by_event_type'):
            return await self.event_store.query_by_event_type(
                event_type="SALE_COMPLETED",
                from_time=from_dt,
                to_time=to_dt,
                limit=limit
            )
        
        # Fallback: read all and filter
        all_events = await self.event_store.read_all(limit=limit * 10)
        return [e for e in all_events if e.event_type == "SALE_COMPLETED"][:limit]
    
    # ═══════════════════════════════════════════════════════════════════════════
    # SYNCHRONIZATION
    # ═══════════════════════════════════════════════════════════════════════════
    
    async def _sync_loop(self):
        """Background task for synchronizing with regional agent."""
        consecutive_failures = 0
        while self._running:
            try:
                await self._sync_with_regional()
                self.sync_status.is_connected = True
                self.sync_status.failed_events = 0
                consecutive_failures = 0
                if self.state == AgentState.DEGRADED:
                    self.state = AgentState.ACTIVE
                    logger.info("Sync recovered, state restored to ACTIVE")
            except Exception as e:
                consecutive_failures += 1
                self.sync_status.is_connected = False
                self.sync_status.failed_events += 1
                logger.error(
                    "Sync failed (attempt %d): %s",
                    consecutive_failures, e,
                    exc_info=True,
                )
                if consecutive_failures >= 3 and self.state == AgentState.ACTIVE:
                    self.state = AgentState.DEGRADED
                    logger.warning(
                        "Entering DEGRADED state after %d consecutive sync failures",
                        consecutive_failures,
                    )
            
            await asyncio.sleep(self.config.sync_interval_seconds)
    
    async def _sync_with_regional(self):
        """Synchronize events with regional agent."""
        # TODO: Implement gRPC sync with regional agent
        # For now, just update status
        self.sync_status.last_sync_at = datetime.utcnow()
        logger.debug("Sync with regional agent completed")
    
    async def _load_crdt_state(self):
        """Load CRDT state from persistence."""
        # TODO: Load from SQLite
        logger.debug("CRDT state loaded")
    
    async def _save_crdt_state(self):
        """Save CRDT state to persistence."""
        # TODO: Save to SQLite
        logger.debug("CRDT state saved")
    
    # ═══════════════════════════════════════════════════════════════════════════
    # EVENT HANDLING
    # ═══════════════════════════════════════════════════════════════════════════
    
    def _on_event(self, event: StoredEvent):
        """Handle events for projections."""
        handlers = self._event_handlers.get(event.event_type, [])
        for handler in handlers:
            try:
                handler(event)
            except Exception as e:
                logger.error(
                    "Event handler %s failed for event %s: %s",
                    handler.__name__ if hasattr(handler, '__name__') else repr(handler),
                    event.event_id,
                    e,
                    exc_info=True,
                )
    
    def on_event(
        self,
        event_type: str,
        handler: Callable[[StoredEvent], None]
    ):
        """Register an event handler."""
        if event_type not in self._event_handlers:
            self._event_handlers[event_type] = []
        self._event_handlers[event_type].append(handler)
    
    # ═══════════════════════════════════════════════════════════════════════════
    # SAGA ORCHESTRATION
    # ═══════════════════════════════════════════════════════════════════════════
    
    async def start_saga(
        self,
        saga_type: str,
        context: Dict[str, Any]
    ) -> str:
        """Start a new saga."""
        saga_id = str(uuid.uuid4())
        
        self._active_sagas[saga_id] = {
            "saga_id": saga_id,
            "saga_type": saga_type,
            "state": "INITIATED",
            "context": context,
            "started_at": datetime.utcnow().isoformat(),
            "steps": []
        }
        
        logger.info(f"Saga started: {saga_id}")
        return saga_id
    
    async def complete_saga_step(
        self,
        saga_id: str,
        step_name: str,
        result: Dict[str, Any]
    ):
        """Record completion of a saga step."""
        if saga_id not in self._active_sagas:
            raise ValueError(f"Unknown saga: {saga_id}")
        self._active_sagas[saga_id]["steps"].append({
            "step_name": step_name,
            "result": result,
            "completed_at": datetime.utcnow().isoformat()
        })
    
    async def complete_saga(self, saga_id: str):
        """Mark a saga as completed."""
        if saga_id not in self._active_sagas:
            raise ValueError(f"Unknown saga: {saga_id}")
        self._active_sagas[saga_id]["state"] = "COMPLETED"
        self._active_sagas[saga_id]["completed_at"] = datetime.utcnow().isoformat()
        logger.info(f"Saga completed: {saga_id}")
