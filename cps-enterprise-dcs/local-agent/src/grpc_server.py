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
from typing import AsyncIterator, Optional
import grpc
from datetime import datetime
from google.protobuf import timestamp_pb2
import traceback
import time
import uuid
from contextlib import asynccontextmanager

logger = logging.getLogger("gRPCServer")

# Import generated protobuf code
from .proto import cps_enterprise_v4_pb2 as pb2
from .proto import cps_enterprise_v4_pb2_grpc as pb2_grpc

from .agent import LocalAgent
from .event_store import StoredEvent, EventMetadata, ConcurrencyException


class GrpcErrorHandler:
    """Centralized error handling for gRPC services."""
    
    @staticmethod
    def handle_exception(context: grpc.aio.ServicerContext, exc: Exception, operation: str) -> None:
        """Handle exceptions with proper logging and gRPC status codes."""
        error_id = f"{int(time.time() * 1000)}"
        error_details = traceback.format_exc()
        
        logger.error(
            f"[{error_id}] Operation '{operation}' failed: {exc}\n{error_details}"
        )
        
        # Map exception types to gRPC status codes
        if isinstance(exc, ConcurrencyException):
            context.set_code(grpc.StatusCode.ABORTED)
            context.set_details(f"Concurrency conflict: {str(exc)}")
        elif isinstance(exc, ValueError):
            context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
            context.set_details(f"Invalid argument: {str(exc)}")
        elif isinstance(exc, KeyError):
            context.set_code(grpc.StatusCode.NOT_FOUND)
            context.set_details(f"Resource not found: {str(exc)}")
        else:
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Internal error [{error_id}]: {str(exc)}")
    
    @staticmethod
    @asynccontextmanager
    async def with_error_handling(context: grpc.aio.ServicerContext, operation: str):
        """Context manager for error handling in gRPC methods."""
        try:
            yield
        except grpc.aio.AioRpcError:
            # Re-raise gRPC errors as-is
            raise
        except Exception as exc:
            GrpcErrorHandler.handle_exception(context, exc, operation)
            raise


class AccountingSwarmServicer(pb2_grpc.AccountingSwarmProtocolServicer):
    """gRPC servicer for AccountingSwarmProtocol."""
    
    def __init__(self, agent: LocalAgent):
        self.agent = agent
        self.error_handler = GrpcErrorHandler()
    
    async def BroadcastFinancialEvent(
        self, 
        request: pb2.SovereignFinancialEvent, 
        context: grpc.aio.ServicerContext
    ) -> pb2.AckResponse:
        """Handle single event broadcast with comprehensive error handling."""
        async with self.error_handler.with_error_handling(context, "BroadcastFinancialEvent"):
            try:
                # Validate request
                if not request.event_id:
                    context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                    context.set_details("event_id is required")
                    return pb2.AckResponse(success=False, message="Missing event_id")
                
                # Append to event store
                stream_id = f"{self.agent.config.branch_id}:incoming"
                
                metadata = EventMetadata(
                    correlation_id=request.correlation_id or str(uuid.uuid4()),
                    agent_id=request.agent_id or self.agent.config.agent_id
                )
                
                event = await self.agent.event_store.append(
                    stream_id=stream_id,
                    event_type=pb2.EventType.Name(request.type) if request.type else "UNKNOWN",
                    payload=request.payload.encrypted_data if request.HasField("payload") else b"",
                    metadata=metadata
                )
                
                logger.info(f"Event broadcasted successfully: {event.event_id}")
                
                # Build response
                return pb2.AckResponse(
                    success=True,
                    message="Event recorded successfully",
                    receipt_hash=event.event_hash or "",
                    is_duplicate=False,
                    processing_node=self.agent.config.agent_id
                )
                
            except Exception as exc:
                logger.error(f"BroadcastFinancialEvent failed: {exc}")
                logger.error(traceback.format_exc())
                
                # Determine appropriate status code
                if isinstance(exc, ConcurrencyException):
                    context.set_code(grpc.StatusCode.ABORTED)
                elif isinstance(exc, ValueError):
                    context.set_code(grpc.StatusCode.INVALID_ARGUMENT)
                else:
                    context.set_code(grpc.StatusCode.INTERNAL)
                
                context.set_details(str(exc))
                return pb2.AckResponse(
                    success=False, 
                    message=f"Error: {str(exc)}"
                )
    
    async def SubscribeEvents(
        self, 
        request: pb2.SubscribeRequest, 
        context: grpc.aio.ServicerContext
    ) -> AsyncIterator[pb2.SovereignFinancialEvent]:
        """Stream events to subscriber with error handling."""
        try:
            event_types = [pb2.EventType.Name(t) for t in request.event_types]
            
            # Subscribe to events
            from .event_store import EventStoreSubscription
            subscription = EventStoreSubscription(
                self.agent.event_store,
                event_types=event_types if event_types else None
            )
            
            # Create async generator
            async for event in self._event_generator(subscription, context):
                yield self._stored_event_to_proto(event)
                
        except asyncio.CancelledError:
            logger.info("SubscribeEvents cancelled by client")
            raise
        except Exception as exc:
            logger.error(f"SubscribeEvents failed: {exc}")
            logger.error(traceback.format_exc())
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Subscription error: {str(exc)}")
            raise
    
    async def _event_generator(self, subscription, context):
        """Generate events from subscription with proper cleanup."""
        queue = asyncio.Queue()
        
        def handler(event: StoredEvent):
            try:
                queue.put_nowait(event)
            except asyncio.QueueFull:
                logger.warning("Event queue full, dropping event")
        
        subscription.on_event(handler)
        
        # Start subscription in background
        task = asyncio.create_task(subscription.start())
        
        try:
            while not context.cancelled():
                try:
                    # Wait for event with timeout to allow cancellation checks
                    event = await asyncio.wait_for(queue.get(), timeout=1.0)
                    yield event
                except asyncio.TimeoutError:
                    continue
                except asyncio.CancelledError:
                    break
        finally:
            subscription.stop()
            task.cancel()
            try:
                await task
            except asyncio.CancelledError:
                pass
    
    def _stored_event_to_proto(self, event: StoredEvent) -> pb2.SovereignFinancialEvent:
        """Convert StoredEvent to protobuf message."""
        try:
            proto_event = pb2.SovereignFinancialEvent(
                event_id=event.event_id,
                stream_version=event.version,
                type=pb2.EventType.Value(event.event_type) if event.event_type in pb2.EventType.keys() else pb2.UNKNOWN
            )
            
            # Set timestamp
            if event.created_at:
                proto_event.ts.FromDatetime(event.created_at)
            
            return proto_event
        except Exception as exc:
            logger.error(f"Failed to convert event to proto: {exc}")
            # Return minimal valid event
            return pb2.SovereignFinancialEvent(
                event_id=event.event_id or "unknown",
                type=pb2.UNKNOWN
            )

    async def RequestReconciliation(
        self, 
        request: pb2.ReconciliationRequest, 
        context: grpc.aio.ServicerContext
    ) -> pb2.ReconciliationResponse:
        """Handle reconciliation request."""
        try:
            # TODO: Implement actual reconciliation logic
            return pb2.ReconciliationResponse(
                is_balanced=True,
                actual_balance=request.expected_balance if request.HasField("expected_balance") else 0.0,
                reconciliation_timestamp=pb2.HybridLogicalClock(physical_ms=int(datetime.now().timestamp() * 1000))
            )
        except Exception as exc:
            logger.error(f"RequestReconciliation failed: {exc}")
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Reconciliation error: {str(exc)}")
            return pb2.ReconciliationResponse(
                is_balanced=False,
                actual_balance=0.0
            )


class QueryServicer(pb2_grpc.QueryProtocolServicer):
    """gRPC servicer for QueryProtocol."""
    
    def __init__(self, agent: LocalAgent):
        self.agent = agent
    
    async def GetBranchSummary(
        self, 
        request: pb2.BranchQuery, 
        context: grpc.aio.ServicerContext
    ) -> pb2.BranchSummary:
        """Get branch summary."""
        try:
            summary = await self.agent.get_branch_summary()
            return pb2.BranchSummary(
                branch_id=summary.get("branch_id", "unknown"),
                today_sales=summary.get("today_sales", 0.0),
                today_transactions=summary.get("today_transactions", 0),
                current_balance=summary.get("current_balance", 0.0),
                active_sessions=summary.get("active_sessions", 0)
            )
        except Exception as exc:
            logger.error(f"GetBranchSummary failed: {exc}")
            logger.error(traceback.format_exc())
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to get branch summary: {str(exc)}")
            return pb2.BranchSummary()
    
    async def GetInventoryStatus(
        self, 
        request: pb2.InventoryQuery, 
        context: grpc.aio.ServicerContext
    ) -> pb2.InventoryStatus:
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
        except Exception as exc:
            logger.error(f"GetInventoryStatus failed: {exc}")
            logger.error(traceback.format_exc())
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Failed to get inventory status: {str(exc)}")
            return pb2.InventoryStatus()
    
    async def SubscribeDashboard(
        self, 
        request: pb2.DashboardSubscription, 
        context: grpc.aio.ServicerContext
    ) -> AsyncIterator[pb2.DashboardUpdate]:
        """Stream dashboard updates."""
        update_count = 0
        max_updates = request.max_updates if request.HasField("max_updates") else 0
        
        try:
            while not context.cancelled():
                try:
                    summary = await self.agent.get_branch_summary()
                    
                    ts = timestamp_pb2.Timestamp()
                    ts.GetCurrentTime()
                    
                    yield pb2.DashboardUpdate(
                        metric_name="today_sales",
                        value=summary.get("today_sales", 0.0),
                        display_value=f"${summary.get('today_sales', 0.0):.2f}",
                        timestamp=ts
                    )
                    
                    update_count += 1
                    if max_updates > 0 and update_count >= max_updates:
                        break
                        
                except asyncio.CancelledError:
                    raise
                except Exception as exc:
                    logger.error(f"SubscribeDashboard iteration failed: {exc}")
                    # Continue to next iteration rather than breaking
                
                # Sleep with cancellation check
                try:
                    await asyncio.wait_for(
                        asyncio.sleep(request.update_interval_ms / 1000.0),
                        timeout=request.update_interval_ms / 1000.0 + 1.0
                    )
                except asyncio.TimeoutError:
                    continue
                    
        except asyncio.CancelledError:
            logger.info("SubscribeDashboard cancelled by client")
            raise
        except Exception as exc:
            logger.error(f"SubscribeDashboard failed: {exc}")
            logger.error(traceback.format_exc())
            context.set_code(grpc.StatusCode.INTERNAL)
            context.set_details(f"Dashboard subscription error: {str(exc)}")
            raise


class LocalAgentGRPCServer:
    """gRPC server for the local agent with enhanced error handling."""
    
    def __init__(self, agent: LocalAgent, port: int = 50051, max_workers: int = 10):
        self.agent = agent
        self.port = port
        self.max_workers = max_workers
        self.server: Optional[grpc.aio.Server] = None
        self._running = False
        self._start_time: Optional[datetime] = None
    
    async def start(self) -> None:
        """Start the gRPC server with enhanced error handling."""
        try:
            self.server = grpc.aio.server(
                futures.ThreadPoolExecutor(max_workers=self.max_workers),
                options=[
                    ('grpc.max_send_message_length', 50 * 1024 * 1024),  # 50MB
                    ('grpc.max_receive_message_length', 50 * 1024 * 1024),  # 50MB
                    ('grpc.keepalive_time_ms', 10000),
                    ('grpc.keepalive_timeout_ms', 5000),
                ]
            )
            
            # Register services
            pb2_grpc.add_AccountingSwarmProtocolServicer_to_server(
                AccountingSwarmServicer(self.agent), self.server
            )
            pb2_grpc.add_QueryProtocolServicer_to_server(
                QueryServicer(self.agent), self.server
            )
            
            # Add health check service (if available)
            try:
                from grpc_health.v1 import health, health_pb2_grpc
                health_servicer = health.HealthServicer()
                health_servicer.set("AccountingSwarmProtocol", health.HealthStatus.SERVING)
                health_servicer.set("QueryProtocol", health.HealthStatus.SERVING)
                health_pb2_grpc.add_HealthServicer_to_server(health_servicer, self.server)
            except ImportError:
                logger.debug("grpc-health-checking not available")
            
            self.server.add_insecure_port(f"[::]:{self.port}")
            await self.server.start()
            
            self._running = True
            self._start_time = datetime.utcnow()
            
            logger.info(f"gRPC server started on port {self.port}")
            
        except Exception as exc:
            logger.error(f"Failed to start gRPC server: {exc}")
            logger.error(traceback.format_exc())
            self._running = False
            raise
    
    async def stop(self, grace_period: Optional[int] = None) -> None:
        """Stop the gRPC server gracefully."""
        if not self.server:
            return
            
        grace = grace_period if grace_period is not None else 5
        
        try:
            logger.info(f"Stopping gRPC server (grace period: {grace}s)...")
            await self.server.stop(grace)
            self._running = False
            logger.info("gRPC server stopped")
        except Exception as exc:
            logger.error(f"Error stopping gRPC server: {exc}")
            raise
    
    async def serve_forever(self) -> None:
        """Run server until interrupted with proper signal handling."""
        try:
            await self.server.wait_for_termination()
        except asyncio.CancelledError:
            logger.info("Server termination cancelled")
            raise
        except Exception as exc:
            logger.error(f"Server terminated with error: {exc}")
            raise
    
    def get_stats(self) -> dict:
        """Get server statistics."""
        return {
            "running": self._running,
            "port": self.port,
            "start_time": self._start_time.isoformat() if self._start_time else None,
            "max_workers": self.max_workers,
        }
