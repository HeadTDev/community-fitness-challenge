---
name: api-contract-governance
description: Principal API contract governance for Gin-based endpoints. Use to control compatibility, envelope consistency, and backend-client alignment.
---

# API Contract Governance

Keep API behavior predictable across backend, worker side effects, and iOS clients by enforcing explicit contract governance.

## Decision Criteria

- Every endpoint must conform to the shared response envelope (`success`, `data|error`, `meta.timestamp`, `meta.request_id`).
- Contract changes are backward-compatible by default; breaking changes require versioned route strategy and migration plan.
- Error code names are part of the public contract and must remain stable once shipped.
- Any payload used by iOS offline cache or websocket hydration requires compatibility window definition.

## Principal Practices

- Keep `/v1` resource naming and auth model consistent across challenge, profile, leaderboard, notification, and future daily-log APIs.
- Require explicit request validation rules aligned with model constraints (date ordering, positive ranges, enum-safe status/type fields).
- Define pagination/sorting determinism for list endpoints before scale work starts.
- Document event-to-API coupling (for example log submission triggers worker actions that affect leaderboard and notifications).

## Failure Modes & Anti-Patterns

- Returning raw internal errors in API responses.
- Silent shape drift between POST response DTOs and GET DTOs for the same resource.
- Mixing transport layer status codes with domain semantics inconsistently.
- Introducing optional fields without nullability semantics and migration behavior.

## Project-Specific Examples

- `POST /auth/register-dev` remains development-only; production behavior must be explicit and non-accidental.
- `GET /v1/challenges/:id/leaderboard` and `/relative` must keep consistent score/rank field semantics across Redis and PostgreSQL fallback paths.

## Related Skills

- `go-hexagonal-architect`
- `cross-platform-system-design-authority`
- `integration-contract-testing-engineering`
