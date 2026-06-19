# CP'S Enterprise DCS - REST API Specification (v4.0)

This document defines the RESTful gateway layer for the DCS. While the core communication between agents uses gRPC, this REST API provides a compatible interface for the POS (React/TS) and external integrations.

## Base URL
`http://{agent_host}:{agent_port}/api/v1`

- Local Agent default: `http://localhost:50051` (Proxy/REST Bridge)
- Regional Agent default: `http://localhost:50052` (Proxy/REST Bridge)

## Authentication
All requests must include a Bearer Token in the header:
`Authorization: Bearer <JWT_TOKEN>`

---

## 1. Events (AccountingSwarmProtocol)

### Broadcast Event
Broadcasts a new financial event to the swarm.
- **URL**: `/events`
- **Method**: `POST`
- **Payload**: `SovereignFinancialEvent` (JSON mapping)

**Sample Request**:
```json
{
  "type": "SALE_COMPLETED",
  "branch_id": "BR001",
  "amount": 150.00,
  "currency": "SAR",
  "payload": {
    "encrypted_data": "...",
    "kms_key_id": "vault-key-1"
  }
}
```

### Get Event
- **URL**: `/events/{event_id}`
- **Method**: `GET`

---

## 2. Queries (QueryProtocol)

### Branch Summary
- **URL**: `/branches/{branch_id}/summary`
- **Method**: `GET`

### Inventory Status
- **URL**: `/branches/{branch_id}/inventory/{product_id}`
- **Method**: `GET`

### Sales Report
- **URL**: `/branches/{branch_id}/reports/sales`
- **Method**: `GET`
- **Params**: `start_date`, `end_date`, `group_by`

---

## 3. Agents

### Registration
- **URL**: `/agents/register`
- **Method**: `POST`

### Heartbeat
- **URL**: `/agents/{agent_id}/heartbeat`
- **Method**: `PUT`
