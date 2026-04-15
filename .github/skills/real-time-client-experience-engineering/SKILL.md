---
name: real-time-client-experience-engineering
description: Principal real-time client experience engineering. Use for websocket-driven UI updates, animation correctness, and latency-resilient UX.
---

# Real-Time Client Experience Engineering

Deliver live interactions that remain coherent under network jitter, event bursts, and reconnect cycles.

## Decision Criteria

- Real-time updates must have deterministic merge behavior with existing on-screen state.
- UX should define graceful degradation for delayed or missing events.
- Animation strategy must prioritize clarity over visual noise under high update frequency.
- Reconnect behavior must specify stale-state indicators and rehydration path.

## Principal Practices

- Use event sequencing/versioning checks before mutating leaderboard state.
- Batch or debounce frequent updates to avoid render thrash while preserving critical changes.
- Keep user-centric context stable (my rank and nearby competitors) during live reorder operations.
- Instrument client-side latency from event receipt to visible render.

## Failure Modes & Anti-Patterns

- Applying events blindly out of order.
- UI jumpiness from full list re-render on every rank update.
- No stale-data indicator during websocket disconnect.
- Coupling animation timing directly to variable network delays.

## Project-Specific Examples

- Relative leaderboard view should preserve focus around current user while handling incoming rank updates.
- Notification center badges should update from live events without conflicting with offline cached counts.

## Related Skills

- `real-time-websocket-systems-engineering`
- `swift-ios-performance-lead`
- `offline-sync-conflict-resolution-engineering`
