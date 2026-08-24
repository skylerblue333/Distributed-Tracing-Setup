# Sky Observe Product Definition

**Product number:** 11 in the SKYCOIN4444 standalone-product master plan.

Sky Observe provides a compact trace-intake and inspection service for development, test and small single-node deployments. Its sellable value is the hardened boundary around bounded telemetry ingestion, operational visibility and reusable integration into larger platforms.

## Included

- validated span intake API
- bounded concurrent in-memory storage
- trace lookup
- health/readiness/metrics
- optional bearer protection
- deterministic tests, race detection, vulnerability scanning and non-root container packaging

## Explicit non-claims

Sky Observe is not the OpenTelemetry Collector, Jaeger, Tempo, Elasticsearch, a distributed tracing database, a log platform, an alert manager or a multi-region observability SaaS. Durable storage, OTLP compatibility and distributed collection require separate adapters and evidence.
