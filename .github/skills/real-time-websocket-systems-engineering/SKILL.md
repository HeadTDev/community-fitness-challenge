---
name: real-time-websocket-systems-engineering
description: Principal engineering for real-time WebSocket systems. Use for hub topology, Redis Pub/Sub bridge, and live ranking events.
---

# Real-Time WebSocket Systems Engineering

Deliver low-latency real-time updates with deterministic event semantics and failure-aware client recovery.

## Decision Criteria

- Define event contracts before implementation (`rank_update`, connection state, heartbeat behavior).
- Decide room model and fanout topology from subscription cardinality and update frequency.
- Require reconnect strategy with replay policy for missed events.
- Enforce authentication model consistency between HTTP and websocket channels.

## Principal Practices

- Build `WSHub` with bounded client queues, heartbeat timeouts, and backpressure handling.
- Bridge Redis Pub/Sub to websocket broadcasts through typed event mapping, not raw payload passthrough.
- Preserve monotonic ordering per challenge stream where feasible; include sequence metadata when not guaranteed.
- Track per-room connection counts and broadcast latency metrics.

## Failure Modes & Anti-Patterns

- Unbounded per-client channels that crash process under burst.
- Broadcasting internal payloads directly without contract validation.
- Ignoring reconnect storms after Redis or API restart.
- Treating websocket as source of truth rather than a projection channel.

## Project-Specific Examples

- `log_submitted` processing should publish a rank update event that the iOS client can apply to absolute and relative leaderboard views.
- Nginx `/ws/*` proxy settings must preserve upgrade headers and idle timeout values aligned with ping/pong cadence.

## Related Skills

- `distributed-systems-optimizer`
- `cross-platform-system-design-authority`
- `observability-engineering`
