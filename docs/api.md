# Broker API

All API paths are currently versioned under `/api/v1`.

## Health

```http
GET /healthz
```

Response:

```json
{
  "ok": true
}
```

## Register Volunteer

```http
POST /api/v1/volunteers/register
Authorization: Bearer <registration-token>
Content-Type: application/json
```

Request:

```json
{
  "public_host": "2001:db8::1",
  "public_port": 443,
  "protocol": "vless-reality-vision",
  "client_id": "2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff",
  "reality_public_key": "xray-public-key",
  "short_id": "5f7a8d9c01ab23cd",
  "server_name": "www.microsoft.com",
  "flow": "xtls-rprx-vision",
  "exit_mode": "direct",
  "max_sessions": 8,
  "max_mbps": 20,
  "volunteer_version": "dev"
}
```

Response:

```json
{
  "id": "relay_...",
  "public_host": "2001:db8::1",
  "public_port": 443,
  "protocol": "vless-reality-vision",
  "client_id": "2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff",
  "reality_public_key": "xray-public-key",
  "short_id": "5f7a8d9c01ab23cd",
  "server_name": "www.microsoft.com",
  "flow": "xtls-rprx-vision",
  "exit_mode": "direct",
  "max_sessions": 8,
  "max_mbps": 20,
  "volunteer_version": "dev",
  "registered_at": "2026-06-09T07:00:00Z",
  "last_heartbeat_at": "2026-06-09T07:00:00Z",
  "expires_at": "2026-06-09T07:03:00Z"
}
```

## Heartbeat

```http
POST /api/v1/volunteers/{id}/heartbeat
Authorization: Bearer <registration-token>
```

Response:

```json
{
  "ok": true,
  "expires_at": "2026-06-09T07:03:30Z"
}
```

## List Relays

```http
GET /api/v1/relays?limit=5
```

Response:

```json
{
  "count": 1,
  "server_time": "2026-06-09T07:00:00Z",
  "relays": [
    {
      "id": "relay_...",
      "public_host": "2001:db8::1",
      "public_port": 443,
      "protocol": "vless-reality-vision",
      "client_id": "2c08df10-4ef4-4ab9-95c6-cb1e94cdb2ff",
      "reality_public_key": "xray-public-key",
      "short_id": "5f7a8d9c01ab23cd",
      "server_name": "www.microsoft.com",
      "flow": "xtls-rprx-vision",
      "exit_mode": "direct",
      "max_sessions": 8,
      "max_mbps": 20,
      "volunteer_version": "dev",
      "registered_at": "2026-06-09T07:00:00Z",
      "last_heartbeat_at": "2026-06-09T07:00:00Z",
      "expires_at": "2026-06-09T07:03:00Z"
    }
  ]
}
```
