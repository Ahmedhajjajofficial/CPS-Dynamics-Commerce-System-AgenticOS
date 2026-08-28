"""
HTTP Gateway for Local Agent
=============================
Exposes REST/JSON endpoints for the POS interface and other HTTP clients.

This gateway translates HTTP requests into Local Agent operations,
allowing browser-based POS terminals to interact with the agent
without requiring gRPC-Web or a separate proxy.
"""

from __future__ import annotations

import asyncio
import json
import os
import uuid
from datetime import datetime
from typing import Any, Dict, Optional

from aiohttp import web

from .agent import LocalAgent, AgentConfig, AgentState


class HTTPGateway:
    """Lightweight HTTP gateway for the Local Agent."""

    def __init__(self, agent: LocalAgent, host: str = "0.0.0.0", port: int = 8080):
        self.agent = agent
        self.host = host
        self.port = port
        self.app = web.Application()
        self._setup_routes()
        self.runner: Optional[web.AppRunner] = None
        self.site: Optional[web.TCPSite] = None

    def _setup_routes(self) -> None:
        self.app.add_routes([
            web.get("/health", self._health),
            web.post("/api/v1/sessions", self._start_session),
            web.post("/api/v1/sessions/{session_id}/close", self._close_session),
            web.post("/api/v1/sales", self._record_sale),
            web.get("/api/v1/summary", self._branch_summary),
            web.get("/api/v1/products", self._list_products),
            web.get("/api/v1/inventory/{product_id}", self._get_inventory),
        ])

    async def _health(self, request: web.Request) -> web.Response:
        return web.json_response({
            "status": "ok",
            "agent_id": self.agent.config.agent_id,
            "branch_id": self.agent.config.branch_id,
            "agent_state": self.agent.state.value,
            "timestamp": datetime.utcnow().isoformat(),
        })

    async def _start_session(self, request: web.Request) -> web.Response:
        try:
            payload = await request.json()
            cashier_id = payload.get("cashier_id", "unknown")
            register_id = payload.get("register_id", "unknown")
            opening_balance = float(payload.get("opening_balance", 0))

            session_id = await self.agent.start_sales_session(
                cashier_id=cashier_id,
                register_id=register_id,
                opening_balance=opening_balance,
            )

            return web.json_response({
                "success": True,
                "data": {
                    "session_id": session_id,
                    "cashier_id": cashier_id,
                    "register_id": register_id,
                    "opening_balance": opening_balance,
                    "started_at": datetime.utcnow().isoformat(),
                    "status": "active",
                },
            }, status=201)
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=400)

    async def _close_session(self, request: web.Request) -> web.Response:
        session_id = request.match_info["session_id"]
        try:
            payload = await request.json()
            closing_balance = float(payload.get("closing_balance", 0))
            total_sales = float(payload.get("total_sales", 0))
            transaction_count = int(payload.get("transaction_count", 0))

            event = await self.agent.close_sales_session(
                session_id=session_id,
                closing_balance=closing_balance,
                total_sales=total_sales,
                transaction_count=transaction_count,
            )

            return web.json_response({
                "success": True,
                "data": {
                    "session_id": session_id,
                    "closing_balance": closing_balance,
                    "total_sales": total_sales,
                    "transaction_count": transaction_count,
                    "closed_at": datetime.utcnow().isoformat(),
                    "event_id": event.event_id,
                },
            })
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=400)

    async def _record_sale(self, request: web.Request) -> web.Response:
        try:
            payload = await request.json()
            product_id = payload.get("product_id")
            quantity = int(payload.get("quantity", 1))
            unit_price = float(payload.get("unit_price", 0))
            total_amount = float(payload.get("total_amount", 0))
            cashier_id = payload.get("cashier_id", "unknown")
            session_id = payload.get("session_id")
            customer_id = payload.get("customer_id")
            payment_method = payload.get("payment_method", "cash")

            if not product_id or not session_id:
                return web.json_response({
                    "success": False,
                    "error": "product_id and session_id are required",
                }, status=400)

            event = await self.agent.record_sale(
                product_id=product_id,
                quantity=quantity,
                unit_price=unit_price,
                total_amount=total_amount,
                cashier_id=cashier_id,
                session_id=session_id,
                customer_id=customer_id,
                payment_method=payment_method,
            )

            return web.json_response({
                "success": True,
                "data": {
                    "event_id": event.event_id,
                    "event_type": event.event_type,
                    "stream_id": event.stream_id,
                    "created_at": event.created_at.isoformat() if event.created_at else None,
                },
            }, status=201)
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=400)

    async def _branch_summary(self, request: web.Request) -> web.Response:
        try:
            summary = await self.agent.get_branch_summary()
            return web.json_response({
                "success": True,
                "data": summary,
            })
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=500)

    async def _list_products(self, request: web.Request) -> web.Response:
        try:
            products = await self.agent.get_sales_history(limit=1000)
            return web.json_response({
                "success": True,
                "data": [
                    {
                        "id": p.stream_id,
                        "name": p.event_type,
                        "price": 0,
                        "category": "Unknown",
                        "taxRate": 0,
                        "isActive": True,
                        "stockQuantity": 0,
                    }
                    for p in products
                ],
            })
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=500)

    async def _get_inventory(self, request: web.Request) -> web.Response:
        product_id = request.match_info["product_id"]
        try:
            quantity = self.agent.get_inventory_level(product_id)
            return web.json_response({
                "success": True,
                "data": {
                    "product_id": product_id,
                    "current_quantity": quantity,
                    "available_quantity": max(0, quantity),
                    "is_low_stock": quantity < 10,
                },
            })
        except Exception as exc:
            return web.json_response({
                "success": False,
                "error": str(exc),
            }, status=500)

    async def start(self) -> None:
        self.runner = web.AppRunner(self.app)
        await self.runner.setup()
        self.site = web.TCPSite(self.runner, self.host, self.port)
        await self.site.start()

    async def stop(self) -> None:
        if self.site:
            await self.site.stop()
        if self.runner:
            await self.runner.cleanup()
