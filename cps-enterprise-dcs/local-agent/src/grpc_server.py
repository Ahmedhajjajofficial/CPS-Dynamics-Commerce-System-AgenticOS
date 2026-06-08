"""
gRPC Server for Local Agent
===========================
Exposes Local Agent functionality via gRPC.

Services:
- AccountingSwarmProtocol: Event broadcasting and sync
- QueryProtocol: Read model queries
"""

from __future__ import annotations

import asyncio
import logging
from concurrent import futures
from typing import AsyncIterator
import grpc
from datetime import datetime
from google.protobuf import timestamp_pb2

logger = logging.getLogger("gRPCServer")

# Import generated protobuf code
from .proto import cps_enterprise_v4_pb2 as pb2
from .proto import cps_enterprise_v4_pb2_grpc as pb2_grpc

from .agent import LocalAgent
from .event_store import StoredEvent, EventMetadata


class AccountingSwarmServicer(pb2_grpc.AccountingSwarmProtocolServicer):
    """gRPC servicer for AccountingSwarmProtocol."""
    
    def __init__(self, agent: LocalAgent):
        self.agent = agent
    
    async def BroadcastFinancialEvent(self, request: pb2.SovereignFinancialEvent, context: grpc.aio.ServicerContext) -> pb2.AckResponse:
        """Handle single event broadcast."""
        try:
            # Append to event store
            # In a real system, we would map the proto message to our internal StoredEvent
            # For this baseline, we use the raw payload if present
            stream_id = f"{self.agent.config.branch_id}:incoming"
            
            metadata = EventMetadata(
                correlation_id=request.correlation_id,
                agent_id=request.agent_id
            )
            
            event = await self.agent.event_store.append(
                stream_id=stream_id,
                event_type=pb2.EventType.Name(request.type),
                payload=request.payload.encrypted_data if request.HasField("payload") else b"",
                metadata=metadata
            )
            
            # Build response
            return pb2.AckResponse(
                success=True,
                message="Event recorded",
                receipt_hash=event.event_hash,
                is_duplicate=False,
                processing_node=self.agent.config.agent_id
            )
        except Exception as e:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details("Internal processing error")
            import logging
            logging.getLogger("gRPC").error("BroadcastFinancialEvent failed: %s", e)
            return pb2.AckResponse(success=False, message="Internal processing error")
    
    async def SubscribeEvents(self, request: pb2.SubscribeRequest, context: grpc.aio.ServicerContext) -> AsyncIterator[pb2.SovereignFinancialEvent]:
        """Stream events to subscriber."""
        event_types = [pb2.EventType.Name(t) for t in request.event_types]
        
        # Subscribe to events
        from .event_store import EventStoreSubscription
        subscription = EventStoreSubscription(
            self.agent.event_store,
            event_types=event_types if event_types else None
        )
        
        # Create async generator
        async for event in self._event_generator(subscription):
            yield self._stored_event_to_proto(event)
    
    async def _event_generator(self, subscription):
        """Generate events from subscription."""
        queue = asyncio.Queue()
        
        def handler(event: StoredEvent):
            queue.put_nowait(event)
        
        subscription.on_event(handler)
        
        # Start subscription in background
        task = asyncio.create_task(subscription.start())
        
        try:
            while True:
                event = await queue.get()
                yield event
        finally:
            subscription.stop()
            task.cancel()
    
    def _stored_event_to_proto(self, event: StoredEvent) -> pb2.SovereignFinancialEvent:
        """Convert StoredEvent to protobuf message."""
        # This is a simplified mapping
        proto_event = pb2.SovereignFinancialEvent(
            event_id=event.event_id,
            stream_version=event.version,
            type=pb2.EventType.Value(event.event_type) if hasattr(pb2.EventType, event.event_type) else pb2.UNKNOWN
        )
        
        # Set timestamp
        proto_event.ts.FromDatetime(event.created_at)
        return proto_event

    async def RequestReconciliation(self, request: pb2.ReconciliationRequest, context: grpc.aio.ServicerContext) -> pb2.ReconciliationResponse:
        """Handle reconciliation request."""
        return pb2.ReconciliationResponse(
            is_balanced=True,
            actual_balance=0.0,
            reconciliation_timestamp=pb2.HybridLogicalClock(physical_ms=int(datetime.now().timestamp() * 1000))
        )


class QueryServicer(pb2_grpc.QueryProtocolServicer):
    """gRPC servicer for QueryProtocol."""
    
    def __init__(self, agent: LocalAgent):
        self.agent = agent
    
    async def GetBranchSummary(self, request: pb2.BranchQuery, context: grpc.aio.ServicerContext) -> pb2.BranchSummary:
        """Get branch summary."""
        try:
            summary = await self.agent.get_branch_summary()
            return pb2.BranchSummary(
                branch_id=summary["branch_id"],
                today_sales=summary["today_sales"],
                today_transactions=0,
                current_balance=0.0,
                active_sessions=0
            )
        except Exception as e:
            logger.error("GetBranchSummary failed: %s", e, exc_info=True)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to get branch summary: {e}")
            return pb2.BranchSummary()
    
    async def GetInventoryStatus(self, request: pb2.InventoryQuery, context: grpc.aio.ServicerContext) -> pb2.InventoryStatus:
        """Get inventory status."""
        try:
            product_id = request.product_id
            quantity = self.agent.get_inventory_level(product_id)
            
            return pb2.InventoryStatus(
                product_id=product_id,
                branch_id=self.agent.config.branch_id,
                current_quantity=quantity,
                available_quantity=quantity,
                is_low_stock=quantity < 10
            )
        except Exception as e:
            logger.error("GetInventoryStatus failed: %s", e, exc_info=True)
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to get inventory status: {e}")
            return pb2.InventoryStatus()
    
    async def SubscribeDashboard(self, request: pb2.DashboardSubscription, context: grpc.aio.ServicerContext) -> AsyncIterator[pb2.DashboardUpdate]:
        """Stream dashboard updates."""
        while not context.cancelled():
            try:
                summary = await self.agent.get_branch_summary()
                
                ts = timestamp_pb2.Timestamp()
                ts.GetCurrentTime()
                
                yield pb2.DashboardUpdate(
                    metric_name="today_sales",
                    value=summary["today_sales"],
                    display_value=f"${summary['today_sales']:.2f}",
                    timestamp=ts
                )
            except asyncio.CancelledError:
                break
            except Exception as e:
                logger.error("SubscribeDashboard iteration failed: %s", e, exc_info=True)
            
            await asyncio.sleep(request.update_interval_ms / 1000.0)


class LocalAgentGRPCServer:
    """gRPC server for the local agent."""
    
    def __init__(self, agent: LocalAgent, port: int = 50051):
        self.agent = agent
        self.port = port
        self.server = None
    
    async def start(self):
        """Start the gRPC server."""
        self.server = grpc.aio.server(futures.ThreadPoolExecutor(max_workers=10))
        
        # Register services
        pb2_grpc.add_AccountingSwarmProtocolServicer_to_server(
            AccountingSwarmServicer(self.agent), self.server
        )
        pb2_grpc.add_QueryProtocolServicer_to_server(
            QueryServicer(self.agent), self.server
        )
        
        self.server.add_insecure_port(f"[::]:{self.port}")
        await self.server.start()
        
        print(f"gRPC server started on port {self.port}")
    
    async def stop(self):
        """Stop the gRPC server."""
        if self.server:
            await self.server.stop(5)
            print("gRPC server stopped")
    
    async def serve_forever(self):
        """Run server until interrupted."""
        await self.server.wait_for_termination()
