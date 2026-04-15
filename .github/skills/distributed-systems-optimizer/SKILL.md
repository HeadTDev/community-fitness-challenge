---
name: distributed-systems-optimizer
description: Principal-level distributed systems optimization for Redis/PostgreSQL workloads. Use for ranking pipelines, consistency control, and scale-safe performance.
---

# Principal Distributed Systems Optimization

Optimize throughput and tail-latency while preserving correctness across Redis, PostgreSQL, workers, and websocket fanout.

## Decision Criteria

- Use Redis for hot ranking paths only when rebuild and fallback semantics are explicitly defined.
- Keep PostgreSQL authoritative for audit and recomputation workflows.
- Choose update strategy by contention pattern: direct `ZINCRBY` for single-key updates, Lua/multi-step atomic scripts for coupled updates.
- Require bounded staleness contracts for every cache-backed endpoint.

## Principal Practices

- Maintain deterministic leaderboard reconstruction command path from `daily_logs`/`participations` in PostgreSQL.
- Design Redis keys with explicit namespace/versioning to support migration without destructive cutovers.
- Prevent double-apply in worker-driven score updates through idempotency keys and durable processing state.
- Enforce backpressure control for websocket fanout when rank events spike.

## Failure Modes & Anti-Patterns

- Treating Redis values as source-of-truth without reconciliation path.
- Running fallback SQL without required indexes and then calling it "high availability."
- Updating Redis and PostgreSQL in opposite order without replay policy.
- Ignoring clock skew and out-of-order message delivery in event-driven rank updates.

## Project-Specific Examples

- `challenge_count:%s` and leaderboard keys must be rebuilt from PostgreSQL when Redis restarts or eviction occurs.
- `GET /v1/challenges/:id/leaderboard` must remain correct when Redis is unavailable by switching to indexed PostgreSQL ranking query.

## Related Skills

- `data-consistency-reconciliation-engineering`
- `worker-reliability-engineering`
- `real-time-websocket-systems-engineering`
