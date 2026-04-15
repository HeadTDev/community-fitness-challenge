---
name: cross-platform-system-design-authority
description: Principal authority for backend-iOS system design. Use for end-to-end flows crossing API, worker, websocket, and client state.
---

# Cross-Platform System Design Authority

Design backend and iOS behavior as one system, not two independent implementations.

## Decision Criteria

- Every backend contract change must include iOS impact analysis (state model, cache model, websocket handling, auth flow).
- Prioritize deterministic behavior under offline/online transitions and delayed worker side effects.
- Evaluate latency budget for rank updates across API write -> SQS/Redis -> websocket -> SwiftUI render path.
- Require fallback UX for every server-side asynchronous step.

## Principal Practices

- Define canonical timeline for key journeys: sign-in, join challenge, submit log, rank update, notification delivery.
- Use shared event vocabulary (`log_submitted`, rank update, notification read state) across backend and client.
- Align idempotency semantics between backend write APIs and iOS retry behavior for weak network conditions.
- Ensure authentication/session policies (token refresh, expiry handling) are consistent across API and websocket channels.

## Failure Modes & Anti-Patterns

- Backend emits event payloads the iOS client cannot reconcile with local cache models.
- Client assumes immediate consistency while backend is explicitly eventual via worker pipelines.
- Designing websocket messages without reconnect/replay strategy.
- Ignoring timezone and locale effects in daily-log and challenge-window computations.

## Project-Specific Examples

- Offline log capture must preserve upload intent (`PendingUpload`) and resolve duplicate submissions via backend daily lock rules.
- Creator dashboard publish actions must match backend state machine and return status transitions that UI can animate safely.

## Related Skills

- `api-contract-governance`
- `real-time-client-experience-engineering`
- `offline-sync-conflict-resolution-engineering`
