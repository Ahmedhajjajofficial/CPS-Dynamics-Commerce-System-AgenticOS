# CP'S Enterprise DCS - WebSocket Events (v4.0)

WebSocket streams provide real-time updates for the POS Dashboard and agent coordination.

## Connection
`ws://{agent_host}:{agent_port}/ws/v1`

---

## 1. Topics

### `dashboard.updates.{branch_id}`
Real-time stream of `DashboardUpdate` messages. Used for live sales counters and status alerts.
- **Protocol mapping**: gRPC `QueryProtocol.SubscribeDashboard`

### `events.stream.{branch_id}`
Real-time stream of all financial events happening at the branch or region.
- **Protocol mapping**: gRPC `AccountingSwarmProtocol.SubscribeEvents`

---

## 2. Message Format
All messages are JSON objects with a `topic` and `data` field.

**Sample Dashboard Update**:
```json
{
  "topic": "dashboard.updates.BR001",
  "data": {
    "metric_name": "today_sales",
    "value": 150.00,
    "display_value": "$150.00",
    "timestamp": "2024-04-29T10:00:00Z"
  }
}
```

---

## 3. Client Commands

### `subscribe`
Subscribe to a specific topic.
```json
{
  "command": "subscribe",
  "topic": "dashboard.updates.BR001"
}
```

### `unsubscribe`
Unsubscribe from a topic.
```json
{
  "command": "unsubscribe",
  "topic": "dashboard.updates.BR001"
}
```
