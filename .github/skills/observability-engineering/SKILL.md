---
name: observability-engineering
description: Principal observability engineering for API, worker, and real-time pipelines.
---

# Observability Engineering

Make system behavior explainable in real time across request paths and asynchronous pipelines.

## Decision Criteria

- Instrumentation must support debugging by request, user, challenge, and job correlation identifiers.
- Logging, metrics, and traces should target actionable questions, not vanity dashboards.
- Alert thresholds must map to user-impacting conditions and error budgets.
- Observability changes are required for every new high-risk feature.

## Principal Practices

- Preserve request ID propagation from middleware into downstream logs and event payload correlation fields.
- Emit structured logs for API, worker handlers, queue processing, and websocket broadcasts.
- Track key metrics: queue lag, processing latency, fallback frequency, reconciliation drift, auth failures.
- Build dashboards that separate correctness failures from performance degradation.

## Failure Modes & Anti-Patterns

- Logs without identifiers needed for cross-service tracing.
- Alert storms from noisy low-signal metrics.
- Missing instrumentation on fallback paths (the exact path used in incidents).
- Relying on manual log scraping instead of durable metrics for SLO tracking.

## Project-Specific Examples

- Worker logs for `log_submitted` should include user/challenge IDs, suspicion outcome, and score impact decision.
- Leaderboard endpoint telemetry should expose whether response came from Redis or PostgreSQL fallback.

## Related Skills

- `incident-response-leadership`
- `worker-reliability-engineering`
- `technical-risk-tradeoff-management`
