"""
Unit tests for HTTP Gateway endpoints.
"""

import asyncio
import json
import os
import sys
import pytest
from aiohttp import web
from aiohttp.test_utils import AioHTTPTestCase, unittest_run_loop

sys.path.insert(0, os.path.join(os.path.dirname(__file__), "src"))

from agent import LocalAgent, AgentConfig
from http_gateway import HTTPGateway


class TestHTTPGateway(AioHTTPTestCase):
    async def get_application(self):
        db_path = os.path.join(os.path.dirname(__file__), "test_http_gateway.db")
        cfg = AgentConfig(
            agent_id="agent-http-test",
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
        self.agent = LocalAgent(cfg)
        await self.agent.initialize()
        self.gateway = HTTPGateway(self.agent, host="127.0.0.1", port=0)
        return self.gateway.app

    async def tearDown(self):
        await self.gateway.stop()
        await self.agent.shutdown()
        if os.path.exists(self.agent.config.db_path):
            os.remove(self.agent.config.db_path)
        await super().tearDown()

    @unittest_run_loop
    async def test_health(self):
        resp = await self.client.request("GET", "/health")
        assert resp.status == 200
        data = await resp.json()
        assert data["status"] == "ok"
        assert data["agent_id"] == "agent-http-test"
        assert data["branch_id"] == "BR001"

    @unittest_run_loop
    async def test_start_session(self):
        payload = json.dumps({
            "cashier_id": "c1",
            "register_id": "REG1",
            "opening_balance": 100.0,
        })
        resp = await self.client.request("POST", "/api/v1/sessions", data=payload, headers={"Content-Type": "application/json"})
        assert resp.status == 201
        data = await resp.json()
        assert data["success"] is True
        assert "session_id" in data["data"]

    @unittest_run_loop
    async def test_record_sale_requires_fields(self):
        payload = json.dumps({})
        resp = await self.client.request("POST", "/api/v1/sales", data=payload, headers={"Content-Type": "application/json"})
        assert resp.status == 400
        data = await resp.json()
        assert data["success"] is False
