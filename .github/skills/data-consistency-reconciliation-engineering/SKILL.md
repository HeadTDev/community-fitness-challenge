---
name: data-consistency-reconciliation-engineering
description: Principal consistency and reconciliation engineering across PostgreSQL, Redis, and asynchronous workers.
---

# Data Consistency & Reconciliation Engineering

Guarantee recoverable consistency when fast paths and source-of-truth paths diverge.

## Decision Criteria

- Define authoritative store per datum and projection stores explicitly.
- Set acceptable staleness windows per endpoint and event stream.
- Require reconciliation job design for every projection cache introduced.
- Choose consistency strategy per operation: strong (transactional), eventual (replay/rebuild), or compensating.

## Principal Practices

- Build deterministic rebuild routines for leaderboard and participant counters from PostgreSQL records.
- Use reconciliation metrics (mismatch counts, drift age) as first-class operational signals.
- Keep idempotent repair operations so repeated runs are safe.
- Gate release of high-risk changes with reconciliation dry-runs against production-like datasets.

## Failure Modes & Anti-Patterns

- Hand-maintained cache fixes without durable repair automation.
- Rebuild jobs that depend on ephemeral in-memory ordering.
- Repair routines that mutate source-of-truth data to match stale cache.
- Ignoring clock and timezone effects in daily partition logic.

## Project-Specific Examples

- `make leaderboard-rebuild` must restore Redis ranking from PostgreSQL after Redis outage and preserve deterministic tie handling.
- Challenge participant count in Redis and challenge table must be reconciled after worker/API crash windows.

## Related Skills

- `distributed-systems-optimizer`
- `worker-reliability-engineering`
- `disaster-recovery-business-continuity`
