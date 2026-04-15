---
name: go-hexagonal-architect
description: Principal-level authority for hexagonal architecture in this Go backend. Use for boundary decisions, transaction orchestration, and long-term domain integrity.
---

# Principal Go Hexagonal Architecture

Enforce strict boundary ownership across `internal/domain`, `internal/services`, `internal/adapter`, and `internal/handler` while optimizing for long-term change velocity.

## Decision Criteria

- Put invariants (status transitions, role checks, scoring eligibility) in domain/service code, never in Gin handlers or SQL strings.
- Keep handlers responsible for transport-only concerns: bind/validate input, map domain errors to `response.Error`, emit `response.Success`.
- Start transactions only at service orchestration boundaries (for example join/leave challenge flows spanning participation + challenge count).
- Make adapters map persistence and infra errors into domain sentinels (`ErrNotFound`, `ErrBadRequest`, `ErrUnauthorized`) so handler behavior stays stable.

## Principal Practices

- Every new external dependency (S3, SQS, Redis, SES, Secrets) must enter through an interface port before touching service logic.
- Preserve request context propagation from handler -> service -> adapters for cancellation and deadline correctness.
- Centralize dependency wiring in `internal/app/app.go`; never instantiate infra clients inside handlers.
- Enforce soft-delete semantics consistently in repositories (`deleted_at IS NULL`) so list/get behavior is deterministic.

## Failure Modes & Anti-Patterns

- Handler-to-repository direct calls for business mutations (skipping services) cause duplicated authorization and race-prone logic.
- Domain packages importing adapter packages leaks infrastructure concerns and blocks test isolation.
- Cross-repository write sequences without transaction wrappers produce diverged participant counts and stale leaderboard state.
- Returning raw SQL/SDK errors to handlers breaks stable API error codes and contract predictability.

## Project-Specific Examples

- Daily log submission path: handler validates payload, service applies lock/idempotency/scoring rules, repositories persist, then event is emitted to SQS.
- Challenge join flow: service transaction writes participation, recomputes participant count, updates challenge, commits, then increments Redis counter key.

## Related Skills

- `distributed-systems-optimizer`
- `event-driven-workflow-orchestration`
- `data-consistency-reconciliation-engineering`
