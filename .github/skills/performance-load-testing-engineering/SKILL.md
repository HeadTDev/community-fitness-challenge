---
name: performance-load-testing-engineering
description: Principal performance and load testing engineering. Use for latency/SLA protection across API, DB, Redis, and worker paths.
---

# Performance & Load Testing Engineering

Protect user experience and system stability under realistic and adversarial load.

## Decision Criteria

- Define SLOs and saturation signals before selecting test scenarios.
- Load profiles must reflect real flows: bursty log submissions, leaderboard reads, websocket fanout spikes.
- Performance changes are accepted only with comparative baseline evidence.
- Test plans must include dependency degradation scenarios.

## Principal Practices

- Benchmark critical endpoints and service methods with controlled data volume growth.
- Exercise DB + Redis concurrency with representative key distributions.
- Measure queue lag and worker throughput under peak ingestion.
- Track p50/p95/p99 latency, error rate, and resource saturation jointly.

## Failure Modes & Anti-Patterns

- Single-endpoint microbenchmarks presented as system readiness.
- Ignoring lock contention and transaction retry behavior.
- Load tests without realistic data cardinality (few users/challenges only).
- Treating websocket throughput as "best effort" without objective limits.

## Project-Specific Examples

- Stress `POST daily-log` path with duplicate submissions and anti-cheat pipeline enabled to validate lock and queue behavior.
- Run leaderboard read load during concurrent score updates to validate rank stability and fallback query performance.

## Related Skills

- `distributed-systems-optimizer`
- `observability-engineering`
- `finops-capacity-engineering`
