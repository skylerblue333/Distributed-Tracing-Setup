# Sky Observe

Sky Observe is a bounded, single-node trace intake and query service for the SKYCOIN4444 infrastructure ecosystem.

## Verified capabilities

- concurrent-safe in-memory span storage with a configurable hard cap
- oldest-span eviction when capacity is reached
- validated trace, span, parent, service and operation identifiers
- bounded duration, attribute count, attribute value length and request-body size
- optional constant-time bearer authentication
- trace lookup by trace ID
- health, readiness and operational metrics endpoints
- explicit counters for accepted, rejected and evicted spans
- graceful HTTP shutdown and server timeouts
- Go race-detector and vulnerability-scan CI gates
- non-root distroless container

## HTTP API

- `GET /healthz`
- `GET /readyz`
- `GET /metrics`
- `POST /api/v1/spans`
- `GET /api/v1/traces/{trace_id}`

Example span:

```json
{
  "trace_id": "trace-123",
  "span_id": "span-456",
  "service": "orders",
  "operation": "create_order",
  "duration_ms": 18,
  "timestamp": "2026-08-24T04:00:00Z",
  "status": "ok",
  "attributes": {"region": "us-east"}
}
```

## Run

```bash
go test ./...
go run .
```

Configuration:

- `HTTP_ADDR` defaults to `:8080`
- `OBSERVE_MAX_SPANS` defaults to `10000`
- `OBSERVE_MAX_BODY_BYTES` defaults to `65536`
- `OBSERVE_API_TOKEN` optionally protects trace ingestion/query endpoints

Container:

```bash
docker build -t sky-observe .
docker run --rm -p 8080:8080 sky-observe
```

## Architecture boundary

Sky Observe is intentionally a **local trace service**, not a full observability platform. It does not implement OTLP ingestion/export, W3C trace-context propagation, sampling coordination, durable trace storage, distributed collectors, dashboards, alerting, log aggregation, multi-node replication or managed OpenTelemetry compatibility.

OpenTelemetry Collector/SDK integrations remain the preferred upstream standards for production-scale telemetry. Future integration should use adapters and pinned upstream versions rather than inventing a proprietary telemetry protocol.

See `PRODUCT.md` and `SECURITY.md` for commercial and security boundaries.
