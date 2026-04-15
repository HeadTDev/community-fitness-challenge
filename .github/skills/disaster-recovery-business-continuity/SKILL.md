---
name: disaster-recovery-business-continuity
description: Principal disaster recovery and continuity engineering. Use for backup, restore, failover, and recovery drills across core services.
---

# Disaster Recovery & Business Continuity

Design recovery so major outages degrade service predictably and restore safely.

## Decision Criteria

- Define RPO/RTO targets per domain: auth, challenge state, daily logs, leaderboard projections, notification state.
- Recovery plans must cover data stores, message queues, object storage, and deployment control plane.
- Every backup strategy must include restore verification cadence.
- Continuity design must preserve user trust-critical features during partial outages.

## Principal Practices

- Maintain PostgreSQL backup + restore runbooks with schema version compatibility checks.
- Treat Redis as recoverable projection; prioritize rebuild speed and correctness over persistence assumptions.
- Keep queue replay and DLQ handling procedures documented and tested.
- Run failure drills that simulate Redis loss, DB restore, and worker outage scenarios.

## Failure Modes & Anti-Patterns

- Backups never tested for restore validity.
- Recovery procedures requiring tribal knowledge or specific individuals.
- Replaying jobs without idempotency safeguards.
- Restoring data without reconciling derived caches and websocket projections.

## Project-Specific Examples

- After PostgreSQL restore, run leaderboard reconstruction and participant count reconciliation before reopening live rank updates.
- After queue outage, replay pending `log_submitted` safely and verify anti-cheat + scoring consistency.

## Related Skills

- `data-consistency-reconciliation-engineering`
- `incident-response-leadership`
- `release-management-governance`
