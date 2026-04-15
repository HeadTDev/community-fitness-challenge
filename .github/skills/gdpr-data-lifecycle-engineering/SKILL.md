---
name: gdpr-data-lifecycle-engineering
description: Principal GDPR data lifecycle engineering. Use for deletion/export workflows, retention policies, and privacy-safe system behavior.
---

# GDPR Data Lifecycle Engineering

Implement privacy rights as reliable distributed workflows across all data domains.

## Decision Criteria

- Data lifecycle policies must define collection purpose, retention period, deletion trigger, and deletion evidence.
- GDPR workflows must include all stores and projections, not only primary tables.
- Deletion and export flows require idempotency and replay safety.
- Privacy implementation is accepted only with verifiable completion signals.

## Principal Practices

- Orchestrate deletion via event workflow (`gdpr_delete`) with explicit processing status tracking.
- Remove or anonymize user-linked records according to legal and product constraints while preserving required audit evidence.
- Delete user-owned objects from S3 prefixes (`avatars/`, potential challenge assets ownership cases).
- Reconcile Redis and websocket projections after deletion to prevent ghost identities.

## Failure Modes & Anti-Patterns

- Marking user deleted while leaving linked assets or notifications intact.
- One-shot deletion scripts without retry and observability.
- Hard deletes that destroy required compliance audit trace.
- GDPR workflow that can be triggered without authentication/authorization safeguards.

## Project-Specific Examples

- `DELETE /v1/users/me` must produce auditable workflow status from API acknowledgment to worker completion.
- User deletion should invalidate future leaderboard membership and remove profile artifacts from cached/offline client views.

## Related Skills

- `system-integrity-guardian`
- `worker-reliability-engineering`
- `data-consistency-reconciliation-engineering`
