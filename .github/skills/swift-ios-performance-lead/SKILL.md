---
name: swift-ios-performance-lead
description: Principal iOS performance engineering for SwiftUI, HealthKit, websocket, and offline-first flows.
---

# Principal Swift iOS Performance Engineering

Deliver resilient, low-latency iOS experiences under real network variability and high-frequency real-time updates.

## Decision Criteria

- Prefer architecture choices that preserve correctness under offline mode, delayed server updates, and websocket reconnects.
- Optimize for deterministic state transitions between API responses, local cache mutations, and live event updates.
- Any performance optimization must keep battery impact and memory growth within explicit limits.
- UI behavior must remain contract-consistent with backend score/rank semantics.

## Principal Practices

- Use structured state containers to avoid duplicate renders during rapid leaderboard updates.
- Implement adaptive update cadence for websocket-driven UI changes to prevent animation thrash.
- Keep SwiftData cache models schema-versioned with migration-safe defaults.
- Separate health data ingestion from score presentation to support anti-cheat and reconciliation updates.

## Failure Modes & Anti-Patterns

- UI depending on immediate server consistency for asynchronous workflows.
- Cache writes and websocket updates racing without conflict resolution policy.
- HealthKit queries triggered too frequently, causing battery drain and noisy data.
- Token refresh handling that fails silently and leaves stale authenticated state.

## Project-Specific Examples

- Leaderboard UI must gracefully handle rank updates arriving while offline-cached challenge data is displayed.
- Daily log submission UX should remain coherent when backend accepts log but scoring/notification arrives asynchronously.

## Related Skills

- `offline-sync-conflict-resolution-engineering`
- `real-time-client-experience-engineering`
- `mobile-security-privacy-engineering`
