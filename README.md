# Distributed Tracing Setup

Reusable OpenTelemetry tracing adapter for SKYCOIN4444 services.

## Current implementation

- OpenTelemetry Go tracer factory
- `OTEL_SERVICE_NAME` support with a SKYCOIN4444 default
- Explicit span creation helper with input validation
- OpenTelemetry HTTP instrumentation adapter
- Unit tests for tracing validation and HTTP instrumentation behavior

## Ecosystem role

**Infrastructure / Observability → Distributed Tracing Boundary**

This repository is a reusable instrumentation component. It is **not** a complete observability platform and does not claim to provide a running collector, storage backend, dashboards, alerting, or production telemetry pipeline by itself.

## Why this is commercially reusable

The component can serve as the tracing foundation for enterprise starter kits covering APIs, ShadowChat, HopeAI, wallets, payments, marketplace workflows, and background jobs. The commercial value comes from integration quality, standardized telemetry, deployment configuration, security, documentation, and actual adoption—not from repository size or generic enterprise claims.

## Truthful status

- OpenTelemetry adapter: **implemented**
- HTTP instrumentation wrapper: **implemented**
- Tests: **implemented**
- Collector/backend: **not included**
- Dashboards/alerts: **not included**
- Production telemetry pipeline: **not verified**
- Paying customers: **not verified**
- ARR/revenue: **not claimed**

The previous README used broad “professional-grade” and “enterprise-level” claims while the repository audit showed only a small Go implementation and supporting scaffolding. This README reports the verified capability instead. fileciteturn308file0

## Open-source foundation

This implementation intentionally builds on the OpenTelemetry Go ecosystem rather than inventing a proprietary tracing protocol. Third-party licenses and notices must remain compliant with the dependency licenses.

## Production roadmap

1. Standardize trace/resource attributes across SKYCOIN4444 services.
2. Add OTLP exporter configuration.
3. Deploy a controlled OpenTelemetry Collector.
4. Add metrics/log correlation.
5. Add sampling and PII-safe attribute policies.
6. Add integration tests for propagated trace context.
7. Add dashboards and service-level alerts.
8. Add deployment/rollback verification.
9. Consolidate the adapter into SKYCOIN4444 Infrastructure.

## License

See the checked-in repository license and applicable third-party dependency licenses.
