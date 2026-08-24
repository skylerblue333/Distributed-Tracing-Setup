# Security Policy

Sky Observe accepts telemetry supplied by clients, so every deployment should treat incoming spans as untrusted data.

## Implemented controls

- request-body size limit
- bounded trace identifiers and attributes
- bounded in-memory storage
- optional constant-time bearer token comparison
- HTTP read/write/header timeouts
- non-root distroless container
- Go race detector and `govulncheck` in CI

## Deployment requirements

- configure `OBSERVE_API_TOKEN` when the intake/query API is exposed outside a trusted boundary
- terminate TLS using a trusted reverse proxy or service mesh
- do not place secrets, credentials, payment data or unnecessary PII in span attributes
- apply network and container resource limits appropriate to the environment

## Not provided

The service does not provide tenant isolation, durable encrypted storage, field-level redaction, TLS termination, OTLP authentication, distributed authorization, replication or compliance certification.

Report suspected vulnerabilities through the repository's GitHub security/reporting channel without including production secrets in public issues.
