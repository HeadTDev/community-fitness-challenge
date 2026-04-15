---
name: offline-sync-conflict-resolution-engineering
description: Principal offline sync and conflict resolution engineering. Use for SwiftData-backed offline mode and safe replay to backend APIs.
---

# Offline Sync & Conflict Resolution Engineering

Guarantee predictable user outcomes when local actions and remote state diverge.

## Decision Criteria

- Define source of truth per field and conflict winner policy (server-wins, client-wins, merge) before implementation.
- Offline queue design must preserve intent, ordering constraints, and idempotency keys.
- Sync policies must handle partial success and replay after app restart.
- Conflict behavior must be user-visible when automatic resolution cannot preserve intent.

## Principal Practices

- Persist pending operations (`PendingUpload`-style) with retry metadata and correlation IDs.
- Align client retry semantics with backend duplicate protection (daily lock, idempotent worker side effects).
- Keep cache invalidation policy explicit for leaderboard and profile updates.
- Reconcile websocket updates against local cache snapshots without destructive overwrite.

## Failure Modes & Anti-Patterns

- Blindly replaying stale operations after user state changed remotely.
- Treating local cache as authoritative for mutable server-derived rankings.
- Dropping failed sync operations without user/system visibility.
- Conflict handling that mutates user-entered data silently.

## Project-Specific Examples

- Offline daily log submissions must avoid duplicate score impact when network returns.
- Profile/avatar updates should reconcile local optimistic state with eventual S3 URL persistence from backend.

## Related Skills

- `cross-platform-system-design-authority`
- `event-driven-workflow-orchestration`
- `real-time-client-experience-engineering`
